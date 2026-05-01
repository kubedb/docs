---
title: WeaviateVersion CRD
menu:
  docs_{{ .version }}:
    identifier: weaviate-catalog-concepts
    name: WeaviateVersion
    parent: weaviate-concepts-weaviate
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# WeaviateVersion

## What is WeaviateVersion

`WeaviateVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the Docker images to be used for [Weaviate](https://weaviate.io/) vector databases deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `WeaviateVersion` custom resource will be created automatically for every supported Weaviate version. You have to specify the name of the `WeaviateVersion` CRD in `spec.version` field of the [Weaviate](/docs/guides/weaviate/concepts/weaviate.md) CRD. Then, KubeDB will use the Docker images specified in the `WeaviateVersion` CRD to create your expected database.

Using a separate CRD for specifying respective Docker images allows us to modify images independent of the KubeDB operator. This also allows users to use a custom image for the database.

## WeaviateVersion Specification

As with all other Kubernetes objects, a `WeaviateVersion` needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: WeaviateVersion
metadata:
  name: "1.33.1"
spec:
  version: "1.33.1"
  db:
    image: "kubedb/weaviate:1.33.1"
  deprecated: false
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `WeaviateVersion` CRD. You have to specify this name in `spec.version` field of the [Weaviate](/docs/guides/weaviate/concepts/weaviate.md) CRD.

The naming convention for `WeaviateVersion` CRD follows the pattern `{Original Weaviate version}`.

### spec.version

`spec.version` is a required field that specifies the original version of the Weaviate database that has been used to build the Docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the Docker images specified here are supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will not create the database and other respective resources for this version.

### spec.db.image

`spec.db.image` is a required field that specifies the Docker image which will be used to create the StatefulSet by KubeDB operator to create the expected Weaviate database.

```bash
$ kubectl get weaviateversions
NAME      VERSION   DB_IMAGE                        DEPRECATED   AGE
1.25.0    1.25.0    kubedb/weaviate:1.25.0                       3d
1.28.0    1.28.0    kubedb/weaviate:1.28.0                       3d
1.30.0    1.30.0    kubedb/weaviate:1.30.0                       3d
1.33.1    1.33.1    kubedb/weaviate:1.33.1                       3d
```

## Next Steps

- Learn about the [Weaviate CRD](/docs/guides/weaviate/concepts/weaviate.md).
- Deploy your first Weaviate database with KubeDB by following the guide [here](/docs/guides/weaviate/quickstart/quickstart.md).