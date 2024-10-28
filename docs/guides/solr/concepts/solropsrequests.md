---
title: AppBinding CRD
menu:
  docs_{{ .version }}:
    identifier: sl-solropsrequest-solr
    name: AppBinding
    parent: sl-concepts-solr
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SolrOpsRequest

## What is SolrOpsRequest

`SolrOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for the [Solr](https://solr.apache.org/guide/solr/latest/index.html) administrative operations like database version update, horizontal scaling, vertical scaling, etc. in a Kubernetes native way.

## SolrOpsRequest Specifications

Like any official Kubernetes resource, a `SolrOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: upgrade-solr
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: solr-cluster
  type: UpdateVersion
  updateVersion:
    targetVersion: 9.6.1
status:
  conditions:
    - lastTransitionTime: "2024-10-25T06:40:49Z"
      message: Successfully updated Solr version
      observedGeneration: 1
      reason: Successful
      status: "True"
      type: Successful
  observedGeneration: 1
  phase: Successful
```

Here, we are going to describe the various sections of a `SolrOpsRequest` CRD.

### spec.type

`spec.type` is a `required` field that specifies the kind of operation that will be applied to the Solr. The following types of operations are allowed in the `SolrOpsRequest`:

- `Restart` - is used to perform a smart restart of the Solr cluster.
- `UpdateVersion` - is used to update the version of the Solr in a managed way. The necessary information required for updating the version, must be provided in `spec.updateVersion` field.
- `VerticalScaling` - is used to vertically scale the Solr nodes (ie. pods). The necessary information required for vertical scaling, must be provided in `spec.verticalScaling` field.
- `HorizontalScaling` - is used to horizontally scale the Solr nodes (ie. pods). The necessary information required for horizontal scaling, must be provided in `spec.horizontalScaling` field.
- `VolumeExpansion` - is used to expand the storage of the Solr nodes (ie. pods). The necessary information required for volume expansion, must be provided in `spec.volumeExpansion` field.
- `ReconfigureTLS` - is used to configure the TLS configuration of a running Solr cluster. The necessary information required for reconfiguring the TLS, must be provided in `spec.tls` field.

> Note: You can only perform one type of operation by using an `SolrOpsRequest` custom resource object. For example, if you want to update your database and scale up its replica then you will need to create two separate `SolrOpsRequest`. At first, you will have to create an `SolrOpsRequest` for updating. Once the update is completed, then you can create another `SolrOpsRequest` for scaling. You should not create two `SolrOpsRequest` simultaneously.

### spec.databaseRef

`spec.databaseRef` is a `required` field that points to the [Solr](/docs/guides/solr/concepts/solr.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- `databaseRef.name` - specifies the name of the [Solr](/docs/guides/solr/concepts/solr.md) object.

> Note: The `SolrOpsRequest` should be on the same namespace as the referring `Solr` object.

### spec.updateVersion

`spec.updateVersion` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `UpdateVersion`.
It specifies the desired version information required for the Solr version update. This field consists of the following sub-fields:

- `updateVersion.targetVersion` refers to an [SolrVersion](/docs/guides/solr/concepts/solrversion.md) CR name that contains the Solr version information required to perform the update.

> KubeDB does not support downgrade for Solr.

**Samples:**
Let's assume we have and Solr cluster of version `9.4.1`. The Solr custom resource is named `solr-cluster` and it's provisioned in demo namespace. Now, you want to update your Solr cluster to `9.6.1`. Apply this YAML to update to your desired version.
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: upgrade-solr
  namespace: demo
spec:
  databaseRef:
    name: solr-cluster
  type: UpdateVersion
  updateVersion:
    targetVersion: 9.6.1
```

### spec.horizontalScaling

`spec.horizontalScaling` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `HorizontalScaling`.
It specifies the necessary information required to horizontally scale the Solr nodes (ie. pods). It consists of the following sub-field:

