package common

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func NominatimReverse(ctx context.Context, nomURL string, lat float32, lon float32, lang string) (map[string]any, error) {
	u := fmt.Sprintf("%s/reverse?format=jsonv2&lat=%.6f&lon=%.6f", nomURL, lat, lon)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if lang != "" {
		req.Header.Set("Accept-Language", lang)
	}

	httpClient := NewHttpClient(10)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	return out, nil
}
