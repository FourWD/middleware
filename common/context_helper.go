package common

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

const (
	DefaultFirebaseTimeout = 10 * time.Second
	DefaultDatabaseTimeout = 30 * time.Second
	ShortTimeout           = 5 * time.Second
	LongTimeout            = 60 * time.Second
)

func ContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

func FirebaseContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), DefaultFirebaseTimeout)
}

func DatabaseContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), DefaultDatabaseTimeout)
}

// NewRequestID generates a new unique request ID for logging and tracing
func NewRequestID() string {
	return uuid.NewString()
}

// FirebaseGetDoc gets a document from Firebase with default timeout
func FirebaseGetDoc(docPath string) (*firestore.DocumentSnapshot, error) {
	ctx, cancel := FirebaseContext()
	defer cancel()

	docRef := FirebaseClient.Doc(docPath)
	return docRef.Get(ctx)
}
