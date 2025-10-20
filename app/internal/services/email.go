package services

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

type Emailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SMTPEmailer struct {
	cfg SMTPConfig
	auth smtp.Auth
	addr string
}

func NewSMTPEmailer(cfg SMTPConfig) *SMTPEmailer {
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return &SMTPEmailer{cfg: cfg, auth: auth, addr: addr}
}

func (s *SMTPEmailer) Send(ctx context.Context, to, subject, body string) error {
	// Basic sync implementation. Production: use background queue and templating.
	from := s.cfg.From
	msg := buildMessage(from, to, subject, body)
	// Use a short deadline from context
	done := make(chan error, 1)
	go func() {
		err := smtp.SendMail(s.addr, s.auth, from, []string{to}, []byte(msg))
		done <- err
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	case <-time.After(15 * time.Second):
		return fmt.Errorf("email timeout")
	}
}

func buildMessage(from, to, subject, body string) string {
	headers := []string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"Mime-Version: 1.0;",
		"Content-Type: text/plain; charset=\"utf-8\";",
	}
	return strings.Join(headers, "\r\n") + "\r\n\r\n" + body
}
