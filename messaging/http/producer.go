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
 * @author KAnggara on Saturday 07/02/2026 20.25
 * @project pp
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/messaging/http
 */

package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/PakaiWA/pakaiwa-platform/messaging/producer"
	"github.com/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta"
	logger "github.com/PakaiWA/pakaiwa-platform/observability/logging/logrus"
	"github.com/sirupsen/logrus"
)

type HttpProducer struct {
	client *http.Client
	url    string
}

// NewHttpProducer creates an HttpProducer that posts events to the given URL.
// The URL must use the http or https scheme; any other scheme (file, gopher,
// etc.) is rejected to prevent Server-Side Request Forgery (SSRF).
func NewHttpProducer(rawURL string) (producer.MessageProducer, error) {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook URL %q: %w", rawURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("webhook URL scheme must be http or https, got %q", parsed.Scheme)
	}

	return &HttpProducer{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		url: rawURL,
	}, nil
}

func (h *HttpProducer) Send(ctx context.Context, topic string, key []byte, clientJID []byte, value []byte) error {
	log := ctxmeta.Logger(ctx)
	if log == nil {
		base := logger.NewLogger(logrus.InfoLevel)
		log = logrus.NewEntry(base)
	}

	if topic == "" {
		return fmt.Errorf("topic must not be empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(value))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PakaiWA-Topic", topic)

	if len(key) > 0 {
		req.Header.Set("X-PakaiWA-Key", string(key))
	}

	if len(clientJID) > 0 {
		req.Header.Set("X-Device-ID", string(clientJID))
	}

	start := time.Now()

	resp, err := h.client.Do(req)
	if err != nil {
		log.WithFields(logrus.Fields{
			"device_id": string(clientJID),
			"event":     topic,
			"latency":   time.Since(start).String(),
		}).WithError(err).Error("failed to send http message")
		return err
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body) // drain body for connection reuse
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("http producer returned status=%d", resp.StatusCode)

		log.WithFields(logrus.Fields{
			"device_id": string(clientJID),
			"event":     topic,
			"status":    resp.StatusCode,
			"latency":   time.Since(start).String(),
		}).WithError(err).Error("http producer non-2xx response")

		return err
	}

	log.WithFields(logrus.Fields{
		"device_id": string(clientJID),
		"event":     topic,
		"status":    resp.StatusCode,
		"latency":   time.Since(start).String(),
	}).Debug("http message sent successfully")

	return nil
}

func (h *HttpProducer) Flush(_ int) int {
	// HTTP is synchronous in this implementation, nothing to flush
	return 0
}

func (h *HttpProducer) Close() error {
	return nil
}
