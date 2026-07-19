// Package i18n loads translation strings from JSON files (docs/TRD.md §11)
// and looks them up by key. Phase 2 only ships English content; Hindi and
// Arabic files/the language switcher land in Phase 3/9
// (docs/IMPLEMENTATION_PLAN.md) — the lookup already falls back to English
// so templates never need a hardcoded string in the meantime.
package i18n

import (
	"embed"
	"encoding/json"
)

//go:embed *.json
var filesFS embed.FS

var catalog = map[string]map[string]string{}

func init() {
	for _, lang := range []string{"en"} {
		data, err := filesFS.ReadFile(lang + ".json")
		if err != nil {
			continue
		}
		var strings map[string]string
		if err := json.Unmarshal(data, &strings); err != nil {
			continue
		}
		catalog[lang] = strings
	}
}

// T looks up key in lang's catalog, falling back to English, then to the
// key itself if truly missing (visibly wrong rather than a blank string).
func T(lang, key string) string {
	if strings, ok := catalog[lang]; ok {
		if v, ok := strings[key]; ok {
			return v
		}
	}
	if v, ok := catalog["en"][key]; ok {
		return v
	}
	return key
}
