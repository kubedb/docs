---
title: QdrantVersion CRD
menu:
  docs_{{ .version }}:
    identifier: qdrant-catalog-concepts
    name: QdrantVersion
    parent: qdrant-concepts-qdrant
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# QdrantVersion

## What is QdrantVersion

`QdrantVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the Docker images to be used for [Qdrant](https://qdrant.tech/) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `QdrantVersion` custom resource will be created automatically for every supported Qdrant version. You have to specify the name of the `QdrantVersion` CRD in `spec.version` field of the [Qdrant](/docs/guides/qdrant/concepts/qdrant.md) CRD. Then, KubeDB will use the Docker images specified in the `QdrantVersion` CRD to create your expected database.

Using a separate CRD for specifying respective Docker images allows us to modify images independent of the KubeDB operator. This also allows users to use a custom image for the database.

## QdrantVersion Specification

As with all other Kubernetes objects, a `QdrantVersion` needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: QdrantVersion
metadata:
  name: "1.17.0"
spec:
  version: "1.17.0"
  db:
    image: "qdrant/qdrant:v1.17.0"
  deprecated: false
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `QdrantVersion` CRD. You have to specify this name in `spec.version` field of the [Qdrant](/docs/guides/qdrant/concepts/qdrant.md) CRD.

The naming convention for `QdrantVersion` CRD follows the pattern `{Original Qdrant version}`.

### spec.version

`spec.version` is a required field that specifies the original version of the Qdrant database that has been used to build the Docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the Docker images specified here are supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will not create the database and other respective resources for this version.

### spec.db.image

`spec.db.image` is a required field that specifies the Docker image which will be used to create the Petset by KubeDB operator to create the expected Qdrant database.

```bash
$ kubectl get qdrantversions
NAME      VERSION   DB_IMAGE                    DEPRECATED   AGE
1.7.4     1.7.4     qdrant/qdrant:v1.7.4                     3d
1.10.0    1.10.0    qdrant/qdrant:v1.10.0                    3d
1.14.0    1.14.0    qdrant/qdrant:v1.14.0                    3d
1.17.0    1.17.0    qdrant/qdrant:v1.17.0                    3d
```

## Next Steps

- Learn about the [Qdrant CRD](/docs/guides/qdrant/concepts/qdrant.md).
- Deploy your first Qdrant database with KubeDB by following the guide [here](/docs/guides/qdrant/quickstart/quickstart.md).