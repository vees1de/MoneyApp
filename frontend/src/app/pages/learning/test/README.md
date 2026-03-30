# Page: Learning Test

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Страница прохождения теста по курсу.

## Зависимости
### Features
- `enrollments-progress`

### Entities
- `assessment`
- `enrollment`

## Импорты (концептуально)
```ts
import { EnrollmentsProgressFacade } from '@app/features/enrollments-progress';
import type { Assessment } from '@app/entities/assessment';
import type { Enrollment } from '@app/entities/enrollment';
```

## Что происходит на странице
- Запускает попытку тестирования.
- Сохраняет ответы и рассчитывает результат.
- Обновляет статус прохождения после завершения.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
