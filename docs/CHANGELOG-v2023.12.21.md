---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2023.12.21
    name: Changelog-v2023.12.21
    parent: welcome
    weight: 20231221
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2023.12.21/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2023.12.21/
---

# KubeDB v2023.12.21 (2023-12-21)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.39.0](https://github.com/kubedb/apimachinery/releases/tag/v0.39.0)

- [c99d3ab1](https://github.com/kubedb/apimachinery/commit/c99d3ab1) Update pg arbiter api (#1091)
- [1d455662](https://github.com/kubedb/apimachinery/commit/1d455662) Add nodeSelector, tolerations in es & kafka spec (#1089)
- [3878d59f](https://github.com/kubedb/apimachinery/commit/3878d59f) Update deps
- [ecc6001f](https://github.com/kubedb/apimachinery/commit/ecc6001f) Update deps
- [bf7aa205](https://github.com/kubedb/apimachinery/commit/bf7aa205) Configure node topology for autoscaling compute resources (#1085)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.24.0](https://github.com/kubedb/autoscaler/releases/tag/v0.24.0)

- [f2e9be5d](https://github.com/kubedb/autoscaler/commit/f2e9be5d) Prepare for release v0.24.0 (#166)
- [98f7ce9e](https://github.com/kubedb/autoscaler/commit/98f7ce9e) Utilize topologyInfo in compute autoscalers (#162)
- [fc29fd8a](https://github.com/kubedb/autoscaler/commit/fc29fd8a) Send hourly audit events (#165)
- [ee3e323f](https://github.com/kubedb/autoscaler/commit/ee3e323f) Update autoscaler & ops apis (#161)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.39.0](https://github.com/kubedb/cli/releases/tag/v0.39.0)

- [3d619254](https://github.com/kubedb/cli/commit/3d619254) Prepare for release v0.39.0 (#741)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.15.0](https://github.com/kubedb/dashboard/releases/tag/v0.15.0)

- [1f0ffd6f](https://github.com/kubedb/dashboard/commit/1f0ffd6f) Prepare for release v0.15.0 (#89)
- [0d69c977](https://github.com/kubedb/dashboard/commit/0d69c977) Send hourly audit events (#88)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.39.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.39.0)

- [944aac8b](https://github.com/kubedb/elasticsearch/commit/944aac8bf) Prepare for release v0.39.0 (#686)
- [090de217](https://github.com/kubedb/elasticsearch/commit/090de2176) Send hourly audit events (#685)
- [3f18a9e4](https://github.com/kubedb/elasticsearch/commit/3f18a9e40) Set tolerations & nodeSelectors from esNode (#682)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.2.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.2.0)

- [89c9f39](https://github.com/kubedb/elasticsearch-restic-plugin/commit/89c9f39) Prepare for release v0.2.0 (#12)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2023.12.21](https://github.com/kubedb/installer/releases/tag/v2023.12.21)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.10.0](https://github.com/kubedb/kafka/releases/tag/v0.10.0)

- [89c3fbe](https://github.com/kubedb/kafka/commit/89c3fbe) Prepare for release v0.10.0 (#55)
- [ca738d8](https://github.com/kubedb/kafka/commit/ca738d8) Send hourly audit events (#54)
- [cfd9ea2](https://github.com/kubedb/kafka/commit/cfd9ea2) Set tolerations & nodeSelectors from kafka topology nodes (#53)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.2.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.2.0)

- [e561ae8](https://github.com/kubedb/kubedb-manifest-plugin/commit/e561ae8) Prepare for release v0.2.0 (#32)
- [86311ba](https://github.com/kubedb/kubedb-manifest-plugin/commit/86311ba) Add mysql and mariadb manifest backup and restore support (#31)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.23.0](https://github.com/kubedb/mariadb/releases/tag/v0.23.0)

- [e6cae3c7](https://github.com/kubedb/mariadb/commit/e6cae3c7) Prepare for release v0.23.0 (#240)
- [b0c9a5a9](https://github.com/kubedb/mariadb/commit/b0c9a5a9) Send hourly audit events (#239)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.2.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.2.0)

- [1c1bb1d](https://github.com/kubedb/mariadb-archiver/commit/1c1bb1d) Prepare for release v0.2.0 (#7)
- [e1ada03](https://github.com/kubedb/mariadb-archiver/commit/e1ada03) Use appscode-images as base image (#6)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.19.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.19.0)

- [a82c76e8](https://github.com/kubedb/mariadb-coordinator/commit/a82c76e8) Prepare for release v0.19.0 (#95)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.32.0](https://github.com/kubedb/memcached/releases/tag/v0.32.0)

- [28a0d9b6](https://github.com/kubedb/memcached/commit/28a0d9b6) Prepare for release v0.32.0 (#410)
- [5b0e2cf7](https://github.com/kubedb/memcached/commit/5b0e2cf7) Send hourly audit events (#409)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.32.0](https://github.com/kubedb/mongodb/releases/tag/v0.32.0)

- [6b7b6be2](https://github.com/kubedb/mongodb/commit/6b7b6be2) Prepare for release v0.32.0 (#589)
- [7c9d0105](https://github.com/kubedb/mongodb/commit/7c9d0105) Send hourly audit events (#588)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.2.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.2.0)

- [2bad72d](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/2bad72d) Prepare for release v0.2.0 (#6)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.2.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.2.0)

- [16cdbac](https://github.com/kubedb/mongodb-restic-plugin/commit/16cdbac) Prepare for release v0.2.0 (#17)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.32.0](https://github.com/kubedb/mysql/releases/tag/v0.32.0)

- [1d875c1c](https://github.com/kubedb/mysql/commit/1d875c1c) Prepare for release v0.32.0 (#581)
- [d4323211](https://github.com/kubedb/mysql/commit/d4323211) Send hourly audit events (#580)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.2.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.2.0)

- [e800623](https://github.com/kubedb/mysql-archiver/commit/e800623) Prepare for release v0.2.0 (#8)
- [b9f6ec5](https://github.com/kubedb/mysql-archiver/commit/b9f6ec5) Install mysqlbinlog (#7)
- [c46d991](https://github.com/kubedb/mysql-archiver/commit/c46d991) Use appscode-images as base image (#6)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.17.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.17.0)

- [eb942605](https://github.com/kubedb/mysql-coordinator/commit/eb942605) Prepare for release v0.17.0 (#92)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.2.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.2.0)

- [91eb451](https://github.com/kubedb/mysql-restic-plugin/commit/91eb451) Prepare for release v0.2.0 (#16)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.17.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.17.0)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.26.0](https://github.com/kubedb/ops-manager/releases/tag/v0.26.0)

- [328b13d9](https://github.com/kubedb/ops-manager/commit/328b13d9) Prepare for release v0.26.0 (#503)
- [10100aa9](https://github.com/kubedb/ops-manager/commit/10100aa9) Set tolerations & nodeSelectors while verticalScaling (#500)
- [14fe79e3](https://github.com/kubedb/ops-manager/commit/14fe79e3) Send hourly audit events (#502)
- [b7a5522f](https://github.com/kubedb/ops-manager/commit/b7a5522f) Update opsRequest api (#499)
- [304855b3](https://github.com/kubedb/ops-manager/commit/304855b3) Update daily-opensearch.yml
- [50c7ff53](https://github.com/kubedb/ops-manager/commit/50c7ff53) Update daily workflow for ES and Kafka. (#493)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.26.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.26.0)

- [35495dd3](https://github.com/kubedb/percona-xtradb/commit/35495dd3) Prepare for release v0.26.0 (#338)
- [7bac5129](https://github.com/kubedb/percona-xtradb/commit/7bac5129) Send hourly audit events (#337)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.12.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.12.0)

- [1dc1fbf](https://github.com/kubedb/percona-xtradb-coordinator/commit/1dc1fbf) Prepare for release v0.12.0 (#52)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.23.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.23.0)

- [d18365e6](https://github.com/kubedb/pg-coordinator/commit/d18365e6) Prepare for release v0.23.0 (#142)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.26.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.26.0)

- [ad28cfa4](https://github.com/kubedb/pgbouncer/commit/ad28cfa4) Prepare for release v0.26.0 (#303)
- [dbe23148](https://github.com/kubedb/pgbouncer/commit/dbe23148) Send hourly audit events (#302)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.39.0](https://github.com/kubedb/postgres/releases/tag/v0.39.0)

- [448e81f0](https://github.com/kubedb/postgres/commit/448e81f04) Prepare for release v0.39.0 (#694)
- [745c6555](https://github.com/kubedb/postgres/commit/745c6555d) Send hourly audit events (#693)
- [e4016868](https://github.com/kubedb/postgres/commit/e4016868e) Send hourly audit events (#691)
- [26f68fef](https://github.com/kubedb/postgres/commit/26f68fefa) Set toleration & nodeSelector fields from arbiter spec (#689)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.2.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.2.0)

- [c4f7e11](https://github.com/kubedb/postgres-archiver/commit/c4f7e11) Fix formatting



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.2.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.2.0)

- [bce9779](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/bce9779) Prepare for release v0.2.0 (#9)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.2.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.2.0)

- [7e449e3](https://github.com/kubedb/postgres-restic-plugin/commit/7e449e3) Prepare for release v0.2.0 (#9)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.1.0](https://github.com/kubedb/provider-aws/releases/tag/v0.1.0)

- [3cdbabe](https://github.com/kubedb/provider-aws/commit/3cdbabe) Fix makefile



## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.1.0](https://github.com/kubedb/provider-azure/releases/tag/v0.1.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.1.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.1.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.39.0](https://github.com/kubedb/provisioner/releases/tag/v0.39.0)

- [6ec88b2b](https://github.com/kubedb/provisioner/commit/6ec88b2b0) Prepare for release v0.39.0 (#65)
- [bbb9417d](https://github.com/kubedb/provisioner/commit/bbb9417da) Send hourly audit events (#64)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.26.0](https://github.com/kubedb/proxysql/releases/tag/v0.26.0)

- [71c51c63](https://github.com/kubedb/proxysql/commit/71c51c63) Prepare for release v0.26.0 (#317)
- [30119f2c](https://github.com/kubedb/proxysql/commit/30119f2c) Send hourly audit events (#316)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.32.0](https://github.com/kubedb/redis/releases/tag/v0.32.0)

- [c18c7bbf](https://github.com/kubedb/redis/commit/c18c7bbf) Prepare for release v0.32.0 (#504)
- [8716c93c](https://github.com/kubedb/redis/commit/8716c93c) Send hourly audit events (#503)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.18.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.18.0)

- [a5ddc00b](https://github.com/kubedb/redis-coordinator/commit/a5ddc00b) Prepare for release v0.18.0 (#83)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.2.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.2.0)

- [352a231](https://github.com/kubedb/redis-restic-plugin/commit/352a231) Prepare for release v0.2.0 (#13)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.26.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.26.0)

- [9fbf8da6](https://github.com/kubedb/replication-mode-detector/commit/9fbf8da6) Prepare for release v0.26.0 (#247)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.15.0](https://github.com/kubedb/schema-manager/releases/tag/v0.15.0)

- [bb65f133](https://github.com/kubedb/schema-manager/commit/bb65f133) Prepare for release v0.15.0 (#90)
- [96ddacfe](https://github.com/kubedb/schema-manager/commit/96ddacfe) Send hourly audit events (#89)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.24.0](https://github.com/kubedb/tests/releases/tag/v0.24.0)

- [7bd88b9f](https://github.com/kubedb/tests/commit/7bd88b9f) Prepare for release v0.24.0 (#275)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.15.0](https://github.com/kubedb/ui-server/releases/tag/v0.15.0)

- [7b2351b0](https://github.com/kubedb/ui-server/commit/7b2351b0) Prepare for release v0.15.0 (#99)
- [956aae83](https://github.com/kubedb/ui-server/commit/956aae83) Send hourly audit events (#98)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.15.0](https://github.com/kubedb/webhook-server/releases/tag/v0.15.0)

- [3afa1398](https://github.com/kubedb/webhook-server/commit/3afa1398) Prepare for release v0.15.0 (#76)
- [96da1acd](https://github.com/kubedb/webhook-server/commit/96da1acd) Send hourly audit events (#75)




