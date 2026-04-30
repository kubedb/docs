---
title: Configure TLS/SSL in Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-tls-configure
    name: Configure TLS
    parent: neo4j-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS in Neo4j

This guide will show you how to use KubeDB to configure TLS for Neo4j.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured to talk to it.
- Install KubeDB operator following [the setup guide](/docs/setup/README.md).

```bash
$ kubectl create ns demo
namespace/demo created
```

## Configure TLS

Create an `Issuer` or `ClusterIssuer`, then deploy a Neo4j database referencing TLS certificates.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: tls-neo4j
  namespace: demo
spec:
  version: "2025.11.2"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: neo4j-issuer
  deletionPolicy: WipeOut
```

The operator provisions certificates for the supported Neo4j protocols and mounts them into the database Pods.

## Verify TLS

```bash
$ kubectl get secret -n demo tls-neo4j-server-cert
NAME                   TYPE                DATA   AGE
tls-neo4j-server-cert  kubernetes.io/tls   3      2m
```

## Cleaning up

```bash
kubectl patch -n demo neo4j/tls-neo4j -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo neo4j/tls-neo4j
kubectl delete ns demo
```
