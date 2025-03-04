package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

type Result struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func RequestPost(url string, token string, payload map[string]interface{}) (Result, error) {
	var response Result

	token = strings.ReplaceAll(token, "Bearer ", "")
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	requestPayload := new(bytes.Buffer)
	json.NewEncoder(requestPayload).Encode(payload)
	req, err := http.NewRequest("POST", url, requestPayload)
	if err != nil {
		return response, errors.New("req error")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)
	responseUrl, err := client.Do(req)
	if err != nil {
		return response, errors.New("response error")
	}
	defer responseUrl.Body.Close()

	if responseUrl.StatusCode != 200 {
		return response, errors.New("status is not 200")
	}

	body, err := io.ReadAll(responseUrl.Body)
	if err != nil {
		return response, errors.New("body error")
	}

	if err = json.Unmarshal(body, &response); err != nil {

	}

	if response.Status != 1 {
		return response, errors.New("status error")
	}

	return response, nil
}
