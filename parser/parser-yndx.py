# -*- coding: utf-8 -*-
"""
Парсер каталога Яндекс Практикум (https://practicum.yandex.ru/catalog/).

Стратегия: извлекаем window.__preloadedData__ из HTML — там лежит готовый JSON
со всеми курсами, тегами, ценами и датами старта.

Собирает: название, цена, ближайший старт, длительность, уровень,
категория (специализация), теги (skill_tags), ссылку на курс.

Результат синхронизируется с БД MoneyApp:
  - providers        -> upsert "Яндекс Практикум"
  - course_categories -> upsert категории
  - skill_tags        -> upsert теги
  - courses           -> upsert курсы (по external_url)
  - course_skill_tags -> связь курс <-> тег
"""

import json
import os
import re
import sys
import time
import uuid
import logging
from datetime import datetime, date
from pathlib import Path
from urllib.parse import urljoin

from dotenv import load_dotenv

# Загружаем .env из корня проекта (MoneyApp/.env)
_project_root = Path(__file__).resolve().parent.parent
_env_path = _project_root / ".env"
load_dotenv(_env_path)

import psycopg2
import psycopg2.extras
import undetected_chromedriver as uc
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------
CATALOG_URL = "https://practicum.yandex.ru/catalog/"
PROVIDER_NAME = "Яндекс Практикум"
PROVIDER_WEBSITE = "https://practicum.yandex.ru"


def _build_database_url():
    """Собирает DATABASE_URL из .env переменных или берёт готовый."""
    # Если задан DATABASE_URL напрямую — используем его
    url = os.getenv("DATABASE_URL")
    if url:
        return url

    # Иначе собираем из отдельных переменных
    db = os.getenv("POSTGRES_DB", "moneyapp")
    user = os.getenv("POSTGRES_USER", "moneyapp")
    password = os.getenv("POSTGRES_PASSWORD", "")
    host = os.getenv("POSTGRES_HOST", "localhost")
    port = os.getenv("POSTGRES_PORT", "5432")
    return "postgres://%s:%s@%s:%s/%s?sslmode=disable" % (user, password, host, port, db)


DATABASE_URL = _build_database_url()

logging.basicConfig(
    level=logging.DEBUG if "--verbose" in sys.argv else logging.INFO,
    format="%(asctime)s  %(levelname)-7s  %(message)s",
)
log = logging.getLogger("parser-yndx")

# ---------------------------------------------------------------------------
# Selenium
# ---------------------------------------------------------------------------

def build_driver(headless=True):
    options = uc.ChromeOptions()
    options.add_argument("--window-size=1600,2400")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")

    use_headless = headless and "--no-headless" not in sys.argv
    log.info("Запуск undetected Chrome (headless=%s, version_main=146)...", use_headless)
    driver = uc.Chrome(options=options, headless=use_headless, version_main=146)
    # Даём Chrome стабилизироваться перед навигацией
    time.sleep(2)
    log.info("Chrome запущен")
    return driver


# ---------------------------------------------------------------------------
# Маппинг уровней
# ---------------------------------------------------------------------------

LEVEL_MAP = {
    "start": "beginner",
    "pro": "advanced",
    "degree": "advanced",
}

# slug тегов-категорий (tag_type == "head") -> наше название категории
HEAD_TAG_SLUGS = {
    "programming": "Программирование",
    "data-analysis": "Анализ данных",
    "design": "Дизайн",
    "management": "Менеджмент",
    "marketing": "Маркетинг",
    "iskusstvennyj-intellekt": "Искусственный интеллект",
    "eng": "Английский язык",
    "prof": "Кем стать в IT",
}

CATEGORIES = list(HEAD_TAG_SLUGS.values())


# ---------------------------------------------------------------------------
# Извлечение __preloadedData__
# ---------------------------------------------------------------------------

def extract_preloaded_data(driver):
    """Извлекает window.__preloadedData__ из страницы через JS."""
    log.info("Пытаемся извлечь window.__preloadedData__ через JS...")

    data = driver.execute_script("return window.__preloadedData__;")
    if data:
        log.info("window.__preloadedData__ получен через JS execute")
        return data

    # Фоллбэк: ищем в HTML source через regex
    log.warning("JS execute вернул None, пробуем regex по page_source...")
    html = driver.page_source
    log.debug("Длина page_source: %d символов", len(html))

    pattern = r'window\.__preloadedData__\s*=\s*(\{.*?\})\s*;?\s*</script>'
    m = re.search(pattern, html, re.DOTALL)
    if m:
        raw = m.group(1)
        log.info("Найден __preloadedData__ через regex, длина JSON: %d", len(raw))
        return json.loads(raw)

    # Сохраняем HTML для отладки
    debug_path = os.path.join(os.path.dirname(__file__), "debug_page.html")
    with open(debug_path, "w", encoding="utf-8") as f:
        f.write(html)
    log.error("Не удалось найти __preloadedData__! HTML сохранён в %s", debug_path)
    return None


