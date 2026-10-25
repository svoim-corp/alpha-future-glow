// Команда support-index — подготовка локальной базы знаний; запускается после миграций API.
package main

import (
	"context"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/config"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/support"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		slog.Error("knowledge indexing failed", "err", err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	conn, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer conn.Close()
	m := support.New(module.Deps{DB: conn})
	m.ConfigureAI(support.AIConfig{URL: cfg.OllamaURL, Model: cfg.OllamaModel, EmbeddingModel: cfg.EmbeddingModel})
	if err := m.Migrate(ctx); err != nil {
		return err
	}
	if err := m.IndexKnowledge(ctx); err != nil {
		return err
	}
	manifest, err := m.KnowledgeManifest(ctx)
	if err != nil {
		return err
	}
	slog.Info("knowledge index ready", "manifest", string(manifest))
	return nil
}
