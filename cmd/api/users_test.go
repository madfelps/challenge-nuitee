package main

import (
	"testing"

	"github.com/madfelps/challenge-nuitee/internal/validator"
)

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name     string
		user     CreateUserRequest
		expected bool
	}{
		{
			name: "valid user",
			user: CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			expected: true,
		},
		{
			name: "empty name",
			user: CreateUserRequest{
				Name:     "",
				Email:    "john@example.com",
				Password: "password123",
			},
			expected: false,
		},
		{
			name: "name too long",
			user: CreateUserRequest{
				Name:     "This is a very long name that exceeds the maximum allowed length of 100 characters and should fail validation",
				Email:    "john@example.com",
				Password: "password123",
			},
			expected: false,
		},
		{
			name: "invalid email",
			user: CreateUserRequest{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			expected: false,
		},
		{
			name: "password too short",
			user: CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "123",
			},
			expected: false,
		},
		{
			name: "password too long",
			user: CreateUserRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "this_is_a_very_long_password_that_exceeds_the_maximum_allowed_length_of_72_characters_and_should_fail_validation",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			validateUser(v, &tt.user)
			ValidateEmail(v, tt.user.Email)
			ValidatePasswordPlaintext(v, tt.user.Password)

			if v.Valid() != tt.expected {
				t.Errorf("validateUser() = %v, expected %v. Errors: %v", v.Valid(), tt.expected, v.Errors)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "valid email",
			email:    "test@example.com",
			expected: true,
		},
		{
			name:     "valid email with subdomain",
			email:    "user@mail.example.com",
			expected: true,
		},
		{
			name:     "valid email with numbers",
			email:    "user123@example123.com",
			expected: true,
		},
		{
			name:     "empty email",
			email:    "",
			expected: false,
		},
		{
			name:     "email without @",
			email:    "testexample.com",
			expected: false,
		},
		{
			name:     "email without domain",
			email:    "test@",
			expected: false,
		},
		{
			name:     "email without local part",
			email:    "@example.com",
			expected: false,
		},
		{
			name:     "email with spaces",
			email:    "test @example.com",
			expected: false,
		},
		{
			name:     "email with multiple @",
			email:    "test@@example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateEmail(v, tt.email)

			if v.Valid() != tt.expected {
				t.Errorf("ValidateEmail() = %v, expected %v. Errors: %v", v.Valid(), tt.expected, v.Errors)
			}
		})
	}
}

func TestValidatePasswordPlaintext(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "valid password",
			password: "password123",
			expected: true,
		},
		{
			name:     "valid password with special characters",
			password: "P@ssw0rd!",
			expected: true,
		},
		{
			name:     "valid password with minimum length",
			password: "12345678",
			expected: true,
		},
		{
			name:     "valid password with maximum length",
			password: "this_is_a_password_with_exactly_72_characters_long_and_should_pass_val!",
			expected: true,
		},
		{
			name:     "empty password",
			password: "",
			expected: false,
		},
		{
			name:     "password too short",
			password: "1234567",
			expected: false,
		},
		{
			name:     "password too long",
			password: "this_is_a_password_that_exceeds_the_maximum_allowed_length_of_72_characters_and_should_fail_validation",
			expected: false,
		},
		{
			name:     "password with only spaces",
			password: "        ",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidatePasswordPlaintext(v, tt.password)

			if v.Valid() != tt.expected {
				t.Errorf("ValidatePasswordPlaintext() = %v, expected %v. Errors: %v", v.Valid(), tt.expected, v.Errors)
			}
		})
	}
}
