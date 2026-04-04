package shell

import (
	"net/url"
	"strings"

	ui8layout "github.com/fastygo/ui8kit/layout"
	hubcore "github.com/fastygo/hubcore"
)

type ContentNavItem struct {
	Path  string
	Label string
	Icon  string
}

func LocalizePath(defaultLocale, locale, raw string) string {
	if locale == defaultLocale || strings.TrimSpace(raw) == "" {
		return raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	query := parsed.Query()
	query.Set("lang", locale)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func ResolveNavItems(defaultLocale, locale string, contentItems []ContentNavItem, moduleItems []hubcore.NavItem) []ui8layout.NavItem {
	if len(moduleItems) == 0 {
		navItems := make([]ui8layout.NavItem, 0, len(contentItems))
		for _, item := range contentItems {
			navItems = append(navItems, ui8layout.NavItem{
				Path:  LocalizePath(defaultLocale, locale, item.Path),
				Label: item.Label,
				Icon:  item.Icon,
			})
		}
		return navItems
	}

	localized := make(map[string]ContentNavItem, len(contentItems))
	for _, item := range contentItems {
		localized[item.Path] = item
	}

	navItems := make([]ui8layout.NavItem, 0, len(moduleItems))
	for _, item := range moduleItems {
		label := strings.TrimSpace(item.Label)
		icon := strings.TrimSpace(item.Icon)
		if localizedItem, ok := localized[item.Path]; ok {
			if label == "" {
				label = localizedItem.Label
			}
			if icon == "" {
				icon = localizedItem.Icon
			}
		}

		navItems = append(navItems, ui8layout.NavItem{
			Path:  LocalizePath(defaultLocale, locale, item.Path),
			Label: label,
			Icon:  icon,
		})
	}

	return navItems
}
