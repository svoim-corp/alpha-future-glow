package dashboard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

// Формат v1 (docs/dashboard.md, «Формат v1»). Координаты — компьютерная сетка из 12 колонок.
const (
	schemaVersion = 1
	maxWidgets    = 30
	gridColumns   = 12
	maxY          = 200
	maxWidgetID   = 64
	maxActions    = 5
)

// sourceIDRe — ID договора, заявки или плана в настройках: совместим с ID API, без пользовательского текста.
var sourceIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// config — сохраняемый документ настройки главной.
type config struct {
	SchemaVersion int      `json:"schemaVersion" example:"1"`
	Widgets       []widget `json:"widgets"`
}

type widget struct {
	ID       string                     `json:"id" example:"77a45b12-67cf-4e29-848a-0db98fb07eb8"`
	Type     string                     `json:"type" example:"loan"`
	X        int                        `json:"x" example:"0"`
	Y        int                        `json:"y" example:"0"`
	W        int                        `json:"w" example:"4"`
	H        int                        `json:"h" example:"3"`
	Settings map[string]json.RawMessage `json:"settings" swaggertype:"object"`
}

// refKind — объект другой зоны, к которому привязан виджет.
type refKind int

const (
	refNone refKind = iota
	refLoan
	refApplication
	refPlan
)

// setting — правило одной настройки типа виджета.
type setting struct {
	valid func(json.RawMessage) bool
	ref   refKind
}

type widgetType struct {
	sizes    [][2]int // допустимые [w, h]
	settings map[string]setting
}

// quickActions — коды быстрых действий; пользовательских URL и произвольных команд нет.
var quickActions = []string{"apply", "plan", "credits", "calendar", "support"}

// widgetTypes — 16 типов v1. Новый тип добавляется сюда и в описание формата, schemaVersion при этом не меняется.
var widgetTypes = map[string]widgetType{
	"applications":  {[][2]int{{8, 3}, {8, 4}, {12, 3}}, map[string]setting{"filter": oneOf("all", "attention")}},
	"loan":          {[][2]int{{4, 3}, {4, 4}, {6, 3}}, map[string]setting{"loanId": source(refLoan)}},
	"payment":       {[][2]int{{4, 2}, {4, 3}, {6, 2}}, map[string]setting{"loanId": source(refLoan)}},
	"debt":          {[][2]int{{4, 2}, {4, 3}}, map[string]setting{}},
	"quick":         {[][2]int{{4, 2}, {4, 3}, {8, 2}}, map[string]setting{"actions": {valid: validActions}}},
	"calendar":      {[][2]int{{6, 4}, {8, 4}, {12, 4}}, map[string]setting{"source": oneOf("all", "payments", "plans")}},
	"plans":         {[][2]int{{6, 4}, {6, 5}, {12, 4}}, map[string]setting{"filter": oneOf("upcoming", "ideas")}},
	"application":   {[][2]int{{4, 3}, {6, 3}, {8, 3}}, map[string]setting{"applicationId": source(refApplication)}},
	"loans":         {[][2]int{{8, 3}, {8, 4}, {12, 3}}, map[string]setting{"status": oneOf("ACTIVE", "CLOSED")}},
	"payments":      {[][2]int{{8, 3}, {4, 3}, {8, 4}}, map[string]setting{"days": {valid: validDays}, "loanId": source(refLoan)}},
	"offers":        {[][2]int{{6, 4}, {4, 4}, {8, 4}}, map[string]setting{"filter": oneOf("current", "all")}},
	"drafts":        {[][2]int{{4, 3}, {6, 3}, {8, 3}}, map[string]setting{}},
	"planFunding":   {[][2]int{{6, 4}, {4, 4}, {8, 4}}, map[string]setting{"planId": source(refPlan)}},
	"notifications": {[][2]int{{4, 4}, {6, 4}, {8, 4}}, map[string]setting{"filter": oneOf("unread", "all")}},
	"tickets":       {[][2]int{{6, 3}, {4, 3}, {6, 4}}, map[string]setting{"filter": oneOf("active", "answered", "all")}},
	"funding":       {[][2]int{{4, 3}, {6, 3}, {8, 3}}, map[string]setting{}},
}

