---
title: DB2Version CRD
menu:
  docs_{{ .version }}:
    identifier: db2-catalog-concepts
    name: DB2Version
    parent: db2-concepts-db2
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DB2Version

## What is DB2Version

`DB2Version` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [DB2](https://www.ibm.com/products/db2) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `DB2Version` custom resource will be created automatically for every supported DB2 versions. You have to specify the name of `DB2Version` CR in `spec.version` field of [DB2](/docs/guides/db2/concepts/db2.md) crd. Then, KubeDB will use the docker images specified in the `DB2Version` CR to create your expected database.

Using a separate CRD for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator. This will also allow the users to use a custom image for the database.

## DB2Version Specification

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: DB2Version
metadata:
  name: "11.5.8.0"
spec:
  version: "11.5.8.0"
  coordinator:
    image: "ghcr.io/kubedb/db2-coordinator:v0.5.0-ubi"
  db:
    image: "kubedb/db2:11.5.8.0"
  deprecated: false
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `DB2Version` CR. You have to specify this name in `spec.version` field of [DB2](/docs/guides/db2/concepts/db2.md) CR.

We follow this convention for naming DB2Version CR:

- Name format: `{Original DB2 version}-{modification tag}`

We use official IBM DB2 release images to build docker images for supporting DB2 versions and re-tag the image with v1, v2 etc. modification tag when there's any. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use DB2 CR with the highest modification tag to enjoy the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of DB2 database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will skip processing this CRD object and will add a event to the CRD object specifying that the DB version is deprecated.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create PetSet by KubeDB operator to create expected DB2 database.

### spec.coordinator.image

`spec.coordinator.image` is a required field that specifies the docker image for the DB2 coordinator container. The coordinator image is used by KubeDB to perform health checks on the DB2 database instance.

The coordinator container runs alongside the main DB2 database container and is responsible for:

- Monitoring the health status of the DB2 database
- Performing periodic health checks to ensure the database is running properly
- Reporting the database status to KubeDB operator
- Handling graceful shutdowns and recovery procedures

This separate coordinator image allows KubeDB to reliably detect database failures and take appropriate actions such as pod restarts or alerts.



## Next Steps

- Read the [DB2 CRD concept](/docs/guides/db2/concepts/db2.md).
- Run the [DB2 quickstart](/docs/guides/db2/quickstart/quickstart.md).