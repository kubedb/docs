---
title: MilvusVersion CRD
menu:
  docs_{{ .version }}:
    identifier: milvus-catalog-concepts
    name: MilvusVersion
    parent: milvus-concepts-milvus
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MilvusVersion

## What is MilvusVersion

`MilvusVersion` is the catalog CRD that defines the Milvus engine image and related metadata for KubeDB-managed Milvus deployments.

KubeDB uses this CRD when resolving `Milvus.spec.version`.

## MilvusVersion Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MilvusVersion
metadata:
  name: 2.6.11
spec:
  db:
    image: ghcr.io/appscode-images/milvus:2.6.11
  etcdVersion: v3.5.21
  securityContext:
    runAsUser: 1000
  version: 2.6.11
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `MilvusVersion` CR. You have to specify this name in `spec.version` field of [Milvus](/docs/guides/milvus/concepts/milvus.md) CR.

We follow this convention for naming MilvusVersion CR:

- Name format: `{Original Milvus image version}-{modification tag}`

We use official Apache Milvus release tar files to build docker images for supporting Milvus versions and re-tag the image with v1, v2 etc. modification tag when there's any. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use MilvusVersion CR with the highest modification tag to enjoy the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of Milvus database that has been used to build the docker image specified in `spec.db.image` field.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create PetSet by KubeDB operator to create expected Milvus database.

### spec.etcdVersion

`spec.etcdVersion` specifies the compatible Etcd version required by this Milvus release.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will skip processing this CRD object and will add a event to the CRD object specifying that the DB version is deprecated.

## Next Steps

- Read the [Milvus CRD concept](/docs/guides/milvus/concepts/milvus.md).
- Run the [Milvus quickstart](/docs/guides/milvus/quickstart/quickstart.md).