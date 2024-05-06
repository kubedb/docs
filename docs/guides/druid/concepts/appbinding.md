---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: dr-appbinding-concepts
    name: AppBinding
    parent: dr-concepts-druid
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

[//]: # (## AppBinding CRD Specification)

[//]: # ()
[//]: # (Like any official Kubernetes resource, an `AppBinding` has `TypeMeta`, `ObjectMeta` and `Spec` sections. However, unlike other Kubernetes resources, it does not have a `Status` section.)

[//]: # ()
[//]: # (An `AppBinding` object created by `KubeDB` for PostgreSQL database is shown below,)

[//]: # ()
[//]: # (```yaml)

[//]: # (apiVersion: appcatalog.appscode.com/v1alpha1)

[//]: # (kind: AppBinding)

[//]: # (metadata:)

[//]: # (  name: quick-postgres)

[//]: # (  namespace: demo)

[//]: # (  labels:)

[//]: # (    app.kubernetes.io/component: database)

[//]: # (    app.kubernetes.io/instance: quick-postgres)

[//]: # (    app.kubernetes.io/managed-by: kubedb.com)

[//]: # (    app.kubernetes.io/name: postgres)

[//]: # (    app.kubernetes.io/version: "10.2"-v2)

[//]: # (    app.kubernetes.io/name: postgreses.kubedb.com)

[//]: # (    app.kubernetes.io/instance: quick-postgres)

[//]: # (spec:)

[//]: # (  type: kubedb.com/postgres)

[//]: # (  secret:)

[//]: # (    name: quick-postgres-auth)

[//]: # (  clientConfig:)

[//]: # (    service:)

[//]: # (      name: quick-postgres)

[//]: # (      path: /)

[//]: # (      port: 5432)

[//]: # (      query: sslmode=disable)

[//]: # (      scheme: postgresql)

[//]: # (  secretTransforms:)

[//]: # (    - renameKey:)

[//]: # (        from: POSTGRES_USER)

[//]: # (        to: username)

[//]: # (    - renameKey:)

[//]: # (        from: POSTGRES_PASSWORD)

[//]: # (        to: password)

[//]: # (  version: "10.2")

[//]: # (```)

[//]: # ()
[//]: # (Here, we are going to describe the sections of an `AppBinding` crd.)

[//]: # ()
[//]: # (### AppBinding `Spec`)

[//]: # ()
[//]: # (An `AppBinding` object has the following fields in the `spec` section:)

[//]: # ()
[//]: # (#### spec.type)

[//]: # ()
[//]: # (`spec.type` is an optional field that indicates the type of the app that this `AppBinding` is pointing to. Stash uses this field to resolve the values of `TARGET_APP_TYPE`, `TARGET_APP_GROUP` and `TARGET_APP_RESOURCE` variables of [BackupBlueprint]&#40;https://appscode.com/products/stash/latest/concepts/crds/backupblueprint/&#41; object.)

[//]: # ()
[//]: # (This field follows the following format: `<app group>/<resource kind>`. The above AppBinding is pointing to a `postgres` resource under `kubedb.com` group.)

[//]: # ()
[//]: # (Here, the variables are parsed as follows:)

[//]: # ()
[//]: # (|       Variable        |                                                               Usage                                                               |)

[//]: # (| --------------------- | --------------------------------------------------------------------------------------------------------------------------------- |)

[//]: # (| `TARGET_APP_GROUP`    | Represents the application group where the respective app belongs &#40;i.e: `kubedb.com`&#41;.                                            |)

[//]: # (| `TARGET_APP_RESOURCE` | Represents the resource under that application group that this appbinding represents &#40;i.e: `postgres`&#41;.                           |)

[//]: # (| `TARGET_APP_TYPE`     | Represents the complete type of the application. It's simply `TARGET_APP_GROUP/TARGET_APP_RESOURCE` &#40;i.e: `kubedb.com/postgres`&#41;. |)

[//]: # ()
[//]: # (#### spec.secret)

[//]: # ()
[//]: # (`spec.secret` specifies the name of the secret which contains the credentials that are required to access the database. This secret must be in the same namespace as the `AppBinding`.)

[//]: # ()
[//]: # (This secret must contain the following keys:)

[//]: # ()
[//]: # (PostgreSQL :)

[//]: # ()
[//]: # (| Key                 | Usage                                               |)

[//]: # (| ------------------- | --------------------------------------------------- |)

[//]: # (| `POSTGRES_USER`     | Username of the target database.                    |)

[//]: # (| `POSTGRES_PASSWORD` | Password for the user specified by `POSTGRES_USER`. |)

[//]: # ()
[//]: # (MySQL :)

[//]: # ()
[//]: # (| Key        | Usage                                          |)

[//]: # (| ---------- | ---------------------------------------------- |)

[//]: # (| `username` | Username of the target database.               |)

[//]: # (| `password` | Password for the user specified by `username`. |)

[//]: # ()
[//]: # (MongoDB :)

[//]: # ()
[//]: # (| Key        | Usage                                          |)

[//]: # (| ---------- | ---------------------------------------------- |)

[//]: # (| `username` | Username of the target database.               |)

[//]: # (| `password` | Password for the user specified by `username`. |)

[//]: # ()
[//]: # (Elasticsearch:)

[//]: # ()
[//]: # (|       Key        |          Usage          |)

[//]: # (| ---------------- | ----------------------- |)

[//]: # (| `ADMIN_USERNAME` | Admin username          |)

[//]: # (| `ADMIN_PASSWORD` | Password for admin user |)

[//]: # ()
[//]: # (#### spec.clientConfig)

[//]: # ()
[//]: # (`spec.clientConfig` defines how to communicate with the target database. You can use either an URL or a Kubernetes service to connect with the database. You don't have to specify both of them.)

[//]: # ()
[//]: # (You can configure following fields in `spec.clientConfig` section:)

[//]: # ()
[//]: # (- **spec.clientConfig.url**)

[//]: # ()
[//]: # (  `spec.clientConfig.url` gives the location of the database, in standard URL form &#40;i.e. `[scheme://]host:port/[path]`&#41;. This is particularly useful when the target database is running outside of the Kubernetes cluster. If your database is running inside the cluster, use `spec.clientConfig.service` section instead.)

[//]: # ()
[//]: # (  > Note that, attempting to use a user or basic auth &#40;e.g. `user:password@host:port`&#41; is not allowed. Stash will insert them automatically from the respective secret. Fragments &#40;"#..."&#41; and query parameters &#40;"?..."&#41; are not allowed either.)

[//]: # ()
[//]: # (- **spec.clientConfig.service**)

[//]: # ()
[//]: # (  If you are running the database inside the Kubernetes cluster, you can use Kubernetes service to connect with the database. You have to specify the following fields in `spec.clientConfig.service` section if you manually create an `AppBinding` object.)

[//]: # ()
[//]: # (  - **name :** `name` indicates the name of the service that connects with the target database.)

[//]: # (  - **scheme :** `scheme` specifies the scheme &#40;i.e. http, https&#41; to use to connect with the database.)

[//]: # (  - **port :** `port` specifies the port where the target database is running.)

[//]: # ()
[//]: # (- **spec.clientConfig.insecureSkipTLSVerify**)

[//]: # ()
[//]: # (  `spec.clientConfig.insecureSkipTLSVerify` is used to disable TLS certificate verification while connecting with the database. We strongly discourage to disable TLS verification during backup. You should provide the respective CA bundle through `spec.clientConfig.caBundle` field instead.)

[//]: # ()
[//]: # (- **spec.clientConfig.caBundle**)

[//]: # ()
[//]: # (  `spec.clientConfig.caBundle` is a PEM encoded CA bundle which will be used to validate the serving certificate of the database.)

[//]: # (## Next Steps)

[//]: # ()
[//]: # (- Learn how to use KubeDB to manage various databases [here]&#40;/docs/guides/README.md&#41;.)

[//]: # (- Want to hack on KubeDB? Check our [contribution guidelines]&#40;/docs/CONTRIBUTING.md&#41;.)
