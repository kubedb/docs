---
title: RBAC for Milvus
menu:
  docs_{{ .version }}:
    identifier: milvus-rbac-quickstart
    name: RBAC
    parent: milvus-quickstart
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RBAC Permissions for Milvus

When RBAC is enabled in your cluster, KubeDB automatically creates the necessary Role, ServiceAccount, and RoleBinding for each Milvus instance. This tutorial explains what permissions are granted and how to verify them.

Here is the list of additional permissions required by the PetSet of Milvus:

| Kubernetes Resource | Resource Names    | Permission required |
|---------------------|-------------------|---------------------|
| petsets             | `{milvus-name}`   | get                 |
| pods                |                   | list, patch         |
| pods/exec           |                   | create              |
| milvuses            |                   | get                 |
| configmaps          | `{milvus-name}`   | get, update, create |

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Create a Milvus Database

Below is the Milvus object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: quick-milvus
  namespace: demo
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: my-release-minio
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

Create the above Milvus object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/milvus/quickstart/quick-milvus.yaml
milvus.kubedb.com/quick-milvus created
```

When this Milvus object is created, KubeDB operator creates Role, ServiceAccount and RoleBinding with the matching Milvus name.

### Role

```bash
$ kubectl get role -n demo quick-milvus -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-milvus
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: milvuses.kubedb.com
  name: quick-milvus
  namespace: demo
rules:
- apiGroups:
  - apps
  resourceNames:
  - quick-milvus
  resources:
  - petsets
  verbs:
  - get
- apiGroups:
  - kubedb.com
  resourceNames:
  - quick-milvus
  resources:
  - milvuses
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
$ kubectl get serviceaccount -n demo quick-milvus -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-milvus
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: milvuses.kubedb.com
  name: quick-milvus
  namespace: demo
```

### RoleBinding

```bash
$ kubectl get rolebinding -n demo quick-milvus -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-milvus
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: milvuses.kubedb.com
  name: quick-milvus
  namespace: demo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: quick-milvus
subjects:
- kind: ServiceAccount
  name: quick-milvus
  namespace: demo
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo milvus/quick-milvus -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo milvus/quick-milvus

kubectl delete ns demo
```
