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
	destFileName := destfilePath

	substrings := strings.Split(filePath, "/")
	newFilename := substrings[len(substrings)-1]

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	f, err := os.Create(destFileName)
	if err != nil {
		return err
	}

	rc, err := client.Bucket(bucket).Object("vehicle/" + newFilename).NewReader(ctx)
	if err != nil {
		return err
	}
	defer rc.Close()

	if _, err := io.Copy(f, rc); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	if err = f.Close(); err != nil {
		return err
	}

	return nil
}
