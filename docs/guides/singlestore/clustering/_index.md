---
title: SingleStore Clustering
menu:
  docs_{{ .version }}:
    identifier: sdb-clustering
    name: SingleStore Clustering
    parent: guides-singlestore
    weight: 15
menu_name: docs_{{ .version }}
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - SingleStore Cluster

This tutorial will show you how to use KubeDB to provision a singlestore cluster.

## Before You Begin

Before proceeding:

- Read [mariadb galera cluster concept](/docs/guides/mariadb/clustering/overview) to learn about MariaDB Group Replication.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).


