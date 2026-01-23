---
title: ClickHouseOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: ch-opsrequest-concepts
    name: ClickHouseOpsRequest
    parent: ch-concepts-clickhouse
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# ClickHouseOpsRequest

## What is ClickHouseOpsRequest

`ClickHouseOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [ClickHouse](https://clickhouse.com/) administrative operations like database version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## ClickHouseOpsRequest CRD Specifications

Like any official Kubernetes resource, a `ClickHouseOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `ClickHouseOpsRequest` CRs for different administrative operations is given below:

Sample `ClickHouseOpsRequest` for updating database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: update-version
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: clickhouse-prod
  type: UpdateVersion
  updateVersion:
    targetVersion: 25.7.1
status:
  conditions:
    - lastTransitionTime: "2025-08-21T07:54:21Z"
      message: Successfully completed update clickhouse version
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

Sample `ClickHouseOpsRequest` for Horizontal Scaling of Database Cluster:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-scale-horizontal-up
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: clickhouse-prod
  horizontalScaling:
    replicas: 3
  type: HorizontalScaling
status:
  conditions:
    - lastTransitionTime: "2025-08-21T08:04:41Z"
      message: Successfully completed horizontally scale ClickHouse cluster
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

Sample `ClickHouseOpsRequest` for Vertical Scaling of Database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: ch-scale-vertical-cluster
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: clickhouse-prod
  type: VerticalScaling
  verticalScaling:
    node:
      resources:
        limits:
          cpu: "2"
          memory: 2Gi
        requests:
          cpu: "2"
          memory: 2Gi
status:
  conditions:
    - lastTransitionTime: "2025-08-21T08:15:43Z"
      message: Successfully completed the vertical scaling for ClickHouse
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

Sample `ClickHouseOpsRequest` Objects for Reconfiguring ClickHouse database with config:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-reconfiugre
  namespace: demo
spec:
  apply: IfReady
  configuration:
    applyConfig:
      config.yaml: |
        profiles:
          default:
            max_query_size: 180000
  databaseRef:
    name: clickhouse-prod
  type: Reconfigure
status:
  conditions:
    - lastTransitionTime: "2025-08-21T08:27:41Z"
      message: Successfully completed reconfigure ClickHouse
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```
Sample `ClickHouseOpsRequest` Objects for Reconfiguring ClickHouse database with secret:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-reconfiugre
  namespace: demo
spec:
  apply: IfReady
  configuration:
    configSecret:
      name: ch-custom-config
    restart: "true"
  databaseRef:
    name: clickhouse-prod
  type: Reconfigure
status:
  conditions:
    - lastTransitionTime: "2025-08-21T10:00:04Z"
      message: Successfully completed reconfigure ClickHouse
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

Sample `ClickHouseOpsRequest` Objects for Volume Expansion of ClickHouse:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: ch-offline-volume-expansion
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: clickhouse-prod
  type: VolumeExpansion
  volumeExpansion:
    mode: Offline
    node: 2Gi
status:
  conditions:
    - lastTransitionTime: "2025-08-21T10:36:53Z"
      message: Successfully completed volumeExpansion for ClickHouse
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful

```

Sample `ClickHouseOpsRequest` Objects for Reconfiguring TLS of the database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: clickhouse-prod
  tls:
    sslVerificationMode: "strict"
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: ch-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: clickhouse-prod
  tls:
    rotateCertificates: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: clickhouse-prod
  tls:
    issuerRef:
      name: ch-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: clickhouse-prod
  tls:
    remove: true
```

Here, we are going to describe the various sections of a `ClickHouseOpsRequest` crd.

A `ClickHouseOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [ClickHouse](/docs/guides/clickhouse/concepts/clickhouse.md) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `ClickHouseOpsRequest`.

- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `ClickHouseOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `ClickHouseOpsRequest`. At first, you have to create a `ClickHouseOpsRequest` for updating. Once it is completed, then you can create another `ClickHouseOpsRequest` for scaling.

### spec.updateVersion

If you want to update you ClickHouse version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [ClickHouseVersion](/docs/guides/clickhouse/concepts/clickhouseversion.md) CR that contains the ClickHouse version information where you want to update.

> You can only update between ClickHouse versions. KubeDB does not support downgrade for ClickHouse.

### spec.horizontalScaling.node

If you want to scale-up or scale-down your ClickHouse cluster or different components of it, you have to specify `spec.horizontalScaling.node` section.

### spec.verticalScaling.node

`spec.verticalScaling.node` is a required field specifying the information of `ClickHouse` resources like `cpu`, `memory` etc that will be scaled.
this has the below structure:

```yaml
requests:
  memory: "200Mi"
  cpu: "0.1"
limits:
  memory: "300Mi"
  cpu: "0.2"
