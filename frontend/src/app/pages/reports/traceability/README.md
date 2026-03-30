# Page: Reports Traceability

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Сквозной трекинг заявки от создания до результата.

## Зависимости
### Features
- `reports-analytics`
- `approvals`
- `external-course-requests`

### Entities
- `request`
- `approval-step`
- `certificate`

## Импорты (концептуально)
```ts
import { ReportsAnalyticsFacade } from '@app/features/reports-analytics';
import { ApprovalsFacade } from '@app/features/approvals';
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import type { Request } from '@app/entities/request';
import type { ApprovalStep } from '@app/entities/approval-step';
import type { Certificate } from '@app/entities/certificate';
```

## Что происходит на странице
- Показывает timeline всех шагов и владельцев.
- Показывает итог: пройдено/не пройдено/сертификат.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
