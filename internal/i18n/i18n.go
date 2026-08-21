package i18n

import (
	"encoding/json"
	"fmt"
	"sync"
)

var (
	mu       sync.RWMutex
	catalogs = map[string]map[string]string{}
	loaded   bool
)

func EnsureLoaded() error {
	mu.Lock()
	defer mu.Unlock()
	if loaded {
		return nil
	}
	for _, lang := range []string{"zh", "en"} {
		b, err := localeFS.ReadFile("locales/" + lang + ".json")
		if err != nil {
			return err
		}
		m := map[string]string{}
		if err := json.Unmarshal(b, &m); err != nil {
			return fmt.Errorf("parse locale %s: %w", lang, err)
		}
		catalogs[lang] = m
	}
	loaded = true
	return nil
}

func Normalize(lang string) string {
	if lang == "en" {
		return "en"
	}
	return "zh"
}

func T(lang, key string, args ...any) string {
	_ = EnsureLoaded()
	lang = Normalize(lang)
	mu.RLock()
	defer mu.RUnlock()
	msg, ok := catalogs[lang][key]
	if !ok {
		if msg, ok = catalogs["zh"][key]; !ok {
			return key
		}
	}
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

func Dict(lang string) map[string]string {
	_ = EnsureLoaded()
	lang = Normalize(lang)
	mu.RLock()
	defer mu.RUnlock()
	src := catalogs[lang]
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
