package content

import (
	"fmt"
	"path"
	"strings"

	"hubrelay-dashboard/fixtures"
)

func Load(locale string) (Catalog, error) {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return Catalog{}, fmt.Errorf("locale must not be empty")
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
		if err := fixtures.Decode(path.Join(locale, loader.name), loader.target); err != nil {
			return Catalog{}, err
		}
	}

	return catalog, nil
}
