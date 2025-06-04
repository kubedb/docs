---
title: Run Ignite with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: ig-custom-rbac-quickstart
    name: Custom RBAC
    parent: ig-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC resources

KubeDB (version 0.13.0 and higher) supports finer user control over role based access permissions provided to a Ignite instance. This tutorial will show you how to use KubeDB to run Ignite database with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/ignite](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/ignite) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for Ignite. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in Ignite crd. If this field is left empty, the KubeDB operator will create a service account name matching Ignite crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `Service Account`, `Role`, and `RoleBinding` for a Ignite instance named `quick-ignite` to provide the bare minimum access permissions.

## Custom RBAC for Ignite

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

Now, we need to create a role that has necessary access permissions for the Ignite instance named `quick-ignite`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/custom-rbac/ig-custom-role.yaml
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
  - ignite-db
  resources:
  - podsecuritypolicies
  verbs:
  - use
```

This permission is required for Ignite pods running on PSP enabled clusters.

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
  creationTimestamp: "kubectl get rolebinding -n demo my-custom-rolebinding -o yaml"
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

Now, create a Ignite crd specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/custom-rbac/ig-custom-db.yaml
ignite.kubedb.com/ignite-quickstart created
```

Below is the YAML for the Ignite crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ignite-quickstart
  namespace: demo
spec:
  replicas: 3
  version: "2.17.0"
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
      containers:
        - name: ignite
          resources:
            limits:
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 250m
              memory: 64Mi
  deletionPolicy: DoNotTerminate
```

Now, wait a few minutes. the KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we should see that a pod with the name `quick-ignite-0` has been created.

Check that the pod is running:

```bash
$ kubectl get pods -n demo
NAME                   READY   STATUS    RESTARTS   AGE
ignite-quickstart-0    1/1     Running   0          5m54s
ignite-quickstart-1    1/1     Running   0          4m42s
ignite-quickstart-2    1/1     Running   0          3m31s
```

## Reusing Service Account

An existing service account can be reused in another Ignite instance. No new access permission is required to run the new Ignite instance.

Now, create Ignite crd `minute-ignite` using the existing service account name `my-custom-serviceaccount` in the `spec.podTemplate.spec.serviceAccountName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/custom-rbac/ig-custom-db-two.yaml
ignite.kubedb.com/ignite-quickstart created
```

Below is the YAML for the Ignite crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: minute-ignite
  namespace: demo
spec:
  replicas: 1
  version: "2.17.0"
  podTemplate:
    spec:
      serviceAccountName: my-custom-serviceaccount
      containers:
        - name: ignite
          resources:
            limits:
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 250m
              memory: 64Mi
  deletionPolicy: DoNotTerminate
```

Now, wait a few minutes. the KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we should see that a pod with the name `minute-ignite-0` has been created.

Check that the pod is running:

```bash
$ kubectl get pods -n demo
NAME                READY   STATUS    RESTARTS   AGE
minute-ignite-0     1/1     Running   0          5m52s
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo ig/ignite-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
ignite.kubedb.com/ignite-quickstart patched

$ kubectl delete -n demo ig/ignite-quickstart
ignite.kubedb.com "ignite-quickstart" deleted

$ kubectl patch -n demo ig/minute-ignite -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
ignite.kubedb.com/minute-ignite patched

$ kubectl delete -n demo ig/minute-ignite
ignite.kubedb.com "minute-ignite" deleted

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

- [Quickstart Ignite](/docs/guides/ignite/quickstart/quickstart.md) with KubeDB Operator.
- Monitor your Ignite database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/ignite/monitoring/using-prometheus-operator.md).
- Monitor your Ignite database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/ignite/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/ignite/private-registry/using-private-registry.md) to deploy Ignite with KubeDB.
- Use [kubedb cli](/docs/guides/ignite/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Ignite object](/docs/guides/ignite/concepts/ignite.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

