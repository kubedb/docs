---
title: MongoDBOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: es-opsrequest-concepts
    name: ElasticsearchOpsRequest
    parent: es-concepts-elasticsearch
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# ElasticsearchOpsRequest

## What is MongoDBOpsRequest

`ElasticsearchOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for the [Elasticsearch](https://www.elastic.co/guide/index.html) administrative operations like database version upgrading, horizontal scaling, vertical scaling, etc. in a Kubernetes native way.

## ElasticsearchOpsRequest Specifications

Like any official Kubernetes resource, a `ElasticsearchOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: es-upgrade
  namespace: demo
spec:
  type: Upgrade
  databaseRef:
    name: es
  upgrade:
    targetVersion: 7.5.2-searchguard
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

Here, we are going to describe the various sections of a `ElasticsearchOpsRequest` CRD.

### spec.type

`spec.type` is a `required` field that specifies the kind of operation that will be applied to the Elasticsearch. The following types of operations are allowed in the `ElasticsearchOpsRequest`:

- `Restart` - is used to perform a smart restart of the Elasticsearch cluster.
- `Upgrade` - is used to upgrade the version of the Elasticsearch in a managed way. The necessary information required for upgrading the version, must be provided in `spec.upgrade` field.
- `VerticalScaling` - is used to vertically scale the Elasticsearch nodes (ie. pods). The necessary information required for vertical scaling, must be provided in `spec.verticalScaling` field.
- `HorizontalScaling` - is used to horizontally scale the Elasticsearch nodes (ie. pods). The necessary information required for horizontal scaling, must be provided in `spec.horizontalScaling` field.
- `VolumeExpansion` - is used to expand the storage of the Elasticsearch nodes (ie. pods). The necessary information required for volume expansion, must be provided in `spec.volumeExpansion` field.
- `ReconfigureTLS` - is used to configure the TLS configuration of a running Elasticsearch cluster. The necessary information required for reconfiguring the TLS, must be provided in `spec.tls` field.

> Note: You can only perform one type of operation by using an `ElasticsearchOpsRequest` custom resource object. For example, if you want to upgrade your database and scale up its replica then you will need to create two separate `ElasticsearchOpsRequest`. At first, you will have to create an `ElasticsearchOpsRequest` for upgrading. Once the upgrade is completed, then you can create another `ElasticsearchOpsRequest` for scaling. You should not create two `ElasticsearchOpsRequest` simultaneously.

### spec.databaseRef

