(function () {
  var root = document.documentElement;
  var themeStorageKey = "ui8kit-theme";

  function ready(fn) {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", fn);
      return;
    }
    fn();
  }

  function normalizeLocale(value, allowedLocales) {
    var locale = (value || "").toLowerCase();
    return allowedLocales.indexOf(locale) !== -1 ? locale : "";
  }

  function parseLocales(raw) {
    return (raw || "")
      .split(",")
      .map(function (item) {
        return item.trim().toLowerCase();
      })
      .filter(function (item, index, list) {
        return item && list.indexOf(item) === index;
      });
  }

  function browserLocales() {
    var values = [];

    if (Array.isArray(navigator.languages)) {
      values = values.concat(navigator.languages);
    }

    if (typeof navigator.language === "string") {
      values.push(navigator.language);
    }

    return values
      .map(function (item) {
        return String(item || "")
          .trim()
          .toLowerCase()
          .replace(/_/g, "-");
      })
      .filter(function (item, index, list) {
        return item && list.indexOf(item) === index;
      });
  }

  function detectPreferredLocale(allowedLocales, defaultLocale) {
    var locales = browserLocales();

    for (var i = 0; i < locales.length; i += 1) {
      var locale = locales[i];
      if (allowedLocales.indexOf(locale) !== -1) {
        return locale;
      }

      var baseLocale = locale.split("-")[0];
      if (allowedLocales.indexOf(baseLocale) !== -1) {
        return baseLocale;
      }
    }

    return defaultLocale;
  }

  function readStoredTheme() {
    try {
      return localStorage.getItem(themeStorageKey);
    } catch (_) {
      return null;
    }
  }

  function writeStoredTheme(value) {
    try {
      localStorage.setItem(themeStorageKey, value);
    } catch (_) {}
  }

  function resolvePreferredTheme() {
    var storedTheme = readStoredTheme();
    if (storedTheme === "dark" || storedTheme === "light") {
      return storedTheme;
    }

    var prefersDark =
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches;
    return prefersDark ? "dark" : "light";
  }

  function applyTheme(theme) {
    root.classList.toggle("dark", theme === "dark");
  }

  function applyThemeButtonState() {
    var icon = document.getElementById("theme-toggle-icon");
    var button = document.getElementById("ui8kit-theme-toggle");
    var dark = root.classList.contains("dark");
    var switchToDark =
      button && button.dataset.switchToDarkLabel
        ? button.dataset.switchToDarkLabel
        : "Switch to dark theme";
    var switchToLight =
      button && button.dataset.switchToLightLabel
        ? button.dataset.switchToLightLabel
        : "Switch to light theme";

    if (icon) {
      icon.className = dark
        ? "latty latty-sun h-4 w-4"
        : "latty latty-moon h-4 w-4";
    }

    if (button) {
      button.setAttribute("aria-pressed", dark ? "true" : "false");
      button.setAttribute("title", dark ? switchToLight : switchToDark);
      button.setAttribute("aria-label", dark ? switchToLight : switchToDark);
    }
  }

  applyTheme(resolvePreferredTheme());

  ready(function () {
    var themeButton = document.getElementById("ui8kit-theme-toggle");
    var toggle = document.getElementById("app-language-toggle");
    var storageKey = "gitcourse-language";

    if (themeButton) {
      themeButton.addEventListener("click", function () {
        var nextTheme = root.classList.contains("dark") ? "light" : "dark";
        applyTheme(nextTheme);
        writeStoredTheme(nextTheme);
        applyThemeButtonState();
      });
    }

    applyThemeButtonState();

    if (!toggle) {
      return;
    }

    var availableLocales = parseLocales(toggle.dataset.availableLocales);
    var defaultLocale =
      normalizeLocale(toggle.dataset.defaultLocale, availableLocales) ||
      availableLocales[0] ||
      "";
    var currentLocale =
      normalizeLocale(toggle.dataset.currentLocale, availableLocales) ||
      defaultLocale;
    var nextLocale =
      normalizeLocale(toggle.dataset.nextLocale, availableLocales) ||
      defaultLocale;

    function readStoredLocale() {
      try {
        return normalizeLocale(localStorage.getItem(storageKey), availableLocales);
      } catch (_) {
        return "";
      }
    }

    function writeStoredLocale(locale) {
      try {
        localStorage.setItem(storageKey, locale);
      } catch (_) {}
    }

    function localeURL(locale) {
      var next = new URL(window.location.href);
      if (locale === defaultLocale) {
        next.searchParams.delete("lang");
      } else {
        next.searchParams.set("lang", locale);
      }
      return next.toString();
    }

    var currentURL = new URL(window.location.href);
    var hasExplicitLocaleParam = currentURL.searchParams.has("lang");
    var storedLocale = readStoredLocale();
    if (hasExplicitLocaleParam) {
      writeStoredLocale(currentLocale);
    } else if (!storedLocale) {
      var detectedLocale = detectPreferredLocale(availableLocales, defaultLocale);
      writeStoredLocale(detectedLocale);
      if (detectedLocale !== currentLocale) {
        window.location.replace(localeURL(detectedLocale));
        return;
      }
      currentLocale = detectedLocale;
    } else if (storedLocale !== currentLocale) {
      var targetURL = localeURL(storedLocale);
      if (targetURL !== window.location.href) {
        window.location.replace(targetURL);
        return;
      }
      currentLocale = storedLocale;
    }

    if (currentLocale === defaultLocale) {
      var canonicalURL = localeURL(defaultLocale);
      if (canonicalURL !== window.location.href) {
        window.location.replace(canonicalURL);
        return;
      }
    }

    toggle.addEventListener("click", function () {
      var targetLocale = nextLocale;
      if (!targetLocale || targetLocale === currentLocale) {
        return;
      }
      writeStoredLocale(targetLocale);
      window.location.assign(localeURL(targetLocale));
    });
  });
})();