func oneOf(values ...string) setting {
	return setting{valid: func(raw json.RawMessage) bool {
		var s string
		return json.Unmarshal(raw, &s) == nil && slices.Contains(values, s)
	}}
}

// source — ID объекта или null: источник ещё не выбран (для payment — все действующие кредиты).
func source(kind refKind) setting {
	return setting{ref: kind, valid: func(raw json.RawMessage) bool {
		id, err := sourceID(raw)
		return err == nil && (id == "" || sourceIDRe.MatchString(id))
	}}
}

// sourceID — значение привязки; "" — null.
func sourceID(raw json.RawMessage) (string, error) {
	var id *string
	if err := json.Unmarshal(raw, &id); err != nil {
		return "", err
	}
	if id == nil {
		return "", nil
	}
	if *id == "" {
		return "", fmt.Errorf("empty source id")
	}
	return *id, nil
}

// validDays — платежи на 7 или 30 дней от сегодня.
func validDays(raw json.RawMessage) bool {
	var d int
	return json.Unmarshal(raw, &d) == nil && (d == 7 || d == 30)
}

// validActions — от одного до пяти разных кодов; порядок значим и сохраняется.
func validActions(raw json.RawMessage) bool {
	var actions []string
	if json.Unmarshal(raw, &actions) != nil || len(actions) < 1 || len(actions) > maxActions {
		return false
	}
	for i, a := range actions {
		if !slices.Contains(quickActions, a) || slices.Contains(actions[:i], a) {
			return false
		}
	}
	return true
}

var errSchema = &httpx.Error{Status: http.StatusBadRequest, Code: httpx.CodeDashboardSchema,
	Message: "Формат настройки главной не поддерживается. Обновите страницу"}

// configInput — тело PUT до проверки: указатели отличают отсутствующее поле от нуля.
type configInput struct {
	SchemaVersion int            `json:"schemaVersion"`
	Widgets       *[]widgetInput `json:"widgets"`
}

type widgetInput struct {
	ID       *string                    `json:"id"`
	Type     *string                    `json:"type"`
	X        *int                       `json:"x"`
	Y        *int                       `json:"y"`
	W        *int                       `json:"w"`
	H        *int                       `json:"h"`
	Settings map[string]json.RawMessage `json:"settings"`
}

// parseConfig разбирает и проверяет тело PUT. Неизвестные поля, типы и настройки — ошибка, а не тишина:
// иначе старый клиент молча потерял бы виджеты нового формата.
func parseConfig(body []byte) (*config, error) {
	var head struct {
		SchemaVersion *int `json:"schemaVersion"`
	}
	if err := json.Unmarshal(body, &head); err != nil {
		return nil, httpx.Validation("Некорректный JSON")
	}
	if head.SchemaVersion == nil {
		return nil, httpx.Validation("", required("schemaVersion"))
	}
	if *head.SchemaVersion != schemaVersion {
		return nil, errSchema
	}

	var in configInput
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil || dec.More() {
		return nil, httpx.Validation("Некорректная структура настройки главной")
	}
	if in.Widgets == nil {
		return nil, httpx.Validation("", required("widgets"))
	}
	if len(*in.Widgets) > maxWidgets {
		return nil, httpx.Validation(fmt.Sprintf("На главной может быть не больше %d виджетов", maxWidgets),
			invalid("widgets", "Слишком много виджетов"))
	}

	cfg := &config{SchemaVersion: schemaVersion, Widgets: make([]widget, 0, len(*in.Widgets))}
	var fields []httpx.FieldError
	for i, wi := range *in.Widgets {
		w, errs := checkWidget(fmt.Sprintf("widgets[%d]", i), wi)
		fields = append(fields, errs...)
		cfg.Widgets = append(cfg.Widgets, w)
	}
	if len(fields) == 0 {
		fields = checkLayout(cfg.Widgets)
	}
	if len(fields) > 0 {
		return nil, httpx.Validation("Настройка главной не прошла проверку", fields...)
	}
	return cfg, nil
}

