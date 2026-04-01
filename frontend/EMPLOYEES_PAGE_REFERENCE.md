# Employees Page Reference (Frontend)

## Короткий ответ: есть ли такие возможности

Частично есть.

Что уже можно получить:
- имя сотрудника;
- отдел/специализация (через `employee_profile`);
- роль/permissions (если нужно для UI).

Что сейчас **не отдается готовым API**:
- per-employee счетчики:
  - сколько курсов проходит сейчас;
  - сколько выполнено;
  - сколько просрочено.

## 1) Что брать из текущего backend

## Endpoint: список сотрудников

- `GET /api/v1/admin/users`
- Permission: `users.read`
- Response: `{"items": identity.UserView[]}`

Полезные поля для страницы:
- `item.id`
- `item.employee_profile.first_name`
- `item.employee_profile.last_name`
- `item.employee_profile.position_title` (можно использовать как специализацию)
- `item.employee_profile.department_id`
- `item.status`

Примечание:
- отдельного endpoint каталога департаментов (`/departments`) в публичном API сейчас нет.

## 2) Что по метрикам обучения на сотрудника

Требования страницы:
- `in_progress_count`
- `completed_count`
- `overdue_count`

Текущие API это не закрывают:
- `GET /api/v1/enrollments/my` работает только для текущего пользователя;
- `GET /api/v1/analytics/compliance` дает агрегаты по отделам, не по сотрудникам;
- `GET /api/v1/dashboard/manager` содержит preview команды и общие KPI, но не персональные счетчики курсов для каждого сотрудника.

Итог: для полноценной таблицы сотрудников нужен новый backend endpoint.

## 3) Рекомендованный endpoint (минимальный)

`GET /api/v1/employees/learning-stats?scope=team|all&department_id=<uuid>&search=<text>&limit=...&offset=...`

Permission:
- team scope: `external_requests.approve_manager` или отдельный `employees.read_team`
- all scope: `users.read`/`analytics.read_hr` или отдельный `employees.read_all`

Рекомендуемый response:

```json
{
  "items": [
    {
      "user_id": "uuid",
      "first_name": "Anna",
      "last_name": "Ivanova",
      "position_title": "Backend Developer",
      "department_id": "uuid",
      "in_progress_count": 3,
      "completed_count": 12,
      "overdue_count": 1
    }
  ],
  "total": 128
}
```

## 4) Frontend view-model (готовый)

```ts
export interface EmployeesPageRow {
  user_id: string;
  full_name: string;
  position_title: string | null; // специализация/должность
  department_id: string | null;
  in_progress_count: number;
  completed_count: number;
  overdue_count: number;
}
```

## 5) Временный workaround (если backend не менять сейчас)

Можно сделать страницу только с профилями сотрудников:
- источник: `GET /api/v1/admin/users`
- без 3 метрик обучения, либо с заглушками.

Корректно собрать требуемые три счетчика через текущие endpoint’ы без backend-доработки нельзя.

