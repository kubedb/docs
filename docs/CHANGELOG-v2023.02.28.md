---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2023.02.28
    name: Changelog-v2023.02.28
    parent: welcome
    weight: 20230228
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2023.02.28/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2023.02.28/
---

# KubeDB v2023.02.28 (2023-02-28)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.32.0](https://github.com/kubedb/apimachinery/releases/tag/v0.32.0)

- [40b15db3](https://github.com/kubedb/apimachinery/commit/40b15db3) Add WsrepSSTMethod Field in MariaDB API (#1012)
- [1b9e2bac](https://github.com/kubedb/apimachinery/commit/1b9e2bac) Update `setDefaults()` for pgbouncer (#1022)
- [7bf2fbe1](https://github.com/kubedb/apimachinery/commit/7bf2fbe1) Add separate Security Config Directory constant for Opensearch V2 (#1021)
- [693b7795](https://github.com/kubedb/apimachinery/commit/693b7795) Update TLS Defaulting for ProxySQL & PgBouncer (#1020)
- [48fae91c](https://github.com/kubedb/apimachinery/commit/48fae91c) Fix `GetPersistentSecrets()` function (#1018)
- [b334a5eb](https://github.com/kubedb/apimachinery/commit/b334a5eb) Add postgres streaming and standby mode in horizontal scaling (#1017)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.17.0](https://github.com/kubedb/autoscaler/releases/tag/v0.17.0)

- [4d8bd3b0](https://github.com/kubedb/autoscaler/commit/4d8bd3b0) Prepare for release v0.17.0 (#136)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.32.0](https://github.com/kubedb/cli/releases/tag/v0.32.0)

- [8477bd09](https://github.com/kubedb/cli/commit/8477bd09) Prepare for release v0.32.0 (#698)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.8.0](https://github.com/kubedb/dashboard/releases/tag/v0.8.0)

- [b0ac65a](https://github.com/kubedb/dashboard/commit/b0ac65a) Prepare for release v0.8.0 (#64)
- [987d7ef](https://github.com/kubedb/dashboard/commit/987d7ef) Add support for opensearch-dashboards 2.x (#63)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.32.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.32.0)

- [24948165](https://github.com/kubedb/elasticsearch/commit/249481659) Prepare for release v0.32.0 (#629)
- [7b6f30ed](https://github.com/kubedb/elasticsearch/commit/7b6f30edf) Use separate securityConfig Volume mount path for Opensearch V2 (#627)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2023.02.28](https://github.com/kubedb/installer/releases/tag/v2023.02.28)

- [d7d1197d](https://github.com/kubedb/installer/commit/d7d1197d) Prepare for release v2023.02.28 (#601)
- [73115439](https://github.com/kubedb/installer/commit/73115439) Update MariaDB initContainer Image with 0.5.0 (#600)
- [fa9aab3c](https://github.com/kubedb/installer/commit/fa9aab3c) Add `SecurityContext` & remove `initContainer` from `pgbouncerVersion.spec` (#596)
- [9959c608](https://github.com/kubedb/installer/commit/9959c608) Add Support for OpenSearch v2.0.1 & v2.5.0 (#599)
- [a9ca09c0](https://github.com/kubedb/installer/commit/a9ca09c0) Update postgres init conatiner and rbac for  pvc (#592)
- [2f3da948](https://github.com/kubedb/installer/commit/2f3da948) Update crds for kubedb/apimachinery@b334a5eb (#591)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.3.0](https://github.com/kubedb/kafka/releases/tag/v0.3.0)

- [be9595d](https://github.com/kubedb/kafka/commit/be9595d) Prepare for release v0.3.0 (#15)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.16.0](https://github.com/kubedb/mariadb/releases/tag/v0.16.0)

- [d092f3a3](https://github.com/kubedb/mariadb/commit/d092f3a3) Prepare for release v0.16.0 (#200)
- [60b6d846](https://github.com/kubedb/mariadb/commit/60b6d846) Add Dynamic `wsrep_sst_method` Selection Code (#193)
- [0ba15d93](https://github.com/kubedb/mariadb/commit/0ba15d93) Update sidekick dependency (#199)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.12.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.12.0)

- [ead9061](https://github.com/kubedb/mariadb-coordinator/commit/ead9061) Prepare for release v0.12.0 (#73)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.25.0](https://github.com/kubedb/memcached/releases/tag/v0.25.0)

- [a6449fc0](https://github.com/kubedb/memcached/commit/a6449fc0) Prepare for release v0.25.0 (#385)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.25.0](https://github.com/kubedb/mongodb/releases/tag/v0.25.0)

- [abfa58ea](https://github.com/kubedb/mongodb/commit/abfa58ea) Prepare for release v0.25.0 (#535)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.25.0](https://github.com/kubedb/mysql/releases/tag/v0.25.0)

- [168c7346](https://github.com/kubedb/mysql/commit/168c7346) Prepare for release v0.25.0 (#522)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.10.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.10.0)

- [570c5f4](https://github.com/kubedb/mysql-coordinator/commit/570c5f4) Prepare for release v0.10.0 (#71)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.10.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.10.0)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.19.0](https://github.com/kubedb/ops-manager/releases/tag/v0.19.0)

- [2ae340f2](https://github.com/kubedb/ops-manager/commit/2ae340f2) Prepare for release v0.19.0 (#419)
- [b91aa423](https://github.com/kubedb/ops-manager/commit/b91aa423) Fix ProxySQL reconfigure tls issues (#418)
- [3d320c64](https://github.com/kubedb/ops-manager/commit/3d320c64) Add PEM encoded output in certificate based on cert-manager feature-gate (#417)
- [817a828e](https://github.com/kubedb/ops-manager/commit/817a828e) Support Acme protocol issued certs for PgBouncer & ProxySQL (#415)
- [fe3ef59a](https://github.com/kubedb/ops-manager/commit/fe3ef59a) Add support for  stand Alone to HA postgres (#409)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.19.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.19.0)

- [97b98029](https://github.com/kubedb/percona-xtradb/commit/97b98029) Prepare for release v0.19.0 (#300)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.5.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.5.0)

- [f33b2dc](https://github.com/kubedb/percona-xtradb-coordinator/commit/f33b2dc) Prepare for release v0.5.0 (#30)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.16.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.16.0)

- [b85f50dc](https://github.com/kubedb/pg-coordinator/commit/b85f50dc) Prepare for release v0.16.0 (#113)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.19.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.19.0)

- [107f81dd](https://github.com/kubedb/pgbouncer/commit/107f81dd) Prepare for release v0.19.0 (#266)
- [7afeb055](https://github.com/kubedb/pgbouncer/commit/7afeb055) Acme TLS support (#262)
- [4abc8090](https://github.com/kubedb/pgbouncer/commit/4abc8090) Fix ownerReference for auth secret (#263)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.32.0](https://github.com/kubedb/postgres/releases/tag/v0.32.0)

- [2daf213e](https://github.com/kubedb/postgres/commit/2daf213e) Prepare for release v0.32.0 (#629)
- [5eecc8b8](https://github.com/kubedb/postgres/commit/5eecc8b8) Refactor  Reconciler to address Standalone to High Availability (#625)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.32.0](https://github.com/kubedb/provisioner/releases/tag/v0.32.0)

- [096872ab](https://github.com/kubedb/provisioner/commit/096872ab7) Prepare for release v0.32.0 (#39)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.19.0](https://github.com/kubedb/proxysql/releases/tag/v0.19.0)

- [d739df42](https://github.com/kubedb/proxysql/commit/d739df42) Prepare for release v0.19.0 (#283)
- [6392951a](https://github.com/kubedb/proxysql/commit/6392951a) Support Acme Protocol Issued Certs (eg LE) (#282)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.25.0](https://github.com/kubedb/redis/releases/tag/v0.25.0)

- [f0c6c4ef](https://github.com/kubedb/redis/commit/f0c6c4ef) Prepare for release v0.25.0 (#450)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.11.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.11.0)

- [a7605af](https://github.com/kubedb/redis-coordinator/commit/a7605af) Prepare for release v0.11.0 (#63)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.19.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.19.0)

- [450a3942](https://github.com/kubedb/replication-mode-detector/commit/450a3942) Prepare for release v0.19.0 (#225)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.8.0](https://github.com/kubedb/schema-manager/releases/tag/v0.8.0)

- [011e1f8c](https://github.com/kubedb/schema-manager/commit/011e1f8c) Prepare for release v0.8.0 (#65)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.17.0](https://github.com/kubedb/tests/releases/tag/v0.17.0)

- [b6e52b82](https://github.com/kubedb/tests/commit/b6e52b82) Prepare for release v0.17.0 (#217)
- [6ccd68ef](https://github.com/kubedb/tests/commit/6ccd68ef) Add MySQL tests (#198)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.8.0](https://github.com/kubedb/ui-server/releases/tag/v0.8.0)

- [77f1095e](https://github.com/kubedb/ui-server/commit/77f1095e) Prepare for release v0.8.0 (#68)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.8.0](https://github.com/kubedb/webhook-server/releases/tag/v0.8.0)

- [72058d49](https://github.com/kubedb/webhook-server/commit/72058d49) Prepare for release v0.8.0 (#52)




