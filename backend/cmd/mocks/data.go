package main

import (
	"fmt"
	"math"
)

// Ответы моков (API_SPEC.md, 4.14). Даты — строки ISO 8601: мок отдаёт их как есть.

type bankClient struct {
	INN           string       `json:"inn"`
	Name          string       `json:"name"`
	Type          string       `json:"type"`
	ClientSince   string       `json:"clientSince"`
	RegAddress    string       `json:"regAddress"`
	ActualAddress string       `json:"actualAddress"`
	TaxRegime     string       `json:"taxRegime"`
	Persons       []bankPerson `json:"persons"`
}

type bankPerson struct {
	PersonID   string   `json:"personId"`
	FullName   string   `json:"fullName"`
	INN        string   `json:"inn"`
	Roles      []string `json:"roles"`
	Share      *float64 `json:"share,omitempty"`
	MCHDNumber *string  `json:"mchdNumber,omitempty"`
}

type turnover struct {
	AvgMonthlyIncome int64           `json:"avgMonthlyIncome"`
	Months           []turnoverMonth `json:"months"`
}

type turnoverMonth struct {
	Month   string `json:"month"`
	Income  int64  `json:"income"`
	Outcome int64  `json:"outcome"`
}

type okved struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type egrulRecord struct {
	INN        string         `json:"inn"`
	Name       string         `json:"name"`
	OGRN       string         `json:"ogrn"`
	State      string         `json:"state"`
	RegDate    string         `json:"regDate"`
	RegAddress string         `json:"regAddress"`
	OKVED      okved          `json:"okved"`
	Head       egrulPerson    `json:"head"`
	Founders   []egrulFounder `json:"founders"`
}

type egrulPerson struct {
	FullName string `json:"fullName"`
	INN      string `json:"inn"`
}

type egrulFounder struct {
	FullName string  `json:"fullName"`
	Share    float64 `json:"share"`
}

type egripRecord struct {
	INN      string `json:"inn"`
	OGRNIP   string `json:"ogrnip"`
	FullName string `json:"fullName"`
	State    string `json:"state"`
	RegDate  string `json:"regDate"`
	OKVED    okved  `json:"okved"`
}

type transparentBusiness struct {
	MassAddress bool   `json:"massAddress"`
	MassHead    bool   `json:"massHead"`
	TaxRegime   string `json:"taxRegime"`
	Employees   int    `json:"employees"`
}

type bkiReport struct {
	Loans          []bkiLoan `json:"loans"`
	TotalDebt      int64     `json:"totalDebt"`
	MaxOverdueDays int       `json:"maxOverdueDays"`
}

type bkiLoan struct {
	Amount      int64 `json:"amount"`
	Balance     int64 `json:"balance"`
	OverdueDays int   `json:"overdueDays"`
}

type fnsDebts struct {
	HasDebt bool  `json:"hasDebt"`
	Amount  int64 `json:"amount"`
}

type girboReport struct {
	Year      int   `json:"year"`
	Revenue   int64 `json:"revenue"`
	NetProfit int64 `json:"netProfit"`
}

type courts struct {
	Arbitration []arbitrationCase `json:"arbitration"`
	Enforcement []enforcement     `json:"enforcement"`
	Bankruptcy  bool              `json:"bankruptcy"`
}

type arbitrationCase struct {
	CaseNumber string `json:"caseNumber"`
	Role       string `json:"role"`
	Amount     int64  `json:"amount"`
}

type enforcement struct {
	Number string `json:"number"`
	Amount int64  `json:"amount"`
}

type mchdRecord struct {
	Valid          bool        `json:"valid"`
	PrincipalINN   string      `json:"principalInn"`
	Representative egrulPerson `json:"representative"`
	ValidUntil     string      `json:"validUntil"`
	Powers         []string    `json:"powers"`
}

