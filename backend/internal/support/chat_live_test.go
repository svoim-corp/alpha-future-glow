package support

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"
)

// Запускается явно на локальном Docker/Ollama; обычный набор тестов не требует моделей.
func TestLivePlanExtraction(t *testing.T) {
	if os.Getenv("SUPPORT_LIVE_MODEL") != "1" {
		t.Skip("локальная модель не запрошена")
	}
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		result, err := exec.CommandContext(r.Context(), "docker", "exec", "credit-cabinet-support-backend-1", "wget", "-qO-", "--header=Content-Type:application/json", "--post-data="+string(raw), "http://ollama:11434"+r.URL.Path).Output()
		if err != nil {
			t.Error(err)
			w.WriteHeader(502)
			return
		}
		_, _ = w.Write(result)
	}))
	defer proxy.Close()
	m := &Module{}
	m.ConfigureAI(AIConfig{URL: proxy.URL, Model: "qwen3:8b-q4_K_M"})
	for _, q := range []string{"Дай мне кредит, хочу на 150к рублей и на 12 месяца. Хочу пару телефонов взять", "Создай план «Новая печь»: купить печь для пекарни, бюджет 500 тысяч рублей, 10 ноября 2026 года", "В плане p-42 измени бюджет на 600 тысяч рублей", "Рассчитай платёж по экспресс-кредиту: 150к на год", "Покажи мои кредиты"} {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		plan, err := m.ai.plan(ctx, []llmMessage{{Role: "user", Content: q}})
		cancel()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s => tools=%v params=%+v", q, plan.Tools, plan.Params)
		if plan.Params == nil {
			t.Error("нет параметров")
			continue
		}
		if len(plan.Tools) == 0 {
			t.Error("нет намерения")
			continue
		}
		if plan.Tools[0] == "propose_plan" && plan.Params.Plan == nil {
			t.Error("нет параметров плана")
		}
		if plan.Tools[0] == "calculate" && plan.Params.ProductCode != "EXPRESS" {
			t.Error("потерян выбранный продукт")
		}
		if q == "Покажи мои кредиты" && (len(plan.Tools) != 1 || plan.Tools[0] != "loans") {
			t.Error("список кредитов подменён другой функцией", plan.Tools)
		}
	}
}
