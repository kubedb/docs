---
title: Run Weaviate with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: weaviate-custom-rbac-quickstart
    name: Custom RBAC
    parent: weaviate-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC Resources

KubeDB supports finer user control over role-based access permissions provided to a Weaviate instance. This tutorial will show you how to use KubeDB to run a Weaviate instance with custom RBAC resources.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/custom-rbac](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/custom-rbac) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows users to provide custom RBAC resources, namely `ServiceAccount`, `Role`, and `RoleBinding` for Weaviate. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in the Weaviate CRD. If this field is left empty, the KubeDB operator will create a ServiceAccount matching the Weaviate name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this ServiceAccount.

If a ServiceAccount name is given but there is no existing ServiceAccount by that name, the KubeDB operator will create one, and Role and RoleBinding will also be generated automatically.

If a ServiceAccount name is given and there is an existing ServiceAccount by that name, the KubeDB operator will use that existing ServiceAccount. Since this ServiceAccount is not managed by KubeDB, users are responsible for providing necessary access permissions manually.

This guide will show you how to create custom `ServiceAccount`, `Role`, and `RoleBinding` for a Weaviate instance named `weaviate-custom-rbac` to provide the bare minimum access permissions.

## Custom RBAC for Weaviate

At first, let's create a `ServiceAccount` in the `demo` namespace:

```bash
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

Verify that the ServiceAccount was created:

```yaml
$ kubectl get serviceaccount -n demo my-custom-serviceaccount -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-custom-serviceaccount
  namespace: demo
```

Now, create a Role that has the necessary access permissions for the Weaviate database named `weaviate-custom-rbac`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/custom-rbac/weaviate-custom-role.yaml
role.rbac.authorization.k8s.io/my-custom-role created
```

Below is the YAML for the Role we just created:

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
  - weaviate-custom-rbac
  resources:
  - statefulsets
  verbs:
  - get
- apiGroups:
  - kubedb.com
  resourceNames:
  - weaviate-custom-rbac
  resources:
  - weaviates
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - list
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - pods/exec
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - get
  - list
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - create
  - get
  - update
```

Note that `resourceNames` like `weaviate-custom-rbac` are unique to this particular Weaviate instance. Another database instance `weaviate-custom-rbac-2` would require these `resourceNames` to be updated accordingly.

Now, create a `RoleBinding` to bind this `Role` with the already created ServiceAccount:

```bash
$ kubectl create rolebinding my-custom-rolebinding \
  --role=my-custom-role \
  --serviceaccount=demo:my-custom-serviceaccount \
  --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created
```

Verify the RoleBinding was created:

```yaml
$ kubectl get rolebinding -n demo my-custom-rolebinding -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: my-custom-rolebinding
  namespace: demo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: my-custom-role
subjects:
- kind: ServiceAccount
  name: my-custom-serviceaccount
  namespace: demo
```

Now, create a `Weaviate` CR specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/custom-rbac/weaviate-custom-db.yaml
weaviate.kubedb.com/weaviate-custom-rbac created
```

Below is the YAML for the `Weaviate` CR we just created:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-custom-rbac
  namespace: demo
spec:
  version: "1.33.1"
  replicas: 3
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
  deletionPolicy: WipeOut
```

Now, wait a few minutes. If everything goes well, we will see that the Weaviate pods are running with the custom ServiceAccount:

```bash
$ kubectl get weaviate -n demo weaviate-custom-rbac
NAME                   VERSION   STATUS   AGE
weaviate-custom-rbac   1.33.1    Ready    2m

$ kubectl get pod -n demo -l app.kubernetes.io/instance=weaviate-custom-rbac -o=custom-columns=NAME:.metadata.name,SERVICEACCOUNT:.spec.serviceAccountName
NAME                     SERVICEACCOUNT
weaviate-custom-rbac-0   my-custom-serviceaccount
weaviate-custom-rbac-1   my-custom-serviceaccount
weaviate-custom-rbac-2   my-custom-serviceaccount
```

The output confirms that all Weaviate pods are running with our custom `my-custom-serviceaccount` ServiceAccount.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviate -n demo weaviate-custom-rbac
kubectl delete rolebinding -n demo my-custom-rolebinding
kubectl delete role -n demo my-custom-role
kubectl delete serviceaccount -n demo my-custom-serviceaccount
kubectl delete ns demo
```