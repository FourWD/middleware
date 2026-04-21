package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/FourWD/middleware/kit"
	"google.golang.org/api/iterator"
)

// FirebaseValueToInt coerces an arbitrary value to int via fmt.Sprintf.
// Returns 0 on failure.
func FirebaseValueToInt(a interface{}) int {
	str := fmt.Sprintf("%d", a)
	intValue, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return intValue
}

// FirebaseUpdate merges data into a Firestore document at the given path.
func FirebaseUpdate(client *firestore.Client, path string, updateData map[string]interface{}) error {
	_, err := client.Doc(path).Set(context.Background(), updateData, firestore.MergeAll)
	if err != nil {
		AppLog.EventError(err, "FirebaseUpdate", map[string]interface{}{
			"path": path,
			"data": updateData,
		}, "")
	}
	return err
}

// FirebaseSaveBySqlLimit1 runs a SQL query that must return exactly one row,
// then merges that row into the Firestore document at the given path.
// Uses the infra-level DatabaseSql global populated by MigrateInfra.
func FirebaseSaveBySqlLimit1(client *firestore.Client, path string, sql string, values ...interface{}) error {
	if DatabaseSql == nil {
		return errors.New("DatabaseSql not initialized")
	}

	rows, err := DatabaseSql.Query(sql, values...)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	var result []map[string]interface{}
	for rows.Next() {
		scanValues := make([]interface{}, len(columns))
		scanPtrs := make([]interface{}, len(columns))
		for i := range columns {
			scanPtrs[i] = &scanValues[i]
		}
		if err := rows.Scan(scanPtrs...); err != nil {
			return err
		}
		row := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			if b, ok := scanValues[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = scanValues[i]
			}
		}
		result = append(result, row)
	}

	// Preserve original byte-level semantics by round-tripping through json
	// (keeps the caller's expectation for types like timestamps unchanged).
	if _, err := json.Marshal(result); err != nil {
		return err
	}

	if len(result) != 1 {
		return errors.New("total row != 1")
	}

	return FirebaseUpdate(client, path, result[0])
}

// FirebaseDelete removes a Firestore document at the given path.
func FirebaseDelete(client *firestore.Client, docPath string) error {
	_, err := client.Doc(docPath).Delete(context.Background())
	return err
}

// FirebaseCount iterates the iterator to its end and returns the total count.
// Iteration errors are logged but do not stop counting.
func FirebaseCount(documents *firestore.DocumentIterator) int {
	count := 0
	for {
		_, err := documents.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			AppLog.EventError(err, "FIREBASE_ITERATOR_ERROR", nil, "")
		}
		count++
	}
	return count
}

// FirebaseCountByField iterates the documents and returns the number of unique
// string values observed at the given field.
func FirebaseCountByField(documents *firestore.DocumentIterator, groupByField string) int {
	uniqueValues := []string{}

	for {
		doc, err := documents.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			AppLog.EventError(err, "FIREBASE_ITERATOR_ERROR", nil, "")
		}

		fieldValue, ok := doc.Data()[groupByField]
		if !ok {
			AppLog.EventWarn("FIREBASE_FIELD_NOT_FOUND", map[string]interface{}{"field": groupByField}, "")
		}

		fieldStr, ok := fieldValue.(string)
		if !ok {
			AppLog.EventWarn("FIREBASE_FIELD_CONVERT_ERROR", map[string]interface{}{"field": groupByField}, "")
		}

		if !kit.StringExistsInList(fieldStr, uniqueValues) {
			uniqueValues = append(uniqueValues, fieldStr)
		}
	}

	return len(uniqueValues)
}
