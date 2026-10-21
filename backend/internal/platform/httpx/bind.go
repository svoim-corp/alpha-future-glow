package httpx

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
)

func init() {
	// В ошибках валидации используем json-имена полей.
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
			if name == "-" || name == "" {
				return f.Name
			}
			return name
		})
	}
}

// Bind разбирает JSON-тело и валидирует его по тегам binding.
// Возвращает *Error, пригодный для Fail.
func Bind(c *gin.Context, req any) error {
	return bindError(c.ShouldBindJSON(req))
}

// BindStrict — как Bind, но поле, которого нет в структуре запроса, — ошибка, а не тишина. Для запросов, где
// клиент не должен пытаться менять лишнее: так он узнаёт, что поле не меняется, вместо молчаливого игнора.
func BindStrict(c *gin.Context, req any) error {
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(req); err != nil {
		if name, ok := strings.CutPrefix(err.Error(), "json: unknown field "); ok {
			return Validation("Эти данные нельзя изменить в заявке",
				FieldError{Field: strings.Trim(name, `"`), Code: "INVALID", Message: "Поле не меняется через заявку"})
		}
		return Validation("Некорректный JSON")
	}
	return bindError(binding.Validator.ValidateStruct(req))
}

// bindError превращает ошибку разбора или валидации в *Error, пригодный для Fail.
func bindError(err error) error {
	if err == nil {
		return nil
	}
	var ves validator.ValidationErrors
	if errors.As(err, &ves) {
		fields := make([]FieldError, 0, len(ves))
		for _, fe := range ves {
			fields = append(fields, fieldError(fe))
		}
		return Validation("Проверьте заполнение полей", fields...)
	}
	return Validation("Некорректный JSON")
}

func fieldError(fe validator.FieldError) FieldError {
	// Namespace вида "Req.company.inn" — отбрасываем имя корневой структуры.
	name := fe.Namespace()
	if i := strings.Index(name, "."); i >= 0 {
		name = name[i+1:]
	}
	switch fe.Tag() {
	case "required", "required_if":
		return FieldError{name, "REQUIRED", "Обязательное поле"}
	case "min", "max", "gte", "lte", "gt", "lt":
		return FieldError{name, "INVALID", "Значение вне допустимого диапазона"}
	case "oneof":
		return FieldError{name, "INVALID", "Недопустимое значение"}
	default:
		return FieldError{name, "INVALID", "Некорректное значение"}
	}
}
