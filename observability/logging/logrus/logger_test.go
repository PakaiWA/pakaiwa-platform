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

package logrus

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNewLogger_ReturnsNonNil(t *testing.T) {
	l := NewLogger(logrus.InfoLevel)
	if l == nil {
		t.Fatal("Expected non-nil logger")
	}
}

func TestNewLogger_LevelSet(t *testing.T) {
	levels := []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.TraceLevel,
	}
	for _, lvl := range levels {
		l := NewLogger(lvl)
		if l.GetLevel() != lvl {
			t.Errorf("Expected level %v, got %v", lvl, l.GetLevel())
		}
	}
}

func TestNewLogger_HasOrderedJSONFormatter(t *testing.T) {
	l := NewLogger(logrus.InfoLevel)
	if _, ok := l.Formatter.(*OrderedJSONFormatter); !ok {
		t.Error("Expected formatter to be *OrderedJSONFormatter")
	}
}

// ---- Format tests ----

func logEntryWith(level logrus.Level, msg string, fields logrus.Fields) *logrus.Entry {
	l := logrus.New()
	l.SetReportCaller(false) // keep caller nil for cleaner test output
	entry := l.WithFields(fields)
	entry.Level = level
	entry.Message = msg
	entry.Time = time.Date(2026, 2, 22, 10, 0, 0, 0, time.UTC)
	return entry
}

func TestFormat_ProducesValidJSON(t *testing.T) {
	f := &OrderedJSONFormatter{
		TimestampFormat: fixedRFC3339Nano,
		LevelKey:        "level",
		TimeKey:         "time",
		MsgKey:          "msg",
		TraceIDKey:      "trace_id",
	}

	entry := logEntryWith(logrus.InfoLevel, "hello world", nil)
	out, err := f.Format(entry)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, out)
	}
}

func TestFormat_ContainsRequiredKeys(t *testing.T) {
	f := &OrderedJSONFormatter{
		TimestampFormat: fixedRFC3339Nano,
		LevelKey:        "level",
		TimeKey:         "time",
		MsgKey:          "msg",
		TraceIDKey:      "trace_id",
	}

	entry := logEntryWith(logrus.InfoLevel, "test message", nil)
	out, err := f.Format(entry)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var m map[string]any
	if err = json.Unmarshal(out, &m); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	for _, key := range []string{"level", "time", "msg"} {
		if _, ok := m[key]; !ok {
			t.Errorf("Expected key %q in output", key)
		}
	}
}

func TestFormat_TraceIDIncluded(t *testing.T) {
	f := &OrderedJSONFormatter{
		LevelKey:   "level",
		TimeKey:    "time",
		MsgKey:     "msg",
		TraceIDKey: "trace_id",
	}

	entry := logEntryWith(logrus.InfoLevel, "with trace", logrus.Fields{
		"trace_id": "abc-123",
	})
	out, err := f.Format(entry)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(string(out), "abc-123") {
		t.Errorf("Expected trace_id in output, got: %s", out)
	}
}

func TestFormat_ModuleFieldLastInOutput(t *testing.T) {
	f := &OrderedJSONFormatter{
		LevelKey:   "level",
		TimeKey:    "time",
		MsgKey:     "msg",
		TraceIDKey: "trace_id",
	}

	entry := logEntryWith(logrus.InfoLevel, "with module", logrus.Fields{
		"module": "auth",
		"user":   "alice",
	})
	out, err := f.Format(entry)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	s := string(out)
	moduleIdx := strings.Index(s, `"module"`)
	userIdx := strings.Index(s, `"user"`)
	if moduleIdx == -1 || userIdx == -1 {
		t.Fatalf("Expected both 'module' and 'user' in output: %s", s)
	}
	if moduleIdx < userIdx {
		t.Errorf("Expected 'module' to appear after 'user', output: %s", s)
	}
}

func TestFormat_ErrorFieldEncoded(t *testing.T) {
	f := &OrderedJSONFormatter{
		LevelKey:   "level",
		TimeKey:    "time",
		MsgKey:     "msg",
		TraceIDKey: "trace_id",
	}

	entry := logEntryWith(logrus.ErrorLevel, "an error occurred", logrus.Fields{
		"error": fmt_error("something went wrong"),
	})
	out, err := f.Format(entry)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if !strings.Contains(string(out), "something went wrong") {
		t.Errorf("Expected error message in output, got: %s", out)
	}
}

// fmt_error is a trivial error type helper for tests.
type testError struct{ msg string }

func (e testError) Error() string { return e.msg }

func fmt_error(msg string) error { return testError{msg} }

func TestFormat_DefaultPadLevelTo(t *testing.T) {
	// PadLevelTo=0 should default to 5
	f := &OrderedJSONFormatter{
		LevelKey:   "level",
		TimeKey:    "time",
		MsgKey:     "msg",
		TraceIDKey: "trace_id",
	}

	entry := logEntryWith(logrus.InfoLevel, "pad default", nil)
	out, err := f.Format(entry)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	if len(out) == 0 {
		t.Error("Expected non-empty output")
	}
}

func TestFormat_DefaultTimestampFormat(t *testing.T) {
	// TimestampFormat="" should use fixedRFC3339Nano
	f := &OrderedJSONFormatter{
		TimestampFormat: "",
		LevelKey:        "level",
		TimeKey:         "time",
		MsgKey:          "msg",
		TraceIDKey:      "trace_id",
	}

	entry := logEntryWith(logrus.InfoLevel, "ts default", nil)
	out, err := f.Format(entry)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
}

