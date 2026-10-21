// Package httpx — общий формат ошибок, ответы, привязка запроса, пагинация и middleware.
package httpx

import "net/http"

// Коды ошибок из API_SPEC.md, раздел 3.
const (
	CodeValidation        = "VALIDATION_ERROR"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeNotFound          = "NOT_FOUND"
	CodeMethodNotAllowed  = "METHOD_NOT_ALLOWED"
	CodeInvalidTransition = "INVALID_TRANSITION"
	CodeVersionConflict   = "VERSION_CONFLICT"
	CodeBusinessRule      = "BUSINESS_RULE"
	CodeUpstreamUnavail   = "UPSTREAM_UNAVAILABLE"
	CodeInternal          = "INTERNAL"
)

// FieldError — ошибка конкретного поля.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error — прикладная ошибка с HTTP-статусом; отдаётся клиенту как есть.
type Error struct {
	Status  int
	Code    string
	Message string
	Fields  []FieldError
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// ErrorBody — тело ответа с ошибкой.
type ErrorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
	TraceID string       `json:"traceId"`
}

func orDefault(msg, def string) string {
	if msg == "" {
		return def
	}
	return msg
}

func Validation(msg string, fields ...FieldError) *Error {
	return &Error{http.StatusBadRequest, CodeValidation, orDefault(msg, "Проверьте заполнение полей"), fields}
}

func Unauthorized() *Error {
	return &Error{Status: http.StatusUnauthorized, Code: CodeUnauthorized, Message: "Требуется вход в систему"}
}

func Forbidden(msg string) *Error {
	return &Error{Status: http.StatusForbidden, Code: CodeForbidden, Message: orDefault(msg, "Недостаточно прав")}
}

func NotFound(msg string) *Error {
	return &Error{Status: http.StatusNotFound, Code: CodeNotFound, Message: orDefault(msg, "Не найдено")}
}

func InvalidTransition(msg string) *Error {
	return &Error{Status: http.StatusConflict, Code: CodeInvalidTransition, Message: orDefault(msg, "Действие недопустимо в текущем статусе")}
}

func VersionConflict() *Error {
	return &Error{Status: http.StatusPreconditionFailed, Code: CodeVersionConflict, Message: "Заявку изменили параллельно, обновите страницу"}
}

func BusinessRule(msg string) *Error {
	return &Error{Status: http.StatusUnprocessableEntity, Code: CodeBusinessRule, Message: orDefault(msg, "Нарушено правило продукта")}
}

func Upstream(msg string) *Error {
	return &Error{Status: http.StatusBadGateway, Code: CodeUpstreamUnavail, Message: orDefault(msg, "Внешний сервис не ответил")}
}
