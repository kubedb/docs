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

The following diagram shows how KubeDB Ops-manager operator reconfigures `Elasticsearch` components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring process of Elasticsearch" src="/docs/images/elasticsearch/es-reconfigure.svg">
<figcaption align="center">Fig: Reconfiguring process of Elasticsearch</figcaption>
</figure>

The Reconfiguring Elasticsearch process consists of the following steps:

1. At first, a user creates an `Elasticsearch` Custom Resource (CR).

2. `KubeDB` Provisioner operator watches the `Elasticsearch` CR.

3. When the operator finds an `Elasticsearch` CR, it creates the required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the various components (ie. Combined, Topology) of the `Elasticsearch`, the user creates an `ElasticsearchOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `ElasticsearchOpsRequest` CR.

6. When it finds an `ElasticsearchOpsRequest` CR, it halts the `Elasticsearch` object which is referred from the `ElasticsearchOpsRequest`. So, the `KubeDB` Provisioner operator doesn't perform any operations on the `Elasticsearch` object during the reconfiguring process.

7. Then the `KubeDB` Ops-manager operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `ElasticsearchOpsRequest` CR.

8. Then the `KubeDB` Ops-manager operator will restart the related PetSet Pods so that they restart with the new configuration defined in the `ElasticsearchOpsRequest` CR.

9. After the successful reconfiguring of the `Elasticsearch` components, the `KubeDB` Ops-manager operator resumes the `Elasticsearch` object so that the `KubeDB` Provisioner operator resumes its usual operations.
### spec.configuration Fields

The `ElasticsearchOpsRequest` with `type: Reconfigure` uses the following sub-fields under `spec.configuration`:

| Field | Description |
|---|---|
| `configSecret` | Reference to a `Secret` containing the full custom configuration file(s) for the database. |
| `secureConfigSecret` | Reference to a `Secret` containing secure settings (e.g., keystore passwords). |
| `applyConfig` | Inline map of `filename: \| content` entries. Merged into the existing `ConfigSecret`. If no `ConfigSecret` exists, a new secret named `{db-name}-user-config` is created. |
| `removeCustomConfig` | Set to `true` to remove all user-provided configuration and revert to the operator-generated defaults. |
| `removeSecureCustomConfig` | Set to `true` to remove user-provided secure settings and revert to the default empty keystore. |

In the next docs, we are going to show a step by step guide on reconfiguring Elasticsearch components using `ElasticsearchOpsRequest` CRD.
