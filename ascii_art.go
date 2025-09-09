package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

// ASCII characters ordered from darkest to lightest
// Using a more detailed character set for better image representation
const asciiChars = "█▉▊▋▌▍▎▏▓▒░@%#*+=-:. "

// convertImageToASCII converts an image to ASCII art
func convertImageToASCII(imageData []byte, maxWidth, maxHeight, terminalWidth, terminalHeight int) (string, error) {
	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate scaling to fit within terminal dimensions
	// Account for character aspect ratio (characters are taller than wide, roughly 2:1)
	const charAspectRatio = 0.5 // Width/Height ratio of terminal characters

	// Adjust the image aspect ratio to account for character dimensions
	imageAspectRatio := float64(width) / float64(height)
	adjustedAspectRatio := imageAspectRatio / charAspectRatio

	var newWidth, newHeight int
	if adjustedAspectRatio > float64(maxWidth)/float64(maxHeight) {
		// Width-constrained
		newWidth = maxWidth
		newHeight = int(float64(maxWidth) / adjustedAspectRatio)
	} else {
		// Height-constrained
		newHeight = maxHeight
		newWidth = int(float64(maxHeight) * adjustedAspectRatio)
	}

	// Ensure minimum dimensions
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("┌─ Image: %dx%d (%s) ─┐\n", width, height, format))
	result.WriteString(fmt.Sprintf("├─ ASCII: %dx%d (term: %dx%d, max: %dx%d) ─┤\n", newWidth, newHeight, terminalWidth, terminalHeight, maxWidth, maxHeight))

	// Show coordinate ranges being sampled
	maxImgX := int(float64(newWidth-1) * float64(width) / float64(newWidth))
	maxImgY := int(float64(newHeight-1) * float64(height) / float64(newHeight))
	midImgX := int(float64(newWidth/2) * float64(width) / float64(newWidth))
	midImgY := int(float64(newHeight/2) * float64(height) / float64(newHeight))
	result.WriteString(fmt.Sprintf("├─ Sampling: X[0,%d,%d] Y[0,%d,%d] of %dx%d ─┤\n", midImgX, maxImgX, midImgY, maxImgY, width, height))
	result.WriteString("└" + strings.Repeat("─", newWidth+2) + "┘\n")

	// Convert to ASCII
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Map ASCII coordinates back to image coordinates
			imgX := int(float64(x) * float64(width) / float64(newWidth))
			imgY := int(float64(y) * float64(height) / float64(newHeight))

			// Get pixel color including alpha channel
			r, g, b, a := img.At(imgX, imgY).RGBA()

			// Handle transparency by blending with white background
			// If pixel is transparent, blend with white (65535, 65535, 65535)
			alpha := float64(a) / 65535.0
			r = uint32(float64(r)*alpha + 65535.0*(1.0-alpha))
			g = uint32(float64(g)*alpha + 65535.0*(1.0-alpha))
			b = uint32(float64(b)*alpha + 65535.0*(1.0-alpha))

			// Convert to grayscale (ITU-R BT.709 standard luminance formula)
			gray := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)

			// Normalize to 0-1 range
			intensity := gray / 65535.0

			// Apply gamma correction for better visual representation
			intensity = 1.0 - intensity // Invert for ASCII mapping

			// Map to ASCII character with better distribution
			charIndex := int(intensity * float64(len(asciiChars)))
			if charIndex >= len(asciiChars) {
				charIndex = len(asciiChars) - 1
			}

			result.WriteRune(rune(asciiChars[charIndex]))
		}
		result.WriteByte('\n')
	}

	return result.String(), nil
}

// isImageFile checks if the filename suggests it's an image file
func isImageFile(filename string) bool {
	filename = strings.ToLower(filename)
	return strings.HasSuffix(filename, ".jpg") ||
		strings.HasSuffix(filename, ".jpeg") ||
		strings.HasSuffix(filename, ".png") ||
		strings.HasSuffix(filename, ".gif") ||
		strings.HasSuffix(filename, ".bmp") ||
		strings.HasSuffix(filename, ".webp")
}

// isImageData checks if the data appears to be image data by examining magic bytes
func isImageData(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	// Check for common image file signatures
	// JPEG
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8 {
		return true
	}

	// PNG
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return true
	}

	// GIF
	if len(data) >= 6 && (string(data[0:6]) == "GIF87a" || string(data[0:6]) == "GIF89a") {
		return true
	}

	// BMP
	if len(data) >= 2 && data[0] == 0x42 && data[1] == 0x4D {
		return true
	}

	// WebP
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return true
	}

	return false
}

// convertToASCIIArt attempts to convert image data to ASCII art
// Returns the ASCII art string and a boolean indicating if conversion was successful
func convertToASCIIArt(data []byte, filename string, terminalWidth, terminalHeight int) (string, bool) {
	// Check if this is likely an image file
	if !isImageFile(filename) && !isImageData(data) {
		return "", false
	}

	// Calculate reasonable dimensions for ASCII art in terminal
	maxWidth := terminalWidth - 6   // Leave margin for border and scrollbar
	maxHeight := terminalHeight - 8 // Account for headers and UI elements

	// Ensure reasonable minimum sizes but be more responsive to terminal size
	if maxWidth < 20 {
		maxWidth = 20
	}
	if maxHeight < 10 {
		maxHeight = 10
	}

	// For extremely large terminals, apply reasonable limits to avoid performance issues
	// Only limit if terminal is exceptionally large (> 200 cols or > 100 rows)
	if terminalWidth > 200 && maxWidth > 180 {
		maxWidth = 180
	}
	if terminalHeight > 100 && maxHeight > 80 {
		maxHeight = 80
	}

	ascii, err := convertImageToASCII(data, maxWidth, maxHeight, terminalWidth, terminalHeight)
	if err != nil {
		return fmt.Sprintf("Error converting image to ASCII: %v", err), false
	}

	return ascii, true
}

// Enhanced image decoder that handles more formats
func init() {
	// Standard formats are automatically registered by importing the packages
	// Additional formats like BMP and WebP are registered by importing their packages with _
}
