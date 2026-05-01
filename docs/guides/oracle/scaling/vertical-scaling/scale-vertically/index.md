---
title: Scale Oracle Vertically
menu:
  docs_{{ .version }}:
    identifier: oracle-scale-vertically
    name: Scale Vertically
    parent: oracle-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Oracle Cluster

This guide will show you how to use `KubeDB` Ops Manager to update the resources of a `Oracle` instance.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/oracle/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/oracle/scaling/vertical-scaling/scale-vertically/yamls](/docs/guides/oracle/scaling/vertical-scaling/scale-vertically/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Apply Vertical Scaling on Oracle Cluster

Here, we are going to deploy a `Oracle` cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

**Deploy Oracle:**

In this section, we are going to deploy a Oracle cluster. Then, in the next section, we will update the resources of the database nodes using vertical scaling. Below is the YAML of the `Oracle` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Oracle` CR we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/scaling/vertical-scaling/scale-vertically/yamls/oracle.yaml
oracle.kubedb.com/oracle-sample created
```

**Wait for the cluster to be ready:**

```bash
$ watch -n 3 kubectl get oracle -n demo oracle-sample
Every 3.0s: kubectl get oracle -n demo oracle-sample

NAME             VERSION   STATUS   AGE
oracle-sample    1.17.0    Ready    3m16s

$ watch -n 3 kubectl get petset -n demo oracle-sample
Every 3.0s: kubectl get petset -n demo oracle-sample

NAME              READY   AGE
oracle-sample     3/3     3m54s

$ watch -n 3 kubectl get pod -n demo
Every 3.0s: kubectl get pod -n demo

NAME                READY   STATUS    RESTARTS   AGE
oracle-sample-0     1/1     Running   0          4m51s
oracle-sample-1     1/1     Running   0          3m50s
oracle-sample-2     1/1     Running   0          3m46s
```

Let's check the resources of the `oracle-sample-0` pod:

```bash
$ kubectl get pod -n demo oracle-sample-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "250m",
    "memory": "512Mi"
  }
}
```

We are ready to apply the `OracleOpsRequest` CR to vertically scale the cluster.

#### Create OracleOpsRequest for Vertical Scaling

In order to update the resources of the database, we have to create a `OracleOpsRequest` CR with our desired resources. Below is the YAML of the `OracleOpsRequest` CR that we are going to create:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: oracle-sample
  verticalScaling:
    node:
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "1"
          memory: "2Gi"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling on `oracle-sample` Oracle database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.verticalScaling.node.resources` specifies the desired CPU and memory resources for the Oracle nodes.

Let's create the `OracleOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/scaling/vertical-scaling/scale-vertically/yamls/vscale.yaml
oracleopsrequest.ops.kubedb.com/qdops-vscale created
```

#### Verify Oracle vertical scaling completed successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of the `Oracle` object and related `PetSet`.

Let's wait for `OracleOpsRequest` to be `Successful`:

```bash
$ watch -n 3 kubectl get OracleOpsRequest -n demo qdops-vscale
Every 3.0s: kubectl get OracleOpsRequest -n demo qdops-vscale

NAME            TYPE              STATUS       AGE
qdops-vscale    VerticalScaling   Successful   3m12s
```

Now, let's verify that the resources of the pods have been updated:

```bash
$ kubectl get pod -n demo oracle-sample-0 -o json | jq '.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see from the above output that the resources of the `oracle-sample-0` pod have been updated successfully. All pods in the cluster will have the same updated resource configuration.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracle -n demo oracle-sample
kubectl delete OracleOpsRequest -n demo qdops-vscale
```