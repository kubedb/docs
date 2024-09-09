---
title: SchemaRegistry CRD
menu:
  docs_{{ .version }}:
    identifier: kf-schemaregistry-concepts
    name: SchemaRegistry
    parent: kf-concepts-kafka
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SchemaRegistry

## What is SchemaRegistry

`SchemaRegistry` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [SchemaRegistry](https://www.apicur.io/registry/) in a Kubernetes native way. You only need to describe the desired configuration in a `SchemaRegistry` object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## SchemaRegistry Spec

As with all other Kubernetes objects, a SchemaRegistry needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example SchemaRegistry object.

```yaml
apiVersion: kafka.kubedb.com/v1alpha1
kind: SchemaRegistry
metadata:
  name: schemaregistry
  namespace: demo
spec:
  version: 2.5.11.final
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  replicas: 3
  kafkaRef:
    name: kafka
    namespace: demo
  podTemplate:
    metadata:
      annotations:
        passMe: ToDatabasePod
      labels:
        thisLabel: willGoToPod
    controller:
      annotations:
        passMe: ToPetSet
      labels:
        thisLabel: willGoToSts
  deletionPolicy: WipeOut
```

### spec.version

`spec.version` is a required field specifying the name of the [SchemaRegistryVersion](/docs/guides/kafka/concepts/schemaregistryversion.md) CR where the docker images are specified. Currently, when you install KubeDB, it creates the following `SchemaRegistryVersion` resources,
- `2.5.11.final`
- `3.15.0`

### spec.replicas

`spec.replicas` the number of instances in SchemaRegistry.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

### spec.kafkaRef

`spec.kafkaRef` is a optional field that specifies the name and namespace of the appbinding for `Kafka` object that the `SchemaRegistry` object is associated with.
```yaml
kafkaRef:
  name: <kafka-object-appbinding-name>
  namespace: <kafka-object-appbinding-namespace>
```

### spec.podTemplate

KubeDB allows providing a template for pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for SchemaRegistry.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
    - annotations (pod's annotation)
    - labels (pod's labels)
- controller:
    - annotations (petset's annotation)
    - labels (petset's labels)
- spec:
    - volumes
    - initContainers
    - containers
    - imagePullSecrets
    - nodeSelector
    - affinity
    - serviceAccountName
    - schedulerName
    - tolerations
    - priorityClassName
    - priority
    - securityContext
    - livenessProbe
    - readinessProbe
    - lifecycle

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/39bf8b2/api/v2/types.go#L44-L279). Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.resources

`spec.podTemplate.spec.resources` is an optional field. This can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplates

You can also provide template for the services created by KubeDB operator for Kafka cluster through `spec.serviceTemplates`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:
- `alias` represents the identifier of the service. It has the following possible value:
    - `stats` is used for the exporter service identification.
- metadata:
    - labels
    - annotations
- spec:
    - type
    - ports
    - clusterIP
    - externalIPs
    - loadBalancerIP
    - loadBalancerSourceRanges
    - externalTrafficPolicy
    - healthCheckNodePort
    - sessionAffinityConfig

See [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.21.1/api/v1/types.go#L237) to understand these fields in detail.

### spec.deletionPolicy

`spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `SchemaRegistry` crd or which resources KubeDB should keep or delete when you delete `SchemaRegistry` crd. KubeDB provides following four deletion policies:

- Delete
- DoNotTerminate
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).

## Next Steps

- Learn how to use KubeDB to run a Apache Kafka Connect cluster [here](/docs/guides/kafka/README.md).
- Monitor your SchemaRegistry with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Detail concepts of [KafkaConnectorVersion object](/docs/guides/kafka/concepts/kafkaconnectorversion.md).
- Learn to use KubeDB managed Kafka objects using [CLIs](/docs/guides/kafka/cli/cli.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
