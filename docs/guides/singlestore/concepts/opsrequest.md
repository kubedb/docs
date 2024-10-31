---
title: SingleStoreOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: sdb-opsrequest-concepts
    name: SingleStoreOpsRequest
    parent:  sdb-concepts-singlestore
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SingleStoreOpsRequest

## What is SingleStoreOpsRequest

`SingleStoreOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [SingleStore](https://www.singlestore.com/) administrative operations like database version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## SingleStoreOpsRequest CRD Specifications

Like any official Kubernetes resource, a `SingleStoreOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `SingleStoreOpsRequest` CRs for different administrative operations is given below:

**Sample `SingleStoreOpsRequest` for updating database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-version-upd
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: sdb
  updateVersion:
    targetVersion: 8.7.10
  timeout: 5m
  apply: IfReady
```

**Sample `SingleStoreOpsRequest` Objects for Horizontal Scaling of different component of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-hscale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sdb
  horizontalScaling:
    aggregator: 2
    leaf: 3
```

**Sample `SingleStoreOpsRequest` Objects for Vertical Scaling of different component of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-scale
  namespace: demo
spec:
  type: VerticalScaling  
  databaseRef:
    name: sdb-sample
  verticalScaling:
    leaf:
      resources:
        requests:
          memory: "2500Mi"
          cpu: "0.7"
        limits:
          memory: "2500Mi"
          cpu: "0.7"
    coordinator:
      resources:
        requests:
          memory: "2500Mi"
          cpu: "0.7"
        limits:
          memory: "2500Mi"
          cpu: "0.7"
    node:
      resources:
        requests:
          memory: "2500Mi"
          cpu: "0.7"
        limits:
          memory: "2500Mi"
          cpu: "0.7"
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-scale
  namespace: demo
spec:
  type: VerticalScaling  
  databaseRef:
    name: sdb-standalone
  verticalScaling:
    node:
      resources:
        requests:
          memory: "2500Mi"
          cpu: "0.7"
        limits:
          memory: "2500Mi"
          cpu: "0.7"
```

**Sample `SingleStoreOpsRequest` Objects for Reconfiguring different database components:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-reconfigure-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sdb-sample
  configuration:
    aggregator:  
      applyConfig:
        sdb-apply.cnf: |-
          max_connections = 550
    leaf:  
      applyConfig:
        sdb-apply.cnf: |-
          max_connections = 550
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-reconfigure-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sdb-standalone
  configuration:
    node:  
      applyConfig:
        sdb-apply.cnf: |-
          max_connections = 550
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-reconfigure-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sdb-sample
  configuration:
    aggregator:  
      configSecret:
        name: sdb-new-custom-config
    leaf:  
      configSecret:
        name: sdb-new-custom-config
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-reconfigure-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sdb-standalone
  configuration:
    node:  
      configSecret:
        name: sdb-new-custom-config
```

**Sample `SingleStoreOpsRequest` Objects for Volume Expansion of different database components:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-volume-ops
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: sdb-sample
  volumeExpansion:
    mode: "Offline"
    aggregator: 10Gi
    leaf: 20Gi
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-volume-ops
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: sdb-standalone
  volumeExpansion:
    mode: "Online"
    node: 20Gi
```

**Sample `SingleStoreOpsRequest` Objects for Reconfiguring TLS of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-tls-reconfigure
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sdb-sample
  tls:
    issuerRef:
      name: sdb-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - singlestore
          organizationalUnits:
            - client
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-tls-reconfigure
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sdb-sample
  tls:
    rotateCertificates: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdb-tls-reconfigure
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sdb-sample
  tls:
    remove: true
```

Here, we are going to describe the various sections of a `SingleStoreOpsRequest` crd.

A `SingleStoreOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [SingleStore](/docs/guides/singlestore/concepts/singlestore.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [SingleStore](/docs/guides/singlestore/concepts/singlestore.md) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `SingleStoreOpsRequest`.

- `Upgrade` / `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `SingleStoreOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `SingleStoreOpsRequest`. At first, you have to create a `SingleStoreOpsRequest` for updating. Once it is completed, then you can create another `SingleStoreOpsRequest` for scaling. 

> Note: There is an exception to the above statement. It is possible to specify both `spec.configuration` & `spec.verticalScaling` in a OpsRequest of type `VerticalScaling`.

### spec.updateVersion

If you want to update you SingleStore version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [SingleStoreVersion](/docs/guides/singlestore/concepts/catalog.md) CR that contains the SingleStore version information where you want to update.

Have a look on the [`updateConstraints`](/docs/guides/singlestore/concepts/catalog.md#specupdateconstraints) of the singlestoreVersion spec to know which versions are supported for updating from the current version.
```yaml
kubectl get sdbversion <current-version> -o=jsonpath='{.spec.updateConstraints}' | jq
```

> You can only update between SingleStore versions. KubeDB does not support downgrade for SingleStore.

### spec.horizontalScaling

If you want to scale-up or scale-down your SingleStore cluster or different components of it, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.aggregator.replicas` indicates the desired number of aggregator nodes for cluster mode after scaling.
- `spec.horizontalScaling.leaf.replicas` indicates the desired number of leaf nodes for cluster mode after scaling.

### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `SingleStore` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.node` indicates the desired resources for standalone SingleStore database after scaling.
- `spec.verticalScaling.aggregator` indicates the desired resources for aggregator node of SingleStore cluster after scaling.
- `spec.verticalScaling.leaf` indicates the desired resources for leaf nodes of SingleStore cluster after scaling.
- `spec.verticalScaling.coordinator` indicates the desired resources for the coordinator container.

