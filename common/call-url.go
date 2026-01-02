package common

import (
	"io"
	"net/http"
)

func CallUrl(url string) string {
	response, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return ""
	}

	return string(body)
}
