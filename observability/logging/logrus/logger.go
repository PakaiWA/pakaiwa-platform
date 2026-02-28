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
 * @author KAnggara on Sunday 08/02/2026 10.21
 * @project pp
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/observability/logging/logrus
 */

package logrus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

const moduleKey = "module"

type OrderedJSONFormatter struct {
	TimestampFormat string // default RFC3339Nano
	LevelKey        string // default "level"
	TimeKey         string // default "time"
	MsgKey          string // default "msg"
	TraceIDKey      string // default "trace_id"
	EscapeHTML      bool
}

type Loggers struct {
	App   *logrus.Entry
	DB    *logrus.Entry
	HTTP  *logrus.Entry
	Kafka *logrus.Entry
	WA    *logrus.Entry
}

type LogConfig struct {
	Level      logrus.Level
	DBLevel    logrus.Level
	HTTPLevel  logrus.Level
	KafkaLevel logrus.Level
	WALevel    logrus.Level
}

const fixedRFC3339Nano = "2006-01-02T15:04:05.000Z07:00"

func NewLogger(logLevel logrus.Level) *logrus.Logger {
	l := logrus.New()
	l.SetLevel(logLevel)

	l.SetFormatter(&OrderedJSONFormatter{
		TimestampFormat: fixedRFC3339Nano,
		LevelKey:        "level",
		TimeKey:         "time",
		MsgKey:          "msg",
		TraceIDKey:      "trace_id",
		EscapeHTML:      false,
	})

	return l
}

func buildLogger(level logrus.Level, formatter logrus.Formatter, output io.Writer) *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(formatter)
	l.SetOutput(output)
	l.SetLevel(level)
	return l
}

func NewLoggers(cfg LogConfig) *Loggers {
	// Shared formatter & output
	formatter := &OrderedJSONFormatter{
		TimestampFormat: fixedRFC3339Nano,
		LevelKey:        "level",
		TimeKey:         "time",
		MsgKey:          "msg",
		TraceIDKey:      "trace_id",
		EscapeHTML:      false,
	}

	var output io.Writer = os.Stdout

	appLogger := buildLogger(cfg.Level, formatter, output)
	dbLogger := buildLogger(cfg.DBLevel, formatter, output)
	waLogger := buildLogger(cfg.WALevel, formatter, output)
	httpLogger := buildLogger(cfg.HTTPLevel, formatter, output)
	kafkaLogger := buildLogger(cfg.KafkaLevel, formatter, output)

	return &Loggers{
		App:   logrus.NewEntry(appLogger).WithField("scope", "app"),
		DB:    logrus.NewEntry(dbLogger).WithField("scope", "db"),
		HTTP:  logrus.NewEntry(httpLogger).WithField("scope", "http"),
		Kafka: logrus.NewEntry(kafkaLogger).WithField("scope", "kafka"),
		WA:    logrus.NewEntry(waLogger).WithField("scope", "wa"),
	}
}

func (f *OrderedJSONFormatter) Format(e *logrus.Entry) ([]byte, error) {
	tsFmt := f.TimestampFormat
	if tsFmt == "" {
		tsFmt = fixedRFC3339Nano
	}

	msgKey := keyOr(f.MsgKey, "msg")
	timeKey := keyOr(f.TimeKey, "time")
	levelKey := keyOr(f.LevelKey, "level")
	traceKey := keyOr(f.TraceIDKey, "trace_id")

	lvl := normalizeLevel(e.Level)

	trace := ""
	if v, ok := e.Data[traceKey]; ok {
		trace = fmt.Sprint(v)
	}

	buf := &bytes.Buffer{}
	buf.Grow(256)
	buf.WriteByte('{')

	writeKV(buf, timeKey, e.Time.Format(tsFmt), true, f.EscapeHTML)
	writeKV(buf, levelKey, lvl, false, f.EscapeHTML)
	if trace != "" {
		writeKV(buf, traceKey, trace, false, f.EscapeHTML)
	}
	writeKV(buf, msgKey, e.Message, false, f.EscapeHTML)

	// caller: file:line
	caller := resolveCaller()
	writeKV(buf, "caller", caller, false, f.EscapeHTML)

	if len(e.Data) > 0 {
		keys := make([]string, 0, len(e.Data))
		var moduleValue any
		for k, v := range e.Data {
			switch k {
			case traceKey:
				continue
			case moduleKey:
				moduleValue = v
			default:
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			buf.WriteByte(',')
			writeKey(buf, k, f.EscapeHTML)
			buf.WriteByte(':')

			// marshal value dengan SetEscapeHTML(f.EscapeHTML)
			var vb bytes.Buffer
			enc := json.NewEncoder(&vb)
			enc.SetEscapeHTML(f.EscapeHTML)
			v := e.Data[k]

			// SPECIAL CASE: error
			if err, ok := v.(error); ok {
				_ = enc.Encode(err.Error()) //nolint:errcheck
			} else {
				_ = enc.Encode(v) //nolint:errcheck
			}
			val := vb.Bytes()
			if len(val) > 0 && val[len(val)-1] == '\n' {
				val = val[:len(val)-1]
			}
			buf.Write(val)
		}

		if moduleValue != nil {
			buf.WriteByte(',')
			writeKey(buf, moduleKey, f.EscapeHTML)
			buf.WriteByte(':')

			var vb bytes.Buffer
			enc := json.NewEncoder(&vb)
			enc.SetEscapeHTML(f.EscapeHTML)
			_ = enc.Encode(moduleValue) //nolint:errcheck

			val := vb.Bytes()
			if len(val) > 0 && val[len(val)-1] == '\n' {
				val = val[:len(val)-1]
			}
			buf.Write(val)
		}
	}

	buf.WriteByte('}')
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func keyOr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func writeKey(buf *bytes.Buffer, k string, escapeHTML bool) {
	writeJSONString(buf, k, escapeHTML)
}

// mempengaruhi urutan log
func writeKV(buf *bytes.Buffer, k, v string, first bool, escapeHTML bool) {
	if !first {
		buf.WriteByte(',')
	}
	writeKey(buf, k, escapeHTML)
	buf.WriteByte(':')
	writeJSONString(buf, v, escapeHTML)
}

func writeJSONString(buf *bytes.Buffer, s string, escapeHTML bool) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(escapeHTML)
	if err := enc.Encode(s); err != nil {
		b.WriteString(`""`)
	}
	out := b.Bytes()
	if n := len(out); n > 0 && out[n-1] == '\n' {
		out = out[:len(out)-1]
	}
	buf.Write(out)
}

func normalizeLevel(level logrus.Level) string {
	switch level {
	case logrus.WarnLevel:
		return "warn"
	case logrus.ErrorLevel:
		return "error"
	case logrus.FatalLevel:
		return "fatal"
	case logrus.PanicLevel:
		return "panic"
	case logrus.InfoLevel:
		return "info"
	case logrus.DebugLevel:
		return "debug"
	case logrus.TraceLevel:
		return "trace"
	default:
		return level.String()
	}
}

func resolveCaller() string {
	// skip chain:
	// 0 = resolveCaller
	// 1 = Format()
	// 2 = logrus internal
	// 3 = adapter / logger wrapper
	// 4 = business code
	for i := 4; i < 10; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			continue
		}

		// skip file internal logrus
		if strings.Contains(file, "sirupsen/logrus") {
			continue
		}

		// skip formatter file
		if strings.Contains(file, "logrus_adapter.go") {
			continue
		}

		return fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	return "unknown"
}
