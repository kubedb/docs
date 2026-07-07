---
title: WeaviateVersion CRD
menu:
  docs_{{ .version }}:
    identifier: weaviate-concepts-catalog
    name: WeaviateVersion
    parent: weaviate-concepts
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# WeaviateVersion

## What is WeaviateVersion

`WeaviateVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative way to specify the docker images and other version-specific metadata used to deploy a Weaviate database with the KubeDB operator.

When you install KubeDB, a `WeaviateVersion` CR is created for each supported Weaviate version. You must specify the name of a `WeaviateVersion` CR in `spec.version` of your `Weaviate` object.

## WeaviateVersion Spec

Below is an example `WeaviateVersion` object:

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: WeaviateVersion
metadata:
  name: 1.33.1
spec:
  db:
    image: ghcr.io/appscode-images/weaviate:1.33.1
  version: 1.33.1
```

### metadata.name

`metadata.name` is the name of the `WeaviateVersion` CR. You specify this name in `spec.version` of your `Weaviate` object.

### spec.version

`spec.version` is the actual Weaviate version released by the Weaviate project.

### spec.db.image

`spec.db.image` is the docker image used to create the Weaviate database pods.

### spec.deprecated

`spec.deprecated` (optional, boolean) indicates whether this `WeaviateVersion` is deprecated. KubeDB will not provision a database with a deprecated version.

## List available versions

```bash
kubectl get weaviateversions
```
NAME     VERSION   DB_IMAGE                                  DEPRECATED   AGE
1.33.1   1.33.1    ghcr.io/appscode-images/weaviate:1.33.1                34h

## Next Steps

- Learn about the [Weaviate CRD](/docs/guides/weaviate/concepts/weaviate.md).
- Deploy your first Weaviate database with the [Quickstart](/docs/guides/weaviate/quickstart/quickstart.md).
