package fixtures

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
)

// FS stores embedded UI copy and mock payloads.
//
//go:embed */*.json */mocks/*.json
var FS embed.FS

func Decode(name string, target any) error {
	body, err := FS.ReadFile(path.Clean(name))
	if err != nil {
		return fmt.Errorf("read fixture %q: %w", name, err)
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("decode fixture %q: %w", name, err)
	}
	return nil
}

func Locales() ([]string, error) {
	entries, err := fs.ReadDir(FS, ".")
	if err != nil {
		return nil, fmt.Errorf("read fixture locales: %w", err)
	}

	locales := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			locales = append(locales, entry.Name())
		}
	}
	sort.Strings(locales)
	return locales, nil
}
