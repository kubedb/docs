package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	inUse = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "connpoolstats",
		Name:      "in_use",
		Help:      "Corresponds to the total number of client connections to mongo.",
		// TODO: tags
	}, []string{"host"})

	available = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "connpoolstats",
		// TODO
		Name: "available",
		Help: "Corresponds to the total number of client connections to mongo.",
	}, []string{"host"})

	created = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "connpoolstats",
		// TODO
		Name: "created",
		Help: "Corresponds to the total number of client connections to mongo.",
	}, []string{"host"})
)

// ServerStatus keeps the data returned by the serverStatus() method.
type HostConnPoolStats struct {
	InUse     float64 `bson:"inUse"`
	Available float64 `bson:"available"`
	Created   float64 `bson:"created"`
}

// Export exports the server status to be consumed by prometheus.
func (stats *HostConnPoolStats) Export(hostname string, ch chan<- prometheus.Metric) {
	inUse.WithLabelValues(hostname).Set(float64(stats.InUse))
	inUse.Collect(ch)
	inUse.Reset()

	available.WithLabelValues(hostname).Set(float64(stats.Available))
	available.Collect(ch)
	available.Reset()

	created.WithLabelValues(hostname).Set(float64(stats.Created))
	created.Collect(ch)
	created.Reset()
}

// Describe describes the server status for prometheus.
func (stats *HostConnPoolStats) Describe(ch chan<- *prometheus.Desc) {
	inUse.Describe(ch)

	available.Describe(ch)

	created.Describe(ch)
}
