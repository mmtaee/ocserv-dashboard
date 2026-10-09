// Package i18n holds Telegram HTML message templates loaded from JSON.
// Override or extend with TELEGRAM_I18N_PATH pointing to a JSON file with the same shape
// as default.json (language code -> key -> format string for fmt.Sprintf).
// See docs/telegram-translations.md.
package i18n

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

//go:embed default.json
var defaultEmbedded []byte

var (
	mu    sync.Mutex
	store map[string]map[string]string // lang -> key -> template
	once  sync.Once
)

// Init loads embedded defaults and optional TELEGRAM_I18N_PATH merge. Safe to call many times.
func Init() {
	once.Do(func() {
		mu.Lock()
		store = make(map[string]map[string]string)
		mu.Unlock()
		_ = mergeJSON(defaultEmbedded)
		if p := strings.TrimSpace(os.Getenv("TELEGRAM_I18N_PATH")); p != "" {
			if b, err := os.ReadFile(p); err == nil {
				_ = mergeJSON(b)
			}
		}
	})
}

func mergeJSON(b []byte) error {
	var raw map[string]map[string]string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	for lang, m := range raw {
		lang = strings.ToLower(strings.TrimSpace(lang))
		if store[lang] == nil {
			store[lang] = make(map[string]string)
		}
		for k, v := range m {
			store[lang][k] = v
		}
	}
	return nil
}

// T returns a formatted template for lang and key, falling back to English, then to key name.
func T(lang, key string, args ...any) string {
	Init()
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		lang = "en"
	}
	mu.Lock()
	defer mu.Unlock()
	msg, ok := store[lang][key]
	if !ok && lang != "en" {
		msg, ok = store["en"][key]
	}
	if !ok {
		return key
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}
