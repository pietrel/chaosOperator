package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Client interface {
	FetchMetric(ctx context.Context, budgetType string, window string, namespace string, labels map[string]string) (float64, error)
}

type promClient struct {
	v1api v1.API
}

func NewClient(address string) (Client, error) {
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return nil, err
	}

	return &promClient{
		v1api: v1.NewAPI(client),
	}, nil
}

func (p *promClient) FetchMetric(ctx context.Context, budgetType string, window string, namespace string, labels map[string]string) (float64, error) {
	query := ""
	labelStr := ""
	for k, v := range labels {
		labelStr += fmt.Sprintf(",%s=\"%s\"", k, v)
	}
	if labelStr != "" {
		labelStr = labelStr[1:] // remove leading comma
	}

	switch budgetType {
	case "error-rate":
		// query = errors / total_requests
		// Example: sum(rate(http_requests_total{status=~"5..", namespace="default", app="foo"}[1h])) / sum(rate(http_requests_total{namespace="default", app="foo"}[1h]))
		query = fmt.Sprintf("sum(rate(http_requests_total{status=~\"5..\",%s}[%s])) / sum(rate(http_requests_total{%s}[%s]))", labelStr, window, labelStr, window)
	case "latency":
		// query = histogram_quantile(0.95, ...)
		// Example: histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{namespace="default", app="foo"}[1h])) by (le))
		query = fmt.Sprintf("histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{%s}[%s])) by (le))", labelStr, window)
	case "availability":
		// query = successful / total
		// Example: sum(rate(http_requests_total{status!~"5..", namespace="default", app="foo"}[1h])) / sum(rate(http_requests_total{namespace="default", app="foo"}[1h]))
		query = fmt.Sprintf("sum(rate(http_requests_total{status!~\"5..\",%s}[%s])) / sum(rate(http_requests_total{%s}[%s]))", labelStr, window, labelStr, window)
	default:
		return 0, fmt.Errorf("unsupported budget type: %s", budgetType)
	}

	result, _, err := p.v1api.Query(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}

	switch v := result.(type) {
	case model.Vector:
		if v.Len() == 0 {
			return 0, nil
		}
		return float64(v[0].Value), nil
	default:
		return 0, fmt.Errorf("unexpected metric result type: %T", result)
	}
}
