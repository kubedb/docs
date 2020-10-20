---
title: PerconaXtraDBVersion
menu:
  docs_{{ .version }}:
    identifier: percona-xtradb-version
    name: PerconaXtraDBVersion
    parent: catalog
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

# PerconaXtraDBVersion

## What is PerconaXtraDBVersion

`PerconaXtraDBVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for PerconaXtraDB deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `PerconaXtraDBVersion` custom resource will be created automatically for every supported PerconaXtraDB versions. You have to specify the name of `PerconaXtraDBVersion` object in `.spec.version` field of [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/percona-xtradb.md) object. Then, KubeDB will use the docker images specified in the `PerconaXtraDBVersion` crd to create your expected PerconaXtraDBVersion instance.

Using a separate object for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator. This will also allow the users to use a custom image for the database.

## PerconaXtraDBVersion Specification

As with all other Kubernetes objects, a PerconaXtraDBVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: PerconaXtraDBVersion
metadata:
  name: "5.7"
  labels:
    app: kubedb
spec:
  version: "5.7"
  deprecated: false
  db:
    image: "${KUBEDB_CATALOG_REGISTRY}/percona-xtradb-cluster:5.7"
  exporter:
    image: "${KUBEDB_CATALOG_REGISTRY}/mysqld-exporter:v0.11.0"
  podSecurityPolicies:
    databasePolicyName: "percona-xtradb-db"
```

### .metadata.name

`.metadata.name` is a required field that specifies the name of the `PerconaXtraDBVersion` object. You have to specify this name in `.spec.version` field of [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/percona-xtradb.md) object.

We follow this convention for naming PerconaXtraDBVersion object:

- Name format: `{Original PerconaXtraDB image version}-{modification tag}`

We modify original PerconaXtraDB docker image to support additional features. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use PerconaXtraDBVersion object with highest modification tag to take advantage of the latest features.

### .spec.version

`.spec.version` is a required field that specifies the original version of PerconaXtraDB database that has been used to build the docker image specified in `.spec.db.image` field.

### .spec.deprecated

`.spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `.spec.deprecated` is set `true`, KubeDB operator will not create the database and other respective resources for this version.

### .spec.db.image

`.spec.db.image` is a required field that specifies the docker image which will be used to create Statefulset by KubeDB operator to create expected PerconaXtraDB database.

### .spec.exporter.image

`.spec.exporter.image` is a required field that specifies the image which will be used to export Prometheus metrics.

### .spec.podSecurityPolicies.databasePolicyName

`.spec.podSecurityPolicies.databasePolicyName` is a required field that specifies the name of the pod security policy required to get the database server Pod(s) running.

## Next Steps

- Learn about PerconaXtraDB CRD [here](/docs/guides/percona-xtradb/concepts/percona-xtradb.md).
- Deploy your first PerconaXtraDB database with KubeDB by following the guide [here](/docs/guides/percona-xtradb/quickstart/quickstart.md).
