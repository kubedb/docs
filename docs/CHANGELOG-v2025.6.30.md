---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2025.6.30
    name: Changelog-v2025.6.30
    parent: welcome
    weight: 20250630
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2025.6.30/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2025.6.30/
---

# KubeDB v2025.6.30 (2025-07-02)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.56.0](https://github.com/kubedb/apimachinery/releases/tag/v0.56.0)

- [07b6689c](https://github.com/kubedb/apimachinery/commit/07b6689c9) Update for release KubeStash@v2025.6.30 (#1482)
- [d083581f](https://github.com/kubedb/apimachinery/commit/d083581f1) fix ch-service (#1481)
- [f2e464a5](https://github.com/kubedb/apimachinery/commit/f2e464a5b) Add Redis `Announce` API Fields (#1465)
- [117e2306](https://github.com/kubedb/apimachinery/commit/117e2306d) Allow scaling down Postgres Database to Standalone (#1479)
- [bd2271f5](https://github.com/kubedb/apimachinery/commit/bd2271f5e) Add SQL Server Distributed AG APIs (#1470)
- [68c02c68](https://github.com/kubedb/apimachinery/commit/68c02c68a) make gen fmt (#1480)
- [141bf9a0](https://github.com/kubedb/apimachinery/commit/141bf9a0d) Add ignite ops (#1478)
- [70801591](https://github.com/kubedb/apimachinery/commit/70801591a) Add ClickHouse TLS Support (#1471)
- [1dcd8005](https://github.com/kubedb/apimachinery/commit/1dcd80052) Add hazelcast ops api (#1475)
- [7eff4b3c](https://github.com/kubedb/apimachinery/commit/7eff4b3c5) Add AutoOps & updateConstraints field for missing DBs (#1477)
- [8ce5da48](https://github.com/kubedb/apimachinery/commit/8ce5da487) Add mariadb maxscale field for supporting vertical scaling (#1472)
- [72ef63b9](https://github.com/kubedb/apimachinery/commit/72ef63b9c) Add Kafka Init Container Details from v4.0.0 (#1473)
- [2f14197a](https://github.com/kubedb/apimachinery/commit/2f14197a9) Add support for Cassandra TLS and ops requests (#1463)
- [97c18a62](https://github.com/kubedb/apimachinery/commit/97c18a62d) Fix pre-conditions in db validator (#1476)
- [66964219](https://github.com/kubedb/apimachinery/commit/669642194) Remove LockToDefault from Default (#1469)
- [cd795a57](https://github.com/kubedb/apimachinery/commit/cd795a57b) Update openapi schema (#1468)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.41.0](https://github.com/kubedb/autoscaler/releases/tag/v0.41.0)

- [f372b37f](https://github.com/kubedb/autoscaler/commit/f372b37f) Prepare for release v0.41.0 (#251)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.9.0](https://github.com/kubedb/cassandra/releases/tag/v0.9.0)

- [c702826b](https://github.com/kubedb/cassandra/commit/c702826b) Prepare for release v0.9.0 (#38)
- [433299d7](https://github.com/kubedb/cassandra/commit/433299d7) Add TLS support (#35)
- [e10d1904](https://github.com/kubedb/cassandra/commit/e10d1904) Remove verbosity flag (#37)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.3.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.3.0)

- [794a7a2](https://github.com/kubedb/cassandra-medusa-plugin/commit/794a7a2) Prepare for release v0.3.0 (#6)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.56.0](https://github.com/kubedb/cli/releases/tag/v0.56.0)

- [0eab26b0](https://github.com/kubedb/cli/commit/0eab26b04) Prepare for release v0.56.0 (#798)
- [a99fd338](https://github.com/kubedb/cli/commit/a99fd338a) Add mssql server dag config command (#797)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.11.0](https://github.com/kubedb/clickhouse/releases/tag/v0.11.0)

- [196b4433](https://github.com/kubedb/clickhouse/commit/196b4433) Prepare for release v0.11.0 (#55)
- [079952e5](https://github.com/kubedb/clickhouse/commit/079952e5) Add TLS (#54)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.11.0](https://github.com/kubedb/crd-manager/releases/tag/v0.11.0)

- [cde2301d](https://github.com/kubedb/crd-manager/commit/cde2301d) Prepare for release v0.11.0 (#83)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.14.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.14.0)

- [2adcbef](https://github.com/kubedb/dashboard-restic-plugin/commit/2adcbef) Prepare for release v0.14.0 (#40)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.11.0](https://github.com/kubedb/db-client-go/releases/tag/v0.11.0)

- [3c7982b7](https://github.com/kubedb/db-client-go/commit/3c7982b7) Prepare for release v0.11.0 (#186)
- [cd33436c](https://github.com/kubedb/db-client-go/commit/cd33436c) Add resty client (#185)
- [332d74bd](https://github.com/kubedb/db-client-go/commit/332d74bd) Add tls for ignite (#184)
- [95952fe8](https://github.com/kubedb/db-client-go/commit/95952fe8) Add TLS support for Cassandra (#183)
- [524e2679](https://github.com/kubedb/db-client-go/commit/524e2679) Update Kafka version configuration (#182)
- [f948e1d2](https://github.com/kubedb/db-client-go/commit/f948e1d2) Add druid apis for monitoring (#181)
- [5786aaf4](https://github.com/kubedb/db-client-go/commit/5786aaf4) Add solr metrics client apis (#179)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.11.0](https://github.com/kubedb/druid/releases/tag/v0.11.0)

- [5a6d8a5a](https://github.com/kubedb/druid/commit/5a6d8a5a) Prepare for release v0.11.0 (#88)
- [4f66c8e3](https://github.com/kubedb/druid/commit/4f66c8e3) Fixed patching db for authSecret reference (#87)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.56.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.56.0)

- [908cc210](https://github.com/kubedb/elasticsearch/commit/908cc2106) Prepare for release v0.56.0 (#767)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.19.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.19.0)

- [0ae7afa2](https://github.com/kubedb/elasticsearch-restic-plugin/commit/0ae7afa2) Prepare for release v0.19.0 (#64)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.11.0](https://github.com/kubedb/ferretdb/releases/tag/v0.11.0)

- [cefd46bd](https://github.com/kubedb/ferretdb/commit/cefd46bd) Prepare for release v0.11.0 (#78)
- [e4246d4f](https://github.com/kubedb/ferretdb/commit/e4246d4f) Set containers & annotations from backend podTemplate



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.4.0](https://github.com/kubedb/gitops/releases/tag/v0.4.0)

- [66ba799e](https://github.com/kubedb/gitops/commit/66ba799e) Prepare for release v0.4.0 (#17)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.2.0](https://github.com/kubedb/hazelcast/releases/tag/v0.2.0)

- [3a24d061](https://github.com/kubedb/hazelcast/commit/3a24d061) Prepare for release v0.2.0 (#4)
- [e5d96b31](https://github.com/kubedb/hazelcast/commit/e5d96b31) Changes for integrating with Ops manager (#3)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.3.0](https://github.com/kubedb/ignite/releases/tag/v0.3.0)

- [d2868afc](https://github.com/kubedb/ignite/commit/d2868afc) Prepare for release v0.3.0 (#10)
- [c5380059](https://github.com/kubedb/ignite/commit/c5380059) Add Ignite TLS (#9)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2025.6.30](https://github.com/kubedb/installer/releases/tag/v2025.6.30)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.27.0](https://github.com/kubedb/kafka/releases/tag/v0.27.0)

- [22c04abb](https://github.com/kubedb/kafka/commit/22c04abb) Prepare for release v0.27.0 (#153)
- [47ec6ebf](https://github.com/kubedb/kafka/commit/47ec6ebf) Add Kafka Init Container changes for v4.0.0 (#152)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.32.0](https://github.com/kubedb/kibana/releases/tag/v0.32.0)

- [ceff9782](https://github.com/kubedb/kibana/commit/ceff9782) Prepare for release v0.32.0 (#153)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.19.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.19.0)

- [d92058de](https://github.com/kubedb/kubedb-manifest-plugin/commit/d92058de) Prepare for release v0.19.0 (#95)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.7.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.7.0)

- [f7e6a32](https://github.com/kubedb/kubedb-verifier/commit/f7e6a32) Prepare for release v0.7.0 (#20)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.40.0](https://github.com/kubedb/mariadb/releases/tag/v0.40.0)

- [8f03696f](https://github.com/kubedb/mariadb/commit/8f03696fe) Prepare for release v0.40.0 (#334)
- [cd2f4f2e](https://github.com/kubedb/mariadb/commit/cd2f4f2ee) Set DB User as Default SecurityContext for Restore JobTemplate (#333)
- [35a47efa](https://github.com/kubedb/mariadb/commit/35a47efa2) add max-concurrent-reconciles flag to operator (#331)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.16.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.16.0)

- [428cea2e](https://github.com/kubedb/mariadb-archiver/commit/428cea2e) Prepare for release v0.16.0 (#52)
- [ca24ffbf](https://github.com/kubedb/mariadb-archiver/commit/ca24ffbf) Add MariaDB Replication Support (#51)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.36.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.36.0)

- [cf404a16](https://github.com/kubedb/mariadb-coordinator/commit/cf404a16) Prepare for release v0.36.0 (#144)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.16.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.16.0)

- [816b7a85](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/816b7a85) Update deps
- [2147ed2b](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/2147ed2b) Merge pull request #48 from kubedb/maxscale



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.14.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.14.0)

- [afb59a5](https://github.com/kubedb/mariadb-restic-plugin/commit/afb59a5) Prepare for release v0.14.0 (#47)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.49.0](https://github.com/kubedb/memcached/releases/tag/v0.49.0)

- [4ddfa95e](https://github.com/kubedb/memcached/commit/4ddfa95ef) Prepare for release v0.49.0 (#501)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.49.0](https://github.com/kubedb/mongodb/releases/tag/v0.49.0)

- [0ebe664b](https://github.com/kubedb/mongodb/commit/0ebe664be) Prepare for release v0.49.0 (#709)
- [fa641f59](https://github.com/kubedb/mongodb/commit/fa641f594) remove replicaCount=3 (#708)
- [451525f6](https://github.com/kubedb/mongodb/commit/451525f64) Set DB User as Default SecurityContext for Restore JobTemplate (#707)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.17.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.17.0)

- [d4ff7543](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/d4ff7543) Prepare for release v0.17.0 (#52)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.19.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.19.0)

- [39813bf](https://github.com/kubedb/mongodb-restic-plugin/commit/39813bf) Prepare for release v0.19.0 (#85)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.11.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.11.0)

- [ecd5f1f0](https://github.com/kubedb/mssql-coordinator/commit/ecd5f1f0) Prepare for release v0.11.0 (#39)
- [873a0e2a](https://github.com/kubedb/mssql-coordinator/commit/873a0e2a) Add Distributed AG Support  (#38)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.11.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.11.0)

- [241028ea](https://github.com/kubedb/mssqlserver/commit/241028ea) Prepare for release v0.11.0 (#81)
- [e3a8001d](https://github.com/kubedb/mssqlserver/commit/e3a8001d) Add SQL Server Distributed AG Support (#79)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.10.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.10.0)




## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.10.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.10.0)

- [574cc4c](https://github.com/kubedb/mssqlserver-walg-plugin/commit/574cc4c) Prepare for release v0.10.0 (#27)
- [f034cd9](https://github.com/kubedb/mssqlserver-walg-plugin/commit/f034cd9) Add Backup Support for Local Storage



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.49.0](https://github.com/kubedb/mysql/releases/tag/v0.49.0)

- [9bdf8e80](https://github.com/kubedb/mysql/commit/9bdf8e808) Prepare for release v0.49.0 (#689)
- [21378cc3](https://github.com/kubedb/mysql/commit/21378cc34) Do not default securityContext for backupConfig jobTemplate (#688)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.17.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.17.0)

- [fe7e924d](https://github.com/kubedb/mysql-archiver/commit/fe7e924d) Prepare for release v0.17.0 (#61)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.34.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.34.0)

- [b6b87ddc](https://github.com/kubedb/mysql-coordinator/commit/b6b87ddc) Prepare for release v0.34.0 (#144)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.17.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.17.0)

- [fe0c6363](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/fe0c6363) Prepare for release v0.17.0 (#48)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.19.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.19.0)

- [56f7c0f](https://github.com/kubedb/mysql-restic-plugin/commit/56f7c0f) Prepare for release v0.19.0 (#75)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.34.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.34.0)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.43.0](https://github.com/kubedb/ops-manager/releases/tag/v0.43.0)

- [da8fdefe](https://github.com/kubedb/ops-manager/commit/da8fdefee) Prepare for release v0.43.0 (#757)
- [6f824976](https://github.com/kubedb/ops-manager/commit/6f824976f) Add DAG mode changes (#750)
- [3d224462](https://github.com/kubedb/ops-manager/commit/3d2244622) Add Announce for Redis (#752)
- [23d5c41a](https://github.com/kubedb/ops-manager/commit/23d5c41a8) add tls support (#748)
- [5879e1fb](https://github.com/kubedb/ops-manager/commit/5879e1fba) Add hazelcast opsrequest (#756)
- [1cc09df5](https://github.com/kubedb/ops-manager/commit/1cc09df57) Add  Cassandra TLS & Ops requests (#751)
- [ccb42fe4](https://github.com/kubedb/ops-manager/commit/ccb42fe48) Add Ignite TLS (#754)
- [5052c649](https://github.com/kubedb/ops-manager/commit/5052c649f) Allow scaling down Postgres Database to Standalone (#755)
- [f026ff13](https://github.com/kubedb/ops-manager/commit/f026ff132) Add Recommendation Engine support for Druid (#753)
- [28792d2f](https://github.com/kubedb/ops-manager/commit/28792d2f2) Add Scaling and Volume Expansion Support for MaxScale Server (#749)



## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.2.0](https://github.com/kubedb/oracle/releases/tag/v0.2.0)

- [77566893](https://github.com/kubedb/oracle/commit/77566893) Prepare for release v0.2.0 (#5)
- [90e857e1](https://github.com/kubedb/oracle/commit/90e857e1) Remove Zap Logger (#4)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.2.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.2.0)

- [8abae9c](https://github.com/kubedb/oracle-coordinator/commit/8abae9c) Prepare for release v0.2.0 (#4)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.43.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.43.0)

- [68298742](https://github.com/kubedb/percona-xtradb/commit/682987422) Prepare for release v0.43.0 (#409)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.29.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.29.0)

- [0ff16588](https://github.com/kubedb/percona-xtradb-coordinator/commit/0ff16588) Prepare for release v0.29.0 (#97)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.40.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.40.0)

- [9f5c68e3](https://github.com/kubedb/pg-coordinator/commit/9f5c68e3) Prepare for release v0.40.0 (#201)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.43.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.43.0)

- [02383560](https://github.com/kubedb/pgbouncer/commit/02383560) Prepare for release v0.43.0 (#373)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.11.0](https://github.com/kubedb/pgpool/releases/tag/v0.11.0)

- [7719e231](https://github.com/kubedb/pgpool/commit/7719e231) Prepare for release v0.11.0 (#76)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.56.0](https://github.com/kubedb/postgres/releases/tag/v0.56.0)

- [c51c8797](https://github.com/kubedb/postgres/commit/c51c8797e) Prepare for release v0.56.0 (#819)
- [e6a4e275](https://github.com/kubedb/postgres/commit/e6a4e2753) Requeue if virtualSecret get/mount call error 'already exists' (#818)
- [495a8889](https://github.com/kubedb/postgres/commit/495a88895) Do not default securityContext for backupConfig jobTemplate (#817)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.17.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.17.0)

- [dc133d1e](https://github.com/kubedb/postgres-archiver/commit/dc133d1e) Prepare for release v0.17.0 (#62)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.17.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.17.0)

- [5a8a9f88](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/5a8a9f88) Prepare for release v0.17.0 (#58)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.19.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.19.0)

- [fc8f11d](https://github.com/kubedb/postgres-restic-plugin/commit/fc8f11d) Prepare for release v0.19.0 (#72)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.17.0](https://github.com/kubedb/provider-aws/releases/tag/v0.17.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.17.0](https://github.com/kubedb/provider-azure/releases/tag/v0.17.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.17.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.17.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.56.0](https://github.com/kubedb/provisioner/releases/tag/v0.56.0)

- [edffbd95](https://github.com/kubedb/provisioner/commit/edffbd95a) Prepare for release v0.56.0 (#155)
- [dc514ac6](https://github.com/kubedb/provisioner/commit/dc514ac62) Fix archiver security context (#153)
- [9bfdea4f](https://github.com/kubedb/provisioner/commit/9bfdea4fd) Add virtual secret scheme; Remove Zap Logger (#154)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.43.0](https://github.com/kubedb/proxysql/releases/tag/v0.43.0)

- [2b873e15](https://github.com/kubedb/proxysql/commit/2b873e15a) Prepare for release v0.43.0 (#395)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.11.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.11.0)

- [28c55813](https://github.com/kubedb/rabbitmq/commit/28c55813) Prepare for release v0.11.0 (#88)
- [6b779a8e](https://github.com/kubedb/rabbitmq/commit/6b779a8e) Remove Zap Logger (#87)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.49.0](https://github.com/kubedb/redis/releases/tag/v0.49.0)

- [a8c565fe](https://github.com/kubedb/redis/commit/a8c565fe0) Prepare for release v0.49.0 (#596)
- [de478eef](https://github.com/kubedb/redis/commit/de478eefc) Expose busport via standby service (#595)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.35.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.35.0)

- [0b1934a2](https://github.com/kubedb/redis-coordinator/commit/0b1934a2) Prepare for release v0.35.0 (#129)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.19.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.19.0)

- [1a106a0](https://github.com/kubedb/redis-restic-plugin/commit/1a106a0) Prepare for release v0.19.0 (#67)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.43.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.43.0)

- [cf06bc84](https://github.com/kubedb/replication-mode-detector/commit/cf06bc84) Prepare for release v0.43.0 (#294)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.32.0](https://github.com/kubedb/schema-manager/releases/tag/v0.32.0)

- [0a422715](https://github.com/kubedb/schema-manager/commit/0a422715) Prepare for release v0.32.0 (#140)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.11.0](https://github.com/kubedb/singlestore/releases/tag/v0.11.0)

- [59ec2aa4](https://github.com/kubedb/singlestore/commit/59ec2aa4) Prepare for release v0.11.0 (#74)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.11.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.11.0)

- [3d492fe](https://github.com/kubedb/singlestore-coordinator/commit/3d492fe) Prepare for release v0.11.0 (#44)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.14.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.14.0)

- [9a77d96](https://github.com/kubedb/singlestore-restic-plugin/commit/9a77d96) Prepare for release v0.14.0 (#45)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.11.0](https://github.com/kubedb/solr/releases/tag/v0.11.0)

- [c8257fce](https://github.com/kubedb/solr/commit/c8257fce) Prepare for release v0.11.0 (#86)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.41.0](https://github.com/kubedb/tests/releases/tag/v0.41.0)

- [5b2bf0d3](https://github.com/kubedb/tests/commit/5b2bf0d3) Prepare for release v0.41.0 (#467)
- [7527283a](https://github.com/kubedb/tests/commit/7527283a) Postgres Restic backup CI (add k3s cluster setup) (#465)
- [75ecca1f](https://github.com/kubedb/tests/commit/75ecca1f) add arbiter mode (#466)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.32.0](https://github.com/kubedb/ui-server/releases/tag/v0.32.0)

- [b9ccc61d](https://github.com/kubedb/ui-server/commit/b9ccc61d) Prepare for release v0.32.0 (#164)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.32.0](https://github.com/kubedb/webhook-server/releases/tag/v0.32.0)

- [33291324](https://github.com/kubedb/webhook-server/commit/33291324) Prepare for release v0.32.0 (#158)



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.5.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.5.0)

- [f6898af](https://github.com/kubedb/xtrabackup-restic-plugin/commit/f6898af) Prepare for release v0.5.0 (#14)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.11.0](https://github.com/kubedb/zookeeper/releases/tag/v0.11.0)

- [e4cd6b4b](https://github.com/kubedb/zookeeper/commit/e4cd6b4b) Prepare for release v0.11.0 (#78)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.12.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.12.0)

- [45187ce](https://github.com/kubedb/zookeeper-restic-plugin/commit/45187ce) Prepare for release v0.12.0 (#36)




