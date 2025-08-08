---
title: CassandraOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: cas-opsrequest-concepts
    name: CassandraOpsRequest
    parent: cas-concepts-cassandra
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# CassandraOpsRequest

## What is CassandraOpsRequest

`CassandraOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [Cassandra](https://cassandra.apache.org/) administrative operations like database version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## CassandraOpsRequest CRD Specifications

Like any official Kubernetes resource, a `CassandraOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `CassandraOpsRequest` CRs for different administrative operations is given below:

Sample `CassandraOpsRequest` for updating database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: cassandra-prod
  updateVersion:
    targetVersion: 5.0.3
status:
  conditions:
    - lastTransitionTime: "2025-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

Sample `CassandraOpsRequest` Objects for Horizontal Scaling of different component of the database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-hscale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: cassandra-prod
  horizontalScaling:
    node: 4
status:
  conditions:
    - lastTransitionTime: "2025-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

Sample `CassandraOpsRequest` Objects for Vertical Scaling of different component of the database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: cassandra-prod
  verticalScaling:
    node:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "4Gi"
          cpu: "3"
status:
  conditions:
    - lastTransitionTime: "2025-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

Sample `CassandraOpsRequest` Objects for Reconfiguring different cassandra mode:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-reconfiugre
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: cassandra-prod
  configuration:
    applyConfig:
      cassandra.yaml: |
        authenticator: PasswordAuthenticator
status:
  conditions:
    - lastTransitionTime: "2025-07-25T18:22:38Z"
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
kind: CassandraOpsRequest
metadata:
  name: casops-reconfiugre
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: cassandra-prod
  configuration:
    configSecret:
      name: new-configsecret
status:
  conditions:
    - lastTransitionTime: "2025-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

Sample `CassandraOpsRequest` Objects for Volume Expansion of different database components:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-volume-exp
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: cassandra-prod
  volumeExpansion:
    mode: "Online"
    node: 2Gi
status:
  conditions:
    - lastTransitionTime: "2025-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

Sample `CassandraOpsRequest` Objects for Reconfiguring TLS of the database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: cassandra-prod
  tls:
    issuerRef:
      name: cas-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        emailAddresses:
          - abc@appscode.com
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: cassandra-dev
  tls:
    rotateCertificates: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: cassandra-prod
  tls:
    issuerRef:
      name: cas-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: cassandra-prod
  tls:
    remove: true
```

Here, we are going to describe the various sections of a `CassandraOpsRequest` crd.

A `CassandraOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [Cassandra](/docs/guides/cassandra/concepts/cassandra.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Cassandra](/docs/guides/cassandra/concepts/cassandra.md) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `CassandraOpsRequest`.

- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `CassandraOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `CassandraOpsRequest`. At first, you have to create a `CassandraOpsRequest` for updating. Once it is completed, then you can create another `CassandraOpsRequest` for scaling.

### spec.updateVersion

If you want to update you Cassandra version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [CassandraVersion](/docs/guides/cassandra/concepts/cassandraversion.md) CR that contains the Cassandra version information where you want to update.

> You can only update between Cassandra versions. KubeDB does not support downgrade for Cassandra.

### spec.horizontalScaling.node

If you want to scale-up or scale-down your Cassandra cluster or different components of it, you have to specify `spec.horizontalScaling.node` section. 

### spec.verticalScaling.node

`spec.verticalScaling.node` is a required field specifying the information of `Cassandra` resources like `cpu`, `memory` etc that will be scaled. 
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

If you want to expand the volume of your Cassandra cluster or different components of it, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.mode` specifies the volume expansion mode. Supported values are `Online` & `Offline`. The default is `Online`.
- `spec.volumeExpansion.node` indicates the desired size for the persistent volume for a Cassandra cluster.


All of them refer to [Quantity](https://v1-22.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#quantity-resource-core) types of Kubernetes.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    node: "2Gi"
```

This will expand the volume size of all the combined nodes to 2 GB.

### spec.configuration

If you want to reconfigure your Running Cassandra cluster or different components of it with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-field:

- `spec.configuration.configSecret` points to a secret in the same namespace of a Cassandra resource, which contains the new custom configurations. If there are any configSecret set before in the database, this secret will replace it.

- `applyConfig` is a map where key supports 3 values, namely `server.properties`, `broker.properties`, `controller.properties`. And value represents the corresponding configurations.

```yaml
  applyConfig:
    cassandra.yaml: |
      authenticator: PasswordAuthenticator
```

- `removeCustomConfig` is a boolean field. Specify this field to true if you want to remove all the custom configuration from the deployed cassandra cluster.

### spec.tls

If you want to reconfigure the TLS configuration of your Cassandra i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/cassandra/concepts/cassandra.md#spectls).
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this cassandra.
- `spec.tls.remove` specifies that we want to remove tls from this cassandra.

### spec.timeout
As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second).
If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.

### CassandraOpsRequest `Status`

`.status` describes the current state and progress of a `CassandraOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `CassandraOpsRequest`. It can have the following three values:

| Phase       | Meaning                                                                          |
|-------------|----------------------------------------------------------------------------------|
| Successful  | KubeDB has successfully performed the operation requested in the CassandraOpsRequest |
| Progressing | KubeDB has started the execution of the applied CassandraOpsRequest                  |
| Failed      | KubeDB has failed the operation requested in the CassandraOpsRequest                 |
| Denied      | KubeDB has denied the operation requested in the CassandraOpsRequest                 |
| Skipped     | KubeDB has skipped the operation requested in the CassandraOpsRequest                |

Important: Ops-manager Operator can skip an opsRequest, only if its execution has not been started yet & there is a newer opsRequest applied in the cluster. `spec.type` has to be same as the skipped one, in this case.

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `CassandraOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `CassandraOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. CassandraOpsRequest has the following types of conditions:

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
