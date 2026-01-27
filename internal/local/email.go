package local

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"gobot/internal/config"

	"github.com/zeromicro/go-zero/core/logx"
)

// EmailService handles email sending via SMTP or Outlet.sh
type EmailService struct {
	config     config.Config
	httpClient *http.Client
}

// NewEmailService creates a new email service
func NewEmailService(cfg config.Config) *EmailService {
	return &EmailService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	To           string            `json:"to"`
	Subject      string            `json:"subject,omitempty"`
	Body         string            `json:"body,omitempty"`         // HTML body
	TextBody     string            `json:"textBody,omitempty"`     // Plain text body
	TemplateSlug string            `json:"templateSlug,omitempty"` // Outlet.sh template
	Variables    map[string]string `json:"variables,omitempty"`    // Template variables
	Tags         []string          `json:"tags,omitempty"`
	Meta         map[string]string `json:"meta,omitempty"`
}

// SendEmailResponse represents the response from sending an email
type SendEmailResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"messageId,omitempty"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}

// SubscribeRequest represents a request to subscribe to a list (Outlet.sh only)
type SubscribeRequest struct {
	Email      string            `json:"email"`
	Name       string            `json:"name,omitempty"`
	CustomData map[string]string `json:"customData,omitempty"`
}

// SubscribeResponse represents the response from subscribing
type SubscribeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// EnrollSequenceRequest represents a request to enroll in a sequence (Outlet.sh only)
type EnrollSequenceRequest struct {
	Email      string            `json:"email"`
	CustomData map[string]string `json:"customData,omitempty"`
}

// EnrollSequenceResponse represents the response from enrolling
type EnrollSequenceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// IsConfigured returns true if any email service is configured (SMTP or Outlet.sh)
func (s *EmailService) IsConfigured() bool {
	return s.hasSMTP() || s.hasOutlet()
}

// hasSMTP returns true if SMTP is configured
func (s *EmailService) hasSMTP() bool {
	return s.config.Email.SMTPHost != "" && s.config.Email.FromAddress != ""
}

// hasOutlet returns true if Outlet.sh is configured
func (s *EmailService) hasOutlet() bool {
	return s.config.Outlet.BaseURL != "" && s.config.Outlet.APIKey != ""
}

// SendEmail sends an email via SMTP or Outlet.sh (prefers Outlet.sh if configured)
func (s *EmailService) SendEmail(ctx context.Context, req SendEmailRequest) (*SendEmailResponse, error) {
	// Prefer Outlet.sh if configured (supports templates)
	if s.hasOutlet() {
		return s.sendViaOutlet(ctx, req)
	}

	// Fall back to SMTP
	if s.hasSMTP() {
		return s.sendViaSMTP(ctx, req)
	}

	logx.Info("No email service configured - skipping email send")
	return &SendEmailResponse{
		Success: true,
		Status:  "skipped",
		Message: "Email service not configured",
	}, nil
}

// SendTemplateEmail sends an email using an Outlet.sh template (falls back to plain email for SMTP)
func (s *EmailService) SendTemplateEmail(ctx context.Context, to, templateSlug string, variables map[string]string) (*SendEmailResponse, error) {
	return s.SendEmail(ctx, SendEmailRequest{
		To:           to,
		TemplateSlug: templateSlug,
		Variables:    variables,
	})
}

// SendSimpleEmail sends a simple email with subject and body (works with both SMTP and Outlet.sh)
func (s *EmailService) SendSimpleEmail(ctx context.Context, to, subject, htmlBody, textBody string) (*SendEmailResponse, error) {
	return s.SendEmail(ctx, SendEmailRequest{
		To:       to,
		Subject:  subject,
		Body:     htmlBody,
		TextBody: textBody,
	})
}

