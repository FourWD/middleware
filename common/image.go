package common

import (
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"

	"github.com/disintegration/imaging"
)

func ImageResizeByRatio(imageURL string, aspectRatio float64) (image.Image, error) {
	// Step 1: Download the image
	response, err := http.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %v", err)
	}
	defer response.Body.Close()

	// Decode the image
	img, _, err := image.Decode(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Step 2: Get current dimensions
	origWidth := img.Bounds().Dx()
	origHeight := img.Bounds().Dy()

	// Step 3: Calculate the target width based on the desired aspect ratio
	targetWidth := int(float64(origHeight) * aspectRatio)

	// Step 4: Ensure the target width is smaller than the original width to avoid stretching
	if targetWidth > origWidth {
		return nil, fmt.Errorf("the target aspect ratio requires a wider image than the original width")
	}

	// Step 5: Crop the image from the center (left and right sides)
	croppedImg := imaging.CropCenter(img, targetWidth, origHeight)

	return croppedImg, nil
}

// SaveImage saves the image to a file
func SaveImage(img image.Image, filePath string) error {
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	err = jpeg.Encode(out, img, nil)
	if err != nil {
		return fmt.Errorf("failed to encode and save image: %v", err)
	}

	return nil
}
