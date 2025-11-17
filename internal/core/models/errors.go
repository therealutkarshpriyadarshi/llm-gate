package models

import (
	"errors"
	"fmt"
)

// Common validation errors
var (
	ErrInvalidModel            = errors.New("invalid model")
	ErrNoMessages              = errors.New("no messages provided")
	ErrInvalidTemperature      = errors.New("temperature must be between 0 and 2.0")
	ErrInvalidTopP             = errors.New("top_p must be between 0 and 1.0")
	ErrInvalidPresencePenalty  = errors.New("presence_penalty must be between -2.0 and 2.0")
	ErrInvalidFrequencyPenalty = errors.New("frequency_penalty must be between -2.0 and 2.0")
)

// ValidationError represents a validation error
type ValidationError struct {
	Field string
	Index int
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Index >= 0 {
		return fmt.Sprintf("validation error in %s[%d]: %s", e.Field, e.Index, e.Message)
	}
	return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field string, index int, message string) error {
	return &ValidationError{
		Field:   field,
		Index:   index,
		Message: message,
	}
}
