package common

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/spf13/viper"
)

func SendMail(sender, subject, body, recipient string) (string, error) {
	domain := viper.GetString("mail.domain")
	apiKey := viper.GetString("mail.key")
	mg := mailgun.NewMailgun(domain, apiKey)

	message := mailgun.NewMessage(sender, subject, body, recipient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, id, err := mg.Send(ctx, message)
	// fmt.Printf("ID: %s Resp: %s\n", id, resp)

	logData := map[string]interface{}{
		"domain":    domain,
		"sender":    sender,
		"subject":   subject,
		"body":      body,
		"recipient": recipient,
	}

	if err == nil {
		Log("Send mail", logData, "")
	} else {
		logData["error"] = err
		LogError("Send mail", logData, "")
	}

	return id, err
}
