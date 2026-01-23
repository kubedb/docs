---
title: MariaDBOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-concepts-mariadbopsrequest
    name: MariaDBOpsRequest
    parent: guides-mariadb-concepts
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MariaDBOpsRequest

## What is MariaDBOpsRequest

`MariaDBOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [MariaDB](https://www.mariadb.com/) administrative operations like database version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## MariaDBOpsRequest CRD Specifications

Like any official Kubernetes resource, a `MariaDBOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `MariaDBOpsRequest` CRs for different administrative operations is given below:

**Sample `MariaDBOpsRequest` for updating database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdops-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: sample-mariadb
  updateVersion:
    targetVersion: 10.5.23
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

**Sample `MariaDBOpsRequest` Objects for Horizontal Scaling of database cluster:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: mdps-scale-horizontal
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: sample-mariadb
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

**Sample `MariaDBOpsRequest` Objects for Vertical Scaling of the database cluster:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: md-scale-vertical
  namespace: demo
spec:
  type: VerticalScaling  
  databaseRef:
    name: sample-mariadb
  verticalScaling:
    mariadb:
      resources:
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

**Sample `MariaDBOpsRequest` Objects for Reconfiguring MariaDB Database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: md-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-mariadb
  configuration:
    applyConfig:
      my-apply.cnf: |-
        [mysqld]
        max_connections = 300
        read_buffer_size = 1234567
    restart: "true"
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

**Sample `MariaDBOpsRequest` Objects for Volume Expansion of MariaDB:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: md-volume-expansion
  namespace: demo
spec:
  type: VolumeExpansion  
  databaseRef:
    name: sample-mariadb
  volumeExpansion:   
    mode: "Online"
    mariadb: 2Gi
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

**Sample `MariaDBOpsRequest` Objects for Reconfiguring TLS of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MariaDBOpsRequest
metadata:
  name: md-recon-tls-add
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-mariadb
  tls:
    requireSSL: true
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: md-issuer
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
kind: MariaDBOpsRequest
metadata:
  name: md-recon-tls-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-mariadb
  tls:
    rotateCertificates: true
```


Here, we are going to describe the various sections of a `MariaDBOpsRequest` crd.

A `MariaDBOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [MariaDB](/docs/guides/mariadb/concepts/mariadb) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [MariaDB](/docs/guides/mariadb/concepts/mariadb) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `MariaDBOpsRequest`.

- `Upgrade` / `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `MariaDBOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `MariaDBOpsRequest`. At first, you have to create a `MariaDBOpsRequest` for updating. Once it is completed, then you can create another `MariaDBOpsRequest` for scaling. You should not create two `MariaDBOpsRequest` simultaneously.

### spec.updateVersion

If you want to update your MariaDB version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [MariaDBVersion](/docs/guides/mariadb/concepts/mariadb-version/index.md) CR that contains the MariaDB version information where you want to update.

> You can only update between MariaDB versions. KubeDB does not support downgrade for MariaDB.

### spec.horizontalScaling

If you want to scale-up or scale-down your MariaDB cluster or different components of it, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:
- `spec.horizontalScaling.member` indicates the desired number of nodes for MariaDB cluster after scaling. For example, if your cluster currently has 4 nodes, and you want to add additional 2 nodes then you have to specify 6 in `spec.horizontalScaling.member` field. Similarly, if you want to remove one node from the cluster, you have to specify 3 in `spec.horizontalScaling.` field.

### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `MariaDB` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-field:

- `spec.verticalScaling.mariadb` indicates the desired resources for MariaDB standalone or cluster after scaling.
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

If you want to expand the volume of your MariaDB standalone or cluster, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.volumeExpansion.mariadb` indicates the desired size for the persistent volume of a MariaDB.
- `spec.volumeExpansion.mode` indicates the mode of volume expansion. It can be `online` or `offline` based on the storage class.


All of them refer to Quantity types of Kubernetes.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    mariadb: "2Gi"
```

This will expand the volume size of all the mariadb nodes to 2 GB.

### spec.configuration

If you want to reconfigure your Running MariaDB cluster with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-fields:
- `configSecret` points to a secret in the same namespace of a MariaDB resource, which contains the new custom configurations. If there are any configSecret set before in the database, this secret will replace it.
- `applyConfig` contains the new custom config as a string which will be merged with the previous configuration.
- `removeCustomConfig` reomoves all the custom configs of the MariaDB server.
- `restart` significantly reduces unnecessary downtime.
  - `auto` (default): restart only if required (determined by ops manager operator)
  - `false`: If user set the restart to false, then pod will not be restarted and new config will not be synced with database.
  - `true`: always restart



### spec.tls

If you want to reconfigure the TLS configuration of your database i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/mariadb/concepts/mariadb/#spectls).
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.
- `spec.tls.remove` specifies that we want to remove tls from this database.


### MariaDBOpsRequest `Status`

`.status` describes the current state and progress of a `MariaDBOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `MariaDBOpsRequest`. It can have the following three values:

| Phase      | Meaning                                                                            |
| ---------- | ---------------------------------------------------------------------------------- |
| Successful | KubeDB has successfully performed the operation requested in the MariaDBOpsRequest |
| Failed     | KubeDB has failed the operation requested in the MariaDBOpsRequest                 |
| Denied     | KubeDB has denied the operation requested in the MariaDBOpsRequest                 |

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `MariaDBOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `MariaDBOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. MariaDBOpsRequest has the following types of conditions:

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
