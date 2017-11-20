package collector

import (
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// server connections -- all of these!
var (
	syncClientConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "connpoolstats",
		Name:      "connection_sync",
		Help:      "Corresponds to the total number of client connections to mongo.",
	})

	numAScopedConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "connpoolstats",
		Name:      "connections_scoped_sync",
		Help:      "Corresponds to the number of active and stored outgoing scoped synchronous connections from the current instance to other members of the sharded cluster or replica set.",
	})

	totalInUse = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "connpoolstats",
		Name:      "connections_in_use",
		Help:      "Corresponds to the total number of client connections to mongo currently in use.",
	})

	totalAvailable = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "connpoolstats",
		Name:      "connections_available",
		Help:      "Corresponds to the total number of client connections to mongo that are currently available.",
	})

	totalCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: "connpoolstats",
		Name:      "connections_created_total",
		Help:      "Corresponds to the total number of client connections to mongo created since instance start",
	})
)

// ServerStatus keeps the data returned by the serverStatus() method.
type ConnPoolStats struct {
	SyncClientConnections float64 `bson:"numClientConnections"`
	ASScopedConnections   float64 `bson:"numAScopedConnections"`
	TotalInUse            float64 `bson:"totalInUse"`
	TotalAvailable        float64 `bson:"totalAvailable"`
	TotalCreated          float64 `bson:"totalCreated"`

	Hosts map[string]*HostConnPoolStats `bson:hosts"`
	// TODO:? not sure if *this* level of granularity is helpful
	//ReplicaSets map[string]ConnPoolReplicaSetStats `bson:"replicaSets"`
}

// Export exports the server status to be consumed by prometheus.
func (stats *ConnPoolStats) Export(ch chan<- prometheus.Metric) {
	syncClientConnections.Set(stats.SyncClientConnections)
	syncClientConnections.Collect(ch)

	numAScopedConnections.Set(stats.ASScopedConnections)
	numAScopedConnections.Collect(ch)

	totalInUse.Set(stats.TotalInUse)
	totalInUse.Collect(ch)

	totalAvailable.Set(stats.TotalAvailable)
	totalAvailable.Collect(ch)

	totalCreated.Set(stats.TotalCreated)
	totalCreated.Collect(ch)

	for hostname, hostStat := range stats.Hosts {
		hostStat.Export(hostname, ch)
	}
}

// Describe describes the server status for prometheus.
func (stats *ConnPoolStats) Describe(ch chan<- *prometheus.Desc) {
	syncClientConnections.Describe(ch)

	numAScopedConnections.Describe(ch)

	totalInUse.Describe(ch)

	totalAvailable.Describe(ch)

	totalCreated.Describe(ch)

	for _, hostStat := range stats.Hosts {
		hostStat.Describe(ch)
	}
}

// GetServerStatus returns the server status info.
func GetConnPoolStats(session *mgo.Session) *ConnPoolStats {
	result := &ConnPoolStats{}
	err := session.DB("admin").Run(bson.D{{"connPoolStats", 1}, {"recordStats", 0}}, result)
	if err != nil {
		glog.Error("Failed to get server status.")
		return nil
	}
	return result
}
