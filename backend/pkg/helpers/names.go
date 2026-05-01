package helpers

import "strings"

func NormalizateNames(name string) string {
	nom := strings.TrimSpace(name)

	if !strings.Contains(nom, " ") {
		return strings.ToUpper(nom[:1]) + nom[1:]
	}

	names := strings.Split(nom, " ")

	return strings.ToUpper(names[0][:1]) + names[0][1:] + " " + strings.ToUpper(names[1][:1]) + names[1][1:]
}
