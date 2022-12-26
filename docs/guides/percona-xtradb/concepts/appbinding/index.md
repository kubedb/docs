---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-concepts-appbinding
    name: AppBinding
    parent: guides-perconaxtradb-concepts
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# AppBinding

## What is AppBinding

An `AppBinding` is a Kubernetes `CustomResourceDefinition`(CRD) which points to an application using either its URL (usually for a non-Kubernetes resident service instance) or a Kubernetes service object (if self-hosted in a Kubernetes cluster), some optional parameters and a credential secret. To learn more about AppBinding and the problems it solves, please read this blog post: [The case for AppBinding](https://blog.byte.builders/post/the-case-for-appbinding).

If you deploy a database using [KubeDB](https://kubedb.com/docs/0.11.0/concepts/), `AppBinding` object will be created automatically for it. Otherwise, you have to create an `AppBinding` object manually pointing to your desired database.

KubeDB uses [Stash](https://appscode.com/products/stash/) to perform backup/recovery of databases. Stash needs to know how to connect with a target database and the credentials necessary to access it. This is done via an `AppBinding`.

## AppBinding CRD Specification

Like any official Kubernetes resource, an `AppBinding` has `TypeMeta`, `ObjectMeta` and `Spec` sections. However, unlike other Kubernetes resources, it does not have a `Status` section.

An `AppBinding` object created by `KubeDB` for PerconaXtraDB database is shown below,

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-pxc
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: perconaxtradbs.kubedb.com
  name: sample-pxc
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJekNDQWd1Z0F3SUJBZ0lVVUg1V24wOSt6MnR6RU5ESnF4N1AxZFg5aWM4d0RRWUpLb1pJaHZjTkFRRUwKQlFBd0lURU9NQXdHQTFVRUF3d0ZiWGx6Y1d3eER6QU5CZ05WQkFvTUJtdDFZbVZrWWpBZUZ3MHlNVEF5TURrdwpPVEkxTWpCYUZ3MHlNakF5TURrd09USTFNakJhTUNFeERqQU1CZ05WQkFNTUJXMTVjM0ZzTVE4d0RRWURWUVFLCkRBWnJkV0psWkdJd2dnRWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUUM3ZDl5YUtMQ3UKYy9NclRBb0NkV1VORld3ckdqbVdvUEVTRWNMR0pjT0JVSTZ5NXZ5QXVGMG1TakZvNzR3SEdSbWRmS2ExMWh0Ygo4TWZ2UFNwWXVGWFpUSi9GbnkwNnU2ekZMVm5xV2h3MUdiZ2ZCUE5XK0w1ZGkzZmVjanBEZmtLbTcrd0ZUVnNmClVzWGVVcUR0VHFpdlJHVUQ5aURUTzNTUmdheVI5U0J0RnRxcHRtV0YrODFqZGlSS2pRTVlCVGJ2MDRueW9UdHUKK0hJZlFjbE40Q1p3NzJPckpUdFdiYnNiWHVlMU5RZU9nQzJmSVhkZEF0WEkxd3lOT04zckxuTFF1SUIrakVLSQpkZTlPandKSkJhSFVzRVZEbllnYlJLSTdIcVdFdk5kL29OU2VZRXF2TXk3K1hwTFV0cDBaVXlxUDV2cC9PSXJ3CmlVMWVxZGNZMzJDcEFnTUJBQUdqVXpCUk1CMEdBMVVkRGdRV0JCUlNnNDVpazFlT3lCU1VKWHkvQllZVDVLck8KeWpBZkJnTlZIU01FR0RBV2dCUlNnNDVpazFlT3lCU1VKWHkvQllZVDVLck95akFQQmdOVkhSTUJBZjhFQlRBRApBUUgvTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFCNTlhNlFGQm1YMTh1b1dzQ3dGT1Y0R25GYnVBL2NoaVN6CkFwRVVhcjI1L2RNK2RpT0RVNkJuMXM3Wmpqem45WU9aVVJJL3UyRGRhdmNnQUNYR2FXWHJhSS84UUYxZDB3OGsKNXFlRmMxQWE5UEhFeEsxVm1xb21MV2xhMkdTOW1EbFpEOEtueDdjU3lpRmVqRXJYdWtld1B1VXA0dUUzTjAraApwQjQ0MDVRU2d4VVc3SmVhamFQdTNXOHgyRUFKMnViTkdMVEk5L0x4V1Z0YkxGcUFoSFphbGRjaXpOSHdTUGYzCkdMWEo3YTBWTW1JY0NuMWh5a0k2UkNrUTRLSE9tbDNOcXRRS2F5RnhUVHVpdzRiZ3k3czA1UnNzRlVUaWN1VmcKc3hnMjFVQUkvYW9WaXpQOVpESGE2TmV0YnpNczJmcmZBeHhBZk9pWDlzN1JuTmM0WHd4VAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    service:
      name: sample-pxc
      port: 3306
      scheme: mysql
  secret:
    name: sample-pxc-auth
  type: kubedb.com/perconaxtradb
  version: 8.0.26
```

Here, we are going to describe the sections of an `AppBinding` crd.

### AppBinding `Spec`

An `AppBinding` object has the following fields in the `spec` section:

#### spec.type

`spec.type` is an optional field that indicates the type of the app that this `AppBinding` is pointing to. Stash uses this field to resolve the values of `TARGET_APP_TYPE`, `TARGET_APP_GROUP` and `TARGET_APP_RESOURCE` variables of [BackupBlueprint](https://appscode.com/products/stash/latest/concepts/crds/backupblueprint/) object.

This field follows the following format: `<app group>/<resource kind>`. The above AppBinding is pointing to a `perconaxtradb` resource under `kubedb.com` group.

Here, the variables are parsed as follows:

|       Variable        |                                                               Usage                                                               |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `TARGET_APP_GROUP`    | Represents the application group where the respective app belongs (i.e: `kubedb.com`).                                            |
| `TARGET_APP_RESOURCE` | Represents the resource under that application group that this appbinding represents (i.e: `perconaxtradb`).                            |
| `TARGET_APP_TYPE`     | Represents the complete type of the application. It's simply `TARGET_APP_GROUP/TARGET_APP_RESOURCE` (i.e: `kubedb.com/perconaxtradb`).  |

#### spec.secret

`spec.secret` specifies the name of the secret which contains the credentials that are required to access the database. This secret must be in the same namespace as the `AppBinding`.

This secret must contain the following keys:

PostgreSQL :

| Key                 | Usage                                               |
| ------------------- | --------------------------------------------------- |
| `POSTGRES_USER`     | Username of the target database.                    |
| `POSTGRES_PASSWORD` | Password for the user specified by `POSTGRES_USER`. |

MySQL :

| Key        | Usage                                          |
| ---------- | ---------------------------------------------- |
| `username` | Username of the target database.               |
| `password` | Password for the user specified by `username`. |

PerconaXtraDB :

| Key        | Usage                                          |
| ---------- | ---------------------------------------------- |
| `username` | Username of the target database.               |
| `password` | Password for the user specified by `username`. |

MongoDB :

| Key        | Usage                                          |
| ---------- | ---------------------------------------------- |
| `username` | Username of the target database.               |
| `password` | Password for the user specified by `username`. |

PerconaXtraDB:

|       Key        |          Usage          |
| ---------------- | ----------------------- |
| `ADMIN_USERNAME` | Admin username          |
| `ADMIN_PASSWORD` | Password for admin user |

#### spec.clientConfig

`spec.clientConfig` defines how to communicate with the target database. You can use either an URL or a Kubernetes service to connect with the database. You don't have to specify both of them.

You can configure following fields in `spec.clientConfig` section:

- **spec.clientConfig.url**

  `spec.clientConfig.url` gives the location of the database, in standard URL form (i.e. `[scheme://]host:port/[path]`). This is particularly useful when the target database is running outside of the Kubernetes cluster. If your database is running inside the cluster, use `spec.clientConfig.service` section instead.
  Note that, attempting to use a user or basic auth (e.g. `user:password@host:port`) is not allowed. Stash will insert them automatically from the respective secret. Fragments ("#...") and query parameters ("?...") are not allowed either.

- **spec.clientConfig.service**

  If you are running the database inside the Kubernetes cluster, you can use Kubernetes service to connect with the database. You have to specify the following fields in `spec.clientConfig.service` section if you manually create an `AppBinding` object.

  - **name :** `name` indicates the name of the service that connects with the target database.
  - **scheme :** `scheme` specifies the scheme (i.e. http, https) to use to connect with the database.
  - **port :** `port` specifies the port where the target database is running.

- **spec.clientConfig.insecureSkipTLSVerify**

  `spec.clientConfig.insecureSkipTLSVerify` is used to disable TLS certificate verification while connecting with the database. We strongly discourage to disable TLS verification during backup. You should provide the respective CA bundle through `spec.clientConfig.caBundle` field instead.

- **spec.clientConfig.caBundle**

  `spec.clientConfig.caBundle` is a PEM encoded CA bundle which will be used to validate the serving certificate of the database.
