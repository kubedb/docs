---
title: Install
menu:
  docs_{{ .version }}:
    identifier: install-kubedb
    name: Install
    parent: setup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: setup
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Installation Guide

There are 2 parts to installing KubeDB. You need to install a Kubernetes operator in your cluster using scripts or via Helm and download kubedb cli on your workstation. You can also use kubectl cli with KubeDB custom resource objects.

## Install KubeDB Operator

To use `kubedb`, you will need to install KubeDB [operator](https://github.com/kubedb/operator). KubeDB operator can be installed via a script or as a Helm chart.

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="helm3-tab" data-toggle="tab" href="#helm3" role="tab" aria-controls="helm3" aria-selected="true">Helm 3 (Recommended)</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="helm2-tab" data-toggle="tab" href="#helm2" role="tab" aria-controls="helm2" aria-selected="false">Helm 2</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="helm3" role="tabpanel" aria-labelledby="helm3-tab">

## Using Helm 3

KubeDB can be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/kubedb/installer/tree/{{< param "info.version" >}}/charts/kubedb) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `my-release`:

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search repo appscode/kubedb
NAME                    CHART VERSION APP VERSION   DESCRIPTION
appscode/kubedb         {{< param "info.version" >}}  {{< param "info.version" >}}  KubeDB by AppsCode - Production ready databases on Kubern...
appscode/kubedb-catalog {{< param "info.version" >}}  {{< param "info.version" >}}  KubeDB Catalog by AppsCode - Catalog for database versions

# Step 1: Install kubedb operator chart
$ helm install kubedb-operator appscode/kubedb --version {{< param "info.version" >}} --namespace kube-system

# Step 2: wait until crds are registered
$ kubectl get crds -l app=kubedb -w
NAME                               AGE
dormantdatabases.kubedb.com        6s
elasticsearches.kubedb.com         12s
elasticsearchversions.kubedb.com   8s
etcds.kubedb.com                   8s
etcdversions.kubedb.com            8s
memcacheds.kubedb.com              6s
memcachedversions.kubedb.com       6s
mongodbs.kubedb.com                7s
mongodbversions.kubedb.com         6s
mysqls.kubedb.com                  7s
mysqlversions.kubedb.com           7s
postgreses.kubedb.com              8s
postgresversions.kubedb.com        7s
redises.kubedb.com                 6s
redisversions.kubedb.com           6s
snapshots.kubedb.com               6s

# Step 3(a): Install KubeDB catalog of database versions
$ helm install kubedb-catalog appscode/kubedb-catalog --version {{< param "info.version" >}} --namespace kube-system

# Step 3(b): Or, if previously installed, upgrade KubeDB catalog of database versions
$ helm upgrade kubedb-catalog appscode/kubedb-catalog --version {{< param "info.version" >}} --namespace kube-system
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.version" >}}/charts/kubedb).

</div>
<div class="tab-pane fade" id="helm2" role="tabpanel" aria-labelledby="helm2-tab">

## Using Helm 2

KubeDB can be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/kubedb/installer/tree/{{< param "info.version" >}}/charts/kubedb) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `my-release`:

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search appscode/kubedb
NAME                   	CHART VERSION	APP VERSION 	DESCRIPTION
appscode/kubedb        	{{< param "info.version" >}} 	{{< param "info.version" >}}	KubeDB by AppsCode - Production ready databases on Kubern...
appscode/kubedb-catalog	{{< param "info.version" >}} 	{{< param "info.version" >}}	KubeDB Catalog by AppsCode - Catalog for database versions

# Step 1: Install kubedb operator chart
$ helm install appscode/kubedb --name kubedb-operator --version {{< param "info.version" >}} \
  --namespace kube-system

# Step 2: wait until crds are registered
$ kubectl get crds -l app=kubedb -w
NAME                               AGE
dormantdatabases.kubedb.com        6s
elasticsearches.kubedb.com         12s
elasticsearchversions.kubedb.com   8s
etcds.kubedb.com                   8s
etcdversions.kubedb.com            8s
memcacheds.kubedb.com              6s
memcachedversions.kubedb.com       6s
mongodbs.kubedb.com                7s
mongodbversions.kubedb.com         6s
mysqls.kubedb.com                  7s
mysqlversions.kubedb.com           7s
postgreses.kubedb.com              8s
postgresversions.kubedb.com        7s
redises.kubedb.com                 6s
redisversions.kubedb.com           6s
snapshots.kubedb.com               6s

# Step 3(a): Install KubeDB catalog of database versions
$ helm install appscode/kubedb-catalog --name kubedb-catalog --version {{< param "info.version" >}} \
  --namespace kube-system

# Step 3(b): Or, if previously installed, upgrade KubeDB catalog of database versions
$ helm upgrade kubedb-catalog appscode/kubedb-catalog --version {{< param "info.version" >}} \
  --namespace kube-system
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.version" >}}/charts/kubedb).

