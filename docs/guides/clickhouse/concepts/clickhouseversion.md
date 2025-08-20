---
title: ClickHouseVersion CRD
menu:
  docs_{{ .version }}:
    identifier: cas-catalog-concepts
    name: ClickHouseVersion
    parent: cas-concepts-clickhouse
    weight: 45
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ClickHouseVersion

## What is ClickHouseVersion

`ClickHouseVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [ClickHouse](https://clickhouse.com/) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `ClickHouseVersion` custom resource will be created automatically for every supported ClickHouse versions. You have to specify the name of `ClickHouseVersion` CR in `spec.version` field of [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md) crd. Then, KubeDB will use the docker images specified in the `ClickHouseVersion` CR to create your expected database.

Using a separate CRD for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator.This will also allow the users to use a custom image for the database.

## ClickHouseVersion Spec

As with all other Kubernetes objects, a ClickHouseVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: ClickHouseVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2025-08-18T11:01:50Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2025.7.31
    helm.sh/chart: kubedb-catalog-v2025.7.31
  name: 24.4.1
  resourceVersion: "1097"
  uid: 023551b4-dd59-42e5-92ab-7b18e7962637
spec:
  clickHouseKeeper:
    image: clickhouse/clickhouse-keeper:24.4.1
  db:
    image: clickhouse/clickhouse-server:24.4.1
  initContainer:
    image: ghcr.io/kubedb/clickhouse-init:24.4.1-v3
  securityContext:
    runAsUser: 101
  version: 24.4.1
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `ClickHouseVersion` CR. You have to specify this name in `spec.version` field of [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md) CR.

We follow this convention for naming ClickHouseVersion CR:

- Name format: `{Original ClickHouse image version}-{modification tag}`

We use the official ClickHouse release tar files to build Docker images for supported ClickHouse versions and re-tag the images with modification tags like v1, v2, etc. Each higher modification tag includes additional features or improvements compared to lower ones. Therefore, it is recommended to use the ClickHouseVersion CR with the highest modification tag to benefit from the latest features.
### spec.version

`spec.version` is a required field that specifies the original version of ClickHouse database that has been used to build the docker image specified in `spec.db.image` field.


### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create PetSet by KubeDB operator to create expected ClickHouse database.


### spec.initContainer.image

`spec.initContainer.image` is a required field that specifies the image which will be used to mount some scripts in database container.

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --namespace kubedb --create-namespace \
  --set global.featureGates.ClickHouse=true --set global.featureGates.ClickHouse=true \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

## Next Steps

- Learn about ClickHouse CRD [here](/docs/guides/clickhouse/concepts/clickhouse.md).
