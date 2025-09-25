---
title: Ignite
menu:
  docs_{{ .version }}:
    identifier: ig-ignite-concepts
    name: Ignite
    parent: ig-concepts-ignite
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Ignite

## What is Ignite

`Ignite` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Ignite](https://ignite.apache.org/) in a Kubernetes native way. You only need to describe the desired database configuration in a Ignite object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## Ignite Spec

As with all other Kubernetes objects, a Ignite needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example of a Ignite object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ignite-quickstart
  namespace: demo
spec:
  replicas: 3
  version: 2.17.0
  configSecret: 
    name: ignite-configuration
  authSecret:
    kind: Secret
    name: ignite-quickstart-auth
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      containers:
        - name: "ignite"
          resources:
            requests:
              cpu: "500m"
            limits:
              cpu: "600m"
              memory: "1.5Gi"
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
  deletionPolicy: WipeOut
```

### spec.replicas

`spec.replicas` is an optional field that specifies the number of desired Instances/Replicas of Ignite server. If you do not specify .spec.replicas, then it defaults to 1.

KubeDB uses `PodDisruptionBudget` to ensure that majority of these replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions) so that quorum is maintained.

### spec.version

`spec.version` is a required field specifying the name of the [IgniteVersion](/docs/guides/ignite/concepts/ignite-version.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `IgniteVersion` crds,

- `2.17.0`

### spec.monitor

Ignite managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor Ignite with builtin Prometheus](/docs/guides/ignite/monitoring/using-builtin-prometheus.md)
- [Monitor Ignite with Prometheus operator](/docs/guides/ignite/monitoring/using-prometheus-operator.md)

### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for Ignite. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). So you can use any Kubernetes supported volume source such as `configMap`, `secret`, `azureDisk` etc. To learn more about how to use a custom configuration file see [here](/docs/guides/ignite/custom-configuration/using-config-file.md).

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the Petset created for Ignite server.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata
  - annotations (pod's annotation)
- controller
  - annotations (petset's annotation)
- spec:
  - containers
  - volumes
  - podPlacementPolicy
  - initContainers
  - containers
  - podPlacementPolicy
  - imagePullSecrets
  - nodeSelector
  - serviceAccountName
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext

Uses of some field of `spec.podTemplate` is described below,

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/master/api/v2/types.go#L26C1-L279C1).
Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.tolerations

The `spec.podTemplate.spec.tolerations` is an optional field. This can be used to specify the pod's tolerations.

#### spec.podTemplate.spec.volumes

The `spec.podTemplate.spec.volumes` is an optional field. This can be used to provide the list of volumes that can be mounted by containers belonging to the pod.

#### spec.podTemplate.spec.podPlacementPolicy

`spec.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the podPlacementPolicy. This will be used by our Petset controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes kubernetes affinity & podTopologySpreadContraints feature to do so.

#### spec.podTemplate.spec.containers

The `spec.podTemplate.spec.containers` can be used to provide the list containers and their configurations for to the database pod. some of the fields are described below,

##### spec.podTemplate.spec.containers[].name
The `spec.podTemplate.spec.containers[].name` field used to specify the name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.

##### spec.podTemplate.spec.containers[].args
`spec.podTemplate.spec.containers[].args` is an optional field. This can be used to provide additional arguments to database installation.

##### spec.podTemplate.spec.containers[].env

`.env` is an optional field that specifies the environment variables to pass to the Ignite containers.

Note that, KubeDB does not allow to update the environment variables. If you try to update environment variables, KubeDB operator will reject the request with following error,

```ini
Error from server (BadRequest): error when applying patch:
...
for: "./ig.yaml": admission webhook "ignite.validators.kubedb.com" denied the request: precondition failed for:
...
At least one of the following was changed:
	apiVersion
	kind
	name
	namespace
	spec.podTemplate.spec.nodeSelector
    spec.podTemplate.spec.env
```

#### spec.podTemplate.spec.containers[].resources

`spec.podTemplate.spec.containers[].resources` is an optional field. This can be used to request compute resources required by containers of the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

#### spec.podTemplate.spec.imagePullSecrets

`KubeDB` provides the flexibility of deploying Ignite server from a private Docker registry. To learn how to deploym Ignite from a private registry, please visit [here](/docs/guides/ignite/private-registry/using-private-registry.md).

#### spec.podTemplate.spec.nodeSelector

`spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

#### spec.podTemplate.spec.serviceAccountName

`serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

If this field is left empty, the KubeDB operator will create a service account name matching Ignite crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. Follow the guide [here](/docs/guides/ignite/custom-rbac/using-custom-rbac.md) to grant necessary permissions in this scenario.

#### spec.podTemplate.spec.resources

`spec.resources` is an optional field. This can be used to request compute resources required by the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

### spec.serviceTemplate

You can also provide a template for the services created by KubeDB operator for Ignite server through `spec.serviceTemplate`. This will allow you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplate`:

- `alias` represents the identifier of the service. It has the following possible value:
  - `primary` is used for the primary service identification.
  - `standby` is used for the secondary service identification.
  - `stats` is used for the exporter service identification.

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

See [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.16.3/api/v1/types.go#L163) to understand these fields in details.


### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Ignite` crd or which resources KubeDB should keep or delete when you delete `Ignite` crd. KubeDB provides following four termination policies:

- DoNotTerminate
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete Ignite crd for different termination policies,

| Behavior                   |  DoNotTerminate |  Delete  | WipeOut  |
| ---------------------------| :------------:  | :------: | :------: |
| 1. Block Delete operation  |    &#10003;     | &#10007; | &#10007; |
| 2. Delete PetSet           |    &#10007;     | &#10003; | &#10003; |
| 3. Delete Services         |    &#10007;     | &#10003; | &#10003; |
| 4. Delete Secrets          |    &#10007;     | &#10007; | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` termination policy by default.

## spec.helathChecker
It defines the attributes for the health checker.
- spec.healthChecker.periodSeconds specifies how often to perform the health check.
- spec.healthChecker.timeoutSeconds specifies the number of seconds after which the probe times out.
- spec.healthChecker.failureThreshold specifies minimum consecutive failures for the healthChecker to be considered failed.
- spec.healthChecker.disableWriteCheck specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this blog post.

## Next Steps

- Learn how to use KubeDB to run a Ignite server [here](/docs/guides/ignite/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
