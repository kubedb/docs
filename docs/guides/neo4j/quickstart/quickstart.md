---
title: Neo4j Quickstart
menu:
  docs_{{ .version }}:
    identifier: neo4j-quickstart-overview
    name: Overview
    parent: neo4j-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Neo4j

This tutorial shows how to run Neo4j with KubeDB.

> Note: YAML files used in this tutorial are stored in [docs/examples/neo4j/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/neo4j/quickstart).

## Before You Begin

- Prepare a Kubernetes cluster and `kubectl`.
- Install KubeDB from [/docs/setup/README.md](/docs/setup/README.md).
- This tutorial uses `docs/examples/neo4j/quickstart/neo4j.yaml` as the working example manifest.
- Create namespace:

```bash
kubectl create ns demo
```

## Check Available StorageClass

```bash
kubectl get storageclass
```

## Check Available Neo4jVersion

```bash
kubectl get neo4jversions
```

## Create a Neo4j Database

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  replicas: 3
  deletionPolicy: WipeOut
  version: "2025.12.1"
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/quickstart/neo4j.yaml
kubectl get neo4j -n demo neo4j-test -w
```

## Verify Neo4j Database

```bash
kubectl get neo4j -n demo
kubectl describe neo4j -n demo neo4j-test
```

When `status.phase` becomes `Ready`, the Neo4j cluster is ready to accept Bolt and HTTP connections.

## Cleaning up

```bash
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```