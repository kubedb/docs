---
title: ZooKeeper
menu:
  docs_{{ .version }}:
    identifier: zk-readme-zookeeper
    name: ZooKeeper
    parent: zk-zookeeper-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/zookeeper/
aliases:
  - /docs/{{ .version }}/guides/zookeeper/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported ZooKeeper Features
| Features                                                   | Community | Enterprise |
|------------------------------------------------------------|:---------:|:----------:|
| Ensemble                                                   | &#10003;  |  &#10003;  |
| Standalone                                                 | &#10003;  |  &#10003;  |
| Authentication & Autorization                              | &#10003;  |  &#10003;  |
| Persistent Volume                                          | &#10003;  |  &#10003;  |
| Initializing from Snapshot ( [Stash](https://stash.run/) ) | &#10003;  |  &#10003;  |
| Instant Backup                                             | &#10003;  |  &#10003;  |
| Scheduled Backup                                           | &#10003;  |  &#10003;  |
| Builtin Prometheus Discovery                               | &#10003;  |  &#10003;  |
| Using Prometheus operator                                  | &#10003;  |  &#10003;  |

## Life Cycle of a ZooKeeper Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/zookeeper/zookeeper-lifecycle.png">
</p>

## User Guide

- [Quickstart ZooKeeper](/docs/guides/zookeeper/quickstart/quickstart.md) with KubeDB Operator.

