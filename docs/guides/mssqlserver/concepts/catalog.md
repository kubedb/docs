---
title: MSSQLServerVersion CRD
menu:
  docs_{{ .version }}:
    identifier: ms-concepts-catalog
    name: MSSQLServerVersion
    parent: ms-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQLServerVersion

## What is MSSQLServerVersion

`MSSQLServerVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [Microsoft SQL Server](https://mcr.microsoft.com/en-us/product/mssql/server/about) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `MSSQLServerVersion` custom resource will be created automatically for every supported MSSQLServer versions. You have to specify the name of `MSSQLServerVersion` CR in `spec.version` field of [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md) CR. Then, KubeDB will use the docker images specified in the `MSSQLServerVersion` CR to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator. This will also allow the users to use a custom image for the database. For more details about how to use custom image with SQL Server in KubeDB, please visit [here](/docs/guides/mssqlserver/custom-versions/setup.md).

## MSSQLServerVersion Specification

As with all other Kubernetes objects, a MSSQLServerVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MSSQLServerVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb-catalog
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2024-10-14T09:42:21Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb-catalog
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2024.9.30
    helm.sh/chart: kubedb-catalog-v2024.9.30
  name: 2022-cu14
  resourceVersion: "379704"
  uid: 1213b070-5eaa-456b-afea-d9ea7b937837
spec:
  archiver:
    addon:
      name: mssqlserver-addon
      tasks:
        fullBackup:
          name: logical-backup
        fullBackupRestore:
          name: logical-backup-restore
        manifestBackup:
          name: manifest-backup
        manifestRestore:
          name: manifest-restore
    walg:
      image: ghcr.io/kubedb/mssqlserver-archiver:v0.0.1
  coordinator:
    image: ghcr.io/kubedb/mssql-coordinator:v0.3.0
  db:
    image: mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04
  exporter:
    image: ghcr.io/kubedb/mssql-exporter:1.1.0
  initContainer:
    image: ghcr.io/kubedb/mssql-init:2022-ubuntu-22-v3
  securityContext:
    runAsUser: 10001
  version: "2022"
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `MSSQLServerVersion` CR. You have to specify this name in `spec.version` field of [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md) CR.

We follow this convention for naming MSSQLServerVersion CR:
- Name format: `{Original sql server image version year}-{cu-number}`

We modify original MS SQL Server docker image to support additional features like WAL archiving, clustering etc. and re-tag the image with v1, v2 etc. modification tag. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use MSSQLServerVersion CR with highest modification tag to take advantage of the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of MSSQLServer database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set `true`, KubeDB operator will not create the database and other respective resources for this version.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create Petset by KubeDB operator to create expected MSSQLServer database.

### spec.exporter.image

`spec.exporter.image` is a required field that specifies the image which will be used to export Prometheus metrics.

## Next Steps

- Learn about MSSQLServer CR [here](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Deploy your first MSSQLServer database with KubeDB by following the guide [here](/docs/guides/mssqlserver/quickstart/quickstart.md).