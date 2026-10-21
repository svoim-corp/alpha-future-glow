package shared

import (
	"context"
	"time"
)

// Companies — данные клиента банка из зоны Company. Реализует: разработчик B. Используют: Identity, Origination, Offer (A), Scoring, Underwriting (B).
type Companies interface {
	// GetByID — краткие данные компании; (nil, nil) — компании нет.
	GetByID(ctx context.Context, id string) (*CompanyView, error)
	// Profile — профиль из клиентской базы банка, сверенный с реестрами (API_SPEC.md, company-profile); (nil, nil) — компании нет.
	Profile(ctx context.Context, companyID string) (*CompanyProfile, error)
}

type CompanyType string

const (
	CompanyLLC CompanyType = "LLC"
	CompanyIE  CompanyType = "IE"
)

type CompanyView struct {
	ID, INN, Name string
	Type          CompanyType
}

// CompanyProfile — всё, что банк знает о клиенте. Клиент эти данные не вводит и не правит (Уточн. ТЗ).
type CompanyProfile struct {
	CompanyView
	OGRN                string
	State               string // ACTIVE, LIQUIDATING, …
	RegDate             time.Time
	ClientSince         time.Time
	RegAddress          string
	ActualAddress       string
	OKVED               OKVED
	TaxRegime           string
	Head                Person
	Founders            []Founder
	GuarantorCandidates []Person      // руководитель и учредители-физлица; для ИП — пусто
	Turnover            Turnover      // по счетам в банке
	Sources             []CheckSource // откуда собран профиль: BANK, EGRUL
	GIRBOAccounted      bool          // отчётность в ГИР БО учтена; цифры клиенту не показываются
}

type OKVED struct{ Code, Name string }

type PersonRole string

const (
	PersonHead           PersonRole = "HEAD"
	PersonFounder        PersonRole = "FOUNDER"
	PersonRepresentative PersonRole = "REPRESENTATIVE"
)

type Person struct {
	PersonID string
	FullName string
	INN      string
	Roles    []PersonRole
}

type Founder struct {
	Person
	Share float64 // доля в процентах
}

type Turnover struct {
	AvgMonthlyIncome int64
	PeriodMonths     int
}

// Authorities — полномочия заявителя по данным банка и реестру МЧД. Реализует: разработчик B. Использует: Origination (A)
// при создании черновика; результат Origination хранит в заявке как снимок.
type Authorities interface {
	// Resolve определяет, может ли человек (personINN из пользователя) действовать от имени компании.
	// Отсутствие полномочий — не ошибка, а AuthorityNone с причиной.
	Resolve(ctx context.Context, companyID, personINN string) (AuthorityView, error)
}

type AuthorityType string

const (
	AuthorityHead AuthorityType = "HEAD"
	AuthorityMCHD AuthorityType = "MCHD"
	AuthorityNone AuthorityType = "NONE"
)

// Полномочия по МЧД (API_SPEC.md, 4.14).
const (
	PowerCreditApplication = "CREDIT_APPLICATION"
	PowerBKIConsent        = "BKI_CONSENT"
)

type AuthorityView struct {
	Type       AuthorityType
	MCHDNumber *string
	ValidUntil *time.Time
	Powers     []string // у руководителя — все
	Reason     string   // для AuthorityNone: понятный клиенту текст для 403
}
