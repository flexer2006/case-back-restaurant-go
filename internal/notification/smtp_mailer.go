// Package notification contains components for working with notifications,
// including sending emails via SMTP
package notification

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/flexer2006/case-back-restaurant-go/configs"
	"github.com/flexer2006/case-back-restaurant-go/internal/domain"
	"gopkg.in/gomail.v2"
)

const (
	smtpConnectTimeout = 10 * time.Second
	smtpSendTimeout    = 30 * time.Second
)

type Dialer interface {
	DialAndSend(m ...*gomail.Message) error
}

var SMTPSendTimeout = smtpSendTimeout

var NewDialer = func(host string, port int, username, password string) Dialer {
	return gomail.NewDialer(host, port, username, password)
}

type SMTPMailer struct {
	config *configs.SMTPConfig
	dialer Dialer
}

func NewSMTPMailer(config *configs.SMTPConfig) domain.EmailSender {
	if config == nil {
		return &SMTPMailer{config: nil}
	}

	dialer := NewDialer(config.Host, config.Port, config.Username, config.Password)

	if d, ok := dialer.(*gomail.Dialer); ok && config.Secure {
		d.TLSConfig = &tls.Config{
			ServerName:         config.Host,
			InsecureSkipVerify: false,
		}
	}

	return &SMTPMailer{
		config: config,
		dialer: dialer,
	}
}

func (s *SMTPMailer) SendEmail(to, subject, body string) error {

	if s.config == nil {
		return fmt.Errorf("%s: %s", common.ErrInitSMTP, common.ErrNilConfig)
	}

	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)

	if to == "" {
		return fmt.Errorf("%s: %s", common.ErrInvalidEmailParams, common.ErrEmptyRecipient)
	}

	if subject == "" {
		return fmt.Errorf("%s: %s", common.ErrInvalidEmailParams, common.ErrEmptySubject)
	}

	if !strings.Contains(to, "@") {
		return fmt.Errorf("%s: %s", common.ErrInvalidEmailParams, common.ErrSMTPInvalidRecipient)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", s.config.From)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	ctx, cancel := context.WithTimeout(context.Background(), SMTPSendTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- s.dialer.DialAndSend(msg)
	}()

	select {
	case err := <-done:
		if err != nil {

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return fmt.Errorf("%s: %s: %w", common.ErrDialSMTP, common.ErrSMTPTimeout, err)
			}

			if opErr, ok := err.(*net.OpError); ok {
				return fmt.Errorf("%s: network operation error %s: %w", common.ErrDialSMTP, opErr.Op, err)
			}

			return fmt.Errorf("%s: %w", common.ErrSendEmail, err)
		}
		return nil

	case <-ctx.Done():
		return fmt.Errorf("%s: %s", common.ErrSendEmail, common.ErrSMTPTimeout)
	}
}
