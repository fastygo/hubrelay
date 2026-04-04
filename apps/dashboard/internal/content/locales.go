package content

import (
	corelocale "github.com/fastygo/hubcore/locale"

	"hubrelay-dashboard/fixtures"
)

const DefaultLocale = corelocale.DefaultLocale

func AvailableLocales() ([]string, error) {
	return corelocale.AvailableLocales(fixtures.Locales)
}

func OrderLocales(locales []string) []string {
	return corelocale.OrderLocales(locales)
}
