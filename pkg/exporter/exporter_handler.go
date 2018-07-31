package exporter

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/appscode/go/ioutil"
	"github.com/appscode/go/runtime"
	"github.com/appscode/pat"
	mgoe "github.com/dcu/mongodb_exporter/collector"
	"github.com/go-kit/kit/log"
	ese "github.com/justwatchcom/elasticsearch_exporter/collector"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	esCtrl "github.com/kubedb/elasticsearch/pkg/controller"
	mgCtrl "github.com/kubedb/mongodb/pkg/controller"
	msCtrl "github.com/kubedb/mysql/pkg/controller"
	pgCtrl "github.com/kubedb/postgres/pkg/controller"
	rde "github.com/oliver006/redis_exporter/exporter"
	"github.com/orcaman/concurrent-map"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	plog "github.com/prometheus/common/log"
	memEx "github.com/prometheus/memcached_exporter/collector"
	mse "github.com/prometheus/mysqld_exporter/collector"
	pge "github.com/wrouesnel/postgres_exporter/collector"
)

const (
	PathParamNamespace = ":namespace"
	PathParamType      = ":type"
	PathParamName      = ":name"
	QueryParamPodIP    = "pod"
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
	namespace := params.Get(PathParamNamespace)
	if namespace == "" {
		http.Error(w, "Missing parameter "+PathParamNamespace, http.StatusBadRequest)
		return
	}
	dbType := params.Get(PathParamType)
	if dbType == "" {
		http.Error(w, "Missing parameter "+PathParamType, http.StatusBadRequest)
		return
	}
	dbName := params.Get(PathParamName)
	if dbName == "" {
		http.Error(w, "Missing parameter "+PathParamName, http.StatusBadRequest)
		return
	}
	podIP := r.URL.Query().Get(QueryParamPodIP)
	if podIP == "" {
		podIP = "127.0.0.1"
	}

	switch dbType {
	case api.ResourcePluralPostgres:
		var reg *prometheus.Registry
		if val, ok := registerers.Get(r.URL.Path); ok {
			reg = val.(*prometheus.Registry)
		} else {
			reg = prometheus.NewRegistry()
			if absent := registerers.SetIfAbsent(r.URL.Path, reg); !absent {
				r2, _ := registerers.Get(r.URL.Path)
				reg = r2.(*prometheus.Registry)
			} else {
				plog.Infof("Configuring exporter for PostgreSQL %s in namespace %s", dbName, namespace)

				conn, err := getPostgresURL(podIP)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				reg.MustRegister(pge.NewExporter(conn, ""))
			}
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		return
	case api.ResourcePluralElasticsearch:
		logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
		var reg *prometheus.Registry
		if val, ok := registerers.Get(r.URL.Path); ok {
			reg = val.(*prometheus.Registry)
		} else {
			reg = prometheus.NewRegistry()
			if absent := registerers.SetIfAbsent(r.URL.Path, reg); !absent {
				r2, _ := registerers.Get(r.URL.Path)
				reg = r2.(*prometheus.Registry)
			} else {
				plog.Infof("Configuring exporter for Elasticsearch %s in namespace %s", dbName, namespace)

				password, err := getElasticsearchPassword()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				esAllNodes := false
				httpClient := &http.Client{
					Timeout: time.Second * 5,
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
				}
				u := &url.URL{
					Scheme: "https",
					User:   url.UserPassword(esCtrl.AdminUser, password),
					Host:   fmt.Sprintf("%s:%d", podIP, 9200),
				}
				reg.MustRegister(ese.NewClusterHealth(logger, httpClient, u))
				reg.MustRegister(ese.NewNodes(logger, httpClient, u, esAllNodes))
			}
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		return
	case api.ResourcePluralMySQL:
		var reg *prometheus.Registry
		if val, ok := registerers.Get(r.URL.Path); ok {
			reg = val.(*prometheus.Registry)
		} else {
			reg = prometheus.NewRegistry()
			if absent := registerers.SetIfAbsent(r.URL.Path, reg); !absent {
				r2, _ := registerers.Get(r.URL.Path)
				reg = r2.(*prometheus.Registry)
			} else {
				plog.Infof("Configuring exporter for MySQL %s in namespace %s", dbName, namespace)
				conn, err := getMySQLURL(podIP)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				reg.MustRegister(mse.New(conn, mse.Collect{
					GlobalStatus: true,
				}))
			}
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		return
	case api.ResourcePluralMongoDB:
		var reg *prometheus.Registry
		if val, ok := registerers.Get(r.URL.Path); ok {
			reg = val.(*prometheus.Registry)
		} else {
			reg = prometheus.NewRegistry()
			if absent := registerers.SetIfAbsent(r.URL.Path, reg); !absent {
				r2, _ := registerers.Get(r.URL.Path)
				reg = r2.(*prometheus.Registry)
			} else {
				plog.Infof("Configuring exporter for MongoDB %s in namespace %s", dbName, namespace)
				conn, err := getMongoDBURL(podIP)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				reg.MustRegister(mgoe.NewMongodbCollector(mgoe.MongodbCollectorOpts{
					URI: conn,
				}))
			}
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		return
	case api.ResourcePluralRedis:
		var reg *prometheus.Registry
		if val, ok := registerers.Get(r.URL.Path); ok {
			reg = val.(*prometheus.Registry)
		} else {
			reg = prometheus.NewRegistry()
			if absent := registerers.SetIfAbsent(r.URL.Path, reg); !absent {
				r2, _ := registerers.Get(r.URL.Path)
				reg = r2.(*prometheus.Registry)
			} else {
				plog.Infof("Configuring exporter for Redis %s in namespace %s", dbName, namespace)

				conn := fmt.Sprintf("redis://%s:6379", podIP)
				exp, err := rde.NewRedisExporter(
					rde.RedisHost{Addrs: []string{conn}, Aliases: []string{""}},
					"",
					"",
					"")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				reg.MustRegister(exp)
			}
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		return
	case api.ResourcePluralMemcached:
		var reg *prometheus.Registry
		if val, ok := registerers.Get(r.URL.Path); ok {
			reg = val.(*prometheus.Registry)
		} else {
			reg = prometheus.NewRegistry()
			if absent := registerers.SetIfAbsent(r.URL.Path, reg); !absent {
				r2, _ := registerers.Get(r.URL.Path)
				reg = r2.(*prometheus.Registry)
			} else {
				plog.Infof("Configuring exporter for Redis %s in namespace %s", dbName, namespace)

				conn := fmt.Sprintf("%s:11211", podIP)
				reg.MustRegister(memEx.NewExporter(conn, 0)) //timeout: if zero,then default timeout will be used
			}
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

func getPostgresURL(podIP string) (string, error) {
	if _, err := os.Stat(pgCtrl.ExporterSecretPath); err != nil {
		return "", err
	}
	user := pgCtrl.PostgresUser
	password, err := ioutil.ReadFile(filepath.Join(pgCtrl.ExporterSecretPath, pgCtrl.KeyPostgresPassword))
	if err != nil {
		return "", fmt.Errorf("error in reading Password of Postgres: %v", err)
	}
	conn := fmt.Sprintf("postgres://%s:%s@%s:5432/?sslmode=disable", user, password, podIP)
	return conn, nil
}

func getElasticsearchPassword() (string, error) {
	if _, err := os.Stat(esCtrl.ExporterSecretPath); err != nil {
		return "", err
	}
	password, err := ioutil.ReadFile(filepath.Join(esCtrl.ExporterSecretPath, esCtrl.KeyAdminPassword))
	if err != nil {
		return "", fmt.Errorf("error in reading Password of Elasticsearch: %v", err)
	}
	return password, nil
}

func getMySQLURL(podIP string) (string, error) {
	if _, err := os.Stat(msCtrl.ExporterSecretPath); err != nil {
		return "", err
	}
	user, err := ioutil.ReadFile(filepath.Join(msCtrl.ExporterSecretPath, msCtrl.KeyMySQLUser))
	if err != nil {
		return "", fmt.Errorf("error in reading Username of MySQL: %v", err)
	}
	password, err := ioutil.ReadFile(filepath.Join(msCtrl.ExporterSecretPath, msCtrl.KeyMySQLPassword))
	if err != nil {
		return "", fmt.Errorf("error in reading Password of MySQL: %v", err)
	}
	conn := fmt.Sprintf("%s:%s@(%s:3306)/", user, password, podIP)
	return conn, nil
}

func getMongoDBURL(podIP string) (string, error) {
	if _, err := os.Stat(mgCtrl.ExporterSecretPath); err != nil {
		return "", err
	}
	user, err := ioutil.ReadFile(filepath.Join(mgCtrl.ExporterSecretPath, mgCtrl.KeyMongoDBUser))
	if err != nil {
		return "", fmt.Errorf("error in reading Username of MongoDB: %v", err)
	}
	password, err := ioutil.ReadFile(filepath.Join(mgCtrl.ExporterSecretPath, mgCtrl.KeyMongoDBPassword))
	if err != nil {
		return "", fmt.Errorf("error in reading Password of MongoDB: %v", err)
	}
	conn := fmt.Sprintf("mongodb://%s:%s@%s:27017", user, password, podIP)
	return conn, nil
}
