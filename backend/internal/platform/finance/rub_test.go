package finance

import "testing"

func TestRub(t *testing.T) {
	for in, want := range map[int64]string{
		0: "0 ₽", 999: "999 ₽", 1000: "1 000 ₽", 300_000: "300 000 ₽", 10_000_000: "10 000 000 ₽", -62_200: "−62 200 ₽",
	} {
		if got := Rub(in); got != want {
			t.Errorf("Rub(%d) = %q, want %q", in, got, want)
		}
	}
}
