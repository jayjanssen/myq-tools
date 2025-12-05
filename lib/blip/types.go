package blip

import (
	"github.com/cashapp/blip"
)

// Type aliases for blip types used elsewhere in the codebase.
// This allows other packages to use these types without directly
// importing github.com/cashapp/blip.

type (
	Metrics     = blip.Metrics
	MetricValue = blip.MetricValue
)
