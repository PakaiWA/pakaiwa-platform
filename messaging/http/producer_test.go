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

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta"
	logrushelper "github.com/PakaiWA/pakaiwa-platform/observability/logging/logrus"
	"github.com/sirupsen/logrus"
)

// contextWithLogger creates a context with a logger for testing.
func contextWithLogger() context.Context {
	l := logrushelper.NewLogger(logrus.DebugLevel)
	entry := l.WithField("test", "http_producer")
	return ctxmeta.WithLogger(context.Background(), entry)
}

func TestNewHttpProducer_ReturnsNonNil(t *testing.T) {
	p := NewHttpProducer("http://localhost:8080/webhook")
	if p == nil {
		t.Fatal("Expected non-nil producer")
	}
}

func TestHttpProducer_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify expected headers
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("X-PakaiWA-Topic") != "test-topic" {
			t.Errorf("Expected X-PakaiWA-Topic header 'test-topic', got %s", r.Header.Get("X-PakaiWA-Topic"))
		}
		if r.Header.Get("X-PakaiWA-Key") != "key-1" {
			t.Errorf("Expected X-PakaiWA-Key header 'key-1', got %s", r.Header.Get("X-PakaiWA-Key"))
		}
		if r.Header.Get("X-Device-Id") != "device-abc" {
			t.Errorf("Expected X-Device-Id header 'device-abc', got %s", r.Header.Get("X-Device-Id"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p := NewHttpProducer(server.URL)
	ctx := contextWithLogger()

	err := p.Send(ctx, "test-topic", []byte("key-1"), []byte("device-abc"), []byte(`{"msg":"hello"}`))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHttpProducer_Send_Non2xxStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p := NewHttpProducer(server.URL)
	ctx := contextWithLogger()

	err := p.Send(ctx, "topic", []byte("key"), []byte("device"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error for non-2xx status code")
	}
}

func TestHttpProducer_Send_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	p := NewHttpProducer(server.URL)
	ctx := contextWithLogger()

	err := p.Send(ctx, "topic", []byte("key"), []byte("device"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error for 400 Bad Request")
	}
}

func TestHttpProducer_Send_201Created(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated) // 201 is still 2xx
	}))
	defer server.Close()

	p := NewHttpProducer(server.URL)
	ctx := contextWithLogger()

	err := p.Send(ctx, "topic", []byte("key"), []byte("device"), []byte(`{}`))
	if err != nil {
		t.Fatalf("Expected no error for 201 Created, got %v", err)
	}
}

func TestHttpProducer_Send_InvalidURL(t *testing.T) {
	p := NewHttpProducer("://invalid-url")
	ctx := contextWithLogger()

	err := p.Send(ctx, "topic", []byte("key"), []byte("device"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestHttpProducer_Send_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p := NewHttpProducer(server.URL)

	ctx, cancel := context.WithCancel(contextWithLogger())
	cancel() // cancel immediately

	err := p.Send(ctx, "topic", []byte("key"), []byte("device"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

func TestHttpProducer_Flush(t *testing.T) {
	p := NewHttpProducer("http://localhost")
	hp := p.(*HttpProducer)

	result := hp.Flush(1000)
	if result != 0 {
		t.Errorf("Expected Flush to return 0 (HTTP is synchronous), got %d", result)
	}
}

func TestHttpProducer_Close(t *testing.T) {
	p := NewHttpProducer("http://localhost")
	hp := p.(*HttpProducer)

	err := hp.Close()
	if err != nil {
		t.Errorf("Expected nil error from Close, got %v", err)
	}
}

func TestHttpProducer_Send_EmptyPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p := NewHttpProducer(server.URL)
	ctx := contextWithLogger()

	err := p.Send(ctx, "topic", nil, nil, nil)
	if err != nil {
		t.Fatalf("Expected no error for nil payload, got %v", err)
	}
}

func TestHttpProducer_Send_Status299(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(299) // Edge of 2xx range
	}))
	defer server.Close()

	p := NewHttpProducer(server.URL)
	ctx := contextWithLogger()

	err := p.Send(ctx, "topic", []byte("k"), []byte("d"), []byte(`{}`))
	if err != nil {
		t.Fatalf("Expected no error for status 299, got %v", err)
	}
}

func TestHttpProducer_Send_Status300(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(300) // Just outside 2xx
	}))
	defer server.Close()

	p := NewHttpProducer(server.URL)
	ctx := contextWithLogger()

	err := p.Send(ctx, "topic", []byte("k"), []byte("d"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error for status 300 (outside 2xx)")
	}
}
