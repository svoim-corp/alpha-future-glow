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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/svoim-corp/alpha-future-glow/backend/docs"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/catalog"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/company"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/identity"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/integrations"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/notifications"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/offer"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/origination"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/planner"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/config"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/events"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/idempotency"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/module"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/scoring"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/servicing"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/support"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/underwriting"
)

// @title			Умный кредитный кабинет API
// @version		1.0
// @description	REST API кредитного кабинета для юрлиц и ИП и рабочего места банка.
// @BasePath		/api/v1
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

	deps := module.Deps{DB: conn, Bus: events.New(), Idem: idempotency.New(conn)}
	if err := idempotency.Migrate(ctx, conn); err != nil {
		return err
	}

	// Зоны. Владельцы — в README.md; каждая зона меняет только свои строки ниже.
	companyMod := company.New(deps)
	identityMod := identity.New(deps)
	catalogMod := catalog.New(deps)
	originationMod := origination.New(deps)
	integrationsMod := integrations.New(deps, cfg.MocksBaseURL)
	scoringMod := scoring.New(deps)
	underwritingMod := underwriting.New(deps)
	offerMod := offer.New(deps)
	servicingMod := servicing.New(deps)
	supportMod := support.New(deps)
	notificationsMod := notifications.New(deps)
	plannerMod := planner.New(deps)

	// Связи между зонами (shared-интерфейсы). Пока у поставщика нет реализации, он отдаёт заглушку из своего ports.go.
	companyMod.Wire(company.Ports{
		BankData:     integrationsMod.BankData(),
		Registries:   integrationsMod.Registries(),
		Applications: originationMod.Applications(),
	})
	identityMod.Wire(identity.Ports{Companies: companyMod.Companies()})
	originationMod.Wire(origination.Ports{
		Products:    catalogMod.Products(),
		Companies:   companyMod.Companies(),
		Authorities: companyMod.Authorities(),
		Prelims:     scoringMod.Prelims(),
	})
	scoringMod.Wire(scoring.Ports{
		Applications: originationMod.Applications(),
		Products:     catalogMod.Products(),
		Companies:    companyMod.Companies(),
		Checks:       integrationsMod.Checks(),
	})
	underwritingMod.Wire(underwriting.Ports{
		Applications: originationMod.Applications(),
		Companies:    companyMod.Companies(),
		Products:     catalogMod.Products(),
		Notifier:     notificationsMod.Notifier(),
	})
	offerMod.Wire(offer.Ports{
		Applications: originationMod.Applications(),
		Decisions:    scoringMod.Decisions(),
		Companies:    companyMod.Companies(),
	})
	servicingMod.Wire(servicing.Ports{Notifier: notificationsMod.Notifier()})
	plannerMod.Wire(planner.Ports{
		Products: catalogMod.Products(),
		Notifier: notificationsMod.Notifier(),
	})

	zones := []zone{
		companyMod, identityMod, catalogMod, originationMod, integrationsMod, scoringMod,
		underwritingMod, offerMod, servicingMod, supportMod, notificationsMod, plannerMod,
	}
	for _, z := range zones {
		if err := z.Migrate(ctx); err != nil {
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
	for _, z := range zones {
		z.Register(routes)
	}

	registerTech(r, cfg, conn, zones)

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

// zone — то, что cmd/api требует от каждой доменной зоны.
type zone interface {
	Migrate(ctx context.Context) error
	Register(r httpx.Routes)
}

// demoResetter — зона, у которой есть демо-данные для POST /admin/seed/reset.
// Реализуйте ResetDemo в зоне, если она хранит то, что меняется в ходе демо (заявки, решения, кредиты…).
type demoResetter interface {
	ResetDemo(ctx context.Context) error
}

// registerTech — технические эндпоинты вне /api/v1 (API_SPEC.md, раздел 4.13).
func registerTech(r *gin.Engine, cfg config.Config, conn *sqlx.DB, zones []zone) {
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

	r.GET("/api/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(docs.SwaggerInfo.ReadDoc()))
	})
	r.GET("/api/docs/*any", func(c *gin.Context) {
		if p := c.Param("any"); p == "" || p == "/" {
			c.Redirect(http.StatusFound, "/api/docs/index.html")
			return
		}
		ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/api/openapi.json"))(c)
	})

	if cfg.IsDemo() {
		r.POST("/admin/seed/reset", func(c *gin.Context) {
			for _, z := range zones {
				if d, ok := z.(demoResetter); ok {
					if err := d.ResetDemo(c.Request.Context()); err != nil {
						httpx.Fail(c, err)
						return
					}
				}
			}
			c.Status(http.StatusNoContent)
		})
	}
}
