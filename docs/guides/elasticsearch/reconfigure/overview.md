---
title: Reconfiguring Elasticsearch
menu:
  docs_{{ .version }}:
    identifier: es-reconfigure-overview
    name: Overview
    parent: es-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Elasticsearch

This guide will give an overview on how KubeDB Ops-manager operator reconfigures `Elasticsearch` components such as Combined, Topology (master, data, ingest), etc.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)

## How Reconfiguring Elasticsearch Process Works

For an `ElasticsearchOpsRequest` with `spec.type: Reconfigure`, KubeDB Ops-manager:

1. Pauses provisioner reconciliation for the target database.
2. Builds the desired configuration from the configuration Secret and inline `applyConfig`, with inline settings taking precedence.
3. Compares the current and desired configuration and chooses live application or restart according to `spec.configuration.restart`.
4. For live application, sends the changed settings in one persistent cluster settings API update and checks that the settings were set or reset. Elasticsearch/OpenSearch rejects the update if any setting is static, unknown, or has an invalid value.
5. Updates the database configuration and PetSets so future Pods load the desired configuration. Updating a PetSet does not itself restart the existing Pods.
6. For the restart path, restarts the Pods and clears persistent cluster settings whose keys appear in the old or new user configuration, allowing the newly loaded configuration files to take effect.
7. Marks the OpsRequest `Successful` and resumes provisioner reconciliation.

The same behavior applies to OpenSearch, using `opensearch.yml` as the main configuration file instead of `elasticsearch.yml`.

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

Live updates are candidates only when every changed setting belongs to the main configuration file and can be applied through the cluster settings API. The running Elasticsearch/OpenSearch server determines which settings are dynamic; the operator does not maintain a fixed list.

A restart is required when:

- A file other than the main configuration file changes, including `jvm.options`, `log4j2.properties`, or a role-specific file such as `data-elasticsearch.yml`.
- A changed main-file setting is also defined in a role-specific file. A cluster-wide API update would override that role-specific value.
- The secure configuration Secret reference changes.
- A changed value contains environment variable substitution, such as `${ZONE}`.
- A removed setting was loaded by any running node at startup. Resetting a persistent API setting alone would fall back to that startup value.
- The server rejects a changed setting as not dynamically updateable.

Unknown settings and invalid values fail the live update attempt rather than triggering a restart.

### Removing Settings

A YAML `null` value inside `applyConfig.elasticsearch.yml` (or `applyConfig.opensearch.yml`) removes that setting from the merged inline configuration. The operator removes null-valued entries before writing node configuration files. A value still supplied by the configuration Secret remains effective; to remove it entirely, use a replacement Secret without that setting and remove any inline override.

A setting previously introduced only through live reconfiguration can be reset live when no running node has it in its startup settings. Otherwise, removal requires a restart. After a restart, the operator clears persistent overrides for keys in the old or new user configuration; persistent settings with other keys are left untouched.

### spec.configuration Fields

The `ElasticsearchOpsRequest` with `type: Reconfigure` uses the following sub-fields under `spec.configuration`:

| Field | Description |
|---|---|
| `restart` | `"auto"`, `"true"`, or `"false"`; controls whether Pods are restarted. Defaults to `"auto"`. |
| `configSecret` | Reference to a `Secret` containing the full custom configuration file(s) for the database. |
| `secureConfigSecret` | Reference to a `Secret` containing secure settings (e.g., keystore passwords). |
| `applyConfig` | Inline map of filenames to configuration content. Merged with the existing inline configuration and layered over the configuration Secret; unchanged files are preserved. |
| `removeCustomConfig` | Set to `true` to remove all user-provided configuration and revert to the operator-generated defaults. |
| `removeSecureCustomConfig` | Set to `true` to remove user-provided secure settings and revert to the default empty keystore. |

Follow the [dynamic reconfiguration guide](/docs/guides/elasticsearch/reconfigure/dynamic-configuration.md) for a live update, or the combined and topology guides for configuration-file changes.
