package main

import (
	"unicode"
)

type Processor interface {
	Process(password string) (Result, error)
}

type PasswordProcessor struct{}

func (p PasswordProcessor) Process(password string) (Result, error) {
	if err := ValidatePassword(password); err != nil {
		return Result{}, err
	}

	r := Result{Length: len([]rune(password))}

	for _, ch := range password {
		switch {
		case unicode.IsDigit(ch):
			r.HasDigit = true
		case unicode.IsUpper(ch):
			r.HasUpper = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			r.HasSymbol = true
		}
	}

	return r, nil
}
