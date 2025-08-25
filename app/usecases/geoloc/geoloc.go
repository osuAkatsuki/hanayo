package geoloc

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slog"

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

	if s == "XK" {
		return "Kosovo"
	}

	country, err := countrySelector.FindCountryByAlpha(s)

	if err != nil {
		slog.Error("Could not find country", "error", err)
		return ""
	}

	return country.Name.Common
}
