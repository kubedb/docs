---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: ch-appbinding-concepts
    name: AppBinding
    parent: ch-concepts-clickhouse
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# AppBinding

## What is AppBinding

An `AppBinding` is a Kubernetes `CustomResourceDefinition`(CRD) which points to an application using either its URL (usually for a non-Kubernetes resident service instance) or a Kubernetes service object (if self-hosted in a Kubernetes cluster), some optional parameters and a credential secret. To learn more about AppBinding and the problems it solves, please read this blog post: [The case for AppBinding](https://appscode.com/blog/post/the-case-for-appbinding).

If you deploy a database using [KubeDB](https://kubedb.com/docs/latest/welcome/), `AppBinding` object will be created automatically for it. Otherwise, you have to create an `AppBinding` object manually pointing to your desired database.

KubeDB uses [Stash](https://appscode.com/products/stash/) to perform backup/recovery of databases. Stash needs to know how to connect with a target database and the credentials necessary to access it. This is done via an `AppBinding`.

## AppBinding CRD Specification

Like any official Kubernetes resource, an `AppBinding` has `TypeMeta`, `ObjectMeta` and `Spec` sections. However, unlike other Kubernetes resources, it does not have a `Status` section.

An `AppBinding` object created by `KubeDB` for ClickHouse database is shown below,

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"ClickHouse","metadata":{"annotations":{},"name":"ch-cluster","namespace":"demo"},"spec":{"clusterTopology":{"clickHouseKeeper":{"externallyManaged":false,"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}}}},"cluster":[{"name":"appscode-cluster","podTemplate":{"spec":{"containers":[{"name":"clickhouse","resources":{"limits":{"memory":"4Gi"},"requests":{"cpu":"500m","memory":"2Gi"}}}],"initContainers":[{"name":"clickhouse-init","resources":{"limits":{"memory":"1Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}}]}},"replicas":2,"shards":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}}}]},"deletionPolicy":"WipeOut","version":"24.4.1"}}
  creationTimestamp: "2025-08-20T06:14:23Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: ch-cluster
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: clickhouses.kubedb.com
  name: ch-cluster
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: ClickHouse
      name: ch-cluster
      uid: 2aca010c-c9ef-4b87-b07e-b72c0c252668
  resourceVersion: "154030"
  uid: 36a15d9a-bad9-4892-8376-ae2ad0c87056
spec:
  appRef:
    apiGroup: kubedb.com
    kind: ClickHouse
    name: ch-cluster
    namespace: demo
  clientConfig:
    service:
      name: ch-cluster
      port: 9000
      scheme: http
  secret:
    name: ch-cluster-auth
    kind: Secret
  type: kubedb.com/clickhouse
  version: 24.4.1
```
Here, we are going to describe the sections of an `AppBinding` crd.

### AppBinding `Spec`

An `AppBinding` object has the following fields in the `spec` section:

#### spec.type

`spec.type` is an optional field that indicates the type of the app that this `AppBinding` is pointing to.

<!--- Add when Stash support is added --->
<!---
Stash uses this field to resolve the values of `TARGET_APP_TYPE`, `TARGET_APP_GROUP` and `TARGET_APP_RESOURCE` variables of [BackupBlueprint](https://appscode.com/products/stash/latest/concepts/crds/backupblueprint/) object.

This field follows the following format: `<app group>/<resource kind>`. The above AppBinding is pointing to a `clickhouse` resource under `kubedb.com` group.

Here, the variables are parsed as follows:

|       Variable        | Usage                                                                                                                          |
| --------------------- |--------------------------------------------------------------------------------------------------------------------------------|
| `TARGET_APP_GROUP`    | Represents the application group where the respective app belongs (i.e: `kubedb.com`).                                         |
| `TARGET_APP_RESOURCE` | Represents the resource under that application group that this appbinding represents (i.e: `clickhouse`).                           |
| `TARGET_APP_TYPE`     | Represents the complete type of the application. It's simply `TARGET_APP_GROUP/TARGET_APP_RESOURCE` (i.e: `kubedb.com/clickhouse`). |

--->

#### spec.secret

`spec.secret` specifies the name of the secret which contains the credentials that are required to access the database. This secret must be in the same namespace as the `AppBinding`.

This secret must contain the following keys for ClickHouse:

| Key        | Usage                                          |
| ---------- |------------------------------------------------|
| `username` | Username of the target ClickHouse instance.    |
| `password` | Password for the user specified by `username`. |


#### spec.appRef
appRef refers to the underlying application. It has 4 fields named `apiGroup`, `kind`, `name` & `namespace`.

#### spec.clientConfig

`spec.clientConfig` defines how to communicate with the target database. You can use either a URL or a Kubernetes service to connect with the database. You don't have to specify both of them.

You can configure following fields in `spec.clientConfig` section:

- **spec.clientConfig.service**

  If you are running the database inside the Kubernetes cluster, you can use Kubernetes service to connect with the database. You have to specify the following fields in `spec.clientConfig.service` section if you manually create an `AppBinding` object.

    - **name :** `name` indicates the name of the service that connects with the target database.
    - **scheme :** `scheme` specifies the scheme (i.e. http, https) to use to connect with the database.
    - **port :** `port` specifies the port where the target database is running.

## Next Steps

- Learn how to use KubeDB to manage various databases [here](/docs/guides/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
