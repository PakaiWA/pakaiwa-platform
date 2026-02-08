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
 * @author KAnggara on Tuesday 16/12/2025 06.35
 * @project PakaiWA
 * ~/work/PakaiWA/PakaiWA/internal/pkg/logger/ctxmeta
 * https://github.com/PakaiWA/PakaiWA/tree/main/internal/pkg/logger/ctxmeta
 */

package ctxmeta

import "context"

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

func TraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(TraceIDKey).(string); ok {
		return v
	}
	return ""
}
