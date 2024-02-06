package common

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

func UploadFileToGoogle(filePath string, appID string, bucket string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Get the file size
	stat, err := file.Stat()
	if err != nil {
		return "", err
	}

	// Read the file into a byte slice
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		return "", err
	}
	// fmt.Println(bs)

	substrings := strings.Split(filePath, "/")
	newFilename := substrings[len(substrings)-1]
	month := fmt.Sprintf("%02d", int(time.Now().Month()))
	year := fmt.Sprintf("%04d", int(time.Now().Year()))
	savePath := fmt.Sprintf("uploads/%s/%s/%s", year, month, newFilename)

	//os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "google.json")
	// bucket := "bucket-name"
	// object := "object-name"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	b := []byte(bs)
	buf := bytes.NewBuffer(b)

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := client.Bucket(bucket).Object(savePath).NewWriter(ctx)
	wc.ChunkSize = 0 // note retries are not supported for chunk size 0.

	if _, err = io.Copy(wc, buf); err != nil {
		return "", err
	}
	// Data can continue to be added to the file until the writer is closed.
	if err := wc.Close(); err != nil {
		return "", err
		// return fmt.Errorf("Writer.Close: %w", err)
	}
	// fmt.Fprintf(, "%v uploaded to %v.\n", object, bucket)
	resultPath := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, savePath)
	// common.Print("last pdf to google path", resultPath)

	return resultPath, nil
}
