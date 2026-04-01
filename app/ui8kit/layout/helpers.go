package layout

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/fastygo/ui8kit/utils"
)

// MobileSheetCheckboxID toggles the mobile nav sheet (checkbox + label pattern, same idea as ui8kit-core Sheet.tsx).
const MobileSheetCheckboxID = "ui8kit-mobile-sheet"

// MobileSheetPanelID is the dialog surface referenced by aria-controls on the menu trigger.
const MobileSheetPanelID = "ui8kit-mobile-sheet-panel"

func rawScript(js string) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<script>"+js+"</script>")
		return err
	})
}

func sidebarLinkClass(active, path string) string {
	if active == path {
		return "bg-accent text-accent-foreground"
	}
	return "text-muted-foreground hover:bg-accent"
}

func sidebarItemClasses(active, path string) string {
	return utils.Cn("flex items-center gap-2 rounded px-4 py-2 text-sm", sidebarLinkClass(active, path))
}

func shellBrand(name string) string {
	if name == "" {
		return "App"
	}
	return name
}

func shellCSS(path string) string {
	if path == "" {
		return "/static/css/app.css"
	}
	return path
}

const themeScript = `(function(){var root=document.documentElement;var storageKey="ui8kit-theme";var readStoredTheme=function(){try{return localStorage.getItem(storageKey)}catch(_){return null}};var writeStoredTheme=function(next){try{localStorage.setItem(storageKey,next)}catch(_){}};
var applyThemeState=function(){var icon=document.getElementById("theme-toggle-icon");var button=document.getElementById("ui8kit-theme-toggle");var dark=root.classList.contains("dark");if(icon){icon.className=dark?"latty latty-sun h-4 w-4":"latty latty-moon h-4 w-4"}if(button){button.setAttribute("aria-pressed",dark?"true":"false");button.setAttribute("title",dark?"Switch to light theme":"Switch to dark theme")}};
var stored=readStoredTheme();var prefersDark=window.matchMedia&&window.matchMedia("(prefers-color-scheme:dark)").matches;var shouldUseDark=stored==="dark"||(!stored&&prefersDark);root.classList.toggle("dark",shouldUseDark);applyThemeState();
window.ui8kitToggleTheme=function(){var nextDark=!root.classList.contains("dark");root.classList.toggle("dark",nextDark);writeStoredTheme(nextDark?"dark":"light");applyThemeState()};
document.addEventListener("DOMContentLoaded",applyThemeState);})();`
