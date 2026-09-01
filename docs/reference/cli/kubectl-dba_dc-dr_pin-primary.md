---
title: Kubectl-Dba Dc-Dr Pin-Primary
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-dc-dr-pin-primary
    name: Kubectl-Dba Dc-Dr Pin-Primary
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba dc-dr pin-primary

Pin this data center as primary (break-glass override): no failover, writable through a control-plane outage

### Synopsis

Creates the human-owned break-glass override ConfigMap<scope> -override in the coordination namespace of the CURRENT cluster, which must be the data center you are pinning (its own spoke).

 Two effects, both live-proven: this DC's agent mirrors the pin onto the Lease so every other Member defers permanently, and this DC's coordinator forces its leader ACTIVE regardless of marker state, so a sustained coordination-plane outage no longer fences it read-only after the usual marker TTL plus uncertainty hold.

 Use it when the failover authority is unreachable and the surviving primary must keep accepting writes, or as a deliberate "never fail this scope over" policy. While it stands there is no split-brain protection for the scope, and nothing takes over if this DC dies.

 KUBECONFIG: the SPOKE of the data center being pinned.

```
kubectl-dba dc-dr pin-primary (--scope LEASE | --db DB_NAME) [flags]
```

### Examples

```
  # Keep dc-b primary for the global scope, come what may (run against dc-b)
  kubectl dba dc-dr pin-primary --scope primary-dc --yes --kubeconfig ~/.kube/dc-b.yaml
  
  # Clear it once the emergency is over
  kubectl dba dc-dr pin-primary --scope primary-dc --remove --yes --kubeconfig ~/.kube/dc-b.yaml
```

### Options

```
      --coord-kubeconfig string             Path to a kubeconfig file for the coordination control plane (overrides the secret/configmap sources)
      --coord-kubeconfig-configmap string   ConfigMap ([namespace/]name, key "kubeconfig") on the current cluster holding the coordination-plane kubeconfig
      --coord-kubeconfig-secret string      Secret ([namespace/]name, key "kubeconfig") on the current cluster holding the coordination-plane kubeconfig (default "dc-failover/coord-kubeconfig")
      --coord-namespace string              Namespace on the coordination plane that holds the primary-DC Leases (default "dc-failover")
      --db string                           Resolve the scope from this database instead (requires the kubeconfig to reach the hub)
      --force                               With --remove: also clear the override-hold annotation directly on the Lease, for when the pinned DC is dead and its agent cannot clear it (run against any live cluster; needs the --coord-* flags to reach the coordination plane)
  -h, --help                                help for pin-primary
      --remove                              Remove the pin instead of creating it
      --scope string                        Primary-DC Lease name of the scope (for example primary-dc or primary-dc-orders)
      --yes                                 Confirm
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

