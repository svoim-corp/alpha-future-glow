package finance

import "time"

// PaymentDate — дата n-го ежемесячного платежа после выдачи from: то же число месяца, а если такого числа
// в месяце нет — последний день месяца. time.AddDate здесь не подходит: 31 января + 1 месяц у него — 3 марта.
func PaymentDate(from time.Time, n int) time.Time {
	y, m, d := from.UTC().Date()
	first := time.Date(y, m+time.Month(n), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(d, last)-1)
}
