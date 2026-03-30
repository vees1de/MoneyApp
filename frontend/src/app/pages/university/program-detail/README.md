# Page: University Program Detail

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Детали программы: состав, формат, правила прохождения.

## Зависимости
### Features
- `corporate-university`
- `enrollments-progress`

### Entities
- `course`
- `enrollment`
- `learning-direction`

## Импорты (концептуально)
```ts
import { CorporateUniversityFacade } from '@app/features/corporate-university';
import { EnrollmentsProgressFacade } from '@app/features/enrollments-progress';
import type { Course } from '@app/entities/course';
import type { Enrollment } from '@app/entities/enrollment';
import type { LearningDirection } from '@app/entities/learning-direction';
```

## Что происходит на странице
- Показывает структуру программы и требования.
- Показывает доступные потоки/группы для участия.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
