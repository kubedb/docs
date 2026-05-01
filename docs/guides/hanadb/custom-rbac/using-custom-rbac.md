---
title: Run HanaDB with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: hanadb-custom-rbac-quickstart
    name: Custom RBAC
    parent: hanadb-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC Resources

KubeDB supports finer user control over role based access permissions provided to a HanaDB instance. This tutorial will show you how to use KubeDB to run HanaDB instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for HanaDB. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in HanaDB CRD. If this field is left empty, the KubeDB operator will create a service account name matching the HanaDB CRD name.

If a service account name is given with an existing service account, the KubeDB operator will use that existing service account. Users are responsible for providing necessary access permissions manually.

## Custom RBAC for HanaDB

At first, let's create a `Service Account` in `demo` namespace.

```bash
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

Now, we need to create a role that has necessary access permissions for the HanaDB database named `quick-hanadb`.

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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/custom-rbac/hanadb-custom-role.yaml
role.rbac.authorization.k8s.io/my-custom-role created
```

Now create a `RoleBinding` to bind this `Role` with the already created service account.

```bash
$ kubectl create rolebinding my-custom-rolebinding \
  --role=my-custom-role \
  --serviceaccount=demo:my-custom-serviceaccount \
  --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created
```

Now, create a HanaDB CRD specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

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
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/custom-rbac/hanadb-custom-db.yaml
hanadb.kubedb.com/quick-hanadb created
```

Check that the pod is running:

```bash
$ kubectl get pod -n demo quick-hanadb-0
NAME              READY   STATUS    RESTARTS   AGE
quick-hanadb-0    1/1     Running   0          5m
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/quick-hanadb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/quick-hanadb

kubectl delete -n demo serviceaccount my-custom-serviceaccount
kubectl delete -n demo role my-custom-role
kubectl delete -n demo rolebinding my-custom-rolebinding
kubectl delete ns demo
```
