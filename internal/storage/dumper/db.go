package dumper

import (
	"database/sql"
	"errors"
	"github.com/dlomanov/mon/internal/apperrors"
	"github.com/dlomanov/mon/internal/entities"
	"go.uber.org/zap"
	"sync"
)

func init() {
	var _ Dumper = (*DBDumper)(nil)
}

func NewDBDumper(logger *zap.Logger, db *sql.DB) *DBDumper {
	return &DBDumper{
		logger:      logger,
		db:          db,
		mu:          sync.Mutex{},
		migrationUp: false,
	}
}

type DBDumper struct {
	logger      *zap.Logger
	db          *sql.DB
	mu          sync.Mutex
	migrationUp bool
}

func (d *DBDumper) Load(dest *map[string]string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	db := d.db

	d.logger.Debug("loading metrics")

	if !d.migrationUp {
		if err := upMigration(db); err != nil {
			return err
		}
		d.migrationUp = true
		d.logger.Debug("migration up")
	}

	rows, err := db.Query(`select "name", "type", "delta", "value" from metrics;`)
	if err != nil {
		d.logger.Error("fail to query metrics", zap.Error(err))
		return err
	}
	defer func(rows *sql.Rows) { _ = rows.Close() }(rows)

	result := make(map[string]string)

	for rows.Next() {
		var (
			name  string
			mtype string
			delta sql.NullInt64
			value sql.NullFloat64
		)

		err = rows.Scan(&name, &mtype, &delta, &value)
		if err != nil {
			d.logger.Error("fail to scan metrics", zap.Error(err))
			return err
		}

		t, ok := entities.ParseMetricType(mtype)
		if !ok {
			d.logger.Error("fail to parse metric type",
				zap.Error(err),
				zap.String("metric_type", mtype))
			return apperrors.ErrUnsupportedMetricType.New(mtype)
		}

		m := entities.Metric{MetricsKey: entities.MetricsKey{Name: name, Type: t}}
		if value.Valid {
			m.Value = &value.Float64
		}
		if delta.Valid {
			m.Delta = &delta.Int64
		}

		result[m.String()] = m.StringValue()
	}
	err = rows.Err()
	if err != nil {
		d.logger.Error("error occurred while reading rows", zap.Error(err))
		return err
	}

	*dest = result
	d.logger.Debug("loaded metrics")
	return nil
}

func (d *DBDumper) Dump(source map[string]string) error {
	if len(source) == 0 {
		d.logger.Debug("nothing to dump")
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.logger.Debug("dumping metrics")

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
		insert into metrics ("name", "type", "delta", "value") values ($1, $2, $3, $4)
		on conflict ("name", "type")
		    do update
		    	set "delta" = excluded."delta",
		    	    "value" = excluded."value";`)
	if err != nil {
		d.logger.Error("metric upsert query preparing failed", zap.Error(err))
		return errors.Join(tx.Rollback(), err)
	}
	defer func(stmt *sql.Stmt) { _ = stmt.Close() }(stmt)

	for k, v := range source {
		mk, err := entities.NewMetricsKey(k)
		if err != nil {
			d.logger.Error("metric key parsing failed", zap.Error(err))
			return errors.Join(tx.Rollback(), err)
		}

		m, err := entities.NewMetric(mk, v)
		if err != nil {
			d.logger.Error("metric value parsing failed", zap.Error(err))
			return errors.Join(tx.Rollback(), err)
		}

		_, err = stmt.Exec(m.Name, string(m.Type), m.Delta, m.Value)
		if err != nil {
			d.logger.Error("metric upsert failed", zap.Error(err))
			return errors.Join(tx.Rollback(), err)
		}
	}

	if err = tx.Commit(); err != nil {
		d.logger.Error("metric upsert commit failed", zap.Error(err))
		return err
	}

	d.logger.Debug("metrics dumped")
	return nil
}

func upMigration(db *sql.DB) error {
	_, err := db.Exec(`
create table if not exists metrics (
    "name" text not null,
    "type" text not null,
    "delta" bigint,
    "value" double precision,
    primary key ("name", "type")
);
	`)
	return err
}
