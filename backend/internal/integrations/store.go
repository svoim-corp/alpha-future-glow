package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

type Store struct{ db *sqlx.DB }

type checkRow struct {
	Source      string                  `db:"source"`
	Status      string                  `db:"status"`
	RequestedAt time.Time               `db:"requested_at"`
	Summary     string                  `db:"summary"`
	Details     db.JSON[map[string]any] `db:"details"`
	Payload     *string                 `db:"payload"` // разобранный ответ источника; NULL — данных нет
}

// Save пишет снимки одного запуска проверок вместе: экран разбора не должен видеть половину запуска.
func (s *Store) Save(ctx context.Context, appID string, rep *shared.CheckReport) error {
	return db.WithTx(ctx, s.db, func(tx *sqlx.Tx) error {
		for _, r := range rep.Results {
			payload, err := encodeFact(factOf(rep.Facts, r.Source))
			if err != nil {
				return err
			}
			if r.Details == nil {
				r.Details = map[string]any{}
			}
			_, err = tx.ExecContext(ctx, `
				INSERT INTO integrations.check_results (app_id, source, status, requested_at, summary, details, payload)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				appID, r.Source, r.Status, r.RequestedAt, r.Summary, db.JSON[map[string]any]{V: r.Details}, payload)
			if err != nil {
				return fmt.Errorf("insert check %s: %w", r.Source, err)
			}
		}
		return nil
	})
}

// Latest — последний снимок по каждому источнику в порядке shared.AllSources.
func (s *Store) Latest(ctx context.Context, appID string) (*shared.CheckReport, error) {
	var rows []checkRow
	err := s.db.SelectContext(ctx, &rows, `
		SELECT DISTINCT ON (source) source, status, requested_at, summary, details, payload
		FROM integrations.check_results
		WHERE app_id = $1
		ORDER BY source, requested_at DESC, id DESC`, appID)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(rows, func(a, b checkRow) int {
		return slices.Index(shared.AllSources, shared.CheckSource(a.Source)) - slices.Index(shared.AllSources, shared.CheckSource(b.Source))
	})

	rep := &shared.CheckReport{Results: make([]shared.CheckResult, 0, len(rows))}
	for _, r := range rows {
		src := shared.CheckSource(r.Source)
		rep.Results = append(rep.Results, shared.CheckResult{
			Source: src, Status: shared.CheckStatus(r.Status), RequestedAt: r.RequestedAt.UTC(), Summary: r.Summary, Details: r.Details.V,
		})
		if r.Payload == nil {
			continue
		}
		f, err := decodeFact(src, []byte(*r.Payload))
		if err != nil {
			return nil, fmt.Errorf("decode check %s of %s: %w", src, appID, err)
		}
		setFact(&rep.Facts, f)
	}
	return rep, nil
}

func (s *Store) Reset(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `TRUNCATE integrations.check_results`)
	return err
}

func orNil[T any](p *T) any {
	if p == nil {
		return nil
	}
	return p
}

func factOf(fs shared.CheckFacts, s shared.CheckSource) any {
	switch s {
	case shared.SourceBank:
		return orNil(fs.Client)
	case shared.SourceBankAccounts:
		return orNil(fs.Turnover)
	case shared.SourceEGRUL:
		return orNil(fs.Registration)
	case shared.SourceTransparentBusiness:
		return orNil(fs.Transparent)
	case shared.SourceBKI:
		return orNil(fs.BKI)
	case shared.SourceFNS:
		return orNil(fs.FNS)
	case shared.SourceGIRBO:
		return orNil(fs.GIRBO)
	case shared.SourceCourts:
		return orNil(fs.Courts)
	}
	return nil
}

func encodeFact(f any) (*string, error) {
	if f == nil {
		return nil, nil
	}
	b, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	s := string(b)
	return &s, nil
}

func decodeFact(s shared.CheckSource, raw []byte) (any, error) {
	var f any
	switch s {
	case shared.SourceBank:
		f = new(shared.BankClient)
	case shared.SourceBankAccounts:
		f = new(shared.TurnoverReport)
	case shared.SourceEGRUL:
		f = new(shared.Registration)
	case shared.SourceTransparentBusiness:
		f = new(shared.TransparentBusiness)
	case shared.SourceBKI:
		f = new(shared.BKIReport)
	case shared.SourceFNS:
		f = new(shared.TaxDebt)
	case shared.SourceGIRBO:
		f = new(shared.GIRBOReport)
	case shared.SourceCourts:
		f = new(shared.CourtRecords)
	default:
		return nil, fmt.Errorf("unknown check source %q", s)
	}
	if err := json.Unmarshal(raw, f); err != nil {
		return nil, err
	}
	return f, nil
}
