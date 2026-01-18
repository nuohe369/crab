package errors

import (
	"errors"
	"strings"
	"testing"

	"github.com/nuohe369/crab/common/response"
)

func TestNew(t *testing.T) {
	err := New(response.CodeParamError, "invalid parameter")
	if err.Code != response.CodeParamError {
		t.Errorf("Expected code %d, got %d", response.CodeParamError, err.Code)
	}

	if err.Msg != "invalid parameter" {
		t.Errorf("Expected msg 'invalid parameter', got '%s'", err.Msg)
	}

	if err.Err != nil {
		t.Error("Expected nil underlying error")
	}
}

func TestNewf(t *testing.T) {
	err := Newf(response.CodeParamError, "invalid parameter: %s", "username")
	if err.Msg != "invalid parameter: username" {
		t.Errorf("Expected formatted message, got '%s'", err.Msg)
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("database connection failed")
	err := Wrap(response.CodeDBError, originalErr)

	if err.Code != response.CodeDBError {
		t.Errorf("Expected code %d, got %d", response.CodeDBError, err.Code)
	}

	if err.Err != originalErr {
		t.Error("Expected wrapped error to match original")
	}

	if !strings.Contains(err.Error(), "database connection failed") {
		t.Errorf("Error message should contain original error: %s", err.Error())
	}
}

func TestWrapNil(t *testing.T) {
	err := Wrap(response.CodeDBError, nil)
	if err != nil {
		t.Error("Wrapping nil error should return nil")
	}
}

func TestWrapf(t *testing.T) {
	originalErr := errors.New("connection timeout")
	err := Wrapf(response.CodeDBError, originalErr, "failed to connect to %s", "database")

	if err.Msg != "failed to connect to database" {
		t.Errorf("Expected custom message, got '%s'", err.Msg)
	}

	if err.Err != originalErr {
		t.Error("Expected wrapped error to match original")
	}
}

func TestErrorMethod(t *testing.T) {
	// Error without underlying error
	err1 := New(response.CodeParamError, "invalid input")
	expected1 := "[2001] invalid input"
	if err1.Error() != expected1 {
		t.Errorf("Expected '%s', got '%s'", expected1, err1.Error())
	}

	// Error with underlying error
	originalErr := errors.New("original")
	err2 := Wrap(response.CodeDBError, originalErr)
	if !strings.Contains(err2.Error(), "original") {
		t.Errorf("Error message should contain underlying error: %s", err2.Error())
	}
}

func TestUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := Wrap(response.CodeDBError, originalErr)

	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Error("Unwrap should return original error")
	}
}

func TestErrUnauthorized(t *testing.T) {
	// Default message
	err1 := ErrUnauthorized()
	if err1.Code != response.CodeUnauth {
		t.Errorf("Expected code %d, got %d", response.CodeUnauth, err1.Code)
	}

	// Custom message
	err2 := ErrUnauthorized("custom auth error")
	if err2.Msg != "custom auth error" {
		t.Errorf("Expected custom message, got '%s'", err2.Msg)
	}
}

func TestErrForbidden(t *testing.T) {
	err := ErrForbidden()
	if err.Code != response.CodeForbid {
		t.Errorf("Expected code %d, got %d", response.CodeForbid, err.Code)
	}
}

func TestErrParamInvalid(t *testing.T) {
	err := ErrParamInvalid("username is required")
	if err.Code != response.CodeParamInvalid {
		t.Errorf("Expected code %d, got %d", response.CodeParamInvalid, err.Code)
	}
	if err.Msg != "username is required" {
		t.Errorf("Expected custom message, got '%s'", err.Msg)
	}
}

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound()
	if err.Code != response.CodeNotFound {
		t.Errorf("Expected code %d, got %d", response.CodeNotFound, err.Code)
	}
}

func TestErrUserNotFound(t *testing.T) {
	err := ErrUserNotFound()
	if err.Code != response.CodeUserNotFound {
		t.Errorf("Expected code %d, got %d", response.CodeUserNotFound, err.Code)
	}
}

func TestErrServerError(t *testing.T) {
	err := ErrServerError("internal error")
	if err.Code != response.CodeServerError {
		t.Errorf("Expected code %d, got %d", response.CodeServerError, err.Code)
	}
}

func TestErrDBError(t *testing.T) {
	originalErr := errors.New("connection failed")
	err := ErrDBError(originalErr)
	if err.Code != response.CodeDBError {
		t.Errorf("Expected code %d, got %d", response.CodeDBError, err.Code)
	}
	if err.Err != originalErr {
		t.Error("Expected wrapped error to match original")
	}
}

func TestIsBizError(t *testing.T) {
	bizErr := New(response.CodeParamError, "test")
	if !IsBizError(bizErr) {
		t.Error("Expected IsBizError to return true for BizError")
	}

	stdErr := errors.New("standard error")
	if IsBizError(stdErr) {
		t.Error("Expected IsBizError to return false for standard error")
	}
}

func TestGetCode(t *testing.T) {
	bizErr := New(response.CodeParamError, "test")
	code := GetCode(bizErr)
	if code != response.CodeParamError {
		t.Errorf("Expected code %d, got %d", response.CodeParamError, code)
	}

	stdErr := errors.New("standard error")
	code = GetCode(stdErr)
	if code != response.CodeError {
		t.Errorf("Expected default code %d, got %d", response.CodeError, code)
	}
}
