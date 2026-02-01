package common

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func DBCreate(requestID string, model interface{}) error {
	data, _ := toMap(model)
	logData := map[string]interface{}{
		"data": data,
	}

	err := Database.Create(model).Error

	if err != nil {
		logData["status"] = "error"
		logData["error"] = err
		LogError("DBCreate", logData, requestID)
	} else {
		logData["status"] = "success"
		Log("DBCreate", logData, requestID)
	}

	return err
}

func DBUpdate(requestID string, model interface{}) error {
	data, _ := toMap(model)
	logData := map[string]interface{}{
		"data": data,
	}

	err := Database.Updates(model).Error

	if err != nil {
		logData["status"] = "error"
		logData["error"] = err
		LogError("DBUpdate", logData, requestID)
	} else {
		logData["status"] = "success"
		Log("DBUpdate", logData, requestID)
	}

	return err
}

func DBUpdateField(requestID string, model any, id string, updateData map[string]interface{}) error {
	for key, value := range updateData {
		if floatValue, ok := value.(float64); ok {
			updateData[key] = parseToFloat(fmt.Sprintf("%.6f", floatValue))
		}
	}

	logData := map[string]interface{}{
		"data": updateData,
	}

	err := Database.Model(model).Where("id = ?", id).Updates(updateData).Error

	if err != nil {
		logData["status"] = "error"
		logData["error"] = err
		LogError("DBUpdateField", logData, requestID)
	} else {
		logData["status"] = "success"
		Log("DBUpdateField", logData, requestID)
	}

	return err
}

func DBDelete(requestID string, model any, id string, DeletedBy string) error {
	updateData := map[string]interface{}{}
	updateData["deleted_at"] = time.Now()
	updateData["deleted_by"] = DeletedBy
	return DBUpdateField(requestID, model, id, updateData)
}

func toMap(v interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func parseToFloat(str string) float64 {
	parsedValue, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return parsedValue
}
