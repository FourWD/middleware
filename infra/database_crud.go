package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// DBCreate inserts a model via the primary GORM DB. Success is implied by a
// successful HTTP response — only failures are logged.
func DBCreate(requestID string, model interface{}) error {
	if Database == nil {
		return errors.New("infra.Database not initialized")
	}
	if err := Database.Create(model).Error; err != nil {
		AppLog.EventError(err, "DB_CREATE_FAILURE", map[string]any{
			"data": modelSnapshot(model, requestID),
		}, requestID,
			WithComponent(ComponentDB),
			WithOperation("create"),
			WithLogKind(LogKindError))
		return err
	}
	return nil
}

// DBUpdate updates a model (non-zero fields) via the primary GORM DB.
func DBUpdate(requestID string, model interface{}) error {
	if Database == nil {
		return errors.New("infra.Database not initialized")
	}
	if err := Database.Updates(model).Error; err != nil {
		AppLog.EventError(err, "DB_UPDATE_FAILURE", map[string]any{
			"data": modelSnapshot(model, requestID),
		}, requestID,
			WithComponent(ComponentDB),
			WithOperation("update"),
			WithLogKind(LogKindError))
		return err
	}
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

	if err := Database.Model(model).Where("id = ?", id).Updates(updateData).Error; err != nil {
		AppLog.EventError(err, "DB_UPDATE_FIELD_FAILURE", map[string]any{
			"record_id": id,
			"data":      updateData,
		}, requestID,
			WithComponent(ComponentDB),
			WithOperation("update_field"),
			WithLogKind(LogKindError))
		return err
	}
	return nil
}

// DBDelete soft-deletes a row by setting deleted_at = now and deleted_by.
func DBDelete(requestID string, model any, id string, deletedBy string) error {
	return DBUpdateField(requestID, model, id, map[string]interface{}{
		"deleted_at": time.Now(),
		"deleted_by": deletedBy,
	})
}

// LogDBErrorCtx is a thin wrapper around LogDBError so callers in this file
// can stay on the request-id path.
func logDBSnapshotFailure(ctx context.Context, err error, requestID string) {
	AppLog.EventError(err, "DB_LOG_SNAPSHOT_FAILURE", nil, requestID,
		WithComponent(ComponentDB),
		WithOperation("log_snapshot"),
		WithLogKind(LogKindDiagnostic))
	_ = ctx
}

// modelSnapshot is a best-effort JSON round-trip used only for log payloads.
// A marshal failure is recorded but never blocks the surrounding DB call.
func modelSnapshot(model interface{}, requestID string) map[string]interface{} {
	data, err := toDBMap(model)
	if err != nil {
		logDBSnapshotFailure(context.Background(), err, requestID)
		return nil
	}
	return data
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
