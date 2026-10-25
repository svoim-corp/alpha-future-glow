// Package searchtext — нормализация поискового текста и безопасные LIKE-шаблоны без доменной логики.
package searchtext

import "strings"

// Normalize приводит регистр, ё/е и пробелы к единому виду.
func Normalize(text string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(strings.ToLower(text), "ё", "е")), " ")
}

// Patterns экранирует пользовательские метасимволы LIKE; SQL всегда использует параметры.
func Patterns(text string) (prefix, contains string) {
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(text)
	return escaped + "%", "%" + escaped + "%"
}
