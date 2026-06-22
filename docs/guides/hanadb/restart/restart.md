---
title: Restart HanaDB
menu:
  docs_{{ .version }}:
    identifier: hanadb-restart-guide
    name: Restart HanaDB
    parent: hanadb-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# Restart HanaDB

This guide shows how to restart a KubeDB-managed HanaDB database using a `HanaDBOpsRequest`.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl`.
- Install KubeDB by following the [setup guide](/docs/setup/README.md).
- Create a namespace:

```bash
kubectl create ns demo
```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/restart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/restart).

## Deploy HanaDB

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/restart/hanadb-cluster.yaml
kubectl wait -n demo hanadb/hanadb-cluster --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Apply Restart

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: hanadb-cluster
  timeout: 30m
  apply: Always
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/restart/hdbops-restart.yaml
kubectl wait -n demo hanadbopsrequest/hdbops-restart --for=jsonpath='{.status.phase}'=Successful --timeout=1800s
kubectl wait -n demo hanadb/hanadb-cluster --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Verify

```bash
kubectl get hdbops -n demo hdbops-restart
kubectl get pods -n demo -l app.kubernetes.io/instance=hanadb-cluster
```

## Cleanup

```bash
kubectl delete hdbops -n demo hdbops-restart
kubectl delete hanadb -n demo hanadb-cluster
kubectl delete ns demo
```
