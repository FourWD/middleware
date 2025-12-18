package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type pushBody struct {
	To       string        `json:"to"`
	Messages []textMessage `json:"messages"`
}

type textMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func SendLineNotify(channelToken, to, text string) error {
	body := pushBody{
		To: to,
		Messages: []textMessage{
			{Type: "text", Text: text},
		},
	}

	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "https://api.line.me/v2/bot/message/push", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+channelToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LINE push failed: %s, body=%s", resp.Status, string(raw))
	}
	return nil
}

// import (
// 	"bytes"
// 	"fmt"
// 	"net/http"
// 	"net/url"
// )

// func SendLineNotify(token, message string) error {
// 	apiURL := "https://notify-api.line.me/api/notify"

// 	// Prepare form data
// 	data := url.Values{}
// 	data.Set("message", message)

// 	// Create a new request
// 	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
// 	if err != nil {
// 		return err
// 	}

// 	// Add headers
// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 	req.Header.Set("Authorization", "Bearer "+token)

// 	// Send the request
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("failed to send message: %s", resp.Status)
// 	}

// 	return nil
// }