func TestFormat_CustomKeys(t *testing.T) {
	f := &OrderedJSONFormatter{
		LevelKey:   "lvl",
		TimeKey:    "ts",
		MsgKey:     "message",
		TraceIDKey: "tid",
	}

	entry := logEntryWith(logrus.InfoLevel, "custom keys", nil)
	out, err := f.Format(entry)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	s := string(out)
	for _, key := range []string{"lvl", "ts", "message"} {
		if !strings.Contains(s, `"`+key+`"`) {
			t.Errorf("Expected key %q in output: %s", key, s)
		}
	}
}

func TestFormat_AllLevels(t *testing.T) {
	f := &OrderedJSONFormatter{
		LevelKey:   "level",
		TimeKey:    "time",
		MsgKey:     "msg",
		TraceIDKey: "trace_id",
	}

	cases := []struct {
		level    logrus.Level
		expected string
	}{
		{logrus.WarnLevel, "warn"},
		{logrus.ErrorLevel, "error"},
		{logrus.InfoLevel, "info"},
		{logrus.DebugLevel, "debug"},
		{logrus.TraceLevel, "trace"},
		{logrus.FatalLevel, "fatal"},
		{logrus.PanicLevel, "panic"},
	}

	for _, tc := range cases {
		entry := logEntryWith(tc.level, "level test", nil)
		out, err := f.Format(entry)
		if err != nil {
			t.Fatalf("Format error for level %v: %v", tc.level, err)
		}

		if !strings.Contains(string(out), tc.expected) {
			t.Errorf("Expected level %q in output for %v, got: %s", tc.expected, tc.level, out)
		}
	}
}

func TestFormat_EscapeHTML(t *testing.T) {
	f := &OrderedJSONFormatter{
		LevelKey:   "level",
		TimeKey:    "time",
		MsgKey:     "msg",
		TraceIDKey: "trace_id",
		EscapeHTML: true,
	}

	entry := logEntryWith(logrus.InfoLevel, "escape <html> & stuff", nil)
	out, err := f.Format(entry)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Valid JSON is still expected
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("Invalid JSON with EscapeHTML=true: %v", err)
	}
}

func TestFormat_NoEscapeHTML(t *testing.T) {
	f := &OrderedJSONFormatter{
		LevelKey:   "level",
		TimeKey:    "time",
		MsgKey:     "msg",
		TraceIDKey: "trace_id",
		EscapeHTML: false,
	}

	entry := logEntryWith(logrus.InfoLevel, "no escape <html>", nil)
	out, err := f.Format(entry)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Output should still be valid JSON
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("Invalid JSON with EscapeHTML=false: %v", err)
	}
}

// TestFormat_WithCaller tests the caller info branch (e.Caller != nil).
func TestFormat_WithCaller(t *testing.T) {
	l := logrus.New()
	l.SetReportCaller(true) // enable caller info

	f := &OrderedJSONFormatter{
		LevelKey:   "level",
		TimeKey:    "time",
		MsgKey:     "msg",
		TraceIDKey: "trace_id",
	}

	// Build an entry via the real logger (so Caller is populated)
	var buf bytes.Buffer
	l.SetOutput(&buf)
	l.SetFormatter(f)
	l.Info("caller test")

	out := buf.Bytes()
	if len(out) == 0 {
		t.Fatal("Expected log output")
	}

	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("Invalid JSON with caller: %v\nOutput: %s", err, out)
	}

	if _, ok := m["caller"]; !ok {
		t.Errorf("Expected 'caller' key in output, got: %s", out)
	}
}

// ---- writeJSONString ----

func TestWriteJSONString_NormalString(t *testing.T) {
	var buf bytes.Buffer
	writeJSONString(&buf, "hello", false)
	if buf.String() != `"hello"` {
		t.Errorf("Expected `\"hello\"`, got %s", buf.String())
	}
}

func TestWriteJSONString_EmptyString(t *testing.T) {
	var buf bytes.Buffer
	writeJSONString(&buf, "", false)
	if buf.String() != `""` {
		t.Errorf("Expected empty JSON string, got %s", buf.String())
	}
}

// ---- keyOr ----

func TestKeyOr_ReturnsValue(t *testing.T) {
	result := keyOr("custom", "default")
	if result != "custom" {
		t.Errorf("Expected 'custom', got %s", result)
	}
}

func TestKeyOr_ReturnsDefault(t *testing.T) {
	result := keyOr("", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got %s", result)
	}
}

// ---- normalizeLevel ----

func TestNormalizeLevel(t *testing.T) {
	cases := map[logrus.Level]string{
		logrus.WarnLevel:  "warn",
		logrus.ErrorLevel: "error",
		logrus.FatalLevel: "fatal",
		logrus.PanicLevel: "panic",
		logrus.InfoLevel:  "info",
		logrus.DebugLevel: "debug",
		logrus.TraceLevel: "trace",
	}
	for level, want := range cases {
		got := normalizeLevel(level)
		if got != want {
			t.Errorf("normalizeLevel(%v) = %q, want %q", level, got, want)
		}
	}

	// Test the default branch with an unknown level value
	unknownLevel := logrus.Level(99)
	got := normalizeLevel(unknownLevel)
	if got == "" {
		t.Error("Expected non-empty string for unknown level")
	}
}

// ---- Integration: NewLogger output is parseable JSON ----

func TestNewLogger_OutputIsJSON(t *testing.T) {
	l := NewLogger(logrus.InfoLevel)
	var buf bytes.Buffer
	l.SetOutput(&buf)
	l.SetReportCaller(false)

	l.Info("integration test message")

	out := buf.Bytes()
	if len(out) == 0 {
		t.Fatal("Expected log output, got empty")
	}

	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("Logger output is not valid JSON: %v\nOutput: %s", err, out)
	}

	if m["msg"] != "integration test message" {
		t.Errorf("Expected msg='integration test message', got %v", m["msg"])
	}
}
