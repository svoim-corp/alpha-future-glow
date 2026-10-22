package servicing

import (
	"bytes"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/finance"
)

// Шрифт с кириллицей встроен в сервер: PDF не зависит от шрифтов системы (DejaVu, свободная лицензия).
var (
	//go:embed fonts/DejaVuSansCondensed.ttf
	fontRegular []byte
	//go:embed fonts/DejaVuSansCondensed-Bold.ttf
	fontBold []byte
)

const (
	pdfFont      = "DejaVu"
	pdfMargin    = 10.0 // мм
	pdfRowHeight = 6.0
	humanTime    = "02.01.2006 15:04"
)

// Колонки таблицы графика: ширина в мм (A4 без полей — 190) и заголовок.
var pdfColumns = []struct {
	width float64
	title string
}{
	{10, "№"}, {21, "Дата"}, {26, "Осн. долг"}, {24, "Проценты"}, {26, "Платёж"}, {28, "Оплачено"}, {28, "Осталось"}, {27, "Статус"},
}

var statusLabels = map[string]string{paymentPlanned: "Ожидается", paymentPaid: "Оплачен", paymentOverdue: "Просрочен"}

var closeLabels = map[string]string{
	closeScheduleCompleted: "погашен по графику",
	closeEarlyRepayment:    "погашен досрочно",
	closeLimit:             "лимит закрыт клиентом",
}

// num — сумма в таблице: разряды через пробел, без знака рубля (единица — в заголовке документа).
func num(v int64) string { return strings.TrimSuffix(finance.Rub(v), " ₽") }

