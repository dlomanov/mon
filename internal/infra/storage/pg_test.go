package storage_test

import (
	"context"
	"github.com/dlomanov/mon/internal/infra/storage"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dlomanov/mon/internal/entities"
	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const (
	teardownTimeout = 10 * time.Second
	postgresStartup = 5 * time.Second
)

type (
	TestSuit struct {
		suite.Suite
		ctx      context.Context
		logger   *zap.Logger
		pgc      *postgres.PostgresContainer
		db       *sqlx.DB
		teardown func()
	}
)

func TestRun(t *testing.T) {
	suite.Run(t, new(TestSuit))
}

func (s *TestSuit) SetupSuite() {
	logger := zaptest.NewLogger(s.T(), zaptest.Level(zap.DebugLevel))
	s.logger = logger
	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.teardown = cancel

	dsn := "host=localhost port=5432 user=postgres password=1 dbname=sprint2 sslmode=disable"
	s.pgc, dsn = createPosgres(s.T(), dsn)
	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	require.NoError(s.T(), err)
	s.db = db
}

func (s *TestSuit) TearDownSuite() {
	timeout, cancel := context.WithTimeout(context.Background(), teardownTimeout)
	defer cancel()

	s.teardown()

	if err := s.db.Close(); err != nil {
		s.logger.Error("failed to close postgres db", zap.Error(err))
	}
	if err := s.pgc.Terminate(timeout); err != nil {
		s.logger.Error("failed to terminate postgres container", zap.Error(err))
	}
}

func (s *TestSuit) TestPGStorage() {
	db, err := storage.NewPGStorage(s.ctx, s.logger, s.db)
	require.NoError(s.T(), err)

	// Test Set
	key := entities.MetricsKey{Type: entities.MetricGauge, Name: "cpu_usage"}
	value := 0.5
	err = db.Set(s.ctx, entities.Metric{
		MetricsKey: key,
		Value:      &value,
	})
	require.NoError(s.T(), err)

	// Test Get
	metric1, ok, err := db.Get(s.ctx, key)
	require.NoError(s.T(), err, "failed to get metric")
	require.True(s.T(), ok, "failed to get metric")
	require.Equal(s.T(), value, *metric1.Value, "invalid metric value")
	require.Equal(s.T(), key, metric1.MetricsKey, "invalid metric key")

	// Test All
	metrics, err := db.All(s.ctx)
	require.NoError(s.T(), err, "failed to get all metrics")
	require.Len(s.T(), metrics, 1, "invalid metrics length")
}

func createPosgres(t *testing.T, dsn string) (*postgres.PostgresContainer, string) {
	values := strings.Split(dsn, " ")
	require.NotEmpty(t, values, "failed to parse database uri")
	kmap := make(map[int]string, len(values))
	vmap := make(map[string]string, len(values))
	for i, v := range values {
		kv := strings.Split(v, "=")
		require.Len(t, kv, 2, "failed to parse database uri value")
		kmap[i] = kv[0]
		vmap[kv[0]] = kv[1]
	}
	port, ok := vmap["port"]
	require.True(t, ok, "failed to get database port")
	username, ok := vmap["user"]
	require.True(t, ok, "failed to get database user")
	password, ok := vmap["password"]
	require.True(t, ok, "failed to get database password")
	dbname, ok := vmap["dbname"]
	require.True(t, ok, "failed to get database name")

	ctx := context.Background()
	pgc, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:latest"),
		postgres.WithDatabase(dbname),
		postgres.WithUsername(username),
		postgres.WithPassword(password),
		postgres.WithInitScripts(),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(postgresStartup),
		),
	)
	require.NoError(t, err, "failed to start postgres container")

	newHost, err := pgc.Host(ctx)
	require.NoError(t, err, "failed to get postgres container host")
	newPort, err := pgc.MappedPort(ctx, nat.Port(port))
	require.NoError(t, err, "failed to get postgres container port")

	var sb strings.Builder
	for i := 0; i < len(values); i++ {
		k := kmap[i]

		_, _ = sb.WriteString(k)
		_ = sb.WriteByte('=')
		switch {
		case k == "host":
			_, _ = sb.WriteString(newHost)
		case k == "port":
			_, _ = sb.WriteString(strconv.Itoa(newPort.Int()))
		default:
			_, _ = sb.WriteString(vmap[k])
		}
		_ = sb.WriteByte(' ')
	}

	dsn = strings.TrimRight(sb.String(), " ")
	return pgc, dsn
}
