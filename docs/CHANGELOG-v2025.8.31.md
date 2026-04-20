---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2025.8.31
    name: Changelog-v2025.8.31
    parent: welcome
    weight: 20250831
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2025.8.31/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2025.8.31/
---

# KubeDB v2025.8.31 (2025-08-31)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.58.0](https://github.com/kubedb/apimachinery/releases/tag/v0.58.0)

- [ce48b0d0](https://github.com/kubedb/apimachinery/commit/ce48b0d05) Update deps
- [079f9604](https://github.com/kubedb/apimachinery/commit/079f96043) Add spec for MySQL & postgres StorageClassMigration opsrequest (#1498)
- [09a35f28](https://github.com/kubedb/apimachinery/commit/09a35f28d) Find the domain from resolve.conf file (#1508)
- [96d4d2fd](https://github.com/kubedb/apimachinery/commit/96d4d2fd3) Update petset api related changes (#1505)
- [b8bdbe2c](https://github.com/kubedb/apimachinery/commit/b8bdbe2c9) Add git-sync in redis pgbouncer pgpool  (#1506)
- [dc61a043](https://github.com/kubedb/apimachinery/commit/dc61a0432) Deploy multiple clickhouse CR if multi cluster needed (#1507)
- [f56c4f01](https://github.com/kubedb/apimachinery/commit/f56c4f014) Add Ignite autoscaler (#1502)
- [e310bc75](https://github.com/kubedb/apimachinery/commit/e310bc751) Add clickhouse ReconfigureTLS, RotateAuth (#1496)
- [6608a4cc](https://github.com/kubedb/apimachinery/commit/6608a4ccf) Use Go 1.25 (#1504)
- [e962b44b](https://github.com/kubedb/apimachinery/commit/e962b44b1) Test against k8s 1.33.2 (#1503)
- [0ae8b7da](https://github.com/kubedb/apimachinery/commit/0ae8b7dae) Fix memcached buggy code (#1501)
- [b314b1ab](https://github.com/kubedb/apimachinery/commit/b314b1ab6) Add RotateAuth OPS Validation (#1500)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.43.0](https://github.com/kubedb/autoscaler/releases/tag/v0.43.0)

- [41eca2aa](https://github.com/kubedb/autoscaler/commit/41eca2aa) Prepare for release v0.43.0 (#258)
- [ab213e24](https://github.com/kubedb/autoscaler/commit/ab213e24) Add support for Ignite (#256)
- [fbba3a3f](https://github.com/kubedb/autoscaler/commit/fbba3a3f) Use Go 1.25 (#257)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.11.0](https://github.com/kubedb/cassandra/releases/tag/v0.11.0)

- [d669103e](https://github.com/kubedb/cassandra/commit/d669103e) Prepare for release v0.11.0 (#47)
- [09a394ca](https://github.com/kubedb/cassandra/commit/09a394ca) Set domain (#46)
- [acb70c4e](https://github.com/kubedb/cassandra/commit/acb70c4e) Ignore notFound error on deletion (#45)
- [8ef5360d](https://github.com/kubedb/cassandra/commit/8ef5360d) Use Go 1.25 (#44)
- [33039dd4](https://github.com/kubedb/cassandra/commit/33039dd4) Test against k8s 1.33.2 (#43)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.5.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.5.0)

- [2742636](https://github.com/kubedb/cassandra-medusa-plugin/commit/2742636) Prepare for release v0.5.0 (#11)
- [4640036](https://github.com/kubedb/cassandra-medusa-plugin/commit/4640036) Use Go 1.25 (#10)
- [dcdd79e](https://github.com/kubedb/cassandra-medusa-plugin/commit/dcdd79e) Test against k8s 1.33.2 (#9)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.58.0](https://github.com/kubedb/cli/releases/tag/v0.58.0)

- [cbed3e12](https://github.com/kubedb/cli/commit/cbed3e123) Prepare for release v0.58.0 (#802)
- [1595907f](https://github.com/kubedb/cli/commit/1595907f3) Use Go 1.25 (#801)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.13.0](https://github.com/kubedb/clickhouse/releases/tag/v0.13.0)

- [f2f4e546](https://github.com/kubedb/clickhouse/commit/f2f4e546) Prepare for release v0.13.0 (#64)
- [dac8281d](https://github.com/kubedb/clickhouse/commit/dac8281d) Deploy multiple clickhouse CR if multi cluster needed (#63)
- [28f74f35](https://github.com/kubedb/clickhouse/commit/28f74f35) Export EnsureSecrets (#59)
- [c69becf2](https://github.com/kubedb/clickhouse/commit/c69becf2) Use Go 1.25 (#62)
- [dd1342e7](https://github.com/kubedb/clickhouse/commit/dd1342e7) Test against k8s 1.33.2 (#61)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.13.0](https://github.com/kubedb/crd-manager/releases/tag/v0.13.0)

- [659181fe](https://github.com/kubedb/crd-manager/commit/659181fe) Prepare for release v0.13.0 (#87)
- [237fba02](https://github.com/kubedb/crd-manager/commit/237fba02) Register clickhouse binding crd
- [f6c95628](https://github.com/kubedb/crd-manager/commit/f6c95628) Use Go 1.25 (#86)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.16.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.16.0)

- [e42a773](https://github.com/kubedb/dashboard-restic-plugin/commit/e42a773) Prepare for release v0.16.0 (#46)
- [7922ad9](https://github.com/kubedb/dashboard-restic-plugin/commit/7922ad9) Use Go 1.25 (#45)
- [3bfebe8](https://github.com/kubedb/dashboard-restic-plugin/commit/3bfebe8) Test against k8s 1.33.2 (#44)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.13.0](https://github.com/kubedb/db-client-go/releases/tag/v0.13.0)

- [981dd5ed](https://github.com/kubedb/db-client-go/commit/981dd5ed) Prepare for release v0.13.0 (#194)
- [68321723](https://github.com/kubedb/db-client-go/commit/68321723) Update dynamic k8s host (#193)
- [435a48d0](https://github.com/kubedb/db-client-go/commit/435a48d0) Use Go 1.25 (#192)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.13.0](https://github.com/kubedb/druid/releases/tag/v0.13.0)

- [86748406](https://github.com/kubedb/druid/commit/86748406) Prepare for release v0.13.0 (#98)
- [25f08469](https://github.com/kubedb/druid/commit/25f08469) set domain (#97)
- [b6f7df3d](https://github.com/kubedb/druid/commit/b6f7df3d) make fmt (#96)
- [9b02466b](https://github.com/kubedb/druid/commit/9b02466b) Ignore notFound error on deletion (#95)
- [83957046](https://github.com/kubedb/druid/commit/83957046) Use Go 1.25 (#94)
- [9e4adb30](https://github.com/kubedb/druid/commit/9e4adb30) Test against k8s 1.33.2 (#93)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.58.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.58.0)

- [7a5ebc5d](https://github.com/kubedb/elasticsearch/commit/7a5ebc5d6) Prepare for release v0.58.0 (#774)
- [9381bab2](https://github.com/kubedb/elasticsearch/commit/9381bab20) Ignore notFound error on deletion (#773)
- [3ec57f3a](https://github.com/kubedb/elasticsearch/commit/3ec57f3ab) Use Go 1.25 (#772)
- [98484d3c](https://github.com/kubedb/elasticsearch/commit/98484d3ce) Test against k8s 1.33.2 (#771)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.21.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.21.0)

- [6d83b242](https://github.com/kubedb/elasticsearch-restic-plugin/commit/6d83b242) Prepare for release v0.21.0 (#70)
- [7ae5f210](https://github.com/kubedb/elasticsearch-restic-plugin/commit/7ae5f210) Use Go 1.25 (#69)
- [c19e72ec](https://github.com/kubedb/elasticsearch-restic-plugin/commit/c19e72ec) Test against k8s 1.33.2 (#68)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.13.0](https://github.com/kubedb/ferretdb/releases/tag/v0.13.0)

- [2e5bd231](https://github.com/kubedb/ferretdb/commit/2e5bd231) Prepare for release v0.13.0 (#84)
- [962f0bdf](https://github.com/kubedb/ferretdb/commit/962f0bdf) Ignore notFound error on deletion
- [f2ea8316](https://github.com/kubedb/ferretdb/commit/f2ea8316) Use Go 1.25 (#83)
- [22b52b26](https://github.com/kubedb/ferretdb/commit/22b52b26) Test against k8s 1.33.2 (#82)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.6.0](https://github.com/kubedb/gitops/releases/tag/v0.6.0)

- [ee522eb9](https://github.com/kubedb/gitops/commit/ee522eb9) Prepare for release v0.6.0 (#24)
- [46e0e8c7](https://github.com/kubedb/gitops/commit/46e0e8c7) Use Go 1.25 (#23)
- [b40b2928](https://github.com/kubedb/gitops/commit/b40b2928) Test against k8s 1.33.2 (#22)
- [ca143ff8](https://github.com/kubedb/gitops/commit/ca143ff8) Test against k8s 1.33.2 (#21)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.4.0](https://github.com/kubedb/hazelcast/releases/tag/v0.4.0)

- [d7db99da](https://github.com/kubedb/hazelcast/commit/d7db99da) Prepare for release v0.4.0 (#12)
- [a9c28eab](https://github.com/kubedb/hazelcast/commit/a9c28eab) Various changes (#9)
- [afa2f105](https://github.com/kubedb/hazelcast/commit/afa2f105) Use Go 1.25 (#11)
- [d208a416](https://github.com/kubedb/hazelcast/commit/d208a416) Test against k8s 1.33.2 (#10)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.5.0](https://github.com/kubedb/ignite/releases/tag/v0.5.0)

- [72293e6b](https://github.com/kubedb/ignite/commit/72293e6b) Prepare for release v0.5.0 (#20)
- [a22235d9](https://github.com/kubedb/ignite/commit/a22235d9) Ignore notFound error on deletion (#19)
- [ea7cd18c](https://github.com/kubedb/ignite/commit/ea7cd18c) Use Go 1.25 (#17)
- [657dab56](https://github.com/kubedb/ignite/commit/657dab56) Add PodPlacementPolicy field (#18)
- [d9d8ccb8](https://github.com/kubedb/ignite/commit/d9d8ccb8) Test against k8s 1.33.2 (#16)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2025.8.31](https://github.com/kubedb/installer/releases/tag/v2025.8.31)

- [91447ebd](https://github.com/kubedb/installer/commit/91447ebdb) Prepare for release v2025.8.31 (#1821)
- [ed3d2ff4](https://github.com/kubedb/installer/commit/ed3d2ff45) Update cve report (#1819)
- [6cb86bec](https://github.com/kubedb/installer/commit/6cb86bec0) Update petset operator v2025.8.31 (#1820)
- [d75eaee7](https://github.com/kubedb/installer/commit/d75eaee70) Git-Sync for redis pgpool pgbouncer (#1817)
- [ae70c072](https://github.com/kubedb/installer/commit/ae70c072c) Update crds for kubedb/apimachinery@079f9604 (#1818)
- [a11c50fc](https://github.com/kubedb/installer/commit/a11c50fc4) Add  git-sync image to mariadb  (#1815)
- [e55958df](https://github.com/kubedb/installer/commit/e55958df0) Fix CRD install fail issue for KubeSlice enabled namespace (#1816)
- [8502e461](https://github.com/kubedb/installer/commit/8502e4615) Update cve report (#1814)
- [cb1c7722](https://github.com/kubedb/installer/commit/cb1c7722c) Introduce "security.enableNetworkPolicy" (#1813)
- [8c931dab](https://github.com/kubedb/installer/commit/8c931dab7) Update crds for kubedb/apimachinery@b8bdbe2c (#1812)
- [d6207051](https://github.com/kubedb/installer/commit/d6207051a) Update crds for kubedb/apimachinery@dc61a043 (#1811)
- [649c09a7](https://github.com/kubedb/installer/commit/649c09a74) Update cve report (#1810)
- [a05171b7](https://github.com/kubedb/installer/commit/a05171b73) Update crds for kubedb/apimachinery@f56c4f01 (#1809)
- [5d3aacf1](https://github.com/kubedb/installer/commit/5d3aacf1b) Update cve report (#1807)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.29.0](https://github.com/kubedb/kafka/releases/tag/v0.29.0)

- [e367577d](https://github.com/kubedb/kafka/commit/e367577d) Prepare for release v0.29.0 (#161)
- [40fdcf94](https://github.com/kubedb/kafka/commit/40fdcf94) Update dynamic k8s domain (#160)
- [3ad3d393](https://github.com/kubedb/kafka/commit/3ad3d393) Handle monitor not found error (#159)
- [bfd1b42c](https://github.com/kubedb/kafka/commit/bfd1b42c) Use Go 1.25 (#158)
- [fd1f4975](https://github.com/kubedb/kafka/commit/fd1f4975) Test against k8s 1.33.2 (#157)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.34.0](https://github.com/kubedb/kibana/releases/tag/v0.34.0)

- [6a4afa04](https://github.com/kubedb/kibana/commit/6a4afa04) Prepare for release v0.34.0 (#158)
- [08d16c5b](https://github.com/kubedb/kibana/commit/08d16c5b) Use Go 1.25 (#157)
- [a9ee0b96](https://github.com/kubedb/kibana/commit/a9ee0b96) Test against k8s 1.33.2 (#156)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.21.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.21.0)

- [d2de8d3e](https://github.com/kubedb/kubedb-manifest-plugin/commit/d2de8d3e) Prepare for release v0.21.0 (#102)
- [cdcf6609](https://github.com/kubedb/kubedb-manifest-plugin/commit/cdcf6609) Use Go 1.25 (#101)
- [487f185c](https://github.com/kubedb/kubedb-manifest-plugin/commit/487f185c) Test against k8s 1.33.2 (#100)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.9.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.9.0)

- [15926d6](https://github.com/kubedb/kubedb-verifier/commit/15926d6) Prepare for release v0.9.0 (#24)
- [05e675e](https://github.com/kubedb/kubedb-verifier/commit/05e675e) Use Go 1.25 (#23)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.42.0](https://github.com/kubedb/mariadb/releases/tag/v0.42.0)

- [d2f2eeb4](https://github.com/kubedb/mariadb/commit/d2f2eeb43) Prepare for release v0.42.0 (#348)
- [fe8ae380](https://github.com/kubedb/mariadb/commit/fe8ae3805) Add mariadb git-sync feature (#346)
- [68044b2f](https://github.com/kubedb/mariadb/commit/68044b2f9) Split Distributed MariaDB Primary ServiceExport(Ops-TLS) (#344)
- [cb1eea24](https://github.com/kubedb/mariadb/commit/cb1eea241) set domain (#347)
- [eabab36f](https://github.com/kubedb/mariadb/commit/eabab36f4) Ignore notFound error on monitor deletion (#345)
- [b1e5701f](https://github.com/kubedb/mariadb/commit/b1e5701fa) Use Go 1.25 (#343)
- [632b070d](https://github.com/kubedb/mariadb/commit/632b070d6) Test against k8s 1.33.2 (#342)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.18.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.18.0)

- [0b44200f](https://github.com/kubedb/mariadb-archiver/commit/0b44200f) Prepare for release v0.18.0 (#58)
- [b7ef933c](https://github.com/kubedb/mariadb-archiver/commit/b7ef933c) Use Go 1.25 (#57)
- [c7a8956f](https://github.com/kubedb/mariadb-archiver/commit/c7a8956f) Test against k8s 1.33.2 (#56)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.38.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.38.0)

- [fc568c63](https://github.com/kubedb/mariadb-coordinator/commit/fc568c63) Prepare for release v0.38.0 (#150)
- [41e55cd0](https://github.com/kubedb/mariadb-coordinator/commit/41e55cd0) Add gRPC Server (#148)
- [91446b3a](https://github.com/kubedb/mariadb-coordinator/commit/91446b3a) Use Go 1.25 (#149)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.18.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.18.0)

- [fc72c621](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/fc72c621) Prepare for release v0.18.0 (#53)
- [75e657af](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/75e657af) Use Go 1.25 (#52)
- [416573f1](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/416573f1) Test against k8s 1.33.2 (#51)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.16.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.16.0)

- [dc88d18](https://github.com/kubedb/mariadb-restic-plugin/commit/dc88d18) Prepare for release v0.16.0 (#53)
- [7e5afc1](https://github.com/kubedb/mariadb-restic-plugin/commit/7e5afc1) Use Go 1.25 (#52)
- [a918a98](https://github.com/kubedb/mariadb-restic-plugin/commit/a918a98) Test against k8s 1.33.2 (#51)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.51.0](https://github.com/kubedb/memcached/releases/tag/v0.51.0)

- [dec0b55c](https://github.com/kubedb/memcached/commit/dec0b55c9) Prepare for release v0.51.0 (#508)
- [e5c0f387](https://github.com/kubedb/memcached/commit/e5c0f387c) Use Go 1.25 (#506)
- [0d4493d8](https://github.com/kubedb/memcached/commit/0d4493d87) Test against k8s 1.33.2 (#505)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.51.0](https://github.com/kubedb/mongodb/releases/tag/v0.51.0)

- [34997974](https://github.com/kubedb/mongodb/commit/34997974c) Prepare for release v0.51.0 (#716)
- [ea449c6e](https://github.com/kubedb/mongodb/commit/ea449c6ef) Update for gitsync const
- [c1a02a2c](https://github.com/kubedb/mongodb/commit/c1a02a2c7) Set domain (#715)
- [1b50cb25](https://github.com/kubedb/mongodb/commit/1b50cb253) Use Go 1.25 (#714)
- [d524b37d](https://github.com/kubedb/mongodb/commit/d524b37d1) Test against k8s 1.33.2 (#713)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.19.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.19.0)

- [9b2a9e52](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/9b2a9e52) Prepare for release v0.19.0 (#57)
- [9869a61a](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/9869a61a) Use Go 1.25 (#56)
- [7e8ab012](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/7e8ab012) Test against k8s 1.33.2 (#55)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.21.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.21.0)

- [3ab2ede4](https://github.com/kubedb/mongodb-restic-plugin/commit/3ab2ede4) Prepare for release v0.21.0 (#91)
- [a69b17df](https://github.com/kubedb/mongodb-restic-plugin/commit/a69b17df) Use Go 1.25 (#90)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.13.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.13.0)

- [778ee906](https://github.com/kubedb/mssql-coordinator/commit/778ee906) Prepare for release v0.13.0 (#43)
- [e4ec2ecb](https://github.com/kubedb/mssql-coordinator/commit/e4ec2ecb) Use Go 1.25 (#42)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.13.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.13.0)

- [0dd2b786](https://github.com/kubedb/mssqlserver/commit/0dd2b786) Prepare for release v0.13.0 (#89)
- [2dbba7e0](https://github.com/kubedb/mssqlserver/commit/2dbba7e0) Set domain (#88)
- [c7907819](https://github.com/kubedb/mssqlserver/commit/c7907819) Use Go 1.25 (#87)
- [f08da8ea](https://github.com/kubedb/mssqlserver/commit/f08da8ea) Test against k8s 1.33.2 (#86)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.12.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.12.0)

- [6eb46c0](https://github.com/kubedb/mssqlserver-archiver/commit/6eb46c0) Use Go 1.25 (#13)



## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.12.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.12.0)

- [4bf6ee1](https://github.com/kubedb/mssqlserver-walg-plugin/commit/4bf6ee1) Prepare for release v0.12.0 (#32)
- [85d2ed4](https://github.com/kubedb/mssqlserver-walg-plugin/commit/85d2ed4) Use Go 1.25 (#31)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.51.0](https://github.com/kubedb/mysql/releases/tag/v0.51.0)

- [58693144](https://github.com/kubedb/mysql/commit/586931448) Prepare for release v0.51.0 (#699)
- [f24ca971](https://github.com/kubedb/mysql/commit/f24ca9719) Adjust git-sync mountpath constant (#698)
- [b0b7b95d](https://github.com/kubedb/mysql/commit/b0b7b95d5) Ignore notFound error on monitor deletion (#697)
- [2ff5ec27](https://github.com/kubedb/mysql/commit/2ff5ec27d) Add Ops-Manager to CI (#686)
- [73cc0a5f](https://github.com/kubedb/mysql/commit/73cc0a5f1) Use Go 1.25 (#696)
- [b8d33dbb](https://github.com/kubedb/mysql/commit/b8d33dbb4) Test against k8s 1.33.2 (#695)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.19.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.19.0)

- [c03b1853](https://github.com/kubedb/mysql-archiver/commit/c03b1853) Prepare for release v0.19.0 (#66)
- [54de850e](https://github.com/kubedb/mysql-archiver/commit/54de850e) Use Go 1.25 (#65)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.36.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.36.0)

- [44aeb425](https://github.com/kubedb/mysql-coordinator/commit/44aeb425) Prepare for release v0.36.0 (#150)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.19.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.19.0)

- [64cc16d7](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/64cc16d7) Prepare for release v0.19.0 (#53)
- [259c5775](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/259c5775) Use Go 1.25 (#52)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.21.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.21.0)

- [e06b074](https://github.com/kubedb/mysql-restic-plugin/commit/e06b074) Prepare for release v0.21.0 (#81)
- [772df9c](https://github.com/kubedb/mysql-restic-plugin/commit/772df9c) Use Go 1.25 (#80)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.36.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.36.0)

- [b75f385](https://github.com/kubedb/mysql-router-init/commit/b75f385) Use Go 1.25 (#52)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.45.0](https://github.com/kubedb/ops-manager/releases/tag/v0.45.0)

- [f9af4d9e](https://github.com/kubedb/ops-manager/commit/f9af4d9e0) Prepare for release v0.45.0 (#790)
- [732c00e2](https://github.com/kubedb/ops-manager/commit/732c00e26) Hazelcast rotate auth, reconfigure TLS, Recommendation Engine (#777)
- [6c5afbe1](https://github.com/kubedb/ops-manager/commit/6c5afbe1a) Distributed MariaDB TLS reconfigure, version upgrade, rotate auth (#779)
- [bf00b3f4](https://github.com/kubedb/ops-manager/commit/bf00b3f4e) Recommandation for PgBouncer and Pgpool (#785)
- [c80e6e44](https://github.com/kubedb/ops-manager/commit/c80e6e447) Add Clickhouse Rotate-Auth,Reconfigure-TLS (#776)
- [13f5085b](https://github.com/kubedb/ops-manager/commit/13f5085b2) Set Domain properly (#788)
- [ac400d32](https://github.com/kubedb/ops-manager/commit/ac400d329) Update Offline Volume Expansion flow  (#786)
- [7bc3a659](https://github.com/kubedb/ops-manager/commit/7bc3a6591) Update cluster.local (#787)



## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.4.0](https://github.com/kubedb/oracle/releases/tag/v0.4.0)

- [f0c77e1f](https://github.com/kubedb/oracle/commit/f0c77e1f) Prepare for release v0.4.0 (#12)
- [ba6a9bb1](https://github.com/kubedb/oracle/commit/ba6a9bb1) Use domain (#11)
- [95008420](https://github.com/kubedb/oracle/commit/95008420) Use Go 1.25 (#10)
- [bd56d458](https://github.com/kubedb/oracle/commit/bd56d458) Test against k8s 1.33.2 (#9)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.4.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.4.0)

- [79dc50e](https://github.com/kubedb/oracle-coordinator/commit/79dc50e) Prepare for release v0.4.0 (#9)
- [d277aa4](https://github.com/kubedb/oracle-coordinator/commit/d277aa4) Use domain utils (#8)
- [8f23d88](https://github.com/kubedb/oracle-coordinator/commit/8f23d88) Use Go 1.25 (#7)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.45.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.45.0)

- [01f946c7](https://github.com/kubedb/percona-xtradb/commit/01f946c76) Prepare for release v0.45.0 (#416)
- [b5b0681a](https://github.com/kubedb/percona-xtradb/commit/b5b0681ae) Ignore notFound error on monitor deletion (#415)
- [46e3394f](https://github.com/kubedb/percona-xtradb/commit/46e3394f9) Use Go 1.25 (#414)
- [b9795106](https://github.com/kubedb/percona-xtradb/commit/b9795106a) Test against k8s 1.33.2 (#413)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.31.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.31.0)

- [72259ff7](https://github.com/kubedb/percona-xtradb-coordinator/commit/72259ff7) Prepare for release v0.31.0 (#101)
- [176770ef](https://github.com/kubedb/percona-xtradb-coordinator/commit/176770ef) Use Go 1.25 (#100)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.42.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.42.0)

- [f95f4f7a](https://github.com/kubedb/pg-coordinator/commit/f95f4f7a) Prepare for release v0.42.0 (#209)
- [497bd29b](https://github.com/kubedb/pg-coordinator/commit/497bd29b) Use FindDomain function (#208)
- [c6edd36e](https://github.com/kubedb/pg-coordinator/commit/c6edd36e) Use Go 1.25 (#206)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.45.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.45.0)

- [50f78972](https://github.com/kubedb/pgbouncer/commit/50f78972) Prepare for release v0.45.0 (#380)
- [f37f96fb](https://github.com/kubedb/pgbouncer/commit/f37f96fb) Add Init script and script from git-sync (#379)
- [69c72451](https://github.com/kubedb/pgbouncer/commit/69c72451) Use Go 1.25 (#378)
- [04e8bbe5](https://github.com/kubedb/pgbouncer/commit/04e8bbe5) Test against k8s 1.33.2 (#377)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.13.0](https://github.com/kubedb/pgpool/releases/tag/v0.13.0)

- [9d3edd94](https://github.com/kubedb/pgpool/commit/9d3edd94) Prepare for release v0.13.0 (#83)
- [535d110b](https://github.com/kubedb/pgpool/commit/535d110b) Use Go 1.25 (#81)
- [92925a00](https://github.com/kubedb/pgpool/commit/92925a00) Test against k8s 1.33.2 (#80)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.58.0](https://github.com/kubedb/postgres/releases/tag/v0.58.0)

- [67233a87](https://github.com/kubedb/postgres/commit/67233a879) Prepare for release v0.58.0 (#831)
- [d26d4d40](https://github.com/kubedb/postgres/commit/d26d4d404) Incorporate new petset api changes (#830)
- [fb69ef80](https://github.com/kubedb/postgres/commit/fb69ef802) Use FindDomain Func (#829)
- [a050ff26](https://github.com/kubedb/postgres/commit/a050ff263) Fix arbiter storage spec (#826)
- [46c13e2f](https://github.com/kubedb/postgres/commit/46c13e2f1) Use Go 1.25 (#828)
- [c6cae10d](https://github.com/kubedb/postgres/commit/c6cae10d8) Test against k8s 1.33.2 (#827)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.19.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.19.0)

- [29ae7ceb](https://github.com/kubedb/postgres-archiver/commit/29ae7ceb) Prepare for release v0.19.0 (#67)
- [4df4f154](https://github.com/kubedb/postgres-archiver/commit/4df4f154) Use Go 1.25 (#66)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.19.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.19.0)

- [518abd3c](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/518abd3c) Prepare for release v0.19.0 (#63)
- [f95399ad](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/f95399ad) Use Go 1.25 (#62)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.21.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.21.0)

- [0ebc088](https://github.com/kubedb/postgres-restic-plugin/commit/0ebc088) Prepare for release v0.21.0 (#78)
- [5e1f68b](https://github.com/kubedb/postgres-restic-plugin/commit/5e1f68b) Use Go 1.25 (#77)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.19.0](https://github.com/kubedb/provider-aws/releases/tag/v0.19.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.19.0](https://github.com/kubedb/provider-azure/releases/tag/v0.19.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.19.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.19.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.58.0](https://github.com/kubedb/provisioner/releases/tag/v0.58.0)

- [f5ae95b1](https://github.com/kubedb/provisioner/commit/f5ae95b19) Prepare for release v0.58.0 (#163)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.45.0](https://github.com/kubedb/proxysql/releases/tag/v0.45.0)

- [a63df178](https://github.com/kubedb/proxysql/commit/a63df1787) Prepare for release v0.45.0 (#402)
- [eaf0e003](https://github.com/kubedb/proxysql/commit/eaf0e0031) Ignore not found error on deletion (#401)
- [1aa6b519](https://github.com/kubedb/proxysql/commit/1aa6b5196) Use Go 1.25 (#400)
- [7d08c5ab](https://github.com/kubedb/proxysql/commit/7d08c5ab6) Test against k8s 1.33.2 (#399)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.13.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.13.0)

- [852448e6](https://github.com/kubedb/rabbitmq/commit/852448e6) Prepare for release v0.13.0 (#96)
- [5865b721](https://github.com/kubedb/rabbitmq/commit/5865b721) Use Go 1.25 (#95)
- [4075bef9](https://github.com/kubedb/rabbitmq/commit/4075bef9) Run Provisioner Tests On Monday + Friday, Update MakeFile (#74)
- [5580dc0e](https://github.com/kubedb/rabbitmq/commit/5580dc0e) set domain (#94)
- [91720f8f](https://github.com/kubedb/rabbitmq/commit/91720f8f) Disable mtls by default (#86)
- [446d6729](https://github.com/kubedb/rabbitmq/commit/446d6729) Use Go 1.25 (#93)
- [9841c20a](https://github.com/kubedb/rabbitmq/commit/9841c20a) Test against k8s 1.33.2 (#92)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.51.0](https://github.com/kubedb/redis/releases/tag/v0.51.0)

- [dc60b5d3](https://github.com/kubedb/redis/commit/dc60b5d38) Prepare for release v0.51.0 (#603)
- [d48fd534](https://github.com/kubedb/redis/commit/d48fd5348) Redis git-sync (#602)
- [8a463626](https://github.com/kubedb/redis/commit/8a4636265) Use Go 1.25 (#601)
- [6203f72e](https://github.com/kubedb/redis/commit/6203f72ec) Test against k8s 1.33.2 (#600)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.37.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.37.0)

- [8355b01d](https://github.com/kubedb/redis-coordinator/commit/8355b01d) Prepare for release v0.37.0 (#134)
- [db2b37f0](https://github.com/kubedb/redis-coordinator/commit/db2b37f0) Use Go 1.25 (#133)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.21.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.21.0)

- [227f7aa](https://github.com/kubedb/redis-restic-plugin/commit/227f7aa) Prepare for release v0.21.0 (#73)
- [7cb3f3b](https://github.com/kubedb/redis-restic-plugin/commit/7cb3f3b) Use Go 1.25 (#72)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.45.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.45.0)

- [f7ac75e5](https://github.com/kubedb/replication-mode-detector/commit/f7ac75e5) Prepare for release v0.45.0 (#298)
- [6b506280](https://github.com/kubedb/replication-mode-detector/commit/6b506280) Use Go 1.25 (#297)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.34.0](https://github.com/kubedb/schema-manager/releases/tag/v0.34.0)

- [a9ed20f3](https://github.com/kubedb/schema-manager/commit/a9ed20f3) Prepare for release v0.34.0 (#145)
- [baafc879](https://github.com/kubedb/schema-manager/commit/baafc879) Use Go 1.25 (#144)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.13.0](https://github.com/kubedb/singlestore/releases/tag/v0.13.0)

- [c0e6f76a](https://github.com/kubedb/singlestore/commit/c0e6f76a) Prepare for release v0.13.0 (#82)
- [422669e4](https://github.com/kubedb/singlestore/commit/422669e4) Use Go 1.25 (#81)
- [ca01869b](https://github.com/kubedb/singlestore/commit/ca01869b) Test against k8s 1.33.2 (#80)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.13.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.13.0)

- [f0dbf80](https://github.com/kubedb/singlestore-coordinator/commit/f0dbf80) Prepare for release v0.13.0 (#48)
- [3ea792a](https://github.com/kubedb/singlestore-coordinator/commit/3ea792a) Use Go 1.25 (#47)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.16.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.16.0)

- [b8cd8be](https://github.com/kubedb/singlestore-restic-plugin/commit/b8cd8be) Prepare for release v0.16.0 (#52)
- [3835009](https://github.com/kubedb/singlestore-restic-plugin/commit/3835009) Use Go 1.25 (#51)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.13.0](https://github.com/kubedb/solr/releases/tag/v0.13.0)

- [6def247b](https://github.com/kubedb/solr/commit/6def247b) Prepare for release v0.13.0 (#96)
- [2460042e](https://github.com/kubedb/solr/commit/2460042e) Replaced hardcoded cluster.local with dynamic cluster domain (#95)
- [beda554f](https://github.com/kubedb/solr/commit/beda554f) Fix Delete Monitor  (#94)
- [ac90d42b](https://github.com/kubedb/solr/commit/ac90d42b) Use Go 1.25 (#93)
- [e94509e2](https://github.com/kubedb/solr/commit/e94509e2) Test against k8s 1.33.2 (#92)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.43.0](https://github.com/kubedb/tests/releases/tag/v0.43.0)

- [c20305de](https://github.com/kubedb/tests/commit/c20305de) Prepare for release v0.43.0 (#478)
- [655060a3](https://github.com/kubedb/tests/commit/655060a3) MSSQLServer Ops Request Tests (#461)
- [07637efe](https://github.com/kubedb/tests/commit/07637efe) Use Go 1.25 (#476)
- [02756ec1](https://github.com/kubedb/tests/commit/02756ec1) Test against k8s 1.33.2 (#475)
- [ba7c73ac](https://github.com/kubedb/tests/commit/ba7c73ac) MySQL TLS Configuration Fix for 5.7.44 (#471)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.34.0](https://github.com/kubedb/ui-server/releases/tag/v0.34.0)

- [e7d6c916](https://github.com/kubedb/ui-server/commit/e7d6c916) Prepare for release v0.34.0 (#172)
- [affa1102](https://github.com/kubedb/ui-server/commit/affa1102) Set Domain (#171)
- [bad1998c](https://github.com/kubedb/ui-server/commit/bad1998c) Use Go 1.25 (#170)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.34.0](https://github.com/kubedb/webhook-server/releases/tag/v0.34.0)

- [c8b29cff](https://github.com/kubedb/webhook-server/commit/c8b29cff) Prepare for release v0.34.0 (#168)
- [28888a41](https://github.com/kubedb/webhook-server/commit/28888a41) Use Go 1.25 (#167)
- [c8f19b6a](https://github.com/kubedb/webhook-server/commit/c8f19b6a) Test against k8s 1.33.2 (#166)



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.7.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.7.0)

- [999e4a0](https://github.com/kubedb/xtrabackup-restic-plugin/commit/999e4a0) Prepare for release v0.7.0 (#20)
- [b90b267](https://github.com/kubedb/xtrabackup-restic-plugin/commit/b90b267) Use Go 1.25 (#19)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.13.0](https://github.com/kubedb/zookeeper/releases/tag/v0.13.0)

- [1242283e](https://github.com/kubedb/zookeeper/commit/1242283e) Prepare for release v0.13.0 (#86)
- [350e875e](https://github.com/kubedb/zookeeper/commit/350e875e) Ignore notFound error on monitor deletion (#85)
- [c0b52734](https://github.com/kubedb/zookeeper/commit/c0b52734) Use Go 1.25 (#84)
- [4cc59e87](https://github.com/kubedb/zookeeper/commit/4cc59e87) Test against k8s 1.33.2 (#83)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.14.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.14.0)

- [ed4e17b](https://github.com/kubedb/zookeeper-restic-plugin/commit/ed4e17b) Prepare for release v0.14.0 (#42)
- [2d04379](https://github.com/kubedb/zookeeper-restic-plugin/commit/2d04379) Use Go 1.25 (#41)




