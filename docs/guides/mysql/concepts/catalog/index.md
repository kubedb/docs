---
title: MySQLVersion CRD
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-concepts-catalog
    name: MySQLVersion
    parent: guides-mysql-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQLVersion

## What is MySQLVersion

`MySQLVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [MySQL](https://www.mysql.com) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `MySQLVersion` custom resource will be created automatically for every supported MySQL versions. You have to specify the name of `MySQLVersion` crd in `spec.version` field of [MySQL](/docs/guides/mysql/concepts/catalog/index.md) crd. Then, KubeDB will use the docker images specified in the `MySQLVersion` crd to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator.  This will also allow the users to use a custom image for the database.

## MySQLVersion Specification

As with all other Kubernetes objects, a MySQLVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MySQLVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb-catalog
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2022-06-16T13:52:58Z"
  generation: 3
  labels:
    app.kubernetes.io/instance: kubedb-catalog
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2022.03.28
    helm.sh/chart: kubedb-catalog-v2022.03.28
  name: 8.0.29
  resourceVersion: "1575483"
  uid: 4e605d5f-a6f0-42cb-a125-b4b4fd02e41e
spec:
  coordinator:
    image: kubedb/mysql-coordinator:v0.4.0-2-g49a2d26-dirty_linux_amd64
  db:
    image: mysql:8.0.29
  distribution: Official
  exporter:
    image: kubedb/mysqld-exporter:v0.13.1
  initContainer:
    image: kubedb/mysql-init:8.0.29_linux_amd64
  podSecurityPolicies:
    databasePolicyName: mysql-db
  replicationModeDetector:
    image: kubedb/replication-mode-detector:v0.13.0
  stash:
    addon:
      backupTask:
        name: mysql-backup-8.0.21
      restoreTask:
        name: mysql-restore-8.0.21
  updateConstraints:
    denylist:
      groupReplication:
      - < 8.0.29
      standalone:
      - < 8.0.29
  version: 8.0.29
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `MySQLVersion` crd. You have to specify this name in `spec.version` field of [MySQL](/docs/guides/mysql/concepts/database/index.md) crd.

We follow this convention for naming MySQLVersion crd:

- Name format: `{Original MySQL image version}-{modification tag}`

We modify original MySQL docker image to support additional features. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use MySQLVersion crd with highest modification tag to take advantage of the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of MySQL database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator. For example, we have modified `kubedb/mysql:8.0` docker image to support custom configuration and re-tagged as `kubedb/mysql:8.0-v2`. Now, KubeDB `0.9.0-rc.0` supports providing custom configuration which required `kubedb/mysql:8.0-v2` docker image. So, we have marked `kubedb/mysql:8.0` as deprecated for KubeDB `0.9.0-rc.0`.

The default value of this field is `false`. If `spec.deprecated` is set `true`, KubeDB operator will not create the database and other respective resources for this version.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create Statefulset by KubeDB operator to create expected MySQL database.

### spec.exporter.image

`spec.exporter.image` is a required field that specifies the image which will be used to export Prometheus metrics.

### spec.initContainer.image

`spec.initContainer.image` is a required field that specifies the image which will be used to remove `lost+found` directory and mount an `EmptyDir` data volume.

### spec.replicationModeDetector.image

`spec.replicationModeDetector.image` is only required field for MySQL Group Replication. This field specifies that image which will be used to detect primary member/replica/node in Group Replication.

### spec.tools.image

`spec.tools.image` is a optional field that specifies the image which will be used to take backup and initialize database from snapshot.

### spec.updateConstraints

`spec.updateConstraints` specifies a specific database version update constraints in a mathematical expression that describes whether it is possible or not to update from the current version to any other valid version. This field consists of the following sub-fields:

- `denylist` specifies that it is not possible to update from the current version to any other version. This field has two sub-fields:
  - `groupReplication` : Suppose you have an expression like, `< 8.0.21` under `groupReplication`, it indicates that it's not possible to update from the current version to any other lower version `8.0.21` for group replication.
  - `standalone`: Suppose you have an expression like, `< 8.0.21` under `standalone`, it indicates that it's not possible to update from the current version to any other lower version `8.0.21` for standalone.
- `allowlist` specifies that it is possible to update from the current version to any other version. This field has two sub-fields:
  - `groupReplication` : Suppose you have an expression like, `8.0.3`, it indicates that it's possible to update from the current version to `8.0.3` for group replication.
  - `standalone`: Suppose you have an expression like, `8.0.3`, it indicates that it's possible to update from the current version to `8.0.3` for standalone.

### spec.podSecurityPolicies.databasePolicyName

`spec.podSecurityPolicies.databasePolicyName` is a required field that specifies the name of the pod security policy required to get the database server pod(s) running.

## Next Steps

- Learn about MySQL crd [here](/docs/guides/mysql/concepts/database/index.md).
- Deploy your first MySQL database with KubeDB by following the guide [here](/docs/guides/mysql/quickstart/index.md).
