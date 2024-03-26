package common

import (
	"encoding/json"
	"strings"
)

func StructToString(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	jsonString := string(jsonData)
	jsonString = strings.Replace(jsonString, `\"`, `"`, -1)
	return jsonString
}

func StructToJson(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	jsonString := string(jsonData)
	jsonString = strings.Replace(jsonString, `\"`, `"`, -1)
	return jsonString, nil
}
