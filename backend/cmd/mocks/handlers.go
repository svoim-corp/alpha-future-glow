package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// server отдаёт синтетические ответы внешних сервисов (API_SPEC.md, 4.14). Поведение задаётся ИНН из seed.
type server struct{ data dataset }

func (s *server) register(r gin.IRouter) {
	m := r.Group("/mock")
	m.GET("/core-banking/clients/:inn", s.bankClient)
	m.GET("/core-banking/turnover/:inn", s.turnover)
	m.GET("/egrul/:inn", s.egrul)
	m.GET("/egrip/:inn", s.egrip)
	m.GET("/transparent-business/:inn", s.transparentBusiness)
	m.POST("/bki/report", s.bkiReport)
	m.GET("/fns/debts/:inn", s.fnsDebts)
	m.GET("/girbo/:inn", s.girbo)
	m.GET("/courts/:inn", s.courts)
	m.GET("/mchd/:number", s.mchd)
}

func notFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": msg})
}

// company находит компанию по :inn или отвечает 404.
func (s *server) company(c *gin.Context) (*company, bool) {
	co, ok := s.data.companies[c.Param("inn")]
	if !ok {
		notFound(c, "Компания не найдена")
	}
	return co, ok
}

func (s *server) bankClient(c *gin.Context) {
	if co, ok := s.company(c); ok {
		c.JSON(http.StatusOK, co.client)
	}
}

func (s *server) turnover(c *gin.Context) {
	if co, ok := s.company(c); ok {
		c.JSON(http.StatusOK, co.turnover)
	}
}

func (s *server) egrul(c *gin.Context) {
	co, ok := s.company(c)
	if !ok {
		return
	}
	if co.egrul == nil {
		notFound(c, "Юрлицо не найдено в ЕГРЮЛ")
		return
	}
	c.JSON(http.StatusOK, co.egrul)
}

func (s *server) egrip(c *gin.Context) {
	co, ok := s.company(c)
	if !ok {
		return
	}
	if co.egrip == nil {
		notFound(c, "ИП не найден в ЕГРИП")
		return
	}
	c.JSON(http.StatusOK, co.egrip)
}

func (s *server) transparentBusiness(c *gin.Context) {
	if co, ok := s.company(c); ok {
		c.JSON(http.StatusOK, co.tb)
	}
}

type bkiRequest struct {
	INN       string `json:"inn" binding:"required"`
	ConsentID string `json:"consentId" binding:"required"`
}

// bkiReport — БКИ отдаёт отчёт только при согласии субъекта, поэтому consentId обязателен.
func (s *server) bkiReport(c *gin.Context) {
	var req bkiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": "Нужны inn и consentId"})
		return
	}
	co, ok := s.data.companies[req.INN]
	if !ok {
		notFound(c, "Субъект кредитной истории не найден")
		return
	}
	c.JSON(http.StatusOK, co.bki)
}

func (s *server) fnsDebts(c *gin.Context) {
	if co, ok := s.company(c); ok {
		c.JSON(http.StatusOK, co.fns)
	}
}

// girbo — бухгалтерская отчётность есть только у ООО.
func (s *server) girbo(c *gin.Context) {
	co, ok := s.company(c)
	if !ok {
		return
	}
	if co.girbo == nil {
		notFound(c, "Отчётность в ГИР БО не найдена")
		return
	}
	c.JSON(http.StatusOK, co.girbo)
}

func (s *server) courts(c *gin.Context) {
	if co, ok := s.company(c); ok {
		c.JSON(http.StatusOK, co.courts)
	}
}

func (s *server) mchd(c *gin.Context) {
	rec, ok := s.data.mchd[c.Param("number")]
	if !ok {
		notFound(c, "Доверенность не найдена в реестре")
		return
	}
	c.JSON(http.StatusOK, rec)
}