// sendViaSMTP sends email using SMTP
func (s *EmailService) sendViaSMTP(_ context.Context, req SendEmailRequest) (*SendEmailResponse, error) {
	cfg := s.config.Email

	// Build email message
	var msg bytes.Buffer

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", cfg.FromName, cfg.FromAddress))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", req.To))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", req.Subject))

	if cfg.ReplyTo != "" {
		msg.WriteString(fmt.Sprintf("Reply-To: %s\r\n", cfg.ReplyTo))
	}

	msg.WriteString("MIME-Version: 1.0\r\n")

	// Determine body content
	body := req.Body
	contentType := "text/html"
	if body == "" {
		body = req.TextBody
		contentType = "text/plain"
	}

	// If using template but no Outlet.sh, log warning and use variables as body
	if req.TemplateSlug != "" && body == "" {
		logx.Infof("Template '%s' requested but Outlet.sh not configured - sending variable data", req.TemplateSlug)
		var parts []string
		for k, v := range req.Variables {
			parts = append(parts, fmt.Sprintf("%s: %s", k, v))
		}
		body = strings.Join(parts, "\n")
		contentType = "text/plain"
	}

	msg.WriteString(fmt.Sprintf("Content-Type: %s; charset=UTF-8\r\n", contentType))
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)

	var auth smtp.Auth
	if cfg.SMTPUser != "" && cfg.SMTPPass != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
	}

	// Send email
	var err error
	if cfg.SMTPPort == 465 {
		// SSL/TLS connection
		err = s.sendMailTLS(addr, auth, cfg.FromAddress, []string{req.To}, msg.Bytes())
	} else {
		// STARTTLS connection (port 587) or plain (port 25)
		err = smtp.SendMail(addr, auth, cfg.FromAddress, []string{req.To}, msg.Bytes())
	}

	if err != nil {
		logx.Errorf("SMTP send failed: %v", err)
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	logx.Infof("Email sent via SMTP to %s", req.To)
	return &SendEmailResponse{
		Success: true,
		Status:  "sent",
		Message: "Email sent via SMTP",
	}, nil
}

// sendMailTLS sends email over TLS (for port 465)
func (s *EmailService) sendMailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	cfg := s.config.Email

	tlsConfig := &tls.Config{
		ServerName: cfg.SMTPHost,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err = client.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// sendViaOutlet sends email using Outlet.sh API
func (s *EmailService) sendViaOutlet(ctx context.Context, req SendEmailRequest) (*SendEmailResponse, error) {
	var resp SendEmailResponse
	if err := s.doOutletRequest(ctx, "POST", "/api/v1/sdk/emails/send", req, &resp); err != nil {
		return nil, err
	}
	logx.Infof("Email sent via Outlet.sh to %s", req.To)
	return &resp, nil
}

// Subscribe adds an email to a list (Outlet.sh only)
func (s *EmailService) Subscribe(ctx context.Context, listSlug string, req SubscribeRequest) (*SubscribeResponse, error) {
	if !s.hasOutlet() {
		logx.Info("Outlet.sh not configured - skipping subscribe")
		return &SubscribeResponse{
			Success: true,
			Message: "List subscription requires Outlet.sh",
		}, nil
	}

	path := fmt.Sprintf("/api/v1/sdk/lists/%s/subscribe", listSlug)
	var resp SubscribeResponse
	if err := s.doOutletRequest(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// EnrollInSequence enrolls a contact in an email sequence (Outlet.sh only)
func (s *EmailService) EnrollInSequence(ctx context.Context, sequenceSlug string, req EnrollSequenceRequest) (*EnrollSequenceResponse, error) {
	if !s.hasOutlet() {
		logx.Info("Outlet.sh not configured - skipping sequence enroll")
		return &EnrollSequenceResponse{
			Success: true,
			Message: "Sequence enrollment requires Outlet.sh",
		}, nil
	}

	path := fmt.Sprintf("/api/v1/sdk/sequences/%s/enroll", sequenceSlug)
	var resp EnrollSequenceResponse
	if err := s.doOutletRequest(ctx, "POST", path, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// doOutletRequest makes an HTTP request to Outlet.sh
func (s *EmailService) doOutletRequest(ctx context.Context, method, path string, body, result any) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := s.config.Outlet.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.Outlet.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("outlet.sh returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
