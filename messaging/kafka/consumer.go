/*
 * Copyright (c) 2025 KAnggara75
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * See <https://www.gnu.org/licenses/gpl-3.0.html>.
 *
 * @author KAnggara75 on Sat 06/09/25 11.04
 * @project PakaiWA kafka
 * https://github.com/PakaiWA/PakaiWA/tree/main/internal/pkg/kafka
 */

package kafka

import (
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type ConsumerConfig struct {
	Brokers []string
	GroupID string
	Options map[string]any
}

func NewKafkaConsumer(cfg ConsumerConfig) (*kafka.Consumer, error) {
	m := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.Brokers, ","),
		"group.id":          cfg.GroupID,
	}

	for k, v := range cfg.Options {
		_ = m.SetKey(k, v)
	}

	return kafka.NewConsumer(m)
}
