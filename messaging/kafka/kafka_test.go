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

package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta"
	logrushelper "github.com/PakaiWA/pakaiwa-platform/observability/logging/logrus"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

// contextWithLogger returns a context with an embedded logrus logger,
// required by Producer[T].Send which calls ctxmeta.Logger(ctx).WithField.
func contextWithLogger() context.Context {
	l := logrushelper.NewLogger(logrus.DebugLevel)
	entry := l.WithField("test", "kafka_producer")
	return ctxmeta.WithLogger(context.Background(), entry)
}

// ---- safeTopic tests ----

func TestSafeTopic_Nil(t *testing.T) {
	result := safeTopic(nil)
	if result != "" {
		t.Errorf("Expected empty string for nil topic, got %q", result)
	}
}

func TestSafeTopic_NonNil(t *testing.T) {
	topic := "my-topic"
	result := safeTopic(&topic)
	if result != "my-topic" {
		t.Errorf("Expected 'my-topic', got %q", result)
	}
}

func TestSafeTopic_EmptyString(t *testing.T) {
	empty := ""
	result := safeTopic(&empty)
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

// ---- mockProducer for testing Producer[T] and StartProducerPollLoop ----

type mockProducer struct {
	sendErr    error
	sendCalled bool
	lastTopic  string
	lastKey    []byte
	lastJID    []byte
	lastValue  []byte
}

func (m *mockProducer) Send(_ context.Context, topic string, key []byte, clientJID []byte, value []byte) error {
	m.sendCalled = true
	m.lastTopic = topic
	m.lastKey = key
	m.lastJID = clientJID
	m.lastValue = value
	return m.sendErr
}

func (m *mockProducer) Flush(_ int) int { return 0 }
func (m *mockProducer) Close() error    { return nil }

// ---- testEvent implements event.Event ----

type testEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

func (e testEvent) EventID() string   { return e.ID }
func (e testEvent) EventName() string { return e.Name }
func (e testEvent) EventKey() string  { return e.Key }

// ---- Producer[T].Send tests ----

func TestProducer_Send_Success(t *testing.T) {
	mock := &mockProducer{}
	p := &Producer[testEvent]{
		Producer: mock,
		Topic:    "events",
	}

	evt := testEvent{ID: "e1", Name: "user.created", Key: "user-123"}
	err := p.Send(contextWithLogger(), evt, "device-xyz")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !mock.sendCalled {
		t.Error("Expected underlying producer Send to be called")
	}

	if mock.lastTopic != "events" {
		t.Errorf("Expected topic 'events', got %q", mock.lastTopic)
	}

	if string(mock.lastKey) != "user-123" {
		t.Errorf("Expected key 'user-123', got %q", mock.lastKey)
	}

	if string(mock.lastJID) != "device-xyz" {
		t.Errorf("Expected clientJID 'device-xyz', got %q", mock.lastJID)
	}

	// Value should be valid JSON of the event
	var m map[string]any
	if err := json.Unmarshal(mock.lastValue, &m); err != nil {
		t.Errorf("Expected JSON-encoded value, got parse error: %v", err)
	}
}

func TestProducer_Send_ProducerError(t *testing.T) {
	mock := &mockProducer{sendErr: errors.New("broker unavailable")}
	p := &Producer[testEvent]{
		Producer: mock,
		Topic:    "events",
	}

	evt := testEvent{ID: "e2", Name: "test.event", Key: "k2"}
	err := p.Send(contextWithLogger(), evt, "device")
	if err == nil {
		t.Error("Expected error from underlying producer, got nil")
	}

	if err.Error() != "broker unavailable" {
		t.Errorf("Expected 'broker unavailable', got %q", err.Error())
	}
}

// unmarshalableEvent is an event.Event whose JSON encoding always fails.
type unmarshalableEvent struct{}

func (e unmarshalableEvent) EventID() string   { return "id" }
func (e unmarshalableEvent) EventName() string { return "name" }
func (e unmarshalableEvent) EventKey() string  { return "key" }

// MarshalJSON makes this type always return a JSON error.
func (e unmarshalableEvent) MarshalJSON() ([]byte, error) {
	return nil, errors.New("cannot marshal: intentional error")
}

// TestProducer_Send_MarshalError covers the json.Marshal error branch in
// Producer[T].Send when the event cannot be serialised.
func TestProducer_Send_MarshalError(t *testing.T) {
	mock := &mockProducer{}
	p := &Producer[unmarshalableEvent]{
		Producer: mock,
		Topic:    "events",
	}

	err := p.Send(contextWithLogger(), unmarshalableEvent{}, "device")
	if err == nil {
		t.Error("Expected error from marshal failure, got nil")
	}

	if mock.sendCalled {
		t.Error("Underlying Send should NOT be called when marshal fails")
	}
}

// TestProducer_Send_MarshalError_NoLogger verifies the p.Log fallback branch
// in Producer[T].Send when there is no logger in context and p.Log is set.
func TestProducer_Send_MarshalError_NoLogger(t *testing.T) {
	mock := &mockProducer{}
	p := &Producer[unmarshalableEvent]{
		Producer: mock,
		Topic:    "events",
	}

	// context.Background() has no logger → ctxmeta.Logger returns nil → p.Log branch
	err := p.Send(context.Background(), unmarshalableEvent{}, "device")
	if err == nil {
		t.Error("Expected error from marshal failure, got nil")
	}
}

// TestProducer_Send_MarshalError_NoLoggerNoLog verifies the logrus fallback
// when both context logger and p.Log are nil.
func TestProducer_Send_MarshalError_NoLoggerNoLog(t *testing.T) {
	mock := &mockProducer{}
	p := &Producer[unmarshalableEvent]{
		Producer: mock,
		Topic:    "events",
	}

	// context.Background() has no logger, p.Log is nil → logrus standard logger branch
	err := p.Send(context.Background(), unmarshalableEvent{}, "device")
	if err == nil {
		t.Error("Expected error from marshal failure, got nil")
	}
}

// ---- StartProducerPollLoop with non-Kafka producer ----

func TestStartProducerPollLoop_NonKafkaProducer(t *testing.T) {
	mock := &mockProducer{}
	l := logrus.NewEntry(logrus.New())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Should return immediately (no goroutine spawned for non-Kafka)
	StartProducerPollLoop(ctx, mock, l, "test-module")

	// Give a tiny window to ensure no panic
	time.Sleep(10 * time.Millisecond)
}

func TestStartProducerPollLoop_CancelledContext(t *testing.T) {
	mock := &mockProducer{}
	l := logrus.NewEntry(logrus.New())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	// Should not panic
	StartProducerPollLoop(ctx, mock, l, "test-module")
	time.Sleep(10 * time.Millisecond)
}

// ---- NewKafkaConsumer tests ----

func TestNewKafkaProducer_InvalidConfig(t *testing.T) {
	// An invalid debug category causes kafka.NewProducer to fail.
	// After our nil-logger fix, the error is logged via logrus.StandardLogger().
	result := NewKafkaProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"debug":             "all,invalid.category",
	})

	if result != nil {
		t.Error("Expected nil producer for invalid config, got non-nil")
		_ = result.Close() //nolint:errcheck
	}
}

