package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

const (
	requestTimeout = 3 * time.Second
	dateLayout     = "2006-01-02"
)

// Названия источников для сообщения «Сервис … не ответил».
const (
	nameBank        = "АБС банка"
	nameEGRUL       = "ЕГРЮЛ"
	nameEGRIP       = "ЕГРИП"
	nameTransparent = "Прозрачный бизнес"
	nameBKI         = "БКИ"
	nameFNS         = "ФНС"
	nameGIRBO       = "ГИР БО"
	nameCourts      = "КАД / ФССП / ЕФРСБ"
	nameMCHD        = "Реестр МЧД"
)

// client — HTTP-клиент к контейнеру моков (API_SPEC.md, 4.14). Реализует shared.BankData и shared.Registries.
type client struct {
	base string
	http *http.Client
}

func newClient(base string, timeout time.Duration) *client {
	return &client{base: strings.TrimRight(base, "/"), http: &http.Client{Timeout: timeout}}
}

// do выполняет запрос к моку и читает JSON в out; false без ошибки — мок ответил 404.
// Сеть, таймаут, неожиданный код или битый ответ — httpx.Upstream: подробности остаются в цепочке ошибки, клиенту не уходят.
func (c *client) do(ctx context.Context, method, path string, body any, source string, out any) (bool, error) {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return false, err
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, rd)
	if err != nil {
		return false, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false, upstream(source, path, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return false, nil
	default:
		return false, upstream(source, path, fmt.Errorf("status %d", resp.StatusCode))
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return false, upstream(source, path, fmt.Errorf("decode: %w", err))
	}
	return true, nil
}

func upstream(source, path string, cause error) error {
	return fmt.Errorf("mock %s: %v: %w", path, cause, httpx.Upstream("Сервис «"+source+"» не ответил, повторите попытку позже"))
}

func parseDate(source, s string) (time.Time, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, httpx.Upstream("Сервис «" + source + "» вернул некорректные данные")
	}
	return t, nil
}

func esc(s string) string { return url.PathEscape(s) }

// Ответы моков. Схема мока и доменные типы меняются независимо, поэтому DTO свои.

type clientDTO struct {
	INN           string      `json:"inn"`
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	ClientSince   string      `json:"clientSince"`
	RegAddress    string      `json:"regAddress"`
	ActualAddress string      `json:"actualAddress"`
	TaxRegime     string      `json:"taxRegime"`
	Persons       []personDTO `json:"persons"`
}

type personDTO struct {
	PersonID   string   `json:"personId"`
	FullName   string   `json:"fullName"`
	INN        string   `json:"inn"`
	Roles      []string `json:"roles"`
	Share      float64  `json:"share"`
	MCHDNumber *string  `json:"mchdNumber"`
}

func (c *client) BankClient(ctx context.Context, inn string) (*shared.BankClient, error) {
	var d clientDTO
	if ok, err := c.do(ctx, http.MethodGet, "/mock/core-banking/clients/"+esc(inn), nil, nameBank, &d); !ok || err != nil {
		return nil, err
	}
	since, err := parseDate(nameBank, d.ClientSince)
	if err != nil {
		return nil, err
	}
	bc := &shared.BankClient{
		INN: d.INN, Name: d.Name, Type: shared.CompanyType(d.Type), ClientSince: since,
		RegAddress: d.RegAddress, ActualAddress: d.ActualAddress, TaxRegime: d.TaxRegime,
		Persons: make([]shared.BankPerson, 0, len(d.Persons)),
	}
	for _, p := range d.Persons {
		roles := make([]shared.PersonRole, len(p.Roles))
		for i, r := range p.Roles {
			roles[i] = shared.PersonRole(r)
		}
		bc.Persons = append(bc.Persons, shared.BankPerson{
			PersonID: p.PersonID, FullName: p.FullName, INN: p.INN, Roles: roles, Share: p.Share, MCHDNumber: p.MCHDNumber,
		})
	}
	return bc, nil
}

type turnoverDTO struct {
	AvgMonthlyIncome int64 `json:"avgMonthlyIncome"`
	Months           []struct {
		Month   string `json:"month"`
		Income  int64  `json:"income"`
		Outcome int64  `json:"outcome"`
	} `json:"months"`
}

