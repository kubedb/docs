---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: es-appbinding-catalog
    name: AppBinding
    parent: es-concepts-elasticsearch
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# AppBinding

## What is AppBinding

An `AppBinding` is a Kubernetes `CustomResourceDefinition`(CRD) which points to an application using either its URL (usually for a non-Kubernetes resident service instance) or a Kubernetes service object (if self-hosted in a Kubernetes cluster), some optional parameters and a credential secret. To learn more about AppBinding and the problems it solves, please read this blog post: [The case for AppBinding](https://appscode.com/blog/post/the-case-for-appbinding).

If you deploy a database using [KubeDB](https://kubedb.com/docs/latest/welcome/), the `AppBinding` object will be created automatically for it. Otherwise, you have to create an `AppBinding` object manually pointing to your desired database.

KubeDB uses [Stash](https://appscode.com/products/stash/) to perform backup/recovery of databases. Stash needs to know how to connect with a target database and the credentials necessary to access it. This is done via an `AppBinding`.

## AppBinding CRD Specification

Like any official Kubernetes resource, an `AppBinding` has `TypeMeta`, `ObjectMeta` and `Spec` sections. However, unlike other Kubernetes resources, it does not have a `Status` section.

An `AppBinding` object created by `KubeDB` for Elasticsearch database is shown below,

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"Elasticsearch","metadata":{"annotations":{},"name":"es-quickstart","namespace":"demo"},"spec":{"enableSSL":true,"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"Delete","version":"xpack-8.2.3"}}
  creationTimestamp: "2022-12-29T05:03:33Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: es-quickstart
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: elasticsearches.kubedb.com
  name: es-quickstart
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: Elasticsearch
      name: es-quickstart
      uid: 4ae0339e-e86a-4032-9146-a7eac1a780db
  resourceVersion: "425575"
  uid: aef62276-dc61-4ec3-9883-9f4093bb1186
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: es-quickstart
    namespace: demo
  clientConfig:
    service:
      name: es-quickstart
      port: 9200
      scheme: https
  parameters:
    apiVersion: appcatalog.appscode.com/v1alpha1
    kind: StashAddon
    stash:
      addon:
        backupTask:
          name: elasticsearch-backup-8.2.0
          params:
            - name: args
              value: --match=^(?![.])(?!apm-agent-configuration)(?!kubedb-system).+
        restoreTask:
          name: elasticsearch-restore-8.2.0
          params:
            - name: args
              value: --match=^(?![.])(?!apm-agent-configuration)(?!kubedb-system).+
  secret:
    name: es-quickstart-elastic-cred
  tlsSecret:
    name: es-quickstart-client-cert
  type: kubedb.com/elasticsearch
  version: 8.2.0

```

Here, we are going to describe the sections of an `AppBinding` crd.

### AppBinding `Spec`

An `AppBinding` object has the following fields in the `spec` section:

#### spec.type

`spec.type` is an optional field that indicates the type of the app that this `AppBinding` is pointing to. Stash uses this field to resolve the values of `TARGET_APP_TYPE`, `TARGET_APP_GROUP` and `TARGET_APP_RESOURCE` variables of [BackupBlueprint](https://appscode.com/products/stash/latest/concepts/crds/backupblueprint/) object.

This field follows the following format: `<app group>/<resource kind>`. The above AppBinding is pointing to a `elasticsearch` resource under `kubedb.com` group.

Here, the variables are parsed as follows:

|       Variable        | Usage                                                                                                                                  |
| --------------------- |----------------------------------------------------------------------------------------------------------------------------------------|
| `TARGET_APP_GROUP`    | Represents the application group where the respective app belongs (i.e: `kubedb.com`).                                                 |
| `TARGET_APP_RESOURCE` | Represents the resource under that application group that this AppBinding represents (i.e: `elasticsearch`).                           |
| `TARGET_APP_TYPE`     | Represents the complete type of the application. It's simply `TARGET_APP_GROUP/TARGET_APP_RESOURCE` (i.e: `kubedb.com/elasticsearch`). |

#### spec.secret

`spec.secret` specifies the name of the secret which contains the credentials that are required to access the database. This secret must be in the same namespace as the `AppBinding`.

This secret must contain the following keys:

|       Key           |          Usage                  |
| :----------------:  | -----------------------         |
| `username`          | Admin/Elastic user name         |
| `password`          | Admin/Elastic user password     |

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

#### spec.appRef
appRef refers to the underlying application. It has 4 fields named `apiGroup`, `kind`, `name` & `namespace`.

#### spec.tlsSecret
`spec.tlsSecret` specifies the name of the secret which contains the tls configuration files that are required to access the database. This secret must be in the same namespace as the `AppBinding`.

## Next Steps

- Learn how to use KubeDB to manage various databases [here](/docs/guides/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
