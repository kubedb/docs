---
title: Kubectl-Dba Adopt Milvus Rollback
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-adopt-milvus-rollback
    name: Kubectl-Dba Adopt Milvus Rollback
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba adopt milvus rollback

Give an adopted (or partly adopted) Milvus back to the chart's workloads

### Synopsis

Deletes the Milvus object with kubedb.com/milvus-orphan-on-delete, so its PetSets keep serving, then moves each role back to the chart's workloads, starting the chart's pods before it stops KubeDB's. Once the chart's proxies are Ready, the client Service gets back its original selector, labels and ports. The data is never touched.

 With --cleanup it then deletes the PetSets (at 0 replicas), the Services other than the client Service and the rendered config KubeDB left, and recreates the chart Services the adoption removed. PVCs are kept.

 For a Milvus adopted by hand (no state ConfigMap), give its --name: the chart's replicas are then taken from the PetSets, and the client Service gets the chart proxies' selector.

```
kubectl-dba adopt milvus rollback [RELEASE] [flags]
```

### Examples

```
  kubectl dba adopt milvus rollback my-release -n milvus --yes --cleanup
  kubectl dba adopt milvus rollback -f ./adopt/adopt.yaml --yes
```

### Options

```
      --cleanup                    Delete what KubeDB left behind once the chart serves again (PVCs are kept)
  -f, --filename string            File with the adoption options (kind MilvusAdoption; "plan --output-dir" writes one). Flags given on the command line win
  -h, --help                       help for rollback
      --name string                Name of the Milvus, for an adoption done by hand (no state ConfigMap)
      --querynode-delay duration   Overlap: time a new QueryNode serves before an old one stops (raise it for large collections) (default 1m0s)
      --role-label string          Pod label that names a workload's Milvus role (default "component")
      --selector string            Label selector of the cluster's workloads (default app.kubernetes.io/instance=RELEASE)
      --settle duration            Pause after each role (default 20s)
      --timeout duration           Longest wait for one step (pods Ready or gone) (default 15m0s)
      --workload stringArray       Map a workload to a role, as role[/group]=Kind/name (repeat for each workload; replaces --selector)
      --yes                        Proceed; without it, rollback only shows what it would do
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

* [kubectl-dba adopt milvus](/docs/reference/cli/kubectl-dba_adopt_milvus.md)	 - Adopt a running Distributed Milvus 2.6 in place

