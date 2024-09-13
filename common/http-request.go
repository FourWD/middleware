package common

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func HttpRequest(url string, method string, token string, jsonString string) (string, error) {
	// jsonData, err := json.Marshal(p)
	client := &http.Client{}

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

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	// var resp ApiResponse
	// errJson := json.Unmarshal(body, &resp)
	// if errJson != nil {
	// 	return "", errJson
	// }
	// result = &resp.Data

	return string(body), nil
}
