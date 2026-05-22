---
title: Neo4jOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: neo4j-opsrequest-concepts
    name: Neo4jOpsRequest
    parent: neo4j-concepts-neo4j
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4jOpsRequest

## What is Neo4jOpsRequest

`Neo4jOpsRequest` is the CRD for day-2 operational workflows for KubeDB-managed Neo4j databases.

## Neo4jOpsRequest CRD Specifications

Like any Kubernetes resource, a `Neo4jOpsRequest` contains `TypeMeta`, `ObjectMeta`, `Spec`, and `Status` sections. The `spec.type` field determines which operation KubeDB should perform on the target Neo4j database.

## Supported operation types

- `Reconfigure`
- `ReconfigureTLS`
- `Restart`
- `RotateAuth`
- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `StorageMigration`

## Sample Neo4jOpsRequest manifests

**Sample `Neo4jOpsRequest` for updating database version:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: neo4j-test
  updateVersion:
    targetVersion: 2025.12.1
```

**Sample `Neo4jOpsRequest` for vertical scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: neo4j-test
  verticalScaling:
    server:
      resources:
        limits:
          cpu: 1500m
          memory: 4Gi
        requests:
          cpu: 700m
          memory: 4Gi
```

**Sample `Neo4jOpsRequest` objects for volume expansion:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-volumeexpansion
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: neo4j
  volumeExpansion:
    mode: "Offline"
    server: 4Gi
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-volumeexpansiononline
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: neo4j
  volumeExpansion:
    mode: "Online"
    server: 6Gi
```

**Sample `Neo4jOpsRequest` for horizontal scaling:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neoops-hscale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: neo4j-test
  horizontalScaling:
    server: 5
    reallocate:
      strategy: "incremental"
      batchSize: 1
```

**Sample `Neo4jOpsRequest` objects for rotating auth credentials:**

Rotate authentication without a user-provided Secret:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: neo4j-test
  timeout: 5m
  apply: IfReady
```

Rotate authentication using a user-provided Secret:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neoops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: neo4j
  authentication:
    secretRef:
      kind: Secret
      name: external-neo4j-auth
  timeout: 5m
  apply: IfReady
```

**Sample `Neo4jOpsRequest` objects for reconfiguring Neo4j:**

Reconfigure using a new custom configuration Secret and inline overrides:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: neo4j
  configuration:
    configSecret:
      name: new-custom-config
    removeCustomConfig: true
    applyConfig:
      server.metrics.csv.interval: "40s"
  timeout: 5m
  apply: IfReady
```

Reconfigure using inline `applyConfig` values:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: reconfigure-apply
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: neo4j
  configuration:
    configSecret:
      name: new-custom-config
    applyConfig:
      server.metrics.enabled: "false"
  timeout: 5m
  apply: IfReady
```

**Sample `Neo4jOpsRequest` for restart:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: neo4j
  timeout: 5m
  apply: Always
```

**Sample `Neo4jOpsRequest` objects for TLS reconfiguration:**

Rotate TLS certificates and update Bolt mode:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: neo4j-tls
  tls:
    rotateCertificates: true
    bolt:
      mode: mTLS
```

Remove TLS from the database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: neo4j-tls
  tls:
    remove: true
```

Add or replace TLS using a cert-manager issuer:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: neo4j-tls
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: neo4j-ca-issuer
```

**Sample `Neo4jOpsRequest` for storage class migration:**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: neo4j-test
  migration:
    storageClassName: custom-longhorn
    oldPVReclaimPolicy: Delete
  timeout: 3000s
