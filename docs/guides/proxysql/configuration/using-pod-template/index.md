---
title: Run ProxySQL with Custom PodTemplate
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-configuration-usingpodtemplate
    name: Customize PodTemplate
    parent: guides-proxysql-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run ProxySQL with Custom PodTemplate

KubeDB supports providing custom configuration for ProxySQL via [PodTemplate](/docs/guides/proxysql/concepts/proxysql/index.md#specpodtemplate). This tutorial will show you how to use KubeDB to run a ProxySQL server with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/proxysql/configuration/using-pod-template/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare MySQL Backend

ProxySQL needs a backend server to proxy. We will use a 3 node MySQL Group Replication cluster set up with KubeDB.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "8.4.3"
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/backends/mysqlgrp/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created
```

Wait for the MySQL cluster to be `Ready`.

```bash
$ kubectl get my -n demo
NAME           VERSION   STATUS   AGE
mysql-server   8.4.3     Ready    7m6s
```

> You can use MariaDB or Percona XtraDB as a backend as well. Have a look at the other [ProxySQL backend examples](/docs/guides/proxysql/backends/).

## Overview

KubeDB allows providing a template for the ProxySQL pod through `.spec.podTemplate`. KubeDB operator will pass the information provided in `.spec.podTemplate` to the PetSet created for ProxySQL.

KubeDB accepts the following fields to set in `.spec.podTemplate`:

- metadata:
  - annotations (pod's annotation)
- controller:
  - annotations (petset's annotation)
- spec:
  - containers
  - env
  - resources
  - initContainers
  - imagePullSecrets
  - nodeSelector
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext
  - serviceAccountName

Read about the fields in detail in [PodTemplate concept](/docs/guides/proxysql/concepts/proxysql/index.md#specpodtemplate).

## CRD Configuration

Below is the YAML for the ProxySQL used in this example. Here, `spec.podTemplate.metadata.annotations` adds a custom annotation to the pod and `spec.podTemplate.spec.containers[].resources` requests compute resources for the `proxysql` container.

```yaml
apiVersion: kubedb.com/v1
kind: ProxySQL
metadata:
  name: sample-proxysql
  namespace: demo
spec:
  version: "2.7.3-debian"
  replicas: 1
  backend:
    name: mysql-server
  podTemplate:
    metadata:
      annotations:
        passMe: ToProxySQLPod
    spec:
      containers:
      - name: proxysql
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/configuration/using-pod-template/examples/proxysql-misc-config.yaml
proxysql.kubedb.com/sample-proxysql created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `sample-proxysql-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                READY   STATUS    RESTARTS   AGE
sample-proxysql-0   1/1     Running   0          45s
```

Now, we will check if the pod has started with the custom configuration we have provided.

Check that the annotation has been set on the pod.

```bash
$ kubectl get pod -n demo sample-proxysql-0 -o jsonpath='{.metadata.annotations}'
{"passMe":"ToProxySQLPod"}⏎   
```

Check that the resource requests have been set on the `proxysql` container.

```bash
$ kubectl get pod -n demo sample-proxysql-0 -o jsonpath='{.spec.containers[?(@.name=="proxysql")].resources}'
{"limits":{"memory":"256Mi"},"requests":{"cpu":"250m","memory":"256Mi"}}
```

We can see that both the annotation and the resource requests we provided through `spec.podTemplate` have been applied to the pod.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete proxysql -n demo sample-proxysql
proxysql.kubedb.com "sample-proxysql" deleted
$ kubectl delete mysql -n demo mysql-server
mysql.kubedb.com "mysql-server" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
