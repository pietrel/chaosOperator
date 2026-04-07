package budget

import (
	v1 "chaosOperator/api/v1"
)

type Calculator interface {
	Calculate(spec v1.BudgetSpec, metric float64) float64
}

type defaultCalculator struct{}

func NewCalculator() Calculator {
	return &defaultCalculator{}
}

func (c *defaultCalculator) Calculate(spec v1.BudgetSpec, metric float64) float64 {
	// consumed = metrics value as per spec
	consumed := metric
	if consumed < 0 {
		consumed = 0
	}
	if consumed > spec.Max {
		// We can consume more than max, but it will result in allowed=false
		// The spec says "Clamp values to [0,1]", but if max is say 0.05,
		// and metric is 0.1, consumed is 0.1.
		// If the spec meant normalized consumed, it would be different.
		// "remaining = max - consumed"
	}
	// Normalizing to [0,1] as requested by "Return value must be normalized to [0,1]" in metrics integration
	// but budget calculation says "consumed = metrics value".

	if consumed > 1.0 {
		consumed = 1.0
	}

	return consumed
}
