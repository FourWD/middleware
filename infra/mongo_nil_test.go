package infra

import (
	"context"
	"testing"
)

func TestMongoClientNilReceiver(t *testing.T) {
	var mc *MongoClient
	if err := mc.Close(context.Background()); err != nil {
		t.Fatalf("Close on nil = %v, want nil", err)
	}
	if c := mc.Collection("x"); c != nil {
		t.Fatalf("Collection on nil = %v, want nil", c)
	}
	if c := (&MongoClient{}).Collection("x"); c != nil {
		t.Fatalf("Collection with nil database = %v, want nil", c)
	}
}
