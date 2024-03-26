package common

import (
	"encoding/json"
)

func StructToString(data interface{}) string {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "Format Error"
	}
	return string(jsonData)
}

func StructToJson(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
