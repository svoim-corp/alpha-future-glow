package support

import (
	"time"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

// AIConfig — только локальный inference, без облачного fallback.
type AIConfig struct{ URL, Model, EmbeddingModel string }

type conversation struct {
	ID            string         `json:"id" db:"id"`
	UserID        string         `json:"-" db:"user_id"`
	CompanyID     string         `json:"-" db:"company_id"`
	ApplicationID *string        `json:"applicationId" db:"application_id"`
	CreatedAt     time.Time      `json:"createdAt" db:"created_at"`
	Title         string         `json:"title" db:"title"`
	Viewer        *viewerContext `json:"-" db:"-"`
}

type source struct {
	ID      string `json:"id" db:"id"`
	Title   string `json:"title" db:"title"`
	Href    string `json:"href" db:"href"`
	Content string `json:"excerpt" db:"content"`
}

type actionParams struct {
	ProductName string                      `json:"productName,omitempty"`
	GoalLabel   string                      `json:"goalLabel,omitempty"`
	Kind        string                      `json:"kind" binding:"required,oneof=draft ticket plan"`
	Draft       *shared.AssistantDraft      `json:"draft,omitempty"`
	Subject     string                      `json:"subject,omitempty" binding:"max=200"`
	Text        string                      `json:"text,omitempty" binding:"max=4000"`
	PlanChange  *shared.AssistantPlanChange `json:"planChange,omitempty"`
	PlanPreview *shared.AssistantPlan       `json:"planPreview,omitempty"`
}

type applicationCard struct {
	ID          string  `json:"id"`
	Number      *string `json:"number"`
	ProductCode string  `json:"productCode"`
	Amount      int64   `json:"amount"`
	Term        int     `json:"term"`
	Status      string  `json:"status"`
	NextStep    string  `json:"nextStep"`
	CanDelete   bool    `json:"canDelete"`
}

type productCard struct {
	Code      string   `json:"code"`
	Name      string   `json:"name"`
	Goals     []string `json:"goals"`
	AmountMin int64    `json:"amountMin"`
	AmountMax int64    `json:"amountMax"`
	TermMin   int      `json:"termMin"`
	TermMax   int      `json:"termMax"`
	RateMin   float64  `json:"rateMin"`
	RateMax   float64  `json:"rateMax"`
}

type block struct {
	Kind        string                  `json:"kind"`
	Title       string                  `json:"title"`
	Product     *productCard            `json:"product,omitempty"`
	Calculation *shared.PaymentEstimate `json:"calculation,omitempty"`
	Application *applicationCard        `json:"application,omitempty"`
	Payment     *shared.UpcomingPayment `json:"payment,omitempty"`
	ActionID    string                  `json:"actionId,omitempty"`
	Action      *actionParams           `json:"action,omitempty"`
	ResultID    string                  `json:"resultId,omitempty"`
	Navigation  *navigationCard         `json:"navigation,omitempty"`
	Plan        *shared.AssistantPlan   `json:"plan,omitempty"`
	Loan        *shared.AssistantLoan   `json:"loan,omitempty"`
	Company     *companyCard            `json:"company,omitempty"`
	Ticket      *ticketCard             `json:"ticket,omitempty"`
}

type navigationCard struct {
	Section string `json:"section"`
	Label   string `json:"label"`
	Href    string `json:"href"`
}
type companyCard struct {
	Name           string `json:"name"`
	INN            string `json:"inn"`
	RegAddress     string `json:"regAddress"`
	ActualAddress  string `json:"actualAddress"`
	MonthlyIncome  int64  `json:"monthlyIncome"`
	ViewerName     string `json:"viewerName,omitempty"`
	ViewerPosition string `json:"viewerPosition,omitempty"`
}
type ticketCard struct {
	ID      string  `json:"id"`
	Subject string  `json:"subject"`
	Text    string  `json:"text"`
	Status  string  `json:"status"`
	Reply   *string `json:"reply,omitempty"`
}

type messagePayload struct {
	Sources []source        `json:"sources"`
	Blocks  []block         `json:"blocks"`
	Finance *financeContext `json:"finance,omitempty"`
	Viewer  *viewerContext  `json:"viewer,omitempty"`
}

type viewerContext struct {
	Name     string `json:"name"`
	Position string `json:"position"`
}

type chatMessage struct {
	ID             string                  `json:"id" db:"id"`
	ConversationID string                  `json:"-" db:"conversation_id"`
	Role           string                  `json:"role" db:"role"`
	Status         string                  `json:"status" db:"status"`
	Text           string                  `json:"text" db:"text"`
	Payload        db.JSON[messagePayload] `json:"-" db:"payload"`
	CreatedAt      time.Time               `json:"createdAt" db:"created_at"`
}

type messageDTO struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	Text      string    `json:"text"`
	Sources   []source  `json:"sources"`
	Blocks    []block   `json:"blocks"`
	CreatedAt time.Time `json:"createdAt"`
}

func (m chatMessage) dto() messageDTO {
	p := m.Payload.V
	if p.Sources == nil {
		p.Sources = []source{}
	}
	if p.Blocks == nil {
		p.Blocks = []block{}
	}
	return messageDTO{m.ID, m.Role, m.Status, m.Text, p.Sources, p.Blocks, m.CreatedAt}
}
