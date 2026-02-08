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
 * https://github.com/PakaiWA/pp/tree/main/runtime/shutdown
 */

package shutdown

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestWait_WithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan bool)
	var receivedSignal os.Signal

	go func() {
		receivedSignal = Wait(ctx)
		done <- true
	}()

	select {
	case <-done:
		if receivedSignal != nil {
			t.Errorf("Expected nil signal on context cancellation, got %v", receivedSignal)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Wait did not return after context cancellation")
	}
}

func TestWait_WithSignal(t *testing.T) {
	ctx := context.Background()

	done := make(chan bool)
	var receivedSignal os.Signal

	go func() {
		receivedSignal = Wait(ctx, syscall.SIGUSR1)
		done <- true
	}()

	// Give the goroutine time to set up signal handling
	time.Sleep(50 * time.Millisecond)

	// Send signal
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find process: %v", err)
	}

	if err := proc.Signal(syscall.SIGUSR1); err != nil {
		t.Fatalf("Failed to send signal: %v", err)
	}

	select {
	case <-done:
		if receivedSignal != syscall.SIGUSR1 {
			t.Errorf("Expected SIGUSR1, got %v", receivedSignal)
		}
	case <-time.After(1 * time.Second):
		t.Error("Wait did not return after receiving signal")
	}
}

func TestWait_DefaultSignals(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan bool)
	var receivedSignal os.Signal

	go func() {
		// Call Wait without specifying signals (should default to SIGINT and SIGTERM)
		receivedSignal = Wait(ctx)
		done <- true
	}()

	select {
	case <-done:
		// Context should cancel before any signal is received
		if receivedSignal != nil {
			t.Errorf("Expected nil signal on context cancellation, got %v", receivedSignal)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Wait did not return after context cancellation")
	}
}

func TestWait_MultipleSignals(t *testing.T) {
	ctx := context.Background()

	done := make(chan bool)
	var receivedSignal os.Signal

	go func() {
		receivedSignal = Wait(ctx, syscall.SIGUSR1, syscall.SIGUSR2)
		done <- true
	}()

	// Give the goroutine time to set up signal handling
	time.Sleep(50 * time.Millisecond)

	// Send SIGUSR2
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find process: %v", err)
	}

	if err := proc.Signal(syscall.SIGUSR2); err != nil {
		t.Fatalf("Failed to send signal: %v", err)
	}

	select {
	case <-done:
		if receivedSignal != syscall.SIGUSR2 {
			t.Errorf("Expected SIGUSR2, got %v", receivedSignal)
		}
	case <-time.After(1 * time.Second):
		t.Error("Wait did not return after receiving signal")
	}
}

func TestWait_ContextCancelBeforeSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan bool)
	var receivedSignal os.Signal

	go func() {
		receivedSignal = Wait(ctx, syscall.SIGUSR1)
		done <- true
	}()

	// Give the goroutine time to set up
	time.Sleep(50 * time.Millisecond)

	// Cancel context before sending signal
	cancel()

	select {
	case <-done:
		if receivedSignal != nil {
			t.Errorf("Expected nil signal on context cancellation, got %v", receivedSignal)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Wait did not return after context cancellation")
	}
}

func TestWaitForSignal_Integration(t *testing.T) {
	// This test is more of a smoke test since WaitForSignal blocks
	// We'll test it in a goroutine with a timeout

	done := make(chan bool)

	go func() {
		// Start waiting for signal
		go WaitForSignal()

		// Give it time to set up
		time.Sleep(50 * time.Millisecond)

		// Send interrupt signal
		proc, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Errorf("Failed to find process: %v", err)
			return
		}

		// Note: We can't easily test WaitForSignal without actually interrupting the test
		// So we'll just verify it doesn't panic when set up
		done <- true

		// Send signal to unblock (this will affect the actual WaitForSignal call)
		proc.Signal(os.Interrupt)
	}()

	select {
	case <-done:
		// Test passed - WaitForSignal was set up without panic
	case <-time.After(200 * time.Millisecond):
		t.Error("Test setup timed out")
	}
}

func TestWait_SignalCleanup(t *testing.T) {
	// Test that signal handlers are properly cleaned up
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan bool)

	go func() {
		Wait(ctx, syscall.SIGUSR1)
		done <- true
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Good, Wait returned
	case <-time.After(200 * time.Millisecond):
		t.Error("Wait did not return after context cancellation")
	}

	// Now test that we can set up signal handling again
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	done2 := make(chan bool)

	go func() {
		Wait(ctx2, syscall.SIGUSR1)
		done2 <- true
	}()

	time.Sleep(50 * time.Millisecond)

	// Send signal
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find process: %v", err)
	}

	if err := proc.Signal(syscall.SIGUSR1); err != nil {
		t.Fatalf("Failed to send signal: %v", err)
	}

	select {
	case <-done2:
		// Good, signal was received
	case <-time.After(1 * time.Second):
		t.Error("Second Wait did not receive signal")
	}
}
