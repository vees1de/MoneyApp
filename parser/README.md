# Parser

Парсеры внешних образовательных платформ для импорта курсов в MoneyApp.

## Установка

```bash
cd parser
python3 -m venv .venv
.venv/bin/pip install -r requirements.txt
```

## Яндекс Практикум (`parser-yndx.py`)

Парсит каталог [practicum.yandex.ru/catalog](https://practicum.yandex.ru/catalog/) — 160+ курсов.

Извлекает: название, цена, длительность, уровень, категория, теги, ссылка на курс.

### Dry run (только парсинг, без записи в БД)

```bash
.venv/bin/python parser-yndx.py --dry-run --no-headless
```

### Синхронизация с БД

Подключение к PostgreSQL берётся из `../.env` (корень проекта) — переменные `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST`, `POSTGRES_DB`.
Или можно задать `DATABASE_URL` напрямую.

```bash
# Через .env (автоматически)
.venv/bin/python parser-yndx.py --no-headless

# Или через DATABASE_URL
DATABASE_URL="postgres://user:pass@localhost:5432/moneyapp?sslmode=disable" \
  .venv/bin/python parser-yndx.py --no-headless
```

### Флаги

| Флаг | Описание |
|------|----------|
| `--dry-run` | Только парсинг, без записи в БД |
| `--no-headless` | Открывает Chrome с UI (обязательно — иначе капча) |
| `--verbose` | Подробные логи (DEBUG) |
