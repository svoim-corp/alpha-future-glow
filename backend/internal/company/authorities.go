package company

import (
	"context"
	"slices"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Причины отказа в полномочиях — текст для 403 при создании черновика.
const (
	reasonNotRelated   = "По данным банка вы не можете действовать от имени компании. Обратитесь к руководителю или оформите МЧД"
	reasonMCHDNotFound = "Доверенность не найдена в реестре МЧД"
	reasonMCHDOther    = "Доверенность выдана не вам или не от имени этой компании"
	reasonMCHDExpired  = "Срок действия доверенности истёк "
	reasonMCHDRevoked  = "Доверенность отозвана или недействительна"
	reasonMCHDNoPower  = "Доверенность не даёт права подавать кредитную заявку"
	reasonDateLayout   = "02.01.2006"
)

// headPowers — руководитель действует без доверенности, ему доступно всё.
var headPowers = []string{shared.PowerCreditApplication, shared.PowerBKIConsent}

// authorities определяет полномочия заявителя по данным банка и реестру МЧД (API_SPEC.md, 4.14).
type authorities struct {
	store *Store
	bank  shared.BankData
	reg   shared.Registries
	now   func() time.Time
}

func none(reason string) shared.AuthorityView {
	return shared.AuthorityView{Type: shared.AuthorityNone, Reason: reason}
}

// Resolve: HEAD в АБС — руководитель; REPRESENTATIVE с МЧД — проверка доверенности в реестре; иначе полномочий нет.
// Недоступность АБС или реестра — ошибка (502), а не отказ: иначе клиент получил бы неверный 403.
func (a *authorities) Resolve(ctx context.Context, companyID, personINN string) (shared.AuthorityView, error) {
	if personINN == "" {
		return none(reasonNotRelated), nil
	}
	cv, err := a.store.GetByID(ctx, companyID)
	if err != nil {
		return shared.AuthorityView{}, err
	}
	if cv == nil {
		return none(reasonNotRelated), nil
	}
	return a.resolve(ctx, *cv, personINN)
}

func (a *authorities) resolve(ctx context.Context, cv shared.CompanyView, personINN string) (shared.AuthorityView, error) {
	bank, err := a.bank.BankClient(ctx, cv.INN)
	if err != nil {
		return shared.AuthorityView{}, err
	}
	if bank == nil {
		return none(reasonNotRelated), nil
	}
	i := slices.IndexFunc(bank.Persons, func(p shared.BankPerson) bool { return p.INN == personINN })
	if i < 0 {
		return none(reasonNotRelated), nil
	}
	p := bank.Persons[i]

	if slices.Contains(p.Roles, shared.PersonHead) {
		return shared.AuthorityView{Type: shared.AuthorityHead, Powers: slices.Clone(headPowers)}, nil
	}
	if !slices.Contains(p.Roles, shared.PersonRepresentative) || p.MCHDNumber == nil {
		return none(reasonNotRelated), nil
	}
	rec, err := a.reg.MCHD(ctx, *p.MCHDNumber)
	if err != nil {
		return shared.AuthorityView{}, err
	}
	return checkMCHD(rec, *p.MCHDNumber, cv.INN, personINN, a.now()), nil
}

// checkMCHD проверяет доверенность. Срок проверяется раньше признака valid: у истёкшей МЧД реестр тоже отдаёт
// valid = false, а клиенту полезнее узнать, что именно срок истёк.
func checkMCHD(rec *shared.MCHDRecord, number, principalINN, personINN string, now time.Time) shared.AuthorityView {
	switch {
	case rec == nil:
		return none(reasonMCHDNotFound)
	case rec.PrincipalINN != principalINN || rec.Representative.INN != personINN:
		return none(reasonMCHDOther)
	case !now.Before(rec.ValidUntil.AddDate(0, 0, 1)): // действует по дату окончания включительно
		return none(reasonMCHDExpired + rec.ValidUntil.Format(reasonDateLayout))
	case !rec.Valid:
		return none(reasonMCHDRevoked)
	case !slices.Contains(rec.Powers, shared.PowerCreditApplication):
		return none(reasonMCHDNoPower)
	}
	until := rec.ValidUntil
	return shared.AuthorityView{Type: shared.AuthorityMCHD, MCHDNumber: &number, ValidUntil: &until, Powers: slices.Clone(rec.Powers)}
}
