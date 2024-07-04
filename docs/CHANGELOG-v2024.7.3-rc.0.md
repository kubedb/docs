---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2024.7.3-rc.0
    name: Changelog-v2024.7.3-rc.0
    parent: welcome
    weight: 20240703
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2024.7.3-rc.0/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2024.7.3-rc.0/
---

# KubeDB v2024.7.3-rc.0 (2024-07-04)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.47.0-rc.0](https://github.com/kubedb/apimachinery/releases/tag/v0.47.0-rc.0)

- [1302836d](https://github.com/kubedb/apimachinery/commit/1302836dc) Update deps
- [01fcb668](https://github.com/kubedb/apimachinery/commit/01fcb6683) Introduce v1 api (#1236)
- [42019af5](https://github.com/kubedb/apimachinery/commit/42019af5f) Update for release KubeStash@v2024.7.1 (#1245)
- [519c2389](https://github.com/kubedb/apimachinery/commit/519c2389b) Fix druid defaulter (#1243)
- [735c4683](https://github.com/kubedb/apimachinery/commit/735c4683a) Update Druid API for internally managed metadatastore and zookeeper (#1238)
- [b4f0c7ae](https://github.com/kubedb/apimachinery/commit/b4f0c7ae5) Add AppBinding PostgresRef in FerretDB API (#1239)
- [b88f519b](https://github.com/kubedb/apimachinery/commit/b88f519ba) Add Pgpool ops-request api(Horizontal Scaling) (#1241)
- [7a9cbb53](https://github.com/kubedb/apimachinery/commit/7a9cbb53c) auth_mode changes (#1235)
- [d9228be3](https://github.com/kubedb/apimachinery/commit/d9228be31) Make package path match package name
- [dd0bd4e6](https://github.com/kubedb/apimachinery/commit/dd0bd4e6f) Move feature gates from crd-manager (#1240)
- [c1a2e274](https://github.com/kubedb/apimachinery/commit/c1a2e2745) Reset RabbitMQ default healthcker periodSecond (#1237)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.47.0-rc.0](https://github.com/kubedb/cli/releases/tag/v0.47.0-rc.0)

- [b4118de7](https://github.com/kubedb/cli/commit/b4118de7) Prepare for release v0.47.0-rc.0 (#770)
- [2e7131a6](https://github.com/kubedb/cli/commit/2e7131a6) update api to v1 (#771)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.2.0-rc.0](https://github.com/kubedb/clickhouse/releases/tag/v0.2.0-rc.0)

- [205ad288](https://github.com/kubedb/clickhouse/commit/205ad288) Prepare for release v0.2.0-rc.0 (#6)
- [a763c285](https://github.com/kubedb/clickhouse/commit/a763c285) Update constants to use kubedb package (#5)
- [d16c8c0b](https://github.com/kubedb/clickhouse/commit/d16c8c0b) Fix Auth Secret Issue (#3)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.2.0-rc.0](https://github.com/kubedb/crd-manager/releases/tag/v0.2.0-rc.0)

- [8392c6cd](https://github.com/kubedb/crd-manager/commit/8392c6cd) Prepare for release v0.2.0-rc.0 (#34)
- [a4f9e562](https://github.com/kubedb/crd-manager/commit/a4f9e562) Preserve crd conversion config on update (#31)
- [5c05c9ba](https://github.com/kubedb/crd-manager/commit/5c05c9ba) Move features to apimachinery



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.23.0-rc.0](https://github.com/kubedb/dashboard/releases/tag/v0.23.0-rc.0)

- [cc962aff](https://github.com/kubedb/dashboard/commit/cc962aff) Prepare for release v0.23.0-rc.0 (#119)
- [4981533c](https://github.com/kubedb/dashboard/commit/4981533c) Update constants with apiv1 (#118)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.5.0-rc.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.5.0-rc.0)

- [d0c1465](https://github.com/kubedb/dashboard-restic-plugin/commit/d0c1465) Prepare for release v0.5.0-rc.0 (#13)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.2.0-rc.0](https://github.com/kubedb/db-client-go/releases/tag/v0.2.0-rc.0)

- [01905848](https://github.com/kubedb/db-client-go/commit/01905848) Prepare for release v0.2.0-rc.0 (#120)
- [3b94bb3e](https://github.com/kubedb/db-client-go/commit/3b94bb3e) Add v1 api to db clients (#119)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.2.0-rc.0](https://github.com/kubedb/druid/releases/tag/v0.2.0-rc.0)

- [8047a2d](https://github.com/kubedb/druid/commit/8047a2d) Prepare for release v0.2.0-rc.0 (#32)
- [3a3deb0](https://github.com/kubedb/druid/commit/3a3deb0) Update druid for creating metadata storage and zk (#30)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.47.0-rc.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.47.0-rc.0)

- [f3bd7f56](https://github.com/kubedb/elasticsearch/commit/f3bd7f56f) Prepare for release v0.47.0-rc.0 (#724)
- [05142253](https://github.com/kubedb/elasticsearch/commit/051422532) Use v1 api (#723)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.10.0-rc.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.10.0-rc.0)

- [10c7f69](https://github.com/kubedb/elasticsearch-restic-plugin/commit/10c7f69) Prepare for release v0.10.0-rc.0 (#35)
- [0f742fd](https://github.com/kubedb/elasticsearch-restic-plugin/commit/0f742fd) Use v1 api (#34)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.2.0-rc.0](https://github.com/kubedb/ferretdb/releases/tag/v0.2.0-rc.0)

- [c9abee71](https://github.com/kubedb/ferretdb/commit/c9abee71) Prepare for release v0.2.0-rc.0 (#32)
- [0c36fb43](https://github.com/kubedb/ferretdb/commit/0c36fb43) Fix apimachinery constants (#31)
- [0afda2a8](https://github.com/kubedb/ferretdb/commit/0afda2a8) Add e2e ci (#25)
- [652e0d81](https://github.com/kubedb/ferretdb/commit/652e0d81) Fix Backend TLS connection (#30)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2024.7.3-rc.0](https://github.com/kubedb/installer/releases/tag/v2024.7.3-rc.0)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.18.0-rc.0](https://github.com/kubedb/kafka/releases/tag/v0.18.0-rc.0)

- [c3f92486](https://github.com/kubedb/kafka/commit/c3f92486) Prepare for release v0.18.0-rc.0 (#96)
- [19b65b86](https://github.com/kubedb/kafka/commit/19b65b86) Update Statefulset with PetSet and apiversion (#95)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.10.0-rc.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.10.0-rc.0)

- [7a27d11](https://github.com/kubedb/kubedb-manifest-plugin/commit/7a27d11) Prepare for release v0.10.0-rc.0 (#57)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.31.0-rc.0](https://github.com/kubedb/mariadb/releases/tag/v0.31.0-rc.0)

- [63504dc0](https://github.com/kubedb/mariadb/commit/63504dc0d) Prepare for release v0.31.0-rc.0 (#272)
- [1bf03c34](https://github.com/kubedb/mariadb/commit/1bf03c34d) Use v1 api (#271)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.7.0-rc.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.7.0-rc.0)

- [9d5d985c](https://github.com/kubedb/mariadb-archiver/commit/9d5d985c) Prepare for release v0.7.0-rc.0 (#21)
- [10687b97](https://github.com/kubedb/mariadb-archiver/commit/10687b97) Use v1 api (#20)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.27.0-rc.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.27.0-rc.0)

- [30064e39](https://github.com/kubedb/mariadb-coordinator/commit/30064e39) Prepare for release v0.27.0-rc.0 (#120)
- [f5a9ceda](https://github.com/kubedb/mariadb-coordinator/commit/f5a9ceda) Use v1 api (#119)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.7.0-rc.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.7.0-rc.0)

- [cb1cfbc](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/cb1cfbc) Prepare for release v0.7.0-rc.0 (#24)
- [7df4615](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/7df4615) Use v1 api (#23)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.5.0-rc.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.5.0-rc.0)

- [cb733d7](https://github.com/kubedb/mariadb-restic-plugin/commit/cb733d7) Prepare for release v0.5.0-rc.0 (#16)
- [0703bcd](https://github.com/kubedb/mariadb-restic-plugin/commit/0703bcd) Use v1 api (#15)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.40.0-rc.0](https://github.com/kubedb/memcached/releases/tag/v0.40.0-rc.0)

- [2190a3c8](https://github.com/kubedb/memcached/commit/2190a3c8f) Prepare for release v0.40.0-rc.0 (#450)
- [cf78ad00](https://github.com/kubedb/memcached/commit/cf78ad006) Use v1 api (#449)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.40.0-rc.0](https://github.com/kubedb/mongodb/releases/tag/v0.40.0-rc.0)

- [3f019fad](https://github.com/kubedb/mongodb/commit/3f019fadf) Prepare for release v0.40.0-rc.0 (#637)
- [dca43b32](https://github.com/kubedb/mongodb/commit/dca43b32c) Use v1 api (#635)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.8.0-rc.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.8.0-rc.0)

- [5df0d96](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/5df0d96) Prepare for release v0.8.0-rc.0 (#29)
- [f1770b5](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/f1770b5) Use v1 api (#28)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.10.0-rc.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.10.0-rc.0)

- [702d565](https://github.com/kubedb/mongodb-restic-plugin/commit/702d565) Prepare for release v0.10.0-rc.0 (#50)
- [a10270b](https://github.com/kubedb/mongodb-restic-plugin/commit/a10270b) Use v1 api (#49)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.2.0-rc.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.2.0-rc.0)

- [56e423f5](https://github.com/kubedb/mssql-coordinator/commit/56e423f5) Prepare for release v0.2.0-rc.0 (#10)
- [482a349a](https://github.com/kubedb/mssql-coordinator/commit/482a349a) Update constants to use kubedb package (#9)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.2.0-rc.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.2.0-rc.0)

- [ed78c962](https://github.com/kubedb/mssqlserver/commit/ed78c962) Prepare for release v0.2.0-rc.0 (#19)
- [bfe83703](https://github.com/kubedb/mssqlserver/commit/bfe83703) Update constants to use kubedb package (#18)
- [9cdb65f5](https://github.com/kubedb/mssqlserver/commit/9cdb65f5) Remove license check for webhook-server (#17)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.40.0-rc.0](https://github.com/kubedb/mysql/releases/tag/v0.40.0-rc.0)

- [3ca73ddd](https://github.com/kubedb/mysql/commit/3ca73ddda) Prepare for release v0.40.0-rc.0 (#629)
- [54cb812e](https://github.com/kubedb/mysql/commit/54cb812ec) Add PetSet and move on V1 API (#628)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.8.0-rc.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.8.0-rc.0)

- [b2e2904b](https://github.com/kubedb/mysql-archiver/commit/b2e2904b) Prepare for release v0.8.0-rc.0 (#35)
- [3d92a58f](https://github.com/kubedb/mysql-archiver/commit/3d92a58f) Use v1 api (#34)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.25.0-rc.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.25.0-rc.0)

- [b8c377fd](https://github.com/kubedb/mysql-coordinator/commit/b8c377fd) Prepare for release v0.25.0-rc.0 (#116)
- [f29b8f56](https://github.com/kubedb/mysql-coordinator/commit/f29b8f56) Update constants to use kubedb package (#115)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.8.0-rc.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.8.0-rc.0)

- [10e977c](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/10e977c) Prepare for release v0.8.0-rc.0 (#22)
- [94ec3c9](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/94ec3c9) Use v1 api (#21)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.10.0-rc.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.10.0-rc.0)

- [83efb51](https://github.com/kubedb/mysql-restic-plugin/commit/83efb51) Prepare for release v0.10.0-rc.0 (#45)
- [fdfd535](https://github.com/kubedb/mysql-restic-plugin/commit/fdfd535) Update API and Skip mysql.user Table (#44)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.25.0-rc.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.25.0-rc.0)




## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.34.0-rc.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.34.0-rc.0)

- [e65c886f](https://github.com/kubedb/percona-xtradb/commit/e65c886f8) Prepare for release v0.34.0-rc.0 (#370)
- [9e8f5c8b](https://github.com/kubedb/percona-xtradb/commit/9e8f5c8b7) Use v1 api (#369)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.20.0-rc.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.20.0-rc.0)

- [67e20d0d](https://github.com/kubedb/percona-xtradb-coordinator/commit/67e20d0d) Prepare for release v0.20.0-rc.0 (#75)
- [6b8544b7](https://github.com/kubedb/percona-xtradb-coordinator/commit/6b8544b7) Use v1 api (#74)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.31.0-rc.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.31.0-rc.0)

- [a26d4398](https://github.com/kubedb/pg-coordinator/commit/a26d4398) Prepare for release v0.31.0-rc.0 (#167)
- [cdd1b821](https://github.com/kubedb/pg-coordinator/commit/cdd1b821) Add PetSet Support; Use api-v1 (#156)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.34.0-rc.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.34.0-rc.0)

- [cde85494](https://github.com/kubedb/pgbouncer/commit/cde85494) Prepare for release v0.34.0-rc.0 (#336)
- [a266f397](https://github.com/kubedb/pgbouncer/commit/a266f397) Use v1 api (#334)
- [d12eb869](https://github.com/kubedb/pgbouncer/commit/d12eb869) Auth_type md5 hashing fixed (#335)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.2.0-rc.0](https://github.com/kubedb/pgpool/releases/tag/v0.2.0-rc.0)

- [fa50af41](https://github.com/kubedb/pgpool/commit/fa50af41) Prepare for release v0.2.0-rc.0 (#37)
- [64bc921d](https://github.com/kubedb/pgpool/commit/64bc921d) Update constants to use kubedb package (#36)
- [b1d4d232](https://github.com/kubedb/pgpool/commit/b1d4d232) Remove redundant TLS secret getter and make method Exportable (#35)
- [497c9eae](https://github.com/kubedb/pgpool/commit/497c9eae) Disable clickhouse in makefile (#34)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.47.0-rc.0](https://github.com/kubedb/postgres/releases/tag/v0.47.0-rc.0)

- [96a43728](https://github.com/kubedb/postgres/commit/96a43728c) Prepare for release v0.47.0-rc.0 (#737)
- [9233ae5c](https://github.com/kubedb/postgres/commit/9233ae5c6) Integrate PetSet in Postgres; Use apiv1 (#718)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.8.0-rc.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.8.0-rc.0)

- [78bba5ae](https://github.com/kubedb/postgres-archiver/commit/78bba5ae) Prepare for release v0.8.0-rc.0 (#33)
- [6d9a8d20](https://github.com/kubedb/postgres-archiver/commit/6d9a8d20) Use v1 api (#32)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.8.0-rc.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.8.0-rc.0)

- [5167fac](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/5167fac) Prepare for release v0.8.0-rc.0 (#31)
- [9cbbfce](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/9cbbfce) Use v1 api (#30)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.10.0-rc.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.10.0-rc.0)

- [c55dd9c](https://github.com/kubedb/postgres-restic-plugin/commit/c55dd9c) Prepare for release v0.10.0-rc.0 (#38)
- [5de7901](https://github.com/kubedb/postgres-restic-plugin/commit/5de7901) Use v1 api (#37)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.9.0-rc.0](https://github.com/kubedb/provider-aws/releases/tag/v0.9.0-rc.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.9.0-rc.0](https://github.com/kubedb/provider-azure/releases/tag/v0.9.0-rc.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.9.0-rc.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.9.0-rc.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.47.0-rc.0](https://github.com/kubedb/provisioner/releases/tag/v0.47.0-rc.0)

- [986b3657](https://github.com/kubedb/provisioner/commit/986b36574) Prepare for release v0.47.0-rc.0 (#100)
- [28e4e1af](https://github.com/kubedb/provisioner/commit/28e4e1af3) Update deps
- [141dbbe9](https://github.com/kubedb/provisioner/commit/141dbbe97) Update deps
- [41a56a3c](https://github.com/kubedb/provisioner/commit/41a56a3c9) Add petset implementation for postgres (#83)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.34.0-rc.0](https://github.com/kubedb/proxysql/releases/tag/v0.34.0-rc.0)

- [0b324506](https://github.com/kubedb/proxysql/commit/0b324506a) Prepare for release v0.34.0-rc.0 (#347)
- [88b4dd7f](https://github.com/kubedb/proxysql/commit/88b4dd7fe) Use v1 api (#346)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.2.0-rc.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.2.0-rc.0)

- [ad06e69b](https://github.com/kubedb/rabbitmq/commit/ad06e69b) Prepare for release v0.2.0-rc.0 (#35)
- [4d025872](https://github.com/kubedb/rabbitmq/commit/4d025872) Update constants to use kubedb package (#34)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.40.0-rc.0](https://github.com/kubedb/redis/releases/tag/v0.40.0-rc.0)

- [824f81d9](https://github.com/kubedb/redis/commit/824f81d9b) Prepare for release v0.40.0-rc.0 (#544)
- [5fadc940](https://github.com/kubedb/redis/commit/5fadc9404) Use v1 api (#542)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.26.0-rc.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.26.0-rc.0)

- [e206e7ac](https://github.com/kubedb/redis-coordinator/commit/e206e7ac) Prepare for release v0.26.0-rc.0 (#106)
- [65403ff6](https://github.com/kubedb/redis-coordinator/commit/65403ff6) Use v1 api (#105)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.10.0-rc.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.10.0-rc.0)

- [11149d9](https://github.com/kubedb/redis-restic-plugin/commit/11149d9) Prepare for release v0.10.0-rc.0 (#37)
- [1588d95](https://github.com/kubedb/redis-restic-plugin/commit/1588d95) Use v1 api (#36)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.34.0-rc.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.34.0-rc.0)

- [aa1d5719](https://github.com/kubedb/replication-mode-detector/commit/aa1d5719) Prepare for release v0.34.0-rc.0 (#272)
- [915f548e](https://github.com/kubedb/replication-mode-detector/commit/915f548e) Use v1 api (#271)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.23.0-rc.0](https://github.com/kubedb/schema-manager/releases/tag/v0.23.0-rc.0)

- [cd898070](https://github.com/kubedb/schema-manager/commit/cd898070) Prepare for release v0.23.0-rc.0 (#114)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.2.0-rc.0](https://github.com/kubedb/singlestore/releases/tag/v0.2.0-rc.0)

- [7e011ab0](https://github.com/kubedb/singlestore/commit/7e011ab0) Prepare for release v0.2.0-rc.0 (#37)
- [17623577](https://github.com/kubedb/singlestore/commit/17623577) Update API constants package (#36)
- [67d1ecb6](https://github.com/kubedb/singlestore/commit/67d1ecb6) Update Makefile and Daily test (#35)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.2.0-rc.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.2.0-rc.0)

- [06e4926](https://github.com/kubedb/singlestore-coordinator/commit/06e4926) Prepare for release v0.2.0-rc.0 (#21)
- [458fa6a](https://github.com/kubedb/singlestore-coordinator/commit/458fa6a) Update constants to use kubedb package (#20)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.5.0-rc.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.5.0-rc.0)

- [efca7ae](https://github.com/kubedb/singlestore-restic-plugin/commit/efca7ae) Prepare for release v0.5.0-rc.0 (#16)
- [ae76f5a](https://github.com/kubedb/singlestore-restic-plugin/commit/ae76f5a) Update constants to use kubedb package (#15)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.2.0-rc.0](https://github.com/kubedb/solr/releases/tag/v0.2.0-rc.0)

- [38d0c569](https://github.com/kubedb/solr/commit/38d0c569) Prepare for release v0.2.0-rc.0 (#35)
- [ca47b7be](https://github.com/kubedb/solr/commit/ca47b7be) Update constants to use kubedb package (#34)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.23.0-rc.0](https://github.com/kubedb/ui-server/releases/tag/v0.23.0-rc.0)

- [c1d29bcb](https://github.com/kubedb/ui-server/commit/c1d29bcb) Prepare for release v0.23.0-rc.0 (#122)
- [107fee8b](https://github.com/kubedb/ui-server/commit/107fee8b) version converted into v1 from v1alpha2 (#121)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.23.0-rc.0](https://github.com/kubedb/webhook-server/releases/tag/v0.23.0-rc.0)

- [d07cc360](https://github.com/kubedb/webhook-server/commit/d07cc360) Prepare for release v0.23.0-rc.0 (#112)
- [e7b7c671](https://github.com/kubedb/webhook-server/commit/e7b7c671) Add v1 api conversion webhooks (#111)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.2.0-rc.0](https://github.com/kubedb/zookeeper/releases/tag/v0.2.0-rc.0)

- [75a1fa49](https://github.com/kubedb/zookeeper/commit/75a1fa49) Prepare for release v0.2.0-rc.0 (#30)
- [bc8d242d](https://github.com/kubedb/zookeeper/commit/bc8d242d) Update constants to use kubedb package (#29)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.3.0-rc.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.3.0-rc.0)

- [3235ccf](https://github.com/kubedb/zookeeper-restic-plugin/commit/3235ccf) Prepare for release v0.3.0-rc.0 (#9)




