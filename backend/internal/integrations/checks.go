package integrations

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Пороги описательной оценки ответа. Это подсказка сотруднику, а не решение: стоп-причины и балл считает Scoring.
const (
	bkiWarnOverdueDays = 1
	bkiFailOverdueDays = 31
	stateActive        = "ACTIVE"
	courtDefendant     = "DEFENDANT"
)

// checker — реализация shared.Checks.
type checker struct {
	client *client
	store  *Store
	now    func() time.Time
}

func (ch *checker) Run(ctx context.Context, req shared.CheckRequest) (*shared.CheckReport, error) {
	rep := ch.collect(ctx, req)
	if err := ch.store.Save(ctx, req.AppID, rep); err != nil {
		return nil, fmt.Errorf("save checks %s: %w", req.AppID, err)
	}
	return rep, nil
}

func (ch *checker) Latest(ctx context.Context, appID string) (*shared.CheckReport, error) {
	return ch.store.Latest(ctx, appID)
}

// sources — какие источники опрашивать: запрошенные или все; ГИР БО — только у ООО.
func sources(req shared.CheckRequest) []shared.CheckSource {
	want := req.Sources
	if len(want) == 0 {
		want = shared.AllSources
	}
	out := make([]shared.CheckSource, 0, len(want))
	for _, s := range shared.AllSources {
		if !slices.Contains(want, s) || (s == shared.SourceGIRBO && req.CompanyType == shared.CompanyIE) {
			continue
		}
		out = append(out, s)
	}
	return out
}

// collect опрашивает источники параллельно: моки независимы, а конвейер и этап 1 ждут их все.
func (ch *checker) collect(ctx context.Context, req shared.CheckRequest) *shared.CheckReport {
	srcs := sources(req)
	results := make([]shared.CheckResult, len(srcs))
	facts := make([]any, len(srcs))
	var wg sync.WaitGroup
	for i, s := range srcs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			at := ch.now().UTC()
			if s == shared.SourceBKI && req.ConsentID == "" {
				results[i] = shared.CheckResult{Source: s, Status: shared.CheckUnavailable, RequestedAt: at, Summary: "Нет согласия на запрос в БКИ"}
				return
			}
			f, err := ch.fetch(ctx, req, s)
			if err != nil {
				results[i] = shared.CheckResult{Source: s, Status: shared.CheckUnavailable, RequestedAt: at, Summary: "Источник не ответил, повторите проверку"}
				return
			}
			results[i] = assess(s, f, at)
			facts[i] = f
		}()
	}
	wg.Wait()

	rep := &shared.CheckReport{Results: results}
	for _, f := range facts {
		setFact(&rep.Facts, f)
	}
	return rep
}

// fact убирает типизированный nil, чтобы «данных нет» было обычным nil в any.
func fact[T any](v *T, err error) (any, error) {
	if err != nil || v == nil {
		return nil, err
	}
	return v, nil
}

func (ch *checker) fetch(ctx context.Context, req shared.CheckRequest, s shared.CheckSource) (any, error) {
	switch s {
	case shared.SourceBank:
		return fact(ch.client.BankClient(ctx, req.INN))
	case shared.SourceBankAccounts:
		return fact(ch.client.Turnover(ctx, req.INN))
	case shared.SourceEGRUL:
		return fact(ch.client.Registration(ctx, req.INN, req.CompanyType))
	case shared.SourceTransparentBusiness:
		return fact(ch.client.transparent(ctx, req.INN))
	case shared.SourceBKI:
		return fact(ch.client.bki(ctx, req.INN, req.ConsentID))
	case shared.SourceFNS:
		return fact(ch.client.fns(ctx, req.INN))
	case shared.SourceGIRBO:
		return fact(ch.client.girbo(ctx, req.INN))
	case shared.SourceCourts:
		return fact(ch.client.courts(ctx, req.INN))
	}
	return nil, fmt.Errorf("unknown check source %q", s)
}

func setFact(fs *shared.CheckFacts, f any) {
	switch v := f.(type) {
	case *shared.BankClient:
		fs.Client = v
	case *shared.TurnoverReport:
		fs.Turnover = v
	case *shared.Registration:
		fs.Registration = v
	case *shared.TransparentBusiness:
		fs.Transparent = v
	case *shared.BKIReport:
		fs.BKI = v
	case *shared.TaxDebt:
		fs.FNS = v
	case *shared.GIRBOReport:
		fs.GIRBO = v
	case *shared.CourtRecords:
		fs.Courts = v
	}
}

