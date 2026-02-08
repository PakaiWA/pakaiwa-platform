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
 * @author KAnggara on Saturday 08/02/2026 11.08
 * @project pp
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/validation
 */

package validation

import (
	"context"
	"errors"
	"testing"

	"github.com/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestGet40Space(t *testing.T) {
	result := Get40Space()

	if len(result) != 40 {
		t.Errorf("Expected 40 spaces, got %d", len(result))
	}

	for i, char := range result {
		if char != ' ' {
			t.Errorf("Expected space at position %d, got %c", i, char)
		}
	}
}

func TestTraceIDFromContext_WithTraceID(t *testing.T) {
	ctx := context.WithValue(context.Background(), "trace_id", "test-trace-123")

	result := TraceIDFromContext(ctx)

	if result != "test-trace-123" {
		t.Errorf("Expected 'test-trace-123', got %s", result)
	}
}

func TestTraceIDFromContext_WithoutTraceID(t *testing.T) {
	ctx := context.Background()

	result := TraceIDFromContext(ctx)

	if len(result) != 40 {
		t.Errorf("Expected 40 spaces as fallback, got length %d", len(result))
	}
}

func TestTraceIDFromContext_WithEmptyTraceID(t *testing.T) {
	ctx := context.WithValue(context.Background(), "trace_id", "")

	result := TraceIDFromContext(ctx)

	// Empty string should return fallback (40 spaces)
	if len(result) != 40 {
		t.Errorf("Expected 40 spaces for empty trace_id, got length %d", len(result))
	}
}

func TestTraceIDFromContext_WithWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), "trace_id", 12345) // int instead of string

	result := TraceIDFromContext(ctx)

	// Wrong type should return fallback (40 spaces)
	if len(result) != 40 {
		t.Errorf("Expected 40 spaces for wrong type, got length %d", len(result))
	}
}

func TestLogValidationErrors_NilLogger(t *testing.T) {
	ctx := context.Background() // No logger in context
	err := errors.New("test error")

	// Should not panic when logger is nil
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LogValidationErrors panicked with nil logger: %v", r)
		}
	}()

	LogValidationErrors(ctx, err, "test message", "/test/path")
}

func TestLogValidationErrors_WithGenericError(t *testing.T) {
	logger, hook := test.NewNullLogger()
	entry := logrus.NewEntry(logger)
	ctx := ctxmeta.WithLogger(context.Background(), entry)

	err := errors.New("generic error")

	LogValidationErrors(ctx, err, "test message", "/test/path")

	// Verify that error was logged
	if len(hook.Entries) == 0 {
		t.Fatal("Expected log entry, got none")
	}

	logEntry := hook.LastEntry()

	if logEntry.Level != logrus.ErrorLevel {
		t.Errorf("Expected error level, got %v", logEntry.Level)
	}

	if logEntry.Message != "test message at /test/path" {
		t.Errorf("Expected 'test message at /test/path', got %s", logEntry.Message)
	}

	if logEntry.Data["event"] != "validation_failed" {
		t.Errorf("Expected event=validation_failed, got %v", logEntry.Data["event"])
	}
}

func TestLogValidationErrors_WithValidationErrors(t *testing.T) {
	logger, hook := test.NewNullLogger()
	entry := logrus.NewEntry(logger)
	ctx := ctxmeta.WithLogger(context.Background(), entry)

	// Create a validator and trigger validation error
	v := NewValidator()

	type TestStruct struct {
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"gte=0,lte=150"`
	}

	// Invalid data
	data := TestStruct{
		Email: "invalid-email",
		Age:   200,
	}

	err := v.Struct(data)
	if err == nil {
		t.Fatal("Expected validation error")
	}

	LogValidationErrors(ctx, err, "validation failed", "/api/users")

	// Should have logged multiple entries (one per validation error)
	if len(hook.Entries) == 0 {
		t.Fatal("Expected log entries, got none")
	}

	// Check that all entries are warnings
	for _, entry := range hook.Entries {
		if entry.Level != logrus.WarnLevel {
			t.Errorf("Expected warn level for validation errors, got %v", entry.Level)
		}

		if entry.Message != "validation failed at /api/users" {
			t.Errorf("Expected 'validation failed at /api/users', got %s", entry.Message)
		}

		if entry.Data["event"] != "validation_failed" {
			t.Errorf("Expected event=validation_failed, got %v", entry.Data["event"])
		}

		// Should have field, tag, and param
		if _, ok := entry.Data["field"]; !ok {
			t.Error("Expected 'field' in log data")
		}
		if _, ok := entry.Data["tag"]; !ok {
			t.Error("Expected 'tag' in log data")
		}
	}
}

func TestLogValidationErrors_WithDefaultMessage(t *testing.T) {
	logger, hook := test.NewNullLogger()
	entry := logrus.NewEntry(logger)
	ctx := ctxmeta.WithLogger(context.Background(), entry)

	err := errors.New("test error")

	// Call without message and path
	LogValidationErrors(ctx, err)

	if len(hook.Entries) == 0 {
		t.Fatal("Expected log entry, got none")
	}

	logEntry := hook.LastEntry()

	// Should use default "unknown" for both message and path
	if logEntry.Message != "unknown at unknown" {
		t.Errorf("Expected 'unknown at unknown', got %s", logEntry.Message)
	}
}

func TestLogValidationErrors_WithMessageOnly(t *testing.T) {
	logger, hook := test.NewNullLogger()
	entry := logrus.NewEntry(logger)
	ctx := ctxmeta.WithLogger(context.Background(), entry)

	err := errors.New("test error")

	// Call with message but no path
	LogValidationErrors(ctx, err, "custom message")

	if len(hook.Entries) == 0 {
		t.Fatal("Expected log entry, got none")
	}

	logEntry := hook.LastEntry()

	// Should use custom message and default "unknown" for path
	if logEntry.Message != "custom message at unknown" {
		t.Errorf("Expected 'custom message at unknown', got %s", logEntry.Message)
	}
}

func TestLogValidationErrors_WithMultipleValidationErrors(t *testing.T) {
	logger, hook := test.NewNullLogger()
	entry := logrus.NewEntry(logger)
	ctx := ctxmeta.WithLogger(context.Background(), entry)

	v := NewValidator()

	type TestStruct struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"required,gte=18"`
	}

	// All fields invalid
	data := TestStruct{
		Name:  "",
		Email: "not-an-email",
		Age:   10,
	}

	err := v.Struct(data)
	if err == nil {
		t.Fatal("Expected validation error")
	}

	LogValidationErrors(ctx, err, "multiple errors", "/test")

	// Should have multiple log entries
	if len(hook.Entries) < 3 {
		t.Errorf("Expected at least 3 log entries, got %d", len(hook.Entries))
	}

	// Collect all fields that were logged
	fields := make(map[string]bool)
	for _, entry := range hook.Entries {
		if field, ok := entry.Data["field"].(string); ok {
			fields[field] = true
		}
	}

	// Should have logged errors for name, email, and age
	expectedFields := []string{"name", "email", "age"}
	for _, field := range expectedFields {
		if !fields[field] {
			t.Errorf("Expected field '%s' to be logged", field)
		}
	}
}
