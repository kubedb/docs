---
title: Kubectl-Dba Adopt Milvus Plan
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-adopt-milvus-plan
    name: Kubectl-Dba Adopt Milvus Plan
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba adopt milvus plan

Check a running Milvus and show what adopting it creates

### Synopsis

Read-only. Finds the workloads, effective configuration, image and client Service of the running cluster, checks everything an in-place adoption depends on (Woodpecker WAL, etcd root path, object storage, a matching MilvusVersion, the current root password), and shows the Secrets, the state ConfigMap and the staged Milvus that "run" would create.

 With --output-dir the objects are written there, unredacted. Edit them if needed and apply them; "run" then uses what exists.

```
kubectl-dba adopt milvus plan RELEASE [flags]
```

### Examples

```
  kubectl dba adopt milvus plan my-release -n milvus
  kubectl dba adopt milvus plan my-release -n milvus --output-dir ./adopt
```

### Options

```
      --auth-secret string          Existing basic-auth Secret with the current root user and password
      --config-file string          The effective milvus.yaml, when it is not mounted from a ConfigMap or Secret
      --etcd-auth-secret string     Basic-auth Secret for an etcd that uses authentication
  -h, --help                        help for plan
      --name string                 Name of the Milvus object (default: the name of the Service clients use, which KubeDB then takes over)
      --output-dir string           Write the objects to this directory (Secrets unredacted, mode 0600)
      --restricted-scc              Set spec.restrictedSCC (OpenShift)
      --role-label string           Pod label that names a workload's Milvus role (default "component")
      --root-password-file string   File holding the current root password (default: common.security.defaultRootPassword of the running config)
      --selector string             Label selector of the cluster's workloads (default app.kubernetes.io/instance=RELEASE)
      --skip-credential-check       Do not log in to the running cluster to check the root credentials
      --version string              MilvusVersion to use (default: the one matching the running image)
      --workload stringArray        Map a workload to a role, as role[/group]=Kind/name (repeat for each workload; replaces --selector)
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

