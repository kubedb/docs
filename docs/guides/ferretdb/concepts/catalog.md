---
title: FerretDBVersion CRD
menu:
  docs_{{ .version }}:
    identifier: fr-catalog-concepts
    name: FerretDBVersion
    parent: fr-concepts-ferretdb
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# FerretDBVersion

## What is FerretDBVersion

`FerretDBVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [FerretDB](https://ferretdb.com/) server deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `FerretDBVersion` custom resource will be created automatically for every supported FerretDB release versions. You have to specify the name of `FerretDBVersion` crd in `spec.version` field of [FerretDB](/docs/guides/ferretdb/concepts/ferretdb.md) crd. Then, KubeDB will use the docker images specified in the `FerretDBVersion` crd to create your expected FerretDB instance.

Using a separate crd for specifying respective docker image names allow us to modify the images independent of KubeDB operator. This will also allow the users to use a custom FerretDB image for their server.

## FerretDBVersion Specification

As with all other Kubernetes objects, a FerretDBVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: FerretDBVersion
metadata:
  name: 1.23.0
spec:
  db:
    image: ghcr.io/appscode-images/ferretdb:1.23.0
  postgres:
    version: 17.4-documentdb    
  securityContext:
    runAsUser: 1000
  version: 1.23.0
  deprecated: false
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `FerretDBVersion` crd. You have to specify this name in `spec.version` field of [FerretDB](/docs/guides/ferretdb/concepts/ferretdb.md) crd.

We follow this convention for naming FerretDBVersion crd:

- Name format: `{Original ferretdb image version}-{modification tag}`

We plan to modify original FerretDB docker images to support additional features. Re-tagging the image with v1, v2 etc. modification tag help separating newer iterations from the older ones. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use FerretDBVersion crd with higher modification tag to take advantage of the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of FerretDB that has been used to build the docker image specified in `spec.server.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set `true`, KubeDB operator will not create the server and other respective resources for this version.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create PetSet by KubeDB operator to create expected FerretDB server.

### spec.postgres.version

`spec.postgres.version` is required field that specifies which KubeDB Postgres version will be used as backend of FerretDB.

### spec.securityContext

`spec.securityContext` holds pod-level security attributes and common container settings for FerretDB pod.

## Next Steps

- Learn about FerretDB crd [here](/docs/guides/ferretdb/concepts/catalog.md).
- Deploy your first FerretDB server with KubeDB by following the guide [here](/docs/guides/ferretdb/quickstart/quickstart.md).