---
title: Reconfigure TLS of Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-reconfigure-tls-cluster
    name: Cluster
    parent: qdrant-reconfigure-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure TLS in Qdrant

This guide shows how to reconfigure or rotate TLS certificates of a Qdrant database using `QdrantOpsRequest`.

## Before You Begin

- Install KubeDB Community and Enterprise operators.
- Ensure your database is already running with TLS enabled.

```bash
$ kubectl create ns demo
```

## Apply ReconfigureTLS OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdrant-reconfigure-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-qdrant
  tls:
    rotateCertificates: true
    client: true
    p2p: true
```

  ```bash
  $ kubectl apply -f qdrant-reconfigure-tls.yaml
  qdrantopsrequest.ops.kubedb.com/qdrant-reconfigure-tls created
  ```

  ## Verify

  ```bash
  $ kubectl get qdrantopsrequest -n demo qdrant-reconfigure-tls
  $ kubectl describe qdrantopsrequest -n demo qdrant-reconfigure-tls
  ```

  ## Cleaning up

  ```bash
  kubectl delete qdrantopsrequest -n demo qdrant-reconfigure-tls
  kubectl delete ns demo
  ```