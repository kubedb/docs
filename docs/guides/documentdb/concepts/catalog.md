---
title: DocumentDBVersion CRD
menu:
  docs_{{ .version }}:
    identifier: documentdb-catalog-concepts
    name: DocumentDBVersion
    parent: documentdb-concepts-documentdb
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DocumentDBVersion

## What is DocumentDBVersion

`DocumentDB` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [DocumentDB](https://documentdb.apache.org) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `DocumentDBVersion` custom resource will be created automatically for every supported DocumentDB versions. You have to specify the name of `DocumentDBVersion` CR in `spec.version` field of [DocumentDB](/docs/guides/documentdb/concepts/documentdb.md) crd. Then, KubeDB will use the docker images specified in the `DocumentDB` CR to create your expected database.

Using a separate CRD for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator.This will also allow the users to use a custom image for the database.


## DocumentDBVersion Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: DocumentDBVersion
metadata:
  name: "pg17-0.109.0"
spec:
  version: "pg17-0.109.0"
  db:
    image: "kubedb/documentdb:pg17-0.109.0"
  deprecated: false
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `DocumentDB` CR. You have to specify this name in `spec.version` field of [DocumentDB](/docs/guides/documentdb/concepts/documentdb.md) CR.

We follow this convention for naming DocumentDBVersion CR:

- Name format: `{Original DocumentDB image version}-{modification tag}`

We use official Apache DocumentDB release tar files to build docker images for supporting DocumentDB versions and re-tag the image with v1, v2 etc. modification tag when there's any. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use DocumentDB CR with the highest modification tag to enjoy the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of DocumentDB database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will skip processing this CRD object and will add a event to the CRD object specifying that the DB version is deprecated.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create PetSet by KubeDB operator to create expected DocumentDB database.


## Next Steps

- Learn about DocumentDB CRD [here](/docs/guides/documentdb/concepts/documentdb.md).
- Deploy your first DocumentDB database with KubeDB by following the guide [here](/docs/guides/documentdb/quickstart/guides/index.md).