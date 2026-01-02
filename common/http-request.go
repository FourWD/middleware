package common

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// Default HTTP client timeout
const DefaultHTTPTimeout = 30 * time.Second

// httpClient is a shared HTTP client with timeout configured
var httpClient = &http.Client{
	Timeout: DefaultHTTPTimeout,
}

func HttpRequest(url string, method string, token string, jsonString string) (string, error) {
	jsonByte, err := json.Marshal(jsonString)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonByte))
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	response, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
