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

## Key fields

- `spec.databaseRef.name` identifies the target `Neo4j` object.
- `spec.type` selects the operation category. Valid values are `Restart`, `ReconfigureTLS`, `RotateAuth`, `Reconfigure`, `HorizontalScaling`, `VerticalScaling`, `VolumeExpansion`, and `UpdateVersion`.
- Set the operation-specific section that matches `spec.type`, such as `spec.updateVersion`, `spec.verticalScaling`, `spec.volumeExpansion`, `spec.horizontalScaling`, `spec.authentication`, `spec.configuration`, or `spec.tls`.
- `spec.updateVersion.targetVersion` selects the target `Neo4jVersion` for an `UpdateVersion` request.
- `spec.verticalScaling.server.resources` defines the new CPU and memory requests and limits for Neo4j server Pods.
- `spec.volumeExpansion.mode` chooses whether storage expansion runs in `Online` or `Offline` mode, and `spec.volumeExpansion.server` sets the new PVC size for the server volume.
- `spec.horizontalScaling.server` sets the desired number of Neo4j servers. `spec.horizontalScaling.reallocate.strategy` controls post-scaling reallocation, and `spec.horizontalScaling.reallocate.batchSize` is used with the `incremental` strategy.
- `spec.authentication.secretRef` optionally points to a user-managed Secret for `RotateAuth`. If it is omitted, KubeDB can rotate credentials using an operator-managed Secret.
- `spec.configuration.configSecret.name` points to a Secret containing new custom configuration, `spec.configuration.removeCustomConfig` removes the existing custom config, and `spec.configuration.applyConfig` applies inline configuration changes.
- `spec.tls.rotateCertificates`, `spec.tls.remove`, and `spec.tls.issuerRef` control TLS certificate rotation, removal, and issuer-based TLS configuration. Protocol-specific settings such as `spec.tls.bolt.mode` can also be updated there.
- `spec.timeout` sets the timeout for each step of the operation.
- `spec.apply` controls when KubeDB should execute the OpsRequest, and `spec.maxRetries` controls how many times the operator retries a failed step.

## Next Steps

- See [Neo4j ops overview](/docs/guides/neo4j/ops-request/overview.md) for operation links.
- Read the [Reconfigure guide](/docs/guides/neo4j/reconfigure/overview.md) for configuration changes.
- Read the [Reconfigure TLS guide](/docs/guides/neo4j/reconfigure-tls/overview.md) for certificate rotation, removal, or issuer updates.
- Read the [Restart guide](/docs/guides/neo4j/restart/restart.md), [Rotate Auth guide](/docs/guides/neo4j/rotate-auth/overview.md), and [Update Version guide](/docs/guides/neo4j/update-version/overview.md).
- Read the [Horizontal Scaling guide](/docs/guides/neo4j/scaling/horizontal-scaling/overview.md), [Vertical Scaling guide](/docs/guides/neo4j/scaling/vertical-scaling/overview.md), and [Volume Expansion guide](/docs/guides/neo4j/volume-expansion/overview.md).
