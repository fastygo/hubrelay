package content

import (
	corelocale "github.com/fastygo/hubcore/locale"
	"hubrelay-dashboard/fixtures"
)

func Load(locale string) (Catalog, error) {
	return corelocale.Load(locale, fixtures.Decode)
}
