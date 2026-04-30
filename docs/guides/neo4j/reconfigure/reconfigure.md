---
title: Reconfigure Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-cluster
    name: Cluster
    parent: neo4j-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Neo4j

This guide shows how to reconfigure a Neo4j database using `Neo4jOpsRequest`.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured.
- Install KubeDB Community and Enterprise operators following [the setup guide](/docs/setup/README.md).
- Review [Neo4j](/docs/guides/neo4j/concepts/neo4j.md), [OpsRequest](/docs/guides/neo4j/concepts/opsrequest.md), and [Reconfigure Overview](/docs/guides/neo4j/reconfigure/overview.md).

```bash
$ kubectl create ns demo
```

## Prepare Database

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-prod
  namespace: demo
spec:
  version: "2025.11.2"
  replicas: 3
  configuration:
    secretName: neo4j-config
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Create an updated config secret and apply a `Neo4jOpsRequest`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: neo4j-prod
  configuration:
    configSecret:
      name: neo4j-config-updated
```

```bash
$ kubectl apply -f neo4j-reconfigure.yaml
neo4jopsrequest.ops.kubedb.com/neo4j-reconfigure created
```

## Verify Reconfiguration

```bash
$ kubectl get neo4jopsrequest -n demo neo4j-reconfigure
NAME                TYPE          STATUS       AGE
neo4j-reconfigure   Reconfigure   Successful   3m
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-reconfigure
kubectl patch -n demo neo4j/neo4j-prod -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo neo4j/neo4j-prod
kubectl delete ns demo
```
