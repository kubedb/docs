---
title: Run PostgreSQL with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: pg-custom-rbac-quickstart
    name: Custom RBAC
    parent: pg-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a PostgreSQL instance. This tutorial will show you how to use KubeDB to run PostgreSQL instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/postgres](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for PostgreSQL. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in Postgres CRD. If this field is left empty, the KubeDB operator will create a service account name matching Postgres crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a PostgreSQL instance named `quick-postges` to provide the bare minimum access permissions.

## Custom RBAC for PostgreSQL

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

Now, we need to create a role that has necessary access permissions for the PostgreSQl Database named `quick-postgres`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/custom-rbac/pg-custom-role.yaml
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
  - quick-postgres
  resources:
  - statefulsets
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
  - configmaps
  verbs:
  - create
- apiGroups:
  - ""
  resourceNames:
  - quick-postgres-leader-lock
  resources:
  - configmaps
  verbs:
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

Please note that resourceNames `quick-postgres` and `quick-postgres-leader-lock` are unique to `quick-postgres` PostgreSQL instance. Another database `quick-postgres-2`, for exmaple, will require these resourceNames to be `quick-postgres-2`, and `quick-postgres-2-leader-lock`.

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

Now, create a Postgres CRD specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/custom-rbac/pg-custom-db.yaml
postgres.kubedb.com/quick-postgres created
```

Below is the YAML for the Postgres crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: quick-postgres
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-postgres
spec:
  version: "10.2-v5"
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
        storage: 50Mi

```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we should see that a pod with the name `quick-postgres-0` has been created.

Check that the statefulset's pod is running

```bash
$ kubectl get pod -n demo quick-postgres-0
NAME                READY     STATUS    RESTARTS   AGE
quick-postgres-0   1/1       Running   0          14m
```

Check the pod's log to see if the database is ready

```bash
$ kubectl logs -f -n demo quick-postgres-0
I0705 12:05:51.697190       1 logs.go:19] FLAG: --alsologtostderr="false"
I0705 12:05:51.717485       1 logs.go:19] FLAG: --enable-analytics="true"
I0705 12:05:51.717543       1 logs.go:19] FLAG: --help="false"
I0705 12:05:51.717558       1 logs.go:19] FLAG: --log_backtrace_at=":0"
I0705 12:05:51.717566       1 logs.go:19] FLAG: --log_dir=""
I0705 12:05:51.717573       1 logs.go:19] FLAG: --logtostderr="false"
I0705 12:05:51.717581       1 logs.go:19] FLAG: --stderrthreshold="0"
I0705 12:05:51.717589       1 logs.go:19] FLAG: --v="0"
I0705 12:05:51.717597       1 logs.go:19] FLAG: --vmodule=""
We want "quick-postgres-0" as our leader
I0705 12:05:52.753464       1 leaderelection.go:175] attempting to acquire leader lease  demo/quick-postgres-leader-lock...
I0705 12:05:52.822093       1 leaderelection.go:184] successfully acquired lease demo/quick-postgres-leader-lock
Got leadership, now do your jobs
Running as Primary
sh: locale: not found

WARNING: enabling "trust" authentication for local connections
You can change this by editing pg_hba.conf or using the option -A, or
--auth-local and --auth-host, the next time you run initdb.
ALTER ROLE
/scripts/primary/start.sh: ignoring /var/initdb/*

LOG:  database system was shut down at 2018-07-05 12:07:51 UTC
LOG:  MultiXact member wraparound protections are now enabled
LOG:  database system is ready to accept connections
LOG:  autovacuum launcher started
```

Once we see `LOG: database system is ready to accept connections` in the log, the database is ready.

## Reusing Service Account

An existing service account can be reused in another Postgres Database. However, users need to create a new Role specific to that Postgres and bind it to the existing service account so that all the necessary access permissions are available to run the new Postgres Database.

For example, to reuse `my-custom-serviceaccount` in a new Database `minute-postgres`, create a role that has all the necessary access permissions for this PostgreSQl Database.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/custom-rbac/pg-custom-role-two.yaml
role.rbac.authorization.k8s.io/my-custom-role created
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
  - miniute-postgres
  resources:
  - statefulsets
  verbs:
  - get
- apiGroups:
  - ""
  resourceNames:
  - miniute-postgres-leader-lock
  resources:
  - configmaps
  verbs:
  - get
  - update
```

Now create a `RoleBinding` to bind `my-custom-role-two` with the already created `my-custom-serviceaccount`.

```bash
$ kubectl create rolebinding my-custom-rolebinding-two --role=my-custom-role-two --serviceaccount=demo:my-custom-serviceaccount --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding-two created

```

Now, create Postgres CRD `minute-postgres` using the existing service account name `my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/custom-rbac/pg-custom-db-two.yaml
postgres.kubedb.com/quick-postgres created
```

Below is the YAML for the Postgres crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: minute-postgres
  namespace: demo
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-postgres
spec:
  version: "10.2-v5"
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
        storage: 50Mi

```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we should see that a pod with the name `minute-postgres-0` has been created.

Check that the statefulset's pod is running

```bash
$ kubectl get pod -n demo minute-postgres-0
NAME                READY     STATUS    RESTARTS   AGE
minute-postgres-0   1/1       Running   0          14m
```

Check the pod's log to see if the database is ready

```bash
$ kubectl logs -f -n demo minute-postgres-0
I0705 12:05:51.697190       1 logs.go:19] FLAG: --alsologtostderr="false"
I0705 12:05:51.717485       1 logs.go:19] FLAG: --enable-analytics="true"
I0705 12:05:51.717543       1 logs.go:19] FLAG: --help="false"
I0705 12:05:51.717558       1 logs.go:19] FLAG: --log_backtrace_at=":0"
I0705 12:05:51.717566       1 logs.go:19] FLAG: --log_dir=""
I0705 12:05:51.717573       1 logs.go:19] FLAG: --logtostderr="false"
I0705 12:05:51.717581       1 logs.go:19] FLAG: --stderrthreshold="0"
I0705 12:05:51.717589       1 logs.go:19] FLAG: --v="0"
I0705 12:05:51.717597       1 logs.go:19] FLAG: --vmodule=""
We want "minute-postgres-0" as our leader
I0705 12:05:52.753464       1 leaderelection.go:175] attempting to acquire leader lease  demo/minute-postgres-leader-lock...
I0705 12:05:52.822093       1 leaderelection.go:184] successfully acquired lease demo/minute-postgres-leader-lock
Got leadership, now do your jobs
Running as Primary
sh: locale: not found

WARNING: enabling "trust" authentication for local connections
You can change this by editing pg_hba.conf or using the option -A, or
--auth-local and --auth-host, the next time you run initdb.
ALTER ROLE
/scripts/primary/start.sh: ignoring /var/initdb/*

LOG:  database system was shut down at 2018-07-05 12:07:51 UTC
LOG:  MultiXact member wraparound protections are now enabled
LOG:  database system is ready to accept connections
LOG:  autovacuum launcher started
```

`LOG: database system is ready to accept connections` in the log signifies that the database is running successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo pg/quick-postgres -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo pg/quick-postgres

kubectl patch -n demo pg/minute-postgres -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo pg/minute-postgres

kubectl delete -n demo role my-custom-role
kubectl delete -n demo role my-custom-role-two

kubectl delete -n demo rolebinding my-custom-rolebinding
kubectl delete -n demo rolebinding my-custom-rolebinding-two

kubectl delete sa -n demo my-custom-serviceaccount

kubectl delete ns demo
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- Learn about [backup & restore](/docs/guides/postgres/backup/stash.md) of PostgreSQL databases using Stash.
- Learn about initializing [PostgreSQL with Script](/docs/guides/postgres/initialization/script_source.md).
- Want to setup PostgreSQL cluster? Check how to [configure Highly Available PostgreSQL Cluster](/docs/guides/postgres/clustering/ha_cluster.md)
- Monitor your PostgreSQL instance with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL instance with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
