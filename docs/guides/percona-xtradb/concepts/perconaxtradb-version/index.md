---
title: PerconaXtraDBVersion CRD
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-concepts-perconaxtradbversion
    name: PerconaXtraDBVersion
    parent: guides-perconaxtradb-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PerconaXtraDBVersion

## What is PerconaXtraDBVersion

`PerconaXtraDBVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [PerconaXtraDB](https://docs.percona.com/percona-xtradb-cluster/8.0/) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `PerconaXtraDBVersion` custom resource will be created automatically for every supported PerconaXtraDB versions. You have to specify the name of `PerconaXtraDBVersion` crd in `spec.version` field of [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb) crd. Then, KubeDB will use the docker images specified in the `PerconaXtraDBVersion` crd to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator.  This will also allow the users to use a custom image for the database.

## PerconaXtraDBVersion Specification

As with all other Kubernetes objects, a PerconaXtraDBVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: PerconaXtraDBVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2022-12-19T09:39:14Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2022.12.13-rc.0
    helm.sh/chart: kubedb-catalog-v2022.12.13-rc.0
  name: 8.0.26
  resourceVersion: "1611"
  uid: 38161f93-0501-4caf-98a5-4d8d168951ca
spec:
  coordinator:
    image: kubedb/percona-xtradb-coordinator:v0.3.0-rc.0
  db:
    image: percona/percona-xtradb-cluster:8.0.26
  exporter:
    image: prom/mysqld-exporter:v0.13.0
  initContainer:
    image: kubedb/percona-xtradb-init:0.2.0
  podSecurityPolicies:
    databasePolicyName: percona-xtradb-db
  stash:
    addon:
      backupTask:
        name: perconaxtradb-backup-5.7
      restoreTask:
        name: perconaxtradb-restore-5.7
  version: 8.0.26
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `PerconaXtraDBVersion` crd. You have to specify this name in `spec.version` field of [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb) crd.

We follow this convention for naming PerconaXtraDBVersion crd:

- Name format: `{Original PerconaXtraDB image version}-{modification tag}`

We modify original PerconaXtraDB docker image to support additional features. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use PerconaXtraDBVersion crd with highest modification tag to take advantage of the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of PerconaXtraDB database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set `true`, KubeDB operator will not create the database and other respective resources for this version.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create Statefulset by KubeDB operator to create expected PerconaXtraDB database.

### spec.exporter.image

`spec.exporter.image` is a required field that specifies the image which will be used to export Prometheus metrics.

### spec.initContainer.image

`spec.initContainer.image` is a required field that specifies the image which will be used to remove `lost+found` directory and mount an `EmptyDir` data volume.

### spec.podSecurityPolicies.databasePolicyName

`spec.podSecurityPolicies.databasePolicyName` is a required field that specifies the name of the pod security policy required to get the database server pod(s) running.

### spec.stash

`spec.stash` is an optional field that specifies the name of the task for stash backup and restore. Learn more about [Stash PerconaXtraDB addon](https://stash.run/docs/v2022.12.11/addons/percona-xtradb/)

