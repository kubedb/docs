---
title: Vertical Scaling SingleStore Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-scaling-vertical-cluster
    name: Vertical Scaling OpsRequest
    parent: guides-sdb-scaling-vertical
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale SingleStore Cluster

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a SingleStore cluster database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
  - [Clustering](/docs/guides/singlestore/clustering/singlestore-clustering/) 
  - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)
  - [Vertical Scaling Overview](/docs/guides/singlestore/scaling/vertical-scaling/overview/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Apply Vertical Scaling on Cluster

Here, we are going to deploy a  `SingleStore` cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

### Deploy SingleStore Cluster 

In this section, we are going to deploy a SingleStore cluster database. Then, in the next section we will update the resources of the database using `SingleStoreOpsRequest` CRD. Below is the YAML of the `SingleStore` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sample-sdb
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
  deletionPolicy: WipeOut
```

Let's create the `SingleStore` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/scaling/vertical-scaling/cluster/example/sample-sdb.yaml
singlestore.kubedb.com/sample-sdb created
```

Now, wait until `sample-sdb` has status `Ready`. i.e,

```bash
$ kubectl get sdb -n demo
NAME         TYPE                  VERSION   STATUS   AGE
sample-sdb   kubedb.com/v1alpha2   8.7.10    Ready    101s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo sample-sdb-aggregator-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "2Gi"
  }
}

```

We are now ready to apply the `SingleStoreOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the database to meet the desired resources after scaling.

#### Create SingleStoreOpsRequest

In order to update the resources of the database, we have to create a `SingleStoreOpsRequest` CR with our desired resources. Below is the YAML of the `SingleStoreOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-vscale
  namespace: demo
spec:
  type: VerticalScaling  
  databaseRef:
    name: sample-sdb
  verticalScaling:
    aggregator:
      resources:
        requests:
          memory: "2500Mi"
          cpu: "0.7"
        limits:
          memory: "2500Mi"
          cpu: "0.7"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `sample-sdb` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.aggregator` specifies the desired `aggregator` nodes resources after scaling. As well you can scale resources for leaf node, standalone node and coordinator container.

Let's create the `SingleStoreOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/scaling/vertical-scaling/cluster/example/sdbops-vscale.yaml
singlestoreopsrequest.ops.kubedb.com/sdbops-vscale created
```

#### Verify SingleStore Cluster resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the resources of `SingleStore` object and related `PetSets` and `Pods`.

Let's wait for `SingleStoreOpsRequest` to be `Successful`.  Run the following command to watch `SingleStoreOpsRequest` CR,

```bash
$ kubectl get singlestoreopsrequest -n demo
NAME            TYPE              STATUS       AGE
sdbops-vscale   VerticalScaling   Successful   7m30s
```

We can see from the above output that the `SingleStoreOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo sample-sdb-aggregator-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "700m",
    "memory": "2500Mi"
  },
  "requests": {
    "cpu": "700m",
    "memory": "2500Mi"
  }
}

```

The above output verifies that we have successfully scaled up the resources of the SingleStore database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete sdb -n demo sample-sdb
$ kubectl delete singlestoreopsrequest -n demo sdbops-vscale
```