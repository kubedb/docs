---
title: MariaDBVersion CRD
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-concepts-mariadbversion
    name: MariaDBVersion
    parent: guides-mariadb-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MariaDBVersion

## What is MariaDBVersion

`MariaDBVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [MariaDB](https://www.mariadb.com) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `MariaDBVersion` custom resource will be created automatically for every supported MariaDB versions. You have to specify the name of `MariaDBVersion` crd in `spec.version` field of [MariaDB](/docs/guides/mariadb/concepts/mariadb) crd. Then, KubeDB will use the docker images specified in the `MariaDBVersion` crd to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator.  This will also allow the users to use a custom image for the database.

## MariaDBVersion Specification

As with all other Kubernetes objects, a MariaDBVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MariaDBVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2024-12-19T05:12:32Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2024.11.18
    helm.sh/chart: kubedb-catalog-v2024.11.18
  name: 10.5.23
  resourceVersion: "1585"
  uid: d6b463f0-032a-4832-8b87-4648da0b38a6
spec:
  archiver:
    addon:
      name: mariadb-addon
      tasks:
        fullBackup:
          name: logical-backup
        fullBackupRestore:
          name: logical-backup-restore
        manifestBackup:
          name: manifest-backup
        manifestRestore:
          name: manifest-restore
        volumeSnapshot:
          name: volume-snapshot
    walg:
      image: ghcr.io/kubedb/mariadb-archiver:v0.9.0_10.5.23-focal
  coordinator:
    image: ghcr.io/kubedb/mariadb-coordinator:v0.29.0
  db:
    image: ghcr.io/appscode-images/mariadb:10.5.23-focal
  exporter:
    image: prom/mysqld-exporter:v0.13.0
  initContainer:
    image: ghcr.io/kubedb/mariadb-init:0.5.2
  podSecurityPolicies:
    databasePolicyName: maria-db
  securityContext:
    runAsUser: 999
  stash:
    addon:
      backupTask:
        name: mariadb-backup-10.5.8
      restoreTask:
        name: mariadb-restore-10.5.8
  ui:
    - name: phpmyadmin
      version: v2024.4.27
  updateConstraints:
    allowlist:
      - '> 10.5.23, <= 11.5.2'
  version: 10.5.23
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `MariaDBVersion` crd. You have to specify this name in `spec.version` field of [MariaDB](/docs/guides/mariadb/concepts/mariadb) crd.

We follow this convention for naming MariaDBVersion crd:

- Name format: `{Original MariaDB image version}-{modification tag}`

We modify original MariaDB docker image to support additional features. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use MariaDBVersion crd with highest modification tag to take advantage of the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of MariaDB database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set `true`, KubeDB operator will not create the database and other respective resources for this version.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create Petset by KubeDB operator to create expected MariaDB database.

### spec.exporter.image

`spec.exporter.image` is a required field that specifies the image which will be used to export Prometheus metrics.

### spec.initContainer.image

`spec.initContainer.image` is a required field that specifies the image which will be used to remove `lost+found` directory and mount an `EmptyDir` data volume.

### spec.podSecurityPolicies.databasePolicyName

`spec.podSecurityPolicies.databasePolicyName` is a required field that specifies the name of the pod security policy required to get the database server pod(s) running.

### spec.stash

`spec.stash` is an optional field that specifies the name of the task for stash backup and restore. Learn more about [Stash MariaDB addon](https://stash.run/docs/v2021.03.08/addons/mariadb/)

