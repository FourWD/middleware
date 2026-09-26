package infra

import (
	"context"
	"strings"
	"testing"

	"cloud.google.com/go/pubsub/v2"
)

func TestProcessRecoveredPanicIsContained(t *testing.T) {
	buf := CaptureLogs(t)
	msg := &pubsub.Message{}

	ok := processRecovered(context.Background(), msg, func(*pubsub.Message) { panic("boom") })
	if ok {
		t.Fatal("want ok=false after a panic")
	}
	if !strings.Contains(buf.String(), "PUBSUB_HANDLER_PANIC") {
		t.Fatalf("panic not logged: %s", buf.String())
	}

	if !processRecovered(context.Background(), msg, func(*pubsub.Message) {}) {
		t.Fatal("want ok=true when process returns normally")
	}
}
