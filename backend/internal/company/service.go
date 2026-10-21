package company

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// profileTTL — сколько профиль живёт в кэше. Данные банка меняются редко, а клиент открывает заявку часто.
const profileTTL = time.Hour

// profiles собирает профиль клиента из АБС и реестров (API_SPEC.md, company-profile).
type profiles struct {
	store *Store
	bank  shared.BankData
	reg   shared.Registries
	now   func() time.Time
}

// Profile — профиль компании; (nil, nil) — компании нет.
func (p *profiles) Profile(ctx context.Context, companyID string) (*shared.CompanyProfile, error) {
	cv, err := p.store.GetByID(ctx, companyID)
	if err != nil || cv == nil {
		return nil, err
	}
	cached, at, err := p.store.CachedProfile(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("load cached profile %s: %w", companyID, err)
	}
	if cached != nil && p.now().Sub(at) < profileTTL {
		return cached, nil
	}

	prof, complete, err := p.fetch(ctx, *cv)
	if err != nil {
		return nil, err
	}
	// Неполный профиль (реестр не ответил) не кэшируем, чтобы следующий запрос его дособрал.
	if complete {
		if err := p.store.SaveProfile(ctx, prof, p.now().UTC()); err != nil {
			return nil, fmt.Errorf("save profile %s: %w", companyID, err)
		}
	}
	return prof, nil
}

// fetch опрашивает источники. АБС обязательна: без неё профиля нет (502). Реестры только сверяют данные банка,
// поэтому их недоступность не ломает профиль — он возвращается без них и с complete = false.
func (p *profiles) fetch(ctx context.Context, cv shared.CompanyView) (prof *shared.CompanyProfile, complete bool, err error) {
	bank, err := p.bank.BankClient(ctx, cv.INN)
	if err != nil {
		return nil, false, err
	}
	if bank == nil {
		return nil, false, fmt.Errorf("company %s (inn %s) not found in core banking", cv.ID, cv.INN)
	}
	turnover, err := p.bank.Turnover(ctx, cv.INN)
	if err != nil {
		return nil, false, err
	}

	complete = true
	reg, err := p.reg.Registration(ctx, cv.INN, cv.Type)
	if err != nil {
		if !isUpstream(err) {
			return nil, false, err
		}
		reg, complete = nil, false
	}
	var girbo *shared.GIRBOReport
	if cv.Type == shared.CompanyLLC {
		if girbo, err = p.reg.GIRBO(ctx, cv.INN); err != nil {
			if !isUpstream(err) {
				return nil, false, err
			}
			girbo, complete = nil, false
		}
	}
	return buildProfile(cv, bank, turnover, reg, girbo), complete, nil
}

func isUpstream(err error) bool {
	var e *httpx.Error
	return errors.As(err, &e) && e.Code == httpx.CodeUpstreamUnavail
}

// buildProfile сводит данные банка и реестров. Адреса и налоговый режим — из АБС: банк главный источник (Уточн.).
func buildProfile(cv shared.CompanyView, bank *shared.BankClient, turnover *shared.TurnoverReport, reg *shared.Registration, girbo *shared.GIRBOReport) *shared.CompanyProfile {
	p := &shared.CompanyProfile{
		CompanyView:         cv,
		ClientSince:         bank.ClientSince,
		RegAddress:          bank.RegAddress,
		ActualAddress:       bank.ActualAddress,
		TaxRegime:           bank.TaxRegime,
		Founders:            []shared.Founder{},
		GuarantorCandidates: []shared.Person{},
		Sources:             []shared.CheckSource{shared.SourceBank},
		GIRBOAccounted:      girbo != nil,
	}
	if reg != nil {
		p.OGRN, p.State, p.RegDate, p.OKVED = reg.OGRN, reg.State, reg.RegDate, reg.OKVED
		p.Sources = append(p.Sources, shared.SourceEGRUL)
	}
	if turnover != nil {
		p.Turnover = shared.Turnover{AvgMonthlyIncome: turnover.AvgMonthlyIncome, PeriodMonths: len(turnover.Months)}
	}

	for _, bp := range bank.Persons {
		person := shared.Person{PersonID: bp.PersonID, FullName: bp.FullName, INN: bp.INN, Roles: bp.Roles}
		isHead, isFounder := slices.Contains(bp.Roles, shared.PersonHead), slices.Contains(bp.Roles, shared.PersonFounder)
		if isHead && p.Head.PersonID == "" {
			p.Head = person
		}
		if isFounder {
			p.Founders = append(p.Founders, shared.Founder{Person: person, Share: bp.Share})
		}
		// Поручитель по ООО — руководитель или учредитель-физлицо; у ИП поручителя нет (API_SPEC.md, 4.5).
		if cv.Type == shared.CompanyLLC && (isHead || isFounder) {
			candidate := person
			candidate.Roles = slices.DeleteFunc(slices.Clone(bp.Roles), func(r shared.PersonRole) bool {
				return r != shared.PersonHead && r != shared.PersonFounder
			})
			p.GuarantorCandidates = append(p.GuarantorCandidates, candidate)
		}
	}
	return p
}
