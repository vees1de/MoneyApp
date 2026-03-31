# Parser

Парсеры внешних образовательных платформ для импорта курсов в MoneyApp.

Архитектура: **парсер** сохраняет JSON -> **sync** применяет к БД.

## Установка

```bash
cd parser
python3 -m venv .venv
.venv/bin/pip install -r requirements.txt
```

## Яндекс Практикум

### 1. Парсинг (parser-yndx.py)

Парсит каталог и сохраняет в `data/yndx-courses.json`.

```bash
# Стандартный запуск
.venv/bin/python parser-yndx.py --no-headless

# С подробными логами
.venv/bin/python parser-yndx.py --no-headless --verbose

# В свой файл
.venv/bin/python parser-yndx.py --no-headless -o data/my-output.json
```

### 2. Синхронизация с БД (sync-to-db.py)

Читает JSON и upsert-ит в PostgreSQL.

```bash
# Dry run — посмотреть что будет записано, без изменений в БД
.venv/bin/python sync-to-db.py --dry-run

# Применить к БД (подключение из ../.env)
.venv/bin/python sync-to-db.py

# Указать конкретный файл
.venv/bin/python sync-to-db.py data/my-output.json
```

### Флаги

| Скрипт | Флаг | Описание |
|--------|------|----------|
| parser-yndx.py | `--no-headless` | Chrome с UI (обязательно — иначе капча) |
| parser-yndx.py | `--verbose` | DEBUG-логи |
| parser-yndx.py | `-o <path>` | Путь для JSON-файла |
| sync-to-db.py | `--dry-run` | Только чтение, без записи в БД |
| sync-to-db.py | `--verbose` | DEBUG-логи |

### Формат JSON

```json
{
  "provider": "Яндекс Практикум",
  "provider_website": "https://practicum.yandex.ru",
  "parsed_at": "2026-03-31",
  "total": 163,
  "courses": [
    {
      "title": "Go-разработчик с нуля",
      "external_url": "https://practicum.yandex.ru/profession/go-developer/",
      "provider": "Яндекс Практикум",
      "category": "Программирование",
      "level": "beginner",
      "price": 127000.0,
      "price_currency": "RUB",
      "duration_hours": 720,
      "next_start_date": null,
      "tags": ["С нуля", "Язык Go", "Backend-разработка"]
    }
  ]
}
```

### Подключение к БД

Берётся из `../.env` (корень проекта):

```env
POSTGRES_HOST=193.187.92.116
POSTGRES_PORT=5432
POSTGRES_DB=moneyapp
POSTGRES_USER=moneyapp
POSTGRES_PASSWORD=...
```

Или через переменную окружения: `DATABASE_URL="postgres://..."`.
