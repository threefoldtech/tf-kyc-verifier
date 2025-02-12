package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	instance  *Metrics
	once      sync.Once
	registery *prometheus.Registry
)

type Metrics struct {
	HTTPRequestsReceived     *prometheus.CounterVec
	HTTPRequestLatency       *prometheus.HistogramVec
	MongoDBOperationsLatency *prometheus.HistogramVec
	IdenfyResponseTime       *prometheus.HistogramVec
	SubstrateResponseTime    *prometheus.HistogramVec
	ServiceUpTime            prometheus.GaugeFunc
	InternalServerErrorRate  *prometheus.CounterVec
	MongoDBOperationsError   *prometheus.CounterVec
}

func GetInstance() *Metrics {
	startTime := time.Now().Unix()

	once.Do(
		func() {
			instance = &Metrics{
				HTTPRequestsReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
					Name: "http_request_count",
					Help: "Number of HTTP requests",
				}, []string{"method", "path"}),

				HTTPRequestLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
					Name: "http_request_latency",
					Help: "Latency of HTTP requests",
				}, []string{"method", "path"}),

				MongoDBOperationsLatency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
					Name: "mongodb_operations_latency",
					Help: "Latency of MongoDB operations",
				}, []string{"operation", "repo"}),

				IdenfyResponseTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
					Name: "idenfy_response_time",
					Help: "Latency of Idenfy API responses",
				}, []string{"operation"}),

				SubstrateResponseTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
					Name: "substrate_response_time",
					Help: "Latency of Substrate API responses",
				}, []string{"operation"}),

				ServiceUpTime: prometheus.NewGaugeFunc(prometheus.GaugeOpts{
					Name: "service_uptime",
					Help: "Service uptime",
				},
					func() float64 { return float64(time.Now().Unix() - startTime) },
				),

				InternalServerErrorRate: prometheus.NewCounterVec(prometheus.CounterOpts{
					Name: "internal_server_error_rate",
					Help: "Internal server error rate",
				}, []string{"method", "path"}),

				MongoDBOperationsError: prometheus.NewCounterVec(
					prometheus.CounterOpts{
						Name: "mongo_db_operations_error",
						Help: "Operations that gives error and does not succed",
					}, []string{"operation", "repo"}),
			}
		})
	return instance
}

func (m *Metrics) Register() error {
	registery = prometheus.NewRegistry()

	for _, metric := range []prometheus.Collector{
		m.HTTPRequestsReceived,
		m.HTTPRequestLatency,
		m.MongoDBOperationsLatency,
		m.IdenfyResponseTime,
		m.SubstrateResponseTime,
		m.ServiceUpTime,
		m.InternalServerErrorRate,
	} {
		if err := registery.Register(metric); err != nil {
			return fmt.Errorf("failed to register metric %v: %w", metric, err)
		}
	}

	return nil
}

func (m *Metrics) GetMetrics() http.Handler {
	return promhttp.HandlerFor(registery, promhttp.HandlerOpts{
		Registry:          registery,
		EnableOpenMetrics: true,
	})
}
