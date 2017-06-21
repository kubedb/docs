package report

import (
	"encoding/json"
	"fmt"
	"net/http"

	tapi "github.com/k8sdb/apimachinery/api"
	tcs "github.com/k8sdb/apimachinery/client/clientset"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

func ExportReport(
	kubeClient clientset.Interface,
	dbClient tcs.ExtensionInterface,
	namespace string,
	kubedbName string,
	index string,
	w http.ResponseWriter,
) {
	startTime := metav1.Now()

	elastic, err := dbClient.Elastics(namespace).Get(kubedbName)
	if err != nil {
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

	esSummary := make(map[string]*tapi.ElasticSummary)
	for _, index := range indices {
		info, err := getDataFromIndex(client, index)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		esSummary[index] = info
	}

	completionTime := metav1.Now()

	r := &tapi.Report{
		TypeMeta:   elastic.TypeMeta,
		ObjectMeta: elastic.ObjectMeta,
		Summary: tapi.ReportSummary{
			Elastic: esSummary,
		},
		Status: tapi.ReportStatus{
			StartTime:      &startTime,
			CompletionTime: &completionTime,
		},
	}
	r.ResourceVersion = ""
	r.SelfLink = ""
	r.UID = ""

	data, err := json.MarshalIndent(r, "", "  ")
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
