package common

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/spf13/viper"
)

func SendMail(sender, subject, body, recipient string) (string, error) {
	message := mailgun.NewMessage(sender, subject, body, recipient)
	return sendMail(message)
}

func SendMailWithTemplate(sender, subject, body, recipient, template string, data map[string]interface{}) (string, error) {
	message := mailgun.NewMessage(sender, subject, body, recipient)
	message.SetTemplate(template)
	addTemplateVariable(message, data)

	return sendMail(message)
}

func addTemplateVariable(message *mailgun.Message, data map[string]interface{}) {
	for key, value := range data {
		strValue, ok := value.(string)
		if !ok {
			// log.Printf("Skipping key %s because it's not a string", key)
			continue
		}
		if err := message.AddTemplateVariable(key, strValue); err != nil {
			// log.Printf("Failed to add template variable: %v", err)
			continue
		}
	}
}

func sendMail(message *mailgun.Message) (string, error) {
	domain := viper.GetString("mail.domain")
	apiKey := viper.GetString("mail.key")
	mg := mailgun.NewMailgun(domain, apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, id, err := mg.Send(ctx, message)
	// fmt.Printf("ID: %s Resp: %s\n", id, resp)

	// logData := map[string]interface{}{
	// 	"sender":    MessageConfig.Sender,
	// 	"subject":   message.().Subject,
	// 	"body":      body,
	// 	"recipient": recipient,
	// }
	logData := map[string]interface{}{}

	if err == nil {
		Log("Send mail", logData, "")
	} else {
		logData["error"] = err
		LogError("Send mail", logData, "")
	}

	return id, err
}
