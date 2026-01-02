package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"cloud.google.com/go/firestore"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/iterator"
)

var FirebaseCtx context.Context
var FirebaseClient *firestore.Client
var AuthClient *firebaseAuth.Client

func ConnectFirebase(key string) {
	FirebaseCtx = context.Background()

	app, err := initFirebaseApp(key)
	if err != nil {
		panic(err)
	}

	FirebaseClient, err = app.Firestore(FirebaseCtx)
	if err != nil {
		LogError("FIREBASE_FIRESTORE_ERROR", map[string]interface{}{"error": err.Error()}, "")
		panic(err)
	}

	AuthClient, err = app.Auth(context.Background())
	if err != nil {
		LogError("FIREBASE_AUTH_CLIENT_ERROR", map[string]interface{}{"error": err.Error()}, "")
	}
}

func FirebaseValueToInt(a interface{}) int {
	str := fmt.Sprintf("%d", a)
	intValue, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return intValue
}

func FirebaseUpdate(client *firestore.Client, path string, updateData map[string]interface{}) error {
	_, err := client.Doc(path).Set(context.Background(), updateData, firestore.MergeAll)
	if err != nil {
		updateData["path"] = path
		updateData["error"] = err
		LogError("FirebaseUpdate", updateData, "")
	}
	return err
}

func FirebaseSaveBySqlLimit1(client *firestore.Client, path string, sql string, values ...interface{}) error {
	jsonBytes, _, err := queryToJSON(DatabaseSql, sql, values...)
	var result []map[string]interface{}
	if json.Unmarshal(jsonBytes, &result); err != nil {
		return err
	}

	if len(result) != 1 {
		return errors.New("total row != 1")
	}

	firstRow := result[0]
	return FirebaseUpdate(client, path, firstRow)
}

func FirebaseDelete(client *firestore.Client, docPath string) error {
	_, err := client.Doc(docPath).Delete(FirebaseCtx)
	return err
}

func FirebaseCount(documents *firestore.DocumentIterator) int {
	count := 0
	for {
		_, err := documents.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			LogError("FIREBASE_ITERATOR_ERROR", map[string]interface{}{"error": err.Error()}, "")
		}
		count++
	}
	return count
}

func FirebaseCountByField(documents *firestore.DocumentIterator, groupByField string) int {
	uniqueValues := []string{}

	for {
		doc, err := documents.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			LogError("FIREBASE_ITERATOR_ERROR", map[string]interface{}{"error": err.Error()}, "")
		}

		fieldValue, ok := doc.Data()[groupByField]
		if !ok {
			LogWarning("FIREBASE_FIELD_NOT_FOUND", map[string]interface{}{"field": groupByField}, "")
		}

		fieldStr, ok := fieldValue.(string)
		if !ok {
			LogWarning("FIREBASE_FIELD_CONVERT_ERROR", map[string]interface{}{"field": groupByField}, "")
		}

		if !StringExistsInList(fieldStr, uniqueValues) {
			uniqueValues = append(uniqueValues, fieldStr)
		}
	}

	return len(uniqueValues)
}
