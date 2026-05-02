package helpers

import (
	"math"
	"strings"
)

func NormalizateNames(name string) string {
	nom := strings.TrimSpace(name)

	if !strings.Contains(nom, " ") {
		return strings.ToUpper(nom[:1]) + nom[1:]
	}

	names := strings.Split(nom, " ")

	for i, name := range names {
		name = strings.TrimSpace(name)
		names[i] = strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
	}

	return strings.Join(names, " ")
}

func CalculateTotalPages(count uint, limit uint) uint {
	return uint(math.Ceil(float64(count) / float64(limit)))
}
