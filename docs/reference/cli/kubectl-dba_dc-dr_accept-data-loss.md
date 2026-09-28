---
title: Kubectl-Dba Dc-Dr Accept-Data-Loss
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-dc-dr-accept-data-loss
    name: Kubectl-Dba Dc-Dr Accept-Data-Loss
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba dc-dr accept-data-loss

Release a failover held by the RPO budget, explicitly accepting the data loss

### Synopsis

When the surviving data center lags more than spec.replication.bestEffortCrossDCLagBytesForFailover (or its lag cannot be measured), the promotion is HELD: the un-replicated WAL of the lost data center is unrecoverable, so choosing between an outage and a loss larger than the budget belongs to a human. This command records that decision by setting dr.kubedb.com/accept-failover-data-loss=true; both promotion paths (the hub gate and the coordinator's data-plane gate) honor it within seconds, and the operator removes the annotation automatically once the failover it authorized lands, so it cannot linger and approve a later, unrelated loss.

 Where the hold is visible before you decide: status.disasterRecovery.protectionMessage (the measured lag), condition DCDRPromotionStalled, and dc-dr status.

 KUBECONFIG: the hub cluster.

```
kubectl-dba dc-dr accept-data-loss DB_NAME --yes [flags]
```

### Examples

```
  # See what would be lost first
  kubectl dba dc-dr status pg-dcdr -n demo
  
  # Accept it
  kubectl dba dc-dr accept-data-loss pg-dcdr -n demo --yes
```

### Options

```
  -h, --help   help for accept-data-loss
      --yes    Confirm accepting data loss beyond the configured RPO budget
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

