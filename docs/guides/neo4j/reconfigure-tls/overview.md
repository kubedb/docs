---
title: Reconfiguring Neo4j TLS
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-tls-overview
    name: Overview
    parent: neo4j-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring TLS of Neo4j Database

This guide gives an overview of how KubeDB Ops-manager reconfigures TLS for a `Neo4j` database, including adding TLS, rotating certificates, updating issuer reference, and removing TLS through `Neo4jOpsRequest`.

## Before You Begin

- You should be familiar with [Neo4j](/docs/guides/neo4j/concepts/neo4j.md).
- You should be familiar with [Neo4jOpsRequest](/docs/guides/neo4j/concepts/opsrequest.md).

## How Reconfiguring Neo4j TLS Works

The following diagram shows the TLS reconfiguration flow for a `Neo4j` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring TLS process of Neo4j" src="/docs/images/neo4j/reconfigureTLS.png">
  <figcaption align="center">Fig: Reconfiguring TLS process of Neo4j</figcaption>
</figure>

The process consists of the following steps:

1. A user creates a `Neo4j` Custom Resource.
2. KubeDB Provisioner reconciles the database and creates required workloads and secrets.
3. To update TLS settings, the user creates a `Neo4jOpsRequest` with `spec.type: ReconfigureTLS`.
4. KubeDB Ops-manager watches the `Neo4jOpsRequest` and validates the `spec.tls` fields.
5. Ops-manager temporarily pauses conflicting reconciliation for the target database.
6. It applies the requested TLS action (add/update via `issuerRef`, rotate via `rotateCertificates`, or disable via `remove`).
7. It rolls/restarts the required pods so updated TLS configuration is picked up.
8. After successful checks, Ops-manager marks the request `Successful` and resumes normal reconciliation.

In the next guide, we show the step-by-step workflow for each TLS reconfiguration operation.

## Next Step

- Follow: [Reconfigure TLS in Neo4j](/docs/guides/neo4j/reconfigure-tls/reconfigure-tls.md).
