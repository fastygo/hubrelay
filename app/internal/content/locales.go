package content

import (
	"fmt"
	"slices"

	"hubrelay-dashboard/fixtures"
)

const DefaultLocale = "en"

func AvailableLocales() ([]string, error) {
	locales, err := fixtures.Locales()
	if err != nil {
		return nil, err
	}
	if !slices.Contains(locales, DefaultLocale) {
		return nil, fmt.Errorf("default locale %q is missing from fixtures", DefaultLocale)
	}
	return OrderLocales(locales), nil
}

func OrderLocales(locales []string) []string {
	ordered := make([]string, 0, len(locales))
	if slices.Contains(locales, DefaultLocale) {
		ordered = append(ordered, DefaultLocale)
	}
	for _, locale := range locales {
		if locale == DefaultLocale {
			continue
		}
		ordered = append(ordered, locale)
	}
	return ordered
}
