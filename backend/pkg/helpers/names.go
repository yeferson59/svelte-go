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

	return strings.ToUpper(names[0][:1]) + names[0][1:] + " " + strings.ToUpper(names[1][:1]) + names[1][1:]
}

func CalculateTotalPages(count uint, limit uint) uint {
	return uint(math.Ceil(float64(count) / float64(limit)))
}
