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

package ctxmeta

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
)

// ---- Logger tests ----

func TestWithLogger_NilContext(t *testing.T) {
	entry := logrus.NewEntry(logrus.New())
	var nilCtx context.Context //nolint:SA1012
	result := WithLogger(nilCtx, entry)
	if result != nil {
		t.Error("Expected nil context when input is nil")
	}
}

func TestWithLogger_NilLogger(t *testing.T) {
	ctx := context.Background()
	result := WithLogger(ctx, nil)
	// Should return original ctx unchanged
	if result != ctx {
		t.Error("Expected original context when logger is nil")
	}
}

func TestWithLogger_ValidInputs(t *testing.T) {
	ctx := context.Background()
	l := logrus.New()
	entry := logrus.NewEntry(l)

	result := WithLogger(ctx, entry)
	if result == nil {
		t.Fatal("Expected non-nil context")
	}
}

func TestLogger_NilContext(t *testing.T) {
	var nilCtx context.Context //nolint:SA1012
	got := Logger(nilCtx)
	if got != nil {
		t.Errorf("Expected nil logger from nil context, got %v", got)
	}
}

func TestLogger_NoLoggerInContext(t *testing.T) {
	ctx := context.Background()
	got := Logger(ctx)
	if got != nil {
		t.Errorf("Expected nil logger when none stored, got %v", got)
	}
}

func TestLogger_RoundTrip(t *testing.T) {
	l := logrus.New()
	entry := l.WithField("test", "value")

	ctx := WithLogger(context.Background(), entry)
	got := Logger(ctx)

	if got == nil {
		t.Fatal("Expected non-nil entry from context")
	}

	if got != entry {
		t.Error("Expected the same logrus.Entry that was stored")
	}
}

func TestLogger_WrongTypeInContext(t *testing.T) {
	// Manually store a non-*logrus.Entry under LoggerKey
	ctx := context.WithValue(context.Background(), LoggerKey, "not-a-logger")
	got := Logger(ctx)
	if got != nil {
		t.Errorf("Expected nil when wrong type is stored, got %v", got)
	}
}

// ---- TraceID tests ----

func TestWithTraceID_StoresValue(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-abc-123")

	got := TraceID(ctx)
	if got != "trace-abc-123" {
		t.Errorf("Expected trace-abc-123, got %s", got)
	}
}

func TestTraceID_NilContext(t *testing.T) {
	var nilCtx context.Context //nolint:SA1012
	got := TraceID(nilCtx)
	if got != "" {
		t.Errorf("Expected empty string from nil context, got %s", got)
	}
}

func TestTraceID_NoValueInContext(t *testing.T) {
	ctx := context.Background()
	got := TraceID(ctx)
	if got != "" {
		t.Errorf("Expected empty string when no trace ID stored, got %s", got)
	}
}

func TestTraceID_WrongTypeInContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), TraceIDKey, 12345)
	got := TraceID(ctx)
	if got != "" {
		t.Errorf("Expected empty string for wrong type, got %s", got)
	}
}

func TestTraceID_EmptyString(t *testing.T) {
	ctx := WithTraceID(context.Background(), "")
	got := TraceID(ctx)
	if got != "" {
		t.Errorf("Expected empty string, got %s", got)
	}
}

// ---- Key type test ----

func TestKey_Constants(t *testing.T) {
	if LoggerKey != "logger" {
		t.Errorf("Expected LoggerKey to be 'logger', got %s", LoggerKey)
	}
	if TraceIDKey != "trace_id" {
		t.Errorf("Expected TraceIDKey to be 'trace_id', got %s", TraceIDKey)
	}
}
