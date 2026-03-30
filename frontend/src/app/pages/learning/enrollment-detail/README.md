# Page: Enrollment Detail

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Карточка конкретного назначения: прогресс, материалы, результаты.

## Зависимости
### Features
- `enrollments-progress`
- `corporate-university`

### Entities
- `enrollment`
- `course`
- `assessment`
- `certificate`

## Импорты (концептуально)
```ts
import { EnrollmentsProgressFacade } from '@app/features/enrollments-progress';
import { CorporateUniversityFacade } from '@app/features/corporate-university';
import type { Enrollment } from '@app/entities/enrollment';
import type { Course } from '@app/entities/course';
import type { Assessment } from '@app/entities/assessment';
import type { Certificate } from '@app/entities/certificate';
```

## Что происходит на странице
- Показывает этапы прохождения и дедлайны.
- Показывает результаты тестов и доступ к материалам.
- Показывает блок сертификата после completion.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
