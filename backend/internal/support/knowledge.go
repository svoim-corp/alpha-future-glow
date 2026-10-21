package support

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"io/fs"
	"math"
	"strconv"
	"strings"
)

//go:embed knowledge/*.md
var knowledge embed.FS

type knowledgeChunk struct {
	ID, Title, Href, Content, Hash string
	Vector                         []float32
}

func splitKnowledge(name, text string) []knowledgeChunk {
	out := []knowledgeChunk{}
	for _, section := range strings.Split(text, "\n## ") {
		title, body, ok := strings.Cut(section, "\n")
		if !ok {
			continue
		}
		title = strings.TrimLeft(title, "# ")
		words := strings.Fields(body)
		for start := 0; start < len(words); start += 250 {
			end := start + 300
			if end > len(words) {
				end = len(words)
			}
			content := title + "\n" + strings.Join(words[start:end], " ")
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
			out = append(out, knowledgeChunk{ID: fmt.Sprintf("%s-%d", name, len(out)), Title: title, Href: "/support?tab=faq", Content: content, Hash: hash})
			if end == len(words) {
				break
			}
		}
	}
	return out
}

func vectorLiteral(v []float32) (string, error) {
	if len(v) != 1024 {
		return "", fmt.Errorf("embedding dimensions: %d", len(v))
	}
	out := make([]string, len(v))
	for i, n := range v {
		if math.IsNaN(float64(n)) || math.IsInf(float64(n), 0) {
			return "", fmt.Errorf("invalid embedding value")
		}
		out[i] = strconv.FormatFloat(float64(n), 'g', -1, 32)
	}
	return "[" + strings.Join(out, ",") + "]", nil
}

// IndexKnowledge обновляет корпус целиком после успешной подготовки embeddings.
// Требует установленного pgvector; обычные миграции Support от него не зависят.
func (m *Module) IndexKnowledge(ctx context.Context) error {
	if m.ai == nil || m.ai.cfg.URL == "" {
		return fmt.Errorf("OLLAMA_BASE_URL is required")
	}
	if _, err := m.deps.DB.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS vector`); err != nil {
		return err
	}
	chunks := []knowledgeChunk{}
	err := fs.WalkDir(knowledge, "knowledge", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := knowledge.ReadFile(path)
		if err != nil {
			return err
		}
		chunks = append(chunks, splitKnowledge(path, string(b))...)
		return nil
	})
	if err != nil {
		return err
	}
	faq, err := m.store.FAQ(ctx, "")
	if err != nil {
		return err
	}
	for _, f := range faq {
		chunks = append(chunks, knowledgeChunk{ID: "faq-" + f.ID, Title: f.Question, Href: "/support?tab=faq", Content: f.Question + "\n" + f.Answer, Hash: fmt.Sprintf("%x", sha256.Sum256([]byte(f.Question+f.Answer)))})
	}
	var previous []struct {
		ID        string `db:"id"`
		Hash      string `db:"hash"`
		Version   string `db:"version"`
		Embedding string `db:"embedding"`
	}
	if err := m.deps.DB.SelectContext(ctx, &previous, `SELECT id,hash,version,array_to_json(embedding)::text AS embedding FROM support.knowledge_chunks ORDER BY id`); err != nil {
		return err
	}
	cached := map[string]knowledgeChunk{}
	version := "1:" + m.ai.cfg.EmbeddingModel
	for _, p := range previous {
		if p.Version != version {
			continue
		}
		var vector []float32
		if err := json.Unmarshal([]byte(p.Embedding), &vector); err != nil {
			return err
		}
		cached[p.ID] = knowledgeChunk{Hash: p.Hash, Vector: vector}
	}
	for i := range chunks {
		if old, ok := cached[chunks[i].ID]; ok && old.Hash == chunks[i].Hash && len(old.Vector) == 1024 {
			chunks[i].Vector = old.Vector
			continue
		}
		v, err := m.ai.embed(ctx, chunks[i].Content)
		if err != nil {
			return err
		}
		chunks[i].Vector = v
	}
	return db.WithTx(ctx, m.deps.DB, func(tx *sqlx.Tx) error {
		// Индекс маленький; атомарная замена убирает удалённые и изменённые фрагменты.
		if _, err := tx.ExecContext(ctx, `DELETE FROM support.knowledge_chunks`); err != nil {
			return err
		}
		for _, c := range chunks {
			v, err := vectorLiteral(c.Vector)
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO support.knowledge_chunks(id,title,href,version,hash,content,embedding) VALUES($1,$2,$3,$4,$5,$6,($7::vector)::real[])`, c.ID, c.Title, c.Href, version, c.Hash, c.Content, v); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *Module) retrieve(ctx context.Context, question string, vector []float32) ([]source, error) {
	v, err := vectorLiteral(vector)
	if err != nil {
		return nil, err
	}
	out := []source{}
	err = m.deps.DB.SelectContext(ctx, &out, `
	 WITH dense AS (
	  SELECT id,row_number() OVER(ORDER BY embedding::vector <=> $1::vector,id) AS rank
	  FROM support.knowledge_chunks ORDER BY embedding::vector <=> $1::vector,id LIMIT 15
	 ), lexical AS (
	  SELECT id,row_number() OVER(ORDER BY ts_rank(search,plainto_tsquery('russian',$2)) DESC,id) AS rank
	  FROM support.knowledge_chunks WHERE search @@ plainto_tsquery('russian',$2)
	  ORDER BY ts_rank(search,plainto_tsquery('russian',$2)) DESC,id LIMIT 15
	 ), scores AS (
	  SELECT id,sum(score) score FROM (
	   SELECT id,1.0/(60+rank) score FROM dense UNION ALL SELECT id,1.0/(60+rank) score FROM lexical
	  ) ranked GROUP BY id
	 ) SELECT k.id,k.title,k.href,k.content FROM support.knowledge_chunks k JOIN scores s ON s.id=k.id
	 ORDER BY s.score DESC,k.id LIMIT 5`, v, question)
	return out, err
}

// KnowledgeManifest возвращает состав корпуса для диагностики без содержимого клиентских диалогов.
func (m *Module) KnowledgeManifest(ctx context.Context) ([]byte, error) {
	var rows []struct {
		ID      string `db:"id" json:"id"`
		Hash    string `db:"hash" json:"hash"`
		Version string `db:"version" json:"version"`
	}
	if err := m.deps.DB.SelectContext(ctx, &rows, `SELECT id,hash,version FROM support.knowledge_chunks ORDER BY id`); err != nil {
		return nil, err
	}
	return json.Marshal(rows)
}
