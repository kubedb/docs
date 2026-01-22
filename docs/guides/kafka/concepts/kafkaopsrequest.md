---
title: KafkaOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: kf-opsrequest-concepts
    name: KafkaOpsRequest
    parent: kf-concepts-kafka
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# KafkaOpsRequest

## What is KafkaOpsRequest

`KafkaOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [Kafka](https://kafka.apache.org/) administrative operations like database version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## KafkaOpsRequest CRD Specifications

Like any official Kubernetes resource, a `KafkaOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `KafkaOpsRequest` CRs for different administrative operations is given below:

**Sample `KafkaOpsRequest` for updating database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: kafka-prod
  updateVersion:
    targetVersion: 3.9.0
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `KafkaOpsRequest` Objects for Horizontal Scaling of different component of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-hscale-combined
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: kafka-dev
  horizontalScaling:
    node: 3
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-hscale-down-topology
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: kafka-prod
  horizontalScaling:
    topology: 
      broker: 2
      controller: 2
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `KafkaOpsRequest` Objects for Vertical Scaling of different component of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-vscale-combined
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: kafka-dev
  verticalScaling:
    node:
      resources:
        requests:
          memory: "1.5Gi"
          cpu: "0.7"
        limits:
          memory: "2Gi"
          cpu: "1"
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-vscale-topology
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: kafka-prod
  verticalScaling:
    broker:
      resources:
        requests:
          memory: "1.5Gi"
          cpu: "0.7"
        limits:
          memory: "2Gi"
          cpu: "1"
    controller:
      resources:
        requests:
          memory: "1.5Gi"
          cpu: "0.7"
        limits:
          memory: "2Gi"
          cpu: "1"
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `KafkaOpsRequest` Objects for Reconfiguring different kafka mode:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-reconfiugre-combined
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: kafka-dev
  configuration:
    applyConfig:
      server.properties: |
        log.retention.hours=100
        default.replication.factor=2
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-reconfiugre-topology
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: kafka-prod
  configuration:
    applyConfig:
      broker.properties: |
        log.retention.hours=100
        default.replication.factor=2
      controller.properties: |
        metadata.log.dir=/var/log/kafka/metadata-custom
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-reconfiugre-combined
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: kafka-dev
  configuration:
    configSecret:
      name: new-configsecret-combined
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-reconfiugre-topology
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: kafka-prod
  configuration:
    configSecret:
      name: new-configsecret-topology
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-reconfiugre-combined
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: kafka-dev
  configuration:
    restart: true
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `KafkaOpsRequest` Objects for Volume Expansion of different database components:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-volume-exp-combined
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: kafka-dev
  volumeExpansion:
    mode: "Online"
    node: 2Gi
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-volume-exp-topology
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: kafka-prod
  volumeExpansion:
    mode: "Online"
    broker: 2Gi
    controller: 2Gi
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `KafkaOpsRequest` Objects for Reconfiguring TLS of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: kafka-prod
  tls:
    issuerRef:
      name: kf-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        emailAddresses:
          - abc@appscode.com
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: kafka-dev
  tls:
    rotateCertificates: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: kafka-prod
  tls:
    issuerRef:
      name: kf-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: kafka-prod
  tls:
    remove: true
```

Here, we are going to describe the various sections of a `KafkaOpsRequest` crd.

A `KafkaOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [Kafka](/docs/guides/kafka/concepts/kafka.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Kafka](/docs/guides/kafka/concepts/kafka.md) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `KafkaOpsRequest`.

- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `KafkaOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `KafkaOpsRequest`. At first, you have to create a `KafkaOpsRequest` for updating. Once it is completed, then you can create another `KafkaOpsRequest` for scaling.

### spec.updateVersion

If you want to update you Kafka version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [KafkaVersion](/docs/guides/kafka/concepts/kafkaversion.md) CR that contains the Kafka version information where you want to update.

> You can only update between Kafka versions. KubeDB does not support downgrade for Kafka.

### spec.horizontalScaling

If you want to scale-up or scale-down your Kafka cluster or different components of it, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.node` indicates the desired number of nodes for Kafka combined cluster after scaling. For example, if your cluster currently has 4 replica with combined node, and you want to add additional 2 nodes then you have to specify 6 in `spec.horizontalScaling.node` field. Similarly, if you want to remove one node from the cluster, you have to specify 3 in `spec.horizontalScaling.node` field.
- `spec.horizontalScaling.topology` indicates the configuration of topology nodes for Kafka topology cluster after scaling. This field consists of the following sub-field:
  - `spec.horizontalScaling.topoloy.broker` indicates the desired number of broker nodes for Kafka topology cluster after scaling.
  - `spec.horizontalScaling.topology.controller` indicates the desired number of controller nodes for Kafka topology cluster after scaling.

> If the reference kafka object is combined cluster, then you can only specify `spec.horizontalScaling.node` field. If the reference kafka object is topology cluster, then you can only specify `spec.horizontalScaling.topology` field. You can not specify both fields at the same time.

### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `Kafka` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.node` indicates the desired resources for combined Kafka cluster after scaling.
- `spec.verticalScaling.broker` indicates the desired resources for broker of Kafka topology cluster after scaling.
- `spec.verticalScaling.controller` indicates the desired resources for controller of Kafka topology cluster after scaling.

> If the reference kafka object is combined cluster, then you can only specify `spec.verticalScaling.node` field. If the reference kafka object is topology cluster, then you can only specify `spec.verticalScaling.broker` or `spec.verticalScaling.controller` or both fields. You can not specify `spec.verticalScaling.node` field with any other fields at the same time, but you can specify `spec.verticalScaling.broker` and `spec.verticalScaling.controller` fields at the same time.

All of them has the below structure:

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

