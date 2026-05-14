---
title: RBAC for Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-rbac-quickstart
    name: RBAC
    parent: neo4j-quickstart
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RBAC Permissions for Neo4j

If RBAC is enabled in your cluster, KubeDB creates Neo4j-specific RBAC resources so Neo4j pods can discover Services and Endpoints during cluster operations.

Here are the additional permissions used by Neo4j pods:

| Kubernetes Resource | Resource Names | Permission required |
|---------------------|----------------|---------------------|
| services            |                | get, list, watch    |
| endpoints           |                | get, list, watch    |

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB CLI on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/neo4j/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/neo4j/quickstart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a Neo4j Database

Below is the Neo4j object used in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-test
  namespace: demo
spec:
  replicas: 3
  version: "2025.12.1"
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

Create the above Neo4j object with the following command:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/quickstart/neo4j.yaml
neo4j.kubedb.com/neo4j-test created
```

When this `Neo4j` object is created, KubeDB operator creates a `Role`, `ServiceAccount`, and `RoleBinding` with matching names and uses that ServiceAccount in the Neo4j pods.

Let's inspect what KubeDB creates.

### Role

KubeDB operator creates a Role object `neo4j-test-role` in the same namespace as the Neo4j object.

```yaml
$ kubectl get role -n demo neo4j-test-role -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  creationTimestamp: "2026-05-14T06:54:08Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: neo4j-test
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: neo4js.kubedb.com
  name: neo4j-test-role
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Neo4j
      name: neo4j-test
      uid: 0034a30c-d33d-4596-a6d8-7cf47aa3d9e6
  resourceVersion: "1461745"
  uid: 1f3850bc-4d28-4780-88ad-b31e9c7fa21e
rules:
  - apiGroups:
      - ""
    resources:
      - services
      - endpoints
    verbs:
      - get
      - list
      - watch
```

### ServiceAccount

KubeDB operator creates a ServiceAccount object `neo4j-test` in the same namespace as the Neo4j object.

```yaml
$ kubectl get serviceaccount -n demo neo4j-test -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2026-05-14T06:54:08Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: neo4j-test
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: neo4js.kubedb.com
  name: neo4j-test
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Neo4j
      name: neo4j-test
      uid: 0034a30c-d33d-4596-a6d8-7cf47aa3d9e6
  resourceVersion: "1461744"
  uid: 8bb16bdc-2a76-454c-8a58-284f0cc33da3
```

This ServiceAccount is used by Neo4j pods created for the `neo4j-test` database.

### RoleBinding

KubeDB operator creates a RoleBinding object `neo4j-test-rolebinding` in the same namespace as the Neo4j object.

```yaml
$ kubectl get rolebinding -n demo neo4j-test-rolebinding -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "2026-05-14T06:54:09Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: neo4j-test
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: neo4js.kubedb.com
  name: neo4j-test-rolebinding
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Neo4j
      name: neo4j-test
      uid: 0034a30c-d33d-4596-a6d8-7cf47aa3d9e6
  resourceVersion: "1461748"
  uid: f5e8ae7f-62a9-4390-bae0-918f4d5b54d1
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: neo4j-test-role
subjects:
  - kind: ServiceAccount
    name: neo4j-test
```

This object binds Role `neo4j-test-role` with ServiceAccount `neo4j-test`.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete -n demo neo4j/neo4j-test
neo4j.kubedb.com "neo4j-test" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```

