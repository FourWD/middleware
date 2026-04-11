package kit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func NominatimReverse(ctx context.Context, client *http.Client, nomURL string, lat float32, lon float32, lang string) (map[string]any, error) {
	url := fmt.Sprintf("%s/reverse?format=jsonv2&lat=%.6f&lon=%.6f", nomURL, lat, lon)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if lang != "" {
		req.Header.Set("Accept-Language", lang)
	}

	resp, err := client.Do(req)
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
