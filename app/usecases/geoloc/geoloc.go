package geoloc

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/pariz/gountries"
)

var countrySelector = gountries.New()

func CountryToCodepoints(isoCode string) string {
	var charList []string
	isoCode = strings.ToUpper(isoCode)

	for _, char := range isoCode {
		charList = append(charList, fmt.Sprintf("%x", int(char)+127397))
	}
	return strings.Join(charList, "-")
}

func CountryReadable(s string) string {
	s = strings.ToUpper(s)
	if s == "XX" || s == "" {
		return ""
	}

	country, err := countrySelector.FindCountryByAlpha(s)

	if err != nil {
		slog.Error("error", "Could not find country!", err)
		return ""
	}

	return country.Name.Common
}
