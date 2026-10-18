package main

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 30
)

func ValidatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("%w: password is empty", ErrInvalidInput)
	}

	length := len([]rune(password))
	if length < MinPasswordLength {
		return fmt.Errorf("%w: password is shorter than %d characters", ErrInvalidInput, MinPasswordLength)
	}
	if length > MaxPasswordLength {
		return fmt.Errorf("%w: password is longer than %d characters", ErrInvalidInput, MaxPasswordLength)
	}

	hasUpper := false
	hasSymbol := false

	for _, ch := range password {
		if ch == ' ' {
			return fmt.Errorf("%w: password must not contain spaces", ErrInvalidInput)
		}
		if ch > unicode.MaxASCII || !unicode.IsPrint(ch) {
			return fmt.Errorf("%w: password must contain only Latin letters, digits and symbols", ErrInvalidInput)
		}
		if unicode.IsUpper(ch) {
			hasUpper = true
		}
		if unicode.IsPunct(ch) || unicode.IsSymbol(ch) {
			hasSymbol = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("%w: password must contain an uppercase letter", ErrInvalidInput)
	}
	if !hasSymbol {
		return fmt.Errorf("%w: password must contain a symbol", ErrInvalidInput)
	}

	return nil
}
