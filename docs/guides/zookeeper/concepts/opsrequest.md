---
title: ZooKeeperOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: zk-opsrequest
    name: ZooKeeperOpsRequest
    parent: zk-concepts-zookeeper
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ZooKeeperOpsRequest

## What is ZooKeeperOpsRequest

`ZooKeeperOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [ZooKeeper](https://zookeeper.apache.org/) administrative operations like database version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## ZooKeeperOpsRequest CRD Specifications

Like any official Kubernetes resource, a `ZooKeeperOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `ZooKeeperOpsRequest` CRs for different administrative operations is given below:

**Sample `ZooKeeperOpsRequest` for updating database:**

Let's assume that you have a KubeDB managed ZooKeeper cluster named `zk-quickstart` running on your kubernetes with version `3.8.3`. Now, You can update it's version to `3.9.1` using the following manifest.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: upgrade-topology
  namespace: demo
spec:
  databaseRef:
    name: zk-quickstart
  type: UpdateVersion
  updateVersion:
    targetVersion: 3.9.1
```

**Sample `ZooKeeperOpsRequest` Objects for Horizontal Scaling of the database Cluster:**

You can scale up and down your zookeeper cluster horizontally. 
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: zk-quickstart
  horizontalScaling:
    replicas: 5
```

**Sample `ZooKeeperOpsRequest` Objects for Vertical Scaling of the database cluster:**

You can vertically scale up or down your cluster by updating the requested cpu, memory or, by limiting them.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zookeeper-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: zookeeper
  verticalScaling:
    node:
      resources:
        requests:
          cpu: 600m
          memory: 1.2Gi
        limits:
          cpu: 1
          memory: 2Gi
```

**Sample `ZooKeeperOpsRequest` Objects for Reconfiguring database cluster:**

Reconfigure your cluster by applying new configuration via `zoo.conf` file.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: myops-reconfigure-config
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: zk-quickstart
  type: Reconfigure
  configuration:
    applyConfig:
      zoo.cfg: |
        max_connections = 300
        read_buffer_size = 1048576
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: myops-reconfigure-config
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: zk-quickstart
  type: Reconfigure
  configuration:
    removeCustomConfig: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: myops-reconfigure-config
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: zk-quickstart
  type: Reconfigure
  configuration:
    configSecret:
      name: new-config-secret
    restart: "true"
```

**Sample `ZooKeeperOpsRequest` Objects for Volume Expansion of database cluster:**

You can expand ZooKeeper storage volume in both online and offline mode.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: online-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: zk-quickstart
  volumeExpansion:
    mode: "Online"
    node: 3Gi
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: offline-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: zk-quickstart
  volumeExpansion:
    mode: "Offline"
    node: 4Gi
```

**Sample `ZooKeeperOpsRequest` Objects for Reconfiguring TLS of the database:**

You can use this Ops-Request to Add, Update, Remove or Rotate Your certificates used in TLS connectivity.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: myops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: zk-quickstart
  tls:
    rotateCertificates: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zkops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: zk-quickstart
  tls:
    issuerRef:
      name: zk-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: myops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: zk-quickstart
  tls:
    remove: true
```

Here, we are going to describe the various sections of a `ZooKeeperOpsRequest` crd.

A `ZooKeeperOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [ZooKeeper](/docs/guides/zookeeper/concepts/zookeeper.md) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `ZooKeeperOpsRequest`.

- `Upgrade` / `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `ZooKeeperOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `ZooKeeperOpsRequest`. At first, you have to create a `ZooKeeperOpsRequest` for updating. Once it is completed, then you can create another `ZooKeeperOpsRequest` for scaling.

> Note: There is an exception to the above statement. It is possible to specify both `spec.configuration` & `spec.verticalScaling` in a OpsRequest of type `VerticalScaling`.

### spec.updateVersion

If you want to update your ZooKeeper version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [ZooKeeperVersion](/docs/guides/zookeeper/concepts/catalog.md) CR that contains the ZooKeeper version information where you want to update.


### spec.horizontalScaling

If you want to scale-up or scale-down your ZooKeeper cluster or different components of it, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.replicas` indicates the desired number of pods for ZooKeeper cluster after scaling. For example, if your cluster currently has 4 pods, and you want to add additional 2 pods then you have to specify 6 in `spec.horizontalScaling.replicas` field. Similarly, if you want to remove one pod from the cluster, you have to specify 3 in `spec.horizontalScaling.replicas` field.

### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `ZooKeeper` resources like `cpu`, `memory` etc. that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.node` indicates the desired resources for PetSet of ZooKeeper after scaling.

It has the below structure:

```yaml
requests:
  memory: "600Mi"
  cpu: "0.5"
limits:
  memory: "800Mi"
  cpu: "0.8"
```

Here, when you specify the resource request, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for the container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. You can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).

### spec.volumeExpansion

> To use the volume expansion feature the storage class must support volume expansion

If you want to expand the volume of your ZooKeeper standalone or cluster, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.volumeExpansion.node` indicates the desired size for the persistent volume of a ZooKeeper.
- `spec.volumeExpansion.mode` indicates the mode of volume expansion. It can be `online` or `offline` based on the storage class.


All of them refer to Quantity types of Kubernetes.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    node: "2Gi"
```

This will expand the volume size of all the ZooKeeper nodes to 2 GB.

### spec.configuration

If you want to reconfigure your Running ZooKeeper cluster or different components of it with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-field:

- `configSecret` points to a secret in the same namespace of a ZooKeeper resource, which contains the new custom configurations. If there are any configSecret set before in the database, this secret will replace it.
- `applyConfig` contains the new custom config as a string which will be merged with the previous configuration.

- `applyConfig` is a map where key supports 1 values, namely `zoo.conf`.

```yaml
  applyConfig:
    zoo.cfg: |
      max_connections = 300
      read_buffer_size = 1048576
```

- `removeCustomConfig` is a boolean field. Specify this field to true if you want to remove all the custom configuration from the deployed ZooKeeper server.

- `restart` significantly reduces unnecessary downtime.
  - `auto` (default): restart only if required (determined by ops manager operator)
  - `false`: no restart
  - `true`: always restart



### spec.tls

If you want to reconfigure the TLS configuration of your ZooKeeper cluster i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/zookeeper/concepts/zookeeper.md#spectls).
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.
- `spec.tls.remove` specifies that we want to remove tls from this database.
- `spec.tls.sslMode` specifies what will be the ssl mode of the cluster allowed values are: disable,allow,prefer,require,verify-ca,verify-full
- `spec.tls.clientAuthMode` specifies what will be the client authentication mode of the cluster allowed values are: md5,scram,cert

### spec.timeout
As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second).
If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.


### ZooKeeperOpsRequest `Status`

`.status` describes the current state and progress of a `ZooKeeperOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `ZooKeeperOpsRequest`. It can have the following three values:

| Phase       | Meaning                                                                             |
|-------------|-------------------------------------------------------------------------------------|
| Successful  | KubeDB has successfully performed the operation requested in the ZooKeeperOpsRequest |
| Progressing | KubeDB has started the execution of the applied ZooKeeperOpsRequest                  |
| Failed      | KubeDB has failed the operation requested in the ZooKeeperOpsRequest                 |
| Denied      | KubeDB has denied the operation requested in the ZooKeeperOpsRequest                 |
| Skipped     | KubeDB has skipped the operation requested in the ZooKeeperOpsRequest                |

Important: Ops-manager Operator can skip an opsRequest, only if its execution has not been started yet & there is a newer opsRequest applied in the cluster. `spec.type` has to be same as the skipped one, in this case.

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `ZooKeeperOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `ZooKeeperOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. ZooKeeperOpsRequest has the following types of conditions:

| Type                           |   | Meaning                                                                   |
|--------------------------------|---|---------------------------------------------------------------------------|
| `Progressing`                  |   | Specifies that the operation is now in the progressing state              |
| `Successful`                   |   | Specifies such a state that the operation on the database was successful. |
| `DatabasePauseSucceeded`       |   | Specifies such a state that the database is paused by the operator        |
| `ResumeDatabase`               |   | Specifies such a state that the database is resumed by the operator       |
| `Failed`                       |   | Specifies such a state that the operation on the database failed.         |
| `UpdatePetSetResources`        |   | Specifies such a state that the PetSet resources has been updated         |
| `UpdatePetSet`                 |   | Specifies such a state that the PetSet  has been updated                  |
| `IssueCertificatesSucceeded`   |   | Specifies such a state that the tls certificate issuing is successful     |
| `UpdateDatabase`               |   | Specifies such a state that the CR of ZooKeeper is updated                 |

- The `status` field is a string, with possible values `True`, `False`, and `Unknown`.
    - `status` will be `True` if the current transition succeeded.
    - `status` will be `False` if the current transition failed.
    - `status` will be `Unknown` if the current transition was denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