// account — расчётный счёт клиента в АБС. Остаток меняется списаниями, поэтому счёт хранится по указателю.
type account struct {
	ID       string `json:"id"`
	Number   string `json:"number"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Balance  int64  `json:"balance"`
}

// company — всё, что моки знают о компании. egrul — только у ООО, egrip и пустой girbo — у ИП.
type company struct {
	client   bankClient
	turnover turnover
	egrul    *egrulRecord
	egrip    *egripRecord
	tb       transparentBusiness
	bki      bkiReport
	fns      fnsDebts
	girbo    *girboReport
	courts   courts
	accounts []*account
}

type dataset struct {
	companies map[string]*company // по ИНН компании
	mchd      map[string]mchdRecord
}

const (
	roleHead           = "HEAD"
	roleFounder        = "FOUNDER"
	roleRepresentative = "REPRESENTATIVE"

	powerCreditApplication = "CREDIT_APPLICATION"
	powerBKIConsent        = "BKI_CONSENT"

	stateActive      = "ACTIVE"
	stateLiquidating = "LIQUIDATING"

	mchdPetrova = "a1b2c3d4-5e6f-4a7b-8c9d-0e1f2a3b4c5d"
	mchdVolkov  = "5f0c1e2d-3b4a-4c5d-8e9f-0a1b2c3d4e5f"

	// Доля исходящих платежей от поступлений в синтетических оборотах.
	outcomeRatio = 0.88

	accountActive  = "ACTIVE"
	accountBlocked = "BLOCKED"
	currencyRUB    = "RUB"

	// Остаток основного счёта — три средних месячных оборота: хватает на платежи и досрочное погашение демо-кредита.
	balanceMonths = 3
	// У smirnov кроме основного — заблокированный счёт и счёт с малым остатком: для сценариев отказа в списании.
	smallBalance = 10_000
	innSmirnov   = "7712345678"
)

// Профили колебаний месячной выручки; среднее каждого — 1.
var (
	stablePattern   = []float64{0.97, 1.02, 0.95, 1.04, 0.99, 1.01, 1.03, 0.98, 1.00, 1.02, 0.97, 1.02}
	unstablePattern = []float64{1.6, 0.3, 1.4, 0.2, 1.5, 0.6, 1.7, 0.4, 1.3, 0.5, 1.8, 0.7}
)

// months строит обороты за len(pattern) месяцев, начиная с from (год, месяц). Среднее считается по месяцам,
// чтобы avgMonthlyIncome всегда совпадал с данными.
func months(year, month int, base int64, pattern []float64) turnover {
	t := turnover{Months: make([]turnoverMonth, 0, len(pattern))}
	var sum int64
	for _, k := range pattern {
		income := roundTo(float64(base)*k, 1000)
		t.Months = append(t.Months, turnoverMonth{
			Month:   fmt.Sprintf("%04d-%02d", year, month),
			Income:  income,
			Outcome: roundTo(float64(income)*outcomeRatio, 1000),
		})
		sum += income
		if month++; month > 12 {
			year, month = year+1, 1
		}
	}
	t.AvgMonthlyIncome = sum / int64(len(pattern))
	return t
}

func roundTo(v float64, step int64) int64 {
	return int64(math.Round(v/float64(step))) * step
}

func ptr[T any](v T) *T { return &v }

// loans собирает отчёт БКИ: totalDebt и maxOverdueDays выводятся из кредитов.
func loans(ls ...bkiLoan) bkiReport {
	r := bkiReport{Loans: ls}
	if r.Loans == nil {
		r.Loans = []bkiLoan{}
	}
	for _, l := range ls {
		r.TotalDebt += l.Balance
		r.MaxOverdueDays = max(r.MaxOverdueDays, l.OverdueDays)
	}
	return r
}

func noCourts() courts {
	return courts{Arbitration: []arbitrationCase{}, Enforcement: []enforcement{}}
}

// newDataset — синтетические данные: по одной компании на ветку «Схемы решения» (backend/SCENARIOS.md).
// Значения подобраны так, чтобы правила этапа 1 и скоринга вели ровно в ожидаемую ветку.
func newDataset() dataset {
	cs := []*company{
		// c-1, smirnov и petrova (МЧД): всё в порядке, одно некритичное арбитражное дело — одобрение.
		{
			client: bankClient{
				INN: "7712345678", Name: "ООО «Северная пекарня»", Type: "LLC", ClientSince: "2016-02-01",
				RegAddress: "г. Москва, ул. Пекарская, 7", ActualAddress: "г. Москва, ул. Складская, 3",
				TaxRegime: "USN_INCOME_MINUS_EXPENSES",
				Persons: []bankPerson{
					{PersonID: "p-1", FullName: "Смирнов Олег Викторович", INN: "771234567890", Roles: []string{roleHead, roleFounder}, Share: ptr(100.0)},
					{PersonID: "p-2", FullName: "Петрова Анна Сергеевна", INN: "771298765432", Roles: []string{roleRepresentative}, MCHDNumber: ptr(mchdPetrova)},
				},
			},
			turnover: months(2025, 10, 1_840_000, stablePattern),
			egrul: &egrulRecord{
				INN: "7712345678", Name: "ООО «Северная пекарня»", OGRN: "1157746000000", State: stateActive,
				RegDate: "2015-03-12", RegAddress: "г. Москва, ул. Пекарская, 7",
				OKVED:    okved{Code: "10.71", Name: "Производство хлеба и мучных кондитерских изделий"},
				Head:     egrulPerson{FullName: "Смирнов Олег Викторович", INN: "771234567890"},
				Founders: []egrulFounder{{FullName: "Смирнов Олег Викторович", Share: 100}},
			},
			tb:     transparentBusiness{TaxRegime: "USN_INCOME_MINUS_EXPENSES", Employees: 34},
			bki:    loans(bkiLoan{Amount: 1_500_000, Balance: 600_000}, bkiLoan{Amount: 500_000, Balance: 300_000}),
			girbo:  &girboReport{Year: 2025, Revenue: 21_400_000, NetProfit: 2_100_000},
			courts: courts{Arbitration: []arbitrationCase{{CaseNumber: "А40-118734/2026", Role: "DEFENDANT", Amount: 120_000}}, Enforcement: []enforcement{}},
		},
		// c-2, volkov: представитель по МЧД, срок которой истёк, — 403 при создании черновика.
		{
			client: bankClient{
				INN: "7704211530", Name: "ООО «Вектор Логистик»", Type: "LLC", ClientSince: "2018-01-15",
				RegAddress: "г. Москва, Варшавское ш., 118", ActualAddress: "г. Москва, Варшавское ш., 118",
				TaxRegime: "OSN",
				Persons: []bankPerson{
					{PersonID: "p-21", FullName: "Никитин Роман Андреевич", INN: "770455667701", Roles: []string{roleHead, roleFounder}, Share: ptr(100.0)},
					{PersonID: "p-22", FullName: "Волков Игорь Николаевич", INN: "770411223344", Roles: []string{roleRepresentative}, MCHDNumber: ptr(mchdVolkov)},
				},
			},
			turnover: months(2025, 10, 2_600_000, stablePattern),
			egrul: &egrulRecord{
				INN: "7704211530", Name: "ООО «Вектор Логистик»", OGRN: "1177746123450", State: stateActive,
				RegDate: "2017-06-20", RegAddress: "г. Москва, Варшавское ш., 118",
				OKVED:    okved{Code: "52.29", Name: "Деятельность вспомогательная, связанная с перевозками"},
				Head:     egrulPerson{FullName: "Никитин Роман Андреевич", INN: "770455667701"},
				Founders: []egrulFounder{{FullName: "Никитин Роман Андреевич", Share: 100}},
			},
			tb:     transparentBusiness{TaxRegime: "OSN", Employees: 58},
			bki:    loans(bkiLoan{Amount: 3_000_000, Balance: 1_100_000}),
			girbo:  &girboReport{Year: 2025, Revenue: 30_900_000, NetProfit: 2_700_000},
			courts: noCourts(),
		},
		// c-3, orlova: ИП зарегистрирован меньше 6 месяцев назад — стоп BUSINESS_TOO_YOUNG.
		{
			client: bankClient{
				INN: "772512345601", Name: "ИП Орлова Мария Игоревна", Type: "IE", ClientSince: "2026-07-08",
				RegAddress: "г. Москва, ул. Профсоюзная, 41, кв. 12", ActualAddress: "г. Москва, ул. Профсоюзная, 41, кв. 12",
				TaxRegime: "USN_INCOME",
				Persons: []bankPerson{
					{PersonID: "p-31", FullName: "Орлова Мария Игоревна", INN: "772512345601", Roles: []string{roleHead}},
				},
			},
			turnover: months(2026, 7, 420_000, []float64{0.8, 1.0, 1.2}),
			egrip: &egripRecord{
				INN: "772512345601", OGRNIP: "326774600123456", FullName: "Орлова Мария Игоревна", State: stateActive,
				RegDate: "2026-07-01", OKVED: okved{Code: "47.91", Name: "Торговля розничная по почте или через интернет"},
			},
			tb:     transparentBusiness{TaxRegime: "USN_INCOME", Employees: 1},
			bki:    loans(),
			courts: noCourts(),
		},
		// c-4, belov: просрочка 95 дней в БКИ — стоп BKI_OVERDUE.
		{
			client: bankClient{
				INN: "7728334410", Name: "ООО «СтройТехМонтаж»", Type: "LLC", ClientSince: "2013-03-11",
				RegAddress: "г. Москва, ул. Обручева, 23", ActualAddress: "г. Москва, ул. Обручева, 23, стр. 2",
				TaxRegime: "OSN",
				Persons: []bankPerson{
					{PersonID: "p-41", FullName: "Белов Сергей Андреевич", INN: "772855667788", Roles: []string{roleHead, roleFounder}, Share: ptr(60.0)},
					{PersonID: "p-42", FullName: "Белова Ирина Павловна", INN: "772855667799", Roles: []string{roleFounder}, Share: ptr(40.0)},
				},
			},
			turnover: months(2025, 10, 3_100_000, stablePattern),
			egrul: &egrulRecord{
				INN: "7728334410", Name: "ООО «СтройТехМонтаж»", OGRN: "1127746234561", State: stateActive,
				RegDate: "2012-09-04", RegAddress: "г. Москва, ул. Обручева, 23",
				OKVED: okved{Code: "43.21", Name: "Производство электромонтажных работ"},
				Head:  egrulPerson{FullName: "Белов Сергей Андреевич", INN: "772855667788"},
				Founders: []egrulFounder{
					{FullName: "Белов Сергей Андреевич", Share: 60},
					{FullName: "Белова Ирина Павловна", Share: 40},
				},
			},
			tb:     transparentBusiness{TaxRegime: "OSN", Employees: 41},
			bki:    loans(bkiLoan{Amount: 5_000_000, Balance: 2_300_000, OverdueDays: 95}, bkiLoan{Amount: 1_200_000, Balance: 400_000}),
			girbo:  &girboReport{Year: 2025, Revenue: 36_800_000, NetProfit: 1_400_000},
			courts: noCourts(),
		},
		// c-5, gromov: компания в процессе ликвидации — стоп COMPANY_INACTIVE.
		{
			client: bankClient{
				INN: "7719456720", Name: "ООО «Ромашка»", Type: "LLC", ClientSince: "2014-06-02",
				RegAddress: "г. Москва, Измайловский пр-т, 9", ActualAddress: "г. Москва, Измайловский пр-т, 9",
				TaxRegime: "USN_INCOME",
				Persons: []bankPerson{
					{PersonID: "p-51", FullName: "Громов Павел Евгеньевич", INN: "771966778899", Roles: []string{roleHead, roleFounder}, Share: ptr(100.0)},
				},
			},
			turnover: months(2025, 10, 450_000, stablePattern),
			egrul: &egrulRecord{
				INN: "7719456720", Name: "ООО «Ромашка»", OGRN: "1147746345672", State: stateLiquidating,
				RegDate: "2014-04-18", RegAddress: "г. Москва, Измайловский пр-т, 9",
				OKVED:    okved{Code: "56.10", Name: "Деятельность ресторанов и услуги по доставке продуктов питания"},
				Head:     egrulPerson{FullName: "Громов Павел Евгеньевич", INN: "771966778899"},
				Founders: []egrulFounder{{FullName: "Громов Павел Евгеньевич", Share: 100}},
			},
			tb:     transparentBusiness{TaxRegime: "USN_INCOME", Employees: 6},
			bki:    loans(),
			girbo:  &girboReport{Year: 2025, Revenue: 5_600_000, NetProfit: -800_000},
			courts: noCourts(),
		},
		// c-6, zaitsev: сообщение о банкротстве в ЕФРСБ — стоп CRITICAL_RISK.
		{
			client: bankClient{
				INN: "7733567830", Name: "ООО «ТехноСервис»", Type: "LLC", ClientSince: "2017-01-20",
				RegAddress: "г. Москва, ул. Тушинская, 16", ActualAddress: "г. Москва, ул. Тушинская, 16",
				TaxRegime: "OSN",
				Persons: []bankPerson{
					{PersonID: "p-61", FullName: "Зайцев Дмитрий Олегович", INN: "773377889900", Roles: []string{roleHead, roleFounder}, Share: ptr(100.0)},
				},
			},
			turnover: months(2025, 10, 1_200_000, stablePattern),
			egrul: &egrulRecord{
				INN: "7733567830", Name: "ООО «ТехноСервис»", OGRN: "1167746456783", State: stateActive,
				RegDate: "2016-11-02", RegAddress: "г. Москва, ул. Тушинская, 16",
				OKVED:    okved{Code: "33.12", Name: "Ремонт машин и оборудования"},
				Head:     egrulPerson{FullName: "Зайцев Дмитрий Олегович", INN: "773377889900"},
				Founders: []egrulFounder{{FullName: "Зайцев Дмитрий Олегович", Share: 100}},
			},
			tb:    transparentBusiness{TaxRegime: "OSN", Employees: 17},
			bki:   loans(bkiLoan{Amount: 2_000_000, Balance: 1_700_000, OverdueDays: 20}),
			girbo: &girboReport{Year: 2025, Revenue: 14_100_000, NetProfit: -2_300_000},
			courts: courts{
				Arbitration: []arbitrationCase{
					{CaseNumber: "А40-204517/2026", Role: "DEFENDANT", Amount: 4_800_000},
					{CaseNumber: "А40-211093/2026", Role: "DEFENDANT", Amount: 1_250_000},
				},
				Enforcement: []enforcement{{Number: "88412/26/77001-ИП", Amount: 630_000}},
				Bankruptcy:  true,
			},
		},
		// c-7, sokolova: небольшие обороты — лимит меньше запрашиваемой суммы (amountFits: false).
		{
			client: bankClient{
				INN: "7716678940", Name: "ООО «Зелёный сад»", Type: "LLC", ClientSince: "2019-04-10",
				RegAddress: "г. Москва, Ярославское ш., 54", ActualAddress: "Московская обл., г. Мытищи, ул. Садовая, 2",
				TaxRegime: "USN_INCOME",
				Persons: []bankPerson{
					{PersonID: "p-71", FullName: "Соколова Елена Викторовна", INN: "771688990011", Roles: []string{roleHead, roleFounder}, Share: ptr(100.0)},
				},
			},
			turnover: months(2025, 10, 520_000, stablePattern),
			egrul: &egrulRecord{
				INN: "7716678940", Name: "ООО «Зелёный сад»", OGRN: "1197746567894", State: stateActive,
				RegDate: "2019-03-25", RegAddress: "г. Москва, Ярославское ш., 54",
				OKVED:    okved{Code: "01.30", Name: "Выращивание рассады"},
				Head:     egrulPerson{FullName: "Соколова Елена Викторовна", INN: "771688990011"},
				Founders: []egrulFounder{{FullName: "Соколова Елена Викторовна", Share: 100}},
			},
			tb:     transparentBusiness{TaxRegime: "USN_INCOME", Employees: 9},
			bki:    loans(),
			girbo:  &girboReport{Year: 2025, Revenue: 6_100_000, NetProfit: 700_000},
			courts: noCourts(),
		},
		// c-8, morozov: долговая нагрузка выше половины годового оборота — иные условия (COUNTER_OFFER).
		{
			client: bankClient{
				INN: "7707789050", Name: "ООО «Альфа-Трейд»", Type: "LLC", ClientSince: "2014-02-17",
				RegAddress: "г. Москва, ул. Новослободская, 31", ActualAddress: "г. Москва, ул. Новослободская, 31",
				TaxRegime: "OSN",
				Persons: []bankPerson{
					{PersonID: "p-81", FullName: "Морозов Алексей Игоревич", INN: "770799001122", Roles: []string{roleHead, roleFounder}, Share: ptr(70.0)},
					{PersonID: "p-82", FullName: "Морозова Ольга Сергеевна", INN: "770799001133", Roles: []string{roleFounder}, Share: ptr(30.0)},
				},
			},
			turnover: months(2025, 10, 2_200_000, stablePattern),
			egrul: &egrulRecord{
				INN: "7707789050", Name: "ООО «Альфа-Трейд»", OGRN: "1137746678905", State: stateActive,
				RegDate: "2013-08-14", RegAddress: "г. Москва, ул. Новослободская, 31",
				OKVED: okved{Code: "46.90", Name: "Торговля оптовая неспециализированная"},
				Head:  egrulPerson{FullName: "Морозов Алексей Игоревич", INN: "770799001122"},
				Founders: []egrulFounder{
					{FullName: "Морозов Алексей Игоревич", Share: 70},
					{FullName: "Морозова Ольга Сергеевна", Share: 30},
				},
			},
			tb: transparentBusiness{TaxRegime: "OSN", Employees: 26},
			bki: loans(
				bkiLoan{Amount: 8_000_000, Balance: 6_500_000},
				bkiLoan{Amount: 6_000_000, Balance: 4_800_000},
				bkiLoan{Amount: 4_000_000, Balance: 2_700_000},
			),
			girbo:  &girboReport{Year: 2025, Revenue: 25_900_000, NetProfit: 1_100_000},
			courts: noCourts(),
		},
		// c-9, kozlov: обороты по счетам на ~58% выше выручки в ГИР БО — TURNOVER_MISMATCH, сотрудник запрашивает пояснения.
		{
			client: bankClient{
				INN: "7724890160", Name: "ООО «Мебельная фабрика «Дуб»»", Type: "LLC", ClientSince: "2012-05-03",
				RegAddress: "г. Москва, ул. Дубининская, 57", ActualAddress: "Московская обл., г. Подольск, ул. Мебельная, 4",
				TaxRegime: "OSN",
				Persons: []bankPerson{
					{PersonID: "p-91", FullName: "Козлов Виктор Петрович", INN: "772400112233", Roles: []string{roleHead, roleFounder}, Share: ptr(100.0)},
				},
			},
			turnover: months(2025, 10, 2_000_000, stablePattern),
			egrul: &egrulRecord{
				INN: "7724890160", Name: "ООО «Мебельная фабрика «Дуб»»", OGRN: "1117746789016", State: stateActive,
				RegDate: "2011-02-07", RegAddress: "г. Москва, ул. Дубининская, 57",
				OKVED:    okved{Code: "31.09", Name: "Производство прочей мебели"},
				Head:     egrulPerson{FullName: "Козлов Виктор Петрович", INN: "772400112233"},
				Founders: []egrulFounder{{FullName: "Козлов Виктор Петрович", Share: 100}},
			},
			tb:     transparentBusiness{TaxRegime: "OSN", Employees: 48},
			bki:    loans(bkiLoan{Amount: 2_000_000, Balance: 600_000}),
			girbo:  &girboReport{Year: 2025, Revenue: 15_200_000, NetProfit: 900_000},
			courts: noCourts(),
		},
		// c-10, polyakov: молодая компания, рваные обороты, мелкие просрочки, налоговый долг, суды, массовый адрес —
		// стоп-причин нет, но балл низкий — отказ.
		{
			client: bankClient{
				INN: "7731901270", Name: "ООО «Полюс»", Type: "LLC", ClientSince: "2024-06-03",
				RegAddress: "г. Москва, ул. Кировоградская, 13А", ActualAddress: "г. Москва, ул. Кировоградская, 13А",
				TaxRegime: "USN_INCOME_MINUS_EXPENSES",
				Persons: []bankPerson{
					{PersonID: "p-101", FullName: "Поляков Роман Сергеевич", INN: "773111223344", Roles: []string{roleHead, roleFounder}, Share: ptr(100.0)},
				},
			},
			turnover: months(2025, 10, 700_000, unstablePattern),
			egrul: &egrulRecord{
				INN: "7731901270", Name: "ООО «Полюс»", OGRN: "1247700890127", State: stateActive,
				RegDate: "2024-05-16", RegAddress: "г. Москва, ул. Кировоградская, 13А",
				OKVED:    okved{Code: "46.73", Name: "Торговля оптовая лесоматериалами и строительными материалами"},
				Head:     egrulPerson{FullName: "Поляков Роман Сергеевич", INN: "773111223344"},
				Founders: []egrulFounder{{FullName: "Поляков Роман Сергеевич", Share: 100}},
			},
			tb:    transparentBusiness{MassAddress: true, TaxRegime: "USN_INCOME_MINUS_EXPENSES", Employees: 3},
			bki:   loans(bkiLoan{Amount: 1_500_000, Balance: 1_200_000, OverdueDays: 12}, bkiLoan{Amount: 900_000, Balance: 700_000}),
			fns:   fnsDebts{HasDebt: true, Amount: 85_000},
			girbo: &girboReport{Year: 2025, Revenue: 7_900_000, NetProfit: -350_000},
			courts: courts{
				Arbitration: []arbitrationCase{
					{CaseNumber: "А40-97310/2026", Role: "DEFENDANT", Amount: 340_000},
					{CaseNumber: "А40-150288/2026", Role: "DEFENDANT", Amount: 95_000},
				},
				Enforcement: []enforcement{{Number: "51907/26/77028-ИП", Amount: 60_000}},
			},
		},
		// c-11, kuznetsov: ИП без поручителя и ГИР БО, всё чисто — одобрение.
		{
			client: bankClient{
				INN: "773498765432", Name: "ИП Кузнецов Андрей Павлович", Type: "IE", ClientSince: "2018-06-01",
				RegAddress: "г. Москва, ул. Генерала Белова, 28, кв. 40", ActualAddress: "г. Москва, ул. Генерала Белова, 28, кв. 40",
				TaxRegime: "USN_INCOME",
				Persons: []bankPerson{
					{PersonID: "p-111", FullName: "Кузнецов Андрей Павлович", INN: "773498765432", Roles: []string{roleHead}},
				},
			},
			turnover: months(2025, 10, 900_000, stablePattern),
			egrip: &egripRecord{
				INN: "773498765432", OGRNIP: "318774600234567", FullName: "Кузнецов Андрей Павлович", State: stateActive,
				RegDate: "2018-05-14", OKVED: okved{Code: "49.41", Name: "Деятельность автомобильного грузового транспорта"},
			},
			tb:     transparentBusiness{TaxRegime: "USN_INCOME", Employees: 4},
			bki:    loans(bkiLoan{Amount: 800_000, Balance: 300_000}),
			courts: noCourts(),
		},
	}

	d := dataset{companies: make(map[string]*company, len(cs)), mchd: map[string]mchdRecord{
		mchdPetrova: {
			Valid: true, PrincipalINN: "7712345678", ValidUntil: "2027-03-01",
			Representative: egrulPerson{FullName: "Петрова Анна Сергеевна", INN: "771298765432"},
			Powers:         []string{powerCreditApplication, powerBKIConsent},
		},
		mchdVolkov: {
			Valid: false, PrincipalINN: "7704211530", ValidUntil: "2026-09-30",
			Representative: egrulPerson{FullName: "Волков Игорь Николаевич", INN: "770411223344"},
			Powers:         []string{powerCreditApplication, powerBKIConsent},
		},
	}}
	for _, c := range cs {
		c.accounts = []*account{newAccount(c, 1, accountActive, c.turnover.AvgMonthlyIncome*balanceMonths)}
		if c.client.INN == innSmirnov {
			c.accounts = append(c.accounts, newAccount(c, 2, accountBlocked, 0), newAccount(c, 3, accountActive, smallBalance))
		}
		d.companies[c.client.INN] = c
	}
	return d
}

// newAccount — n-й рублёвый счёт компании. Номер — балансовый счёт ООО (40702) или ИП (40802), код рубля 810 и
// последние 8 цифр ИНН с номером счёта: синтетический, но в формате 20 цифр.
func newAccount(c *company, n int, status string, balance int64) *account {
	prefix := "40702810"
	if c.client.Type == "IE" {
		prefix = "40802810"
	}
	return &account{
		ID: fmt.Sprintf("acc-%s-%d", c.client.INN, n), Number: fmt.Sprintf("%s%s%04d", prefix, c.client.INN[len(c.client.INN)-8:], n),
		Currency: currencyRUB, Status: status, Balance: balance,
	}
}
