---
title: Reconfiguring Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-overview
    name: Overview
    parent: neo4j-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Neo4j

This guide gives an overview of how KubeDB Ops-manager reconfigures `Neo4j` database components.

## Before You Begin

- You should be familiar with [Neo4j](/docs/guides/neo4j/concepts/neo4j.md).
- You should be familiar with [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).

## How Reconfiguring Neo4j Process Works

For a `Neo4jOpsRequest` with `spec.type: Reconfigure`, KubeDB Ops-manager:

1. Validates the requested configuration and builds the desired settings from the configuration Secret and inline `applyConfig` values. Inline settings take precedence.
2. Pauses conflicting reconciliations and compares the current and desired configuration.
3. Reads `SHOW SETTINGS` from a running server to determine whether each changed setting is dynamic.
4. Chooses live application or restart according to `spec.configuration.restart`.
5. For live application, applies the changed settings on one server first using `dbms.setConfigValue`. If a value is rejected, it attempts to restore settings already changed on that server and fails the request.
6. Updates the stored configuration so future Pods load the desired settings. For live application, applies the changes directly to every server and verifies that the servers report consistent values. For the restart path, restarts the servers to load the new configuration.
7. Marks the request `Successful` and resumes normal reconciliation.

## Restart Policy

Use `spec.configuration.restart` to control how the configuration is applied:

| Value | Behavior |
|-------|----------|
| `"auto"` (default) | Apply the configuration without restarting Pods when every change supports a live update. Otherwise, restart the database Pods. |
| `"true"` | Restart the database Pods to load the new configuration. |
| `"false"` | Apply the configuration only if every change supports a live update. Fail the OpsRequest if a restart is required. |

Quote `"true"` and `"false"` in YAML: this field takes a string, not a boolean. An omitted value behaves as `"auto"`. A request containing both dynamic and startup-only changes uses the restart path with `"auto"`; the operator does not apply only the dynamic subset.

The `ReconfigureMode` status condition records `Live` or `Restart` in its `reason` and explains the decision in its `message`. A failed request that requires a restart with `restart: "false"` reports the reason in its failure conditions. A request that changes no settings completes without a restart unless a restart was explicitly requested.

## Which Changes Require a Restart?

A setting can be added or changed live only if the running server reports `isDynamic: true` in `SHOW SETTINGS`. Dynamic changes are applied on every server, since `dbms.setConfigValue` affects only the server where it runs. The stored configuration is also updated so that the new values survive Pod replacement.

A restart is required when:

- A changed setting is not dynamic.
- A known setting is removed from the user configuration, including through `removeCustomConfig` or a replacement configuration Secret.
- `server.jvm.additional` changes.
- Another configuration file changes, such as `apoc.conf` or a logging XML file.
- The remote alias keystore configuration changes.
- A changed value uses command expansion, such as `$(...)`.

Unknown settings and values rejected during live application fail the request. Neo4j setting removal requires a restart even when the setting supports live value changes. Do not use the Elasticsearch YAML-null removal syntax for Neo4j settings.

## Configuration Fields

| Field | Description |
|-------|-------------|
| `restart` | `"auto"`, `"true"`, or `"false"`; controls whether Pods are restarted. Defaults to `"auto"`. |
| `configSecret` | Reference to a Secret containing custom Neo4j configuration. |
| `applyConfig` | Inline configuration merged with existing inline values and layered over the configuration Secret. |
| `removeCustomConfig` | Remove existing custom configuration. Removing known settings requires a restart. |
| `remoteAliasKeystore` | Configure the remote alias keystore. Changes require a restart. |

## Next Step

Follow the [dynamic reconfiguration guide](/docs/guides/neo4j/reconfigure/dynamic-configuration.md) for a live update. For a configuration change that includes JVM arguments, follow: [Reconfigure Neo4j Cluster](/docs/guides/neo4j/reconfigure/reconfigure.md).
