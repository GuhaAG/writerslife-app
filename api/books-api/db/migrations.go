package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/guhaag/writerslife-books-api/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	database, err := sql.Open("pgx", pool.Config().ConnString())
	if err != nil {
		return fmt.Errorf("open migration database: %w", err)
	}
	defer database.Close()

	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}
	source, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}
	migration, err := migrate.NewWithInstance("iofs", source, "writerslife", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer migration.Close()

	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
