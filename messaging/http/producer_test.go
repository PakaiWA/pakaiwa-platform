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

// mustNewHttpProducer is a test helper that calls NewHttpProducer and
// fatals if construction fails (used for URLs we know are valid).
func mustNewHttpProducer(t *testing.T, rawURL string) *HttpProducer {
	t.Helper()
	p, err := NewHttpProducer(rawURL)
	if err != nil {
		t.Fatalf("NewHttpProducer(%q) unexpected error: %v", rawURL, err)
	}
	return p.(*HttpProducer)
}

// ---- NewHttpProducer validation ----

func TestNewHttpProducer_ValidHTTP(t *testing.T) {
	p, err := NewHttpProducer("http://localhost:8080/webhook")
	if err != nil {
		t.Fatalf("Expected no error for http URL, got %v", err)
	}
	if p == nil {
		t.Fatal("Expected non-nil producer")
	}
}

func TestNewHttpProducer_ValidHTTPS(t *testing.T) {
	p, err := NewHttpProducer("https://example.com/webhook")
	if err != nil {
		t.Fatalf("Expected no error for https URL, got %v", err)
	}
	if p == nil {
		t.Fatal("Expected non-nil producer")
	}
}

func TestNewHttpProducer_InvalidURL(t *testing.T) {
	_, err := NewHttpProducer("://bad-url")
	if err == nil {
		t.Error("Expected error for unparseable URL, got nil")
	}
}

func TestNewHttpProducer_BadScheme_File(t *testing.T) {
	_, err := NewHttpProducer("file:///etc/passwd")
	if err == nil {
		t.Error("Expected error for file:// scheme (SSRF risk), got nil")
	}
}

func TestNewHttpProducer_BadScheme_Gopher(t *testing.T) {
	_, err := NewHttpProducer("gopher://evil.example.com")
	if err == nil {
		t.Error("Expected error for gopher:// scheme (SSRF risk), got nil")
	}
}

func TestNewHttpProducer_BadScheme_FTP(t *testing.T) {
	_, err := NewHttpProducer("ftp://files.example.com/data")
	if err == nil {
		t.Error("Expected error for ftp:// scheme, got nil")
	}
}

// ---- Send tests ----

func TestHttpProducer_Send_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("X-PakaiWA-Topic") != "test-topic" {
			t.Errorf("Expected X-PakaiWA-Topic 'test-topic', got %s", r.Header.Get("X-PakaiWA-Topic"))
		}
		if r.Header.Get("X-PakaiWA-Key") != "key-1" {
			t.Errorf("Expected X-PakaiWA-Key 'key-1', got %s", r.Header.Get("X-PakaiWA-Key"))
		}
		if r.Header.Get("X-Device-Id") != "device-abc" {
			t.Errorf("Expected X-Device-Id 'device-abc', got %s", r.Header.Get("X-Device-Id"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p := mustNewHttpProducer(t, server.URL)
	err := p.Send(contextWithLogger(), "test-topic", []byte("key-1"), []byte("device-abc"), []byte(`{"msg":"hello"}`))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHttpProducer_Send_Non2xxStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	p := mustNewHttpProducer(t, server.URL)
	err := p.Send(contextWithLogger(), "topic", []byte("key"), []byte("device"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error for non-2xx status code")
	}
}

func TestHttpProducer_Send_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	p := mustNewHttpProducer(t, server.URL)
	err := p.Send(contextWithLogger(), "topic", []byte("key"), []byte("device"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error for 400 Bad Request")
	}
}

func TestHttpProducer_Send_201Created(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated) // 201 is still 2xx
	}))
	defer server.Close()

	p := mustNewHttpProducer(t, server.URL)
	err := p.Send(contextWithLogger(), "topic", []byte("key"), []byte("device"), []byte(`{}`))
	if err != nil {
		t.Fatalf("Expected no error for 201 Created, got %v", err)
	}
}

func TestHttpProducer_Send_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p := mustNewHttpProducer(t, server.URL)

	ctx, cancel := context.WithCancel(contextWithLogger())
	cancel() // cancel immediately

	err := p.Send(ctx, "topic", []byte("key"), []byte("device"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

// TestHttpProducer_Send_ContextCancelled_NoLogger covers the logrus fallback
// branch in Send's error path (no logger in context).
func TestHttpProducer_Send_ContextCancelled_NoLogger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p := mustNewHttpProducer(t, server.URL)

	ctx, cancel := context.WithCancel(context.Background()) // no logger
	cancel()

	err := p.Send(ctx, "topic", []byte("key"), []byte("device"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error for cancelled context without logger")
	}
}

// TestHttpProducer_Send_Non2xx_NoLogger covers the logrus fallback branch
// in the status-code error path when there is no logger in context.
func TestHttpProducer_Send_Non2xx_NoLogger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	p := mustNewHttpProducer(t, server.URL)
	err := p.Send(context.Background(), "topic", []byte("key"), []byte("device"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error for 502 without context logger")
	}
}

func TestHttpProducer_Flush(t *testing.T) {
	p := mustNewHttpProducer(t, "http://localhost")
	result := p.Flush(1000)
	if result != 0 {
		t.Errorf("Expected Flush to return 0 (HTTP is synchronous), got %d", result)
	}
}

func TestHttpProducer_Close(t *testing.T) {
	p := mustNewHttpProducer(t, "http://localhost")
	err := p.Close()
	if err != nil {
		t.Errorf("Expected nil error from Close, got %v", err)
	}
}

func TestHttpProducer_Send_EmptyPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p := mustNewHttpProducer(t, server.URL)
	err := p.Send(contextWithLogger(), "topic", nil, nil, nil)
	if err != nil {
		t.Fatalf("Expected no error for nil payload, got %v", err)
	}
}

func TestHttpProducer_Send_Status299(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(299) // Top edge of 2xx
	}))
	defer server.Close()

	p := mustNewHttpProducer(t, server.URL)
	err := p.Send(contextWithLogger(), "topic", []byte("k"), []byte("d"), []byte(`{}`))
	if err != nil {
		t.Fatalf("Expected no error for status 299, got %v", err)
	}
}

func TestHttpProducer_Send_Status300(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(300) // Just outside 2xx
	}))
	defer server.Close()

	p := mustNewHttpProducer(t, server.URL)
	err := p.Send(contextWithLogger(), "topic", []byte("k"), []byte("d"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error for status 300 (outside 2xx)")
	}
}
