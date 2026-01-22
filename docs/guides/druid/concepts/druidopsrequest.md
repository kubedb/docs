---
title: DruidOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: guides-druid-concepts-druidopsrequest
    name: DruidOpsRequest
    parent: guides-druid-concepts
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# DruidOpsRequest

## What is DruidOpsRequest

`DruidOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [Druid](https://druid.apache.org/) administrative operations like database version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## DruidOpsRequest CRD Specifications

Like any official Kubernetes resource, a `DruidOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `DruidOpsRequest` CRs for different administrative operations is given below:

**Sample `DruidOpsRequest` for updating database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: druid-prod
  updateVersion:
    targetVersion: 30.0.1
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `DruidOpsRequest` Objects for Horizontal Scaling of different component of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-hscale-down
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: druid-prod
  horizontalScaling:
    topology: 
      coordinators: 2
      historicals: 2
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `DruidOpsRequest` Objects for Vertical Scaling of different component of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: druid-prod
  verticalScaling:
    coordinators:
      resources:
        requests:
          memory: "1.5Gi"
          cpu: "0.7"
        limits:
          memory: "2Gi"
          cpu: "1"
    historicals:
      resources:
        requests:
          memory: "1.5Gi"
          cpu: "0.7"
        limits:
          memory: "2Gi"
          cpu: "1"
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `DruidOpsRequest` Objects for Reconfiguring different druid mode:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-reconfiugre
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: druid-prod
  configuration:
    applyConfig:
      middleManager.properties: |
        druid.worker.capacity=5
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
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
kind: DruidOpsRequest
metadata:
  name: drops-reconfiugre
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: druid-prod
  configuration:
    configSecret:
      name: new-configsecret
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
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
kind: DruidOpsRequest
metadata:
  name: drops-reconfiugre
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: druid-prod
  configuration:
    restart: "true"
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `DruidOpsRequest` Objects for Volume Expansion of different database components:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-volume-exp
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: druid-prod
  volumeExpansion:
    mode: "Online"
    historicals: 2Gi
    middleManagers: 2Gi
status:
  conditions:
    - lastTransitionTime: "2024-07-25T18:22:38Z"
      message: Successfully completed the modification process
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

**Sample `DruidOpsRequest` Objects for Reconfiguring TLS of the database:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: druid-prod
  tls:
    issuerRef:
      name: dr-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        emailAddresses:
          - abc@appscode.com
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: druid-dev
  tls:
    rotateCertificates: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: druid-prod
  tls:
    issuerRef:
      name: dr-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DruidOpsRequest
metadata:
  name: drops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: druid-prod
  tls:
    remove: true
```

Here, we are going to describe the various sections of a `DruidOpsRequest` crd.

A `DruidOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [Druid](/docs/guides/druid/concepts/druid.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Druid](/docs/guides/druid/concepts/druid.md) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `DruidOpsRequest`.

- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`

> You can perform only one type of operation on a single `DruidOpsRequest` CR. For example, if you want to update your database and scale up its replica then you have to create two separate `DruidOpsRequest`. At first, you have to create a `DruidOpsRequest` for updating. Once it is completed, then you can create another `DruidOpsRequest` for scaling.

### spec.updateVersion

If you want to update you Druid version, you have to specify the `spec.updateVersion` section that specifies the desired version information. This field consists of the following sub-field:

- `spec.updateVersion.targetVersion` refers to a [DruidVersion](/docs/guides/druid/concepts/druidversion.md) CR that contains the Druid version information where you want to update.

> You can only update between Druid versions. KubeDB does not support downgrade for Druid.

### spec.horizontalScaling

If you want to scale-up or scale-down your Druid cluster or different components of it, you have to specify `spec.horizontalScaling` section. This field consists of the following sub-field:

- `spec.horizontalScaling.topology` indicates the configuration of topology nodes for Druid topology cluster after scaling. This field consists of the following sub-field:
  - `spec.horizontalScaling.topoloy.coordinators` indicates the desired number of coordinators nodes for Druid topology cluster after scaling.
  - `spec.horizontalScaling.topology.overlords` indicates the desired number of overlords nodes for Druid topology cluster after scaling.
  - `spec.horizontalScaling.topology.brokers` indicates the desired number of brokers nodes for Druid topology cluster after scaling.
  - `spec.horizontalScaling.topology.routers` indicates the desired number of routers nodes for Druid topology cluster after scaling.
  - `spec.horizontalScaling.topology.historicals` indicates the desired number of historicals nodes for Druid topology cluster after scaling.
  - `spec.horizontalScaling.topology.middleManagers` indicates the desired number of middleManagers nodes for Druid topology cluster after scaling.

### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `Druid` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-fields:
- `spec.verticalScaling.coordinators` indicates the desired resources for coordinators of Druid topology cluster after scaling.
- `spec.verticalScaling.overlords` indicates the desired resources for overlords of Druid topology cluster after scaling.
- `spec.verticalScaling.brokers` indicates the desired resources for brokers of Druid topology cluster after scaling.
- `spec.verticalScaling.routers` indicates the desired resources for routers of Druid topology cluster after scaling.
- `spec.verticalScaling.historicals` indicates the desired resources for historicals of Druid topology cluster after scaling.
- `spec.verticalScaling.middleManagers` indicates the desired resources for middleManagers of Druid topology cluster after scaling.

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

If you want to expand the volume of your Druid cluster or different components of it, you have to specify `spec.volumeExpansion` section. This field consists of the following sub-field:

- `spec.mode` specifies the volume expansion mode. Supported values are `Online` & `Offline`. The default is `Online`.
- `spec.volumeExpansion.historicals` indicates the desired size for the persistent volume for historicals of a Druid topology cluster.
- `spec.volumeExpansion.middleManagers` indicates the desired size for the persistent volume for middleManagers of a Druid topology cluster.

> It is only possible to expand the data servers ie. `historicals` and `middleManagers` as they only comes with persistent volumes.

All of them refer to [Quantity](https://v1-22.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#quantity-resource-core) types of Kubernetes.

Example usage of this field is given below:

```yaml
spec:
  volumeExpansion:
    node: "2Gi"
```

This will expand the volume size of all the combined nodes to 2 GB.

### spec.configuration

If you want to reconfigure your Running Druid cluster or different components of it with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-field:

- `spec.configuration.configSecret` points to a secret in the same namespace of a Druid resource, which contains the new custom configurations. If there are any configSecret set before in the database, this secret will replace it. The value of the field `spec.stringData` of the secret like below:
```yaml
common.runtime.properties: |
  druid.storage.archiveBucket="my-druid-archive-bucket"
middleManagers.properties: |
  druid.worker.capacity=5
```
> Similarly, it is possible to provide configs for `coordinators`, `overlords`, `brokers`, `routers` and `historicals` through `coordinators.properties`, `overlords.properties`, `brokers.properties`, `routers.properties` and `historicals.properties` respectively.

- `applyConfig` contains the new custom config as a string which will be merged with the previous configuration.

- `applyConfig` is a map where key supports 3 values, namely `server.properties`, `broker.properties`, `controller.properties`. And value represents the corresponding configurations.

```yaml
  applyConfig:
    common.runtime.properties: |
      druid.storage.archiveBucket="my-druid-archive-bucket"
    middleManagers.properties: |
      druid.worker.capacity=5
```

- `removeCustomConfig` is a boolean field. Specify this field to true if you want to remove all the custom configuration from the deployed druid cluster.

- `restart` significantly reduces unnecessary downtime.
  - `auto` (default): restart only if required (determined by ops manager operator)
  - `false`: no restart
  - `true`: always restart



### spec.tls

If you want to reconfigure the TLS configuration of your Druid i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates, you have to specify `spec.tls` section. This field consists of the following sub-field:

- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/druid/concepts/druid.md#spectls).
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this druid.
- `spec.tls.remove` specifies that we want to remove tls from this druid.

### spec.timeout
As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second).
If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.

### DruidOpsRequest `Status`

`.status` describes the current state and progress of a `DruidOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `DruidOpsRequest`. It can have the following three values:

| Phase       | Meaning                                                                          |
|-------------|----------------------------------------------------------------------------------|
| Successful  | KubeDB has successfully performed the operation requested in the DruidOpsRequest |
| Progressing | KubeDB has started the execution of the applied DruidOpsRequest                  |
| Failed      | KubeDB has failed the operation requested in the DruidOpsRequest                 |
| Denied      | KubeDB has denied the operation requested in the DruidOpsRequest                 |
| Skipped     | KubeDB has skipped the operation requested in the DruidOpsRequest                |

Important: Ops-manager Operator can skip an opsRequest, only if an
