package collector

import (
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// TopStatus represents top metrics
type TopStatus struct {
	TopStats TopStatsMap `bson:"totals,omitempty"`
}

// GetTopStats fetches top stats
func GetTopStats(session *mgo.Session) (*TopStatus, error) {
	results := &TopStatus{}
	err := session.DB("admin").Run(bson.D{{"top", 1}}, &results)
	return results, err
}

// Export exports metrics to Prometheus
func (status *TopStatus) Export(ch chan<- prometheus.Metric) {
	status.TopStats.Export(ch)
}

// Describe describes metrics collected
func (status *TopStatus) Describe(ch chan<- *prometheus.Desc) {
	status.TopStats.Describe(ch)
}

// GetTopStatus fetches top stats
func GetTopStatus(session *mgo.Session) *TopStatus {
	topStatus, err := GetTopStats(session)
	if err != nil {
		glog.Error("Failed to get top status.")
		return nil
	}

	return topStatus
}
