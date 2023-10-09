---
title: Kubectl-Dba Remote-Config Mysql
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-remote-config-mysql
    name: Kubectl-Dba Remote-Config Mysql
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba remote-config mysql

generate appbinding , secrets for remote replica

### Synopsis

generate appbinding , secrets for remote replica

```
kubectl-dba remote-config mysql [flags]
```

### Examples

```
kubectl dba remote-config mysql -n <ns> -u <user_name> -p$<password> -d<dns_name>  <db_name>
 kubectl dba remote-config mysql -n <ns> -u <user_name> -p$<password> -d<dns_name>  <db_name> 

```

### Options

```
  -d, --dns string         dns name for the remote replica (default "localhost")
  -h, --help               help for mysql
  -n, --namespace string   host namespace for the remote replica (default "default")
  -p, --pass string        password name for the remote replica (default "password")
  -u, --user string        user name for the remote replica (default "postgres")
  -y, --yes                permission for alter password  for the remote replica
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
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --username string                Username for basic authentication to the API server
```

### SEE ALSO

* [kubectl-dba remote-config](/docs/reference/cli/kubectl-dba_remote-config.md)	 - generate appbinding , secrets for remote replica

