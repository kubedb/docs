---
title: Vertical Scaling MariaDB Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-scaling-vertical-cluster
    name: Cluster
    parent: guides-mariadb-scaling-vertical
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Vertical Scale MariaDB Cluster

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a MariaDB cluster database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
  - [Clustering](/docs/guides/mariadb/clustering/galera-cluster) 
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)
  - [Vertical Scaling Overview](/docs/guides/mariadb/scaling/vertical-scaling/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Apply Vertical Scaling on Cluster

Here, we are going to deploy a  `MariaDB` cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare MariaDB Cluster

Now, we are going to deploy a `MariaDB` cluster database with version `10.5.8`.
> Vertical Scaling for `MariaDB Standalone` can be performed in the same way as `MariaDB Cluster`. Only remove the `spec.replicas` field from the below yaml to deploy a MariaDB Standalone.

### Deploy MariaDB Cluster 

In this section, we are going to deploy a MariaDB cluster database. Then, in the next section we will update the resources of the database using `MariaDBOpsRequest` CRD. Below is the YAML of the `MariaDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.8"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's create the `MariaDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/scaling/vertical-scaling/cluster/example/sample-pxc.yaml
mariadb.kubedb.com/sample-mariadb created
```

Now, wait until `sample-mariadb` has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo
NAME             VERSION    STATUS     AGE
sample-mariadb    10.5.8     Ready     3m46s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo sample-mariadb-0 -o json | jq '.spec.containers[].resources'
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

You can see the Pod has the default resources which is assigned by Kubedb operator.

We are now ready to apply the `MariaDBOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of the database to meet the desired resources after scaling.

#### Create MariaDBOpsRequest

In order to update the resources of the database, we have to create a `MariaDBOpsRequest` CR with our desired resources. Below is the YAML of the `MariaDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: sample-mariadb
  verticalScaling:
    mariadb:
      requests:
        memory: "1.2Gi"
        cpu: "0.6"
      limits:
        memory: "1.2Gi"
        cpu: "0.6"
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `sample-mariadb` database.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.mariadb` specifies the desired resources after scaling.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/scaling/vertical-scaling/cluster/example/mdops-vscale.yaml
mariadbopsrequest.ops.kubedb.com/mdops-vscale created
```

#### Verify MariaDB Cluster resources updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the resources of `MariaDB` object and related `StatefulSets` and `Pods`.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ kubectl get mariadbopsrequest -n demo
Every 2.0s: kubectl get mariadbopsrequest -n demo
NAME                     TYPE              STATUS       AGE
mdops-vscale        VerticalScaling      Successful    3m56s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo sample-mariadb-0 -o json | jq '.spec.containers[].resources'
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

The above output verifies that we have successfully scaled up the resources of the MariaDB database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo sample-mariadb
$ kubectl delete mariadbopsrequest -n demo mdops-vscale
```