```

## Key fields

- `spec.databaseRef.name` identifies the target `Neo4j` object.
- `spec.type` selects the operation category. Valid values are `Restart`, `ReconfigureTLS`, `RotateAuth`, `Reconfigure`, `HorizontalScaling`, `VerticalScaling`, `VolumeExpansion`, `StorageMigration`, and `UpdateVersion`.
- Set the operation-specific section that matches `spec.type`, such as `spec.updateVersion`, `spec.verticalScaling`, `spec.volumeExpansion`, `spec.horizontalScaling`, `spec.migration`, `spec.authentication`, `spec.configuration`, or `spec.tls`.
- `spec.updateVersion.targetVersion` selects the target `Neo4jVersion` for an `UpdateVersion` request.
- `spec.verticalScaling.server.resources` defines the new CPU and memory requests and limits for Neo4j server Pods.
- `spec.volumeExpansion.mode` chooses whether storage expansion runs in `Online` or `Offline` mode, and `spec.volumeExpansion.server` sets the new PVC size for the server volume.
- `spec.migration.storageClassName` selects the destination StorageClass for `StorageMigration`, and `spec.migration.oldPVReclaimPolicy` controls old PV reclaim behavior (`Delete` or `Retain`).
- `spec.horizontalScaling.server` sets the desired number of Neo4j servers. `spec.horizontalScaling.reallocate.strategy` controls post-scaling reallocation, and `spec.horizontalScaling.reallocate.batchSize` is used with the `incremental` strategy.
- `spec.authentication.secretRef` optionally points to a user-managed Secret for `RotateAuth`. If it is omitted, KubeDB can rotate credentials using an operator-managed Secret.
- `spec.configuration.configSecret.name` points to a Secret containing new custom configuration, `spec.configuration.removeCustomConfig` removes the existing custom config, and `spec.configuration.applyConfig` applies inline configuration changes.
- `spec.tls.rotateCertificates`, `spec.tls.remove`, and `spec.tls.issuerRef` control TLS certificate rotation, removal, and issuer-based TLS configuration. Protocol-specific settings such as `spec.tls.bolt.mode` can also be updated there.
- `spec.timeout` sets the timeout for each step of the operation.
- `spec.apply` controls when KubeDB should execute the OpsRequest, and `spec.maxRetries` controls how many times the operator retries a failed step.

### Neo4jOpsRequest `Status`

`.status` describes the current state and progress of the `Neo4jOpsRequest` operation. It has the following fields:

#### status.phase

`status.phase` indicates the overall phase of the operation for this `Neo4jOpsRequest`.

| Phase        | Meaning                                                                           |
|--------------|-----------------------------------------------------------------------------------|
| `Progressing` | KubeDB has started processing the requested operation                            |
| `Successful` | KubeDB has successfully completed the requested operation                         |
| `Failed`     | KubeDB has failed to complete the requested operation                             |
| `Denied`     | KubeDB has denied the requested operation                                         |

#### status.observedGeneration

`status.observedGeneration` shows the most recent generation observed by the `Neo4jOpsRequest` controller.

#### status.conditions

`status.conditions` is an array that tracks step-by-step state transitions during `Neo4jOpsRequest` processing. Each condition entry includes:

- `type`: category of the condition transition.
- `status`: one of `True`, `False`, or `Unknown`.
- `reason`: machine-readable reason for the latest transition.
- `message`: human-readable details for the transition.
- `lastTransitionTime`: timestamp for the latest state transition.
- `observedGeneration`: generation observed for that condition update.

Common `type` values:

| Type                  | Meaning                                                                 |
|-----------------------|-------------------------------------------------------------------------|
| `UpdateVersion`       | Version update step has completed                                       |
| `VerticalScaling`     | Vertical scaling step has completed                                     |
| `HorizontalScaling`   | Horizontal scaling step has completed                                   |
| `VolumeExpansion`     | Volume expansion step has completed                                     |
| `StorageMigration`    | Storage class migration step has completed                              |
| `Reconfigure`         | Reconfiguration step has completed                                      |
| `ReconfigureTLS`      | TLS reconfiguration step has completed                                  |
| `Restart`             | Restart step has completed                                              |
| `RotateAuth`          | Auth rotation step has completed                                        |

Common `reason` values:

| Reason                                  | Meaning                                                                 |
|-----------------------------------------|-------------------------------------------------------------------------|
| `OpsRequestProgressingStarted`          | Operator has started processing the OpsRequest                          |
| `OpsRequestFailedToProgressing`         | Operator failed to start processing                                     |
| `OpsRequestProcessedSuccessfully`       | Operator has completed the requested operation                          |
| `DatabaseVersionUpdatingStarted`        | Version update has started                                              |
| `SuccessfullyUpdatedDatabaseVersion`    | Version update has completed successfully                               |
| `FailedToUpdateDatabaseVersion`         | Version update has failed                                               |
| `VerticalScalingStarted`                | Vertical scaling has started                                            |
| `SuccessfullyPerformedVerticalScaling`  | Vertical scaling has completed successfully                             |
| `FailedToPerformVerticalScaling`        | Vertical scaling has failed                                             |
| `HorizontalScalingStarted`              | Horizontal scaling has started                                          |
| `SuccessfullyPerformedHorizontalScaling`| Horizontal scaling has completed successfully                           |
| `FailedToPerformHorizontalScaling`      | Horizontal scaling has failed                                           |
| `StorageMigrationStarted`               | Storage migration has started                                           |
| `SuccessfullyPerformedStorageMigration` | Storage migration has completed successfully                            |
| `FailedToPerformStorageMigration`       | Storage migration has failed                                            |

## Next Steps

- See [Neo4j ops overview](/docs/guides/neo4j/concepts/opsrequest.md) for operation links.
- Read the [Reconfigure guide](/docs/guides/neo4j/reconfigure/overview.md) for configuration changes.
- Read the [Reconfigure TLS guide](/docs/guides/neo4j/reconfigure-tls/overview.md) for certificate rotation, removal, or issuer updates.
- Read the [Restart guide](/docs/guides/neo4j/restart/restart.md), [Rotate Auth guide](/docs/guides/neo4j/rotate-auth/overview.md), and [Update Version guide](/docs/guides/neo4j/update-version/overview.md).
- Read the [Horizontal Scaling guide](/docs/guides/neo4j/scaling/horizontal-scaling/overview.md), [Vertical Scaling guide](/docs/guides/neo4j/scaling/vertical-scaling/overview.md), and [Volume Expansion guide](/docs/guides/neo4j/volume-expansion/overview.md).
- Read the [Storage Migration guide](/docs/guides/neo4j/migration/storageMigration.md) for persistent volume and storage class migration.
