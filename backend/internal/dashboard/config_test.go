package dashboard

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
)

// allTypes — по одному виджету каждого из 16 типов без пересечений.
const allTypes = `{"schemaVersion":1,"widgets":[
	{"id":"w1","type":"applications","x":0,"y":0,"w":8,"h":3,"settings":{"filter":"attention"}},
	{"id":"w2","type":"loan","x":8,"y":0,"w":4,"h":3,"settings":{"loanId":"l-55"}},
	{"id":"w3","type":"payment","x":0,"y":3,"w":4,"h":2,"settings":{"loanId":null}},
	{"id":"w4","type":"debt","x":4,"y":3,"w":4,"h":2,"settings":{}},
	{"id":"w5","type":"quick","x":8,"y":3,"w":4,"h":2,"settings":{"actions":["apply","plan","support"]}},
	{"id":"w6","type":"calendar","x":0,"y":5,"w":12,"h":4,"settings":{"source":"payments"}},
	{"id":"w7","type":"plans","x":0,"y":9,"w":6,"h":5,"settings":{"filter":"ideas"}},
	{"id":"w8","type":"application","x":6,"y":9,"w":6,"h":3,"settings":{"applicationId":null}},
	{"id":"w9","type":"loans","x":0,"y":14,"w":12,"h":3,"settings":{"status":"CLOSED"}},
	{"id":"w10","type":"payments","x":0,"y":17,"w":8,"h":3,"settings":{"days":30,"loanId":null}},
	{"id":"w11","type":"offers","x":8,"y":17,"w":4,"h":4,"settings":{"filter":"all"}},
	{"id":"w12","type":"drafts","x":0,"y":21,"w":4,"h":3,"settings":{}},
	{"id":"w13","type":"planFunding","x":4,"y":21,"w":8,"h":4,"settings":{"planId":"p-18"}},
	{"id":"w14","type":"notifications","x":0,"y":25,"w":4,"h":4,"settings":{"filter":"unread"}},
	{"id":"w15","type":"tickets","x":4,"y":25,"w":6,"h":4,"settings":{"filter":"active"}},
	{"id":"w16","type":"funding","x":0,"y":29,"w":8,"h":3,"settings":{}}
]}`

// one — документ из одного виджета.
func one(w string) string { return `{"schemaVersion":1,"widgets":[` + w + `]}` }

