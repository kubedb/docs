---
title: Updating MariaDB Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-updating-cluster
    name: Cluster
    parent: guides-mariadb-updating
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# update version of MariaDB Cluster

This guide will show you how to use `KubeDB` Enterprise operator to update the version of `MariaDB` Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
  - [Cluster](/docs/guides/mariadb/clustering/overview)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)
  - [Updating Overview](/docs/guides/mariadb/update-version/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare MariaDB Cluster

Now, we are going to deploy a `MariaDB` cluster database with version `10.4.32`.

### Deploy MariaDB cluster

In this section, we are going to deploy a MariaDB Cluster. Then, in the next section we will update the version of the database using `MariaDBOpsRequest` CRD. Below is the YAML of the `MariaDB` CR that we are going to create,

> If you want to update `MariaDB Standalone`, Just remove the `spec.Replicas` from the below yaml and rest of the steps are same.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.4.32"
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

Let's create the `MariaDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/update-version/cluster/examples/sample-mariadb.yaml
mariadb.kubedb.com/sample-mariadb created
```

Now, wait until `sample-mariadb` created has status `Ready`. i.e,

```bash
$ kubectl get mariadb -n demo                                                                                                                                             
NAME             VERSION    STATUS     AGE
sample-mariadb   10.4.32    Ready     3m15s
```

We are now ready to apply the `MariaDBOpsRequest` CR to update this database.

### update MariaDB Version

Here, we are going to update `MariaDB` cluster from `10.4.32` to `10.5.23`.

#### Create MariaDBOpsRequest:

In order to update the database cluster, we have to create a `MariaDBOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `MariaDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdops-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: sample-mariadb
  updateVersion:
    targetVersion: "10.5.23"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `sample-mariadb` MariaDB database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `10.5.23`.

Let's create the `MariaDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/update-version/cluster/examples/mdops-update.yaml
mariadbopsrequest.ops.kubedb.com/mdops-update created
```

#### Verify MariaDB version updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the image of `MariaDB` object and related `PetSets` and `Pods`.

Let's wait for `MariaDBOpsRequest` to be `Successful`.  Run the following command to watch `MariaDBOpsRequest` CR,

```bash
$ kubectl get mariadbopsrequest -n demo
Every 2.0s: kubectl get mariadbopsrequest -n demo
NAME              TYPE            STATUS       AGE
mdops-update      UpdateVersion   Successful    84s
```

We can see from the above output that the `MariaDBOpsRequest` has succeeded.

Now, we are going to verify whether the `MariaDB` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get mariadb -n demo sample-mariadb -o=jsonpath='{.spec.version}{"\n"}'
10.5.23

$ kubectl get sts -n demo sample-mariadb -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
mariadb:10.5.23

$ kubectl get pods -n demo sample-mariadb-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
mariadb:10.5.23
```

You can see from above, our `MariaDB` cluster database has been updated with the new version. So, the update process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo sample-mariadb
$ kubectl delete mariadbopsrequest -n demo mdops-update
```