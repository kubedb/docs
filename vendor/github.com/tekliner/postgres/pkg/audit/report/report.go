package report

import (
	"encoding/json"
	"fmt"
	"net/http"

	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	cs "github.com/kubedb/apimachinery/client/clientset/versioned"
	"github.com/kubedb/postgres/pkg/controller"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ExportReport(
	kubeClient kubernetes.Interface,
	dbClient cs.Interface,
	namespace string,
	kubedbName string,
	dbname string,
	w http.ResponseWriter,
) {
	startTime := metav1.Now()

	postgres, err := dbClient.KubedbV1alpha1().Postgreses(namespace).Get(kubedbName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			http.Error(w, fmt.Sprintf(`Postgres "%v" not found`, kubedbName), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	secret, err := kubeClient.CoreV1().Secrets(namespace).Get(postgres.Spec.DatabaseSecret.SecretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			http.Error(w, fmt.Sprintf(`Secret "%v" not found`, postgres.Spec.DatabaseSecret.SecretName), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	username := string(secret.Data[controller.PostgresUser])
	password := string(secret.Data[controller.PostgresPassword])

	host := fmt.Sprintf("%v.%v", kubedbName, namespace)
	port := controller.PostgresPort

	databases := make([]string, 0)
	if dbname == "" {
		engine, err := newXormEngine(username, password, host, port, "postgres")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		databases, err = getAllDatabase(engine)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		databases = append(databases, dbname)
	}

	pgSummary := make(map[string]*api.PostgresSummary)
	for _, db := range databases {
		engine, err := newXormEngine(username, password, host, port, db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		info, err := getDataFromDB(engine)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pgSummary[db] = info
	}

	completionTime := metav1.Now()

	r := &api.Report{
		TypeMeta:   postgres.TypeMeta,
		ObjectMeta: postgres.ObjectMeta,
		Summary: api.ReportSummary{
			Postgres: pgSummary,
		},
		Status: api.ReportStatus{
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
