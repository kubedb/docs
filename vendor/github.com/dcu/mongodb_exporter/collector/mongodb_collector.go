package collector

import (
	"time"

	"github.com/dcu/mongodb_exporter/shared"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/mgo.v2"
)

var (
	// Namespace is the namespace of the metrics
	Namespace = "mongodb"
)

var (
	upGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Name:      "up",
		Help:      "To show if we can connect to mongodb instance",
	}, []string{})
)

// MongodbCollectorOpts is the options of the mongodb collector.
type MongodbCollectorOpts struct {
	URI                      string
	TLSCertificateFile       string
	TLSPrivateKeyFile        string
	TLSCaFile                string
	TLSHostnameValidation    bool
	CollectReplSet           bool
	CollectOplog             bool
	TailOplog                bool
	CollectTopMetrics        bool
	CollectDatabaseMetrics   bool
	CollectCollectionMetrics bool
	CollectProfileMetrics    bool
	CollectConnPoolStats     bool
	UserName                 string
	AuthMechanism            string
	SocketTimeout            time.Duration
}

func (in MongodbCollectorOpts) toSessionOps() shared.MongoSessionOpts {
	return shared.MongoSessionOpts{
		URI:                   in.URI,
		TLSCertificateFile:    in.TLSCertificateFile,
		TLSPrivateKeyFile:     in.TLSPrivateKeyFile,
		TLSCaFile:             in.TLSCaFile,
		TLSHostnameValidation: in.TLSHostnameValidation,
		UserName:              in.UserName,
		AuthMechanism:         in.AuthMechanism,
		SocketTimeout:         in.SocketTimeout,
	}
}

// MongodbCollector is in charge of collecting mongodb's metrics.
type MongodbCollector struct {
	Opts MongodbCollectorOpts
}

// NewMongodbCollector returns a new instance of a MongodbCollector.
func NewMongodbCollector(opts MongodbCollectorOpts) *MongodbCollector {
	exporter := &MongodbCollector{
		Opts: opts,
	}

	return exporter
}

// Describe describes all mongodb's metrics.
func (exporter *MongodbCollector) Describe(ch chan<- *prometheus.Desc) {
	(&ServerStatus{}).Describe(ch)
	(&ReplSetStatus{}).Describe(ch)
	(&ReplSetConf{}).Describe(ch)
	(&DatabaseStatus{}).Describe(ch)

	if exporter.Opts.CollectTopMetrics {
		(&TopStatus{}).Describe(ch)
	}
}

// Collect collects all mongodb's metrics.
func (exporter *MongodbCollector) Collect(ch chan<- prometheus.Metric) {
	mongoSess := shared.MongoSession(exporter.Opts.toSessionOps())
	if mongoSess != nil {
		upGauge.WithLabelValues().Set(float64(1))
		upGauge.Collect(ch)
		upGauge.Reset()
		defer mongoSess.Close()
		glog.Info("Collecting Server Status")
		exporter.collectServerStatus(mongoSess, ch)
		if exporter.Opts.CollectReplSet {
			glog.Info("Collecting ReplSet Status")
			exporter.collectReplSetStatus(mongoSess, ch)
			exporter.collectReplSetConf(mongoSess, ch)
		}
		if exporter.Opts.CollectOplog {
			glog.Info("Collecting Oplog Status")
			exporter.collectOplogStatus(mongoSess, ch)
		}

		if exporter.Opts.TailOplog {
			glog.Info("Collecting Oplog Tail Stats")
			exporter.collectOplogTailStats(mongoSess, ch)
		}

		if exporter.Opts.CollectTopMetrics {
			glog.Info("Collecting Top Metrics")
			exporter.collectTopStatus(mongoSess, ch)
		}

		if exporter.Opts.CollectDatabaseMetrics {
			glog.Info("Collecting Database Metrics")
			exporter.collectDatabaseStatus(mongoSess, ch)
		}

		if exporter.Opts.CollectCollectionMetrics {
			glog.Info("Collection Collection Metrics")
			exporter.collectCollectionStatus(mongoSess, ch)
		}

		if exporter.Opts.CollectProfileMetrics {
			glog.Info("Collection Profile Metrics")
			exporter.collectProfileStatus(mongoSess, ch)
		}

		if exporter.Opts.CollectConnPoolStats {
			glog.Info("Collecting Connection Pool Stats")
			exporter.collectConnPoolStats(mongoSess, ch)
		}
	} else {
		upGauge.WithLabelValues().Set(float64(0))
		upGauge.Collect(ch)
		upGauge.Reset()
	}
}

