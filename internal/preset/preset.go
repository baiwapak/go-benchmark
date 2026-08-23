package preset

// Preset describes a reference server profile and suggested load-test parameters.
type Preset struct {
	ID          string
	NameKey     string
	SpecKey     string
	RationaleKey string
	Concurrency int
	DurationSec int
	TimeoutSec  int
}

// All returns built-in presets ordered from smallest to largest capacity.
var All = []Preset{
	{
		ID:           "starter",
		NameKey:      "home.preset.starter.name",
		SpecKey:      "home.preset.starter.spec",
		RationaleKey: "home.preset.starter.rationale",
		Concurrency:  5,
		DurationSec:  30,
		TimeoutSec:   15,
	},
	{
		ID:           "basic",
		NameKey:      "home.preset.basic.name",
		SpecKey:      "home.preset.basic.spec",
		RationaleKey: "home.preset.basic.rationale",
		Concurrency:  10,
		DurationSec:  45,
		TimeoutSec:   12,
	},
	{
		ID:           "standard",
		NameKey:      "home.preset.standard.name",
		SpecKey:      "home.preset.standard.spec",
		RationaleKey: "home.preset.standard.rationale",
		Concurrency:  20,
		DurationSec:  60,
		TimeoutSec:   10,
	},
	{
		ID:           "growth",
		NameKey:      "home.preset.growth.name",
		SpecKey:      "home.preset.growth.spec",
		RationaleKey: "home.preset.growth.rationale",
		Concurrency:  35,
		DurationSec:  60,
		TimeoutSec:   10,
	},
	{
		ID:           "enterprise",
		NameKey:      "home.preset.enterprise.name",
		SpecKey:      "home.preset.enterprise.spec",
		RationaleKey: "home.preset.enterprise.rationale",
		Concurrency:  60,
		DurationSec:  90,
		TimeoutSec:   10,
	},
	{
		ID:           "performance",
		NameKey:      "home.preset.performance.name",
		SpecKey:      "home.preset.performance.spec",
		RationaleKey: "home.preset.performance.rationale",
		Concurrency:  100,
		DurationSec:  120,
		TimeoutSec:   8,
	},
}
