---
title: Restart Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-restart-overview
    name: Restart Neo4j
    parent: neo4j-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Neo4j

This guide shows how to restart Neo4j pods using `Neo4jOpsRequest`.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured.
- Install KubeDB and Ops-manager from [here](/docs/setup/README.md).
- Create an isolated namespace:

```bash
kubectl create ns demo
```

## Deploy Neo4j

Save this as `neo4j.yaml`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  version: "2025.10.1"
  replicas: 3
  storage:
    resources:
      requests:
        storage: 2Gi
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
  deletionPolicy: WipeOut
```

Apply it and wait for the cluster to become ready:

```bash
kubectl apply -f neo4j.yaml

kubectl get neo4j -n demo neo4j-test -w
```

Wait until `STATUS` shows `Ready` before proceeding.

```
NAME         VERSION     STATUS   AGE
neo4j-test   2025.10.1   Ready    3m
```

## Apply Restart OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: neo4j-test
  timeout: 5m
  apply: Always
```

`apply: Always` tells KubeDB to execute the restart even if the database is not currently in the ready state.

```bash
$ cat <<'EOF' | kubectl apply -f -
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: neo4j-test
  timeout: 5m
  apply: Always
EOF
neo4jopsrequest.ops.kubedb.com/neo4j-restart created

$ kubectl wait --for=jsonpath='{.status.phase}'=Successful neo4jopsrequest/neo4j-restart -n demo --timeout=600s
neo4jopsrequest.ops.kubedb.com/neo4j-restart condition met
```

## Verify

```bash
$ kubectl get neo4jopsrequest -n demo neo4j-restart
NAME            TYPE      STATUS       AGE
neo4j-restart   Restart   Successful   1m

$ kubectl describe neo4jopsrequest -n demo neo4j-restart
Name:         neo4j-restart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         Neo4jOpsRequest
Metadata:
  ...
Spec:
  Type:  Restart
  Database Ref:
    Name:  neo4j-test
  Apply:  Always
Status:
  Phase:  Successful
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-restart
kubectl delete neo4j -n demo neo4j-test
kubectl delete ns demo
```