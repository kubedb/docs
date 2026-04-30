---
title: Run Qdrant with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: qdrant-custom-rbac-quickstart
    name: Custom RBAC
    parent: qdrant-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC Resources

This tutorial shows how to run Qdrant with custom `ServiceAccount`, `Role`, and `RoleBinding` resources.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured for that cluster.
- Install KubeDB operator following [the setup guide](/docs/setup/README.md).
- This tutorial uses `demo` namespace.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Create Custom RBAC Resources

Create the service account first:

```bash
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

Create a role with the required permissions for Qdrant Pods and related resources:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-custom-role
  namespace: demo
rules:
- apiGroups:
  - kubedb.com
  resources:
  - qdrants
  resourceNames:
  - qdrant-custom-rbac
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - pods
  - pods/exec
  verbs:
  - get
  - list
  - create
```

Bind the role to the service account:

```bash
$ kubectl create rolebinding my-custom-rolebinding \
  --role=my-custom-role \
  --serviceaccount=demo:my-custom-serviceaccount \
  --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created
```

## Deploy Qdrant with Custom Service Account

## Deploy Qdrant with Custom Service Account

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-custom-rbac
  namespace: demo
spec:
  version: 1.17.0
  mode: Distributed
  replicas: 3
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f qdrant-custom-rbac.yaml
qdrant.kubedb.com/qdrant-custom-rbac created
```

## Verify

```bash
$ kubectl get qdrant -n demo qdrant-custom-rbac
NAME                VERSION   STATUS   AGE
qdrant-custom-rbac  1.17.0    Ready    2m

$ kubectl get pod -n demo -l app.kubernetes.io/instance=qdrant-custom-rbac
```

## Cleaning up

```bash
kubectl delete qdrant -n demo qdrant-custom-rbac
kubectl delete rolebinding -n demo my-custom-rolebinding
kubectl delete role -n demo my-custom-role
kubectl delete serviceaccount -n demo my-custom-serviceaccount
kubectl delete ns demo
```