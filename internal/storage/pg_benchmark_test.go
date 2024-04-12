package storage_test

import (
	"context"
	"testing"

	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const dsn = "host=localhost port=5432 user=postgres password=1 dbname=sprint2 sslmode=disable"

func BenchmarkPGStorageGet(b *testing.B) {
	ctx := context.Background()
	logger := zaptest.NewLogger(b, zaptest.Level(zap.DebugLevel))
	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	require.NoError(b, err)
	defer func() { require.NoError(b, db.Close()) }()

	ps, err := storage.NewPGStorage(ctx, logger, db)
	require.NoError(b, err)

	key := entities.MetricsKey{
		Type: entities.MetricCounter,
		Name: "test_name",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := ps.Get(ctx, key)
		require.NoError(b, err)
	}
}
