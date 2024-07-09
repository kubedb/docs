---
title: ZooKeeperVersion CRD
menu:
  docs_{{ .version }}:
    identifier: zk-catalog-concepts
    name: ZooKeeperVersion
    parent: zk-concepts-zookeeper
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ZooKeeperVersion

## What is ZooKeeperVersion

`ZooKeeperVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [ZooKeeper](https://zookeeper.apache.org/) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `ZooKeeperVersion` custom resource will be created automatically for every supported ZooKeeper versions. You have to specify the name of `ZooKeeperVersion` crd in `spec.version` field of [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md) crd. Then, KubeDB will use the docker images specified in the `ZooKeeperVersion` crd to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator. This will also allow the users to use a custom image for the database.

## ZooKeeperVersion Specification

As with all other Kubernetes objects, a ZooKeeperVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: ZooKeeperVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2024-05-02T09:41:52Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2024.4.27
    helm.sh/chart: kubedb-catalog-v2024.4.27
  name: 3.9.1
  resourceVersion: "1455"
  uid: 3c5a4714-4ce2-4b41-8ad9-4899c3127dcc
spec:
  db:
    image: ghcr.io/appscode-images/zookeeper:3.9.1
  initContainer:
    image: ghcr.io/kubedb/zookeeper-init:3.7-v1
  securityContext:
    runAsUser: 1000
  version: 3.9.1
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `ZooKeeperVersion` crd. You have to specify this name in `spec.version` field of [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md) crd.


### spec.version

`spec.version` is a required field that specifies the original version of ZooKeeper server that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will skip processing this CRD object and will add a event to the CRD object specifying that the DB version is deprecated.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create Petset by KubeDB operator to create expected ZooKeeper server.

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
  --set gloabal.featureGates.ZooKeeper=true \
  --wait --burst-limit=10000 --debug
```

## Next Steps

- Learn about ZooKeeper crd [here](/docs/guides/zookeeper/concepts/zookeeper.md).
- Deploy your first ZooKeeper server with KubeDB by following the guide [here](/docs/guides/zookeeper/quickstart/quickstart.md).
