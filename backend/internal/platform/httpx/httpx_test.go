package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func do(r http.Handler, method, path, body string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestFail(t *testing.T) {
	r := gin.New()
	r.Use(TraceID())
	r.GET("/known", func(c *gin.Context) { Fail(c, NotFound("нет")) })
	r.GET("/unknown", func(c *gin.Context) { Fail(c, errors.New("boom")) })

	var b ErrorBody
	w := do(r, "GET", "/known", "", nil)
	_ = json.Unmarshal(w.Body.Bytes(), &b)
	if w.Code != 404 || b.Code != CodeNotFound || b.Message != "нет" || b.TraceID == "" {
		t.Fatalf("known: %d %+v", w.Code, b)
	}
	if strings.Contains(w.Body.String(), "fields") {
		t.Fatal("пустые fields должны быть опущены")
	}

	w = do(r, "GET", "/unknown", "", map[string]string{"X-Request-Id": "abc"})
	b = ErrorBody{}
	_ = json.Unmarshal(w.Body.Bytes(), &b)
	if w.Code != 500 || b.Code != CodeInternal || b.TraceID != "abc" || strings.Contains(w.Body.String(), "boom") {
		t.Fatalf("unknown: %d %+v", w.Code, b)
	}
}

func TestBind(t *testing.T) {
	type req struct {
		Login string `json:"login" binding:"required"`
		Age   int    `json:"age" binding:"gte=18"`
	}
	r := gin.New()
	r.Use(TraceID())
	r.POST("/x", func(c *gin.Context) {
		var q req
		if err := Bind(c, &q); err != nil {
			Fail(c, err)
			return
		}
		c.Status(204)
	})
	var b ErrorBody
	w := do(r, "POST", "/x", `{"age":10}`, nil)
	_ = json.Unmarshal(w.Body.Bytes(), &b)
	if w.Code != 400 || len(b.Fields) != 2 || b.Fields[0].Field != "login" || b.Fields[0].Code != "REQUIRED" || b.Fields[1].Field != "age" {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	w = do(r, "POST", "/x", `{bad`, nil)
	b = ErrorBody{}
	_ = json.Unmarshal(w.Body.Bytes(), &b)
	if w.Code != 400 || b.Message != "Некорректный JSON" {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	if w = do(r, "POST", "/x", `{"login":"a","age":20}`, nil); w.Code != 204 {
		t.Fatalf("ok: %d", w.Code)
	}
}

func TestParsePage(t *testing.T) {
	cases := []struct {
		q          string
		page, size int
	}{
		{"", 1, 20}, {"?page=3&size=50", 3, 50}, {"?size=1000", 1, 100},
		{"?page=0&size=0", 1, 20}, {"?page=x&size=-5", 1, 20},
	}
	for _, tc := range cases {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/"+tc.q, nil)
		p, s := ParsePage(c)
		if p != tc.page || s != tc.size {
			t.Errorf("%q: got %d/%d", tc.q, p, s)
		}
	}
	b, _ := json.Marshal(NewPage[int](nil, 1, 20, 0))
	if !strings.Contains(string(b), `"items":[]`) {
		t.Fatalf("items не должен быть null: %s", b)
	}
}

func TestIfMatch(t *testing.T) {
	for in, want := range map[string]int{"7": 7, `"7"`: 7} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("PATCH", "/", nil)
		c.Request.Header.Set("If-Match", in)
		if v, err := IfMatch(c); err != nil || v != want {
			t.Errorf("%q: %d %v", in, v, err)
		}
	}
	for _, in := range []string{"", "abc"} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("PATCH", "/", nil)
		c.Request.Header.Set("If-Match", in)
		if _, err := IfMatch(c); err == nil {
			t.Errorf("%q: ожидалась ошибка", in)
		}
	}
}

func TestNoRouteAndRecovery(t *testing.T) {
	log := discardLogger()
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(TraceID(), Logger(log), Recovery(log))
	r.NoRoute(NoRoute)
	r.NoMethod(NoMethod)
	r.GET("/panic", func(c *gin.Context) { panic("x") })
	r.GET("/only-get", func(c *gin.Context) { c.Status(200) })

	for _, tc := range []struct {
		method, path string
		status       int
		code         string
	}{
		{"GET", "/nope", 404, CodeNotFound},
		{"POST", "/only-get", 405, CodeMethodNotAllowed},
		{"GET", "/panic", 500, CodeInternal},
	} {
		w := do(r, tc.method, tc.path, "", nil)
		var b ErrorBody
		_ = json.Unmarshal(w.Body.Bytes(), &b)
		if w.Code != tc.status || b.Code != tc.code || b.TraceID == "" {
			t.Errorf("%s %s: %d %s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
}

func TestBindStrict(t *testing.T) {
	type nested struct {
		PersonID string `json:"personId" binding:"required"`
	}
	type req struct {
		Amount *int64  `json:"amount" binding:"omitempty,gt=0"`
		Nested *nested `json:"nested"`
	}
	r := gin.New()
	r.Use(TraceID())
	r.PATCH("/x", func(c *gin.Context) {
		var q req
		if err := BindStrict(c, &q); err != nil {
			Fail(c, err)
			return
		}
		c.Status(204)
	})
	for _, tc := range []struct {
		body, field string
		status      int
	}{
		{`{"amount":5}`, "", 204},
		{`{"nested":{"personId":"p-1"}}`, "", 204},
		{`{"company":{"inn":"1"}}`, "company", 400},
		{`{"nested":{"personId":"p-1","fullName":"x"}}`, "fullName", 400},
		{`{"amount":-1}`, "amount", 400},
		{`{"nested":{}}`, "nested.personId", 400},
		{`{bad`, "", 400},
	} {
		w := do(r, "PATCH", "/x", tc.body, nil)
		var b ErrorBody
		_ = json.Unmarshal(w.Body.Bytes(), &b)
		if w.Code != tc.status || tc.field != "" && (len(b.Fields) != 1 || b.Fields[0].Field != tc.field) {
			t.Errorf("%s: %d %s", tc.body, w.Code, w.Body.String())
		}
	}
}
