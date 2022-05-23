package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	Observations      prometheus.Counter
	ObservationsBytes *prometheus.CounterVec

	Duration *prometheus.GaugeVec
)

type NetStat struct {
}
