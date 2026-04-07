package budget

import (
	"testing"
	v1 "chaosOperator/api/v1"
)

func TestCalculate(t *testing.T) {
	calc := NewCalculator()
	spec := v1.BudgetSpec{Max: 0.05}

	tests := []struct {
		name     string
		metric   float64
		expected float64
	}{
		{"Normal", 0.02, 0.02},
		{"Over Max", 0.07, 0.07},
		{"Zero", 0.0, 0.0},
		{"Negative", -0.1, 0.0},
		{"Over One", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := calc.Calculate(spec, tt.metric)
			if res != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, res)
			}
		})
	}
}
