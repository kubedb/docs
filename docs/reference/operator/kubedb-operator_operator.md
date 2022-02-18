---
title: Kubedb-Operator Operator
menu:
  docs_{{ .version }}:
    identifier: kubedb-operator-operator
    name: Kubedb-Operator Operator
    parent: reference-operator
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubedb-operator operator

Launch KubeDB Provisioner

### Synopsis

Launch KubeDB Provisioner

```
kubedb-operator operator [flags]
```

### Options

```
      --burst int                           The maximum burst for throttle (default 1000000)
      --health-probe-bind-address string    The address the probe endpoint binds to. (default ":8081")
  -h, --help                                help for operator
      --kubeconfig string                   Path to kubeconfig file with authorization information (the master location is set by the master flag).
      --leader-elect                        Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.
      --license-file string                 Path to license file
      --master string                       The address of the Kubernetes API server (overrides any value in kubeconfig)
      --metrics-bind-address string         The address the metric endpoint binds to. (default ":8080")
      --qps float                           The maximum QPS to the master from this client (default 1e+06)
      --readiness-probe-interval duration   The time between two consecutive health checks that the operator performs to the database. (default 10s)
      --resync-period duration              If non-zero, will re-list this often. Otherwise, re-list will be delayed aslong as possible (until the upstream source closes the watch or times out. (default 10m0s)
```

### Options inherited from parent commands

```
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
```

### SEE ALSO

* [kubedb-operator](/docs/reference/operator/kubedb-operator.md)	 - 

