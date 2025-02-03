package errors

import (
	"errors"
	"fmt"
	"testing"
)

type CustomError struct {
	Err error
}

func (e *CustomError) Error() string { return e.Err.Error() }

func IsCustomError(err error) bool {
	var asError *CustomError
	return errors.As(err, &asError)
}

func TestWrapping(t *testing.T) {
	t.Parallel()
	originalErr := errors.New("original error")
	err1 := NewError(originalErr)
	if !errors.Is(err1, originalErr) {
		t.Errorf("original error is not wrapped")
	}
	if !Is(err1, originalErr) {
		t.Errorf("original error is not wrapped")
	}
	unwrappedErr := errors.Unwrap(err1)
	if !errors.Is(unwrappedErr, originalErr) {
		t.Errorf("original error is not wrapped")
	}
	if !Is(unwrappedErr, originalErr) {
		t.Errorf("original error is not wrapped")
	}
	unwrappedErr = Unwrap(err1)
	if !errors.Is(unwrappedErr, originalErr) {
		t.Errorf("original error is not wrapped")
	}
	if !Is(unwrappedErr, originalErr) {
		t.Errorf("original error is not wrapped")
	}
	wrappedErr := fmt.Errorf("wrapped: %w", originalErr)
	if !errors.Is(wrappedErr, originalErr) {
		t.Errorf("original error is not wrapped")
	}
	if !Is(wrappedErr, originalErr) {
		t.Errorf("original error is not wrapped")
	}
}

func TestNewError(t *testing.T) {
	t.Parallel()
	originalErr := &CustomError{errors.New("err1")}
	if !IsCustomError(originalErr) {
		t.Errorf("TestNewError fail #1")
	}
	err1 := NewError(originalErr)
	if !IsCustomError(err1) {
		t.Errorf("TestNewError fail #2")
	}
	wrappedErr1 := fmt.Errorf("wrapped %w", err1)
	if !IsCustomError(wrappedErr1) {
		t.Errorf("TestNewError fail #3")
	}
	unwrappedErr1 := Unwrap(wrappedErr1)
	if !IsCustomError(unwrappedErr1) {
		t.Errorf("TestNewError fail #4")
	}
}
