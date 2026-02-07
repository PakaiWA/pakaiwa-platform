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
 * https://github.com/PakaiWA/PakaiWA/tree/main/~/work/PakaiWA/pp/db/postgres
 */

package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"github.com/PakaiWA/pakaiwa-platform/errors"
)

// NewDatabase creates and configures a new PostgreSQL connection pool.
// It initializes the pool with the provided configuration and verifies connectivity.
// The function will panic if the DSN is invalid or if pool creation fails.
func NewDatabase(ctx context.Context, log *logrus.Logger, cfg Config) *pgxpool.Pool {
	log.Info("Connecting to database...")

	pgxCfg := errors.Must(pgxpool.ParseConfig(cfg.DSN))

	pgxCfg.MinConns = cfg.MinConns
	pgxCfg.MaxConns = cfg.MaxConns
	pgxCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	pgxCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	pgxCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	start := time.Now()
	pool := errors.Must(pgxpool.NewWithConfig(ctx, pgxCfg))
	log.WithField("duration", time.Since(start)).Debug("pgxpool initialized")

	pingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		log.WithError(err).Fatal("database ping failed")
	}

	return pool
}
