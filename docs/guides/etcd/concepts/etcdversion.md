---
title: EtcdVersion CRD
menu:
  docs_{{ .version }}:
    identifier: etcd-catalog-concepts
    name: EtcdVersion
    parent: etcd-concepts-etcd
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# EtcdVersion

## What is EtcdVersion

`EtcdVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [etcd](https://etcd.io/) deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, an `EtcdVersion` custom resource will be created automatically for every supported etcd version. You have to specify the name of the `EtcdVersion` crd in the `spec.version` field of the [Etcd](/docs/guides/etcd/concepts/etcd.md) crd. Then, KubeDB will use the docker images specified in the `EtcdVersion` crd to create your expected database.

Using a separate crd for specifying the respective docker images and pod security policy names allows us to modify the images and policies independently of the KubeDB operator. This also allows users to use a custom image for the database.

> Unlike most other databases, etcd has no vendor distributions — there is a single upstream implementation — so `EtcdVersion` deliberately has **no** `spec.distribution` field.

`EtcdVersion` is a cluster-scoped resource, so it is not created in any namespace.

## EtcdVersion Specification

As with all other Kubernetes objects, an EtcdVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: EtcdVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2026-01-14T09:41:52Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
  name: 3.6.4
  resourceVersion: "12345"
  uid: 3c5a4714-4ce2-4b41-8ad9-4899c3127dcc
spec:
  db:
    image: ghcr.io/appscode-images/etcd:v3.6.4
  exporter:
    image: ghcr.io/appscode-images/etcd:v3.6.4
  securityContext:
    runAsUser: 1000
  updateConstraints:
    allowlist:
      - '>= 3.6.4, <= 3.7.0'
  version: 3.6.4
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `EtcdVersion` crd. You have to specify this name in the `spec.version` field of the [Etcd](/docs/guides/etcd/concepts/etcd.md) crd.

### spec.version

`spec.version` is a required field that specifies the original version of the etcd server that has been used to build the docker image specified in the `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here are supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, the KubeDB operator will skip processing this CRD object and will add an event to the CRD object specifying that the DB version is deprecated.

### spec.endOfLife

`spec.endOfLife` is an optional field that marks whether this etcd version has reached its end of life according to [endoflife.date](https://endoflife.date/). It is informational — it does not by itself block provisioning — but the Recommendation Engine takes it into account when proposing an `UpdateVersion` [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md).

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used by the KubeDB operator to create the PetSet running the expected etcd members.

### spec.initContainer.image

`spec.initContainer.image` is an optional field that specifies the image for the init container (`etcd-init`).

### spec.exporter.image

`spec.exporter.image` is a required field in the schema and specifies the image used for metrics related tooling.

> For etcd, **no exporter sidecar is actually deployed**. etcd serves its own Prometheus metrics endpoint natively, so KubeDB points the stats Service and the ServiceMonitor directly at etcd. The field is kept for structural parity with the other KubeDB catalog types.

### spec.securityContext

`spec.securityContext` holds the additional security settings applied to the etcd container.

- `spec.securityContext.runAsUser` is the default UID for the DB container. The upstream etcd image runs as uid `1000`.
- `spec.securityContext.runAsAnyNonRoot` will be `true` if the user is allowed to change the default DB container user to something other than the image's default user.

### spec.updateConstraints

`updateConstraints` specifies the constraints that need to be considered during a version update. Here, `allowlist` contains the versions that are allowed as the target of an update from the current version. An empty `allowlist` indicates that all versions are accepted except those in the `denylist`. Conversely, `denylist` contains all the rejected versions for the update request; an empty list indicates that no version is rejected.

For etcd, the shipped constraints follow the upstream upgrade rules: updates must stay within the same major version, must never go backwards, and must move at most one minor version at a time. For example, `3.5.21` allows `>= 3.5.21, <= 3.6.4`.

### spec.stash

This holds the backup & restore task definitions, where a `TaskRef` has a `Name` & `Params` section. `Params` specifies a list of parameters to pass to the task.

### spec.archiver

`spec.archiver` holds the KubeStash addon related specification used by the [EtcdArchiver](/docs/guides/etcd/concepts/etcdarchiver.md) driven snapshot backup.

### spec.podSecurityPolicies.databasePolicyName

`spec.podSecurityPolicies.databasePolicyName` specifies the name of the pod security policy required to get the database server pod(s) running. To use a user-defined policy, the name of the policy has to be set in `spec.podSecurityPolicies` and in the list of allowed policy names in the KubeDB operator like below:

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --namespace kubedb --create-namespace \
  --set additionalPodSecurityPolicies[0]=custom-db-policy \
  --set-file global.license=/path/to/the/license.txt \
  --set global.featureGates.Etcd=true \
  --wait --burst-limit=10000 --debug
```

## Next Steps

- Learn about the Etcd crd [here](/docs/guides/etcd/concepts/etcd.md).
- Deploy your first etcd cluster with KubeDB by following the guide [here](/docs/guides/etcd/quickstart/quickstart.md).
