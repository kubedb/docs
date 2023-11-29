---
title: Kubectl-Dba Data
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-data
    name: Kubectl-Dba Data
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba data

Insert, Drop or Verify data in a database

### Synopsis

Insert random data or verify data in a database.

```
kubectl-dba data
```

### Examples

```
  # Insert 100 rows in mysql table
  kubectl dba data insert mysql mysql-demo -n demo --rows=100
  
  # Verify that postgres has at least 100 rows data
  kubectl dba data verify postgres sample-postgres -n demo --rows=100
  
  # Drop all the CLI inserted data from mongodb
  kubectl dba data drop mg -n demo sample-mg
  
  
  Valid resource types include:
  * elasticsearch
  * mongodb
  * mariadb
  * mysql
  * postgres
  * redis
```

### Options

```
  -h, --help   help for data
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
      --insecure-skip-tls-verify              If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                     Path to the kubeconfig file to use for CLI requests.
      --match-server-version                  Require server version to match client version
  -n, --namespace string                      If present, the namespace scope for this CLI request
      --password string                       Password for basic authentication to the API server
      --request-timeout string                The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                         The address and port of the Kubernetes API server
      --tls-server-name string                Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                          Bearer token for authentication to the API server
      --user string                           The name of the kubeconfig user to use
      --username string                       Username for basic authentication to the API server
```

### SEE ALSO

* [kubectl-dba](/docs/reference/cli/kubectl-dba.md)	 - kubectl plugin for KubeDB
* [kubectl-dba data drop](/docs/reference/cli/kubectl-dba_data_drop.md)	 - Drop data from a database
* [kubectl-dba data insert](/docs/reference/cli/kubectl-dba_data_insert.md)	 - Insert random data in a database
* [kubectl-dba data verify](/docs/reference/cli/kubectl-dba_data_verify.md)	 - Verify data in a database

