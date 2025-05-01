package utils

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"sort"
)

func GetImageFromURL(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

func GetImageSize(url string) (int, int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	return width, height, nil
}

func GetImageDominantColors(img image.Image) ([]string, error) {
	if img == nil {
		return nil, fmt.Errorf("nil image provided")
	}

	// Get image bounds
	bounds := img.Bounds()

	// Create a map to count color occurrences
	colorCount := make(map[uint32]int)

	// Sample pixels (sample every 5th pixel to improve performance)
	sampleRate := 5
	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleRate {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleRate {
			r, g, b, _ := img.At(x, y).RGBA()
			// Convert to 8-bit per channel
			r, g, b = r>>8, g>>8, b>>8
			// Create a single key for the RGB color
			colorKey := (r << 16) | (g << 8) | b
			colorCount[colorKey]++
		}
	}

	// Sort colors by occurrence
	type colorFreq struct {
		color uint32
		count int
	}

	colorFreqs := make([]colorFreq, 0, len(colorCount))
	for color, count := range colorCount {
		colorFreqs = append(colorFreqs, colorFreq{color, count})
	}

	// Sort by count (descending)
	sort.Slice(colorFreqs, func(i, j int) bool {
		return colorFreqs[i].count > colorFreqs[j].count
	})

	// Take top 5 colors (or fewer if there aren't 5)
	numColors := 5
	if len(colorFreqs) < numColors {
		numColors = len(colorFreqs)
	}

	// Convert to hex strings
	result := make([]string, numColors)
	for i := 0; i < numColors; i++ {
		color := colorFreqs[i].color
		r := (color >> 16) & 0xFF
		g := (color >> 8) & 0xFF
		b := color & 0xFF
		result[i] = fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}

	return result, nil
}
