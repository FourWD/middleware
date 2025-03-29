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

/* new mailgun attach file
func SendMailWithTemplate(sender, subject, body, recipient, template string, data map[string]interface{}, atttachURL string, atttachURL2 string) (string, error) {
	message := mailgun.NewMessage(sender, subject, body, recipient)
	message.SetTemplate(template)
	addTemplateVariable(message, data)
	if atttachURL != "" {
		// Download the PDF from GCS URL
		pdfURL := "https://storage.googleapis.com/fourd-aot/receipt/da92fffe-39c7-456d-b9ac-4de20e3eb949.pdf"
		pdfData, err := downloadFileFromURL(pdfURL)
		if err != nil {
			log.Fatalf("Failed to download PDF: %v", err)
		}

		// Attach PDF file
		message.AddBufferAttachment("receipt.pdf", pdfData)
		// if err != nil {
		// 	log.Fatalf("Failed to attach PDF: %v", err)
		// }
		// message.AddAttachment(atttachURL)
	}
	if atttachURL2 != "" {
		// Download the PDF from GCS URL
		pdfURL := "https://storage.googleapis.com/fourd-aot/receipt/da92fffe-39c7-456d-b9ac-4de20e3eb949.pdf"
		pdfData, err := downloadFileFromURL(pdfURL)
		if err != nil {
			log.Fatalf("Failed to download PDF: %v", err)
		}

		// Attach PDF file
		message.AddBufferAttachment("receipt2.pdf", pdfData)
		// if err != nil {
		// 	log.Fatalf("Failed to attach PDF: %v", err)
		// }
		// message.AddAttachment(atttachURL)
	}

	return sendMail(message)
}

func downloadFileFromURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch file, status: %d", resp.StatusCode)
	}

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}

	return fileData, nil
}

*/
