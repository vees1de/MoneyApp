# -*- coding: utf-8 -*-
"""
Синхронизация спарсенных курсов из JSON в PostgreSQL.

Читает JSON-файл (от parser-yndx.py) и upsert-ит в БД MoneyApp:
  - providers, course_categories, skill_tags, courses, course_skill_tags

Запуск:
  .venv/bin/python sync-to-db.py                          # по умолчанию data/yndx-courses.json
  .venv/bin/python sync-to-db.py data/yndx-courses.json   # явный путь
  .venv/bin/python sync-to-db.py --dry-run                # только проверка, без записи
"""

import json
import os
import re
import sys
import uuid
import logging
from datetime import datetime, timezone
from pathlib import Path

from dotenv import load_dotenv

_project_root = Path(__file__).resolve().parent.parent
load_dotenv(_project_root / ".env")

import psycopg2

# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------

DEFAULT_INPUT = str(Path(__file__).resolve().parent / "data" / "yndx-courses.json")

CATEGORIES = [
    "Программирование",
    "Анализ данных",
    "Дизайн",
    "Менеджмент",
    "Маркетинг",
    "Искусственный интеллект",
    "Английский язык",
    "Кем стать в IT",
]

logging.basicConfig(
    level=logging.DEBUG if "--verbose" in sys.argv else logging.INFO,
    format="%(asctime)s  %(levelname)-7s  %(message)s",
)
log = logging.getLogger("sync-to-db")


# ---------------------------------------------------------------------------
# Database URL
# ---------------------------------------------------------------------------

def _build_database_url():
    url = os.getenv("DATABASE_URL")
    if url:
        return url

    db = os.getenv("POSTGRES_DB", "moneyapp")
    user = os.getenv("POSTGRES_USER", "moneyapp")
    password = os.getenv("POSTGRES_PASSWORD", "")
    host = os.getenv("POSTGRES_HOST", "localhost")
    port = os.getenv("POSTGRES_PORT", "5432")

    if ":" in host:
        host, port = host.rsplit(":", 1)

    return "postgres://%s:%s@%s:%s/%s?sslmode=disable" % (user, password, host, port, db)


DATABASE_URL = _build_database_url()


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def slugify(text):
    text = text.lower().strip()
    text = re.sub(r"[^\w\s-]", "", text)
    text = re.sub(r"[\s_]+", "-", text)
    return text[:120]


# ---------------------------------------------------------------------------
# Sync
# ---------------------------------------------------------------------------

