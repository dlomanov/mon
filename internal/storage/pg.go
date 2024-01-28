package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dlomanov/mon/internal/apperrors"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func NewPGStorage(
	ctx context.Context,
	logger *zap.Logger,
	db *sqlx.DB,
) (*PGStorage, error) {
	ps := &PGStorage{
		logger:      logger,
		db:          db,
		migrationUp: false,
	}

	err := ps.migrate(ctx)
	return ps, err
}

type PGStorage struct {
	logger      *zap.Logger
	db          *sqlx.DB
	migrationUp bool
}

func (ps *PGStorage) Get(
	ctx context.Context,
	key entities.MetricsKey,
) (result entities.Metric, ok bool, err error) {
	m := metric{}

	err = ps.db.GetContext(ctx, &m,
		`select "name", "type", "delta", "value" from metrics where "name"= $1 and "type" = $2`,
		key.Name, string(key.Type))
	if errors.Is(err, sql.ErrNoRows) {
		return result, false, nil
	}
	if err != nil {
		return result, false, err
	}

	result, err = m.toEntity()
	if err != nil {
		return result, false, err
	}

	return result, true, nil
}

func (ps *PGStorage) All(ctx context.Context) (result []entities.Metric, err error) {
	var metrics []metric

	err = ps.db.SelectContext(ctx, &metrics, `select "name", "type", "delta", "value" from metrics`)
	if errors.Is(err, sql.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return result, err
	}

	result = make([]entities.Metric, 0, len(metrics))
	for _, v := range metrics {
		entity, err := v.toEntity()
		if err != nil {
			return result, err
		}

		result = append(result, entity)
	}

	return result, err
}

func (ps *PGStorage) Set(ctx context.Context, metrics ...entities.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	tx, err := ps.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		insert into metrics ("name", "type", "delta", "value") values ($1, $2, $3, $4)
		on conflict ("name", "type")
		    do update
		    	set "delta" = excluded."delta",
		    	    "value" = excluded."value";`)
	if err != nil {
		ps.logger.Error("metric upsert query preparing failed", zap.Error(err))
		return errors.Join(tx.Rollback(), err)
	}
	defer func(stmt *sql.Stmt) { _ = stmt.Close() }(stmt)

	for _, v := range metrics {
		_, err = stmt.ExecContext(ctx, v.Name, string(v.Type), v.Delta, v.Value)
		if err != nil {
			ps.logger.Error("metric upsert failed", zap.Error(err))
			return errors.Join(tx.Rollback(), err)
		}
	}

	if err = tx.Commit(); err != nil {
		ps.logger.Error("metric upsert commit failed", zap.Error(err))
		return err
	}

	return nil
}

func (ps *PGStorage) migrate(ctx context.Context) error {
	if ps.migrationUp {
		ps.logger.Debug("already migrated")
		return nil
	}
	ps.migrationUp = true

	_, err := ps.db.ExecContext(ctx, `
create table if not exists metrics (
    "name" text not null,
    "type" text not null,
    "delta" bigint,
    "value" double precision,
    primary key ("name", "type")
);
	`)

	if err != nil {
		ps.logger.Error("migration failed", zap.Error(err))
		return err
	}

	ps.logger.Debug("successfully migrated")
	return nil
}

type metric struct {
	Name  string          `db:"name"`
	Type  string          `db:"type"`
	Delta sql.NullInt64   `db:"delta"`
	Value sql.NullFloat64 `db:"value"`
}

func (m *metric) toEntity() (result entities.Metric, err error) {
	mtype, parsed := entities.ParseMetricType(m.Type)
	if !parsed {
		return result, apperrors.ErrUnsupportedMetricType.New(m.Type)
	}

	result.Name = m.Name
	result.Type = mtype
	if m.Delta.Valid {
		result.Delta = &m.Delta.Int64
	}
	if m.Value.Valid {
		result.Value = &m.Value.Float64
	}

	return result, nil
}
