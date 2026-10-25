package support

import (
	"context"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/svoim-corp/alpha-future-glow/backend/internal/platform/db"
)

func TestMigrationVersionsUnique(t *testing.T) {
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	versions := make(map[int]string)
	for _, entry := range entries {
		version, err := strconv.Atoi(strings.SplitN(entry.Name(), "_", 2)[0])
		if err != nil {
			t.Fatal(err)
		}
		if previous, exists := versions[version]; exists {
			t.Fatalf("duplicate migration version %d: %s, %s", version, previous, entry.Name())
		}
		versions[version] = entry.Name()
	}
}

func TestMigrateFromBothVersionFourHistories(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задан")
	}
	ctx := context.Background()
	conn, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	for _, history := range []string{"fresh", "chat", "search"} {
		t.Run(history, func(t *testing.T) {
			// Собственные схемы теста: существующую support и её данные не трогаем.
			schema := "support_upgrade_" + history
			if _, err := conn.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if _, err := conn.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
					t.Error(err)
				}
			})
			current := fstest.MapFS{}
			previous := fstest.MapFS{}
			for _, entry := range entries {
				path := "migrations/" + entry.Name()
				body, err := migrations.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				file := &fstest.MapFile{Data: []byte(strings.ReplaceAll(string(body), "support.", schema+"."))}
				current[path] = file
				if !strings.HasPrefix(entry.Name(), "0005_") {
					previous[path] = file
				}
			}
			if history == "search" {
				delete(previous, "migrations/0004_chat.sql")
				// Точное содержимое исторической 0004_search.sql из main.
				legacy := "-- +goose Up\n" +
					"CREATE INDEX faq_search_text_idx ON support.faq USING gin (to_tsvector('russian', replace(lower(id || ' ' || question || ' ' || answer), 'ё', 'е')));\n" +
					"CREATE INDEX tickets_search_text_idx ON support.tickets USING gin (to_tsvector('russian', replace(lower(id || ' ' || subject || ' ' || text || ' ' || coalesce(reply, '')), 'ё', 'е')));\n" +
					"-- +goose Down\nDROP INDEX support.tickets_search_text_idx;\nDROP INDEX support.faq_search_text_idx;\n"
				previous["migrations/0004_search.sql"] = &fstest.MapFile{
					Data: []byte(strings.ReplaceAll(legacy, "support.", schema+".")),
				}
			}
			if history != "fresh" {
				if err := db.Migrate(ctx, conn, schema, previous); err != nil {
					t.Fatal(err)
				}
				if _, err := conn.ExecContext(ctx, "UPDATE "+schema+".faq SET answer = $1 WHERE id = $2", "Сохранённый ответ", "f-4"); err != nil {
					t.Fatal(err)
				}
			}
			if history == "chat" {
				if _, err := conn.ExecContext(ctx, "INSERT INTO "+schema+".conversations (id,user_id,company_id) VALUES ($1,$2,$3)", "preserved-chat", "u-1", "c-1"); err != nil {
					t.Fatal(err)
				}
			}
			for attempt := 0; attempt < 2; attempt++ {
				if err := db.Migrate(ctx, conn, schema, current); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range []string{"conversations", "messages", "actions", "knowledge_chunks", "faq_search_text_idx", "tickets_search_text_idx"} {
				var exists bool
				if err := conn.GetContext(ctx, &exists, "SELECT to_regclass($1) IS NOT NULL", schema+"."+name); err != nil || !exists {
					t.Fatalf("missing %s: %v", name, err)
				}
			}
			if history != "fresh" {
				var answer string
				if err := conn.GetContext(ctx, &answer, "SELECT answer FROM "+schema+".faq WHERE id = $1", "f-4"); err != nil || answer != "Сохранённый ответ" {
					t.Fatalf("FAQ data lost: %q, %v", answer, err)
				}
			}
			if history == "chat" {
				var count int
				if err := conn.GetContext(ctx, &count, "SELECT count(*) FROM "+schema+".conversations WHERE id = $1", "preserved-chat"); err != nil || count != 1 {
					t.Fatalf("chat data lost: %d, %v", count, err)
				}
			}
		})
	}
}
