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
 * https://github.com/PakaiWA/pp/tree/main/http/client
 */

package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetClient_Singleton(t *testing.T) {
	client1 := getClient()
	client2 := getClient()

	if client1 != client2 {
		t.Error("Expected getClient to return the same instance (singleton)")
	}

	if client1.Timeout != 5*time.Second {
		t.Errorf("Expected timeout to be 5 seconds, got %v", client1.Timeout)
	}
}

func TestGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	ctx := context.Background()
	resp, err := Get(ctx, server.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "success" {
		t.Errorf("Expected 'success', got %s", string(body))
	}
}

func TestGet_InvalidURL(t *testing.T) {
	ctx := context.Background()
	_, err := Get(ctx, "://invalid-url")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestGet_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := Get(ctx, server.URL)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

func TestPost_Success(t *testing.T) {
	type TestPayload struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		var payload TestPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if payload.Name != "John" || payload.Email != "john@example.com" {
			t.Errorf("Unexpected payload: %+v", payload)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	ctx := context.Background()
	payload := TestPayload{Name: "John", Email: "john@example.com"}

	resp, err := Post(ctx, server.URL, payload)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestPut_Success(t *testing.T) {
	type TestPayload struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		var payload TestPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	ctx := context.Background()
	payload := TestPayload{ID: 1, Name: "Updated"}

	resp, err := Put(ctx, server.URL, payload)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestPatch_Success(t *testing.T) {
	type TestPayload struct {
		Name string `json:"name"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}

		var payload TestPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	ctx := context.Background()
	payload := TestPayload{Name: "Patched"}

	resp, err := Patch(ctx, server.URL, payload)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestDoJSON_InvalidJSON(t *testing.T) {
	// Create a type that can't be marshaled to JSON
	type InvalidPayload struct {
		Channel chan int // channels can't be marshaled to JSON
	}

	ctx := context.Background()
	payload := InvalidPayload{Channel: make(chan int)}

	_, err := Post(ctx, "http://example.com", payload)
	if err == nil {
		t.Error("Expected error for invalid JSON payload")
	}
}

func TestDoJSON_EmptyPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "{}" {
			t.Errorf("Expected empty object {}, got %s", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	type EmptyPayload struct{}

	resp, err := Post(ctx, server.URL, EmptyPayload{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestDoJSON_ComplexPayload(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type User struct {
		Name    string   `json:"name"`
		Age     int      `json:"age"`
		Tags    []string `json:"tags"`
		Address Address  `json:"address"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		if user.Name != "John" {
			t.Errorf("Expected name John, got %s", user.Name)
		}
		if len(user.Tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(user.Tags))
		}
		if user.Address.City != "NYC" {
			t.Errorf("Expected city NYC, got %s", user.Address.City)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	payload := User{
		Name: "John",
		Age:  30,
		Tags: []string{"developer", "golang"},
		Address: Address{
			Street: "123 Main St",
			City:   "NYC",
		},
	}

	resp, err := Post(ctx, server.URL, payload)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestDoJSON_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	ctx := context.Background()
	type Payload struct {
		Data string `json:"data"`
	}

	resp, err := Post(ctx, server.URL, Payload{Data: "test"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}
