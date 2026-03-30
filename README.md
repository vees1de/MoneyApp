# ИМПУЛЬС

> Единая платформа обучения сотрудников

ИМПУЛЬС - цифровая платформа для управления обучением сотрудников ИТ-компании. Она объединяет внутренние образовательные программы и внешние профессиональные курсы в единую систему, автоматизирует согласование, планирование, контроль и аналитику обучения, снижает нагрузку на HR и делает развитие сотрудников прозрачным для бизнеса.

## Что умеет платформа

- управлять каталогом внутренних и внешних курсов;
- назначать обучение, отслеживать прогресс и результаты тестирования;
- вести заявки на внешнее обучение и маршруты согласования;
- контролировать бюджетные лимиты и подтверждение сертификатов;
- поддерживать корпоративный университет: направления, программы, группы, сессии и обратную связь;
- показывать ролевые дашборды, отчеты и экспорт данных;
- управлять пользователями, ролями, правами и журналом аудита;
- подключать внешние сервисы: Outlook, GitHub, YouGile, Yandex OAuth, Telegram.

## Для кого

| Роль | Основные сценарии |
| --- | --- |
| Сотрудник | каталог курсов, мое обучение, заявки на внешние курсы, сертификаты, профиль |
| Руководитель | согласование заявок, контроль прогресса команды, менеджерский дашборд |
| HR / L&D | программы обучения, бюджет, аналитика, уведомления, корпоративный университет |
| Тренер | работа с группами, сессиями, обратной связью и учебными потоками |
| Администратор | пользователи, роли, права доступа, системные настройки и интеграции |

## Архитектура

ИМПУЛЬС собран как монорепозиторий из backend, frontend и инфраструктурных сценариев.

### Backend

- Go 1.26
- `net/http` + `chi`
- PostgreSQL
- JWT-аутентификация + refresh sessions
- RBAC и feature-based access control
- modular monolith
- DB-backed outbox и фоновые очереди
- OpenAPI / Swagger

Основные backend-модули:

- `identity`
- `org`
- `admin`
- `catalog`
- `learning`
- `testing`
- `certificates`
- `course_requests`
- `external_training`
- `notifications`
- `outlook`
- `university`
- `analytics`
- `audit`
- `github_integration`
- `yougile`

### Frontend

- Angular 21
- Angular Material / CDK
- lazy-loaded страницы
- role guards и feature guards
- отдельные рабочие столы для `EMPLOYEE`, `MANAGER`, `HR_LND`, `TRAINER`, `ADMINISTRATOR`

### Инфраструктура

- Docker Compose для локальной инфраструктуры и миграций
- фоновой worker для интеграций и асинхронных задач
- SSH-скрипты для серверных операций
- GitHub Actions workflow для деплоя

## Ключевые пользовательские разделы

- `Каталог`: внутренние и внешние курсы, карточка курса, материалы.
- `Мое обучение`: назначения, прохождение, фиксация прогресса.
- `Внешние заявки`: создание, отправка, ревизии, согласование, загрузка сертификатов.
- `Согласования`: inbox для руководителей и HR.
- `Корпоративный университет`: направления, программы, группы, сессии, обратная связь.
- `Календарь`: обзор обучения и конфликтов расписания.
- `Отчеты`: сводка, прогресс, бюджет, трассируемость, экспорт.
- `Уведомления`: центр уведомлений и массовое прочтение.
- `Администрирование`: пользователи, роли, права доступа.

## Структура репозитория

```text
.
├── backend/               # Go API, worker, migrations, OpenAPI, domain modules
│   ├── cmd/api            # HTTP API
│   ├── cmd/worker         # Фоновые задачи и интеграции
│   ├── internal/docs      # Swagger / OpenAPI
│   ├── local_test         # Простая HTML-панель для smoke-тестов API
│   └── migrations         # Схема БД и seed-данные
├── frontend/              # Angular-приложение
├── deploy/                # Скрипты серверных операций
├── docker-compose.yml     # Compose-стек для локальной инфраструктуры
└── .env.example           # Базовая конфигурация окружения
```

## Быстрый старт

### Требования

- Go 1.26+
- Node.js + npm
- Docker и Docker Compose

### 1. Подготовить окружение

```bash
cp .env.example .env
```

Минимально проверьте значения:

- `POSTGRES_PASSWORD`
- `AUTH_JWT_SECRET`
- `POSTGRES_DB`
- `POSTGRES_USER`

### 2. Поднять PostgreSQL

```bash
docker compose up -d postgres
```

### 3. Применить миграции

```bash
docker compose run --rm migrate
```

### 4. Запустить backend API

```bash
cd backend
go run ./cmd/api
```

API по умолчанию будет доступен на [http://localhost:8080](http://localhost:8080).

### 5. Запустить worker

В отдельном терминале:

```bash
cd backend
go run ./cmd/worker
```

### 6. Запустить frontend

```bash
cd frontend
npm install
npm run start
```

Frontend по умолчанию доступен на [http://localhost:4200](http://localhost:4200).

## API и диагностика

- Base path: `/api/v1`
- Swagger UI: [http://localhost:8080/swagger](http://localhost:8080/swagger)
- OpenAPI: [http://localhost:8080/openapi.yaml](http://localhost:8080/openapi.yaml)
- Swagger JSON: [http://localhost:8080/swagger.json](http://localhost:8080/swagger.json)
- Liveness probe: [http://localhost:8080/healthz](http://localhost:8080/healthz)
- Readiness probe: [http://localhost:8080/readyz](http://localhost:8080/readyz)

Полезные материалы:

- [backend/README.md](./backend/README.md)
- [backend/FRONTEND_API_GUIDE.md](./backend/FRONTEND_API_GUIDE.md)
- [deploy/README.md](./deploy/README.md)

## Seed-данные для локальной среды

Миграции добавляют в базу:

- тестовые курсы и skill tags;
- demo-пользователей для ролей `employee`, `manager`, `hr`;
- demo-структуру подразделения и связку сотрудник -> руководитель.

Для ручного smoke-тестирования можно использовать [backend/local_test](./backend/local_test/).

<details>
<summary>Демо-учетки</summary>

| Роль | Email | Password |
| --- | --- | --- |
| Employee | `employee.demo@moneyapp.local` | `DemoEmployee123!` |
| Manager | `manager.demo@moneyapp.local` | `DemoManager123!` |
| HR | `hr.demo@moneyapp.local` | `DemoHR123!` |

</details>

## Деплой

В репозитории уже есть заготовки для серверных сценариев:

- [deploy/apply_backend_migrations_ssh.sh](./deploy/apply_backend_migrations_ssh.sh)
- [deploy/build_frontend_local_to_server.sh](./deploy/build_frontend_local_to_server.sh)
- [deploy/build_frontend_on_server.sh](./deploy/build_frontend_on_server.sh)
- [ci-cd.yml](./.github/workflows/ci-cd.yml)

Важно: часть deploy-артефактов в текущем рабочем дереве требует синхронизации с фактической структурой репозитория. Для локальной разработки ориентируйтесь на команды из раздела "Быстрый старт".

## Текущее состояние

Проект находится в активной переработке из legacy-наследия в единую LMS/L&D платформу. На момент обновления этого README:

- backend проходит `go test ./...`;
- frontend собирается через `npm run build`;
- корневой README отражает актуальную продуктовую модель ИМПУЛЬС и текущую структуру кода.