def parse_preloaded_data(data):
    """Парсит структуру __preloadedData__ и возвращает список курсов."""
    api_data = data.get("apiData", {})
    log.info("Ключи apiData: %s", list(api_data.keys()))

    # --- Теги (getProfessionTags) ---
    tags_raw = api_data.get("getProfessionTags", [])
    log.info("Найдено тегов (getProfessionTags): %d", len(tags_raw))

    # Строим справочник тегов по id
    tags_by_id = {}
    for t in tags_raw:
        tags_by_id[t["id"]] = {
            "name": t["name"],
            "slug": t["slug"],
            "tag_type": t.get("tag_type", ""),
        }
        log.debug("  Тег: %-30s  type=%-8s  slug=%s", t["name"], t.get("tag_type"), t["slug"])

    # --- Курсы / профессии ---
    # Ищем все ключи, которые содержат массив курсов
    profession_keys = [
        "getV2CatalogProfessions",
        "getCatalogProfessions",
        "getProfessions",
        "getCatalog",
        "getProfessionCatalog",
        "catalog",
        "professions",
    ]

    professions_raw = []
    used_key = None
    for key in profession_keys:
        val = api_data.get(key)
        if isinstance(val, list) and len(val) > 0:
            professions_raw = val
            used_key = key
            break
        elif isinstance(val, dict):
            # Может быть dict с items/results
            for sub in ("items", "results", "professions", "courses"):
                if isinstance(val.get(sub), list):
                    professions_raw = val[sub]
                    used_key = "%s.%s" % (key, sub)
                    break

    if not professions_raw:
        # Перебираем все ключи apiData для поиска массивов с курсами
        log.warning("Известные ключи не найдены, сканируем все ключи apiData...")
        for key, val in api_data.items():
            if key == "getProfessionTags" or key == "errors":
                continue
            if isinstance(val, list) and len(val) > 2:
                # Проверяем, похоже ли на курсы (есть title или name)
                sample = val[0] if val else {}
                if isinstance(sample, dict) and ("title" in sample or "name" in sample or "slug" in sample):
                    professions_raw = val
                    used_key = key
                    log.info("Найден подходящий ключ: '%s' (%d элементов)", key, len(val))
                    break
            elif isinstance(val, dict):
                for sub_key, sub_val in val.items():
                    if isinstance(sub_val, list) and len(sub_val) > 2:
                        sample = sub_val[0] if sub_val else {}
                        if isinstance(sample, dict) and ("title" in sample or "name" in sample):
                            professions_raw = sub_val
                            used_key = "%s.%s" % (key, sub_key)
                            log.info("Найден подходящий вложенный ключ: '%s' (%d элементов)", used_key, len(sub_val))
                            break

    if not professions_raw:
        log.error("Не найден массив курсов в apiData!")
        log.error("Доступные ключи и типы:")
        for k, v in api_data.items():
            if isinstance(v, list):
                log.error("  '%s': list[%d]  sample=%s", k, len(v), str(v[0])[:200] if v else "empty")
            elif isinstance(v, dict):
                log.error("  '%s': dict keys=%s", k, list(v.keys())[:10])
            else:
                log.error("  '%s': %s = %s", k, type(v).__name__, str(v)[:100])
        return []

    log.info("Используем ключ '%s': %d элементов", used_key, len(professions_raw))

    # Логируем первый элемент для понимания структуры
    if professions_raw:
        sample = professions_raw[0]
        log.info("Пример элемента (ключи): %s", list(sample.keys()) if isinstance(sample, dict) else type(sample))
        # Дампим первый элемент полностью для дебага полей
        log.info("=== ПОЛНЫЙ ДАМП ПЕРВОГО ЭЛЕМЕНТА ===")
        for k, v in sample.items():
            log.info("  field %-30s = %s", k, repr(v)[:200])
        log.info("=== КОНЕЦ ДАМПА ===")

    # --- Парсим каждый курс ---
    courses = []
    for idx, item in enumerate(professions_raw):
        if not isinstance(item, dict):
            log.debug("  [%d] пропуск — не dict: %s", idx, type(item))
            continue

        # name — основное поле названия в Практикуме
        title = item.get("name") or item.get("title") or ""
        # Убираем неразрывные пробелы
        title = title.replace("\xa0", " ")
        if not title:
            log.debug("  [%d] пропуск — нет name/title", idx)
            continue

        slug = item.get("slug", "")
        # URL строится как /profession/<slug>/ или /course/<slug>/
        item_type = item.get("type", "default")
        if slug:
            if item_type in ("degree",):
                external_url = PROVIDER_WEBSITE + "/degree/" + slug + "/"
            else:
                external_url = PROVIDER_WEBSITE + "/profession/" + slug + "/"
        else:
            external_url = ""

        # Цена — число или null
        price = None
        price_currency = "RUB"
        price_raw = item.get("price")
        currency_raw = item.get("currency")
        if isinstance(price_raw, (int, float)):
            price = float(price_raw)
        elif price_raw is None:
            # Бесплатный курс или цена не указана
            price = 0.0 if item.get("able_to_purchase") is False else None

        if currency_raw and isinstance(currency_raw, str):
            price_currency = currency_raw.upper()

        # partial_price — цена в рассрочку (для справки)
        partial_price = item.get("partial_price")

        # Длительность — приходит в месяцах, конвертируем в часы (1 мес ~ 80 ч)
        duration_hours = None
        duration_raw = item.get("duration")
        if isinstance(duration_raw, (int, float)) and duration_raw > 0:
            duration_hours = int(duration_raw) * 80

        # Ближайший старт — нет в каталоге, но может быть в landingsData
        next_start = None

        # Теги
        tag_names = []
        category = None
        level = None

        # Теги могут быть как массив id, так и массив объектов
        item_tags = item.get("tags") or item.get("tag_ids") or item.get("tagIds") or []
        for tag_ref in item_tags:
            tag_id = tag_ref if isinstance(tag_ref, str) else tag_ref.get("id", "") if isinstance(tag_ref, dict) else ""
            tag_info = tags_by_id.get(tag_id)

            if not tag_info:
                # Если tag_ref — объект с name
                if isinstance(tag_ref, dict) and tag_ref.get("name"):
                    tag_names.append(tag_ref["name"])
                    tag_slug = tag_ref.get("slug", "")
                    if tag_slug in HEAD_TAG_SLUGS:
                        category = HEAD_TAG_SLUGS[tag_slug]
                    if tag_slug in LEVEL_MAP:
                        level = LEVEL_MAP[tag_slug]
                continue

            tag_names.append(tag_info["name"])
            if tag_info["tag_type"] == "head" and tag_info["slug"] in HEAD_TAG_SLUGS:
                category = HEAD_TAG_SLUGS[tag_info["slug"]]
            if tag_info["slug"] in LEVEL_MAP:
                level = LEVEL_MAP[tag_info["slug"]]

        # Уровень может быть и отдельным полем
        if not level:
            level_raw = item.get("level") or item.get("difficulty") or ""
            if isinstance(level_raw, str):
                level_raw_lower = level_raw.lower()
                if level_raw_lower in LEVEL_MAP:
                    level = LEVEL_MAP[level_raw_lower]
                elif level_raw_lower in ("beginner", "intermediate", "advanced"):
                    level = level_raw_lower

        log.info(
            "  [%d] %-45s | level=%-10s | price=%-8s | start=%-12s | dur=%-5s | cat=%-20s | tags=%s",
            idx,
            title[:45],
            level or "-",
            price if price is not None else "-",
            str(next_start) if next_start else "-",
            duration_hours or "-",
            category or "-",
            ", ".join(tag_names[:5]) or "-",
        )

        courses.append({
            "title": title,
            "external_url": external_url,
            "category": category,
            "level": level,
            "price": price,
            "price_currency": price_currency,
            "duration_hours": duration_hours,
            "next_start_date": next_start,
            "tags": tag_names,
        })

    return courses


