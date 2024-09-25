---
title: Vertical Scaling ProxySQL Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-scaling-vertical-cluster
    name: Demo
    parent: guides-proxysql-scaling-vertical
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale ProxySQL Cluster

This guide will show you how to use `KubeDB` Enterprise operator to update the resources of a ProxySQL cluster .

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [ProxySQL](/docs/guides/proxysql/concepts/proxysql)
    - [Clustering](/docs/guides/proxysql/clustering/proxysql-cluster)
    - [ProxySQLOpsRequest](/docs/guides/proxysql/concepts/opsrequest)
    - [Vertical Scaling Overview](/docs/guides/proxysql/scaling/vertical-scaling/overview)

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
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/scaling/vertical-scaling/cluster/example/sample-mysql.yaml
mysql.kubedb.com/mysql-server created 
```

After applying the above yaml wait for the MySQL to be Ready.

## Apply Vertical Scaling on Cluster

Here, we are going to deploy a  `ProxySQL` cluster using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare ProxySQL Cluster

Now, we are going to deploy a `ProxySQL` cluster database with version `2.3.2-debian`.

In this section, we are going to deploy a ProxySQL cluster. Then, in the next section we will update the resources of the servers using `ProxySQLOpsRequest` CRD. Below is the YAML of the `ProxySQL` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
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
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
      - name: proxysql
        resources:
          limits:
            cpu: 500m
            memory: 1Gi
          requests:
            cpu: 500m
            memory: 1Gi
```

Let's create the `ProxySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/scaling/vertical-scaling/cluster/example/sample-proxysql.yaml
proxysql.kubedb.com/proxy-server created
```

Now, wait until `proxy-server` has status `Ready`. i.e,

```bash
$ kubectl get proxysql -n demo
NAME             VERSION         STATUS     AGE
proxy-server    2.3.2-debian     Ready     3m46s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo proxy-server-0 -o json | jq '.spec.containers[].resources'
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

We are now ready to apply the `ProxySQLOpsRequest` CR to update the resources of this server.

### Scale Vertically

Here, we are going to update the resources of the server to meet the desired resources after scaling.

#### Create ProxySQLOpsRequest

In order to update the resources of the database, we have to create a `ProxySQLOpsRequest` CR with our desired resources. Below is the YAML of the `ProxySQLOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: proxyops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  proxyRef:
    name: proxy-server
  verticalScaling:
    proxysql:
      resources:
        requests:
          memory: "1.2Gi"
          cpu: "0.6"
        limits:
          memory: "1.2Gi"
          cpu: "0.6"
```

Here,

- `spec.proxyRef.name` specifies that we are performing vertical scaling operation on `proxy-server` instance.
- `spec.type` specifies that we are performing `VerticalScaling` on our server.
- `spec.verticalScaling.proxysql` specifies the desired resources after scaling.

Let's create the `ProxySQLOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/scaling/vertical-scaling/cluster/example/proxyops-vscale.yaml
proxysqlopsrequest.ops.kubedb.com/proxyops-vscale created
```

#### Verify ProxySQL Cluster resources updated successfully

If everything goes well, `KubeDB` Enterprise operator will update the resources of `ProxySQL` object and related `PetSets` and `Pods`.

Let's wait for `ProxySQLOpsRequest` to be `Successful`.  Run the following command to watch `ProxySQLOpsRequest` CR,

```bash
$ kubectl get proxysqlopsrequest -n demo
Every 2.0s: kubectl get proxysqlopsrequest -n demo
NAME                       TYPE              STATUS       AGE
proxyops-vscale        VerticalScaling      Successful    3m56s
```

We can see from the above output that the `ProxySQLOpsRequest` has succeeded. Now, we are going to verify from one of the Pod yaml whether the resources of the database has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo proxy-server-0 -o json | jq '.spec.containers[].resources'
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

The above output verifies that we have successfully scaled up the resources of the ProxySQL instance.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete proxysql -n demo proxy-server
$ kubectl delete proxysqlopsrequest -n demo proxyops-vscale
```