`spec.databaseRef` is a `required` field that points to the [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- `databaseRef.name` - specifies the name of the [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch.md) object.

> Note: The `ElasticsearchOpsRequest` should be on the same namespace as the referring `Elasticsearch` object.

### spec.upgrade

`spec.upgrade` is an `optional` field, but it will act as a `required` field when the `spec.type` is set to `Upgrade`.
It specifies the desired version information required for the Elasticsearch version upgrade. This field consists of the following sub-fields:

- `upgrade.targetVersion` refers to an [ElasticsearchVersion](/docs/guides/elasticsearch/concepts/catalog.md) CR name that contains the Elasticsearch version information required to perform the upgrade.

> KubeDB does not support downgrade for Elasticsearch.

#### spec.horizontalScaling

If you want to scale-up or scale-down your MongoDB cluster or different components of it, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.replicas` indicates the desired number of nodes for MongoDB replicaset cluster after scaling. For example, if your cluster currently has 4 replicaset nodes, and you want to add additional 2 nodes then you have to specify 6 in `spec.horizontalScaling.replicas` field. Similarly, if you want to remove one node from the cluster, you have to specify 3 in `spec.horizontalScaling.replicas` field.
- `spec.horizontalScaling.configServer.replicas` indicates the desired number of ConfigServer nodes for Sharded MongoDB cluster after scaling.
- `spec.horizontalScaling.mongos.replicas` indicates the desired number of Mongos nodes for Sharded MongoDB cluster after scaling.
- `spec.horizontalScaling.shard` indicates the configuration of shard nodes for Sharded MongoDB cluster after scaling. This field consists of the following sub-field:
  - `spec.horizontalScaling.shard.replicas` indicates the number of replicas each shard will have after scaling.
  - `spec.horizontalScaling.shard.shards` indicates the number of shards after scaling

#### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `MongoDB` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.standalone` indicates the desired resources for standalone MongoDB database after scaling.
- `spec.verticalScaling.replicaSet` indicates the desired resources for replicaSet of MongoDB database after scaling.
- `spec.verticalScaling.mongos` indicates the desired resources for Mongos nodes of Sharded MongoDB database after scaling.
- `spec.verticalScaling.configServer` indicates the desired resources for ConfigServer nodes of Sharded MongoDB database after scaling.
- `spec.verticalScaling.shard` indicates the desired resources for Shard nodes of Sharded MongoDB database after scaling.
- `spec.verticalScaling.exporter` indicates the desired resources for the `exporter` container.

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

#### spec.volumeExpansion

> To use the volume expansion feature the storage class must support volume expansion

If you want to expand the volume of your MongoDB cluster or different components of it, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.volumeExpansion.standalone` indicates the desired size for the persistent volume of a standalone MongoDB database.
- `spec.volumeExpansion.replicaSet` indicates the desired size for the persistent volume of replicaSets of a MongoDB database.
- `spec.volumeExpansion.configServer` indicates the desired size for the persistent volume of the config server of a sharded MongoDB database.
- `spec.volumeExpansion.shard` indicates the desired size for the persistent volume of shards of a sharded MongoDB database.

All of them refer to [Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#quantity-resource-core) types of Kubernetes.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    shard: "2Gi"
```

This will expand the volume size of all the shard nodes to 2 GB.

#### spec.customConfig

If you want to reconfigure your Running MongoDB cluster or different components of it with new custom configuration, you have to specify `spec.customConfig` section. This field consists of the following sub-field:

- `spec.customConfig.standalone` indicates the desired new custom configuration for a standalone MongoDB database.
- `spec.customConfig.replicaSet` indicates the desired new custom configuration for replicaSet of a MongoDB database.
- `spec.customConfig.configServer` indicates the desired new custom configuration for config servers of a sharded MongoDB database.
- `spec.customConfig.mongos` indicates the desired new custom configuration for the mongos nodes of a sharded MongoDB database.
- `spec.customConfig.shard` indicates the desired new custom configuration for the shard nodes of a sharded MongoDB database.

All of them has the following sub-fields:

- `configMap` points to a configMap in the same namespace of a MongoDB resource, which contains the new custom configurations. If there are any configmap sources before, this configmap will replace it.
- `data` contains the new custom config which will be merged with the previous configuration.

### MongoDBOpsRequest `Status`

`.status` describes the current state and progress of a `MongoDBOpsRequest` operation. It has the following fields:

#### status.phase

`status.phase` indicates the overall phase of the operation for this `MongoDBOpsRequest`. It can have the following three values:

| Phase      | Meaning                                                                            |
| ---------- | ---------------------------------------------------------------------------------- |
| Successful | KubeDB has successfully performed the operation requested in the MongoDBOpsRequest |
| Failed     | KubeDB has failed the operation requested in the MongoDBOpsRequest                 |
| Denied     | KubeDB has denied the operation requested in the MongoDBOpsRequest                 |

#### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `MongoDBOpsRequest` controller.

#### status.conditions

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





**Sample `MongoDBOpsRequest` Objects for Horizontal Scaling of different component of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MongoDBOpsRequest
metadata:
  name: mops-hscale-down-configserver
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: mg-sharding
  horizontalScaling:
    shard:
      shards: 3
      replicas: 2
    configServer:
      replicas: 2
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
  customConfig:
    replicaSet:
      data:
        mongod.conf: |
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
  customConfig:
    shard:
      data:
        mongod.conf: |
          net:
            maxIncomingConnections: 30000
    configServer:
      data:
        mongod.conf: |
          net:
            maxIncomingConnections: 30000
    mongos:
      data:
        mongod.conf: |
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
  customConfig:
    standalone:
      data:
        mongod.conf: |
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
  customConfig:
    replicaSet:
      configMap:
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
  customConfig:
    shard:
      configMap:
        name: new-custom-config
    confiServer:
      configMap:
        name: new-custom-config
    mongos:
      configMap:
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
  customConfig:
    standalone:
      configMap:
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

