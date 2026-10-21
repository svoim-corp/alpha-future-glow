package company

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func init() { gin.SetMode(gin.TestMode) }

var (
	llc = shared.CompanyView{ID: "c-1", INN: "7712345678", Name: "ООО «Северная пекарня»", Type: shared.CompanyLLC}
	ie  = shared.CompanyView{ID: "c-11", INN: "773498765432", Name: "ИП Кузнецов Андрей Павлович", Type: shared.CompanyIE}

	head    = shared.BankPerson{PersonID: "p-1", FullName: "Смирнов Олег Викторович", INN: "771234567890", Roles: []shared.PersonRole{shared.PersonHead, shared.PersonFounder}, Share: 60}
	founder = shared.BankPerson{PersonID: "p-3", FullName: "Смирнова Ирина Павловна", INN: "771234567899", Roles: []shared.PersonRole{shared.PersonFounder}, Share: 40}
	rep     = shared.BankPerson{PersonID: "p-2", FullName: "Петрова Анна Сергеевна", INN: "771298765432", Roles: []shared.PersonRole{shared.PersonRepresentative}}
	ieHead  = shared.BankPerson{PersonID: "p-111", FullName: "Кузнецов Андрей Павлович", INN: "773498765432", Roles: []shared.PersonRole{shared.PersonHead}}
)

func ids(ps []shared.Person) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.PersonID
	}
	return out
}

func TestBuildProfile(t *testing.T) {
	reg := &shared.Registration{OGRN: "1157746000000", State: "ACTIVE", OKVED: shared.OKVED{Code: "10.71"}}
	turnover := &shared.TurnoverReport{AvgMonthlyIncome: 1840000, Months: make([]shared.TurnoverMonth, 12)}

	tests := []struct {
		name       string
		cv         shared.CompanyView
		persons    []shared.BankPerson
		reg        *shared.Registration
		girbo      *shared.GIRBOReport
		head       string
		founders   int
		candidates []string
		sources    []shared.CheckSource
	}{
		{"ООО: руководитель-учредитель, учредитель и представитель", llc, []shared.BankPerson{head, rep, founder}, reg, &shared.GIRBOReport{},
			"p-1", 2, []string{"p-1", "p-3"}, []shared.CheckSource{shared.SourceBank, shared.SourceEGRUL}},
		{"ООО: реестр не ответил", llc, []shared.BankPerson{head}, nil, nil,
			"p-1", 1, []string{"p-1"}, []shared.CheckSource{shared.SourceBank}},
		{"ИП: поручителя нет", ie, []shared.BankPerson{ieHead}, reg, nil,
			"p-111", 0, []string{}, []shared.CheckSource{shared.SourceBank, shared.SourceEGRUL}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := buildProfile(tt.cv, &shared.BankClient{Persons: tt.persons, TaxRegime: "OSN"}, turnover, tt.reg, tt.girbo)
			if p.Head.PersonID != tt.head || len(p.Founders) != tt.founders || !slices.Equal(ids(p.GuarantorCandidates), tt.candidates) ||
				!slices.Equal(p.Sources, tt.sources) || p.GIRBOAccounted != (tt.girbo != nil) {
				t.Fatalf("%+v", p)
			}
			if p.Turnover.AvgMonthlyIncome != 1840000 || p.Turnover.PeriodMonths != 12 || p.TaxRegime != "OSN" {
				t.Fatalf("turnover/tax: %+v", p)
			}
		})
	}

	// У кандидата в поручители только роли HEAD/FOUNDER, а исходные роли человека не меняются.
	both := head
	both.Roles = []shared.PersonRole{shared.PersonHead, shared.PersonRepresentative, shared.PersonFounder}
	p := buildProfile(llc, &shared.BankClient{Persons: []shared.BankPerson{both}}, nil, nil, nil)
	if !slices.Equal(p.GuarantorCandidates[0].Roles, []shared.PersonRole{shared.PersonHead, shared.PersonFounder}) || len(p.Head.Roles) != 3 {
		t.Fatalf("roles: %+v / %+v", p.GuarantorCandidates[0].Roles, p.Head.Roles)
	}
}

type fakeBank struct {
	client *shared.BankClient
	err    error
}

func (f fakeBank) BankClient(context.Context, string) (*shared.BankClient, error) {
	return f.client, f.err
}
func (f fakeBank) Turnover(context.Context, string) (*shared.TurnoverReport, error) {
	return &shared.TurnoverReport{}, nil
}

