---
title: HazelcastVersion CRD
menu:
  docs_{{ .version }}:
    identifier: hz-catalog-concepts
    name: HazelcastVersion
    parent: hz-concepts-hazelcast
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HazelcastVersion

## What is HazelcastVersion

`HazelcastVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [Hazelcast](https://hazelcast.com/) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `HazelcastVersion` custom resource will be created automatically for every supported Hazelcast versions. You have to specify the name of `HazelcastVersion` crd in `spec.version` field of [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md) crd. Then, KubeDB will use the docker images specified in the `HazelcastVersion` crd to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator. This will also allow the users to use a custom image for the database.

## HazelcastVersion Specification

As with all other Kubernetes objects, a HazelcastVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: HazelcastVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2025-06-11T06:11:37Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2025.5.30
    helm.sh/chart: kubedb-catalog-v2025.5.30
  name: 5.5.2
  resourceVersion: "1171464"
  uid: d4918f95-4b81-4824-a73f-3dc1b3fbadc1
spec:
  db:
    image: hazelcast/hazelcast-enterprise:5.5.2
  initContainer:
    image: ghcr.io/kubedb/hazelcast-init:5.5.2
  securityContext:
    runAsUser: 65534
  updateConstraints:
    allowlist:
      - '>= 5.5.2, <= 6.0.0'
  version: 5.5.2

```

### metadata.name

`metadata.name` is a required field that specifies the name of the `HazelcastVersion` crd. You have to specify this name in `spec.version` field of [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md) crd.


### spec.version

`spec.version` is a required field that specifies the original version of Hazelcast server that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will skip processing this CRD object and will add a event to the CRD object specifying that the DB version is deprecated.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create StatefulSet by KubeDB operator to create expected Hazelcast server.

### spec.initContainer.image

`spec.initContainer.image` is a required field that specifies the image for init container.

### spec.securityContext

DB specific security context which will be added in statefulset.

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --namespace kubedb --create-namespace \
  --set additionalPodSecurityPolicies[0]=custom-db-policy \
  --set-file global.license=/path/to/the/license.txt \
  --set gloabal.featureGates.Hazelcast=true \
  --wait --burst-limit=10000 --debug
```

## Next Steps

- Learn about Hazelcast crd [here](/docs/guides/hazelcast/concepts/hazelcast.md).
- Deploy your first Hazelcast server with KubeDB by following the guide [here](/docs/guides/hazelcast/quickstart/overview/index.md).
