---
title: FerretDBOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: fr-opsrequest-concepts
    name: FerretDBOpsRequest
    parent: fr-concepts-ferretdb
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# FerretDBOpsRequest

## What is FerretDBOpsRequest

`FerretDBOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for `FerretDB` administrative operations like version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## FerretDBOpsRequest CRD Specifications

Like any official Kubernetes resource, a `FerretDBOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `FerretDBOpsRequest` CRs for different administrative operations is given below:

**Sample `FerretDBOpsRequest` for updating version:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: ferretdb-version-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: ferretdb
  updateVersion:
    targetVersion: 1.23.0
```

**Sample `FerretDBOpsRequest` Objects for Horizontal Scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: ferretdb-horizontal-scale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: ferretdb
  horizontalScaling:
    primary:
      replicas: 3
    secondary:
      replicas: 3
```

**Sample `FerretDBOpsRequest` Objects for Vertical Scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: ferretdb-vertical-scale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: ferretdb
  verticalScaling:
    primary:
      resources:
        requests:
          memory: "1200Mi"
          cpu: "0.7"
        limits:
          memory: "1200Mi"
          cpu: "0.7"
```

**Sample `FerretDBOpsRequest` Objects for Reconfiguring TLS:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ferretdb
  tls:
    sslMode: requireSSL
    issuerRef:
      name: ferretdb-ca-issuer
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
kind: FerretDBOpsRequest
metadata:
  name: tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ferretdb
  tls:
    rotateCertificates: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ferretdb
  tls:
    remove: true
```

Here, we are going to describe the various sections of a `FerretDBOpsRequest` crd.

A `FerretDBOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [FerretDB](/docs/guides/ferretdb/concepts/ferretdb.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [FerretDB](/docs/guides/ferretdb/concepts/ferretdb.md) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `FerretDBOpsRequest`.

- `Upgrade` / `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `FerretDBOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `FerretDBOpsRequest`. At first, you have to create a `FerretDBOpsRequest` for updating. Once it is completed, then you can create another `FerretDBOpsRequest` for scaling.

### spec.updateVersion

If you want to update your FerretDB version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [FerretDBVersion](/docs/guides/ferretdb/concepts/catalog.md) CR that contains the FerretDB version information where you want to update.


### spec.horizontalScaling

If you want to scale-up or scale-down your FerretDB cluster or different components of it, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.primary.replicas` indicates the desired number of primary pods for FerretDB cluster after scaling. For example, if your cluster currently has 4 pods, and you want to add additional 2 pods then you have to specify 6 in `spec.horizontalScaling.node` field. Similarly, if you want to remove one pod from the cluster, you have to specify 3 in `spec.horizontalScaling.node` field.
- `spec.horizontalScaling.secondary.replicas` indicates the desired number of secondary pods for FerretDB cluster after scaling. For example, if your cluster currently has 4 pods, and you want to add additional 2 pods then you have to specify 6 in `spec.horizontalScaling.node` field. Similarly, if you want to remove one pod from the cluster, you have to specify 3 in `spec.horizontalScaling.node` field.

### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `FerretDB` resources like `cpu`, `memory` etc. that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.primary` indicates the desired resources for PetSet of FerretDB primary server after scaling.
- `spec.verticalScaling.secondary` indicates the desired resources for PetSet of FerretDB secondary server after scaling.

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

### spec.tls

If you want to reconfigure the TLS configuration of your ferretdb cluster i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/ferretdb/concepts/ferretdb.md#spectls).
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


### FerretDBOpsRequest `Status`

`.status` describes the current state and progress of a `FerretDBOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `FerretDBOpsRequest`. It can have the following three values:

| Phase       | Meaning                                                                            |
|-------------|------------------------------------------------------------------------------------|
| Successful  | KubeDB has successfully performed the operation requested in the FerretDBOpsRequest |
| Progressing | KubeDB has started the execution of the applied FerretDBOpsRequest                  |
| Failed      | KubeDB has failed the operation requested in the FerretDBOpsRequest                 |
| Denied      | KubeDB has denied the operation requested in the FerretDBOpsRequest                 |
| Skipped     | KubeDB has skipped the operation requested in the FerretDBOpsRequest                |

Important: Ops-manager Operator can skip an opsRequest, only if its execution has not been started yet & there is a newer opsRequest applied in the cluster. `spec.type` has to be same as the skipped one, in this case.

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `FerretDBOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `FerretDBOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. FerretDBOpsRequest has the following types of conditions:

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
| `UpdateDatabase`               | Specifies such a state that the CR of FerretDB is updated                   |

- The `status` field is a string, with possible values `True`, `False`, and `Unknown`.
    - `status` will be `True` if the current transition succeeded.
    - `status` will be `False` if the current transition failed.
    - `status` will be `Unknown` if the current transition was denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.
