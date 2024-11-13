---
title: Restart Memcached
menu:
  docs_{{ .version }}:
    identifier: mc-restart
    name: Restart Memcached
    parent: restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Memcached

KubeDB supports restarting the Memcached database via a MemcachedOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/memcached](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/memcached) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Memcached

In this section, we are going to deploy a Memcached database using KubeDB.

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcd-quickstart
  namespace: demo
spec:
  replicas: 1
  version: "1.6.22"
  podTemplate:
    spec:
      containers:
        - name: memcached
          resources:
            limits:
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 250m
              memory: 64Mi
  deletionPolicy: WipeOut
```

Let's create the `Memcached` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/restart/memcached.yaml
memcached.kubedb.com/memcd-quickstart created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: memcd-quickstart
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the Memcached database. The database should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields can be found [here](/docs/guides/Memcached/concepts/opsrequest.md)

Let's create the `MemcachedOpsRequest` CR we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/restart/opsrequest-restart.yaml
memcachedopsrequest.ops.kubedb.com/restart created
```

Now the Ops-manager operator will first restart the pods one by one.

```shell
$ kubectl get mcops -n demo restart 
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   3m25s

$ kubectl get mcops -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"MemcachedOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"databaseRef":{"name":"memcd-quickstart"},"type":"Restart"}}
  creationTimestamp: "2024-11-13T13:13:42Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "1740073"
  uid: dbe5f074-f894-4c89-b34f-a4fad2322576
spec:
  apply: IfReady
  databaseRef:
    name: memcd-quickstart
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2024-11-13T13:13:42Z"
    message: Memcached ops request is restarting the database nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2024-11-13T13:13:52Z"
    message: Successfully restarted pods
    observedGeneration: 1
    reason: RestartPods
    status: "True"
    type: RestartPods
  - lastTransitionTime: "2024-11-13T13:13:47Z"
    message: evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--memcd-quickstart-0
  - lastTransitionTime: "2024-11-13T13:13:47Z"
    message: is pod ready; ConditionStatus:False
    observedGeneration: 1
    status: "False"
    type: IsPodReady
  - lastTransitionTime: "2024-11-13T13:13:52Z"
    message: is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
    observedGeneration: 1
    status: "True"
    type: IsPodReady--memcd-quickstart-0
  - lastTransitionTime: "2024-11-13T13:13:52Z"
    message: Successfully Restarted Database
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
kubectl delete memcachedopsrequest -n demo restart
kubectl delete memcached -n demo memcd-quickstart
kubectl delete ns demo
```

## Next Steps

- Monitor your Memcached database with KubeDB using [built-in Prometheus](/docs/guides/memcached/monitoring/using-builtin-prometheus.md).
- Monitor your MemcachedQL database with KubeDB using [Prometheus operator](/docs/guides/Memcached/monitoring/using-prometheus-operator.md).
- Detail concepts of [Memcached object](/docs/guides/Memcached/concepts/Memcached.md).
- Use [private Docker registry](/docs/guides/Memcached/private-registry/using-private-registry.md) to deploy MemcachedQL with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
