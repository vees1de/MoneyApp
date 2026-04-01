package smart_export

import "strings"

// statusTranslations maps English status codes to Russian labels.
var statusTranslations = map[string]string{
	// Application / enrollment / request statuses
	"draft":                "Черновик",
	"pending":              "На рассмотрении",
	"pending_manager":      "Ожидает руководителя",
	"pending_hr":           "Ожидает HR",
	"approved":             "Одобрено",
	"rejected":             "Отклонено",
	"withdrawn":            "Отозвано",
	"enrolled":             "Записан",
	"in_progress":          "В процессе",
	"completed":            "Завершено",
	"canceled":             "Отменено",
	"cancelled":            "Отменено",
	"expired":              "Истекло",
	"on_hold":              "Приостановлено",
	"failed":               "Не сдано",
	"sent_to_revision":     "На доработке",
	"waiting_certificate":  "Ожидание сертификата",
	"certificate_uploaded": "Сертификат загружен",
	"certificate_approved": "Сертификат одобрен",
	"certificate_rejected": "Сертификат отклонён",
	"not_started":          "Не начат",

	// Course statuses
	"published": "Опубликован",
	"archived":  "В архиве",

	// Intake statuses
	"open":   "Открыт",
	"closed": "Закрыт",

	// Suggestion statuses
	"new":      "Новое",
	"reviewed": "Рассмотрено",

	// Payment statuses
	"unpaid": "Не оплачено",
	"paid":   "Оплачено",

	// Employment statuses
	"active":     "Активен",
	"inactive":   "Неактивен",
	"terminated": "Уволен",
	"on_leave":   "В отпуске",
	"probation":  "Испытательный срок",
}

// levelTranslations maps English level codes to Russian labels.
var levelTranslations = map[string]string{
	"beginner":     "Начальный",
	"junior":       "Junior",
	"intern":       "Стажёр",
	"starter":      "Начинающий",
	"middle":       "Middle",
	"intermediate": "Средний",
	"senior":       "Senior",
	"advanced":     "Продвинутый",
	"expert":       "Эксперт",
	"lead":         "Lead",
	"principal":    "Principal",
	"staff":        "Staff",
}

// sourceTypeTranslations maps source_type / source codes.
var sourceTypeTranslations = map[string]string{
	"internal":   "Внутренний",
	"external":   "Внешний",
	"imported":   "Импортирован",
	"catalog":    "Каталог",
	"manual":     "Ручная запись",
	"assignment": "Назначение",
	"self":       "Самостоятельно",
	"intake":     "Набор",
	"suggestion": "Предложение",
}

// courseTypeTranslations maps course type codes.
var courseTypeTranslations = map[string]string{
	"online":   "Онлайн",
	"offline":  "Очный",
	"blended":  "Смешанный",
	"video":    "Видеокурс",
	"webinar":  "Вебинар",
	"workshop": "Воркшоп",
	"self_paced": "Самостоятельный",
}

// languageTranslations maps language codes.
var languageTranslations = map[string]string{
	"ru":  "Русский",
	"en":  "Английский",
	"de":  "Немецкий",
	"fr":  "Французский",
	"es":  "Испанский",
	"zh":  "Китайский",
	"ja":  "Японский",
	"ko":  "Корейский",
	"pt":  "Португальский",
	"it":  "Итальянский",
}

// priorityTranslations maps priority codes.
var priorityTranslations = map[string]string{
	"low":      "Низкий",
	"medium":   "Средний",
	"high":     "Высокий",
	"critical": "Критичный",
	"urgent":   "Срочный",
}

// currencyTranslations maps currency codes.
var currencyTranslations = map[string]string{
	"RUB": "₽ (Рубль)",
	"USD": "$ (Доллар)",
	"EUR": "€ (Евро)",
	"GBP": "£ (Фунт)",
	"KZT": "₸ (Тенге)",
}

// columnKeyTranslationMap maps column keys to their translation dictionaries.
var columnKeyTranslationMap = map[string]map[string]string{
	"status":            statusTranslations,
	"employment_status": statusTranslations,
	"payment_status":    statusTranslations,

	"level":        levelTranslations,
	"course_level": levelTranslations,

	"source":      sourceTypeTranslations,
	"source_type": sourceTypeTranslations,
	"type":        courseTypeTranslations,

	"language": languageTranslations,
	"priority": priorityTranslations,

	"price_currency": currencyTranslations,
}

// LocalizeRows translates known enum values in query results to Russian.
func LocalizeRows(qr *QueryResult) {
	if qr == nil || len(qr.Rows) == 0 {
		return
	}

	// Build a per-column index of which translation map to use.
	translators := make([]map[string]string, len(qr.Columns))
	for i, col := range qr.Columns {
		translators[i] = columnKeyTranslationMap[col.Key]
	}

	for _, row := range qr.Rows {
		for colIdx, val := range row {
			dict := translators[colIdx]
			if dict == nil {
				continue
			}

			strVal := ""
			switch v := val.(type) {
			case string:
				strVal = v
			case []byte:
				strVal = string(v)
			default:
				continue
			}

			key := strings.TrimSpace(strings.ToLower(strVal))
			if translated, ok := dict[key]; ok {
				row[colIdx] = translated
			}
		}
	}
}
