# Chaos Budget Operator

The Chaos Budget Operator is a Kubernetes-native controller that manages chaos experiment permissions based on reliability budgets. It continuously monitors system health via Prometheus and ensures that chaos experiments are only allowed when the system has sufficient reliability budget.

## Features

- **Reliability Budgets:** Define error-rate, latency (p95), or availability budgets.
- **Prometheus Integration:** Real-time budget consumption monitoring.
- **Policy Engine:** Customizable policies (`deny`, `throttle`, `noop`) to override decisions.
- **Decision API:** Fast HTTP endpoint for chaos agents to query permissions.
- **Kubernetes Native:** Managed via `ChaosBudget` Custom Resource Definitions (CRDs).

## Architecture

The operator consists of several core modules:

- **Controller:** Reconciles `ChaosBudget` resources and updates their status.
- **Metrics Client:** Fetches live metrics from Prometheus.
- **Budget Calculator:** Computes budget consumption based on metrics.
- **Policy Evaluator:** Applies additional logic to allow/deny decisions.
- **Decision API:** Provides a `/check?target=<name>` endpoint for external agents.

## Getting Started

### Prerequisites

- Go 1.25+
- Kubernetes cluster
- Prometheus server

### Installation

1. **Apply the CRD:**

```bash
kubectl apply -f config/crd/chaos_v1_chaosbudget.yaml
```

2. **Run the Operator:**

```bash
go run cmd/manager/main.go --prometheus-address http://your-prometheus:9090
```

## Configuration

### ChaosBudget Resource

Example `ChaosBudget` definition:

```yaml
apiVersion: chaos.example.com/v1
kind: ChaosBudget
metadata:
  name: api-error-budget
spec:
  target:
    name: api-service
  budget:
    type: "error-rate"
    max: 0.05
  window: "1h"
  policies:
    - type: "throttle"
```

### Command Line Flags

- `--metrics-bind-address`: The address the metric endpoint binds to (default `:8080`).
- `--health-probe-bind-address`: The address the probe endpoint binds to (default `:8081`).
- `--prometheus-address`: Prometheus server address (default `http://prometheus-k8s.monitoring.svc:9090`).
- `--api-bind-address`: The address the decision API binds to (default `:8082`).
- `--reconcile-interval`: Reconciliation interval (default `60s`).

## API for Chaos Agents

Agents can check if chaos is allowed for a specific target:

```bash
curl "http://operator-address:8082/check?target=api-service"
```

**Response:**

```json
{
  "allowed": true,
  "remaining": 0.02,
  "reason": "within budget"
}
```

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

- `api/v1/`: CRD definitions and Go types.
- `cmd/manager/`: Entrypoint for the operator.
- `internal/controller/`: Reconciliation logic.
- `internal/metrics/`: Prometheus client.
- `internal/budget/`: Budget calculation logic.
- `internal/policy/`: Policy evaluation.
- `internal/api/`: Agent communication API.
