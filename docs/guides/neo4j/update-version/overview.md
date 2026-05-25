---
title: Updating Neo4j Version
menu:
  docs_{{ .version }}:
    identifier: neo4j-update-version-overview
    name: Overview
    parent: neo4j-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Neo4j Update Version Overview

This page explains how KubeDB Ops-manager upgrades Neo4j using `Neo4jOpsRequest`.

## Before You Begin

- You should be familiar with [Neo4j](/docs/guides/neo4j/concepts/neo4j.md).
- You should be familiar with [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).

## How Update Version Works

The following diagram shows how KubeDB Ops-manager updates the version of a `Neo4j` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Update version process of Neo4j" src="/docs/images/neo4j/UpdateVersion.png">
  <figcaption align="center">Fig: Update version process of Neo4j</figcaption>
</figure>

The update version process consists of the following steps:

For a `Neo4jOpsRequest` with `spec.type: UpdateVersion`, KubeDB Ops-manager:

1. Validates target version from `spec.updateVersion.targetVersion`.
2. Pauses conflicting reconciliations for safe upgrade.
3. Updates Neo4j image/version references.
4. Performs controlled rolling update of Neo4j members.
5. Waits for all pods and cluster status to become healthy.
6. Marks the request `Successful` and resumes reconciliation.

## Next Step

Follow the detailed guide: [Upgrade Neo4j Version](/docs/guides/neo4j/update-version/versionupgrading/index.md).
