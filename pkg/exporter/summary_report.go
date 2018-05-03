package exporter

import (
	"fmt"
	"net/http"

	"github.com/appscode/go/runtime"
	"github.com/appscode/pat"
	tapi "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	esaudit "github.com/kubedb/elasticsearch/pkg/audit/report"
	pgaudit "github.com/kubedb/postgres/pkg/audit/report"
)

func (e Options) ExportSummaryReport(w http.ResponseWriter, r *http.Request) {
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
	case tapi.ResourceSingularPostgres:
		pgaudit.ExportReport(e.KubeClient, e.DbClient, namespace, kubedbName, index, w)
	case tapi.ResourceSingularElasticsearch:
		esaudit.ExportReport(e.KubeClient, e.DbClient, namespace, kubedbName, index, w)
	default:
		http.Error(w, fmt.Sprintf(`Invalid kubedb type "%v"`, kubedbType), http.StatusBadRequest)
	}
}
