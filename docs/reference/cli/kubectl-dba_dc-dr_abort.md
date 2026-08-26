---
title: Kubectl-Dba Dc-Dr Abort
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-dc-dr-abort
    name: Kubectl-Dba Dc-Dr Abort
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba dc-dr abort

Abort an in-flight planned switchover and restore writes to the current active DC

### Synopsis

Sets the dr.kubedb.com/switchover-abort annotation. Its PRESENCE aborts: the hub clears the quiesce so the original active data center resumes accepting writes, and removes every switchover annotation once done.

 Do NOT abort by deleting the switchover-to annotation: in a scope shared by several databases the hub re-propagates it to every sibling each pass, so a bare removal silently reappears. This explicit abort signal is propagated and honored scope-wide. A switchover that cannot complete also auto-aborts on its own after the switchover timeout (default 10m, dr.kubedb.com/switchover-timeout to override).

 KUBECONFIG: the hub cluster.

```
kubectl-dba dc-dr abort DB_NAME [flags]
```

### Examples

```
  kubectl dba dc-dr abort pg-dcdr -n demo
```

### Options

```
  -h, --help   help for abort
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
      --request-timeout string                The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                         The address and port of the Kubernetes API server
      --tls-server-name string                Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                          Bearer token for authentication to the API server
      --user string                           The name of the kubeconfig user to use
      --username string                       Username for basic authentication to the API server
```

### SEE ALSO

* [kubectl-dba dc-dr](/docs/reference/cli/kubectl-dba_dc-dr.md)	 - Cross data center DR operations: switchover, failover, pins, and diagnosis

