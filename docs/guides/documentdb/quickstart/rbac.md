---
title: RBAC for DocumentDB
menu:
  docs_{{ .version }}:
    identifier: documentdb-rbac-quickstart
    name: RBAC
    parent: documentdb-quickstart
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RBAC Permissions for DocumentDB

When RBAC is enabled in your cluster, KubeDB automatically creates the necessary Role, ServiceAccount, and RoleBinding for each DocumentDB instance. This tutorial explains what permissions are granted and how to verify them.

Here is the list of additional permissions required by the PetSet of DocumentDB:

| Kubernetes Resource | Resource Names        | Permission required |
|---------------------|-----------------------|---------------------|
| petsets             | `{documentdb-name}`   | get                 |
| pods                |                       | list, patch         |
| pods/exec           |                       | create              |
| documentdbs         |                       | get                 |
| configmaps          | `{documentdb-name}`   | get, update, create |

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a DocumentDB Database

Below is the DocumentDB object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: quick-docdb
  namespace: demo
spec:
  version: "5.0.6"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

Create the above DocumentDB object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/quickstart/quick-docdb.yaml
documentdb.kubedb.com/quick-docdb created
```

When this DocumentDB object is created, KubeDB operator creates Role, ServiceAccount and RoleBinding with the matching DocumentDB name and uses that ServiceAccount in the corresponding PetSet.

### Role

```bash
$ kubectl get role -n demo quick-docdb -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-docdb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: documentdbs.kubedb.com
  name: quick-docdb
  namespace: demo
rules:
- apiGroups:
  - apps
  resourceNames:
  - quick-docdb
  resources:
  - petsets
  verbs:
  - get
- apiGroups:
  - kubedb.com
  resourceNames:
  - quick-docdb
  resources:
  - documentdbs
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
$ kubectl get serviceaccount -n demo quick-docdb -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-docdb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: documentdbs.kubedb.com
  name: quick-docdb
  namespace: demo
```

### RoleBinding

```bash
$ kubectl get rolebinding -n demo quick-docdb -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-docdb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: documentdbs.kubedb.com
  name: quick-docdb
  namespace: demo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: quick-docdb
subjects:
- kind: ServiceAccount
  name: quick-docdb
  namespace: demo
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo documentdb/quick-docdb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo documentdb/quick-docdb

kubectl delete ns demo
```
