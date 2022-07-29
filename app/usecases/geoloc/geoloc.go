package geoloc

import (
	"fmt"
	"strings"

	"github.com/biter777/countries"
)

var countriesMap = map[string]string{}

func CreateCountryList() {
	if len(countriesMap) == 0 {
		for _, country := range countries.All() {
			countriesMap[country.Alpha2()] = country.String()
		}
	}
}

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

	if _, found := countriesMap[s]; !found {
		return ""
	}

	return countriesMap[s]
}
