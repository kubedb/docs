---
title: Run MongoDB with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: mg-custom-rbac-quickstart
    name: Custom RBAC
    parent: mg-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a MongoDB instance. This tutorial will show you how to use KubeDB to run MongoDB instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for MongoDB. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in MongoDB crd.   If this field is left empty, the KubeDB operator will create a service account name matching MongoDB crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a MongoDB instance named `quick-mongodb` to provide the bare minimum access permissions.

## Custom RBAC for MongoDB

At first, let's create a `Service Acoount` in `demo` namespace.

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
  selfLink: /api/v1/namespaces/demo/serviceaccounts/myserviceaccount
  uid: b2ec2b05-8292-11e9-8d10-080027a8b217
secrets:
- name: myserviceaccount-token-t8zxd
```

Now, we need to create a role that has necessary access permissions for the MongoDB instance named `quick-mongodb`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/custom-rbac/mg-custom-role.yaml
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
  - policy
  resourceNames:
  - mongodb-db
  resources:
  - podsecuritypolicies
  verbs:
  - use
```

This permission is required for MongoDB pods running on PSP enabled clusters.

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
  creationTimestamp: "2019-05-30T04:33:39Z"
  name: my-custom-rolebinding
  namespace: demo
  resourceVersion: "1405"
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

Now, create a MongoDB crd specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/custom-rbac/mg-custom-db.yaml
mongodb.kubedb.com/quick-mongodb created
```

Below is the YAML for the MongoDB crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: quick-mongodb
  namespace: demo
spec:
  version: "4.4.26"
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
  deletionPolicy: DoNotTerminate
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, deployment, petsets, services, secret etc. If everything goes well, we should see that a pod with the name `quick-mongodb-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo quick-mongodb-0
NAME              READY   STATUS    RESTARTS   AGE
quick-mongodb-0   1/1     Running   0          28s
```

Check the pod's log to see if the database is ready

```bash
$ kubectl logs -f -n demo quick-mongodb-0
about to fork child process, waiting until server is ready for connections.
forked process: 17
2019-06-10T08:56:45.259+0000 I CONTROL  [main] ***** SERVER RESTARTED *****
2019-06-10T08:56:45.263+0000 I CONTROL  [initandlisten] MongoDB starting : pid=17 port=27017 dbpath=/data/db 64-bit host=quick-mongodb-0
...
...
MongoDB init process complete; ready for start up.
...
..
2019-06-10T08:56:49.287+0000 I NETWORK  [thread1] waiting for connections on port 27017
2019-06-10T08:56:57.179+0000 I NETWORK  [thread1] connection accepted from 127.0.0.1:39214 #1 (1 connection now open)
```

Once we see `connection accepted` in the log, the database is ready.

## Reusing Service Account

An existing service account can be reused in another MongoDB instance. No new access permission is required to run the new MongoDB instance.

Now, create MongoDB crd `minute-mongodb` using the existing service account name `my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/custom-rbac/mg-custom-db-two.yaml
mongodb.kubedb.com/quick-mongodb created
```

Below is the YAML for the MongoDB crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: minute-mongodb
  namespace: demo
spec:
  version: "4.4.26"
  podTemplate:
      spec:
        serviceAccountName: my-custom-serviceaccount
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: DoNotTerminate
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, petset, deployment, services, secret etc. If everything goes well, we should see that a pod with the name `minute-mongodb-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo minute-mongodb-0
NAME                READY   STATUS    RESTARTS   AGE
minute-mongodb-0    1/1     Running   0          50s
```

Check the pod's log to see if the database is ready

```bash
$ kubectl logs -f -n demo minute-mongodb-0
about to fork child process, waiting until server is ready for connections.
forked process: 17
2019-06-10T08:56:45.259+0000 I CONTROL  [main] ***** SERVER RESTARTED *****
2019-06-10T08:56:45.263+0000 I CONTROL  [initandlisten] MongoDB starting : pid=17 port=27017 dbpath=/data/db 64-bit host=quick-mongodb-0
...
...
MongoDB init process complete; ready for start up.
...
..
2019-06-10T08:56:49.287+0000 I NETWORK  [thread1] waiting for connections on port 27017
2019-06-10T08:56:57.179+0000 I NETWORK  [thread1] connection accepted from 127.0.0.1:39214 #1 (1 connection now open)
```

`connection accepted` in the log signifies that the database is running successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mg/quick-mongodb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mg/quick-mongodb

kubectl patch -n demo mg/minute-mongodb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mg/minute-mongodb

kubectl delete -n demo role my-custom-role
kubectl delete -n demo rolebinding my-custom-rolebinding

kubectl delete sa -n demo my-custom-serviceaccount

kubectl delete ns demo
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart MongoDB](/docs/guides/mongodb/quickstart/quickstart.md) with KubeDB Operator.
- [Backup and Restore](/docs/guides/mongodb/backup/overview/index.md) MongoDB instances using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB instance with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB instance with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Use [kubedb cli](/docs/guides/mongodb/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

