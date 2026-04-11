package kit

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type FirestoreClient struct {
	Client *firestore.Client
	Auth   *firebaseAuth.Client
}

func NewFirestoreClient(ctx context.Context, credentialsFile string) (*FirestoreClient, error) {
	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("creating firebase app: %w", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating firestore client: %w", err)
	}

	auth, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating auth client: %w", err)
	}

	return &FirestoreClient{Client: client, Auth: auth}, nil
}

func (fc *FirestoreClient) Close() error {
	return fc.Client.Close()
}

func (fc *FirestoreClient) Set(ctx context.Context, path string, data map[string]interface{}) error {
	_, err := fc.Client.Doc(path).Set(ctx, data, firestore.MergeAll)
	return err
}

func (fc *FirestoreClient) Delete(ctx context.Context, docPath string) error {
	_, err := fc.Client.Doc(docPath).Delete(ctx)
	return err
}

func (fc *FirestoreClient) BatchSet(ctx context.Context, docs map[string]map[string]interface{}) error {
	if len(docs) == 0 {
		return nil
	}

	bw := fc.Client.BulkWriter(ctx)

	var firstErr error
	for path, data := range docs {
		docRef := fc.Client.Doc(path)
		if _, err := bw.Set(docRef, data, firestore.MergeAll); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	bw.Flush()
	bw.End()
	return firstErr
}

func (fc *FirestoreClient) BatchSetSameData(ctx context.Context, paths []string, data map[string]interface{}) error {
	if len(paths) == 0 {
		return nil
	}

	docs := make(map[string]map[string]interface{}, len(paths))
	for _, path := range paths {
		docs[path] = data
	}
	return fc.BatchSet(ctx, docs)
}

func IteratorCount(iter *firestore.DocumentIterator) (int, error) {
	count := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
