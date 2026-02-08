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
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/db/postgres
 */

package postgres

import (
	"testing"
	"time"
)

func TestConfig_Creation(t *testing.T) {
	cfg := Config{
		DSN:               "postgres://user:pass@localhost:5432/testdb",
		MinConns:          2,
		MaxConns:          10,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
		ConnectTimeout:    5 * time.Second,
	}

	if cfg.DSN != "postgres://user:pass@localhost:5432/testdb" {
		t.Errorf("Expected DSN to be set correctly, got %s", cfg.DSN)
	}

	if cfg.MinConns != 2 {
		t.Errorf("Expected MinConns to be 2, got %d", cfg.MinConns)
	}

	if cfg.MaxConns != 10 {
		t.Errorf("Expected MaxConns to be 10, got %d", cfg.MaxConns)
	}

	if cfg.MaxConnIdleTime != 30*time.Minute {
		t.Errorf("Expected MaxConnIdleTime to be 30 minutes, got %v", cfg.MaxConnIdleTime)
	}

	if cfg.HealthCheckPeriod != 1*time.Minute {
		t.Errorf("Expected HealthCheckPeriod to be 1 minute, got %v", cfg.HealthCheckPeriod)
	}

	if cfg.ConnectTimeout != 5*time.Second {
		t.Errorf("Expected ConnectTimeout to be 5 seconds, got %v", cfg.ConnectTimeout)
	}
}

func TestConfig_ZeroValues(t *testing.T) {
	cfg := Config{}

	if cfg.DSN != "" {
		t.Errorf("Expected empty DSN, got %s", cfg.DSN)
	}

	if cfg.MinConns != 0 {
		t.Errorf("Expected MinConns to be 0, got %d", cfg.MinConns)
	}

	if cfg.MaxConns != 0 {
		t.Errorf("Expected MaxConns to be 0, got %d", cfg.MaxConns)
	}
}