- `horizontalScaling.node` - specifies the desired number of nodes for the Solr cluster running in combined mode (ie. `Solr.spec.topology` is `empty`).  The value should be greater than the maximum value of replication for the shard of any index. For example, if a shard has `x` replicas, `x+1` data nodes are required to allocate them.
- `horizontalScaling.overseer` - specifies the desired number of overseer nodes. The value should be greater than zero ( >= 1 ).
- `horizontalScaling.data` - specifies the desired number of data nodes. The value should be greater than zero ( >= 1 ).
- `horizontalScaling.coordinator` - specifies the desired number of coordinator nodes. ( >= 1)

**Samples:**

- Horizontally scale combined nodes:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: hscale-solr-combined
  namespace: demo
spec:
  databaseRef:
    name: solr-combined
  type: HorizontalScaling
  horizontalScaling:
    node: 2
 ```

- Horizontally scale cluster topology:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: hscale-solr-topology
  namespace: demo
spec:
  databaseRef:
    name: solr-cluster
  type: HorizontalScaling
  horizontalScaling:
    coordinator: 2
    data: 2
    overseer: 2
```

- Horizontally scale only data nodes:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: SolrOpsRequest
  metadata:
    name: hscale-data-nodes
    namespace: demo
  spec:
    type: HorizontalScaling
    databaseRef:
      name: solr-cluster
    horizontalScaling:
      topology:
        data: 4
  ```

### spec.verticalScaling

`spec.verticalScaling` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `VerticalScaling`. It specifies the necessary information required to vertically scale the Solr node resources (ie. `cpu`, `memory`). It consists of the following sub-field:

- `verticalScaling.node` - specifies the desired node resources for the Solr cluster running in combined mode (ie. `Solr.spec.topology` is `empty`).
- `verticalScaling.overseer` - specifies the desired resources for the overseer nodes. It takes input same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).
- `verticalScaling.data` - specifies the desired node resources for the data nodes. It takes input  same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).
- `verticalScaling.coordinator` - specifies the desired node resources for the coordinator nodes. It takes input  same as the k8s [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-types).

> Note: It is recommended not to use resources below the default one; `cpu: 900m, memory: 2Gi`.

**Samples:**

- Vertically scale combined nodes:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: vertical-scale-combined
  namespace: demo
spec:
  databaseRef:
    name: solr-combined
  type: VerticalScaling
  verticalScaling:
    node:
      resources:
        limits:
          cpu: 1
          memory: 2.5Gi
        requests:
          cpu: 1
          memory: 2.5Gi
```

- For topology cluster

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: vertical-scale-topology
  namespace: demo
spec:
  databaseRef:
    name: solr-cluster
  type: VerticalScaling
  verticalScaling:
    data:
      resources:
        limits:
          cpu: 1
          memory: 2.5Gi
        requests:
          cpu: 1
          memory: 2.5Gi
    overseer:
      resources:
        limits:
          cpu: 1
          memory: 2.5Gi
        requests:
          cpu: 1
          memory: 2.5Gi
    coordinator:
      resources:
        limits:
          cpu: 1
          memory: 2.5Gi
        requests:
          cpu: 1
          memory: 2.5Gi
```

- Vertically scale only data nodes:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: vertical-scale-topology
  namespace: demo
spec:
  databaseRef:
    name: solr-cluster
  type: VerticalScaling
  verticalScaling:
    data:
      resources:
        limits:
          cpu: 1
          memory: 2.5Gi
        requests:
          cpu: 1
          memory: 2.5Gi
```

### spec.volumeExpansion

> Note: To use the volume expansion feature the StorageClass must support volume expansion.

`spec.volumeExpansion` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `VolumeExpansion`. It specifies the necessary information required to expand the storage of the Solr node. It consists of the following sub-field:

- `volumeExpansion.node` - specifies the desired size of the persistent volume for the Solr node running in combined mode (ie. `Solr.spec.topology` is `empty`).
- `volumeExpansion.overseer` - specifies the desired size of the persistent volume for the overseer nodes.
- `volumeExpansion.data` - specifies the desired size of the persistent volume for the data nodes.
- `volumeExpansion.coordinator` - specifies the desired size of the persistent volume for the ingest nodes.

