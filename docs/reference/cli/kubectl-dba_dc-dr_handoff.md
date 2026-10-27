---
title: Kubectl-Dba Dc-Dr Handoff
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-dc-dr-handoff
    name: Kubectl-Dba Dc-Dr Handoff
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba dc-dr handoff

Move the failover authority for a scope by handing off its primary-DC Lease

### Synopsis

Writes dr.open-cluster-management.io/handoff-to on the scope's primary-DC Lease. The holding data center's agent releases the Lease once, the target acquires it within a retry tick, and the annotation clears itself.

 This is the scope-local FAILOVER lever, and the correct tool when the active data center's database is down but its data center is alive: no quiesce and no catch-up wait happen, so loss is bounded by the RPO budget rather than zero. For a healthy primary prefer "dc-dr switchover", which is zero-RPO.

 It moves EVERY database sharing the scope. Do NOT stop a DC's agent to force a failover instead: one agent serves every scope its DC holds, so that expires all of them together.

 KUBECONFIG: the hub cluster (to resolve a database's scope and read the coordination kubeconfig Secret). The Lease is written on the coordination plane via the --coord-* flags.

```
kubectl-dba dc-dr handoff (DB_NAME | --lease NAME) --to DC [flags]
```

### Examples

```
  # Fail a database's scope over to dc-b
  kubectl dba dc-dr handoff pg-dcdr -n demo --to dc-b --yes
  
  # Move a scope by Lease name (works with no database left)
  kubectl dba dc-dr handoff --lease primary-dc-orders --to dc-a --yes
```

### Options

```
      --coord-kubeconfig string             Path to a kubeconfig file for the coordination control plane (overrides the secret/configmap sources)
      --coord-kubeconfig-configmap string   ConfigMap ([namespace/]name, key "kubeconfig") on the current cluster holding the coordination-plane kubeconfig
      --coord-kubeconfig-secret string      Secret ([namespace/]name, key "kubeconfig") on the current cluster holding the coordination-plane kubeconfig (default "dc-failover/coord-kubeconfig")
      --coord-namespace string              Namespace on the coordination plane that holds the primary-DC Leases (default "dc-failover")
  -h, --help                                help for handoff
      --lease string                        Act on this Lease directly instead of resolving a database's scope
      --to string                           Target data center
      --yes                                 Confirm the handoff (it moves every database in the scope)
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

