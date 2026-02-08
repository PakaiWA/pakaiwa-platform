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
 * @author KAnggara on Saturday 08/02/2026 10.57
 * @project pp
 * https://github.com/PakaiWA/pp/tree/main/cache/redis
 */

package redis

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestConfig_Creation(t *testing.T) {
	cfg := Config{
		Addr:         "localhost:6379",
		Password:     "secret",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	if cfg.Addr != "localhost:6379" {
		t.Errorf("Expected Addr to be localhost:6379, got %s", cfg.Addr)
	}

	if cfg.Password != "secret" {
		t.Errorf("Expected Password to be secret, got %s", cfg.Password)
	}

	if cfg.DB != 0 {
		t.Errorf("Expected DB to be 0, got %d", cfg.DB)
	}

	if cfg.DialTimeout != 5*time.Second {
		t.Errorf("Expected DialTimeout to be 5s, got %v", cfg.DialTimeout)
	}
}

func TestConfig_ZeroValues(t *testing.T) {
	cfg := Config{}

	if cfg.Addr != "" {
		t.Errorf("Expected empty Addr, got %s", cfg.Addr)
	}

	if cfg.Password != "" {
		t.Errorf("Expected empty Password, got %s", cfg.Password)
	}

	if cfg.DB != 0 {
		t.Errorf("Expected DB to be 0, got %d", cfg.DB)
	}

	if cfg.DialTimeout != 0 {
		t.Errorf("Expected DialTimeout to be 0, got %v", cfg.DialTimeout)
	}
}

func TestNewRedisClient_InvalidAddress(t *testing.T) {
	cfg := Config{
		Addr:         "invalid:99999",
		Password:     "",
		DB:           0,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	ctx := context.Background()
	client, err := NewRedisClient(ctx, cfg)

	if err == nil {
		if client != nil {
			_ = client.Close() //nolint:errcheck
		}
		t.Error("Expected error for invalid Redis address")
	}

	if client != nil {
		t.Error("Expected nil client on error")
	}
}

func TestNewRedisClient_ContextCancellation(t *testing.T) {
	cfg := Config{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client, err := NewRedisClient(ctx, cfg)

	if err == nil {
		if client != nil {
			_ = client.Close() //nolint:errcheck
		}
		t.Error("Expected error for cancelled context")
	}
}

func TestNewRedisClient_Success(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("Skipping: TEST_REDIS_URL not set")
	}

	cfg := Config{
		Addr:         redisURL,
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	ctx := context.Background()
	client, err := NewRedisClient(ctx, cfg)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	defer func() {
		_ = client.Close() //nolint:errcheck
	}()

	// Test basic operation
	err = client.Set(ctx, "test_key", "test_value", 10*time.Second).Err()
	if err != nil {
		t.Errorf("Failed to set key: %v", err)
	}

	val, err := client.Get(ctx, "test_key").Result()
	if err != nil {
		t.Errorf("Failed to get key: %v", err)
	}

	if val != "test_value" {
		t.Errorf("Expected test_value, got %s", val)
	}

	// Cleanup
	client.Del(ctx, "test_key")
}

func TestNewRedisClient_WithPassword(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL_WITH_AUTH")
	redisPassword := os.Getenv("TEST_REDIS_PASSWORD")

	if redisURL == "" || redisPassword == "" {
		t.Skip("Skipping: TEST_REDIS_URL_WITH_AUTH or TEST_REDIS_PASSWORD not set")
	}

	cfg := Config{
		Addr:         redisURL,
		Password:     redisPassword,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	ctx := context.Background()
	client, err := NewRedisClient(ctx, cfg)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	defer func() {
		_ = client.Close() //nolint:errcheck
	}()
}

func TestNewRedisClient_DifferentDB(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("Skipping: TEST_REDIS_URL not set")
	}

	cfg := Config{
		Addr:         redisURL,
		Password:     "",
		DB:           1, // Use DB 1 instead of 0
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	ctx := context.Background()
	client, err := NewRedisClient(ctx, cfg)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	defer func() {
		_ = client.Close() //nolint:errcheck
	}()
}

func TestNewRedisClient_CustomTimeouts(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("Skipping: TEST_REDIS_URL not set")
	}

	cfg := Config{
		Addr:         redisURL,
		Password:     "",
		DB:           0,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	ctx := context.Background()
	client, err := NewRedisClient(ctx, cfg)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	defer func() {
		_ = client.Close() //nolint:errcheck
	}()
}
