# Page: Admin Dashboard

## Route
$(System.Collections.Hashtable.Route)

## Назначение
ЛК администратора со статусом системы и ключевыми метриками.

## Зависимости
### Features
- `admin`
- `reports-analytics`

### Entities
- `user`
- `department`

## Импорты (концептуально)
```ts
import { AdminFacade } from '@app/features/admin';
import { ReportsAnalyticsFacade } from '@app/features/reports-analytics';
import type { User } from '@app/entities/user';
import type { Department } from '@app/entities/department';
```

## Что происходит на странице
- Показывает техническое состояние ключевых подсистем.
- Показывает короткие KPI и ссылки в админ-разделы.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
