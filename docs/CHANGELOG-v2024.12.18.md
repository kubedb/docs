---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2024.12.18
    name: Changelog-v2024.12.18
    parent: welcome
    weight: 20241218
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2024.12.18/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2024.12.18/
---

# KubeDB v2024.12.18 (2024-12-20)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.50.0](https://github.com/kubedb/apimachinery/releases/tag/v0.50.0)

- [34e3d3f1](https://github.com/kubedb/apimachinery/commit/34e3d3f1f) Update deps
- [61e95e45](https://github.com/kubedb/apimachinery/commit/61e95e454) Add available database list in connection (#1371)
- [e9eda10c](https://github.com/kubedb/apimachinery/commit/e9eda10ca) Fix availabilityGroup panic; Set defults to logBackupOptions (#1370)
- [a471a5b0](https://github.com/kubedb/apimachinery/commit/a471a5b03) Update RabitMQ Constants (#1365)
- [e677abc2](https://github.com/kubedb/apimachinery/commit/e677abc2b) Add WAl backup failed and success history limit field (#1368)
- [d35054f5](https://github.com/kubedb/apimachinery/commit/d35054f53) Add opensearch rotateauth constants (#1369)
- [52bb8340](https://github.com/kubedb/apimachinery/commit/52bb83407) Fix RabbitMQ Default Authsecret name (#1367)
- [886ff390](https://github.com/kubedb/apimachinery/commit/886ff3903) Add kafka controller quorum constant (#1366)
- [f9bb8f9d](https://github.com/kubedb/apimachinery/commit/f9bb8f9d8) Update deps
- [90dd367e](https://github.com/kubedb/apimachinery/commit/90dd367e0) add remote to group replication ops request (#1359)
- [fb7e8ddf](https://github.com/kubedb/apimachinery/commit/fb7e8ddf6) Remove redundant database phase (#1358)
- [5c754b01](https://github.com/kubedb/apimachinery/commit/5c754b01e) Set Default Postgres StandbyMode (#1364)
- [97a0eb7c](https://github.com/kubedb/apimachinery/commit/97a0eb7c1) Add const for ES Master node shard allocation disabling (#1361)
- [9ab2b7bd](https://github.com/kubedb/apimachinery/commit/9ab2b7bdd) Add AutoOps for Kafka (#1362)
- [cac99cb9](https://github.com/kubedb/apimachinery/commit/cac99cb95) Update for release KubeStash@v2024.12.9 (#1360)
- [392b6fe3](https://github.com/kubedb/apimachinery/commit/392b6fe3e) Add ReplicationStrategy Default Value for MySQL (#1356)
- [2198d910](https://github.com/kubedb/apimachinery/commit/2198d9109) Add helpers to use different labels in sidekick (#1357)
- [cfbad6a2](https://github.com/kubedb/apimachinery/commit/cfbad6a2d) Add Memcahced Authentication Constant (#1355)
- [5c29840d](https://github.com/kubedb/apimachinery/commit/5c29840da) Add archiver contants (#1352)
- [d61bcdd7](https://github.com/kubedb/apimachinery/commit/d61bcdd70) Update deps
- [07482109](https://github.com/kubedb/apimachinery/commit/074821090) Update sidekick
- [1e6df971](https://github.com/kubedb/apimachinery/commit/1e6df9713) Rename auth-active-from annotation (#1354)
- [6ec77631](https://github.com/kubedb/apimachinery/commit/6ec77631d) Add ExtractDatabaseInfo Func for Cassandra (#1351)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.35.0](https://github.com/kubedb/autoscaler/releases/tag/v0.35.0)

- [d6768b22](https://github.com/kubedb/autoscaler/commit/d6768b22) Prepare for release v0.35.0 (#232)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.3.0](https://github.com/kubedb/cassandra/releases/tag/v0.3.0)

- [f9865ce8](https://github.com/kubedb/cassandra/commit/f9865ce8) Prepare for release v0.3.0 (#13)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.50.0](https://github.com/kubedb/cli/releases/tag/v0.50.0)

- [f7e7b8fd](https://github.com/kubedb/cli/commit/f7e7b8fd) Prepare for release v0.50.0 (#782)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.5.0](https://github.com/kubedb/clickhouse/releases/tag/v0.5.0)

- [bf104e52](https://github.com/kubedb/clickhouse/commit/bf104e52) Prepare for release v0.5.0 (#27)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.5.0](https://github.com/kubedb/crd-manager/releases/tag/v0.5.0)

- [debe8b5f](https://github.com/kubedb/crd-manager/commit/debe8b5f) Prepare for release v0.5.0 (#57)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.8.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.8.0)

- [cdd6918](https://github.com/kubedb/dashboard-restic-plugin/commit/cdd6918) Prepare for release v0.8.0 (#26)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.5.0](https://github.com/kubedb/db-client-go/releases/tag/v0.5.0)

- [d6abb5b9](https://github.com/kubedb/db-client-go/commit/d6abb5b9) Prepare for release v0.5.0 (#154)
- [61b1dcbb](https://github.com/kubedb/db-client-go/commit/61b1dcbb) Return amqp channel with RabbitMQ client (#153)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.5.0](https://github.com/kubedb/druid/releases/tag/v0.5.0)

- [0504c4f6](https://github.com/kubedb/druid/commit/0504c4f6) Prepare for release v0.5.0 (#64)
- [498b668c](https://github.com/kubedb/druid/commit/498b668c) Use `DatabasePhase` instead of `DruidPhase` (#63)
- [678142c7](https://github.com/kubedb/druid/commit/678142c7) Update `AuthActiveFromAnnotation` const (#62)
- [039004a7](https://github.com/kubedb/druid/commit/039004a7) Fix Druid active-from Annotation check in Auth Secret (#61)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.50.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.50.0)

- [b616d048](https://github.com/kubedb/elasticsearch/commit/b616d0488) Prepare for release v0.50.0 (#744)
- [21a3fb0f](https://github.com/kubedb/elasticsearch/commit/21a3fb0ff) Add method to get internal config for opensearch (#743)
- [c1e9afd3](https://github.com/kubedb/elasticsearch/commit/c1e9afd34) Updated annotation name (#742)
- [f302a147](https://github.com/kubedb/elasticsearch/commit/f302a1479) Remove old statefulset (#741)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.13.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.13.0)

- [7447e2da](https://github.com/kubedb/elasticsearch-restic-plugin/commit/7447e2da) Prepare for release v0.13.0 (#50)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.5.0](https://github.com/kubedb/ferretdb/releases/tag/v0.5.0)

- [d25c38ce](https://github.com/kubedb/ferretdb/commit/d25c38ce) Prepare for release v0.5.0 (#53)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2024.12.18](https://github.com/kubedb/installer/releases/tag/v2024.12.18)

- [eccf9762](https://github.com/kubedb/installer/commit/eccf9762) Prepare for release v2024.12.18 (#1480)
- [d7bae1ed](https://github.com/kubedb/installer/commit/d7bae1ed) Add MysQL 8.4.3 & 9.1.0 (#1477)
- [4a14d08d](https://github.com/kubedb/installer/commit/4a14d08d) Update wal-g image for mongo (#1458)
- [e5cda56a](https://github.com/kubedb/installer/commit/e5cda56a) Add mongo 8.0.4 (#1476)
- [4c8df0a5](https://github.com/kubedb/installer/commit/4c8df0a5) Add kafka connector version (#1474)
- [a6a9317c](https://github.com/kubedb/installer/commit/a6a9317c) Update cve report (#1473)
- [ed4bca04](https://github.com/kubedb/installer/commit/ed4bca04) Use postgres init 0.17.0
- [9f421c0b](https://github.com/kubedb/installer/commit/9f421c0b) Add druid version 31.0.0 (#1472)
- [65a4fd96](https://github.com/kubedb/installer/commit/65a4fd96) Update crds for kubedb/apimachinery@e677abc2 (#1471)
- [bd60d90c](https://github.com/kubedb/installer/commit/bd60d90c) Update init images for ES & OS (#1469)
- [1e10e58c](https://github.com/kubedb/installer/commit/1e10e58c) Fix Kafka updateConstraints allowList (#1470)
- [c3611296](https://github.com/kubedb/installer/commit/c3611296) Update cve report (#1468)
- [73609f81](https://github.com/kubedb/installer/commit/73609f81) Add new redis versions (#1425)
- [b4e15c46](https://github.com/kubedb/installer/commit/b4e15c46) Add RabbitMQ v4.0.4 and Memcached v1.6.33 (#1454)
- [97abfca7](https://github.com/kubedb/installer/commit/97abfca7) Add MariaDB 11.6.2 Version Support (#1466)
- [8b9aac45](https://github.com/kubedb/installer/commit/8b9aac45) Add MySQL and SingleStore new Version (#1467)
- [5e27a293](https://github.com/kubedb/installer/commit/5e27a293) Add mssql new version: 2022-cu16 (#1449)
- [1a894a77](https://github.com/kubedb/installer/commit/1a894a77) Update crds for kubedb/apimachinery@52bb8340 (#1465)
- [de09e33d](https://github.com/kubedb/installer/commit/de09e33d) Update opensearch init image for openshift (#1462)
- [9e94042e](https://github.com/kubedb/installer/commit/9e94042e) Add/Deprecate Kafka Versions (#1456)
- [aca49c1d](https://github.com/kubedb/installer/commit/aca49c1d) Update crds for kubedb/apimachinery@886ff390 (#1464)
- [6d574d16](https://github.com/kubedb/installer/commit/6d574d16) Update cve report (#1463)
- [eaff2a8a](https://github.com/kubedb/installer/commit/eaff2a8a) Update deps
- [f9c4d04d](https://github.com/kubedb/installer/commit/f9c4d04d) Update cve report (#1461)
- [7c90c55f](https://github.com/kubedb/installer/commit/7c90c55f) Update cve report (#1460)
- [4d396b83](https://github.com/kubedb/installer/commit/4d396b83) Add verification function (#1426)
- [8a7412fb](https://github.com/kubedb/installer/commit/8a7412fb) Update cve report (#1455)
- [373544f4](https://github.com/kubedb/installer/commit/373544f4) Add FerretDB version 1.24.0 (#1428)
- [f3ab9862](https://github.com/kubedb/installer/commit/f3ab9862) Add new version for ES, OS & Solr (#1429)
- [519d3e3c](https://github.com/kubedb/installer/commit/519d3e3c) Update crds for kubedb/apimachinery@fb7e8ddf (#1452)
- [dab84f58](https://github.com/kubedb/installer/commit/dab84f58) Update cve report (#1451)
- [c230da3a](https://github.com/kubedb/installer/commit/c230da3a) Update crds for kubedb/apimachinery@9ab2b7bd (#1450)
- [6f943d37](https://github.com/kubedb/installer/commit/6f943d37) Update cve report (#1448)
- [47efe72a](https://github.com/kubedb/installer/commit/47efe72a) Update deps
- [dda9a5b3](https://github.com/kubedb/installer/commit/dda9a5b3) Check for image architecture (#1446)
- [76ca0d85](https://github.com/kubedb/installer/commit/76ca0d85) Update cve report (#1447)
- [f069d099](https://github.com/kubedb/installer/commit/f069d099) Update cve report (#1445)
- [22795fd1](https://github.com/kubedb/installer/commit/22795fd1) Update toybox images to v0.8.11 with ARM64 support (#1444)
- [76b7bd4c](https://github.com/kubedb/installer/commit/76b7bd4c) Update cve report (#1443)
- [31567515](https://github.com/kubedb/installer/commit/31567515) Update cve report (#1442)
- [164ed05d](https://github.com/kubedb/installer/commit/164ed05d) Update cve report (#1441)
- [9b3ca184](https://github.com/kubedb/installer/commit/9b3ca184) Update cve report (#1440)
- [7bc3619b](https://github.com/kubedb/installer/commit/7bc3619b) Update cve report (#1438)
- [5ae176a2](https://github.com/kubedb/installer/commit/5ae176a2) Update cve report (#1437)
- [4732ae81](https://github.com/kubedb/installer/commit/4732ae81) Update cve report (#1436)
- [f3c8c57d](https://github.com/kubedb/installer/commit/f3c8c57d) Update cve report (#1435)
- [19ace325](https://github.com/kubedb/installer/commit/19ace325) Update cve report (#1434)
- [6c8bb627](https://github.com/kubedb/installer/commit/6c8bb627) Update cve report (#1433)
- [9f1c07e1](https://github.com/kubedb/installer/commit/9f1c07e1) Update cve report (#1432)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.21.0](https://github.com/kubedb/kafka/releases/tag/v0.21.0)

- [f07497c0](https://github.com/kubedb/kafka/commit/f07497c0) Prepare for release v0.21.0 (#126)
- [a4216f60](https://github.com/kubedb/kafka/commit/a4216f60) Fix .spec.authSecret Patch overriding user provided rotateAfter (#125)
- [d65a2fff](https://github.com/kubedb/kafka/commit/d65a2fff) Update configuration for new version and improve logs (#123)
- [f0762832](https://github.com/kubedb/kafka/commit/f0762832) Fix kafka connector deletion panic (#122)
- [01081fc9](https://github.com/kubedb/kafka/commit/01081fc9) Update auth active from annotations (#121)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.26.0](https://github.com/kubedb/kibana/releases/tag/v0.26.0)

- [1c698737](https://github.com/kubedb/kibana/commit/1c698737) Prepare for release v0.26.0 (#132)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.13.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.13.0)

- [b9ef54b](https://github.com/kubedb/kubedb-manifest-plugin/commit/b9ef54b) Prepare for release v0.13.0 (#81)
- [cbe4eac](https://github.com/kubedb/kubedb-manifest-plugin/commit/cbe4eac) Add `os.Stat` error check and unnecessary boolean init (#79)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.1.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.1.0)

- [ba99265](https://github.com/kubedb/kubedb-verifier/commit/ba99265) Make scripts executable (#7)
- [7ce9481](https://github.com/kubedb/kubedb-verifier/commit/7ce9481) Prepare for release v0.1.0 (#6)
- [5aaf694](https://github.com/kubedb/kubedb-verifier/commit/5aaf694) Update pkg name (#5)
- [d251cc0](https://github.com/kubedb/kubedb-verifier/commit/d251cc0) Use kind v0.25.0 (#4)
- [4921a00](https://github.com/kubedb/kubedb-verifier/commit/4921a00) Update plugin name (#3)
- [484c28c](https://github.com/kubedb/kubedb-verifier/commit/484c28c) Remove cherry-pick workflow (#2)
- [e9e4d82](https://github.com/kubedb/kubedb-verifier/commit/e9e4d82) Add support for backup verification (#1)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.34.0](https://github.com/kubedb/mariadb/releases/tag/v0.34.0)

- [b0d77276](https://github.com/kubedb/mariadb/commit/b0d77276d) Prepare for release v0.34.0 (#299)
- [6c3d91ef](https://github.com/kubedb/mariadb/commit/6c3d91efc) Stop checking archiver allowness if ref given (#298)
- [c8213ab5](https://github.com/kubedb/mariadb/commit/c8213ab58) Add Increamental Snapshot (#297)
- [d4e04a2a](https://github.com/kubedb/mariadb/commit/d4e04a2a2) Use different labels for sidekick (#293)
- [dbda72bb](https://github.com/kubedb/mariadb/commit/dbda72bb3) Add Support for Rotate Auth  (#292)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.30.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.30.0)

- [449d50d8](https://github.com/kubedb/mariadb-coordinator/commit/449d50d8) Prepare for release v0.30.0 (#131)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.10.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.10.0)

- [de217a7](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/de217a7) Prepare for release v0.10.0 (#35)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.8.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.8.0)

- [2daabda](https://github.com/kubedb/mariadb-restic-plugin/commit/2daabda) Prepare for release v0.8.0 (#32)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.43.0](https://github.com/kubedb/memcached/releases/tag/v0.43.0)

- [05dd3217](https://github.com/kubedb/memcached/commit/05dd3217e) Prepare for release v0.43.0 (#476)
- [6ce84f7b](https://github.com/kubedb/memcached/commit/6ce84f7bc) Add Rotate Auth Stuffs (#474)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.43.0](https://github.com/kubedb/mongodb/releases/tag/v0.43.0)

- [2c7aeb18](https://github.com/kubedb/mongodb/commit/2c7aeb186) Prepare for release v0.43.0 (#673)
- [c7aa9063](https://github.com/kubedb/mongodb/commit/c7aa90638) update snapshot comps phase to succeed before database and arch deletion (#672)
- [e751b02a](https://github.com/kubedb/mongodb/commit/e751b02af) pass inc. snapshot limit as args to wal-g (#671)
- [9f301c9c](https://github.com/kubedb/mongodb/commit/9f301c9c7) fix auth secret panic in mongodb archiver manifest restore (#669)
- [44777c86](https://github.com/kubedb/mongodb/commit/44777c86a) Use different labels for sidekick (#668)
- [d69f0f40](https://github.com/kubedb/mongodb/commit/d69f0f405) update `AuthActiveFromAnnotation` const in secret (#667)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.11.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.11.0)

- [e46c5ae](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/e46c5ae) Prepare for release v0.11.0 (#40)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.13.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.13.0)

- [1a813ed](https://github.com/kubedb/mongodb-restic-plugin/commit/1a813ed) Prepare for release v0.13.0 (#72)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.5.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.5.0)

- [80570f6a](https://github.com/kubedb/mssql-coordinator/commit/80570f6a) Prepare for release v0.5.0 (#23)
- [3c1d7695](https://github.com/kubedb/mssql-coordinator/commit/3c1d7695) Update the database object at the beginning of each iteration (#22)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.5.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.5.0)

- [d584daa9](https://github.com/kubedb/mssqlserver/commit/d584daa9) Prepare for release v0.5.0 (#49)
- [ff81045d](https://github.com/kubedb/mssqlserver/commit/ff81045d) Add incremental snapshot feature for last log archive time  (#48)
- [72c4eee5](https://github.com/kubedb/mssqlserver/commit/72c4eee5) Add RotateAuth related changes (#42)
- [492a76e3](https://github.com/kubedb/mssqlserver/commit/492a76e3) Use different labels for sidekick (#44)
- [f2b6e388](https://github.com/kubedb/mssqlserver/commit/f2b6e388) Use Archiver constants from apimachinery (#41)
- [80259ceb](https://github.com/kubedb/mssqlserver/commit/80259ceb) Fix Availability Group Databases adding log (#43)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.4.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.4.0)

- [a2ce047](https://github.com/kubedb/mssqlserver-archiver/commit/a2ce047) Update deps (#7)
- [503c917](https://github.com/kubedb/mssqlserver-archiver/commit/503c917) Update wal-g to v2024.12.18 (#6)
- [bc4c5a9](https://github.com/kubedb/mssqlserver-archiver/commit/bc4c5a9) Update log push start/end time in incremental snapshot (#5)



## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.4.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.4.0)

- [075fdbc](https://github.com/kubedb/mssqlserver-walg-plugin/commit/075fdbc) Prepare for release v0.4.0 (#11)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.43.0](https://github.com/kubedb/mysql/releases/tag/v0.43.0)

- [f825bfbc](https://github.com/kubedb/mysql/commit/f825bfbc6) Prepare for release v0.43.0 (#657)
- [c5c16712](https://github.com/kubedb/mysql/commit/c5c167123) Stop checking archiver allowness if ref given
- [c5b77115](https://github.com/kubedb/mysql/commit/c5b771157) Add Increamental Snapshot (#656)
- [f0234d44](https://github.com/kubedb/mysql/commit/f0234d449) Update CI for Daily Test (#655)
- [9eceaab8](https://github.com/kubedb/mysql/commit/9eceaab8a) Set PITR_RESTORE env to InitContainer (#654)
- [7ed83cc1](https://github.com/kubedb/mysql/commit/7ed83cc11) Add XtraBackup Base Backup Support for Archiver (#647)
- [e2bb04af](https://github.com/kubedb/mysql/commit/e2bb04afd) Use different labels for sidekick (#652)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.11.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.11.0)

- [5e43ade5](https://github.com/kubedb/mysql-archiver/commit/5e43ade5) Prepare for release v0.11.0 (#47)
- [3e469aa5](https://github.com/kubedb/mysql-archiver/commit/3e469aa5) Update wal-g to v2024.12.18 (#46)
- [d8e596e7](https://github.com/kubedb/mysql-archiver/commit/d8e596e7) Update Binlog Push Time to Snapshot (#45)
- [0bd79d60](https://github.com/kubedb/mysql-archiver/commit/0bd79d60) Add XtraBackup Base Backup Support for Archiver (#42)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.28.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.28.0)

- [8fc7747b](https://github.com/kubedb/mysql-coordinator/commit/8fc7747b) Prepare for release v0.28.0 (#129)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.11.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.11.0)

- [0dad487](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/0dad487) Prepare for release v0.11.0 (#36)
- [111ada4](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/111ada4) Add Support for Standalone (#35)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.13.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.13.0)

- [da314e7](https://github.com/kubedb/mysql-restic-plugin/commit/da314e7) Prepare for release v0.13.0 (#63)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.28.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.28.0)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.37.0](https://github.com/kubedb/ops-manager/releases/tag/v0.37.0)

- [626de627](https://github.com/kubedb/ops-manager/commit/626de6273) Prepare for release v0.37.0 (#692)
- [4f6c6473](https://github.com/kubedb/ops-manager/commit/4f6c6473f) Fix postgres version upgrade, reconfigure, reconfigure tls (#684)
- [6a9280e5](https://github.com/kubedb/ops-manager/commit/6a9280e56) Fix opensearch rotateauth issue (#690)
- [a1fcd3a7](https://github.com/kubedb/ops-manager/commit/a1fcd3a7b) Add RotateAuth ops request for mssqlserver (#673)
- [c739da4f](https://github.com/kubedb/ops-manager/commit/c739da4f4) Fix Semi Sync SQL syntax for v8.4.2 (#689)
- [fae25715](https://github.com/kubedb/ops-manager/commit/fae257156) Add MySQL Replication Mode Transformation Ops-request (#686)
- [9105a353](https://github.com/kubedb/ops-manager/commit/9105a353f) Fix RabbitMQ and Druid phase const (#688)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.37.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.37.0)

- [2eaf102a](https://github.com/kubedb/percona-xtradb/commit/2eaf102ae) Prepare for release v0.37.0 (#387)
- [1c679aac](https://github.com/kubedb/percona-xtradb/commit/1c679aac3) Add Support for Rotate Auth  (#385)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.23.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.23.0)

- [247ad142](https://github.com/kubedb/percona-xtradb-coordinator/commit/247ad142) Prepare for release v0.23.0 (#85)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.34.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.34.0)

- [c05b8ff9](https://github.com/kubedb/pg-coordinator/commit/c05b8ff9) Prepare for release v0.34.0 (#179)
- [04b387c5](https://github.com/kubedb/pg-coordinator/commit/04b387c5) Update LSN position in raft continuously  (#178)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.37.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.37.0)

- [a5139b71](https://github.com/kubedb/pgbouncer/commit/a5139b71) Prepare for release v0.37.0 (#353)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.5.0](https://github.com/kubedb/pgpool/releases/tag/v0.5.0)

- [258de2d9](https://github.com/kubedb/pgpool/commit/258de2d9) Prepare for release v0.5.0 (#54)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.50.0](https://github.com/kubedb/postgres/releases/tag/v0.50.0)

- [cdd642cf](https://github.com/kubedb/postgres/commit/cdd642cf1) Prepare for release v0.50.0 (#778)
- [213338de](https://github.com/kubedb/postgres/commit/213338dea) Add support for pitr restore to the latest point (#776)
- [d5b49d54](https://github.com/kubedb/postgres/commit/d5b49d544) Only use patch call when patching active from annotations (#777)
- [49471781](https://github.com/kubedb/postgres/commit/494717817) Add User in manifest restore (#774)
- [37a0cbba](https://github.com/kubedb/postgres/commit/37a0cbba6) Fix Daily (#769)
- [686c6240](https://github.com/kubedb/postgres/commit/686c62408) Use different labels for sidekick (#773)
- [3213e090](https://github.com/kubedb/postgres/commit/3213e0902) Update rotate auth annotation (#772)
- [99cb90f1](https://github.com/kubedb/postgres/commit/99cb90f17) Fix Archiver for Minio Backend (#771)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.11.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.11.0)

- [e105966d](https://github.com/kubedb/postgres-archiver/commit/e105966d) Prepare for release v0.11.0 (#45)
- [7cf65461](https://github.com/kubedb/postgres-archiver/commit/7cf65461) Update wal-g to v2024.12.18 (#44)
- [2186a7fd](https://github.com/kubedb/postgres-archiver/commit/2186a7fd) Add last commit timestamp for pitr recovery (#43)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.11.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.11.0)

- [ebc6acf](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/ebc6acf) Prepare for release v0.11.0 (#44)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.13.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.13.0)

- [757d2fb](https://github.com/kubedb/postgres-restic-plugin/commit/757d2fb) Prepare for release v0.13.0 (#58)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.12.0](https://github.com/kubedb/provider-aws/releases/tag/v0.12.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.12.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.12.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.50.0](https://github.com/kubedb/provisioner/releases/tag/v0.50.0)

- [6fa74bc8](https://github.com/kubedb/provisioner/commit/6fa74bc8f) Prepare for release v0.50.0 (#127)
- [a3014bcd](https://github.com/kubedb/provisioner/commit/a3014bcd9) Parse verbosity flag and use it for RabbitMQ controller (#126)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.37.0](https://github.com/kubedb/proxysql/releases/tag/v0.37.0)

- [8a1f6058](https://github.com/kubedb/proxysql/commit/8a1f6058b) Prepare for release v0.37.0 (#369)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.5.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.5.0)

- [88965dca](https://github.com/kubedb/rabbitmq/commit/88965dca) Prepare for release v0.5.0 (#60)
- [199afd73](https://github.com/kubedb/rabbitmq/commit/199afd73) Add support for RabbitMQ v4.x.x (#59)
- [4ab000f8](https://github.com/kubedb/rabbitmq/commit/4ab000f8) Use Single channel for health checking and Update logger (#58)
- [ac0c09a0](https://github.com/kubedb/rabbitmq/commit/ac0c09a0) Use `DatabasePhase` instead of `RabbitMQPhase` (#57)
- [c2124756](https://github.com/kubedb/rabbitmq/commit/c2124756) Fix configSecret patch issue due to unsorted configs (#56)
- [94584794](https://github.com/kubedb/rabbitmq/commit/94584794) Add Daily test CI (#55)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.43.0](https://github.com/kubedb/redis/releases/tag/v0.43.0)

- [e0e75e30](https://github.com/kubedb/redis/commit/e0e75e30a) Prepare for release v0.43.0 (#566)
- [157bd04a](https://github.com/kubedb/redis/commit/157bd04a7) Improve log (#562)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.29.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.29.0)

- [f2264e1e](https://github.com/kubedb/redis-coordinator/commit/f2264e1e) Prepare for release v0.29.0 (#116)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.13.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.13.0)

- [683d72a](https://github.com/kubedb/redis-restic-plugin/commit/683d72a) Prepare for release v0.13.0 (#53)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.37.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.37.0)

- [9421cbe6](https://github.com/kubedb/replication-mode-detector/commit/9421cbe6) Prepare for release v0.37.0 (#282)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.26.0](https://github.com/kubedb/schema-manager/releases/tag/v0.26.0)

- [1ee0ace2](https://github.com/kubedb/schema-manager/commit/1ee0ace2) Prepare for release v0.26.0 (#126)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.5.0](https://github.com/kubedb/singlestore/releases/tag/v0.5.0)

- [caf5065b](https://github.com/kubedb/singlestore/commit/caf5065b) Prepare for release v0.5.0 (#54)
- [f9edc334](https://github.com/kubedb/singlestore/commit/f9edc334) Update Auth Secret Name (#53)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.5.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.5.0)

- [0a1d843](https://github.com/kubedb/singlestore-coordinator/commit/0a1d843) Prepare for release v0.5.0 (#31)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.8.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.8.0)

- [197f322](https://github.com/kubedb/singlestore-restic-plugin/commit/197f322) Prepare for release v0.8.0 (#30)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.5.0](https://github.com/kubedb/solr/releases/tag/v0.5.0)

- [b20a5a32](https://github.com/kubedb/solr/commit/b20a5a32) Prepare for release v0.5.0 (#61)
- [6d2fce5c](https://github.com/kubedb/solr/commit/6d2fce5c) Update annotation name (#60)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.35.0](https://github.com/kubedb/tests/releases/tag/v0.35.0)

- [59be12e0](https://github.com/kubedb/tests/commit/59be12e0) Prepare for release v0.35.0 (#424)
- [12bd698c](https://github.com/kubedb/tests/commit/12bd698c) Fix MongoDB forbidden env variable panic (#394)
- [1763ba5a](https://github.com/kubedb/tests/commit/1763ba5a) Add e2e-test for  MSSQL custom-config, env_var, health-check, generic changes (#405)
- [0951fa8e](https://github.com/kubedb/tests/commit/0951fa8e) Add MongoDB Archiver Backup-Restore Test (replicaset, shard) (#387)
- [ba878d7c](https://github.com/kubedb/tests/commit/ba878d7c) Add es test profile  fix (#410)
- [433ca25b](https://github.com/kubedb/tests/commit/433ca25b) Add Mysql test profile Fix (#409)
- [9547ab1a](https://github.com/kubedb/tests/commit/9547ab1a) cleanup issue fix for backup-restore tests (#421)
- [47cba329](https://github.com/kubedb/tests/commit/47cba329) Add Druid version upgrade (#415)
- [e3e7523c](https://github.com/kubedb/tests/commit/e3e7523c) Fix connect cluster deletion policy test (#419)
- [ed10d935](https://github.com/kubedb/tests/commit/ed10d935) Druid reconfigure custom config (#418)
- [1ea737fb](https://github.com/kubedb/tests/commit/1ea737fb) Add Druid Reconfigure TLS Test (#416)
- [37e4ab1b](https://github.com/kubedb/tests/commit/37e4ab1b) Add RabbitMQ tests (#361)
- [540ec432](https://github.com/kubedb/tests/commit/540ec432) Disable Pg Exporter test with ssl test when ssl is disabled (#414)
- [44c197dc](https://github.com/kubedb/tests/commit/44c197dc) Make common function for SetEnvs, fix panic cases (MongoDB) (#417)
- [3bf65c00](https://github.com/kubedb/tests/commit/3bf65c00) Add Postgres TLS enabled(md5) part for TerminationPolicy (#411)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.26.0](https://github.com/kubedb/ui-server/releases/tag/v0.26.0)

- [0fc8de68](https://github.com/kubedb/ui-server/commit/0fc8de68) Prepare for release v0.26.0 (#142)
- [92a52ad9](https://github.com/kubedb/ui-server/commit/92a52ad9) Set available database list for connection (#141)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.26.0](https://github.com/kubedb/webhook-server/releases/tag/v0.26.0)

- [aa70196e](https://github.com/kubedb/webhook-server/commit/aa70196e) Prepare for release v0.26.0 (#137)
- [8d4cf8c8](https://github.com/kubedb/webhook-server/commit/8d4cf8c8) Remove go.sum



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.5.0](https://github.com/kubedb/zookeeper/releases/tag/v0.5.0)

- [ae4d951c](https://github.com/kubedb/zookeeper/commit/ae4d951c) Prepare for release v0.5.0 (#52)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.6.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.6.0)

- [9bed507](https://github.com/kubedb/zookeeper-restic-plugin/commit/9bed507) Prepare for release v0.6.0 (#22)




