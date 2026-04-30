---
title: Run Weaviate with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: weaviate-custom-rbac-quickstart
    name: Custom RBAC
    parent: weaviate-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC Resources

This tutorial shows how to run Weaviate with custom `ServiceAccount`, `Role`, and `RoleBinding` resources.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured for that cluster.
- Install KubeDB operator from [setup guide](/docs/setup/README.md).

```bash
$ kubectl create ns demo
namespace/demo created
```

## Create Custom RBAC Resources

```bash
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created

$ kubectl create role my-custom-role -n demo --verb=get,list,watch --resource=pods
role.rbac.authorization.k8s.io/my-custom-role created

$ kubectl create rolebinding my-custom-rolebinding \
  --role=my-custom-role \
  --serviceaccount=demo:my-custom-serviceaccount \
  --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created
```

## Deploy Weaviate with Custom Service Account

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-custom-rbac
  namespace: demo
spec:
  version: 1.33.1
  replicas: 3
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f weaviate-custom-rbac.yaml
weaviate.kubedb.com/weaviate-custom-rbac created
```

## Verify

```bash
$ kubectl get weaviate -n demo weaviate-custom-rbac
```

## Cleaning up

```bash
kubectl delete weaviate -n demo weaviate-custom-rbac
kubectl delete rolebinding -n demo my-custom-rolebinding
kubectl delete role -n demo my-custom-role
kubectl delete serviceaccount -n demo my-custom-serviceaccount
kubectl delete ns demo
```