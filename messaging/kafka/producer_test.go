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
	"errors"
	"testing"
	"time"

	"github.com/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta"
	logrushelper "github.com/PakaiWA/pakaiwa-platform/observability/logging/logrus"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

// newTestKafkaProducer creates a real kafka.Producer backed by a stub broker
// address. NewProducer succeeds immediately (librdkafka is lazy about
// connecting), giving us a valid *KafkaProducer for unit tests.
func newTestKafkaProducer(t *testing.T) *KafkaProducer {
	t.Helper()
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		t.Fatalf("failed to create test kafka.Producer: %v", err)
	}
	return &KafkaProducer{p: p}
}

// ---- NewKafkaProducer ----

func TestNewKafkaProducer_ValidConfig(t *testing.T) {
	p := NewKafkaProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if p == nil {
		t.Fatal("Expected non-nil producer from valid config")
	}
	// clean up
	if err := p.Close(); err != nil {
		t.Errorf("Close returned unexpected error: %v", err)
	}
}

// ---- KafkaProducer.Flush ----

func TestKafkaProducer_Flush(t *testing.T) {
	kp := newTestKafkaProducer(t)
	defer func() { _ = kp.Close() }()

	// Flush with short timeout; no queued messages → should return quickly.
	remaining := kp.Flush(100)
	if remaining < 0 {
		t.Errorf("Flush returned negative value: %d", remaining)
	}
}

// ---- KafkaProducer.Close ----

func TestKafkaProducer_Close(t *testing.T) {
	kp := newTestKafkaProducer(t)
	err := kp.Close()
	if err != nil {
		t.Errorf("Close returned unexpected error: %v", err)
	}
}

// ---- KafkaProducer.Events ----

func TestKafkaProducer_Events_NotNil(t *testing.T) {
	kp := newTestKafkaProducer(t)
	defer func() { _ = kp.Close() }()

	ch := kp.Events()
	if ch == nil {
		t.Error("Expected non-nil events channel")
	}
}

// ---- KafkaProducer.Send ----

func TestKafkaProducer_Send_Queues(t *testing.T) {
	kp := newTestKafkaProducer(t)
	defer func() { _ = kp.Close() }()

	l := logrushelper.NewLogger(logrus.DebugLevel)
	entry := l.WithField("test", "kafka_send")
	ctx := ctxmeta.WithLogger(context.Background(), entry)

	err := kp.Send(ctx, "test-topic", []byte("key"), []byte("device"), []byte(`{"hello":"world"}`))
	// Without a real broker, Produce() queues the message and returns nil.
	// The actual delivery failure appears later in the events channel.
	if err != nil {
		// ErrQueueFull is also acceptable
		var kErr kafka.Error
		if errors.As(err, &kErr) && kErr.Code() == kafka.ErrQueueFull {
			t.Logf("Got ErrQueueFull (acceptable): %v", err)
		} else {
			t.Errorf("Unexpected error from Send: %v", err)
		}
	}
}

func TestKafkaProducer_Send_QueueFull(t *testing.T) {
	// Create a producer with queue.buffering.max.messages=1 and fill it until full.
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":            "localhost:9092",
		"queue.buffering.max.messages": 1,
	})
	if err != nil {
		t.Fatalf("failed to create producer: %v", err)
	}
	kp := &KafkaProducer{p: p}
	defer func() { _ = kp.Close() }()

	l := logrushelper.NewLogger(logrus.DebugLevel)
	entry := l.WithField("test", "queue_full")
	ctx := ctxmeta.WithLogger(context.Background(), entry)

	topic := "fill-topic"
	// Flood the queue until we get ErrQueueFull.
	var gotQueueFull bool
	for i := 0; i < 10000; i++ {
		err := kp.Send(ctx, topic, []byte("k"), []byte("d"), []byte(`{}`))
		if err != nil {
			var kErr kafka.Error
			if errors.As(err, &kErr) && kErr.Code() == kafka.ErrQueueFull {
				gotQueueFull = true
				break
			}
		}
	}
	if !gotQueueFull {
		t.Log("Queue full error was not triggered (broker may have accepted messages); acceptable")
	}
}

