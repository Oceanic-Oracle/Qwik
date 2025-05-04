package utils

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"mailer/internal/config"
	"net/smtp"
)

type SenderInterface interface {
	SendMsg(string, string, string) error
	Close() error
}

type Sender struct {
	client       *smtp.Client
	smtpUsername string

	log *slog.Logger
}

func NewSender(cfg config.Sender, log *slog.Logger) (SenderInterface, error) {
	if cfg.SmtpUsername == "" {
		return &MokSender{log: log}, nil
	}
	tlsConfig := &tls.Config{
		ServerName: cfg.SmtpHost,
	}

	conn, err := tls.Dial("tcp", cfg.SmtpHost+":"+cfg.SmtpPort, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("TLS connection failed: %w", err)
	}

	client, err := smtp.NewClient(conn, cfg.SmtpHost)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("SMTP client creation failed: %w", err)
	}

	auth := smtp.PlainAuth("", cfg.SmtpUsername, cfg.SmtpPassword, cfg.SmtpHost)
	if err := client.Auth(auth); err != nil {
		client.Quit()
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return &Sender{client: client, smtpUsername: cfg.SmtpUsername}, nil
}

func (s *Sender) SendMsg(email, theme, msg string) error {
	s.log.Info("attempting to send email",
		"to", email,
		"subject", theme,
		"message", msg,
	)

	if err := s.client.Mail(s.smtpUsername); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	if err := s.client.Rcpt(email); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}
	w, err := s.client.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("message write failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("message close failed: %w", err)
	}
	return nil
}

func (s *Sender) Close() error {
	return s.client.Quit()
}

type MokSender struct {
	log *slog.Logger
}

func (ms *MokSender) SendMsg(email, theme, msg string) error {
	ms.log.Debug("sending message")
	return nil
}

func (ms *MokSender) Close() error {
	return nil
}
