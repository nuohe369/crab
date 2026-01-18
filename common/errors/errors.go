package errors

import (
	"fmt"

	"github.com/nuohe369/crab/common/response"
)

// BizError represents a business error with error code
// BizError 表示带有错误码的业务错误
type BizError struct {
	Code response.Code
	Msg  string
	Err  error // Underlying error (optional) | 底层错误（可选）
}

// Error implements error interface
// Error 实现 error 接口
func (e *BizError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Msg, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

// Unwrap returns the underlying error
// Unwrap 返回底层错误
func (e *BizError) Unwrap() error {
	return e.Err
}

// New creates a new business error
// New 创建一个新的业务错误
func New(code response.Code, msg string) *BizError {
	return &BizError{
		Code: code,
		Msg:  msg,
	}
}

// Newf creates a new business error with formatted message
// Newf 创建一个带格式化消息的业务错误
func Newf(code response.Code, format string, args ...any) *BizError {
	return &BizError{
		Code: code,
		Msg:  fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with error code
// Wrap 用错误码包装一个已存在的错误
func Wrap(code response.Code, err error) *BizError {
	if err == nil {
		return nil
	}
	return &BizError{
		Code: code,
		Msg:  code.Msg(),
		Err:  err,
	}
}

// Wrapf wraps an existing error with error code and custom message
// Wrapf 用错误码和自定义消息包装一个已存在的错误
func Wrapf(code response.Code, err error, format string, args ...any) *BizError {
	if err == nil {
		return nil
	}
	return &BizError{
		Code: code,
		Msg:  fmt.Sprintf(format, args...),
		Err:  err,
	}
}

// Common error constructors for frequently used errors
// 常用错误的构造函数

// ErrUnauthorized creates an unauthorized error
// ErrUnauthorized 创建一个未授权错误
func ErrUnauthorized(msg ...string) *BizError {
	if len(msg) > 0 {
		return New(response.CodeUnauth, msg[0])
	}
	return New(response.CodeUnauth, response.CodeUnauth.Msg())
}

// ErrForbidden creates a forbidden error
// ErrForbidden 创建一个禁止访问错误
func ErrForbidden(msg ...string) *BizError {
	if len(msg) > 0 {
		return New(response.CodeForbid, msg[0])
	}
	return New(response.CodeForbid, response.CodeForbid.Msg())
}

// ErrParamInvalid creates a parameter invalid error
// ErrParamInvalid 创建一个参数无效错误
func ErrParamInvalid(msg ...string) *BizError {
	if len(msg) > 0 {
		return New(response.CodeParamInvalid, msg[0])
	}
	return New(response.CodeParamInvalid, response.CodeParamInvalid.Msg())
}

// ErrNotFound creates a not found error
// ErrNotFound 创建一个资源未找到错误
func ErrNotFound(msg ...string) *BizError {
	if len(msg) > 0 {
		return New(response.CodeNotFound, msg[0])
	}
	return New(response.CodeNotFound, response.CodeNotFound.Msg())
}

// ErrUserNotFound creates a user not found error
// ErrUserNotFound 创建一个用户未找到错误
func ErrUserNotFound() *BizError {
	return New(response.CodeUserNotFound, response.CodeUserNotFound.Msg())
}

// ErrServerError creates a server error
// ErrServerError 创建一个服务器错误
func ErrServerError(msg ...string) *BizError {
	if len(msg) > 0 {
		return New(response.CodeServerError, msg[0])
	}
	return New(response.CodeServerError, response.CodeServerError.Msg())
}

// ErrDBError creates a database error
// ErrDBError 创建一个数据库错误
func ErrDBError(err error) *BizError {
	return Wrap(response.CodeDBError, err)
}

// IsBizError checks if error is a business error
// IsBizError 检查错误是否为业务错误
func IsBizError(err error) bool {
	_, ok := err.(*BizError)
	return ok
}

// GetCode extracts error code from error
// GetCode 从错误中提取错误码
func GetCode(err error) response.Code {
	if bizErr, ok := err.(*BizError); ok {
		return bizErr.Code
	}
	return response.CodeError
}
