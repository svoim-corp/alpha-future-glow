package shared

import "context"

// SearchScope — область доступа из проверенной серверной сессии, не из query-параметров.
type SearchScope struct {
	UserID    string
	CompanyID string
}

// SearchPosition — последняя строка стабильного порядка score DESC, id ASC.
type SearchPosition struct {
	Score int
	ID    string
}

// SearchQuery — нормализованный запрос и ограничение чтения одной зоны.
type SearchQuery struct {
	Text  string
	Limit int
	After SearchPosition
}

// SearchHit — краткая проекция источника; SQL-строки домена и финансовые модели не покидают зону.
type SearchHit struct {
	ID      string
	Title   string
	Snippet string
	State   string
	Step    string
	Score   int
}

// SearchProvider читает только собственную зону, применяя scope до формирования выдачи.
type SearchProvider func(context.Context, SearchScope, SearchQuery) ([]SearchHit, error)

// SearchLabels — названия кодов каталога для поиска по человекочитаемым целям и продуктам.
type SearchLabels struct {
	Products map[string]string
	Goals    map[string]string
}

// SearchLabelsProvider не расширяет контракт Products, используемый финансовыми сценариями.
type SearchLabelsProvider func(context.Context) (SearchLabels, error)
