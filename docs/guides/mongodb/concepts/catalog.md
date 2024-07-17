---
title: MongoDBVersion CRD
menu:
  docs_{{ .version }}:
    identifier: mg-catalog-concepts
    name: MongoDBVersion
    parent: mg-concepts-mongodb
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDBVersion

## What is MongoDBVersion

`MongoDBVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [MongoDB](https://www.mongodb.com/) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `MongoDBVersion` custom resource will be created automatically for every supported MongoDB versions. You have to specify the name of `MongoDBVersion` crd in `spec.version` field of [MongoDB](/docs/guides/mongodb/concepts/mongodb.md) crd. Then, KubeDB will use the docker images specified in the `MongoDBVersion` crd to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator.This will also allow the users to use a custom image for the database.

## MongoDBVersion Spec

As with all other Kubernetes objects, a MongoDBVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MongoDBVersion
metadata:
  name: "4.4.26"
  labels:
    app: kubedb
spec:
  db:
    image: mongo:4.4.26
  distribution: Official
  exporter:
    image: kubedb/mongodb_exporter:v0.32.0
  initContainer:
    image: kubedb/mongodb-init:4.2-v7
  podSecurityPolicies:
    databasePolicyName: mongodb-db
  replicationModeDetector:
    image: kubedb/replication-mode-detector:v0.16.0
  stash:
    addon:
      backupTask:
        name: mongodb-backup-4.4.6
      restoreTask:
        name: mongodb-restore-4.4.6
  updateConstraints:
    allowlist:
      - '>= 4.4.0, < 5.0.0'
  version: 4.4.26
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `MongoDBVersion` crd. You have to specify this name in `spec.version` field of [MongoDB](/docs/guides/mongodb/concepts/mongodb.md) crd.

We follow this convention for naming MongoDBVersion crd:

- Name format: `{Original MongoDB image verion}-{modification tag}`

We modify original MongoDB docker image to support MongoDB clustering and re-tag the image with v1, v2 etc. modification tag. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use MongoDBVersion crd with highest modification tag to enjoy the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of MongoDB database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator. For example, we have modified `kubedb/mongo:3.6` docker image to support MongoDB clustering and re-tagged as `kubedb/mongo:3.6-v1`. So, we have marked `kubedb/mongo:3.6` as deprecated for KubeDB `0.9.0-rc.0`.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will skip processing this CRD object and will add a event to the CRD object specifying that the DB version is deprecated.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create Petset by KubeDB operator to create expected MongoDB database.

### spec.initContainer.image
`spec.initContainer.image` is a required field that specifies the image for init container.


### spec.exporter.image

`spec.exporter.image` is a required field that specifies the image which will be used to export Prometheus metrics.

### spec.stash
This holds the Backup & Restore task definitions, where a `TaskRef` has a `Name` & `Params` section. Params specifies a list of parameters to pass to the task.

### spec.updateConstraints
updateConstraints specifies the constraints that need to be considered during version update. Here `allowList` contains the versions those are allowed for updating from the current version.
An empty list of AllowList indicates all the versions are accepted except the denyList.
On the other hand, `DenyList` contains all the rejected versions for the update request. An empty list indicates no version is rejected.

### spec.podSecurityPolicies.databasePolicyName

`spec.podSecurityPolicies.databasePolicyName` is a required field that specifies the name of the pod security policy required to get the database server pod(s) running.

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --namespace kubedb --create-namespace \
  --set additionalPodSecurityPolicies[0]=custom-db-policy \
  --set additionalPodSecurityPolicies[1]=custom-snapshotter-policy \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

## Next Steps

- Learn about MongoDB crd [here](/docs/guides/mongodb/concepts/mongodb.md).
- Deploy your first MongoDB database with KubeDB by following the guide [here](/docs/guides/mongodb/quickstart/quickstart.md).
