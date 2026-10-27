---
title: Rotating YugabyteDB Authentication
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-rotate-auth-overview
    name: Overview
    parent: yugabytedb-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotating YugabyteDB Authentication

A `YugabyteDBOpsRequest` with `spec.type: RotateAuth` changes the password of the `yugabyte` user. If `spec.authentication.secretRef` is set, KubeDB uses the password in that Secret; otherwise it generates a random one.

KubeDB Ops-manager:

1. Stages the new credential next to the old one.
2. Runs `ALTER ROLE yugabyte PASSWORD ...` over a SQL connection (condition `UpdateCredential`). It tries the new password first, so a retry after a crash does not alter the role twice.
3. Verifies that the new password is accepted (condition `VerifyCredential`), then updates `spec.authSecret` and the Secret reference of the AppBinding.

No pod restarts. The username must stay `yugabyte`. With `spec.ycql.enabled`, the YCQL runner is not available in this release, so RotateAuth fails early on such a database.

## Next Step

[Rotate YugabyteDB authentication](/docs/guides/yugabytedb/rotate-auth/rotateauth.md).
