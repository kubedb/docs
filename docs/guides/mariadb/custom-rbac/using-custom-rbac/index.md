---
title: Run MariaDB with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-customrbac-usingcustomrbac
    name: Custom RBAC
    parent: guides-mariadb-customrbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a MariaDB instance. This tutorial will show you how to use KubeDB to run MariaDB instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mariadb/custom-rbac/using-custom-rbac/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for MariaDB. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in MariaDB crd.   If this field is left empty, the KubeDB operator will create a service account name matching MariaDB crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a MariaDB instance named `quick-postges` to provide the bare minimum access permissions.

## Custom RBAC for MariaDB

At first, let's create a `Service Acoount` in `demo` namespace.

```bash
$ kubectl create serviceaccount -n demo md-custom-serviceaccount
serviceaccount/md-custom-serviceaccount created
```

It should create a service account.

```yaml
$ kubectl get serviceaccount -n demo md-custom-serviceaccount -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2021-03-18T04:38:59Z"
  name: md-custom-serviceaccount
  namespace: demo
  resourceVersion: "84669"
  selfLink: /api/v1/namespaces/demo/serviceaccounts/md-custom-serviceaccount
  uid: 788bd6c6-3eae-4797-b6ca-5722ef64c9dc
secrets:
- name: md-custom-serviceaccount-token-jnhvd
```

Now, we need to create a role that has necessary access permissions for the MariaDB instance named `sample-mariadb`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/custom-rbac/using-custom-rbac/examples/md-custom-role.yaml
role.rbac.authorization.k8s.io/md-custom-role created
```

Below is the YAML for the Role we just created.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: md-custom-role
  namespace: demo
rules:
- apiGroups:
  - policy
  resourceNames:
  - maria-db
  resources:
  - podsecuritypolicies
  verbs:
  - use
```

This permission is required for MariaDB pods running on PSP enabled clusters.

Now create a `RoleBinding` to bind this `Role` with the already created service account.

```bash
$ kubectl create rolebinding md-custom-rolebinding --role=md-custom-role --serviceaccount=demo:md-custom-serviceaccount --namespace=demo
rolebinding.rbac.authorization.k8s.io/md-custom-rolebinding created
```

It should bind `md-custom-role` and `md-custom-serviceaccount` successfully.

SO, All required resources for RBAC are created.

```bash
$ kubectl get serviceaccount,role,rolebindings -n demo
NAME                                      SECRETS   AGE
serviceaccount/default                    1         38m
serviceaccount/md-custom-serviceaccount   1         36m

NAME                                            CREATED AT
role.rbac.authorization.k8s.io/md-custom-role   2021-03-18T05:13:27Z

NAME                                                          ROLE                  AGE
rolebinding.rbac.authorization.k8s.io/md-custom-rolebinding   Role/md-custom-role   79s
```

Now, create a MariaDB crd specifying `spec.podTemplate.spec.serviceAccountName` field to `md-custom-serviceaccount`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/custom-rbac/using-custom-rbac/examples/md-custom-db.yaml
mariadb.kubedb.com/sample-mariadb created
```

Below is the YAML for the MariaDB crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.23"
  storageType: Durable
  podTemplate:
    spec:
      serviceAccountName: md-custom-serviceaccount
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, StatefulSet, services, secret etc. If everything goes well, we should see that a pod with the name `sample-mariadb-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo sample-mariadb-0
NAME            READY   STATUS    RESTARTS   AGE
sample-mariadb-0   1/1     Running   0          2m44s
```

Check the pod's log to see if the database is ready

```bash
$ kubectl logs -f -n demo sample-mariadb-0
2021-03-18 05:35:13+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 1:10.5.23+maria~focal started.
2021-03-18 05:35:13+00:00 [Note] [Entrypoint]: Switching to dedicated user 'mysql'
2021-03-18 05:35:13+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 1:10.5.23+maria~focal started.
2021-03-18 05:35:14+00:00 [Note] [Entrypoint]: Initializing database files
...
2021-03-18  5:35:22 0 [Note] Reading of all Master_info entries succeeded
2021-03-18  5:35:22 0 [Note] Added new Master_info '' to hash table
2021-03-18  5:35:22 0 [Note] mysqld: ready for connections.
Version: '10.5.23-MariaDB-1:10.5.23+maria~focal'  socket: '/run/mysqld/mysqld.sock'  port: 3306  mariadb.org binary distribution
```

Once we see `mysqld: ready for connections.` in the log, the database is ready.

## Reusing Service Account

An existing service account can be reused in another MariaDB instance. No new access permission is required to run the new MariaDB instance.

Now, create MariaDB crd `another-mariadb` using the existing service account name `md-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/custom-rbac/using-custom-rbac/examples/md-custom-db-2.yaml
mariadb.kubedb.com/another-mariadb created
```

Below is the YAML for the MariaDB crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: another-mariadb
  namespace: demo
spec:
  version: "10.5.23"
  storageType: Durable
  podTemplate:
    spec:
      serviceAccountName: md-custom-serviceaccount
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we should see that a pod with the name `another-mariadb` has been created.

Check that the statefulset's pod is running

```bash
$ kubectl get pod -n demo another-mariadb-0
NAME                READY   STATUS    RESTARTS   AGE
another-mariadb-0   1/1     Running   0          37s
```

Check the pod's log to see if the database is ready

```bash
...
$ kubectl logs -f -n demo another-mariadb-0
2021-03-18 05:39:50+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 1:10.5.23+maria~focal started.
2021-03-18 05:39:50+00:00 [Note] [Entrypoint]: Switching to dedicated user 'mysql'
2021-03-18 05:39:50+00:00 [Note] [Entrypoint]: Entrypoint script for MySQL Server 1:10.5.23+maria~focal started.
2021-03-18 05:39:50+00:00 [Note] [Entrypoint]: Initializing database files
...
2021-03-18  5:39:59 0 [Note] mysqld: ready for connections.
Version: '10.5.23-MariaDB-1:10.5.23+maria~focal'  socket: '/run/mysqld/mysqld.sock'  port: 3306  mariadb.org binary distribution
```

`mysqld: ready for connections.` in the log signifies that the database is running successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo sample-mariadb
mariadb.kubedb.com "sample-mariadb" deleted
$ kubectl delete mariadb -n demo another-mariadb
mariadb.kubedb.com "another-mariadb" deleted
$ kubectl delete -n demo role md-custom-role
role.rbac.authorization.k8s.io "md-custom-role" deleted
$ kubectl delete -n demo rolebinding md-custom-rolebinding
rolebinding.rbac.authorization.k8s.io "md-custom-rolebinding" deleted
$ kubectl delete sa -n demo md-custom-serviceaccount
serviceaccount "md-custom-serviceaccount" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```


