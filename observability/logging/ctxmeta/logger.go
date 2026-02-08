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
 * https://github.com/PakaiWA/PakaiWA/tree/main/internal/pkg/logger/ctxmeta
 */

package ctxmeta

import (
	"context"

	"github.com/sirupsen/logrus"
)

func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	if ctx == nil || logger == nil {
		return ctx
	}
	return context.WithValue(ctx, LoggerKey, logger)
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
