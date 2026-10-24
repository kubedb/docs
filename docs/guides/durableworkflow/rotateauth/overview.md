---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: dw-rotate-auth-overview
    name: Overview
    parent: dw-rotateauth-durableworkflow
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication Overview

A `RotateAuth` ops request changes the password of the database users that the servers connect with, without a failed connection. Which credentials are these? The ones of the stores in `spec.persistence`: the user `dw_<name>` and, for a separate visibility store, `dw_<name>_visibility`. The credentials of a database user that you created yourself (`createDatabase: false`) are yours to rotate, and a request fails with `no store user is managed by the operator` when there is nothing to rotate.

A user has one password at a time, and the servers read it when they start. Changing the password in place would break every pod that still holds the old one. The ops-manager therefore swaps users:

1. It creates the **next user** next to the current one, with the same access and the new password. The names alternate between `dw_<name>` and `dw_<name>_b`.
2. It updates the Secret of the store and the static configuration, and [rolls the pods one at a time](/docs/guides/durableworkflow/update-version/overview.md#how-the-pods-are-rolled). The pods that have not been replaced yet still work, because the old user exists.
3. When no pod uses it anymore, it **drops the old user**.

The new password is generated, or taken from the `password` key of the Secret in `spec.authentication.secretRef`.

On a highly available cluster this has no downtime. It was verified with 0 failed calls for Temporal on Postgres. It has **not** been verified for Cadence, and not for MySQL or Cassandra.

The Elasticsearch and OpenSearch visibility store uses the account of its `AppBinding`, which the operator does not own, so it is not rotated by this request.

The username changes after a rotation. Anything else that connects to the same database with `dw_<name>` must read the new Secret.

## Next Steps

- Run a [Rotate Authentication](/docs/guides/durableworkflow/rotateauth/durableworkflow.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
