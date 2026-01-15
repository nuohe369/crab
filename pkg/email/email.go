package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// Config email configuration
type Config struct {
	Host     string // SMTP server address
	Port     int    // SMTP port
	Username string // Username
	Password string // Password/authorization code
	From     string // Sender address
	FromName string // Sender name
	UseTLS   bool   // Whether to use TLS
}

// Client email client
type Client struct {
	config Config
}

// Message email message
type Message struct {
	To      []string // Recipients
	Cc      []string // Carbon copy
	Bcc     []string // Blind carbon copy
	Subject string   // Subject
	Body    string   // Body (HTML)
	IsHTML  bool     // Whether HTML format
}

// New creates an email client
func New(cfg Config) *Client {
	return &Client{config: cfg}
}

// Send sends an email
func (c *Client) Send(msg *Message) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("recipient cannot be empty")
	}

	// Build email content
	content := c.buildMessage(msg)

	// All recipients
	recipients := append(append(msg.To, msg.Cc...), msg.Bcc...)

	// Send
	if c.config.UseTLS {
		return c.sendWithTLS(recipients, content)
	}
	return c.sendPlain(recipients, content)
}

// SendSimple simple send (single recipient, plain text)
func (c *Client) SendSimple(to, subject, body string) error {
	return c.Send(&Message{
		To:      []string{to},
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	})
}

// SendHTML sends HTML email
func (c *Client) SendHTML(to, subject, htmlBody string) error {
	return c.Send(&Message{
		To:      []string{to},
		Subject: subject,
		Body:    htmlBody,
		IsHTML:  true,
	})
}

// buildMessage builds email content
func (c *Client) buildMessage(msg *Message) []byte {
	var builder strings.Builder

	// From
	if c.config.FromName != "" {
		builder.WriteString(fmt.Sprintf("From: %s <%s>\r\n", c.config.FromName, c.config.From))
	} else {
		builder.WriteString(fmt.Sprintf("From: %s\r\n", c.config.From))
	}

	// To
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))

	// Cc
	if len(msg.Cc) > 0 {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.Cc, ", ")))
	}

	// Subject
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))

	// Date
	builder.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))

	// MIME
	builder.WriteString("MIME-Version: 1.0\r\n")
	if msg.IsHTML {
		builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	// Empty line separates header and body
	builder.WriteString("\r\n")
	builder.WriteString(msg.Body)

	return []byte(builder.String())
}

// sendPlain plain send (port 25)
func (c *Client) sendPlain(to []string, content []byte) error {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
	return smtp.SendMail(addr, auth, c.config.From, to, content)
}

// sendWithTLS TLS send (port 465/587)
func (c *Client) sendWithTLS(to []string, content []byte) error {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// TLS configuration
	tlsConfig := &tls.Config{
		ServerName: c.config.Host,
	}

	// Connect
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS connection failed: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, c.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Sender
	if err := client.Mail(c.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Recipients
	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	// Write email content
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get writer: %w", err)
	}

	if _, err := w.Write(content); err != nil {
		return fmt.Errorf("failed to write email content: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return client.Quit()
}
