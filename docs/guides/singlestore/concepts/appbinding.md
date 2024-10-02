---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: sdb-appbinding-concepts
    name: AppBinding
    parent: sdb-concepts-singlestore
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# AppBinding

## What is AppBinding

An `AppBinding` is a Kubernetes `CustomResourceDefinition`(CRD) which points to an application using either its URL (usually for a non-Kubernetes resident service instance) or a Kubernetes service object (if self-hosted in a Kubernetes cluster), some optional parameters and a credential secret. To learn more about AppBinding and the problems it solves, please read this blog post: [The case for AppBinding](https://appscode.com/blog/post/the-case-for-appbinding).

If you deploy a database using [KubeDB](https://kubedb.com/docs/0.11.0/concepts/), `AppBinding` object will be created automatically for it. Otherwise, you have to create an `AppBinding` object manually pointing to your desired database.

KubeDB uses [Stash](https://appscode.com/products/stash/) to perform backup/recovery of databases. Stash needs to know how to connect with a target database and the credentials necessary to access it. This is done via an `AppBinding`.

## AppBinding CRD Specification

Like any official Kubernetes resource, an `AppBinding` has `TypeMeta`, `ObjectMeta` and `Spec` sections. However, unlike other Kubernetes resources, it does not have a `Status` section.

An `AppBinding` object created by `KubeDB` for SingleStore database is shown below,

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Singlestore","metadata":{"annotations":{},"name":"sdb","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","licenseSecret":{"name":"license-secret"},"storageType":"Durable","topology":{"aggregator":{"podTemplate":{"spec":{"containers":[{"name":"singlestore","resources":{"limits":{"cpu":"600m","memory":"2Gi"},"requests":{"cpu":"600m","memory":"2Gi"}}}]}},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}}},"leaf":{"podTemplate":{"spec":{"containers":[{"name":"singlestore","resources":{"limits":{"cpu":"600m","memory":"2Gi"},"requests":{"cpu":"600m","memory":"2Gi"}}}]}},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"10Gi"}}}}},"version":"8.7.10"}}
  creationTimestamp: "2024-08-15T09:04:57Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sdb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: singlestores.kubedb.com
  name: sdb
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Singlestore
    name: sdb
    uid: efeededa-7dc8-4e98-b4e4-0e0603adc32c
  resourceVersion: "5633763"
  uid: 77dc8123-52f6-4832-a6e2-478795e487b7
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Singlestore
    name: sdb
    namespace: demo
  clientConfig:
    service:
      name: sdb
      path: /
      port: 3306
      scheme: tcp
    url: tcp(sdb.demo.svc:3306)/
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: SinglestoreConfiguration
    masterAggregator: sdb-aggregator-0.sdb-pods.demo.svc
    stash:
      addon:
        backupTask:
          name: ""
        restoreTask:
          name: ""
  secret:
    name: sdb-root-cred
  type: kubedb.com/singlestore
  version: 8.7.10
```

Here, we are going to describe the sections of an `AppBinding` crd.

### AppBinding `Spec`

An `AppBinding` object has the following fields in the `spec` section:

#### spec.type

`spec.type` is an optional field that indicates the type of the app that this `AppBinding` is pointing to. Stash uses this field to resolve the values of `TARGET_APP_TYPE`, `TARGET_APP_GROUP` and `TARGET_APP_RESOURCE` variables of [BackupBlueprint](https://appscode.com/products/stash/latest/concepts/crds/backupblueprint/) object.

This field follows the following format: `<app group>/<resource kind>`. The above AppBinding is pointing to a `postgres` resource under `kubedb.com` group.

Here, the variables are parsed as follows:

|       Variable        | Usage                                                                                                                                |
| --------------------- |--------------------------------------------------------------------------------------------------------------------------------------|
| `TARGET_APP_GROUP`    | Represents the application group where the respective app belongs (i.e: `kubedb.com`).                                               |
| `TARGET_APP_RESOURCE` | Represents the resource under that application group that this appbinding represents (i.e: `singlestore`).                           |
| `TARGET_APP_TYPE`     | Represents the complete type of the application. It's simply `TARGET_APP_GROUP/TARGET_APP_RESOURCE` (i.e: `kubedb.com/singlestore`). |

#### spec.secret

`spec.secret` specifies the name of the secret which contains the credentials that are required to access the database. This secret must be in the same namespace as the `AppBinding`.

This secret must contain the following keys:

SingleStore :

| Key        | Usage                                          |
|------------|------------------------------------------------|
| `username` | Username of the target database.               |
| `password` | Password for the user specified by `password`. |

#### spec.appRef
appRef refers to the underlying application. It has 4 fields named `apiGroup`, `kind`, `name` & `namespace`.

#### spec.clientConfig

`spec.clientConfig` defines how to communicate with the target database. You can use either an URL or a Kubernetes service to connect with the database. You don't have to specify both of them.

You can configure following fields in `spec.clientConfig` section:

- **spec.clientConfig.url**

  `spec.clientConfig.url` gives the location of the database, in standard URL form (i.e. `[scheme://]host:port/[path]`). This is particularly useful when the target database is running outside of the Kubernetes cluster. If your database is running inside the cluster, use `spec.clientConfig.service` section instead.

> Note that, attempting to use a user or basic auth (e.g. `user:password@host:port`) is not allowed. Stash will insert them automatically from the respective secret. Fragments ("#...") and query parameters ("?...") are not allowed either.

- **spec.clientConfig.service**

  If you are running the database inside the Kubernetes cluster, you can use Kubernetes service to connect with the database. You have to specify the following fields in `spec.clientConfig.service` section if you manually create an `AppBinding` object.

  - **name :** `name` indicates the name of the service that connects with the target database.
  - **scheme :** `scheme` specifies the scheme (i.e. http, https) to use to connect with the database.
  - **port :** `port` specifies the port where the target database is running.

- **spec.clientConfig.insecureSkipTLSVerify**

  `spec.clientConfig.insecureSkipTLSVerify` is used to disable TLS certificate verification while connecting with the database. We strongly discourage to disable TLS verification during backup. You should provide the respective CA bundle through `spec.clientConfig.caBundle` field instead.

- **spec.clientConfig.caBundle**

  `spec.clientConfig.caBundle` is a PEM encoded CA bundle which will be used to validate the serving certificate of the database.

## Next Steps

- Learn how to use KubeDB to manage various databases [here](/docs/guides/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).