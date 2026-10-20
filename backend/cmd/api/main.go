// Команда api — REST API сервера /api/v1.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/catalog"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/company"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/identity"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/integrations"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/notifications"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/offer"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/origination"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/config"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/idempotency"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/scoring"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/servicing"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/support"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/underwriting"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)
	if err := run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if !cfg.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conn, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Шина событий: зоны подписываются на события друг друга через неё.
	// Пока подписчиков нет; передавайте bus в модули по мере надобности.
	_ = events.New()

	if err := idempotency.Migrate(ctx, conn); err != nil {
		return err
	}
	// Экземпляр idempotency.New(conn) передавайте в модули, которым нужен .Middleware() на роутах-командах.

	// Сборка модулей. Порядок важен для миграций: company -> identity -> остальные.
	companyMod, err := company.New(ctx, conn)
	if err != nil {
		return err
	}
	identityMod, err := identity.New(ctx, conn, companyMod.Companies())
	if err != nil {
		return err
	}

	// Заготовки остальных зон: подключайте зависимости (shared-интерфейсы, bus) в New по мере надобности.
	stubs := []interface{ Register(httpx.Routes) }{}
	add := func(m interface{ Register(httpx.Routes) }, err error) error {
		if err != nil {
			return err
		}
		stubs = append(stubs, m)
		return nil
	}
	for _, f := range []func() error{
		func() error { return add(catalog.New(ctx, conn)) },
		func() error { return add(origination.New(ctx, conn)) },
		func() error { return add(integrations.New(ctx, conn)) },
		func() error { return add(scoring.New(ctx, conn)) },
		func() error { return add(underwriting.New(ctx, conn)) },
		func() error { return add(offer.New(ctx, conn)) },
		func() error { return add(servicing.New(ctx, conn)) },
		func() error { return add(support.New(ctx, conn)) },
		func() error { return add(notifications.New(ctx, conn)) },
	} {
		if err := f(); err != nil {
			return err
		}
	}

	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(httpx.TraceID(), httpx.Logger(log), httpx.Recovery(log), auth.Middleware(identityMod))
	r.NoRoute(httpx.NoRoute)
	r.NoMethod(httpx.NoMethod)

	api := r.Group("/api/v1")
	routes := httpx.Routes{
		Public: api,
		Client: api.Group("", auth.RequireAuth()),
		Bank:   api.Group("/bank", auth.RequireAuth(), auth.RequireRole(shared.RoleBankOfficer)),
	}
	identityMod.Register(routes)
	companyMod.Register(routes)
	for _, m := range stubs {
		m.Register(routes)
	}

	registerTech(r, cfg, conn)

	srv := &http.Server{
		Addr: cfg.HTTPAddr, Handler: r,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.Info("api listening", "addr", cfg.HTTPAddr, "profile", cfg.Profile)
		errCh <- srv.ListenAndServe()
	}()
	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(sctx)
	}
	return nil
}

// registerTech — технические эндпоинты вне /api/v1 (API_SPEC.md, раздел 4.13).
func registerTech(r *gin.Engine, cfg config.Config, conn *sqlx.DB) {
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "UP"}) })

	mocksClient := &http.Client{Timeout: 2 * time.Second}
	r.GET("/ready", func(c *gin.Context) {
		dbStatus, mocksStatus := "UP", "UP"
		pctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := conn.PingContext(pctx); err != nil {
			dbStatus = "DOWN"
		}
		req, _ := http.NewRequestWithContext(pctx, http.MethodGet, cfg.MocksBaseURL+"/health", nil)
		if resp, err := mocksClient.Do(req); err != nil {
			mocksStatus = "DOWN"
		} else {
			if resp.StatusCode != http.StatusOK {
				mocksStatus = "DOWN"
			}
			resp.Body.Close()
		}
		status, code := "UP", http.StatusOK
		if dbStatus != "UP" || mocksStatus != "UP" {
			status, code = "DOWN", http.StatusServiceUnavailable
		}
		c.JSON(code, gin.H{"status": status, "db": dbStatus, "mocks": mocksStatus})
	})

	if cfg.IsDemo() {
		// TODO: каждая зона добавит свой сброс демо-данных.
		r.POST("/admin/seed/reset", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	}
}
