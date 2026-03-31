# Page: Catalog List

## Route
`/catalog`

## Назначение
Страница всех курсов с загрузкой из backend и фильтрами через query params.

## Data Source
- `GET https://bims.su/api/v1/courses`

## Текущие фильтры
- `status`
- `source_type`
- `level`
- `limit`
- `offset`

## Что происходит на странице
- При открытии загружает список курсов.
- Пользователь применяет фильтры, фронт отправляет их в query к `/courses`.
- Каждая карточка ведет на детали курса: `/catalog/:courseId`.
- Обрабатываются состояния: loading/error/empty.
