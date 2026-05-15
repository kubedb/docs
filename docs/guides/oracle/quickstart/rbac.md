---
title: Oracle RBAC
menu:
  docs_{{ .version }}:
    identifier: oracle-sample-rbac
    name: RBAC
    parent: oracle-sample
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RBAC Permissions for Oracle

If RBAC is enabled in clusters, some Oracle-specific RBAC permissions are required. These permissions are required for the KubeDB operator to manage Oracle pods properly.

Here is the list of additional permissions required by the StatefulSet of Oracle:

| Kubernetes Resource | Resource Names  | Permission required    |
|---------------------|-----------------|------------------------|
| statefulsets        | `{oracle-name}` | get                    |
| pods                |                 | list, patch            |
| pods/exec           |                 | create                 |
| oracles             |                 | get                    |
| configmaps          | `{oracle-name}` | get, update, create    |
| secrets             |                 | get, list              |

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/oracle/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/quickstart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a Oracle Database

Below is the `Oracle` object created in this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sample
  namespace: demo
spec:
  version: "21.3.0"
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/quickstart/oracle-sample.yaml
oracle.kubedb.com/oracle-sample created
```

When this `Oracle` object is created, KubeDB operator creates a Role, ServiceAccount, and RoleBinding with the matching Oracle name and uses that ServiceAccount in the corresponding StatefulSet.

Let's see what KubeDB operator has created for additional RBAC permissions.

### Role

KubeDB operator creates a Role object `oracle-sample` in the same namespace as the Oracle object:

```yaml
$ kubectl get role -n demo oracle-sample -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: oracle-sample
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: oracles.kubedb.com
  name: oracle-sample
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Oracle
    name: oracle-sample
rules:
- apiGroups:
  - apps
  resourceNames:
  - oracle-sample
  resources:
  - statefulsets
  verbs:
  - get
- apiGroups:
  - kubedb.com
  resourceNames:
  - oracle-sample
  resources:
  - oracles
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

KubeDB operator creates a ServiceAccount object `oracle-sample` in the same namespace as the Oracle object:

```yaml
$ kubectl get serviceaccount -n demo oracle-sample -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: oracle-sample
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: oracles.kubedb.com
  name: oracle-sample
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Oracle
    name: oracle-sample
```

This ServiceAccount is used in the StatefulSet created for the Oracle object.

### RoleBinding

KubeDB operator creates a RoleBinding object `oracle-sample` in the same namespace as the Oracle object:

```yaml
$ kubectl get rolebinding -n demo oracle-sample -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: oracle-sample
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: oracles.kubedb.com
  name: oracle-sample
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Oracle
    name: oracle-sample
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: oracle-sample
subjects:
- kind: ServiceAccount
  name: oracle-sample
  namespace: demo
```

This object binds Role `oracle-sample` with ServiceAccount `oracle-sample`.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo oracle/oracle-sample -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo oracle/oracle-sample
kubectl delete ns demo
```