// checkWidget проверяет одну запись по правилам её типа.
func checkWidget(path string, in widgetInput) (widget, []httpx.FieldError) {
	var fields []httpx.FieldError
	present := []struct {
		name string
		ok   bool
	}{
		{"id", in.ID != nil}, {"type", in.Type != nil}, {"x", in.X != nil}, {"y", in.Y != nil},
		{"w", in.W != nil}, {"h", in.H != nil}, {"settings", in.Settings != nil},
	}
	for _, p := range present {
		if !p.ok {
			fields = append(fields, required(path+"."+p.name))
		}
	}
	if len(fields) > 0 {
		return widget{}, fields
	}

	w := widget{ID: *in.ID, Type: *in.Type, X: *in.X, Y: *in.Y, W: *in.W, H: *in.H, Settings: in.Settings}
	if n := utf8.RuneCountInString(w.ID); n < 1 || n > maxWidgetID {
		fields = append(fields, invalid(path+".id", "Идентификатор виджета — от 1 до 64 символов"))
	}
	t, ok := widgetTypes[w.Type]
	if !ok {
		return w, append(fields, invalid(path+".type", "Неподдерживаемый тип виджета"))
	}
	if w.X < 0 || w.X >= gridColumns || w.Y < 0 || w.Y > maxY || w.X+w.W > gridColumns {
		fields = append(fields, invalid(path+".x", "Виджет выходит за пределы сетки"))
	}
	if !slices.Contains(t.sizes, [2]int{w.W, w.H}) {
		fields = append(fields, invalid(path+".w", "Недопустимый размер для этого типа виджета"))
	}
	for name := range w.Settings {
		if _, ok := t.settings[name]; !ok {
			fields = append(fields, invalid(path+".settings."+name, "Неподдерживаемая настройка"))
		}
	}
	for name, s := range t.settings {
		raw, ok := w.Settings[name]
		switch {
		case !ok:
			fields = append(fields, required(path+".settings."+name))
		case !s.valid(raw):
			fields = append(fields, invalid(path+".settings."+name, "Недопустимое значение"))
		}
	}
	// Настройки — map: порядок ошибок не должен зависеть от обхода.
	slices.SortFunc(fields, func(a, b httpx.FieldError) int { return strings.Compare(a.Field, b.Field) })
	return w, fields
}

// checkLayout — уникальные ID и непересекающиеся прямоугольники.
func checkLayout(ws []widget) []httpx.FieldError {
	var fields []httpx.FieldError
	for i, a := range ws {
		for j, b := range ws[:i] {
			if a.ID == b.ID {
				fields = append(fields, invalid(fmt.Sprintf("widgets[%d].id", i), "Идентификатор уже занят"))
			}
			if a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H {
				fields = append(fields, invalid(fmt.Sprintf("widgets[%d].x", i),
					fmt.Sprintf("Виджет пересекается с виджетом %d", j+1)))
			}
		}
	}
	return fields
}

// ref — привязка виджета к объекту другой зоны.
type ref struct {
	kind  refKind
	id    string
	field string
}

// refs — заполненные привязки виджетов. Записи неизвестного типа пропускаются: у сохранённого документа их нет.
func refs(ws []widget) []ref {
	var out []ref
	for i, w := range ws {
		for name, s := range widgetTypes[w.Type].settings {
			if s.ref == refNone {
				continue
			}
			if id, err := sourceID(w.Settings[name]); err == nil && id != "" {
				out = append(out, ref{kind: s.ref, id: id, field: fmt.Sprintf("widgets[%d].settings.%s", i, name)})
			}
		}
	}
	return out
}

func required(field string) httpx.FieldError {
	return httpx.FieldError{Field: field, Code: "REQUIRED", Message: "Обязательное поле"}
}

func invalid(field, msg string) httpx.FieldError {
	return httpx.FieldError{Field: field, Code: "INVALID", Message: msg}
}
