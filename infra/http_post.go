package infra

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Result is the legacy upstream response envelope used by RequestPost.
type Result struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

var requestPostClient = &http.Client{
	Timeout: 10 * time.Second,
}

// RequestPost sends a JSON POST and expects a Result envelope. The token
// may include or omit the "Bearer " prefix.
func RequestPost(url string, token string, payload map[string]interface{}) (Result, error) {
	requestID := uuid.NewString()
	logData := map[string]any{"url": url}

	AppLog.Event("HTTP_POST_START", logData, requestID,
		WithComponent(ComponentHTTPClient),
		WithOperation("post"),
		WithLogKind(LogKindBusiness))

	var response Result

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(payload); err != nil {
		logHTTPPostFailure(err, "HTTP_POST_ENCODE_FAILURE", url, requestID)
		return response, fmt.Errorf("encode payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		logHTTPPostFailure(err, "HTTP_POST_REQUEST_BUILD_FAILURE", url, requestID)
		return response, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	token = strings.TrimPrefix(token, "Bearer ")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := requestPostClient.Do(req)
	if err != nil {
		logHTTPPostFailure(err, "HTTP_POST_EXECUTE_FAILURE", url, requestID)
		return response, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		AppLog.EventError(nil, "HTTP_POST_STATUS_FAILURE", map[string]any{
			"url":    url,
			"status": res.StatusCode,
		}, requestID,
			WithComponent(ComponentHTTPClient),
			WithOperation("post"),
			WithLogKind(LogKindError))
		return response, fmt.Errorf("unexpected status %d", res.StatusCode)
	}

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		logHTTPPostFailure(err, "HTTP_POST_READ_FAILURE", url, requestID)
		return response, fmt.Errorf("read body: %w", err)
	}

	if err := json.Unmarshal(raw, &response); err != nil {
		logHTTPPostFailure(err, "HTTP_POST_UNMARSHAL_FAILURE", url, requestID)
		return response, fmt.Errorf("unmarshal body: %w", err)
	}

	if response.Status != 1 {
		AppLog.EventError(nil, "HTTP_POST_RESULT_FAILURE", map[string]any{
			"url":           url,
			"result_status": response.Status,
			"result_code":   response.Code,
		}, requestID,
			WithComponent(ComponentHTTPClient),
			WithOperation("post"),
			WithLogKind(LogKindError))
		return response, fmt.Errorf("upstream returned status %d (%s)", response.Status, response.Code)
	}

	return response, nil
}

func logHTTPPostFailure(err error, label, url, requestID string) {
	AppLog.EventError(err, label, map[string]any{"url": url}, requestID,
		WithComponent(ComponentHTTPClient),
		WithOperation("post"),
		WithLogKind(LogKindError))
}
