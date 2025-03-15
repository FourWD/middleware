package common

import (
	"encoding/json"
)

func GetJsonValues(jsonStr string, keys ...string) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, key := range keys {
		if val, ok := data[key]; ok {
			result[key] = val
		}
	}
	return result, nil
}
