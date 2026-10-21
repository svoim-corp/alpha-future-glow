package company

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

type mchdRegistry struct {
	fakeRegistries
	recs map[string]*shared.MCHDRecord
	err  error
}

func (r *mchdRegistry) MCHD(_ context.Context, number string) (*shared.MCHDRecord, error) {
	return r.recs[number], r.err
}

func TestResolve(t *testing.T) {
	today := time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)
	date := func(s string) time.Time { d, _ := time.Parse("2006-01-02", s); return d }
	num := func(s string) *string { return &s }

	const company, petrova, volkov = "7712345678", "771298765432", "770411223344"
	representative := func(inn, mchd string) shared.BankPerson {
		return shared.BankPerson{PersonID: "p-" + inn, INN: inn, Roles: []shared.PersonRole{shared.PersonRepresentative}, MCHDNumber: num(mchd)}
	}
	mchd := func(principal, person, until string, valid bool, powers ...string) *shared.MCHDRecord {
		return &shared.MCHDRecord{Valid: valid, PrincipalINN: principal, Representative: shared.RegistryPerson{INN: person}, ValidUntil: date(until), Powers: powers}
	}
	all := []string{shared.PowerCreditApplication, shared.PowerBKIConsent}

	tests := []struct {
		name      string
		person    string
		persons   []shared.BankPerson
		recs      map[string]*shared.MCHDRecord
		bankErr   error
		want      shared.AuthorityType
		reason    string // начало причины
		upstream  bool
		mchdPower []string
	}{
		{name: "руководитель", person: head.INN, persons: []shared.BankPerson{head, rep}, want: shared.AuthorityHead, mchdPower: all},
		{name: "ИП", person: ieHead.INN, persons: []shared.BankPerson{ieHead}, want: shared.AuthorityHead, mchdPower: all},
		{name: "действующая МЧД (petrova)", person: petrova, persons: []shared.BankPerson{head, representative(petrova, "m-1")},
			recs: map[string]*shared.MCHDRecord{"m-1": mchd(company, petrova, "2027-03-01", true, all...)}, want: shared.AuthorityMCHD, mchdPower: all},
		{name: "МЧД действует в последний день", person: petrova, persons: []shared.BankPerson{representative(petrova, "m-1")},
			recs: map[string]*shared.MCHDRecord{"m-1": mchd(company, petrova, "2026-10-21", true, all...)}, want: shared.AuthorityMCHD, mchdPower: all},
		{name: "истёкшая МЧД (volkov)", person: volkov, persons: []shared.BankPerson{representative(volkov, "m-2")},
			recs: map[string]*shared.MCHDRecord{"m-2": mchd(company, volkov, "2026-09-30", false, all...)}, want: shared.AuthorityNone, reason: "Срок действия доверенности истёк 30.09.2026"},
		{name: "отозванная МЧД", person: petrova, persons: []shared.BankPerson{representative(petrova, "m-1")},
			recs: map[string]*shared.MCHDRecord{"m-1": mchd(company, petrova, "2027-03-01", false, all...)}, want: shared.AuthorityNone, reason: reasonMCHDRevoked},
		{name: "МЧД без права подавать заявку", person: petrova, persons: []shared.BankPerson{representative(petrova, "m-1")},
			recs: map[string]*shared.MCHDRecord{"m-1": mchd(company, petrova, "2027-03-01", true, shared.PowerBKIConsent)}, want: shared.AuthorityNone, reason: reasonMCHDNoPower},
		{name: "МЧД от другой компании", person: petrova, persons: []shared.BankPerson{representative(petrova, "m-1")},
			recs: map[string]*shared.MCHDRecord{"m-1": mchd("7704211530", petrova, "2027-03-01", true, all...)}, want: shared.AuthorityNone, reason: reasonMCHDOther},
		{name: "МЧД на другого человека", person: petrova, persons: []shared.BankPerson{representative(petrova, "m-1")},
			recs: map[string]*shared.MCHDRecord{"m-1": mchd(company, volkov, "2027-03-01", true, all...)}, want: shared.AuthorityNone, reason: reasonMCHDOther},
		{name: "МЧД нет в реестре", person: petrova, persons: []shared.BankPerson{representative(petrova, "m-404")}, want: shared.AuthorityNone, reason: reasonMCHDNotFound},
		{name: "только учредитель", person: founder.INN, persons: []shared.BankPerson{head, founder}, want: shared.AuthorityNone, reason: reasonNotRelated},
		{name: "человека нет в АБС", person: "000000000000", persons: []shared.BankPerson{head}, want: shared.AuthorityNone, reason: reasonNotRelated},
		{name: "АБС недоступна — ошибка, а не отказ", person: head.INN, bankErr: httpx.Upstream(""), upstream: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *shared.BankClient
			if tt.persons != nil {
				client = &shared.BankClient{Persons: tt.persons}
			}
			a := &authorities{bank: fakeBank{client: client, err: tt.bankErr}, reg: &mchdRegistry{recs: tt.recs}, now: func() time.Time { return today }}
			got, err := a.resolve(context.Background(), shared.CompanyView{ID: "c-1", INN: company}, tt.person)
			if tt.upstream {
				if !isUpstream(err) {
					t.Fatalf("err = %v", err)
				}
				return
			}
			if err != nil || got.Type != tt.want || !strings.HasPrefix(got.Reason, tt.reason) {
				t.Fatalf("got %+v, err %v; want %s %q", got, err, tt.want, tt.reason)
			}
			if tt.want != shared.AuthorityNone && (len(got.Powers) != len(tt.mchdPower) || got.Reason != "") {
				t.Fatalf("powers: %+v", got)
			}
			if tt.want == shared.AuthorityMCHD && (got.MCHDNumber == nil || *got.MCHDNumber != "m-1" || got.ValidUntil == nil) {
				t.Fatalf("mchd: %+v", got)
			}
		})
	}
}

func TestResolveRegistryUnavailable(t *testing.T) {
	p := rep
	p.MCHDNumber = new(string)
	a := &authorities{
		bank: fakeBank{client: &shared.BankClient{Persons: []shared.BankPerson{p}}},
		reg:  &mchdRegistry{err: httpx.Upstream("")},
		now:  time.Now,
	}
	if _, err := a.resolve(context.Background(), llc, p.INN); !isUpstream(err) {
		t.Fatalf("реестр МЧД недоступен — ожидается 502, получено %v", err)
	}
}

func TestResolveEmptyPerson(t *testing.T) {
	a := &authorities{}
	got, err := a.Resolve(context.Background(), "c-1", "")
	if err != nil || got.Type != shared.AuthorityNone {
		t.Fatalf("%+v %v", got, err)
	}
}
