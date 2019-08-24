---
title: Run Redis with Custom RBAC resources
menu:
  docs_0.12.0:
    identifier: rd-custom-rbac-quickstart
    name: Custom RBAC
    parent: rd-custom-rbac
    weight: 10
menu_name: docs_0.12.0
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a Redis instance. This tutorial will show you how to use KubeDB to run Redis instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [minikube](https://github.com/kubernetes/minikube).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/0.12.0/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for Redis. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in Redis crd. If this field is left empty, the KubeDB operator will create a service account name matching Redis crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a Redis instance named `quick-postges` to provide the bare minimum access permissions.

## Custom RBAC for Redis

At first, let's create a `Service Acoount` in `demo` namespace.

```console
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

Now, we need to create a role that has necessary access permissions for the Redis instance named `quick-redis`.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/0.12.0/docs/examples/redis/custom-rbac/rd-custom-role.yaml
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
  - redis-db
  resources:
  - podsecuritypolicies
  verbs:
  - use
```

This permission is required for Redis pods running on PSP enabled clusters.

Now create a `RoleBinding` to bind this `Role` with the already created service account.

```console
$ kubectl create rolebinding my-custom-rolebinding --role=my-custom-role --serviceaccount=demo:my-custom-serviceaccount --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created

```

It should bind `my-custom-role` and `my-custom-serviceaccount` successfully.

```yaml
$ kubectl get rolebinding -n demo my-custom-rolebinding -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
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

Now, create a Redis crd specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/0.12.0/docs/examples/redis/custom-rbac/rd-custom-db.yaml
redis.kubedb.com/quick-redis created
```

Below is the YAML for the Redis crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Redis
metadata:
  name: quick-redis
  namespace: demo
spec:
  version: "4.0-v2"
  storageType: Durable
  storage:
    podTemplate:
      spec:
        serviceAccountName: my-custom-serviceaccount
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate

```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we should see that a pod with the name `quick-redis-0` has been created.

Check that the statefulset's pod is running

```console
$ kubectl get pod -n demo quick-redis-0
NAME                READY     STATUS    RESTARTS   AGE
quick-redis-0   1/1       Running   0          14m
```

Check the pod's log to see if the database is ready

```console
$ kubectl logs -f -n demo quick-redis-0
1:C 10 Jun 04:32:25.537 # oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
1:C 10 Jun 04:32:25.537 # Redis version=4.0.11, bits=64, commit=00000000, modified=0, pid=1, just started
1:C 10 Jun 04:32:25.537 # Warning: no config file specified, using the default config. In order to specify a config file use redis-server /path/to/redis.conf
1:M 10 Jun 04:32:25.537 * Running mode=standalone, port=6379.
1:M 10 Jun 04:32:25.537 # WARNING: The TCP backlog setting of 511 cannot be enforced because /proc/sys/net/core/somaxconn is set to the lower value of 128.
1:M 10 Jun 04:32:25.537 # Server initialized
1:M 10 Jun 04:32:25.537 * Ready to accept connections
```

Once we see `Ready to accept connections` in the log, the database is ready.

## Reusing Service Account

An existing service account can be reused in another Redis instance. No new access permission is required to run the new Redis instance.

Now, create Redis crd `minute-redis` using the existing service account name `my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/0.12.0/docs/examples/redis/custom-rbac/rd-custom-db-two.yaml
redis.kubedb.com/quick-redis created
```

Below is the YAML for the Redis crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Redis
metadata:
  name: minute-redis
  namespace: demo
spec:
  version: "4.0-v2"
  storageType: Durable
  storage:
    podTemplate:
      spec:
        serviceAccountName: my-custom-serviceaccount
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate

```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, statefulset, services, secret etc. If everything goes well, we should see that a pod with the name `minute-redis-0` has been created.

Check that the statefulset's pod is running

```console
$ kubectl get pod -n demo minute-redis-0
NAME                READY     STATUS    RESTARTS   AGE
minute-redis-0   1/1       Running   0          14m
```

Check the pod's log to see if the database is ready

```console
$ kubectl logs -f -n demo minute-redis-0
1:C 10 Jun 04:32:25.537 # oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
1:C 10 Jun 04:32:25.537 # Redis version=4.0.11, bits=64, commit=00000000, modified=0, pid=1, just started
1:C 10 Jun 04:32:25.537 # Warning: no config file specified, using the default config. In order to specify a config file use redis-server /path/to/redis.conf
1:M 10 Jun 04:32:25.537 * Running mode=standalone, port=6379.
1:M 10 Jun 04:32:25.537 # WARNING: The TCP backlog setting of 511 cannot be enforced because /proc/sys/net/core/somaxconn is set to the lower value of 128.
1:M 10 Jun 04:32:25.537 # Server initialized
1:M 10 Jun 04:32:25.537 * Ready to accept connections
```

`Ready to accept connections` in the log signifies that the database is running successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl patch -n demo rd/quick-redis -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo rd/quick-redis

kubectl patch -n demo rd/minute-redis -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo rd/minute-redis

kubectl delete -n demo role my-custom-role
kubectl delete -n demo rolebinding my-custom-rolebinding

kubectl delete sa -n demo my-custom-serviceaccount

kubectl delete ns demo
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/uninstall.md).

## Next Steps

- [Quickstart Redis](/docs/guides/redis/quickstart/quickstart.md) with KubeDB Operator.
- [Snapshot and Restore](/docs/guides/redis/snapshot/backup-and-restore.md) process of Redis instances using KubeDB.
- Take [Scheduled Snapshot](/docs/guides/redis/snapshot/scheduled-backup.md) of Redis instances using KubeDB.
- Initialize [Redis with Script](/docs/guides/redis/initialization/using-script.md).
- Initialize [Redis with Snapshot](/docs/guides/redis/initialization/using-snapshot.md).
- Monitor your Redis instance with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/redis/monitoring/using-coreos-prometheus-operator.md).
- Monitor your Redis instance with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Use [kubedb cli](/docs/guides/redis/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Redis object](/docs/concepts/databases/redis.md).
- Detail concepts of [Snapshot object](/docs/concepts/snapshot.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

