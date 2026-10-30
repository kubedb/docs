---
title: Kubectl-Dba Adopt Milvus Run
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-adopt-milvus-run
    name: Kubectl-Dba Adopt Milvus Run
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba adopt milvus run

Adopt a running Milvus: create the Milvus object and hand the roles over

### Synopsis

Runs the checks of "plan", creates the Secrets, the state ConfigMap NAME-adopt (NAME is the Milvus) and the Milvus with every role at 0 replicas, and waits for the operator. Before any pod moves, it compares the config the operator rendered with the running one on every setting that decides where the data is read from, and stops on a difference.

 Then it hands the roles over one at a time:

  *  --method overlap (default): start KubeDB's pods of a role, wait until they are Ready, then stop the chart's. QueryNodes move one at a time. No role stops serving while it moves; it needs room for one extra set of pods of a role (one extra QueryNode).
  *  --method stop-start: stop the chart's pods of a role, then start KubeDB's. Each role is briefly unavailable; no extra capacity is needed.

 Progress is recorded in the state ConfigMap: if the command stops, run it again and it continues where it was. "rollback" gives the cluster back.

 Do not remove the chart with "helm uninstall" afterwards: it would delete the client Service that KubeDB now uses.

```
kubectl-dba adopt milvus run [RELEASE] [flags]
```

### Examples

```
  kubectl dba adopt milvus run my-release -n milvus --yes
  kubectl dba adopt milvus run my-release -n milvus --method stop-start --yes
  kubectl dba adopt milvus run -f ./adopt/adopt.yaml --yes
```

### Options

```
      --auth-secret string          Existing basic-auth Secret with the current root user and password
      --config-file string          The effective milvus.yaml, when it is not mounted from a ConfigMap or Secret
      --etcd-auth-secret string     Basic-auth Secret for an etcd that uses authentication
  -f, --filename string             File with the adoption options (kind MilvusAdoption; "plan --output-dir" writes one). Flags given on the command line win
  -h, --help                        help for run
      --method string               Handover method: overlap or stop-start (default "overlap")
      --name string                 Name of the Milvus object (default: the name of the Service clients use, which KubeDB then takes over)
      --querynode-delay duration    Overlap: time a new QueryNode serves before an old one stops (raise it for large collections) (default 1m0s)
      --ready-timeout duration      Longest wait for the operator, and for the Milvus to be Ready at the end (default 15m0s)
      --restricted-scc              Set spec.restrictedSCC (OpenShift)
      --role-label string           Pod label that names a workload's Milvus role (default "component")
      --root-password-file string   File holding the current root password (default: common.security.defaultRootPassword of the running config)
      --selector string             Label selector of the cluster's workloads (default app.kubernetes.io/instance=RELEASE)
      --settle duration             Pause after each role (default 20s)
      --skip-config-check           Do not stop when the rendered config differs from the running one
      --skip-credential-check       Do not log in to the running cluster to check the root credentials
      --timeout duration            Longest wait for one step (pods Ready or gone) (default 15m0s)
      --version string              MilvusVersion to use (default: the one matching the running image)
      --workload stringArray        Map a workload to a role, as role[/group]=Kind/name (repeat for each workload; replaces --selector)
      --yes                         Proceed; without it, run only shows what it would do
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

