# HR Dashboard

## Route
`/dashboard/hr`

## Layout
- Two-column desktop layout:
  - main: quick actions, requests, process
  - side: calendar/upcoming events
- Single-column on narrow screens.

## Widget set
- `quick-actions-widget`
- `course-requests-widget` (status tabs + counts)
- `current-learning-widget` as process board (`enrollments`)
- `upcoming-events-widget`

## Data sources
- `CalendarApiService`
- `ExternalRequestsApiService`
- `EnrollmentsApiService`
- `CoursesApiService`
