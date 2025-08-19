---
title: HazelcastOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: hz-opsrequest-concepts
    name: HazelcastOpsRequest
    parent: hz-concepts-hazelcast
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HazelcastOpsRequest

## What is HazelcastOpsRequest

`HazelcastOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for Hazelcast administrative operations like database version updating, horizontal scaling, vertical scaling, reconfigure TLS, restart, etc. in a Kubernetes native way.

## HazelcastOpsRequest CRD Specifications

Like any official Kubernetes resource, a `HazelcastOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `HazelcastOpsRequest` CRs for different administrative operations is given below.

Sample HazelcastOpsRequest for updating database version:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hzops-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: hz-prod
  updateVersion:
    targetVersion: 5.5.6
```

Sample `HazelcastOpsRequest` for horizontal scaling:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hazelcast-scale-up
  namespace: demo
spec:
  databaseRef:
    name: hz-prod
  type: HorizontalScaling
  horizontalScaling:
    hazelcast: 4
```

Sample `HazelcastOpsRequest` for vertical scaling:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hazelcast-vertical-scaling
  namespace: demo
spec:
  databaseRef:
    name: hz-prod
  type: VerticalScaling
  verticalScaling:
    hazelcast:
      resources:
        limits:
          cpu: 1
          memory: 2.5Gi
        requests:
          cpu: 1
          memory: 2.5Gi
```

Sample `HazelcastOpsRequest` for reconfiguring TLS:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hzops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: hz-prod
  tls:
    issuerRef:
      name: hz-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - hazelcast
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

Sample `HazelcastOpsRequest` for restart:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hazelcast-restart
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: hz-prod
  type: Restart
```
Sample `HazelcastOpsRequest` for reconfigure:
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hz-reconfigure-custom-config
  namespace: demo
spec:
  apply: IfReady
  configuration:
    configSecret:
      name: hazelcast-custom-config
    applyConfig:
      hazelcast.yaml: |-
        hazelcast:
          persistence:
            enabled: true
            validation-timeout-seconds: 2500
            data-load-timeout-seconds: 3000
            auto-remove-stale-data: false
      hazelcast-client.yaml: |-
        hazelcast-client: {}
  databaseRef:
    name: hz-prod
  type: Reconfigure

```

Here, we are going to describe the various sections of a `HazelcastOpsRequest` crd.

## HazelcastOpsRequest `Spec`

A `HazelcastOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `HazelcastOpsRequest`.

- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `HazelcastOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `HazelcastOpsRequest`. At first, you have to create a `HazelcastOpsRequest` for updating. Once it is completed, then you can create another `HazelcastOpsRequest` for scaling.

### spec.updateVersion

If you want to update your Hazelcast version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [HazelcastVersion](/docs/guides/hazelcast/concepts/catalog.md) CR that contains the Hazelcast version information where you want to update.

> You can only update between Hazelcast versions. KubeDB does not support downgrade for Hazelcast.

### spec.horizontalScaling

If you want to scale-up or scale-down your Hazelcast cluster, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.member` indicates the desired number of member nodes for Hazelcast cluster after scaling. For example, if your cluster currently has 3 member nodes and you want to add additional 2 member nodes then you have to specify 5 in `spec.horizontalScaling.member` field. Similarly, if you want to remove 1 node from the cluster, you have to specify 2 in `spec.horizontalScaling.member` field.

### spec.verticalScaling

`spec.verticalScaling` is used to specify the new resources requirements to vertical scale the database. This field consists of the following sub-fields:

- `spec.verticalScaling.member` indicates the Hazelcast member resources. It has the below structure:

```yaml
requests:
  memory: "2Gi"
  cpu: "1"
limits:
  memory: "2Gi"  
  cpu: "1"
```

Here, when you specify the resource request for Hazelcast member, KubeDB will create a new [PetSet](https://github.com/kubeops/petset) and [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) with the new resource requirements and drop the old PetSet and StatefulSet.

### spec.volumeExpansion

To expand the storage of the Hazelcast cluster, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.volumeExpansion.member` indicates the desired size for the persistent volume claim of the member nodes.

All the volumes of member nodes will be expanded when the ops request type is `VolumeExpansion`.

### spec.reconfigure

`spec.reconfigure` specifies the information of the custom configuration. This field consists of the following sub-field:

- `spec.reconfigure.configSecret` points to a secret in the same namespace of a Hazelcast resource, which contains the new custom configurations. If there are any configSecret is already associated with the database, the new custom configuration will be merged and will be applied to the database.

### spec.reconfigureTLS

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing Hazelcast via a HazelcastOpsRequest. This field consists of the following sub-field:

- `spec.reconfigureTLS.issuerRef` specifies the issuer name, api group and kind of the desired issuer. For example,

```yaml
issuerRef:
  apiGroup: cert-manager.io
  kind: Issuer
  name: hz-ca-issuer
```

- `spec.reconfigureTLS.certificates` specifies the certificates. For example,

```yaml
certificates:
- alias: server
  subject:
    organizations:
    - hazelcast
    organizationalUnits:
    - server
- alias: client  
  subject:
    organizations:
    - hazelcast
    organizationalUnits:
    - client
```

- `spec.reconfigureTLS.rotateCertificates` specifies that we want to rotate the certificate of this database. Set it to `true` to rotate certificates.

- `spec.reconfigureTLS.remove` specifies that we want to remove tls of this database. Set it to `true` to remove tls.

### spec.timeout

As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second). If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply

This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use `IfReady` if you want to process the opsRequest only when the database is Ready. And use `Always` if you want to process the execution of opsReq irrespective of the Database state.

## HazelcastOpsRequest `Status`

`.status` describes the current state and progress of the `HazelcastOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `HazelcastOpsRequest`. It can have the following three values:

| Phase      | Meaning                                                                              |
| ---------- | ------------------------------------------------------------------------------------ |
| Successful | KubeDB has successfully performed the operation requested in the HazelcastOpsRequest |
| Failed     | KubeDB has failed to perform the operation requested in the HazelcastOpsRequest     |
| Progressing| KubeDB is performing the operation requested in the HazelcastOpsRequest             |

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `HazelcastOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `HazelcastOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition. HazelcastOpsRequest has the following types of conditions:

| Type                         | Meaning                                                                   |
| ---------------------------- | ------------------------------------------------------------------------- |
| `Progressing`                | Specifies that the operation is now in the progressing state             |
| `Successful`                 | Specifies that the operation phase succeeded                              |
| `Failed`                     | Specifies that the operation phase failed                                |
| `UpdateVersion`              | Specifies that the UpdateVersion operation succeeded                      |
| `HorizontalScaling`          | Specifies that the HorizontalScaling operation succeeded                 |
| `VerticalScaling`            | Specifies that the VerticalScaling operation succeeded                   |
| `VolumeExpansion`            | Specifies that the VolumeExpansion operation succeeded                   |
| `Reconfigure`                | Specifies that the Reconfigure operation succeeded                       |
| `ReconfigureTLS`             | Specifies that the ReconfigureTLS operation succeeded                    |
| `Restart`                    | Specifies that the Restart operation succeeded                           |

- `status` specifies the status of the condition. It can be `True`, `False` or `Unknown`.
- `lastTransitionTime` specifies the last time the condition transitioned from one status to another.
- `reason` specifies the reason for the last transition of the condition.
- `message` provides a human readable message indicating details about the last transition.

## Next Steps

- Learn about [Hazelcast CRD](/docs/guides/hazelcast/concepts/hazelcast.md).
- Deploy your first Hazelcast database with KubeDB by following the guide [here](/docs/guides/hazelcast/quickstart/quickstart.md).
