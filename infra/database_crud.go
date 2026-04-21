package infra

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// DBCreate inserts a model via the primary GORM DB and logs the outcome.
func DBCreate(requestID string, model interface{}) error {
	if Database == nil {
		return errors.New("infra.Database not initialized")
	}

	data, _ := toDBMap(model)
	logData := map[string]interface{}{"data": data}

	if err := Database.Create(model).Error; err != nil {
		AppLog.EventError(err, "DBCreate", logData, requestID)
		return err
	}
	logData["status"] = "success"
	AppLog.Event("DBCreate", logData, requestID)
	return nil
}

// DBUpdate updates a model (non-zero fields) via the primary GORM DB.
func DBUpdate(requestID string, model interface{}) error {
	if Database == nil {
		return errors.New("infra.Database not initialized")
	}

	data, _ := toDBMap(model)
	logData := map[string]interface{}{"data": data}

	if err := Database.Updates(model).Error; err != nil {
		AppLog.EventError(err, "DBUpdate", logData, requestID)
		return err
	}
	logData["status"] = "success"
	AppLog.Event("DBUpdate", logData, requestID)
	return nil
}

// DBUpdateField updates a single row identified by id with a map of fields.
// float64 values are normalised to 6-decimal precision to match downstream SQL
// column semantics.
func DBUpdateField(requestID string, model any, id string, updateData map[string]interface{}) error {
	if Database == nil {
		return errors.New("infra.Database not initialized")
	}

	for key, value := range updateData {
		if floatValue, ok := value.(float64); ok {
			updateData[key] = parseDBFloat(fmt.Sprintf("%.6f", floatValue))
		}
	}

	logData := map[string]interface{}{"data": updateData}

	if err := Database.Model(model).Where("id = ?", id).Updates(updateData).Error; err != nil {
		AppLog.EventError(err, "DBUpdateField", logData, requestID)
		return err
	}
	logData["status"] = "success"
	AppLog.Event("DBUpdateField", logData, requestID)
	return nil
}

// DBDelete soft-deletes a row by setting deleted_at = now and deleted_by.
func DBDelete(requestID string, model any, id string, deletedBy string) error {
	return DBUpdateField(requestID, model, id, map[string]interface{}{
		"deleted_at": time.Now(),
		"deleted_by": deletedBy,
	})
}

func toDBMap(v interface{}) (map[string]interface{}, error) {
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

func parseDBFloat(str string) float64 {
	parsedValue, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return parsedValue
}