If you want to expand the volume of your Kafka cluster or different components of it, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.mode` specifies the volume expansion mode. Supported values are `Online` & `Offline`. The default is `Online`.
- `spec.volumeExpansion.node` indicates the desired size for the persistent volume of a combined Kafka cluster.
- `spec.volumeExpansion.broker` indicates the desired size for the persistent volume for broker of a Kafka topology cluster.
- `spec.volumeExpansion.controller` indicates the desired size for the persistent volume for controller of a Kafka topology cluster.

> If the reference kafka object is combined cluster, then you can only specify `spec.volumeExpansion.node` field. If the reference kafka object is topology cluster, then you can only specify `spec.volumeExpansion.broker` or `spec.volumeExpansion.controller` or both fields. You can not specify `spec.volumeExpansion.node` field with any other fields at the same time, but you can specify `spec.volumeExpansion.broker` and `spec.volumeExpansion.controller` fields at the same time.

All of them refer to [Quantity](https://v1-22.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#quantity-resource-core) types of Kubernetes.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    node: "2Gi"
```

This will expand the volume size of all the combined nodes to 2 GB.

### spec.configuration

If you want to reconfigure your Running Kafka cluster or different components of it with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-field:

- `spec.configuration.configSecret` points to a secret in the same namespace of a Kafka resource, which contains the new custom configurations. If there are any configSecret set before in the database, this secret will replace it. The value of the field `spec.stringData` of the secret like below:
```yaml
server.properties: |
  default.replication.factor=3
  offsets.topic.replication.factor=3
  log.retention.hours=100
broker.properties: |
  default.replication.factor=3
  offsets.topic.replication.factor=3
  log.retention.hours=100
controller.properties: |
  default.replication.factor=3
  offsets.topic.replication.factor=3
  log.retention.hours=100
```
> If you want to reconfigure a combined Kafka cluster, then you can only specify `server.properties` field. If you want to reconfigure a topology Kafka cluster, then you can specify `broker.properties` or `controller.properties` or both fields. You can not specify `server.properties` field with any other fields at the same time, but you can specify `broker.properties` and `controller.properties` fields at the same time.

- `applyConfig` contains the new custom config as a string which will be merged with the previous configuration.

- `applyConfig` is a map where key supports 3 values, namely `server.properties`, `broker.properties`, `controller.properties`. And value represents the corresponding configurations.

```yaml
  applyConfig:
    server.properties: |
      default.replication.factor=3
      offsets.topic.replication.factor=3
      log.retention.hours=100
    broker.properties: |
      default.replication.factor=3
      offsets.topic.replication.factor=3
      log.retention.hours=100
    controller.properties: |
      metadata.log.dir=/var/log/kafka/metadata-custom
```
- `restart` significantly reduces unnecessary downtime.
  - `auto` (default): restart only if required (determined by ops manager operator)
  - `false`: no restart
  - `true`: always restart

    
- `removeCustomConfig` is a boolean field. Specify this field to true if you want to remove all the custom configuration from the deployed kafka cluster.

### spec.tls

If you want to reconfigure the TLS configuration of your Kafka i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/kafka/concepts/kafka.md#spectls).
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this kafka.
- `spec.tls.remove` specifies that we want to remove tls from this kafka.

### spec.timeout
As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second).
If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.

### KafkaOpsRequest `Status`

`.status` describes the current state and progress of a `KafkaOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `KafkaOpsRequest`. It can have the following three values:

| Phase       | Meaning                                                                          |
|-------------|----------------------------------------------------------------------------------|
| Successful  | KubeDB has successfully performed the operation requested in the KafkaOpsRequest |
| Progressing | KubeDB has started the execution of the applied KafkaOpsRequest                  |
| Failed      | KubeDB has failed the operation requested in the KafkaOpsRequest                 |
| Denied      | KubeDB has denied the operation requested in the KafkaOpsRequest                 |
| Skipped     | KubeDB has skipped the operation requested in the KafkaOpsRequest                |

Important: Ops-manager Operator can skip an opsRequest, only if its execution has not been started yet & there is a newer opsRequest applied in the cluster. `spec.type` has to be same as the skipped one, in this case.

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `KafkaOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `KafkaOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. KafkaOpsRequest has the following types of conditions:

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
| `UpdateMongosImage`           | Specifies such a state that the Mongos Image has been updated             |
| `UpdatePetSetResources`       | Specifies such a state that the Petset resources has been updated         |
| `UpdateShardResources`        | Specifies such a state that the Shard resources has been updated          |
| `UpdateReplicaSetResources`   | Specifies such a state that the Replicaset resources has been updated     |
| `UpdateConfigServerResources` | Specifies such a state that the ConfigServer resources has been updated   |
| `UpdateMongosResources`       | Specifies such a state that the Mongos resources has been updated         |
| `ScaleDownReplicaSet`         | Specifies such a state that the scale down operation of replicaset        |
| `ScaleUpReplicaSet`           | Specifies such a state that the scale up operation of replicaset          |
| `ScaleUpShardReplicas`        | Specifies such a state that the scale up operation of shard replicas      |
| `ScaleDownShardReplicas`      | Specifies such a state that the scale down operation of shard replicas    |
| `ScaleDownConfigServer`       | Specifies such a state that the scale down operation of config server     |
| `ScaleUpConfigServer`         | Specifies such a state that the scale up operation of config server       |
| `ScaleMongos`                 | Specifies such a state that the scale down operation of replicaset        |
| `VolumeExpansion`             | Specifies such a state that the volume expansion operaton of the database |
| `ReconfigureReplicaset`       | Specifies such a state that the reconfiguration of replicaset nodes       |
| `ReconfigureMongos`           | Specifies such a state that the reconfiguration of mongos nodes           |
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
