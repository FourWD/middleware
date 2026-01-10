package common

import (
	"time"

	"cloud.google.com/go/firestore"

	"github.com/google/uuid"
)

func FirebaseBatchUpdate(docPaths map[string]map[string]interface{}) error {
	requestID := uuid.NewString()

	Log("FirebaseBatchUpdate Start", map[string]interface{}{
		"total_docs": len(docPaths),
	}, requestID)

	if len(docPaths) == 0 {
		return nil
	}

	ctx, cancel := ContextWithTimeout(60 * time.Second)
	defer cancel()

	bw := FirebaseClient.BulkWriter(ctx)

	var firstErr error
	failCount := 0

	for path, updateData := range docPaths {
		Log("FirebaseBatchUpdate Add", map[string]interface{}{
			"path": path,
			"data": updateData,
		}, requestID)

		docRef := FirebaseClient.Doc(path)
		_, err := bw.Set(docRef, updateData, firestore.MergeAll)
		if err != nil {
			Log("FirebaseBatchUpdate Set Error", map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			}, requestID)
			if firstErr == nil {
				firstErr = err
			}
			failCount++
			continue
		}
	}

	bw.Flush()
	bw.End()

	successCount := len(docPaths) - failCount

	Log("FirebaseBatchUpdate Complete", map[string]interface{}{
		"total_docs":    len(docPaths),
		"success_count": successCount,
		"fail_count":    failCount,
	}, requestID)

	return firstErr
}
