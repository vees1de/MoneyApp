# Page: Reports Export

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Экспорт отчетов в Excel/PDF.

## Зависимости
### Features
- `reports-analytics`

### Entities
- `enrollment`
- `request`
- `budget-limit`

## Импорты (концептуально)
```ts
import { ReportsAnalyticsFacade } from '@app/features/reports-analytics';
import type { Enrollment } from '@app/entities/enrollment';
import type { Request } from '@app/entities/request';
import type { BudgetLimit } from '@app/entities/budget-limit';
```

## Что происходит на странице
- Позволяет выбрать шаблон, период и формат.
- Формирует задачу экспорта и выдает файл пользователю.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
