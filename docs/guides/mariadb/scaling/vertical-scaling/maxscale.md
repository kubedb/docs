---
title: Vertical Scaling MaxScale Server
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-scaling-vertical-maxscale
    name: MaxScale
    parent: guides-mariadb-scaling-vertical
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale MaxScale Server

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a MaxScale server.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
    - [MariaDB Replication](/docs/guides/mariadb/clustering/mariadb-replication)
    - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)
    - [Vertical Scaling Overview](/docs/guides/mariadb/scaling/vertical-scaling/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Apply Vertical Scaling on MaxScale Server

Here, we are going to deploy a  `MariaDB` cluster in replication mode using a supported version by `KubeDB` operator. Then we will apply vertical scaling on `MaxScale` server.

### Deploy MariaDB Cluster

In this section, we are going to deploy a MariaDB cluster database in replication mode. Then, in the next section we will update the resources of the `MaxScale` server using `MariaDBOpsRequest` CRD. Below is the YAML of the `MariaDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: md-replication
  namespace: demo
spec:
  version: "10.5.23"
  replicas: 3
  topology:
    mode: MariaDBReplication
    maxscale:
      replicas: 3
      enableUI: true
      storageType: Durable
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 50Mi
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MariaDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/scaling/md-replication.yaml
mariadb.kubedb.com/md-replication created
```

Now, wait until `md-replication` has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo
NAME             VERSION   STATUS   AGE
md-replication   10.5.23   Ready    2m39s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo md-replication-mx-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "512Mi"
  },
  "requests": {
    "cpu": "200m",
    "memory": "256Mi"
  }
}
```

You can see the Pod has the default resources which is assigned by KubeDB operator.

We are now ready to apply the `MariaDBOpsRequest` CR to update the resources of this database.

### Vertical Scaling

Here, we are going to update the resources of `MaxScale` server to meet the desired resources after scaling.

#### Create MariaDBOpsRequest

In order to update the resources of the database, we have to create a `MariaDBOpsRequest` CR with our desired resources. Below is the YAML of the `MariaDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: maxscale-vertical-scale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: md-replication
  verticalScaling:
    maxscale:
      resources:
        requests:
          memory: "512Mi"
          cpu: "0.3"
        limits:
          memory: "1Gi"
          cpu: "0.6"
```

Here,
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `sample-mariadb` database.
- `spec.VerticalScaling.maxscale` specifies the desired resources of maxscale server after scaling.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/scaling/vertical-scaling/mx-vscale.yaml
mariadbopsrequest.ops.kubedb.com/maxscale-vertical-scale created
```

#### Verify MaxScale server resources updated successfully

If everything goes well, `KubeDB` Enterprise operator will update the resources of `MariaDB` object and related `PetSets` and `Pods`.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ watch kubectl get mariadbopsrequest -n demo
Every 2.0s: kubectl get mariadbopsrequest -n demo 

NAME                      TYPE              STATUS       AGE
maxscale-vertical-scale   VerticalScaling   Successful   3m8s

```

We can see from the above output that the `MariaDBOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of maxscale server has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo md-replication-mx-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "600m",
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "300m",
    "memory": "512Mi"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the MariaDB database.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo md-replication
$ kubectl delete mariadbopsrequest -n demo maxscale-vertical-scale
$ kubectl delete ns demo
```