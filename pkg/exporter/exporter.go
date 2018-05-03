package exporter

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/appscode/go/log"
	"github.com/appscode/pat"
	tcs "github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	MasterURL      string
	KubeconfigPath string
	Address        string

	KubeClient kubernetes.Interface
	DbClient   tcs.KubedbV1alpha1Interface
}

func New(
	masterURL string,
	kubeconfigPath string,
	address string,
	kubeClient kubernetes.Interface,
	dbClient tcs.KubedbV1alpha1Interface,
) Options {
	return Options{
		MasterURL:      masterURL,
		KubeconfigPath: kubeconfigPath,
		Address:        address,
		KubeClient:     kubeClient,
		DbClient:       dbClient,
	}
}

func (e Options) Export() {
	fmt.Println("Starting exporter...")

	m := pat.New()
	m.Get("/metrics", promhttp.Handler())
	pattern := fmt.Sprintf("/kubedb.com/v1alpha1/namespaces/%s/%s/%s/metrics", PathParamNamespace, PathParamType, PathParamName)
	log.Infoln("URL pattern:", pattern)
	m.Get(pattern, http.HandlerFunc(ExportMetrics))
	m.Del(pattern, http.HandlerFunc(DeleteRegistry))
	http.Handle("/", m)

	log.Infof("Starting Server: %s", e.Address)
	log.Fatal(http.ListenAndServe(e.Address, nil))
}