// schedulePDF — полный график кредита из снимка s на дату today (docs/servicing.md).
func schedulePDF(s *loanSnapshot, today time.Time) ([]byte, error) {
	l := s.loan
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	pdf.SetAutoPageBreak(true, pdfMargin+5)
	pdf.SetCreationDate(s.at)
	pdf.SetTitle("График платежей "+l.ContractNumber, true)
	pdf.AddUTF8FontFromBytes(pdfFont, "", fontRegular)
	pdf.AddUTF8FontFromBytes(pdfFont, "B", fontBold)
	pdf.AliasNbPages("")

	inTable := false
	tableHeader := func() {
		pdf.SetFont(pdfFont, "B", 8)
		pdf.SetFillColor(235, 238, 242)
		for _, col := range pdfColumns {
			pdf.CellFormat(col.width, pdfRowHeight, col.title, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetFont(pdfFont, "", 8)
	}
	// Шапка таблицы повторяется на каждой новой странице.
	pdf.SetHeaderFunc(func() {
		if inTable {
			tableHeader()
		}
	})
	pdf.SetFooterFunc(func() {
		pdf.SetY(-pdfMargin - 2)
		pdf.SetFont(pdfFont, "", 7)
		pdf.CellFormat(0, 4, fmt.Sprintf("%s · стр. %d из {nb}", l.ContractNumber, pdf.PageNo()), "", 0, "R", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetFont(pdfFont, "B", 13)
	pdf.MultiCell(0, 7, "График платежей по кредитному договору "+l.ContractNumber, "", "L", false)
	pdf.Ln(2)

	pdf.SetFont(pdfFont, "", 9)
	state := "действует"
	if l.Status == loanClosed && l.ClosedAt != nil && l.CloseReason != nil {
		state = "закрыт " + l.ClosedAt.UTC().Format(humanDate) + ", " + closeLabels[*l.CloseReason]
	}
	grace := "нет"
	if l.GraceMonths > 0 {
		grace = strconv.Itoa(l.GraceMonths) + " мес."
	}
	lines := [][2]string{
		{"Продукт", l.ProductName},
		{"Дата выдачи", l.IssuedAt.UTC().Format(humanDate)},
		{"Сумма кредита", finance.Rub(l.Amount)},
		{"Срок", strconv.Itoa(l.Term) + " мес."},
		{"Ставка", strings.Replace(strconv.FormatFloat(l.Rate, 'f', -1, 64), ".", ",", 1) + "% годовых"},
		{"Отсрочка", grace},
		{"Ежемесячный платёж", finance.Rub(l.MonthlyPayment)},
		{"Состояние кредита", state},
		{"Сформирован", s.at.Format(humanTime) + " UTC, операционная дата " + today.Format(humanDate)},
		{"Версии", fmt.Sprintf("кредита — %d, графика — %d; расчёт по политике %s", l.Version, l.ScheduleVersion, policyVersion)},
	}
	for _, kv := range lines {
		pdf.SetFont(pdfFont, "B", 9)
		pdf.CellFormat(45, 5, kv[0], "", 0, "L", false, 0, "")
		pdf.SetFont(pdfFont, "", 9)
		pdf.MultiCell(0, 5, kv[1], "", "L", false)
	}
	if early := earlyRepayments(s.ops); len(early) > 0 || l.ScheduleVersion > 1 {
		pdf.Ln(1)
		pdf.MultiCell(0, 5, "График пересчитан после досрочного погашения. Строки «Досрочно» — суммы, внесённые в дату "+
			"исполнения; заменённые будущие платежи не показаны и оплаченными не считаются."+early, "", "L", false)
	}
	pdf.Ln(3)

	inTable = true
	tableHeader()
	var planned, plannedPrincipal, plannedInterest, paid, remaining int64
	for _, p := range s.rows {
		status := statusLabels[paymentStatus(p, today)]
		if p.Early {
			status = "Досрочно"
		}
		cells := []string{
			strconv.Itoa(p.N), p.DueDate.Format(humanDate), num(p.Principal), num(p.Interest), num(p.Total),
			num(p.PaidPrincipal + p.PaidInterest), num(p.remaining()), status,
		}
		for i, col := range pdfColumns {
			align := "R"
			if i == 1 || i == len(pdfColumns)-1 {
				align = "C"
			}
			pdf.CellFormat(col.width, pdfRowHeight, cells[i], "1", 0, align, false, 0, "")
		}
		pdf.Ln(-1)
		planned, plannedPrincipal, plannedInterest = planned+p.Total, plannedPrincipal+p.Principal, plannedInterest+p.Interest
		paid, remaining = paid+p.PaidPrincipal+p.PaidInterest, remaining+p.remaining()
	}
	pdf.SetFont(pdfFont, "B", 8)
	totals := []string{"", "Итого", num(plannedPrincipal), num(plannedInterest), num(planned), num(paid), num(remaining), ""}
	for i, col := range pdfColumns {
		pdf.CellFormat(col.width, pdfRowHeight, totals[i], "1", 0, "R", false, 0, "")
	}
	pdf.Ln(-1)
	inTable = false

	pdf.Ln(3)
	pdf.SetFont(pdfFont, "", 8)
	pdf.MultiCell(0, 4, "Суммы — в рублях. «Оплачено» и «Осталось» — основной долг и проценты строки. Проценты будущих "+
		"платежей рассчитаны на весь срок и не равны сумме полного досрочного погашения: её даёт расчёт досрочного "+
		"погашения в кабинете. Пени в график не входят — полная задолженность показана в кабинете.", "", "L", false)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("render schedule pdf %s: %w", l.ID, err)
	}
	return buf.Bytes(), nil
}

// earlyRepayments — исполненные досрочные погашения для пометки о перерасчёте: « Досрочные погашения: 21.10.2026 — 200 000 ₽.»
func earlyRepayments(ops []opRow) string {
	var items []string
	for i := len(ops) - 1; i >= 0; i-- {
		op := ops[i]
		if op.Type == opEarly && op.Status == opSucceeded && op.PostedAt != nil && op.PostedAmount != nil {
			items = append(items, op.PostedAt.UTC().Format(humanDate)+" — "+finance.Rub(*op.PostedAmount))
		}
	}
	if len(items) == 0 {
		return ""
	}
	return " Досрочные погашения: " + strings.Join(items, "; ") + "."
}
