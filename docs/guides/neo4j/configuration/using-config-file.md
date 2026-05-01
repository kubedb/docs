---
title: Run Neo4j with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: neo4j-using-config-file
    name: Config File
    parent: neo4j-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Neo4j. This tutorial will show you how to use KubeDB to run Neo4j with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Overview

Neo4j supports configuration via `neo4j.conf`. KubeDB uses `spec.configuration.secretName` to allow users to provide a custom configuration Secret. The operator mounts and applies this configuration automatically.

In this tutorial, we will configure `dbms.logs.query.enabled` and `server.memory.heap.initial_size`.

## Custom Configuration

At first, let's create a custom `neo4j.conf` file:

```properties
dbms.logs.query.enabled=INFO
server.memory.heap.initial_size=512m
server.memory.heap.max_size=512m
```

Now, create a Secret with this configuration file.

```bash
$ kubectl create secret generic -n demo neo4j-configuration \
  --from-file=neo4j.conf=./neo4j.conf
secret/neo4j-configuration created
```

Verify the Secret has the configuration file.

```bash
$ kubectl get secret -n demo neo4j-configuration -o yaml
apiVersion: v1
data:
  neo4j.conf: <base64-encoded-content>
kind: Secret
metadata:
  name: neo4j-configuration
  namespace: demo
```

Now, create Neo4j CRD specifying `spec.configuration.secretName` field.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: custom-neo4j
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  configuration:
    secretName: neo4j-configuration
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/configuration/neo4j-configuration.yaml
neo4j.kubedb.com/custom-neo4j created
```

Now, wait for the Neo4j to be ready.

```bash
$ kubectl get neo4j -n demo custom-neo4j
NAME           VERSION   STATUS   AGE
custom-neo4j   2025.11.2 Ready    3m
```

Verify the applied configuration:

```bash
$ kubectl exec -it -n demo custom-neo4j-0 -- cat /var/lib/neo4j/conf/neo4j.conf | grep -E "dbms.logs.query.enabled|server.memory.heap"
dbms.logs.query.enabled=INFO
server.memory.heap.initial_size=512m
server.memory.heap.max_size=512m
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo neo4j/custom-neo4j -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo neo4j/custom-neo4j

kubectl delete -n demo secret neo4j-configuration
kubectl delete ns demo
```
