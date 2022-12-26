---
title: Upgrading PerconaXtraDB Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-upgrading-cluster
    name: Cluster
    parent: guides-perconaxtradb-upgrading
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Upgrade version of PerconaXtraDB Cluster

This guide will show you how to use `KubeDB` Enterprise operator to upgrade the version of `PerconaXtraDB` Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb)
  - [Cluster](/docs/guides/percona-xtradb/clustering/overview)
  - [PerconaXtraDBOpsRequest](/docs/guides/percona-xtradb/concepts/opsrequest)
  - [Upgrading Overview](/docs/guides/percona-xtradb/upgrading/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare PerconaXtraDB Cluster

Now, we are going to deploy a `PerconaXtraDB` cluster database with version `10.4.17`.

### Deploy PerconaXtraDB cluster

In this section, we are going to deploy a PerconaXtraDB Cluster. Then, in the next section we will upgrade the version of the database using `PerconaXtraDBOpsRequest` CRD. Below is the YAML of the `PerconaXtraDB` CR that we are going to create,

> If you want to upgrade `PerconaXtraDB Standalone`, Just remove the `spec.Replicas` from the below yaml and rest of the steps are same.

```yaml
apiVersion: kubedb.com/v1alpha2
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
  terminationPolicy: WipeOut

```

Let's create the `PerconaXtraDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/upgrading/cluster/examples/sample-pxc.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Now, wait until `sample-pxc` created has status `Ready`. i.e,

```bash
$ kubectl get perconaxtradb -n demo                                                                                                                                             
NAME             VERSION    STATUS     AGE
sample-pxc   8.0.26    Ready     3m15s
```

We are now ready to apply the `PerconaXtraDBOpsRequest` CR to upgrade this database.

### Upgrade PerconaXtraDB Version

Here, we are going to upgrade `PerconaXtraDB` cluster from `8.0.26` to `8.0.28`.

#### Create PerconaXtraDBOpsRequest:

In order to upgrade the database cluster, we have to create a `PerconaXtraDBOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `PerconaXtraDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-upgrade
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: sample-pxc
  upgrade:
    targetVersion: "8.0.28"
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `sample-pxc` PerconaXtraDB database.
- `spec.type` specifies that we are going to perform `Upgrade` on our database.
- `spec.upgrade.targetVersion` specifies the expected version of the database `8.0.28`.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/upgrading/cluster/examples/pxops-upgrade.yaml
perconaxtradbopsrequest.ops.kubedb.com/pxops-upgrade created
```

#### Verify PerconaXtraDB version upgraded successfully 

If everything goes well, `KubeDB` Enterprise operator will update the image of `PerconaXtraDB` object and related `StatefulSets` and `Pods`.

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CR,

```bash
$ kubectl get perconaxtradbopsrequest -n demo
Every 2.0s: kubectl get perconaxtradbopsrequest -n demo
NAME                TYPE      STATUS       AGE
pxops-upgrade      Upgrade   Successful    84s
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded.

Now, we are going to verify whether the `PerconaXtraDB` and the related `StatefulSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get perconaxtradb -n demo sample-pxc -o=jsonpath='{.spec.version}{"\n"}'
8.0.28

$ kubectl get sts -n demo sample-pxc -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
percona/percona-xtradb-cluster:8.0.28

$ kubectl get pods -n demo sample-pxc-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
percona/percona-xtradb-cluster:8.0.28
```

You can see from above, our `PerconaXtraDB` cluster database has been updated with the new version. So, the upgrade process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete perconaxtradb -n demo sample-pxc
$ kubectl delete perconaxtradbopsrequest -n demo pxops-upgrade
```