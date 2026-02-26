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
 * @author KAnggara on Saturday 07/02/2026 18.18
 * @project pp
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/db/postgres
 */

package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func NewDatabase(ctx context.Context, cfg Config, module string) (*pgxpool.Pool, error) {
	var log *logrus.Entry
	if entry := ctxmeta.Logger(ctx); entry != nil {
		log = entry.WithField("module", module)
	} else {
		log = logrus.WithField("module", module)
	}

	// ===== Validate Config (fail-fast) =====
	if cfg.MinConns > cfg.MaxConns {
		return nil, fmt.Errorf("db config invalid: min_conns (%d) > max_conns (%d)", cfg.MinConns, cfg.MaxConns)
	}

	if cfg.MaxConnIdleTime <= 0 || cfg.MaxConnLifetime <= 0 {
		return nil, fmt.Errorf("db config invalid: conn idle/lifetime must be > 0")
	}

	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}

	// ===== Parse pgx config =====
	pgxCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}

	pgxCfg.MinConns = cfg.MinConns
	pgxCfg.MaxConns = cfg.MaxConns
	pgxCfg.MaxConnLifetime = cfg.MaxConnLifetime
	pgxCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	pgxCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	pgxCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	// Optional: set application_name untuk observability PostgreSQL
	if module != "" {
		if pgxCfg.ConnConfig.RuntimeParams == nil {
			pgxCfg.ConnConfig.RuntimeParams = map[string]string{}
		}
		pgxCfg.ConnConfig.RuntimeParams["application_name"] = module
	}

	// ===== Create Pool =====
	start := time.Now()
	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, err
	}
	log.WithFields(logrus.Fields{
		"min_conns": cfg.MinConns,
		"max_conns": cfg.MaxConns,
		"duration":  time.Since(start),
	}).Info("pgxpool initialized")

	// ===== Ping (liveness check) =====
	pingTimeout := max(cfg.ConnectTimeout, 3*time.Second)

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		log.WithError(err).Error("database ping failed")
		pool.Close()
		return nil, err
	}

	log.Info("database connection is healthy")
	return pool, nil
}
