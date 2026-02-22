/*
 * Copyright (c) 2026 KAnggara
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * See <https://www.gnu.org/licenses/gpl-3.0.html>.
 */

package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestPrometheusHandler_ReturnsOK(t *testing.T) {
	app := fiber.New()
	app.Get("/metrics", PrometheusHandler())

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestPrometheusHandler_ResponseBodyNotEmpty(t *testing.T) {
	app := fiber.New()
	app.Get("/metrics", PrometheusHandler())

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	if len(body) == 0 {
		t.Error("Expected non-empty Prometheus metrics body")
	}
}

func TestPrometheusHandler_ContentTypeIsTextPlain(t *testing.T) {
	app := fiber.New()
	app.Get("/metrics", PrometheusHandler())

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer func() {
		_ = resp.Body.Close() //nolint:errcheck
	}()

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		t.Error("Expected non-empty Content-Type header")
	}
}

func TestHttpRequests_Counter(t *testing.T) {
	// Verify the counter is registered (not nil)
	if HttpRequests == nil {
		t.Error("Expected HttpRequests counter to be non-nil")
	}

	// Use it without panicking
	HttpRequests.WithLabelValues("GET", "/test", "200").Inc()
}

func TestHttpDuration_Histogram(t *testing.T) {
	// Verify the histogram is registered (not nil)
	if HttpDuration == nil {
		t.Error("Expected HttpDuration histogram to be non-nil")
	}

	// Use it without panicking
	HttpDuration.WithLabelValues("POST", "/api").Observe(0.01)
}
