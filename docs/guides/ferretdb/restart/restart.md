---
title: Restart FerretDB
menu:
  docs_{{ .version }}:
    identifier: fr-restart-details
    name: Restart FerretDB
    parent: fr-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart FerretDB

KubeDB supports restarting the FerretDB via a FerretDBOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/ferretdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/ferretdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy FerretDB

In this section, we are going to deploy a FerretDB using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferretdb
  namespace: demo
spec:
  version: "2.0.0"
  backend:
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 500Mi
  deletionPolicy: WipeOut
```

Let's create the `FerretDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/restart/ferretdb.yaml
ferretdb.kubedb.com/ferretdb created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: restart-ferretdb
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: ferretdb
  timeout: 3m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the FerretDB.  The ferretdb should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/ferretdb/concepts/opsrequest.md#spectimeout)

Let's create the `FerretDBOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/restart/ops.yaml
ferretdbopsrequest.ops.kubedb.com/restart-ferretdb created
```

Now the Ops-manager operator will restart the pods one by one.

```shell
$ kubectl get frops -n demo
NAME               TYPE      STATUS       AGE
restart-ferretdb   Restart   Successful   2m15s

$ kubectl get frops -n demo -oyaml restart-ferretdb
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"FerretDBOpsRequest","metadata":{"annotations":{},"name":"restart-ferretdb","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"ferretdb"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2024-10-21T12:38:38Z"
  generation: 1
  name: restart-ferretdb
  namespace: demo
  resourceVersion: "367859"
  uid: 0ca77cab-d354-43a4-ba85-c31f1f6e685d
spec:
  apply: Always
  databaseRef:
    name: ferretdb
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2024-10-21T12:38:38Z"
    message: FerretDBOpsRequest has started to restart FerretDB nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2024-10-21T12:38:46Z"
    message: get pod; ConditionStatus:True; PodName:ferretdb-0
    observedGeneration: 1
    status: "True"
    type: GetPod--ferretdb-0
  - lastTransitionTime: "2024-10-21T12:38:46Z"
    message: evict pod; ConditionStatus:True; PodName:ferretdb-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--ferretdb-0
  - lastTransitionTime: "2024-10-21T12:38:51Z"
    message: check pod running; ConditionStatus:True; PodName:ferretdb-0
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--ferretdb-0
  - lastTransitionTime: "2024-10-21T12:38:56Z"
    message: Successfully restarted FerretDB nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2024-10-21T12:38:56Z"
    message: Controller has successfully restart the FerretDB replicas
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```


## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ferretdbopsrequest -n demo restart-ferretdb
kubectl delete ferretdb -n demo ferretdb
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [FerretDB object](/docs/guides/ferretdb/concepts/ferretdb.md).
- Monitor your FerretDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/ferretdb/monitoring/using-prometheus-operator.md).
- Monitor your FerretDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/ferretdb/monitoring/using-builtin-prometheus.md).
- Detail concepts of [FerretDB object](/docs/guides/ferretdb/concepts/ferretdb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
