package main

import (
	"fmt"
	"net/http"

	"github.com/appscode/go/runtime"
	"github.com/appscode/pat"
	tapi "github.com/k8sdb/apimachinery/api"
	esaudit "github.com/k8sdb/elasticsearch/pkg/audit/report"
	pgaudit "github.com/k8sdb/postgres/pkg/audit/report"
)

func ExportSummaryReport(w http.ResponseWriter, r *http.Request) {
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

	kubedbType := params.Get(PathParamType)
	if kubedbType == "" {
		http.Error(w, "Missing parameter "+PathParamType, http.StatusBadRequest)
		return
	}
	kubedbName := params.Get(PathParamName)
	if kubedbName == "" {
		http.Error(w, "Missing parameter "+PathParamName, http.StatusBadRequest)
		return
	}

	index := r.URL.Query().Get("index")

	switch kubedbType {
	case tapi.ResourceTypePostgres:
		pgaudit.ExportReport(kubeClient, dbClient, namespace, kubedbName, index, w)
	case tapi.ResourceTypeElastic:
		esaudit.ExportReport(kubeClient, dbClient, namespace, kubedbName, index, w)
	default:
		http.Error(w, fmt.Sprintf(`Invalid kubedb type "%v"`, kubedbType), http.StatusBadRequest)
	}
}