# ---------------------------------------------------------------------------
# Scraping
# ---------------------------------------------------------------------------

def wait_for_page_load(driver, max_wait=30):
    """Ждём пока капча пройдёт и страница загрузится."""
    log.info("Проверяем наличие капчи...")
    for attempt in range(max_wait):
        title = driver.title or ""
        log.info("  [%d/%d] title='%s'", attempt + 1, max_wait, title)

        if "робот" in title.lower() or "captcha" in title.lower():
            # Капча — SmartCaptcha Яндекса автоматически решается
            # для undetected_chromedriver, просто ждём
            log.info("  Капча обнаружена, ждём автопрохождение...")
            time.sleep(2)
            continue

        # Проверяем есть ли __preloadedData__
        has_data = driver.execute_script(
            "return typeof window.__preloadedData__ !== 'undefined'"
        )
        if has_data:
            log.info("  __preloadedData__ доступен!")
            return True

        # Страница загружена но без preloadedData — ждём гидратацию
        time.sleep(1)

    log.warning("Не дождались загрузки страницы за %d сек", max_wait)
    return False


def scrape_catalog():
    """Загружает страницу каталога и извлекает данные."""
    max_retries = 3
    driver = None

    for attempt in range(1, max_retries + 1):
        # Закрываем предыдущий драйвер если был
        if driver:
            try:
                driver.quit()
            except Exception:
                pass

        driver = build_driver()

        try:
            log.info("[попытка %d/%d] Загрузка %s ...", attempt, max_retries, CATALOG_URL)
            driver.get(CATALOG_URL)

            log.info("Ожидание загрузки body...")
            WebDriverWait(driver, 25).until(
                EC.presence_of_element_located((By.TAG_NAME, "body"))
            )
            time.sleep(2)

            # Ждём прохождения капчи и загрузки данных
            page_ok = wait_for_page_load(driver, max_wait=40)
            if not page_ok:
                log.warning("Страница не загрузилась, пробуем извлечь что есть...")

            # Попробуем извлечь preloadedData
            preloaded = extract_preloaded_data(driver)
            if not preloaded:
                log.error("preloadedData не найден — парсинг невозможен")
                log.error("Попробуйте запустить с --no-headless для ручного прохождения капчи")
                return []

            # Если дошли сюда — всё ок, выходим из retry-цикла
            break

        except Exception as e:
            log.warning("[попытка %d/%d] Ошибка: %s", attempt, max_retries, e)
            if attempt == max_retries:
                log.error("Все %d попыток исчерпаны", max_retries)
                try:
                    driver.quit()
                except Exception:
                    pass
                return []
            log.info("Повтор через 3 сек...")
            time.sleep(3)
            continue

        # Логируем верхнеуровневые ключи
        log.info("Верхние ключи __preloadedData__: %s", list(preloaded.keys()))
        api_data = preloaded.get("apiData", {})
        log.info("Ключи apiData: %s", list(api_data.keys()))
        for k, v in api_data.items():
            if isinstance(v, list):
                log.info("  apiData['%s']: list, %d элементов", k, len(v))
            elif isinstance(v, dict):
                log.info("  apiData['%s']: dict, ключи: %s", k, list(v.keys())[:8])
            else:
                log.info("  apiData['%s']: %s", k, type(v).__name__)

    # Вне retry-цикла — данные получены
    try:
        courses = parse_preloaded_data(preloaded)
        log.info("Итого спарсено курсов: %d", len(courses))
        return courses
    finally:
        if driver:
            try:
                driver.quit()
            except Exception:
                pass
            log.info("Chrome закрыт")


