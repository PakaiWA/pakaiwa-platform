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
 * @author KAnggara on Sunday 08/02/2026 10.07
 * @project pp
 * https://github.com/PakaiWA/pp/tree/main/observability/metrics
 */

package metrics

import (
	"net/http"
	"net/http/httptest"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HttpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed.",
		},
		[]string{"method", "path", "status"},
	)

	HttpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Execution Duration HTTP handler",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(HttpRequests)
	prometheus.MustRegister(HttpDuration)
}

func PrometheusHandler() fiber.Handler {
	return func(c fiber.Ctx) error {
		rec := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/metrics", nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("failed to create request")
		}

		promhttp.Handler().ServeHTTP(rec, req)

		for k, v := range rec.Header() {
			c.Set(k, v[0])
		}

		c.Status(rec.Code)
		return c.Send(rec.Body.Bytes())
	}
}
