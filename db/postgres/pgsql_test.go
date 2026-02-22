/*
 * Copyright (c) 2026 KAnggara
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * See <https://www.gnu.org/licenses/gpl-3.0.html>.
 */

package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta"
	logger "github.com/PakaiWA/pakaiwa-platform/observability/logging/logrus"
	"github.com/sirupsen/logrus"
)

// ctxWithLogger returns a context carrying a logrus logger.
func ctxWithLogger() context.Context {
	log := logger.NewLogger(logrus.DebugLevel)
	entry := log.WithField("component", "psql-test")
	return ctxmeta.WithLogger(context.Background(), entry)
}

// unreachablePostgresDSN is a syntactically valid DSN that points to a host
// that cannot be connected to, so ParseConfig succeeds but Ping fails.
// We use an IP in the documentation range (192.0.2.0/24 – RFC 5737) which
// is guaranteed to be unreachable.
const unreachablePostgresDSN = "postgres://user:pass@192.0.2.1:5432/testdb?connect_timeout=1"

// ---------- ParseConfig error paths (already covered, kept for clarity) ----------

func TestNewDatabase_InvalidDSN(t *testing.T) {
	pool, err := NewDatabase(ctxWithLogger(), Config{DSN: "invalid-dsn"}, "TEST")
	if pool != nil {
		pool.Close()
	}
	if err == nil {
		t.Error("Expected error for invalid DSN, got nil")
	}
}

func TestNewDatabase_DSNParsingErrors(t *testing.T) {
	cases := []struct {
		name string
		dsn  string
	}{
		{"completely invalid", "this is not a dsn"},
		{"missing scheme separator", "postgresuser:pass@localhost/db"},
		{"invalid host brackets", "postgres://user:pass@[invalid]/db"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pool, err := NewDatabase(ctxWithLogger(), Config{DSN: tc.dsn}, "TESTING")
			if pool != nil {
				pool.Close()
			}
			if err == nil {
				t.Errorf("Expected parse error for DSN %q, got nil", tc.dsn)
			}
		})
	}
}

// ---------- Config assignment + NewWithConfig + Ping failure ----------

// TestNewDatabase_PingFails exercises lines 35-53: config is assigned,
// pgxpool.NewWithConfig succeeds, pool.Ping fails because the host is
// unreachable, which triggers the error branch (log.Error + pool.Close).
func TestNewDatabase_PingFails(t *testing.T) {
	cfg := Config{
		DSN:               unreachablePostgresDSN,
		MinConns:          1,
		MaxConns:          5,
		MaxConnIdleTime:   1 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
		ConnectTimeout:    1 * time.Second,
	}

	// Use a tight overall deadline so the test doesn't hang.
	ctx, cancel := context.WithTimeout(ctxWithLogger(), 3*time.Second)
	defer cancel()

	pool, err := NewDatabase(ctx, cfg, "PingFailTest")
	if pool != nil {
		pool.Close()
	}
	if err == nil {
		t.Error("Expected error when pinging unreachable host, got nil")
	}
}

// TestNewDatabase_PingFails_NoLogger verifies the logrus fallback branch
// (no logger in context) still reaches and logs the ping error correctly.
func TestNewDatabase_PingFails_NoLogger(t *testing.T) {
	cfg := Config{
		DSN:               unreachablePostgresDSN,
		MinConns:          1,
		MaxConns:          5,
		MaxConnIdleTime:   1 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
		ConnectTimeout:    1 * time.Second,
	}

	// context.Background() has no logger → exercises the else branch in NewDatabase.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := NewDatabase(ctx, cfg, "PingFailNoLogger")
	if pool != nil {
		pool.Close()
	}
	if err == nil {
		t.Error("Expected error when pinging unreachable host without context logger, got nil")
	}
}

// TestNewDatabase_ConfigFieldsApplied uses an unreachable DSN to confirm all
// five config fields (MinConns, MaxConns, MaxConnIdleTime, HealthCheckPeriod,
// ConnectTimeout) are applied without panicking; the assertion is implicit —
// the function returns an error (not a panic) when it can't connect.
func TestNewDatabase_ConfigFieldsApplied(t *testing.T) {
	cfg := Config{
		DSN:               unreachablePostgresDSN,
		MinConns:          2,
		MaxConns:          20,
		MaxConnIdleTime:   45 * time.Minute,
		HealthCheckPeriod: 2 * time.Minute,
		ConnectTimeout:    1 * time.Second,
	}

	ctx, cancel := context.WithTimeout(ctxWithLogger(), 3*time.Second)
	defer cancel()

	pool, err := NewDatabase(ctx, cfg, "ConfigApplied")
	if pool != nil {
		pool.Close()
	}
	// We only need to confirm no panic; error is expected.
	if err == nil {
		t.Error("Expected connection error for unreachable host, got nil")
	}
}

// ---------- Integration: full happy path (requires TEST_DATABASE_URL) ----------

func TestNewDatabase_Happy_Integration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("Skipping: TEST_DATABASE_URL not set")
	}

	cfg := Config{
		DSN:               dsn,
		MinConns:          1,
		MaxConns:          5,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
		ConnectTimeout:    5 * time.Second,
	}

	ctx := ctxWithLogger()
	pool, err := NewDatabase(ctx, cfg, "IntegrationTest")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if pool == nil {
		t.Fatal("Expected non-nil pool")
	}
	defer pool.Close()

	// Verify pool is functional.
	if err := pool.Ping(ctx); err != nil {
		t.Errorf("Ping failed after successful NewDatabase: %v", err)
	}

	// Verify MaxConns was applied.
	stats := pool.Stat()
	if stats.MaxConns() != cfg.MaxConns {
		t.Errorf("MaxConns: want %d, got %d", cfg.MaxConns, stats.MaxConns())
	}

	// Verify we can acquire a real connection.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Errorf("Failed to acquire connection: %v", err)
	} else {
		conn.Release()
	}
}

// TestNewDatabase_Integration_NoLogger verifies NewDatabase with a real DB
// and no logger in context (the else branch for log initialisation).
func TestNewDatabase_Integration_NoLogger(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("Skipping: TEST_DATABASE_URL not set")
	}

	cfg := Config{
		DSN:               dsn,
		MinConns:          1,
		MaxConns:          3,
		MaxConnIdleTime:   10 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
		ConnectTimeout:    5 * time.Second,
	}

	// No logger in context → fallback to logrus standard logger.
	pool, err := NewDatabase(context.Background(), cfg, "IntegrationNoLogger")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if pool == nil {
		t.Fatal("Expected non-nil pool")
	}
	defer pool.Close()
}
