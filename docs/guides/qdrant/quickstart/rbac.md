---
title: Qdrant RBAC
menu:
  docs_{{ .version }}:
    identifier: qdrant-quickstart-rbac
    name: RBAC
    parent: qdrant-quickstart
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RBAC Permissions for Qdrant

If RBAC is enabled in clusters, some Qdrant-specific RBAC permissions are required. These permissions are required for the KubeDB operator to manage Qdrant pods properly.

Here is the list of additional permissions required by the StatefulSet of Qdrant:

| Kubernetes Resource | Resource Names  | Permission required    |
|---------------------|-----------------|------------------------|
| statefulsets        | `{qdrant-name}` | get                    |
| pods                |                 | list, patch            |
| pods/exec           |                 | create                 |
| qdrants             |                 | get                    |
| configmaps          | `{qdrant-name}` | get, update, create    |
| secrets             |                 | get, list              |

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/qdrant/quickstart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a Qdrant Database

Below is the `Qdrant` object created in this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-rbac
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/qdrant-rbac.yaml
qdrant.kubedb.com/qdrant-rbac created
```

When this `Qdrant` object is created, KubeDB operator creates a Role, ServiceAccount, and RoleBinding with the matching Qdrant name and uses that ServiceAccount in the corresponding StatefulSet.

Let's see what KubeDB operator has created for additional RBAC permissions.

### Role

KubeDB operator creates a Role object `qdrant-rbac` in the same namespace as the Qdrant object:

```yaml
$ kubectl get role -n demo qdrant-rbac -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: qdrant-rbac
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: qdrants.kubedb.com
  name: qdrant-rbac
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Qdrant
    name: qdrant-rbac
rules:
- apiGroups:
  - apps
  resourceNames:
  - qdrant-rbac
  resources:
  - statefulsets
  verbs:
  - get
- apiGroups:
  - kubedb.com
  resourceNames:
  - qdrant-rbac
  resources:
  - qdrants
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
  - delete
- apiGroups:
  - ""
  resources:
  - pods/exec
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
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

KubeDB operator creates a ServiceAccount object `qdrant-rbac` in the same namespace as the Qdrant object:

```yaml
$ kubectl get serviceaccount -n demo qdrant-rbac -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: qdrant-rbac
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: qdrants.kubedb.com
  name: qdrant-rbac
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Qdrant
    name: qdrant-rbac
```

This ServiceAccount is used in the StatefulSet created for the Qdrant object.

### RoleBinding

KubeDB operator creates a RoleBinding object `qdrant-rbac` in the same namespace as the Qdrant object:

```yaml
$ kubectl get rolebinding -n demo qdrant-rbac -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: qdrant-rbac
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: qdrants.kubedb.com
  name: qdrant-rbac
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Qdrant
    name: qdrant-rbac
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: qdrant-rbac
subjects:
- kind: ServiceAccount
  name: qdrant-rbac
  namespace: demo
```

This object binds Role `qdrant-rbac` with ServiceAccount `qdrant-rbac`.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo qdrant/qdrant-rbac -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo qdrant/qdrant-rbac
kubectl delete ns demo
```