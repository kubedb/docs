---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2025.5.30
    name: Changelog-v2025.5.30
    parent: welcome
    weight: 20250530
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2025.5.30/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2025.5.30/
---

# KubeDB v2025.5.30 (2025-05-31)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.55.0](https://github.com/kubedb/apimachinery/releases/tag/v0.55.0)

- [e711ccbd](https://github.com/kubedb/apimachinery/commit/e711ccbd7) Report cluster mode in billing event (#1467)
- [d84b0ec5](https://github.com/kubedb/apimachinery/commit/d84b0ec51) Add zap logger (#1466)
- [5cb03546](https://github.com/kubedb/apimachinery/commit/5cb035464) Add Clickhouse opsreq: restart,vertical scale (#1456)
- [e9f8caba](https://github.com/kubedb/apimachinery/commit/e9f8cabaa) Oracle API (#1464)
- [de1412fc](https://github.com/kubedb/apimachinery/commit/de1412fcb) Add hazelcast api (#1413)
- [02cb3e12](https://github.com/kubedb/apimachinery/commit/02cb3e121) Add support for Cassandra UpdateVersion and Vertical Scale (#1458)
- [fc30e50e](https://github.com/kubedb/apimachinery/commit/fc30e50e2) Add RotateAuth opsrequest validator for Redis (#1459)
- [33f103be](https://github.com/kubedb/apimachinery/commit/33f103be5) Add ferretdb 'backend' spec (#1462)
- [8b490644](https://github.com/kubedb/apimachinery/commit/8b490644a) Add RabbitMQ RotateAuth Validation (#1461)
- [6dba2e99](https://github.com/kubedb/apimachinery/commit/6dba2e99e) Add ignite's monitor field (#1460)
- [02d581cc](https://github.com/kubedb/apimachinery/commit/02d581ccd) Redis Validation code added in webhook (#1457)
- [bd55299c](https://github.com/kubedb/apimachinery/commit/bd55299ce) Add Client Certificate API to ClickHouse (#1455)
- [b555d303](https://github.com/kubedb/apimachinery/commit/b555d3030) Fix conversion for mg & kf (#1454)
- [8d88b2b0](https://github.com/kubedb/apimachinery/commit/8d88b2b04) Fix Replicas Ready Check (#1453)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.40.0](https://github.com/kubedb/autoscaler/releases/tag/v0.40.0)

- [ae84f9ec](https://github.com/kubedb/autoscaler/commit/ae84f9ec) Prepare for release v0.40.0 (#250)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.8.0](https://github.com/kubedb/cassandra/releases/tag/v0.8.0)

- [add86028](https://github.com/kubedb/cassandra/commit/add86028) Prepare for release v0.8.0 (#36)
- [7b462ed1](https://github.com/kubedb/cassandra/commit/7b462ed1) Export methods and Update Webhook Setup (#34)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.2.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.2.0)

- [e6d035b](https://github.com/kubedb/cassandra-medusa-plugin/commit/e6d035b) Prepare for release v0.2.0 (#5)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.55.0](https://github.com/kubedb/cli/releases/tag/v0.55.0)

- [b43c9c14](https://github.com/kubedb/cli/commit/b43c9c14c) Prepare for release v0.55.0 (#796)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.10.0](https://github.com/kubedb/clickhouse/releases/tag/v0.10.0)

- [2c1ca884](https://github.com/kubedb/clickhouse/commit/2c1ca884) Prepare for release v0.10.0 (#53)
- [ab75c636](https://github.com/kubedb/clickhouse/commit/ab75c636) Update Health Check for Pod Joining in Cluster  (#49)
- [e3d9a3cf](https://github.com/kubedb/clickhouse/commit/e3d9a3cf) Add Client Certificate (#51)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.10.0](https://github.com/kubedb/crd-manager/releases/tag/v0.10.0)

- [24d7c90e](https://github.com/kubedb/crd-manager/commit/24d7c90e) Prepare for release v0.10.0 (#78)
- [fe37b349](https://github.com/kubedb/crd-manager/commit/fe37b349) Add Oracle Crds (#77)
- [d024dc2a](https://github.com/kubedb/crd-manager/commit/d024dc2a) Configure hazelcast for crd manager (#72)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.13.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.13.0)

- [b59b15d](https://github.com/kubedb/dashboard-restic-plugin/commit/b59b15d) Prepare for release v0.13.0 (#39)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.10.0](https://github.com/kubedb/db-client-go/releases/tag/v0.10.0)

- [999446c9](https://github.com/kubedb/db-client-go/commit/999446c9) Prepare for release v0.10.0 (#180)
- [0c501494](https://github.com/kubedb/db-client-go/commit/0c501494) Update virtual secrets implementation (#174)
- [7d604391](https://github.com/kubedb/db-client-go/commit/7d604391) Export pgbouncer.KubeDBClientBuilder's "getBackendAuth" method (#175)
- [e91e0389](https://github.com/kubedb/db-client-go/commit/e91e0389) Fix config (#178)
- [564ffcc7](https://github.com/kubedb/db-client-go/commit/564ffcc7) Add hazelcast client (#177)
- [e4ea3665](https://github.com/kubedb/db-client-go/commit/e4ea3665) Add oracle db client (#176)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.10.0](https://github.com/kubedb/druid/releases/tag/v0.10.0)

- [02c68fa7](https://github.com/kubedb/druid/commit/02c68fa7) Prepare for release v0.10.0 (#86)
- [07cd0d7e](https://github.com/kubedb/druid/commit/07cd0d7e) Add max-concurrent-reconciles flag (#85)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.55.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.55.0)

- [78daed93](https://github.com/kubedb/elasticsearch/commit/78daed93c) Prepare for release v0.55.0 (#766)
- [fff278ee](https://github.com/kubedb/elasticsearch/commit/fff278ee5) Add max-concurrent-reconciles flag (#765)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.18.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.18.0)

- [079b86d4](https://github.com/kubedb/elasticsearch-restic-plugin/commit/079b86d4) Prepare for release v0.18.0 (#63)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.10.0](https://github.com/kubedb/ferretdb/releases/tag/v0.10.0)

- [5ef56625](https://github.com/kubedb/ferretdb/commit/5ef56625) Prepare for release v0.10.0 (#77)
- [047b5538](https://github.com/kubedb/ferretdb/commit/047b5538) Set backend pg resources (#76)
- [d99de6d1](https://github.com/kubedb/ferretdb/commit/d99de6d1) Add max-concurrent-reconciles flag and update shedule (#75)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.3.0](https://github.com/kubedb/gitops/releases/tag/v0.3.0)

- [e7a6421a](https://github.com/kubedb/gitops/commit/e7a6421a) Prepare for release v0.3.0 (#16)
- [0f6c5a8a](https://github.com/kubedb/gitops/commit/0f6c5a8a) Add MySQL, MSSQLServer and Elasticsearch Support (#15)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.1.0](https://github.com/kubedb/hazelcast/releases/tag/v0.1.0)

- [5861b0f3](https://github.com/kubedb/hazelcast/commit/5861b0f3) Prepare for release v0.1.0 (#2)
- [776a2847](https://github.com/kubedb/hazelcast/commit/776a2847) Hazelcast Bootstrap (#1)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.2.0](https://github.com/kubedb/ignite/releases/tag/v0.2.0)

- [991ee23d](https://github.com/kubedb/ignite/commit/991ee23d) Prepare for release v0.2.0 (#8)
- [6d00303c](https://github.com/kubedb/ignite/commit/6d00303c) Add monitoring Support (#7)
- [998cfa31](https://github.com/kubedb/ignite/commit/998cfa31) Fix v2.17.0 issue (#6)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2025.5.30](https://github.com/kubedb/installer/releases/tag/v2025.5.30)

- [2443b351](https://github.com/kubedb/installer/commit/2443b3518) Prepare for release v2025.5.30 (#1722)
- [79bfb53a](https://github.com/kubedb/installer/commit/79bfb53ae) Update init images; make gen fmt; Update catalog (#1719)
- [da9b4329](https://github.com/kubedb/installer/commit/da9b4329c) Add Oracle Database Support (#1711)
- [867da9ef](https://github.com/kubedb/installer/commit/867da9ef1) Add hazelcast support (#1626)
- [650bb7b6](https://github.com/kubedb/installer/commit/650bb7b68) Update crds for kubedb/apimachinery@d84b0ec5 (#1718)
- [70ad4934](https://github.com/kubedb/installer/commit/70ad4934f) Add updateConstraints for Cassandra (#1696)
- [ec6d0f5d](https://github.com/kubedb/installer/commit/ec6d0f5d2) Add Postgres New Versions 17.5, 16.9, 15.13, 14.18, 13.21 (#1702)
- [b5d2b884](https://github.com/kubedb/installer/commit/b5d2b884b) Add Cassandra addons (#1686)
- [16420f46](https://github.com/kubedb/installer/commit/16420f46d) Add Ignite Dashboards (#1701)
- [b855d983](https://github.com/kubedb/installer/commit/b855d983b) Add New ProxySQL Version 3.0.1 (#1699)
- [5a747db3](https://github.com/kubedb/installer/commit/5a747db33) Add databases flag for singlestore addon (#1688)
- [22379995](https://github.com/kubedb/installer/commit/223799959) Update crds for kubedb/apimachinery@e9f8caba (#1715)
- [feff483b](https://github.com/kubedb/installer/commit/feff483b8) Update cve report (#1708)
- [24e5e470](https://github.com/kubedb/installer/commit/24e5e4708) Update crds for kubedb/apimachinery@02cb3e12 (#1713)
- [0a7e2dec](https://github.com/kubedb/installer/commit/0a7e2decb) Update crds for kubedb/apimachinery@fc30e50e (#1712)
- [0481aed1](https://github.com/kubedb/installer/commit/0481aed1c) Update crds for kubedb/apimachinery@6dba2e99 (#1709)
- [132e9fcd](https://github.com/kubedb/installer/commit/132e9fcd1) Update cve report (#1707)
- [82c0ac76](https://github.com/kubedb/installer/commit/82c0ac769) Update cve report (#1706)
- [1e659c45](https://github.com/kubedb/installer/commit/1e659c45d) Update cve report (#1705)
- [fba10bbb](https://github.com/kubedb/installer/commit/fba10bbbc) Update cve report (#1704)
- [7f21fadc](https://github.com/kubedb/installer/commit/7f21fadca) Update cve report (#1700)
- [1cb4a04e](https://github.com/kubedb/installer/commit/1cb4a04e2) Add SideKick, Kubestash Storage Permission (#1703)
- [5e96cf86](https://github.com/kubedb/installer/commit/5e96cf862) Update cve report (#1698)
- [7de684e6](https://github.com/kubedb/installer/commit/7de684e62) Update cve report (#1697)
- [cf862da6](https://github.com/kubedb/installer/commit/cf862da6b) Update cve report (#1695)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.26.0](https://github.com/kubedb/kafka/releases/tag/v0.26.0)

- [3a1234c5](https://github.com/kubedb/kafka/commit/3a1234c5) Prepare for release v0.26.0 (#151)
- [f21ca6e8](https://github.com/kubedb/kafka/commit/f21ca6e8) Add `max-concurrent-reconciles` flag (#150)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.31.0](https://github.com/kubedb/kibana/releases/tag/v0.31.0)

- [e5d3b649](https://github.com/kubedb/kibana/commit/e5d3b649) Prepare for release v0.31.0 (#152)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.18.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.18.0)

- [427c40ee](https://github.com/kubedb/kubedb-manifest-plugin/commit/427c40ee) Prepare for release v0.18.0 (#94)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.6.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.6.0)

- [a2a2fde](https://github.com/kubedb/kubedb-verifier/commit/a2a2fde) Prepare for release v0.6.0 (#19)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.39.0](https://github.com/kubedb/mariadb/releases/tag/v0.39.0)

- [3181d186](https://github.com/kubedb/mariadb/commit/3181d1868) Prepare for release v0.39.0 (#332)
- [f5d6d7cc](https://github.com/kubedb/mariadb/commit/f5d6d7ccd) Add TLS Support to Maxscale (#327)
- [b9689cb4](https://github.com/kubedb/mariadb/commit/b9689cb47) Fix Archiver for TLS Enabled Minio Backend (#330)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.15.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.15.0)

- [881ed377](https://github.com/kubedb/mariadb-archiver/commit/881ed377) Prepare for release v0.15.0 (#50)
- [7fa5ec27](https://github.com/kubedb/mariadb-archiver/commit/7fa5ec27) Fix Archiver for TLS Enabled Minio Backend (#49)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.35.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.35.0)

- [a9088109](https://github.com/kubedb/mariadb-coordinator/commit/a9088109) Prepare for release v0.35.0 (#143)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.15.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.15.0)

- [a9f0a3ed](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/a9f0a3ed) Prepare for release v0.15.0 (#47)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.13.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.13.0)

- [4c92acc](https://github.com/kubedb/mariadb-restic-plugin/commit/4c92acc) Prepare for release v0.13.0 (#46)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.48.0](https://github.com/kubedb/memcached/releases/tag/v0.48.0)

- [0a5ca5b9](https://github.com/kubedb/memcached/commit/0a5ca5b99) Prepare for release v0.48.0 (#500)
- [be261477](https://github.com/kubedb/memcached/commit/be261477b) Add max-concurrent-reconciles flag (#499)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.16.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.16.0)

- [9fa95203](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/9fa95203) Prepare for release v0.16.0 (#51)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.18.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.18.0)

- [c20f034](https://github.com/kubedb/mongodb-restic-plugin/commit/c20f034) Prepare for release v0.18.0 (#84)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.10.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.10.0)

- [ff3a2edc](https://github.com/kubedb/mssql-coordinator/commit/ff3a2edc) Prepare for release v0.10.0 (#37)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.10.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.10.0)

- [dc35ad0e](https://github.com/kubedb/mssqlserver/commit/dc35ad0e) Prepare for release v0.10.0 (#78)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.9.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.9.0)




## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.9.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.9.0)

- [09d3f93](https://github.com/kubedb/mssqlserver-walg-plugin/commit/09d3f93) Prepare for release v0.9.0 (#25)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.48.0](https://github.com/kubedb/mysql/releases/tag/v0.48.0)

- [a07d4e9e](https://github.com/kubedb/mysql/commit/a07d4e9ed) Prepare for release v0.48.0 (#687)
- [864661fc](https://github.com/kubedb/mysql/commit/864661fc5) Fix Archiver for TLS Enabled Minio Backend (#685)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.16.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.16.0)

- [47cd4436](https://github.com/kubedb/mysql-archiver/commit/47cd4436) Prepare for release v0.16.0 (#60)
- [1bc6de63](https://github.com/kubedb/mysql-archiver/commit/1bc6de63) Fix TLS Enabled Minio S3 (#59)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.33.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.33.0)

- [b08856d0](https://github.com/kubedb/mysql-coordinator/commit/b08856d0) Prepare for release v0.33.0 (#143)
- [8e3a873c](https://github.com/kubedb/mysql-coordinator/commit/8e3a873c) Fix semi sync version upgrade issue for version <8.4.2 to >=8.4.2 (#142)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.16.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.16.0)

- [a70cb81b](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/a70cb81b) Prepare for release v0.16.0 (#47)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.18.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.18.0)

- [418db6c](https://github.com/kubedb/mysql-restic-plugin/commit/418db6c) Prepare for release v0.18.0 (#74)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.33.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.33.0)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.42.0](https://github.com/kubedb/ops-manager/releases/tag/v0.42.0)

- [1b5ecc78](https://github.com/kubedb/ops-manager/commit/1b5ecc780) Prepare for release v0.42.0 (#744)
- [fdaa591b](https://github.com/kubedb/ops-manager/commit/fdaa591b0) Use zap logger from apimachinery (#743)
- [1b34a52b](https://github.com/kubedb/ops-manager/commit/1b34a52b3) Change status of Redis Ops Request failed condition False (#738)
- [1031fed2](https://github.com/kubedb/ops-manager/commit/1031fed24) Hazelcast cert methods (#741)
- [0eeeace2](https://github.com/kubedb/ops-manager/commit/0eeeace2d) Add rotateauth opsrequest for redis (#736)
- [14fdfd75](https://github.com/kubedb/ops-manager/commit/14fdfd756) Add RedisSentinel Rotate Auth (#740)
- [e2222861](https://github.com/kubedb/ops-manager/commit/e22228618) Don't enable client tls in mssql reconfigure tls (#742)
- [757ec88e](https://github.com/kubedb/ops-manager/commit/757ec88ec) Add ClickHous Restart & verticalScaling (#725)
- [2053ac54](https://github.com/kubedb/ops-manager/commit/2053ac54e) Add TLS Reconfigure Support to MariaDBReplication (#735)
- [43507f00](https://github.com/kubedb/ops-manager/commit/43507f00c) Fix semi sync version upgrade issue for version <8.4.2 to >=8.4.2 (#733)
- [ba32a084](https://github.com/kubedb/ops-manager/commit/ba32a084b) Add support for Cassandra UpdateVersion & VerticalScaling (#737)
- [75db9507](https://github.com/kubedb/ops-manager/commit/75db95074) Add RabbitMQ auth rotate (#734)
- [1f7a50a9](https://github.com/kubedb/ops-manager/commit/1f7a50a9a) Set backend pg resources



## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.1.0](https://github.com/kubedb/oracle/releases/tag/v0.1.0)

- [228096cc](https://github.com/kubedb/oracle/commit/228096cc) Prepare for release v0.1.0 (#3)
- [fef9cee2](https://github.com/kubedb/oracle/commit/fef9cee2) Implement Oracle



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.1.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.1.0)

- [c31da55](https://github.com/kubedb/oracle-coordinator/commit/c31da55) Prepare for release v0.1.0 (#3)
- [be2d18f](https://github.com/kubedb/oracle-coordinator/commit/be2d18f) Oracle Coordinator For DataGuard Mode (#2)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.42.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.42.0)

- [744acb1f](https://github.com/kubedb/percona-xtradb/commit/744acb1fa) Prepare for release v0.42.0 (#408)
- [6d1d5228](https://github.com/kubedb/percona-xtradb/commit/6d1d5228c) Add max-concurrent-reconciles flag to operator (#407)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.28.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.28.0)

- [58ac9747](https://github.com/kubedb/percona-xtradb-coordinator/commit/58ac9747) Prepare for release v0.28.0 (#96)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.39.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.39.0)

- [c0856f18](https://github.com/kubedb/pg-coordinator/commit/c0856f18) Prepare for release v0.39.0 (#200)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.42.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.42.0)

- [9caa0b59](https://github.com/kubedb/pgbouncer/commit/9caa0b59) Prepare for release v0.42.0 (#372)
- [13d97001](https://github.com/kubedb/pgbouncer/commit/13d97001) Add max-concurrent-reconciles flag (#371)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.10.0](https://github.com/kubedb/pgpool/releases/tag/v0.10.0)

- [af86700f](https://github.com/kubedb/pgpool/commit/af86700f) Prepare for release v0.10.0 (#75)
- [d8f1db30](https://github.com/kubedb/pgpool/commit/d8f1db30) Add max-concurrent-reconciles flag (#74)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.55.0](https://github.com/kubedb/postgres/releases/tag/v0.55.0)

- [2cff2ac0](https://github.com/kubedb/postgres/commit/2cff2ac00) Prepare for release v0.55.0 (#816)
- [f719b859](https://github.com/kubedb/postgres/commit/f719b8598) Update Virtual-Secrets Implementation (#813)
- [3169b565](https://github.com/kubedb/postgres/commit/3169b5653) Add max-concurrent-reconciles flag (#815)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.16.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.16.0)

- [f99e631e](https://github.com/kubedb/postgres-archiver/commit/f99e631e) Prepare for release v0.16.0 (#61)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.16.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.16.0)

- [dbee5a03](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/dbee5a03) Prepare for release v0.16.0 (#57)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.18.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.18.0)

- [38e3457](https://github.com/kubedb/postgres-restic-plugin/commit/38e3457) Prepare for release v0.18.0 (#71)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.16.0](https://github.com/kubedb/provider-aws/releases/tag/v0.16.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.16.0](https://github.com/kubedb/provider-azure/releases/tag/v0.16.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.16.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.16.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.55.0](https://github.com/kubedb/provisioner/releases/tag/v0.55.0)

- [591de3a4](https://github.com/kubedb/provisioner/commit/591de3a4e) Prepare for release v0.55.0 (#152)
- [0047c550](https://github.com/kubedb/provisioner/commit/0047c550e) Add Oracle Database Support (#151)
- [e8e345bb](https://github.com/kubedb/provisioner/commit/e8e345bbb) Remove .DS_Store



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.42.0](https://github.com/kubedb/proxysql/releases/tag/v0.42.0)

- [6ff3d865](https://github.com/kubedb/proxysql/commit/6ff3d865b) Prepare for release v0.42.0 (#394)
- [e73b19b1](https://github.com/kubedb/proxysql/commit/e73b19b17) Pass proxyql version for mysql new auth plugin setup (#393)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.10.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.10.0)

- [3b744fc5](https://github.com/kubedb/rabbitmq/commit/3b744fc5) Prepare for release v0.10.0 (#85)
- [2615eeb9](https://github.com/kubedb/rabbitmq/commit/2615eeb9) Ensure RabbitMQadminCredSecret (#83)
- [bc8cf613](https://github.com/kubedb/rabbitmq/commit/bc8cf613) Add max-concurrent-reconciles flag (#84)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.48.0](https://github.com/kubedb/redis/releases/tag/v0.48.0)

- [b932e992](https://github.com/kubedb/redis/commit/b932e9925) Prepare for release v0.48.0 (#594)
- [42cddee8](https://github.com/kubedb/redis/commit/42cddee89) Signed-off-by: HiranmoyChowdhury <hiranmoy@appscode.com> (#593)
- [c470ef7c](https://github.com/kubedb/redis/commit/c470ef7c9) Add REDISCLI_AUTH env for valkey (#592)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.34.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.34.0)

- [66c710e3](https://github.com/kubedb/redis-coordinator/commit/66c710e3) Prepare for release v0.34.0 (#128)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.18.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.18.0)

- [05a57a5](https://github.com/kubedb/redis-restic-plugin/commit/05a57a5) Prepare for release v0.18.0 (#66)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.42.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.42.0)

- [0f4b6da2](https://github.com/kubedb/replication-mode-detector/commit/0f4b6da2) Prepare for release v0.42.0 (#293)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.31.0](https://github.com/kubedb/schema-manager/releases/tag/v0.31.0)

- [dd78a5b9](https://github.com/kubedb/schema-manager/commit/dd78a5b9) Prepare for release v0.31.0 (#139)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.10.0](https://github.com/kubedb/singlestore/releases/tag/v0.10.0)

- [39ed6595](https://github.com/kubedb/singlestore/commit/39ed6595) Prepare for release v0.10.0 (#73)
- [036f3e00](https://github.com/kubedb/singlestore/commit/036f3e00) Update Schedule in daily.yml (#72)
- [107e9c8b](https://github.com/kubedb/singlestore/commit/107e9c8b) Update Webhook Setup (#71)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.10.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.10.0)

- [5269ad4](https://github.com/kubedb/singlestore-coordinator/commit/5269ad4) Prepare for release v0.10.0 (#43)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.13.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.13.0)

- [74866f9](https://github.com/kubedb/singlestore-restic-plugin/commit/74866f9) Prepare for release v0.13.0 (#44)
- [33e171a](https://github.com/kubedb/singlestore-restic-plugin/commit/33e171a) Fix for empty database (#43)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.10.0](https://github.com/kubedb/solr/releases/tag/v0.10.0)

- [a3588381](https://github.com/kubedb/solr/commit/a3588381) Prepare for release v0.10.0 (#85)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.40.0](https://github.com/kubedb/tests/releases/tag/v0.40.0)

- [1c24b4ef](https://github.com/kubedb/tests/commit/1c24b4ef) Prepare for release v0.40.0 (#464)
- [ae1e8e90](https://github.com/kubedb/tests/commit/ae1e8e90) Update deps
- [d854f970](https://github.com/kubedb/tests/commit/d854f970) Add pgbouncer tls (#398)
- [8ec76bc1](https://github.com/kubedb/tests/commit/8ec76bc1) kakfa change completed (#408)
- [fab46597](https://github.com/kubedb/tests/commit/fab46597) Fix MySQL semi-sync ops-request, general (#423)
- [202542ef](https://github.com/kubedb/tests/commit/202542ef) Add Archiver to backup-restore CI (#457)
- [ddbf9c5f](https://github.com/kubedb/tests/commit/ddbf9c5f) Add restic backup-restore e2e test for mysql innodb, semi-sync mode (#422)
- [3ae387db](https://github.com/kubedb/tests/commit/3ae387db) Add Redis Cluster Health Test (#448)
- [4f454ea2](https://github.com/kubedb/tests/commit/4f454ea2) Add Tests for MariaDB Replication Topology (#456)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.31.0](https://github.com/kubedb/ui-server/releases/tag/v0.31.0)

- [08c7ea52](https://github.com/kubedb/ui-server/commit/08c7ea52) Prepare for release v0.31.0 (#163)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.31.0](https://github.com/kubedb/webhook-server/releases/tag/v0.31.0)

- [2e1a0c20](https://github.com/kubedb/webhook-server/commit/2e1a0c20) Prepare for release v0.31.0 (#157)
- [b972d532](https://github.com/kubedb/webhook-server/commit/b972d532) Add Oracle & Hazlecast Webhook Support (#156)



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.4.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.4.0)

- [f793b7d](https://github.com/kubedb/xtrabackup-restic-plugin/commit/f793b7d) Prepare for release v0.4.0 (#13)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.10.0](https://github.com/kubedb/zookeeper/releases/tag/v0.10.0)

- [da5c78d6](https://github.com/kubedb/zookeeper/commit/da5c78d6) Prepare for release v0.10.0 (#77)
- [2c0d2bc3](https://github.com/kubedb/zookeeper/commit/2c0d2bc3) Add flag to configure max concurrent reconciles (#76)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.11.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.11.0)

- [46d762e](https://github.com/kubedb/zookeeper-restic-plugin/commit/46d762e) Prepare for release v0.11.0 (#35)




