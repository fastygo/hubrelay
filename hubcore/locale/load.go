package locale

import (
	"fmt"
	"path"
	"slices"
	"strings"
)

const DefaultLocale = "en"

type DecodeFunc func(name string, target any) error

type LocalesFunc func() ([]string, error)

func Load(locale string, decode DecodeFunc) (Catalog, error) {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return Catalog{}, fmt.Errorf("locale must not be empty")
	}
	if decode == nil {
		return Catalog{}, fmt.Errorf("decode function must not be nil")
	}

	catalog := Catalog{Locale: locale}
	loaders := []struct {
		name   string
		target any
	}{
		{name: "common.json", target: &catalog.Common},
		{name: "shell.json", target: &catalog.Shell},
		{name: "login.json", target: &catalog.Login},
		{name: "health.json", target: &catalog.Health},
		{name: "capabilities.json", target: &catalog.Capabilities},
		{name: "ask.json", target: &catalog.Ask},
		{name: "egress.json", target: &catalog.Egress},
		{name: "audit.json", target: &catalog.Audit},
	}

	for _, loader := range loaders {
		if err := decode(path.Join(locale, loader.name), loader.target); err != nil {
			return Catalog{}, err
		}
	}

	return catalog, nil
}

func AvailableLocales(locales LocalesFunc) ([]string, error) {
	if locales == nil {
		return nil, fmt.Errorf("locales function must not be nil")
	}

	items, err := locales()
	if err != nil {
		return nil, err
	}
	if !slices.Contains(items, DefaultLocale) {
		return nil, fmt.Errorf("default locale %q is missing from fixtures", DefaultLocale)
	}
	return OrderLocales(items), nil
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
