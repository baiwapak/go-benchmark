package i18n

import "embed"

//go:embed locales/*.json
var localeFS embed.FS
