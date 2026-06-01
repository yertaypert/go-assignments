package pkg

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type Postgres struct {
	Conn *sql.DB
}

func NewPostgres() (*Postgres, error) {
	host := getEnv("PG_HOST", "127.0.0.1")
	port := getEnv("PG_PORT", "5432")
	user := getEnv("PG_USER", "postgres")
	password := getEnv("PG_PASSWORD", "postgres")
	dbName := getEnv("PG_DB", "assignment7")
	sslMode := getEnv("PG_SSLMODE", "disable")

	if err := ensureDatabase(host, port, user, password, dbName, sslMode); err != nil {
		return nil, err
	}

	dsn := buildDSN(host, port, user, password, dbName, sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	pg := &Postgres{Conn: db}
	if err := pg.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return pg, nil
}

func (p *Postgres) Close() error {
	if p == nil || p.Conn == nil {
		return nil
	}

	return p.Conn.Close()
}

func (p *Postgres) migrate() error {
	const query = `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    verified BOOLEAN NOT NULL DEFAULT FALSE
);`

	if _, err := p.Conn.Exec(query); err != nil {
		return fmt.Errorf("create users table: %w", err)
	}

	return nil
}

func ensureDatabase(host, port, user, password, dbName, sslMode string) error {
	adminDSN := buildDSN(host, port, user, password, "postgres", sslMode)

	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return fmt.Errorf("sql open admin db: %w", err)
	}
	defer adminDB.Close()

	if err := adminDB.Ping(); err != nil {
		return fmt.Errorf("admin db ping: %w", err)
	}

	var exists bool
	if err := adminDB.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`,
		dbName,
	).Scan(&exists); err != nil {
		return fmt.Errorf("check database exists: %w", err)
	}

	if exists {
		return nil
	}

	query := fmt.Sprintf(`CREATE DATABASE "%s"`, strings.ReplaceAll(dbName, `"`, `""`))
	if _, err := adminDB.Exec(query); err != nil {
		return fmt.Errorf("create database %s: %w", dbName, err)
	}

	return nil
}

func buildDSN(host, port, user, password, dbName, sslMode string) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslMode,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
