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
 * @author KAnggara on Sunday 08/02/2026 10.14
 * @project pp
 * https://github.com/PakaiWA/pp/tree/main/validation
 */

package validation

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

func LogValidationErrors(ctx context.Context, err error, data ...string) {
	log := ctxmeta.Logger(ctx)
	if log == nil {
		return // atau fallback logger infra jika mau
	}

	message := "unknown"
	if len(data) > 0 {
		message = data[0]
	}

	path := "unknown"
	if len(data) > 1 {
		path = data[1]
	}

	msg := fmt.Sprintf("%s at %s", message, path)

	var validationError validator.ValidationErrors
	if errors.As(err, &validationError) {
		for _, v := range validationError {
			log.WithFields(logrus.Fields{
				"event": "validation_failed",
				"field": v.Field(),
				"tag":   v.Tag(),
				"param": v.Param(),
			}).Warn(msg)
		}
	} else {
		log.WithError(err).
			WithField("event", "validation_failed").
			Error(msg)
	}
}

func TraceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value("trace_id").(string); ok && v != "" {
		return v
	}
	return Get40Space()
}

func Get40Space() string {
	return strings.Repeat(" ", 40)
}
