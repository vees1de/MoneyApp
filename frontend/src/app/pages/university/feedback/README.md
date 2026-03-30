# Page: University Feedback

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Сбор и просмотр обратной связи по внутреннему курсу.

## Зависимости
### Features
- `corporate-university`
- `reports-analytics`

### Entities
- `course`
- `user`

## Импорты (концептуально)
```ts
import { CorporateUniversityFacade } from '@app/features/corporate-university';
import { ReportsAnalyticsFacade } from '@app/features/reports-analytics';
import type { Course } from '@app/entities/course';
import type { User } from '@app/entities/user';
```

## Что происходит на странице
- Собирает оценку и комментарий участника.
- Показывает агрегированные результаты по курсу.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
