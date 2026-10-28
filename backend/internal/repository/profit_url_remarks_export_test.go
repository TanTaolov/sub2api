//go:build integration

package repository

import (
	"context"
	"database/sql"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

// Reuse the existing container-only, migrated private database fixture. HTTP
// tests commit real transactions; they must never reset shared suite tables.
func ProfitURLRemarksTestDatabase(t *testing.T) (*sql.DB, *dbent.Client) {
	t.Helper()
	return newImageUpgradeDatabase(t, context.Background())
}
