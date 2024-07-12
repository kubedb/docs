---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2024.7.11-rc.1
    name: Changelog-v2024.7.11-rc.1
    parent: welcome
    weight: 20240711
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2024.7.11-rc.1/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2024.7.11-rc.1/
---

# KubeDB v2024.7.11-rc.1 (2024-07-12)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.47.0-rc.1](https://github.com/kubedb/apimachinery/releases/tag/v0.47.0-rc.1)

- [d3914aa5](https://github.com/kubedb/apimachinery/commit/d3914aa5e) Update deps
- [79ca732b](https://github.com/kubedb/apimachinery/commit/79ca732bd) Update Redis Ops Master -> Shards (#1255)
- [0fdb8074](https://github.com/kubedb/apimachinery/commit/0fdb8074c) Add UI chart info & remove status.gateway from db (#1256)
- [88ec29e7](https://github.com/kubedb/apimachinery/commit/88ec29e74) Set the default resources correctly (#1253)
- [cea4a328](https://github.com/kubedb/apimachinery/commit/cea4a328b) update scaling field for pgbouncer ops-request (#1244)
- [9809d94e](https://github.com/kubedb/apimachinery/commit/9809d94ee) Add API for Solr Restart OpsRequest (#1247)
- [abc86bb9](https://github.com/kubedb/apimachinery/commit/abc86bb90) Fix druid by adding Postgres as metadata storage type (#1252)
- [f8063159](https://github.com/kubedb/apimachinery/commit/f8063159a) Rename Master -> Shards in Redis (#1249)
- [5760b1e2](https://github.com/kubedb/apimachinery/commit/5760b1e2e) Fix phase tests and  use ensure container utilities (#1250)
- [d03a54a6](https://github.com/kubedb/apimachinery/commit/d03a54a6c) Report control plane and worker node stats
- [2e35ad03](https://github.com/kubedb/apimachinery/commit/2e35ad031) Use v1 api for schema-manager phase calulation (#1248)
- [41c7c89a](https://github.com/kubedb/apimachinery/commit/41c7c89a5) Correctly package up the solr constants (#1246)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.32.0-rc.1](https://github.com/kubedb/autoscaler/releases/tag/v0.32.0-rc.1)

- [9aa8ef3a](https://github.com/kubedb/autoscaler/commit/9aa8ef3a) Prepare for release v0.32.0-rc.1 (#212)
- [ed522899](https://github.com/kubedb/autoscaler/commit/ed522899) Update constants and petset with apiv1 (#211)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.47.0-rc.1](https://github.com/kubedb/cli/releases/tag/v0.47.0-rc.1)

- [a0aab82d](https://github.com/kubedb/cli/commit/a0aab82d) Prepare for release v0.47.0-rc.1 (#772)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.2.0-rc.1](https://github.com/kubedb/clickhouse/releases/tag/v0.2.0-rc.1)

- [69f6e117](https://github.com/kubedb/clickhouse/commit/69f6e117) Prepare for release v0.2.0-rc.1 (#7)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.2.0-rc.1](https://github.com/kubedb/crd-manager/releases/tag/v0.2.0-rc.1)

- [abdfe6d4](https://github.com/kubedb/crd-manager/commit/abdfe6d4) Prepare for release v0.2.0-rc.1 (#35)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.23.0-rc.1](https://github.com/kubedb/dashboard/releases/tag/v0.23.0-rc.1)

- [f47fe1fb](https://github.com/kubedb/dashboard/commit/f47fe1fb) Prepare for release v0.23.0-rc.1 (#120)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.5.0-rc.1](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.5.0-rc.1)

- [e6dae6e](https://github.com/kubedb/dashboard-restic-plugin/commit/e6dae6e) Prepare for release v0.5.0-rc.1 (#14)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.2.0-rc.1](https://github.com/kubedb/db-client-go/releases/tag/v0.2.0-rc.1)

- [57a5122f](https://github.com/kubedb/db-client-go/commit/57a5122f) Prepare for release v0.2.0-rc.1 (#121)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.2.0-rc.1](https://github.com/kubedb/druid/releases/tag/v0.2.0-rc.1)

- [a72889d](https://github.com/kubedb/druid/commit/a72889d) Prepare for release v0.2.0-rc.1 (#34)
- [c03ae84](https://github.com/kubedb/druid/commit/c03ae84) Update druid requeuing strategy for waiting (#33)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.47.0-rc.1](https://github.com/kubedb/elasticsearch/releases/tag/v0.47.0-rc.1)

- [1f8f6a49](https://github.com/kubedb/elasticsearch/commit/1f8f6a495) Prepare for release v0.47.0-rc.1 (#726)
- [66d09cc2](https://github.com/kubedb/elasticsearch/commit/66d09cc27) Fix error handling for validators (#725)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.10.0-rc.1](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.10.0-rc.1)

- [5a8fe42](https://github.com/kubedb/elasticsearch-restic-plugin/commit/5a8fe42) Prepare for release v0.10.0-rc.1 (#36)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.2.0-rc.1](https://github.com/kubedb/ferretdb/releases/tag/v0.2.0-rc.1)

- [fcc68498](https://github.com/kubedb/ferretdb/commit/fcc68498) Prepare for release v0.2.0-rc.1 (#34)
- [e8dfe581](https://github.com/kubedb/ferretdb/commit/e8dfe581) make client funcs accessible (#33)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2024.7.11-rc.1](https://github.com/kubedb/installer/releases/tag/v2024.7.11-rc.1)

- [16022316](https://github.com/kubedb/installer/commit/16022316) Prepare for release v2024.7.11-rc.1 (#1165)
- [954cdacc](https://github.com/kubedb/installer/commit/954cdacc) Update cve report 2024-07-11 (#1164)
- [bd137cca](https://github.com/kubedb/installer/commit/bd137cca) Don't import ui-chart CRDs (#1163)
- [9e14db28](https://github.com/kubedb/installer/commit/9e14db28) add ui section to postgresversions (#1156)
- [a72e9a26](https://github.com/kubedb/installer/commit/a72e9a26) Update crds for kubedb/apimachinery@79ca732b (#1162)
- [6798829a](https://github.com/kubedb/installer/commit/6798829a) add ops apiservice (#1149)
- [c99972ee](https://github.com/kubedb/installer/commit/c99972ee) Update redis init image for shards replica count changes (#1153)
- [951ebc5e](https://github.com/kubedb/installer/commit/951ebc5e) fix metrics configuration (#1154)
- [97dd8e83](https://github.com/kubedb/installer/commit/97dd8e83) Update cve report 2024-07-10 (#1158)
- [f744756b](https://github.com/kubedb/installer/commit/f744756b) Fix MariaDB Restic Image Version (#1159)
- [7fca858f](https://github.com/kubedb/installer/commit/7fca858f) Update Memcached Exporter Image & Add Metrics Configuration (#1077)
- [df73ad95](https://github.com/kubedb/installer/commit/df73ad95) Remove ui charts (#1157)
- [41d5a8ac](https://github.com/kubedb/installer/commit/41d5a8ac) Disable Pgadmin CSRF Check (#1147)
- [18699788](https://github.com/kubedb/installer/commit/18699788) Update crds for kubedb/apimachinery@9809d94e (#1155)
- [739d3e17](https://github.com/kubedb/installer/commit/739d3e17) Check for kubedb-webhook-server.enabled before waiting (#1152)
- [fc911e3a](https://github.com/kubedb/installer/commit/fc911e3a) Update cve report 2024-07-09 (#1151)
- [ff44a502](https://github.com/kubedb/installer/commit/ff44a502) Update crds for kubedb/apimachinery@f8063159 (#1150)
- [7f748314](https://github.com/kubedb/installer/commit/7f748314) Wait for the webhook-server svc to be ready (#1148)
- [004f3d71](https://github.com/kubedb/installer/commit/004f3d71) Update cve report 2024-07-08 (#1146)
- [7e91921a](https://github.com/kubedb/installer/commit/7e91921a) Fix ca.crt key for service monitors
- [6d639777](https://github.com/kubedb/installer/commit/6d639777) Update cve report 2024-07-04 (#1145)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.18.0-rc.1](https://github.com/kubedb/kafka/releases/tag/v0.18.0-rc.1)

- [b2cc90d4](https://github.com/kubedb/kafka/commit/b2cc90d4) Prepare for release v0.18.0-rc.1 (#98)
- [7a56e529](https://github.com/kubedb/kafka/commit/7a56e529) Install petset kafka daily (#97)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.10.0-rc.1](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.10.0-rc.1)

- [599fa89](https://github.com/kubedb/kubedb-manifest-plugin/commit/599fa89) Prepare for release v0.10.0-rc.1 (#58)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.31.0-rc.1](https://github.com/kubedb/mariadb/releases/tag/v0.31.0-rc.1)

- [c16d25c7](https://github.com/kubedb/mariadb/commit/c16d25c72) Prepare for release v0.31.0-rc.1 (#274)
- [823748e1](https://github.com/kubedb/mariadb/commit/823748e1a) Fix Env Validation (#273)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.7.0-rc.1](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.7.0-rc.1)

- [9e93d807](https://github.com/kubedb/mariadb-archiver/commit/9e93d807) Prepare for release v0.7.0-rc.1 (#22)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.27.0-rc.1](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.27.0-rc.1)

- [ca636ee2](https://github.com/kubedb/mariadb-coordinator/commit/ca636ee2) Prepare for release v0.27.0-rc.1 (#121)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.7.0-rc.1](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.7.0-rc.1)

- [e20539b](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/e20539b) Prepare for release v0.7.0-rc.1 (#25)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.5.0-rc.1](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.5.0-rc.1)

- [7af0211](https://github.com/kubedb/mariadb-restic-plugin/commit/7af0211) Prepare for release v0.5.0-rc.1 (#17)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.40.0-rc.1](https://github.com/kubedb/memcached/releases/tag/v0.40.0-rc.1)

- [028b7d98](https://github.com/kubedb/memcached/commit/028b7d98d) Prepare for release v0.40.0-rc.1 (#453)
- [ba86e1ca](https://github.com/kubedb/memcached/commit/ba86e1ca6) Update Validator (#452)
- [aa177b55](https://github.com/kubedb/memcached/commit/aa177b551) Fix Webhook Provisioner Restart Issue (#451)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.40.0-rc.1](https://github.com/kubedb/mongodb/releases/tag/v0.40.0-rc.1)

- [a5a96a7b](https://github.com/kubedb/mongodb/commit/a5a96a7ba) Prepare for release v0.40.0-rc.1 (#640)
- [87a1e446](https://github.com/kubedb/mongodb/commit/87a1e446f) fix error handling in validator (#639)
- [bda4f0c8](https://github.com/kubedb/mongodb/commit/bda4f0c85) Add petset to daily CI (#638)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.8.0-rc.1](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.8.0-rc.1)

- [96328df](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/96328df) Prepare for release v0.8.0-rc.1 (#30)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.10.0-rc.1](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.10.0-rc.1)

- [c22ff40](https://github.com/kubedb/mongodb-restic-plugin/commit/c22ff40) Prepare for release v0.10.0-rc.1 (#52)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.2.0-rc.1](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.2.0-rc.1)

- [080b930e](https://github.com/kubedb/mssql-coordinator/commit/080b930e) Prepare for release v0.2.0-rc.1 (#11)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.40.0-rc.1](https://github.com/kubedb/mysql/releases/tag/v0.40.0-rc.1)

- [9801f22d](https://github.com/kubedb/mysql/commit/9801f22db) Prepare for release v0.40.0-rc.1 (#631)
- [695750a5](https://github.com/kubedb/mysql/commit/695750a55) fix validator for MySQL (#630)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.8.0-rc.1](https://github.com/kubedb/mysql-archiver/releases/tag/v0.8.0-rc.1)

- [f7adcd27](https://github.com/kubedb/mysql-archiver/commit/f7adcd27) Prepare for release v0.8.0-rc.1 (#36)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.25.0-rc.1](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.25.0-rc.1)

- [7565022c](https://github.com/kubedb/mysql-coordinator/commit/7565022c) Prepare for release v0.25.0-rc.1 (#118)
- [e15adb2d](https://github.com/kubedb/mysql-coordinator/commit/e15adb2d) Update StatefulSet to PetSet (#117)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.8.0-rc.1](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.8.0-rc.1)

- [d289a41](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/d289a41) Prepare for release v0.8.0-rc.1 (#23)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.10.0-rc.1](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.10.0-rc.1)

- [ab39345](https://github.com/kubedb/mysql-restic-plugin/commit/ab39345) Prepare for release v0.10.0-rc.1 (#46)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.25.0-rc.1](https://github.com/kubedb/mysql-router-init/releases/tag/v0.25.0-rc.1)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.34.0-rc.1](https://github.com/kubedb/ops-manager/releases/tag/v0.34.0-rc.1)

- [81eee969](https://github.com/kubedb/ops-manager/commit/81eee9696) Prepare for release v0.34.0-rc.1 (#603)
- [d20609cb](https://github.com/kubedb/ops-manager/commit/d20609cbc) Add support for api V1 (#541)
- [a0612b60](https://github.com/kubedb/ops-manager/commit/a0612b607) Update Condition in MemcachedRetries (#591)
- [86af04a9](https://github.com/kubedb/ops-manager/commit/86af04a9f) Add Pgpool ops-request (Horizontal Scaling, Update Version, Reconfigure TLS) (#590)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.34.0-rc.1](https://github.com/kubedb/percona-xtradb/releases/tag/v0.34.0-rc.1)

- [eded3d05](https://github.com/kubedb/percona-xtradb/commit/eded3d05d) Prepare for release v0.34.0-rc.1 (#372)
- [1966d11a](https://github.com/kubedb/percona-xtradb/commit/1966d11a3) Fix Env Validation (#371)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.20.0-rc.1](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.20.0-rc.1)

- [fc57007d](https://github.com/kubedb/percona-xtradb-coordinator/commit/fc57007d) Prepare for release v0.20.0-rc.1 (#76)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.31.0-rc.1](https://github.com/kubedb/pg-coordinator/releases/tag/v0.31.0-rc.1)

- [1c067e4c](https://github.com/kubedb/pg-coordinator/commit/1c067e4c) Prepare for release v0.31.0-rc.1 (#168)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.34.0-rc.1](https://github.com/kubedb/pgbouncer/releases/tag/v0.34.0-rc.1)

- [0f12bc22](https://github.com/kubedb/pgbouncer/commit/0f12bc22) Prepare for release v0.34.0-rc.1 (#338)
- [3f9a8665](https://github.com/kubedb/pgbouncer/commit/3f9a8665) Signed-off-by: Hiranmoy Das Chowdhury <hiranmoy@appscode.com> (#337)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.2.0-rc.1](https://github.com/kubedb/pgpool/releases/tag/v0.2.0-rc.1)

- [60867940](https://github.com/kubedb/pgpool/commit/60867940) Prepare for release v0.2.0-rc.1 (#38)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.47.0-rc.1](https://github.com/kubedb/postgres/releases/tag/v0.47.0-rc.1)

- [2bf47c9e](https://github.com/kubedb/postgres/commit/2bf47c9e4) Prepare for release v0.47.0-rc.1 (#739)
- [bcfe0a48](https://github.com/kubedb/postgres/commit/bcfe0a488) Fix validator for postgres (#738)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.8.0-rc.1](https://github.com/kubedb/postgres-archiver/releases/tag/v0.8.0-rc.1)

- [5a8c6ec9](https://github.com/kubedb/postgres-archiver/commit/5a8c6ec9) Prepare for release v0.8.0-rc.1 (#34)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.8.0-rc.1](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.8.0-rc.1)

- [624c851](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/624c851) Prepare for release v0.8.0-rc.1 (#32)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.10.0-rc.1](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.10.0-rc.1)

- [f50a13e](https://github.com/kubedb/postgres-restic-plugin/commit/f50a13e) Prepare for release v0.10.0-rc.1 (#39)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.9.0-rc.1](https://github.com/kubedb/provider-aws/releases/tag/v0.9.0-rc.1)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.9.0-rc.1](https://github.com/kubedb/provider-azure/releases/tag/v0.9.0-rc.1)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.9.0-rc.1](https://github.com/kubedb/provider-gcp/releases/tag/v0.9.0-rc.1)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.47.0-rc.1](https://github.com/kubedb/provisioner/releases/tag/v0.47.0-rc.1)

- [6767c852](https://github.com/kubedb/provisioner/commit/6767c8527) Prepare for release v0.47.0-rc.1 (#102)
- [8429d3b2](https://github.com/kubedb/provisioner/commit/8429d3b2a) Update deps (#101)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.34.0-rc.1](https://github.com/kubedb/proxysql/releases/tag/v0.34.0-rc.1)

- [7da3c423](https://github.com/kubedb/proxysql/commit/7da3c4235) Prepare for release v0.34.0-rc.1 (#349)
- [0ea35fb6](https://github.com/kubedb/proxysql/commit/0ea35fb68) fix validator for ProxySQL (#348)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.2.0-rc.1](https://github.com/kubedb/rabbitmq/releases/tag/v0.2.0-rc.1)

- [8b6b97bb](https://github.com/kubedb/rabbitmq/commit/8b6b97bb) Prepare for release v0.2.0-rc.1 (#36)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.40.0-rc.1](https://github.com/kubedb/redis/releases/tag/v0.40.0-rc.1)

- [9cb53e47](https://github.com/kubedb/redis/commit/9cb53e470) Prepare for release v0.40.0-rc.1 (#546)
- [8af74f1a](https://github.com/kubedb/redis/commit/8af74f1a0) Update master -> shards and replica count for cluster (#545)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.26.0-rc.1](https://github.com/kubedb/redis-coordinator/releases/tag/v0.26.0-rc.1)

- [d15ce249](https://github.com/kubedb/redis-coordinator/commit/d15ce249) Prepare for release v0.26.0-rc.1 (#107)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.10.0-rc.1](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.10.0-rc.1)

- [95dd894](https://github.com/kubedb/redis-restic-plugin/commit/95dd894) Prepare for release v0.10.0-rc.1 (#38)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.34.0-rc.1](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.34.0-rc.1)

- [c0197572](https://github.com/kubedb/replication-mode-detector/commit/c0197572) Prepare for release v0.34.0-rc.1 (#273)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.23.0-rc.1](https://github.com/kubedb/schema-manager/releases/tag/v0.23.0-rc.1)

- [3f84503b](https://github.com/kubedb/schema-manager/commit/3f84503b) Prepare for release v0.23.0-rc.1 (#116)
- [db6996aa](https://github.com/kubedb/schema-manager/commit/db6996aa) Directly use phase from DB status section (#115)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.2.0-rc.1](https://github.com/kubedb/singlestore/releases/tag/v0.2.0-rc.1)

- [fd637835](https://github.com/kubedb/singlestore/commit/fd637835) Prepare for release v0.2.0-rc.1 (#38)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.2.0-rc.1](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.2.0-rc.1)

- [e0bc384](https://github.com/kubedb/singlestore-coordinator/commit/e0bc384) Prepare for release v0.2.0-rc.1 (#22)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.5.0-rc.1](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.5.0-rc.1)

- [9bf8b9c](https://github.com/kubedb/singlestore-restic-plugin/commit/9bf8b9c) Prepare for release v0.5.0-rc.1 (#17)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.2.0-rc.1](https://github.com/kubedb/solr/releases/tag/v0.2.0-rc.1)

- [4d896266](https://github.com/kubedb/solr/commit/4d896266) Prepare for release v0.2.0-rc.1 (#37)
- [d8f02861](https://github.com/kubedb/solr/commit/d8f02861) fix constants (#36)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.32.0-rc.1](https://github.com/kubedb/tests/releases/tag/v0.32.0-rc.1)

- [8b29ad4f](https://github.com/kubedb/tests/commit/8b29ad4f) Prepare for release v0.32.0-rc.1 (#332)
- [522ce4dd](https://github.com/kubedb/tests/commit/522ce4dd) Add api V1 support for e2e test cases (#330)
- [074319cb](https://github.com/kubedb/tests/commit/074319cb) Kubestash test (#328)
- [3d86cc15](https://github.com/kubedb/tests/commit/3d86cc15) Add MS SQL Server Provisioning Tests  (#321)
- [ac5c8e4a](https://github.com/kubedb/tests/commit/ac5c8e4a) Add FerretDB test (#323)
- [3b09f127](https://github.com/kubedb/tests/commit/3b09f127) Reprovision test (#311)
- [cbb366d5](https://github.com/kubedb/tests/commit/cbb366d5) Update SingleStore Tests Regarding API Changes (#327)
- [7568498c](https://github.com/kubedb/tests/commit/7568498c) Fix Pgpool sync users test (#326)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.23.0-rc.1](https://github.com/kubedb/ui-server/releases/tag/v0.23.0-rc.1)

- [4cad3a5d](https://github.com/kubedb/ui-server/commit/4cad3a5d) Prepare for release v0.23.0-rc.1 (#123)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.23.0-rc.1](https://github.com/kubedb/webhook-server/releases/tag/v0.23.0-rc.1)

- [0ddff1f5](https://github.com/kubedb/webhook-server/commit/0ddff1f5) Prepare for release v0.23.0-rc.1 (#115)
- [acd7e03f](https://github.com/kubedb/webhook-server/commit/acd7e03f) Fix ops webhook (#114)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.2.0-rc.1](https://github.com/kubedb/zookeeper/releases/tag/v0.2.0-rc.1)

- [68219ffe](https://github.com/kubedb/zookeeper/commit/68219ffe) Prepare for release v0.2.0-rc.1 (#31)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.3.0-rc.1](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.3.0-rc.1)

- [6333dd9](https://github.com/kubedb/zookeeper-restic-plugin/commit/6333dd9) Prepare for release v0.3.0-rc.1 (#10)




