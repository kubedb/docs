---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2026.6.5-rc.1
    name: Changelog-v2026.6.5-rc.1
    parent: welcome
    weight: 20260605
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.6.5-rc.1/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.6.5-rc.1/
---

# KubeDB v2026.6.5-rc.1 (2026-06-05)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.65.0-rc.1](https://github.com/kubedb/apimachinery/releases/tag/v0.65.0-rc.1)

- [23c81b3b](https://github.com/kubedb/apimachinery/commit/23c81b3b6) Fix go.mod
- [007ef3ea](https://github.com/kubedb/apimachinery/commit/007ef3ea6) Remove FerretDB support (#1750)
- [101341c2](https://github.com/kubedb/apimachinery/commit/101341c29) Added wv-ops-validation and autoscaling webhook (#1722)
- [fb439fff](https://github.com/kubedb/apimachinery/commit/fb439fff4) Register Neo4j Autoscaler (#1746)
- [62f69e28](https://github.com/kubedb/apimachinery/commit/62f69e286) Add CiliumNetworkPolicy flavor support (#1714)
- [9b1a220b](https://github.com/kubedb/apimachinery/commit/9b1a220b4) Bump go.bytebuilders.dev/audit to v0.0.52 (#1745)
- [dd3d0e3d](https://github.com/kubedb/apimachinery/commit/dd3d0e3d3) Update go.bytebuilders.dev/audit to v0.0.51 (#1744)
- [0e48a4da](https://github.com/kubedb/apimachinery/commit/0e48a4da6) Add common configuration for milvus storage migration (#1741)
- [ea56f8dd](https://github.com/kubedb/apimachinery/commit/ea56f8dd2) Add pkg/secret helpers for dual-path auth secret access (#1726)
- [ef57da77](https://github.com/kubedb/apimachinery/commit/ef57da777) Fix DatabaseConfiguration resource name casing and remove stale CRDs (#1727)
- [9bdcc3b7](https://github.com/kubedb/apimachinery/commit/9bdcc3b7e) DatabaseInfo -> DatabaseConfiguration (#1725)
- [66b31784](https://github.com/kubedb/apimachinery/commit/66b31784e) Introduce summary api (#1724)
- [59b4eaa3](https://github.com/kubedb/apimachinery/commit/59b4eaa36) Add weaviate ops helpers
- [f7c72195](https://github.com/kubedb/apimachinery/commit/f7c721957) Review autoscaler api (#1719)
- [97db127e](https://github.com/kubedb/apimachinery/commit/97db127e6) Update Elasticsearch StorageMigration Ops Api (#1721)
- [7b38d914](https://github.com/kubedb/apimachinery/commit/7b38d9144) Add Milvus Ops Request (#1666)
- [46001228](https://github.com/kubedb/apimachinery/commit/46001228f) add weaviate ops_req (#1716)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.50.0-rc.1](https://github.com/kubedb/autoscaler/releases/tag/v0.50.0-rc.1)

- [9d983c07](https://github.com/kubedb/autoscaler/commit/9d983c07) Prepare for release v0.50.0-rc.1 (#305)
- [9d7909f8](https://github.com/kubedb/autoscaler/commit/9d7909f8) Remove FerretDB support (#304)
- [9366ad79](https://github.com/kubedb/autoscaler/commit/9366ad79) Set Ops Request Options Default (#293)
- [587219ce](https://github.com/kubedb/autoscaler/commit/587219ce) Add Qdrant Autoscaler (#285)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.18.0-rc.1](https://github.com/kubedb/cassandra/releases/tag/v0.18.0-rc.1)

- [4855098e](https://github.com/kubedb/cassandra/commit/4855098e) Prepare for release v0.18.0-rc.1 (#87)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.12.0-rc.1](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.12.0-rc.1)

- [aefdc390](https://github.com/kubedb/cassandra-medusa-plugin/commit/aefdc390) Prepare for release v0.12.0-rc.1 (#39)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.20.0-rc.1](https://github.com/kubedb/clickhouse/releases/tag/v0.20.0-rc.1)

- [18935285](https://github.com/kubedb/clickhouse/commit/18935285) Prepare for release v0.20.0-rc.1 (#112)



## [kubedb/clickhouse-backup-plugin](https://github.com/kubedb/clickhouse-backup-plugin)

### [v0.2.0-rc.1](https://github.com/kubedb/clickhouse-backup-plugin/releases/tag/v0.2.0-rc.1)

- [8ee159c7](https://github.com/kubedb/clickhouse-backup-plugin/commit/8ee159c7) Prepare for release v0.2.0-rc.1 (#25)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.20.0-rc.1](https://github.com/kubedb/crd-manager/releases/tag/v0.20.0-rc.1)

- [f63e3c4d](https://github.com/kubedb/crd-manager/commit/f63e3c4d) Prepare for release v0.20.0-rc.1 (#139)
- [33485516](https://github.com/kubedb/crd-manager/commit/33485516) Remove FerretDB support (#138)
- [f4f684e9](https://github.com/kubedb/crd-manager/commit/f4f684e9) Add all missing CRDs (#137)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.23.0-rc.1](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.23.0-rc.1)

- [dbe58a1d](https://github.com/kubedb/dashboard-restic-plugin/commit/dbe58a1d) Prepare for release v0.23.0-rc.1 (#77)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.20.0-rc.1](https://github.com/kubedb/db-client-go/releases/tag/v0.20.0-rc.1)

- [a24ee1cd](https://github.com/kubedb/db-client-go/commit/a24ee1cd) Prepare for release v0.20.0-rc.1 (#247)
- [0efcbce1](https://github.com/kubedb/db-client-go/commit/0efcbce1) Use shared pkg/secret helpers for dual-path auth secret access (#245)
- [0878c362](https://github.com/kubedb/db-client-go/commit/0878c362) Bump kubedb.dev/apimachinery to drop FerretDB (#246)
- [523e0304](https://github.com/kubedb/db-client-go/commit/523e0304) Tighten CI/release workflow secrets, perms, and release notes



## [kubedb/db2](https://github.com/kubedb/db2)

### [v0.6.0-rc.1](https://github.com/kubedb/db2/releases/tag/v0.6.0-rc.1)

- [deec0ca0](https://github.com/kubedb/db2/commit/deec0ca0) Prepare for release v0.6.0-rc.1 (#29)



## [kubedb/db2-coordinator](https://github.com/kubedb/db2-coordinator)

### [v0.6.0-rc.1](https://github.com/kubedb/db2-coordinator/releases/tag/v0.6.0-rc.1)




## [kubedb/documentdb](https://github.com/kubedb/documentdb)

### [v0.2.0-rc.1](https://github.com/kubedb/documentdb/releases/tag/v0.2.0-rc.1)

- [481e035d](https://github.com/kubedb/documentdb/commit/481e035d) Prepare for release v0.2.0-rc.1 (#18)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.20.0-rc.1](https://github.com/kubedb/druid/releases/tag/v0.20.0-rc.1)

- [7970cf8b](https://github.com/kubedb/druid/commit/7970cf8b) Prepare for release v0.20.0-rc.1 (#137)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.65.0-rc.1](https://github.com/kubedb/elasticsearch/releases/tag/v0.65.0-rc.1)

- [890e13c1](https://github.com/kubedb/elasticsearch/commit/890e13c16) Prepare for release v0.65.0-rc.1 (#817)
- [613e034e](https://github.com/kubedb/elasticsearch/commit/613e034ea) Add StorageMigration OpsRequest support (#810)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.28.0-rc.1](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.28.0-rc.1)

- [650b7682](https://github.com/kubedb/elasticsearch-restic-plugin/commit/650b7682) Prepare for release v0.28.0-rc.1 (#100)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.13.0-rc.1](https://github.com/kubedb/gitops/releases/tag/v0.13.0-rc.1)

- [0da911fa](https://github.com/kubedb/gitops/commit/0da911fa) Prepare for release v0.13.0-rc.1 (#68)



## [kubedb/hanadb](https://github.com/kubedb/hanadb)

### [v0.6.0-rc.1](https://github.com/kubedb/hanadb/releases/tag/v0.6.0-rc.1)

- [da2fb620](https://github.com/kubedb/hanadb/commit/da2fb620) Prepare for release v0.6.0-rc.1 (#43)



## [kubedb/hanadb-coordinator](https://github.com/kubedb/hanadb-coordinator)

### [v0.5.0-rc.1](https://github.com/kubedb/hanadb-coordinator/releases/tag/v0.5.0-rc.1)

- [97c79d78](https://github.com/kubedb/hanadb-coordinator/commit/97c79d78) Prepare for release v0.5.0-rc.1 (#14)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.11.0-rc.1](https://github.com/kubedb/hazelcast/releases/tag/v0.11.0-rc.1)

- [e45daa28](https://github.com/kubedb/hazelcast/commit/e45daa28) Prepare for release v0.11.0-rc.1 (#49)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.12.0-rc.1](https://github.com/kubedb/ignite/releases/tag/v0.12.0-rc.1)

- [46fc9cb0](https://github.com/kubedb/ignite/commit/46fc9cb0) Prepare for release v0.12.0-rc.1 (#58)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2026.6.5-rc.1](https://github.com/kubedb/installer/releases/tag/v2026.6.5-rc.1)

- [b06451d3](https://github.com/kubedb/installer/commit/b06451d33) Prepare for release v2026.6.5-rc.1 (#2337)
- [78341132](https://github.com/kubedb/installer/commit/78341132c) Add global.networkPolicy.flavor with cilium support (#2289)
- [6fc62ced](https://github.com/kubedb/installer/commit/6fc62ced1) Remove FerretDB support (#2335)
- [abc8e994](https://github.com/kubedb/installer/commit/abc8e994d) Add Qdrant and Ignite Autoscaler Validating and Mutating Webhook (#2260)
- [f3095296](https://github.com/kubedb/installer/commit/f30952964) Use ace-user-roles v2026.6.12 with audit cluster role (#2332)
- [67ef5f22](https://github.com/kubedb/installer/commit/67ef5f229) Add Milvus Webhook Validation (#2328)
- [0b68dd1e](https://github.com/kubedb/installer/commit/0b68dd1ee) update redis init (#2323)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.36.0-rc.1](https://github.com/kubedb/kafka/releases/tag/v0.36.0-rc.1)

- [a1068e0b](https://github.com/kubedb/kafka/commit/a1068e0b) Prepare for release v0.36.0-rc.1 (#200)
- [c6a0dee3](https://github.com/kubedb/kafka/commit/c6a0dee3) Add StorageMigration OpsRequest support (#194)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.41.0-rc.1](https://github.com/kubedb/kibana/releases/tag/v0.41.0-rc.1)

- [7bec93e8](https://github.com/kubedb/kibana/commit/7bec93e8) Prepare for release v0.41.0-rc.1 (#183)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.28.0-rc.1](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.28.0-rc.1)

- [5fe85111](https://github.com/kubedb/kubedb-manifest-plugin/commit/5fe85111) Prepare for release v0.28.0-rc.1 (#133)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.16.0-rc.1](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.16.0-rc.1)

- [bf5f23ea](https://github.com/kubedb/kubedb-verifier/commit/bf5f23ea) Prepare for release v0.16.0-rc.1 (#52)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.49.0-rc.1](https://github.com/kubedb/mariadb/releases/tag/v0.49.0-rc.1)

- [f0f7b7e0](https://github.com/kubedb/mariadb/commit/f0f7b7e0f) Prepare for release v0.49.0-rc.1 (#406)
- [1bb8a02a](https://github.com/kubedb/mariadb/commit/1bb8a02ac) Add StorageMigration OpsRequest support (#398)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.25.0-rc.1](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.25.0-rc.1)

- [06784371](https://github.com/kubedb/mariadb-archiver/commit/06784371) Prepare for release v0.25.0-rc.1 (#93)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.45.0-rc.1](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.45.0-rc.1)

- [1358936a](https://github.com/kubedb/mariadb-coordinator/commit/1358936a) Prepare for release v0.45.0-rc.1 (#180)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.25.0-rc.1](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.25.0-rc.1)

- [70b5cdd2](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/70b5cdd2) Prepare for release v0.25.0-rc.1 (#79)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.23.0-rc.1](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.23.0-rc.1)

- [3b08f7c6](https://github.com/kubedb/mariadb-restic-plugin/commit/3b08f7c6) Prepare for release v0.23.0-rc.1 (#91)



## [kubedb/migrator-cli](https://github.com/kubedb/migrator-cli)

### [v0.5.0-rc.1](https://github.com/kubedb/migrator-cli/releases/tag/v0.5.0-rc.1)

- [f64c95c](https://github.com/kubedb/migrator-cli/commit/f64c95c) Prepare for release v0.5.0-rc.1 (#23)



## [kubedb/migrator-operator](https://github.com/kubedb/migrator-operator)

### [v0.5.0-rc.1](https://github.com/kubedb/migrator-operator/releases/tag/v0.5.0-rc.1)

- [593fa3a](https://github.com/kubedb/migrator-operator/commit/593fa3a) Prepare for release v0.5.0-rc.1 (#18)



## [kubedb/milvus](https://github.com/kubedb/milvus)

### [v0.6.0-rc.1](https://github.com/kubedb/milvus/releases/tag/v0.6.0-rc.1)

- [703de898](https://github.com/kubedb/milvus/commit/703de898) Prepare for release v0.6.0-rc.1 (#42)
- [83890286](https://github.com/kubedb/milvus/commit/83890286) Add StorageMigration Ops-Request Support (#39)
- [718988e1](https://github.com/kubedb/milvus/commit/718988e1) Add Milvus Ops Request (#33)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.58.0-rc.1](https://github.com/kubedb/mongodb/releases/tag/v0.58.0-rc.1)

- [4591f101](https://github.com/kubedb/mongodb/commit/4591f1011) Prepare for release v0.58.0-rc.1 (#765)
- [0ec52968](https://github.com/kubedb/mongodb/commit/0ec529684) Add virtual secret support (#762)
- [2946c711](https://github.com/kubedb/mongodb/commit/2946c7119) Honor user-provided renewBefore in TLS certificate ops (#764)
- [75fb23b2](https://github.com/kubedb/mongodb/commit/75fb23b2d) Implement cilium networkpolicy (#763)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.26.0-rc.1](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.26.0-rc.1)

- [ec3b81a8](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/ec3b81a8) Prepare for release v0.26.0-rc.1 (#84)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.28.0-rc.1](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.28.0-rc.1)

- [a8f1b59b](https://github.com/kubedb/mongodb-restic-plugin/commit/a8f1b59b) Prepare for release v0.28.0-rc.1 (#129)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.20.0-rc.1](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.20.0-rc.1)

- [acc28e43](https://github.com/kubedb/mssql-coordinator/commit/acc28e43) Prepare for release v0.20.0-rc.1 (#72)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.20.0-rc.1](https://github.com/kubedb/mssqlserver/releases/tag/v0.20.0-rc.1)

- [1ec9a0f1](https://github.com/kubedb/mssqlserver/commit/1ec9a0f1) Prepare for release v0.20.0-rc.1 (#138)
- [0eb8b1c8](https://github.com/kubedb/mssqlserver/commit/0eb8b1c8) Add StorageMigration OpsRequest support for MSSQLServer (#132)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.19.0-rc.1](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.19.0-rc.1)




## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.19.0-rc.1](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.19.0-rc.1)

- [51ea4f2](https://github.com/kubedb/mssqlserver-walg-plugin/commit/51ea4f2) Prepare for release v0.19.0-rc.1 (#61)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.26.0-rc.1](https://github.com/kubedb/mysql-archiver/releases/tag/v0.26.0-rc.1)

- [33eaf828](https://github.com/kubedb/mysql-archiver/commit/33eaf828) Prepare for release v0.26.0-rc.1 (#108)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.43.0-rc.1](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.43.0-rc.1)

- [c813bf0e](https://github.com/kubedb/mysql-coordinator/commit/c813bf0e) Prepare for release v0.43.0-rc.1 (#181)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.26.0-rc.1](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.26.0-rc.1)

- [5c73df2a](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/5c73df2a) Prepare for release v0.26.0-rc.1 (#80)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.28.0-rc.1](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.28.0-rc.1)

- [8751f570](https://github.com/kubedb/mysql-restic-plugin/commit/8751f570) Prepare for release v0.28.0-rc.1 (#115)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.43.0-rc.1](https://github.com/kubedb/mysql-router-init/releases/tag/v0.43.0-rc.1)




## [kubedb/neo4j](https://github.com/kubedb/neo4j)

### [v0.6.0-rc.1](https://github.com/kubedb/neo4j/releases/tag/v0.6.0-rc.1)

- [c5dcc324](https://github.com/kubedb/neo4j/commit/c5dcc324) Prepare for release v0.6.0-rc.1 (#38)
- [0bf9da80](https://github.com/kubedb/neo4j/commit/0bf9da80) Add Backup port in primary service (#31)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.52.0-rc.1](https://github.com/kubedb/ops-manager/releases/tag/v0.52.0-rc.1)

- [93fc27d3](https://github.com/kubedb/ops-manager/commit/93fc27d38) Prepare for release v0.52.0-rc.1 (#868)
- [7b18fff4](https://github.com/kubedb/ops-manager/commit/7b18fff4c) Remove FerretDB support (#867)
- [9b4ef955](https://github.com/kubedb/ops-manager/commit/9b4ef9551) Add Recommendation Engine support for Neo4j (#857)
- [a0bcd0df](https://github.com/kubedb/ops-manager/commit/a0bcd0df2) Add Milvus Ops Request (#846)



## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.11.0-rc.1](https://github.com/kubedb/oracle/releases/tag/v0.11.0-rc.1)

- [789d3804](https://github.com/kubedb/oracle/commit/789d3804) Prepare for release v0.11.0-rc.1 (#58)
- [13c00d3e](https://github.com/kubedb/oracle/commit/13c00d3e) add ImagePullSecret for observer (#54)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.11.0-rc.1](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.11.0-rc.1)

- [1a66329](https://github.com/kubedb/oracle-coordinator/commit/1a66329) Prepare for release v0.11.0-rc.1 (#37)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.52.0-rc.1](https://github.com/kubedb/percona-xtradb/releases/tag/v0.52.0-rc.1)

- [db4f1d80](https://github.com/kubedb/percona-xtradb/commit/db4f1d809) Prepare for release v0.52.0-rc.1 (#459)
- [023944e1](https://github.com/kubedb/percona-xtradb/commit/023944e11) Add StorageMigration OpsRequest support for PerconaXtraDB (#452)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.38.0-rc.1](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.38.0-rc.1)

- [a0b5004e](https://github.com/kubedb/percona-xtradb-coordinator/commit/a0b5004e) Prepare for release v0.38.0-rc.1 (#128)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.49.0-rc.1](https://github.com/kubedb/pg-coordinator/releases/tag/v0.49.0-rc.1)

- [412aedcf](https://github.com/kubedb/pg-coordinator/commit/412aedcf) Prepare for release v0.49.0-rc.1 (#252)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.52.0-rc.1](https://github.com/kubedb/pgbouncer/releases/tag/v0.52.0-rc.1)

- [8e5905a4](https://github.com/kubedb/pgbouncer/commit/8e5905a4b) Prepare for release v0.52.0-rc.1 (#416)
- [c5e7a81e](https://github.com/kubedb/pgbouncer/commit/c5e7a81e0) Add --network-policy-flavor flag with cilium support (#415)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.20.0-rc.1](https://github.com/kubedb/pgpool/releases/tag/v0.20.0-rc.1)

- [1750b479](https://github.com/kubedb/pgpool/commit/1750b479) Prepare for release v0.20.0-rc.1 (#123)
- [eff93373](https://github.com/kubedb/pgpool/commit/eff93373) Add --network-policy-flavor flag with cilium support (#122)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.65.0-rc.1](https://github.com/kubedb/postgres/releases/tag/v0.65.0-rc.1)

- [da3dca05](https://github.com/kubedb/postgres/commit/da3dca059) Prepare for release v0.65.0-rc.1 (#894)
- [f203a907](https://github.com/kubedb/postgres/commit/f203a9079) Add --network-policy-flavor flag with cilium support (#886)
- [aa3c1fd5](https://github.com/kubedb/postgres/commit/aa3c1fd54) Honor user-provided renewBefore in TLS certificate ops (#893)
- [5d3d25d5](https://github.com/kubedb/postgres/commit/5d3d25d55) (Skip coordinator + Fix health check) for Remote Replica (#891)
- [4f7c4133](https://github.com/kubedb/postgres/commit/4f7c41333) Update cluster.local -> slice.local (#890)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.26.0-rc.1](https://github.com/kubedb/postgres-archiver/releases/tag/v0.26.0-rc.1)

- [10de04a8](https://github.com/kubedb/postgres-archiver/commit/10de04a8) Prepare for release v0.26.0-rc.1 (#109)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.26.0-rc.1](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.26.0-rc.1)

- [a578dc28](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/a578dc28) Prepare for release v0.26.0-rc.1 (#90)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.28.0-rc.1](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.28.0-rc.1)

- [229651a4](https://github.com/kubedb/postgres-restic-plugin/commit/229651a4) Prepare for release v0.28.0-rc.1 (#110)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.26.0-rc.1](https://github.com/kubedb/provider-aws/releases/tag/v0.26.0-rc.1)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.26.0-rc.1](https://github.com/kubedb/provider-azure/releases/tag/v0.26.0-rc.1)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.26.0-rc.1](https://github.com/kubedb/provider-gcp/releases/tag/v0.26.0-rc.1)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.65.0-rc.1](https://github.com/kubedb/provisioner/releases/tag/v0.65.0-rc.1)

- [ae1b2445](https://github.com/kubedb/provisioner/commit/ae1b24456) Prepare for release v0.65.0-rc.1 (#210)
- [eb7cffb1](https://github.com/kubedb/provisioner/commit/eb7cffb1d) Implement cilium networkpolicy (#207)
- [f2a28deb](https://github.com/kubedb/provisioner/commit/f2a28deba) Remove FerretDB support (#209)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.52.0-rc.1](https://github.com/kubedb/proxysql/releases/tag/v0.52.0-rc.1)

- [fe85fcd0](https://github.com/kubedb/proxysql/commit/fe85fcd09) Prepare for release v0.52.0-rc.1 (#438)



## [kubedb/qdrant](https://github.com/kubedb/qdrant)

### [v0.6.0-rc.1](https://github.com/kubedb/qdrant/releases/tag/v0.6.0-rc.1)

- [6ea63e53](https://github.com/kubedb/qdrant/commit/6ea63e53) Prepare for release v0.6.0-rc.1 (#46)
- [88c4b011](https://github.com/kubedb/qdrant/commit/88c4b011) Add --network-policy-flavor flag with cilium support (#45)
- [928e3450](https://github.com/kubedb/qdrant/commit/928e3450) Add StorageMigration OpsRequest support for Qdrant (#40)
- [b144b7b6](https://github.com/kubedb/qdrant/commit/b144b7b6) Add Reconfigure TLS (#37)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.20.0-rc.1](https://github.com/kubedb/rabbitmq/releases/tag/v0.20.0-rc.1)

- [d1511a11](https://github.com/kubedb/rabbitmq/commit/d1511a11) Prepare for release v0.20.0-rc.1 (#137)
- [b2d406fb](https://github.com/kubedb/rabbitmq/commit/b2d406fb) Add StorageMigration OpsRequest support for RabbitMQ (#132)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.58.0-rc.1](https://github.com/kubedb/redis/releases/tag/v0.58.0-rc.1)

- [ca8050ef](https://github.com/kubedb/redis/commit/ca8050ef7) Prepare for release v0.58.0-rc.1 (#649)
- [82e9701e](https://github.com/kubedb/redis/commit/82e9701e5) Add --network-policy-flavor flag with cilium support (#648)
- [7f56a8b9](https://github.com/kubedb/redis/commit/7f56a8b98) Health Check updated (#641)
- [7452ec78](https://github.com/kubedb/redis/commit/7452ec78a) Add StorageMigration OpsRequest support for Redis (#644)
- [21037f13](https://github.com/kubedb/redis/commit/21037f135) Merge ACL in reconfigure merger (#642)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.44.0-rc.1](https://github.com/kubedb/redis-coordinator/releases/tag/v0.44.0-rc.1)

- [b0fc9339](https://github.com/kubedb/redis-coordinator/commit/b0fc9339) Prepare for release v0.44.0-rc.1 (#161)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.28.0-rc.1](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.28.0-rc.1)

- [b2769079](https://github.com/kubedb/redis-restic-plugin/commit/b2769079) Prepare for release v0.28.0-rc.1 (#107)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.41.0-rc.1](https://github.com/kubedb/schema-manager/releases/tag/v0.41.0-rc.1)

- [3d0da9a2](https://github.com/kubedb/schema-manager/commit/3d0da9a2) Prepare for release v0.41.0-rc.1 (#171)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.20.0-rc.1](https://github.com/kubedb/singlestore/releases/tag/v0.20.0-rc.1)

- [0ac62594](https://github.com/kubedb/singlestore/commit/0ac62594) Prepare for release v0.20.0-rc.1 (#127)
- [9860ed42](https://github.com/kubedb/singlestore/commit/9860ed42) Add --network-policy-flavor flag with cilium support (#126)
- [9b843f68](https://github.com/kubedb/singlestore/commit/9b843f68) Add StorageMigration OpsRequest support for Singlestore (#120)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.20.0-rc.1](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.20.0-rc.1)

- [ae11d990](https://github.com/kubedb/singlestore-coordinator/commit/ae11d990) Prepare for release v0.20.0-rc.1 (#73)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.23.0-rc.1](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.23.0-rc.1)

- [d6832394](https://github.com/kubedb/singlestore-restic-plugin/commit/d6832394) Prepare for release v0.23.0-rc.1 (#86)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.20.0-rc.1](https://github.com/kubedb/solr/releases/tag/v0.20.0-rc.1)

- [0d6417ef](https://github.com/kubedb/solr/commit/0d6417ef) Prepare for release v0.20.0-rc.1 (#133)
- [4f32f195](https://github.com/kubedb/solr/commit/4f32f195) Add --network-policy-flavor flag with cilium support (#132)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.50.0-rc.1](https://github.com/kubedb/tests/releases/tag/v0.50.0-rc.1)

- [59d384bc](https://github.com/kubedb/tests/commit/59d384bc4) Prepare for release v0.50.0-rc.1 (#538)
- [d5a306ba](https://github.com/kubedb/tests/commit/d5a306baf) Remove FerretDB support (#537)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.41.0-rc.1](https://github.com/kubedb/ui-server/releases/tag/v0.41.0-rc.1)

- [62d7f163](https://github.com/kubedb/ui-server/commit/62d7f163b) Prepare for release v0.41.0-rc.1 (#208)
- [5fa49a75](https://github.com/kubedb/ui-server/commit/5fa49a752) Remove FerretDB support (#207)



## [kubedb/weaviate](https://github.com/kubedb/weaviate)

### [v0.6.0-rc.1](https://github.com/kubedb/weaviate/releases/tag/v0.6.0-rc.1)

- [513e4bf4](https://github.com/kubedb/weaviate/commit/513e4bf4) Prepare for release v0.6.0-rc.1 (#39)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.41.0-rc.1](https://github.com/kubedb/webhook-server/releases/tag/v0.41.0-rc.1)

- [29608ec8](https://github.com/kubedb/webhook-server/commit/29608ec8b) Prepare for release v0.41.0-rc.1 (#220)
- [9baa5eb5](https://github.com/kubedb/webhook-server/commit/9baa5eb56) Remove FerretDB support (#219)
- [0f68066f](https://github.com/kubedb/webhook-server/commit/0f68066fa) Setup Qdrant Autoscaler Webhook (#209)
- [fe08c7c4](https://github.com/kubedb/webhook-server/commit/fe08c7c4d) Add Milvus OPS Webhook (#217)



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.13.0-rc.1](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.13.0-rc.1)

- [36817455](https://github.com/kubedb/xtrabackup-restic-plugin/commit/36817455) Prepare for release v0.13.0-rc.1 (#55)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.20.0-rc.1](https://github.com/kubedb/zookeeper/releases/tag/v0.20.0-rc.1)

- [b3e8b477](https://github.com/kubedb/zookeeper/commit/b3e8b477) Prepare for release v0.20.0-rc.1 (#125)
- [37106c9c](https://github.com/kubedb/zookeeper/commit/37106c9c) Add --network-policy-flavor flag with cilium support (#124)
- [dca38f13](https://github.com/kubedb/zookeeper/commit/dca38f13) Use PatchStatus instead of CreateOrPatch to avoid timing issues on db deletion (#122)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.20.0-rc.1](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.20.0-rc.1)

- [b5ff2162](https://github.com/kubedb/zookeeper-restic-plugin/commit/b5ff2162) Prepare for release v0.20.0-rc.1 (#71)




