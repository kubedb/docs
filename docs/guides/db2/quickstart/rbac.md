---
title: RBAC for DB2
menu:
  docs_{{ .version }}:
    identifier: db2-rbac-quickstart
    name: RBAC
    parent: db2-quickstart-db2
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RBAC Permissions for DB2

When RBAC is enabled in your cluster, KubeDB automatically creates the necessary Role, ServiceAccount, and RoleBinding for each DB2 instance. This tutorial explains what permissions are granted and how to verify them.

Here is the list of additional permissions required by the PetSet of DB2:

| Kubernetes Resource | Resource Names    | Permission required |
|---------------------|-------------------|---------------------|
| petsets             | `{db2-name}`      | get                 |
| pods                |                   | list, patch         |
| pods/exec           |                   | create              |
| db2s                |                   | get                 |
| configmaps          | `{db2-name}`      | get, update, create |

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/db2](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/db2) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a DB2 Database

Below is the DB2 object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DB2
metadata:
  name: quick-db2
  namespace: demo
spec:
  version: "11.5.9"
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

Create the above DB2 object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/db2/quickstart/quick-db2.yaml
db2.kubedb.com/quick-db2 created
```

When this DB2 object is created, KubeDB operator creates Role, ServiceAccount and RoleBinding with the matching DB2 name and uses that ServiceAccount in the corresponding PetSet.

Let's see what KubeDB operator has created for additional RBAC permission.

### Role

KubeDB operator creates a Role object `quick-db2` in same namespace as the DB2 object.

```bash
$ kubectl get role -n demo quick-db2 -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-db2
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: db2s.kubedb.com
  name: quick-db2
  namespace: demo
rules:
- apiGroups:
  - apps
  resourceNames:
  - quick-db2
  resources:
  - petsets
  verbs:
  - get
- apiGroups:
  - kubedb.com
  resourceNames:
  - quick-db2
  resources:
  - db2s
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

KubeDB operator creates a ServiceAccount object `quick-db2` in same namespace as the DB2 object.

```bash
$ kubectl get serviceaccount -n demo quick-db2 -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-db2
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: db2s.kubedb.com
  name: quick-db2
  namespace: demo
```

This ServiceAccount is used in the PetSet created for the DB2 object.

### RoleBinding

KubeDB operator creates a RoleBinding object `quick-db2` in same namespace as the DB2 object.

```bash
$ kubectl get rolebinding -n demo quick-db2 -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-db2
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: db2s.kubedb.com
  name: quick-db2
  namespace: demo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: quick-db2
subjects:
- kind: ServiceAccount
  name: quick-db2
  namespace: demo
```

This object binds Role `quick-db2` with ServiceAccount `quick-db2`.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo db2/quick-db2 -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo db2/quick-db2

kubectl delete ns demo
```
