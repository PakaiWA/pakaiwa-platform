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

package httpserver

import (
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if !opts.TrustProxy {
		t.Error("Expected TrustProxy to be true")
	}

	if !opts.EnableIPValidation {
		t.Error("Expected EnableIPValidation to be true")
	}

	if opts.AppName != "" {
		t.Errorf("Expected AppName to be empty, got %s", opts.AppName)
	}

	if opts.ErrorHandler != nil {
		t.Error("Expected ErrorHandler to be nil")
	}

	if opts.TrustedProxies != nil {
		t.Error("Expected TrustedProxies to be nil")
	}
}

func TestNewFiber_WithDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	app := NewFiber(opts)

	if app == nil {
		t.Fatal("Expected non-nil fiber app")
	}
}

func TestNewFiber_WithCustomOptions(t *testing.T) {
	opts := Options{
		AppName:            "TestApp",
		TrustProxy:         true,
		EnableIPValidation: false,
		TrustedProxies:     []string{"127.0.0.1", "10.0.0.1"},
	}

	app := NewFiber(opts)

	if app == nil {
		t.Fatal("Expected non-nil fiber app")
	}
}

func TestNewFiber_EmptyOptions(t *testing.T) {
	app := NewFiber(Options{})

	if app == nil {
		t.Fatal("Expected non-nil fiber app with empty options")
	}
}

func TestNewFiber_WithNoTrustProxy(t *testing.T) {
	opts := Options{
		TrustProxy:         false,
		EnableIPValidation: false,
	}

	app := NewFiber(opts)

	if app == nil {
		t.Fatal("Expected non-nil fiber app")
	}
}

func TestOptions_Fields(t *testing.T) {
	proxies := []string{"192.168.0.1", "10.0.0.0/8"}
	opts := Options{
		AppName:            "MyApp",
		TrustProxy:         true,
		EnableIPValidation: true,
		TrustedProxies:     proxies,
	}

	if opts.AppName != "MyApp" {
		t.Errorf("Expected AppName 'MyApp', got %s", opts.AppName)
	}

	if !opts.TrustProxy {
		t.Error("Expected TrustProxy to be true")
	}

	if !opts.EnableIPValidation {
		t.Error("Expected EnableIPValidation to be true")
	}

	if len(opts.TrustedProxies) != 2 {
		t.Errorf("Expected 2 trusted proxies, got %d", len(opts.TrustedProxies))
	}
}
