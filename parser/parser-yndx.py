# -*- coding: utf-8 -*-
"""
Парсер каталога Яндекс Практикум.

Парсит https://practicum.yandex.ru/catalog/ и сохраняет результат
в JSON-файл (data/yndx-courses.json).

Запуск:
  .venv/bin/python parser-yndx.py --no-headless
  .venv/bin/python parser-yndx.py --no-headless --verbose
  .venv/bin/python parser-yndx.py --no-headless -o my-output.json
"""

import json
import os
import re
import sys
import time
import logging
from datetime import date
from pathlib import Path
from urllib.parse import urljoin

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

DEFAULT_OUTPUT = str(Path(__file__).resolve().parent / "data" / "yndx-courses.json")

logging.basicConfig(
    level=logging.DEBUG if "--verbose" in sys.argv else logging.INFO,
    format="%(asctime)s  %(levelname)-7s  %(message)s",
)
log = logging.getLogger("parser-yndx")

# ---------------------------------------------------------------------------
# Selenium
# ---------------------------------------------------------------------------

def build_driver():
    options = uc.ChromeOptions()
    options.add_argument("--window-size=1600,2400")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")

    use_headless = "--no-headless" not in sys.argv
    log.info("Запуск Chrome (headless=%s, version_main=146)...", use_headless)
    driver = uc.Chrome(options=options, headless=use_headless, version_main=146)
    time.sleep(2)
    log.info("Chrome запущен")
    return driver


# ---------------------------------------------------------------------------
# Маппинг
# ---------------------------------------------------------------------------

LEVEL_MAP = {
    "start": "beginner",
    "pro": "advanced",
    "degree": "advanced",
}

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


# ---------------------------------------------------------------------------
# Извлечение __preloadedData__
# ---------------------------------------------------------------------------

def extract_preloaded_data(driver):
    log.info("Извлекаем window.__preloadedData__...")

    data = driver.execute_script("return window.__preloadedData__;")
    if data:
        log.info("Получен через JS execute")
        return data

    log.warning("JS вернул None, пробуем regex...")
    html = driver.page_source

    m = re.search(r'window\.__preloadedData__\s*=\s*(\{.*?\})\s*;?\s*</script>', html, re.DOTALL)
    if m:
        return json.loads(m.group(1))

    debug_path = os.path.join(os.path.dirname(__file__), "debug_page.html")
    with open(debug_path, "w", encoding="utf-8") as f:
        f.write(html)
    log.error("__preloadedData__ не найден! HTML -> %s", debug_path)
    return None


# ---------------------------------------------------------------------------
# Парсинг данных
# ---------------------------------------------------------------------------

def parse_preloaded_data(data):
    api_data = data.get("apiData", {})
    log.info("Ключи apiData: %s", list(api_data.keys()))

    # Теги
    tags_raw = api_data.get("getProfessionTags", [])
    log.info("Тегов: %d", len(tags_raw))

    tags_by_id = {}
    for t in tags_raw:
        tags_by_id[t["id"]] = {
            "name": t["name"],
            "slug": t["slug"],
            "tag_type": t.get("tag_type", ""),
        }

    # Курсы
    profession_keys = [
        "getV2CatalogProfessions",
        "getCatalogProfessions",
        "getProfessions",
        "getCatalog",
    ]

    professions_raw = []
    used_key = None
    for key in profession_keys:
        val = api_data.get(key)
        if isinstance(val, list) and len(val) > 0:
            professions_raw = val
            used_key = key
            break

    if not professions_raw:
        for key, val in api_data.items():
            if key in ("getProfessionTags", "errors", "getProfessionTagGroups"):
                continue
            if isinstance(val, list) and len(val) > 2:
                sample = val[0] if val else {}
                if isinstance(sample, dict) and ("name" in sample or "title" in sample):
                    professions_raw = val
                    used_key = key
                    break

    if not professions_raw:
        log.error("Массив курсов не найден в apiData!")
        return []

    log.info("Ключ '%s': %d курсов", used_key, len(professions_raw))

    if professions_raw:
        sample = professions_raw[0]
        log.info("Поля первого элемента: %s", list(sample.keys()) if isinstance(sample, dict) else type(sample))

    courses = []
    for idx, item in enumerate(professions_raw):
        if not isinstance(item, dict):
            continue

        title = item.get("name") or item.get("title") or ""
        title = title.replace("\xa0", " ")
        if not title:
            continue

        slug = item.get("slug", "")
        item_type = item.get("type", "default")
        if slug:
            prefix = "degree" if item_type == "degree" else "profession"
            external_url = PROVIDER_WEBSITE + "/" + prefix + "/" + slug + "/"
        else:
            external_url = ""

        # Цена
        price = None
        price_currency = "RUB"
        price_raw = item.get("price")
        currency_raw = item.get("currency")
        if isinstance(price_raw, (int, float)):
            price = float(price_raw)
        elif price_raw is None:
            price = 0.0 if item.get("able_to_purchase") is False else None
        if currency_raw and isinstance(currency_raw, str):
            price_currency = currency_raw.upper()

        # Длительность (месяцы -> часы)
        duration_hours = None
        duration_raw = item.get("duration")
        if isinstance(duration_raw, (int, float)) and duration_raw > 0:
            duration_hours = int(duration_raw) * 80

        # Теги, категория, уровень
        tag_names = []
        category = None
        level = None

        for tag_ref in (item.get("tags") or []):
            tag_id = tag_ref if isinstance(tag_ref, str) else tag_ref.get("id", "") if isinstance(tag_ref, dict) else ""
            tag_info = tags_by_id.get(tag_id)

            if not tag_info:
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

        if not level:
            level_raw = item.get("level") or item.get("difficulty") or ""
            if isinstance(level_raw, str):
                ll = level_raw.lower()
                if ll in LEVEL_MAP:
                    level = LEVEL_MAP[ll]
                elif ll in ("beginner", "intermediate", "advanced"):
                    level = ll

        log.info(
            "  [%d] %-45s | %-10s | price=%-8s | dur=%-5s | cat=%-20s | tags=%s",
            idx, title[:45], level or "-",
            price if price is not None else "-",
            duration_hours or "-", category or "-",
            ", ".join(tag_names[:5]) or "-",
        )

        courses.append({
            "title": title,
            "external_url": external_url,
            "provider": PROVIDER_NAME,
            "category": category,
            "level": level,
            "price": price,
            "price_currency": price_currency,
            "duration_hours": duration_hours,
            "next_start_date": None,
            "tags": tag_names,
        })

    return courses


