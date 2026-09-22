package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/new-er/schufa-dsgvo-requester/internal/config"
)

func Send(cfg *config.Config, subject, body string) error {
	auth := smtp.PlainAuth("", cfg.Mail.Username, cfg.Mail.Password, cfg.Mail.SMTPHost)

	// Build message
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s", cfg.Mail.FromAddr, cfg.Mail.ToAddr, subject, body)

	// TLS config
	tlsConfig := &tls.Config{
		ServerName: cfg.Mail.SMTPHost,
	}

	// Connect to server
	addr := fmt.Sprintf("%s:%d", cfg.Mail.SMTPHost, cfg.Mail.SMTPPort)
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer client.Close()

	// Start TLS
	if err := client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("starttls: %w", err)
	}

	// Auth
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	// Set sender and recipient
	if err := client.Mail(cfg.Mail.FromAddr); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := client.Rcpt(cfg.Mail.ToAddr); err != nil {
		return fmt.Errorf("rcpt to: %w", err)
	}

	// Data
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("close: %w", err)
	}

	// Quit
	return client.Quit()
}