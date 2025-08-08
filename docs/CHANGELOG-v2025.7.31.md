---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2025.7.31
    name: Changelog-v2025.7.31
    parent: welcome
    weight: 20250731
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2025.7.31/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2025.7.31/
---

# KubeDB v2025.7.31 (2025-08-07)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.57.0](https://github.com/kubedb/apimachinery/releases/tag/v0.57.0)

- [ed84b233](https://github.com/kubedb/apimachinery/commit/ed84b2330) Use KubeStash v2025.7.31 (#1499)
- [3837d7a2](https://github.com/kubedb/apimachinery/commit/3837d7a28) Postgres GRPC TLS Secret  (#1497)
- [d432a1eb](https://github.com/kubedb/apimachinery/commit/d432a1eb7) Add MSSQL Init Database Contants (#1494)
- [d5817eb1](https://github.com/kubedb/apimachinery/commit/d5817eb18) fix-hazelcast-ops-api (#1495)
- [c6c60155](https://github.com/kubedb/apimachinery/commit/c6c60155a) Update deps
- [6888d138](https://github.com/kubedb/apimachinery/commit/6888d138a) Setup Manifestwork Watcher (#1492)
- [73c19cfe](https://github.com/kubedb/apimachinery/commit/73c19cfe3) Add Hazelcast autoscaling webhook (#1493)
- [ccd67105](https://github.com/kubedb/apimachinery/commit/ccd67105f) Fix PVCName (#1491)
- [a1f8ed00](https://github.com/kubedb/apimachinery/commit/a1f8ed009) Add custom config field for maxscale (#1484)
- [c0c0e6b7](https://github.com/kubedb/apimachinery/commit/c0c0e6b7b) Add Clickhouse Ops-req support (#1474)
- [b67dcb13](https://github.com/kubedb/apimachinery/commit/b67dcb132) Horizontal Scaling Ops Request for Redis Hostname (#1486)
- [4b62f261](https://github.com/kubedb/apimachinery/commit/4b62f2618) add  hazelcast autoscalling crd (#1489)
- [a416271d](https://github.com/kubedb/apimachinery/commit/a416271da) Add support for Cassandra Rotate Auth (#1485)
- [2a430ba3](https://github.com/kubedb/apimachinery/commit/2a430ba32) Update ignite constants (#1490)
- [19a3bbd4](https://github.com/kubedb/apimachinery/commit/19a3bbd48) Add support for cross zone archiver restore (#1488)
- [b9a3506e](https://github.com/kubedb/apimachinery/commit/b9a3506e4) Allow mongo shards to have only one replica
- [e4feb7d0](https://github.com/kubedb/apimachinery/commit/e4feb7d0e) Test against k8s 1.33 (#1483)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.42.0](https://github.com/kubedb/autoscaler/releases/tag/v0.42.0)

- [0b7fc20b](https://github.com/kubedb/autoscaler/commit/0b7fc20b) Prepare for release v0.42.0 (#255)
- [356eb7af](https://github.com/kubedb/autoscaler/commit/356eb7af) Prepare for release v0.42.0-rc.0 (#254)
- [08d73d72](https://github.com/kubedb/autoscaler/commit/08d73d72) Add support for Cassandra (#252)
- [3edb10e3](https://github.com/kubedb/autoscaler/commit/3edb10e3) Add Autoscaler for Hazelcast (#253)
- [9dfa1fcc](https://github.com/kubedb/autoscaler/commit/9dfa1fcc) Fix crashing if refered db not found



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.10.0](https://github.com/kubedb/cassandra/releases/tag/v0.10.0)

- [25fe3b50](https://github.com/kubedb/cassandra/commit/25fe3b50) Prepare for release v0.10.0 (#42)
- [9819b1b4](https://github.com/kubedb/cassandra/commit/9819b1b4) Prepare for release v0.10.0-rc.0 (#41)
- [a5e74a21](https://github.com/kubedb/cassandra/commit/a5e74a21) Test against k8s 1.33 (#39)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.4.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.4.0)

- [0c069f8](https://github.com/kubedb/cassandra-medusa-plugin/commit/0c069f8) Prepare for release v0.4.0 (#8)
- [e6bb628](https://github.com/kubedb/cassandra-medusa-plugin/commit/e6bb628) Prepare for release v0.4.0-rc.0 (#7)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.57.0](https://github.com/kubedb/cli/releases/tag/v0.57.0)

- [a470ced8](https://github.com/kubedb/cli/commit/a470ced87) Prepare for release v0.57.0 (#800)
- [f584c273](https://github.com/kubedb/cli/commit/f584c2733) Prepare for release v0.57.0-rc.0 (#799)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.12.0](https://github.com/kubedb/clickhouse/releases/tag/v0.12.0)

- [5db4080b](https://github.com/kubedb/clickhouse/commit/5db4080b) Prepare for release v0.12.0 (#60)
- [0c3b51ae](https://github.com/kubedb/clickhouse/commit/0c3b51ae) Prepare for release v0.12.0-rc.0 (#58)
- [815c95ee](https://github.com/kubedb/clickhouse/commit/815c95ee) Add ClickHouse Ops-req (#57)
- [360fe315](https://github.com/kubedb/clickhouse/commit/360fe315) Test against k8s 1.33 (#56)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.12.0](https://github.com/kubedb/crd-manager/releases/tag/v0.12.0)

- [f6c3d8f2](https://github.com/kubedb/crd-manager/commit/f6c3d8f2) Prepare for release v0.12.0 (#85)
- [b8cb042f](https://github.com/kubedb/crd-manager/commit/b8cb042f) Prepare for release v0.12.0-rc.0 (#84)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.15.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.15.0)

- [244fa70](https://github.com/kubedb/dashboard-restic-plugin/commit/244fa70) Prepare for release v0.15.0 (#43)
- [ae4dc6b](https://github.com/kubedb/dashboard-restic-plugin/commit/ae4dc6b) Prepare for release v0.15.0-rc.0 (#42)
- [41c5309](https://github.com/kubedb/dashboard-restic-plugin/commit/41c5309) Add Automatic Restic Unlock feature (#41)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.12.0](https://github.com/kubedb/db-client-go/releases/tag/v0.12.0)

- [f20de9e7](https://github.com/kubedb/db-client-go/commit/f20de9e7) Prepare for release v0.12.0 (#191)
- [4ea2ed88](https://github.com/kubedb/db-client-go/commit/4ea2ed88) Prepare for release v0.12.0-rc.0 (#190)
- [f70a3871](https://github.com/kubedb/db-client-go/commit/f70a3871) Update ignite client (#188)
- [e73f8f69](https://github.com/kubedb/db-client-go/commit/e73f8f69) Fix es client generation from kubedb/ui-server
- [7b20b24d](https://github.com/kubedb/db-client-go/commit/7b20b24d) Fix config version for kafka sarama (#187)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.12.0](https://github.com/kubedb/druid/releases/tag/v0.12.0)

- [b5c8b0ea](https://github.com/kubedb/druid/commit/b5c8b0ea) Prepare for release v0.12.0 (#92)
- [88a643f1](https://github.com/kubedb/druid/commit/88a643f1) Prepare for release v0.12.0-rc.0 (#91)
- [3e59d167](https://github.com/kubedb/druid/commit/3e59d167) Test against k8s 1.33 (#89)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.57.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.57.0)

- [34dffed2](https://github.com/kubedb/elasticsearch/commit/34dffed2d) Prepare for release v0.57.0 (#770)
- [6a63ef51](https://github.com/kubedb/elasticsearch/commit/6a63ef51f) Prepare for release v0.57.0-rc.0 (#769)
- [d2ffd889](https://github.com/kubedb/elasticsearch/commit/d2ffd8890) Test against k8s 1.33 (#768)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.20.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.20.0)

- [b80fd437](https://github.com/kubedb/elasticsearch-restic-plugin/commit/b80fd437) Prepare for release v0.20.0 (#67)
- [df4ea2e2](https://github.com/kubedb/elasticsearch-restic-plugin/commit/df4ea2e2) Prepare for release v0.20.0-rc.0 (#66)
- [d4476789](https://github.com/kubedb/elasticsearch-restic-plugin/commit/d4476789) Add Automatic Restic Unlock Feature (#65)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.12.0](https://github.com/kubedb/ferretdb/releases/tag/v0.12.0)

- [260d686b](https://github.com/kubedb/ferretdb/commit/260d686b) Prepare for release v0.12.0 (#81)
- [e8f362c9](https://github.com/kubedb/ferretdb/commit/e8f362c9) Prepare for release v0.12.0-rc.0 (#80)
- [00808e7e](https://github.com/kubedb/ferretdb/commit/00808e7e) Test against k8s 1.33 (#79)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.5.0](https://github.com/kubedb/gitops/releases/tag/v0.5.0)

- [e51dc17a](https://github.com/kubedb/gitops/commit/e51dc17a) Prepare for release v0.5.0 (#20)
- [9868d099](https://github.com/kubedb/gitops/commit/9868d099) Prepare for release v0.5.0-rc.0 (#19)
- [26efd7e5](https://github.com/kubedb/gitops/commit/26efd7e5) Test against k8s 1.33 (#18)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.3.0](https://github.com/kubedb/hazelcast/releases/tag/v0.3.0)

- [13e3ce4f](https://github.com/kubedb/hazelcast/commit/13e3ce4f) Prepare for release v0.3.0 (#8)
- [eeea106d](https://github.com/kubedb/hazelcast/commit/eeea106d) Prepare for release v0.3.0-rc.0 (#7)
- [68cb321a](https://github.com/kubedb/hazelcast/commit/68cb321a) Merge pull request #5 from kubedb/gha-up



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.4.0](https://github.com/kubedb/ignite/releases/tag/v0.4.0)

- [71fbc350](https://github.com/kubedb/ignite/commit/71fbc350) Prepare for release v0.4.0 (#15)
- [5f380116](https://github.com/kubedb/ignite/commit/5f380116) Prepare for release v0.4.0-rc.0 (#14)
- [9d476909](https://github.com/kubedb/ignite/commit/9d476909) Fix persistence issue (#12)
- [0273ed45](https://github.com/kubedb/ignite/commit/0273ed45) Test against k8s 1.33 (#11)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2025.7.31](https://github.com/kubedb/installer/releases/tag/v2025.7.31)

- [d7a55a5e](https://github.com/kubedb/installer/commit/d7a55a5e7) Prepare for release v2025.7.31 (#1795)
- [9488bf35](https://github.com/kubedb/installer/commit/9488bf35a) Update crds for kubedb/apimachinery@ed84b233 (#1794)
- [b654c092](https://github.com/kubedb/installer/commit/b654c0922) ADD Petset GET, LIST, WATCH Permission to Webhook Server (#1793)
- [f8264e6b](https://github.com/kubedb/installer/commit/f8264e6b0) Update mssql init image for sql server init database via scripts support (#1792)
- [749bce50](https://github.com/kubedb/installer/commit/749bce509) Prepare for release v2025.7.30-rc.0 (#1790)
- [0d0e855c](https://github.com/kubedb/installer/commit/0d0e855c9) Fix expression for disk usage (#1777)
- [fbea58d3](https://github.com/kubedb/installer/commit/fbea58d3f) Update crds for kubedb/apimachinery@6888d138 (#1788)
- [3694a956](https://github.com/kubedb/installer/commit/3694a9563) Add ManifestWork ClusterRole to Provisioner, OpsManager (#1779)
- [bb4ddf15](https://github.com/kubedb/installer/commit/bb4ddf150) Re-write recommendation flags; Add deadline (#1776)
- [74f731b0](https://github.com/kubedb/installer/commit/74f731b0c) Update cve report 2025-07-31 (#1775)
- [dba697ea](https://github.com/kubedb/installer/commit/dba697ea2) Add Clickhouse Version 25.7.1 (#1778)
- [d32f771c](https://github.com/kubedb/installer/commit/d32f771c8) Add webhook configuration for hazelcast autoscaler and ops-manager (#1786)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.28.0](https://github.com/kubedb/kafka/releases/tag/v0.28.0)

- [1b7314b0](https://github.com/kubedb/kafka/commit/1b7314b0) Prepare for release v0.28.0 (#156)
- [a62d4125](https://github.com/kubedb/kafka/commit/a62d4125) Prepare for release v0.28.0-rc.0 (#155)
- [24e806ab](https://github.com/kubedb/kafka/commit/24e806ab) Test against k8s 1.33 (#154)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.33.0](https://github.com/kubedb/kibana/releases/tag/v0.33.0)

- [49817589](https://github.com/kubedb/kibana/commit/49817589) Prepare for release v0.33.0 (#155)
- [9ddffbf9](https://github.com/kubedb/kibana/commit/9ddffbf9) Prepare for release v0.33.0-rc.0 (#154)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.20.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.20.0)

- [8595ee83](https://github.com/kubedb/kubedb-manifest-plugin/commit/8595ee83) Prepare for release v0.20.0 (#99)
- [de61847a](https://github.com/kubedb/kubedb-manifest-plugin/commit/de61847a) Prepare for release v0.20.0-rc.0 (#98)
- [2a560940](https://github.com/kubedb/kubedb-manifest-plugin/commit/2a560940) Add Automatic Restic Unlock feature (#97)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.8.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.8.0)

- [4b86938](https://github.com/kubedb/kubedb-verifier/commit/4b86938) Prepare for release v0.8.0 (#22)
- [bf5a26b](https://github.com/kubedb/kubedb-verifier/commit/bf5a26b) Prepare for release v0.8.0-rc.0 (#21)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.41.0](https://github.com/kubedb/mariadb/releases/tag/v0.41.0)

- [c4380c8e](https://github.com/kubedb/mariadb/commit/c4380c8ef) Prepare for release v0.41.0 (#341)
- [9203d564](https://github.com/kubedb/mariadb/commit/9203d564e) Prepare for release v0.41.0-rc.0 (#340)
- [5dacea53](https://github.com/kubedb/mariadb/commit/5dacea536) Add Distributed MariaDB Support (#338)
- [cff8f52d](https://github.com/kubedb/mariadb/commit/cff8f52d9) Add custom config support for maxscale server (#336)
- [02a5b30f](https://github.com/kubedb/mariadb/commit/02a5b30f8) Fix Unnecessary DB Container Patch Issue (#339)
- [8d7cfeef](https://github.com/kubedb/mariadb/commit/8d7cfeef8) Fix MariaDB Local PVC BS Panic Issue (#337)
- [e6d6fe2b](https://github.com/kubedb/mariadb/commit/e6d6fe2b1) Test against k8s 1.33 (#335)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.17.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.17.0)

- [eb59c6ff](https://github.com/kubedb/mariadb-archiver/commit/eb59c6ff) Prepare for release v0.17.0 (#55)
- [af4a00ce](https://github.com/kubedb/mariadb-archiver/commit/af4a00ce) Prepare for release v0.17.0-rc.0 (#54)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.37.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.37.0)

- [73654241](https://github.com/kubedb/mariadb-coordinator/commit/73654241) Prepare for release v0.37.0 (#147)
- [bb0f9013](https://github.com/kubedb/mariadb-coordinator/commit/bb0f9013) Prepare for release v0.37.0-rc.0 (#146)
- [a1f65df7](https://github.com/kubedb/mariadb-coordinator/commit/a1f65df7) Add MultiCluster Support (#145)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.17.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.17.0)

- [21dda36d](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/21dda36d) Prepare for release v0.17.0 (#50)
- [36397d20](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/36397d20) Prepare for release v0.17.0-rc.0 (#49)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.15.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.15.0)

- [c8a1780](https://github.com/kubedb/mariadb-restic-plugin/commit/c8a1780) Prepare for release v0.15.0 (#50)
- [b178010](https://github.com/kubedb/mariadb-restic-plugin/commit/b178010) Prepare for release v0.15.0-rc.0 (#49)
- [8995cb0](https://github.com/kubedb/mariadb-restic-plugin/commit/8995cb0) Add Automatic Restic Unlock Feature (#48)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.50.0](https://github.com/kubedb/memcached/releases/tag/v0.50.0)

- [edee5b22](https://github.com/kubedb/memcached/commit/edee5b224) Prepare for release v0.50.0 (#504)
- [9c7cb565](https://github.com/kubedb/memcached/commit/9c7cb5654) Prepare for release v0.50.0-rc.0 (#503)
- [55609e30](https://github.com/kubedb/memcached/commit/55609e30e) Test against k8s 1.33 (#502)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.50.0](https://github.com/kubedb/mongodb/releases/tag/v0.50.0)

- [7fdaafad](https://github.com/kubedb/mongodb/commit/7fdaafad7) Prepare for release v0.50.0 (#712)
- [b4c09d93](https://github.com/kubedb/mongodb/commit/b4c09d93e) Prepare for release v0.50.0-rc.0 (#711)
- [26720bfc](https://github.com/kubedb/mongodb/commit/26720bfcf) Test against k8s 1.33 (#710)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.18.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.18.0)

- [582a18b3](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/582a18b3) Prepare for release v0.18.0 (#54)
- [cb811e0c](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/cb811e0c) Prepare for release v0.18.0-rc.0 (#53)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.20.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.20.0)

- [c2debfb9](https://github.com/kubedb/mongodb-restic-plugin/commit/c2debfb9) Prepare for release v0.20.0 (#88)
- [1e1dd930](https://github.com/kubedb/mongodb-restic-plugin/commit/1e1dd930) Prepare for release v0.20.0-rc.0 (#87)
- [c22d6f05](https://github.com/kubedb/mongodb-restic-plugin/commit/c22d6f05) Add Automatic Restic Unlock feature (#86)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.12.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.12.0)

- [600317db](https://github.com/kubedb/mssql-coordinator/commit/600317db) Prepare for release v0.12.0 (#41)
- [2a81caa0](https://github.com/kubedb/mssql-coordinator/commit/2a81caa0) Prepare for release v0.12.0-rc.0 (#40)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.12.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.12.0)

- [81f0538a](https://github.com/kubedb/mssqlserver/commit/81f0538a) Prepare for release v0.12.0 (#85)
- [98dd099b](https://github.com/kubedb/mssqlserver/commit/98dd099b) Add MSSQL Init Database Support via Scripts (#84)
- [7bab2549](https://github.com/kubedb/mssqlserver/commit/7bab2549) Prepare for release v0.12.0-rc.0 (#83)
- [4cd91b76](https://github.com/kubedb/mssqlserver/commit/4cd91b76) Test against k8s 1.33 (#82)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.11.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.11.0)




## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.11.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.11.0)

- [85f65fe](https://github.com/kubedb/mssqlserver-walg-plugin/commit/85f65fe) Prepare for release v0.11.0 (#29)
- [6a85654](https://github.com/kubedb/mssqlserver-walg-plugin/commit/6a85654) Prepare for release v0.11.0-rc.0 (#28)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.50.0](https://github.com/kubedb/mysql/releases/tag/v0.50.0)

- [1ce49fab](https://github.com/kubedb/mysql/commit/1ce49fabd) Prepare for release v0.50.0 (#694)
- [bb520cdf](https://github.com/kubedb/mysql/commit/bb520cdf7) Prepare for release v0.50.0-rc.0 (#693)
- [74f97c84](https://github.com/kubedb/mysql/commit/74f97c841) Fix Container Env Insertion (#691)
- [1c58c870](https://github.com/kubedb/mysql/commit/1c58c8704) Fix Unnecessary DB Container Patch Issue (#692)
- [40d65e64](https://github.com/kubedb/mysql/commit/40d65e643) Test against k8s 1.33 (#690)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.18.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.18.0)

- [7757eb7f](https://github.com/kubedb/mysql-archiver/commit/7757eb7f) Prepare for release v0.18.0 (#63)
- [b847d6d8](https://github.com/kubedb/mysql-archiver/commit/b847d6d8) Prepare for release v0.18.0-rc.0 (#62)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.35.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.35.0)

- [c07dbef0](https://github.com/kubedb/mysql-coordinator/commit/c07dbef0) Prepare for release v0.35.0 (#147)
- [255642b6](https://github.com/kubedb/mysql-coordinator/commit/255642b6) Prepare for release v0.35.0-rc.0 (#146)
- [93c3754d](https://github.com/kubedb/mysql-coordinator/commit/93c3754d) Get root password from secret instead of env variable (#145)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.18.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.18.0)

- [48e190e5](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/48e190e5) Prepare for release v0.18.0 (#50)
- [95608b82](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/95608b82) Prepare for release v0.18.0-rc.0 (#49)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.20.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.20.0)

- [039a6a8](https://github.com/kubedb/mysql-restic-plugin/commit/039a6a8) Prepare for release v0.20.0 (#78)
- [63528f1](https://github.com/kubedb/mysql-restic-plugin/commit/63528f1) Prepare for release v0.20.0-rc.0 (#77)
- [1940e43](https://github.com/kubedb/mysql-restic-plugin/commit/1940e43) Add Automatic Restic Unlock feature (#76)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.35.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.35.0)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.44.0](https://github.com/kubedb/ops-manager/releases/tag/v0.44.0)

- [a2d396e0](https://github.com/kubedb/ops-manager/commit/a2d396e00) Prepare for release v0.44.0 (#778)
- [2003a7c9](https://github.com/kubedb/ops-manager/commit/2003a7c9b) Add TLS Secret For Postgres GRPC Server and Client (#775)
- [7d57ff5a](https://github.com/kubedb/ops-manager/commit/7d57ff5a7) Horizontal Fix for Redis (#774)
- [4735de90](https://github.com/kubedb/ops-manager/commit/4735de907) Update deps
- [b7d5dc90](https://github.com/kubedb/ops-manager/commit/b7d5dc906) Prepare for release v0.44.0-rc.0 (#773)
- [1e7063de](https://github.com/kubedb/ops-manager/commit/1e7063dec) Add MariaDB MultiCluster Support (#770)
- [4b574394](https://github.com/kubedb/ops-manager/commit/4b574394a) Horizontal scaling for redis (#767)
- [b82654f5](https://github.com/kubedb/ops-manager/commit/b82654f58) update in same-version-update-recommendation (#758)
- [c006f340](https://github.com/kubedb/ops-manager/commit/c006f3405) Improve & Fix Mssql rotate auth with new secret (#772)
- [5cf86161](https://github.com/kubedb/ops-manager/commit/5cf861610) Add Ignite OpsRequest Rotate Auth reconfigureTLS & version update (#769)
- [0e8f2416](https://github.com/kubedb/ops-manager/commit/0e8f24161) Add Cassandra Rotate Auth (#761)
- [a7642f3f](https://github.com/kubedb/ops-manager/commit/a7642f3f5) Add Clickhouse Ops-Req (#763)
- [c2dffe06](https://github.com/kubedb/ops-manager/commit/c2dffe06c) Add rotate auth support for mysql and mariadb replication mode (#760)



## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.3.0](https://github.com/kubedb/oracle/releases/tag/v0.3.0)

- [dd5b4cfe](https://github.com/kubedb/oracle/commit/dd5b4cfe) Prepare for release v0.3.0 (#8)
- [0e520420](https://github.com/kubedb/oracle/commit/0e520420) Prepare for release v0.3.0-rc.0 (#7)
- [df62795d](https://github.com/kubedb/oracle/commit/df62795d) Test against k8s 1.33 (#6)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.3.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.3.0)

- [0050013](https://github.com/kubedb/oracle-coordinator/commit/0050013) Prepare for release v0.3.0 (#6)
- [bfc05a1](https://github.com/kubedb/oracle-coordinator/commit/bfc05a1) Prepare for release v0.3.0-rc.0 (#5)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.44.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.44.0)

- [6b77be9a](https://github.com/kubedb/percona-xtradb/commit/6b77be9aa) Prepare for release v0.44.0 (#412)
- [ca075aa5](https://github.com/kubedb/percona-xtradb/commit/ca075aa5a) Prepare for release v0.44.0-rc.0 (#411)
- [2ddd3843](https://github.com/kubedb/percona-xtradb/commit/2ddd3843f) Test against k8s 1.33 (#410)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.30.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.30.0)

- [c3da701e](https://github.com/kubedb/percona-xtradb-coordinator/commit/c3da701e) Prepare for release v0.30.0 (#99)
- [bcfd2a0b](https://github.com/kubedb/percona-xtradb-coordinator/commit/bcfd2a0b) Prepare for release v0.30.0-rc.0 (#98)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.41.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.41.0)

- [0fcd5662](https://github.com/kubedb/pg-coordinator/commit/0fcd5662) Prepare for release v0.41.0 (#205)
- [ae5d1d71](https://github.com/kubedb/pg-coordinator/commit/ae5d1d71) Add GRPC Support and Fix Arbiter (#204)
- [48a00fdc](https://github.com/kubedb/pg-coordinator/commit/48a00fdc) Prepare for release v0.41.0-rc.0 (#203)
- [ed3fbb81](https://github.com/kubedb/pg-coordinator/commit/ed3fbb81) Add Grpc Server and Add Support for Distributed Petset (#202)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.44.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.44.0)

- [3287bd39](https://github.com/kubedb/pgbouncer/commit/3287bd39) Prepare for release v0.44.0 (#376)
- [4dff5acb](https://github.com/kubedb/pgbouncer/commit/4dff5acb) Prepare for release v0.44.0-rc.0 (#375)
- [ed0bb5c1](https://github.com/kubedb/pgbouncer/commit/ed0bb5c1) Test against k8s 1.33 (#374)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.12.0](https://github.com/kubedb/pgpool/releases/tag/v0.12.0)

- [f8517f13](https://github.com/kubedb/pgpool/commit/f8517f13) Prepare for release v0.12.0 (#79)
- [fc054ccd](https://github.com/kubedb/pgpool/commit/fc054ccd) Prepare for release v0.12.0-rc.0 (#78)
- [58e9fafb](https://github.com/kubedb/pgpool/commit/58e9fafb) Test against k8s 1.33 (#77)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.57.0](https://github.com/kubedb/postgres/releases/tag/v0.57.0)

- [40250991](https://github.com/kubedb/postgres/commit/402509914) Prepare for release v0.57.0 (#824)
- [8183702e](https://github.com/kubedb/postgres/commit/8183702eb) Add TLS support for GRPC (#823)
- [3cb475c3](https://github.com/kubedb/postgres/commit/3cb475c30) Prepare for release v0.57.0-rc.0 (#822)
- [68119e3e](https://github.com/kubedb/postgres/commit/68119e3e3) Allow Postgres Clusterring On Multiple Kubernetes Clusters (#821)
- [ae5ac199](https://github.com/kubedb/postgres/commit/ae5ac1999) Fix in CI (add ops-manager, change cron-schedule) (#814)
- [e7622ac2](https://github.com/kubedb/postgres/commit/e7622ac26) Test against k8s 1.33 (#820)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.18.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.18.0)

- [af6f0218](https://github.com/kubedb/postgres-archiver/commit/af6f0218) Prepare for release v0.18.0 (#64)
- [54158927](https://github.com/kubedb/postgres-archiver/commit/54158927) Prepare for release v0.18.0-rc.0 (#63)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.18.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.18.0)

- [050bdf83](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/050bdf83) Prepare for release v0.18.0 (#60)
- [7d2b4455](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/7d2b4455) Prepare for release v0.18.0-rc.0 (#59)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.20.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.20.0)

- [44ec9e0](https://github.com/kubedb/postgres-restic-plugin/commit/44ec9e0) Prepare for release v0.20.0 (#75)
- [73cedc9](https://github.com/kubedb/postgres-restic-plugin/commit/73cedc9) Prepare for release v0.20.0-rc.0 (#74)
- [4a83f61](https://github.com/kubedb/postgres-restic-plugin/commit/4a83f61) Add Automatic Restic Unlock feature (#73)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.18.0](https://github.com/kubedb/provider-aws/releases/tag/v0.18.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.18.0](https://github.com/kubedb/provider-azure/releases/tag/v0.18.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.18.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.18.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.57.0](https://github.com/kubedb/provisioner/releases/tag/v0.57.0)

- [43bbc6b4](https://github.com/kubedb/provisioner/commit/43bbc6b4d) Prepare for release v0.57.0 (#160)
- [bd9cfe9c](https://github.com/kubedb/provisioner/commit/bd9cfe9ce) Update Provisioner (#159)
- [af72f8a3](https://github.com/kubedb/provisioner/commit/af72f8a39) Prepare for release v0.57.0-rc.0 (#158)
- [2610cb80](https://github.com/kubedb/provisioner/commit/2610cb806) Add Distributed Postgres and MariaDB Support (#157)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.44.0](https://github.com/kubedb/proxysql/releases/tag/v0.44.0)

- [c27243ac](https://github.com/kubedb/proxysql/commit/c27243ace) Prepare for release v0.44.0 (#398)
- [1a5396e8](https://github.com/kubedb/proxysql/commit/1a5396e8e) Prepare for release v0.44.0-rc.0 (#397)
- [2bb2d02f](https://github.com/kubedb/proxysql/commit/2bb2d02f4) Test against k8s 1.33 (#396)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.12.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.12.0)

- [3861b737](https://github.com/kubedb/rabbitmq/commit/3861b737) Prepare for release v0.12.0 (#91)
- [74164cf8](https://github.com/kubedb/rabbitmq/commit/74164cf8) Prepare for release v0.12.0-rc.0 (#90)
- [462aa28e](https://github.com/kubedb/rabbitmq/commit/462aa28e) Test against k8s 1.33 (#89)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.50.0](https://github.com/kubedb/redis/releases/tag/v0.50.0)

- [b08afa04](https://github.com/kubedb/redis/commit/b08afa04e) Prepare for release v0.50.0 (#599)
- [444f71bf](https://github.com/kubedb/redis/commit/444f71bf1) Prepare for release v0.50.0-rc.0 (#598)
- [02232f6c](https://github.com/kubedb/redis/commit/02232f6c8) Test against k8s 1.33 (#597)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.36.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.36.0)

- [a2cdae65](https://github.com/kubedb/redis-coordinator/commit/a2cdae65) Prepare for release v0.36.0 (#132)
- [c6f29217](https://github.com/kubedb/redis-coordinator/commit/c6f29217) Prepare for release v0.36.0-rc.0 (#131)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.20.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.20.0)

- [8aa9439](https://github.com/kubedb/redis-restic-plugin/commit/8aa9439) Prepare for release v0.20.0 (#70)
- [5dbf0dd](https://github.com/kubedb/redis-restic-plugin/commit/5dbf0dd) Prepare for release v0.20.0-rc.0 (#69)
- [975bb58](https://github.com/kubedb/redis-restic-plugin/commit/975bb58) Add Automatic Restic Unlock feature (#68)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.44.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.44.0)

- [c2ebe7ab](https://github.com/kubedb/replication-mode-detector/commit/c2ebe7ab) Prepare for release v0.44.0 (#296)
- [dde5e4a1](https://github.com/kubedb/replication-mode-detector/commit/dde5e4a1) Prepare for release v0.44.0-rc.0 (#295)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.33.0](https://github.com/kubedb/schema-manager/releases/tag/v0.33.0)

- [53a54d60](https://github.com/kubedb/schema-manager/commit/53a54d60) Prepare for release v0.33.0 (#142)
- [b80acd86](https://github.com/kubedb/schema-manager/commit/b80acd86) Prepare for release v0.33.0-rc.0 (#141)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.12.0](https://github.com/kubedb/singlestore/releases/tag/v0.12.0)

- [77b98a3f](https://github.com/kubedb/singlestore/commit/77b98a3f) Prepare for release v0.12.0 (#79)
- [65038345](https://github.com/kubedb/singlestore/commit/65038345) Prepare for release v0.12.0-rc.0 (#77)
- [a0bc041a](https://github.com/kubedb/singlestore/commit/a0bc041a) Test against k8s 1.33 (#75)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.12.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.12.0)

- [763497b](https://github.com/kubedb/singlestore-coordinator/commit/763497b) Prepare for release v0.12.0 (#46)
- [2822921](https://github.com/kubedb/singlestore-coordinator/commit/2822921) Prepare for release v0.12.0-rc.0 (#45)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.15.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.15.0)

- [6c108f8](https://github.com/kubedb/singlestore-restic-plugin/commit/6c108f8) Prepare for release v0.15.0 (#49)
- [1c216d8](https://github.com/kubedb/singlestore-restic-plugin/commit/1c216d8) Prepare for release v0.15.0-rc.0 (#48)
- [da671fc](https://github.com/kubedb/singlestore-restic-plugin/commit/da671fc) Test against k8s 1.33 (#47)
- [fc53bcb](https://github.com/kubedb/singlestore-restic-plugin/commit/fc53bcb) Add Automatic Restic Unlock feature (#46)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.12.0](https://github.com/kubedb/solr/releases/tag/v0.12.0)

- [99ea9c43](https://github.com/kubedb/solr/commit/99ea9c43) Prepare for release v0.12.0 (#91)
- [687f318a](https://github.com/kubedb/solr/commit/687f318a) Prepare for release v0.12.0-rc.0 (#90)
- [1a5534f9](https://github.com/kubedb/solr/commit/1a5534f9) Test against k8s 1.33 (#87)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.42.0](https://github.com/kubedb/tests/releases/tag/v0.42.0)

- [6bf4f554](https://github.com/kubedb/tests/commit/6bf4f554) Prepare for release v0.42.0 (#474)
- [426563d0](https://github.com/kubedb/tests/commit/426563d0) Prepare for release v0.42.0-rc.0 (#472)
- [9c3418da](https://github.com/kubedb/tests/commit/9c3418da) Add MySQL Archiver Test (#460)
- [9e4d0a64](https://github.com/kubedb/tests/commit/9e4d0a64) Test against k8s 1.33 (#468)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.33.0](https://github.com/kubedb/ui-server/releases/tag/v0.33.0)

- [c7426a91](https://github.com/kubedb/ui-server/commit/c7426a91) Prepare for release v0.33.0 (#169)
- [976c6624](https://github.com/kubedb/ui-server/commit/976c6624) Prepare for release v0.33.0-rc.0 (#168)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.33.0](https://github.com/kubedb/webhook-server/releases/tag/v0.33.0)

- [97f7ce19](https://github.com/kubedb/webhook-server/commit/97f7ce19) Prepare for release v0.33.0 (#163)
- [034d8712](https://github.com/kubedb/webhook-server/commit/034d8712) Add Petset Scheme (#162)
- [9ddc27d5](https://github.com/kubedb/webhook-server/commit/9ddc27d5) Prepare for release v0.33.0-rc.0 (#161)
- [38ab286a](https://github.com/kubedb/webhook-server/commit/38ab286a) setup hazelcast weebhook manager (#160)



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.6.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.6.0)

- [d0e8962](https://github.com/kubedb/xtrabackup-restic-plugin/commit/d0e8962) Prepare for release v0.6.0 (#17)
- [ec90e55](https://github.com/kubedb/xtrabackup-restic-plugin/commit/ec90e55) Prepare for release v0.6.0-rc.0 (#16)
- [281eabf](https://github.com/kubedb/xtrabackup-restic-plugin/commit/281eabf) Add Automatic Restic Unlock Feature (#15)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.12.0](https://github.com/kubedb/zookeeper/releases/tag/v0.12.0)

- [a6377217](https://github.com/kubedb/zookeeper/commit/a6377217) Prepare for release v0.12.0 (#82)
- [aa9b5b0a](https://github.com/kubedb/zookeeper/commit/aa9b5b0a) Prepare for release v0.12.0-rc.0 (#81)
- [fddd3804](https://github.com/kubedb/zookeeper/commit/fddd3804) Test against k8s 1.33 (#79)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.13.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.13.0)

- [12e5093](https://github.com/kubedb/zookeeper-restic-plugin/commit/12e5093) Prepare for release v0.13.0 (#39)
- [06daf14](https://github.com/kubedb/zookeeper-restic-plugin/commit/06daf14) Prepare for release v0.13.0-rc.0 (#38)
- [e9c8c7c](https://github.com/kubedb/zookeeper-restic-plugin/commit/e9c8c7c) Add Automatic Restic Unlock feature (#37)




