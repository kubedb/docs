---
title: Configure TLS in Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-tls-configure
    name: Configure TLS
    parent: qdrant-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS in Qdrant

This guide shows how to configure TLS for both client and peer-to-peer traffic in Qdrant.

## Before You Begin

- Install KubeDB operator.
- Install cert-manager and create an `Issuer` or `ClusterIssuer`.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Deploy Qdrant with TLS

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: tls-qdrant
  namespace: demo
spec:
  version: 1.17.0
  mode: Distributed
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: qdrant-issuer
    client: true
    p2p: true
  deletionPolicy: WipeOut
```

## Verify

```bash
$ kubectl get qdrant -n demo tls-qdrant
$ kubectl get secret -n demo | grep tls-qdrant
```

## Cleaning up

```bash
kubectl delete qdrant -n demo tls-qdrant
kubectl delete ns demo
```