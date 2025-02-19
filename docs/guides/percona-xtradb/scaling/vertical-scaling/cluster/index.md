---
title: Vertical Scaling PerconaXtraDB Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-scaling-vertical-cluster
    name: Cluster
    parent: guides-perconaxtradb-scaling-vertical
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale PerconaXtraDB Cluster

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a PerconaXtraDB cluster database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb)
  - [Clustering](/docs/guides/percona-xtradb/clustering/galera-cluster) 
  - [PerconaXtraDBOpsRequest](/docs/guides/percona-xtradb/concepts/opsrequest)
  - [Vertical Scaling Overview](/docs/guides/percona-xtradb/scaling/vertical-scaling/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Apply Vertical Scaling on Cluster

Here, we are going to deploy a  `PerconaXtraDB` cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare PerconaXtraDB Cluster

Now, we are going to deploy a `PerconaXtraDB` cluster database with version `8.0.26`.
> Vertical Scaling for `PerconaXtraDB Standalone` can be performed in the same way as `PerconaXtraDB Cluster`. Only remove the `spec.replicas` field from the below yaml to deploy a PerconaXtraDB Standalone.

### Deploy PerconaXtraDB Cluster 

In this section, we are going to deploy a PerconaXtraDB cluster database. Then, in the next section we will update the resources of the database using `PerconaXtraDBOpsRequest` CRD. Below is the YAML of the `PerconaXtraDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.26"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `PerconaXtraDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/scaling/vertical-scaling/cluster/example/sample-pxc.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Now, wait until `sample-pxc` has status `Ready`. i.e,

```bash
$ kubectl get perconaxtradb -n demo
NAME             VERSION    STATUS     AGE
sample-pxc    8.0.26     Ready     3m46s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo sample-pxc-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "500m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see the Pod has the default resources which is assigned by KubeDB operator.

We are now ready to apply the `PerconaXtraDBOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the database to meet the desired resources after scaling.

#### Create PerconaXtraDBOpsRequest

In order to update the resources of the database, we have to create a `PerconaXtraDBOpsRequest` CR with our desired resources. Below is the YAML of the `PerconaXtraDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: sample-pxc
  verticalScaling:
    perconaxtradb:
      resources:
        requests:
          memory: "1.2Gi"
          cpu: "0.6"
        limits:
          memory: "1.2Gi"
          cpu: "0.6"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `sample-pxc` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.perconaxtradb` specifies the desired resources after scaling.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/scaling/vertical-scaling/cluster/example/pxops-vscale.yaml
perconaxtradbopsrequest.ops.kubedb.com/pxops-vscale created
```

#### Verify PerconaXtraDB Cluster resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the resources of `PerconaXtraDB` object and related `PetSets` and `Pods`.

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CR,

```bash
$ kubectl get perconaxtradbopsrequest -n demo
Every 2.0s: kubectl get perconaxtradbopsrequest -n demo
NAME                     TYPE              STATUS       AGE
pxops-vscale        VerticalScaling      Successful    3m56s
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo sample-pxc-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "1288490188800m"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1288490188800m"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the PerconaXtraDB database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete perconaxtradb -n demo sample-pxc
$ kubectl delete perconaxtradbopsrequest -n demo pxops-vscale
```