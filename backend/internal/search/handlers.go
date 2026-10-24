package search

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/auth"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/httpx"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/searchtext"
	"github.com/svoim-corp/alpha-future-glow/backend/internal/shared"
)

var kinds = []string{"applications", "loans", "plans", "products", "faq", "tickets"}

type itemDTO struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	State   string `json:"state"`
	Step    string `json:"step"`
}

type groupDTO struct {
	Type        string    `json:"type"`
	Items       []itemDTO `json:"items"`
	HasMore     bool      `json:"hasMore"`
	NextCursor  *string   `json:"nextCursor"`
	Unavailable bool      `json:"unavailable"`
}

type responseDTO struct {
	Query  string     `json:"query"`
	Groups []groupDTO `json:"groups"`
}

type cursor struct {
	Kind  string `json:"kind"`
	Hash  string `json:"hash"`
	Score int    `json:"score"`
	ID    string `json:"id"`
}

func fingerprint(scope shared.SearchScope, query string) string {
	sum := sha256.Sum256([]byte(scope.UserID + "\x00" + scope.CompanyID + "\x00" + query))
	return hex.EncodeToString(sum[:])
}

func parseKinds(value string) ([]string, error) {
	if value == "" {
		return append([]string(nil), kinds...), nil
	}
	out := make([]string, 0, len(kinds))
	seen := map[string]bool{}
	for _, kind := range strings.Split(value, ",") {
		valid := false
		for _, allowed := range kinds {
			valid = valid || kind == allowed
		}
		if !valid || seen[kind] {
			return nil, httpx.Validation("Неизвестная или повторяющаяся категория поиска")
		}
		seen[kind] = true
		out = append(out, kind)
	}
	return out, nil
}

// search godoc
//
// @Summary Общий поиск клиентского кабинета
// @Tags search
// @Produce json
// @Param q query string true "Запрос, 2–120 символов"
// @Param types query string false "Категории через запятую"
// @Param limitPerType query int false "Результатов на категорию, 1–20" default(3)
// @Param cursor query string false "Продолжение одной категории"
// @Success 200 {object} responseDTO
// @Failure 400 {object} httpx.ErrorBody
// @Failure 401 {object} httpx.ErrorBody
// @Failure 403 {object} httpx.ErrorBody
// @Failure 503 {object} httpx.ErrorBody
// @Router /search [get]
func (m *Module) search(c *gin.Context) {
	u := auth.UserFrom(c)
	if u.Role != shared.RoleClient || u.CompanyID == nil {
		httpx.Fail(c, httpx.Forbidden("Поиск доступен клиенту банка с компанией"))
		return
	}
	scope := shared.SearchScope{UserID: u.ID, CompanyID: *u.CompanyID}
	text := searchtext.Normalize(c.Query("q"))
	if strings.IndexFunc(text, unicode.IsControl) >= 0 {
		httpx.Fail(c, httpx.Validation("В поиске недопустимы управляющие символы"))
		return
	}
	if utf8.RuneCountInString(text) < 2 || utf8.RuneCountInString(text) > 120 {
		httpx.Fail(c, httpx.Validation("Введите от 2 до 120 символов для поиска"))
		return
	}
	types, err := parseKinds(c.Query("types"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	limit := 3
	if raw := c.Query("limitPerType"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 20 {
			httpx.Fail(c, httpx.Validation("Размер поисковой выдачи должен быть от 1 до 20"))
			return
		}
	}
	after := shared.SearchPosition{}
	if raw := c.Query("cursor"); raw != "" {
		var position cursor
		if len(raw) > 2048 {
			httpx.Fail(c, httpx.Validation("Некорректное продолжение поиска"))
			return
		}
		decoded, decodeErr := base64.RawURLEncoding.DecodeString(raw)
		if len(raw) > 2048 || len(types) != 1 || decodeErr != nil || json.Unmarshal(decoded, &position) != nil ||
			position.Kind != types[0] || position.Hash != fingerprint(scope, text) ||
			position.Score < 1 || position.Score > 4 || position.ID == "" || len(position.ID) > 128 || strings.IndexFunc(position.ID, func(char rune) bool {
			return !(char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '-' || char == '_')
		}) >= 0 {
			httpx.Fail(c, httpx.Validation("Продолжение поиска устарело или не соответствует запросу"))
			return
		}
		after = shared.SearchPosition{Score: position.Score, ID: position.ID}
	}
	groups, err := m.find(c.Request.Context(), scope, text, types, limit, after)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.Header("Cache-Control", "private, no-store")
	httpx.JSON(c, http.StatusOK, responseDTO{Query: text, Groups: groups})
}

func nextCursor(scope shared.SearchScope, text, kind string, hit shared.SearchHit) (*string, error) {
	body, err := json.Marshal(cursor{Kind: kind, Hash: fingerprint(scope, text), Score: hit.Score, ID: hit.ID})
	if err != nil {
		return nil, fmt.Errorf("encode search cursor: %w", err)
	}
	value := base64.RawURLEncoding.EncodeToString(body)
	return &value, nil
}
