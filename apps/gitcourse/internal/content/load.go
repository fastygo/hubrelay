package content

import (
	"path"

	"gitcourse/fixtures"
)

func Load(locale string) (Catalog, error) {
	var catalog Catalog
	if err := fixtures.Decode(path.Join(locale, "common.json"), &catalog.Common); err != nil {
		return Catalog{}, err
	}
	if err := fixtures.Decode(path.Join(locale, "login.json"), &catalog.Login); err != nil {
		return Catalog{}, err
	}
	if err := fixtures.Decode(path.Join(locale, "ask.json"), &catalog.Ask); err != nil {
		return Catalog{}, err
	}
	if err := fixtures.Decode(path.Join(locale, "course.json"), &catalog.Course); err != nil {
		return Catalog{}, err
	}
	return catalog, nil
}
