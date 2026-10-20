package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSON — обёртка для jsonb-колонок: поле типа JSON[T] читается и пишется как T.
type JSON[T any] struct{ V T }

func (j *JSON[T]) Scan(src any) error {
	switch s := src.(type) {
	case nil:
		var zero T
		j.V = zero
		return nil
	case []byte:
		return json.Unmarshal(s, &j.V)
	case string:
		return json.Unmarshal([]byte(s), &j.V)
	default:
		return fmt.Errorf("db.JSON: неподдерживаемый тип %T", src)
	}
}

func (j JSON[T]) Value() (driver.Value, error) {
	b, err := json.Marshal(j.V)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}
