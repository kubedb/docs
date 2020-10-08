/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"text/template"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha2"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

			appBinding, err := c.AppCatalogClient.AppcatalogV1alpha1().AppBindings(namespace).Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					log.Warning(err)
				} else {
					log.Error(err)
				}
				continue //Dont add pgbouncer databse base for this non existent appBinding
			}
			var hostname string
			if appBinding.Spec.ClientConfig.URL == nil {
				if appBinding.Spec.ClientConfig.Service != nil {
					hostname = appBinding.Spec.ClientConfig.Service.Name + "." + namespace + ".svc"
					hostPort := appBinding.Spec.ClientConfig.Service.Port
					buf.WriteString(fmt.Sprint(db.Alias, "= host=", hostname, " port=", hostPort, " dbname=", db.DatabaseName))
				}
			} else {
				// Reminder URL should contain host=localhost port=5432
				// TODO: Test against RDS
				buf.WriteString(fmt.Sprint(db.Alias + " = " + *(appBinding.Spec.ClientConfig.URL) + " dbname=" + db.DatabaseName))
			}
			if db.DatabaseSecretRef != nil {
				secret, err := c.Client.CoreV1().Secrets(pgbouncer.Namespace).Get(context.TODO(), db.DatabaseSecretRef.Name, metav1.GetOptions{})
				if err == nil {
					buf.WriteString(fmt.Sprint(" user=", string(secret.Data["username"])))
					buf.WriteString(fmt.Sprint(" password=", string(secret.Data["password"])))
				}
			}
			buf.WriteRune('\n')
		}
	}

	buf.WriteString("\n[pgbouncer]\n")
	buf.WriteString("logfile = /tmp/pgbouncer.log\n")
	buf.WriteString("pidfile = /tmp/pgbouncer.pid\n")

	if pgbouncer.Spec.TLS != nil {
		if pgbouncer.Spec.TLS.IssuerRef != nil {
			//SSL is enabled
			buf.WriteString("client_tls_sslmode = verify-full\n")
			buf.WriteString(fmt.Sprintln("client_tls_ca_file = " + filepath.Join(ServingServerCertMountPath, "ca.crt")))
			buf.WriteString(fmt.Sprintln("client_tls_key_file = " + filepath.Join(ServingServerCertMountPath, "tls.key")))
			buf.WriteString(fmt.Sprintln("client_tls_cert_file = " + filepath.Join(ServingServerCertMountPath, "tls.crt")))
		}
	}
	upstreamServerCAExists, err := c.isUpStreamServerCAExist(pgbouncer)
	if err != nil {
		log.Infoln(err)
		return "", err
	}
	if upstreamServerCAExists {
		buf.WriteString("server_tls_sslmode = verify-full\n")
		buf.WriteString(fmt.Sprintln("server_tls_ca_file = " + filepath.Join(UserListMountPath, api.PgBouncerUpstreamServerCA)))
	}

	secretFileName, err := c.getUserListFileName(pgbouncer)
	if err != nil {
		log.Infoln(err)
		return "", err
	}

	if pgbouncer.Spec.ConnectionPool == nil || (pgbouncer.Spec.ConnectionPool != nil && pgbouncer.Spec.ConnectionPool.AuthType != "any") {
		buf.WriteString(fmt.Sprintln("auth_file = ", filepath.Join(UserListMountPath, secretFileName)))
	}
	//TODO: what about auth type md5 and or something else?
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
	owner := metav1.NewControllerRef(pgbouncer, api.SchemeGroupVersion.WithKind(api.ResourceKindPgBouncer))

	cfg, err := c.generateConfig(pgbouncer)
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	_, vt, err := core_util.CreateOrPatchConfigMap(context.TODO(), c.Client, configMapMeta, func(in *core.ConfigMap) *core.ConfigMap {
		in.Labels = pgbouncer.OffshootLabels()
		core_util.EnsureOwnerReference(&in.ObjectMeta, owner)
		in.Data = map[string]string{
			pbConfigFile: cfg,
		}
		return in
	}, metav1.PatchOptions{})

	return vt, err
}

func (c *Controller) getUserListFileName(pgbouncer *api.PgBouncer) (string, error) {
	defaultSecretSpec := c.GetDefaultSecretSpec(pgbouncer)
	defaultSecret, err := c.Client.CoreV1().Secrets(pgbouncer.Namespace).Get(context.TODO(), defaultSecretSpec.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if data, exists := defaultSecret.Data[pbUserData]; exists && data != nil {
		return pbUserData, nil
	}
	return pbAdminData, nil
}

func (c *Controller) isUpStreamServerCAExist(pgbouncer *api.PgBouncer) (bool, error) {
	defaultSecretSpec := c.GetDefaultSecretSpec(pgbouncer)
	defaultSecret, err := c.Client.CoreV1().Secrets(pgbouncer.Namespace).Get(context.TODO(), defaultSecretSpec.Name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if _, exists := defaultSecret.Data[api.PgBouncerUpstreamServerCA]; exists {
		return true, nil
	}
	return false, nil
}
