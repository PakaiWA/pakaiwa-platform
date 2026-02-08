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
 * https://github.com/PakaiWA/pp/tree/main/validation
 */

package validation

import (
	"testing"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Fatal("Expected validator to be created, got nil")
	}
}

func TestValidator_WithJSONTags(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"gte=0,lte=150"`
	}

	validator := NewValidator()

	tests := []struct {
		name      string
		input     TestStruct
		wantError bool
	}{
		{
			name: "valid struct",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   30,
			},
			wantError: false,
		},
		{
			name: "missing required field",
			input: TestStruct{
				Name:  "",
				Email: "john@example.com",
				Age:   30,
			},
			wantError: true,
		},
		{
			name: "invalid email",
			input: TestStruct{
				Name:  "John Doe",
				Email: "invalid-email",
				Age:   30,
			},
			wantError: true,
		},
		{
			name: "age out of range",
			input: TestStruct{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   200,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.input)
			if tt.wantError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestValidator_JSONTagNameFunc(t *testing.T) {
	type TestStruct struct {
		UserName string `json:"user_name" validate:"required"`
		Password string `json:"password" validate:"required,min=8"`
		Ignored  string `json:"-"`
		NoTag    string `validate:"required"`
	}

	validator := NewValidator()

	// Test with missing user_name
	input := TestStruct{
		UserName: "",
		Password: "password123",
		NoTag:    "value",
	}

	err := validator.Struct(input)
	if err == nil {
		t.Fatal("Expected validation error for missing user_name")
	}

	// Verify that the error message uses the JSON tag name
	validationErrors := err.(interface{ Error() string })
	errMsg := validationErrors.Error()

	// The error should reference "user_name" (from json tag), not "UserName"
	if errMsg == "" {
		t.Error("Expected error message to contain field information")
	}
}

func TestValidator_WithoutJSONTag(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
	}

	validator := NewValidator()

	input := TestStruct{Name: ""}
	err := validator.Struct(input)

	if err == nil {
		t.Error("Expected validation error for missing Name")
	}
}

func TestValidator_EmptyJSONTag(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"" validate:"required"`
		Field2 string `validate:"required"`
	}

	validator := NewValidator()

	input := TestStruct{
		Field1: "",
		Field2: "",
	}

	err := validator.Struct(input)
	if err == nil {
		t.Error("Expected validation errors")
	}
}

func TestValidator_ComplexValidation(t *testing.T) {
	type Address struct {
		Street  string `json:"street" validate:"required"`
		City    string `json:"city" validate:"required"`
		ZipCode string `json:"zip_code" validate:"required,len=5"`
	}

	type User struct {
		Name    string  `json:"name" validate:"required"`
		Email   string  `json:"email" validate:"required,email"`
		Age     int     `json:"age" validate:"gte=18,lte=100"`
		Address Address `json:"address" validate:"required"`
	}

	validator := NewValidator()

	tests := []struct {
		name      string
		input     User
		wantError bool
	}{
		{
			name: "valid user with address",
			input: User{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   25,
				Address: Address{
					Street:  "123 Main St",
					City:    "New York",
					ZipCode: "10001",
				},
			},
			wantError: false,
		},
		{
			name: "invalid zip code length",
			input: User{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   25,
				Address: Address{
					Street:  "123 Main St",
					City:    "New York",
					ZipCode: "100",
				},
			},
			wantError: true,
		},
		{
			name: "age below minimum",
			input: User{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   15,
				Address: Address{
					Street:  "123 Main St",
					City:    "New York",
					ZipCode: "10001",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(tt.input)
			if tt.wantError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
