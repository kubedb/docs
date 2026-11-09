---
title: Running Branch Post-Actions
menu:
  docs_{{ .version }}:
    identifier: branching-post-actions
    name: Post-Actions
    parent: branching
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Branch Post-Actions

Post-actions prepare cloned data for the people and applications that will use a branch. Define each action in `spec.postActions` with a container image. For example, you can:

- Mask names, email addresses, and other personal information before giving developers or AI agents access.
- Drop tables containing sensitive information that the branch does not need.
- Delete confidential rows or data outside the scope of a test.
- Insert test fixtures or application-specific seed data.

These changes run against the branch database and leave the source data unchanged.

## Execution and readiness

Courier runs the actions **in list order**, creating a Kubernetes Job for each action in the target namespace. Actions start after the target database is ready and requested credential isolation is complete. The Branch stays in `ActionsRunning` until all actions succeed, then becomes `Ready`.

A scheduled refresh reruns the full action list against the replacement data. Make scripts safe for Job retries and repeated execution. The target database Service can exist before the actions finish; applications consuming prepared data should wait for the **Branch** to become `Ready`.

For cross-cluster branching, Jobs and their Secrets live on the **target** cluster. See [PostgreSQL post-actions](/docs/guides/postgres/branch/post-actions.md) for a complete SQL anonymization example.

## Connection environment

Courier injects the database-specific connection environment using the target Service and auth Secret. PostgreSQL actions receive `PGHOST`, `PGPORT`, `PGUSER`, and `PGPASSWORD`. Keep credentials out of image layers, command arguments, and application logs.

## Job configuration

| Field | Purpose |
|---|---|
| `image` | Action image with the required database client or application tooling. |
| `jobDefaults.imagePullPolicy` | Image pull policy. |
| `jobDefaults.backoffLimit` | Job retry limit. |
| `jobDefaults.activeDeadlineSeconds` | Overall Job execution deadline. |
| `jobDefaults.ttlSecondsAfterFinished` | Retention period after a Job finishes. |
| `jobTemplate.metadata` | Action pod labels and annotations. |
| `jobTemplate.spec.args` | Arguments passed to the image entrypoint. |
| `jobTemplate.spec.env` | Additional environment variables. |
| `jobTemplate.spec.resources` | Action container resource requests and limits. |
| `jobTemplate.spec.serviceAccountName` | ServiceAccount in the target namespace. |
| `jobTemplate.spec.imagePullSecrets` | Image pull Secrets available in the target namespace. |
| `jobTemplate.spec.nodeSelector`, `jobTemplate.spec.affinity`, `jobTemplate.spec.tolerations` | Action pod placement. |

`jobTemplate` uses KubeDB's pod-template shape. Set action container options directly under `jobTemplate.spec`; it does not take a Kubernetes `containers` list. Choose a retention period long enough for Courier to observe completion and for operators to inspect logs.
