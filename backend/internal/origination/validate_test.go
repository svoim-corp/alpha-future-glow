package origination

import (
	"net/http"
	"testing"
)

func TestValidate(t *testing.T) {
	e := newEnv(t)
	id := e.draft(t, "smirnov").ID
	path := "/api/v1/applications/" + id + "/validate"

	got := decode[validateResponse](t, e.do("POST", path, "smirnov", ""))
	if got.Valid || len(got.Missing) != 4 || got.Missing[0].Field != "prelim" {
		t.Fatalf("новый черновик: %+v", got)
	}

	e.passPrelim(t, id)
	e.do("POST", "/api/v1/applications/"+id+"/consents/bki", "smirnov", consentBody)
	e.patch(id, "smirnov", "1", `{"stage2":{"guarantor":{"personId":"p-1"}}}`)
	e.do("POST", "/api/v1/applications/"+id+"/guarantor/consent-bki", "smirnov", `{"given":true}`)
	got = decode[validateResponse](t, e.do("POST", path, "smirnov", ""))
	if got.Valid || len(got.Missing) != 1 || got.Missing[0].Field != "stage2.truthConfirmed" || got.Missing[0].Message == "" {
		t.Fatalf("осталось только подтвердить достоверность: %+v", got)
	}

	e.companies.upstream = true
	if w := e.do("POST", path, "smirnov", ""); w.Code != http.StatusBadGateway {
		t.Fatalf("данные банка недоступны: %d", w.Code)
	}
	e.companies.upstream = false
	if w := e.do("POST", path, "orlova", ""); w.Code != http.StatusForbidden {
		t.Fatalf("чужая заявка: %d", w.Code)
	}
}
