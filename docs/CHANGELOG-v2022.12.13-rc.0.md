---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2022.12.13-rc.0
    name: Changelog-v2022.12.13-rc.0
    parent: welcome
    weight: 20221213
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2022.12.13-rc.0/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2022.12.13-rc.0/
---

# KubeDB v2022.12.13-rc.0 (2022-12-12)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.30.0-rc.0](https://github.com/kubedb/apimachinery/releases/tag/v0.30.0-rc.0)

- [70bc1ca7](https://github.com/kubedb/apimachinery/commit/70bc1ca7) Fix build
- [c051e053](https://github.com/kubedb/apimachinery/commit/c051e053) Update deps (#1007)
- [2a1d4b0b](https://github.com/kubedb/apimachinery/commit/2a1d4b0b) Set PSP in KafkaVersion Spec to optional (#1005)
- [69bc9dec](https://github.com/kubedb/apimachinery/commit/69bc9dec) Add kafka api (#998)
- [b9528283](https://github.com/kubedb/apimachinery/commit/b9528283) Run GH actions on ubuntu-20.04 (#1004)
- [d498e8e9](https://github.com/kubedb/apimachinery/commit/d498e8e9) Add ```TransferLeadershipInterval``` and ```TransferLeadershipTimeout``` for Postgres (#1001)
- [b8f88e70](https://github.com/kubedb/apimachinery/commit/b8f88e70) Add sidekick api to kubebuilder client (#1000)
- [89a71807](https://github.com/kubedb/apimachinery/commit/89a71807) Change DatabaseRef to ProxyRef in ProxySQLAutoscaler (#997)
- [f570aabe](https://github.com/kubedb/apimachinery/commit/f570aabe) Add support for ProxySQL autoscaler (#996)
- [01c07593](https://github.com/kubedb/apimachinery/commit/01c07593) Add ProxySQL Vertical-Scaling spec (#995)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.15.0-rc.0](https://github.com/kubedb/autoscaler/releases/tag/v0.15.0-rc.0)

- [2e6d15fd](https://github.com/kubedb/autoscaler/commit/2e6d15fd) Prepare for release v0.15.0-rc.0 (#126)
- [a5bc7afd](https://github.com/kubedb/autoscaler/commit/a5bc7afd) Update deps (#125)
- [56ebf3fd](https://github.com/kubedb/autoscaler/commit/56ebf3fd) Run GH actions on ubuntu-20.04 (#124)
- [ef402f45](https://github.com/kubedb/autoscaler/commit/ef402f45) Add ProxySQL autoscaler support (#121)
- [36165599](https://github.com/kubedb/autoscaler/commit/36165599) Acquire license from proxyserver (#123)
- [f727dc6e](https://github.com/kubedb/autoscaler/commit/f727dc6e) Reduce logs; Fix RecommendationProvider's parameters for sharded mongo (#122)
- [835632d9](https://github.com/kubedb/autoscaler/commit/835632d9) Clean up go.mod



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.30.0-rc.0](https://github.com/kubedb/cli/releases/tag/v0.30.0-rc.0)

- [1bf92e06](https://github.com/kubedb/cli/commit/1bf92e06) Prepare for release v0.30.0-rc.0 (#689)
- [76426575](https://github.com/kubedb/cli/commit/76426575) Update deps (#688)
- [2f35bac1](https://github.com/kubedb/cli/commit/2f35bac1) Run GH actions on ubuntu-20.04 (#687)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.6.0-rc.0](https://github.com/kubedb/dashboard/releases/tag/v0.6.0-rc.0)

- [a7952c3](https://github.com/kubedb/dashboard/commit/a7952c3) Prepare for release v0.6.0-rc.0 (#52)
- [722df43](https://github.com/kubedb/dashboard/commit/722df43) Update deps (#51)
- [600877d](https://github.com/kubedb/dashboard/commit/600877d) Run GH actions on ubuntu-20.04 (#50)
- [cc2b95b](https://github.com/kubedb/dashboard/commit/cc2b95b) Acquire license from proxyserver (#49)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.30.0-rc.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.30.0-rc.0)

- [6b883d16](https://github.com/kubedb/elasticsearch/commit/6b883d16e) Prepare for release v0.30.0-rc.0 (#617)
- [40ab6ecf](https://github.com/kubedb/elasticsearch/commit/40ab6ecf5) Update deps (#616)
- [732ba4c2](https://github.com/kubedb/elasticsearch/commit/732ba4c2f) Run GH actions on ubuntu-20.04 (#615)
- [ba032204](https://github.com/kubedb/elasticsearch/commit/ba0322041) Fix PDB deletion issue (#614)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2022.12.13-rc.0](https://github.com/kubedb/installer/releases/tag/v2022.12.13-rc.0)

- [9fca52a4](https://github.com/kubedb/installer/commit/9fca52a4) Prepare for release v2022.12.13-rc.0 (#574)
- [a1811331](https://github.com/kubedb/installer/commit/a1811331) Add support for elasticsearch 8.5.2 (#566)
- [7288df17](https://github.com/kubedb/installer/commit/7288df17) Update redis-init image (#573)
- [a9e2070d](https://github.com/kubedb/installer/commit/a9e2070d) Add kafka versions (#571)
- [9d3c3255](https://github.com/kubedb/installer/commit/9d3c3255) Update crds for kubedb/apimachinery@2a1d4b0b (#572)
- [0c3cfd8b](https://github.com/kubedb/installer/commit/0c3cfd8b) Update crds for kubedb/apimachinery@69bc9dec (#570)
- [d8cf2cfd](https://github.com/kubedb/installer/commit/d8cf2cfd) Update crds for kubedb/apimachinery@b9528283 (#569)
- [15601eeb](https://github.com/kubedb/installer/commit/15601eeb) Run GH actions on ubuntu-20.04 (#568)
- [833df418](https://github.com/kubedb/installer/commit/833df418) Add proxysql to kubedb grafana dashboard values and resources (#567)
- [bb368507](https://github.com/kubedb/installer/commit/bb368507) Add support for Postgres 15.1 12.13 13.9 14.6 (#563)
- [5c43e598](https://github.com/kubedb/installer/commit/5c43e598) Update Grafana dashboards (#564)
- [641023f5](https://github.com/kubedb/installer/commit/641023f5) Update crds for kubedb/apimachinery@89a71807 (#561)
- [be777e86](https://github.com/kubedb/installer/commit/be777e86) Update crds for kubedb/apimachinery@f570aabe (#560)
- [c0473ea7](https://github.com/kubedb/installer/commit/c0473ea7) Update crds for kubedb/apimachinery@01c07593 (#559)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.1.0-rc.0](https://github.com/kubedb/kafka/releases/tag/v0.1.0-rc.0)

- [41f3a22](https://github.com/kubedb/kafka/commit/41f3a22) Prepare for release v0.1.0-rc.0 (#4)
- [6cb7882](https://github.com/kubedb/kafka/commit/6cb7882) Refactor SetupControllers
- [f4c8eb1](https://github.com/kubedb/kafka/commit/f4c8eb1) Update deps (#3)
- [61ab7f6](https://github.com/kubedb/kafka/commit/61ab7f6) Acquire license from proxyserver (#2)
- [11f6df2](https://github.com/kubedb/kafka/commit/11f6df2) Add Operator for Kafka (#1)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.14.0-rc.0](https://github.com/kubedb/mariadb/releases/tag/v0.14.0-rc.0)

- [fbc128ad](https://github.com/kubedb/mariadb/commit/fbc128ad) Prepare for release v0.14.0-rc.0 (#188)
- [6048437a](https://github.com/kubedb/mariadb/commit/6048437a) Update deps (#187)
- [649bb98e](https://github.com/kubedb/mariadb/commit/649bb98e) Run GH actions on ubuntu-20.04 (#186)
- [b14ab86f](https://github.com/kubedb/mariadb/commit/b14ab86f) Update PDB Deletion (#185)
- [897068c5](https://github.com/kubedb/mariadb/commit/897068c5) Use constants from apimachinery (#184)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.10.0-rc.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.10.0-rc.0)

- [02c4399](https://github.com/kubedb/mariadb-coordinator/commit/02c4399) Prepare for release v0.10.0-rc.0 (#66)
- [bf28b66](https://github.com/kubedb/mariadb-coordinator/commit/bf28b66) Update deps (#65)
- [a00947d](https://github.com/kubedb/mariadb-coordinator/commit/a00947d) Run GH actions on ubuntu-20.04 (#64)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.23.0-rc.0](https://github.com/kubedb/memcached/releases/tag/v0.23.0-rc.0)

- [8f5172f6](https://github.com/kubedb/memcached/commit/8f5172f6) Prepare for release v0.23.0-rc.0 (#378)
- [cb73ec86](https://github.com/kubedb/memcached/commit/cb73ec86) Update deps (#377)
- [e8b780d6](https://github.com/kubedb/memcached/commit/e8b780d6) Run GH actions on ubuntu-20.04 (#376)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.23.0-rc.0](https://github.com/kubedb/mongodb/releases/tag/v0.23.0-rc.0)

- [2602cc08](https://github.com/kubedb/mongodb/commit/2602cc08) Prepare for release v0.23.0-rc.0 (#524)
- [a53e0b6e](https://github.com/kubedb/mongodb/commit/a53e0b6e) Update deps (#523)
- [6f68602b](https://github.com/kubedb/mongodb/commit/6f68602b) Run GH actions on ubuntu-20.04 (#522)
- [d9448103](https://github.com/kubedb/mongodb/commit/d9448103) Fix PDB issues (#521)
- [6f9b3325](https://github.com/kubedb/mongodb/commit/6f9b3325) Copy missing fields from podTemplate & serviceTemplate (#520)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.23.0-rc.0](https://github.com/kubedb/mysql/releases/tag/v0.23.0-rc.0)

- [22382a39](https://github.com/kubedb/mysql/commit/22382a39) Prepare for release v0.23.0-rc.0 (#512)
- [8e7fb1a7](https://github.com/kubedb/mysql/commit/8e7fb1a7) Update deps (#511)
- [15f8ba0b](https://github.com/kubedb/mysql/commit/15f8ba0b) Run GH actions on ubuntu-20.04 (#510)
- [83335edb](https://github.com/kubedb/mysql/commit/83335edb) Update PDB Deletion (#509)
- [b5b8cadd](https://github.com/kubedb/mysql/commit/b5b8cadd) Use constants from apimachinery (#508)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.8.0-rc.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.8.0-rc.0)

- [cc3258d](https://github.com/kubedb/mysql-coordinator/commit/cc3258d) Prepare for release v0.8.0-rc.0 (#63)
- [25da659](https://github.com/kubedb/mysql-coordinator/commit/25da659) Update deps (#62)
- [c2cd415](https://github.com/kubedb/mysql-coordinator/commit/c2cd415) Run GH actions on ubuntu-20.04 (#61)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.8.0-rc.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.8.0-rc.0)

- [a8c367e](https://github.com/kubedb/mysql-router-init/commit/a8c367e) Update deps (#28)
- [e11c7ff](https://github.com/kubedb/mysql-router-init/commit/e11c7ff) Run GH actions on ubuntu-20.04 (#27)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.17.0-rc.0](https://github.com/kubedb/ops-manager/releases/tag/v0.17.0-rc.0)

- [13107ce9](https://github.com/kubedb/ops-manager/commit/13107ce9) Prepare for release v0.17.0-rc.0 (#393)
- [96f289a0](https://github.com/kubedb/ops-manager/commit/96f289a0) Update deps (#392)
- [ab83bb02](https://github.com/kubedb/ops-manager/commit/ab83bb02) Update Evict pod with kmodules api (#388)
- [028a4a29](https://github.com/kubedb/ops-manager/commit/028a4a29) Fix condition check for pvc update (#384)
- [f85db652](https://github.com/kubedb/ops-manager/commit/f85db652) Add TLS support for Kafka (#391)
- [93e1fcf4](https://github.com/kubedb/ops-manager/commit/93e1fcf4) Fix: compareTables() function for postgresql logical replication (#385)
- [d6225c57](https://github.com/kubedb/ops-manager/commit/d6225c57) Run GH actions on ubuntu-20.04 (#390)
- [eb9f8b0c](https://github.com/kubedb/ops-manager/commit/eb9f8b0c) Remove usage of `UpgradeVersion` constant (#389)
- [f682a359](https://github.com/kubedb/ops-manager/commit/f682a359) Skip Managing TLS if DB is paused for MariaDB, PXC and ProxySQL (#387)
- [1ba7dc05](https://github.com/kubedb/ops-manager/commit/1ba7dc05) Add ProxySQL Vertical Scaling Ops-Request (#381)
- [db89b9c9](https://github.com/kubedb/ops-manager/commit/db89b9c9) Adding `UpdateVersion` in mongo validator (#382)
- [7c373593](https://github.com/kubedb/ops-manager/commit/7c373593) Acquire license from proxyserver (#383)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.17.0-rc.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.17.0-rc.0)

- [f7ba9bfc](https://github.com/kubedb/percona-xtradb/commit/f7ba9bfc) Prepare for release v0.17.0-rc.0 (#290)
- [806df3d2](https://github.com/kubedb/percona-xtradb/commit/806df3d2) Update deps (#289)
- [a55bb0f2](https://github.com/kubedb/percona-xtradb/commit/a55bb0f2) Run GH actions on ubuntu-20.04 (#288)
- [37fab686](https://github.com/kubedb/percona-xtradb/commit/37fab686) Update PDB Deletion (#287)
- [55c35a72](https://github.com/kubedb/percona-xtradb/commit/55c35a72) Use constants from apimachinery (#286)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.3.0-rc.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.3.0-rc.0)

- [7e53d31](https://github.com/kubedb/percona-xtradb-coordinator/commit/7e53d31) Prepare for release v0.3.0-rc.0 (#23)
- [bd5e0b3](https://github.com/kubedb/percona-xtradb-coordinator/commit/bd5e0b3) Update deps (#22)
- [b970f14](https://github.com/kubedb/percona-xtradb-coordinator/commit/b970f14) Run GH actions on ubuntu-20.04 (#21)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.14.0-rc.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.14.0-rc.0)

- [34cb5a6c](https://github.com/kubedb/pg-coordinator/commit/34cb5a6c) Prepare for release v0.14.0-rc.0 (#105)
- [7394e6b7](https://github.com/kubedb/pg-coordinator/commit/7394e6b7) Update deps (#104)
- [228b1ae2](https://github.com/kubedb/pg-coordinator/commit/228b1ae2) Merge pull request #102 from kubedb/leader-switch
- [11a3c127](https://github.com/kubedb/pg-coordinator/commit/11a3c127) Merge branch 'master' into leader-switch
- [f8d04c52](https://github.com/kubedb/pg-coordinator/commit/f8d04c52) Add PG Reset Wal for Single user mode failed #101
- [8eaa5f11](https://github.com/kubedb/pg-coordinator/commit/8eaa5f11) retry eviction of pod and delete pod if fails
- [d2a23fa9](https://github.com/kubedb/pg-coordinator/commit/d2a23fa9) Update deps
- [febd8aab](https://github.com/kubedb/pg-coordinator/commit/febd8aab) Refined
- [5a2005cf](https://github.com/kubedb/pg-coordinator/commit/5a2005cf) Fix: Transfer Leadership issue fix with pod delete
- [7631cb84](https://github.com/kubedb/pg-coordinator/commit/7631cb84) Add PG Reset Wal for Single user mode failed
- [a951c00e](https://github.com/kubedb/pg-coordinator/commit/a951c00e) Run GH actions on ubuntu-20.04 (#103)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.17.0-rc.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.17.0-rc.0)

- [8d39e418](https://github.com/kubedb/pgbouncer/commit/8d39e418) Prepare for release v0.17.0-rc.0 (#251)
- [991cbaec](https://github.com/kubedb/pgbouncer/commit/991cbaec) Update deps (#250)
- [8af0a2f0](https://github.com/kubedb/pgbouncer/commit/8af0a2f0) Run GH actions on ubuntu-20.04 (#248)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.30.0-rc.0](https://github.com/kubedb/postgres/releases/tag/v0.30.0-rc.0)

- [da9e88bb](https://github.com/kubedb/postgres/commit/da9e88bb) Prepare for release v0.30.0-rc.0 (#615)
- [f2e2da36](https://github.com/kubedb/postgres/commit/f2e2da36) Update deps (#614)
- [296bb241](https://github.com/kubedb/postgres/commit/296bb241) Run GH actions on ubuntu-20.04 (#613)
- [d67b529a](https://github.com/kubedb/postgres/commit/d67b529a) Add tranferLeadership env for co-ordinator (#612)
- [fab00b44](https://github.com/kubedb/postgres/commit/fab00b44) Update PDB Deletion (#611)
- [c104c2b2](https://github.com/kubedb/postgres/commit/c104c2b2) Check for old auth secret label (#610)
- [932d6851](https://github.com/kubedb/postgres/commit/932d6851) Fix shared buffer for version 10 (#609)
- [60dba4ae](https://github.com/kubedb/postgres/commit/60dba4ae) Use constants from apimachinery (#608)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.30.0-rc.0](https://github.com/kubedb/provisioner/releases/tag/v0.30.0-rc.0)

- [1104e9f6](https://github.com/kubedb/provisioner/commit/1104e9f68) Prepare for release v0.30.0-rc.0 (#28)
- [f37503db](https://github.com/kubedb/provisioner/commit/f37503dbb) Add kafka controller (#27)
- [c8618da0](https://github.com/kubedb/provisioner/commit/c8618da0b) Update deps (#26)
- [2db07a7d](https://github.com/kubedb/provisioner/commit/2db07a7dc) Run GH actions on ubuntu-20.04 (#25)
- [9949d569](https://github.com/kubedb/provisioner/commit/9949d5692) Acquire license from proxyserver (#24)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.17.0-rc.0](https://github.com/kubedb/proxysql/releases/tag/v0.17.0-rc.0)

- [587d8b97](https://github.com/kubedb/proxysql/commit/587d8b97) Prepare for release v0.17.0-rc.0 (#267)
- [32b9cc71](https://github.com/kubedb/proxysql/commit/32b9cc71) Update deps (#266)
- [05e7a3a4](https://github.com/kubedb/proxysql/commit/05e7a3a4) Add MariaDB and Percona-XtraDB Backend (#264)
- [a1e7c91d](https://github.com/kubedb/proxysql/commit/a1e7c91d) Fix CI workflow for private deps
- [effb7617](https://github.com/kubedb/proxysql/commit/effb7617) Run GH actions on ubuntu-20.04 (#265)
- [38391814](https://github.com/kubedb/proxysql/commit/38391814) Use constants from apimachinery (#263)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.23.0-rc.0](https://github.com/kubedb/redis/releases/tag/v0.23.0-rc.0)

- [175547fa](https://github.com/kubedb/redis/commit/175547fa) Prepare for release v0.23.0-rc.0 (#438)
- [265332d0](https://github.com/kubedb/redis/commit/265332d0) Update deps (#437)
- [f1a8f85f](https://github.com/kubedb/redis/commit/f1a8f85f) Run GH actions on ubuntu-20.04 (#436)
- [9263f404](https://github.com/kubedb/redis/commit/9263f404) Fix PDB deletion issue (#435)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.9.0-rc.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.9.0-rc.0)

- [61aefbb](https://github.com/kubedb/redis-coordinator/commit/61aefbb) Prepare for release v0.9.0-rc.0 (#56)
- [94a6eea](https://github.com/kubedb/redis-coordinator/commit/94a6eea) Update deps (#55)
- [4454cf1](https://github.com/kubedb/redis-coordinator/commit/4454cf1) Run GH actions on ubuntu-20.04 (#54)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.17.0-rc.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.17.0-rc.0)

- [865f05e0](https://github.com/kubedb/replication-mode-detector/commit/865f05e0) Prepare for release v0.17.0-rc.0 (#218)
- [8d0fa119](https://github.com/kubedb/replication-mode-detector/commit/8d0fa119) Update deps (#217)
- [e6a86096](https://github.com/kubedb/replication-mode-detector/commit/e6a86096) Run GH actions on ubuntu-20.04 (#216)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.6.0-rc.0](https://github.com/kubedb/schema-manager/releases/tag/v0.6.0-rc.0)

- [64bf4d7a](https://github.com/kubedb/schema-manager/commit/64bf4d7a) Prepare for release v0.6.0-rc.0 (#56)
- [c0bd9699](https://github.com/kubedb/schema-manager/commit/c0bd9699) Update deps (#55)
- [ab5098c9](https://github.com/kubedb/schema-manager/commit/ab5098c9) Run GH actions on ubuntu-20.04 (#54)
- [3a7c5fb9](https://github.com/kubedb/schema-manager/commit/3a7c5fb9) Acquire license from proxyserver (#53)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.15.0-rc.0](https://github.com/kubedb/tests/releases/tag/v0.15.0-rc.0)

- [d212a7d2](https://github.com/kubedb/tests/commit/d212a7d2) Prepare for release v0.15.0-rc.0 (#208)
- [1c9c1627](https://github.com/kubedb/tests/commit/1c9c1627) Update deps (#207)
- [b3bfac83](https://github.com/kubedb/tests/commit/b3bfac83) Run GH actions on ubuntu-20.04 (#206)
- [986dd480](https://github.com/kubedb/tests/commit/986dd480) Add Redis Sentinel e2e Tests (#199)
- [5c2fc0b9](https://github.com/kubedb/tests/commit/5c2fc0b9) Update MongoDB Autoscaler tests (#204)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.6.0-rc.0](https://github.com/kubedb/ui-server/releases/tag/v0.6.0-rc.0)

- [8e1be757](https://github.com/kubedb/ui-server/commit/8e1be757) Prepare for release v0.6.0-rc.0 (#59)
- [05f138aa](https://github.com/kubedb/ui-server/commit/05f138aa) Update deps (#58)
- [87c75073](https://github.com/kubedb/ui-server/commit/87c75073) Run GH actions on ubuntu-20.04 (#56)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.6.0-rc.0](https://github.com/kubedb/webhook-server/releases/tag/v0.6.0-rc.0)

- [2df0f44e](https://github.com/kubedb/webhook-server/commit/2df0f44e) Prepare for release v0.6.0-rc.0 (#41)
- [f1ea74a2](https://github.com/kubedb/webhook-server/commit/f1ea74a2) Add kafka webhooks (#39)
- [b15ff051](https://github.com/kubedb/webhook-server/commit/b15ff051) Update deps (#40)
- [6246a9cf](https://github.com/kubedb/webhook-server/commit/6246a9cf) Run GH actions on ubuntu-20.04 (#38)




