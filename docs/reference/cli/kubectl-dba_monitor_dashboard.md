---
title: Kubectl-Dba Monitor Dashboard
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-monitor-dashboard
    name: Kubectl-Dba Monitor Dashboard
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba monitor dashboard

Check availability of a grafana dashboard

### Synopsis

Check availability of metrics in prometheus server used in a grafana dashboard.

```
kubectl-dba monitor dashboard
```

### Examples

```
  kubectl dba monitor dashboard [DATABASE] [DASHBOARD_NAME] \
  --prom-svc=[PROM_SVC_NAME] --prom-svc-namespace=[PROM_SVC_NS] --prom-svc-port=[PROM_SVC_PORT]
  
  # Check availability of a postgres grafana dashboard
  kubectl-dba monitor dashboard postgres postgres_databases_dashboard \
  --prom-svc-name=prometheus-kube-prometheus-prometheus --prom-svc-namespace=monitoring --prom-svc-port=9090
  
  Valid dashboards include:
  * elasticsearch
  * kafka
  * mariadb
  * mongodb
  * mysql
  * perconaxtradb
  * postgres
  * proxysql
  * redis
```

### Options

```
  -b, --branch string   branch name of the github repo (default "master")
  -h, --help            help for dashboard
```

### Options inherited from parent commands

```
      --as string                             Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray                  Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                         UID to impersonate for the operation.
      --cache-dir string                      Default cache directory (default "/home/runner/.kube/cache")
      --certificate-authority string          Path to a cert file for the certificate authority
      --client-certificate string             Path to a client certificate file for TLS
      --client-key string                     Path to a client key file for TLS
      --cluster string                        The name of the kubeconfig cluster to use
      --context string                        The name of the kubeconfig context to use
      --default-seccomp-profile-type string   Default seccomp profile
      --disable-compression                   If true, opt-out of response compression for all requests to the server
      --insecure-skip-tls-verify              If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                     Path to the kubeconfig file to use for CLI requests.
      --match-server-version                  Require server version to match client version
  -n, --namespace string                      If present, the namespace scope for this CLI request
      --password string                       Password for basic authentication to the API server
      --prom-svc-name string                  name of the prometheus service
      --prom-svc-namespace string             namespace of the prometheus service
      --prom-svc-port int                     port of the prometheus service (default 9090)
      --request-timeout string                The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                         The address and port of the Kubernetes API server
      --tls-server-name string                Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                          Bearer token for authentication to the API server
      --user string                           The name of the kubeconfig user to use
      --username string                       Username for basic authentication to the API server
```

### SEE ALSO

* [kubectl-dba monitor](/docs/reference/cli/kubectl-dba_monitor.md)	 - Monitoring related commands for a database

