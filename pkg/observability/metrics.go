package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds core prometheus metrics.
type Metrics struct {
	Registry *prometheus.Registry
	Requests *prometheus.CounterVec
	Costs    prometheus.Counter
}

// NewMetrics initializes metrics collectors.
func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "spawn_requests_total",
		Help: "Total processed requests",
	}, []string{"component", "status"})
	costs := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "spawn_cost_usd_total",
		Help: "Total model spend in USD",
	})
	reg.MustRegister(requests, costs)
	return &Metrics{Registry: reg, Requests: requests, Costs: costs}
}

// Handler returns /metrics handler.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}
