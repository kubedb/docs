---
title: Resetting a Branch Root Password
menu:
  docs_{{ .version }}:
    identifier: branching-reset-root-password
    name: ResetRootPassword
    parent: branching
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Resetting a Branch Root Password

Set `spec.resetRootPassword: true` when the branch should have its own root credential. Without this option, the branch starts with the source credential copied into a separate target Secret.

## Execution order

Courier coordinates credential isolation after the target has been provisioned and before running post-actions or reporting the Branch `Ready`. The database-specific implementation determines whether credentials are generated during provisioning or changed through a database operation.

A database engine that rotates credentials through an OpsRequest requires the corresponding KubeDB Ops-manager support on the target cluster. PostgreSQL generates the branch credential during provisioning; follow its [password-reset guide](/docs/guides/postgres/branch/reset-root-password.md) for the exact behavior and verification commands.

## Using the branch credential

Read connection credentials from the target database's auth Secret after the Branch reaches `Ready`. Configure applications to use that credential with the target Service. The source database retains its credential.

Post-actions receive the branch's current connection environment, so they run after credential isolation without embedding passwords in the Branch manifest.

## Scheduled branches

A refresh replaces cloned data and can change the target credential state. Wait for the refreshed Branch to return to `Ready`, then read the target auth Secret again. Applications must use the current branch credential rather than assuming a password remains unchanged across refreshes.

## Verification

Verify authentication through the target database Service using both the branch password and the source password. The branch password should succeed and the source password should be rejected. A local database socket can use a different authentication policy and does not establish password isolation.

The [PostgreSQL example](/docs/guides/postgres/branch/reset-root-password.md) demonstrates this check without displaying decoded passwords.
