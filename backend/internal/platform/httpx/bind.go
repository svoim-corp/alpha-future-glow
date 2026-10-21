package httpx

import (
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
	err := c.ShouldBindJSON(req)
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
