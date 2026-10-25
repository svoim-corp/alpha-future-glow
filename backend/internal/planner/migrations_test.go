package planner

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

func TestMigrateFromBothVersionTwoHistories(t *testing.T) {
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
	for _, history := range []string{"fresh", "assistant", "search"} {
		t.Run(history, func(t *testing.T) {
			// Изолированные схемы теста не затрагивают планы в planner.
			schema := "planner_upgrade_" + history
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
				file := &fstest.MapFile{Data: []byte(strings.ReplaceAll(string(body), "planner.", schema+"."))}
				current[path] = file
				if !strings.HasPrefix(entry.Name(), "0003_") {
					previous[path] = file
				}
			}
			if history == "search" {
				delete(previous, "migrations/0002_assistant_commands.sql")
				legacy := "-- +goose Up\n" +
					"CREATE INDEX plans_search_text_idx ON planner.plans USING gin (to_tsvector('russian', replace(lower(id || ' ' || title || ' ' || goal || ' ' || coalesce(note, '')), 'ё', 'е')));\n" +
					"-- +goose Down\nDROP INDEX planner.plans_search_text_idx;\n"
				previous["migrations/0002_search.sql"] = &fstest.MapFile{
					Data: []byte(strings.ReplaceAll(legacy, "planner.", schema+".")),
				}
			}
			if history != "fresh" {
				if err := db.Migrate(ctx, conn, schema, previous); err != nil {
					t.Fatal(err)
				}
			}
			if history == "assistant" {
				if _, err := conn.ExecContext(ctx, "INSERT INTO "+schema+".assistant_commands (user_id,command_id,fingerprint,result) VALUES ($1,$2,$3,$4)", "u-1", "preserved-command", "original", `{"id":"original-plan"}`); err != nil {
					t.Fatal(err)
				}
			}
			for attempt := 0; attempt < 2; attempt++ {
				if err := db.Migrate(ctx, conn, schema, current); err != nil {
					t.Fatal(err)
				}
			}
			for _, name := range []string{"plans", "assistant_commands", "plans_search_text_idx"} {
				var exists bool
				if err := conn.GetContext(ctx, &exists, "SELECT to_regclass($1) IS NOT NULL", schema+"."+name); err != nil || !exists {
					t.Fatalf("missing %s: %v", name, err)
				}
			}
			if history == "assistant" {
				var id string
				if err := conn.GetContext(ctx, &id, "SELECT result->>'id' FROM "+schema+".assistant_commands WHERE command_id = $1", "preserved-command"); err != nil || id != "original-plan" {
					t.Fatalf("command result lost: %q, %v", id, err)
				}
			}
		})
	}
}
