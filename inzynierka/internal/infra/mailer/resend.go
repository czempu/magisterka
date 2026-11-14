package mailer

import (
	"context"
	"errors"

	domainMailer "erpbadger/internal/domain/mailer"
	"github.com/resend/resend-go/v2"
)

type ResendMailer struct {
	client *resend.Client
	apiKey string
}

func NewResendMailer(apiKey string) domainMailer.Mailer {
	if apiKey == "" {
		return &NoopMailer{}
	}
	return &ResendMailer{
		client: resend.NewClient(apiKey),
		apiKey: apiKey,
	}
}

func (m *ResendMailer) SendMail(ctx context.Context, from string, to []string, html, subject string) error {
	if m.client == nil {
		return errors.New("resend client not initialized")
	}
	params := &resend.SendEmailRequest{
		From:    from,
		To:      to,
		Html:    html,
		Subject: subject,
	}
	_, err := m.client.Emails.Send(params)
	return err
}

// NoopMailer fallback for development / test when API key is not configured.
type NoopMailer struct{}

func (m *NoopMailer) SendMail(ctx context.Context, from string, to []string, html, subject string) error {
	// Silently succeed in dev mode without throwing errors
	return nil
}
