---
title: Weaviate Storage Autoscaler Cluster
menu:
  docs_{{ .version }}:
    identifier: weaviate-autoscaler-storage-cluster
    name: Cluster
    parent: weaviate-autoscaler-storage
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Weaviate Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a Weaviate cluster database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community, Enterprise and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation).

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack).

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [WeaviateOpsRequest](/docs/guides/weaviate/concepts/opsrequest.md)
  - [Storage Autoscaling Overview](/docs/guides/weaviate/autoscaler/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Storage Autoscaling of Cluster Database

At first, verify that your cluster has a storage class that supports volume expansion:

```bash
$ kubectl get storageclass
NAME                  PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)    rancher.io/local-path   Delete          WaitForFirstConsumer   false                  79m
topolvm-provisioner   topolvm.cybozu.com      Delete          WaitForFirstConsumer   true                   78m
```

We can see from the output that `topolvm-provisioner` storage class has `ALLOWVOLUMEEXPANSION` set to `true`. We will use it for this tutorial. You can install topolvm from [here](https://github.com/topolvm/topolvm).

Now, we are going to deploy a `Weaviate` cluster using a supported version by `KubeDB` operator. Then we are going to apply `WeaviateAutoscaler` to set up autoscaling.

### Deploy Weaviate Cluster

In this section, we are going to deploy a Weaviate cluster database with version `1.26.4`. Then, in the next section we will set up autoscaling for this database using `WeaviateAutoscaler` CRD. Below is the YAML of the `Weaviate` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: "1.26.4"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "topolvm-provisioner"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Weaviate` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/autoscaler/storage/weaviate-cluster.yaml
weaviate.kubedb.com/weaviate-sample created
```

Now, wait until `weaviate-sample` has status `Ready`:

```bash
$ kubectl get weaviate -n demo
NAME              VERSION   STATUS   AGE
weaviate-sample   1.26.4    Ready    3m46s
```

Let's check the volume size from the StatefulSet and from the persistent volumes:

```bash
$ kubectl get sts -n demo weaviate-sample -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                               STORAGECLASS          REASON   AGE
pvc-43266d76-f280-4cca-bd78-d13660a84db9   1Gi        RWO            Delete           Bound    demo/data-weaviate-sample-2         topolvm-provisioner            57s
pvc-4a509b05-774b-42d9-b36d-599c9056af37   1Gi        RWO            Delete           Bound    demo/data-weaviate-sample-0         topolvm-provisioner            58s
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   1Gi        RWO            Delete           Bound    demo/data-weaviate-sample-1         topolvm-provisioner            57s
```

You can see the StatefulSet has 1GB storage and the capacity of all the persistent volumes is also 1GB.

We are now ready to apply the `WeaviateAutoscaler` CRD to set up storage autoscaling for this database.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a `WeaviateAutoscaler` Object.

#### Create WeaviateAutoscaler Object

In order to set up storage autoscaling for this cluster database, we have to create a `WeaviateAutoscaler` CR with our desired configuration. Below is the YAML of the `WeaviateAutoscaler` object that we are going to create:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: WeaviateAutoscaler
metadata:
  name: weaviate-as-storage
  namespace: demo
spec:
  databaseRef:
    name: weaviate-sample
  storage:
    node:
      trigger: "On"
      usageThreshold: 20
      scalingThreshold: 20
      expansionMode: "Online"
```

Here,

- `spec.databaseRef.name` specifies that we are performing storage autoscaling on `weaviate-sample` database.
- `spec.storage.node.trigger` specifies that storage autoscaling is enabled for the Weaviate nodes.
- `spec.storage.node.usageThreshold` specifies the storage usage threshold - if storage usage exceeds `20%`, storage autoscaling will be triggered.
- `spec.storage.node.scalingThreshold` specifies the scaling threshold - storage will be scaled to `20%` of the current amount.
- `spec.storage.node.expansionMode` specifies the expansion mode of the volume expansion `WeaviateOpsRequest` created by `WeaviateAutoscaler`. topolvm-provisioner supports online volume expansion so `expansionMode` is set to `Online`.

Let's create the `WeaviateAutoscaler` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/autoscaler/storage/weaviate-as-storage.yaml
weaviateautoscaler.autoscaling.kubedb.com/weaviate-as-storage created
```

#### Verify Autoscaler is set up successfully

Let's check that the `WeaviateAutoscaler` resource is created successfully:

```bash
$ kubectl get weaviateautoscaler -n demo
NAME                  AGE
weaviate-as-storage   33s

$ kubectl describe weaviateautoscaler weaviate-as-storage -n demo
Name:         weaviate-as-storage
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         WeaviateAutoscaler
Spec:
  Database Ref:
    Name:  weaviate-sample
  Storage:
    Node:
      Expansion Mode:   Online
      Scaling Threshold:  20
      Trigger:            On
      Usage Threshold:    20
Events:                   <none>
```

So, the `WeaviateAutoscaler` resource is created successfully. The operator will now continuously watch the storage usage of the Weaviate pods. When the usage crosses the `usageThreshold`, it will create a `WeaviateOpsRequest` to expand the storage.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using the `dd` command to see if storage autoscaling is working:

```bash
$ kubectl exec -it -n demo weaviate-sample-0 -- bash
root@weaviate-sample-0:/var/lib/weaviate# df -h /var/lib/weaviate
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/57cd4330-784f-42c1-bf8e-e743241df164 1014M   32M  983M   4% /var/lib/weaviate
root@weaviate-sample-0:/var/lib/weaviate# dd if=/dev/zero of=/var/lib/weaviate/file.img bs=800M count=1
1+0 records in
1+0 records out
838860800 bytes (839 MB, 800 MiB) copied, 6.47 s, 130 MB/s
root@weaviate-sample-0:/var/lib/weaviate# df -h /var/lib/weaviate
Filesystem                                         Size  Used Avail Use% Mounted on
/dev/topolvm/57cd4330-784f-42c1-bf8e-e743241df164 1014M  832M  183M  82% /var/lib/weaviate
```

Now let's watch the `WeaviateOpsRequest` in the demo namespace:

```bash
$ kubectl get weaviateopsrequest -n demo -w
NAME                                     TYPE              STATUS        AGE
wvops-weaviate-sample-xxxxxxxx           VolumeExpansion   Progressing   10s
wvops-weaviate-sample-xxxxxxxx           VolumeExpansion   Successful    2m
```

After the `WeaviateOpsRequest` completes successfully, let's check the updated storage:

```bash
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                               STORAGECLASS          REASON   AGE
pvc-43266d76-f280-4cca-bd78-d13660a84db9   1217Mi     RWO            Delete           Bound    demo/data-weaviate-sample-2         topolvm-provisioner            15m
pvc-4a509b05-774b-42d9-b36d-599c9056af37   1217Mi     RWO            Delete           Bound    demo/data-weaviate-sample-0         topolvm-provisioner            15m
pvc-c27eee12-cd86-4410-b39e-b1dd735fc14d   1217Mi     RWO            Delete           Bound    demo/data-weaviate-sample-1         topolvm-provisioner            15m
```

The storage has been automatically scaled from 1Gi to ~1.2Gi (120% of 1Gi) as we specified a `scalingThreshold` of 20%.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviate -n demo weaviate-sample
kubectl delete weaviateautoscaler -n demo weaviate-as-storage
kubectl delete ns demo
```
