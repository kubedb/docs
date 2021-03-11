---
title: Memcached
menu:
  docs_{{ .version }}:
    identifier: mc-memcached-concepts
    name: Memcached
    parent: mc-concepts-memcached
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Memcached

## What is Memcached

`Memcached` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Memcached](https://memcached.org/) in a Kubernetes native way. You only need to describe the desired database configuration in a Memcached object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Memcached Spec

As with all other Kubernetes objects, a Memcached needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example of a Memcached object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Memcached
metadata:
  name: mc1
  namespace: demo
spec:
  replicas: 3
  version: 1.5.22
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          app: kubedb
        interval: 10s
  configSecret:
    name: mc-custom-config
  podTemplate:
    metadata:
      annotations:
        passMe: ToDatabasePod
    controller:
      annotations:
        passMe: ToDeployment
    spec:
      serviceAccountName: my-service-account
      schedulerName: my-scheduler
      nodeSelector:
        disktype: ssd
      imagePullSecrets:
      - name: myregistrykey
      args:
      - "-u memcache"
      env:
      - name: TEST_ENV
        value: "value"
      resources:
        requests:
          memory: "64Mi"
          cpu: "250m"
        limits:
          memory: "128Mi"
          cpu: "500m"
  serviceTemplates:
  - alias: primary
    metadata:
      annotations:
        passMe: ToService
    spec:
      type: NodePort
      ports:
      - name:  http
        port:  9200
  terminationPolicy: Halt
```

### spec.replicas

`spec.replicas` is an optional field that specifies the number of desired Instances/Replicas of Memcached server. If you do not specify .spec.replicas, then it defaults to 1.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

### spec.version

`spec.version` is a required field specifying the name of the [MemcachedVersion](/docs/guides/memcached/concepts/catalog.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `MemcachedVersion` resources,

- `1.5.4`, `1.5.4-v1`, `1.5`, `1.5-v1`

### spec.monitor

Memcached managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor Memcached with builtin Prometheus](/docs/guides/memcached/monitoring/using-builtin-prometheus.md)
- [Monitor Memcached with Prometheus operator](/docs/guides/memcached/monitoring/using-prometheus-operator.md)

### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for Memcached. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). So you can use any kubernetes supported volume source such as `configMap`, `secret`, `azureDisk` etc. To learn more about how to use a custom configuration file see [here](/docs/guides/memcached/configuration/using-config-file.md).

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the Deployment created for Memcached server.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata
  - annotations (pod's annotation)
- controller
  - annotations (statefulset's annotation)
- spec:
  - args
  - env
  - resources
  - initContainers
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

Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.args

`spec.podTemplate.spec.args` is an optional field. This can be used to provide additional arguments to database installation.

#### spec.podTemplate.spec.env

`spec.env` is an optional field that specifies the environment variables to pass to the Memcached docker image.

Note that, KubeDB does not allow to update the environment variables. If you try to update environment variables, KubeDB operator will reject the request with following error,

```ini
Error from server (BadRequest): error when applying patch:
...
for: "./mc.yaml": admission webhook "memcached.validators.kubedb.com" denied the request: precondition failed for:
...
At least one of the following was changed:
	apiVersion
	kind
	name
	namespace
	spec.podTemplate.spec.nodeSelector
    spec.podTemplate.spec.env
```

#### spec.podTemplate.spec.imagePullSecrets

`KubeDB` provides the flexibility of deploying Memcached server from a private Docker registry. To learn how to deploym Memcached from a private registry, please visit [here](/docs/guides/memcached/private-registry/using-private-registry.md).

#### spec.podTemplate.spec.nodeSelector

`spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.serviceAccountName

  `serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

  If this field is left empty, the KubeDB operator will create a service account name matching Memcached crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

  If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

  If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. Follow the guide [here](/docs/guides/memcached/custom-rbac/using-custom-rbac.md) to grant necessary permissions in this scenario.

#### spec.podTemplate.spec.resources

`spec.resources` is an optional field. This can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplate

You can also provide a template for the services created by KubeDB operator for Memcached server through `spec.serviceTemplate`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplate`:

- metadata:
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

See [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.16.3/api/v1/types.go#L163) to understand these fields in detail.

### spec.terminationPolicy

`terminationPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Memcached` crd or which resources KubeDB should keep or delete when you delete `Memcached` crd. KubeDB provides following four termination policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `terminationPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete Memcached crd for different termination policies,

| Behavior                            | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ----------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Create Dormant Database          |    &#10007;    | &#10003; | &#10007; | &#10007; |
| 3. Delete StatefulSet               |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |

If you don't specify `spec.terminationPolicy` KubeDB uses `Halt` termination policy by default.

## Next Steps

- Learn how to use KubeDB to run a Memcached server [here](/docs/guides/memcached/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
