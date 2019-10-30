/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
)

const (
	PostgresPassword = "POSTGRES_PASSWORD"
	PostgresUser     = "POSTGRES_USER"
	DefaultHostPort  = 5432
	pbConfigFile     = "pgbouncer.ini"
)

var (
	cfgtpl = template.Must(template.New("cfg").Parse(`listen_port = {{ .Port }}
listen_addr = *
pool_mode = {{ .PoolMode }}
ignore_startup_parameters = extra_float_digits{{ if .IgnoreStartupParameters }}, {{.IgnoreStartupParameters}}{{ end }}
{{- if .MaxClientConnections  }}
max_client_conn = {{ .MaxClientConnections }}
{{- end }}
{{- if .MaxDBConnections  }}
max_db_connections = {{ .MaxDBConnections }}
{{- end }}
{{- if .MaxUserConnections  }}
max_user_connections = {{ .MaxUserConnections }}
{{- end }}
{{- if .MinPoolSize  }}
min_pool_size = {{ .MinPoolSize }}
{{- end }}
{{- if .DefaultPoolSize  }}
default_pool_size = {{ .DefaultPoolSize }}
{{- end }}
{{- if .ReservePoolSize  }}
reserve_pool_size = {{ .ReservePoolSize }}
{{- end }}
{{- if .ReservePoolTimeoutSeconds  }}
reserve_pool_timeout = {{ .ReservePoolTimeoutSeconds }}
{{- end }}
{{- if .StatsPeriodSeconds  }}
stats_period = {{ .StatsPeriodSeconds }}
{{- end }}
{{- if .AuthType }}
auth_type = {{ .AuthType }}
{{- end }}
{{- if .AuthUser }}
auth_user = {{ .AuthUser }}
{{- end }}
admin_users = kubedb{{range .AdminUsers }},{{.}}{{end}}
`))
)

func (c *Controller) generateConfig(pgbouncer *api.PgBouncer) (string, error) {
	var buf bytes.Buffer
	buf.WriteString("[databases]\n")
	if pgbouncer.Spec.Databases != nil {
		for _, db := range pgbouncer.Spec.Databases {
			name := db.DatabaseRef.Name
			namespace := pgbouncer.GetNamespace()

			appBinding, err := c.AppCatalogClient.AppcatalogV1alpha1().AppBindings(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					log.Warning(err)
				} else {
					log.Error(err)
				}
				continue //Dont add pgbouncer databse base for this non existent appbinding
			}
			//if appBinding.Spec.ClientConfig.Service != nil {
			//	name = appBinding.Spec.ClientConfig.Service.Name
			//	namespace = appBinding.Namespace
			//	hostPort = appBinding.Spec.ClientConfig.Service.Port
			//}
			var hostname string
			if appBinding.Spec.ClientConfig.URL == nil {
				if appBinding.Spec.ClientConfig.Service != nil {
					//urlString, err := appBinding.URL()
					//if err != nil {
					//	log.Errorln(err)
					//}
					//parsedURL, err := pq.ParseURL(urlString)
					//if err != nil {
					//	log.Errorln(err)
					//}
					//println(":::Parsed URL= ", parsedURL)
					//parsedURL = strings.ReplaceAll(parsedURL, " sslmode=disable","")
					//dbinfo = dbinfo +fmt.Sprintln(db.Alias +" = " + parsedURL +" dbname="+db.DbName )
					hostname = appBinding.Spec.ClientConfig.Service.Name + "." + namespace + ".svc"
					hostPort := appBinding.Spec.ClientConfig.Service.Port
					buf.WriteString(fmt.Sprint(db.Alias, "= host=", hostname, " port=", hostPort, " dbname=", db.DatabaseName))
					//dbinfo = dbinfo +fmt.Sprintln(db.Alias +"x = host=" + hostname +" port="+strconv.Itoa(int(hostPort))+" dbname="+db.DbName )
				}
			} else {
				//Reminder URL should contain host=localhost port=5432
				buf.WriteString(fmt.Sprint(db.Alias + " = " + *(appBinding.Spec.ClientConfig.URL) + " dbname=" + db.DatabaseName))
			}
			if db.UserName != "" {
				buf.WriteString(fmt.Sprint(" user=", db.UserName))
			}
			if db.Password != "" {
				buf.WriteString(fmt.Sprint(" password=", db.Password))
			}
			buf.WriteRune('\n')
		}
	}

	buf.WriteString("\n[pgbouncer]\n")
	buf.WriteString("logfile = /tmp/pgbouncer.log\n") // TODO: send log to stdout ?
	buf.WriteString("pidfile = /tmp/pgbouncer.pid\n")

	secretFileName, err := c.getUserListFileName(pgbouncer)
	if err != nil {
		return "", err
	}
	if secretFileName == "" {
		secretFileName = pbAdminData
	}

	if pgbouncer.Spec.ConnectionPool == nil || (pgbouncer.Spec.ConnectionPool != nil && pgbouncer.Spec.ConnectionPool.AuthType != "any") {
		buf.WriteString(fmt.Sprintln("auth_file = ", filepath.Join(userListMountPath, secretFileName)))
	}
	if pgbouncer.Spec.ConnectionPool != nil {
		err := cfgtpl.Execute(&buf, pgbouncer.Spec.ConnectionPool)
		if err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

func (c *Controller) ensureConfigMapFromCRD(pgbouncer *api.PgBouncer) (kutil.VerbType, error) {
	configMapMeta := metav1.ObjectMeta{
		Name:      pgbouncer.OffshootName(),
		Namespace: pgbouncer.Namespace,
	}
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, pgbouncer)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	cfg, err := c.generateConfig(pgbouncer)
	if err != nil {
		return kutil.VerbUnchanged, rerr
	}

	_, vt, err := core_util.CreateOrPatchConfigMap(c.Client, configMapMeta, func(in *core.ConfigMap) *core.ConfigMap {
		in.Labels = pgbouncer.OffshootLabels()
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Data = map[string]string{
			pbConfigFile: cfg,
		}
		return in
	})
	return vt, err
}

func (c *Controller) getUserListFileName(pgbouncer *api.PgBouncer) (string, error) {
	defaultSecretSpec := c.GetDefaultSecretSpec(pgbouncer)
	defaultSecret, err := c.Client.CoreV1().Secrets(pgbouncer.Namespace).Get(defaultSecretSpec.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if _, exists := defaultSecret.Data[pbUserData]; exists {
		return pbUserData, nil
	}
	return pbAdminData, nil
}
