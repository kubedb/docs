---
title: DurableWorkflowOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: dw-opsrequest-concepts
    name: DurableWorkflowOpsRequest
    parent: dw-concepts-durableworkflow
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DurableWorkflowOpsRequest

## What is DurableWorkflowOpsRequest

`DurableWorkflowOpsRequest` is a Kubernetes `Custom Resource Definition` (CRD). It provides a declarative way to run day-2 operations on a [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md) in a Kubernetes native way. The short name is `dwops`.

## DurableWorkflowOpsRequest Spec

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-hscale
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: temporal
  horizontalScaling:
    frontend: 3
    history: 3
    matching: 3
  timeout: 30m
  apply: IfReady
```

### spec.databaseRef

The name of the `DurableWorkflow`, which must be in the namespace of the request.

### spec.type

| Type                | Guide                                                                              | Fields                |
|---------------------|------------------------------------------------------------------------------------|-----------------------|
| `Restart`           | [Restart](/docs/guides/durableworkflow/restart/restart.md)                          | `restart`             |
| `UpdateVersion`     | [Update Version](/docs/guides/durableworkflow/update-version/overview.md)           | `updateVersion`       |
| `HorizontalScaling` | [Scaling](/docs/guides/durableworkflow/scaling/overview.md)                         | `horizontalScaling`   |
| `VerticalScaling`   | [Scaling](/docs/guides/durableworkflow/scaling/overview.md)                         | `verticalScaling`     |
| `Reconfigure`       | [Reconfigure](/docs/guides/durableworkflow/reconfigure/overview.md)                 | `configuration`       |
| `ReconfigureTLS`    | [Reconfigure TLS](/docs/guides/durableworkflow/reconfigure_tls/overview.md)         | `tls`                 |
| `RotateAuth`        | [Rotate Authentication](/docs/guides/durableworkflow/rotateauth/overview.md)        | `authentication`      |

Volume expansion and storage migration do not exist for this kind, because the servers have no volumes.

### spec.restart

`roles` lists the roles to restart, from `frontend`, `history`, `matching`, `worker` and `internalFrontend`. When omitted, every role is restarted.

### spec.updateVersion

- `targetVersion`: the `DurableWorkflowVersion` to move to.
- `allowDowntime`: needed for a hop whose schema update cannot run online, and acknowledges a cluster that is not highly available.
- `surge`: a map from a role to `true` or `false`. By default every role except `history` gets one extra pod while it is rolled.

### spec.horizontalScaling and spec.verticalScaling

`horizontalScaling` takes the new replicas of `frontend`, `history`, `matching`, `worker` and `internalFrontend` (Temporal only), each at least 1. `verticalScaling` takes the same role names, each with `resources`, `nodeSelectionPolicy` and `topology`.

### spec.configuration

`configSecret`, `applyConfig` (a map from `server.yaml` or `dynamic_config.yaml` to its content), `removeCustomConfig`, and `restart` (`auto`, `true` or `false`). `numHistoryShards` cannot be changed.

### spec.tls

`issuerRef`, `certificates`, `rotateCertificates` or `remove`.

### spec.authentication

`secretRef` is an optional Secret with a `password` key. Without it, a password is generated.

### spec.timeout, spec.apply and spec.maxRetries

- `timeout` is the time that a single step may take. A replaced pod gets a third of it to become ready.
- `apply` is `IfReady` (the default) or `Always`.
- `maxRetries` defaults to `1`.

## DurableWorkflowOpsRequest Status

`status.phase` is `Pending`, `Progressing`, `Successful` or `Failed`, and `status.conditions` holds one condition per step. The spec of a request cannot be changed after it is created.

```bash
$ kubectl get dwops -n demo
NAME        TYPE                STATUS       AGE
dw-hscale   HorizontalScaling   Successful   6m
```

While a request runs, the `DurableWorkflow` is paused for the provisioner. Only one request should run at a time.

## Next Steps

- Learn about the [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md) CRD.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
