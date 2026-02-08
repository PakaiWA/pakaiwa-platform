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
 * @author KAnggara on Saturday 07/02/2026 19.36
 * @project pp
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/http/client
 */

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	client *http.Client
	once   sync.Once
)

func Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return getClient().Do(req)
}

func Post[T any](ctx context.Context, url string, body T) (*http.Response, error) {
	return doJSON(ctx, http.MethodPost, url, body)
}

func Put[T any](ctx context.Context, url string, body T) (*http.Response, error) {
	return doJSON(ctx, http.MethodPut, url, body)
}

func Patch[T any](ctx context.Context, url string, body T) (*http.Response, error) {
	return doJSON(ctx, http.MethodPatch, url, body)
}

func doJSON[T any](ctx context.Context, method, url string, body T) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// validate JSON
	var js json.RawMessage
	if err := json.Unmarshal(b, &js); err != nil {
		return nil, errors.New("invalid JSON payload")
	}

	fmt.Println("Raw bytes: ", b)         // tampil seperti angka
	fmt.Println("As string: ", string(b)) // {"name":"Vin"}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return getClient().Do(req)
}

func getClient() *http.Client {
	once.Do(func() {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	})
	return client
}