All of them has the below structure:

```yaml
requests:
  memory: "2000Mi"
  cpu: "0.7"
limits:
  memory: "3000Mi"
  cpu: "0.9"
```

Here, when you specify the resource request, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for the container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. You can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).

### spec.volumeExpansion

> To use the volume expansion feature the storage class must support volume expansion

If you want to expand the volume of your SingleStore cluster or different components of it, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.mode` specifies the volume expansion mode. Supported values are `Online` & `Offline`. The default is `Online`.
- `spec.volumeExpansion.node` indicates the desired size for the persistent volume of a standalone SingleStore database.
- `spec.volumeExpansion.aggregator` indicates the desired size for the persistent volume of aggregator node of cluster.
- `spec.volumeExpansion.leaf` indicates the desired size for the persistent volume of leaf node of cluster.

All of them refer to [Quantity](https://v1-22.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#quantity-resource-core) types of Kubernetes.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    aggregator: "20Gi"
```

This will expand the volume size of all the shard nodes to 20 GB.

### spec.configuration

If you want to reconfigure your Running SingleStore cluster or different components of it with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-field:

- `spec.configuration.standalone` indicates the desired new custom configuration for a standalone SingleStore database.
- `spec.configuration.aggregator` indicates the desired new custom configuration for aggregator node of cluster mode.
- `spec.configuration.leaf` indicates the desired new custom configuration for leaf node of cluster mode.

All of them has the following sub-fields:

- `configSecret` points to a secret in the same namespace of a SingleStore resource, which contains the new custom configurations. If there are any configSecret set before in the database, this secret will replace it.
- `applyConfig` contains the new custom config as a string which will be merged with the previous configuration. 

- `applyConfig` is a map where key supports values, namely `sdb-apply.cnf`. And value represents the corresponding configurations.
KubeDB provisioner operator applies these two directly while reconciling.

```yaml
  applyConfig:
    sdb-apply.cnf: |-
      max_connections = 550
```

- `removeCustomConfig` is a boolean field. Specify this field to true if you want to remove all the custom configuration from the deployed singlestore server.

### spec.tls

If you want to reconfigure the TLS configuration of your database i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/singlestore/concepts/singlestore.md#spectls).
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.
- `spec.tls.remove` specifies that we want to remove tls from this database.


### spec.timeout
As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second). 
If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.


### SingleStoreOpsRequest `Status`

`.status` describes the current state and progress of a `SingleStoreOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `SingleStoreOpsRequest`. It can have the following three values:

| Phase       | Meaning                                                                                |
|-------------|----------------------------------------------------------------------------------------|
| Successful  | KubeDB has successfully performed the operation requested in the SingleStoreOpsRequest |
| Progressing | KubeDB has started the execution of the applied SingleStoreOpsRequest                  |
| Failed      | KubeDB has failed the operation requested in the SingleStoreOpsRequest                 |
| Denied      | KubeDB has denied the operation requested in the SingleStoreOpsRequest                 |
| Skipped     | KubeDB has skipped the operation requested in the SingleStoreOpsRequest                |

Important: Ops-manager Operator can skip an opsRequest, only if its execution has not been started yet & there is a newer opsRequest applied in the cluster. `spec.type` has to be same as the skipped one, in this case.

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `SingleStoreOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `SingleStoreOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. SingleStoreOpsRequest has the following types of conditions:

| Type                        | Meaning                                                                    |
|-----------------------------|----------------------------------------------------------------------------|
| `Progressing`               | Specifies that the operation is now in the progressing state               |
| `Successful`                | Specifies such a state that the operation on the database was successful.  |
| `HaltDatabase`              | Specifies such a state that the database is halted by the operator         |
| `ResumeDatabase`            | Specifies such a state that the database is resumed by the operator        |
| `Failed`                    | Specifies such a state that the operation on the database failed.          |
| `StartingBalancer`          | Specifies such a state that the balancer has successfully started          |
| `StoppingBalancer`          | Specifies such a state that the balancer has successfully stopped          |
| `UpdatePetSetResources`     | Specifies such a state that the PetSet resources has been updated          |
| `UpdateAggregatorResources` | Specifies such a state that the Aggregator resources has been updated      |
| `UpdateLeafResources`       | Specifies such a state that the Leaf resources has been updated            |
| `UpdateNodeResources`       | Specifies such a state that the node has been updated                      |
| `ScaleDownAggregator`       | Specifies such a state that the scale down operation of aggregator         |
| `ScaleUpAggregator`         | Specifies such a state that the scale up operation of aggregator           |
| `ScaleUpLeaf`               | Specifies such a state that the scale up operation of leaf                 |
| `ScaleDownleaf`             | Specifies such a state that the scale down operation of leaf               |
| `VolumeExpansion`           | Specifies such a state that the volume expansion operation of the database |
| `ReconfigureAggregator`     | Specifies such a state that the reconfiguration of aggregator nodes        |
| `ReconfigureLeaf`           | Specifies such a state that the reconfiguration of leaf nodes              |
| `ReconfigureNode`           | Specifies such a state that the reconfiguration of standalone nodes        |

- The `status` field is a string, with possible values `True`, `False`, and `Unknown`.
  - `status` will be `True` if the current transition succeeded.
  - `status` will be `False` if the current transition failed.
  - `status` will be `Unknown` if the current transition was denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.