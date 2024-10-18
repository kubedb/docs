---
title: Restart Druid
menu:
  docs_{{ .version }}:
    identifier: guides-druid-restart-guide
    name: Restart Druid
    parent: guides-druid-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Druid

KubeDB supports restarting the Druid database via a DruidOpsRequest. Restarting is useful if some pods are got stuck in some phase, or they are not working correctly. This tutorial will show you how to use that.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/druid](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/druid) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Druid

In this section, we are going to deploy a Druid database using KubeDB.

### Create External Dependency (Deep Storage)

Before proceeding further, we need to prepare deep storage, which is one of the external dependency of Druid and used for storing the segments. It is a storage mechanism that Apache Druid does not provide. **Amazon S3**, **Google Cloud Storage**, or **Azure Blob Storage**, **S3-compatible storage** (like **Minio**), or **HDFS** are generally convenient options for deep storage.

In this tutorial, we will run a `minio-server` as deep storage in our local `kind` cluster using `minio-operator` and create a bucket named `druid` in it, which the deployed druid database will use.

```bash

$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1

$ helm upgrade --install --namespace "demo" --create-namespace druid-minio minio/tenant \
--set tenant.pools[0].servers=1 \
--set tenant.pools[0].volumesPerServer=1 \
--set tenant.pools[0].size=1Gi \
--set tenant.certificate.requestAutoCert=false \
--set tenant.buckets[0].name="druid" \
--set tenant.pools[0].name="default"

```

Now we need to create a `Secret` named `deep-storage-config`. It contains the necessary connection information using which the druid database will connect to the deep storage.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deep-storage-config
  namespace: demo
stringData:
  druid.storage.type: "s3"
  druid.storage.bucket: "druid"
  druid.storage.baseKey: "druid/segments"
  druid.s3.accessKey: "minio"
  druid.s3.secretKey: "minio123"
  druid.s3.protocol: "http"
  druid.s3.enablePathStyleAccess: "true"
  druid.s3.endpoint.signingRegion: "us-east-1"
  druid.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

Let’s create the `deep-storage-config` Secret shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/restart/yamls/deep-storage-config.yaml
secret/deep-storage-config created
```

Now, lets go ahead and create a druid database.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-cluster
  namespace: demo
spec:
  version: 28.0.1
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: Delete
```

Let's create the `Druid` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/update-version/yamls/druid-cluster.yaml
druid.kubedb.com/druid-cluster created
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: druid-cluster
  timeout: 5m
  apply: Always
```

- `spec.type` specifies the Type of the ops Request
- `spec.databaseRef` holds the name of the Druid CR. It should be available in the same namespace as the opsRequest
- The meaning of `spec.timeout` & `spec.apply` fields will be found [here](/docs/guides/druid/concepts/druidopsrequest.md#spectimeout)

Let's create the `DruidOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/druid/restart/restart.yaml
druidopsrequest.ops.kubedb.com/restart created
```

Now the Ops-manager operator will first restart the controller pods, then broker of the referenced druid.

```shell
$ kubectl get kfops -n demo
NAME      TYPE      STATUS       AGE
restart   Restart   Successful   119s

$ kubectl get kfops -n demo restart -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"DruidOpsRequest","metadata":{"annotations":{},"name":"restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"druid-cluster"},"timeout":"3m","type":"Restart"}}
  creationTimestamp: "2024-07-26T10:12:10Z"
  generation: 1
  name: restart
  namespace: demo
  resourceVersion: "24434"
  uid: 956a374e-1d6f-4f68-828f-cfed4410b175
spec:
  apply: Always
  databaseRef:
    name: druid-cluster
  timeout: 3m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2024-07-26T10:12:10Z"
    message: Druid ops-request has started to restart druid nodes
    observedGeneration: 1
    reason: Restart
    status: "True"
    type: Restart
  - lastTransitionTime: "2024-07-26T10:12:18Z"
    message: get pod; ConditionStatus:True; PodName:druid-cluster-controller-0
    observedGeneration: 1
    status: "True"
    type: GetPod--druid-cluster-controller-0
  - lastTransitionTime: "2024-07-26T10:12:18Z"
    message: evict pod; ConditionStatus:True; PodName:druid-cluster-controller-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--druid-cluster-controller-0
  - lastTransitionTime: "2024-07-26T10:12:23Z"
    message: check pod running; ConditionStatus:True; PodName:druid-cluster-controller-0
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--druid-cluster-controller-0
  - lastTransitionTime: "2024-07-26T10:12:28Z"
    message: get pod; ConditionStatus:True; PodName:druid-cluster-controller-1
    observedGeneration: 1
    status: "True"
    type: GetPod--druid-cluster-controller-1
  - lastTransitionTime: "2024-07-26T10:12:28Z"
    message: evict pod; ConditionStatus:True; PodName:druid-cluster-controller-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--druid-cluster-controller-1
  - lastTransitionTime: "2024-07-26T10:12:38Z"
    message: check pod running; ConditionStatus:True; PodName:druid-cluster-controller-1
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--druid-cluster-controller-1
  - lastTransitionTime: "2024-07-26T10:12:43Z"
    message: get pod; ConditionStatus:True; PodName:druid-cluster-broker-0
    observedGeneration: 1
    status: "True"
    type: GetPod--druid-cluster-broker-0
  - lastTransitionTime: "2024-07-26T10:12:43Z"
    message: evict pod; ConditionStatus:True; PodName:druid-cluster-broker-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--druid-cluster-broker-0
  - lastTransitionTime: "2024-07-26T10:13:18Z"
    message: check pod running; ConditionStatus:True; PodName:druid-cluster-broker-0
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--druid-cluster-broker-0
  - lastTransitionTime: "2024-07-26T10:13:23Z"
    message: get pod; ConditionStatus:True; PodName:druid-cluster-broker-1
    observedGeneration: 1
    status: "True"
    type: GetPod--druid-cluster-broker-1
  - lastTransitionTime: "2024-07-26T10:13:23Z"
    message: evict pod; ConditionStatus:True; PodName:druid-cluster-broker-1
    observedGeneration: 1
    status: "True"
    type: EvictPod--druid-cluster-broker-1
  - lastTransitionTime: "2024-07-26T10:13:28Z"
    message: check pod running; ConditionStatus:True; PodName:druid-cluster-broker-1
    observedGeneration: 1
    status: "True"
    type: CheckPodRunning--druid-cluster-broker-1
  - lastTransitionTime: "2024-07-26T10:13:33Z"
    message: Successfully Restarted Druid nodes
    observedGeneration: 1
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - lastTransitionTime: "2024-07-26T10:13:33Z"
    message: Controller has successfully restart the Druid replicas
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
kubectl delete druidopsrequest -n demo restart
kubectl delete druid -n demo druid-cluster
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Different Druid topology clustering modes [here](/docs/guides/druid/clustering/_index.md).
- Monitor your Druid database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Druid database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/druid/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
