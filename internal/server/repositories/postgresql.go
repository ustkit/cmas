package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	// Register pgx stdlib
	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/ustkit/cmas/internal/server/config"
	"github.com/ustkit/cmas/internal/types"
)

var errNoDBConn = errors.New("no database connection")

type RepoPostgreSQL struct {
	db     *sql.DB
	config *config.Config
}

func NewRepositoryPostgreSQL(serverConfig *config.Config) (repo RepoPostgreSQL, err error) {
	db, err := sql.Open("pgx", serverConfig.DataBaseDSN)
	repo = RepoPostgreSQL{
		db:     db,
		config: serverConfig,
	}

	if err != nil {
		return repo, err
	}

	return
}

func (repo RepoPostgreSQL) Save(ctx context.Context, name string, value types.Value) error {
	if repo.db == nil {
		return errNoDBConn
	}

	_, err := repo.db.ExecContext(ctx,
		`INSERT INTO metrics (id, type, delta, gauge) VALUES($1, $2, $3, $4)  
		 ON CONFLICT (id, type) 
		 DO UPDATE SET delta = metrics.delta + excluded.delta, gauge = $4`,
		name, value.TValue, value.CValue, value.GValue)
	if err != nil {
		return err
	}

	if repo.config.StoreInterval == "0" {
		return repo.SaveToFile()
	}

	return nil
}

func (repo RepoPostgreSQL) SaveAll(ctx context.Context, values []types.ValueJSON) (err error) {
	if repo.db == nil {
		return errNoDBConn
	}

	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil && tx != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("save all: tx err %w: roll back err %v", err, rbErr)
			}
		}
	}()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO metrics (id, type, delta, gauge) VALUES($1, $2, $3, $4)  
		 ON CONFLICT (id, type) 
		 DO UPDATE SET delta = metrics.delta + excluded.delta, gauge = $4`)
	if err != nil {
		return err
	}

	defer stmt.Close()

	for _, v := range values {
		var (
			delta types.Counter
			value types.Gauge
		)

		if v.Delta != nil {
			delta = *v.Delta
		}

		if v.Value != nil {
			value = *v.Value
		}

		if _, err = stmt.ExecContext(ctx, v.ID, v.MType, delta, value); err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if repo.config.StoreInterval == "0" {
		return repo.SaveToFile()
	}

	return nil
}

func (repo RepoPostgreSQL) FindByName(ctx context.Context, name string) (value types.Value, err error) {
	if repo.db == nil {
		return value, errNoDBConn
	}

	err = repo.db.QueryRowContext(ctx,
		`SELECT type, delta, gauge FROM metrics WHERE id = $1`, name).
		Scan(&value.TValue, &value.CValue, &value.GValue)

	return
}

func (repo RepoPostgreSQL) FindAll(ctx context.Context) (values types.Values, err error) {
	if repo.db == nil {
		return values, errNoDBConn
	}

	values = make(types.Values)

	rows, err := repo.db.QueryContext(ctx, "SELECT id, type, delta, gauge FROM metrics")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			mName  string
			mValue types.Value
		)

		err = rows.Scan(&mName, &mValue.TValue, &mValue.CValue, &mValue.GValue)
		if err != nil {
			return nil, err
		}

		values[mName] = &mValue
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (repo RepoPostgreSQL) Restore() error {
	return nil
}

func (repo RepoPostgreSQL) SaveToFile() error {
	return nil
}

func (repo RepoPostgreSQL) Close() error {
	if repo.db == nil {
		return errNoDBConn
	}

	return repo.db.Close()
}

func (repo RepoPostgreSQL) Ping(ctx context.Context) error {
	if repo.db == nil {
		return errNoDBConn
	}

	return repo.db.PingContext(ctx)
}