type fakeRegistries struct {
	regErr, girboErr error
	girboCalls       int
}

func (f *fakeRegistries) Registration(context.Context, string, shared.CompanyType) (*shared.Registration, error) {
	if f.regErr != nil {
		return nil, f.regErr
	}
	return &shared.Registration{State: "ACTIVE"}, nil
}
func (f *fakeRegistries) MCHD(context.Context, string) (*shared.MCHDRecord, error) { return nil, nil }
func (f *fakeRegistries) GIRBO(context.Context, string) (*shared.GIRBOReport, error) {
	f.girboCalls++
	if f.girboErr != nil {
		return nil, f.girboErr
	}
	return &shared.GIRBOReport{Year: 2025}, nil
}

func TestFetch(t *testing.T) {
	upstream := httpx.Upstream("")
	bank := &shared.BankClient{Persons: []shared.BankPerson{head}}
	tests := []struct {
		name     string
		cv       shared.CompanyView
		bank     fakeBank
		reg      *fakeRegistries
		wantErr  func(error) bool
		complete bool
		girbo    int
	}{
		{"всё ответило", llc, fakeBank{client: bank}, &fakeRegistries{}, nil, true, 1},
		{"АБС недоступна — 502", llc, fakeBank{err: upstream}, &fakeRegistries{}, isUpstream, false, 0},
		{"клиента нет в АБС — внутренняя ошибка", llc, fakeBank{}, &fakeRegistries{}, func(err error) bool { return err != nil && !isUpstream(err) }, false, 0},
		{"ЕГРЮЛ недоступен — профиль без реестра", llc, fakeBank{client: bank}, &fakeRegistries{regErr: upstream}, nil, false, 1},
		{"ГИР БО недоступен — профиль без отчётности", llc, fakeBank{client: bank}, &fakeRegistries{girboErr: upstream}, nil, false, 1},
		{"ошибка реестра не сетевая — наружу", llc, fakeBank{client: bank}, &fakeRegistries{regErr: errors.New("boom")}, func(err error) bool { return err != nil }, false, 0},
		{"ИП — ГИР БО не запрашивается", ie, fakeBank{client: bank}, &fakeRegistries{}, nil, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &profiles{bank: tt.bank, reg: tt.reg, now: time.Now}
			prof, complete, err := p.fetch(context.Background(), tt.cv)
			if tt.wantErr != nil {
				if !tt.wantErr(err) || prof != nil {
					t.Fatalf("err = %v, prof = %+v", err, prof)
				}
				return
			}
			if err != nil || prof == nil || complete != tt.complete || tt.reg.girboCalls != tt.girbo {
				t.Fatalf("err = %v, complete = %v, girbo calls = %d", err, complete, tt.reg.girboCalls)
			}
		})
	}
}

type fakeApplications struct {
	shared.Applications
	apps map[string]*shared.ApplicationView
}

func (f fakeApplications) Application(_ context.Context, id string) (*shared.ApplicationView, error) {
	return f.apps[id], nil
}

type fakeResolver map[string]*auth.User

func (f fakeResolver) Resolve(_ context.Context, token string) (*auth.User, error) {
	return f[token], nil
}

func TestCompanyProfileAccess(t *testing.T) {
	c1, c2 := "c-1", "c-2"
	m := New(module.Deps{}) // без БД: до профиля запрос не доходит
	m.Wire(Ports{Applications: fakeApplications{apps: map[string]*shared.ApplicationView{"a-1": {ID: "a-1", CompanyID: c1}}}})

	r := gin.New()
	r.Use(httpx.TraceID(), auth.Middleware(fakeResolver{
		"other":   {ID: "u-3", Role: shared.RoleClient, CompanyID: &c2},
		"officer": {ID: "u-9", Role: shared.RoleBankOfficer},
	}))
	m.Register(httpx.Routes{Client: r.Group("", auth.RequireAuth())})

	tests := []struct {
		token, app string
		code       int
	}{
		{"other", "a-404", http.StatusNotFound},
		{"other", "a-1", http.StatusForbidden},
		{"officer", "a-1", http.StatusForbidden},
		{"", "a-1", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "/applications/"+tt.app+"/company-profile", nil)
		if tt.token != "" {
			req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: tt.token})
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tt.code {
			t.Errorf("%s %s: %d, want %d", tt.token, tt.app, w.Code, tt.code)
		}
	}
}
