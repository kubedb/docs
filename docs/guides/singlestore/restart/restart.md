---
title: Restart SingleStore
menu:
  docs_{{ .version }}:
    identifier: sdb-restart-details
    name: Restart SingleStore
    parent: sdb-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart SingleStore

KubeDB supports restarting the SingleStore database via a SingleStoreOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/singlestore/restart/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/singlestore/restart/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

## Deploy SingleStore

In this section, we are going to deploy a SingleStore database using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-sample
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 1
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"
      storage:
        storageClassName: "longhorn"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"            
      storage:
        storageClassName: "longhorn"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    kind: Secret
    name: license-secret
  storageType: Durable
```

Let's create the `SingleStore` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/restart/yamls/sdb-sample.yaml
singlestore.kubedb.com/sdb-sample created
```
**Wait for the database to be ready:**

Now, wait for `SingleStore` going on `Ready` state

```bash
kubectl get singlestore -n demo
NAME         TYPE                  VERSION   STATUS   AGE
sdb-sample   kubedb.com/v1alpha2   8.7.10    Ready    2m

```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: sdb-sample
  timeout: 10m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the SingleStore database. The db should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/rabbitmq/concepts/opsrequest.md#spectimeout)

> Note: The method of restarting the standalone & clustered singlestore is exactly same as above. All you need, is to specify the corresponding SingleStore name in `spec.databaseRef.name` section.

Let's create the `SingleStoreOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/restart/yamls/restart-ops.yaml
singlestoreopsrequest.ops.kubedb.com/restart created
```

Now the Ops-manager operator will restart the pods sequentially by their cardinal suffix.

```shell
$ kubectl get singlestoreopsrequest -n demo
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   10m

$ kubectl get singlestoreopsrequest -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"SinglestoreOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"sdb-sample"},"timeout":"10m","type":"Restart"}}
  creationTimestamp: "2024-10-28T05:31:00Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "3549386"
  uid: b2512e44-89eb-4f1b-ae0d-232caee94f01
spec:
  apply: Always
  databaseRef:
    name: sdb-sample
  timeout: 10m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2024-10-28T05:31:00Z"
    message: Singlestore ops-request has started to restart singlestore nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2024-10-28T05:31:03Z"
    message: Successfully paused database
    observedGeneration: 1
    reason: DatabasePauseSucceeded
    status: "True"
    type: DatabasePauseSucceeded
  - lastTransitionTime: "2024-10-28T05:33:33Z"
    message: Successfully restarted Singlestore nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2024-10-28T05:31:08Z"
    message: get pod; ConditionStatus:True; PodName:sdb-sample-aggregator-0
    observedGeneration: 1
    status: "True"
    type: GetPod--sdb-sample-aggregator-0
  - lastTransitionTime: "2024-10-28T05:31:08Z"
    message: evict pod; ConditionStatus:True; PodName:sdb-sample-aggregator-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--sdb-sample-aggregator-0
  - lastTransitionTime: "2024-10-28T05:31:48Z"
    message: check pod ready; ConditionStatus:True; PodName:sdb-sample-aggregator-0
    observedGeneration: 1
    status: "True"
    type: CheckPodReady--sdb-sample-aggregator-0
  - lastTransitionTime: "2024-10-28T05:31:53Z"
    message: get pod; ConditionStatus:True; PodName:sdb-sample-leaf-0
    observedGeneration: 1
    status: "True"
    type: GetPod--sdb-sample-leaf-0
  - lastTransitionTime: "2024-10-28T05:31:53Z"
    message: evict pod; ConditionStatus:True; PodName:sdb-sample-leaf-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--sdb-sample-leaf-0
  - lastTransitionTime: "2024-10-28T05:32:38Z"
    message: check pod ready; ConditionStatus:True; PodName:sdb-sample-leaf-0
    observedGeneration: 1
    status: "True"
    type: CheckPodReady--sdb-sample-leaf-0
  - lastTransitionTime: "2024-10-28T05:32:43Z"
    message: get pod; ConditionStatus:True; PodName:sdb-sample-leaf-1
    observedGeneration: 1
    status: "True"
    type: GetPod--sdb-sample-leaf-1
  - lastTransitionTime: "2024-10-28T05:32:43Z"
    message: evict pod; ConditionStatus:True; PodName:sdb-sample-leaf-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--sdb-sample-leaf-1
  - lastTransitionTime: "2024-10-28T05:33:28Z"
    message: check pod ready; ConditionStatus:True; PodName:sdb-sample-leaf-1
    observedGeneration: 1
    status: "True"
    type: CheckPodReady--sdb-sample-leaf-1
  - lastTransitionTime: "2024-10-28T05:33:33Z"
    message: Controller has successfully restart the Singlestore replicas
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```


## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete singlestoreopsrequest -n demo restart
kubectl delete singlestore -n demo sdb-sample
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [SingleStore object](/docs/guides/singlestore/concepts/singlestore.md).
- Monitor your SingleStore database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/singlestore/monitoring/prometheus-operator/index.md).
- Monitor your SingleStore database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/singlestore/monitoring/builtin-prometheus/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
