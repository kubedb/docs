---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: rd-appbinding-concepts
    name: AppBinding
    parent: rd-concepts-redis
    weight: 30
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

An `AppBinding` object created by `KubeDB` for Redis database is shown below,

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Redis","metadata":{"annotations":{},"name":"redis1","namespace":"demo"},"spec":{"authSecret":{"externallyManaged":false,"name":"redis1-auth"},"autoOps":{"disabled":true},"cluster":{"master":3,"replicas":1},"configSecret":{"name":"rd-custom-config"},"disableAuth":false,"halted":false,"healthChecker":{"disableWriteCheck":false,"failureThreshold":2,"periodSeconds":15,"timeoutSeconds":10},"mode":"Cluster","monitor":{"agent":"prometheus.io/operator","prometheus":{"serviceMonitor":{"interval":"10s","labels":{"app":"kubedb"}}}},"podTemplate":{"controller":{"annotations":{"passMe":"ToPetSet"}},"metadata":{"annotations":{"passMe":"ToDatabasePod"}},"spec":{"args":["--loglevel verbose"],"env":[{"name":"ENV_VARIABLE","value":"value"}],"imagePullSecrets":[{"name":"regcred"}],"resources":{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"250m","memory":"64Mi"}},"serviceAccountName":"my-service-account"}},"serviceTemplates":[{"alias":"primary","metadata":{"annotations":{"passMe":"ToService"}},"spec":{"ports":[{"name":"http","port":9200}],"type":"NodePort"}}],"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"deletionPolicy":"Halt","tls":{"certificates":[{"alias":"client","emailAddresses":["abc@appscode.com"],"subject":{"organizations":["kubedb"]}},{"alias":"server","emailAddresses":["abc@appscode.com"],"subject":{"organizations":["kubedb"]}}],"issuerRef":{"apiGroup":"cert-manager.io","kind":"Issuer","name":"redis-ca-issuer"}},"version":"6.2.14"}}
  creationTimestamp: "2023-02-01T05:27:19Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: redis1
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: redises.kubedb.com
  name: redis1
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: Redis
      name: redis1
      uid: a01272d3-97b6-4e8c-912f-67eff07e3811
  resourceVersion: "398775"
  uid: 336988b4-5805-48ac-9d06-e3375fa4c435
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Redis
    name: redis1
    namespace: demo
  clientConfig:
    service:
      name: redis1
      port: 6379
      scheme: rediss
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    clientCertSecret:
      name: redis1-client-cert
    kind: RedisConfiguration
    stash:
      addon:
        backupTask:
          name: redis-backup-6.2.5
        restoreTask:
          name: redis-restore-6.2.5
  secret:
    name: redis1-auth
  tlsSecret:
    name: redis1-client-cert
  type: kubedb.com/redis
  version: 6.2.14

```

Here, we are going to describe the sections of an `AppBinding` crd.

### AppBinding `Spec`

An `AppBinding` object has the following fields in the `spec` section:

#### spec.type

`spec.type` is an optional field that indicates the type of the app that this `AppBinding` is pointing to. Stash uses this field to resolve the values of `TARGET_APP_TYPE`, `TARGET_APP_GROUP` and `TARGET_APP_RESOURCE` variables of [BackupBlueprint](https://appscode.com/products/stash/latest/concepts/crds/backupblueprint/) object.

This field follows the following format: `<app group>/<resource kind>`. The above AppBinding is pointing to a `redis` resource under `kubedb.com` group.

Here, the variables are parsed as follows:

|       Variable        | Usage                                                                                                                          |
| --------------------- |--------------------------------------------------------------------------------------------------------------------------------|
| `TARGET_APP_GROUP`    | Represents the application group where the respective app belongs (i.e: `kubedb.com`).                                         |
| `TARGET_APP_RESOURCE` | Represents the resource under that application group that this appbinding represents (i.e: `redis`).                           |
| `TARGET_APP_TYPE`     | Represents the complete type of the application. It's simply `TARGET_APP_GROUP/TARGET_APP_RESOURCE` (i.e: `kubedb.com/redis`). |

#### spec.secret

`spec.secret` specifies the name of the secret which contains the credentials that are required to access the database. This secret must be in the same namespace as the `AppBinding`.

This secret must contain the following keys:


Redis :

| Key        | Usage                                          |
| ---------- | ---------------------------------------------- |
| `username` | Username of the target database.               |
| `password` | Password for the user specified by `username`. |

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

MongoDB :

| Key        | Usage                                          |
| ---------- | ---------------------------------------------- |
| `username` | Username of the target database.               |
| `password` | Password for the user specified by `username`. |

Elasticsearch:

|       Key        |          Usage          |
| ---------------- | ----------------------- |
| `ADMIN_USERNAME` | Admin username          |
| `ADMIN_PASSWORD` | Password for admin user |

#### spec.appRef
appRef refers to the underlying application. It has 4 fields named `apiGroup`, `kind`, `name` & `namespace`.

#### spec.clientConfig

`spec.clientConfig` defines how to communicate with the target database. You can use either a URL or a Kubernetes service to connect with the database. You don't have to specify both of them.

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
