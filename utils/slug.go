package utils

import (
	"math/rand"
	"strings"
)

func GenerateUniqueSlug(name string) string {
	slug := Slugify(name, true)
	return slug
}

func Slugify(name string, unique bool) string {
	// unique -> true: add random number to slug
	if unique {
		slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		slug = strings.ReplaceAll(slug, "_", "-")
		return slug + "-" + GenerateRandomString(4)
	}
	// unique -> false: just slugify the name
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	slug = strings.ReplaceAll(slug, "_", "-")
	return slug
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	var result strings.Builder
	for i := 0; i < length; i++ {
		result.WriteByte(charset[GenerateRandomInt(len(charset))])
	}
	return result.String()
}

func GenerateRandomInt(max int) int {
	rand.Seed(rand.Int63())
	return rand.Intn(max)
}

func GenerateRandomFloat64() float64 {
	rand.Seed(rand.Int63())
	return rand.Float64()
}
