package common

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

func DownloadFile(filePath string, destfilePath string, appID string, bucket string) error {
	// bucket := "bucket-name"
	// object := "object-name"
	destFileName := destfilePath

	substrings := strings.Split(filePath, "/")
	newFilename := substrings[len(substrings)-1]

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		// return fmt.Errorf("storage.NewClient: %w", err)
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	f, err := os.Create(destFileName)
	if err != nil {
		// return fmt.Errorf("os.Create: %w", err)
		return err
	}

	// object := "vehicle/" + newFilename
	rc, err := client.Bucket(bucket).Object("vehicle/" + newFilename).NewReader(ctx)
	if err != nil {
		// return fmt.Errorf("Object(%q).NewReader: %w", object, err)
		return err
	}
	defer rc.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	if err = f.Close(); err != nil {
		// return fmt.Errorf("f.Close: %w", err)
		return err
	}

	// fmt.Fprintf(w, "Blob %v downloaded to local file %v\n", object, destFileName)
	// common.Print("downloaded to local file "+object, destFileName)

	return nil
}
