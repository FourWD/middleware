package common

import "encoding/json"

func StructToString(data interface{}) string {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "Format Error"
	}
	return string(jsonData)
}
