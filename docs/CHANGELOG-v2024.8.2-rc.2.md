---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2024.8.2-rc.2
    name: Changelog-v2024.8.2-rc.2
    parent: welcome
    weight: 20240802
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2024.8.2-rc.2/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2024.8.2-rc.2/
---

# KubeDB v2024.8.2-rc.2 (2024-08-03)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.47.0-rc.2](https://github.com/kubedb/apimachinery/releases/tag/v0.47.0-rc.2)

- [0f09eede](https://github.com/kubedb/apimachinery/commit/0f09eedec) Add DefaultUserCredSecretName utility for mssql
- [527c5936](https://github.com/kubedb/apimachinery/commit/527c59365) Add MSSQLServer Archiver Backup and Restore API (#1265)
- [6779210b](https://github.com/kubedb/apimachinery/commit/6779210b4) Fix build error (#1274)
- [1d9ee37f](https://github.com/kubedb/apimachinery/commit/1d9ee37f6) Add Solr ops for vertical scaling and volume expansion (#1261)
- [45c637cc](https://github.com/kubedb/apimachinery/commit/45c637cc8) Add helpers to get the archiver with maximum priority (#1266)
- [9cb2a307](https://github.com/kubedb/apimachinery/commit/9cb2a3076) Update deps
- [739f7f6f](https://github.com/kubedb/apimachinery/commit/739f7f6f1) Add MSSQL Server Monitoring APIs (#1271)
- [3f96b907](https://github.com/kubedb/apimachinery/commit/3f96b9077) Add ExtractDatabaseInfo Func for ClickHouse (#1270)
- [6d44bfc1](https://github.com/kubedb/apimachinery/commit/6d44bfc1c) Upsert config-merger initContainer via ES defaults (#1259)
- [67d23948](https://github.com/kubedb/apimachinery/commit/67d239489) Add FerretDBOpsManager (#1267)
- [08af58b8](https://github.com/kubedb/apimachinery/commit/08af58b87) pgbouncerAutoScalerSpec Update and Scheme Register (#1268)
- [c57d4c2b](https://github.com/kubedb/apimachinery/commit/c57d4c2bc) Keep the druid dependency references if given (#1272)
- [88b60875](https://github.com/kubedb/apimachinery/commit/88b60875b) Add Kafka RestProxy APIs (#1262)
- [4af05fd8](https://github.com/kubedb/apimachinery/commit/4af05fd8b) Add types for all autoscaler CRDs (#1264)
- [bf2f0aeb](https://github.com/kubedb/apimachinery/commit/bf2f0aeb4) Update Memcached Scaling APIs (#1260)
- [7b594eb1](https://github.com/kubedb/apimachinery/commit/7b594eb14) Use DeletionPolicy in etcd (#1258)
- [84475028](https://github.com/kubedb/apimachinery/commit/844750289) Skip RedisClusterSpec conversion for sentinel and standalone mode (#1257)
- [c028472b](https://github.com/kubedb/apimachinery/commit/c028472b9) Update deps



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.32.0-rc.2](https://github.com/kubedb/autoscaler/releases/tag/v0.32.0-rc.2)

- [a03964ec](https://github.com/kubedb/autoscaler/commit/a03964ec) Prepare for release v0.32.0-rc.2 (#215)
- [fc76bfff](https://github.com/kubedb/autoscaler/commit/fc76bfff) Set db as the OwnerRef in autoscaler CR (#213)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.47.0-rc.2](https://github.com/kubedb/cli/releases/tag/v0.47.0-rc.2)

- [26dca461](https://github.com/kubedb/cli/commit/26dca461) Prepare for release v0.47.0-rc.2 (#773)
- [a82c4b22](https://github.com/kubedb/cli/commit/a82c4b22) Make changes to run cli from the appscode/grafana-dashboards CI (#766)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.2.0-rc.2](https://github.com/kubedb/clickhouse/releases/tag/v0.2.0-rc.2)

- [352d87ca](https://github.com/kubedb/clickhouse/commit/352d87ca) Prepare for release v0.2.0-rc.2 (#9)
- [06fb2f51](https://github.com/kubedb/clickhouse/commit/06fb2f51) Fix Finalizer Removal and Remove PetSet Ready Condition Check (#8)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.2.0-rc.2](https://github.com/kubedb/crd-manager/releases/tag/v0.2.0-rc.2)

- [6a17ac75](https://github.com/kubedb/crd-manager/commit/6a17ac75) Prepare for release v0.2.0-rc.2 (#40)
- [2772153d](https://github.com/kubedb/crd-manager/commit/2772153d) Scale Down Provisioner if Older (#38)
- [f5470e5b](https://github.com/kubedb/crd-manager/commit/f5470e5b) Add Kafka RestProxy CRD (#37)
- [da1fedc4](https://github.com/kubedb/crd-manager/commit/da1fedc4) Add ferretdb ops-manager CRD (#39)
- [d7f0c41b](https://github.com/kubedb/crd-manager/commit/d7f0c41b) Install autoscaler CRDs (#36)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.23.0-rc.2](https://github.com/kubedb/dashboard/releases/tag/v0.23.0-rc.2)

- [b23df248](https://github.com/kubedb/dashboard/commit/b23df248) Prepare for release v0.23.0-rc.2 (#121)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.5.0-rc.2](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.5.0-rc.2)

- [6188e57](https://github.com/kubedb/dashboard-restic-plugin/commit/6188e57) Prepare for release v0.5.0-rc.2 (#15)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.2.0-rc.2](https://github.com/kubedb/db-client-go/releases/tag/v0.2.0-rc.2)

- [99d096fb](https://github.com/kubedb/db-client-go/commit/99d096fb) Prepare for release v0.2.0-rc.2 (#126)
- [49bebb7e](https://github.com/kubedb/db-client-go/commit/49bebb7e) Add Kafka RestProxy (#123)
- [495ccff1](https://github.com/kubedb/db-client-go/commit/495ccff1) Add solr client. (#106)
- [01231603](https://github.com/kubedb/db-client-go/commit/01231603) Add method to set database for redis client (#125)
- [877df856](https://github.com/kubedb/db-client-go/commit/877df856) Add ZooKeeper Client (#124)
- [48d0e46f](https://github.com/kubedb/db-client-go/commit/48d0e46f) Add druid client (#122)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.2.0-rc.2](https://github.com/kubedb/druid/releases/tag/v0.2.0-rc.2)

- [af49e11](https://github.com/kubedb/druid/commit/af49e11) Prepare for release v0.2.0-rc.2 (#37)
- [c0f7a40](https://github.com/kubedb/druid/commit/c0f7a40) Fix druid healthcheck (#36)
- [d0bf458](https://github.com/kubedb/druid/commit/d0bf458) Update makefile and add druid db-client-go (#35)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.47.0-rc.2](https://github.com/kubedb/elasticsearch/releases/tag/v0.47.0-rc.2)

- [fd8abfca](https://github.com/kubedb/elasticsearch/commit/fd8abfca3) Prepare for release v0.47.0-rc.2 (#728)
- [be009364](https://github.com/kubedb/elasticsearch/commit/be009364b) Fix PodTemplate assignment to config-merger initContainer (#727)
- [28668a0c](https://github.com/kubedb/elasticsearch/commit/28668a0c3) Revert "Fix podTemplate assignment for init container"
- [9b7e0aa0](https://github.com/kubedb/elasticsearch/commit/9b7e0aa0c) Fix podTemplate assignment for init container



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.10.0-rc.2](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.10.0-rc.2)

- [2fabca1](https://github.com/kubedb/elasticsearch-restic-plugin/commit/2fabca1) Prepare for release v0.10.0-rc.2 (#37)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.2.0-rc.2](https://github.com/kubedb/ferretdb/releases/tag/v0.2.0-rc.2)

- [0a38354a](https://github.com/kubedb/ferretdb/commit/0a38354a) Prepare for release v0.2.0-rc.2 (#36)
- [0d005f64](https://github.com/kubedb/ferretdb/commit/0d005f64) Make some changes for ops manager (#35)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2024.8.2-rc.2](https://github.com/kubedb/installer/releases/tag/v2024.8.2-rc.2)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.18.0-rc.2](https://github.com/kubedb/kafka/releases/tag/v0.18.0-rc.2)

- [98b8a404](https://github.com/kubedb/kafka/commit/98b8a404) Prepare for release v0.18.0-rc.2 (#100)
- [ced0da95](https://github.com/kubedb/kafka/commit/ced0da95) Add Kafka RestProxy (#99)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.10.0-rc.2](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.10.0-rc.2)

- [8b1207d](https://github.com/kubedb/kubedb-manifest-plugin/commit/8b1207d) Prepare for release v0.10.0-rc.2 (#60)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.31.0-rc.2](https://github.com/kubedb/mariadb/releases/tag/v0.31.0-rc.2)

- [86038a20](https://github.com/kubedb/mariadb/commit/86038a201) Prepare for release v0.31.0-rc.2 (#276)
- [d66565b2](https://github.com/kubedb/mariadb/commit/d66565b29) Fix Archiver BackupConfig Not Ready Issue (#275)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.7.0-rc.2](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.7.0-rc.2)

- [138bcf7c](https://github.com/kubedb/mariadb-archiver/commit/138bcf7c) Prepare for release v0.7.0-rc.2 (#23)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.27.0-rc.2](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.27.0-rc.2)

- [92a607fa](https://github.com/kubedb/mariadb-coordinator/commit/92a607fa) Prepare for release v0.27.0-rc.2 (#122)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.7.0-rc.2](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.7.0-rc.2)

- [6df0b0f](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/6df0b0f) Prepare for release v0.7.0-rc.2 (#26)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.5.0-rc.2](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.5.0-rc.2)

- [73a3ff8](https://github.com/kubedb/mariadb-restic-plugin/commit/73a3ff8) Prepare for release v0.5.0-rc.2 (#18)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.40.0-rc.2](https://github.com/kubedb/memcached/releases/tag/v0.40.0-rc.2)

- [f2ab956b](https://github.com/kubedb/memcached/commit/f2ab956b4) Prepare for release v0.40.0-rc.2 (#457)
- [8a4aaac1](https://github.com/kubedb/memcached/commit/8a4aaac1d) Add Reconciler (#456)
- [7ee57761](https://github.com/kubedb/memcached/commit/7ee577616) Add Rule and Petset Watcher (#455)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.40.0-rc.2](https://github.com/kubedb/mongodb/releases/tag/v0.40.0-rc.2)

- [2f717399](https://github.com/kubedb/mongodb/commit/2f717399e) Prepare for release v0.40.0-rc.2 (#643)
- [a387e83d](https://github.com/kubedb/mongodb/commit/a387e83d1) Modify the archiver selection process (#636)
- [6d8e3468](https://github.com/kubedb/mongodb/commit/6d8e3468d) Copy secrets to DB namespace; Refactor (#642)
- [88843d22](https://github.com/kubedb/mongodb/commit/88843d22f) Copy Toleration & placementPolicy field to petset (#641)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.8.0-rc.2](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.8.0-rc.2)

- [5fd0cf8](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/5fd0cf8) Prepare for release v0.8.0-rc.2 (#31)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.10.0-rc.2](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.10.0-rc.2)

- [9edee3a](https://github.com/kubedb/mongodb-restic-plugin/commit/9edee3a) Prepare for release v0.10.0-rc.2 (#53)
- [2039750](https://github.com/kubedb/mongodb-restic-plugin/commit/2039750) fix tls enable mongodb ping issue (#51)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.2.0-rc.2](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.2.0-rc.2)

- [b6e43327](https://github.com/kubedb/mssql-coordinator/commit/b6e43327) Prepare for release v0.2.0-rc.2 (#12)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.2.0-rc.2](https://github.com/kubedb/mssqlserver/releases/tag/v0.2.0-rc.2)

- [76b74614](https://github.com/kubedb/mssqlserver/commit/76b74614) Update deps
- [47446bb1](https://github.com/kubedb/mssqlserver/commit/47446bb1) Update deps



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.40.0-rc.2](https://github.com/kubedb/mysql/releases/tag/v0.40.0-rc.2)

- [b2cc7f23](https://github.com/kubedb/mysql/commit/b2cc7f236) Prepare for release v0.40.0-rc.2 (#632)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.8.0-rc.2](https://github.com/kubedb/mysql-archiver/releases/tag/v0.8.0-rc.2)

- [a54178b9](https://github.com/kubedb/mysql-archiver/commit/a54178b9) Prepare for release v0.8.0-rc.2 (#37)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.25.0-rc.2](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.25.0-rc.2)

- [cbae32d4](https://github.com/kubedb/mysql-coordinator/commit/cbae32d4) Prepare for release v0.25.0-rc.2 (#119)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.8.0-rc.2](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.8.0-rc.2)

- [1b66901](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/1b66901) Prepare for release v0.8.0-rc.2 (#24)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.10.0-rc.2](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.10.0-rc.2)

- [4c64334](https://github.com/kubedb/mysql-restic-plugin/commit/4c64334) Prepare for release v0.10.0-rc.2 (#47)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.25.0-rc.2](https://github.com/kubedb/mysql-router-init/releases/tag/v0.25.0-rc.2)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.34.0-rc.2](https://github.com/kubedb/ops-manager/releases/tag/v0.34.0-rc.2)

- [2cfccf3e](https://github.com/kubedb/ops-manager/commit/2cfccf3ec) Prepare for release v0.34.0-rc.2 (#614)
- [597c24a1](https://github.com/kubedb/ops-manager/commit/597c24a1f) Add solr ops request (#611)
- [9626eaff](https://github.com/kubedb/ops-manager/commit/9626eaffd) Fix Ops Requests for Redis and Sentinel  (#604)
- [59f17bfe](https://github.com/kubedb/ops-manager/commit/59f17bfe6) Fix ES for PetSet changes (#613)
- [74bb1b19](https://github.com/kubedb/ops-manager/commit/74bb1b19a) Add FerretDB OpsManager (#612)
- [d366a497](https://github.com/kubedb/ops-manager/commit/d366a4978) Fix and reorg MongoDB AddTLS and RemoveTLS (#607)
- [0665166a](https://github.com/kubedb/ops-manager/commit/0665166a1) Memcached Ops Request (Horizontal Scaling & Version Update) (#609)
- [593dc522](https://github.com/kubedb/ops-manager/commit/593dc5227) horizontal and vertical scaling for pgbouncer (#594)
- [fb62665c](https://github.com/kubedb/ops-manager/commit/fb62665ce) Fix memcached Ops Request (#606)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.34.0-rc.2](https://github.com/kubedb/percona-xtradb/releases/tag/v0.34.0-rc.2)

- [4db202b7](https://github.com/kubedb/percona-xtradb/commit/4db202b76) Prepare for release v0.34.0-rc.2 (#374)
- [4fa06ced](https://github.com/kubedb/percona-xtradb/commit/4fa06cedc) Fix Init Container Volume Mount Issue (#373)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.20.0-rc.2](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.20.0-rc.2)

- [b43e1a42](https://github.com/kubedb/percona-xtradb-coordinator/commit/b43e1a42) Prepare for release v0.20.0-rc.2 (#77)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.31.0-rc.2](https://github.com/kubedb/pg-coordinator/releases/tag/v0.31.0-rc.2)

- [8c6d9de3](https://github.com/kubedb/pg-coordinator/commit/8c6d9de3) Prepare for release v0.31.0-rc.2 (#169)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.34.0-rc.2](https://github.com/kubedb/pgbouncer/releases/tag/v0.34.0-rc.2)

- [ff8c5491](https://github.com/kubedb/pgbouncer/commit/ff8c5491) Prepare for release v0.34.0-rc.2 (#339)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.2.0-rc.2](https://github.com/kubedb/pgpool/releases/tag/v0.2.0-rc.2)

- [7df31597](https://github.com/kubedb/pgpool/commit/7df31597) Prepare for release v0.2.0-rc.2 (#39)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.47.0-rc.2](https://github.com/kubedb/postgres/releases/tag/v0.47.0-rc.2)

- [25b10bb6](https://github.com/kubedb/postgres/commit/25b10bb6e) Prepare for release v0.47.0-rc.2 (#743)
- [2a6b188e](https://github.com/kubedb/postgres/commit/2a6b188e7) trgger backup once after appbinding is created from provisioner (#741)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.8.0-rc.2](https://github.com/kubedb/postgres-archiver/releases/tag/v0.8.0-rc.2)

- [0949ab0c](https://github.com/kubedb/postgres-archiver/commit/0949ab0c) Prepare for release v0.8.0-rc.2 (#35)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.8.0-rc.2](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.8.0-rc.2)

- [744dd4a](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/744dd4a) Prepare for release v0.8.0-rc.2 (#33)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.10.0-rc.2](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.10.0-rc.2)

- [e27dc68](https://github.com/kubedb/postgres-restic-plugin/commit/e27dc68) Prepare for release v0.10.0-rc.2 (#41)
- [67a5b04](https://github.com/kubedb/postgres-restic-plugin/commit/67a5b04) Add postgres multiple db version support for kubestash (#40)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.9.0-rc.2](https://github.com/kubedb/provider-aws/releases/tag/v0.9.0-rc.2)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.9.0-rc.2](https://github.com/kubedb/provider-azure/releases/tag/v0.9.0-rc.2)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.9.0-rc.2](https://github.com/kubedb/provider-gcp/releases/tag/v0.9.0-rc.2)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.47.0-rc.2](https://github.com/kubedb/provisioner/releases/tag/v0.47.0-rc.2)

- [f792d7a5](https://github.com/kubedb/provisioner/commit/f792d7a5a) Prepare for release v0.47.0-rc.2 (#107)
- [fb027117](https://github.com/kubedb/provisioner/commit/fb0271174) Update deps
- [53179301](https://github.com/kubedb/provisioner/commit/53179301d) Update deps (#106)
- [8b7a2a6d](https://github.com/kubedb/provisioner/commit/8b7a2a6d7) Add Kafka RestProxy (#103)
- [5ff0a5c8](https://github.com/kubedb/provisioner/commit/5ff0a5c83) Add Separate controller for Redis sentinel (#105)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.34.0-rc.2](https://github.com/kubedb/proxysql/releases/tag/v0.34.0-rc.2)

- [c7df11cb](https://github.com/kubedb/proxysql/commit/c7df11cb2) Prepare for release v0.34.0-rc.2 (#351)
- [c56d021e](https://github.com/kubedb/proxysql/commit/c56d021e7) Elevate privilege for monitor user (#350)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.2.0-rc.2](https://github.com/kubedb/rabbitmq/releases/tag/v0.2.0-rc.2)

- [6e8ac555](https://github.com/kubedb/rabbitmq/commit/6e8ac555) Prepare for release v0.2.0-rc.2 (#37)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.40.0-rc.2](https://github.com/kubedb/redis/releases/tag/v0.40.0-rc.2)

- [7e9c8648](https://github.com/kubedb/redis/commit/7e9c8648c) Prepare for release v0.40.0-rc.2 (#549)
- [38b4b380](https://github.com/kubedb/redis/commit/38b4b3807) Update deps (#547)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.26.0-rc.2](https://github.com/kubedb/redis-coordinator/releases/tag/v0.26.0-rc.2)

- [3a92ab81](https://github.com/kubedb/redis-coordinator/commit/3a92ab81) Prepare for release v0.26.0-rc.2 (#108)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.10.0-rc.2](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.10.0-rc.2)

- [15072a8](https://github.com/kubedb/redis-restic-plugin/commit/15072a8) Prepare for release v0.10.0-rc.2 (#39)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.34.0-rc.2](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.34.0-rc.2)

- [77ca2092](https://github.com/kubedb/replication-mode-detector/commit/77ca2092) Prepare for release v0.34.0-rc.2 (#274)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.23.0-rc.2](https://github.com/kubedb/schema-manager/releases/tag/v0.23.0-rc.2)

- [10ca5613](https://github.com/kubedb/schema-manager/commit/10ca5613) Prepare for release v0.23.0-rc.2 (#117)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.2.0-rc.2](https://github.com/kubedb/singlestore/releases/tag/v0.2.0-rc.2)

- [09de410f](https://github.com/kubedb/singlestore/commit/09de410f) Prepare for release v0.2.0-rc.2 (#39)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.2.0-rc.2](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.2.0-rc.2)

- [33a157b](https://github.com/kubedb/singlestore-coordinator/commit/33a157b) Prepare for release v0.2.0-rc.2 (#23)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.5.0-rc.2](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.5.0-rc.2)

- [1a8b875](https://github.com/kubedb/singlestore-restic-plugin/commit/1a8b875) Prepare for release v0.5.0-rc.2 (#18)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.2.0-rc.2](https://github.com/kubedb/solr/releases/tag/v0.2.0-rc.2)

- [a7534064](https://github.com/kubedb/solr/commit/a7534064) Prepare for release v0.2.0-rc.2 (#39)
- [1dd26676](https://github.com/kubedb/solr/commit/1dd26676) Changes related to ops manager (#38)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.32.0-rc.2](https://github.com/kubedb/tests/releases/tag/v0.32.0-rc.2)

- [b279ee9f](https://github.com/kubedb/tests/commit/b279ee9f) Prepare for release v0.32.0-rc.2 (#343)
- [92599f33](https://github.com/kubedb/tests/commit/92599f33) Add Druid Tests (#306)
- [d4762475](https://github.com/kubedb/tests/commit/d4762475) Fix ES env variable tests for V1 api changes (#336)
- [1d5a9926](https://github.com/kubedb/tests/commit/1d5a9926) Add Resource for PerconaXtraDB, MariaDB when creating object (#334)
- [43aa9e97](https://github.com/kubedb/tests/commit/43aa9e97) Fix ES test for V1 changes (#335)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.23.0-rc.2](https://github.com/kubedb/ui-server/releases/tag/v0.23.0-rc.2)

- [893dc3ac](https://github.com/kubedb/ui-server/commit/893dc3ac) Prepare for release v0.23.0-rc.2 (#125)
- [8fa3f1cb](https://github.com/kubedb/ui-server/commit/8fa3f1cb) resource matrics dep updation (#124)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.23.0-rc.2](https://github.com/kubedb/webhook-server/releases/tag/v0.23.0-rc.2)




## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.2.0-rc.2](https://github.com/kubedb/zookeeper/releases/tag/v0.2.0-rc.2)

- [1051b1db](https://github.com/kubedb/zookeeper/commit/1051b1db) Prepare for release v0.2.0-rc.2 (#34)
- [2f695af7](https://github.com/kubedb/zookeeper/commit/2f695af7) Add ZooKeeper Client (#33)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.3.0-rc.2](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.3.0-rc.2)

- [0f2644b](https://github.com/kubedb/zookeeper-restic-plugin/commit/0f2644b) Prepare for release v0.3.0-rc.2 (#11)




