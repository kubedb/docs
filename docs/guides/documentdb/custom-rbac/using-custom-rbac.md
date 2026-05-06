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

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a DocumentDB instance. This tutorial will show you how to use KubeDB to run DocumentDB instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for DocumentDB. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in DocumentDB CRD. If this field is left empty, the KubeDB operator will create a service account name matching DocumentDB crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a DocumentDB instance named `quick-docdb` to provide the bare minimum access permissions.

## Custom RBAC for DocumentDB

At first, let's create a `Service Account` in `demo` namespace.

```bash
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

It should create a service account.

```yaml
$ kubectl get serviceaccount -n demo my-custom-serviceaccount -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2019-05-30T04:23:39Z"
  name: my-custom-serviceaccount
  namespace: demo
  resourceVersion: "21657"
  selfLink: /api/v1/namespaces/demo/serviceaccounts/my-custom-serviceaccount
  uid: b2ec2b05-8292-11e9-8d10-080027a8b217
secrets:
- name: my-custom-serviceaccount-token-t8zxd
```

Now, we need to create a role that has necessary access permissions for the DocumentDB Database named `quick-docdb`.

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

Please note that resourceName `quick-docdb` is unique to `quick-docdb` DocumentDB instance. Another database `quick-docdb-2`, for example, will require the resourceName to be `quick-docdb-2`.

Now create a `RoleBinding` to bind this `Role` with the already created service account.

```bash
$ kubectl create rolebinding my-custom-rolebinding --role=my-custom-role --serviceaccount=demo:my-custom-serviceaccount --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created
```

It should bind `my-custom-role` and `my-custom-serviceaccount` successfully.

```yaml
$ kubectl get rolebinding -n demo my-custom-rolebinding -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "2019-05-30T04:54:56Z"
  name: my-custom-rolebinding
  namespace: demo
  resourceVersion: "23944"
  selfLink: /apis/rbac.authorization.k8s.io/v1/namespaces/demo/rolebindings/my-custom-rolebinding
  uid: 123afc02-8297-11e9-8d10-080027a8b217
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: my-custom-role
subjects:
- kind: ServiceAccount
  name: my-custom-serviceaccount
  namespace: demo
```

Now, create a DocumentDB CRD specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/custom-rbac/docdb-custom-db.yaml
documentdb.kubedb.com/quick-docdb created
```

Below is the YAML for the DocumentDB crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: quick-docdb
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-docdb
spec:
  version: "pg17-0.109.0"
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
        storage: 5Gi
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we should see that a pod with the name `quick-docdb-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo quick-docdb-0
NAME             READY   STATUS    RESTARTS   AGE
quick-docdb-0    1/1     Running   0          3m
```

Check the pod's log to see if the database is ready
```bash
```bash
$ kubectl logs -f -n demo second-docdb-0
```

Once we see `database system is ready to accept connections` in the log, the database is ready.

## Reusing Service Account

An existing service account can be reused in another DocumentDB Database. However, users need to create a new Role specific to that DocumentDB and bind it to the existing service account so that all the necessary access permissions are available to run the new DocumentDB Database.

For example, to reuse `my-custom-serviceaccount` in a new Database `second-docdb`, create a role that has all the necessary access permissions for this DocumentDB Database.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/custom-rbac/docdb-custom-role-two.yaml
role.rbac.authorization.k8s.io/my-custom-role-two created
```

Below is the YAML for the Role we just created.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: my-custom-role-two
  namespace: demo
rules:
- apiGroups:
  - apps
  resourceNames:
  - second-docdb
  resources:
  - petsets
  verbs:
  - get
- apiGroups:
  - kubedb.com
  resourceNames:
  - second-docdb
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

Now create a `RoleBinding` to bind `my-custom-role-two` with the already created `my-custom-serviceaccount`.

```bash
$ kubectl create rolebinding my-custom-rolebinding-two --role=my-custom-role-two --serviceaccount=demo:my-custom-serviceaccount --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding-two created
```

Now, create DocumentDB CRD `second-docdb` using the existing service account name `my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/custom-rbac/docdb-custom-db-two.yaml
documentdb.kubedb.com/second-docdb created
```

Below is the YAML for the DocumentDB crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: second-docdb
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: second-docdb
spec:
  version: "pg17-0.109.0"
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
        storage: 5Gi
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we should see that a pod with the name `second-docdb-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo second-docdb-0
NAME             READY   STATUS    RESTARTS   AGE
second-docdb-0   1/1     Running   0          3m
```

Check the pod's log to see if the database is ready

```bash
$ kubectl logs -f -n demo second-docdb-0
```

`database system is ready to accept connections` in the log signifies that the database is running successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo documentdb/quick-docdb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo documentdb/quick-docdb

kubectl patch -n demo documentdb/second-docdb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo documentdb/second-docdb

kubectl delete -n demo role my-custom-role
kubectl delete -n demo role my-custom-role-two

kubectl delete -n demo rolebinding my-custom-rolebinding
kubectl delete -n demo rolebinding my-custom-rolebinding-two

kubectl delete sa -n demo my-custom-serviceaccount

kubectl delete ns demo
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- Learn about initializing [DocumentDB with Script](/docs/guides/documentdb/initialization/script_source.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

