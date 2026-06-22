---
title: Vertical Scaling HanaDB
menu:
  docs_{{ .version }}:
    identifier: hanadb-scaling-vertical-guide
    name: Scale Vertically
    parent: hanadb-scaling-vertical
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# Vertical Scaling HanaDB

This guide shows how to update CPU and memory resources for the HanaDB container using a `HanaDBOpsRequest`.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl`.
- Install KubeDB by following the [setup guide](/docs/setup/README.md).
- Create a namespace:

```bash
kubectl create ns demo
```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/scaling/vertical-scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/scaling/vertical-scaling).

## Deploy HanaDB

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/scaling/vertical-scaling/hanadb-cluster.yaml
kubectl wait -n demo hanadb/hanadb-cluster --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Apply Vertical Scaling

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: hanadb-cluster
  verticalScaling:
    hanadb:
      resources:
        requests:
          cpu: "2100m"
          memory: "8448Mi"
        limits:
          cpu: "4"
          memory: "14Gi"
  timeout: 30m
  apply: IfReady
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/scaling/vertical-scaling/hdbops-vscale.yaml
kubectl wait -n demo hanadbopsrequest/hdbops-vscale --for=jsonpath='{.status.phase}'=Successful --timeout=1800s
kubectl wait -n demo hanadb/hanadb-cluster --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Verify

```bash
kubectl get hanadb -n demo hanadb-cluster -o jsonpath='{.spec.podTemplate.spec.containers[?(@.name=="hanadb")].resources}'
kubectl get petsets.apps.k8s.appscode.com -n demo hanadb-cluster -o jsonpath='{.spec.template.spec.containers[?(@.name=="hanadb")].resources}'
```

## Cleanup

```bash
kubectl delete hdbops -n demo hdbops-vscale
kubectl delete hanadb -n demo hanadb-cluster
kubectl delete ns demo
```
