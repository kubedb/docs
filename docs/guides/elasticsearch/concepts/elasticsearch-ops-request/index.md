---
title: ElasticsearchOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: es-opsrequest-concepts
    name: ElasticsearchOpsRequest
    parent: es-concepts-elasticsearch
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# ElasticsearchOpsRequest

## What is ElasticsearchOpsRequest

`ElasticsearchOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for the [Elasticsearch](https://www.elastic.co/guide/index.html) and [OpenSearch](https://opensearch.org/) administrative operations like database version update, horizontal scaling, vertical scaling, etc. in a Kubernetes native way.

## ElasticsearchOpsRequest Specifications

Like any official Kubernetes resource, a `ElasticsearchOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: es-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: es
  updateVersion:
    targetVersion: searchguard-7.5.2-v1
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
- `UpdateVersion` - is used to update the version of the Elasticsearch in a managed way. The necessary information required for updating the version, must be provided in `spec.updateVersion` field.
- `VerticalScaling` - is used to vertically scale the Elasticsearch nodes (ie. pods). The necessary information required for vertical scaling, must be provided in `spec.verticalScaling` field.
- `HorizontalScaling` - is used to horizontally scale the Elasticsearch nodes (ie. pods). The necessary information required for horizontal scaling, must be provided in `spec.horizontalScaling` field.
- `VolumeExpansion` - is used to expand the storage of the Elasticsearch nodes (ie. pods). The necessary information required for volume expansion, must be provided in `spec.volumeExpansion` field.
- `ReconfigureTLS` - is used to configure the TLS configuration of a running Elasticsearch cluster. The necessary information required for reconfiguring the TLS, must be provided in `spec.tls` field.

> Note: You can only perform one type of operation by using an `ElasticsearchOpsRequest` custom resource object. For example, if you want to update your database and scale up its replica then you will need to create two separate `ElasticsearchOpsRequest`. At first, you will have to create an `ElasticsearchOpsRequest` for updating. Once the update is completed, then you can create another `ElasticsearchOpsRequest` for scaling. You should not create two `ElasticsearchOpsRequest` simultaneously.

### spec.databaseRef

