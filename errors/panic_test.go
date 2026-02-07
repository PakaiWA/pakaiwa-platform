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
 * @author KAnggara on Saturday 07/02/2026 18.39
 * @project pp
 * https://github.com/PakaiWA/PakaiWA/tree/main/~/work/PakaiWA/pp/errors
 */

package errors

import (
	"errors"
	"testing"
)

func TestMust_Success(t *testing.T) {
	result := Must(42, nil)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

func TestMust_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't panic")
		}
	}()

	Must(0, errors.New("test error"))
}

func TestMust_WithString(t *testing.T) {
	result := Must("hello", nil)
	if result != "hello" {
		t.Errorf("Expected 'hello', got %s", result)
	}
}

func TestMust_WithStruct(t *testing.T) {
	type TestStruct struct {
		Value int
	}

	expected := TestStruct{Value: 100}
	result := Must(expected, nil)
	if result.Value != expected.Value {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCheck_NoError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	Check(nil)
}

func TestCheck_WithError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't panic")
		}
	}()

	Check(errors.New("test error"))
}

func TestPanicIfError_NoError(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Unexpected panic: %v", r)
		}
	}()

	PanicIfError(nil)
}

func TestPanicIfError_WithError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic, but didn't panic")
		}
	}()

	PanicIfError(errors.New("test error"))
}
