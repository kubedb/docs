---
title: Kubectl-Dba Dc-Dr Active-Dc
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-dc-dr-active-dc
    name: Kubectl-Dba Dc-Dr Active-Dc
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba dc-dr active-dc

Print the data center that currently holds the primary role

### Synopsis

Reads the primary-DC Lease from the coordination control plane, the authority for which data center is active.

 Given a database name, its failover scope is resolved first (the PlacementPolicy's failoverPolicy trigger, exactly as the operator resolves it) and the matching Lease is read. Given --lease, that Lease is read directly, which also works for a scope whose database is gone.

 KUBECONFIG: the hub cluster (to read the Postgres and its PlacementPolicy, and by default to read the coordination kubeconfig Secret). The coordination plane itself is reached with the --coord-* flags.

```
kubectl-dba dc-dr active-dc [DB_NAME] [--lease NAME] [flags]
```

### Examples

```
  # By database
  kubectl dba dc-dr active-dc pg-dcdr -n demo
  
  # By Lease name, with an explicit coordination kubeconfig file
  kubectl dba dc-dr active-dc --lease primary-dc --coord-kubeconfig /tmp/coord.yaml
  
  # Scriptable: just the DC name
  kubectl dba dc-dr active-dc pg-dcdr -n demo -q
```

### Options

```
      --coord-kubeconfig string             Path to a kubeconfig file for the coordination control plane (overrides the secret/configmap sources)
      --coord-kubeconfig-configmap string   ConfigMap ([namespace/]name, key "kubeconfig") on the current cluster holding the coordination-plane kubeconfig
      --coord-kubeconfig-secret string      Secret ([namespace/]name, key "kubeconfig") on the current cluster holding the coordination-plane kubeconfig (default "dc-failover/coord-kubeconfig")
      --coord-namespace string              Namespace on the coordination plane that holds the primary-DC Leases (default "dc-failover")
  -h, --help                                help for active-dc
      --lease string                        Read this Lease directly instead of resolving a database's scope
  -q, --quiet                               Print only the active DC name
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

