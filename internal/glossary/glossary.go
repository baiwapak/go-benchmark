package glossary

// Section groups related terms on the glossary page.
type Section struct {
	TitleKey string
	TermIDs  []string
}

// Sections defines glossary layout (i18n keys: glossary.term.<id>.title / .body).
var Sections = []Section{
	{
		TitleKey: "glossary.section.metrics",
		TermIDs:  []string{"rps", "qps", "tps", "throughput", "latency", "ttfb", "success_rate", "error_rate", "p50", "p90", "p95", "p99"},
	},
	{
		TitleKey: "glossary.section.params",
		TermIDs:  []string{"concurrency", "duration", "timeout", "ramp"},
	},
	{
		TitleKey: "glossary.section.http",
		TermIDs:  []string{"http", "status_code", "header", "sse"},
	},
	{
		TitleKey: "glossary.section.other",
		TermIDs:  []string{"sla", "ssrf", "api", "cdn"},
	},
}
