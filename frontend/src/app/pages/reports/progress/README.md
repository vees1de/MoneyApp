# Page: Reports Progress

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Отчет по прогрессу прохождения и обязательному обучению.

## Зависимости
### Features
- `reports-analytics`
- `enrollments-progress`

### Entities
- `enrollment`
- `user`
- `department`

## Импорты (концептуально)
```ts
import { ReportsAnalyticsFacade } from '@app/features/reports-analytics';
import { EnrollmentsProgressFacade } from '@app/features/enrollments-progress';
import type { Enrollment } from '@app/entities/enrollment';
import type { User } from '@app/entities/user';
import type { Department } from '@app/entities/department';
```

## Что происходит на странице
- Показывает completion/overdue в срезах.
- Показывает проблемные подразделения.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
