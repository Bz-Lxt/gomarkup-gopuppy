package repo

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(ctx context.Context, pool *pgxpool.Pool, files fs.FS) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock(88442201)"); err != nil {
		return err
	}
	defer func() { _, _ = conn.Exec(ctx, "SELECT pg_advisory_unlock(88442201)") }()

	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		entries, err = fs.ReadDir(files, "migrations")
		if err != nil {
			return err
		}
	}
	names := make([]string, 0, len(entries))
	root := "."
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
		if e.IsDir() && e.Name() == "migrations" {
			root = "migrations"
			sub, err := fs.ReadDir(files, "migrations")
			if err != nil {
				return err
			}
			names = names[:0]
			for _, s := range sub {
				if !s.IsDir() && strings.HasSuffix(s.Name(), ".sql") {
					names = append(names, s.Name())
				}
			}
		}
	}
	sort.Strings(names)
	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		return err
	}
	for _, name := range names {
		var exists bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		path := name
		if root != "." {
			path = root + "/" + name
		}
		sqlb, err := fs.ReadFile(files, path)
		if err != nil {
			return err
		}
		if _, err := conn.Exec(ctx, string(sqlb)); err != nil {
			return fmt.Errorf("migrate %s: %w", name, err)
		}
		if _, err := conn.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES ($1)`, name); err != nil {
			return err
		}
	}
	return nil
}