func TestParseConfig(t *testing.T) {
	many := make([]string, maxWidgets+1)
	for i := range many {
		many[i] = fmt.Sprintf(`{"id":"w%d","type":"debt","x":0,"y":%d,"w":4,"h":2,"settings":{}}`, i, i*2)
	}
	tests := []struct {
		name  string
		body  string
		code  string // "" — документ принят
		field string // поле первой ошибки
	}{
		{"все 16 типов", allTypes, "", ""},
		{"пустой список", `{"schemaVersion":1,"widgets":[]}`, "", ""},
		{"нет версии", `{"widgets":[]}`, httpx.CodeValidation, "schemaVersion"},
		{"другая версия", `{"schemaVersion":2,"widgets":[{"type":"future"}]}`, httpx.CodeDashboardSchema, ""},
		{"нет widgets", `{"schemaVersion":1}`, httpx.CodeValidation, "widgets"},
		{"лишнее поле документа", `{"schemaVersion":1,"widgets":[],"userId":"u-2"}`, httpx.CodeValidation, ""},
		{"не JSON", `{`, httpx.CodeValidation, ""},
		{"больше 30", `{"schemaVersion":1,"widgets":[` + strings.Join(many, ",") + `]}`, httpx.CodeValidation, "widgets"},
		{"неизвестный тип", one(`{"id":"a","type":"chart","x":0,"y":0,"w":4,"h":3,"settings":{}}`), httpx.CodeValidation, "widgets[0].type"},
		{"нет координаты", one(`{"id":"a","type":"debt","y":0,"w":4,"h":2,"settings":{}}`), httpx.CodeValidation, "widgets[0].x"},
		{"дробная координата", one(`{"id":"a","type":"debt","x":0.5,"y":0,"w":4,"h":2,"settings":{}}`), httpx.CodeValidation, ""},
		{"пустой id", one(`{"id":"","type":"debt","x":0,"y":0,"w":4,"h":2,"settings":{}}`), httpx.CodeValidation, "widgets[0].id"},
		{"id длиннее 64", one(`{"id":"` + strings.Repeat("a", 65) + `","type":"debt","x":0,"y":0,"w":4,"h":2,"settings":{}}`), httpx.CodeValidation, "widgets[0].id"},
		{"недопустимый размер", one(`{"id":"a","type":"debt","x":0,"y":0,"w":6,"h":2,"settings":{}}`), httpx.CodeValidation, "widgets[0].w"},
		{"за правым краем", one(`{"id":"a","type":"debt","x":10,"y":0,"w":4,"h":2,"settings":{}}`), httpx.CodeValidation, "widgets[0].x"},
		{"y больше 200", one(`{"id":"a","type":"debt","x":0,"y":201,"w":4,"h":2,"settings":{}}`), httpx.CodeValidation, "widgets[0].x"},
		{"отрицательный x", one(`{"id":"a","type":"debt","x":-1,"y":0,"w":4,"h":2,"settings":{}}`), httpx.CodeValidation, "widgets[0].x"},
		{"нет settings", one(`{"id":"a","type":"debt","x":0,"y":0,"w":4,"h":2}`), httpx.CodeValidation, "widgets[0].settings"},
		{"лишняя настройка", one(`{"id":"a","type":"debt","x":0,"y":0,"w":4,"h":2,"settings":{"url":"https://x"}}`), httpx.CodeValidation, "widgets[0].settings.url"},
		{"нет обязательной настройки", one(`{"id":"a","type":"loan","x":0,"y":0,"w":4,"h":3,"settings":{}}`), httpx.CodeValidation, "widgets[0].settings.loanId"},
		{"неизвестный фильтр", one(`{"id":"a","type":"tickets","x":0,"y":0,"w":6,"h":3,"settings":{"filter":"closed"}}`), httpx.CodeValidation, "widgets[0].settings.filter"},
		{"days 14", one(`{"id":"a","type":"payments","x":0,"y":0,"w":8,"h":3,"settings":{"days":14,"loanId":null}}`), httpx.CodeValidation, "widgets[0].settings.days"},
		{"ID с недопустимыми символами", one(`{"id":"a","type":"loan","x":0,"y":0,"w":4,"h":3,"settings":{"loanId":"l 55"}}`), httpx.CodeValidation, "widgets[0].settings.loanId"},
		{"пустой ID источника", one(`{"id":"a","type":"loan","x":0,"y":0,"w":4,"h":3,"settings":{"loanId":""}}`), httpx.CodeValidation, "widgets[0].settings.loanId"},
		{"ID источника длиннее 128", one(`{"id":"a","type":"planFunding","x":0,"y":0,"w":6,"h":4,"settings":{"planId":"` + strings.Repeat("p", 129) + `"}}`), httpx.CodeValidation, "widgets[0].settings.planId"},
		{"повтор быстрого действия", one(`{"id":"a","type":"quick","x":0,"y":0,"w":4,"h":2,"settings":{"actions":["apply","apply"]}}`), httpx.CodeValidation, "widgets[0].settings.actions"},
		{"нет быстрых действий", one(`{"id":"a","type":"quick","x":0,"y":0,"w":4,"h":2,"settings":{"actions":[]}}`), httpx.CodeValidation, "widgets[0].settings.actions"},
		{"неизвестное действие", one(`{"id":"a","type":"quick","x":0,"y":0,"w":4,"h":2,"settings":{"actions":["javascript:alert(1)"]}}`), httpx.CodeValidation, "widgets[0].settings.actions"},
		{"повтор id", `{"schemaVersion":1,"widgets":[
			{"id":"a","type":"debt","x":0,"y":0,"w":4,"h":2,"settings":{}},
			{"id":"a","type":"debt","x":4,"y":0,"w":4,"h":2,"settings":{}}]}`, httpx.CodeValidation, "widgets[1].id"},
		{"пересечение", `{"schemaVersion":1,"widgets":[
			{"id":"a","type":"debt","x":0,"y":0,"w":4,"h":2,"settings":{}},
			{"id":"b","type":"debt","x":2,"y":1,"w":4,"h":2,"settings":{}}]}`, httpx.CodeValidation, "widgets[1].x"},
		{"касание краями", `{"schemaVersion":1,"widgets":[
			{"id":"a","type":"debt","x":0,"y":0,"w":4,"h":2,"settings":{}},
			{"id":"b","type":"debt","x":4,"y":0,"w":4,"h":2,"settings":{}},
			{"id":"c","type":"debt","x":0,"y":2,"w":4,"h":2,"settings":{}}]}`, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseConfig([]byte(tt.body))
			if tt.code == "" {
				if err != nil {
					t.Fatalf("документ отклонён: %v", err)
				}
				if cfg.SchemaVersion != schemaVersion {
					t.Errorf("schemaVersion = %d", cfg.SchemaVersion)
				}
				return
			}
			var e *httpx.Error
			if !errors.As(err, &e) {
				t.Fatalf("ошибка = %v, ожидалась %s", err, tt.code)
			}
			if e.Code != tt.code {
				t.Errorf("код = %s, ожидался %s", e.Code, tt.code)
			}
			if tt.field != "" && (len(e.Fields) == 0 || e.Fields[0].Field != tt.field) {
				t.Errorf("поля = %+v, ожидалось %s", e.Fields, tt.field)
			}
		})
	}
}

func TestRefs(t *testing.T) {
	cfg, err := parseConfig([]byte(allTypes))
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]refKind{}
	for _, r := range refs(cfg.Widgets) {
		got[r.id+" "+r.field] = r.kind
	}
	want := map[string]refKind{
		"l-55 widgets[1].settings.loanId":  refLoan,
		"p-18 widgets[12].settings.planId": refPlan,
	}
	if len(got) != len(want) {
		t.Fatalf("привязки = %v, ожидались %v", got, want)
	}
	for k, kind := range want {
		if got[k] != kind {
			t.Errorf("привязка %s = %v, ожидалась %v", k, got[k], kind)
		}
	}
}