# ---------------------------------------------------------------------------
# Database sync
# ---------------------------------------------------------------------------

def slugify(text):
    text = text.lower().strip()
    text = re.sub(r"[^\w\s-]", "", text)
    text = re.sub(r"[\s_]+", "-", text)
    return text[:120]


def sync_to_db(courses):
    """Синхронизирует спарсенные курсы с PostgreSQL."""
    log.info("Подключение к БД: %s", DATABASE_URL.split("@")[-1])  # без пароля
    conn = psycopg2.connect(DATABASE_URL)
    conn.autocommit = False
    cur = conn.cursor()
    now = datetime.utcnow()

    try:
        # 1. Upsert provider
        provider_id = str(uuid.uuid5(uuid.NAMESPACE_URL, PROVIDER_WEBSITE))
        cur.execute(
            """
            INSERT INTO providers (id, type, name, website_url, is_active, created_at, updated_at)
            VALUES (%s, 'external', %s, %s, true, %s, %s)
            ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, updated_at = EXCLUDED.updated_at
            """,
            (provider_id, PROVIDER_NAME, PROVIDER_WEBSITE, now, now),
        )
        log.info("Provider upserted: %s (%s)", PROVIDER_NAME, provider_id)

        # 2. Upsert categories
        cat_ids = {}
        for cat_name in CATEGORIES:
            cat_code = slugify(cat_name)
            cat_id = str(uuid.uuid5(uuid.NAMESPACE_URL, "yndx-cat:" + cat_code))
            cur.execute(
                """
                INSERT INTO course_categories (id, name, code, is_active)
                VALUES (%s, %s, %s, true)
                ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
                RETURNING id
                """,
                (cat_id, cat_name, cat_code),
            )
            cat_ids[cat_name] = str(cur.fetchone()[0])
        log.info("Категории upserted: %d", len(cat_ids))

        # 3. Upsert skill tags
        tag_ids = {}

        def ensure_tag(tag_name):
            if tag_name in tag_ids:
                return tag_ids[tag_name]
            slug = slugify(tag_name)
            tag_id = str(uuid.uuid5(uuid.NAMESPACE_URL, "tag:" + slug))
            cur.execute(
                """
                INSERT INTO skill_tags (id, name, slug)
                VALUES (%s, %s, %s)
                ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
                RETURNING id
                """,
                (tag_id, tag_name, slug),
            )
            tag_ids[tag_name] = str(cur.fetchone()[0])
            return tag_ids[tag_name]

        # 4. Upsert courses
        created = 0
        updated = 0

        for c in courses:
            course_slug = slugify(c["title"])[:550]
            category_id = cat_ids.get(c["category"]) if c["category"] else None

            cur.execute(
                "SELECT id FROM courses WHERE external_url = %s",
                (c["external_url"],),
            )
            row = cur.fetchone()

            if row:
                course_id = str(row[0])
                cur.execute(
                    """
                    UPDATE courses SET
                        title = %s,
                        slug = %s,
                        category_id = %s,
                        level = %s,
                        duration_hours = %s,
                        price = %s,
                        price_currency = %s,
                        next_start_date = %s,
                        status = 'published',
                        updated_at = %s
                    WHERE id = %s
                    """,
                    (
                        c["title"], course_slug, category_id, c["level"],
                        c["duration_hours"], c["price"], c["price_currency"],
                        c["next_start_date"], now, course_id,
                    ),
                )
                updated += 1
                log.debug("  UPDATE '%s'", c["title"])
            else:
                course_id = str(uuid.uuid4())
                cur.execute(
                    """
                    INSERT INTO courses (
                        id, type, source_type, title, slug, provider_id, category_id,
                        level, duration_hours, language, is_mandatory_default, status,
                        external_url, price, price_currency, next_start_date,
                        created_at, updated_at
                    ) VALUES (
                        %s, 'external', 'imported', %s, %s, %s, %s,
                        %s, %s, 'ru', false, 'published',
                        %s, %s, %s, %s,
                        %s, %s
                    )
                    """,
                    (
                        course_id, c["title"], course_slug, provider_id, category_id,
                        c["level"], c["duration_hours"],
                        c["external_url"], c["price"], c["price_currency"],
                        c["next_start_date"], now, now,
                    ),
                )
                created += 1
                log.debug("  INSERT '%s'", c["title"])

            # 5. Sync skill tags
            cur.execute("DELETE FROM course_skill_tags WHERE course_id = %s", (course_id,))
            for tag_name in c.get("tags", []):
                tid = ensure_tag(tag_name)
                cur.execute(
                    """
                    INSERT INTO course_skill_tags (course_id, skill_tag_id)
                    VALUES (%s, %s) ON CONFLICT DO NOTHING
                    """,
                    (course_id, tid),
                )

        conn.commit()
        log.info("Синхронизация завершена: создано %d, обновлено %d", created, updated)

    except Exception:
        conn.rollback()
        log.exception("Ошибка при синхронизации с БД")
        raise
    finally:
        cur.close()
        conn.close()


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    log.info("=" * 60)
    log.info("Парсер Яндекс Практикум — старт")
    log.info("=" * 60)

    courses = scrape_catalog()

    if not courses:
        log.warning("Курсы не найдены!")
        log.warning("Проверьте debug_page.html (если создан) и логи выше")
        sys.exit(1)

    log.info("-" * 60)
    log.info("РЕЗУЛЬТАТ: %d курсов", len(courses))
    log.info("-" * 60)
    for c in courses:
        log.info(
            "  %-45s | %-10s | price=%-8s | start=%-12s | dur=%-5s | tags=%s",
            c["title"][:45],
            c["level"] or "-",
            c["price"] if c["price"] is not None else "-",
            str(c["next_start_date"]) if c["next_start_date"] else "-",
            c["duration_hours"] or "-",
            ", ".join(c["tags"][:5]),
        )

    if "--dry-run" in sys.argv:
        log.info("Dry run — синхронизация с БД пропущена")
        return

    sync_to_db(courses)
    log.info("Готово!")


if __name__ == "__main__":
    main()
