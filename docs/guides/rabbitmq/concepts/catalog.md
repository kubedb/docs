---
title: RabbitMQVersion CRD
menu:
  docs_{{ .version }}:
    identifier: rm-catalog
    name: RabbitMQVersion
    parent: rm-concepts-guides
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# RabbitMQVersion

## What is RabbitMQVersion

`RabbitMQVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [RabbitMQ](https://www.rabbitmq.com/) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `RabbitMQVersion` custom resource will be created automatically for every supported RabbitMQ versions. You have to specify the name of `RabbitMQVersion` crd in `spec.version` field of [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md) crd. Then, KubeDB will use the docker images specified in the `RabbitMQVersion` crd to create your expected database.

Using a separate crd for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of KubeDB operator.This will also allow the users to use a custom image for the database.

## RabbitMQVersion Spec

As with all other Kubernetes objects, a RabbitMQVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Get `RabbitMQVersion` CR with a simple kubectl command.

```bash
$ kubectl get rmversion 3.13.2 -oyaml
```

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: RabbitMQVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2024-09-10T05:57:12Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2024.8.21
    helm.sh/chart: kubedb-catalog-v2024.8.21
  name: 3.13.2
  resourceVersion: "46385"
  uid: d853aaf9-e9b8-40b8-9663-a201a5a645c1
spec:
  db:
    image: ghcr.io/appscode-images/rabbitmq:3.13.2-management-alpine
  initContainer:
    image: ghcr.io/kubedb/rabbitmq-init:3.13.2
  securityContext:
    runAsUser: 999
  version: 3.13.2
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `RabbitMQVersion` crd. You have to specify this name in `spec.version` field of [RabbitMQ](/docs/guides/rabbitmq/concepts/rabbitmq.md) crd.

We follow this convention for naming RabbitMQVersion crd:

- Name format: `{Original RabbitMQ image version}-{modification tag}`

We modify original RabbitMQ docker image to support RabbitMQ clustering and re-tag the image with v1, v2 etc. modification tag. An image with higher modification tag will have more features than the images with lower modification tag. Hence, it is recommended to use RabbitMQVersion crd with the highest modification tag to enjoy the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of RabbitMQ database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will skip processing this CRD object and will add an event to the CRD object specifying that the DB version is deprecated.

### spec.db.image

`spec.db.image` is a required field that specifies the docker image which will be used to create PetSet by KubeDB operator to create expected RabbitMQ database.

### spec.initContainer.image
`spec.initContainer.image` is a required field that specifies the image for init container.


## Next Steps

- Learn about RabbitMQ crd [here](/docs/guides/rabbitmq/concepts/rabbitmq.md).
- Deploy your first RabbitMQ database with KubeDB by following the guide [here](/docs/guides/rabbitmq/concepts/rabbitmq.md).
