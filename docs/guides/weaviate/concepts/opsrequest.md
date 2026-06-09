---
title: WeaviateOpsRequest CRD
menu:
  docs_{{ .version }}:
    identifier: weaviate-opsrequest-concepts
    name: WeaviateOpsRequest
    parent: weaviate-concepts-weaviate
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# WeaviateOpsRequest

## What is WeaviateOpsRequest
`WeaviateOpsRequest` is a Kubernetes `CustomResource` that lets you run day-2 operations on a [Weaviate](/docs/guides/weaviate/concepts/weaviate.md) database in a declarative way.

Instead of editing many low-level resources manually, you submit a `WeaviateOpsRequest` and KubeDB Ops Manager performs the requested workflow (for example restart, version update, or resource scaling) and reports progress in the request status.

## Supported operation types
The following operation types are supported for Weaviate:

- `Reconfigure`
- `Restart`
- `RotateAuth`
- `UpdateVersion`
- `VerticalScaling`
- `VolumeExpansion`

Use one operation type per request. If you need multiple changes, apply multiple ops requests sequentially.

## Sample WeaviateOpsRequest
Sample request for `UpdateVersion`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: weaviate-sample
  updateVersion:
    targetVersion: 1.34.0
```

Sample request for `VerticalScaling`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-vertical-scale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: weaviate-sample
  verticalScaling:
    node:
      resources:
        requests:
          cpu: "500m"
          memory: 1Gi
        limits:
          cpu: "1"
          memory: 2Gi
```

## Key fields

- `spec.databaseRef.name` points to the target `Weaviate` object.
- `spec.type` defines the operation category.
- `spec.timeout` (optional) controls operation timeout.
- `spec.apply` (optional) controls when the request is executed.

Operation-specific sections:

- `spec.configuration` for `Reconfigure`
- `spec.updateVersion.targetVersion` for `UpdateVersion`
- `spec.verticalScaling.node.resources` for `VerticalScaling`
- `spec.volumeExpansion` for `VolumeExpansion`
- `spec.authSecret` for user-defined credential rotation in `RotateAuth`

Always verify the exact shape from operation-specific guides because fields vary by operation.

## Status fields

After you apply a request, watch status to track progress:

```bash
kubectl get weaviateopsrequest -n demo
kubectl describe weaviateopsrequest -n demo <opsrequest-name>
```

Important status information:

- `.status.phase` (`Successful`, `Failed`, or `Denied`)
- `.status.conditions` for step-by-step controller progress
- `.status.observedGeneration` for reconciliation state

## Next Steps

- See [Weaviate ops overview](/docs/guides/weaviate/ops-request/overview.md) for operation links.
- Follow operation tutorials like [Restart](/docs/guides/weaviate/restart/restart.md) and [UpdateVersion](/docs/guides/weaviate/update-version/overview.md).