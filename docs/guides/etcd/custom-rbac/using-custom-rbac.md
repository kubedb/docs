---
title: Run Etcd with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: etcd-custom-rbac-quickstart
    name: Custom RBAC
    parent: etcd-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC resources

KubeDB supports finer user control over the role based access permissions provided to an `Etcd`
cluster. This tutorial shows you how to run an etcd cluster with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be
configured to communicate with your cluster. If you do not already have a cluster, you can create one
by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install the KubeDB CLI on your workstation and the KubeDB operator in your cluster following the
steps [here](/docs/setup/README.md). etcd support is an alpha feature gate, so it must be enabled
explicitly — pass `--set featureGates.Etcd=true` to the KubeDB Helm chart.

To keep things isolated, this tutorial uses a separate namespace called `demo`.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: the YAML files used in this tutorial are stored in
> [docs/examples/etcd/custom-rbac](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd/custom-rbac)
> of the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Overview

KubeDB allows users to provide custom RBAC resources — namely a `ServiceAccount`, a `Role` and a
`RoleBinding` — for an `Etcd` cluster. This is done through the
`spec.podTemplate.spec.serviceAccountName` field of the `Etcd` CRD.

- If the field is left empty, the KubeDB operator creates a `ServiceAccount` whose name matches the
  `Etcd` object's name. A `Role` and a `RoleBinding` granting the necessary access permissions are
  generated automatically for that service account.
- If a service account name is given but no service account by that name exists, the KubeDB operator
  creates one, along with the `Role` and `RoleBinding` that grant the necessary permissions.
- If a service account name is given and a service account by that name already exists, the KubeDB
  operator uses the existing one. Since that service account is not managed by KubeDB, **you** are
  responsible for granting it the necessary permissions.

This guide shows how to create a custom `ServiceAccount`, `Role` and `RoleBinding` for an `Etcd`
cluster named `etcd-quickstart` with the bare minimum access permissions.

## Custom RBAC for Etcd

At first, let's create a `ServiceAccount` in the `demo` namespace.

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
  creationTimestamp: "2026-08-15T04:23:39Z"
  name: my-custom-serviceaccount
  namespace: demo
  resourceVersion: "21657"
  uid: b2ec2b05-8292-11e9-8d10-080027a8b217
```

Now, we need to create a `Role` with the access permissions the `Etcd` cluster named
`etcd-quickstart` needs.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/custom-rbac/etcd-custom-role.yaml
role.rbac.authorization.k8s.io/my-custom-role created
```

Below is the YAML for the `Role` we just created.

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
  - etcd-db
  resources:
  - podsecuritypolicies
  verbs:
  - use
```

This permission is required for etcd pods running on PSP enabled clusters.

Now create a `RoleBinding` to bind this `Role` with the service account we created.

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
  creationTimestamp: "2026-08-15T04:25:12Z"
  name: my-custom-rolebinding
  namespace: demo
  resourceVersion: "1405"
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

Now, create an `Etcd` object with `spec.podTemplate.spec.serviceAccountName` set to
`my-custom-serviceaccount`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/custom-rbac/etcd-custom-db.yaml
etcd.kubedb.com/etcd-quickstart created
```

Below is the YAML for the `Etcd` object we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-quickstart
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
      containers:
        - name: etcd
          resources:
            limits:
              cpu: 500m
              memory: 1Gi
            requests:
              cpu: 250m
              memory: 512Mi
  deletionPolicy: DoNotTerminate
```

> The container name in `spec.podTemplate.spec.containers` must be `etcd` — that is the name of the
> main etcd container in a KubeDB-managed pod.

Now, wait a few minutes. The KubeDB operator will create the necessary PetSet, Services, Secrets and
so on. Members are brought up one ordinal at a time, so if everything goes well we should end up with
three running pods.

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-quickstart
NAME                READY   STATUS    RESTARTS   AGE
etcd-quickstart-0   1/1     Running   0          5m54s
etcd-quickstart-1   1/1     Running   0          4m42s
etcd-quickstart-2   1/1     Running   0          3m31s
```

Verify that the pods are actually using the custom service account:

```bash
$ kubectl get pod -n demo etcd-quickstart-0 -o jsonpath='{.spec.serviceAccountName}{"\n"}'
my-custom-serviceaccount
```

And that the cluster reached `Ready`:

```bash
$ kubectl get etcd -n demo etcd-quickstart
NAME              VERSION   STATUS   AGE
etcd-quickstart   3.6.4     Ready    6m10s
```

## Reusing a Service Account

An existing service account can be reused by another `Etcd` cluster. No new access permission is
required to run the new cluster.

Now, create an `Etcd` object named `minute-etcd` using the same service account name
`my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/custom-rbac/etcd-custom-db-two.yaml
etcd.kubedb.com/minute-etcd created
```

Below is the YAML for the `Etcd` object we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: minute-etcd
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 1
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
      containers:
        - name: etcd
          resources:
            limits:
              cpu: 500m
              memory: 1Gi
            requests:
              cpu: 250m
              memory: 512Mi
  deletionPolicy: DoNotTerminate
```

Now, wait a few minutes. The KubeDB operator will create the necessary PVC, PetSet, Services, Secrets
and so on. If everything goes well, we should see a pod named `minute-etcd-0`.

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=minute-etcd
NAME            READY   STATUS    RESTARTS   AGE
minute-etcd-0   1/1     Running   0          5m52s
```

> A single-member cluster is a perfectly valid `Etcd` object — etcd has no separate "standalone" mode
> in KubeDB, `spec.replicas: 1` is all there is to it. It has no fault tolerance, though: use an odd
> number of members (3 or 5) for anything you care about.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo etcd/etcd-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
etcd.kubedb.com/etcd-quickstart patched

$ kubectl delete -n demo etcd/etcd-quickstart
etcd.kubedb.com "etcd-quickstart" deleted

$ kubectl patch -n demo etcd/minute-etcd -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
etcd.kubedb.com/minute-etcd patched

$ kubectl delete -n demo etcd/minute-etcd
etcd.kubedb.com "minute-etcd" deleted

$ kubectl delete -n demo role my-custom-role
role.rbac.authorization.k8s.io "my-custom-role" deleted

$ kubectl delete -n demo rolebinding my-custom-rolebinding
rolebinding.rbac.authorization.k8s.io "my-custom-rolebinding" deleted

$ kubectl delete sa -n demo my-custom-serviceaccount
serviceaccount "my-custom-serviceaccount" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```

If you would like to uninstall the KubeDB operator, please follow the steps
[here](/docs/setup/README.md).

## Next Steps

- Use a [private Docker registry](/docs/guides/etcd/private-registry/using-private-registry.md) to
  deploy etcd with KubeDB.
- Back up and restore your etcd cluster with
  [KubeStash](/docs/guides/etcd/backup/kubestash/overview/index.md).
- Detail concepts of the [Etcd](/docs/guides/etcd/concepts/etcd.md) object.
- Monitor your etcd cluster with KubeDB using the
  [out-of-the-box Prometheus operator](/docs/guides/etcd/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
