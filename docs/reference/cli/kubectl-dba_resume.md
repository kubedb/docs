---
title: Kubectl-Dba Resume
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-resume
    name: Kubectl-Dba Resume
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba resume

Resume processing of an object.

### Synopsis

Resume the community-operator's watch for the objects. The community-operator will continue to process the object.

Use "kubectl api-resources" for a complete list of supported resources.

```
kubectl-dba resume (-f FILENAME | TYPE [NAME_PREFIX | -l label] | TYPE/NAME)
```

### Examples

```
  # Resume an elasticsearch
  dba resume elasticsearch elasticsearch-demo
  
  # Resume a postgres
  dba resume pg/postgres-demo
  
  # Resume all postgres
  dba resume postgreses
  
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
      --all-namespaces      If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.
  -f, --filename strings    Filename, directory, or URL to files containing the resource to resume
  -h, --help                help for resume
  -k, --kustomize string    Process the kustomization directory. This flag can't be used together with -f or -R.
      --only-backupconfig   If provided, only the backupconfiguration for the database is resumed.
      --only-db             If provided, only the database is resumed.
  -R, --recursive           Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.
  -l, --selector string     Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default cache directory (default "/home/runner/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --enable-analytics               Send analytical events to Google Analytics (default true)
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
      --match-server-version           Require server version to match client version
  -n, --namespace string               If present, the namespace scope for this CLI request
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
```

### SEE ALSO

* [kubectl-dba](/docs/reference/cli/kubectl-dba.md)	 - kubectl plugin for KubeDB

