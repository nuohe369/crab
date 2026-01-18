// Package email provides email sending functionality with SMTP support
// Package email 提供支持 SMTP 的邮件发送功能
package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// Config represents email configuration
// Config 表示邮件配置
type Config struct {
	Host     string // SMTP server address | SMTP 服务器地址
	Port     int    // SMTP port | SMTP 端口
	Username string // Username | 用户名
	Password string // Password/authorization code | 密码/授权码
	From     string // Sender address | 发件人地址
	FromName string // Sender name | 发件人名称
	UseTLS   bool   // Whether to use TLS | 是否使用 TLS
}

// Client represents an email client
// Client 表示邮件客户端
type Client struct {
	config Config // Email configuration | 邮件配置
}

// Message represents an email message
// Message 表示邮件消息
type Message struct {
	To      []string // Recipients | 收件人
	Cc      []string // Carbon copy | 抄送
	Bcc     []string // Blind carbon copy | 密送
	Subject string   // Subject | 主题
	Body    string   // Body (HTML or plain text) | 正文（HTML 或纯文本）
	IsHTML  bool     // Whether HTML format | 是否 HTML 格式
}

// New creates an email client
// New 创建邮件客户端
func New(cfg Config) *Client {
	return &Client{config: cfg}
}

// Send sends an email
// Send 发送邮件
func (c *Client) Send(msg *Message) error {
	if len(msg.To) == 0 {
		return fmt.Errorf("recipient cannot be empty")
	}

	// Build email content | 构建邮件内容
	content := c.buildMessage(msg)

	// All recipients | 所有收件人
	recipients := append(append(msg.To, msg.Cc...), msg.Bcc...)

	// Send | 发送
	if c.config.UseTLS {
		return c.sendWithTLS(recipients, content)
	}
	return c.sendPlain(recipients, content)
}

// SendSimple sends a simple email (single recipient, plain text)
// SendSimple 发送简单邮件（单个收件人，纯文本）
func (c *Client) SendSimple(to, subject, body string) error {
	return c.Send(&Message{
		To:      []string{to},
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	})
}

// SendHTML sends an HTML email
// SendHTML 发送 HTML 邮件
func (c *Client) SendHTML(to, subject, htmlBody string) error {
	return c.Send(&Message{
		To:      []string{to},
		Subject: subject,
		Body:    htmlBody,
		IsHTML:  true,
	})
}

// buildMessage builds email content
// buildMessage 构建邮件内容
func (c *Client) buildMessage(msg *Message) []byte {
	var builder strings.Builder

	// From | 发件人
	if c.config.FromName != "" {
		builder.WriteString(fmt.Sprintf("From: %s <%s>\r\n", c.config.FromName, c.config.From))
	} else {
		builder.WriteString(fmt.Sprintf("From: %s\r\n", c.config.From))
	}

	// To | 收件人
	builder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))

	// Cc | 抄送
	if len(msg.Cc) > 0 {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.Cc, ", ")))
	}

	// Subject | 主题
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))

	// Date | 日期
	builder.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))

	// MIME | MIME 类型
	builder.WriteString("MIME-Version: 1.0\r\n")
	if msg.IsHTML {
		builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	// Empty line separates header and body | 空行分隔头部和正文
	builder.WriteString("\r\n")
	builder.WriteString(msg.Body)

	return []byte(builder.String())
}

// sendPlain sends email using plain SMTP (port 25)
// sendPlain 使用普通 SMTP 发送邮件（端口 25）
func (c *Client) sendPlain(to []string, content []byte) error {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
	return smtp.SendMail(addr, auth, c.config.From, to, content)
}

// sendWithTLS sends email using TLS (port 465/587)
// sendWithTLS 使用 TLS 发送邮件（端口 465/587）
func (c *Client) sendWithTLS(to []string, content []byte) error {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// TLS configuration | TLS 配置
	tlsConfig := &tls.Config{
		ServerName: c.config.Host,
	}

	// Connect | 连接
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS connection failed: %w", err)
	}
	defer conn.Close()

	// Create SMTP client | 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, c.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate | 认证
	auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Sender | 发件人
	if err := client.Mail(c.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Recipients | 收件人
	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	// Write email content | 写入邮件内容
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
