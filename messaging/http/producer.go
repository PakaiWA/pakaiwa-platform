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
 * https://github.com/PakaiWA/PakaiWA/tree/main/~/work/PakaiWA/pp/messaging/http
 */

package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/PakaiWA/pakaiwa-platform/messaging/producer"
	"github.com/sirupsen/logrus"
)

type HttpProducer struct {
	client *http.Client
	url    string
	log    *logrus.Logger
}

func NewHttpProducer(url string, log *logrus.Logger) producer.MessageProducer {
	return &HttpProducer{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		url: url,
		log: log,
	}
}

func (h *HttpProducer) Send(ctx context.Context, topic string, key []byte, clientJID []byte, value []byte) error {
	req, err := http.NewRequestWithContext(ctx, "POST", h.url, bytes.NewBuffer(value))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PakaiWA-Topic", topic)
	req.Header.Set("X-PakaiWA-Key", string(key))
	req.Header.Set("X-Device-Id", string(clientJID))

	resp, err := h.client.Do(req)
	if err != nil {
		h.log.WithError(err).Error("failed to send http message")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("http producer returned status: %d", resp.StatusCode)
		h.log.WithError(err).Error("failed to send http message")
		return err
	}

	return nil
}

func (h *HttpProducer) Flush(_ int) int {
	// HTTP is synchronous in this implementation, nothing to flush
	return 0
}

func (h *HttpProducer) Close() error {
	return nil
}
