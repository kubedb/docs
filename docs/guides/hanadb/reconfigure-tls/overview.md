---
title: Reconfigure HanaDB TLS Overview
menu:
  docs_{{ .version }}:
    identifier: hanadb-reconfigure-tls-overview
    name: Overview
    parent: hanadb-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# Reconfigure HanaDB TLS

KubeDB supports adding TLS to an existing HanaDB database and rotating existing TLS certificates with a `HanaDBOpsRequest` of type `ReconfigureTLS`.

## Before You Begin

You should be familiar with the following KubeDB concepts:

- [HanaDB](/docs/guides/hanadb/concepts/hanadb.md)
- [HanaDBOpsRequest](/docs/guides/hanadb/concepts/opsrequest.md)
- [HanaDB TLS](/docs/guides/hanadb/tls/overview.md)

## How ReconfigureTLS Works

The ReconfigureTLS process consists of the following steps:

1. A user creates a `HanaDB` object.
2. The KubeDB Provisioner provisions the required PetSet, services, secrets, and related resources.
3. To add or rotate TLS, the user creates a `HanaDBOpsRequest` with `spec.type: ReconfigureTLS`.
4. The KubeDB Ops Manager pauses the referenced `HanaDB` object while the operation is running.
5. Ops Manager updates the TLS configuration and certificate resources according to the OpsRequest.
6. Ops Manager restarts the HanaDB pods so the database and KubeDB clients use the updated TLS configuration.
7. After the operation succeeds, Ops Manager resumes the `HanaDB` object.

See the [Reconfigure TLS guide](/docs/guides/hanadb/reconfigure-tls/reconfigure-tls.md) for a step-by-step example.
