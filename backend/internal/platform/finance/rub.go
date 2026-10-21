package finance

import "strconv"

// Rub — сумма для текста клиенту: «1 500 000 ₽», разряды через пробел.
func Rub(n int64) string {
	s := strconv.FormatInt(n, 10)
	sign := ""
	if n < 0 {
		sign, s = "−", s[1:]
	}
	out := make([]byte, 0, len(s)+len(s)/3)
	for i := range len(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, s[i])
	}
	return sign + string(out) + " ₽"
}
