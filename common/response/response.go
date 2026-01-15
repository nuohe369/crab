package response

import "github.com/gofiber/fiber/v2"

// Code represents a response code.
type Code int

// Common error codes
const (
	CodeSuccess Code = 0
	CodeError   Code = 1
)

// Authentication related codes (1000-1999)
const (
	CodeUnauth       Code = 1001 // unauthenticated
	CodeTokenExpired Code = 1002 // token expired
	CodeTokenInvalid Code = 1003 // token invalid
	CodeForbid       Code = 1004 // forbidden
)

// Parameter related codes (2000-2999)
const (
	CodeParamError   Code = 2001 // parameter error
	CodeParamMissing Code = 2002 // parameter missing
	CodeParamInvalid Code = 2003 // parameter invalid
)

// Resource related codes (3000-3999)
const (
	CodeNotFound  Code = 3001 // resource not found
	CodeDuplicate Code = 3002 // resource duplicate
)

// Business related codes (4000-4999)
const (
	CodeUserNotFound   Code = 4001 // user not found
	CodePasswordWrong  Code = 4002 // wrong password
	CodeUserDisabled   Code = 4003 // user disabled
	CodeUserExists     Code = 4004 // user already exists
	CodeBizError       Code = 4000 // general business error
	CodeAuthError      Code = 4005 // authentication error
)

// Organization related codes (4100-4199)
const (
	CodeOrgCodeExists        Code = 4100 // organization code already exists
	CodeOrgNotFound          Code = 4101 // organization not found
	CodeOrgPermissionDenied  Code = 4102 // permission denied
	CodeOrgMemberFull        Code = 4103 // organization member limit reached
	CodeOrgMemberExists      Code = 4104 // user is already a member
	CodeOrgInviteExpired     Code = 4105 // invitation expired
	CodeOrgInviteUsed        Code = 4106 // invitation already used
	CodeOrgInviteInvalid     Code = 4107 // invalid invitation code
	CodeOrgCannotRemoveOwner Code = 4108 // cannot remove owner
	CodeOrgOwnerCannotLeave  Code = 4109 // owner cannot leave
	CodeOrgTargetNotMember   Code = 4110 // target user is not a member
	CodeOrgCannotInviteOwner Code = 4111 // cannot invite as owner
	CodeOrgDisabled          Code = 4112 // organization disabled
	CodeOrgNotSelected       Code = 4113 // organization not selected
	CodeOrgAlreadyMember     Code = 4114 // user is already a member
	CodeOrgMemberLimit       Code = 4115 // member limit reached
	CodeOrgNotMember         Code = 4116 // not an organization member
	CodeOrgUserNotFound      Code = 4117 // user not found
)

// System related codes (5000-5999)
const (
	CodeServerError     Code = 5001 // server error
	CodeDBError         Code = 5002 // database error
	CodeRedisError      Code = 5003 // Redis error
	CodeTooManyRequests Code = 5004 // too many requests
)

// Error code message mapping
var codeMsg = map[Code]string{
	CodeSuccess:      "success",
	CodeError:        "error",
	CodeUnauth:       "Unauthenticated",
	CodeTokenExpired: "Token expired",
	CodeTokenInvalid: "Invalid token",
	CodeForbid:       "Forbidden",
	CodeParamError:   "Parameter error",
	CodeParamMissing: "Parameter missing",
	CodeParamInvalid: "Invalid parameter",
	CodeNotFound:     "Resource not found",
	CodeDuplicate:    "Resource duplicate",
	CodeUserNotFound:  "User not found",
	CodePasswordWrong: "Wrong password",
	CodeUserDisabled:  "User disabled",
	CodeUserExists:    "User already exists",
	CodeBizError:      "Business error",
	CodeAuthError:     "Authentication error",
	CodeServerError:     "Server error",
	CodeDBError:         "Database error",
	CodeRedisError:      "Redis error",
	CodeTooManyRequests: "Too many requests",
	// Organization related
	CodeOrgCodeExists:        "Organization code already exists",
	CodeOrgNotFound:          "Organization not found",
	CodeOrgPermissionDenied:  "Permission denied",
	CodeOrgMemberFull:        "Organization member limit reached",
	CodeOrgMemberExists:      "User is already a member",
	CodeOrgInviteExpired:     "Invitation expired",
	CodeOrgInviteUsed:        "Invitation already used",
	CodeOrgInviteInvalid:     "Invalid invitation code",
	CodeOrgCannotRemoveOwner: "Cannot remove owner",
	CodeOrgOwnerCannotLeave:  "Owner cannot leave",
	CodeOrgTargetNotMember:   "Target user is not a member",
	CodeOrgCannotInviteOwner: "Cannot invite as owner",
	CodeOrgDisabled:          "Organization disabled",
	CodeOrgNotSelected:       "Organization not selected",
	CodeOrgAlreadyMember:     "User is already a member",
	CodeOrgMemberLimit:       "Member limit reached",
	CodeOrgNotMember:         "Not an organization member",
	CodeOrgUserNotFound:      "User not found",
}

// Msg returns the message for the error code.
func (c Code) Msg() string {
	if msg, ok := codeMsg[c]; ok {
		return msg
	}
	return "Unknown error"
}

// Response represents a unified response structure.
type Response struct {
	Code Code   `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// PageData represents paginated data.
type PageData struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}

// OK returns a successful response.
func OK(c *fiber.Ctx, data any) error {
	return c.JSON(Response{
		Code: CodeSuccess,
		Msg:  CodeSuccess.Msg(),
		Data: data,
	})
}

// Fail returns a failure response with a custom message.
func Fail(c *fiber.Ctx, msg string) error {
	return c.JSON(Response{
		Code: CodeError,
		Msg:  msg,
	})
}

// FailCode returns a failure response using the default message for the error code.
func FailCode(c *fiber.Ctx, code Code) error {
	return c.JSON(Response{
		Code: code,
		Msg:  code.Msg(),
	})
}

// FailCodeMsg returns a failure response with an error code and custom message.
func FailCodeMsg(c *fiber.Ctx, code Code, msg string) error {
	return c.JSON(Response{
		Code: code,
		Msg:  msg,
	})
}

// Page returns a paginated response.
func Page(c *fiber.Ctx, list any, total int64, page, size int) error {
	return OK(c, PageData{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// FailMsg is an alias for FailCodeMsg.
func FailMsg(c *fiber.Ctx, code Code, msg string) error {
	return FailCodeMsg(c, code, msg)
}

// OKList returns a paginated list response.
func OKList(c *fiber.Ctx, list any, total int64, page, size int) error {
	return OK(c, PageData{
		List:  list,
		Total: total,
		Page:  page,
		Size:  size,
	})
}