```

Here, when you specify the resource request, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for the container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. You can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).

### spec.volumeExpansion

> To use the volume expansion feature the storage class must support volume expansion

If you want to expand the volume of your ClickHouse cluster or different components of it, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.mode` specifies the volume expansion mode. Supported values are `Online` & `Offline`. The default is `Online`.
- `spec.volumeExpansion.node` indicates the desired size for the persistent volume for a ClickHouse cluster.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    node: "2Gi"
```

This will expand the volume size of all the combined nodes to 2 GB.

### spec.configuration

If you want to reconfigure your Running ClickHouse cluster or different components of it with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-field:

- `spec.configuration.configSecret` points to a secret in the same namespace of a ClickHouse resource, which contains the new custom configurations. If there are any configSecret set before in the database, this secret will replace it.

- `applyConfig` is a map where the key represents the target config file (e.g., config.yaml) and the value contains the corresponding configuration content.

```yaml
  applyConfig:
    config.yaml: |
      profiles:
        default:
          max_query_size: 180000
```

- `removeCustomConfig` is a boolean field. Specify this field to true if you want to remove all the custom configuration from the deployed clickhouse cluster.

- `restart` significantly reduces unnecessary downtime.
  - `auto` (default): restart only if required (determined by ops manager operator)
  - `false`: no restart
  - `true`: always restart
### spec.tls

If you want to reconfigure the TLS configuration of your ClickHouse i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/clickhouse/concepts/clickhouse.md#spectls).
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this clickhouse.
- `spec.tls.remove` specifies that we want to remove tls from this clickhouse.

### spec.timeout
As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second).
If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.

### ClickHouseOpsRequest `Status`

`.status` describes the current state and progress of a `ClickHouseOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `ClickHouseOpsRequest`. It can have the following three values:

| Phase       | Meaning                                                                          |
|-------------|----------------------------------------------------------------------------------|
| Successful  | KubeDB has successfully performed the operation requested in the ClickHouseOpsRequest |
| Progressing | KubeDB has started the execution of the applied ClickHouseOpsRequest                  |
| Failed      | KubeDB has failed the operation requested in the ClickHouseOpsRequest                 |
| Denied      | KubeDB has denied the operation requested in the ClickHouseOpsRequest                 |
| Skipped     | KubeDB has skipped the operation requested in the ClickHouseOpsRequest                |

Important: Ops-manager Operator can skip an opsRequest, only if its execution has not been started yet & there is a newer opsRequest applied in the cluster. `spec.type` has to be same as the skipped one, in this case.

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `ClickHouseOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `ClickHouseOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. ClickHouseOpsRequest has the following types of conditions:

| Type                          | Meaning                                                                   |
|-------------------------------|---------------------------------------------------------------------------|
| `Progressing`                 | Specifies that the operation is now in the progressing state              |
| `Successful`                  | Specifies such a state that the operation on the database was successful. |
| `HaltDatabase`                | Specifies such a state that the database is halted by the operator        |
| `ResumeDatabase`              | Specifies such a state that the database is resumed by the operator       |
| `Failed`                      | Specifies such a state that the operation on the database failed.         |
| `StartingBalancer`            | Specifies such a state that the balancer has successfully started         |
| `StoppingBalancer`            | Specifies such a state that the balancer has successfully stopped         |
| `UpdateShardImage`            | Specifies such a state that the Shard Images has been updated             |
| `UpdateReplicaSetImage`       | Specifies such a state that the Replicaset Image has been updated         |
| `UpdateConfigServerImage`     | Specifies such a state that the ConfigServer Image has been updated       |
| `UpdatePetSetResources`       | Specifies such a state that the Petset resources has been updated         |
| `UpdateShardResources`        | Specifies such a state that the Shard resources has been updated          |
| `UpdateReplicaSetResources`   | Specifies such a state that the Replicaset resources has been updated     |
| `UpdateConfigServerResources` | Specifies such a state that the ConfigServer resources has been updated   |
| `ScaleDownReplicaSet`         | Specifies such a state that the scale down operation of replicaset        |
| `ScaleUpReplicaSet`           | Specifies such a state that the scale up operation of replicaset          |
| `ScaleUpShardReplicas`        | Specifies such a state that the scale up operation of shard replicas      |
| `ScaleDownShardReplicas`      | Specifies such a state that the scale down operation of shard replicas    |
| `ScaleDownConfigServer`       | Specifies such a state that the scale down operation of config server     |
| `ScaleUpConfigServer`         | Specifies such a state that the scale up operation of config server       |
| `VolumeExpansion`             | Specifies such a state that the volume expansion operaton of the database |
| `ReconfigureReplicaset`       | Specifies such a state that the reconfiguration of replicaset nodes       |
| `ReconfigureShard`            | Specifies such a state that the reconfiguration of shard nodes            |
| `ReconfigureConfigServer`     | Specifies such a state that the reconfiguration of config server nodes    |

- The `status` field is a string, with possible values `True`, `False`, and `Unknown`.
    - `status` will be `True` if the current transition succeeded.
    - `status` will be `False` if the current transition failed.
    - `status` will be `Unknown` if the current transition was denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
