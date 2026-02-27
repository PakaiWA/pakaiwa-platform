/*
 * Copyright (c) 2025 KAnggara
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * See <https://www.gnu.org/licenses/gpl-3.0.html>.
 *
 * @author KAnggara on Thursday 18/12/2025 07.38
 * @project PakaiWA
 * ~/work/PakaiWA/PakaiWA/internal/pkg/logger/ctxmeta
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/observability/logging/ctxmeta
 */

package ctxmeta

import (
	"context"

	"github.com/sirupsen/logrus"
)

type loggerKey string

const (
	loggerAppKey   loggerKey = "logger_app"
	loggerDBKey    loggerKey = "logger_db"
	loggerHTTPKey  loggerKey = "logger_http"
	loggerKafkaKey loggerKey = "logger_kafka"
	loggerWAKey    loggerKey = "logger_wa"
)

func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	if ctx == nil || logger == nil {
		return ctx
	}
	return context.WithValue(ctx, LoggerKey, logger)
}

func WithLoggers(ctx context.Context, app, db, http, kafka, wa *logrus.Entry) context.Context {
	if ctx == nil {
		return ctx
	}

	if app != nil {
		ctx = context.WithValue(ctx, loggerAppKey, app)
	}

	if db != nil {
		ctx = context.WithValue(ctx, loggerDBKey, db)
	}

	if http != nil {
		ctx = context.WithValue(ctx, loggerHTTPKey, http)
	}

	if kafka != nil {
		ctx = context.WithValue(ctx, loggerKafkaKey, kafka)
	}

	if wa != nil {
		ctx = context.WithValue(ctx, loggerWAKey, wa)
	}

	return ctx
}

func Logger(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return nil
	}

	if v, ok := ctx.Value(LoggerKey).(*logrus.Entry); ok {
		return v
	}

	return nil
}

func LoggerApp(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	if v, ok := ctx.Value(loggerAppKey).(*logrus.Entry); ok {
		return v
	}
	return logrus.NewEntry(logrus.StandardLogger())
}

func LoggerDB(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return LoggerApp(ctx)
	}
	if v, ok := ctx.Value(loggerDBKey).(*logrus.Entry); ok {
		return v
	}
	return LoggerApp(ctx)
}

func LoggerHTTP(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return LoggerApp(ctx)
	}
	if v, ok := ctx.Value(loggerHTTPKey).(*logrus.Entry); ok {
		return v
	}
	return LoggerApp(ctx)
}

func LoggerKafka(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return LoggerApp(ctx)
	}
	if v, ok := ctx.Value(loggerKafkaKey).(*logrus.Entry); ok {
		return v
	}
	return LoggerApp(ctx)
}

func LoggerWA(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return LoggerApp(ctx)
	}
	if v, ok := ctx.Value(loggerWAKey).(*logrus.Entry); ok {
		return v
	}
	return LoggerApp(ctx)
}
