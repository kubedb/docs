---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2022.05.24
    name: Changelog-v2022.05.24
    parent: welcome
    weight: 20220524
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2022.05.24/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2022.05.24/
---

# KubeDB v2022.05.24 (2022-05-20)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.27.0](https://github.com/kubedb/apimachinery/releases/tag/v0.27.0)

- [3634eb14](https://github.com/kubedb/apimachinery/commit/3634eb14) Add `HealthCheckPaused` condition and `Unknown` phase (#898)
- [a3a3b1df](https://github.com/kubedb/apimachinery/commit/a3a3b1df) Add Raft metrics port as constants (#896)
- [6f9afd91](https://github.com/kubedb/apimachinery/commit/6f9afd91) Add support for MySQL semi-sync cluster (#890)
- [bf17bf6d](https://github.com/kubedb/apimachinery/commit/bf17bf6d) Add constants for Kibana 8 (#894)
- [a8461374](https://github.com/kubedb/apimachinery/commit/a8461374) Add method and constants for proxysql (#893)
- [a57c9577](https://github.com/kubedb/apimachinery/commit/a57c9577) Add doubleOptIn funcs & shortnames for schema-manager (#889)
- [af6f51f3](https://github.com/kubedb/apimachinery/commit/af6f51f3) Add constants and helpers for ES Internal Users (#886)
- [74c4fc13](https://github.com/kubedb/apimachinery/commit/74c4fc13) Fix typo (#888)
- [023a7988](https://github.com/kubedb/apimachinery/commit/023a7988) Update ProxySQL types and helpers (#883)
- [29217d17](https://github.com/kubedb/apimachinery/commit/29217d17) Fix pgbouncer Version Spec
- [3b994342](https://github.com/kubedb/apimachinery/commit/3b994342) Add support for mariadbdatabase with webhook (#858)
- [4be4a876](https://github.com/kubedb/apimachinery/commit/4be4a876) Add spec for MongoDB arbiter support (#862)
- [36e97b5a](https://github.com/kubedb/apimachinery/commit/36e97b5a) Add TopologySpreadConstraints (#885)
- [27c7483d](https://github.com/kubedb/apimachinery/commit/27c7483d) Add SyncStatefulSetPodDisruptionBudget helper method (#884)
- [0e635b9f](https://github.com/kubedb/apimachinery/commit/0e635b9f) Make ClusterHealth inline in ES insight (#881)
- [761d8ca3](https://github.com/kubedb/apimachinery/commit/761d8ca3) fix: update Postgres shared buffer func (#880)
- [8579cef3](https://github.com/kubedb/apimachinery/commit/8579cef3) Add Support for Opensearch Dashboards (#878)
- [24eadd87](https://github.com/kubedb/apimachinery/commit/24eadd87) Use Go 1.18 (#879)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.12.0](https://github.com/kubedb/autoscaler/releases/tag/v0.12.0)

- [7cb69fae](https://github.com/kubedb/autoscaler/commit/7cb69fae) Prepare for release v0.12.0 (#84)
- [0dd28106](https://github.com/kubedb/autoscaler/commit/0dd28106) Update dependencies (#83)
- [8fb60ad6](https://github.com/kubedb/autoscaler/commit/8fb60ad6) Update dependencies(nats client, mongo-driver) (#81)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.27.0](https://github.com/kubedb/cli/releases/tag/v0.27.0)

- [01318a20](https://github.com/kubedb/cli/commit/01318a20) Prepare for release v0.27.0 (#664)
- [6d399a31](https://github.com/kubedb/cli/commit/6d399a31) Update dependencies (#663)
- [3e9a658f](https://github.com/kubedb/cli/commit/3e9a658f) Update dependencies(nats client, mongo-driver) (#662)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.3.0](https://github.com/kubedb/dashboard/releases/tag/v0.3.0)

- [454bf6a](https://github.com/kubedb/dashboard/commit/454bf6a) Prepare for release v0.3.0 (#27)
- [872bbd9](https://github.com/kubedb/dashboard/commit/872bbd9) Update dependencies (#26)
- [6cafd62](https://github.com/kubedb/dashboard/commit/6cafd62) Add support for Kibana 8 (#25)
- [273d034](https://github.com/kubedb/dashboard/commit/273d034) Update dependencies(nats client, mongo-driver) (#24)
- [7d6c3ec](https://github.com/kubedb/dashboard/commit/7d6c3ec) Add support for Opensearch_Dashboards (#17)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.27.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.27.0)

- [ba7dee10](https://github.com/kubedb/elasticsearch/commit/ba7dee10) Prepare for release v0.27.0 (#577)
- [cfdb4d21](https://github.com/kubedb/elasticsearch/commit/cfdb4d21) Update dependencies (#576)
- [d24dfadc](https://github.com/kubedb/elasticsearch/commit/d24dfadc) Add support for ElasticStack Built-In  Users (#574)
- [cdb4d974](https://github.com/kubedb/elasticsearch/commit/cdb4d974) Update dependencies(nats client, mongo-driver) (#575)
- [865d0703](https://github.com/kubedb/elasticsearch/commit/865d0703) Add support for Elasticsearch 8 (#573)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2022.05.24](https://github.com/kubedb/installer/releases/tag/v2022.05.24)




## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.11.0](https://github.com/kubedb/mariadb/releases/tag/v0.11.0)

- [b2fd680d](https://github.com/kubedb/mariadb/commit/b2fd680d) Prepare for release v0.11.0 (#147)
- [39ac8190](https://github.com/kubedb/mariadb/commit/39ac8190) Update dependencies (#146)
- [f081a5ee](https://github.com/kubedb/mariadb/commit/f081a5ee) Update MariaDB conditions on health check (#138)
- [385f270d](https://github.com/kubedb/mariadb/commit/385f270d) Update dependencies(nats client, mongo-driver) (#145)
- [6879d6a6](https://github.com/kubedb/mariadb/commit/6879d6a6) Cleanup PodDisruptionBudget when the replica count is one or less (#144)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.7.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.7.0)

- [23da6cd](https://github.com/kubedb/mariadb-coordinator/commit/23da6cd) Prepare for release v0.7.0 (#43)
- [e1fca00](https://github.com/kubedb/mariadb-coordinator/commit/e1fca00) Update dependencies (#42)
- [20d90c6](https://github.com/kubedb/mariadb-coordinator/commit/20d90c6) Update dependencies(nats client, mongo-driver) (#41)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.20.0](https://github.com/kubedb/memcached/releases/tag/v0.20.0)

- [439a9398](https://github.com/kubedb/memcached/commit/439a9398) Prepare for release v0.20.0 (#355)
- [73606c44](https://github.com/kubedb/memcached/commit/73606c44) Update dependencies (#354)
- [75cd9209](https://github.com/kubedb/memcached/commit/75cd9209) Update dependencies(nats client, mongo-driver) (#353)
- [2b996ad8](https://github.com/kubedb/memcached/commit/2b996ad8) Cleanup PodDisruptionBudget when the replica count is one or less (#352)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.20.0](https://github.com/kubedb/mongodb/releases/tag/v0.20.0)

- [85063ec7](https://github.com/kubedb/mongodb/commit/85063ec7) Prepare for release v0.20.0 (#477)
- [ab3a33f7](https://github.com/kubedb/mongodb/commit/ab3a33f7) Update dependencies (#476)
- [275fbdc4](https://github.com/kubedb/mongodb/commit/275fbdc4) Fix shard database write check (#475)
- [643c958c](https://github.com/kubedb/mongodb/commit/643c958c) Use updated commit-hash (#474)
- [8ba58693](https://github.com/kubedb/mongodb/commit/8ba58693) Add arbiter support (#470)
- [a8ecbc33](https://github.com/kubedb/mongodb/commit/a8ecbc33) Update dependencies(nats client, mongo-driver) (#472)
- [3073bbec](https://github.com/kubedb/mongodb/commit/3073bbec) Cleanup PodDisruptionBudget when the replica count is one or less (#471)
- [e7c146cb](https://github.com/kubedb/mongodb/commit/e7c146cb) Refactor statefulset-related files (#449)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.20.0](https://github.com/kubedb/mysql/releases/tag/v0.20.0)

- [988eab76](https://github.com/kubedb/mysql/commit/988eab76) Prepare for release v0.20.0 (#470)
- [47c7e612](https://github.com/kubedb/mysql/commit/47c7e612) Update dependencies (#469)
- [a972735f](https://github.com/kubedb/mysql/commit/a972735f) Pass `--set-gtid-purged=OFF` to app binding for stash (#468)
- [a4f2e6a5](https://github.com/kubedb/mysql/commit/a4f2e6a5) Add Raft Server ports for MySQL Semi-sync (#467)
- [b9a3c322](https://github.com/kubedb/mysql/commit/b9a3c322) Add Support for Semi-sync cluster (#464)
- [2d7a0080](https://github.com/kubedb/mysql/commit/2d7a0080) Update dependencies(nats client, mongo-driver) (#466)
- [684d553a](https://github.com/kubedb/mysql/commit/684d553a) Cleanup PodDisruptionBudget when the replica count is one or less (#462)
- [5caa331a](https://github.com/kubedb/mysql/commit/5caa331a) Patch existing Auth secret to db ojbect (#463)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.5.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.5.0)

- [b30fd8e](https://github.com/kubedb/mysql-router-init/commit/b30fd8e) Update dependencies (#20)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.14.0](https://github.com/kubedb/ops-manager/releases/tag/v0.14.0)

- [2727742d](https://github.com/kubedb/ops-manager/commit/2727742d) Prepare for release v0.14.0 (#310)
- [8964e523](https://github.com/kubedb/ops-manager/commit/8964e523) Fix: Redis shard node deletion issue for Horizontal scaling (#304)
- [63ac74e1](https://github.com/kubedb/ops-manager/commit/63ac74e1) Fix product name
- [8e5a457d](https://github.com/kubedb/ops-manager/commit/8e5a457d) Rename to ops-manager package (#309)
- [36a71aa0](https://github.com/kubedb/ops-manager/commit/36a71aa0) Update dependencies (#307) (#308)
- [e0c10f1f](https://github.com/kubedb/ops-manager/commit/e0c10f1f) Update dependencies (#307)
- [d0b2d531](https://github.com/kubedb/ops-manager/commit/d0b2d531) Fix mongodb shard scale down (#306)
- [a65f70f9](https://github.com/kubedb/ops-manager/commit/a65f70f9) update replication user updating condition (#305)
- [88027506](https://github.com/kubedb/ops-manager/commit/88027506) Update Replication User Password (#300)
- [e1b525cb](https://github.com/kubedb/ops-manager/commit/e1b525cb) Use updated commit-hash (#303)
- [4dc359b4](https://github.com/kubedb/ops-manager/commit/4dc359b4) Ensure right master count when scaling down Redis Shard Cluster
- [c3bed80c](https://github.com/kubedb/ops-manager/commit/c3bed80c) Add ProxySQL TLS support (#302)
- [b8e2c085](https://github.com/kubedb/ops-manager/commit/b8e2c085) Add arbiter-support for mongodb (#291)
- [20a24475](https://github.com/kubedb/ops-manager/commit/20a24475) Update dependencies(nats client, mongo-driver) (#298)
- [0da84955](https://github.com/kubedb/ops-manager/commit/0da84955) Fix horizontal scaling to support Redis Shard Dynamic Failover (#297)
- [8d6d42a2](https://github.com/kubedb/ops-manager/commit/8d6d42a2) Add PgBouncer TLS Support (#295)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.14.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.14.0)

- [83f74c17](https://github.com/kubedb/percona-xtradb/commit/83f74c17) Prepare for release v0.14.0 (#258)
- [bfae8113](https://github.com/kubedb/percona-xtradb/commit/bfae8113) Update dependencies (#257)
- [bcc010f8](https://github.com/kubedb/percona-xtradb/commit/bcc010f8) Update dependencies(nats client, mongo-driver) (#256)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.11.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.11.0)

- [373a83e](https://github.com/kubedb/pg-coordinator/commit/373a83e) Prepare for release v0.11.0 (#80)
- [254c361](https://github.com/kubedb/pg-coordinator/commit/254c361) Update dependencies (#79)
- [7f6a6c0](https://github.com/kubedb/pg-coordinator/commit/7f6a6c0) Add Raft Metrics And graceful shutdown of Postgres (#74)
- [c1a5b53](https://github.com/kubedb/pg-coordinator/commit/c1a5b53) Update dependencies(nats client, mongo-driver) (#78)
- [b6da859](https://github.com/kubedb/pg-coordinator/commit/b6da859) Fix: Fast Shut-down Postgres server to avoid single-user mode shutdown failure (#73)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.14.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.14.0)

- [8bb55234](https://github.com/kubedb/pgbouncer/commit/8bb55234) Prepare for release v0.14.0 (#221)
- [ca8efd9a](https://github.com/kubedb/pgbouncer/commit/ca8efd9a) Update dependencies (#220)
- [8122b2c7](https://github.com/kubedb/pgbouncer/commit/8122b2c7) Update dependencies(nats client, mongo-driver) (#218)
- [431839ee](https://github.com/kubedb/pgbouncer/commit/431839ee) Update exporter container to support TLS enabled PgBouncer (#217)
- [766ece71](https://github.com/kubedb/pgbouncer/commit/766ece71) Fix TLS and Config Related Issues, Add health Check (#210)
- [76ebe1ec](https://github.com/kubedb/pgbouncer/commit/76ebe1ec) Cleanup PodDisruptionBudget when the replica count is one or less (#216)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.27.0](https://github.com/kubedb/postgres/releases/tag/v0.27.0)

- [bc3cf38e](https://github.com/kubedb/postgres/commit/bc3cf38e) Prepare for release v0.27.0 (#573)
- [14c87e8f](https://github.com/kubedb/postgres/commit/14c87e8f) Update dependencies (#572)
- [7cb31a1d](https://github.com/kubedb/postgres/commit/7cb31a1d) Add Raft Metrics exporter Port for Monitoring (#569)
- [3a71b165](https://github.com/kubedb/postgres/commit/3a71b165) Update dependencies(nats client, mongo-driver) (#571)
- [131dd7d9](https://github.com/kubedb/postgres/commit/131dd7d9) Cleanup podDiscruptionBudget when the replica count is one or less (#570)
- [44e929d8](https://github.com/kubedb/postgres/commit/44e929d8) Fix: Fast Shut-down Postgres server to avoid single-user mode shutdown failure (#568)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.27.0](https://github.com/kubedb/provisioner/releases/tag/v0.27.0)

- [1a87a7e7](https://github.com/kubedb/provisioner/commit/1a87a7e7) Prepare for release v0.27.0 (#2)
- [53226f1d](https://github.com/kubedb/provisioner/commit/53226f1d) Rename to provisioner module (#1)
- [ae8196d3](https://github.com/kubedb/provisioner/commit/ae8196d3) Update dependencies(nats client, mongo-driver) (#465)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.14.0](https://github.com/kubedb/proxysql/releases/tag/v0.14.0)

- [283f3bf3](https://github.com/kubedb/proxysql/commit/283f3bf3) Prepare for release v0.14.0 (#235)
- [05e4b5dc](https://github.com/kubedb/proxysql/commit/05e4b5dc) Update dependencies (#234)
- [81b98c09](https://github.com/kubedb/proxysql/commit/81b98c09) Fix phase and condition update for ProxySQL (#233)
- [c0561e90](https://github.com/kubedb/proxysql/commit/c0561e90) Add support for ProxySQL clustering and TLS (#231)
- [df6b4688](https://github.com/kubedb/proxysql/commit/df6b4688) Update dependencies(nats client, mongo-driver) (#232)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.20.0](https://github.com/kubedb/redis/releases/tag/v0.20.0)

- [3dcfc3c7](https://github.com/kubedb/redis/commit/3dcfc3c7) Prepare for release v0.20.0 (#397)
- [ac65b0b3](https://github.com/kubedb/redis/commit/ac65b0b3) Update dependencies (#396)
- [177c0329](https://github.com/kubedb/redis/commit/177c0329) Update dependencies(nats client, mongo-driver) (#395)
- [6bf1db27](https://github.com/kubedb/redis/commit/6bf1db27) Redis Shard Cluster Dynamic Failover (#393)
- [4fa76436](https://github.com/kubedb/redis/commit/4fa76436) Refactor StatefulSet ENVs for Redis (#394)
- [b12bfef9](https://github.com/kubedb/redis/commit/b12bfef9) Cleanup PodDisruptionBudget when the replica count is one or less (#392)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.6.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.6.0)

- [fb4f029](https://github.com/kubedb/redis-coordinator/commit/fb4f029) Prepare for release v0.6.0 (#33)
- [69cc834](https://github.com/kubedb/redis-coordinator/commit/69cc834) Update dependencies (#32)
- [9c1cbd9](https://github.com/kubedb/redis-coordinator/commit/9c1cbd9) Update dependencies(nats client, mongo-driver) (#31)
- [33baab6](https://github.com/kubedb/redis-coordinator/commit/33baab6) Update Env Variables (#30)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.14.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.14.0)

- [fcb720f2](https://github.com/kubedb/replication-mode-detector/commit/fcb720f2) Prepare for release v0.14.0 (#194)
- [b59867e3](https://github.com/kubedb/replication-mode-detector/commit/b59867e3) Update dependencies (#193)
- [bc287981](https://github.com/kubedb/replication-mode-detector/commit/bc287981) Update dependencies(nats client, mongo-driver) (#192)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.3.0](https://github.com/kubedb/schema-manager/releases/tag/v0.3.0)

- [e98aaec](https://github.com/kubedb/schema-manager/commit/e98aaec) Prepare for release v0.3.0 (#29)
- [99ca0f7](https://github.com/kubedb/schema-manager/commit/99ca0f7) Fix sharded-mongo restore issue; Use typed doubleOptIn funcs (#28)
- [2a23c38](https://github.com/kubedb/schema-manager/commit/2a23c38) Add support for MariaDB database schema manager (#24)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.12.0](https://github.com/kubedb/tests/releases/tag/v0.12.0)

- [6501852d](https://github.com/kubedb/tests/commit/6501852d) Prepare for release v0.12.0 (#178)
- [68979c56](https://github.com/kubedb/tests/commit/68979c56) Update dependencies (#177)
- [affe5f32](https://github.com/kubedb/tests/commit/affe5f32) Update dependencies(nats client, mongo-driver) (#176)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.3.0](https://github.com/kubedb/ui-server/releases/tag/v0.3.0)

- [4cb89db](https://github.com/kubedb/ui-server/commit/4cb89db) Prepare for release v0.3.0 (#34)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.3.0](https://github.com/kubedb/webhook-server/releases/tag/v0.3.0)

- [5d69aa6](https://github.com/kubedb/webhook-server/commit/5d69aa6) Prepare for release v0.3.0 (#19)
- [ca55fb8](https://github.com/kubedb/webhook-server/commit/ca55fb8) Update dependencies (#18)
- [22b4ab7](https://github.com/kubedb/webhook-server/commit/22b4ab7) Update dependencies(nats client, mongo-driver) (#17)




