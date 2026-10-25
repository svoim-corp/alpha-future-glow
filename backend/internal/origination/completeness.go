package origination

import (
	"slices"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// missingField — чего не хватает для отправки; field — путь поля в GET /applications/{id}, его подсвечивает мастер.
type missingField struct {
	Field   string
	Message string
}

// completeness — проверка полноты перед отправкой (docs/origination.md, validate). Данные банка должны быть
// загружены: без них не проверить поручителя. truthConfirmed — подтверждение достоверности: при отправке оно
// приходит в запросе, а до этого его ещё нет.
func completeness(d *dossier, truthConfirmed bool) []missingField {
	a, st := d.app, d.stage
	var out []missingField
	add := func(field, msg string) { out = append(out, missingField{field, msg}) }

	switch {
	case st.invalidated:
		add("prelim", "Параметры изменились — пройдите предварительную проверку заново")
	case st.prelim == nil:
		add("prelim", "Пройдите предварительную проверку")
	case st.prelim.Outcome == shared.PrelimStop:
		add("prelim", "По результату предварительной проверки заявку подать нельзя")
	case st.stage != stageFull:
		add("prelim.choice", "Выберите: уменьшить сумму до лимита или оставить запрошенную")
	}
	if d.consent == nil {
		add("consents.bki", "Дайте согласие на запрос в БКИ")
	}
	if d.guarantorRequired() {
		g := a.GuarantorPersonID
		switch {
		case g == nil || !slices.ContainsFunc(d.profile.GuarantorCandidates, func(p shared.Person) bool { return p.PersonID == *g }):
			add("stage2.guarantor.personId", "Выберите поручителя")
		case !d.guarantorAgreed():
			add("stage2.guarantor.bkiConsent", "Нужно согласие поручителя на запрос в БКИ")
		}
	}
	if !truthConfirmed {
		add("stage2.truthConfirmed", "Подтвердите достоверность сведений")
	}
	return out
}

// guarantorRequired — поручитель нужен ООО, если его требует продукт. У ИП поручителя нет.
func (d *dossier) guarantorRequired() bool {
	return d.profile.Type == shared.CompanyLLC && d.product != nil && d.product.GuarantorForLLC
}
