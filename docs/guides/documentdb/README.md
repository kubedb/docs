---
title: DocumentDB
menu:
  docs_{{ .version }}:
    identifier: dc-documentdb-guides-readme
    name: DocumentDB
    parent: dc-documentdb-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/documentdb/
aliases:
  - /docs/{{ .version }}/guides/documentdb/README/
---
> New to KubeDB? Please start [here](/docs/README.md).

# Overview

DocumentDB is a document database that speaks the MongoDB wire protocol on top of a PostgreSQL engine (via FerretDB), giving you MongoDB-compatible APIs backed by the reliability and tooling of PostgreSQL. KubeDB lets you provision and operate DocumentDB clusters on Kubernetes, handling day-2 operations such as scaling, reconfiguration, failover, and storage management through declarative `DocumentDBOpsRequest` objects.

## Supported DocumentDB Features

| Features                                                      | Availability |
|---------------------------------------------------------------|:------------:|
| Clustering (High Availability)                                |   &#10003;   |
| Custom Configuration (`user.conf`)                            |   &#10003;   |
| Externally manageable Auth Secret (Rotate Auth)               |   &#10003;   |
| Persistent volume                                             |   &#10003;   |
| Automated Vertical & Horizontal Scaling                       |   &#10003;   |
| Automated Volume Expansion                                    |   &#10003;   |
| Storage Migration (StorageClass to StorageClass)              |   &#10003;   |
| Autoscaling ( Compute resources & Storage )                   |   &#10003;   |
| Automated Failover & Disaster Recovery                        |   &#10003;   |
| Automated Restart                                             |   &#10003;   |
| Reconfigurable runtime configuration                          |   &#10003;   |

## Supported DocumentDB Versions

KubeDB supports the following DocumentDB Versions.
- `15.12-documentdb`
- `16.8-documentdb`
- `17.4-documentdb`

## User Guide

- [Custom Configuration](/docs/guides/documentdb/configuration/using-config-file.md) using a config file.
- [Reconfigure](/docs/guides/documentdb/reconfigure/reconfigure.md) a running DocumentDB database.
- [Failover & Disaster Recovery](/docs/guides/documentdb/failure-and-disaster-recovery/failover.md) of a DocumentDB cluster.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
