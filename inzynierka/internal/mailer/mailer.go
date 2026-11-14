package mailer

import ("github.com/resend/resend-go/v2"
	"erpbadger/config"
)

func SendMail(from string, to []string, html,subject string) error {
    apiKey := config.ResendApiKey
    client := resend.NewClient(apiKey)
    params := &resend.SendEmailRequest{
        From:    from,
        To:      to,
        Html:    html,
        Subject: subject,
    }
    _, err := client.Emails.Send(params)
        return err
}
