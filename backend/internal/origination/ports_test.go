package origination

import (
	"context"
	"testing"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

func TestApplicationsPort(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	apps := e.m.Applications()
	id := e.draft(t, "petrova").ID

	v, err := apps.Application(ctx, id)
	if err != nil || v == nil {
		t.Fatalf("заявка: %v", err)
	}
	if v.CompanyID != "c-1" || v.ApplicantID != "u-2" || v.ProductCode != "WORKING" || v.Goal != "WORKING_CAPITAL" ||
		v.Amount != 1_500_000 || v.Term != 24 || v.Status != shared.StatusDraft || v.Stage != stagePrelim || v.Version != 1 ||
		v.Authority.Type != shared.AuthorityMCHD || v.Authority.MCHDNumber == nil || v.Number != nil || v.SubmittedAt != nil {
		t.Fatalf("представление: %+v", v)
	}
	if v, err := apps.Application(ctx, "a-0"); v != nil || err != nil {
		t.Fatalf("нет заявки — (nil, nil): %+v %v", v, err)
	}

	version, err := apps.SetAmount(ctx, id, 900_000, shared.Actor{Type: shared.ActorClient, UserID: "u-2"})
	if err != nil || version != 2 {
		t.Fatalf("SetAmount: %d %v", version, err)
	}
	if v, _ := apps.Application(ctx, id); v.Amount != 900_000 || v.Version != 2 {
		t.Fatalf("после SetAmount: %+v", v)
	}

	if c, err := apps.BKIConsent(ctx, id, shared.ConsentApplicant); c != nil || err != nil {
		t.Fatalf("согласия нет — (nil, nil): %+v %v", c, err)
	}
	if w := e.do("POST", "/api/v1/applications/"+e.draft(t, "smirnov").ID+"/consents/bki", "smirnov", consentBody); w.Code != 201 {
		t.Fatal(w.Body.String())
	}

	if err := apps.Transition(ctx, id, shared.StatusClosed, shared.Actor{Type: shared.ActorSystem}, "TEST", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := apps.SetAmount(ctx, id, 800_000, shared.Actor{}); code(err) != httpx.CodeInvalidTransition {
		t.Fatalf("SetAmount не черновика: %v", err)
	}
	if err := apps.Transition(ctx, id, shared.StatusSubmitted, shared.Actor{Type: shared.ActorSystem}, "TEST", ""); code(err) != httpx.CodeInvalidTransition {
		t.Fatalf("недопустимый переход: %v", err)
	}
}

func TestBKIConsentPort(t *testing.T) {
	e := newEnv(t)
	id := e.draft(t, "smirnov").ID
	if w := e.do("POST", "/api/v1/applications/"+id+"/consents/bki", "smirnov", consentBody); w.Code != 201 {
		t.Fatal(w.Body.String())
	}
	c, err := e.m.Applications().BKIConsent(context.Background(), id, shared.ConsentApplicant)
	if err != nil || c == nil || c.ID == "" || c.GivenBy != "u-1" || !c.GivenAt.Equal(today) {
		t.Fatalf("согласие: %+v %v", c, err)
	}
}
