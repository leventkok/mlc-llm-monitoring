package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	loadEnv()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := connect(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect failed: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "begin tx: %v\n", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE scores SET scored_by = NULL
		WHERE scored_by IN (
			SELECT id FROM users
			WHERE username = $1 OR email = $2
		)`, "admin", "leventkok63@gmail.com")
	if err != nil {
		fmt.Fprintf(os.Stderr, "clear scores: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("scores updated: %d\n", tag.RowsAffected())

	tag, err = tx.Exec(ctx, `
		DELETE FROM users
		WHERE username = $1 OR email = $2`, "admin", "leventkok63@gmail.com")
	if err != nil {
		fmt.Fprintf(os.Stderr, "delete users: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("users deleted: %d\n", tag.RowsAffected())

	if err := tx.Commit(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "commit: %v\n", err)
		os.Exit(1)
	}
}

func connect(ctx context.Context, raw string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(raw)
	if err == nil {
		return pgxpool.NewWithConfig(ctx, cfg)
	}

	rebuilt, err := rebuildDatabaseURL(raw)
	if err != nil {
		return nil, err
	}
	cfg, err = pgxpool.ParseConfig(rebuilt)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}

func rebuildDatabaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	scheme := "postgresql"
	if strings.HasPrefix(raw, "postgresql://") {
		raw = strings.TrimPrefix(raw, "postgresql://")
	} else if strings.HasPrefix(raw, "postgres://") {
		scheme = "postgres"
		raw = strings.TrimPrefix(raw, "postgres://")
	} else {
		return "", fmt.Errorf("invalid DATABASE_URL scheme")
	}

	at := strings.LastIndex(raw, "@")
	if at <= 0 {
		return "", fmt.Errorf("invalid DATABASE_URL")
	}

	userInfo := raw[:at]
	hostPart := raw[at+1:]

	colon := strings.Index(userInfo, ":")
	if colon <= 0 {
		return "", fmt.Errorf("invalid DATABASE_URL userinfo")
	}

	user := userInfo[:colon]
	password := userInfo[colon+1:]

	host := hostPart
	dbName := "postgres"
	if slash := strings.Index(hostPart, "/"); slash >= 0 {
		host = hostPart[:slash]
		dbName = strings.TrimPrefix(hostPart[slash+1:], "/")
	}

	query := "sslmode=require"
	if q := strings.Index(dbName, "?"); q >= 0 {
		query = dbName[q+1:]
		dbName = dbName[:q]
	}

	return fmt.Sprintf("%s://%s:%s@%s/%s?%s",
		scheme,
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		dbName,
		query,
	), nil
}

func loadEnv() {
	candidates := []string{
		filepath.Join("..", ".cursor", "mcp.env"),
		filepath.Join(".cursor", "mcp.env"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			_ = godotenv.Load(p)
			return
		}
	}
}