</div>
<div class="tab-pane fade" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using YAML

If you prefer to not use Helm, you can generate YAMLs from KubeDB chart and deploy using `kubectl`. Here we are going to show the prodecure using Helm 3.

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search repo appscode/kubedb
NAME                    CHART VERSION APP VERSION   DESCRIPTION
appscode/kubedb         {{< param "info.version" >}}  {{< param "info.version" >}}  KubeDB by AppsCode - Production ready databases on Kubern...
appscode/kubedb-catalog {{< param "info.version" >}}  {{< param "info.version" >}}  KubeDB Catalog by AppsCode - Catalog for database versions

# Step 1: Install kubedb operator chart
$ helm template kubedb-operator appscode/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kube-system \
  --no-hooks | kubectl apply -f -

# Step 2: wait until crds are registered
$ kubectl get crds -l app=kubedb -w
NAME                               AGE
dormantdatabases.kubedb.com        6s
elasticsearches.kubedb.com         12s
elasticsearchversions.kubedb.com   8s
etcds.kubedb.com                   8s
etcdversions.kubedb.com            8s
memcacheds.kubedb.com              6s
memcachedversions.kubedb.com       6s
mongodbs.kubedb.com                7s
mongodbversions.kubedb.com         6s
mysqls.kubedb.com                  7s
mysqlversions.kubedb.com           7s
postgreses.kubedb.com              8s
postgresversions.kubedb.com        7s
redises.kubedb.com                 6s
redisversions.kubedb.com           6s
snapshots.kubedb.com               6s

# Step: Install/Upgrade KubeDB catalog of database versions
$ helm template kubedb-catalog appscode/kubedb-catalog \
  --version {{< param "info.version" >}} \
  --namespace kube-system \
  --no-hooks | kubectl apply -f -
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.version" >}}/charts/kubedb).

</div>
</div>

### Installing in GKE Cluster

If you are installing KubeDB on a GKE cluster, you will need cluster admin permissions to install KubeDB operator. Run the following command to grant admin permision to the cluster.

```console
$ kubectl create clusterrolebinding "cluster-admin-$(whoami)" \
  --clusterrole=cluster-admin \
  --user="$(gcloud config get-value core/account)"
```


## Verify operator installation

To check if KubeDB operator pods have started, run the following command:

```console
$ kubectl get pods --all-namespaces -l app=kubedb --watch
```

Once the operator pods are running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm CRD groups have been registered by the operator, run the following command:

```console
$ kubectl get crd -l app=kubedb
```

Now, you are ready to [create your first database](/docs/guides/README.md) using KubeDB.


## Install KubeDB CLI

KubeDB provides a CLI to work with database objects. Download pre-built binaries from [kubedb/cli Github releases](https://github.com/kubedb/cli/releases) and put the binary to some directory in your `PATH`. To install on Linux 64-bit and MacOS 64-bit you can run the following commands:

```console
# Linux amd 64-bit
wget -O kubedb https://github.com/kubedb/cli/releases/download/{{< param "info.version" >}}/kubedb-linux-amd64 \
  && chmod +x kubedb \
  && sudo mv kubedb /usr/local/bin/

# Mac 64-bit
wget -O kubedb https://github.com/kubedb/cli/releases/download/{{< param "info.version" >}}/kubedb-darwin-amd64 \
  && chmod +x kubedb \
  && sudo mv kubedb /usr/local/bin/
```

If you prefer to install KubeDB cli from source code, you will need to set up a GO development environment following [these instructions](https://golang.org/doc/code.html). Then, install `kubedb` CLI using `go get` from source code.

```console
go get github.com/kubedb/cli/...
```

Please note that this will install KubeDB cli from master branch which might include breaking and/or undocumented changes.


## Configuring RBAC

KubeDB installer will create 3 user facing cluster roles:

| ClusterRole       | Aggregates To | Desription |
| ----------------- | --------------| ---------- |
| kubedb:core:admin | admin         | Allows edit access to all `KubeDB` CRDs, intended to be granted within a namespace using a RoleBinding. This grants ability to wipeout dormant database and delete their record. |
| kubedb:core:edit  | edit          | Allows edit access to all `KubeDB` CRDs except `DormantDatabase` CRD, intended to be granted within a namespace using a RoleBinding. |
| kubedb:core:view  | view          | Allows read-only access to `KubeDB` CRDs, intended to be granted within a namespace using a RoleBinding. |

These user facing roles supports [ClusterRole Aggregation](https://kubernetes.io/docs/admin/authorization/rbac/#aggregated-clusterroles) feature in Kubernetes 1.9 or later clusters.


## Upgrade KubeDB

To upgrade KubeDB cli, just replace the old cli with the new version. To upgrade KubeDB operator, please follow the instruction for the corresponding release.