func (exporter *MongodbCollector) collectServerStatus(session *mgo.Session, ch chan<- prometheus.Metric) *ServerStatus {
	serverStatus := GetServerStatus(session)
	if serverStatus != nil {
		glog.Info("exporting ServerStatus Metrics")
		serverStatus.Export(ch)
	}
	return serverStatus
}

func (exporter *MongodbCollector) collectReplSetStatus(session *mgo.Session, ch chan<- prometheus.Metric) *ReplSetStatus {
	replSetStatus := GetReplSetStatus(session)

	if replSetStatus != nil {
		glog.Info("exporting ReplSetStatus Metrics")
		replSetStatus.Export(ch)
	}

	return replSetStatus
}

func (exporter *MongodbCollector) collectReplSetConf(session *mgo.Session, ch chan<- prometheus.Metric) *ReplSetConf {
	replSetConf := GetReplSetConf(session)

	if replSetConf != nil {
		glog.Info("exporting ReplSetConf Metrics")
		replSetConf.Export(ch)
	}

	return replSetConf
}

func (exporter *MongodbCollector) collectOplogStatus(session *mgo.Session, ch chan<- prometheus.Metric) *OplogStatus {
	oplogStatus := GetOplogStatus(session)

	if oplogStatus != nil {
		glog.Info("exporting OplogStatus Metrics")
		oplogStatus.Export(ch)
	}

	return oplogStatus
}

func (exporter *MongodbCollector) collectOplogTailStats(session *mgo.Session, ch chan<- prometheus.Metric) *OplogTailStats {
	oplogTailStats := GetOplogTailStats(session)

	if oplogTailStats != nil {
		glog.Info("exporting oplogTailStats Metrics")
		oplogTailStats.Export(ch)
	}

	return oplogTailStats
}

func (exporter *MongodbCollector) collectTopStatus(session *mgo.Session, ch chan<- prometheus.Metric) *TopStatus {
	topStatus := GetTopStatus(session)
	if topStatus != nil {
		glog.Info("exporting Top Metrics")
		topStatus.Export(ch)
	}
	return topStatus
}

func (exporter *MongodbCollector) collectDatabaseStatus(session *mgo.Session, ch chan<- prometheus.Metric) {
	all, err := session.DatabaseNames()
	if err != nil {
		glog.Error("Failed to get database names")
		return
	}
	for _, db := range all {
		if db == "admin" || db == "test" {
			continue
		}
		dbStatus := GetDatabaseStatus(session, db)
		if dbStatus != nil {
			glog.Infof("exporting Database Metrics for db=%q", dbStatus.Name)
			dbStatus.Export(ch)
		}
	}
}

func (exporter *MongodbCollector) collectCollectionStatus(session *mgo.Session, ch chan<- prometheus.Metric) {
	database_names, err := session.DatabaseNames()
	if err != nil {
		glog.Error("failed to get database names")
		return
	}
	for _, db := range database_names {
		if db == "admin" || db == "test" {
			continue
		}
		CollectCollectionStatus(session, db, ch)
	}
}

func (exporter *MongodbCollector) collectProfileStatus(session *mgo.Session, ch chan<- prometheus.Metric) {
	all, err := session.DatabaseNames()
	if err != nil {
		glog.Error("failed to get database names: %s", err)
		return
	}
	for _, db := range all {
		if db == "admin" || db == "test" {
			continue
		}
		CollectProfileStatus(session, db, ch)
	}
}

func (exporter *MongodbCollector) collectConnPoolStats(session *mgo.Session, ch chan<- prometheus.Metric) {
	connPoolStats := GetConnPoolStats(session)

	if connPoolStats != nil {
		glog.Info("exporting ConnPoolStats Metrics")
		connPoolStats.Export(ch)
	}
}
