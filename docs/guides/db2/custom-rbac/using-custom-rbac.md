---
title: Run DB2 with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: db2-custom-rbac-quickstart
    name: Custom RBAC
    parent: db2-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a DB2 instance. This tutorial will show you how to use KubeDB to run DB2 instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/db2](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/db2) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for DB2. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in DB2 CRD. If this field is left empty, the KubeDB operator will create a service account name matching DB2 crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a DB2 instance named `quick-db2` to provide the bare minimum access permissions.

## Custom RBAC for DB2

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

Now, we need to create a role that has necessary access permissions for the DB2 Database named `quick-db2`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/db2/custom-rbac/db2-custom-role.yaml
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

Please note that resourceName `quick-db2` is unique to `quick-db2` DB2 instance. Another database `quick-db2-2`, for example, will require the resourceName to be `quick-db2-2`.

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

Now, create a DB2 CRD specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/db2/custom-rbac/db2-custom-db.yaml
db2.kubedb.com/quick-db2 created
```

Below is the YAML for the DB2 crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DB2
metadata:
  name: quick-db2
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-db2
spec:
  version: "11.5.9"
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
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we should see that a pod with the name `quick-db2-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo quick-db2-0
NAME           READY   STATUS    RESTARTS   AGE
quick-db2-0    1/1     Running   0          3m
```

Check the pod's log to see if the database is ready

Once we see `database system is ready to accept connections` in the log, the database is ready.

## Reusing Service Account

An existing service account can be reused in another DB2 Database. However, users need to create a new Role specific to that DB2 and bind it to the existing service account so that all the necessary access permissions are available to run the new DB2 Database.

For example, to reuse `my-custom-serviceaccount` in a new Database `second-db2`, create a role that has all the necessary access permissions for this DB2 Database.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/db2/custom-rbac/db2-custom-role-two.yaml
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
  - second-db2
  resources:
  - petsets
  verbs:
  - get
- apiGroups:
  - kubedb.com
  resourceNames:
  - second-db2
  resources:
  - db2s
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

Now, create DB2 CRD `second-db2` using the existing service account name `my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/db2/custom-rbac/db2-custom-db-two.yaml
db2.kubedb.com/second-db2 created
```

Below is the YAML for the DB2 crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DB2
metadata:
  name: second-db2
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: second-db2
spec:
  version: "11.5.9"
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
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we should see that a pod with the name `second-db2-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo second-db2-0
NAME           READY   STATUS    RESTARTS   AGE
second-db2-0   1/1     Running   0          3m
```

Check the pod's log to see if the database is ready

`database system is ready to accept connections` in the log signifies that the database is running successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo db2/quick-db2 -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo db2/quick-db2

kubectl patch -n demo db2/second-db2 -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo db2/second-db2

kubectl delete -n demo role my-custom-role
kubectl delete -n demo role my-custom-role-two

kubectl delete -n demo rolebinding my-custom-rolebinding
kubectl delete -n demo rolebinding my-custom-rolebinding-two

kubectl delete sa -n demo my-custom-serviceaccount

kubectl delete ns demo
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- Learn about initializing [DB2 with Script](/docs/guides/db2/initialization/script_source.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

