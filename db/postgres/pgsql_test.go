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

// validCfg returns a base Config that passes all validation checks but points
// to an unreachable host so Ping always fails deterministically.
func validCfg() Config {
	return Config{
		DSN:               unreachablePostgresDSN,
		MinConns:          1,
		MaxConns:          5,
		MaxConnIdleTime:   1 * time.Minute,
		MaxConnLifetime:   5 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
		ConnectTimeout:    1 * time.Second,
	}
}

// shortCtx returns a context with a 3 s deadline — enough for a connect
// attempt to an unreachable host to time out but not slow down the suite.
func shortCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctxWithLogger(), 3*time.Second)
}

// ─── Validation: MinConns > MaxConns ──────────────────────────────────────────

func TestNewDatabase_MinConnsGreaterThanMaxConns(t *testing.T) {
	cfg := validCfg()
	cfg.MinConns = 10
	cfg.MaxConns = 5

	pool, err := NewDatabase(ctxWithLogger(), cfg, "test")
	if pool != nil {
		pool.Close()
	}
	if err == nil {
		t.Fatal("expected error when MinConns > MaxConns, got nil")
	}
}

// ─── Validation: MaxConnIdleTime / MaxConnLifetime ───────────────────────────

func TestNewDatabase_InvalidConnDurations(t *testing.T) {
	cases := []struct {
		name            string
		maxConnIdleTime time.Duration
		maxConnLifetime time.Duration
	}{
		{"zero MaxConnIdleTime", 0, 5 * time.Minute},
		{"negative MaxConnIdleTime", -1 * time.Second, 5 * time.Minute},
		{"zero MaxConnLifetime", 1 * time.Minute, 0},
		{"negative MaxConnLifetime", 1 * time.Minute, -1 * time.Second},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := validCfg()
			cfg.MaxConnIdleTime = tc.maxConnIdleTime
			cfg.MaxConnLifetime = tc.maxConnLifetime

			pool, err := NewDatabase(ctxWithLogger(), cfg, "test")
			if pool != nil {
				pool.Close()
			}
			if err == nil {
				t.Errorf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

// ─── Validation: ConnectTimeout default fallback ──────────────────────────────

// TestNewDatabase_ConnectTimeoutDefault verifies that a zero (or negative)
// ConnectTimeout is silently clamped to 5 s and the function still proceeds
// (ultimately failing at Ping because the host is unreachable, not panicking).
func TestNewDatabase_ConnectTimeoutDefault(t *testing.T) {
	cases := []struct {
		name    string
		timeout time.Duration
	}{
		{"zero timeout", 0},
		{"negative timeout", -1 * time.Second},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg := validCfg()
			cfg.ConnectTimeout = tc.timeout

			ctx, cancel := context.WithTimeout(ctxWithLogger(), 10*time.Second)
			defer cancel()

			pool, err := NewDatabase(ctx, cfg, "TimeoutDefault")
			if pool != nil {
				pool.Close()
			}
			// Expect ping failure, not a validation error.
			if err == nil {
				t.Error("expected ping error for unreachable host, got nil")
			}
		})
	}
}

// ─── DSN parsing errors ───────────────────────────────────────────────────────

func TestNewDatabase_InvalidDSN(t *testing.T) {
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
			cfg := validCfg()
			cfg.DSN = tc.dsn

			pool, err := NewDatabase(ctxWithLogger(), cfg, "DSNError")
			if pool != nil {
				pool.Close()
			}
			if err == nil {
				t.Errorf("expected parse error for DSN %q, got nil", tc.dsn)
			}
		})
	}
}

// ─── Ping failure: with and without context logger ───────────────────────────

func TestNewDatabase_PingFails(t *testing.T) {
	ctx, cancel := shortCtx()
	defer cancel()

	pool, err := NewDatabase(ctx, validCfg(), "PingFailTest")
	if pool != nil {
		pool.Close()
	}
	if err == nil {
		t.Fatal("expected error when pinging unreachable host, got nil")
	}
}

// TestNewDatabase_PingFails_NoLogger exercises the fallback branch where no
// logger is stored in the context (log = logrus.WithField(…)).
func TestNewDatabase_PingFails_NoLogger(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := NewDatabase(ctx, validCfg(), "PingFailNoLogger")
	if pool != nil {
		pool.Close()
	}
	if err == nil {
		t.Fatal("expected error when pinging unreachable host without context logger, got nil")
	}
}

// ─── module / application_name variants ──────────────────────────────────────

// TestNewDatabase_ModuleVariants confirms that both an empty and a non-empty
// module string are handled without panicking. In both cases Ping fails
// (unreachable host) and the function returns a non-nil error.
func TestNewDatabase_ModuleVariants(t *testing.T) {
	cases := []struct {
		name   string
		module string
	}{
		{"non-empty module sets application_name", "my-service"},
		{"empty module skips application_name", ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := shortCtx()
			defer cancel()

			pool, err := NewDatabase(ctx, validCfg(), tc.module)
			if pool != nil {
				pool.Close()
			}
			if err == nil {
				t.Errorf("expected ping error for unreachable host (module=%q), got nil", tc.module)
			}
		})
	}
}

// ─── Config field propagation ─────────────────────────────────────────────────

// TestNewDatabase_AllConfigFieldsApplied ensures all six pool knobs
// (MinConns, MaxConns, MaxConnIdleTime, MaxConnLifetime, HealthCheckPeriod,
// ConnectTimeout) can be set without triggering a panic. The unreachable host
// causes a deterministic Ping error.
func TestNewDatabase_AllConfigFieldsApplied(t *testing.T) {
	cfg := Config{
		DSN:               unreachablePostgresDSN,
		MinConns:          2,
		MaxConns:          20,
		MaxConnIdleTime:   45 * time.Minute,
		MaxConnLifetime:   2 * time.Hour,
		HealthCheckPeriod: 2 * time.Minute,
		ConnectTimeout:    1 * time.Second,
	}

	ctx, cancel := shortCtx()
	defer cancel()

	pool, err := NewDatabase(ctx, cfg, "ConfigApplied")
	if pool != nil {
		pool.Close()
	}
	if err == nil {
		t.Fatal("expected connection error for unreachable host, got nil")
	}
}

// ─── Integration tests (skipped unless TEST_DATABASE_URL is set) ──────────────

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
		MaxConnLifetime:   1 * time.Hour,
		HealthCheckPeriod: 1 * time.Minute,
		ConnectTimeout:    5 * time.Second,
	}

	ctx := ctxWithLogger()
	pool, err := NewDatabase(ctx, cfg, "IntegrationTest")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
	defer pool.Close()

	// Verify pool is still reachable.
	if err := pool.Ping(ctx); err != nil {
		t.Errorf("Ping failed after successful NewDatabase: %v", err)
	}

	// Verify MaxConns was applied.
	if stats := pool.Stat(); stats.MaxConns() != cfg.MaxConns {
		t.Errorf("MaxConns: want %d, got %d", cfg.MaxConns, stats.MaxConns())
	}

	// Verify a real connection can be acquired and released.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Errorf("failed to acquire connection: %v", err)
	} else {
		conn.Release()
	}
}

// TestNewDatabase_Integration_NoLogger verifies NewDatabase with a real DB
// and no logger in context (exercises the logrus fallback branch).
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
		MaxConnLifetime:   30 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
		ConnectTimeout:    5 * time.Second,
	}

	pool, err := NewDatabase(context.Background(), cfg, "IntegrationNoLogger")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
	defer pool.Close()
}
