---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-rotate-auth-overview
    name: Overview
    parent: guides-oracle-rotate-authentication
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Oracle

**Rotate Authentication** is a feature of the KubeDB Ops-manager operator that allows you to rotate the authentication credentials (the database password) of an `Oracle` database without manual intervention. This is useful for security compliance and credential hygiene.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)

## How Rotate Oracle Authentication Configuration Process Works

The authentication credentials of an `Oracle` database are stored in a Kubernetes `Secret` (by default `<db-name>-auth`) containing the `username` and `password` keys. By default, the privileged user is **`SYS`** (connected as `SYSDBA`).

There are two ways to rotate the authentication of an Oracle database:

1. **Operator generated credentials:** When you create an `OracleOpsRequest` of type `RotateAuth` without referencing any user provided secret, the KubeDB Ops-manager operator generates a new random password, applies it to the database with `ALTER USER <user> IDENTIFIED BY "<new-password>"`, and updates the auth secret. The previous credentials are preserved under the `.prev` (and the upcoming under `.next`) keys of the auth secret, so an application that still holds the old password has a grace window to migrate.

2. **User defined credentials:** You can supply your own credentials by creating a `Secret` of type `kubernetes.io/basic-auth` and referencing it through `spec.authentication.secretRef.name` in the `OracleOpsRequest`. The operator applies the password from that secret to the database.

> **Note:** Oracle does **not** allow renaming the `SYS` user. Therefore, the rotate authentication operation rotates the **password** only; the `username` remains `sys`.

The high level steps the Ops-manager operator performs during a `RotateAuth` operation are:

1. Update the credential (generate a new password or read the user provided secret).
2. Update the related `PetSet`s so the new secret is mounted into the pods.
3. Restart the database pods (one at a time) so they pick up the new credential.
4. Mark the `OracleOpsRequest` as `Successful`.

In the next section, we will walk you through a step-by-step guide on rotating authentication of an Oracle database using `OracleOpsRequest`.

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

> ## ⚠️ Legal Notice
>
> Oracle® and Oracle Database® are registered trademarks of Oracle Corporation.
> KubeDB is not affiliated with, endorsed by, or sponsored by Oracle Corporation.
>
> KubeDB provides only orchestration and management tooling for Kubernetes.
> It does not distribute, bundle, ship, or include any Oracle Database software or binaries.
>
> Users must provide their own Oracle container images and hold valid Oracle licenses.
> Users are solely responsible for compliance with Oracle’s licensing terms, including all rules regarding containers, Docker, and Kubernetes environments.
>
> KubeDB makes no representations or warranties regarding Oracle licensing compliance.