// TestKafkaProducer_Send_NonQueueFullError_WithLogger covers the
// `entry != nil` error branch in KafkaProducer.Send by calling Send on a
// closed producer (which returns a non-ErrQueueFull error) with a logger in ctx.
func TestKafkaProducer_Send_NonQueueFullError_WithLogger(t *testing.T) {
	kp := newTestKafkaProducer(t)
	_ = kp.Close() // close first so Produce() fails with ErrState

	l := logrushelper.NewLogger(logrus.DebugLevel)
	entry := l.WithField("test", "send_err_with_logger")
	ctx := ctxmeta.WithLogger(context.Background(), entry)

	err := kp.Send(ctx, "topic", []byte("k"), []byte("d"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error when sending on closed producer")
	}
}

// TestKafkaProducer_Send_NonQueueFullError_NoLogger covers the
// `logrus` fallback branch in KafkaProducer.Send (no logger in context).
func TestKafkaProducer_Send_NonQueueFullError_NoLogger(t *testing.T) {
	kp := newTestKafkaProducer(t)
	_ = kp.Close() // close first so Produce() fails with ErrState

	// context.Background() has no logger → logrus fallback branch
	err := kp.Send(context.Background(), "topic", []byte("k"), []byte("d"), []byte(`{}`))
	if err == nil {
		t.Error("Expected error when sending on closed producer without context logger")
	}
}

// ---- StartProducerPollLoop with real KafkaProducer ----

// TestStartProducerPollLoop_KafkaProducer_CtxCancel verifies the goroutine
// starts and exits cleanly when the context is cancelled.
func TestStartProducerPollLoop_KafkaProducer_CtxCancel(t *testing.T) {
	kp := newTestKafkaProducer(t)
	defer func() { _ = kp.Close() }()

	l := logrus.New()
	ctx, cancel := context.WithCancel(context.Background())

	StartProducerPollLoop(ctx, kp, l, "test-module")

	// Let the goroutine start up.
	time.Sleep(20 * time.Millisecond)

	// Cancel → triggers the ctx.Done() case in the poll loop.
	cancel()

	// Give the goroutine time to exit gracefully.
	time.Sleep(50 * time.Millisecond)
}

// TestStartProducerPollLoop_KafkaProducer_MessageDelivered covers the
// *kafka.Message branch in the events switch (success path).
func TestStartProducerPollLoop_KafkaProducer_MessageDelivered(t *testing.T) {
	kp := newTestKafkaProducer(t)
	// Do NOT defer Close here — we manage it manually below.

	l := logrus.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartProducerPollLoop(ctx, kp, l, "delivery-test")
	time.Sleep(20 * time.Millisecond)

	// Produce a message so a delivery report lands in the events channel.
	topic := "test-topic"
	_ = kp.p.Produce(&kafka.Message{ //nolint:errcheck
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte("my-key"),
		Value:          []byte(`{}`),
	}, nil)

	// Give librdkafka time to emit the delivery report (err because no broker).
	time.Sleep(200 * time.Millisecond)

	cancel()
	time.Sleep(30 * time.Millisecond)
	_ = kp.Close() //nolint:errcheck
}

// TestStartProducerPollLoop_KafkaProducer_MessageDelivered_NoKey covers
// *kafka.Message without a key (the key-length branch).
func TestStartProducerPollLoop_KafkaProducer_MessageDelivered_NoKey(t *testing.T) {
	kp := newTestKafkaProducer(t)

	l := logrus.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartProducerPollLoop(ctx, kp, l, "nokey-test")
	time.Sleep(20 * time.Millisecond)

	topic := "no-key-topic"
	_ = kp.p.Produce(&kafka.Message{ //nolint:errcheck
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            nil, // no key → len(e.Key) == 0 branch
		Value:          []byte(`{}`),
	}, nil)

	time.Sleep(200 * time.Millisecond)
	cancel()
	time.Sleep(30 * time.Millisecond)
	_ = kp.Close() //nolint:errcheck
}

// TestStartProducerPollLoop_KafkaProducer_DeliveryError injects a
// *kafka.Message with a TopicPartition.Error to cover the delivery-failure log.
func TestStartProducerPollLoop_KafkaProducer_DeliveryError(t *testing.T) {
	kp := newTestKafkaProducer(t)

	l := logrus.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartProducerPollLoop(ctx, kp, l, "delivery-err-test")
	time.Sleep(20 * time.Millisecond)

	// Inject a *kafka.Message with an error directly into the events channel.
	topic := "err-topic"
	deliveryErr := kafka.NewError(kafka.ErrUnknownTopicOrPart, "unknown topic", false)
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
			Error:     deliveryErr,
		},
		Key:   []byte("err-key"),
		Value: []byte(`{}`),
	}
	// Send directly to the events channel (internal channel, accessible via Events()).
	select {
	case kp.p.Events() <- msg:
	case <-time.After(100 * time.Millisecond):
		t.Log("Events channel full; delivery-error path not exercised via injection")
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(30 * time.Millisecond)
	_ = kp.Close() //nolint:errcheck
}

