package migrations

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Run applies embedded SQL migrations in filename order.
func Run(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := fs.ReadDir(Files, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	var filenames []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		filenames = append(filenames, entry.Name())
	}

	sort.Strings(filenames)

	for _, name := range filenames {
		query, err := Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if _, err := pool.Exec(ctx, string(query)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}

		log.Printf("applied migration %s", name)
	}

	return nil
}
