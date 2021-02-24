---
title: ElasticsearchOpsRequests CRD
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

## What is ElasticsearchOpsRequest

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

`spec.upgrade` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `Upgrade`.
It specifies the desired version information required for the Elasticsearch version upgrade. This field consists of the following sub-fields:

- `upgrade.targetVersion` refers to an [ElasticsearchVersion](/docs/guides/elasticsearch/concepts/catalog.md) CR name that contains the Elasticsearch version information required to perform the upgrade.

> KubeDB does not support downgrade for Elasticsearch.

**Samples:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: es-topology-upgrade
  namespace: demo
spec:
  type: Upgrade
  databaseRef:
    name: es
  upgrade:
    targetVersion: 7.5.2-searchguard
```

### spec.horizontalScaling

`spec.horizontalScaling` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `HorizontalScaling`.
It specifies the necessary information required to horizontally scale the Elasticsearch nodes (ie. pods). It consists of the following sub-field:

- `horizontalScaling.node` - specifies the desired number of nodes for the Elasticsearch cluster running in combined mode (ie. `Elasticsearch.spec.topology` is `empty`).  The value should be greater than the maximum value of replication for the shard of any index. For example, if a shard has `x` replicas, `x+1` data nodes are required to allocate them.

- `horizontalScaling.topology` - specifies the desired number of different type of nodes for the Elasticsearch cluster running in cluster topology mode (ie. `Elasticsearch.spec.topology` is `not empty`).
  - `topology.master` - specifies the desired number of master nodes. The value should be greater than zero ( >= 1 ).
  - `toplogy.ingest` - specifies the desired number of ingest nodes. The value should be greater than zero ( >= 1 ).
  - `topology.data` - specifies the desired number of data nodes. The value should be greater than the maximum value of replication for the shard of any index. For example, if a shard has `x` replicas, `x+1` data nodes are required to allocate them.

**Samples:**

- Horizontally scale combined nodes:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: hscale-combined
    namespace: demo
  spec:
    type: HorizontalScaling
    databaseRef:
      name: es
    horizontalScaling:
      node: 4
  ```

- Horizontally scale cluster topology:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: hscale-topology
    namespace: demo
  spec:
    type: HorizontalScaling
    databaseRef:
      name: es
    horizontalScaling:
      topology:
        master: 2
        ingest: 2
        data: 3
  ```

- Horizontally scale only ingest nodes:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: hscale-ingest-nodes
    namespace: demo
  spec:
    type: HorizontalScaling
    databaseRef:
      name: es
    horizontalScaling:
      topology:
        ingest: 4
  ```

### spec.verticalScaling

`spec.verticalScaling` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `VerticalScaling`. It specifies the necessary information required to vertically scale the Elasticsearch node resources (ie. `cpu`, `memory`). It consists of the following sub-field:

- `verticalScaling.node` - specifies the desired node resources for the Elasticsearch cluster running in combined mode (ie. `Elasticsearch.spec.topology` is `empty`).
- `verticalScaling.topology` - specifies the desired node resources for different type of node of the Elasticsearch running in cluster topology mode (ie. `Elasticsearch.spec.topology` is `not empty`).
  - `topology.master` - specifies the desired resources for the master nodes. It takes input same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).
  - `topology.data` - specifies the desired node resources for the data nodes. It takes input  same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).
  - `topology.ingest` - specifies the desired node resources for the ingest nodes. It takes input  same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).

> Note: It is recommended not to use resources below the default one; `cpu: 500m, memory: 1Gi`.

**Samples:**

- Vertically scale combined nodes:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: vscale-combined
    namespace: demo
  spec:
    type: VerticalScaling
    databaseRef:
      name: es
    verticalScaling:
      node:
        limits:
          cpu: 1000m
          memory: 2Gi
        requests:
          cpu: 500m
          memory: 1Gi
  ```

- Vertically scale topology cluster:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: vscale-topology
    namespace: demo
  spec:
    type: VerticalScaling
    databaseRef:
      name: es
    verticalScaling:
      topology:
        master:
          limits:
            cpu: 750m
            memory: 800Mi
        data:
          requests:
            cpu: 760m
            memory: 900Mi
        ingest:
          limits:
            cpu: 900m
            memory: 1.2Gi
          requests:
            cpu: 800m
            memory: 1Gi  
  ```

