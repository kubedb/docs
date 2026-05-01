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

When RBAC is enabled in your cluster, KubeDB automatically creates the necessary Role, ServiceAccount, and RoleBinding for each HanaDB instance. This tutorial explains what permissions are granted and how to verify them.

Here is the list of additional permissions required by the PetSet of HanaDB:

| Kubernetes Resource | Resource Names    | Permission required |
|---------------------|-------------------|---------------------|
| petsets             | `{hanadb-name}`   | get                 |
| pods                |                   | list, patch         |
| pods/exec           |                   | create              |
| hanadbs             |                   | get                 |
| configmaps          | `{hanadb-name}`   | get, update, create |

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Create a HanaDB Database

Below is the HanaDB object created in this tutorial.

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
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: Delete
```

Create the above HanaDB object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/quickstart/quick-hanadb.yaml
hanadb.kubedb.com/quick-hanadb created
```

When this HanaDB object is created, KubeDB operator creates Role, ServiceAccount and RoleBinding with the matching HanaDB name.

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
  - apps
  resourceNames:
  - quick-hanadb
  resources:
  - petsets
  verbs:
  - get
- apiGroups:
  - kubedb.com
  resourceNames:
  - quick-hanadb
  resources:
  - hanadbs
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - patch
- apiGroups:
  - ""
  resources:
  - pods/exec
  verbs:
  - create
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

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/quick-hanadb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/quick-hanadb

kubectl delete ns demo
```
