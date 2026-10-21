package shared

import (
	"context"
	"time"
)

// BankData — данные клиента в АБС банка (API_SPEC.md, 4.14). Реализует: разработчик B (Integrations). Использует: Company (B).
// Недоступность АБС — httpx.Upstream.
type BankData interface {
	// BankClient — карточка клиента и люди, известные банку; (nil, nil) — клиента нет.
	BankClient(ctx context.Context, inn string) (*BankClient, error)
	// Turnover — обороты по счетам в банке; (nil, nil) — клиента нет.
	Turnover(ctx context.Context, inn string) (*TurnoverReport, error)
}

// Registries — государственные реестры. Реализует: разработчик B (Integrations). Использует: Company (B).
// Недоступность реестра — httpx.Upstream.
type Registries interface {
	// Registration — запись ЕГРЮЛ (ООО) или ЕГРИП (ИП); (nil, nil) — записи нет.
	Registration(ctx context.Context, inn string, t CompanyType) (*Registration, error)
	// MCHD — доверенность из реестра МЧД; (nil, nil) — доверенность не найдена.
	MCHD(ctx context.Context, number string) (*MCHDRecord, error)
	// GIRBO — последняя бухгалтерская отчётность ООО; (nil, nil) — отчётности нет.
	GIRBO(ctx context.Context, inn string) (*GIRBOReport, error)
}

// Checks — внешние проверки по заявке со снимками ответов (CheckResult). Реализует: разработчик B (Integrations).
// Используют: Scoring, Underwriting (B). Integrations только описывает ответ источника; решение по нему принимает Scoring.
type Checks interface {
	// Run опрашивает источники и сохраняет снимки. Не ответивший источник — результат UNAVAILABLE, а не ошибка.
	Run(ctx context.Context, req CheckRequest) (*CheckReport, error)
	// Latest — последний снимок по каждому источнику; пустой отчёт — проверок не было.
	Latest(ctx context.Context, appID string) (*CheckReport, error)
}

type BankClient struct {
	INN, Name     string
	Type          CompanyType
	ClientSince   time.Time
	RegAddress    string
	ActualAddress string
	TaxRegime     string
	Persons       []BankPerson
}

type BankPerson struct {
	PersonID   string
	FullName   string
	INN        string
	Roles      []PersonRole
	Share      float64 // доля учредителя в процентах; 0 — не учредитель
	MCHDNumber *string // только у представителя
}

type TurnoverReport struct {
	AvgMonthlyIncome int64
	Months           []TurnoverMonth
}

type TurnoverMonth struct {
	Month           string // "2026-09"
	Income, Outcome int64
}

// Registration — запись ЕГРЮЛ или ЕГРИП. У ИП Name — ФИО, OGRN — ОГРНИП, Head и Founders пустые.
type Registration struct {
	INN, Name, OGRN string
	State           string // ACTIVE, LIQUIDATING, …
	RegDate         time.Time
	RegAddress      string
	OKVED           OKVED
	Head            *RegistryPerson
	Founders        []RegistryFounder
}

type RegistryPerson struct{ FullName, INN string }

type RegistryFounder struct {
	FullName string
	Share    float64
}

type MCHDRecord struct {
	Valid          bool
	PrincipalINN   string
	Representative RegistryPerson
	ValidUntil     time.Time
	Powers         []string
}

type CheckSource string

const (
	SourceBank                CheckSource = "BANK"
	SourceBankAccounts        CheckSource = "BANK_ACCOUNTS"
	SourceEGRUL               CheckSource = "EGRUL" // ЕГРЮЛ или ЕГРИП — по типу компании
	SourceTransparentBusiness CheckSource = "TRANSPARENT_BUSINESS"
	SourceBKI                 CheckSource = "BKI"
	SourceFNS                 CheckSource = "FNS"
	SourceGIRBO               CheckSource = "GIRBO" // только ООО
	SourceCourts              CheckSource = "COURTS"
)

// AllSources — все источники в порядке показа.
var AllSources = []CheckSource{
	SourceBank, SourceBankAccounts, SourceEGRUL, SourceTransparentBusiness, SourceBKI, SourceFNS, SourceGIRBO, SourceCourts,
}

type CheckStatus string

const (
	CheckOK          CheckStatus = "OK"
	CheckWarn        CheckStatus = "WARN"
	CheckFail        CheckStatus = "FAIL"
	CheckUnavailable CheckStatus = "UNAVAILABLE"
)

type CheckRequest struct {
	AppID       string
	INN         string
	CompanyType CompanyType
	ConsentID   string        // согласие на запрос в БКИ; пусто — БКИ не запрашивается
	Sources     []CheckSource // пусто — все источники
}

type CheckResult struct {
	Source      CheckSource
	Status      CheckStatus
	RequestedAt time.Time
	Summary     string         // понятный текст для сотрудника
	Details     map[string]any // ключевые цифры для экрана разбора
}

// CheckReport — результаты для показа и разобранные ответы источников для правил Scoring.
// nil в Facts — источник не запрашивали, не ответил или данных нет.
type CheckReport struct {
	Results []CheckResult
	Facts   CheckFacts
}

type CheckFacts struct {
	Client       *BankClient
	Turnover     *TurnoverReport
	Registration *Registration
	Transparent  *TransparentBusiness
	BKI          *BKIReport
	FNS          *TaxDebt
	GIRBO        *GIRBOReport
	Courts       *CourtRecords
}

type TransparentBusiness struct {
	MassAddress, MassHead bool
	TaxRegime             string
	Employees             int
}

type BKIReport struct {
	Loans          []BKILoan
	TotalDebt      int64
	MaxOverdueDays int
}

type BKILoan struct {
	Amount, Balance int64
	OverdueDays     int
}

type TaxDebt struct {
	HasDebt bool
	Amount  int64
}

type GIRBOReport struct {
	Year               int
	Revenue, NetProfit int64
}

type CourtRecords struct {
	Arbitration []ArbitrationCase
	Enforcement []Enforcement
	Bankruptcy  bool
}

type ArbitrationCase struct {
	CaseNumber string
	Role       string // DEFENDANT, PLAINTIFF
	Amount     int64
}

type Enforcement struct {
	Number string
	Amount int64
}
