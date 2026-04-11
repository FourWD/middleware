package kit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type pushBody struct {
	To       string        `json:"to"`
	Messages []textMessage `json:"messages"`
}

type textMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func SendLineNotify(ctx context.Context, client *http.Client, channelToken, to, text string) error {
	body := pushBody{
		To: to,
		Messages: []textMessage{
			{Type: "text", Text: text},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.line.me/v2/bot/message/push", bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+channelToken)
	req.Header.Set("Content-Type", "application/json")

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
