---
title: Kubectl-Dba Exec
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-exec
    name: Kubectl-Dba Exec
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba exec

Execute script or command to a database.

### Synopsis

Execute commands or scripts to a database.

```
kubectl-dba exec
```

### Options

```
  -h, --help   help for exec
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
* [kubectl-dba exec mariadb](/docs/reference/cli/kubectl-dba_exec_mariadb.md)	 - Execute SQL commands to a mariadb resource
* [kubectl-dba exec mongodb](/docs/reference/cli/kubectl-dba_exec_mongodb.md)	 - Execute commands to a mongodb resource
* [kubectl-dba exec mysql](/docs/reference/cli/kubectl-dba_exec_mysql.md)	 - Execute SQL commands to a mysql resource
* [kubectl-dba exec postgres](/docs/reference/cli/kubectl-dba_exec_postgres.md)	 - Execute SQL commands to a postgres resource
* [kubectl-dba exec redis](/docs/reference/cli/kubectl-dba_exec_redis.md)	 - Execute SQL commands to a redis resource

