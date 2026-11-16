---
title: WorkflowNamespace CRD
menu:
  docs_{{ .version }}:
    identifier: dw-workflownamespace-concepts
    name: WorkflowNamespace
    parent: dw-concepts-durableworkflow
    weight: 35
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# WorkflowNamespace

## What is WorkflowNamespace

`WorkflowNamespace` (short name `wfns`) is a Kubernetes `Custom Resource Definition` that describes a **Temporal namespace** or a **Cadence domain** of a [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md). Workflows run inside a namespace (a domain, in Cadence), and a new cluster has none that you can use. The KubeDB operator creates it in the engine, keeps its retention, description, owner and archival in sync, and deletes it according to `spec.deletionPolicy`.

The same kind covers both engines. The `DurableWorkflowVersion` of the referenced `DurableWorkflow` decides whether the operator creates a namespace or a domain.

## WorkflowNamespace Spec

```yaml
apiVersion: kubedb.com/v1alpha2
kind: WorkflowNamespace
metadata:
  name: orders
  namespace: demo
spec:
  databaseRef:
    name: temporal
  description: Order processing
  retention: 168h
  deletionPolicy: Delete
```

### spec.databaseRef

Required. The name of the `DurableWorkflow`, in the same Kubernetes namespace. It is immutable. The webhook rejects an object whose `DurableWorkflow` does not exist. The operator waits (`Pending`) until the `DurableWorkflow` is `Ready`.

### spec.name

The name in the engine. It defaults to `metadata.name` and is immutable.

### spec.description and spec.ownerEmail

The description, and the contact of the owner. Cadence requires an owner for a domain, so set `ownerEmail` for Cadence. Temporal stores it as given. Both can be changed but not cleared.

### spec.retention

How long the closed workflow executions are kept. The default is `72h`. It must be greater than zero.

Cadence keeps whole days, with a minimum of one day. The webhook rejects less than `24h`, and the operator **rounds up** to the next day: `25h` becomes 2 days and `49h` becomes 3 days. `status.retention` shows the retention that is in force.

### spec.archival

`archival.enabled: true` archives the histories and the visibility records of the closed workflows of this namespace to the URIs in `spec.archival` of the `DurableWorkflow`. The webhook rejects it when the `DurableWorkflow` has no archival configured. The engines fix the URI once it is set. See [Archival](/docs/guides/durableworkflow/configuration/archival.md).

### spec.tokenSecretRef

The name of a Secret in the same namespace. Its key `token` must hold a JWT that the operator sends to the frontend. It is needed only when the `DurableWorkflow` has [authorization](/docs/guides/durableworkflow/authorization/authorization.md) enabled, and then the operator reports an error without it. The token must carry the `admin` permission on Temporal (for example `default:admin`) and the admin claim on Cadence.

### spec.deletionPolicy

| Value            | When the object is deleted                                                                       |
|------------------|--------------------------------------------------------------------------------------------------|
| `Delete` (default) | The namespace or domain is deleted in the engine, **with all its workflows and histories**.   |
| `DoNotTerminate` | The webhook refuses the deletion.                                                                |

For Temporal, the deletion runs in the background of the engine. For Cadence, the domain is deprecated first and then deleted. A domain with running workflows refuses, which keeps the object.

{{< notice type="warning" message="Cadence 1.4.2 has no DeleteDomain. Deleting the object leaves the domain in the DEPRECATED state in the engine and emits a Warning event (DeprecatedOnly). The object itself is released." >}}

## How the operator connects

The operator calls the gRPC API of the frontend Service (`<name>.<namespace>.svc`, on the gRPC port of the version). With TLS enabled it trusts `ca.crt` of the frontend certificate Secret and presents the `client` certificate. Global namespaces are not supported.

A drift made by hand in the engine, for example another retention set with the CLI, is corrected within five minutes.

## WorkflowNamespace Status

| Field                          | Description                                                   |
|--------------------------------|---------------------------------------------------------------|
| `status.phase`                 | `Pending`, `Ready`, `Failed` or `Terminating`.                |
| `status.name`                  | The name used in the engine.                                  |
| `status.id`                    | The id the engine reported.                                   |
| `status.retention`             | The retention in force in the engine.                         |
| `status.conditions`            | The `Ready` condition, with a reason such as `Synced`.        |

```bash
$ kubectl get wfns -n demo
NAME     DATABASE   NAME     RETENTION   STATUS   AGE
orders   temporal   orders   168h        Ready    1m
```

Events on the object: `Registered`, `Updated`, `Deleted`, `Kept` (a normal event for a policy that leaves the namespace), and the warnings `SyncFailed`, `DeleteFailed` and `DeprecatedOnly`.

## Next Steps

- Create one with the [WorkflowNamespace guide](/docs/guides/durableworkflow/namespaces/workflownamespace.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
