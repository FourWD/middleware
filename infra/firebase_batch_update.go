package infra

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"

	"github.com/google/uuid"
)

func FirebaseBatchUpdate(docPaths map[string]map[string]interface{}) error {
	if len(docPaths) == 0 {
		return nil
	}

	batchID := uuid.NewString()
	start := time.Now()

	AppLog.Event("FIREBASE_BATCH_START", map[string]any{
		"batch_id":  batchID,
		"doc_count": len(docPaths),
	}, batchID,
		WithComponent(ComponentFirebase),
		WithOperation("batch_update"),
		WithLogKind(LogKindBusiness))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	bw := FirestoreClient.BulkWriter(ctx)

	var firstErr error
	failCount := 0

	for path, updateData := range docPaths {
		docRef := FirestoreClient.Doc(path)
		if _, err := bw.Set(docRef, updateData, firestore.MergeAll); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			failCount++
		}
	}

	bw.Flush()
	bw.End()

	if firstErr != nil {
		AppLog.EventError(firstErr, "FIREBASE_BATCH_FAILURE", map[string]any{
			"batch_id":     batchID,
			"doc_count":    len(docPaths),
			"fail_count":   failCount,
			"duration_ms":  time.Since(start).Milliseconds(),
		}, batchID,
			WithComponent(ComponentFirebase),
			WithOperation("batch_update"),
			WithLogKind(LogKindError))
		return firstErr
	}

	AppLog.Event("FIREBASE_BATCH_COMPLETE", map[string]any{
		"batch_id":    batchID,
		"doc_count":   len(docPaths),
		"duration_ms": time.Since(start).Milliseconds(),
	}, batchID,
		WithComponent(ComponentFirebase),
		WithOperation("batch_update"),
		WithLogKind(LogKindBusiness))

	return nil
}
