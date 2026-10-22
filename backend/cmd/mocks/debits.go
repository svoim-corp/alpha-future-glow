package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Статусы поручения на списание.
const (
	debitProcessing = "PROCESSING"
	debitSucceeded  = "SUCCEEDED"
	debitFailed     = "FAILED"
)

// debit — поручение на списание со счёта клиента. Банк принимает его сразу (PROCESSING) и проводит при первом
// опросе статуса: так сервер видит асинхронное подтверждение, но демо не ждёт.
type debit struct {
	DebitID      string  `json:"debitId"`
	RequestID    string  `json:"requestId"`
	AccountID    string  `json:"accountId"`
	Status       string  `json:"status"`
	Amount       int64   `json:"amount"`
	PostedAmount *int64  `json:"postedAmount,omitempty"`
	PostedAt     *string `json:"postedAt,omitempty"`
	Reason       *string `json:"reason,omitempty"`

	acc *account
}

// ledger — состояние списаний в памяти: после перезапуска мока поручения забываются, сервер отправит их заново.
type ledger struct {
	mu        sync.Mutex
	seq       int
	debits    map[string]*debit // по debitId
	byRequest map[string]*debit // по requestId: повтор поручения не создаёт второе списание
}

func newLedger() *ledger {
	return &ledger{debits: map[string]*debit{}, byRequest: map[string]*debit{}}
}

// accounts — счета компании. Остатки меняются списаниями, поэтому чтение — под той же блокировкой.
func (s *server) accounts(c *gin.Context) {
	co, ok := s.company(c)
	if !ok {
		return
	}
	s.ledger.mu.Lock()
	defer s.ledger.mu.Unlock()
	out := make([]account, len(co.accounts))
	for i, a := range co.accounts {
		out[i] = *a
	}
	c.JSON(http.StatusOK, out)
}

type debitRequest struct {
	RequestID string `json:"requestId" binding:"required"`
	INN       string `json:"inn" binding:"required"`
	AccountID string `json:"accountId" binding:"required"`
	Amount    int64  `json:"amount" binding:"required,gt=0"`
}

func (s *server) createDebit(c *gin.Context) {
	var req debitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "Нужны requestId, inn, accountId и amount > 0"})
		return
	}
	co, ok := s.data.companies[req.INN]
	if !ok {
		notFound(c, "Компания не найдена")
		return
	}
	var acc *account
	for _, a := range co.accounts {
		if a.ID == req.AccountID {
			acc = a
		}
	}
	if acc == nil {
		notFound(c, "Счёт не найден")
		return
	}

	l := s.ledger
	l.mu.Lock()
	defer l.mu.Unlock()
	d, ok := l.byRequest[req.RequestID]
	if !ok {
		l.seq++
		d = &debit{DebitID: fmt.Sprintf("dbt-%d", l.seq), RequestID: req.RequestID, AccountID: acc.ID, Status: debitProcessing, Amount: req.Amount, acc: acc}
		l.debits[d.DebitID], l.byRequest[req.RequestID] = d, d
	}
	c.JSON(http.StatusOK, gin.H{"debitId": d.DebitID, "status": d.Status})
}

// debitStatus проводит поручение при первом опросе: списывает со счёта или отказывает с причиной.
func (s *server) debitStatus(c *gin.Context) {
	l := s.ledger
	l.mu.Lock()
	defer l.mu.Unlock()
	d, ok := l.debits[c.Param("id")]
	if !ok {
		notFound(c, "Поручение не найдено")
		return
	}
	if d.Status == debitProcessing {
		settle(d, time.Now().UTC())
	}
	c.JSON(http.StatusOK, d)
}

func settle(d *debit, now time.Time) {
	var reason string
	switch {
	case d.acc.Status != accountActive:
		reason = "Счёт заблокирован"
	case d.acc.Currency != currencyRUB:
		reason = "Списание возможно только с рублёвого счёта"
	case d.acc.Balance < d.Amount:
		reason = "Недостаточно средств на счёте"
	}
	if reason != "" {
		d.Status, d.Reason = debitFailed, &reason
		return
	}
	d.acc.Balance -= d.Amount
	at := now.Format(time.RFC3339)
	d.Status, d.PostedAmount, d.PostedAt = debitSucceeded, &d.Amount, &at
}
