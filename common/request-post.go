package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Result struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func RequestPost(url string, token string, payload map[string]interface{}) (Result, error) {
	requestID := uuid.NewString()

	// payloadJSON, _ := json.MarshalIndent(payload, "", "  ")

	logData := map[string]interface{}{
		"token":   token,
		"url":     url,
		"payload": fmt.Sprintf("%+v", payload),
		// "payload_json": payloadJSON,
	}

	Log("RequestPost", logData, requestID)

	// if token == "" {
	// 	LogError("no token", nil, requestID)
	// 	return Result{}, errors.New("no token")
	// }

	var response Result

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	requestPayload := new(bytes.Buffer)
	json.NewEncoder(requestPayload).Encode(payload)
	req, err := http.NewRequest("POST", url, requestPayload)
	if err != nil {
		LogError(err.Error(), nil, requestID)
		return response, err
	}
	req.Header.Add("Content-Type", "application/json")

	token = strings.ReplaceAll(token, "Bearer ", "")
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}
	responseUrl, err := client.Do(req)
	if err != nil {
		LogError(err.Error(), nil, requestID)
		return response, err
	}
	defer responseUrl.Body.Close()

	if responseUrl.StatusCode != 200 {
		LogError("status is not 200", nil, requestID)
		return response, errors.New("status is not 200")
	}

	body, err := io.ReadAll(responseUrl.Body)
	if err != nil {
		LogError(err.Error(), nil, requestID)
		return response, err
	}

	if err = json.Unmarshal(body, &response); err != nil {
		LogError(err.Error(), nil, requestID)
		return response, err
	}

	if response.Status != 1 {
		LogError("status error", nil, requestID)
		return response, errors.New("status error")
	}

	return response, nil
}
