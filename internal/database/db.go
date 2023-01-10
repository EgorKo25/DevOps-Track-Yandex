package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/EgorKo25/DevOps-Track-Yandex/internal/storage"
	"log"
	"time"

	"github.com/EgorKo25/DevOps-Track-Yandex/internal/configuration"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	DB  *sql.DB
	ctx context.Context
	cfg *config.ConfigurationServer
	str *storage.MetricStorage

	Buffer []storage.Metric
}

func NewDB(cfg *config.ConfigurationServer, ctx context.Context, str *storage.MetricStorage) *DB {
	if cfg.DB == "" {
		return nil
	}

	conn, err := sql.Open("pgx", cfg.DB)
	if err != nil {
		log.Println(err)
	}

	return &DB{
		DB:     conn,
		ctx:    ctx,
		cfg:    cfg,
		str:    str,
		Buffer: make([]storage.Metric, 0, 29),
	}
}

func (d *DB) CreateTable() {
	ctx, cancel := context.WithTimeout(d.ctx, 3*time.Second)
	defer cancel()

	query := "CREATE TABLE IF NOT EXISTS metrics (id VARCHAR(30), type VARCHAR(10), hash VARCHAR(100), value DOUBLE PRECISION, delta BIGINT);"

	r, err := d.DB.ExecContext(ctx, query)
	if err != nil {
		log.Println("Field to create db table ", err, r)
	}
}
func (d *DB) Close() error {
	return d.DB.Close()
}

// Flush TODO: go dock
func (d *DB) Flush() (err error) {

	ctx, cancel := context.WithTimeout(d.ctx, 3*time.Second)
	defer cancel()

	if d.DB == nil {
		return errors.New("you haven`t opened the database connection")
	}

	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	var stmt *sql.Stmt
	defer stmt.Close()

	for _, v := range d.Buffer {

		switch v.MType {
		case "gauge":
			query := `INSERT INTO metrics 
    				(id, type, hash, value) 
					VALUES ($1, $2, $3, $4)`

			stmt, err = tx.PrepareContext(ctx, query)
			if err != nil {
				return err
			}

			if _, err = stmt.ExecContext(ctx, v.ID, v.MType, v.Hash, float64(*v.Value)); err != nil {
				if err = tx.Rollback(); err != nil {
					log.Println("update drivers: unable to rollback: ", err)
				}
				return err
			}

		case "counter":
			query := `INSERT INTO metrics 
    				(id, type, hash, delta) 
					VALUES ($1, $2, $3, $4)`

			stmt, err = tx.PrepareContext(ctx, query)
			if err != nil {
				return err
			}

			if _, err = stmt.ExecContext(ctx, v.ID, v.MType, v.Hash, int(*v.Delta)); err != nil {
				if err = tx.Rollback(); err != nil {
					log.Println("update drivers: unable to rollback: ", err)
				}
				return err
			}

		}

	}

	if err = tx.Commit(); err != nil {
		log.Println("update drivers: unable to commit: ", err)
		return err
	}

	d.Buffer = d.Buffer[:0]
	return nil

}

// Run TODO: go dock
func (d *DB) Run(metric *storage.Metric) (err error) {

	ctx, cancel := context.WithTimeout(d.ctx, 3*time.Second)
	defer cancel()

	switch metric.MType {

	case "gauge":
		_, err = d.DB.ExecContext(ctx,
			`INSERT INTO metrics 
    				(id, type, hash, value) 
					VALUES ($1, $2, $3, $4)`,
			metric.ID, metric.MType, metric.Hash, float64(*metric.Value),
		)
		if err != nil {
			log.Println("insert row into table went wrong, ", err)
			return err

		}

	case "counter":

		_, err = d.DB.ExecContext(ctx,
			`INSERT INTO metrics 
    				(id, type, hash, delta) 
					VALUES ($1, $2, $3, $4)`,
			metric.ID, metric.MType, metric.Hash, int(*metric.Delta),
		)
		if err != nil {
			log.Println("insert row into table went wrong, ", err)
			return err
		}
	}

	return nil
}
