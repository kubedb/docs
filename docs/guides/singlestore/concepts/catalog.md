---
title: SingleStoreVersion CRD
menu:
  docs_{{ .version }}:
    identifier: sdb-catalog-concepts
    name: SingleStoreVersion
    parent: sdb-concepts-singlestore
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SingleStoreVersion

## What is SingleStoreVersion

`SingleStoreVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [SingleStore](https://www.singlestore.com/) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `SingleStoreVersion` custom resource will be created automatically for every supported SingleStore versions. You have to specify the name of `SingleStoreVersion` crd in `spec.version` field of [SingleStore](/docs/guides/singlestore/concepts/singlestore.md) crd. Then, KubeDB will use the docker images specified in the `SingleStoreVersion` crd to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator.This will also allow the users to use a custom image for the database.

## SingleStoreVersion Spec

As with all other Kubernetes objects, a SingleStoreVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: SinglestoreVersion
metadata:
  name: 8.5.7
spec:
  coordinator:
    image: ghcr.io/kubedb/singlestore-coordinator:v0.2.0-rc.2
  db:
    image: ghcr.io/appscode-images/singlestore-node:alma-8.5.7-bf633c1a54
  initContainer:
    image: ghcr.io/kubedb/singlestore-init:8.5-v2
  securityContext:
    runAsGroup: 998
    runAsUser: 999
  standalone:
    image: singlestore/cluster-in-a-box:alma-8.5.7-bf633c1a54-4.0.17-1.17.8
  version: 8.5.7
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `SingleStoreVersion` crd. You have to specify this name in `spec.version` field of [SingleStore](/docs/guides/singlestore/concepts/singlestore.md) crd.

We follow this convention for naming SingleStoreVersion crd:

- Name format: `{Original SingleStore image verion}-{modification tag}`

We modify original SingleStore docker image to support SingleStore clustering and re-tag the image with v1, v2 etc. modification tag. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use SingleStoreVersion crd with highest modification tag to enjoy the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of SingleStore database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will skip processing this CRD object and will add a event to the CRD object specifying that the DB version is deprecated.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create Petset by KubeDB operator to create expected SingleStore database.

### spec.coordinator.image

`spec.coordinator.image` is a required field that specifies the docker image which will be used to create Petset by KubeDB operator to create expected SingleStore database.

### spec.initContainer.image

`spec.initContainer.image` is a required field that specifies the image for init container.

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

- Learn about SingleStore crd [here](/docs/guides/singlestore/concepts/singlestore.md).
- Deploy your first SingleStore database with KubeDB by following the guide [here](/docs/guides/singlestore/quickstart/quickstart.md).