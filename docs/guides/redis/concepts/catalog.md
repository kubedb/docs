---
title: RedisVersion CRD
menu:
  docs_{{ .version }}:
    identifier: rd-catalog-concepts
    name: RedisVersion
    parent: rd-concepts-redis
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RedisVersion

## What is RedisVersion

`RedisVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for `Redis/Valkey` database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `RedisVersion` custom resource will be created automatically for every supported Redis versions. You have to specify the name of `RedisVersion` crd in `spec.version` field of [Redis](/docs/guides/redis/concepts/redis.md) crd. Then, KubeDB will use the docker images specified in the `RedisVersion` crd to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator. This will also allow the users to use a custom image for the database.

## RedisVersion Specification

As with all other Kubernetes objects, a RedisVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: RedisVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2025.6.30
    helm.sh/chart: kubedb-catalog-v2025.6.30
  name: 7.2.4
spec:
  coordinator:
    image: ghcr.io/kubedb/redis-coordinator:v0.35.0
  db:
    image: ghcr.io/appscode-images/redis:7.2.4-bookworm
  distribution: Official
  exporter:
    image: ghcr.io/kubedb/redis_exporter:1.66.0
  initContainer:
    image: ghcr.io/kubedb/redis-init:0.12.0
  podSecurityPolicies:
    databasePolicyName: ""
  securityContext:
    runAsUser: 999
  stash:
    addon:
      backupTask:
        name: redis-backup-7.0.5
      restoreTask:
        name: redis-restore-7.0.5
  updateConstraints:
    allowlist:
      - '>= 7.2.4, < 7.4.2'
  version: 7.2.4
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `RedisVersion` crd. You have to specify this name in `spec.version` field of [Redis](/docs/guides/redis/concepts/redis.md) crd.

We follow this convention for naming RedisVersion crd:

- Name format: `{Original Redis image verion}-{modification tag}`

`Note`: If you want to use Valkey instead of Redis, you can add a prefix on the name `valkey`

We modify original Redis docker image to support Redis clustering and re-tag the image with v1, v2 etc. modification tag. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use RedisVersion crd with highest modification tag to enjoy the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of Redis server that has been used to build the docker image specified in `spec.db.image` field.

### spec.distribution

`spec.distribution` is a required field that specifies distribution of the database. Only two values are supported in this field `valkey` and `official`(Redis).

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will skip processing this CRD object and will add a event to the CRD object specifying that the DB version is deprecated.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create Petset by KubeDB operator to create expected Redis server.

### spec.initContainer.image

`spec.initContainer.image` is a required field that specifies the image for init container.

### spec.exporter.image

`spec.exporter.image` is a required field that specifies the image which will be used to export Prometheus metrics.

### spec.stash

This holds the Backup & Restore task definitions, where a `TaskRef` has a `Name` & `Params` section. Params specifies a list of parameters to pass to the task.
To learn more, visit [stash documentation](https://stash.run/)

### spec.updateConstraints
updateConstraints specifies the constraints that need to be considered during version update. Here `allowList` contains the versions those are allowed for updating from the current version.
An empty list of AllowList indicates all the versions are accepted except the denyList.
On the other hand, `DenyList` contains all the rejected versions for the update request. An empty list indicates no version is rejected.

### spec.podSecurityPolicies.databasePolicyName

`spec.podSecurityPolicies.databasePolicyName` is a required field that specifies the name of the pod security policy required to get the database server pod(s) running. To use a user-defined policy, the name of the policy has to be set in `spec.podSecurityPolicies` and in the list of allowed policy names in KubeDB operator like below:

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --namespace kubedb --create-namespace \
  --set additionalPodSecurityPolicies[0]=custom-db-policy \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

## Next Steps

- Learn about Redis crd [here](/docs/guides/redis/concepts/redis.md).
- Deploy your first Redis server with KubeDB by following the guide [here](/docs/guides/redis/quickstart/quickstart.md).