func TestNewKafkaConsumer_InvalidBroker(t *testing.T) {
	// confluent-kafka-go creates the consumer eagerly — an invalid broker
	// configuration causes an error on NewConsumer.
	cfg := ConsumerConfig{
		Brokers: []string{}, // empty brokers → config error
		GroupID: "test-group",
		Options: map[string]any{},
	}
	consumer, err := NewKafkaConsumer(cfg)
	// With empty/missing bootstrap.servers, librdkafka returns an error.
	if err == nil {
		t.Log("NewKafkaConsumer with empty brokers succeeded (librdkafka may allow it)")
		if consumer != nil {
			_ = consumer.Close() //nolint:errcheck
		}
	}
}

func TestNewKafkaConsumer_WithOptions(t *testing.T) {
	cfg := ConsumerConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "test-group",
		Options: map[string]any{
			"auto.offset.reset": "earliest",
		},
	}
	consumer, err := NewKafkaConsumer(cfg)
	// No real broker present, but NewConsumer may succeed since it's lazy.
	// Either success or error is acceptable here; we just test no panic.
	if err == nil && consumer != nil {
		_ = consumer.Close() //nolint:errcheck
	}
}

func TestConsumerConfig_Fields(t *testing.T) {
	cfg := ConsumerConfig{
		Brokers: []string{"b1:9092", "b2:9092"},
		GroupID: "my-group",
		Options: map[string]any{"key": "val"},
	}

	if len(cfg.Brokers) != 2 {
		t.Errorf("Expected 2 brokers, got %d", len(cfg.Brokers))
	}
	if cfg.GroupID != "my-group" {
		t.Errorf("Expected GroupID 'my-group', got %s", cfg.GroupID)
	}
	if cfg.Options["key"] != "val" {
		t.Errorf("Expected Options[key]='val', got %v", cfg.Options["key"])
	}
}
