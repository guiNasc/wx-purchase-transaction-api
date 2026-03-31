package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnvAsInt("DB_PORT", 5432)
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "wx_purchase_transactions")
	sslmode := getEnv("DB_SSLMODE", "disable")
	maxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 25)
	connMaxLifetime := getEnvAsDurationSeconds("DB_CONN_MAX_LIFETIME_SECONDS", 300)
	connMaxIdleTime := getEnvAsDurationSeconds("DB_CONN_MAX_IDLE_TIME_SECONDS", 120)
	connectTimeout := getEnvAsDurationSeconds("DB_CONNECT_TIMEOUT_SECONDS", 5)
	maxRetries := getEnvAsInt("DB_CONNECT_MAX_RETRIES", 8)
	initialBackoff := getEnvAsDurationMilliseconds("DB_CONNECT_BACKOFF_INITIAL_MS", 500)
	maxBackoff := getEnvAsDurationMilliseconds("DB_CONNECT_BACKOFF_MAX_MS", 5000)

	if maxRetries < 0 {
		maxRetries = 0
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	if maxOpenConns > 0 && maxIdleConns > maxOpenConns {
		maxIdleConns = maxOpenConns
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	attempts := maxRetries + 1
	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		pingCtx, cancel := context.WithTimeout(context.Background(), connectTimeout)
		lastErr = db.PingContext(pingCtx)
		cancel()

		if lastErr == nil {
			slog.Info("database connected",
				"db_name", dbname,
				"host", host,
				"port", port,
				"attempt", attempt,
				"max_open_conns", maxOpenConns,
				"max_idle_conns", maxIdleConns,
				"conn_max_lifetime", connMaxLifetime.String(),
				"conn_max_idle_time", connMaxIdleTime.String(),
			)
			return db, nil
		}

		if attempt == attempts {
			break
		}

		backoff := calculateBackoff(initialBackoff, maxBackoff, attempt-1)
		slog.Warn("database ping failed, retrying",
			"attempt", attempt,
			"max_attempts", attempts,
			"retry_in", backoff.String(),
			"error", lastErr,
		)
		time.Sleep(backoff)
	}

	if closeErr := db.Close(); closeErr != nil {
		slog.Warn("failed to close database after connection retries exhausted", "error", closeErr)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("database connection failed after %d attempts: %w", attempts, lastErr)
	}

	return nil, fmt.Errorf("database connection failed after %d attempts", attempts)
}

func calculateBackoff(initial, max time.Duration, step int) time.Duration {
	if initial <= 0 {
		initial = 500 * time.Millisecond
	}

	if max <= 0 {
		max = 5 * time.Second
	}

	if initial > max {
		return max
	}

	backoff := initial
	for i := 0; i < step; i++ {
		next := backoff * 2
		if next <= 0 || next > max {
			return max
		}
		backoff = next
	}

	if backoff > max {
		return max
	}

	return backoff
}

func getEnvAsDurationSeconds(key string, fallbackSeconds int) time.Duration {
	seconds := getEnvAsInt(key, fallbackSeconds)
	if seconds < 0 {
		seconds = fallbackSeconds
	}

	return time.Duration(seconds) * time.Second
}

func getEnvAsDurationMilliseconds(key string, fallbackMilliseconds int) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return time.Duration(fallbackMilliseconds) * time.Millisecond
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return time.Duration(fallbackMilliseconds) * time.Millisecond
	}

	return time.Duration(parsed) * time.Millisecond
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
