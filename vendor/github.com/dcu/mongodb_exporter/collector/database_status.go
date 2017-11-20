package collector

import (
	"strings"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	indexSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "db",
		Name:      "index_size_bytes",
		Help:      "The total size in bytes of all indexes created on this database",
	}, []string{"db", "shard"})
	dataSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "db",
		Name:      "data_size_bytes",
		Help:      "The total size in bytes of the uncompressed data held in this database",
	}, []string{"db", "shard"})
	collectionsTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "db",
		Name:      "collections_total",
		Help:      "Contains a count of the number of collections in that database",
	}, []string{"db", "shard"})
	indexesTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "db",
		Name:      "indexes_total",
		Help:      "Contains a count of the total number of indexes across all collections in the database",
	}, []string{"db", "shard"})
	objectsTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "db",
		Name:      "objects_total",
		Help:      "Contains a count of the number of objects (i.e. documents) in the database across all collections",
	}, []string{"db", "shard"})
)

// DatabaseStatus represents stats about a database
type DatabaseStatus struct {
	Name        string                `bson:"db,omitempty"`
	IndexSize   int                   `bson:"indexSize,omitempty"`
	DataSize    int                   `bson:"dataSize,omitempty"`
	Collections int                   `bson:"collections,omitempty"`
	Objects     int                   `bson:"objects,omitempty"`
	Indexes     int                   `bson:"indexes,omitempty"`
	Shards      map[string]*RawStatus `bson:"raw,omitempty"`
}

// RawStatus represents stats about a database shard
type RawStatus struct {
	Name        string `bson:"db,omitempty"`
	IndexSize   int    `bson:"indexSize,omitempty"`
	DataSize    int    `bson:"dataSize,omitempty"`
	Collections int    `bson:"collections,omitempty"`
	Objects     int    `bson:"objects,omitempty"`
	Indexes     int    `bson:"indexes,omitempty"`
}

// Export exports database stats to prometheus
func (dbStatus *DatabaseStatus) Export(ch chan<- prometheus.Metric) {
	if len(dbStatus.Shards) > 0 {
		for shard, stats := range dbStatus.Shards {
			shard = strings.Split(shard, "/")[0]
			indexSize.WithLabelValues(stats.Name, shard).Set(float64(stats.IndexSize))
			dataSize.WithLabelValues(stats.Name, shard).Set(float64(stats.DataSize))
			collectionsTotal.WithLabelValues(stats.Name, shard).Set(float64(stats.Collections))
			indexesTotal.WithLabelValues(stats.Name, shard).Set(float64(stats.Indexes))
			objectsTotal.WithLabelValues(stats.Name, shard).Set(float64(stats.Objects))
		}
	} else {
		indexSize.WithLabelValues(dbStatus.Name, "").Set(float64(dbStatus.IndexSize))
		dataSize.WithLabelValues(dbStatus.Name, "").Set(float64(dbStatus.DataSize))
		collectionsTotal.WithLabelValues(dbStatus.Name, "").Set(float64(dbStatus.Collections))
		indexesTotal.WithLabelValues(dbStatus.Name, "").Set(float64(dbStatus.Indexes))
		objectsTotal.WithLabelValues(dbStatus.Name, "").Set(float64(dbStatus.Objects))
	}

	indexSize.Collect(ch)
	dataSize.Collect(ch)
	collectionsTotal.Collect(ch)
	indexesTotal.Collect(ch)
	objectsTotal.Collect(ch)

	indexSize.Reset()
	dataSize.Reset()
	collectionsTotal.Reset()
	indexesTotal.Reset()
	objectsTotal.Reset()
}

// Describe describes database stats for prometheus
func (dbStatus *DatabaseStatus) Describe(ch chan<- *prometheus.Desc) {
	indexSize.Describe(ch)
	dataSize.Describe(ch)
	collectionsTotal.Describe(ch)
	indexesTotal.Describe(ch)
	objectsTotal.Describe(ch)
}

// GetDatabaseStatus returns stats for a given database
func GetDatabaseStatus(session *mgo.Session, db string) *DatabaseStatus {
	var dbStatus DatabaseStatus
	err := session.DB(db).Run(bson.D{{"dbStats", 1}, {"scale", 1}}, &dbStatus)
	if err != nil {
		glog.Error(err)
		return nil
	}

	return &dbStatus
}
