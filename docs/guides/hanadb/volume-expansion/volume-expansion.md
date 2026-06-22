---
title: HanaDB Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: hanadb-volume-expansion-guide
    name: Volume Expansion
    parent: hanadb-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDB Volume Expansion

This guide shows how to expand HanaDB persistent volumes using a `HanaDBOpsRequest`.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl`.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- Use a `StorageClass` that supports volume expansion.
- Create a namespace:

```bash
kubectl create ns demo
```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/volume-expansion](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/volume-expansion).

## Deploy HanaDB

The example uses `storageClassName: local-path`. Replace it with an expandable `StorageClass` in your cluster if needed.

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/volume-expansion/hanadb-standalone.yaml
kubectl wait -n demo hanadb/hanadb-standalone --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Apply Volume Expansion

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: hanadb-standalone
  volumeExpansion:
    mode: Online
    hanadb: 65Gi
  timeout: 30m
  apply: IfReady
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/volume-expansion/hdbops-volume-expansion.yaml
kubectl wait -n demo hanadbopsrequest/hdbops-volume-expansion --for=jsonpath='{.status.phase}'=Successful --timeout=1800s
kubectl wait -n demo hanadb/hanadb-standalone --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Verify

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=hanadb-standalone
kubectl get petset -n demo hanadb-standalone -o jsonpath='{.spec.volumeClaimTemplates[?(@.metadata.name=="data")].spec.resources.requests.storage}'
```

## Cleanup

```bash
kubectl delete hdbops -n demo hdbops-volume-expansion
kubectl delete hanadb -n demo hanadb-standalone
kubectl delete ns demo
```
