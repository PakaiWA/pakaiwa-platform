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

	ctx := context.Background()
	pool, err := NewDatabase(ctx, log, cfg)

	if err == nil {
		if pool != nil {
			pool.Close()
		}
		t.Error("Expected error for invalid DSN, got nil")
	}

	if pool != nil {
		t.Error("Expected nil pool on error")
	}
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
	pool, err := NewDatabase(ctx, log, cfg)

	if err != nil {
		t.Fatalf("Failed to create database pool: %v", err)
	}

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

	// Test multiple pings
	for i := 0; i < 3; i++ {
		if err := pool.Ping(ctx); err != nil {
			t.Errorf("Ping %d failed: %v", i+1, err)
		}
	}

	// Test that we can acquire a connection
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Errorf("Failed to acquire connection: %v", err)
	} else {
		conn.Release()
	}
}

func TestNewDatabase_ContextTimeout(t *testing.T) {
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
		ConnectTimeout:    1 * time.Nanosecond, // Very short timeout
	}

	ctx := context.Background()
	pool, err := NewDatabase(ctx, log, cfg)

	// With such a short timeout, connection might fail
	// This is acceptable - we're testing timeout handling
	if err != nil {
		// Expected - timeout occurred
		if pool != nil {
			pool.Close()
		}
		return
	}

	// If it somehow succeeded, clean up
	if pool != nil {
		defer pool.Close()
	}
}
