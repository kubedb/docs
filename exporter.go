package main

import (
	"fmt"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/appscode/go/runtime"
	"github.com/appscode/pat"
	tapi "github.com/k8sdb/apimachinery/api"
	ese "github.com/k8sdb/elasticsearch_exporter/exporter"
	pge "github.com/k8sdb/postgres_exporter/exporter"
	"github.com/orcaman/concurrent-map"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"gopkg.in/ini.v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
)

const (
	ParamNamespace = ":namespace"
	ParamType      = ":type"
	ParamName      = ":name"
	ParamPodIP     = ":ip"
)

var (
	registerers = cmap.New() // URL.path => *prometheus.Registry
)

func DeleteRegistry(w http.ResponseWriter, r *http.Request) {
	defer runtime.HandleCrash()

	registerers.Remove(r.URL.Path)
	w.WriteHeader(http.StatusOK)
}

func ExportMetrics(w http.ResponseWriter, r *http.Request) {
	defer runtime.HandleCrash()

	params, found := pat.FromContext(r.Context())
	if !found {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}
	namespace := params.Get(ParamNamespace)
	if namespace == "" {
		http.Error(w, "Missing parameter "+ParamNamespace, http.StatusBadRequest)
		return
	}
	dbType := params.Get(ParamType)
	if dbType == "" {
		http.Error(w, "Missing parameter "+ParamType, http.StatusBadRequest)
		return
	}
	dbName := params.Get(ParamName)
	if dbName == "" {
		http.Error(w, "Missing parameter "+ParamName, http.StatusBadRequest)
		return
	}
	podIP := params.Get(ParamPodIP)
	if podIP == "" {
		http.Error(w, "Missing parameter "+ParamPodIP, http.StatusBadRequest)
		return
	}

	switch dbType {
	case tapi.ResourceTypePostgres:
		var reg *prometheus.Registry
		if val, ok := registerers.Get(r.URL.Path); ok {
			reg = val.(*prometheus.Registry)
		} else {
			reg = prometheus.NewRegistry()
			if absent := registerers.SetIfAbsent(r.URL.Path, reg); !absent {
				r2, _ := registerers.Get(r.URL.Path)
				reg = r2.(*prometheus.Registry)
			} else {
				log.Infof("Configuring exporter for PostgreSQL %s in namespace %s", dbName, namespace)
				db, err := dbClient.Postgreses(namespace).Get(dbName)
				if kerr.IsNotFound(err) {
					http.NotFound(w, r)
					return
				} else if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				conn, err := getPostgresURL(db, podIP)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				reg.MustRegister(pge.NewExporter(conn, ""))
			}
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		return
	case tapi.ResourceTypeElastic:
		var reg *prometheus.Registry
		if val, ok := registerers.Get(r.URL.Path); ok {
			reg = val.(*prometheus.Registry)
		} else {
			reg = prometheus.NewRegistry()
			if absent := registerers.SetIfAbsent(r.URL.Path, reg); !absent {
				r2, _ := registerers.Get(r.URL.Path)
				reg = r2.(*prometheus.Registry)
			} else {
				log.Infof("Configuring exporter for Elasticsearch %s in namespace %s", dbName, namespace)
				_, err := dbClient.Elastics(namespace).Get(dbName)
				if kerr.IsNotFound(err) {
					http.NotFound(w, r)
					return
				} else if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				esURI := fmt.Sprintf("http://%s:9200", podIP)
				nodesStatsURI := esURI + "/_nodes/_local/stats"
				clusterHealthURI := esURI + "/_cluster/health"
				esTimeout := 5 * time.Second
				esAllNodes := false
				exporter := ese.NewExporter(nodesStatsURI, clusterHealthURI, esTimeout, esAllNodes, nil)
				reg.MustRegister(exporter)
			}
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

func getPostgresURL(db *tapi.Postgres, podIP string) (string, error) {
	secret, err := kubeClient.CoreV1().Secrets(db.Namespace).Get(db.Spec.DatabaseSecret.SecretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	cfg, err := ini.Load(secret.Data[".admin"])
	if err != nil {
		return "", err
	}
	section, err := cfg.GetSection("")
	if err != nil {
		return "", err
	}
	user := "postgres"
	if k, err := section.GetKey("POSTGRES_USER"); err == nil {
		user = k.Value()
	}
	var password string
	if k, err := section.GetKey("POSTGRES_PASSWORD"); err == nil {
		password = k.Value()
	}
	conn := fmt.Sprintf("postgres://%s:%s@%s:5432", user, password, podIP)
	return conn, nil
}
