---
title: Milvus Recommendation Engine
menu:
  docs_{{ .version }}:
    identifier: milvus-recommendation-guide
    name: Guide
    parent: milvus-recommendation
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus Recommendation Engine

The KubeDB Recommendation Engine watches your databases and proactively generates `Recommendation` objects describing maintenance operations you should perform — each carrying a ready-to-apply `MilvusOpsRequest`. For Milvus, recommendations are generated for:

- **RotateAuth** — when the auth credential is older than `spec.authSecret.rotateAfter`.
- **ReconfigureTLS** — when a TLS certificate is approaching expiry.
- **UpdateVersion** — when a newer, non-deprecated `MilvusVersion` is available.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)

- The **Recommendation Engine** and **Supervisor** CRDs must be installed (they ship with the KubeDB Supervisor component):

  ```bash
  $ kubectl get crd recommendations.supervisor.appscode.com
  NAME                                      CREATED AT
  recommendations.supervisor.appscode.com   2026-06-30T05:18:01Z
  ```

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO, creates the `my-release-minio` secret, and installs the etcd operator required by Milvus.

> Milvus has no `autoOps` field, so there is no `autoOps.disabled` toggle — recommendations are always generated for the conditions above when their triggers are met.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/recommendation/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/recommendation/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Milvus with Recommendation-Triggering Settings

The sample manifests deliberately use version **`2.6.9`** (so an `UpdateVersion` recommendation can point to a newer catalog version such as `2.6.11`), a short `authSecret.rotateAfter` of `15m` (so a `RotateAuth` recommendation is generated), and short-lived certificates (so a `ReconfigureTLS` recommendation is generated as expiry approaches).

`standalone.yaml`

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-standalone
  namespace: demo
spec:
  version: "2.6.9"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: "my-release-minio"
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    storageClassName: local-path
    resources:
      requests:
        storage: 1Gi
  authSecret:
    kind: Secret
    name: milvus-standalone-auth
    rotateAfter: 15m
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: milvus-issuer
    certificates:
      - alias: server
        duration: 61m
      - alias: client
        duration: 61m
```

The distributed sample (`distributed.yaml`) is equivalent, targeting `milvus-cluster` with `streamingnode` storage.

- `spec.authSecret.rotateAfter: 15m` — the credential is considered due for rotation 15 minutes after `activeFrom`.
- `spec.tls.certificates[].duration: 61m` — short certificate lifetimes so renewal/rotation recommendations surface quickly.
- `spec.version: "2.6.9"` — an older catalog version, so an update to `2.6.11` is recommended.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/recommendation/yamls/standalone.yaml
milvus.kubedb.com/milvus-standalone created
```

Wait for the database to become `Ready`.

## Observe the Generated Recommendations

After the database is ready (and `rotateAfter` / certificate-renewal thresholds are reached), `Recommendation` objects appear in the namespace:

```bash
$ kubectl get recommendation -n demo
NAME                                                 STATUS    OUTDATED   AGE
milvus-standalone-x-milvus-x-update-version-3rq4py   Pending   false      2m
```

Each `Recommendation` is a `supervisor.appscode.com` object that describes one operation and embeds the exact `MilvusOpsRequest` that resolves it. Here is the `UpdateVersion` recommendation generated because the running database is on `2.6.9` while `2.6.11` is available:

```bash
$ kubectl get recommendation milvus-standalone-x-milvus-x-update-version-3rq4py -n demo -o yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  annotations:
    kubedb.com/recommendation-for-version: 2.6.9
  labels:
    app.kubernetes.io/instance: milvus-standalone
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: version-update
  name: milvus-standalone-x-milvus-x-update-version-3rq4py
  namespace: demo
spec:
  description: Latest patch version is available. Recommending version Update from 2.6.9 to 2.6.11.
  operation:                       # <-- the ready-to-apply MilvusOpsRequest
    apiVersion: ops.kubedb.com/v1alpha1
    kind: MilvusOpsRequest
    metadata:
      name: update-version
      namespace: demo
    spec:
      databaseRef:
        name: milvus-standalone
      type: UpdateVersion
      updateVersion:
        targetVersion: 2.6.11
  recommender:
    name: kubedb-ops-manager
  requireExplicitApproval: true
  target:
    apiGroup: kubedb.com
    kind: Milvus
    name: milvus-standalone
status:
  approvalStatus: Pending
  phase: Pending
```

Key fields:

- **`spec.description`** explains *why* the recommendation was generated.
- **`spec.operation`** is the upcoming `MilvusOpsRequest` — applying it performs the recommended operation.
- **`spec.target`** identifies the `Milvus` the recommendation is for.
- **`spec.requireExplicitApproval`** means the operation is only applied after approval (manually, or by the Supervisor within a maintenance window).
- **`status.phase`** tracks the recommendation lifecycle (`Pending` → ... ; it becomes `Skipped`/`Outdated` if the underlying condition no longer holds — for example after you update the version yourself).

Similarly, once the auth credential exceeds `spec.authSecret.rotateAfter` (15m), a `RotateAuth` recommendation is generated, and as the short-lived certificates approach expiry a `ReconfigureTLS` recommendation appears — each embeds the corresponding `MilvusOpsRequest` (`type: RotateAuth` / `type: ReconfigureTLS`).

## Applying a Recommendation

Each `Recommendation` embeds the exact `MilvusOpsRequest` that resolves it (`spec.operation`). The Supervisor can apply it automatically within a maintenance window, or you can extract and apply it manually:

```bash
$ kubectl get recommendation -n demo <name> -o jsonpath='{.spec.operation}' | kubectl apply -f -
```

## Cleaning up

```bash
$ kubectl delete recommendation -n demo --all
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone
$ kubectl delete ns demo
```

## Next Steps

- Learn about [rotate auth](/docs/guides/milvus/rotate-auth/guide.md), [reconfigure TLS](/docs/guides/milvus/reconfigure-tls/guide.md) and [update version](/docs/guides/milvus/update-version/guide.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