// assess описывает ответ источника для экрана разбора (API_SPEC.md, 4.12: status, summary, details).
// f == nil — источник ответил, что данных нет.
func assess(s shared.CheckSource, f any, at time.Time) shared.CheckResult {
	r := shared.CheckResult{Source: s, Status: shared.CheckOK, RequestedAt: at}
	switch v := f.(type) {
	case *shared.BankClient:
		r.Summary = fmt.Sprintf("Клиент банка с %d года", v.ClientSince.Year())
		r.Details = map[string]any{"clientSince": v.ClientSince.Format(dateLayout)}
	case *shared.TurnoverReport:
		r.Summary = fmt.Sprintf("Обороты по счетам за %d мес., в среднем %s/мес.", len(v.Months), rub(v.AvgMonthlyIncome))
		r.Details = map[string]any{"avgMonthlyIncome": v.AvgMonthlyIncome, "periodMonths": len(v.Months)}
	case *shared.Registration:
		r.Details = map[string]any{"state": v.State, "regDate": v.RegDate.Format(dateLayout)}
		if v.State != stateActive {
			r.Status, r.Summary = shared.CheckFail, "Не действует: "+v.State
			break
		}
		r.Summary = "Действует, " + age(v.RegDate, at)
	case *shared.TransparentBusiness:
		r.Details = map[string]any{"employees": v.Employees, "taxRegime": v.TaxRegime}
		switch {
		case v.MassAddress:
			r.Status, r.Summary = shared.CheckWarn, "Массовый адрес регистрации"
		case v.MassHead:
			r.Status, r.Summary = shared.CheckWarn, "Массовый руководитель"
		default:
			r.Summary = "Нет признаков массового адреса"
		}
	case *shared.BKIReport:
		r.Details = map[string]any{"activeLoans": len(v.Loans), "totalDebt": v.TotalDebt, "overdueDays": v.MaxOverdueDays}
		switch {
		case v.MaxOverdueDays >= bkiFailOverdueDays:
			r.Status, r.Summary = shared.CheckFail, fmt.Sprintf("Просрочка %d дн.", v.MaxOverdueDays)
		case v.MaxOverdueDays >= bkiWarnOverdueDays:
			r.Status, r.Summary = shared.CheckWarn, fmt.Sprintf("Просрочка до 30 дней (%d дн.)", v.MaxOverdueDays)
		case len(v.Loans) == 0:
			r.Summary = "Кредитов нет"
		default:
			r.Summary = plural(len(v.Loans), "кредит", "кредита", "кредитов") + ", просрочек нет"
		}
	case *shared.TaxDebt:
		r.Details = map[string]any{"hasDebt": v.HasDebt, "amount": v.Amount}
		r.Summary = "Налоговой задолженности нет"
		if v.HasDebt {
			r.Status, r.Summary = shared.CheckWarn, "Налоговая задолженность "+rub(v.Amount)
		}
	case *shared.GIRBOReport:
		r.Summary = fmt.Sprintf("Отчётность за %d год учтена", v.Year)
		r.Details = map[string]any{"year": v.Year, "revenue": v.Revenue, "netProfit": v.NetProfit}
	case *shared.CourtRecords:
		assessCourts(&r, v)
	case nil:
		r.Status, r.Summary = notFound(s)
	}
	return r
}

func assessCourts(r *shared.CheckResult, v *shared.CourtRecords) {
	var sued int
	var claims int64
	for _, a := range v.Arbitration {
		if a.Role == courtDefendant {
			sued++
			claims += a.Amount
		}
	}
	r.Details = map[string]any{"arbitration": sued, "claims": claims, "enforcement": len(v.Enforcement), "bankruptcy": v.Bankruptcy}
	switch {
	case v.Bankruptcy:
		r.Status, r.Summary = shared.CheckFail, "Сведения о банкротстве в ЕФРСБ"
	case sued > 0 || len(v.Enforcement) > 0:
		r.Status = shared.CheckWarn
		r.Summary = plural(sued, "арбитражное дело", "арбитражных дела", "арбитражных дел") + ", " + rub(claims)
		if len(v.Enforcement) > 0 {
			r.Summary += "; исполнительных производств: " + strconv.Itoa(len(v.Enforcement))
		}
	default:
		r.Summary = "Судебных дел нет"
	}
}

func notFound(s shared.CheckSource) (shared.CheckStatus, string) {
	switch s {
	case shared.SourceBank:
		return shared.CheckFail, "Клиент не найден в АБС банка"
	case shared.SourceEGRUL:
		return shared.CheckFail, "Нет записи в реестре"
	case shared.SourceGIRBO:
		return shared.CheckWarn, "Отчётность в ГИР БО не найдена"
	}
	return shared.CheckWarn, "Источник не вернул данных"
}

// age — «11 лет» или «3 мес.» с даты регистрации.
func age(reg, now time.Time) string {
	months := (now.Year()-reg.Year())*12 + int(now.Month()) - int(reg.Month())
	if now.Day() < reg.Day() {
		months--
	}
	if months < 12 {
		return fmt.Sprintf("%d мес.", max(months, 0))
	}
	return plural(months/12, "год", "года", "лет")
}

// plural — «1 кредит», «2 кредита», «5 кредитов».
func plural(n int, one, few, many string) string {
	w := many
	switch n10, n100 := n%10, n%100; {
	case n10 == 1 && n100 != 11:
		w = one
	case n10 >= 2 && n10 <= 4 && (n100 < 12 || n100 > 14):
		w = few
	}
	return strconv.Itoa(n) + " " + w
}

// rub — «1 840 000 ₽».
func rub(v int64) string {
	s := strconv.FormatInt(v, 10)
	neg := v < 0
	if neg {
		s = s[1:]
	}
	var out []byte
	for i := range len(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, s[i])
	}
	if neg {
		return "-" + string(out) + " ₽"
	}
	return string(out) + " ₽"
}
