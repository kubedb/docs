---
title: Updating ProxySQL Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-updating-cluster
    name: Demo
    parent: guides-proxysql-updating
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# update version of ProxySQL Cluster

This guide will show you how to use `KubeDB` Enterprise operator to update the version of `ProxySQL` Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql)
  - [Cluster](/docs/guides/proxysql/clustering/overview)
  - [ProxySQLOpsRequest](/docs/guides/proxysql/concepts/opsrequest)
  - [updating Overview](/docs/guides/proxysql/update-version/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

Also we need a mysql backend for the proxysql server. So we are  creating one with the below yaml. 

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "5.7.44"
  replicas: 3
  topology:
    mode: GroupReplication
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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/update-version/cluster/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created 
```

After applying the above yaml wait for the MySQL to be Ready.

## Prepare ProxySQL Cluster

Now, we are going to deploy a `ProxySQL` cluster with version `2.3.2-debian`.

### Deploy ProxySQL cluster

In this section, we are going to deploy a ProxySQL Cluster. Then, in the next section we will update the version of the instance using `ProxySQLOpsRequest` CRD. Below is the YAML of the `ProxySQL` CR that we are going to create,


```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.3.2-debian"
  replicas: 3
  backend:
    name: mysql-server
  syncUsers: true
  terminationPolicy: WipeOut

```

Let's create the `ProxySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/update-version/cluster/examples/sample-proxysql.yaml
proxysql.kubedb.com/proxy-server created
```

Now, wait until `proxy-server` created has status `Ready`. i.e,

```bash
$ kubectl get proxysql -n demo                                                                                                                                             
NAME             VERSION       STATUS     AGE
proxy-server   2.3.2-debian    Ready     3m15s
```

We are now ready to apply the `ProxySQLOpsRequest` CR to update this database.

## update ProxySQL Version

Here, we are going to update `ProxySQL` cluster from `2.3.2-debian` to `2.4.4-debian`.

### Create ProxySQLOpsRequest:

In order to update the database cluster, we have to create a `ProxySQLOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `ProxySQLOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: proxyops-update
  namespace: demo
spec:
  type: UpdateVersion
  proxyRef:
    name: proxy-server
  updateVersion:
    targetVersion: "2.4.4-debian"
```

Here,

- `spec.proxyRef.name` specifies that we are performing operation on `proxy-server` ProxySQL database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `2.4.4-debian`.

Let's create the `ProxySQLOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/update-version/cluster/examples/proxyops-update.yaml
proxysqlopsrequest.ops.kubedb.com/proxyops-update created
```

### Verify ProxySQL version updated successfully 

If everything goes well, `KubeDB` Enterprise operator will update the image of `ProxySQL` object and related `StatefulSets` and `Pods`.

Let's wait for `ProxySQLOpsRequest` to be `Successful`.  Run the following command to watch `ProxySQLOpsRequest` CR,

```bash
$ kubectl get proxysqlopsrequest -n demo
Every 2.0s: kubectl get proxysqlopsrequest -n demo
NAME                 TYPE            STATUS       AGE
proxyops-update      UpdateVersion   Successful    84s
```

We can see from the above output that the `ProxySQLOpsRequest` has succeeded.

Now, we are going to verify whether the `ProxySQL` and the related `StatefulSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get proxysql -n demo proxy-server -o=jsonpath='{.spec.version}{"\n"}'
2.4.4-debian

$ kubectl get sts -n demo proxy-server -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
kubedb/proxysql:2.4.4-debian@sha256....

$ kubectl get pods -n demo proxy-server-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
kubedb/proxysql:2.4.4-debian@sha256....

```

You can see from above, our `ProxySQL` cluster database has been updated with the new version. So, the update process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete proxysql -n demo proxy-server
$ kubectl delete proxysqlopsrequest -n demo proxyops-update
```