All of them refer to [Quantity](https://v1-22.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#quantity-resource-core) types of Kubernetes.

> Note: Make sure that the requested volume is greater than the current volume.

**Samples:**

- Expand volume for combined nodes:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: volume-expansion-topology
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: solr-cluster
  type: VolumeExpansion
  volumeExpansion:
    mode: Offline
    node: 4Gi
  ```

- Expand volume for cluster topology:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: volume-expansion-topology
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: solr-cluster
  type: VolumeExpansion
  volumeExpansion:
    mode: Offline
    data: 4Gi
    overseer : 4Gi
    coordinator: 4Gi
  ```

- Expand volume for only data nodes:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: volume-expansion-topology
  namespace: demo
spec:
  apply: IfReady
  databaseRef:
    name: solr-cluster
  type: VolumeExpansion
  volumeExpansion:
    mode: Offline
    data: 4Gi
  ```

### spec.tls

> The ReconfigureTLS only works with the [Cert-Manager](https://cert-manager.io/docs/concepts/) managed certificates. [Installation guide](https://cert-manager.io/docs/installation/).

`spec.tls` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `ReconfigureTLS`. It specifies the necessary information required to add or remove or update the TLS configuration of the Solr cluster. It consists of the following sub-fields:

- `tls.remove` ( `bool` | `false` ) - tells the operator to remove the TLS configuration for the HTTP layer. The transport layer is always secured with certificates, so the removal process does not affect the transport layer.
- `tls.rotateCertificates` ( `bool` | `false`) - tells the operator to renew all the certificates.
- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for Solr. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA.
    - `apiGroup` - is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
    - `kind` - is the type of resource that is being referenced. The supported values are `Issuer` and `ClusterIssuer`.
    - `name` - is the name of the resource ( `Issuer` or `ClusterIssuer` ) that is being referenced.

- `tls.certificates` - is an `optional` field that specifies a list of certificate configurations used to configure the  certificates. It has the following fields:
    - `alias` - represents the identifier of the certificate. It has the following possible value:
        - `server` - is used for the server certificate configuration.
        - `client` - is used for the client certificate configuration.

    - `secretName` - ( `string` | `"<database-name>-alias-cert"` ) - specifies the k8s secret name that holds the certificates.

    - `subject` - specifies an `X.509` distinguished name (DN). It has the following configurable fields:
        - `organizations` ( `[]string` | `nil` ) - is a list of organization names.
        - `organizationalUnits` ( `[]string` | `nil` ) - is a list of organization unit names.
        - `countries` ( `[]string` | `nil` ) -  is a list of country names (ie. Country Codes).
        - `localities` ( `[]string` | `nil` ) - is a list of locality names.
        - `provinces` ( `[]string` | `nil` ) - is a list of province names.
        - `streetAddresses` ( `[]string` | `nil` ) - is a list of street addresses.
        - `postalCodes` ( `[]string` | `nil` ) - is a list of postal codes.
        - `serialNumber` ( `string` | `""` ) is a serial number.

      For more details, visit [here](https://golang.org/pkg/crypto/x509/pkix/#Name).

    - `duration` ( `string` | `""` ) - is the period during which the certificate is valid. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as `"300m"`, `"1.5h"` or `"20h45m"`. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    - `renewBefore` ( `string` | `""` ) - is a specifiable time before expiration duration.
    - `dnsNames` ( `[]string` | `nil` ) - is a list of subject alt names.
    - `ipAddresses` ( `[]string` | `nil` ) - is a list of IP addresses.
    - `uris` ( `[]string` | `nil` ) - is a list of URI Subject Alternative Names.
    - `emailAddresses` ( `[]string` | `nil` ) - is a list of email Subject Alternative Names.

To enable TLS on the HTTP layer, the configuration for the `http` layer certificate needs to be provided on `tls.certificates[]` list.

**Samples:**

- Add TLS:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: add-tls
  namespace: demo
spec:
  apply: IfReady
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: self-signed-issuer
      kind: ClusterIssuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  databaseRef:
    name: solr-cluster
  type: ReconfigureTLS
  ```

- Remove TLS:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: SolrOpsRequest
  metadata:
    name: remove-tls
    namespace: demo
  spec:
    type: ReconfigureTLS
    databaseRef:
      name: solr-cluster
    tls:
      remove: true
  ```

- Rotate TLS:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: SolrOpsRequest
  metadata:
    name: rotate-tls
    namespace: demo
  spec:
    type: ReconfigureTLS
    databaseRef:
      name: solr-cluster
    tls:
      rotateCertificates: true
  ```

### spec.configuration

If you want to reconfigure your Running Solr cluster or different components of it with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-field:

- `spec.configuration.configsecret`: ConfigSecret is an optional field to provide custom configuration file for database.
- `spec.configuration.applyConfig`: ApplyConfig is an optional field to provide Solr configuration. Provided configuration will be applied to config files stored in ConfigSecret. If the ConfigSecret is missing, the operator will create a new k8s secret by the following naming convention: {db-name}-user-config.
```yaml
  	applyConfig:
      solr.xml: |
        <backup>
          <repository name="kubedb-s3" class="org.apache.solr.s3.S3BackupRepository">
            <str name="s3.bucket.name">solrbackup</str>
            <str name="s3.region">us-east-1</str>
            <str name="s3.endpoint">http://s3proxy-s3.demo.svc:80</str>
          </repository>
        </backup>
```

- `spec.configuration.removeCustomConfig`: If set to "true", the user provided configuration will be removed. The Solr cluster will start will default configuration that is generated by the operator.

### spec.timeout
As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second).
If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.

