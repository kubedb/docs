package main

import (
	"fmt"
	"net/http"

	"github.com/appscode/go/runtime"
	"github.com/appscode/pat"
	tapi "github.com/k8sdb/apimachinery/api"
	pg "github.com/k8sdb/postgres/pkg/audit/summary"
)

func ExportSummaryReport(w http.ResponseWriter, r *http.Request) {
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

	kubedbType := params.Get(ParamType)
	if kubedbType == "" {
		http.Error(w, "Missing parameter "+ParamType, http.StatusBadRequest)
		return
	}
	kubedbName := params.Get(ParamName)
	if kubedbName == "" {
		http.Error(w, "Missing parameter "+ParamName, http.StatusBadRequest)
		return
	}

	switch kubedbType {
	case tapi.ResourceTypePostgres:
		dbname := r.URL.Query().Get("dbname")
		pg.GetSummaryReport(kubeClient, dbClient, namespace, kubedbName, dbname, w)
	case tapi.ResourceTypeElastic:
		return
	default:
		http.Error(w, fmt.Sprintf(`Invalid kubedb type "%v"`, kubedbType), http.StatusBadRequest)
	}
}
