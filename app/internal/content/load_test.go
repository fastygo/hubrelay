package content

import "testing"

func TestAvailableLocales(t *testing.T) {
	locales, err := AvailableLocales()
	if err != nil {
		t.Fatalf("AvailableLocales() error = %v", err)
	}
	if len(locales) < 3 {
		t.Fatalf("expected at least three locales, got %v", locales)
	}
	if locales[0] != DefaultLocale {
		t.Fatalf("expected default locale first, got %v", locales)
	}
	if locales[1] != "es" {
		t.Fatalf("expected spanish locale second for proof-of-concept, got %v", locales)
	}
}

func TestLoadEnglishCatalog(t *testing.T) {
	catalog, err := Load("en")
	if err != nil {
		t.Fatalf("Load(en) error = %v", err)
	}

	if catalog.Shell.BrandName != "HubRelay" {
		t.Fatalf("expected HubRelay brand, got %q", catalog.Shell.BrandName)
	}
	if len(catalog.Shell.Nav) == 0 {
		t.Fatal("expected shell navigation items")
	}
	if catalog.Ask.Errors.PromptRequired == "" {
		t.Fatal("expected ask prompt required copy")
	}
	if catalog.Common.EmptyValue == "" {
		t.Fatal("expected common empty value label")
	}
	if catalog.Common.LocaleCode != "EN" {
		t.Fatalf("expected EN locale code, got %q", catalog.Common.LocaleCode)
	}
	if catalog.Common.SwitchToDarkTheme == "" {
		t.Fatal("expected english theme toggle copy")
	}
}

func TestLoadRussianCatalog(t *testing.T) {
	catalog, err := Load("ru")
	if err != nil {
		t.Fatalf("Load(ru) error = %v", err)
	}

	if catalog.Health.Header.Title != "Состояние HubRelay" {
		t.Fatalf("unexpected russian health title %q", catalog.Health.Header.Title)
	}
	if catalog.Common.LocaleCode != "RU" {
		t.Fatalf("expected RU locale code, got %q", catalog.Common.LocaleCode)
	}
	if catalog.Common.SwitchToLightTheme == "" {
		t.Fatal("expected russian theme toggle copy")
	}
}

func TestLoadSpanishCatalog(t *testing.T) {
	catalog, err := Load("es")
	if err != nil {
		t.Fatalf("Load(es) error = %v", err)
	}

	if catalog.Health.Header.Title != "Estado de HubRelay" {
		t.Fatalf("unexpected spanish health title %q", catalog.Health.Header.Title)
	}
	if catalog.Common.LocaleCode != "ES" {
		t.Fatalf("expected ES locale code, got %q", catalog.Common.LocaleCode)
	}
	if catalog.Common.ToggleTheme == "" {
		t.Fatal("expected spanish generic theme toggle label")
	}
}
