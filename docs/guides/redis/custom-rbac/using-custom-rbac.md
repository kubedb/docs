---
title: Run Redis with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: rd-custom-rbac-quickstart
    name: Custom RBAC
    parent: rd-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a Redis instance. This tutorial will show you how to use KubeDB to run Redis instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for Redis. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in Redis crd. If this field is left empty, the KubeDB operator will create a service account name matching Redis crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a Redis instance named `quick-redis` to provide the bare minimum access permissions.

## Custom RBAC for Redis

At first, let's create a `Service Acoount` in `demo` namespace.

```bash
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

It should create a service account.

```bash
$ kubectl get serviceaccount -n demo my-custom-serviceaccount -o yaml
```
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2023-02-06T10:19:00Z"
  name: my-custom-serviceaccount
  namespace: demo
  resourceVersion: "683509"
  uid: 186702c3-6d84-4ba9-b349-063c4e681622
secrets:
  - name: my-custom-serviceaccount-token-vpr84
```

Now, we need to create a role that has necessary access permissions for the Redis instance named `quick-redis`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/custom-rbac/rd-custom-role.yaml
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

```bash
$ kubectl create rolebinding my-custom-rolebinding --role=my-custom-role --serviceaccount=demo:my-custom-serviceaccount --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created

```

It should bind `my-custom-role` and `my-custom-serviceaccount` successfully.

```bash
$ kubectl get rolebinding -n demo my-custom-rolebinding -o yaml
```
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "2023-02-06T09:46:26Z"
  name: my-custom-rolebinding
  namespace: demo
  resourceVersion: "680621"
  uid: 6f74cce7-bb20-4584-bdc1-bdfb3598604f
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/custom-rbac/rd-custom-db.yaml
redis.kubedb.com/quick-redis created
```

Below is the YAML for the Redis crd we just created.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: quick-redis
  namespace: demo
spec:
  version: 6.2.14
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

Now, wait a few minutes. the KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we should see that a pod with the name `quick-redis-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo quick-redis-0
NAME            READY   STATUS    RESTARTS   AGE
quick-redis-0   1/1     Running   0          61s
```

Check if database is in Ready state

```bash
$ kubectl get redis -n demo
NAME          VERSION   STATUS   AGE
quick-redis   6.2.14     Ready    117s
```

## Reusing Service Account

An existing service account can be reused in another Redis instance. No new access permission is required to run the new Redis instance.

Now, create Redis crd `minute-redis` using the existing service account name `my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/custom-rbac/rd-custom-db-two.yaml
redis.kubedb.com/quick-redis created
```

Below is the YAML for the Redis crd we just created.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: minute-redis
  namespace: demo
spec:
  version: 6.2.14
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

Now, wait a few minutes. the KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we should see that a pod with the name `minute-redis-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo minute-redis-0
NAME                READY     STATUS    RESTARTS   AGE
minute-redis-0   1/1       Running   0          14m
```

Check if database is in Ready state

```bash
$ kubectl get redis -n demo
NAME           VERSION   STATUS   AGE
minute-redis   6.2.14     Ready    76s
quick-redis    6.2.14     Ready    4m26s
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo rd/quick-redis -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/quick-redis patched

$ kubectl delete -n demo rd/quick-redis
redis.kubedb.com "quick-redis" deleted

$ kubectl patch -n demo rd/minute-redis -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/minute-redis patched

$ kubectl delete -n demo rd/minute-redis
redis.kubedb.com "minute-redis" deleted

$ kubectl delete -n demo role my-custom-role
role.rbac.authorization.k8s.io "my-custom-role" deleted

$ kubectl delete -n demo rolebinding my-custom-rolebinding
rolebinding.rbac.authorization.k8s.io "my-custom-rolebinding" deleted

$ kubectl delete sa -n demo my-custom-serviceaccount
serviceaccount "my-custom-serviceaccount" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart Redis](/docs/guides/redis/quickstart/quickstart.md) with KubeDB Operator.
- Monitor your Redis instance with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis instance with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Use [kubedb cli](/docs/guides/redis/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

