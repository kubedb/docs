---
title: MongoDBOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: mg-opsrequest-concepts
    name: MongoDBOpsRequest
    parent: mg-concepts-mongodb
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# MongoDBOpsRequest

## What is MongoDBOpsRequest

`MongoDBOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [MongoDB](https://www.mongodb.com/) administrative operations like database version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## MongoDBOpsRequest CRD Specifications

Like any official Kubernetes resource, a `MongoDBOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `MongoDBOpsRequest` CRs for different administrative operations is given below:

**Sample `MongoDBOpsRequest` for updating database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: mg-standalone
  updateVersion:
    targetVersion: 4.0.5-v3
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `MongoDBOpsRequest` Objects for Horizontal Scaling of different component of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-hscale-configserver
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mg-sharding
  horizontalScaling:
    shard:
      shards: 3
      replicas: 3
    configServer:
      replicas: 3
    mongos:
      replicas: 2
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
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
kind: MongoDBOpsRequest
metadata:
  name: mops-hscale-down-replicaset
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mg-replicaset
  horizontalScaling:
    replicas: 3
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `MongoDBOpsRequest` Objects for Vertical Scaling of different component of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-vscale-configserver
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mg-sharding
  verticalScaling:
    configServer:
      requests:
        memory: "150Mi"
        cpu: "0.1"
      limits:
        memory: "250Mi"
        cpu: "0.2"
    mongos:
      requests:
        memory: "150Mi"
        cpu: "0.1"
      limits:
        memory: "250Mi"
        cpu: "0.2"
    shard:
      requests:
        memory: "150Mi"
        cpu: "0.1"
      limits:
        memory: "250Mi"
        cpu: "0.2"
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
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
kind: MongoDBOpsRequest
metadata:
  name: mops-vscale-standalone
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mg-standalone
  verticalScaling:
    standalone:
      requests:
        memory: "150Mi"
        cpu: "0.1"
      limits:
        memory: "250Mi"
        cpu: "0.2"
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
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
kind: MongoDBOpsRequest
metadata:
  name: mops-vscale-replicaset
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: mg-replicaset
  verticalScaling:
    replicaSet:
      requests:
        memory: "150Mi"
        cpu: "0.1"
      limits:
        memory: "250Mi"
        cpu: "0.2"
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `MongoDBOpsRequest` Objects for Reconfiguring different database components:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-reconfiugre-data-replicaset
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-replicaset
  configuration:
    replicaSet:
      inlineConfig: |
          net:
            maxIncomingConnections: 30000
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
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
kind: MongoDBOpsRequest
metadata:
  name: mops-reconfiugre-data-shard
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-sharding
  configuration:
    shard:
      inlineConfig: |
          net:
            maxIncomingConnections: 30000
    configServer:
      inlineConfig: |
        net:
          maxIncomingConnections: 30000
    mongos:
      inlineConfig: |
        net:
          maxIncomingConnections: 30000
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
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
kind: MongoDBOpsRequest
metadata:
  name: mops-reconfiugre-data-standalone
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-standalone
  configuration:
    standalone:
      inlineConfig: |
        net:
          maxIncomingConnections: 30000
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
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
kind: MongoDBOpsRequest
metadata:
  name: mops-reconfiugre-replicaset
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-replicaset
  configuration:
    replicaSet:
      configSecret:
        name: new-custom-config
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
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
kind: MongoDBOpsRequest
metadata:
  name: mops-reconfiugre-shard
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-sharding
  configuration:
    shard:
      configSecret:
        name: new-custom-config
    configServer:
      configSecret:
        name: new-custom-config
    mongos:
      configSecret:
        name: new-custom-config
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
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
kind: MongoDBOpsRequest
metadata:
  name: mops-reconfiugre-standalone
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-standalone
  configuration:
    standalone:
      configSecret:
        name: new-custom-config
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `MongoDBOpsRequest` Objects for Volume Expansion of different database components:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-volume-exp-replicaset
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: mg-replicaset
  volumeExpansion:
    replicaSet: 2Gi
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
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
kind: MongoDBOpsRequest
metadata:
  name: mops-volume-exp-shard
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: mg-sharding
  volumeExpansion:
    shard: 2Gi
    configServer: 2Gi
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
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
kind: MongoDBOpsRequest
metadata:
  name: mops-volume-exp-standalone
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: mg-standalone
  volumeExpansion:
    standalone: 2Gi
status:
  conditions:
    - lastTransitionTime: "2020-08-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `MongoDBOpsRequest` Objects for Reconfiguring TLS of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mg-rs
  tls:
    issuerRef:
      name: mg-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        emailAddresses:
          - abc@appscode.com
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mg-rs
  tls:
    rotateCertificates: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mg-rs
  tls:
    issuerRef:
      name: mg-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mg-rs
  tls:
    remove: true