- Vertically scale only data nodes:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: vscale-data-nodes
    namespace: demo
  spec:
    type: VerticalScaling
    databaseRef:
      name: es
    verticalScaling:
      topology:
        data:
          limits:
            cpu: 900m
            memory: 1.2Gi
          requests:
            cpu: 800m
            memory: 1Gi  
  ```

### spec.volumeExpansion

> Note: To use the volume expansion feature the StorageClass must support volume expansion.

`spec.volumeExpansion` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `VolumeExpansion`. It specifies the necessary information required to expand the storage of the Elasticsearch node. It consists of the following sub-field:

- `volumeExpansion.node` - specifies the desired size of the persistent volume for the Elasticsearch node running in combined mode (ie. `Elasticsearch.spec.topology` is `empty`).
- `volumeExpansion.topology` - specifies the desired size of the persistent volumes for the different types of nodes of the Elasticsearch cluster running in cluster topology mode (ie. `Elasticsearch.spec.topology` is `not empty`).
  - `topology.master` - specifies the desired size of the persistent volume for the master nodes.
  - `topology.data` - specifies the desired size of the persistent volume for the data nodes.
  - `topology.ingest` - specifies the desired size of the persistent volume for the ingest nodes.

All of them refer to [Quantity](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#quantity-resource-core) types of Kubernetes.

> Note: Make sure that the requested volume is greater than the current volume.

**Samples:**

- Expand volume for combined nodes:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: volume-expansion-combined
    namespace: demo
  spec:
    type: VolumeExpansion
    databaseRef:
      name: es
    volumeExpansion:
      node: 4Gi
  ```

- Expand volume for cluster topology:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: volume-expansion-topology
    namespace: demo
  spec:
    type: VolumeExpansion
    databaseRef:
      name: es
    volumeExpansion:
      topology:
        master: 2Gi
        data: 3Gi
        ingest: 4Gi
  ```

- Expand volume for only data nodes:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: volume-expansion-data-nodes
    namespace: demo
  spec:
    type: VolumeExpansion
    databaseRef:
      name: es
    volumeExpansion:
      topology:
        data: 5Gi
  ```

## ElasticsearchOpsRequest `Status`

`.status` describes the current state and progress of a `ElasticsearchOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `ElasticsearchOpsRequest`. It can have the following three values:

| Phase       | Meaning                                                                             |
| :--------:  | ----------------------------------------------------------------------------------  |
| Progressing | KubeDB has started to process the Ops request                                       |
| Successful  | KubeDB has successfully performed all the operations needed for the Ops request     |
| Failed      | KubeDB has failed while performing the operations needed for the Ops request        |

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `ElasticsearchOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `ElasticsearchOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition.
- The `status` field is a string, with possible values `True`, `False`, and `Unknown`.
  - `status` will be `True` if the current transition succeeded.
  - `status` will be `False` if the current transition failed.
  - `status` will be `Unknown` if the current transition was denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.

ElasticsearchOpsRequest has the following types of conditions:

| Type                            | Meaning                                                                   |
| -----------------------------   | ------------------------------------------------------------------------- |
| `Progressing`                   | The operator has started to process the Ops request                       |
| `Successful`                    | The Ops request has successfully executed                                 |
| `Failed`                        | The operation on the database failed                                      |
| `OrphanStatefulSetPods`         | The statefulSet has deleted leaving the pods orphaned                     |
| `ReadyStatefulSets`             | The StatefulSet are ready                                                 |
| `ScaleDownCombinedNode`         | Scaled down the combined nodes                                            |
| `ScaleDownDataNode`             | Scaled down the data nodes                                                |
| `ScaleDownIngestNode`           | Scaled down the ingest nodes                                              |
| `ScaleDownMasterNode`           | Scaled down the master nodes                                              |
| `ScaleUpCombinedNode`           | Scaled up the combined nodes                                              |
| `ScaleUpDataNode`               | Scaled up the data nodes                                                  |
| `ScaleUpIngestNode`             | Scaled up the ingest nodes                                                |
| `ScaleUpMasterNode`             | Scaled up the master nodes                                                |
| `UpdateCombinedNodePVCs`        | Updated combined node PVCs                                                |
| `UpdateDataNodePVCs`            | Updated data node PVCs                                                    |
| `UpdateIngestNodePVCs`          | Updated ingest node PVCs                                                  |
| `UpdateMasterNodePVCs`          | Updated master node PVCs                                                  |
| `UpdateNodeResources`           | Updated node resources                                                    |
