package kit

import "encoding/json"

func StructToString(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

func StructToJson(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	jsonString := string(jsonData)
	if jsonString == "null" {
		return "", nil
	}
	return jsonString, nil
}
