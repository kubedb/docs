---
title: Kubectl-Dba Dc-Dr Debug Fence
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-dc-dr-debug-fence
    name: Kubectl-Dba Dc-Dr Debug Fence
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba dc-dr debug fence

Diagnose a database whose primary is fenced read-only

### Synopsis

A DC-DR database goes read-only by design when its local marker is missing, stale past its TTL, or names another data center, so that at most one data center is ever writable. This reports which of those applies, whether the authority itself is healthy, and what to do when the coordination plane is the thing that is broken.

 KUBECONFIG: the hub cluster; the coordination plane via --coord-*.

```
kubectl-dba dc-dr debug fence DB_NAME [flags]
```

### Examples

```
  kubectl dba dc-dr debug fence pg-dcdr -n demo
```

### Options

```
      --coord-kubeconfig string             Path to a kubeconfig file for the coordination control plane (overrides the secret/configmap sources)
      --coord-kubeconfig-configmap string   ConfigMap ([namespace/]name, key "kubeconfig") on the current cluster holding the coordination-plane kubeconfig
      --coord-kubeconfig-secret string      Secret ([namespace/]name, key "kubeconfig") on the current cluster holding the coordination-plane kubeconfig (default "dc-failover/coord-kubeconfig")
      --coord-namespace string              Namespace on the coordination plane that holds the primary-DC Leases (default "dc-failover")
  -h, --help                                help for fence
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

* [kubectl-dba dc-dr debug](/docs/reference/cli/kubectl-dba_dc-dr_debug.md)	 - Diagnose DC-DR symptoms: failover not happening, switchover stuck, fenced database

