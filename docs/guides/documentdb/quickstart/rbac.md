---
title: RBAC for DocumentDB
menu:
  docs_{{ .version }}:
    identifier: documentdb-rbac-quickstart
    name: RBAC
    parent: documentdb-quickstart
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RBAC Permissions for DocumentDB

When RBAC is enabled in your cluster, KubeDB automatically creates the necessary Role, ServiceAccount, and RoleBinding for each DocumentDB instance. This tutorial explains what permissions are granted and how to verify them.

Here is the list of additional permissions required by the health checker and DocumentDB operations:

| Kubernetes Resource | Permissions         |
|---------------------|---------------------|
| pods                | get, list, watch    |
| pods/exec           | create              |
| secrets             | get                 |
| services            | get, list           |
| endpoints           | get, list           |

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a DocumentDB Database

Below is the DocumentDB object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: quick-docdb
  namespace: demo
spec:
  version: "pg17-0.109.0"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: Delete
```

Create the above DocumentDB object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/quickstart/quick-docdb.yaml
documentdb.kubedb.com/quick-docdb created
```

When this DocumentDB object is created, KubeDB operator creates Role, ServiceAccount and RoleBinding with the matching DocumentDB name and uses that ServiceAccount in the corresponding PetSet.

## Verify RBAC Resources Created

Let's verify what RBAC resources KubeDB has created for the DocumentDB instance.

### Role

KubeDB operator creates a Role object `quick-docdb` in the same namespace as the DocumentDB object. This Role grants the necessary permissions for the DocumentDB pod to operate.

```bash
$ kubectl get role -n demo quick-docdb -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-docdb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: documentdbs.kubedb.com
  name: quick-docdb
  namespace: demo
rules:
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
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
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - get
  - list
```

This Role grants the minimum required permissions for DocumentDB health checker to monitor the database and access necessary cluster resources. The permissions include pod monitoring, secret access, and service discovery.

### ServiceAccount

KubeDB operator creates a ServiceAccount object `quick-docdb` in the same namespace as the DocumentDB object.

```bash
$ kubectl get serviceaccount -n demo quick-docdb -o yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-docdb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: documentdbs.kubedb.com
  name: quick-docdb
  namespace: demo
```

This ServiceAccount is used by the DocumentDB health checker container to monitor the database. You can verify it's being used by checking the PetSet:

```bash
$ kubectl get petset -n demo quick-docdb -o jsonpath='{.spec.template.spec.serviceAccountName}'
quick-docdb
```

### RoleBinding

KubeDB operator creates a RoleBinding object `quick-docdb` in the same namespace as the DocumentDB object.

```bash
$ kubectl get rolebinding -n demo quick-docdb -o yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: quick-docdb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: documentdbs.kubedb.com
  name: quick-docdb
  namespace: demo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: quick-docdb
subjects:
- kind: ServiceAccount
  name: quick-docdb
  namespace: demo
```

This RoleBinding binds the Role `quick-docdb` with the ServiceAccount `quick-docdb`, granting the health checker permissions to monitor the DocumentDB instance and access necessary resources.

## Verify DocumentDB is Running

Let's verify that the DocumentDB instance is running with the correct RBAC configuration:

```bash
$ kubectl get documentdb -n demo quick-docdb
NAME          VERSION       STATUS    AGE
quick-docdb   pg17-0.109.0  Running   5m
```

Check that the pod is running:

```bash
$ kubectl get pods -n demo quick-docdb-0
NAME            READY   STATUS    RESTARTS   AGE
quick-docdb-0   1/1     Running   0          5m
```

View the pod logs to confirm the database started successfully:

```bash
$ kubectl logs -n demo quick-docdb-0
2025-01-12 10:30:15.123 UTC [1] LOG:  starting DocumentDB
2025-01-12 10:30:15.456 UTC [1] LOG:  listening on IPv4 address "0.0.0.0", port 27017
2025-01-12 10:30:15.789 UTC [1] LOG:  database system is ready to accept connections
```

## Custom RBAC

If you need fine-grained control over RBAC permissions or want to use a custom service account, please refer to the [Custom RBAC Guide](/docs/guides/documentdb/custom-rbac/using-custom-rbac.md) for detailed instructions.

## Automatic Resource Cleanup

When you delete the DocumentDB instance, KubeDB automatically cleans up the associated RBAC resources (Role, ServiceAccount, and RoleBinding) since they are owned by the DocumentDB object.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo documentdb/quick-docdb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo documentdb/quick-docdb

kubectl delete ns demo
```

## Next Steps

- Read the [DocumentDB CRD Concept](/docs/guides/documentdb/concepts/documentdb.md) for detailed DocumentDB specification.
- Read the [DocumentDBVersion CRD Concept](/docs/guides/documentdb/concepts/catalog.md) for DatabaseVersion specification.
- Learn [Custom RBAC](/docs/guides/documentdb/custom-rbac/using-custom-rbac.md) for fine-grained access control.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

