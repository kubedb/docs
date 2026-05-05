---
title: Reconfiguring Neo4j TLS
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-tls-overview
    name: Overview
    parent: neo4j-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Neo4j TLS

This guide shows how to rotate or update TLS materials of a Neo4j database.

## Before You Begin

- Install `cert-manager` in your cluster.
- Install KubeDB and Ops-manager from [here](/docs/setup/README.md).
- Deploy a TLS-enabled Neo4j database first by following the [Configure TLS guide](/docs/guides/neo4j/tls/configure/).
- Use the example file `docs/examples/neo4j/reconfigure-tls/ops-request.yaml` for the OpsRequest.
- Use namespace `demo` for isolation.

```bash
kubectl create ns demo
```

## Deploy a TLS-enabled Neo4j

After following the TLS configuration guide, make sure the `tls-neo4j` database is ready before applying the OpsRequest.

```bash
kubectl get neo4j -n demo tls-neo4j -w
```

## Apply ReconfigureTLS OpsRequest

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/reconfigure-tls/ops-request.yaml
kubectl get neo4jopsrequest -n demo neo4j-reconfigure-tls
```

## Verify

```bash
kubectl describe neo4jopsrequest -n demo neo4j-reconfigure-tls
```

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-reconfigure-tls
kubectl delete neo4j -n demo tls-neo4j
kubectl delete ns demo
```
