package cmds

import (
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
	mgCtrl "github.com/kubedb/mongodb/pkg/controller"
	msCtrl "github.com/kubedb/mysql/pkg/controller"
	rde "github.com/oliver006/redis_exporter/exporter"
	"github.com/orcaman/concurrent-map"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	plog "github.com/prometheus/common/log"
	memEx "github.com/prometheus/memcached_exporter/exporter"
	mse "github.com/prometheus/mysqld_exporter/collector"
	pge "github.com/wrouesnel/postgres_exporter/exporter"
	"gopkg.in/ini.v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	case api.ResourceTypePostgres:
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
				db, err := dbClient.Postgreses(namespace).Get(dbName, metav1.GetOptions{})
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
	case api.ResourceTypeElasticsearch:
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
				_, err := dbClient.Elasticsearchs(namespace).Get(dbName, metav1.GetOptions{})
				if kerr.IsNotFound(err) {
					http.NotFound(w, r)
					return
				} else if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				esURI := fmt.Sprintf("http://%s:9200", podIP)
				esURL, err := url.Parse(esURI)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				esTimeout := 5 * time.Second
				esAllNodes := false
				httpClient := &http.Client{Timeout: esTimeout}
				reg.MustRegister(ese.NewClusterHealth(logger, httpClient, esURL))
				reg.MustRegister(ese.NewNodes(logger, httpClient, esURL, esAllNodes))
			}
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		return
	case api.ResourceTypeMySQL:
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
	case api.ResourceTypeMongoDB:
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
				fmt.Println(">>>>>.conn: ", conn) //todo: delete
				reg.MustRegister(mgoe.NewMongodbCollector(mgoe.MongodbCollectorOpts{
					URI: conn,
				}))
			}
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		return
	case api.ResourceTypeRedis:
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
				_, err := dbClient.Redises(namespace).Get(dbName, metav1.GetOptions{})
				if kerr.IsNotFound(err) {
					http.NotFound(w, r)
					return
				} else if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				conn := fmt.Sprintf("redis://%s:6379", podIP)
				exp, err := rde.NewRedisExporter(
					rde.RedisHost{Addrs: []string{conn}, Aliases: []string{""}},
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
	case api.ResourceTypeMemcached:
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
				_, err := dbClient.Memcacheds(namespace).Get(dbName, metav1.GetOptions{})
				if kerr.IsNotFound(err) {
					http.NotFound(w, r)
					return
				} else if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				conn := fmt.Sprintf("%s:11211", podIP)
				reg.MustRegister(memEx.NewExporter(conn, 0)) //timeout: if zero,then default timeout will be used
			}
		}
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

func getPostgresURL(db *api.Postgres, podIP string) (string, error) {
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

func getMySQLURL(podIP string) (string, error) {
	if _, err := os.Stat(msCtrl.ExporterSecretPath); err != nil {
		return "", err
	}
	userData, _ := ioutil.ReadFile(filepath.Join(msCtrl.ExporterSecretPath, msCtrl.KeyMySQLUser))
	passwordData, _ := ioutil.ReadFile(filepath.Join(msCtrl.ExporterSecretPath, msCtrl.KeyMySQLPassword))

	conn := fmt.Sprintf("%s:%s@(%s:3306)/", userData, passwordData, podIP)
	return conn, nil
}

func getMongoDBURL(podIP string) (string, error) {
	if _, err := os.Stat(mgCtrl.ExporterSecretPath); err != nil {
		return "", err
	}
	userData, _ := ioutil.ReadFile(filepath.Join(mgCtrl.ExporterSecretPath, mgCtrl.KeyMongoDBUser))
	passwordData, _ := ioutil.ReadFile(filepath.Join(mgCtrl.ExporterSecretPath, mgCtrl.KeyMongoDBPassword))

	conn := fmt.Sprintf("mongodb://%s:%s@%s:27017", userData, passwordData, podIP)
	return conn, nil
}
