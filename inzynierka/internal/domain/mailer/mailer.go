package mailer

import "context"

// Mailer defines interface for sending system emails.
type Mailer interface {
	SendMail(ctx context.Context, from string, to []string, html, subject string) error
}
