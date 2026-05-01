---
title: Run Neo4j with Custom RBAC resources
menu:
  docs_{{ .version }}:
    identifier: neo4j-custom-rbac-quickstart
    name: Custom RBAC
    parent: neo4j-custom-rbac
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom RBAC Resources

KubeDB supports finer user control over role based access permissions provided to a Neo4j instance. This tutorial will show you how to use KubeDB to run Neo4j instance with custom RBAC resources.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Overview

KubeDB allows users to provide custom RBAC resources, namely, `ServiceAccount`, `Role`, and `RoleBinding` for Neo4j. This is provided via the `spec.podTemplate.spec.serviceAccountName` field in Neo4j CRD.

## Custom RBAC for Neo4j

At first, let's create a `Service Account` in `demo` namespace.

```bash
$ kubectl create serviceaccount -n demo my-custom-serviceaccount
serviceaccount/my-custom-serviceaccount created
```

Now, we need to create a role that has necessary access permissions for the Neo4j database named `quick-neo4j`.

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
  - quick-neo4j
  resources:
  - petsets
  verbs:
  - get
- apiGroups:
  - kubedb.com
  resourceNames:
  - quick-neo4j
  resources:
  - neo4js
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
  - secrets
  verbs:
  - get
  - create
  - update
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/custom-rbac/neo4j-custom-role.yaml
role.rbac.authorization.k8s.io/my-custom-role created
```

Now create a `RoleBinding` to bind this `Role` with the already created service account.

```bash
$ kubectl create rolebinding my-custom-rolebinding \
  --role=my-custom-role \
  --serviceaccount=demo:my-custom-serviceaccount \
  --namespace=demo
rolebinding.rbac.authorization.k8s.io/my-custom-rolebinding created
```

Now, create a Neo4j CRD specifying `spec.podTemplate.spec.serviceAccountName` field to `my-custom-serviceaccount`.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: quick-neo4j
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
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
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/custom-rbac/neo4j-custom-db.yaml
neo4j.kubedb.com/quick-neo4j created
```

Check that the pod is running:

```bash
$ kubectl get pod -n demo quick-neo4j-0
NAME             READY   STATUS    RESTARTS   AGE
quick-neo4j-0    1/1     Running   0          3m
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo neo4j/quick-neo4j -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo neo4j/quick-neo4j

kubectl delete -n demo serviceaccount my-custom-serviceaccount
kubectl delete -n demo role my-custom-role
kubectl delete -n demo rolebinding my-custom-rolebinding
kubectl delete ns demo
```
