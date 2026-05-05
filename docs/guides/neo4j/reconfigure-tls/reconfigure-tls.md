---
title: Reconfigure TLS of Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-tls-cluster
    name: Cluster
    parent: neo4j-reconfigure-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure TLS in Neo4j

This guide shows how to rotate or reconfigure TLS certificates of a Neo4j database using `Neo4jOpsRequest`.

## Apply TLS Reconfiguration

Rotate TLS certificates and update the Bolt protocol mode:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-reconfigure-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-neo4j
  tls:
    rotateCertificates: true
    bolt:
      mode: mTLS
```

Remove TLS from the database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-neo4j
  tls:
    remove: true
```

Add or replace TLS configuration using a cert-manager issuer:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-neo4j
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: neo4j-ca-issuer
```

```bash
$ kubectl apply -f neo4j-reconfigure-tls.yaml
neo4jopsrequest.ops.kubedb.com/neo4j-reconfigure-tls created
```

## Verify

```bash
$ kubectl get neo4jopsrequest -n demo neo4j-reconfigure-tls
NAME                    TYPE             STATUS       AGE
neo4j-reconfigure-tls   ReconfigureTLS   Successful   2m
```
