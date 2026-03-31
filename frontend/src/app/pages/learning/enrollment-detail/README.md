# Page: Enrollment Detail

## Route
`/learning/:enrollmentId`

## Что реализовано
- Загрузка enrollment через `GET /api/v1/enrollments/{id}`.
- Действия:
  - старт обучения (`POST /enrollments/{id}/start`)
  - обновление прогресса (`POST /enrollments/{id}/progress`)
  - завершение (`POST /enrollments/{id}/complete`)
- Базовый MVP-flow прохождения до завершения.