```

Here, we are going to describe the various sections of a `MongoDBOpsRequest` crd.

A `MongoDBOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [MongoDB](/docs/guides/mongodb/concepts/mongodb.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [MongoDB](/docs/guides/mongodb/concepts/mongodb.md) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `MongoDBOpsRequest`.

- `Upgrade` / `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `MongoDBOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `MongoDBOpsRequest`. At first, you have to create a `MongoDBOpsRequest` for updating. Once it is completed, then you can create another `MongoDBOpsRequest` for scaling. 

> Note: There is an exception to the above statement. It is possible to specify both `spec.configuration` & `spec.verticalScaling` in a OpsRequest of type `VerticalScaling`.

### spec.updateVersion

If you want to update you MongoDB version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [MongoDBVersion](/docs/guides/mongodb/concepts/catalog.md) CR that contains the MongoDB version information where you want to update.

Have a look on the [`updateConstraints`](/docs/guides/mongodb/concepts/catalog.md#specupdateconstraints) of the mongodbVersion spec to know which versions are supported for updating from the current version.
```yaml
kubectl get mgversion <current-version> -o=jsonpath='{.spec.updateConstraints}' | jq
```

> You can only update between MongoDB versions. KubeDB does not support downgrade for MongoDB.

### spec.horizontalScaling

If you want to scale-up or scale-down your MongoDB cluster or different components of it, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.replicas` indicates the desired number of nodes for MongoDB replicaset cluster after scaling. For example, if your cluster currently has 4 replicaset nodes, and you want to add additional 2 nodes then you have to specify 6 in `spec.horizontalScaling.replicas` field. Similarly, if you want to remove one node from the cluster, you have to specify 3 in `spec.horizontalScaling.replicas` field.
- `spec.horizontalScaling.configServer.replicas` indicates the desired number of ConfigServer nodes for Sharded MongoDB cluster after scaling.
- `spec.horizontalScaling.mongos.replicas` indicates the desired number of Mongos nodes for Sharded MongoDB cluster after scaling.
- `spec.horizontalScaling.shard` indicates the configuration of shard nodes for Sharded MongoDB cluster after scaling. This field consists of the following sub-field:
  - `spec.horizontalScaling.shard.replicas` indicates the number of replicas each shard will have after scaling.
  - `spec.horizontalScaling.shard.shards` indicates the number of shards after scaling

### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `MongoDB` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.standalone` indicates the desired resources for standalone MongoDB database after scaling.
- `spec.verticalScaling.replicaSet` indicates the desired resources for replicaSet of MongoDB database after scaling.
- `spec.verticalScaling.mongos` indicates the desired resources for Mongos nodes of Sharded MongoDB database after scaling.
- `spec.verticalScaling.configServer` indicates the desired resources for ConfigServer nodes of Sharded MongoDB database after scaling.
- `spec.verticalScaling.shard` indicates the desired resources for Shard nodes of Sharded MongoDB database after scaling.
- `spec.verticalScaling.exporter` indicates the desired resources for the `exporter` container.
- `spec.verticalScaling.arbiter` indicates the desired resources for arbiter node of MongoDB database after scaling.
- `spec.verticalScaling.coordinator` indicates the desired resources for the coordinator container.

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

If you want to expand the volume of your MongoDB cluster or different components of it, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.mode` specifies the volume expansion mode. Supported values are `Online` & `Offline`. The default is `Online`.
- `spec.volumeExpansion.standalone` indicates the desired size for the persistent volume of a standalone MongoDB database.
- `spec.volumeExpansion.replicaSet` indicates the desired size for the persistent volume of replicaSets of a MongoDB database.
- `spec.volumeExpansion.configServer` indicates the desired size for the persistent volume of the config server of a sharded MongoDB database.
- `spec.volumeExpansion.shard` indicates the desired size for the persistent volume of shards of a sharded MongoDB database.

All of them refer to [Quantity](https://v1-22.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#quantity-resource-core) types of Kubernetes.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    shard: "2Gi"
```

This will expand the volume size of all the shard nodes to 2 GB.

### spec.configuration

If you want to reconfigure your Running MongoDB cluster or different components of it with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-field:

- `spec.configuration.standalone` indicates the desired new custom configuration for a standalone MongoDB database.
- `spec.configuration.replicaSet` indicates the desired new custom configuration for replicaSet of a MongoDB database.
- `spec.configuration.configServer` indicates the desired new custom configuration for config servers of a sharded MongoDB database.
- `spec.configuration.mongos` indicates the desired new custom configuration for the mongos nodes of a sharded MongoDB database.
- `spec.configuration.shard` indicates the desired new custom configuration for the shard nodes of a sharded MongoDB database.
- `spec.verticalScaling.arbiter` indicates the desired new custom configuration for arbiter node of MongoDB database after scaling.

All of them has the following sub-fields:

- `configSecret` points to a secret in the same namespace of a MongoDB resource, which contains the new custom configurations. If there are any configSecret set before in the database, this secret will replace it.
- `inlineConfig` contains the new custom config as a string which will be merged with the previous configuration. 
> Note: You can use `inlineConfig` only for `mongod.conf` configurations. This field is deprecated & will be removed in some future KubeDB release.

- `applyConfig` is the replacement of `inlineConfig`. It is a map where key supports 3 values, namely `mongod.conf`, `replicaset.json`, `configuration.js`. And value represents the corresponding configurations.
For your information, replicaset.json is used to modify replica set configurations, which we see in the output of `rs.config()`. And `configurarion.js` is used to apply a js script to configure mongodb at runtime.
KubeDB provisioner operator applies these two directly while reconciling.

```yaml
  applyConfig:
    configuration.js: |
      print("hello world!!!!")
    replicaset.json: |
      {
        "settings" : {
          "electionTimeoutMillis" : 4000
        }
      }
    mongod.conf: |
      net:
        maxIncomingConnections: 30000   
```

- `removeCustomConfig` is a boolean field. Specify this field to true if you want to remove all the custom configuration from the deployed mongodb server.

### spec.tls

If you want to reconfigure the TLS configuration of your database i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/mongodb/concepts/mongodb.md#spectls).
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.
- `spec.tls.remove` specifies that we want to remove tls from this database.

### spec.readinessCriteria

`spec.readinessCriteria` is the criteria for checking readiness of a MongoDB pod after restarting it. It has two fields. 
- `spec.readinessCriteria.oplogMaxLagSeconds` defines the maximum allowed lagging time between the primary & secondary.
- `spec.readinessCriteria.objectsCountDiffPercentage` denotes the maximum allowed object-count-difference between the primary & secondary.

```yaml
...
spec:
  readinessCriteria:
    oplogMaxLagSeconds: 20
    objectsCountDiffPercentage: 10
...
```
Exceeding these thresholds results in opsRequest failure. One thing to note that, readinessCriteria field will make impact only if pod restarting is associated with the opsRequest type.

### spec.timeout
As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second). 
If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.


### MongoDBOpsRequest `Status`

`.status` describes the current state and progress of a `MongoDBOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `MongoDBOpsRequest`. It can have the following three values:

| Phase       | Meaning                                                                            |
|-------------|------------------------------------------------------------------------------------|
| Successful  | KubeDB has successfully performed the operation requested in the MongoDBOpsRequest |
| Progressing | KubeDB has started the execution of the applied MongoDBOpsRequest                  |
| Failed      | KubeDB has failed the operation requested in the MongoDBOpsRequest                 |
| Denied      | KubeDB has denied the operation requested in the MongoDBOpsRequest                 |
| Skipped     | KubeDB has skipped the operation requested in the MongoDBOpsRequest                |

Important: Ops-manager Operator can skip an opsRequest, only if its execution has not been started yet & there is a newer opsRequest applied in the cluster. `spec.type` has to be same as the skipped one, in this case.

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `MongoDBOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `MongoDBOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. MongoDBOpsRequest has the following types of conditions:

| Type                          | Meaning                                                                   |
| ----------------------------- | ------------------------------------------------------------------------- |
| `Progressing`                 | Specifies that the operation is now in the progressing state              |
| `Successful`                  | Specifies such a state that the operation on the database was successful. |
| `HaltDatabase`               | Specifies such a state that the database is halted by the operator        |
| `ResumeDatabase`              | Specifies such a state that the database is resumed by the operator       |
| `Failed`                      | Specifies such a state that the operation on the database failed.         |
| `StartingBalancer`            | Specifies such a state that the balancer has successfully started         |
| `StoppingBalancer`            | Specifies such a state that the balancer has successfully stopped         |
| `UpdateShardImage`            | Specifies such a state that the Shard Images has been updated             |
| `UpdateReplicaSetImage`       | Specifies such a state that the Replicaset Image has been updated         |
| `UpdateConfigServerImage`     | Specifies such a state that the ConfigServer Image has been updated       |
| `UpdateMongosImage`           | Specifies such a state that the Mongos Image has been updated             |
| `UpdateStatefulSetResources`  | Specifies such a state that the Statefulset resources has been updated    |
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