# ---------------------------------------------------------------------------
# Scraping с retry
# ---------------------------------------------------------------------------

def wait_for_page_load(driver, max_wait=30):
    log.info("Проверяем капчу...")
    for i in range(max_wait):
        title = driver.title or ""
        log.info("  [%d/%d] title='%s'", i + 1, max_wait, title)

        if "робот" in title.lower() or "captcha" in title.lower():
            log.info("  Капча, ждём...")
            time.sleep(2)
            continue

        has_data = driver.execute_script(
            "return typeof window.__preloadedData__ !== 'undefined'"
        )
        if has_data:
            log.info("  __preloadedData__ доступен!")
            return True
        time.sleep(1)

    log.warning("Не дождались загрузки за %d сек", max_wait)
    return False


def scrape_catalog():
    max_retries = 3
    driver = None

    for attempt in range(1, max_retries + 1):
        if driver:
            try:
                driver.quit()
            except Exception:
                pass

        driver = build_driver()

        try:
            log.info("[попытка %d/%d] Загрузка %s", attempt, max_retries, CATALOG_URL)
            driver.get(CATALOG_URL)

            WebDriverWait(driver, 25).until(
                EC.presence_of_element_located((By.TAG_NAME, "body"))
            )
            time.sleep(2)

            wait_for_page_load(driver, max_wait=40)

            preloaded = extract_preloaded_data(driver)
            if not preloaded:
                log.error("preloadedData не найден")
                log.error("Попробуйте --no-headless")
                return []

            break

        except Exception as e:
            log.warning("[попытка %d/%d] Ошибка: %s", attempt, max_retries, e)
            if attempt == max_retries:
                log.error("Все попытки исчерпаны")
                try:
                    driver.quit()
                except Exception:
                    pass
                return []
            log.info("Повтор через 3 сек...")
            time.sleep(3)
            continue

    try:
        courses = parse_preloaded_data(preloaded)
        log.info("Спарсено курсов: %d", len(courses))
        return courses
    finally:
        if driver:
            try:
                driver.quit()
            except Exception:
                pass
            log.info("Chrome закрыт")


# ---------------------------------------------------------------------------
# Сохранение в JSON
# ---------------------------------------------------------------------------

def save_json(courses, output_path):
    os.makedirs(os.path.dirname(output_path), exist_ok=True)

    # date -> str для сериализации
    for c in courses:
        if isinstance(c.get("next_start_date"), date):
            c["next_start_date"] = c["next_start_date"].isoformat()

    with open(output_path, "w", encoding="utf-8") as f:
        json.dump({
            "provider": PROVIDER_NAME,
            "provider_website": PROVIDER_WEBSITE,
            "parsed_at": date.today().isoformat(),
            "total": len(courses),
            "courses": courses,
        }, f, ensure_ascii=False, indent=2)

    log.info("Сохранено в %s (%d курсов)", output_path, len(courses))


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def get_output_path():
    for i, arg in enumerate(sys.argv):
        if arg == "-o" and i + 1 < len(sys.argv):
            return sys.argv[i + 1]
    return DEFAULT_OUTPUT


def main():
    log.info("=" * 60)
    log.info("Парсер Яндекс Практикум")
    log.info("=" * 60)

    courses = scrape_catalog()

    if not courses:
        log.warning("Курсы не найдены!")
        sys.exit(1)

    log.info("ИТОГО: %d курсов", len(courses))

    output_path = get_output_path()
    save_json(courses, output_path)


if __name__ == "__main__":
    main()
