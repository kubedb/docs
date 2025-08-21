---
title: MySQL StorageClass Migration Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-migration-storageclass
    name: StorageClass Migration
    parent: guides-mysql-migration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQL StorageClass Migration

This guide will show you how to use `KubeDB` Ops Manager to  migrate `StorageClass` of MySQL database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have at least two `StorageClass` resources in order to perform a migration.

- Install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [MySQL](/docs/guides/mysql/concepts/mysqldatabase)
    - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare MySQL Database

At first verify that your cluster has at least two `StorageClass`. Let's check,

```bash
➤ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  12d
longhorn               driver.longhorn.io      Delete          Immediate              true                   12d
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   12d
```
From the above output we can see that we have more than two `StorageClass` resources. We will now deploy a `MySQL` database using `local-path` StorageClass and insert some data into it. 
After that, we will apply `MySQLOPSRequest` to migrate StorageClass from `local-path` to `longhorn`. 

KubeDB implements a `MySQL` CRD to define the specification of a MySQL database. Below is the `MySQL` object created in this tutorial. 

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "9.1.0"
  replicas: 3
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/migration/sample-mysql.yaml
mysql.kubedb.com/sample-mysql created
```
Now, wait until sample-mysql has status `Ready`. i.e,

```bash
$ kubectl get mysql -n demo
NAME           VERSION   STATUS   AGE
sample-mysql   9.1.0     Ready    9m32s
```

Lets create a table in the primary.

```bash
# find the primary pod
kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1 }'
sample-mysql-0
```

