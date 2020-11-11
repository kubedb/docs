---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2020.11.11
    name: Changelog-v2020.11.11
    parent: welcome
    weight: 20201111
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2020.11.11/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2020.11.11/
---

# KubeDB v2020.11.11 (2020-11-11)


## [appscode/kubedb-enterprise](https://github.com/appscode/kubedb-enterprise)

### [v0.2.0](https://github.com/appscode/kubedb-enterprise/releases/tag/v0.2.0)

- [b96f1b56](https://github.com/appscode/kubedb-enterprise/commit/b96f1b56) Prepare for release v0.2.0 (#95)



## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.15.0](https://github.com/kubedb/apimachinery/releases/tag/v0.15.0)

- [592d5b47](https://github.com/kubedb/apimachinery/commit/592d5b47) Add default resource limit (#650)
- [b5fa4a10](https://github.com/kubedb/apimachinery/commit/b5fa4a10) Rename MasterServiceName to MasterDiscoveryServiceName
- [849e6c06](https://github.com/kubedb/apimachinery/commit/849e6c06) Remove ElasticsearchMetricsPortName
- [421d760b](https://github.com/kubedb/apimachinery/commit/421d760b) Add HasServiceTemplate & GetServiceTemplate helpers (#649)
- [25e0e4af](https://github.com/kubedb/apimachinery/commit/25e0e4af) Enable separate serviceTemplate for each service (#648)
- [f325af77](https://github.com/kubedb/apimachinery/commit/f325af77) Remove replicaServiceTemplate from Postgres CRD (#646)
- [31286270](https://github.com/kubedb/apimachinery/commit/31286270) Add `ReplicationModeDetector` Image for MongoDB (#645)
- [df5ada37](https://github.com/kubedb/apimachinery/commit/df5ada37) Update Elasticsearch constants  (#639)
- [1387ac1f](https://github.com/kubedb/apimachinery/commit/1387ac1f) Remove version label from database labels (#644)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.15.0](https://github.com/kubedb/cli/releases/tag/v0.15.0)

- [df75044a](https://github.com/kubedb/cli/commit/df75044a) Prepare for release v0.15.0 (#549)
- [08bce120](https://github.com/kubedb/cli/commit/08bce120) Update KubeDB api (#548)
- [3f4e0fd5](https://github.com/kubedb/cli/commit/3f4e0fd5) Update KubeDB api (#547)
- [56429d25](https://github.com/kubedb/cli/commit/56429d25) Update KubeDB api (#546)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.15.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.15.0)

- [bffb46a0](https://github.com/kubedb/elasticsearch/commit/bffb46a0) Prepare for release v0.15.0 (#414)
- [8915b7b5](https://github.com/kubedb/elasticsearch/commit/8915b7b5) Update KubeDB api (#413)
- [6a95dbf1](https://github.com/kubedb/elasticsearch/commit/6a95dbf1) Allow stats service patching
- [93b8501c](https://github.com/kubedb/elasticsearch/commit/93b8501c) Use separate ServiceTemplate for each service (#412)
- [87ba4941](https://github.com/kubedb/elasticsearch/commit/87ba4941) Use container name as constant (#402)
- [a1b6343a](https://github.com/kubedb/elasticsearch/commit/a1b6343a) Update KubeDB api (#411)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v0.15.0](https://github.com/kubedb/installer/releases/tag/v0.15.0)

- [a8c5b9c](https://github.com/kubedb/installer/commit/a8c5b9c) Prepare for release v0.15.0 (#204)
- [f17451e](https://github.com/kubedb/installer/commit/f17451e) Add `ReplicationModeDetector` Image for MongoDB (#202)
- [8dc9941](https://github.com/kubedb/installer/commit/8dc9941) Add permissions to evict pods (#201)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.8.0](https://github.com/kubedb/memcached/releases/tag/v0.8.0)

- [aff58c9e](https://github.com/kubedb/memcached/commit/aff58c9e) Prepare for release v0.8.0 (#243)
- [3dbf5486](https://github.com/kubedb/memcached/commit/3dbf5486) Update KubeDB api (#242)
- [d1821f03](https://github.com/kubedb/memcached/commit/d1821f03) Use separate ServiceTemplate for each service (#241)
- [44ea6d2b](https://github.com/kubedb/memcached/commit/44ea6d2b) Update KubeDB api (#240)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.8.0](https://github.com/kubedb/mongodb/releases/tag/v0.8.0)

- [3b0f1d08](https://github.com/kubedb/mongodb/commit/3b0f1d08) Prepare for release v0.8.0 (#318)
- [b3685ab8](https://github.com/kubedb/mongodb/commit/b3685ab8) Update KubeDB api (#317)
- [bf9d872c](https://github.com/kubedb/mongodb/commit/bf9d872c) Allow stats service patching
- [183f1ac3](https://github.com/kubedb/mongodb/commit/183f1ac3) Use separate ServiceTemplate for each service (#315)
- [3d105d2a](https://github.com/kubedb/mongodb/commit/3d105d2a) Fix Health Check (#305)
- [98fe156b](https://github.com/kubedb/mongodb/commit/98fe156b) Add `ReplicationModeDetector` (#316)
- [539337b0](https://github.com/kubedb/mongodb/commit/539337b0) Update KubeDB api (#314)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.8.0](https://github.com/kubedb/mysql/releases/tag/v0.8.0)

- [b83e2323](https://github.com/kubedb/mysql/commit/b83e2323) Prepare for release v0.8.0 (#305)
- [1916fb3d](https://github.com/kubedb/mysql/commit/1916fb3d) Update KubeDB api (#304)
- [2e2dd9b0](https://github.com/kubedb/mysql/commit/2e2dd9b0) Allow stats service patching
- [18cbe558](https://github.com/kubedb/mysql/commit/18cbe558) Use separate ServiceTemplate for each service (#303)
- [741c9718](https://github.com/kubedb/mysql/commit/741c9718) Fix MySQL args (#295)



## [kubedb/operator](https://github.com/kubedb/operator)

### [v0.15.0](https://github.com/kubedb/operator/releases/tag/v0.15.0)

- [c9c540f0](https://github.com/kubedb/operator/commit/c9c540f0) Prepare for release v0.15.0 (#349)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.2.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.2.0)

- [13b6c2d7](https://github.com/kubedb/percona-xtradb/commit/13b6c2d7) Prepare for release v0.2.0 (#138)
- [6e4a4449](https://github.com/kubedb/percona-xtradb/commit/6e4a4449) Update KubeDB api (#137)
- [717fca92](https://github.com/kubedb/percona-xtradb/commit/717fca92) Use separate ServiceTemplate for each service (#136)
- [3386c10e](https://github.com/kubedb/percona-xtradb/commit/3386c10e) Update KubeDB api (#135)



## [kubedb/pg-leader-election](https://github.com/kubedb/pg-leader-election)

### [v0.3.0](https://github.com/kubedb/pg-leader-election/releases/tag/v0.3.0)




## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.2.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.2.0)

- [a9c88518](https://github.com/kubedb/pgbouncer/commit/a9c88518) Prepare for release v0.2.0 (#108)
- [56132158](https://github.com/kubedb/pgbouncer/commit/56132158) Update KubeDB api (#107)
- [2d9e4490](https://github.com/kubedb/pgbouncer/commit/2d9e4490) Use separate ServiceTemplate for each service (#106)
- [9cfb2ae2](https://github.com/kubedb/pgbouncer/commit/9cfb2ae2) Update KubeDB api (#105)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.15.0](https://github.com/kubedb/postgres/releases/tag/v0.15.0)

- [929217b2](https://github.com/kubedb/postgres/commit/929217b2) Prepare for release v0.15.0 (#424)
- [7782f03d](https://github.com/kubedb/postgres/commit/7782f03d) Update KubeDB api (#423)
- [0216423d](https://github.com/kubedb/postgres/commit/0216423d) Allow stats service patching
- [6f5e3b57](https://github.com/kubedb/postgres/commit/6f5e3b57) Use separate ServiceTemplate for each service (#422)
- [111fddfa](https://github.com/kubedb/postgres/commit/111fddfa) Update KubeDB api (#421)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.2.0](https://github.com/kubedb/proxysql/releases/tag/v0.2.0)

- [71444683](https://github.com/kubedb/proxysql/commit/71444683) Prepare for release v0.2.0 (#120)
- [ce811abf](https://github.com/kubedb/proxysql/commit/ce811abf) Update KubeDB api (#119)
- [4ed10ea2](https://github.com/kubedb/proxysql/commit/4ed10ea2) Use separate ServiceTemplate for each service (#118)
- [d43e7359](https://github.com/kubedb/proxysql/commit/d43e7359) Update KubeDB api (#117)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.8.0](https://github.com/kubedb/redis/releases/tag/v0.8.0)

- [a2fe5b3b](https://github.com/kubedb/redis/commit/a2fe5b3b) Prepare for release v0.8.0 (#262)
- [9de30e41](https://github.com/kubedb/redis/commit/9de30e41) Update KubeDB api (#261)
- [5c8281d2](https://github.com/kubedb/redis/commit/5c8281d2) Use separate ServiceTemplate for each service (#260)
- [3a269916](https://github.com/kubedb/redis/commit/3a269916) Update KubeDB api (#259)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.2.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.2.0)

- [70416f7](https://github.com/kubedb/replication-mode-detector/commit/70416f7) Prepare for release v0.2.0 (#92)
- [d75f103](https://github.com/kubedb/replication-mode-detector/commit/d75f103) Update KubeDB api (#91)
- [5f03577](https://github.com/kubedb/replication-mode-detector/commit/5f03577) Add MongoDB `ReplicationModeDetector` (#90)
- [8456fdc](https://github.com/kubedb/replication-mode-detector/commit/8456fdc) Drop mysql from repo name (#89)




