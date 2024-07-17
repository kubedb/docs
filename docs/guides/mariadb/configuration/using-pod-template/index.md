---
title: Run MariaDB with Custom PodTemplate
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-configuration-usingpodtemplate
    name: Customize PodTemplate
    parent: guides-mariadb-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run MariaDB with Custom PodTemplate

KubeDB supports providing custom configuration for MariaDB via [PodTemplate](/docs/guides/mariadb/concepts/mariadb/#specpodtemplate). This tutorial will show you how to use KubeDB to run a MariaDB database with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mariadb/configuration/using-pod-template/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for MariaDB database.

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

Read about the fields in details in [PodTemplate concept](/docs/guides/mariadb/concepts/mariadb/#specpodtemplate),

## CRD Configuration

Below is the YAML for the MariaDB created in this example. Here, [`spec.podTemplate.spec.env`](/docs/guides/mariadb/concepts/mariadb/#specpodtemplatespecenv) specifies environment variables and [`spec.podTemplate.spec.args`](/docs/guides/mariadb/concepts/mariadb/#specpodtemplatespecargs) provides extra arguments for [MariaDB Docker Image](https://hub.docker.com/_/mariadb/).

In this tutorial, an initial database `mdDB` will be created by providing `env` `MYSQL_DATABASE` while the server character set will be set to `utf8mb4` by adding extra `args`. 

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.23"
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
      - name: mariadb
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/configuration/using-pod-template/examples/md-misc-config.yaml
mariadb.kubedb.com/sample-mariadb created
```

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we will see that a pod with the name `sample-mariadb` has been created.

Check that the petset's pod is running

```bash
$ $ kubectl get pod -n demo
NAME               READY   STATUS    RESTARTS   AGE
sample-mariadb-0   1/1     Running   0          96s
```

Check the pod's log to see if the database is ready

```bash
$ kubectl logs -f -n demo sample-mariadb-0
2021-03-18 06:06:17+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 1:10.5.23+maria~focal started.
2021-03-18 06:06:18+00:00 [Note] [Entrypoint]: Switching to dedicated user 'mysql'
2021-03-18 06:06:18+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 1:10.5.23+maria~focal started.
2021-03-18 06:06:19+00:00 [Note] [Entrypoint]: Initializing database files
...
2021-03-18  6:06:33 0 [Note] mysqld: ready for connections.
Version: '10.5.23-MariaDB-1:10.5.23+maria~focal'  socket: '/run/mysqld/mysqld.sock'  port: 3306  mariadb.org binary distribution
```

Once we see `Note] mysqld: ready for connections.` in the log, the database is ready.

Now, we will check if the database has started with the custom configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 22
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# Check mdDB
MariaDB [(none)]> show databases;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mdDB               |
| mysql              |
| performance_schema |
+--------------------+
4 rows in set (0.001 sec)

# Check character_set_server
MariaDB [(none)]> show variables like 'char%';
+--------------------------+----------------------------+
| Variable_name            | Value                      |
+--------------------------+----------------------------+
| character_set_client     | latin1                     |
| character_set_connection | latin1                     |
| character_set_database   | utf8mb4                    |
| character_set_filesystem | binary                     |
| character_set_results    | latin1                     |
| character_set_server     | utf8mb4                    |
| character_set_system     | utf8                       |
| character_sets_dir       | /usr/share/mysql/charsets/ |
+--------------------------+----------------------------+
8 rows in set (0.001 sec)

MariaDB [(none)]> quit;
Bye
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo sample-mariadb
mariadb.kubedb.com "sample-mariadb" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
