package content

import "gitcourse/fixtures"

const DefaultLocale = "en"

func AvailableLocales() ([]string, error) {
	return fixtures.Locales()
}
