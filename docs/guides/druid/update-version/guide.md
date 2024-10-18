---
title: Update Version of Druid
menu:
  docs_{{ .version }}:
    identifier: guides-druid-update-version-guide
    name: Guide
    parent: guides-druid-update-version
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/README.md).

# Update version of Druid

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `Druid` Combined or Topology.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/druid/concepts/druid.md)
    - [DruidOpsRequest](/docs/guides/druid/concepts/druidopsrequest.md)
    - [Updating Overview](/docs/guides/druid/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/druid](/docs/examples/druid) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare Druid

Now, we are going to deploy a `Druid` cluster with version `28.0.1`.

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/update-version/yamls/deep-storage-config.yaml
secret/deep-storage-config created
```

### Deploy Druid

In this section, we are going to deploy a Druid topology cluster. Then, in the next section we will update the version using `DruidOpsRequest` CRD. Below is the YAML of the `Druid` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-quickstart
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

Now, wait until `druid-cluster` created has status `Ready`. i.e,

```bash
$ kubectl get dr -n demo -w                                                                                                                                           
NAME            TYPE                  VERSION    STATUS         AGE
druid-cluster   kubedb.com/v1aplha2   28.0.1     Provisioning   0s
druid-cluster   kubedb.com/v1aplha2   28.0.1     Provisioning   55s
.
.
druid-cluster   kubedb.com/v1aplha2   28.0.1     Ready          119s
```

We are now ready to apply the `DruidOpsRequest` CR to update.

### update Druid Version

Here, we are going to update `Druid` from `28.0.1` to `30.0.0`.

#### Create DruidOpsRequest:

In order to update the version, we have to create a `DruidOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `DruidOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: druid-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: druid-cluster
  updateVersion:
    targetVersion: 30.0.0
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `druid-cluster` Druid.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `3.6.1`.

Let's create the `DruidOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/update-version/yamls/druid-hscale-up.yaml
druidopsrequest.ops.kubedb.com/druid-update-version created
```

#### Verify Druid version updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the image of `Druid` object and related `PetSets` and `Pods`.

Let's wait for `DruidOpsRequest` to be `Successful`.  Run the following command to watch `DruidOpsRequest` CR,

```bash
$ kubectl get druidopsrequest -n demo
NAME                   TYPE            STATUS        AGE
druid-update-version   UpdateVersion   Successful    2m6s
```

We can see from the above output that the `DruidOpsRequest` has succeeded. If we describe the `DruidOpsRequest` we will get an overview of the steps that were followed to update the database version.

```bash
$ kubectl describe druidopsrequest -n demo druid-update-version
Name:         druid-update-version
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         DruidOpsRequest
Metadata:
  Creation Timestamp:  2024-07-30T10:18:44Z
  Generation:          1
  Resource Version:    90131
  UID:                 a274197b-c379-485b-9a36-9eb1e673eee4
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   druid-cluster
  Timeout:  5m
  Type:     UpdateVersion
  Update Version:
    Target Version:  3.6.1
Status:
  Conditions:
    Last Transition Time:  2024-07-30T10:18:44Z
    Message:               Druid ops-request has started to update version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2024-07-30T10:18:54Z
    Message:               successfully reconciled the Druid with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-30T10:18:59Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-controller-0
    Last Transition Time:  2024-07-30T10:18:59Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-controller-0
    Last Transition Time:  2024-07-30T10:19:19Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-controller-0
    Last Transition Time:  2024-07-30T10:19:24Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-controller-1
    Last Transition Time:  2024-07-30T10:19:24Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-controller-1
    Last Transition Time:  2024-07-30T10:19:49Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-controller-1
    Last Transition Time:  2024-07-30T10:19:54Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-broker-0
    Last Transition Time:  2024-07-30T10:19:54Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-broker-0
    Last Transition Time:  2024-07-30T10:20:14Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-broker-0
    Last Transition Time:  2024-07-30T10:20:19Z
    Message:               get pod; ConditionStatus:True; PodName:druid-cluster-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--druid-cluster-broker-1
    Last Transition Time:  2024-07-30T10:20:19Z
    Message:               evict pod; ConditionStatus:True; PodName:druid-cluster-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--druid-cluster-broker-1
    Last Transition Time:  2024-07-30T10:20:44Z
    Message:               check pod running; ConditionStatus:True; PodName:druid-cluster-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--druid-cluster-broker-1
    Last Transition Time:  2024-07-30T10:20:49Z
    Message:               Successfully Restarted Druid nodes
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-07-30T10:20:50Z
    Message:               Successfully completed update druid version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   3m7s   KubeDB Ops-manager Operator  Start processing for DruidOpsRequest: demo/druid-update-version
  Normal   Starting                                                                   3m7s   KubeDB Ops-manager Operator  Pausing Druid databse: demo/druid-cluster
  Normal   Successful                                                                 3m7s   KubeDB Ops-manager Operator  Successfully paused Druid database: demo/druid-cluster for DruidOpsRequest: druid-update-version
  Normal   UpdatePetSets                                                              2m57s  KubeDB Ops-manager Operator  successfully reconciled the Druid with updated version
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-controller-0             2m52s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-controller-0           2m52s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-controller-0  2m47s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-controller-0   2m32s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-controller-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-controller-1             2m27s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-controller-1           2m27s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-controller-1  2m22s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-controller-1   2m2s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-controller-1
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-broker-0                 117s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-broker-0               117s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-broker-0      112s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-broker-0       97s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-broker-0
  Warning  get pod; ConditionStatus:True; PodName:druid-cluster-broker-1                 92s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:druid-cluster-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:druid-cluster-broker-1               92s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:druid-cluster-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:druid-cluster-broker-1      87s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:druid-cluster-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:druid-cluster-broker-1       67s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:druid-cluster-broker-1
  Normal   RestartPods                                                                62s    KubeDB Ops-manager Operator  Successfully Restarted Druid nodes
  Normal   Starting                                                                   62s    KubeDB Ops-manager Operator  Resuming Druid database: demo/druid-cluster
  Normal   Successful                                                                 61s    KubeDB Ops-manager Operator  Successfully resumed Druid database: demo/druid-cluster for DruidOpsRequest: druid-update-version
```

Now, we are going to verify whether the `Druid` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get dr -n demo druid-cluster -o=jsonpath='{.spec.version}{"\n"}'
3.6.1

$ kubectl get petset -n demo druid-cluster-broker -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/druid-kraft:3.6.1@sha256:e251d3c0ceee0db8400b689e42587985034852a8a6c81b5973c2844e902e6d11

$ kubectl get pods -n demo druid-cluster-broker-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/druid-kraft:3.6.1@sha256:e251d3c0ceee0db8400b689e42587985034852a8a6c81b5973c2844e902e6d11
```

You can see from above, our `Druid` has been updated with the new version. So, the updateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete druidopsrequest -n demo druid-update-version
kubectl delete dr -n demo druid-cluster
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Different Druid topology clustering modes [here](/docs/guides/druid/clustering/_index.md).
- Monitor your Druid database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/druid/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Druid database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/druid/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
