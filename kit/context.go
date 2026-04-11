package kit

import (
	"context"
	"time"

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

func FirebaseContextFrom(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, DefaultFirebaseTimeout)
}

func DatabaseContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), DefaultDatabaseTimeout)
}

func NewRequestID() string {
	return uuid.NewString()
}
