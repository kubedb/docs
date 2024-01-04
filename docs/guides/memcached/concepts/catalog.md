---
title: MemcachedVersion CRD
menu:
  docs_{{ .version }}:
    identifier: mc-catalog-concepts
    name: MemcachedVersion
    parent: mc-concepts-memcached
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MemcachedVersion

## What is MemcachedVersion

`MemcachedVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [Memcached](https://memcached.org) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `MemcachedVersion` custom resource will be created automatically for every supported Memcached versions. You have to specify the name of `MemcachedVersion` crd in `spec.version` field of [Memcached](/docs/guides/memcached/concepts/memcached.md) crd. Then, KubeDB will use the docker images specified in the `MemcachedVersion` crd to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator. This will also allow the users to use a custom image for the database.

## MemcachedVersion Specification

As with all other Kubernetes objects, a MemcachedVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MemcachedVersion
metadata:
  name: "1.6.22"
  labels:
    app: kubedb
spec:
  version: "1.5.4"
  db:
    image: "${KUBEDB_DOCKER_REGISTRY}/memcached:1.6.22"
  exporter:
    image: "${KUBEDB_DOCKER_REGISTRY}/memcached-exporter:v0.4.1"
  podSecurityPolicies:
    databasePolicyName: "memcached-db"
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `MemcachedVersion` crd. You have to specify this name in `spec.version` field of [Memcached](/docs/guides/memcached/concepts/memcached.md) crd.

We follow this convention for naming MemcachedVersion crd:

- Name format: `{Original Memcached image version}-{modification tag}`

We modify original Memcached docker image to support additional features. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use MemcachedVersion crd with highest modification tag to take advantage of the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of Memcached server that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator. For example, we have modified `kubedb/memcached:1.5.4` docker image to support custom configuration and re-tagged as `kubedb/memcached:1.6.22`. Now, KubeDB `0.9.0-rc.0` supports providing custom configuration which required `kubedb/memcached:1.6.22` docker image. So, we have marked `kubedb/memcached:1.5.4` as deprecated for KubeDB `0.9.0-rc.0`.

The default value of this field is `false`. If `spec.deprecated` is set `true`, KubeDB operator will not create the database and other respective resources for this version.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create Statefulset by KubeDB operator to create expected Memcached server.

### spec.exporter.image

`spec.exporter.image` is a required field that specifies the image which will be used to export Prometheus metrics.

### spec.podSecurityPolicies.databasePolicyName

`spec.podSecurityPolicies.databasePolicyName` is a required field that specifies the name of the pod security policy required to get the database server pod(s) running. To use a user-defined policy, the name of the polict has to be set in `spec.podSecurityPolicies` and in the list of allowed policy names in KubeDB operator like below:

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --namespace kubedb --create-namespace \
  --set additionalPodSecurityPolicies[0]=custom-db-policy \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

## Next Steps

- Learn about Memcached crd [here](/docs/guides/memcached/concepts/memcached.md).
- Deploy your first Memcached server with KubeDB by following the guide [here](/docs/guides/memcached/quickstart/quickstart.md).
