---
title: Run PerconaXtraDB with Custom PodTemplate
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-configuration-usingpodtemplate
    name: Customize PodTemplate
    parent: guides-perconaxtradb-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run PerconaXtraDB with Custom PodTemplate

KubeDB supports providing custom configuration for PerconaXtraDB via [PodTemplate](/docs/guides/percona-xtradb/concepts/perconaxtradb/#specpodtemplate). This tutorial will show you how to use KubeDB to run a PerconaXtraDB database with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/percona-xtradb/configuration/using-pod-template/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for PerconaXtraDB database.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
  - annotations (pod's annotation)
- controller:
  - annotations (petset's annotation)
- spec:
  - env
  - resources
  - initContainers
  - imagePullSecrets
  - nodeSelector
  - affinity
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext

Read about the fields in details in [PodTemplate concept](/docs/guides/percona-xtradb/concepts/perconaxtradb/#specpodtemplate),

## CRD Configuration

Below is the YAML for the PerconaXtraDB created in this example. Here, [`spec.podTemplate.spec.env`](/docs/guides/percona-xtradb/concepts/perconaxtradb/#specpodtemplatespecenv) specifies environment variables and [`spec.podTemplate.spec.args`](/docs/guides/percona-xtradb/concepts/perconaxtradb/#specpodtemplatespecargs) provides extra arguments for [PerconaXtraDB Docker Image](https://hub.docker.com/_/perconaxtradb/).

In this tutorial, an initial database `mdDB` will be created by providing `env` `MYSQL_DATABASE` while the server character set will be set to `utf8mb4` by adding extra `args`. 

```yaml
apiVersion: kubedb.com/v1
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.26"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      containers:
      - name: perconaxtradb
        env:
        - name: MYSQL_DATABASE
          value: mdDB
        args:
        - --character-set-server=utf8mb4
        resources:
          requests:
            memory: "1Gi"
            cpu: "250m"
  deletionPolicy: WipeOut
```


```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/configuration/using-pod-template/examples/md-misc-config.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we will see that a pod with the name `sample-pxc` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME           READY   STATUS    RESTARTS   AGE
sample-pxc-0   2/2     Running   0          3m30s
sample-pxc-1   2/2     Running   0          3m30s
sample-pxc-2   2/2     Running   0          3m30s
```

Check the perconaxtradb CRD status if the database is ready

```bash
$ kubectl get perconaxtradb --all-namespaces
NAMESPACE   NAME         VERSION   STATUS   AGE
demo        sample-pxc   8.0.26    Ready    4m8s
```

Once we see `Note] mysqld: ready for connections.` in the log, the database is ready.

Now, we will check if the database has started with the custom configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 110
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| kubedb_system      |
| mdDB               |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
6 rows in set (0.01 sec)

# Check character_set_server
mysql> show variables like 'char%';
+--------------------------+---------------------------------------------+
| Variable_name            | Value                                       |
+--------------------------+---------------------------------------------+
| character_set_client     | latin1                                      |
| character_set_connection | latin1                                      |
| character_set_database   | utf8mb4                                     |
| character_set_filesystem | binary                                      |
| character_set_results    | latin1                                      |
| character_set_server     | utf8mb4                                     |
| character_set_system     | utf8mb3                                     |
| character_sets_dir       | /usr/share/percona-xtradb-cluster/charsets/ |
+--------------------------+---------------------------------------------+
8 rows in set (0.01 sec)

mysql> quit;
Bye
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete perconaxtradb -n demo sample-pxc
perconaxtradb.kubedb.com "sample-pxc" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
