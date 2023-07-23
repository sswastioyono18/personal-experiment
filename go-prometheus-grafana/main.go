package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func recordMetrics() {
	go func() {
		for {
			responseStatus.WithLabelValues("200").Inc()
			time.Sleep(2 * time.Second)
			responseStatus.WithLabelValues("400").Inc()
		}
	}()
}

var responseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_response_status",
		Help: "Status of HTTP response",
	},
	[]string{"status"},
)

func main() {
	prometheus.MustRegister(responseStatus)

	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
