/*
 * Copyright (c) 2026 KAnggara
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * See <https://www.gnu.org/licenses/gpl-3.0.html>.
 *
 * @author KAnggara on Saturday 07/02/2026 18.39
 * @project pp
 * https://github.com/PakaiWA/PakaiWA/tree/main/~/work/PakaiWA/pp/db/postgres
 */

package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNewDatabase_InvalidDSN(t *testing.T) {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	cfg := Config{
		DSN:               "invalid-dsn",
		MinConns:          2,
		MaxConns:          10,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
		ConnectTimeout:    5 * time.Second,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid DSN, but didn't panic")
		}
	}()

	ctx := context.Background()
	NewDatabase(ctx, log, cfg)
}

func TestNewDatabase_ValidConfig(t *testing.T) {
	// Skip if no database is available
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("Skipping database test: TEST_DATABASE_URL not set")
	}

	log := logrus.New()
	log.SetOutput(os.Stdout)

	cfg := Config{
		DSN:               dsn,
		MinConns:          2,
		MaxConns:          10,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
		ConnectTimeout:    5 * time.Second,
	}

	ctx := context.Background()
	pool := NewDatabase(ctx, log, cfg)

	if pool == nil {
		t.Error("Expected pool to be initialized, got nil")
	}

	defer pool.Close()

	// Test ping
	if err := pool.Ping(ctx); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}

	// Verify pool stats
	stats := pool.Stat()
	if stats.MaxConns() != cfg.MaxConns {
		t.Errorf("Expected MaxConns to be %d, got %d", cfg.MaxConns, stats.MaxConns())
	}
}
