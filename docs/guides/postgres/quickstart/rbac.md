---
title: RBAC for PostgreSQL
menu:
  docs_{{ .version }}:
    identifier: pg-rbac-quickstart
    name: RBAC
    parent: pg-quickstart-postgres
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RBAC Permissions for Postgres

If RBAC is enabled in clusters, some PostgreSQL specific RBAC permissions are required. These permissions are required for Leader Election process of PostgreSQL clustering.

Here is the list of additional permissions required by PetSet of Postgres:

| Kubernetes Resource | Resource Names    | Permission required |
|---------------------|-------------------|---------------------|
| petsets        | `{postgres-name}` | get                 |
| pods                |                   | list, patch         |
| pods/exec           |                   | create              |
| Postgreses          |                   | get                 |
| configmaps          | `{postgres-name}` | get, update, create |

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/postgres](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a PostgreSQL database

Below is the Postgres object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: quick-postgres
  namespace: demo
spec:
  version: "13.13"
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

Create above Postgres object with following command

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/quickstart/quick-postgres.yaml
postgres.kubedb.com/quick-postgres created
```

When this Postgres object is created, KubeDB operator creates Role, ServiceAccount and RoleBinding with the matching PostgreSQL name and uses that ServiceAccount name in the corresponding PetSet.

Let's see what KubeDB operator has created for additional RBAC permission

### Role

KubeDB operator create a Role object `quick-postgres` in same namespace as Postgres object.

```yaml
$ kubectl get role -n demo quick-postgres -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  creationTimestamp: "2022-05-31T05:20:19Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-postgres
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: postgreses.kubedb.com
  name: quick-postgres
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Postgres
      name: quick-postgres
      uid: c118d264-85b7-4140-bc3f-d459c58c0523
  resourceVersion: "367334"
  uid: e72f25a5-5945-4687-9e8f-8af33c1a6b13
rules:
  - apiGroups:
      - apps
    resourceNames:
      - quick-postgres
    resources:
      - petsets
    verbs:
      - get
  - apiGroups:
      - kubedb.com
    resourceNames:
      - quick-postgres
    resources:
      - postgreses
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
  - apiGroups:
      - policy
    resourceNames:
      - postgres-db
    resources:
      - podsecuritypolicies
    verbs:
      - use

```

### ServiceAccount

KubeDB operator create a ServiceAccount object `quick-postgres` in same namespace as Postgres object.

```yaml
$ kubectl get serviceaccount -n demo quick-postgres -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2022-05-31T05:20:19Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-postgres
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: postgreses.kubedb.com
  name: quick-postgres
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Postgres
      name: quick-postgres
      uid: c118d264-85b7-4140-bc3f-d459c58c0523
  resourceVersion: "367333"
  uid: 1a1db587-d5a6-4cfc-aa82-dc960b7e1f28

```

This ServiceAccount is used in PetSet created for Postgres object.

### RoleBinding

KubeDB operator create a RoleBinding object `quick-postgres` in same namespace as Postgres object.

```yaml
$ kubectl get rolebinding -n demo quick-postgres -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "2022-05-31T05:20:19Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-postgres
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: postgreses.kubedb.com
  name: quick-postgres
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Postgres
      name: quick-postgres
      uid: c118d264-85b7-4140-bc3f-d459c58c0523
  resourceVersion: "367335"
  uid: 1fc9f872-8adc-4940-b93d-18f70bec38d5
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: quick-postgres
subjects:
  - kind: ServiceAccount
    name: quick-postgres
    namespace: demo

```

This  object binds Role `quick-postgres` with ServiceAccount `quick-postgres`.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo pg/quick-postgres -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo pg/quick-postgres

kubectl delete ns demo
```
