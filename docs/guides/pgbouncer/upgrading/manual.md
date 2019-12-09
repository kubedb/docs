---
title: Upgrading Operator
menu:
  docs_{{ .version }}:
    identifier: pb-upgrade-manual
    name: Manual
    parent: pb-upgrading-pgbouncer
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# KubeDB Upgrade Manual

This tutorial will show you how to upgrade KubeDB from previous version to 0.11.0.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [minikube](https://github.com/kubernetes/minikube).

Now, install KubeDB 0.9.0 cli on your workstation and KubeDB operator in your cluster following the steps [here](https://kubedb.com/docs/0.9.0/setup/install/).

## Previous operator

In this tutorial we are using helm to install kubedb 0.9.0 release. But, user can install kubedb operator from script too.

```console
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

# Step 1: Install kubedb operator chart
$ helm install appscode/kubedb --name kubedb-operator --version 0.9.0 \
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
$ helm install appscode/kubedb-catalog --name kubedb-catalog --version 0.9.0 \
  --namespace kube-system

# Step 3(b): Or, if previously installed, upgrade KubeDB catalog of database versions
$ helm upgrade kubedb-catalog appscode/kubedb-catalog --version 0.9.0 \
  --namespace kube-system

$ helm ls
NAME           	REVISION	UPDATED                 	STATUS  	CHART               	APP VERSION	NAMESPACE
kubedb-catalog 	1       	Fri Feb  8 11:21:34 2019	DEPLOYED	kubedb-catalog-0.9.0	0.9.0      	kube-system
kubedb-operator	1       	Fri Feb  8 11:18:46 2019	DEPLOYED	kubedb-0.9.0        	0.9.0      	kube-system
```

## Upgrade kubedb-operator

For helm, `upgrade` command works fine.

```console
$ helm upgrade --install kubedb-operator appscode/kubedb --version 0.11.0 --namespace kube-system
$ helm upgrade --install kubedb-catalog appscode/kubedb-catalog --version 0.11.0 --namespace kube-system

$ helm ls
NAME           	REVISION	UPDATED                 	STATUS  	CHART               	APP VERSION	NAMESPACE
kubedb-catalog 	2       	Fri Feb  8 12:12:45 2019	DEPLOYED	kubedb-catalog-0.11.0	0.11.0      	kube-system
kubedb-operator	2       	Fri Feb  8 12:11:57 2019	DEPLOYED	kubedb-0.11.0        	0.11.0      	kube-system
```

## Upgrade CRD objects

Note that, if any server becomes stale for using deprecated PgBouncerVersion, upgrading to a non-deprecated PgBouncerVersion will solve the issue as kubedb-operator will update the Statefulsets too.

## Next Steps

- Get started with PgBouncer using  [quickstart](/docs/guides/pgbouncer/quickstart/quickstart.md).
- Learn about [custom PgBouncerVersions](/docs/guides/pgbouncer/custom-versions/setup.md).
- Learn about [private registry](/docs/guides/pgbouncer/private-registry/using-private-registry.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