`spec.databaseRef` is a `required` field that points to the [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- `databaseRef.name` - specifies the name of the [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md) object.

> Note: The `ElasticsearchOpsRequest` should be on the same namespace as the referring `Elasticsearch` object.

### spec.updateVersion

`spec.updateVersion` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `UpdateVersion`.
It specifies the desired version information required for the Elasticsearch version update. This field consists of the following sub-fields:

- `updateVersion.targetVersion` refers to an [ElasticsearchVersion](/docs/guides/elasticsearch/concepts/catalog/index.md) CR name that contains the Elasticsearch version information required to perform the update.

> KubeDB does not support downgrade for Elasticsearch.

**Samples:**
Let's assume we have and Elasticsearch cluster of version `xpack-8.2.0`. The Elasticsearch custom resource is named `es-quickstart` and it's provisioned in demo namespace. Now, you want to update your Elasticsearch cluster to `xpack-8.5.2`. Apply this YAML to update to your desired version.
```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: es-quickstart-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: es-quickstart
  updateVersion:
    targetVersion: xpack-8.5.2
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

All of them refer to [Quantity](https://v1-22.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#quantity-resource-core) types of Kubernetes.

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

### spec.tls

> The ReconfigureTLS only works with the [Cert-Manager](https://cert-manager.io/docs/concepts/) managed certificates. [Installation guide](https://cert-manager.io/docs/installation/).

`spec.tls` is an `optional` field, but it acts as a `required` field when the `spec.type` is set to `ReconfigureTLS`. It specifies the necessary information required to add or remove or update the TLS configuration of the Elasticsearch cluster. It consists of the following sub-fields:

- `tls.remove` ( `bool` | `false` ) - tells the operator to remove the TLS configuration for the HTTP layer. The transport layer is always secured with certificates, so the removal process does not affect the transport layer.
- `tls.rotateCertificates` ( `bool` | `false`) - tells the operator to renew all the certificates.
- `tls.issuerRef` - is an `optional` field that references to the `Issuer` or `ClusterIssuer` custom resource object of [cert-manager](https://cert-manager.io/docs/concepts/issuer/). It is used to generate the necessary certificate secrets for Elasticsearch. If the `issuerRef` is not specified, the operator creates a self-signed CA and also creates necessary certificate (valid: 365 days) secrets using that CA. 
  - `apiGroup` - is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
  - `kind` - is the type of resource that is being referenced. The supported values are `Issuer` and `ClusterIssuer`.
  - `name` - is the name of the resource ( `Issuer` or `ClusterIssuer` ) that is being referenced.

- `tls.certificates` - is an `optional` field that specifies a list of certificate configurations used to configure the  certificates. It has the following fields:
  - `alias` - represents the identifier of the certificate. It has the following possible value:
    - `transport` - is used for the transport layer certificate configuration.
    - `http` - is used for the HTTP layer certificate configuration.
    - `admin` - is used for the admin certificate configuration. Available for the `SearchGuard` and the `OpenDistro` auth-plugins.
    - `metrics-exporter` - is used for the metrics-exporter sidecar certificate configuration.
  
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
  kind: ElasticsearchOpsRequest
  metadata:
    name: add-tls
    namespace: demo
  spec:
    type: ReconfigureTLS
    databaseRef:
      name: es
    tls:
      issuerRef:
        apiGroup: "cert-manager.io"
        kind: Issuer
        name: es-issuer
      certificates:
      - alias: http
        subject:
          organizations:
          - kubedb.com
        emailAddresses:
        - abc@kubedb.com 
  ```

- Remove TLS:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: remove-tls
    namespace: demo
  spec:
    type: ReconfigureTLS
    databaseRef:
      name: es
    tls:
      remove: true
  ```

- Rotate TLS:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: rotate-tls
    namespace: demo
  spec:
    type: ReconfigureTLS
    databaseRef:
      name: es
    tls:
      rotateCertificates: true
  ```

- Update transport layer certificate:

  ```yaml
  apiVersion: ops.kubedb.com/v1alpha1
  kind: ElasticsearchOpsRequest
  metadata:
    name: update-tls
    namespace: demo
  spec:
    type: ReconfigureTLS
    databaseRef:
      name: es
    tls:
      certificates:
        - alias: transport
          subject:
            organizations:
              - mydb.com # say, previously it was "kubedb.com"
  ```

### spec.configuration

If you want to reconfigure your Running Elasticsearch cluster or different components of it with new custom configuration, you have to specify `spec.configuration` section. This field consists of the following sub-field:

- `spec.configuration.configsecret`: ConfigSecret is an optional field to provide custom configuration file for database.
- `spec.configuration.secureConfigSecret`: SecureConfigSecret is an optional field to provide secure settings for database.
- `spec.configuration.applyConfig`: ApplyConfig is an optional field to provide Elasticsearch configuration. Provided configuration will be applied to config files stored in ConfigSecret. If the ConfigSecret is missing, the operator will create a new k8s secret by the following naming convention: {db-name}-user-config.
```yaml
  	applyConfig:
  		file-name.yml: |
  			key: value
  		elasticsearch.yml: |
  			thread_pool:
  				write:
  					size: 30
```

- `spec.configuration.removeCustomConfig`: If set to "true", the user provided configuration will be removed. The Elasticsearch cluster will start will default configuration that is generated by the operator.
- `spec.configuration.removeSecureCustomConfig`: If set to "true", the user provided secure settings will be removed. The elasticsearch.keystore will start will default password (i.e. "").

### spec.timeout
As we internally retry the ops request steps multiple times, This `timeout` field helps the users to specify the timeout for those steps of the ops request (in second).
If a step doesn't finish within the specified timeout, the ops request will result in failure.

### spec.apply
This field controls the execution of obsRequest depending on the database state. It has two supported values: `Always` & `IfReady`.
Use IfReady, if you want to process the opsRequest only when the database is Ready. And use Always, if you want to process the execution of opsReq irrespective of the Database state.

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
