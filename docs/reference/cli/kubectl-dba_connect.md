---
title: Kubectl-Dba Connect
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-connect
    name: Kubectl-Dba Connect
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba connect

Connect to a database.

### Synopsis

Connect to a database.

```
kubectl-dba connect
```

### Options

```
  -h, --help   help for connect
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                  UID to impersonate for the operation.
      --cache-dir string               Default cache directory (default "/home/runner/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
      --match-server-version           Require server version to match client version
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
```

### SEE ALSO

* [kubectl-dba](/docs/reference/cli/kubectl-dba.md)	 - kubectl plugin for KubeDB
* [kubectl-dba connect elasticsearch](/docs/reference/cli/kubectl-dba_connect_elasticsearch.md)	 - Connect to a shell to run elasticsearch api calls
* [kubectl-dba connect mariadb](/docs/reference/cli/kubectl-dba_connect_mariadb.md)	 - Connect to a mariadb object
* [kubectl-dba connect memcached](/docs/reference/cli/kubectl-dba_connect_memcached.md)	 - Connect to a telnet shell to run command against a memcached database
* [kubectl-dba connect mongodb](/docs/reference/cli/kubectl-dba_connect_mongodb.md)	 - Connect to a mongodb object
* [kubectl-dba connect mysql](/docs/reference/cli/kubectl-dba_connect_mysql.md)	 - Connect to a mysql object
* [kubectl-dba connect postgres](/docs/reference/cli/kubectl-dba_connect_postgres.md)	 - Connect to a postgres object
* [kubectl-dba connect redis](/docs/reference/cli/kubectl-dba_connect_redis.md)	 - Connect to a redis object's pod