def sync_to_db(data, dry_run=False):
    provider_name = data["provider"]
    provider_website = data["provider_website"]
    courses = data["courses"]

    log.info("Провайдер: %s", provider_name)
    log.info("Курсов в файле: %d", len(courses))

    if dry_run:
        log.info("--- DRY RUN ---")
        for c in courses:
            log.info(
                "  %-45s | %-10s | price=%-8s | dur=%-5s | tags=%s",
                c["title"][:45],
                c.get("level") or "-",
                c.get("price") if c.get("price") is not None else "-",
                c.get("duration_hours") or "-",
                ", ".join((c.get("tags") or [])[:4]),
            )
        log.info("Dry run завершён — БД не затронута")
        return

    log.info("Подключение к БД: %s", DATABASE_URL.split("@")[-1])
    conn = psycopg2.connect(DATABASE_URL)

    try:
        provider_id = str(uuid.uuid5(uuid.NAMESPACE_URL, provider_website))
        cat_ids = {}
        tag_ids_by_name = {}
        tag_ids_by_slug = {}

        with conn.cursor() as cur:
            now = datetime.now(tz=timezone.utc)

            # 1. Provider
            cur.execute(
                """
                INSERT INTO providers (id, type, name, website_url, is_active, created_at, updated_at)
                VALUES (%s, 'external', %s, %s, true, %s, %s)
                ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, updated_at = EXCLUDED.updated_at
                """,
                (provider_id, provider_name, provider_website, now, now),
            )
            log.info("Provider upserted: %s", provider_name)

            # 2. Categories
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

            # 3. Skill tags (кешируем целиком, чтобы не дергать БД на каждый тег)
            cur.execute("SELECT id, name, slug FROM skill_tags")
            for tag_id, tag_name, tag_slug in cur.fetchall():
                tag_ids_by_name[tag_name] = str(tag_id)
                tag_ids_by_slug[tag_slug] = str(tag_id)

        conn.commit()
        log.info("Категории: %d", len(cat_ids))
        log.info("Тегов в кеше: %d", len(tag_ids_by_slug))

        # 4. Courses
        created = 0
        updated = 0
        errors = 0

        for idx, c in enumerate(courses):
            title = c["title"]
            external_url = c.get("external_url", "")
            if not external_url:
                continue

            try:
                course_now = datetime.now(tz=timezone.utc)
                course_action = None
                pending_tag_ids_by_name = {}
                pending_tag_ids_by_slug = {}

                with conn.cursor() as cur:
                    def ensure_tag(tag_name):
                        tag_name = (tag_name or "").strip()
                        if not tag_name:
                            return None

                        if tag_name in pending_tag_ids_by_name:
                            return pending_tag_ids_by_name[tag_name]

                        if tag_name in tag_ids_by_name:
                            return tag_ids_by_name[tag_name]

                        slug = slugify(tag_name)
                        if slug in pending_tag_ids_by_slug:
                            existing_id = pending_tag_ids_by_slug[slug]
                            pending_tag_ids_by_name[tag_name] = existing_id
                            return existing_id

                        if slug in tag_ids_by_slug:
                            existing_id = tag_ids_by_slug[slug]
                            tag_ids_by_name[tag_name] = existing_id
                            log.debug("Tag alias by slug: '%s' -> %s", tag_name, slug)
                            return existing_id

                        tag_id = str(uuid.uuid5(uuid.NAMESPACE_URL, "tag:" + slug))
                        cur.execute(
                            """
                            INSERT INTO skill_tags (id, name, slug)
                            VALUES (%s, %s, %s)
                            ON CONFLICT DO NOTHING
                            RETURNING id
                            """,
                            (tag_id, tag_name, slug),
                        )
                        row = cur.fetchone()
                        if row:
                            existing_id = str(row[0])
                            pending_tag_ids_by_name[tag_name] = existing_id
                            pending_tag_ids_by_slug[slug] = existing_id
                        else:
                            cur.execute(
                                """
                                SELECT id, name, slug
                                FROM skill_tags
                                WHERE name = %s OR slug = %s
                                ORDER BY CASE WHEN name = %s THEN 0 ELSE 1 END
                                LIMIT 1
                                """,
                                (tag_name, slug, tag_name),
                            )
                            row = cur.fetchone()
                            if not row:
                                raise RuntimeError("tag upsert failed for '%s'" % tag_name)

                            existing_id = str(row[0])
                            if row[1] != tag_name or row[2] != slug:
                                log.debug(
                                    "Using existing tag '%s' (%s) for incoming '%s' (%s)",
                                    row[1], row[2], tag_name, slug,
                                )

                            tag_ids_by_name[tag_name] = existing_id
                            tag_ids_by_slug[slug] = existing_id

                        return existing_id

                    course_slug = slugify(title)[:550]
                    category_id = cat_ids.get(c.get("category")) if c.get("category") else None
                    level = c.get("level")
                    duration_hours = c.get("duration_hours")
                    price = c.get("price")
                    price_currency = c.get("price_currency", "RUB")
                    next_start = c.get("next_start_date")

                    cur.execute(
                        "SELECT id FROM courses WHERE external_url = %s ORDER BY created_at NULLS LAST, id LIMIT 1",
                        (external_url,),
                    )
                    row = cur.fetchone()

                    if row:
                        course_id = str(row[0])
                        cur.execute(
                            """
                            UPDATE courses SET
                                title = %s, slug = %s, category_id = %s, level = %s,
                                duration_hours = %s, price = %s, price_currency = %s,
                                next_start_date = %s, status = 'published', updated_at = %s
                            WHERE id = %s
                            """,
                            (title, course_slug, category_id, level,
                             duration_hours, price, price_currency,
                             next_start, course_now, course_id),
                        )
                        course_action = "updated"
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
                            (course_id, title, course_slug, provider_id, category_id,
                             level, duration_hours,
                             external_url, price, price_currency,
                             next_start, course_now, course_now),
                        )
                        course_action = "created"

                    # Tags
                    cur.execute("DELETE FROM course_skill_tags WHERE course_id = %s", (course_id,))

                    seen_tag_ids = set()
                    for tag_name in (c.get("tags") or []):
                        tid = ensure_tag(tag_name)
                        if not tid or tid in seen_tag_ids:
                            continue
                        seen_tag_ids.add(tid)
                        cur.execute(
                            """
                            INSERT INTO course_skill_tags (course_id, skill_tag_id)
                            VALUES (%s, %s)
                            ON CONFLICT DO NOTHING
                            """,
                            (course_id, tid),
                        )

                conn.commit()
                tag_ids_by_name.update(pending_tag_ids_by_name)
                tag_ids_by_slug.update(pending_tag_ids_by_slug)
                if course_action == "created":
                    created += 1
                elif course_action == "updated":
                    updated += 1

                if (idx + 1) % 20 == 0:
                    log.info("  Прогресс: %d/%d", idx + 1, len(courses))

            except Exception as e:
                conn.rollback()
                errors += 1
                log.warning("  Ошибка курса '%s': %s", title[:40], e)

        log.info("Готово: создано %d, обновлено %d, ошибок %d", created, updated, errors)

    finally:
        conn.close()


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def get_input_path():
    for arg in sys.argv[1:]:
        if not arg.startswith("-") and arg.endswith(".json"):
            return arg
    return DEFAULT_INPUT


def main():
    input_path = get_input_path()
    dry_run = "--dry-run" in sys.argv

    log.info("=" * 60)
    log.info("Синхронизация курсов -> БД")
    log.info("Файл: %s", input_path)
    log.info("=" * 60)

    if not os.path.exists(input_path):
        log.error("Файл не найден: %s", input_path)
        log.error("Сначала запустите парсер: .venv/bin/python parser-yndx.py --no-headless")
        sys.exit(1)

    with open(input_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    log.info("Загружено: %d курсов (спарсено %s)", data["total"], data.get("parsed_at", "?"))

    sync_to_db(data, dry_run=dry_run)


if __name__ == "__main__":
    main()
