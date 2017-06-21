package summary

import (
	"encoding/json"
	"fmt"
	"net/http"

	tcs "github.com/k8sdb/apimachinery/client/clientset"
	"github.com/k8sdb/elasticsearch/pkg/audit/type"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	clientset "k8s.io/client-go/kubernetes"
)

func GetSummaryReport(
	kubeClient clientset.Interface,
	dbClient tcs.ExtensionInterface,
	namespace string,
	kubedbName string,
	index string,
	w http.ResponseWriter,
) {

	if _, err := dbClient.Elastics(namespace).Get(kubedbName); err != nil {
		if kerr.IsNotFound(err) {
			http.Error(w, fmt.Sprintf(`Elastic "%v" not found`, kubedbName), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	host := fmt.Sprintf("%v.%v", kubedbName, namespace)
	port := "9200"

	client, err := newClient(host, port)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	indices := make([]string, 0)
	if index == "" {
		indices, err = getAllIndices(client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		indices = append(indices, index)
	}

	infos := make(map[string]*types.IndexInfo)
	for _, index := range indices {
		info, err := getDataFromIndex(client, index)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		infos[index] = info
	}

	data, err := json.MarshalIndent(infos, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, string(data))
	} else {
		http.Error(w, "audit data not found", http.StatusNotFound)
	}
}
