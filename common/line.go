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
