---
title: SchemaRegistryVersion CRD
menu:
  docs_{{ .version }}:
    identifier: kf-schemaregistryversion-concepts
    name: SchemaRegistryVersion
    parent: kf-concepts-kafka
    weight: 55
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SchemaRegistryVersion

## What is SchemaRegistryVersion

`SchemaRegistryVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for install SchemaRegistry and RestProxy with KubeDB in a Kubernetes native way.

When you install KubeDB, a `SchemaRegistryVersion` custom resource will be created automatically for every supported SchemaRegistry Version. You have to specify list of `SchemaRegistryVersion` CR names in `spec.version` field of SchemaRegistry or RestProxy CR. Then, KubeDB will use the docker images specified in the `SchemaRegistryVersion` cr to install your SchemaRegistry or RestProxy.

Using a separate CR for specifying respective docker images and policies independent of KubeDB operator. This will also allow the users to use a custom image for the SchemaRegistry or RestProxy.

## SchemaRegistryVersion Spec

As with all other Kubernetes objects, a KafkaVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: SchemaRegistryVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb-catalog
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2024-08-30T04:54:14Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb-catalog
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2024.8.21
    helm.sh/chart: kubedb-catalog-v2024.8.21
  name: 2.5.11.final
  resourceVersion: "133199"
  uid: deca9f55-6fef-4477-a66d-7e1fe77d9bbd
spec:
  distribution: Apicurio
  inMemory:
    image: apicurio/apicurio-registry-mem:2.5.11.Final
  registry:
    image: apicurio/apicurio-registry-kafkasql:2.5.11.Final
  securityContext:
    runAsUser: 1001
  version: 2.5.11
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `SchemaRegistryVersion` CR. You have to specify this name in `spec.version` field of SchemaRegistry and RestProxy CR.

### spec.version

`spec.version` is a required field that specifies the original version of SchemaRegistry that has been used to build the docker image specified in `spec.registry` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will skip processing this CRD object and will add a event to the CRD object specifying that the DB version is deprecated.

### spec.registry.image

`spec.registry.image` is a required field that specifies the docker image which will be used to install schema registry or restproxy by KubeDB operator.

### spec.inMemory.image

`spec.inMemory.image` is a optional field that specifies the docker image which will be used to install schema registry in memory by KubeDB operator.

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --namespace kubedb --create-namespace \
  --set additionalPodSecurityPolicies[0]=custom-db-policy \
  --set additionalPodSecurityPolicies[1]=custom-snapshotter-policy \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

## Next Steps

- Learn about Kafka CRD [here](/docs/guides/kafka/concepts/kafka.md).
- Learn about SchemaRegistry CRD [here](/docs/guides/kafka/concepts/schemaregistry.md).
- Deploy your first ConnectCluster with KubeDB by following the guide [here](/docs/guides/kafka/connectcluster/overview.md).
