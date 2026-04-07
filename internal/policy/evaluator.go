package policy

type Decision struct {
	Allowed bool
	Reason  string
}

type Evaluator interface {
	Evaluate(policies []string, remaining float64) Decision
}

type defaultEvaluator struct{}

func NewEvaluator() Evaluator {
	return &defaultEvaluator{}
}

func (e *defaultEvaluator) Evaluate(policies []string, remaining float64) Decision {
	decision := Decision{
		Allowed: remaining > 0,
		Reason:  "within budget",
	}
	if !decision.Allowed {
		decision.Reason = "budget exceeded"
	}

	for _, p := range policies {
		switch p {
		case "deny":
			decision.Allowed = false
			decision.Reason = "force denied by policy"
		case "throttle":
			// allow but mark degraded - in this implementation we still allow if budget ok
			// but we could change the reason.
			if decision.Allowed {
				decision.Reason = "allowed (degraded/throttled)"
			}
		case "noop":
			// no effect
		}
	}

	return decision
}
