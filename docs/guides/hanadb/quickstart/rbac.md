---
title: RBAC for HanaDB
menu:
  docs_{{ .version }}:
    identifier: hanadb-rbac-quickstart
    name: RBAC
    parent: hanadb-quickstart
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RBAC Permissions for HanaDB

When RBAC is enabled in your cluster, KubeDB automatically creates the RBAC resources required by each HanaDB instance. This tutorial explains what permissions are granted and how to verify them.

The HanaDB pods require the following namespace-scoped permissions:

| Kubernetes Resource | Permission required              |
|---------------------|----------------------------------|
| pods                | `*`                              |
| pods/exec           | `create`                         |
| hanadbs             | `get`, `list`, `watch`, `patch`  |
| hanadbs/status      | `patch`                          |
| secrets             | `get`, `list`, `create`, `update` |
| petsets             | `get`, `list`, `watch`           |
| configmaps          | `create`, `get`, `update`        |

They also require cluster-scoped read access to `hanadbversions` and `hanadbs`.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl` to communicate with it. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install the KubeDB CLI on your workstation and the KubeDB operator in your cluster by following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Create a HanaDB Database

The following manifest creates the HanaDB instance used in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: quick-hanadb
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
  deletionPolicy: Delete
```

Create the HanaDB object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/quickstart/quick-hanadb.yaml
hanadb.kubedb.com/quick-hanadb created
```

When the HanaDB object is created, the KubeDB operator creates a `Role`, `ServiceAccount`, `RoleBinding`, `ClusterRole`, and `ClusterRoleBinding` with the matching HanaDB name.

### Role

```bash
$ kubectl get role -n demo quick-hanadb -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-hanadb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: hanadbs.kubedb.com
  name: quick-hanadb
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

### ServiceAccount

```bash
$ kubectl get serviceaccount -n demo quick-hanadb -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-hanadb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: hanadbs.kubedb.com
  name: quick-hanadb
  namespace: demo
```

### RoleBinding

```bash
$ kubectl get rolebinding -n demo quick-hanadb -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-hanadb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: hanadbs.kubedb.com
  name: quick-hanadb
  namespace: demo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: quick-hanadb
subjects:
- kind: ServiceAccount
  name: quick-hanadb
  namespace: demo
```

### ClusterRole

```bash
$ kubectl get clusterrole quick-hanadb -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: quick-hanadb
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

### ClusterRoleBinding

```bash
$ kubectl get clusterrolebinding quick-hanadb -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-hanadb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: hanadbs.kubedb.com
  name: quick-hanadb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: quick-hanadb
subjects:
- kind: ServiceAccount
  name: quick-hanadb
  namespace: demo
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/quick-hanadb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/quick-hanadb

kubectl delete clusterrolebinding quick-hanadb
kubectl delete clusterrole quick-hanadb
kubectl delete ns demo
```
