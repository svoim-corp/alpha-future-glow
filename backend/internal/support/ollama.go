package support

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type ollama struct {
	cfg  AIConfig
	http *http.Client
}
type llmMessage struct {
	Role     string     `json:"role"`
	Content  string     `json:"content"`
	ToolName string     `json:"tool_name,omitempty"`
	Calls    []toolCall `json:"tool_calls,omitempty"`
}
type toolCall struct {
	Function toolFunction `json:"function"`
}
type toolFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}
type toolDefinition struct {
	Type     string   `json:"type"`
	Function toolSpec `json:"function"`
}
type toolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// ConfigureAI подключает локальный Ollama. Пустой URL оставляет FAQ и обращения доступными.
func (m *Module) ConfigureAI(cfg AIConfig) {
	m.ai = &ollama{cfg: cfg, http: &http.Client{Timeout: 120 * time.Second}}
}

func (o *ollama) post(ctx context.Context, path string, body, out any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(o.cfg.URL, "/")+path, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.http.Do(req)
	if err != nil {
		return fmt.Errorf("local inference: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("local inference status %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(out)
}

func (o *ollama) embed(ctx context.Context, text string) ([]float32, error) {
	var result struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := o.post(ctx, "/api/embed", map[string]any{"model": o.cfg.EmbeddingModel, "input": text, "truncate": false, "keep_alive": "10m"}, &result); err != nil {
		return nil, err
	}
	if len(result.Embeddings) != 1 || len(result.Embeddings[0]) != 1024 {
		return nil, fmt.Errorf("invalid embedding dimensions")
	}
	return result.Embeddings[0], nil
}

func (o *ollama) chat(ctx context.Context, messages []llmMessage, tools []toolDefinition) (llmMessage, error) {
	var result struct {
		Message       llmMessage `json:"message"`
		TotalDuration int64      `json:"total_duration"`
		LoadDuration  int64      `json:"load_duration"`
		EvalCount     int        `json:"eval_count"`
		EvalDuration  int64      `json:"eval_duration"`
	}
	err := o.post(ctx, "/api/chat", map[string]any{"model": o.cfg.Model, "messages": messages, "tools": tools, "stream": false, "think": false, "keep_alive": "10m", "options": map[string]any{"num_ctx": 8192, "num_predict": 1000, "temperature": 0.2}}, &result)
	if err == nil {
		slog.Info("local generation completed", "total_ms", result.TotalDuration/int64(time.Millisecond), "load_ms", result.LoadDuration/int64(time.Millisecond), "eval_tokens", result.EvalCount, "eval_ms", result.EvalDuration/int64(time.Millisecond))
	}
	return result.Message, err
}
