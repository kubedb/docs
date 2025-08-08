---
title: Kubectl-Dba
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba
    name: Kubectl-Dba
    parent: reference-cli
    weight: 0

menu_name: docs_{{ .version }}
section_menu_id: reference
url: /docs/{{ .version }}/reference/cli/
aliases:
- /docs/{{ .version }}/reference/cli/kubectl-dba/
---
## kubectl-dba

kubectl plugin for KubeDB

### Synopsis

kubectl plugin for KubeDB by AppsCode - Kubernetes ready production-grade Databases

 Find more information at https://kubedb.com

```
kubectl-dba [flags]
```

### Options

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
  -h, --help                                  help for kubectl-dba
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

* [kubectl-dba completion](/docs/reference/cli/kubectl-dba_completion.md)	 - Generate completion script
* [kubectl-dba connect](/docs/reference/cli/kubectl-dba_connect.md)	 - Connect to a database.
* [kubectl-dba data](/docs/reference/cli/kubectl-dba_data.md)	 - Insert, Drop or Verify data in a database
* [kubectl-dba debug](/docs/reference/cli/kubectl-dba_debug.md)	 - Debug any Database issue
* [kubectl-dba describe](/docs/reference/cli/kubectl-dba_describe.md)	 - Show details of a specific resource or group of resources
* [kubectl-dba exec](/docs/reference/cli/kubectl-dba_exec.md)	 - Execute script or command to a database.
* [kubectl-dba monitor](/docs/reference/cli/kubectl-dba_monitor.md)	 - Monitoring related commands for a database
* [kubectl-dba mssql](/docs/reference/cli/kubectl-dba_mssql.md)	 - MSSQLServer database commands
* [kubectl-dba options](/docs/reference/cli/kubectl-dba_options.md)	 - Print the list of flags inherited by all commands
* [kubectl-dba pause](/docs/reference/cli/kubectl-dba_pause.md)	 - Pause the processing of an object.
* [kubectl-dba remote-config](/docs/reference/cli/kubectl-dba_remote-config.md)	 - generate appbinding , secrets for remote replica
* [kubectl-dba restart](/docs/reference/cli/kubectl-dba_restart.md)	 - Smartly restart the pods of the database.
* [kubectl-dba resume](/docs/reference/cli/kubectl-dba_resume.md)	 - Resume processing of an object.
* [kubectl-dba show-credentials](/docs/reference/cli/kubectl-dba_show-credentials.md)	 - Prints credentials of the database.
* [kubectl-dba version](/docs/reference/cli/kubectl-dba_version.md)	 - Prints binary version number.

