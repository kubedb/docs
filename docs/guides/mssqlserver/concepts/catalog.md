---
title: MSSQLServerVersion CRD
menu:
  docs_{{ .version }}:
    identifier: ms-catalog-concepts
    name: MSSQLServerVersion
    parent: ms-concepts-mssqlserver
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQLServerVersion

## What is MSSQLServerVersion

`MSSQLServerVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [MSSQLServer](https://www.mssqlserverql.org/) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `MSSQLServerVersion` custom resource will be created automatically for every supported MSSQLServer versions. You have to specify the name of `MSSQLServerVersion` crd in `spec.version` field of [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md) crd. Then, KubeDB will use the docker images specified in the `MSSQLServerVersion` crd to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator. This will also allow the users to use a custom image for the database. For more details about how to use custom image with MSSQLServer in KubeDB, please visit [here](/docs/guides/mssqlserver/custom-versions/setup.md).

## MSSQLServerVersion Specification

As with all other Kubernetes objects, a MSSQLServerVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MSSQLServerVersion
metadata:
  name: "13.13"
spec:
  coordinator:
    image: kubedb /ms-coordinator:v0.1.0
  db:
    image: mssqlserver:13.2-alpine
  distribution: MSSQLServer
  exporter:
    image: prometheuscommunity/mssqlserver-exporter:v0.9.0
  initContainer:
    image: kubedb/mssqlserver-init:0.1.0
  podSecurityPolicies:
    databasePolicyName: mssqlserver-db
  stash:
    addon:
      backupTask:
        name: mssqlserver-backup-13.1
      restoreTask:
        name: mssqlserver-restore-13.1
  version: "13.13"
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `MSSQLServerVersion` crd. You have to specify this name in `spec.version` field of [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md) crd.

We follow this convention for naming MSSQLServerVersion crd:
- Name format: `{Original MSSQLServer image version}-{modification tag}`

We modify original MSSQLServer docker image to support additional features like WAL archiving, clustering etc. and re-tag the image with v1, v2 etc. modification tag. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use MSSQLServerVersion crd with highest modification tag to take advantage of the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of MSSQLServer database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator. For example, we have modified `kubedb/mssqlserver:10.2` docker image to support custom configuration and re-tagged as `kubedb/mssqlserver:10.2-v2`. Now, KubeDB `0.9.0-rc.0` supports providing custom configuration which required `kubedb/mssqlserver:10.2-v2` docker image. So, we have marked `kubedb/mssqlserver:10.2` as deprecated in KubeDB `0.9.0-rc.0`.

The default value of this field is `false`. If `spec.deprecated` is set `true`, KubeDB operator will not create the database and other respective resources for this version.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create Petset by KubeDB operator to create expected MSSQLServer database.

### spec.exporter.image

`spec.exporter.image` is a required field that specifies the image which will be used to export Prometheus metrics.

### spec.tools.image

`spec.tools.image` is a required field that specifies the image which will be used to take backup and initialize database from snapshot.

### spec.podSecurityPolicies.databasePolicyName

`spec.podSecurityPolicies.databasePolicyName` is a required field that specifies the name of the pod security policy required to get the database server pod(s) running.

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --namespace kubedb --create-namespace \
  --set additionalPodSecurityPolicies[0]=custom-db-policy \
  --set additionalPodSecurityPolicies[1]=custom-snapshotter-policy \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

## Next Steps

- Learn about MSSQLServer crd [here](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- Deploy your first MSSQLServer database with KubeDB by following the guide [here](/docs/guides/mssqlserver/quickstart/quickstart.md).