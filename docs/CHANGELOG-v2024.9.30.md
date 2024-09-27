---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2024.9.30
    name: Changelog-v2024.9.30
    parent: welcome
    weight: 20240930
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2024.9.30/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2024.9.30/
---

# KubeDB v2024.9.30 (2024-09-27)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.48.0](https://github.com/kubedb/apimachinery/releases/tag/v0.48.0)

- [52824a32](https://github.com/kubedb/apimachinery/commit/52824a32c) Update deps
- [ce1269f7](https://github.com/kubedb/apimachinery/commit/ce1269f7a) Add DB Connection api for UI (#1300)
- [270c350d](https://github.com/kubedb/apimachinery/commit/270c350d4) Add support for NetworkPolicy (#1309)
- [1a812915](https://github.com/kubedb/apimachinery/commit/1a812915f) Update api to configure tls for solr (#1297)
- [83aa723b](https://github.com/kubedb/apimachinery/commit/83aa723b3) Fix DB's restoring phase for application-level restore (#1306)
- [706baeb4](https://github.com/kubedb/apimachinery/commit/706baeb42) Change default wal_keep_size for postgres (#1310)
- [42f85a6e](https://github.com/kubedb/apimachinery/commit/42f85a6e3) Add Support of Monitoring to ClickHouse (#1302)
- [027fb904](https://github.com/kubedb/apimachinery/commit/027fb9041) Add Support for ClickHouse Custom Config (#1299)
- [8657958f](https://github.com/kubedb/apimachinery/commit/8657958fe) Update Kibana API (#1301)
- [2c9d3953](https://github.com/kubedb/apimachinery/commit/2c9d3953d) Add ClickHouse Keeper Api (#1278)
- [9349421e](https://github.com/kubedb/apimachinery/commit/9349421ec) Add update constraints to Pgpool api (#1295)
- [1d90d680](https://github.com/kubedb/apimachinery/commit/1d90d680f) Add MS SQL Ops Requests APIs (#1198)
- [adf7fe76](https://github.com/kubedb/apimachinery/commit/adf7fe76c) Add Kafka Scram Constants (#1304)
- [88b051c9](https://github.com/kubedb/apimachinery/commit/88b051c92) Add init field in druid api for backup (#1305)
- [38154e49](https://github.com/kubedb/apimachinery/commit/38154e492) Reconfigure PGBouncer (#1291)
- [639cafff](https://github.com/kubedb/apimachinery/commit/639cafff0) Add ZooKeeper Ops Request (#1263)
- [f622d2d3](https://github.com/kubedb/apimachinery/commit/f622d2d3c) Add Cassandra Autoscaler Api (#1308)
- [5a256e68](https://github.com/kubedb/apimachinery/commit/5a256e688) Dont use kube-ui-server to detect rancher project namespaces (#1307)
- [c4598e14](https://github.com/kubedb/apimachinery/commit/c4598e143) Add Cassandra API (#1283)
- [e32b81a8](https://github.com/kubedb/apimachinery/commit/e32b81a82) Use KIND v0.24.0 (#1303)
- [9ed3d95a](https://github.com/kubedb/apimachinery/commit/9ed3d95a4) Add UpdateVersion API for pgbouncer (#1284)
- [f4c829bd](https://github.com/kubedb/apimachinery/commit/f4c829bda) Add updateConstraints in Memcached Api (#1298)
- [ecd23db4](https://github.com/kubedb/apimachinery/commit/ecd23db43) Update for release KubeStash@v2024.8.30 (#1296)
- [ceb03dd8](https://github.com/kubedb/apimachinery/commit/ceb03dd8b) Remove config secret check from Pgpool webhook (#1294)
- [eb51db0d](https://github.com/kubedb/apimachinery/commit/eb51db0d7) Update for release Stash@v2024.8.27 (#1293)
- [60f4751e](https://github.com/kubedb/apimachinery/commit/60f4751e3) Ignore Postgres reference when deleting Pgpool (#1292)
- [27985f4e](https://github.com/kubedb/apimachinery/commit/27985f4eb) Add monitoring defaults for mssqlserver (#1290)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.33.0](https://github.com/kubedb/autoscaler/releases/tag/v0.33.0)

- [ba8c02d9](https://github.com/kubedb/autoscaler/commit/ba8c02d9) Prepare for release v0.33.0 (#224)
- [fce8a02b](https://github.com/kubedb/autoscaler/commit/fce8a02b) Add Microsoft SQL Server Autoscaler (#223)
- [326587f5](https://github.com/kubedb/autoscaler/commit/326587f5) Add FerretDB autoscaler (#221)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.1.0](https://github.com/kubedb/cassandra/releases/tag/v0.1.0)

- [3f30b40f](https://github.com/kubedb/cassandra/commit/3f30b40f) Prepare for release v0.1.0 (#4)
- [64cc246e](https://github.com/kubedb/cassandra/commit/64cc246e) Add support for NetworkPolicy (#3)
- [1a175919](https://github.com/kubedb/cassandra/commit/1a175919) Use KIND v0.24.0 (#2)
- [c63d5269](https://github.com/kubedb/cassandra/commit/c63d5269) Add Cassandra Controller (#1)
- [f6700853](https://github.com/kubedb/cassandra/commit/f6700853) Setup initial skeleton
- [ed868ab3](https://github.com/kubedb/cassandra/commit/ed868ab3) Add License and readme



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.48.0](https://github.com/kubedb/cli/releases/tag/v0.48.0)

- [6d82a893](https://github.com/kubedb/cli/commit/6d82a893) Prepare for release v0.48.0 (#778)
- [0d6782c0](https://github.com/kubedb/cli/commit/0d6782c0) add mssql and (#777)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.3.0](https://github.com/kubedb/clickhouse/releases/tag/v0.3.0)

- [df41f2d1](https://github.com/kubedb/clickhouse/commit/df41f2d1) Prepare for release v0.3.0 (#20)
- [1d883404](https://github.com/kubedb/clickhouse/commit/1d883404) Add Support for NetworkPolicy (#19)
- [f0e4114c](https://github.com/kubedb/clickhouse/commit/f0e4114c) Add Internally Managed Keeper, Custom Config, Monitoring Support (#4)
- [d32f7df6](https://github.com/kubedb/clickhouse/commit/d32f7df6) Use KIND v0.24.0 (#18)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.3.0](https://github.com/kubedb/crd-manager/releases/tag/v0.3.0)

- [00607da6](https://github.com/kubedb/crd-manager/commit/00607da6) Prepare for release v0.3.0 (#50)
- [dce1d519](https://github.com/kubedb/crd-manager/commit/dce1d519) Add mssqlserver ops-request crd (#49)
- [a08b59b8](https://github.com/kubedb/crd-manager/commit/a08b59b8) Add support for Cassandra (#45)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.6.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.6.0)

- [f78d04f](https://github.com/kubedb/dashboard-restic-plugin/commit/f78d04f) Prepare for release v0.6.0 (#21)
- [de78d66](https://github.com/kubedb/dashboard-restic-plugin/commit/de78d66) Add timeout for backup and restore (#20)
- [d7eddc2](https://github.com/kubedb/dashboard-restic-plugin/commit/d7eddc2) Use restic 0.17.1 (#19)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.3.0](https://github.com/kubedb/db-client-go/releases/tag/v0.3.0)

- [c1fd8329](https://github.com/kubedb/db-client-go/commit/c1fd8329) Prepare for release v0.3.0 (#140)
- [5f1c5482](https://github.com/kubedb/db-client-go/commit/5f1c5482) Add tls for solr (#135)
- [2182ff59](https://github.com/kubedb/db-client-go/commit/2182ff59) Add DisableSecurity Support for ClickHouse (#137)
- [f15802ff](https://github.com/kubedb/db-client-go/commit/f15802ff) update zk client (#128)
- [51823940](https://github.com/kubedb/db-client-go/commit/51823940) Pass containerPort as parameter (#134)
- [e11cb0a2](https://github.com/kubedb/db-client-go/commit/e11cb0a2) Add Cassandra Client (#130)
- [392630ba](https://github.com/kubedb/db-client-go/commit/392630ba) Fix rabbitmq client builder for disable auth (#138)
- [2af340c4](https://github.com/kubedb/db-client-go/commit/2af340c4) Fix RabbitMQ HTTP client with TLS (#136)
- [b8c278df](https://github.com/kubedb/db-client-go/commit/b8c278df) Fix RabbitMQ client for default Vhost setup (#133)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.3.0](https://github.com/kubedb/druid/releases/tag/v0.3.0)

- [2b6a780e](https://github.com/kubedb/druid/commit/2b6a780e) Prepare for release v0.3.0 (#50)
- [18ef77fc](https://github.com/kubedb/druid/commit/18ef77fc) Add Support for NetworkPolicy (#49)
- [e13d968f](https://github.com/kubedb/druid/commit/e13d968f) Use KIND v0.24.0 (#46)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.48.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.48.0)

- [9cf7cb80](https://github.com/kubedb/elasticsearch/commit/9cf7cb801) Prepare for release v0.48.0 (#734)
- [d948222d](https://github.com/kubedb/elasticsearch/commit/d948222d9) Add network policy support (#733)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.11.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.11.0)

- [d248b00](https://github.com/kubedb/elasticsearch-restic-plugin/commit/d248b00) Prepare for release v0.11.0 (#45)
- [d319899](https://github.com/kubedb/elasticsearch-restic-plugin/commit/d319899) Update helper method and refactor code (#44)
- [095aeb8](https://github.com/kubedb/elasticsearch-restic-plugin/commit/095aeb8) Undo `DBReady` condition check for external database backup (#42)
- [b068c4b](https://github.com/kubedb/elasticsearch-restic-plugin/commit/b068c4b) Add timeout for backup and restore (#43)
- [ee77367](https://github.com/kubedb/elasticsearch-restic-plugin/commit/ee77367) Use restic 0.17.1 (#41)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.3.0](https://github.com/kubedb/ferretdb/releases/tag/v0.3.0)

- [19802a3f](https://github.com/kubedb/ferretdb/commit/19802a3f) Prepare for release v0.3.0 (#45)
- [2d9b8382](https://github.com/kubedb/ferretdb/commit/2d9b8382) Add Support for NetworkPolicy (#44)
- [9c12274c](https://github.com/kubedb/ferretdb/commit/9c12274c) make version accessible (#43)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2024.9.30](https://github.com/kubedb/installer/releases/tag/v2024.9.30)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.19.0](https://github.com/kubedb/kafka/releases/tag/v0.19.0)

- [9810e653](https://github.com/kubedb/kafka/commit/9810e653) Prepare for release v0.19.0 (#109)
- [119eeac9](https://github.com/kubedb/kafka/commit/119eeac9) add network policy support for kafka (#108)
- [b3cc4075](https://github.com/kubedb/kafka/commit/b3cc4075) Add Kafka Security Scram (#107)
- [7981d5b4](https://github.com/kubedb/kafka/commit/7981d5b4) Use KIND v0.24.0 (#106)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.24.0](https://github.com/kubedb/kibana/releases/tag/v0.24.0)

- [3546f26e](https://github.com/kubedb/kibana/commit/3546f26e) Prepare for release v0.24.0 (#127)
- [a6bac65e](https://github.com/kubedb/kibana/commit/a6bac65e) Fix healthcheck and refactor repository (#126)
- [00962224](https://github.com/kubedb/kibana/commit/00962224) Rename repo to kibana



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.11.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.11.0)

- [9e2894b](https://github.com/kubedb/kubedb-manifest-plugin/commit/9e2894b) Prepare for release v0.11.0 (#71)
- [5018874](https://github.com/kubedb/kubedb-manifest-plugin/commit/5018874) Fix druid external dependencies defaulting for restore (#70)
- [5fd90bf](https://github.com/kubedb/kubedb-manifest-plugin/commit/5fd90bf) Add SingleStore Manifest Backup (#69)
- [da14505](https://github.com/kubedb/kubedb-manifest-plugin/commit/da14505) Refactor (#68)
- [bac2445](https://github.com/kubedb/kubedb-manifest-plugin/commit/bac2445) Add druid manifest backup (#61)
- [775288e](https://github.com/kubedb/kubedb-manifest-plugin/commit/775288e) Use v1 api; Reset podTemplate properly on restore (#66)
- [a02b671](https://github.com/kubedb/kubedb-manifest-plugin/commit/a02b671) Add timeout for backup and restore (#65)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.32.0](https://github.com/kubedb/mariadb/releases/tag/v0.32.0)

- [bd16acf1](https://github.com/kubedb/mariadb/commit/bd16acf10) Prepare for release v0.32.0 (#285)
- [8506ab7a](https://github.com/kubedb/mariadb/commit/8506ab7a1) Add Support for NetworkPolicy (#284)
- [842b2f50](https://github.com/kubedb/mariadb/commit/842b2f501) Remove Initial Trigger Backup (#283)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.8.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.8.0)

- [34dcc82f](https://github.com/kubedb/mariadb-archiver/commit/34dcc82f) Prepare for release v0.8.0 (#28)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.28.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.28.0)

- [9a921621](https://github.com/kubedb/mariadb-coordinator/commit/9a921621) Prepare for release v0.28.0 (#127)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.8.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.8.0)

- [c9f7a01](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/c9f7a01) Prepare for release v0.8.0 (#31)
- [23e6751](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/23e6751) Fix Snapshot Fail Status not Updating (#30)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.6.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.6.0)

- [76729c2](https://github.com/kubedb/mariadb-restic-plugin/commit/76729c2) Prepare for release v0.6.0 (#26)
- [b259caa](https://github.com/kubedb/mariadb-restic-plugin/commit/b259caa) Update helper method and refactor code (#25)
- [3edea46](https://github.com/kubedb/mariadb-restic-plugin/commit/3edea46) Add timeout for backup and restore (#24)
- [d4633a1](https://github.com/kubedb/mariadb-restic-plugin/commit/d4633a1) Use restic 0.17.1 (#23)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.41.0](https://github.com/kubedb/memcached/releases/tag/v0.41.0)

- [0d034542](https://github.com/kubedb/memcached/commit/0d0345422) Prepare for release v0.41.0 (#466)
- [cbb3473d](https://github.com/kubedb/memcached/commit/cbb3473df) Add network policy support (#465)
- [93062094](https://github.com/kubedb/memcached/commit/930620948) Add Ensure Memcached Config (#464)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.41.0](https://github.com/kubedb/mongodb/releases/tag/v0.41.0)

- [8d48e934](https://github.com/kubedb/mongodb/commit/8d48e934e) Prepare for release v0.41.0 (#658)
- [3bd0132f](https://github.com/kubedb/mongodb/commit/3bd0132f9) Add support for NetworkPolicy (#657)
- [7499a272](https://github.com/kubedb/mongodb/commit/7499a272a) update kubestash deps (#655)
- [81734ed3](https://github.com/kubedb/mongodb/commit/81734ed3a) use `-f` instead of `--config` for custom config (#654)
- [f2d4f5a2](https://github.com/kubedb/mongodb/commit/f2d4f5a20) Add all utility installation-commands in makefile (#653)
- [682903ca](https://github.com/kubedb/mongodb/commit/682903caa) Copy storageSecret on walg-restore (#652)
- [cfdd1aa9](https://github.com/kubedb/mongodb/commit/cfdd1aa9b) Fix trigger backup once and restore path issue (#651)
- [6f0cb14d](https://github.com/kubedb/mongodb/commit/6f0cb14d9) Make granular packages (#650)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.9.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.9.0)

- [94a4768](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/94a4768) Prepare for release v0.9.0 (#35)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.11.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.11.0)

- [85fa28e](https://github.com/kubedb/mongodb-restic-plugin/commit/85fa28e) Prepare for release v0.11.0 (#66)
- [3bbb7ed](https://github.com/kubedb/mongodb-restic-plugin/commit/3bbb7ed) print mongodb output if get error (#64)
- [7cc56b2](https://github.com/kubedb/mongodb-restic-plugin/commit/7cc56b2) Update helper method and refactor code (#65)
- [5af3981](https://github.com/kubedb/mongodb-restic-plugin/commit/5af3981) Add timeout for backup and restore (#62)
- [fb269f3](https://github.com/kubedb/mongodb-restic-plugin/commit/fb269f3) Initialize the restic components only (#63)
- [2173316](https://github.com/kubedb/mongodb-restic-plugin/commit/2173316) Use restic 0.17.1 (#61)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.3.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.3.0)

- [af0e9071](https://github.com/kubedb/mssql-coordinator/commit/af0e9071) Prepare for release v0.3.0 (#17)
- [ede4a4a8](https://github.com/kubedb/mssql-coordinator/commit/ede4a4a8) Refactor (#16)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.3.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.3.0)

- [68177328](https://github.com/kubedb/mssqlserver/commit/68177328) Prepare for release v0.3.0 (#30)
- [b3ad0719](https://github.com/kubedb/mssqlserver/commit/b3ad0719) Add Network Policy (#29)
- [0ceb7e2e](https://github.com/kubedb/mssqlserver/commit/0ceb7e2e) New version 2022-CU14 related changes (#28)
- [9b56906d](https://github.com/kubedb/mssqlserver/commit/9b56906d) Add application backup/restore support (#27)
- [9ed0de02](https://github.com/kubedb/mssqlserver/commit/9ed0de02) Export variables and add Custom Configuration  (#24)
- [4ddadb8f](https://github.com/kubedb/mssqlserver/commit/4ddadb8f) Use KIND v0.24.0 (#26)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.2.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.2.0)




## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.2.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.2.0)

- [d4fc0aa](https://github.com/kubedb/mssqlserver-walg-plugin/commit/d4fc0aa) Prepare for release v0.2.0 (#7)
- [4692e18](https://github.com/kubedb/mssqlserver-walg-plugin/commit/4692e18) Add Application level backup/restore support (#6)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.41.0](https://github.com/kubedb/mysql/releases/tag/v0.41.0)

- [adff4b59](https://github.com/kubedb/mysql/commit/adff4b592) Prepare for release v0.41.0 (#642)
- [7c215328](https://github.com/kubedb/mysql/commit/7c2153284) Add Support for NetworkPolicy (#641)
- [3b411bb7](https://github.com/kubedb/mysql/commit/3b411bb75) Remove Initial Trigger Backup and Refactor RestoreNamespace Field (#640)
- [0150b29c](https://github.com/kubedb/mysql/commit/0150b29cc) Add  Env and Health Check Condition for v8.4.2 (#639)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.9.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.9.0)

- [812086aa](https://github.com/kubedb/mysql-archiver/commit/812086aa) Prepare for release v0.9.0 (#41)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.26.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.26.0)

- [f145403c](https://github.com/kubedb/mysql-coordinator/commit/f145403c) Prepare for release v0.26.0 (#124)
- [bdb20c37](https://github.com/kubedb/mysql-coordinator/commit/bdb20c37) Add MySQL New Version 8.4.2 (#123)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.9.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.9.0)

- [9c987dd](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/9c987dd) Prepare for release v0.9.0 (#29)
- [81532b9](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/81532b9) Fix Snapshot Fail Status not Updating (#28)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.11.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.11.0)

- [ea38abb](https://github.com/kubedb/mysql-restic-plugin/commit/ea38abb) Prepare for release v0.11.0 (#58)
- [b83f066](https://github.com/kubedb/mysql-restic-plugin/commit/b83f066) Refactor code (#57)
- [7c0df0b](https://github.com/kubedb/mysql-restic-plugin/commit/7c0df0b) Add druid logical backup (#51)
- [ca89c9a](https://github.com/kubedb/mysql-restic-plugin/commit/ca89c9a) Add External Databases Backup/Restore support (#55)
- [a3e6482](https://github.com/kubedb/mysql-restic-plugin/commit/a3e6482) Add timeout for backup and restore (#56)
- [163d4fb](https://github.com/kubedb/mysql-restic-plugin/commit/163d4fb) Use restic 0.17.1 (#54)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.26.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.26.0)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.35.0](https://github.com/kubedb/ops-manager/releases/tag/v0.35.0)




## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.35.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.35.0)

- [2a0bd464](https://github.com/kubedb/percona-xtradb/commit/2a0bd4644) Prepare for release v0.35.0 (#381)
- [0c3572b7](https://github.com/kubedb/percona-xtradb/commit/0c3572b7c) Add Support for NetworkPolicy (#380)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.21.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.21.0)

- [0dbbf87d](https://github.com/kubedb/percona-xtradb-coordinator/commit/0dbbf87d) Prepare for release v0.21.0 (#81)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.32.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.32.0)

- [c69d966c](https://github.com/kubedb/pg-coordinator/commit/c69d966c) Prepare for release v0.32.0 (#173)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.35.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.35.0)

- [4d644e52](https://github.com/kubedb/pgbouncer/commit/4d644e52) Prepare for release v0.35.0 (#347)
- [db6b6365](https://github.com/kubedb/pgbouncer/commit/db6b6365) Add Support for NetworkPolicy (#346)
- [24c2cb96](https://github.com/kubedb/pgbouncer/commit/24c2cb96) config secret merging and using the new merged secret (#345)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.3.0](https://github.com/kubedb/pgpool/releases/tag/v0.3.0)

- [5703b7bc](https://github.com/kubedb/pgpool/commit/5703b7bc) Prepare for release v0.3.0 (#48)
- [5331e699](https://github.com/kubedb/pgpool/commit/5331e699) Add network policy (#47)
- [e8d0477d](https://github.com/kubedb/pgpool/commit/e8d0477d) Add updated apimachinery (#46)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.48.0](https://github.com/kubedb/postgres/releases/tag/v0.48.0)

- [a7f61f7f](https://github.com/kubedb/postgres/commit/a7f61f7f6) Prepare for release v0.48.0 (#754)
- [cb2b30a1](https://github.com/kubedb/postgres/commit/cb2b30a14) Add network policy support (#753)
- [09b14dea](https://github.com/kubedb/postgres/commit/09b14dea4) Add Restic Driver support for postgres archiver (#752)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.9.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.9.0)

- [fac9428f](https://github.com/kubedb/postgres-archiver/commit/fac9428f) Prepare for release v0.9.0 (#40)
- [53d7ba6e](https://github.com/kubedb/postgres-archiver/commit/53d7ba6e) Fix slow wal push issue (#39)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.9.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.9.0)

- [0f4f483](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/0f4f483) Prepare for release v0.9.0 (#38)
- [1415a86](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/1415a86) Remove return error on backup fail (#37)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.11.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.11.0)

- [e75a130](https://github.com/kubedb/postgres-restic-plugin/commit/e75a130) Prepare for release v0.11.0 (#53)
- [d7b516b](https://github.com/kubedb/postgres-restic-plugin/commit/d7b516b) Add Physical backup and restore support for postgres (#51)
- [848bf53](https://github.com/kubedb/postgres-restic-plugin/commit/848bf53) Refactor code (#52)
- [7d3fac6](https://github.com/kubedb/postgres-restic-plugin/commit/7d3fac6) Add druid logical backup (#47)
- [43e6a41](https://github.com/kubedb/postgres-restic-plugin/commit/43e6a41) Undo `DBReady` condition check for external database backup (#49)
- [e335e3e](https://github.com/kubedb/postgres-restic-plugin/commit/e335e3e) Add timeout for backup and restore (#50)
- [d2e8650](https://github.com/kubedb/postgres-restic-plugin/commit/d2e8650) Use restic 0.17.1 (#48)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.10.0](https://github.com/kubedb/provider-aws/releases/tag/v0.10.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.10.0](https://github.com/kubedb/provider-azure/releases/tag/v0.10.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.10.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.10.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.48.0](https://github.com/kubedb/provisioner/releases/tag/v0.48.0)




## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.35.0](https://github.com/kubedb/proxysql/releases/tag/v0.35.0)

- [6398c68f](https://github.com/kubedb/proxysql/commit/6398c68fc) Prepare for release v0.35.0 (#360)
- [4a877c76](https://github.com/kubedb/proxysql/commit/4a877c763) Add Support for NetworkPolicy (#359)
- [be4c0011](https://github.com/kubedb/proxysql/commit/be4c00117) Use KIND v0.24.0 (#358)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.3.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.3.0)

- [4638e298](https://github.com/kubedb/rabbitmq/commit/4638e298) Prepare for release v0.3.0 (#48)
- [4cbfc8a6](https://github.com/kubedb/rabbitmq/commit/4cbfc8a6) Add network policy (#47)
- [be9417ca](https://github.com/kubedb/rabbitmq/commit/be9417ca) Use KIND v0.24.0 (#46)
- [e38f3735](https://github.com/kubedb/rabbitmq/commit/e38f3735) Fix dasboard svc port for http with TLS (#45)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.41.0](https://github.com/kubedb/redis/releases/tag/v0.41.0)

- [8283bb57](https://github.com/kubedb/redis/commit/8283bb57f) Prepare for release v0.41.0 (#560)
- [391e9b6d](https://github.com/kubedb/redis/commit/391e9b6d0) Add support for NetworkPolicy (#559)
- [55c69812](https://github.com/kubedb/redis/commit/55c698121) Add all utility installation-commands in makefile (#558)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.27.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.27.0)

- [1abf13d5](https://github.com/kubedb/redis-coordinator/commit/1abf13d5) Prepare for release v0.27.0 (#112)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.11.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.11.0)

- [e965823](https://github.com/kubedb/redis-restic-plugin/commit/e965823) Prepare for release v0.11.0 (#48)
- [bacecc1](https://github.com/kubedb/redis-restic-plugin/commit/bacecc1) Update helper method and refactor code (#47)
- [e1c4d47](https://github.com/kubedb/redis-restic-plugin/commit/e1c4d47) Add timeout for backup and restore (#45)
- [5dc4ec2](https://github.com/kubedb/redis-restic-plugin/commit/5dc4ec2) Use restic 0.17.1 (#44)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.35.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.35.0)

- [941e7a2a](https://github.com/kubedb/replication-mode-detector/commit/941e7a2a) Prepare for release v0.35.0 (#278)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.24.0](https://github.com/kubedb/schema-manager/releases/tag/v0.24.0)

- [bf0555a1](https://github.com/kubedb/schema-manager/commit/bf0555a1) Prepare for release v0.24.0 (#121)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.3.0](https://github.com/kubedb/singlestore/releases/tag/v0.3.0)

- [78c83807](https://github.com/kubedb/singlestore/commit/78c83807) Prepare for release v0.3.0 (#47)
- [56f2bf71](https://github.com/kubedb/singlestore/commit/56f2bf71) Add Network Policy (#46)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.3.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.3.0)

- [66d0da4](https://github.com/kubedb/singlestore-coordinator/commit/66d0da4) Prepare for release v0.3.0 (#27)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.6.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.6.0)

- [1ff32bb](https://github.com/kubedb/singlestore-restic-plugin/commit/1ff32bb) Prepare for release v0.6.0 (#26)
- [c242852](https://github.com/kubedb/singlestore-restic-plugin/commit/c242852) Update helper method and refactor code (#25)
- [0c2b37f](https://github.com/kubedb/singlestore-restic-plugin/commit/0c2b37f) Add timeout for backup and restore (#24)
- [38ef8d4](https://github.com/kubedb/singlestore-restic-plugin/commit/38ef8d4) Use restic 0.17.1 (#23)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.3.0](https://github.com/kubedb/solr/releases/tag/v0.3.0)

- [1e52e5c3](https://github.com/kubedb/solr/commit/1e52e5c3) Prepare for release v0.3.0 (#49)
- [7c22876e](https://github.com/kubedb/solr/commit/7c22876e) Add Network Policy (#48)
- [01c3c56c](https://github.com/kubedb/solr/commit/01c3c56c) Add tls for solr (#47)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.33.0](https://github.com/kubedb/tests/releases/tag/v0.33.0)

- [2f79ea6b](https://github.com/kubedb/tests/commit/2f79ea6b) Prepare for release v0.33.0 (#379)
- [abf5bbf2](https://github.com/kubedb/tests/commit/abf5bbf2) Add backup and restore CI for MongoDB (#376)
- [bffe9f51](https://github.com/kubedb/tests/commit/bffe9f51) Add Kafka metrics exporter tests (#369)
- [cc8b9e7c](https://github.com/kubedb/tests/commit/cc8b9e7c) Refactor Redis Tests (#360)
- [c22d1ab1](https://github.com/kubedb/tests/commit/c22d1ab1) Add Memcached Autoscaling e2e Test (#354)
- [dcc52bef](https://github.com/kubedb/tests/commit/dcc52bef) Change replica range (#370)
- [c64ef036](https://github.com/kubedb/tests/commit/c64ef036) Add Druid metrics exporter tests (#367)
- [3b43466a](https://github.com/kubedb/tests/commit/3b43466a) Update BackupConfiguration auto-trigger and fix in MongoDB, MariaDB (#365)
- [ce203ceb](https://github.com/kubedb/tests/commit/ce203ceb) Add mssql exporter (#353)
- [f734ec8c](https://github.com/kubedb/tests/commit/f734ec8c) Redis ssl fix exporter (#357)
- [0b8eca92](https://github.com/kubedb/tests/commit/0b8eca92) Add Pgbouncer tests for scaling (#340)
- [bc560566](https://github.com/kubedb/tests/commit/bc560566) Update Debugging printing function (#363)
- [f7b2eae1](https://github.com/kubedb/tests/commit/f7b2eae1) Kafka autoscaling (#359)
- [deec4378](https://github.com/kubedb/tests/commit/deec4378) Add MongoDB Backup-Restore test using KubeStash (#333)
- [e896042c](https://github.com/kubedb/tests/commit/e896042c) Add MariaDB backup-restore test using KubeStash (#337)
- [6b790e6b](https://github.com/kubedb/tests/commit/6b790e6b) Add Pgpool tests (TLS, Ops-requets, Autoscaler, Configuration) (#358)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.24.0](https://github.com/kubedb/ui-server/releases/tag/v0.24.0)




## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.24.0](https://github.com/kubedb/webhook-server/releases/tag/v0.24.0)




## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.3.0](https://github.com/kubedb/zookeeper/releases/tag/v0.3.0)

- [46451567](https://github.com/kubedb/zookeeper/commit/46451567) Prepare for release v0.3.0 (#41)
- [b0f16147](https://github.com/kubedb/zookeeper/commit/b0f16147) Add network policy (#40)
- [d4dd7f50](https://github.com/kubedb/zookeeper/commit/d4dd7f50) Add ZooKeeper Ops Request Changes (#32)
- [f4af8e59](https://github.com/kubedb/zookeeper/commit/f4af8e59) Use KIND v0.24.0 (#39)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.4.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.4.0)

- [c64eefc](https://github.com/kubedb/zookeeper-restic-plugin/commit/c64eefc) Prepare for release v0.4.0 (#17)
- [db28c83](https://github.com/kubedb/zookeeper-restic-plugin/commit/db28c83) Update helper method and refactor code (#16)
- [175447e](https://github.com/kubedb/zookeeper-restic-plugin/commit/175447e) Add timeout for backup and restore (#15)




