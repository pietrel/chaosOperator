package policy

import "testing"

func TestEvaluate(t *testing.T) {
	ev := NewEvaluator()

	tests := []struct {
		name      string
		policies  []string
		remaining float64
		allowed   bool
	}{
		{"Budget OK", []string{}, 0.01, true},
		{"Budget Exceeded", []string{}, -0.01, false},
		{"Deny Policy", []string{"deny"}, 0.01, false},
		{"Throttle Policy OK", []string{"throttle"}, 0.01, true},
		{"Throttle Policy Exceeded", []string{"throttle"}, -0.01, false},
		{"Mixed Policy Deny wins", []string{"throttle", "deny"}, 0.01, false},
		{"Noop Policy", []string{"noop"}, 0.01, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := ev.Evaluate(tt.policies, tt.remaining)
			if res.Allowed != tt.allowed {
				t.Errorf("expected allowed %v, got %v", tt.allowed, res.Allowed)
			}
		})
	}
}
