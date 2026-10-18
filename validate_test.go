package main

import (
	"errors"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid", "Hello123!", false},
		{"empty", "", true},
		{"only spaces", "        ", true},
		{"too short", "Ab1!", true},
		{"too long", "Abcdefghij1234567890Abcdefghij!!", true},
		{"has space", "Hello 123!", true},
		{"cyrillic", "Привет123!", true},
		{"no uppercase", "hello123!", true},
		{"no symbol", "Hello1234", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidInput) {
					t.Errorf("want ErrInvalidInput, got %v", err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
