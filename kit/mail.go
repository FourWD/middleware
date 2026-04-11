package kit

import (
	"context"
	"fmt"

	"github.com/mailgun/mailgun-go/v4"
)

type MailConfig struct {
	Domain string
	APIKey string
}

type MailClient struct {
	client mailgun.Mailgun
}

func NewMailClient(cfg MailConfig) *MailClient {
	return &MailClient{
		client: mailgun.NewMailgun(cfg.Domain, cfg.APIKey),
	}
}

func (mc *MailClient) SendMail(ctx context.Context, sender, subject, body, recipient string) (string, error) {
	message := mailgun.NewMessage(sender, subject, body, recipient)
	_, id, err := mc.client.Send(ctx, message)
	return id, err
}

func (mc *MailClient) SendMailWithTemplate(ctx context.Context, sender, subject, body, recipient, template string, vars map[string]any) (string, error) {
	message := mailgun.NewMessage(sender, subject, body, recipient)
	message.SetTemplate(template)

	for key, value := range vars {
		if err := message.AddTemplateVariable(key, value); err != nil {
			return "", fmt.Errorf("setting template variable %s: %w", key, err)
		}
	}

	_, id, err := mc.client.Send(ctx, message)
	return id, err
}
