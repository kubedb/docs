---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: ms-concepts-appbinding
    name: AppBinding
    parent: ms-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# AppBinding

## What is AppBinding

An `AppBinding` is a Kubernetes `CustomResourceDefinition`(CRD) which points to an application using either its URL (usually for a non-Kubernetes resident service instance) or a Kubernetes service object (if self-hosted in a Kubernetes cluster), some optional parameters and a credential secret. To learn more about AppBinding and the problems it solves, please read this blog post: [The case for AppBinding](https://appscode.com/blog/post/the-case-for-appbinding).

If you deploy a database using [KubeDB](https://kubedb.com/docs/latest/welcome/), `AppBinding` object will be created automatically for it. Otherwise, you have to create an `AppBinding` object manually pointing to your desired database.

KubeDB uses [KubeStash](https://kubestash.com/) to perform backup/recovery of databases. KubeStash needs to know how to connect with a target database and the credentials necessary to access it. This is done via an `AppBinding`.

## AppBinding CRD Specification

Like any official Kubernetes resource, an `AppBinding` has `TypeMeta`, `ObjectMeta` and `Spec` sections. However, unlike other Kubernetes resources, it does not have a `Status` section.

An `AppBinding` object created by `KubeDB` for MSSQLServer database is shown below,

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  creationTimestamp: "2024-10-14T10:13:19Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: mssqlserver
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mssqlservers.kubedb.com
  name: mssqlserver
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: MSSQLServer
      name: mssqlserver
      uid: df41bddb-dfa0-4bfb-bac6-39e124722f28
  resourceVersion: "382556"
  uid: 9d64bc4b-1926-4faa-ac5c-36561471c98a
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MSSQLServer
    name: mssqlserver
    namespace: demo
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURMekNDQWhlZ0F3SUJBZ0lVVG9Sb1Q3KzZ5OUZ5OURrVml1aWFQR2c0TmRvd0RRWUpLb1pJaHZjTkFRRUwKQlFBd0p6RVVNQklHQTFVRUF3d0xUVk5UVVV4VFpYSjJaWEl4RHpBTkJnTlZCQW9NQm10MVltVmtZakFlRncweQpOREV3TURneE1EVTBNekZhRncweU5URXdNRGd4TURVME16RmFNQ2N4RkRBU0JnTlZCQU1NQzAxVFUxRk1VMlZ5CmRtVnlNUTh3RFFZRFZRUUtEQVpyZFdKbFpHSXdnZ0VpTUEwR0NTcUdTSWIzRFFFQkFRVUFBNElCRHdBd2dnRUsKQW9JQkFRREt1Yk1LUHNZVDA5aUZGdS92NGhhMFNpSG05NThxcTlkRlpkNWFaTVVkTTgxanA2Y1oyZ0IrVlJXUgpLVlpGa20vNmRIamdyZnJCelhkMHRQdW1NZ3plRU9WU00yckxWY1lISytZWjh2SWMxYmNuVWlCaCtJU2FCb3UrClZaUGRlZGpienNhb0ozSW9xY3IyWjBtVERYMlJ6TmRWK3E3M3BTc3Q4UnZaQWtLU3BGZ2R3Y1p3dkJlYmduOFkKUm13ZFhzZlNGS0NhU0pORENlTjFEdEtZTGVSRndRV2JwdTBZM0VVUGFpcC9xaDZvNTdKUUhTaFlTQzVoVndSUgpqSzRwTDVVVnlia096MHlTeFVqQzI2MGJnWWNPWmE2OElMdDJWNEVDNVpzRkZNTWc5S3JzcUhvRnRDQmw0ZVNNCmFlaXhQWEljWEVkTktGTWJRazdySFVXU1hSanZBZ01CQUFHalV6QlJNQjBHQTFVZERnUVdCQlIxdDVDQnBhY00KZVY3dVJFUHAwUG91d0t1aGxUQWZCZ05WSFNNRUdEQVdnQlIxdDVDQnBhY01lVjd1UkVQcDBQb3V3S3VobFRBUApCZ05WSFJNQkFmOEVCVEFEQVFIL01BMEdDU3FHU0liM0RRRUJDd1VBQTRJQkFRQjJSaWhSYU4zQStXR1JQbHhqCkZOMzVwOVZtTnduSjRHcEZSOVF0blg5bVNhdkovS1VGUXAwbXVEK29pT0FHYkJvWENBSXBwUTBsOXFCWDV1N3UKeWxjQzN5Q2Z4UHhZS1NiMjNyL2FqQ3J3T3E2Y0xMU3RvaGFDRnhNdHRBOUoxUldURHp0U2E0SER5VWJCUllOUwowb1BBOThEYkJ2clBFVU11ajlUbUZpS0wwQmhHakx0bmttQmJGdkUrZkZ4RWltMzE2ai84TjZlT0xkQlJQUmIwCjhQMmN6dW9wbGFMcDNCUmttR3A2cDJKd1BwNXUrZlZ5UE9wbWhPWW5hTEZaYXdHTHUzS0c4amViZ1psTllJU2MKdUFNZWpqNW1rT0l5VFl0NFJZSThBeW11bWZsYjZBM1g0YkZhS1hsVUdKQlRjWTJGakZ5VkFHcWxTTGJ1ZkZPYQovbXh6Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    service:
      name: mssqlserver
      path: /
      port: 1433
      scheme: tcp
  secret:
    name: mssqlserver-auth
  tlsSecret:
    name: mssqlserver-client-cert
  type: kubedb.com/mssqlserver
  version: "2022"
```

Here, we are going to describe the sections of an `AppBinding` CR.

### AppBinding `Spec`

An `AppBinding` object has the following fields in the `spec` section:

#### spec.type

`spec.type` is an optional field that indicates the type of the app that this `AppBinding` is pointing to. KubeStash uses this field to resolve the values of `TARGET_APP_TYPE`, `TARGET_APP_GROUP` and `TARGET_APP_RESOURCE` variables of [BackupBlueprint](https://appscode.com/products/KubeStash/latest/concepts/crds/backupblueprint/) object.

This field follows the following format: `<app group>/<resource kind>`. The above AppBinding is pointing to a `mssqlserver` resource under `kubedb.com` group.

Here, the variables are parsed as follows:

|       Variable        |                                                               Usage                                                               |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `TARGET_APP_GROUP`    | Represents the application group where the respective app belongs (i.e: `kubedb.com`).                                            |
| `TARGET_APP_RESOURCE` | Represents the resource under that application group that this appbinding represents (i.e: `mssqlserver`).                           |
| `TARGET_APP_TYPE`     | Represents the complete type of the application. It's simply `TARGET_APP_GROUP/TARGET_APP_RESOURCE` (i.e: `kubedb.com/mssqlserver`). |

#### spec.secret

`spec.secret` specifies the name of the secret which contains the credentials that are required to access the database. This secret must be in the same namespace as the `AppBinding`.

This secret must contain the following keys:

MSSQLServer :

| Key         | Usage                                          |
|-------------|------------------------------------------------|
| `username`  | Username of the target database.               |
| `password`  | Password for the user specified by `username`. |


#### spec.clientConfig

`spec.clientConfig` defines how to communicate with the target database. You can use either an URL or a Kubernetes service to connect with the database. You don't have to specify both of them.

You can configure following fields in `spec.clientConfig` section:

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
