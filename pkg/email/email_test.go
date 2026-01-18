package email

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user@example.com",
		Password: "password",
		From:     "sender@example.com",
		FromName: "Test Sender",
		UseTLS:   true,
	}

	client := New(cfg)
	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.config.Host != cfg.Host {
		t.Errorf("Expected host %s, got %s", cfg.Host, client.config.Host)
	}
}

func TestBuildMessage(t *testing.T) {
	client := New(Config{
		From:     "sender@example.com",
		FromName: "Test Sender",
	})

	msg := &Message{
		To:      []string{"recipient@example.com"},
		Cc:      []string{"cc@example.com"},
		Subject: "Test Subject",
		Body:    "Test Body",
		IsHTML:  false,
	}

	content := client.buildMessage(msg)
	contentStr := string(content)

	// Check required headers
	if !strings.Contains(contentStr, "From: Test Sender <sender@example.com>") {
		t.Error("Missing or incorrect From header")
	}

	if !strings.Contains(contentStr, "To: recipient@example.com") {
		t.Error("Missing or incorrect To header")
	}

	if !strings.Contains(contentStr, "Cc: cc@example.com") {
		t.Error("Missing or incorrect Cc header")
	}

	if !strings.Contains(contentStr, "Subject: Test Subject") {
		t.Error("Missing or incorrect Subject header")
	}

	if !strings.Contains(contentStr, "Test Body") {
		t.Error("Missing body content")
	}

	if !strings.Contains(contentStr, "Content-Type: text/plain; charset=UTF-8") {
		t.Error("Missing or incorrect Content-Type header")
	}
}

func TestBuildMessageHTML(t *testing.T) {
	client := New(Config{
		From: "sender@example.com",
	})

	msg := &Message{
		To:      []string{"recipient@example.com"},
		Subject: "HTML Test",
		Body:    "<h1>Hello</h1>",
		IsHTML:  true,
	}

	content := client.buildMessage(msg)
	contentStr := string(content)

	if !strings.Contains(contentStr, "Content-Type: text/html; charset=UTF-8") {
		t.Error("Expected HTML content type")
	}

	if !strings.Contains(contentStr, "<h1>Hello</h1>") {
		t.Error("Missing HTML body content")
	}
}

func TestBuildMessageWithoutFromName(t *testing.T) {
	client := New(Config{
		From: "sender@example.com",
	})

	msg := &Message{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Body",
	}

	content := client.buildMessage(msg)
	contentStr := string(content)

	if !strings.Contains(contentStr, "From: sender@example.com\r\n") {
		t.Error("Expected simple From header without name")
	}
}

func TestSendEmptyRecipients(t *testing.T) {
	client := New(Config{
		Host: "smtp.example.com",
		Port: 587,
		From: "sender@example.com",
	})

	msg := &Message{
		To:      []string{},
		Subject: "Test",
		Body:    "Body",
	}

	err := client.Send(msg)
	if err == nil {
		t.Error("Expected error when sending to empty recipients")
	}

	if !strings.Contains(err.Error(), "recipient cannot be empty") {
		t.Errorf("Expected 'recipient cannot be empty' error, got: %v", err)
	}
}