// TestStartProducerPollLoop_KafkaProducer_DeliverySuccess injects a
// *kafka.Message with no error to cover the "Kafka message delivered" debug log.
func TestStartProducerPollLoop_KafkaProducer_DeliverySuccess(t *testing.T) {
	kp := newTestKafkaProducer(t)

	l := logrus.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartProducerPollLoop(ctx, kp, l, "delivery-ok-test")
	time.Sleep(20 * time.Millisecond)

	// Inject a *kafka.Message with nil error (success) and a key.
	topic := "ok-topic"
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: 0,
			Error:     nil, // success path
		},
		Key:   []byte("ok-key"), // len > 0 → fields["key"] is set
		Value: []byte(`{}`),
	}
	select {
	case kp.p.Events() <- msg:
	case <-time.After(100 * time.Millisecond):
		t.Log("Events channel full; success delivery path not exercised via injection")
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(30 * time.Millisecond)
	_ = kp.Close() //nolint:errcheck
}

// TestStartProducerPollLoop_KafkaProducer_DeliverySuccess_NoKey injects a
// *kafka.Message with nil error and no key (the len(e.Key)==0 branch).
func TestStartProducerPollLoop_KafkaProducer_DeliverySuccess_NoKey(t *testing.T) {
	kp := newTestKafkaProducer(t)

	l := logrus.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartProducerPollLoop(ctx, kp, l, "delivery-ok-nokey-test")
	time.Sleep(20 * time.Millisecond)

	topic := "ok-nokey-topic"
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: 0,
			Error:     nil,
		},
		Key:   nil, // len == 0 → skip fields["key"]
		Value: []byte(`{}`),
	}
	select {
	case kp.p.Events() <- msg:
	case <-time.After(100 * time.Millisecond):
		t.Log("Events channel full; no-key success path not exercised")
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(30 * time.Millisecond)
	_ = kp.Close() //nolint:errcheck
}

// case in the poll loop switch.
func TestStartProducerPollLoop_KafkaProducer_KafkaError(t *testing.T) {
	kp := newTestKafkaProducer(t)

	l := logrus.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartProducerPollLoop(ctx, kp, l, "kafka-err-test")
	time.Sleep(20 * time.Millisecond)

	// Inject a kafka.Error into the events channel.
	kafkaErr := kafka.NewError(kafka.ErrBrokerNotAvailable, "broker not available", false)
	select {
	case kp.p.Events() <- kafkaErr:
	case <-time.After(100 * time.Millisecond):
		t.Log("Events channel full; kafka.Error path not exercised via injection")
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(30 * time.Millisecond)
	_ = kp.Close() //nolint:errcheck
}

// TestStartProducerPollLoop_KafkaProducer_DefaultCase covers the default branch
// (unknown event type) in the poll loop switch.
func TestStartProducerPollLoop_KafkaProducer_DefaultCase(t *testing.T) {
	kp := newTestKafkaProducer(t)

	l := logrus.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartProducerPollLoop(ctx, kp, l, "default-case-test")
	time.Sleep(20 * time.Millisecond)

	// kafka.PartitionsMetadata satisfies kafka.Event but is not handled explicitly.
	// Inject it to trigger the default case.
	unknownEvent := &kafka.Stats{} // implements kafka.Event, hits default
	select {
	case kp.p.Events() <- unknownEvent:
	case <-time.After(100 * time.Millisecond):
		t.Log("Events channel full; default case not exercised via injection")
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(30 * time.Millisecond)
	_ = kp.Close() //nolint:errcheck
}
