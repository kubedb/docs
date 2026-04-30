---
title: Run DocumentDB with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: documentdb-custom-rbac-quickstart
    name: Custom RBAC
    parent: documentdb-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC Resources

KubeDB supports finer user control over role based access permissions provided to a DocumentDB instance. This tutorial will show you how to use KubeDB to run a DocumentDB instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for DocumentDB. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in DocumentDB CRD. If this field is left empty, the KubeDB operator will create a service account name matching the DocumentDB CRD name.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Users are responsible for providing necessary access permissions manually.

## Custom RBAC for DocumentDB

At first, let's create a `Service Account` in `demo` namespace.

```bash
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

Now, we need to create a role that has necessary access permissions for the DocumentDB database named `quick-docdb`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/custom-rbac/docdb-custom-role.yaml
role.rbac.authorization.k8s.io/my-custom-role created
```

Below is the YAML for the Role we just created.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-custom-role
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

Now create a `RoleBinding` to bind this `Role` with the already created service account.

```bash
$ kubectl create rolebinding my-custom-rolebinding \
  --role=my-custom-role \
  --serviceaccount=demo:my-custom-serviceaccount \
  --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created
```

Now, create a DocumentDB CRD specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: quick-docdb
  namespace: demo
spec:
  version: "5.0.6"
  storageType: Durable
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/custom-rbac/docdb-custom-db.yaml
documentdb.kubedb.com/quick-docdb created
```

Check that the pod is running:

```bash
$ kubectl get pod -n demo quick-docdb-0
NAME             READY   STATUS    RESTARTS   AGE
quick-docdb-0    1/1     Running   0          3m
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo documentdb/quick-docdb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo documentdb/quick-docdb

kubectl delete -n demo serviceaccount my-custom-serviceaccount
kubectl delete -n demo role my-custom-role
kubectl delete -n demo rolebinding my-custom-rolebinding
kubectl delete ns demo
```
