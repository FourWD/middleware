package common

import (
	"image"
	"image/jpeg"
	"net/http"
	"os"

	"github.com/nfnt/resize"
)

func ResizeImage(imageURL string, width int, quality int) (string, error) {
	// Fetch the image from the URL
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", err
	}

	// Resize the image
	resizedImg := resize.Resize(uint(width), 0, img, resize.Lanczos3)

	// Create a temporary file
	tempFile, err := os.CreateTemp("", "resized-*.jpg")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Save the resized image to the temporary file
	options := &jpeg.Options{Quality: quality}
	err = jpeg.Encode(tempFile, resizedImg, options)
	if err != nil {
		return "", err
	}

	// Return the path to the temporary file
	//log.Println("Original: ", imageURL)
	//log.Println("Resize: ", tempFile.Name())
	return tempFile.Name(), nil
}
