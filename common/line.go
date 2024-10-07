package common

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

func SendLineNotify(token, message string) error {
	apiURL := "https://notify-api.line.me/api/notify"

	// Prepare form data
	data := url.Values{}
	data.Set("message", message)

	// Create a new request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}

	// Add headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message: %s", resp.Status)
	}

	return nil
}