func (c *client) Turnover(ctx context.Context, inn string) (*shared.TurnoverReport, error) {
	var d turnoverDTO
	if ok, err := c.do(ctx, http.MethodGet, "/mock/core-banking/turnover/"+esc(inn), nil, nameBank, &d); !ok || err != nil {
		return nil, err
	}
	t := &shared.TurnoverReport{AvgMonthlyIncome: d.AvgMonthlyIncome, Months: make([]shared.TurnoverMonth, len(d.Months))}
	for i, m := range d.Months {
		t.Months[i] = shared.TurnoverMonth{Month: m.Month, Income: m.Income, Outcome: m.Outcome}
	}
	return t, nil
}

type okvedDTO struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type registryPersonDTO struct {
	FullName string `json:"fullName"`
	INN      string `json:"inn"`
}

type egrulDTO struct {
	INN        string            `json:"inn"`
	Name       string            `json:"name"`
	OGRN       string            `json:"ogrn"`
	State      string            `json:"state"`
	RegDate    string            `json:"regDate"`
	RegAddress string            `json:"regAddress"`
	OKVED      okvedDTO          `json:"okved"`
	Head       registryPersonDTO `json:"head"`
	Founders   []struct {
		FullName string  `json:"fullName"`
		Share    float64 `json:"share"`
	} `json:"founders"`
}

type egripDTO struct {
	INN      string   `json:"inn"`
	OGRNIP   string   `json:"ogrnip"`
	FullName string   `json:"fullName"`
	State    string   `json:"state"`
	RegDate  string   `json:"regDate"`
	OKVED    okvedDTO `json:"okved"`
}

// Registration ходит в ЕГРИП для ИП и в ЕГРЮЛ для остальных.
func (c *client) Registration(ctx context.Context, inn string, t shared.CompanyType) (*shared.Registration, error) {
	if t == shared.CompanyIE {
		return c.egrip(ctx, inn)
	}
	return c.egrul(ctx, inn)
}

func (c *client) egrul(ctx context.Context, inn string) (*shared.Registration, error) {
	var d egrulDTO
	if ok, err := c.do(ctx, http.MethodGet, "/mock/egrul/"+esc(inn), nil, nameEGRUL, &d); !ok || err != nil {
		return nil, err
	}
	reg, err := parseDate(nameEGRUL, d.RegDate)
	if err != nil {
		return nil, err
	}
	r := &shared.Registration{
		INN: d.INN, Name: d.Name, OGRN: d.OGRN, State: d.State, RegDate: reg, RegAddress: d.RegAddress,
		OKVED:    shared.OKVED{Code: d.OKVED.Code, Name: d.OKVED.Name},
		Head:     &shared.RegistryPerson{FullName: d.Head.FullName, INN: d.Head.INN},
		Founders: make([]shared.RegistryFounder, len(d.Founders)),
	}
	for i, f := range d.Founders {
		r.Founders[i] = shared.RegistryFounder{FullName: f.FullName, Share: f.Share}
	}
	return r, nil
}

func (c *client) egrip(ctx context.Context, inn string) (*shared.Registration, error) {
	var d egripDTO
	if ok, err := c.do(ctx, http.MethodGet, "/mock/egrip/"+esc(inn), nil, nameEGRIP, &d); !ok || err != nil {
		return nil, err
	}
	reg, err := parseDate(nameEGRIP, d.RegDate)
	if err != nil {
		return nil, err
	}
	return &shared.Registration{
		INN: d.INN, Name: d.FullName, OGRN: d.OGRNIP, State: d.State, RegDate: reg,
		OKVED: shared.OKVED{Code: d.OKVED.Code, Name: d.OKVED.Name},
	}, nil
}

type mchdDTO struct {
	Valid          bool              `json:"valid"`
	PrincipalINN   string            `json:"principalInn"`
	Representative registryPersonDTO `json:"representative"`
	ValidUntil     string            `json:"validUntil"`
	Powers         []string          `json:"powers"`
}

func (c *client) MCHD(ctx context.Context, number string) (*shared.MCHDRecord, error) {
	var d mchdDTO
	if ok, err := c.do(ctx, http.MethodGet, "/mock/mchd/"+esc(number), nil, nameMCHD, &d); !ok || err != nil {
		return nil, err
	}
	until, err := parseDate(nameMCHD, d.ValidUntil)
	if err != nil {
		return nil, err
	}
	return &shared.MCHDRecord{
		Valid: d.Valid, PrincipalINN: d.PrincipalINN, ValidUntil: until, Powers: d.Powers,
		Representative: shared.RegistryPerson{FullName: d.Representative.FullName, INN: d.Representative.INN},
	}, nil
}

