---
title: Run HanaDB with Custom RBAC Resources
menu:
  docs_{{ .version }}:
    identifier: hanadb-custom-rbac-quickstart
    name: Custom RBAC
    parent: hanadb-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# Using Custom RBAC Resources

KubeDB supports user-managed role-based access permissions for HanaDB. This tutorial shows how to run a HanaDB instance with custom RBAC resources.

## Before You Begin

Prepare a Kubernetes cluster and configure `kubectl` to communicate with it. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Install the KubeDB CLI on your workstation and the KubeDB operator in your cluster by following the [setup guide](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Overview

KubeDB allows users to provide custom RBAC resources for HanaDB: `ServiceAccount`, `Role`, `RoleBinding`, `ClusterRole`, and `ClusterRoleBinding`. Configure the service account through `spec.podTemplate.spec.serviceAccountName`. If this field is empty, the KubeDB operator creates a service account whose name matches the HanaDB object.

If you reference an existing service account, the KubeDB operator uses it. You are responsible for granting the required permissions.

## Custom RBAC for HanaDB

Create a `ServiceAccount` in the `demo` namespace.

```bash
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

Create a `Role` with the namespace-scoped permissions required by the HanaDB instance named `hanadb-custom-rbac`.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-custom-role
  namespace: demo
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - "*"
- apiGroups:
  - ""
  resources:
  - pods/exec
  verbs:
  - create
- apiGroups:
  - kubedb.com
  resources:
  - hanadbs
  verbs:
  - get
  - list
  - watch
  - patch
- apiGroups:
  - kubedb.com
  resources:
  - hanadbs/status
  verbs:
  - patch
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
  - create
  - update
- apiGroups:
  - apps.k8s.appscode.com
  resources:
  - petsets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - create
  - get
  - update
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/custom-rbac/hanadb-custom-role.yaml
role.rbac.authorization.k8s.io/my-custom-role created
```

Create a `RoleBinding` to bind this `Role` to the custom service account.

```bash
$ kubectl create rolebinding my-custom-rolebinding \
  --role=my-custom-role \
  --serviceaccount=demo:my-custom-serviceaccount \
  --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created
```

Create the cluster-scoped permissions required by the HanaDB pod.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: my-custom-clusterrole
rules:
- apiGroups:
  - catalog.kubedb.com
  resources:
  - hanadbversions
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - kubedb.com
  resources:
  - hanadbs
  verbs:
  - get
  - list
  - watch
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/custom-rbac/hanadb-custom-clusterrole.yaml
clusterrole.rbac.authorization.k8s.io/my-custom-clusterrole created
```

Bind the `ClusterRole` with the custom service account.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: my-custom-clusterrolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: my-custom-clusterrole
subjects:
- kind: ServiceAccount
  name: my-custom-serviceaccount
  namespace: demo
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/custom-rbac/hanadb-custom-clusterrolebinding.yaml
clusterrolebinding.rbac.authorization.k8s.io/my-custom-clusterrolebinding created
```

Create a HanaDB object with `spec.podTemplate.spec.serviceAccountName` set to `my-custom-serviceaccount`.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-custom-rbac
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
  storage:
    storageClassName: local-path
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/custom-rbac/hanadb-custom-db.yaml
hanadb.kubedb.com/hanadb-custom-rbac created
```

Check that the pod is running:

```bash
$ kubectl get pod -n demo hanadb-custom-rbac-0
NAME                   READY   STATUS    RESTARTS   AGE
hanadb-custom-rbac-0   1/1     Running   0          5m
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/hanadb-custom-rbac -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/hanadb-custom-rbac

kubectl delete -n demo serviceaccount my-custom-serviceaccount
kubectl delete -n demo role my-custom-role
kubectl delete -n demo rolebinding my-custom-rolebinding
kubectl delete clusterrole my-custom-clusterrole
kubectl delete clusterrolebinding my-custom-clusterrolebinding
kubectl delete ns demo
```
