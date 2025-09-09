package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestIsImageFile(t *testing.T) {
	testCases := []struct {
		filename string
		expected bool
	}{
		{"test.jpg", true},
		{"test.jpeg", true},
		{"test.png", true},
		{"test.gif", true},
		{"test.bmp", true},
		{"test.webp", true},
		{"test.JPG", true}, // Case insensitive
		{"test.txt", false},
		{"test.pdf", false},
		{"test", false},
		{"image.png.txt", false}, // Should not match
		{"", false},
	}

	for _, tc := range testCases {
		result := isImageFile(tc.filename)
		if result != tc.expected {
			t.Errorf("isImageFile(%q) = %v, expected %v", tc.filename, result, tc.expected)
		}
	}
}

func TestIsImageData(t *testing.T) {
	// Create test data with known magic bytes
	testCases := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "JPEG magic bytes",
			data:     []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10},
			expected: true,
		},
		{
			name:     "PNG magic bytes",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			expected: true,
		},
		{
			name:     "GIF87a magic bytes",
			data:     []byte("GIF87a"),
			expected: true,
		},
		{
			name:     "GIF89a magic bytes",
			data:     []byte("GIF89a"),
			expected: true,
		},
		{
			name:     "BMP magic bytes",
			data:     []byte{0x42, 0x4D, 0x36, 0x48},
			expected: true,
		},
		{
			name:     "WebP magic bytes",
			data:     []byte("RIFF\x24\x08\x00\x00WEBP"),
			expected: true,
		},
		{
			name:     "Text data",
			data:     []byte("Hello, world!"),
			expected: false,
		},
		{
			name:     "Empty data",
			data:     []byte{},
			expected: false,
		},
		{
			name:     "Short data",
			data:     []byte{0x42},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isImageData(tc.data)
			if result != tc.expected {
				t.Errorf("isImageData(%q) = %v, expected %v", tc.name, result, tc.expected)
			}
		})
	}
}

func createTestImage(width, height int) ([]byte, error) {
	// Create a simple test image with a gradient pattern
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a simple gradient pattern
			gray := uint8((x + y) * 255 / (width + height))
			img.Set(x, y, color.RGBA{gray, gray, gray, 255})
		}
	}

	// Encode to PNG
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func TestConvertImageToASCII(t *testing.T) {
	// Create a small test image
	imageData, err := createTestImage(20, 10)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// Test conversion
	ascii, err := convertImageToASCII(imageData, 40, 20, 80, 25)
	if err != nil {
		t.Fatalf("Failed to convert image to ASCII: %v", err)
	}

	// Basic checks
	if ascii == "" {
		t.Error("Expected non-empty ASCII art, got empty string")
	}

	// Check that the output contains expected elements
	if !bytes.Contains([]byte(ascii), []byte("Image:")) {
		t.Error("Expected ASCII output to contain image info header")
	}

	if !bytes.Contains([]byte(ascii), []byte("ASCII:")) {
		t.Error("Expected ASCII output to contain ASCII info header")
	}

	// Count lines to verify approximate dimensions
	lines := bytes.Split([]byte(ascii), []byte("\n"))
	if len(lines) < 5 { // Header lines + some image lines
		t.Errorf("Expected at least 5 lines in ASCII output, got %d", len(lines))
	}
}

func TestConvertToASCIIArt(t *testing.T) {
	// Test with valid image data
	imageData, err := createTestImage(10, 10)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	ascii, isImage := convertToASCIIArt(imageData, "test.png", 80, 25)
	if !isImage {
		t.Error("Expected convertToASCIIArt to detect image data")
	}
	if ascii == "" {
		t.Error("Expected non-empty ASCII art")
	}

	// Test with non-image data
	textData := []byte("This is just text content")
	ascii, isImage = convertToASCIIArt(textData, "test.txt", 80, 25)
	if isImage {
		t.Error("Expected convertToASCIIArt to not detect text as image")
	}

	// Test with image filename but non-image data
	ascii, isImage = convertToASCIIArt(textData, "fake.jpg", 80, 25)
	if isImage {
		t.Error("Expected convertToASCIIArt to not convert invalid image data")
	}
}

func TestConvertImageToASCIIErrors(t *testing.T) {
	// Test with invalid image data
	invalidData := []byte("This is not image data")
	_, err := convertImageToASCII(invalidData, 40, 20, 80, 25)
	if err == nil {
		t.Error("Expected error when converting invalid image data")
	}

	// Test with empty data
	_, err = convertImageToASCII([]byte{}, 40, 20, 80, 25)
	if err == nil {
		t.Error("Expected error when converting empty data")
	}
}

func TestConvertImageToASCIISmallDimensions(t *testing.T) {
	// Create a small test image
	imageData, err := createTestImage(2, 2)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	// Test with very small output dimensions
	ascii, err := convertImageToASCII(imageData, 1, 1, 80, 25)
	if err != nil {
		t.Fatalf("Failed to convert small image: %v", err)
	}

	// Should still produce valid output
	if ascii == "" {
		t.Error("Expected non-empty ASCII art even for small dimensions")
	}
}

// Benchmark the ASCII conversion performance
func BenchmarkConvertImageToASCII(b *testing.B) {
	// Create a moderately sized test image
	imageData, err := createTestImage(100, 100)
	if err != nil {
		b.Fatalf("Failed to create test image: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := convertImageToASCII(imageData, 80, 40, 120, 50)
		if err != nil {
			b.Fatalf("Conversion failed: %v", err)
		}
	}
}
