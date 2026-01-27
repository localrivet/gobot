package email

import (
	"context"
	"fmt"
	"net/smtp"
)

// Service handles email sending via SMTP
type Service struct {
	smtpHost    string
	smtpPort    int
	smtpUser    string
	smtpPass    string
	fromAddress string
	fromName    string
	replyTo     string
}

// Config holds email service configuration
type Config struct {
	SMTPHost    string
	SMTPPort    int
	SMTPUser    string
	SMTPPass    string
	FromAddress string
	FromName    string
	ReplyTo     string
}

// NewService creates a new email service
func NewService(cfg Config) *Service {
	return &Service{
		smtpHost:    cfg.SMTPHost,
		smtpPort:    cfg.SMTPPort,
		smtpUser:    cfg.SMTPUser,
		smtpPass:    cfg.SMTPPass,
		fromAddress: cfg.FromAddress,
		fromName:    cfg.FromName,
		replyTo:     cfg.ReplyTo,
	}
}

// IsConfigured returns true if the email service has valid credentials
func (s *Service) IsConfigured() bool {
	return s.smtpHost != "" && s.smtpUser != "" && s.smtpPass != ""
}

// sendEmail sends an HTML email via SMTP
func (s *Service) sendEmail(to, subject, htmlBody string) error {
	if !s.IsConfigured() {
		return fmt.Errorf("email service not configured: missing SMTP credentials")
	}

	// Build email message
	headers := fmt.Sprintf("From: %s <%s>\r\n", s.fromName, s.fromAddress)
	headers += fmt.Sprintf("To: %s\r\n", to)
	if s.replyTo != "" {
		headers += fmt.Sprintf("Reply-To: %s\r\n", s.replyTo)
	}
	headers += fmt.Sprintf("Subject: %s\r\n", subject)
	headers += "MIME-Version: 1.0\r\n"
	headers += "Content-Type: text/html; charset=UTF-8\r\n"
	headers += "\r\n"

	message := []byte(headers + htmlBody)

	// SMTP auth
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)

	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)
	err := smtp.SendMail(addr, auth, s.fromAddress, []string{to}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendEmail sends an HTML email to a single recipient
func (s *Service) SendEmail(ctx context.Context, to, subject, htmlBody string) error {
	return s.sendEmail(to, subject, htmlBody)
}

// SendEmailFrom sends an HTML email with a custom from address
func (s *Service) SendEmailFrom(ctx context.Context, fromEmail, fromName, to, subject, htmlBody string) error {
	if !s.IsConfigured() {
		return fmt.Errorf("email service not configured: missing SMTP credentials")
	}

	// Use provided from or fall back to defaults
	from := fromEmail
	name := fromName
	if from == "" {
		from = s.fromAddress
	}
	if name == "" {
		name = s.fromName
	}

	headers := fmt.Sprintf("From: %s <%s>\r\n", name, from)
	headers += fmt.Sprintf("To: %s\r\n", to)
	if s.replyTo != "" {
		headers += fmt.Sprintf("Reply-To: %s\r\n", s.replyTo)
	}
	headers += fmt.Sprintf("Subject: %s\r\n", subject)
	headers += "MIME-Version: 1.0\r\n"
	headers += "Content-Type: text/html; charset=UTF-8\r\n"
	headers += "\r\n"

	message := []byte(headers + htmlBody)

	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)

	return smtp.SendMail(addr, auth, from, []string{to}, message)
}

// SendPasswordResetEmail sends a password reset email
func (s *Service) SendPasswordResetEmail(ctx context.Context, to, resetURL string) error {
	subject := "Reset Your Password"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0;">
        <h1 style="margin: 0; font-size: 24px;">Password Reset Request</h1>
    </div>
    <div style="background: #f9fafb; padding: 30px; border-radius: 0 0 8px 8px;">
        <p>You requested a password reset. Click the button below to reset your password:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="display: inline-block; background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; font-weight: 600;">Reset Password</a>
        </div>
        <p style="color: #6b7280; font-size: 14px;">If you didn't request this, you can safely ignore this email. This link expires in 1 hour.</p>
        <p style="color: #6b7280; font-size: 14px;">If the button doesn't work, copy and paste this link into your browser:</p>
        <p style="color: #667eea; font-size: 12px; word-break: break-all;">%s</p>
    </div>
</body>
</html>
	`, resetURL, resetURL)

	return s.sendEmail(to, subject, body)
}

// SendWelcomeEmail sends a welcome email to new users
func (s *Service) SendWelcomeEmail(ctx context.Context, to, name, appURL string) error {
	subject := "Welcome to " + s.fromName

	greeting := "there"
	if name != "" {
		greeting = name
	}

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0;">
        <h1 style="margin: 0; font-size: 24px;">Welcome!</h1>
    </div>
    <div style="background: #f9fafb; padding: 30px; border-radius: 0 0 8px 8px;">
        <p>Hi %s,</p>
        <p>Thanks for signing up! We're excited to have you on board.</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="display: inline-block; background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 6px; font-weight: 600;">Get Started</a>
        </div>
        <p>If you have any questions, just reply to this email. We're always happy to help!</p>
        <p>Best,<br>The Team</p>
    </div>
</body>
</html>
	`, greeting, appURL)

	return s.sendEmail(to, subject, body)
}
