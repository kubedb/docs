---
title: PerconaXtraDBOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-concepts-perconaxtradbopsrequest
    name: PerconaXtraDBOpsRequest
    parent: guides-perconaxtradb-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# PerconaXtraDBOpsRequest

## What is PerconaXtraDBOpsRequest

`PerconaXtraDBOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [PerconaXtraDB](https://docs.percona.com/percona-xtradb-cluster/8.0//) administrative operations like database version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## PerconaXtraDBOpsRequest CRD Specifications

Like any official Kubernetes resource, a `PerconaXtraDBOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `PerconaXtraDBOpsRequest` CRs for different administrative operations is given below:

**Sample `PerconaXtraDBOpsRequest` for updating database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: px-version-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: sample-pxc
  updateVersion:
    targetVersion: 8.0.26
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

**Sample `PerconaXtraDBOpsRequest` Objects for Horizontal Scaling of database cluster:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: px-scale-horizontal
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sample-pxc
  horizontalScaling:
    member : 5
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

**Sample `PerconaXtraDBOpsRequest` Objects for Vertical Scaling of the database cluster:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: px-scale-vertical
  namespace: demo
spec:
  type: VerticalScaling  
  databaseRef:
    name: sample-pxc
  verticalScaling:
    perconaxtradb:
      requests:
        memory: "600Mi"
        cpu: "0.1"
      limits:
        memory: "600Mi"
        cpu: "0.1"
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

**Sample `PerconaXtraDBOpsRequest` Objects for Reconfiguring PerconaXtraDB Database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: px-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-pxc
  configuration:   
    inlineConfig: |
      max_connections = 300
      read_buffer_size = 1234567
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

**Sample `PerconaXtraDBOpsRequest` Objects for Volume Expansion of PerconaXtraDB:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: px-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion  
  databaseRef:
    name: sample-pxc
  volumeExpansion:   
    mode: "Online"
    perconaxtradb: 2Gi
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

**Sample `PerconaXtraDBOpsRequest` Objects for Reconfiguring TLS of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: px-recon-tls-add
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-pxc
  tls:
    requireSSL: true
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: px-issuer
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
kind: PerconaXtraDBOpsRequest
metadata:
  name: px-recon-tls-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-pxc
  tls:
    rotateCertificates: true
```


Here, we are going to describe the various sections of a `PerconaXtraDBOpsRequest` crd.

A `PerconaXtraDBOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `PerconaXtraDBOpsRequest`.

- `Upgrade` / `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `PerconaXtraDBOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `PerconaXtraDBOpsRequest`. At first, you have to create a `PerconaXtraDBOpsRequest` for updating. Once it is completed, then you can create another `PerconaXtraDBOpsRequest` for scaling. You should not create two `PerconaXtraDBOpsRequest` simultaneously.

### spec.updateVersion

If you want to update your PerconaXtraDB version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [PerconaXtraDBVersion](/docs/guides/percona-xtradb/concepts/perconaxtradb-version/index.md) CR that contains the PerconaXtraDB version information where you want to update.

> You can only update between PerconaXtraDB versions. KubeDB does not support downgrade for PerconaXtraDB.

### spec.horizontalScaling

If you want to scale-up or scale-down your PerconaXtraDB cluster or different components of it, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:
- `spec.horizontalScaling.member` indicates the desired number of nodes for PerconaXtraDB cluster after scaling. For example, if your cluster currently has 4 nodes, and you want to add additional 2 nodes then you have to specify 6 in `spec.horizontalScaling.member` field. Similarly, if you want to remove one node from the cluster, you have to specify 3 in `spec.horizontalScaling.` field.

### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `PerconaXtraDB` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-field:

- `spec.verticalScaling.perconaxtradb` indicates the desired resources for PerconaXtraDB standalone or cluster after scaling.
- `spec.verticalScaling.exporter` indicates the desired resources for the `exporter` container.
- `spec.verticalScaling.coordinator` indicates the desired resources for the `coordinator` container.


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

If you want to expand the volume of your PerconaXtraDB standalone or cluster, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.volumeExpansion.perconaxtradb` indicates the desired size for the persistent volume of a PerconaXtraDB.
- `spec.volumeExpansion.mode` indicates the mode of volume expansion. It can be `online` or `offline` based on the storage class.


All of them refer to Quantity types of Kubernetes.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    perconaxtradb: "2Gi"
```

This will expand the volume size of all the perconaxtradb nodes to 2 GB.

### spec.configuration

If you want to reconfigure your Running PerconaXtraDB cluster with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-fields:
- `configSecret` points to a secret in the same namespace of a PerconaXtraDB resource, which contains the new custom configurations. If there are any configSecret set before in the database, this secret will replace it.
- `inlineConfig` contains the new custom config as a string which will be merged with the previous configuration.
- `removeCustomConfig` reomoves all the custom configs of the PerconaXtraDB server.

### spec.tls

If you want to reconfigure the TLS configuration of your database i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/percona-xtradb/concepts/perconaxtradb/#spectls).
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.
- `spec.tls.remove` specifies that we want to remove tls from this database.


### PerconaXtraDBOpsRequest `Status`

`.status` describes the current state and progress of a `PerconaXtraDBOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `PerconaXtraDBOpsRequest`. It can have the following three values:

| Phase      | Meaning                                                                            |
| ---------- | ---------------------------------------------------------------------------------- |
| Successful | KubeDB has successfully performed the operation requested in the PerconaXtraDBOpsRequest |
| Failed     | KubeDB has failed the operation requested in the PerconaXtraDBOpsRequest                 |
| Denied     | KubeDB has denied the operation requested in the PerconaXtraDBOpsRequest                 |

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `PerconaXtraDBOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `PerconaXtraDBOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. PerconaXtraDBOpsRequest has the following types of conditions:

| Type                          | Meaning                                                                   |
| ----------------------------- | ------------------------------------------------------------------------- |
| `Progressing`                 | Specifies that the operation is now in the progressing state              |
| `Successful`                  | Specifies such a state that the operation on the database was successful. |
| `Failed`                      | Specifies such a state that the operation on the database failed.         |
| `ScaleDownCluster`            | Specifies such a state that the scale down operation of replicaset        |
| `ScaleUpCluster`              | Specifies such a state that the scale up operation of replicaset          |
| `VolumeExpansion`             | Specifies such a state that the volume expansion operaton of the database |
| `Reconfigure`                 | Specifies such a state that the reconfiguration of replicaset nodes       |

- The `status` field is a string, with possible values `True`, `False`, and `Unknown`.
  - `status` will be `True` if the current transition succeeded.
  - `status` will be `False` if the current transition failed.
  - `status` will be `Unknown` if the current transition was denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
