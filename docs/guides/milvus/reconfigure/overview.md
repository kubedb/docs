---
title: Reconfigure Milvus Overview
menu:
  docs_{{ .version }}:
    identifier: milvus-reconfigure-overview
    name: Overview
    parent: milvus-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Milvus

This guide will give an overview on how the KubeDB Ops-manager operator reconfigures a `Milvus` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)

## How Reconfigure Process Works

Milvus is configured through a single configuration file that **must** be named `milvus.yaml`. KubeDB renders a base configuration and merges any custom configuration you supply on top of it.

A `MilvusOpsRequest` of type `Reconfigure` lets you change that custom configuration on a running database. The `spec.configuration` block supports:

- **`configSecret`** — a reference to a `Secret` whose `milvus.yaml` key holds the new configuration.
- **`applyConfig`** — an inline map; the value under the `milvus.yaml` key is merged into the existing configuration. This is convenient for small, incremental changes.
- **`removeCustomConfig`** — when `true`, the previously applied custom configuration is discarded before the new configuration is applied.
- **`restart`** — `auto` (default), `true`, or `false`. Controls whether the pods are restarted after the configuration change.

The flow is:

1. A user creates a `MilvusOpsRequest` of type `Reconfigure`.
2. The operator validates the request and pauses the `Milvus` database.
3. The new configuration is merged and written to the rendered configuration secret (`<db>-<hash>`), under the `milvus.yaml` key.
4. The PetSets are reconciled and (depending on `restart`) the pods are recreated to load the new configuration.
5. The operator resumes the database and marks the `MilvusOpsRequest` as `Successful`.

> **Important:** the configuration key is always `milvus.yaml` — use that exact key in your config secret and in `applyConfig`.

In the next doc, we will see a step-by-step guide on reconfiguring a Milvus database.
