package dumper

import (
	"database/sql"
	"errors"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/dlomanov/mon/internal/storage/internal/mem"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"sync"
	"time"
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

var delays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second, 0}

func (d *DBDumper) Load(dest *mem.Storage) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.logger.Debug("loading metrics")

	var (
		result mem.Storage
		err    error
	)
	for _, t := range delays {
		result, err = d.load()
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
			time.Sleep(t)
		} else {
			break
		}
	}
	if err != nil {
		d.logger.Debug("metrics loading  failed", zap.Error(err))
		return err
	}

	*dest = result
	d.logger.Debug("loaded metrics")
	return nil
}

func (d *DBDumper) load() (result mem.Storage, err error) {
	db := d.db

	if !d.migrationUp {
		if err = upMigration(db); err != nil {
			return result, err
		}
		d.migrationUp = true
		d.logger.Debug("migration up")
	}

	rows, err := db.Query(`select "name", "type", "delta", "value" from metrics;`)
	if err != nil {
		d.logger.Error("fail to query metrics", zap.Error(err))
		return result, err
	}
	defer func(rows *sql.Rows) { _ = rows.Close() }(rows)

	result = make(mem.Storage)

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
			return result, err
		}

		m := entities.Metric{
			MetricsKey: entities.MetricsKey{
				Name: name,
				Type: entities.ParseMetricTypeForced(mtype),
			},
		}
		if value.Valid {
			m.Value = &value.Float64
		}
		if delta.Valid {
			m.Delta = &delta.Int64
		}

		result[m.String()] = m
	}
	err = rows.Err()
	if err != nil {
		d.logger.Error("error occurred while reading rows", zap.Error(err))
		return result, err
	}

	return result, nil
}

func (d *DBDumper) Dump(source mem.Storage) error {
	if len(source) == 0 {
		d.logger.Debug("nothing to dump")
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.logger.Debug("dumping metrics")

	var err error
	for _, t := range delays {
		err = dump(d.logger, d.db, source)
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerrcode.IsConnectionException(pgerr.Code) {
			time.Sleep(t)
		} else {
			break
		}
	}

	if err != nil {
		d.logger.Error("metrics dumping failed", zap.Error(err))
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

func dump(
	logger *zap.Logger,
	db *sql.DB,
	source mem.Storage,
) error {
	tx, err := db.Begin()
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
		logger.Error("metric upsert query preparing failed", zap.Error(err))
		return errors.Join(tx.Rollback(), err)
	}
	defer func(stmt *sql.Stmt) { _ = stmt.Close() }(stmt)

	for _, v := range source {
		_, err = stmt.Exec(v.Name, string(v.Type), v.Delta, v.Value)
		if err != nil {
			logger.Error("metric upsert failed", zap.Error(err))
			return errors.Join(tx.Rollback(), err)
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Error("metric upsert commit failed", zap.Error(err))
		return err
	}

	return nil
}
