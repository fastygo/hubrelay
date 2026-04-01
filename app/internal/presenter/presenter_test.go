package presenter

import (
	"testing"

	"hubrelay-dashboard/internal/content"
)

func TestAskPageUsesLocalizedPaths(t *testing.T) {
	locales, err := content.AvailableLocales()
	if err != nil {
		t.Fatalf("AvailableLocales() error = %v", err)
	}

	catalogs := make(map[string]content.Catalog, len(locales))
	for _, locale := range locales {
		catalogs[locale], err = content.Load(locale)
		if err != nil {
			t.Fatalf("Load(%s) error = %v", locale, err)
		}
	}

	enCatalog, err := content.Load("en")
	if err != nil {
		t.Fatalf("Load(en) error = %v", err)
	}

	enPage := New(enCatalog, Config{
		DefaultLocale:    content.DefaultLocale,
		AvailableLocales: locales,
		ToggleLocales:    locales[:2],
		Catalogs:         catalogs,
	}).AskPage(AskPageState{})
	if enPage.FormAction != "/ask" {
		t.Fatalf("expected english form action /ask, got %q", enPage.FormAction)
	}
	if enPage.StreamEndpoint != "/ask/stream" {
		t.Fatalf("expected english stream endpoint /ask/stream, got %q", enPage.StreamEndpoint)
	}
	if enPage.Layout.NavItems[0].Path != "/" {
		t.Fatalf("expected english home nav path /, got %q", enPage.Layout.NavItems[0].Path)
	}

	ruCatalog, err := content.Load("ru")
	if err != nil {
		t.Fatalf("Load(ru) error = %v", err)
	}

	ruPage := New(ruCatalog, Config{
		DefaultLocale:    content.DefaultLocale,
		AvailableLocales: locales,
		ToggleLocales:    locales[:2],
		Catalogs:         catalogs,
	}).AskPage(AskPageState{})
	if ruPage.FormAction != "/ask?lang=ru" {
		t.Fatalf("expected russian form action /ask?lang=ru, got %q", ruPage.FormAction)
	}
	if ruPage.StreamEndpoint != "/ask/stream?lang=ru" {
		t.Fatalf("expected russian stream endpoint /ask/stream?lang=ru, got %q", ruPage.StreamEndpoint)
	}
	if ruPage.Layout.NavItems[0].Path != "/?lang=ru" {
		t.Fatalf("expected russian home nav path /?lang=ru, got %q", ruPage.Layout.NavItems[0].Path)
	}
	if ruPage.Layout.LanguageToggle.CurrentLabel != "RU" {
		t.Fatalf("expected russian toggle label RU, got %q", ruPage.Layout.LanguageToggle.CurrentLabel)
	}
	if !ruPage.Layout.LanguageToggle.Enabled {
		t.Fatal("expected language toggle enabled")
	}
	if enPage.Layout.LanguageToggle.NextLocale != "es" {
		t.Fatalf("expected english toggle to switch to es, got %q", enPage.Layout.LanguageToggle.NextLocale)
	}
	if enPage.Layout.ThemeToggle.SwitchToDarkLabel != "Switch to dark theme" {
		t.Fatalf("unexpected english theme dark label %q", enPage.Layout.ThemeToggle.SwitchToDarkLabel)
	}

	esCatalog, err := content.Load("es")
	if err != nil {
		t.Fatalf("Load(es) error = %v", err)
	}

	esPage := New(esCatalog, Config{
		DefaultLocale:    content.DefaultLocale,
		AvailableLocales: locales,
		ToggleLocales:    locales[:2],
		Catalogs:         catalogs,
	}).AskPage(AskPageState{})
	if esPage.FormAction != "/ask?lang=es" {
		t.Fatalf("expected spanish form action /ask?lang=es, got %q", esPage.FormAction)
	}
	if esPage.Layout.LanguageToggle.CurrentLabel != "ES" {
		t.Fatalf("expected spanish toggle label ES, got %q", esPage.Layout.LanguageToggle.CurrentLabel)
	}
	if esPage.Layout.ThemeToggle.SwitchToLightLabel != "Cambiar al tema claro" {
		t.Fatalf("unexpected spanish theme light label %q", esPage.Layout.ThemeToggle.SwitchToLightLabel)
	}
}
