package response

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestCodeMsg(t *testing.T) {
	tests := []struct {
		code     Code
		expected string
	}{
		{CodeSuccess, "success"},
		{CodeError, "error"},
		{CodeUnauth, "Unauthenticated"},
		{CodeTokenExpired, "Token expired"},
		{CodeParamError, "Parameter error"},
		{CodeNotFound, "Resource not found"},
		{CodeUserNotFound, "User not found"},
		{CodeServerError, "Server error"},
		{Code(9999), "Unknown error"},
	}

	for _, tt := range tests {
		msg := tt.code.Msg()
		if msg != tt.expected {
			t.Errorf("Code(%d).Msg() = %s, want %s", tt.code, msg, tt.expected)
		}
	}
}

func TestOK(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return OK(c, fiber.Map{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Code != CodeSuccess {
		t.Errorf("Expected code %d, got %d", CodeSuccess, result.Code)
	}

	if result.Msg != "success" {
		t.Errorf("Expected msg 'success', got '%s'", result.Msg)
	}
}

func TestFail(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return Fail(c, "custom error message")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Code != CodeError {
		t.Errorf("Expected code %d, got %d", CodeError, result.Code)
	}

	if result.Msg != "custom error message" {
		t.Errorf("Expected msg 'custom error message', got '%s'", result.Msg)
	}
}

func TestFailCode(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return FailCode(c, CodeUnauth)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Code != CodeUnauth {
		t.Errorf("Expected code %d, got %d", CodeUnauth, result.Code)
	}

	if result.Msg != CodeUnauth.Msg() {
		t.Errorf("Expected msg '%s', got '%s'", CodeUnauth.Msg(), result.Msg)
	}
}

func TestFailCodeMsg(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		return FailCodeMsg(c, CodeParamError, "custom param error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	var result Response
	json.Unmarshal(body, &result)

	if result.Code != CodeParamError {
		t.Errorf("Expected code %d, got %d", CodeParamError, result.Code)
	}

	if result.Msg != "custom param error" {
		t.Errorf("Expected msg 'custom param error', got '%s'", result.Msg)
	}
}

func TestPage(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		list := []string{"item1", "item2", "item3"}
		return Page(c, list, 100, 1, 10)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Code Code     `json:"code"`
		Msg  string   `json:"msg"`
		Data PageData `json:"data"`
	}
	json.Unmarshal(body, &result)

	if result.Code != CodeSuccess {
		t.Errorf("Expected code %d, got %d", CodeSuccess, result.Code)
	}

	if result.Data.Total != 100 {
		t.Errorf("Expected total 100, got %d", result.Data.Total)
	}

	if result.Data.Page != 1 {
		t.Errorf("Expected page 1, got %d", result.Data.Page)
	}

	if result.Data.Size != 10 {
		t.Errorf("Expected size 10, got %d", result.Data.Size)
	}
}