type transparentDTO struct {
	MassAddress bool   `json:"massAddress"`
	MassHead    bool   `json:"massHead"`
	TaxRegime   string `json:"taxRegime"`
	Employees   int    `json:"employees"`
}

func (c *client) transparent(ctx context.Context, inn string) (*shared.TransparentBusiness, error) {
	var d transparentDTO
	if ok, err := c.do(ctx, http.MethodGet, "/mock/transparent-business/"+esc(inn), nil, nameTransparent, &d); !ok || err != nil {
		return nil, err
	}
	return &shared.TransparentBusiness{MassAddress: d.MassAddress, MassHead: d.MassHead, TaxRegime: d.TaxRegime, Employees: d.Employees}, nil
}

type bkiRequestDTO struct {
	INN       string `json:"inn"`
	ConsentID string `json:"consentId"`
}

type bkiDTO struct {
	Loans []struct {
		Amount      int64 `json:"amount"`
		Balance     int64 `json:"balance"`
		OverdueDays int   `json:"overdueDays"`
	} `json:"loans"`
	TotalDebt      int64 `json:"totalDebt"`
	MaxOverdueDays int   `json:"maxOverdueDays"`
}

func (c *client) bki(ctx context.Context, inn, consentID string) (*shared.BKIReport, error) {
	var d bkiDTO
	if ok, err := c.do(ctx, http.MethodPost, "/mock/bki/report", bkiRequestDTO{INN: inn, ConsentID: consentID}, nameBKI, &d); !ok || err != nil {
		return nil, err
	}
	r := &shared.BKIReport{TotalDebt: d.TotalDebt, MaxOverdueDays: d.MaxOverdueDays, Loans: make([]shared.BKILoan, len(d.Loans))}
	for i, l := range d.Loans {
		r.Loans[i] = shared.BKILoan{Amount: l.Amount, Balance: l.Balance, OverdueDays: l.OverdueDays}
	}
	return r, nil
}

type fnsDTO struct {
	HasDebt bool  `json:"hasDebt"`
	Amount  int64 `json:"amount"`
}

func (c *client) fns(ctx context.Context, inn string) (*shared.TaxDebt, error) {
	var d fnsDTO
	if ok, err := c.do(ctx, http.MethodGet, "/mock/fns/debts/"+esc(inn), nil, nameFNS, &d); !ok || err != nil {
		return nil, err
	}
	return &shared.TaxDebt{HasDebt: d.HasDebt, Amount: d.Amount}, nil
}

type girboDTO struct {
	Year      int   `json:"year"`
	Revenue   int64 `json:"revenue"`
	NetProfit int64 `json:"netProfit"`
}

func (c *client) girbo(ctx context.Context, inn string) (*shared.GIRBOReport, error) {
	var d girboDTO
	if ok, err := c.do(ctx, http.MethodGet, "/mock/girbo/"+esc(inn), nil, nameGIRBO, &d); !ok || err != nil {
		return nil, err
	}
	return &shared.GIRBOReport{Year: d.Year, Revenue: d.Revenue, NetProfit: d.NetProfit}, nil
}

type courtsDTO struct {
	Arbitration []struct {
		CaseNumber string `json:"caseNumber"`
		Role       string `json:"role"`
		Amount     int64  `json:"amount"`
	} `json:"arbitration"`
	Enforcement []struct {
		Number string `json:"number"`
		Amount int64  `json:"amount"`
	} `json:"enforcement"`
	Bankruptcy bool `json:"bankruptcy"`
}

func (c *client) courts(ctx context.Context, inn string) (*shared.CourtRecords, error) {
	var d courtsDTO
	if ok, err := c.do(ctx, http.MethodGet, "/mock/courts/"+esc(inn), nil, nameCourts, &d); !ok || err != nil {
		return nil, err
	}
	r := &shared.CourtRecords{
		Bankruptcy:  d.Bankruptcy,
		Arbitration: make([]shared.ArbitrationCase, len(d.Arbitration)),
		Enforcement: make([]shared.Enforcement, len(d.Enforcement)),
	}
	for i, a := range d.Arbitration {
		r.Arbitration[i] = shared.ArbitrationCase{CaseNumber: a.CaseNumber, Role: a.Role, Amount: a.Amount}
	}
	for i, e := range d.Enforcement {
		r.Enforcement[i] = shared.Enforcement{Number: e.Number, Amount: e.Amount}
	}
	return r, nil
}
