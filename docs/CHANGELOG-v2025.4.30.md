---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2025.4.30
    name: Changelog-v2025.4.30
    parent: welcome
    weight: 20250430
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2025.4.30/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2025.4.30/
---

# KubeDB v2025.4.30 (2025-04-30)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.54.0](https://github.com/kubedb/apimachinery/releases/tag/v0.54.0)

- [7da7ab31](https://github.com/kubedb/apimachinery/commit/7da7ab313) Update deps
- [b5f3a997](https://github.com/kubedb/apimachinery/commit/b5f3a997d) Add Ignite APIs (#1439)
- [7b89c070](https://github.com/kubedb/apimachinery/commit/7b89c0706) SetDefault for Valkey (#1431)
- [a8ed2b15](https://github.com/kubedb/apimachinery/commit/a8ed2b157) Update Cassandra Version API field for backup (#1448)
- [6c6eea72](https://github.com/kubedb/apimachinery/commit/6c6eea72a) Add mssql helpers for SecondaryAccessMode (#1452)
- [0f3f9fb6](https://github.com/kubedb/apimachinery/commit/0f3f9fb61) Distribution Redis -> Official (#1446)
- [1411575b](https://github.com/kubedb/apimachinery/commit/1411575b8) Add MySQL topology defaults to SetDefaults() (#1442)
- [9071b547](https://github.com/kubedb/apimachinery/commit/9071b547d) Support multiple license restrictions (#1445)
- [739c294b](https://github.com/kubedb/apimachinery/commit/739c294bc) SecondaryAccess -> SecondaryAccessMode (#1451)
- [853d2ab6](https://github.com/kubedb/apimachinery/commit/853d2ab63) Add horizons field in mg & mgOps (#1447)
- [9310d2f4](https://github.com/kubedb/apimachinery/commit/9310d2f43) Add SecondaryAccess field to specify whether to use active or passive secondaries for SQL Server AG (#1450)
- [50acd512](https://github.com/kubedb/apimachinery/commit/50acd5129) Update deps
- [f9c9734e](https://github.com/kubedb/apimachinery/commit/f9c9734e0) Fix defaulting solrAutoscaler (#1444)
- [693605d3](https://github.com/kubedb/apimachinery/commit/693605d3c) Enquque petsets conditionally: standardization (#1441)
- [437d5e8e](https://github.com/kubedb/apimachinery/commit/437d5e8e4) Update Owner Name (#1440)
- [ba449cf3](https://github.com/kubedb/apimachinery/commit/ba449cf3b) Fix Memcached validate (#1438)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.39.0](https://github.com/kubedb/autoscaler/releases/tag/v0.39.0)

- [02241a78](https://github.com/kubedb/autoscaler/commit/02241a78) Prepare for release v0.39.0 (#249)
- [32dfd236](https://github.com/kubedb/autoscaler/commit/32dfd236) Hardcode node typ for es,kf,mg,sl,sdb (#248)
- [480ed1ff](https://github.com/kubedb/autoscaler/commit/480ed1ff) Implement machineProfiles with nodeTopology for all dbs (#247)
- [1bf90192](https://github.com/kubedb/autoscaler/commit/1bf90192) Implement machineProfiles with nodeTopology (#246)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.7.0](https://github.com/kubedb/cassandra/releases/tag/v0.7.0)

- [48b908c7](https://github.com/kubedb/cassandra/commit/48b908c7) Prepare for release v0.7.0 (#33)
- [4d654a13](https://github.com/kubedb/cassandra/commit/4d654a13) Don't PublishNotReadyAddresses for non Headless services (#26)
- [54d2f8c8](https://github.com/kubedb/cassandra/commit/54d2f8c8) Add support for Backup (#32)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.1.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.1.0)

- [6991226](https://github.com/kubedb/cassandra-medusa-plugin/commit/6991226) Prepare for release v0.1.0 (#4)
- [3f4c98a](https://github.com/kubedb/cassandra-medusa-plugin/commit/3f4c98a) Fix license issues
- [45abd2c](https://github.com/kubedb/cassandra-medusa-plugin/commit/45abd2c) Merge pull request #2 from kubedb/cass-backup
- [c779cfb](https://github.com/kubedb/cassandra-medusa-plugin/commit/c779cfb) Merge pull request #3 from kubedb/gha-up
- [354dcfd](https://github.com/kubedb/cassandra-medusa-plugin/commit/354dcfd) Use Go 1.24
- [30b207e](https://github.com/kubedb/cassandra-medusa-plugin/commit/30b207e) Disable image caching in setup-qemu action (#1)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.54.0](https://github.com/kubedb/cli/releases/tag/v0.54.0)

- [a04f28b4](https://github.com/kubedb/cli/commit/a04f28b4f) Prepare for release v0.54.0 (#795)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.9.0](https://github.com/kubedb/clickhouse/releases/tag/v0.9.0)

- [0ff896fe](https://github.com/kubedb/clickhouse/commit/0ff896fe) Prepare for release v0.9.0 (#50)
- [3b26faef](https://github.com/kubedb/clickhouse/commit/3b26faef) Fix ClickHouse Keeper Selector (#48)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.9.0](https://github.com/kubedb/crd-manager/releases/tag/v0.9.0)

- [4362bceb](https://github.com/kubedb/crd-manager/commit/4362bceb) Prepare for release v0.9.0 (#76)
- [869f231f](https://github.com/kubedb/crd-manager/commit/869f231f) Add Ignite (#74)
- [625d5bf3](https://github.com/kubedb/crd-manager/commit/625d5bf3) Install gitops crds only if the flag is on (#75)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.12.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.12.0)

- [e89c675](https://github.com/kubedb/dashboard-restic-plugin/commit/e89c675) Prepare for release v0.12.0 (#38)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.9.0](https://github.com/kubedb/db-client-go/releases/tag/v0.9.0)

- [7d3f54b7](https://github.com/kubedb/db-client-go/commit/7d3f54b7) Prepare for release v0.9.0 (#173)
- [a9b612da](https://github.com/kubedb/db-client-go/commit/a9b612da) fix ignite sqlClient (#172)
- [86f1d163](https://github.com/kubedb/db-client-go/commit/86f1d163) Add Ignite Client (#171)
- [8141b087](https://github.com/kubedb/db-client-go/commit/8141b087) Add Memcached Client (#170)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.9.0](https://github.com/kubedb/druid/releases/tag/v0.9.0)

- [644653c6](https://github.com/kubedb/druid/commit/644653c6) Prepare for release v0.9.0 (#84)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.54.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.54.0)

- [8faa2364](https://github.com/kubedb/elasticsearch/commit/8faa23647) Prepare for release v0.54.0 (#764)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.17.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.17.0)

- [ba37ce8a](https://github.com/kubedb/elasticsearch-restic-plugin/commit/ba37ce8a) Prepare for release v0.17.0 (#62)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.9.0](https://github.com/kubedb/ferretdb/releases/tag/v0.9.0)

- [fd85af44](https://github.com/kubedb/ferretdb/commit/fd85af44) Prepare for release v0.9.0 (#74)
- [3340af01](https://github.com/kubedb/ferretdb/commit/3340af01) fix stats service selector (#73)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.2.0](https://github.com/kubedb/gitops/releases/tag/v0.2.0)

- [a01846f2](https://github.com/kubedb/gitops/commit/a01846f2) Prepare for release v0.2.0 (#14)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.1.0](https://github.com/kubedb/ignite/releases/tag/v0.1.0)

- [5583cbdb](https://github.com/kubedb/ignite/commit/5583cbdb) Fix script permission (#5)
- [3f103f76](https://github.com/kubedb/ignite/commit/3f103f76) Prepare for release v0.1.0 (#4)
- [0800cfc7](https://github.com/kubedb/ignite/commit/0800cfc7) Cleanup ci
- [428b7dd3](https://github.com/kubedb/ignite/commit/428b7dd3) Test against k8s 1.32 (#2)
- [611f01a8](https://github.com/kubedb/ignite/commit/611f01a8) Add Ignite Operator Bootstrap (#3)
- [eb2b1097](https://github.com/kubedb/ignite/commit/eb2b1097) Add skeleton controller



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2025.4.30](https://github.com/kubedb/installer/releases/tag/v2025.4.30)

- [090b8588](https://github.com/kubedb/installer/commit/090b8588a) Prepare for release v2025.4.30 (#1680)
- [d8676aac](https://github.com/kubedb/installer/commit/d8676aacf) Redis Official distribution added (#1679)
- [9872281e](https://github.com/kubedb/installer/commit/9872281e1) Add Ignite (#1643)
- [b86dbf1e](https://github.com/kubedb/installer/commit/b86dbf1e9) Add ValKey Versions (#1677)
- [e87a2659](https://github.com/kubedb/installer/commit/e87a26598) Update catalog
- [8c6ea49f](https://github.com/kubedb/installer/commit/8c6ea49f9) Update cve report 2025-04-28 (#1665)
- [3303d562](https://github.com/kubedb/installer/commit/3303d5624) Add support for Cassandra backup (#1676)
- [455bf1b5](https://github.com/kubedb/installer/commit/455bf1b51) Add New ProxySQL Version 2.7.3 (#1666)
- [9270d064](https://github.com/kubedb/installer/commit/9270d064c) Add '--install-gitops-crds' flag (#1661)
- [0b28a1fa](https://github.com/kubedb/installer/commit/0b28a1fa8) Add Valkey to Redis license restriction (#1660)
- [fa9e198b](https://github.com/kubedb/installer/commit/fa9e198b8) Update crds for kubedb/apimachinery@b5f3a997 (#1675)
- [0d658935](https://github.com/kubedb/installer/commit/0d6589353) Update MySQL Init Image (#1667)
- [ca728491](https://github.com/kubedb/installer/commit/ca728491d) Update cve report (#1664)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.25.0](https://github.com/kubedb/kafka/releases/tag/v0.25.0)

- [ee707865](https://github.com/kubedb/kafka/commit/ee707865) Prepare for release v0.25.0 (#149)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.30.0](https://github.com/kubedb/kibana/releases/tag/v0.30.0)

- [d71bb28f](https://github.com/kubedb/kibana/commit/d71bb28f) Prepare for release v0.30.0 (#151)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.17.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.17.0)

- [1e74db51](https://github.com/kubedb/kubedb-manifest-plugin/commit/1e74db51) Prepare for release v0.17.0 (#93)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.5.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.5.0)

- [717e717](https://github.com/kubedb/kubedb-verifier/commit/717e717) Prepare for release v0.5.0 (#18)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.38.0](https://github.com/kubedb/mariadb/releases/tag/v0.38.0)

- [e7b961fd](https://github.com/kubedb/mariadb/commit/e7b961fd5) Prepare for release v0.38.0 (#328)
- [7d84f316](https://github.com/kubedb/mariadb/commit/7d84f3169) Nightly Test CI Fix (#317)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.14.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.14.0)

- [3357ad29](https://github.com/kubedb/mariadb-archiver/commit/3357ad29) Prepare for release v0.14.0 (#48)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.34.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.34.0)

- [3af49e38](https://github.com/kubedb/mariadb-coordinator/commit/3af49e38) Prepare for release v0.34.0 (#142)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.14.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.14.0)

- [5aa23344](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/5aa23344) Prepare for release v0.14.0 (#46)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.12.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.12.0)

- [99425c9](https://github.com/kubedb/mariadb-restic-plugin/commit/99425c9) Prepare for release v0.12.0 (#45)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.47.0](https://github.com/kubedb/memcached/releases/tag/v0.47.0)

- [933a5aaf](https://github.com/kubedb/memcached/commit/933a5aafa) Prepare for release v0.47.0 (#498)
- [59215715](https://github.com/kubedb/memcached/commit/59215715d) Update Healthcheck (#497)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.47.0](https://github.com/kubedb/mongodb/releases/tag/v0.47.0)

- [41e22e2a](https://github.com/kubedb/mongodb/commit/41e22e2a5) Prepare for release v0.47.0 (#704)
- [d4e2e085](https://github.com/kubedb/mongodb/commit/d4e2e085b) Use primary pod for creating role in repl in archiver resotre (#702)
- [8edd009a](https://github.com/kubedb/mongodb/commit/8edd009a3) Move renaming-related codes on archiver-restore to a seperate file
- [13da2f49](https://github.com/kubedb/mongodb/commit/13da2f496) Fix shard reconciliation (#701)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.15.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.15.0)

- [79d16e38](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/79d16e38) Prepare for release v0.15.0 (#50)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.17.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.17.0)

- [accf597](https://github.com/kubedb/mongodb-restic-plugin/commit/accf597) Prepare for release v0.17.0 (#83)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.9.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.9.0)

- [e450a03b](https://github.com/kubedb/mssql-coordinator/commit/e450a03b) Prepare for release v0.9.0 (#36)
- [0a45cf12](https://github.com/kubedb/mssql-coordinator/commit/0a45cf12) Use option to specify whether to use active or passive secondaries for SQL Server AG (#35)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.9.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.9.0)

- [8c19feb3](https://github.com/kubedb/mssqlserver/commit/8c19feb3) Prepare for release v0.9.0 (#77)
- [35f43968](https://github.com/kubedb/mssqlserver/commit/35f43968) Accept certificates with negative serial number (#75)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.8.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.8.0)




## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.8.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.8.0)

- [1b092f1](https://github.com/kubedb/mssqlserver-walg-plugin/commit/1b092f1) Prepare for release v0.8.0 (#24)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.47.0](https://github.com/kubedb/mysql/releases/tag/v0.47.0)

- [ec5c6e21](https://github.com/kubedb/mysql/commit/ec5c6e21b) Prepare for release v0.47.0 (#684)
- [52f0a8e1](https://github.com/kubedb/mysql/commit/52f0a8e1c) Add MySQL topology defaults to SetDefaults() (#683)
- [fea6344e](https://github.com/kubedb/mysql/commit/fea6344e5) Make InnoDB Buffer Pool Size Dynamic (#680)
- [aacb8217](https://github.com/kubedb/mysql/commit/aacb82173) Divide CI along with db_mode (#682)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.15.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.15.0)

- [611a1cfd](https://github.com/kubedb/mysql-archiver/commit/611a1cfd) Prepare for release v0.15.0 (#58)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.32.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.32.0)

- [f1ada940](https://github.com/kubedb/mysql-coordinator/commit/f1ada940) Prepare for release v0.32.0 (#141)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.15.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.15.0)

- [59822ff4](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/59822ff4) Prepare for release v0.15.0 (#46)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.17.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.17.0)

- [06037ec](https://github.com/kubedb/mysql-restic-plugin/commit/06037ec) Prepare for release v0.17.0 (#73)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.32.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.32.0)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.41.0](https://github.com/kubedb/ops-manager/releases/tag/v0.41.0)

- [076344f4](https://github.com/kubedb/ops-manager/commit/076344f48) Prepare for release v0.41.0 (#731)
- [6e6fc7b8](https://github.com/kubedb/ops-manager/commit/6e6fc7b81) Add Valkey distribution  (#727)
- [e769c065](https://github.com/kubedb/ops-manager/commit/e769c0656) Fix webhook for local deployment (#728)
- [3d82ca3b](https://github.com/kubedb/ops-manager/commit/3d82ca3b0) Update for proxysql controller changes (#730)
- [b5f67337](https://github.com/kubedb/ops-manager/commit/b5f673375) Use option to specify whether to use active or passive secondaries for SQL Server AG (#726)
- [071e805a](https://github.com/kubedb/ops-manager/commit/071e805a9) Implement horizons for MongoDB replicaset (#724)
- [59dcdac7](https://github.com/kubedb/ops-manager/commit/59dcdac78) Update Kafka ops request daily schedule (#714)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.41.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.41.0)

- [7bbf390d](https://github.com/kubedb/percona-xtradb/commit/7bbf390d5) Prepare for release v0.41.0 (#406)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.27.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.27.0)

- [e6259e17](https://github.com/kubedb/percona-xtradb-coordinator/commit/e6259e17) Prepare for release v0.27.0 (#95)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.38.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.38.0)

- [72911d92](https://github.com/kubedb/pg-coordinator/commit/72911d92) Prepare for release v0.38.0 (#199)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.41.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.41.0)

- [a770503f](https://github.com/kubedb/pgbouncer/commit/a770503f) Prepare for release v0.41.0 (#370)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.9.0](https://github.com/kubedb/pgpool/releases/tag/v0.9.0)

- [7d581e91](https://github.com/kubedb/pgpool/commit/7d581e91) Prepare for release v0.9.0 (#73)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.54.0](https://github.com/kubedb/postgres/releases/tag/v0.54.0)

- [b8f06227](https://github.com/kubedb/postgres/commit/b8f062272) Prepare for release v0.54.0 (#812)
- [450a8b08](https://github.com/kubedb/postgres/commit/450a8b085) Fix virtual secrets deletion (#810)
- [db351f9b](https://github.com/kubedb/postgres/commit/db351f9b5) CI-FIX test repo checkout command (#811)
- [e2963887](https://github.com/kubedb/postgres/commit/e2963887d) Add db_mode, separate daily tests (#803)
- [92f7f153](https://github.com/kubedb/postgres/commit/92f7f1534) Fix restore issue (#809)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.15.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.15.0)

- [838f1dd7](https://github.com/kubedb/postgres-archiver/commit/838f1dd7) Prepare for release v0.15.0 (#60)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.15.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.15.0)

- [213533d2](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/213533d2) Prepare for release v0.15.0 (#56)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.17.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.17.0)

- [03c4dbf](https://github.com/kubedb/postgres-restic-plugin/commit/03c4dbf) Prepare for release v0.17.0 (#70)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.15.0](https://github.com/kubedb/provider-aws/releases/tag/v0.15.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.15.0](https://github.com/kubedb/provider-azure/releases/tag/v0.15.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.15.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.15.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.54.0](https://github.com/kubedb/provisioner/releases/tag/v0.54.0)

- [992fd429](https://github.com/kubedb/provisioner/commit/992fd429e) Prepare for release v0.54.0 (#150)
- [68487083](https://github.com/kubedb/provisioner/commit/684870838) Add Ignite (#149)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.41.0](https://github.com/kubedb/proxysql/releases/tag/v0.41.0)

- [5e55f968](https://github.com/kubedb/proxysql/commit/5e55f9684) Prepare for release v0.41.0 (#392)
- [131b3bd5](https://github.com/kubedb/proxysql/commit/131b3bd51) Add PerconaXtraDB Galera Cluster suport (#391)
- [05cdca0f](https://github.com/kubedb/proxysql/commit/05cdca0fb) Add MariaDB suport for ProxySQL, pod dns for mysql_servers config (#383)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.9.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.9.0)

- [f938f3f8](https://github.com/kubedb/rabbitmq/commit/f938f3f8) Prepare for release v0.9.0 (#82)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.47.0](https://github.com/kubedb/redis/releases/tag/v0.47.0)

- [75318363](https://github.com/kubedb/redis/commit/75318363b) Prepare for release v0.47.0 (#591)
- [d0ba7016](https://github.com/kubedb/redis/commit/d0ba7016a) Remove replace
- [155af1e2](https://github.com/kubedb/redis/commit/155af1e24) Add Valkey support (#584)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.33.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.33.0)

- [3bd07eb1](https://github.com/kubedb/redis-coordinator/commit/3bd07eb1) Prepare for release v0.33.0 (#127)
- [f3566c1e](https://github.com/kubedb/redis-coordinator/commit/f3566c1e) valkey integration (#124)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.17.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.17.0)

- [1ec5f1c](https://github.com/kubedb/redis-restic-plugin/commit/1ec5f1c) Prepare for release v0.17.0 (#65)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.41.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.41.0)

- [53576cb1](https://github.com/kubedb/replication-mode-detector/commit/53576cb1) Prepare for release v0.41.0 (#292)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.30.0](https://github.com/kubedb/schema-manager/releases/tag/v0.30.0)

- [8f0e3e2d](https://github.com/kubedb/schema-manager/commit/8f0e3e2d) Prepare for release v0.30.0 (#138)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.9.0](https://github.com/kubedb/singlestore/releases/tag/v0.9.0)

- [2122d0f9](https://github.com/kubedb/singlestore/commit/2122d0f9) Prepare for release v0.9.0 (#70)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.9.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.9.0)

- [1a302d6](https://github.com/kubedb/singlestore-coordinator/commit/1a302d6) Prepare for release v0.9.0 (#42)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.12.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.12.0)

- [189c173](https://github.com/kubedb/singlestore-restic-plugin/commit/189c173) Prepare for release v0.12.0 (#42)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.9.0](https://github.com/kubedb/solr/releases/tag/v0.9.0)

- [e8bc7698](https://github.com/kubedb/solr/commit/e8bc7698) Prepare for release v0.9.0 (#83)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.39.0](https://github.com/kubedb/tests/releases/tag/v0.39.0)

- [5aa1d756](https://github.com/kubedb/tests/commit/5aa1d756) Prepare for release v0.39.0 (#459)
- [04acb1ea](https://github.com/kubedb/tests/commit/04acb1ea) MySQL Test Config, fix profiles (#458)
- [4297ee65](https://github.com/kubedb/tests/commit/4297ee65) Add DB_Mode and Divide test profiles for MySQL (#450)
- [eb2e0520](https://github.com/kubedb/tests/commit/eb2e0520) Postgres Archiver  - Standalone (#454)
- [265b6250](https://github.com/kubedb/tests/commit/265b6250) Fix BackupConfigDeletionPolicy, CI issues (#455)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.30.0](https://github.com/kubedb/ui-server/releases/tag/v0.30.0)

- [7fc6654b](https://github.com/kubedb/ui-server/commit/7fc6654b) Prepare for release v0.30.0 (#162)
- [fa8ad21a](https://github.com/kubedb/ui-server/commit/fa8ad21a) Add connection support for es,fr,sl (#161)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.30.0](https://github.com/kubedb/webhook-server/releases/tag/v0.30.0)

- [f259b5de](https://github.com/kubedb/webhook-server/commit/f259b5de) Prepare for release v0.30.0 (#154)
- [08138276](https://github.com/kubedb/webhook-server/commit/08138276) Add ignite



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.3.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.3.0)

- [95889ff](https://github.com/kubedb/xtrabackup-restic-plugin/commit/95889ff) Prepare for release v0.3.0 (#12)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.9.0](https://github.com/kubedb/zookeeper/releases/tag/v0.9.0)

- [5ec85425](https://github.com/kubedb/zookeeper/commit/5ec85425) Prepare for release v0.9.0 (#75)
- [458caade](https://github.com/kubedb/zookeeper/commit/458caade) Use PatchStatus to avoid timing issues on db deletion (#74)
- [097d6bd3](https://github.com/kubedb/zookeeper/commit/097d6bd3) Fix webhook config and Configure restore session controller (#73)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.10.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.10.0)

- [d76732b](https://github.com/kubedb/zookeeper-restic-plugin/commit/d76732b) Prepare for release v0.10.0 (#34)




