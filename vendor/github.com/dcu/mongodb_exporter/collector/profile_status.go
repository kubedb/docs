package collector

import (
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var (
	profileCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: Namespace,
		Subsystem: "profile",
		Name:      "slow_query_30s_count",
		Help:      "The number of slow queries in this database during last 30 seconds",
	}, []string{"database"})
)

type ProfileStatus struct {
	Name  string `bson:"database"`
	Count int    `bson:"count"`
}

func (profileStatus *ProfileStatus) Export(ch chan<- prometheus.Metric) {
	profileCount.WithLabelValues(profileStatus.Name).Set(float64(profileStatus.Count))
	profileCount.Collect(ch)
	profileCount.Reset()
}

func (profileStatus *ProfileStatus) Describe(ch chan<- *prometheus.Desc) {
	profileCount.Describe(ch)
}

func CollectProfileStatus(session *mgo.Session, db string, ch chan<- prometheus.Metric) {
	ts := time.Now().Add(-time.Duration(time.Second * 30))
	count, err := session.DB(db).C("system.profile").Find(bson.M{"ts": bson.M{"$gt": ts}}).Count()
	if err != nil {
		glog.Error(err)
		return
	}
	profileStatus := ProfileStatus{db, count}
	profileStatus.Export(ch)
}
