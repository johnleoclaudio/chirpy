package middlewares

import (
	"chirpy/metrics"
	"net/http"
)

type Middlewares struct {
	APIMetrics *metrics.API
}

func NewMiddlewares(apiMetrics *metrics.API) *Middlewares {
	return &Middlewares{
		APIMetrics: apiMetrics,
	}
}

func (m *Middlewares) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		m.APIMetrics.IncMetric()
		next.ServeHTTP(w, r)
	})
}
