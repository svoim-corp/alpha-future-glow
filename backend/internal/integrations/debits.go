package integrations

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// Счета и списания АБС банка (API_SPEC.md, 4.14). Реализуют shared.Debits для Servicing.

const accountActive = "ACTIVE"

type accountDTO struct {
	ID       string `json:"id"`
	Number   string `json:"number"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
	Balance  *int64 `json:"balance"`
}

func (c *client) Accounts(ctx context.Context, inn string) ([]shared.BankAccount, error) {
	var d []accountDTO
	if ok, err := c.do(ctx, http.MethodGet, "/mock/core-banking/accounts/"+esc(inn), nil, nameBank, &d); !ok || err != nil {
		return []shared.BankAccount{}, err
	}
	out := make([]shared.BankAccount, len(d))
	for i, a := range d {
		out[i] = shared.BankAccount{ID: a.ID, Number: a.Number, Currency: a.Currency, Active: a.Status == accountActive, Balance: a.Balance}
	}
	return out, nil
}

type debitRequestDTO struct {
	RequestID string `json:"requestId"`
	INN       string `json:"inn"`
	AccountID string `json:"accountId"`
	Amount    int64  `json:"amount"`
}

type debitDTO struct {
	DebitID      string  `json:"debitId"`
	Status       string  `json:"status"`
	PostedAmount *int64  `json:"postedAmount"`
	PostedAt     *string `json:"postedAt"`
	Reason       *string `json:"reason"`
}

// Debit — 404 мока (нет компании или счёта) для сервера — сбой источника: счёт уже проверен по списку счетов.
func (c *client) Debit(ctx context.Context, req shared.DebitRequest) (string, error) {
	var d debitDTO
	body := debitRequestDTO{RequestID: req.RequestID, INN: req.INN, AccountID: req.AccountID, Amount: req.Amount}
	ok, err := c.do(ctx, http.MethodPost, "/mock/core-banking/debits", body, nameBank, &d)
	if err != nil {
		return "", err
	}
	if !ok || d.DebitID == "" {
		return "", upstream(nameBank, "/mock/core-banking/debits", fmt.Errorf("debit %s not accepted", req.RequestID))
	}
	return d.DebitID, nil
}

func (c *client) DebitStatus(ctx context.Context, externalID string) (*shared.DebitResult, error) {
	var d debitDTO
	path := "/mock/core-banking/debits/" + esc(externalID)
	if ok, err := c.do(ctx, http.MethodGet, path, nil, nameBank, &d); !ok || err != nil {
		return nil, err
	}
	invalid := httpx.Upstream("Сервис «" + nameBank + "» вернул некорректные данные")
	switch r := (shared.DebitResult{Status: shared.DebitStatus(d.Status)}); r.Status {
	case shared.DebitProcessing:
		return &r, nil
	case shared.DebitFailed:
		if d.Reason != nil {
			r.Reason = *d.Reason
		}
		return &r, nil
	case shared.DebitSucceeded:
		if d.PostedAmount == nil || d.PostedAt == nil {
			return nil, fmt.Errorf("debit %s: no posting: %w", externalID, invalid)
		}
		at, err := time.Parse(time.RFC3339, *d.PostedAt)
		if err != nil {
			return nil, fmt.Errorf("debit %s: posted at: %w", externalID, invalid)
		}
		r.PostedAmount, r.PostedAt = *d.PostedAmount, at.UTC()
		return &r, nil
	default:
		return nil, fmt.Errorf("debit %s: status %q: %w", externalID, d.Status, invalid)
	}
}
