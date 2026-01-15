package email

import (
	"bytes"
	"html/template"
)

// Predefined templates
var templates = map[string]string{
	// Verification code template
	"verify_code": `
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
  <h2>验证码</h2>
  <p>您好,您的验证码是:</p>
  <p style="font-size: 24px; font-weight: bold; color: #1890ff; letter-spacing: 4px;">{{.Code}}</p>
  <p>验证码有效期为 {{.ExpireMinutes}} 分钟,请勿泄露给他人.</p>
  <p style="color: #999; font-size: 12px;">如非本人操作,请忽略此邮件.</p>
</body>
</html>`,

	// Password reset template
	"password_reset": `
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
  <h2>密码重置</h2>
  <p>您好,您正在重置密码.</p>
  <p>请点击以下链接重置密码({{.ExpireMinutes}} 分钟内有效):</p>
  <p><a href="{{.ResetURL}}" style="color: #1890ff;">{{.ResetURL}}</a></p>
  <p style="color: #999; font-size: 12px;">如非本人操作,请忽略此邮件.</p>
</body>
</html>`,

	// Welcome email template
	"welcome": `
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; padding: 20px;">
  <h2>欢迎加入 {{.SiteName}}</h2>
  <p>您好,{{.Nickname}}！</p>
  <p>感谢您注册 {{.SiteName}},祝您使用愉快！</p>
</body>
</html>`,
}

// RenderTemplate renders a template
func RenderTemplate(name string, data any) (string, error) {
	tplStr, ok := templates[name]
	if !ok {
		return "", nil
	}

	tpl, err := template.New(name).Parse(tplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// VerifyCodeData verification code template data
type VerifyCodeData struct {
	Code          string
	ExpireMinutes int
}

// PasswordResetData password reset template data
type PasswordResetData struct {
	ResetURL      string
	ExpireMinutes int
}

// WelcomeData welcome email template data
type WelcomeData struct {
	SiteName string
	Nickname string
}
