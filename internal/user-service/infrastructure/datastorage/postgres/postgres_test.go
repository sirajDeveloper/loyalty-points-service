//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testPool      *pgxpool.Pool
	testContainer testcontainers.Container
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	testContainer, err = postgres.RunContainer(
		ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(1).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to start container: %v", err))
	}

	host, err := testContainer.Host(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to get host: %v", err))
	}
	port, err := testContainer.MappedPort(ctx, "5432")
	if err != nil {
		panic(fmt.Sprintf("failed to get port: %v", err))
	}
	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	var poolErr error
	for i := 0; i < 10; i++ {
		testPool, poolErr = pgxpool.New(ctx, connStr)
		if poolErr == nil {
			if err := testPool.Ping(ctx); err == nil {
				break
			}
			testPool.Close()
		}
		time.Sleep(1 * time.Second)
	}
	if poolErr != nil {
		panic(fmt.Sprintf("failed to connect after retries: %v", poolErr))
	}
	if testPool == nil {
		panic("pool is nil after connection attempts")
	}

	if err := runMigrations(connStr); err != nil {
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}

	code := m.Run()

	if testPool != nil {
		testPool.Close()
	}
	if testContainer != nil {
		if err := testContainer.Terminate(ctx); err != nil {
			panic(fmt.Sprintf("failed to terminate container: %v", err))
		}
	}

	os.Exit(code)
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	cleanupDB(t, testPool)
	return testPool
}

func cleanupDB(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	_, err := pool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
	if err != nil {
		t.Logf("failed to truncate users: %v", err)
	}

	_, err = pool.Exec(ctx, "ALTER SEQUENCE users_id_seq RESTART WITH 1")
	if err != nil {
		t.Logf("failed to reset sequence users_id_seq: %v", err)
	}
}

func runMigrations(connStr string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "migrations")); err == nil {
			break
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return fmt.Errorf("migrations directory not found")
		}
		wd = parent
	}

	migrationPath := filepath.Join(wd, "migrations", "gophermart")
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationPath),
		connStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
