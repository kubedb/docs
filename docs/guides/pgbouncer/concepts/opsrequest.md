---
title: PgBouncerOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: pb-opsrequest-concepts
    name: PgBouncerOpsRequest
    parent: pb-concepts-pgbouncer
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PgBouncerOpsRequest

## What is PgBouncerOpsRequest

`PgBouncerOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [PgBouncer](https://pgbouncer.net/mediawiki/index.php/Main_Page) administrative operations like version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## PgBouncerOpsRequest CRD Specifications

Like any official Kubernetes resource, a `PgBouncerOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `PgBouncerOpsRequest` CRs for different administrative operations is given below:

**Sample `PgBouncerOpsRequest` for updating version:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pgbouncer-version-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: pgbouncer-server
  updateVersion:
    targetVersion: 1.18.0
```

**Sample `PgBouncerOpsRequest` Objects for Horizontal Scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pgbouncer-horizontal-scale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: pgbouncer-server
  horizontalScaling:
    replicas: 2
```

**Sample `PgBouncerOpsRequest` Objects for Vertical Scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pgbouncer-vertical-scale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: pgbouncer-server
  verticalScaling:
    pgbouncer:
      resources:
        requests:
          memory: "1200Mi"
          cpu: "0.7"
        limits:
          memory: "1200Mi"
          cpu: "0.7"
```

**Sample `PgBouncerOpsRequest` Objects for Reconfiguring:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pgbouncer-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: pgbouncer-server
  configuration:
    pgbouncer:
      applyConfig:
        pgbouncer.conf: |-
          auth_type = scram-sh-256
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pgbouncer-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: pgbouncer-server
  configuration:
    removeCustomConfig: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: pgbouncer-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: pgbouncer-server
  configuration:
    configSecret:
      name: new-custom-config
```


**Sample `PgBouncerOpsRequest` Objects for Reconfiguring TLS:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pgbouncer-server
  tls:
    sslMode: verify-ca
    clientAuthMode: cert
    issuerRef:
      name: pgbouncer-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - kubedb
          organizationalUnits:
            - client
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pgbouncer-server
  tls:
    rotateCertificates: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PgBouncerOpsRequest
metadata:
  name: tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: pgbouncer-server
  tls:
    remove: true
```

Here, we are going to describe the various sections of a `PgBouncerOpsRequest` crd.

A `PgBouncerOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `PgBouncerOpsRequest`.

- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `PgBouncerOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `PgBouncerOpsRequest`. At first, you have to create a `PgBouncerOpsRequest` for updating. Once it is completed, then you can create another `PgBouncerOpsRequest` for scaling. 

> Note: There is an exception to the above statement. It is possible to specify both `spec.configuration` & `spec.verticalScaling` in a OpsRequest of type `VerticalScaling`.

### spec.updateVersion

If you want to update your PgBouncer version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [PgBouncerVersion](/docs/guides/pgbouncer/concepts/catalog.md) CR that contains the PgBouncer version information where you want to update.


### spec.horizontalScaling

If you want to scale-up or scale-down your PgBouncer cluster or different components of it, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.replicas` indicates the desired number of pods for PgBouncer cluster after scaling. For example, if your cluster currently has 4 pods, and you want to add additional 2 pods then you have to specify 6 in `spec.horizontalScaling.replicas` field. Similarly, if you want to remove one pod from the cluster, you have to specify 3 in `spec.horizontalScaling.replicas` field.

### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of resources like `cpu`, `memory` etc. that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.pgbouncer` indicates the desired resources for PetSet of PgBouncer after scaling.
- `spec.verticalScaling.exporter` indicates the desired resources for PetSet of PgBouncer Exporter after scaling.

It has the below structure:

```yaml
requests:
  memory: "200Mi"
  cpu: "0.1"
limits:
  memory: "300Mi"
  cpu: "0.2"
```

Here, when you specify the resource request, the scheduler uses this information to decide which node to place the container of the Pod on and when you specify a resource limit for the container, the `kubelet` enforces those limits so that the running container is not allowed to use more of that resource than the limit you set. You can found more details from [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).


### spec.configuration

If you want to reconfigure your Running PgBouncer cluster or different components of it with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-field:

- `configSecret` points to a secret in the same namespace of a PgBouncer resource, which contains the new custom configurations. If there are any configSecret set before in the database, this secret will replace it.
- `applyConfig` contains the new custom config as a string which will be merged with the previous configuration. 

- `applyConfig` is a map where key supports 1 values, namely `pgbouncer.ini`.

```yaml
  applyConfig:
    pgbouncer.conf: |-
      max_pool = 30  
```

- `removeCustomConfig` is a boolean field. Specify this field to true if you want to remove all the custom configuration from the deployed pgbouncer server.

### spec.tls

If you want to reconfigure the TLS configuration of your pgbouncer cluster i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/pgbouncer/concepts/pgbouncer.md#spectls).
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.
- `spec.tls.remove` specifies that we want to remove tls from this database.
- `spec.tls.sslMode` specifies what will be the ssl mode of the cluster allowed values are: disable,allow,prefer,require,verify-ca,verify-full
- `spec.tls.clientAuthMode` specifies what will be the client authentication mode of the cluster allowed values are: md5,scram,cert

### spec.timeout
As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second). 
If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of opsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.


### PgBouncerOpsRequest `Status`

`.status` describes the current state and progress of a `PgBouncerOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `PgBouncerOpsRequest`. It can have the following three values:

| Phase       | Meaning                                                                            |
|-------------|------------------------------------------------------------------------------------|
| Successful  | KubeDB has successfully performed the operation requested in the PgBouncerOpsRequest |
| Progressing | KubeDB has started the execution of the applied PgBouncerOpsRequest                  |
| Failed      | KubeDB has failed the operation requested in the PgBouncerOpsRequest                 |
| Denied      | KubeDB has denied the operation requested in the PgBouncerOpsRequest                 |
| Skipped     | KubeDB has skipped the operation requested in the PgBouncerOpsRequest                |

Important: Ops-manager Operator can skip an opsRequest, only if its execution has not been started yet & there is a newer opsRequest applied in the cluster. `spec.type` has to be same as the skipped one, in this case.

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `PgBouncerOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `PgBouncerOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. PgBouncerOpsRequest has the following types of conditions:

| Type                           | Meaning                                                                   |
|--------------------------------|---------------------------------------------------------------------------|
| `Progressing`                  | Specifies that the operation is now in the progressing state              |
| `Successful`                   | Specifies such a state that the operation on the database was successful. |
| `DatabasePauseSucceeded`       | Specifies such a state that the database is paused by the operator        |
| `ResumeDatabase`               | Specifies such a state that the database is resumed by the operator       |
| `Failed`                       | Specifies such a state that the operation on the database failed.         |
| `UpdatePetSetResources`        | Specifies such a state that the PetSet resources has been updated         |
| `UpdatePetSet`                 | Specifies such a state that the PetSet  has been updated                  |
| `IssueCertificatesSucceeded`   | Specifies such a state that the tls certificate issuing is successful     |
| `UpdateDatabase`               | Specifies such a state that the CR of PgBouncer is updated                   |

- The `status` field is a string, with possible values `True`, `False`, and `Unknown`.
  - `status` will be `True` if the current transition succeeded.
  - `status` will be `False` if the current transition failed.
  - `status` will be `Unknown` if the current transition was denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
