---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: guides-druid-concepts-appbinding
    name: AppBinding
    parent: guides-druid-concepts
    weight: 20
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

An `AppBinding` object created by `KubeDB` for Druid database is shown below,

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Druid","metadata":{"annotations":{},"name":"druid-quickstart","namespace":"demo"},"spec":{"deepStorage":{"configSecret":{"name":"deep-storage-config"},"type":"s3"},"topology":{"routers":{"replicas":1}},"version":"28.0.1"}}
  creationTimestamp: "2024-10-16T13:28:40Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: druid-quickstart
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: druids.kubedb.com
  name: druid-quickstart
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Druid
    name: druid-quickstart
    uid: 06dc7c5f-65ad-4310-a203-b18c0d33d662
  resourceVersion: "45154"
  uid: 58861709-99f9-4c78-8cf9-b5dc6534102e
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Druid
    name: druid-quickstart
    namespace: demo
  clientConfig:
    caBundle: dGhpcyBpcyBub3QgYSBjZXJ0
    service:
      name: druid-quickstart-pods
      port: 8888
      scheme: http
    url: http://druid-quickstart-coordinators-0.druid-quickstart-pods.demo.svc.cluster.local:8081,http://druid-quickstart-overlords-0.druid-quickstart-pods.demo.svc.cluster.local:8090,http://druid-quickstart-middlemanagers-0.druid-quickstart-pods.demo.svc.cluster.local:8091,http://druid-quickstart-historicals-0.druid-quickstart-pods.demo.svc.cluster.local:8083,http://druid-quickstart-brokers-0.druid-quickstart-pods.demo.svc.cluster.local:8082,http://druid-quickstart-routers-0.druid-quickstart-pods.demo.svc.cluster.local:8888
  secret:
    name: druid-quickstart-admin-cred
  tlsSecret:
    name: druid-client-cert
  type: kubedb.com/druid
  version: 28.0.1
```
Here, we are going to describe the sections of an `AppBinding` crd.

### AppBinding `Spec`

An `AppBinding` object has the following fields in the `spec` section:

#### spec.type

`spec.type` is an optional field that indicates the type of the app that this `AppBinding` is pointing to.

<!--- Add when Stash support is added --->
<!---
Stash uses this field to resolve the values of `TARGET_APP_TYPE`, `TARGET_APP_GROUP` and `TARGET_APP_RESOURCE` variables of [BackupBlueprint](https://appscode.com/products/stash/latest/concepts/crds/backupblueprint/) object.

This field follows the following format: `<app group>/<resource kind>`. The above AppBinding is pointing to a `druid` resource under `kubedb.com` group.

Here, the variables are parsed as follows:

|       Variable        | Usage                                                                                                                          |
| --------------------- |--------------------------------------------------------------------------------------------------------------------------------|
| `TARGET_APP_GROUP`    | Represents the application group where the respective app belongs (i.e: `kubedb.com`).                                         |
| `TARGET_APP_RESOURCE` | Represents the resource under that application group that this appbinding represents (i.e: `druid`).                           |
| `TARGET_APP_TYPE`     | Represents the complete type of the application. It's simply `TARGET_APP_GROUP/TARGET_APP_RESOURCE` (i.e: `kubedb.com/druid`). |

--->

#### spec.secret

`spec.secret` specifies the name of the secret which contains the credentials that are required to access the database. This secret must be in the same namespace as the `AppBinding`.

This secret must contain the following keys for Druid:

| Key        | Usage                                          |
| ---------- |------------------------------------------------|
| `username` | Username of the target Druid instance.         |
| `password` | Password for the user specified by `username`. |


#### spec.appRef
appRef refers to the underlying application. It has 4 fields named `apiGroup`, `kind`, `name` & `namespace`.

#### spec.clientConfig

`spec.clientConfig` defines how to communicate with the target database. You can use either a URL or a Kubernetes service to connect with the database. You don't have to specify both of them.

You can configure following fields in `spec.clientConfig` section:

- **spec.clientConfig.url**

  `spec.clientConfig.url` gives the location of the database, in standard URL form (i.e. `[scheme://]host:port/[path]`). This is particularly useful when the target database is running outside the Kubernetes cluster. If your database is running inside the cluster, use `spec.clientConfig.service` section instead.

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
