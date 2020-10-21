---
title: PgBouncerVersion
menu:
  docs_{{ .version }}:
    identifier: pgbouncer-version
    name: PgBouncerVersion
    parent: catalog
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: concepts
---

# PgBouncerVersion

## What is PgBouncerVersion

`PgBouncerVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [PgBouncer](https://pgbouncer.github.io/) server deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `PgBouncerVersion` custom resource will be created automatically for every supported PgBouncer release versions. You have to specify the name of `PgBouncerVersion` crd in `spec.version` field of [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md) crd. Then, KubeDB will use the docker images specified in the `PgBouncerVersion` crd to create your expected PgBouncer instance.

Using a separate crd for specifying respective docker image names allow us to modify the images independent of KubeDB operator. This will also allow the users to use a custom PgBouncer image for their server. For more details about how to use custom image with PgBouncer in KubeDB, please visit [here](/docs/guides/pgbouncer/custom-versions/setup.md).

## PgBouncerVersion Specification

As with all other Kubernetes objects, a PgBouncerVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: PgBouncerVersion
metadata:
  name: "1.12.0"
  labels:
    app: kubedb
spec:
  deprecated: false
  version: "1.12.0"
  server:
    image: "${KUBEDB_CATALOG_REGISTRY}/pgbouncer:1.12.0"
  exporter:
    image: "${KUBEDB_CATALOG_REGISTRY}/pgbouncer_exporter:v0.1.1"
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `PgBouncerVersion` crd. You have to specify this name in `spec.version` field of [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md) crd.

We follow this convention for naming PgBouncerVersion crd:

- Name format: `{Original pgbouncer image version}-{modification tag}`

We plan to modify original PgBouncer docker images to support additional features. Re-tagging the image with v1, v2 etc. modification tag helps separating newer iterations from the older ones. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use PgBouncerVersion crd with highest modification tag to take advantage of the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of PgBouncer that has been used to build the docker image specified in `spec.server.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set `true`, KubeDB operator will not create the server and other respective resources for this version.

### spec.server.image

`spec.server.image` is a required field that specifies the docker image which will be used to create Statefulset by KubeDB operator to create expected PgBouncer server.

### spec.exporter.image

`spec.exporter.image` is a required field that specifies the image which will be used to export Prometheus metrics.

## Next Steps

- Learn about PgBouncer crd [here](/docs/guides/pgbouncer/concepts/catalog.md).
- Deploy your first PgBouncer server with KubeDB by following the guide [here](/docs/guides/pgbouncer/quickstart/quickstart.md).