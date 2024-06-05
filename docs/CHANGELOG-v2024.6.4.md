---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2024.6.4
    name: Changelog-v2024.6.4
    parent: welcome
    weight: 20240604
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2024.6.4/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2024.6.4/
---

# KubeDB v2024.6.4 (2024-06-05)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.46.0](https://github.com/kubedb/apimachinery/releases/tag/v0.46.0)

- [acca053a](https://github.com/kubedb/apimachinery/commit/acca053ab) Update kubestash api
- [f64ef778](https://github.com/kubedb/apimachinery/commit/f64ef7785) Add pod placement policy to v1 pod template spec (#1233)
- [7cb05566](https://github.com/kubedb/apimachinery/commit/7cb055661) Use k8s client libs 1.30.1 (#1232)
- [faac2b8d](https://github.com/kubedb/apimachinery/commit/faac2b8dc) Use DeletionPolicy instead of TerminationPolicy in new dbs (#1231)
- [e387d0c6](https://github.com/kubedb/apimachinery/commit/e387d0c63) Use PodPlacementPolicy from v2 podTemplate (#1230)
- [74e7190f](https://github.com/kubedb/apimachinery/commit/74e7190f0) Set TLS Defaults for Microsoft SQL Server (#1226)
- [f50d9c10](https://github.com/kubedb/apimachinery/commit/f50d9c109) Refactor Clickhouse Webhook (#1228)
- [04949036](https://github.com/kubedb/apimachinery/commit/049490369) Add TLS constants and defaults for RabbitMQ (#1227)
- [d5eb1d50](https://github.com/kubedb/apimachinery/commit/d5eb1d507) Add Druid autoscaler API (#1219)
- [8f7ef3c8](https://github.com/kubedb/apimachinery/commit/8f7ef3c81) Add Druid ops-request API (#1208)
- [a0277a77](https://github.com/kubedb/apimachinery/commit/a0277a77e) Add Pgpool Autoscaler API (#1223)
- [b0c77ebe](https://github.com/kubedb/apimachinery/commit/b0c77ebe1) Add Pgpool OpsRequest API (#1209)
- [0d84725c](https://github.com/kubedb/apimachinery/commit/0d84725c8) Add RabbitMQ OpsRequests (#1225)
- [31cc9434](https://github.com/kubedb/apimachinery/commit/31cc9434f) pgbouncer reload config based on label (#1224)
- [b3797f47](https://github.com/kubedb/apimachinery/commit/b3797f47c) Add ClickHouse API (#1212)
- [46e50802](https://github.com/kubedb/apimachinery/commit/46e50802d) Add Kafka Schema Registry APIs (#1217)
- [5e6b27ed](https://github.com/kubedb/apimachinery/commit/5e6b27ed3) Add MSSQL Server TLS related APIs and helpers (#1218)
- [e1111569](https://github.com/kubedb/apimachinery/commit/e11115697) Add SingleStore AutoScaler API (#1213)
- [85c1f2a6](https://github.com/kubedb/apimachinery/commit/85c1f2a6a) Add SingleStore Ops-Request API (#1211)
- [2c4d34e0](https://github.com/kubedb/apimachinery/commit/2c4d34e07) Updated Memcached API (#1214)
- [2321968d](https://github.com/kubedb/apimachinery/commit/2321968de) Update druid API for simplifying YAML (#1222)
- [dfb99f1a](https://github.com/kubedb/apimachinery/commit/dfb99f1ab) Set default scalingRules (#1220)
- [0744dd0c](https://github.com/kubedb/apimachinery/commit/0744dd0c9) Allow only one database for pgbouncer connection pooling (#1210)
- [08fb210b](https://github.com/kubedb/apimachinery/commit/08fb210be) Fix error check for ferretdb webhook (#1221)
- [4274e807](https://github.com/kubedb/apimachinery/commit/4274e8077) Update auditor library (#1216)
- [7fb9b8c9](https://github.com/kubedb/apimachinery/commit/7fb9b8c98) Extract databaseInfo from mssql (#1203)
- [945f983c](https://github.com/kubedb/apimachinery/commit/945f983c3) Add `syncUsers` field to pgbouncer api (#1206)
- [bf3d0581](https://github.com/kubedb/apimachinery/commit/bf3d0581d) Update kmodules/client-go deps
- [a74eb3db](https://github.com/kubedb/apimachinery/commit/a74eb3db5) Update ZooKeeper CRD (#1207)
- [bcfb3b3d](https://github.com/kubedb/apimachinery/commit/bcfb3b3da) Add Security Context (#1204)
- [62470fd4](https://github.com/kubedb/apimachinery/commit/62470fd41) Add Pgpool AppBinding (#1205)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.31.0](https://github.com/kubedb/autoscaler/releases/tag/v0.31.0)

- [53c1e362](https://github.com/kubedb/autoscaler/commit/53c1e362) Prepare for release v0.31.0 (#209)
- [41133bbe](https://github.com/kubedb/autoscaler/commit/41133bbe) Use k8s 1.30 client libs (#208)
- [72b0ef60](https://github.com/kubedb/autoscaler/commit/72b0ef60) Add Pgpool Autoscaler (#207)
- [dfbc3cfe](https://github.com/kubedb/autoscaler/commit/dfbc3cfe) Add Druid Autoscaler (#206)
- [bd1b970b](https://github.com/kubedb/autoscaler/commit/bd1b970b) Add SingleStore Autoscaler (#204)
- [cbcf5481](https://github.com/kubedb/autoscaler/commit/cbcf5481) Update auditor library (#205)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.46.0](https://github.com/kubedb/cli/releases/tag/v0.46.0)

- [1399ee75](https://github.com/kubedb/cli/commit/1399ee75) Prepare for release v0.46.0 (#769)
- [de6ea578](https://github.com/kubedb/cli/commit/de6ea578) Use k8s 1.30 client libs (#768)
- [fa28c173](https://github.com/kubedb/cli/commit/fa28c173) Update auditor library (#767)
- [c12d92b1](https://github.com/kubedb/cli/commit/c12d92b1) Add --only-archiver flag to pause/resume DB archiver (#756)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.1.0](https://github.com/kubedb/clickhouse/releases/tag/v0.1.0)

- [4065f54](https://github.com/kubedb/clickhouse/commit/4065f54) Prepare for release v0.1.0 (#2)
- [7e3dd87](https://github.com/kubedb/clickhouse/commit/7e3dd87) Add Support for Provisioning (#1)
- [5d37918](https://github.com/kubedb/clickhouse/commit/5d37918) Add ClickHouseVersion API
- [d5c86a3](https://github.com/kubedb/clickhouse/commit/d5c86a3) Add License
- [edd9534](https://github.com/kubedb/clickhouse/commit/edd9534) Add API and controller Skeleton



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.1.0](https://github.com/kubedb/crd-manager/releases/tag/v0.1.0)

- [da4101ef](https://github.com/kubedb/crd-manager/commit/da4101ef) Prepare for release v0.1.0 (#33)
- [c762bfce](https://github.com/kubedb/crd-manager/commit/c762bfce) Use k8s 1.30 client libs (#32)
- [eb1ed23f](https://github.com/kubedb/crd-manager/commit/eb1ed23f) Add Schema Registry, Ops-Requests and Autoscalers CRD (#30)
- [0e329d08](https://github.com/kubedb/crd-manager/commit/0e329d08) Add ClickHouse (#28)
- [5ad04861](https://github.com/kubedb/crd-manager/commit/5ad04861) Update auditor library (#29)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.22.0](https://github.com/kubedb/dashboard/releases/tag/v0.22.0)

- [40f22f3e](https://github.com/kubedb/dashboard/commit/40f22f3e) Prepare for release v0.22.0 (#117)
- [7e3f9a78](https://github.com/kubedb/dashboard/commit/7e3f9a78) Use k8s 1.30 client libs (#116)
- [6524a4b6](https://github.com/kubedb/dashboard/commit/6524a4b6) Update auditor library (#115)
- [8b7b25c5](https://github.com/kubedb/dashboard/commit/8b7b25c5) Remove license checker for webhook-server (#114)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.4.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.4.0)

- [6a55ad2](https://github.com/kubedb/dashboard-restic-plugin/commit/6a55ad2) Prepare for release v0.4.0 (#12)
- [dd3e588](https://github.com/kubedb/dashboard-restic-plugin/commit/dd3e588) Use k8s 1.30 client libs (#11)
- [8a6f5fc](https://github.com/kubedb/dashboard-restic-plugin/commit/8a6f5fc) Update auditor library (#10)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.1.0](https://github.com/kubedb/db-client-go/releases/tag/v0.1.0)

- [19480f10](https://github.com/kubedb/db-client-go/commit/19480f10) Prepare for release v0.1.0 (#115)
- [1942b7bd](https://github.com/kubedb/db-client-go/commit/1942b7bd) client for health check & multiple user reload for pgBouncer (#107)
- [277b3fb6](https://github.com/kubedb/db-client-go/commit/277b3fb6) Use k8s 1.30 client libs (#114)
- [78f5e4c4](https://github.com/kubedb/db-client-go/commit/78f5e4c4) Update rabbitmq client methods (#113)
- [8beb95c5](https://github.com/kubedb/db-client-go/commit/8beb95c5) Update apimachinery module
- [f25da8ec](https://github.com/kubedb/db-client-go/commit/f25da8ec) Add ClickHouse DB Client (#112)
- [8c9132a7](https://github.com/kubedb/db-client-go/commit/8c9132a7) Add RabbitMQ PubSub methods (#105)
- [ed82b008](https://github.com/kubedb/db-client-go/commit/ed82b008) Add Kafka Schema Registry Client (#110)
- [9c578aba](https://github.com/kubedb/db-client-go/commit/9c578aba) Add MSSQL Server TLS  config  (#111)
- [94f6f276](https://github.com/kubedb/db-client-go/commit/94f6f276) Update auditor library (#109)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.1.0](https://github.com/kubedb/druid/releases/tag/v0.1.0)

- [bbacb4b](https://github.com/kubedb/druid/commit/bbacb4b) Prepare for release v0.1.0 (#29)
- [98c33ef](https://github.com/kubedb/druid/commit/98c33ef) Use k8s 1.30 client libs (#27)
- [6bc1c42](https://github.com/kubedb/druid/commit/6bc1c42) Use PodPlacementPolicy from v2 podTemplate (#26)
- [0358de3](https://github.com/kubedb/druid/commit/0358de3) Fix druid for ops-req and simplifying YAML (#25)
- [c6d8828](https://github.com/kubedb/druid/commit/c6d8828) Update auditor library (#24)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.46.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.46.0)

- [3c94bb71](https://github.com/kubedb/elasticsearch/commit/3c94bb710) Prepare for release v0.46.0 (#722)
- [0c25a1a9](https://github.com/kubedb/elasticsearch/commit/0c25a1a9e) Use k8s 1.30 client libs (#721)
- [c2fa3aca](https://github.com/kubedb/elasticsearch/commit/c2fa3acae) Add new version for nightly CI run (#720)
- [1714633c](https://github.com/kubedb/elasticsearch/commit/1714633c4) Update auditor library (#719)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.9.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.9.0)

- [adc1b75](https://github.com/kubedb/elasticsearch-restic-plugin/commit/adc1b75) Prepare for release v0.9.0 (#33)
- [6d90baf](https://github.com/kubedb/elasticsearch-restic-plugin/commit/6d90baf) Use k8s 1.30 client libs (#32)
- [0f59810](https://github.com/kubedb/elasticsearch-restic-plugin/commit/0f59810) Update auditor library (#31)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.1.0](https://github.com/kubedb/ferretdb/releases/tag/v0.1.0)

- [334c999e](https://github.com/kubedb/ferretdb/commit/334c999e) Prepare for release v0.1.0 (#29)
- [64f92533](https://github.com/kubedb/ferretdb/commit/64f92533) Use k8s 1.30 client libs (#28)
- [f8ba908f](https://github.com/kubedb/ferretdb/commit/f8ba908f) Update TerminationPolicy to DeletionPolicy (#27)
- [34042883](https://github.com/kubedb/ferretdb/commit/34042883) Move TLS to OpsManager,refactor appbinging,makefile (#26)
- [3bf2bd3c](https://github.com/kubedb/ferretdb/commit/3bf2bd3c) Update auditor library (#24)
- [0168006f](https://github.com/kubedb/ferretdb/commit/0168006f) Use kbClient; Fix patching bug; Use custom predicate for postgres watcher (#23)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2024.6.4](https://github.com/kubedb/installer/releases/tag/v2024.6.4)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.17.0](https://github.com/kubedb/kafka/releases/tag/v0.17.0)

- [1214efb9](https://github.com/kubedb/kafka/commit/1214efb9) Prepare for release v0.17.0 (#94)
- [3c354bb9](https://github.com/kubedb/kafka/commit/3c354bb9) Use k8s 1.30 client libs (#93)
- [25c78b67](https://github.com/kubedb/kafka/commit/25c78b67) Update TerminationPolicy to DeletionPolicy (#92)
- [16eece63](https://github.com/kubedb/kafka/commit/16eece63) Add Kafka Schema Registry (#91)
- [b41a6efe](https://github.com/kubedb/kafka/commit/b41a6efe) Update auditor library (#90)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.9.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.9.0)

- [473e302](https://github.com/kubedb/kubedb-manifest-plugin/commit/473e302) Prepare for release v0.9.0 (#56)
- [a85da9a](https://github.com/kubedb/kubedb-manifest-plugin/commit/a85da9a) Fix nil pointer ref for manifest options (#55)
- [ad0bee2](https://github.com/kubedb/kubedb-manifest-plugin/commit/ad0bee2) Add default namespace for restore (#54)
- [9a503e1](https://github.com/kubedb/kubedb-manifest-plugin/commit/9a503e1) Add Support for Cross Namespace Restore (#45)
- [e02604f](https://github.com/kubedb/kubedb-manifest-plugin/commit/e02604f) Use k8s 1.30 client libs (#53)
- [93633f3](https://github.com/kubedb/kubedb-manifest-plugin/commit/93633f3) Update auditor library (#52)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.30.0](https://github.com/kubedb/mariadb/releases/tag/v0.30.0)

- [a382c4b2](https://github.com/kubedb/mariadb/commit/a382c4b2b) Prepare for release v0.30.0 (#270)
- [956d17b6](https://github.com/kubedb/mariadb/commit/956d17b65) Use k8s 1.30 client libs (#269)
- [c6531c18](https://github.com/kubedb/mariadb/commit/c6531c18e) Update auditor library (#268)
- [4add6182](https://github.com/kubedb/mariadb/commit/4add61822) Add Cloud, TLS Support for Archiver Backup and Restore



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.6.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.6.0)

- [dd0a4eb](https://github.com/kubedb/mariadb-archiver/commit/dd0a4eb) Prepare for release v0.6.0 (#19)
- [1a5d018](https://github.com/kubedb/mariadb-archiver/commit/1a5d018) Use k8s 1.30 client libs (#18)
- [33c9ef9](https://github.com/kubedb/mariadb-archiver/commit/33c9ef9) Update auditor library (#17)
- [de771cb](https://github.com/kubedb/mariadb-archiver/commit/de771cb) Add Cloud, TLS Support for Archiver Backup and Restore



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.26.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.26.0)

- [4160a9a4](https://github.com/kubedb/mariadb-coordinator/commit/4160a9a4) Prepare for release v0.26.0 (#118)
- [19af6050](https://github.com/kubedb/mariadb-coordinator/commit/19af6050) Use k8s 1.30 client libs (#117)
- [a90ae803](https://github.com/kubedb/mariadb-coordinator/commit/a90ae803) Update auditor library (#116)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.6.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.6.0)

- [24bc92e](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/24bc92e) Prepare for release v0.6.0 (#22)
- [7e532d3](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/7e532d3) Use k8s 1.30 client libs (#21)
- [57e689b](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/57e689b) Update auditor library (#20)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.4.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.4.0)

- [bf9cc17](https://github.com/kubedb/mariadb-restic-plugin/commit/bf9cc17) Prepare for release v0.4.0 (#14)
- [cdb7954](https://github.com/kubedb/mariadb-restic-plugin/commit/cdb7954) Update target name and namespace ref for restore (#13)
- [52f4777](https://github.com/kubedb/mariadb-restic-plugin/commit/52f4777) Use k8s 1.30 client libs (#12)
- [813f1c1](https://github.com/kubedb/mariadb-restic-plugin/commit/813f1c1) Wait for db provisioned (#10)
- [e5d053a](https://github.com/kubedb/mariadb-restic-plugin/commit/e5d053a) Update auditor library (#11)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.39.0](https://github.com/kubedb/memcached/releases/tag/v0.39.0)

- [410f3b36](https://github.com/kubedb/memcached/commit/410f3b365) Prepare for release v0.39.0 (#448)
- [918a89c6](https://github.com/kubedb/memcached/commit/918a89c61) Use k8s 1.30 client libs (#447)
- [cf2d7206](https://github.com/kubedb/memcached/commit/cf2d72064) Fix Ops-Manager Issue (#446)
- [bdd09321](https://github.com/kubedb/memcached/commit/bdd093216) Add Custom ConfigSecret (#443)
- [bde05333](https://github.com/kubedb/memcached/commit/bde05333d) Update auditor library (#444)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.39.0](https://github.com/kubedb/mongodb/releases/tag/v0.39.0)

- [f7d28ac2](https://github.com/kubedb/mongodb/commit/f7d28ac21) Prepare for release v0.39.0 (#634)
- [4160721f](https://github.com/kubedb/mongodb/commit/4160721f1) MongoDB Archiver shard (#631)
- [6d041566](https://github.com/kubedb/mongodb/commit/6d041566c) Use k8s 1.30 client libs (#633)
- [4a73b0bc](https://github.com/kubedb/mongodb/commit/4a73b0bc3) Update auditor library (#632)
- [52f125ad](https://github.com/kubedb/mongodb/commit/52f125adb) Add archiver support for azure (#630)
- [0bddf18f](https://github.com/kubedb/mongodb/commit/0bddf18f8) Add support for NFS in MongoDB Archiver (#629)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.7.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.7.0)

- [abce7be](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/abce7be) Prepare for release v0.7.0 (#27)
- [cc35c3c](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/cc35c3c) Use k8s 1.30 client libs (#26)
- [3fec803](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/3fec803) Add support for Shard (#18)
- [db7c7a7](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/db7c7a7) Update auditor library (#25)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.9.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.9.0)

- [05c9180](https://github.com/kubedb/mongodb-restic-plugin/commit/05c9180) Prepare for release v0.9.0 (#48)
- [b4e7e31](https://github.com/kubedb/mongodb-restic-plugin/commit/b4e7e31) Update target name and namespace ref for restore (#47)
- [e45acfe](https://github.com/kubedb/mongodb-restic-plugin/commit/e45acfe) Fix specific component restore support (#43)
- [f88fdef](https://github.com/kubedb/mongodb-restic-plugin/commit/f88fdef) Use k8s 1.30 client libs (#46)
- [21cfc60](https://github.com/kubedb/mongodb-restic-plugin/commit/21cfc60) Wait for db provisioned (#44)
- [2a1c819](https://github.com/kubedb/mongodb-restic-plugin/commit/2a1c819) Update auditor library (#45)



## [kubedb/mssql](https://github.com/kubedb/mssql)

### [v0.1.0](https://github.com/kubedb/mssql/releases/tag/v0.1.0)




## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.1.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.1.0)

- [44cac816](https://github.com/kubedb/mssql-coordinator/commit/44cac816) Prepare for release v0.1.0 (#6)
- [8c8c4f41](https://github.com/kubedb/mssql-coordinator/commit/8c8c4f41) Update helpers and logs (#5)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.39.0](https://github.com/kubedb/mysql/releases/tag/v0.39.0)

- [9fb15d4d](https://github.com/kubedb/mysql/commit/9fb15d4dd) Prepare for release v0.39.0 (#626)
- [b8591a75](https://github.com/kubedb/mysql/commit/b8591a757) Use k8s 1.30 client libs (#625)
- [11d498b8](https://github.com/kubedb/mysql/commit/11d498b82) Update auditor library (#624)
- [297e6f89](https://github.com/kubedb/mysql/commit/297e6f899) Refactor ENV and Func Name



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.7.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.7.0)

- [9aec804](https://github.com/kubedb/mysql-archiver/commit/9aec804) Prepare for release v0.7.0 (#33)
- [da8412e](https://github.com/kubedb/mysql-archiver/commit/da8412e) Use k8s 1.30 client libs (#32)
- [5a5c184](https://github.com/kubedb/mysql-archiver/commit/5a5c184) Update auditor library (#31)
- [bd8ed03](https://github.com/kubedb/mysql-archiver/commit/bd8ed03) Refactor ENV and Func Name
- [2817c31](https://github.com/kubedb/mysql-archiver/commit/2817c31) Refactor ENV and func



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.24.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.24.0)

- [84966dd7](https://github.com/kubedb/mysql-coordinator/commit/84966dd7) Prepare for release v0.24.0 (#113)
- [ca8f6759](https://github.com/kubedb/mysql-coordinator/commit/ca8f6759) Use k8s 1.30 client libs (#112)
- [e292c238](https://github.com/kubedb/mysql-coordinator/commit/e292c238) Update auditor library (#111)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.7.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.7.0)

- [b427a30](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/b427a30) Prepare for release v0.7.0 (#20)
- [1bbc1a9](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/1bbc1a9) Use k8s 1.30 client libs (#19)
- [a897efa](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/a897efa) Update auditor library (#18)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.9.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.9.0)

- [dd3c711](https://github.com/kubedb/mysql-restic-plugin/commit/dd3c711) Prepare for release v0.9.0 (#42)
- [7adcbdb](https://github.com/kubedb/mysql-restic-plugin/commit/7adcbdb) Update target name and namespace ref for restore (#41)
- [7f6a752](https://github.com/kubedb/mysql-restic-plugin/commit/7f6a752) Use k8s 1.30 client libs (#40)
- [5ef953f](https://github.com/kubedb/mysql-restic-plugin/commit/5ef953f) Update auditor library (#39)
- [e108ece](https://github.com/kubedb/mysql-restic-plugin/commit/e108ece) Wait for db provisioned (#38)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.24.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.24.0)

- [f1f471e](https://github.com/kubedb/mysql-router-init/commit/f1f471e) Use k8s 1.30 client libs (#44)
- [10d3c4b](https://github.com/kubedb/mysql-router-init/commit/10d3c4b) Update auditor library (#43)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.33.0](https://github.com/kubedb/ops-manager/releases/tag/v0.33.0)

- [9e726183](https://github.com/kubedb/ops-manager/commit/9e726183f) Prepare for release v0.33.0 (#586)
- [8febc50f](https://github.com/kubedb/ops-manager/commit/8febc50fe) Update conditions in Redis and Redis Sentinel retries (#585)
- [8f8a8c84](https://github.com/kubedb/ops-manager/commit/8f8a8c846) Update conditions in Elasticsearch retries (#581)
- [0c8cfd83](https://github.com/kubedb/ops-manager/commit/0c8cfd833) Update conditions in RabbitMQ retries (#584)
- [4c0cb06a](https://github.com/kubedb/ops-manager/commit/4c0cb06a4) Update Conditions in SinglStore, ProxySQL and MySQL retries (#578)
- [9e08e0a5](https://github.com/kubedb/ops-manager/commit/9e08e0a53) Update conditions for Druid (#583)
- [3d5ae8ae](https://github.com/kubedb/ops-manager/commit/3d5ae8ae4) Update Conditions in MariaDB retries (#582)
- [274275c2](https://github.com/kubedb/ops-manager/commit/274275c29) Improve Logging (#580)
- [ca93648c](https://github.com/kubedb/ops-manager/commit/ca93648c2) Update conditions for Pgpool (#579)
- [f4d4d49b](https://github.com/kubedb/ops-manager/commit/f4d4d49b8) Improve Kafka Ops Request logging (#577)
- [04e32a77](https://github.com/kubedb/ops-manager/commit/04e32a776) Update conditions in MongoDB retries (#576)
- [2eaa1742](https://github.com/kubedb/ops-manager/commit/2eaa17422) Use k8s 1.30 client libs (#574)
- [c6f51c1b](https://github.com/kubedb/ops-manager/commit/c6f51c1b5) Improve logging (#575)
- [b47c75db](https://github.com/kubedb/ops-manager/commit/b47c75dbb) Add RabbitMQ OpsRequest and TLS (#573)
- [1090b7b9](https://github.com/kubedb/ops-manager/commit/1090b7b94) Add Singlestore Ops-manager (#562)
- [c4187163](https://github.com/kubedb/ops-manager/commit/c41871638) Add Druid vertical scaling and volume expansion ops-requests (#564)
- [bbb39eb1](https://github.com/kubedb/ops-manager/commit/bbb39eb19) Add Pgpool ops request (vertical scaling, reconfigure) (#565)
- [2502ea5c](https://github.com/kubedb/ops-manager/commit/2502ea5cc) Add FerretDB TLS (#572)
- [5feab765](https://github.com/kubedb/ops-manager/commit/5feab765c) Add MSSQL Server TLS (#571)
- [5a2a61a7](https://github.com/kubedb/ops-manager/commit/5a2a61a7b) Add Memcached's Ops Request (Restart + Vertical Scaling + Reconfiguration) (#568)
- [252c0f9c](https://github.com/kubedb/ops-manager/commit/252c0f9c3) Update auditor library (#570)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.33.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.33.0)

- [9f4ad9f1](https://github.com/kubedb/percona-xtradb/commit/9f4ad9f19) Prepare for release v0.33.0 (#368)
- [51457c7f](https://github.com/kubedb/percona-xtradb/commit/51457c7f9) Use k8s 1.30 client libs (#367)
- [b2f82795](https://github.com/kubedb/percona-xtradb/commit/b2f82795f) Update auditor library (#366)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.19.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.19.0)

- [502d4f25](https://github.com/kubedb/percona-xtradb-coordinator/commit/502d4f25) Prepare for release v0.19.0 (#73)
- [c89a34cb](https://github.com/kubedb/percona-xtradb-coordinator/commit/c89a34cb) Use k8s 1.30 client libs (#72)
- [a13e477d](https://github.com/kubedb/percona-xtradb-coordinator/commit/a13e477d) Update auditor library (#71)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.30.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.30.0)

- [bf7a8900](https://github.com/kubedb/pg-coordinator/commit/bf7a8900) Prepare for release v0.30.0 (#166)
- [90f77d71](https://github.com/kubedb/pg-coordinator/commit/90f77d71) Use k8s 1.30 client libs (#165)
- [a6fe13f5](https://github.com/kubedb/pg-coordinator/commit/a6fe13f5) Update auditor library (#164)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.33.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.33.0)

- [4ba7070c](https://github.com/kubedb/pgbouncer/commit/4ba7070c) Prepare for release v0.33.0 (#333)
- [aec17eaf](https://github.com/kubedb/pgbouncer/commit/aec17eaf) Add health check (#327)
- [87e09f30](https://github.com/kubedb/pgbouncer/commit/87e09f30) Use k8s 1.30 client libs (#332)
- [346a5688](https://github.com/kubedb/pgbouncer/commit/346a5688) pgbouncer support only one postgres DB (#330)
- [be77042b](https://github.com/kubedb/pgbouncer/commit/be77042b) Update auditor library (#331)
- [d6cde1dd](https://github.com/kubedb/pgbouncer/commit/d6cde1dd) Multiple user support (#328)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.1.0](https://github.com/kubedb/pgpool/releases/tag/v0.1.0)

- [3979efd1](https://github.com/kubedb/pgpool/commit/3979efd1) Prepare for release v0.1.0 (#33)
- [7f032308](https://github.com/kubedb/pgpool/commit/7f032308) Use k8s 1.30 client libs (#32)
- [7dc03c00](https://github.com/kubedb/pgpool/commit/7dc03c00) Update deletion and pod placement policy (#31)
- [cc425633](https://github.com/kubedb/pgpool/commit/cc425633) Add pcp port and pdb (#30)
- [f4d0c0f9](https://github.com/kubedb/pgpool/commit/f4d0c0f9) Update auditor library (#29)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.46.0](https://github.com/kubedb/postgres/releases/tag/v0.46.0)

- [575d302b](https://github.com/kubedb/postgres/commit/575d302b2) Prepare for release v0.46.0 (#736)
- [28b5b86e](https://github.com/kubedb/postgres/commit/28b5b86e6) Use k8s 1.30 client libs (#735)
- [754de55c](https://github.com/kubedb/postgres/commit/754de55c2) Add remote replica support for pg 13,14 (#734)
- [34285094](https://github.com/kubedb/postgres/commit/342850945) Remove wait until from operator (#733)
- [8b995cb6](https://github.com/kubedb/postgres/commit/8b995cb6f) Update auditor library (#732)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.7.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.7.0)

- [d3ee4020](https://github.com/kubedb/postgres-archiver/commit/d3ee4020) Prepare for release v0.7.0 (#31)
- [ec608c78](https://github.com/kubedb/postgres-archiver/commit/ec608c78) Use k8s 1.30 client libs (#30)
- [0775a40b](https://github.com/kubedb/postgres-archiver/commit/0775a40b) Update auditor library (#29)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.7.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.7.0)

- [1246683](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/1246683) Prepare for release v0.7.0 (#29)
- [bc60a8e](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/bc60a8e) Use k8s 1.30 client libs (#28)
- [56106ab](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/56106ab) Update auditor library (#27)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.9.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.9.0)

- [07366ba](https://github.com/kubedb/postgres-restic-plugin/commit/07366ba) Prepare for release v0.9.0 (#36)
- [7e2c8cc](https://github.com/kubedb/postgres-restic-plugin/commit/7e2c8cc) Wait for db provisioned (#33)
- [fb986a7](https://github.com/kubedb/postgres-restic-plugin/commit/fb986a7) Use k8s 1.30 client libs (#35)
- [5d7b929](https://github.com/kubedb/postgres-restic-plugin/commit/5d7b929) Update auditor library (#34)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.8.0](https://github.com/kubedb/provider-aws/releases/tag/v0.8.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.8.0](https://github.com/kubedb/provider-azure/releases/tag/v0.8.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.8.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.8.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.46.0](https://github.com/kubedb/provisioner/releases/tag/v0.46.0)




## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.33.0](https://github.com/kubedb/proxysql/releases/tag/v0.33.0)

- [b0cb53d8](https://github.com/kubedb/proxysql/commit/b0cb53d82) Prepare for release v0.33.0 (#345)
- [207a8e31](https://github.com/kubedb/proxysql/commit/207a8e316) Use k8s 1.30 client libs (#344)
- [d48ceec9](https://github.com/kubedb/proxysql/commit/d48ceec9b) Update auditor library (#343)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.1.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.1.0)

- [e5b4e89](https://github.com/kubedb/rabbitmq/commit/e5b4e89) Prepare for release v0.1.0 (#32)
- [ec07f9e](https://github.com/kubedb/rabbitmq/commit/ec07f9e) Use k8s 1.30 client libs (#31)
- [2a564ae](https://github.com/kubedb/rabbitmq/commit/2a564ae) Use DeletionPolicy instead of TerminationPolicy
- [e9ff804](https://github.com/kubedb/rabbitmq/commit/e9ff804) Use PodPlacementPolicy from v2 podTemplate
- [979f9c0](https://github.com/kubedb/rabbitmq/commit/979f9c0) Add support for TLS (#30)
- [4d5b3b2](https://github.com/kubedb/rabbitmq/commit/4d5b3b2) Update auditor library (#29)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.25.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.25.0)

- [0afe9b48](https://github.com/kubedb/redis-coordinator/commit/0afe9b48) Prepare for release v0.25.0 (#104)
- [519a3d70](https://github.com/kubedb/redis-coordinator/commit/519a3d70) Use k8s 1.30 client libs (#103)
- [585fd8e8](https://github.com/kubedb/redis-coordinator/commit/585fd8e8) Update auditor library (#102)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.9.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.9.0)

- [3f308fb](https://github.com/kubedb/redis-restic-plugin/commit/3f308fb) Prepare for release v0.9.0 (#35)
- [c5ca569](https://github.com/kubedb/redis-restic-plugin/commit/c5ca569) Use k8s 1.30 client libs (#34)
- [39f43d9](https://github.com/kubedb/redis-restic-plugin/commit/39f43d9) Update auditor library (#33)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.33.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.33.0)

- [7d8bde2f](https://github.com/kubedb/replication-mode-detector/commit/7d8bde2f) Prepare for release v0.33.0 (#270)
- [4107fe7d](https://github.com/kubedb/replication-mode-detector/commit/4107fe7d) Use k8s 1.30 client libs (#269)
- [46a338bc](https://github.com/kubedb/replication-mode-detector/commit/46a338bc) Update auditor library (#268)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.22.0](https://github.com/kubedb/schema-manager/releases/tag/v0.22.0)

- [8ec78e1f](https://github.com/kubedb/schema-manager/commit/8ec78e1f) Prepare for release v0.22.0 (#113)
- [3dd18fbe](https://github.com/kubedb/schema-manager/commit/3dd18fbe) Use k8s 1.30 client libs (#112)
- [d290e4dc](https://github.com/kubedb/schema-manager/commit/d290e4dc) Update auditor library (#111)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.1.0](https://github.com/kubedb/singlestore/releases/tag/v0.1.0)

- [e98ab47](https://github.com/kubedb/singlestore/commit/e98ab47) Prepare for release v0.1.0 (#34)
- [246ad81](https://github.com/kubedb/singlestore/commit/246ad81) Add Support for DB phase change while restoring using KubeStash (#33)
- [1a1d0ce](https://github.com/kubedb/singlestore/commit/1a1d0ce) Use k8s 1.30 client libs (#32)
- [ad24f22](https://github.com/kubedb/singlestore/commit/ad24f22) Change for Ops-manager and Standalone Custom Config (#30)
- [50cc740](https://github.com/kubedb/singlestore/commit/50cc740) Update auditor library (#29)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.1.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.1.0)

- [5b25f8a](https://github.com/kubedb/singlestore-coordinator/commit/5b25f8a) Prepare for release v0.1.0 (#18)
- [a6a8f8d](https://github.com/kubedb/singlestore-coordinator/commit/a6a8f8d) Use k8s 1.30 client libs (#17)
- [84ea1b8](https://github.com/kubedb/singlestore-coordinator/commit/84ea1b8) Fixed Petset Name Typo (#16)
- [ea90f1b](https://github.com/kubedb/singlestore-coordinator/commit/ea90f1b) Update deps



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.4.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.4.0)

- [cac9e60](https://github.com/kubedb/singlestore-restic-plugin/commit/cac9e60) Prepare for release v0.4.0 (#14)
- [db06087](https://github.com/kubedb/singlestore-restic-plugin/commit/db06087) Use k8s 1.30 client libs (#13)
- [ecff2b2](https://github.com/kubedb/singlestore-restic-plugin/commit/ecff2b2) Update auditor library (#12)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.1.0](https://github.com/kubedb/solr/releases/tag/v0.1.0)

- [de0665a](https://github.com/kubedb/solr/commit/de0665a) Prepare for release v0.1.0 (#32)
- [4a87cd1](https://github.com/kubedb/solr/commit/4a87cd1) Use podPlacementPolicy from v2 template and use deletionPolicy instead of terminationPolicy. (#30)
- [bd23006](https://github.com/kubedb/solr/commit/bd23006) Add petset to run tests. (#29)
- [11e98aa](https://github.com/kubedb/solr/commit/11e98aa) Update auditor library (#28)
- [f2854c4](https://github.com/kubedb/solr/commit/f2854c4) Use version specific bootstrap configurations (#27)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.31.0](https://github.com/kubedb/tests/releases/tag/v0.31.0)

- [c6f2e4d4](https://github.com/kubedb/tests/commit/c6f2e4d4) Prepare for release v0.31.0 (#325)
- [5eb32303](https://github.com/kubedb/tests/commit/5eb32303) Use k8s 1.30 client libs (#324)
- [ff5cbff3](https://github.com/kubedb/tests/commit/ff5cbff3) Add solr tests. (#305)
- [5f3a802e](https://github.com/kubedb/tests/commit/5f3a802e) Update auditor library (#322)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.22.0](https://github.com/kubedb/ui-server/releases/tag/v0.22.0)

- [22148e44](https://github.com/kubedb/ui-server/commit/22148e44) Prepare for release v0.22.0 (#120)
- [338095fc](https://github.com/kubedb/ui-server/commit/338095fc) Use k8s 1.30 client libs (#119)
- [eb426684](https://github.com/kubedb/ui-server/commit/eb426684) Update auditor library (#118)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.22.0](https://github.com/kubedb/webhook-server/releases/tag/v0.22.0)

- [648876fc](https://github.com/kubedb/webhook-server/commit/648876fc) Prepare for release v0.22.0 (#110)
- [90146eaf](https://github.com/kubedb/webhook-server/commit/90146eaf) Add ClickHouse, SchemaRegistry, Ops, Autoscaler Webhook (#109)
- [0c1b6f0f](https://github.com/kubedb/webhook-server/commit/0c1b6f0f) Use k8s 1.30 client libs (#108)
- [d5b3a328](https://github.com/kubedb/webhook-server/commit/d5b3a328) Update auditor library (#106)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.1.0](https://github.com/kubedb/zookeeper/releases/tag/v0.1.0)

- [0f0958b2](https://github.com/kubedb/zookeeper/commit/0f0958b2) Prepare for release v0.1.0 (#28)
- [a0c3a9c3](https://github.com/kubedb/zookeeper/commit/a0c3a9c3) Use k8s 1.30 client libs (#27)
- [41f23f5e](https://github.com/kubedb/zookeeper/commit/41f23f5e) Update PodPlacementPolicy (#26)
- [0b3c259f](https://github.com/kubedb/zookeeper/commit/0b3c259f) Add support for PodDisruptionBudget (#25)
- [4139742c](https://github.com/kubedb/zookeeper/commit/4139742c) Update auditor library (#24)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.2.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.2.0)

- [5953888](https://github.com/kubedb/zookeeper-restic-plugin/commit/5953888) Prepare for release v0.2.0 (#8)
- [33a7841](https://github.com/kubedb/zookeeper-restic-plugin/commit/33a7841) Update Makefile




