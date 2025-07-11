---
title: CassandraVersion CRD
menu:
  docs_{{ .version }}:
    identifier: guides-cassandra-concepts-cassandraversion
    name: CassandraVersion
    parent: guides-cassandra-concepts
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# CassandraVersion

## What is CassandraVersion

`CassandraVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [Cassandra](https://cassandra.apache.org) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `CassandraVersion` custom resource will be created automatically for every supported Cassandra versions. You have to specify the name of `CassandraVersion` CR in `spec.version` field of [Cassandra](/docs/guides/cassandra/concepts/cassandra.md) crd. Then, KubeDB will use the docker images specified in the `CassandraVersion` CR to create your expected database.

Using a separate CRD for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator.This will also allow the users to use a custom image for the database.

## CassandraVersion Spec

As with all other Kubernetes objects, a CassandraVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: CassandraVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2025-10-16T13:10:10Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2024.9.30
    helm.sh/chart: kubedb-catalog-v2025.6.30
  name: 5.0.3
  resourceVersion: "42125"
  uid: e30e23aa-febc-4029-8be7-993afaff1fc6
spec:
  db:
    image: ghcr.io/appscode-images/cassandra-management:5.0.3
  exporter:
    image: ghcr.io/appscode-images/cassandra-exporter:2.3.8
  initContainer:
    image: ghcr.io/kubedb/cassandra-init:5.0.0-v2
  medusa:
    image: ghcr.io/appscode-images/cassandra-medusa:0.24.0
    init:
      image: ghcr.io/appscode-images/cassandra-medusa:0.24.0
  securityContext:
    runAsUser: 999
  version: 5.0.3
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `CassandraVersion` CR. You have to specify this name in `spec.version` field of [Cassandra](/docs/guides/cassandra/concepts/cassandra.md) CR.

We follow this convention for naming CassandraVersion CR:

- Name format: `{Original Cassandra image version}-{modification tag}`

We use official Apache Cassandra release tar files to build docker images for supporting Cassandra versions and re-tag the image with v1, v2 etc. modification tag when there's any. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use CassandraVersion CR with the highest modification tag to enjoy the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of Cassandra database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will skip processing this CRD object and will add a event to the CRD object specifying that the DB version is deprecated.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create PetSet by KubeDB operator to create expected Cassandra database.

### spec.initContainer.image

`spec.initContainer.image` is a required field that specifies the image which will be used to remove `lost+found` directory and mount an `EmptyDir` data volume.

### spec.podSecurityPolicies.databasePolicyName

`spec.podSecurityPolicies.databasePolicyName` is a required field that specifies the name of the pod security policy required to get the database server pod(s) running.

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --namespace kubedb --create-namespace \
  --set global.featureGates.Cassandra=true --set global.featureGates.Cassandra=true \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

## Next Steps

- Learn about Cassandra CRD [here](/docs/guides/cassandra/concepts/cassandra.md).
- Deploy your first Cassandra database with KubeDB by following the guide [here](/docs/guides/cassandra/quickstart/guide/index.md).
