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
 * @author KAnggara on Saturday 07/02/2026 20.01
 * @project pp
 * https://github.com/PakaiWA/pakaiwa-platform/tree/main/messaging/kafka
 */

package kafka

import (
	"context"
	"encoding/json"

	"github.com/PakaiWA/pakaiwa-platform/messaging/event"
	"github.com/PakaiWA/pakaiwa-platform/messaging/producer"
	"github.com/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

type KafkaProducer struct {
	p *kafka.Producer
}

func NewKafkaProducer(cfg *kafka.ConfigMap) producer.MessageProducer {
	p, err := kafka.NewProducer(cfg)
	if err != nil {
		logrus.WithError(err).Error("failed to create kafka producer")
		return nil
	}

	return &KafkaProducer{
		p: p,
	}
}

func (k *KafkaProducer) Send(ctx context.Context, topic string, key []byte, clientJID []byte, value []byte) error {
	entry := ctxmeta.LoggerKafka(ctx)

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   key,
		Value: value,
		Headers: []kafka.Header{
			{
				Key:   "device_id",
				Value: clientJID,
			},
		},
	}

	err := k.p.Produce(msg, nil)
	if err != nil {
		if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrQueueFull {
			// Jika queue penuh, kita bisa push synchronously atau biarkan caller handle.
			// Untuk simplicity, kita return error agar caller bisa retry jika perlu,
			// atau handle di level atas.
			return err
		}
		if entry != nil {
			entry.WithField("device_id", clientJID).WithError(err).Error("failed to produce kafka message")
		} else {
			logrus.WithField("device_id", clientJID).WithError(err).Error("failed to produce kafka message")
		}
		return err
	}

	return nil
}

func (k *KafkaProducer) Flush(timeoutMs int) int {
	return k.p.Flush(timeoutMs)
}

func (k *KafkaProducer) Close() error {
	k.p.Close()
	return nil
}

// Events returns the underlying kafka.Producer events channel.
// This is needed for the poll loop which is Kafka-specific.
func (k *KafkaProducer) Events() chan kafka.Event {
	return k.p.Events()
}

type Producer[T event.Event] struct {
	Producer producer.MessageProducer
	Topic    string
}

func (p *Producer[T]) Send(ctx context.Context, evt T, clientJID string) error {
	entry := ctxmeta.LoggerKafka(ctx)
	value, err := json.Marshal(evt)
	if err != nil {
		if entry != nil {
			entry.WithField("device_id", clientJID).WithError(err).Error("failed to marshal event")
		} else {
			logrus.WithField("device_id", clientJID).WithError(err).Error("failed to marshal event")
		}
		return err
	}

	return p.Producer.Send(
		ctx,
		p.Topic,
		[]byte(evt.EventKey()),
		[]byte(clientJID),
		value,
	)
}

func StartProducerPollLoop(
	ctx context.Context,
	producer producer.MessageProducer,
	log *logrus.Entry,
	module string,
) {
	kp, ok := producer.(*KafkaProducer)
	if !ok {
		log.WithField("module", module).
			Debug("not a Kafka producer, skipping poll loop")
		return
	}

	// inject module sekali saja
	baseLog := log.WithField("module", module)

	go func() {
		baseLog.Info("kafka producer poll loop started")

		for {
			select {
			case <-ctx.Done():
				baseLog.Info("kafka producer poll loop stopping")
				return

			case ev, ok := <-kp.Events():
				if !ok {
					baseLog.Warn("kafka events channel closed")
					return
				}

				switch e := ev.(type) {

				case *kafka.Message:
					fields := logrus.Fields{
						"event":     "kafka_delivery_report",
						"topic":     safeTopic(e.TopicPartition.Topic),
						"partition": e.TopicPartition.Partition,
						"offset":    e.TopicPartition.Offset,
					}

					if len(e.Key) > 0 {
						fields["key"] = string(e.Key)
					}

					if err := e.TopicPartition.Error; err != nil {
						baseLog.WithFields(fields).
							WithError(err).
							Error("kafka delivery failed")
					} else {
						baseLog.WithFields(fields).
							Debug("kafka message delivered")
					}

				case kafka.Error:
					baseLog.WithError(e).
						Error("kafka internal error")

				default:
					// ignore other events
				}
			}
		}
	}()
}

func safeTopic(topic *string) string {
	if topic == nil {
		return ""
	}
	return *topic
}
