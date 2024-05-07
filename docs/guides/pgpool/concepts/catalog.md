---
title: PgpoolVersion CRD
menu:
  docs_{{ .version }}:
    identifier: pp-catalog-concepts
    name: PgpoolVersion
    parent: pp-concepts-pgpool
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PgpoolVersion

## What is PgpoolVersion

`PgpoolVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [Pgpool](https://pgpool.net/) server deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `PgpoolVersion` custom resource will be created automatically for every supported Pgpool release versions. You have to specify the name of `PgpoolVersion` crd in `spec.version` field of [Pgpool](/docs/guides/pgpool/concepts/pgpool.md) crd. Then, KubeDB will use the docker images specified in the `PgpoolVersion` crd to create your expected Pgpool instance.

Using a separate crd for specifying respective docker image names allow us to modify the images independent of KubeDB operator. This will also allow the users to use a custom Pgpool image for their server. For more details about how to use custom image with Pgpool in KubeDB, please visit [here](/docs/guides/pgpool/custom-versions/setup.md).

## PgpoolVersion Specification

As with all other Kubernetes objects, a PgpoolVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: PgpoolVersion
metadata:
  name: 4.5.0
spec:
  exporter:
    image: ghcr.io/appscode-images/pgpool2_exporter:v1.2.2
  pgpool:
    image: ghcr.io/appscode-images/pgpool2:4.5.0
  version: 4.5.0
  deprecated: false
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `PgpoolVersion` crd. You have to specify this name in `spec.version` field of [Pgpool](/docs/guides/pgpool/concepts/pgpool.md) crd.

We follow this convention for naming PgpoolVersion crd:

- Name format: `{Original pgpool image version}-{modification tag}`

We plan to modify original Pgpool docker images to support additional features. Re-tagging the image with v1, v2 etc. modification tag help separating newer iterations from the older ones. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use PgpoolVersion crd with higher modification tag to take advantage of the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of Pgpool that has been used to build the docker image specified in `spec.server.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set `true`, KubeDB operator will not create the server and other respective resources for this version.

### spec.pgpool.image

`spec.pgpool.image` is a required field that specifies the docker image which will be used to create PetSet by KubeDB operator to create expected Pgpool server.

### spec.exporter.image

`spec.exporter.image` is a required field that specifies the image which will be used to export Prometheus metrics.

## Next Steps

- Learn about Pgpool crd [here](/docs/guides/pgpool/concepts/catalog.md).
- Deploy your first Pgpool server with KubeDB by following the guide [here](/docs/guides/pgpool/quickstart/quickstart.md).