## SolrOpsRequest `Status`

`.status` describes the current state and progress of a `SolrOpsRequest` operation. It has the following fields:

### status.phase

`status.phase` indicates the overall phase of the operation for this `SolrOpsRequest`. It can have the following three values:

| Phase       | Meaning                                                                             |
| :--------:  | ----------------------------------------------------------------------------------  |
| Progressing | KubeDB has started to process the Ops request                                       |
| Successful  | KubeDB has successfully performed all the operations needed for the Ops request     |
| Failed      | KubeDB has failed while performing the operations needed for the Ops request        |

### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `SolrOpsRequest` controller.

### status.conditions

`status.conditions` is an array that specifies the conditions of different steps of `SolrOpsRequest` processing. Each condition entry has the following fields:

- `types` specifies the type of the condition.
- The `status` field is a string, with possible values `True`, `False`, and `Unknown`.
    - `status` will be `True` if the current transition succeeded.
    - `status` will be `False` if the current transition failed.
    - `status` will be `Unknown` if the current transition was denied.
- The `message` field is a human-readable message indicating details about the condition.
- The `reason` field is a unique, one-word, CamelCase reason for the condition's last transition.
- The `lastTransitionTime` field provides a timestamp for when the operation last transitioned from one state to another.
- The `observedGeneration` shows the most recent condition transition generation observed by the controller.

SolrOpsRequest has the following types of conditions:

| Type                        | Meaning                                             |
|-----------------------------|-----------------------------------------------------|
| `Progressing`               | The operator has started to process the Ops request |
| `Successful`                | The Ops request has successfully executed           |
| `Failed`                    | The operation on the database failed                |
| `OrphanPetSetPods`          | The petSet has deleted leaving the pods orphaned    |
| `ReadyPetSets`              | The PetSet are ready                                |
| `ScaleDownCombinedNode`     | Scaled down the combined nodes                      |
| `ScaleDownDataNode`         | Scaled down the data nodes                          |
| `ScaleDownCoordinatorNode`  | Scaled down the coordinator nodes                   |
| `ScaleDownOverseerNode`     | Scaled down the overseer nodes                      |
| `ScaleUpCombinedNode`       | Scaled up the combined nodes                        |
| `ScaleUpDataNode`           | Scaled up the data nodes                            |
| `ScaleUpCoordinatorNode`    | Scaled up the coordinator nodes                     |
| `ScaleUpOverseerNode`       | Scaled up the overseer nodes                        |
| `UpdateCombinedNodePVCs`    | Updated combined node PVCs                          |
| `UpdateDataNodePVCs`        | Updated data node PVCs                              |
| `UpdateCoordinatorNodePVCs` | Updated coordinator node PVCs                       |
| `UpdateOverseerNodePVCs`    | Updated overseer node PVCs                          |
| `UpdateNodeResources`       | Updated node resources                              |
