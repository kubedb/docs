---
title: HanaDB
menu:
  docs_{{ .version }}:
    identifier: hanadb-readme
    name: HanaDB
    parent: guides-hanadb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/hanadb/
aliases:
  - /docs/{{ .version }}/guides/hanadb/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Overview

KubeDB operates [SAP HANA](https://www.sap.com/products/technology-platform/hana.html) databases on
Kubernetes through the `HanaDB` Custom Resource Definition (CRD). A single `HanaDB` object describes a
standalone HANA instance or a multi-node HANA **System Replication** cluster, and the KubeDB operator
provisions the PetSet, Services, authentication Secret, and AppBinding required to run it. Day-2
operations such as restart, reconfigure, TLS management, scaling, volume expansion, storage migration,
and credential rotation are driven declaratively through the `HanaDBOpsRequest` CRD.

The guides in this section use [SAP HANA, express edition](https://www.sap.com/products/technology-platform/hana/express-trial.html)
(`hanaexpress`) images.

## Supported HanaDB Features

| Features                                              | Availability |
|-------------------------------------------------------|:------------:|
| Standalone instance                                   |   &#10003;   |
| System Replication cluster (multi-node)               |   &#10003;   |
| Synchronous / async replication modes                 |   &#10003;   |
| Read-enabled secondary (`logreplay_readaccess`)       |   &#10003;   |
| Persistent Volume                                     |   &#10003;   |
| Custom Configuration (`global.ini`)                   |   &#10003;   |
| Custom docker image                                   |   &#10003;   |
| Authentication (auto-generated credentials)           |   &#10003;   |
| TLS/SSL (cert-manager)                                |   &#10003;   |
| Builtin Prometheus Monitoring                         |   &#10003;   |
| Prometheus Operator Monitoring                        |   &#10003;   |
| Restart (`HanaDBOpsRequest`)                          |   &#10003;   |
| Reconfigure (`HanaDBOpsRequest`)                      |   &#10003;   |
| Reconfigure TLS (`HanaDBOpsRequest`)                  |   &#10003;   |
| Vertical Scaling (`HanaDBOpsRequest`)                 |   &#10003;   |
| Volume Expansion (`HanaDBOpsRequest`)                 |   &#10003;   |
| Horizontal Scaling (`HanaDBOpsRequest`)               |   &#10003;   |
| Storage Migration (`HanaDBOpsRequest`)                |   &#10003;   |
| Rotate Authentication (`HanaDBOpsRequest`)            |   &#10003;   |

## User Guide

- [HanaDB Quickstart](/docs/guides/hanadb/quickstart/quickstart.md) — deploy a standalone HanaDB and connect to it.
- [HanaDB System Replication](/docs/guides/hanadb/clustering/system-replication.md) — run a multi-node HANA System Replication cluster.
- [Custom Configuration](/docs/guides/hanadb/configuration/using-config-file.md) — supply a custom `global.ini`.
- [Monitoring with builtin Prometheus](/docs/guides/hanadb/monitoring/using-builtin-prometheus.md).
- [Monitoring with Prometheus Operator](/docs/guides/hanadb/monitoring/using-prometheus-operator.md).
- [TLS/SSL Encryption](/docs/guides/hanadb/tls/overview.md).
- Day-2 operations: [Restart](/docs/guides/hanadb/restart/restart.md), [Reconfigure](/docs/guides/hanadb/reconfigure/reconfigure.md), [Vertical Scaling](/docs/guides/hanadb/scaling/vertical-scaling/vertical-scaling.md), [Volume Expansion](/docs/guides/hanadb/volume-expansion/volume-expansion.md), [Storage Migration](/docs/guides/hanadb/storage-migration/storage-migration.md), [Rotate Authentication](/docs/guides/hanadb/rotate-authentication/rotate-authentication.md).

## Concepts

- [HanaDB CRD](/docs/guides/hanadb/concepts/hanadb.md)
- [HanaDBVersion CRD](/docs/guides/hanadb/concepts/catalog.md)
- [HanaDBOpsRequest CRD](/docs/guides/hanadb/concepts/opsrequest.md)
- [AppBinding](/docs/guides/hanadb/concepts/appbinding.md)

> ## ⚠️ Legal Notice
>
> SAP® and SAP HANA® are registered trademarks of SAP SE. KubeDB is not affiliated with, endorsed by,
> or sponsored by SAP SE.
>
> KubeDB provides only orchestration and management tooling for Kubernetes. It does not distribute,
> bundle, ship, or include any SAP HANA software or binaries. Users must provide their own SAP HANA
> container images and hold valid SAP licenses, and are solely responsible for compliance with SAP's
> licensing terms.
