package common

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var FirebaseCtx context.Context
var FirebaseClient *firestore.Client

func ConnectFirebase(key string) {
	FirebaseCtx = context.Background()
	sa := option.WithCredentialsFile(key)
	app, err := firebase.NewApp(FirebaseCtx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	FirebaseClient, err = app.Firestore(FirebaseCtx)
	if err != nil {
		log.Fatalln(err)
	}
	// defer client.Close()
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
			log.Fatalf("Error iterating over documents: %v", err)
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
			log.Fatalf("Error iterating over documents: %v", err)
		}

		fieldValue, ok := doc.Data()[groupByField]
		if !ok {
			log.Fatalf("Document does not have the specified field: %s", groupByField)
		}

		fieldStr, ok := fieldValue.(string)
		if !ok {
			log.Fatalf("Unable to convert field value to string")
		}

		if !StringExistsInList(fieldStr, uniqueValues) {
			uniqueValues = append(uniqueValues, fieldStr)
		}
	}

	return len(uniqueValues)
}
