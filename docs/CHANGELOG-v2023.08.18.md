---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2023.08.18
    name: Changelog-v2023.08.18
    parent: welcome
    weight: 20230818
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2023.08.18/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2023.08.18/
---

# KubeDB v2023.08.18 (2023-08-21)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.35.0](https://github.com/kubedb/apimachinery/releases/tag/v0.35.0)

- [8e2aab0c](https://github.com/kubedb/apimachinery/commit/8e2aab0c) Add apis for git sync (#1055)
- [eec599f0](https://github.com/kubedb/apimachinery/commit/eec599f0) Add default MaxUnavailable spec for ES (#1053)
- [72a039cd](https://github.com/kubedb/apimachinery/commit/72a039cd) Updated kafka validateVersions webhook for newly added versions (#1054)
- [f86084cd](https://github.com/kubedb/apimachinery/commit/f86084cd) Add Logical Replication Replica Identity conditions (#1050)
- [4ce21bc5](https://github.com/kubedb/apimachinery/commit/4ce21bc5) Add cruise control API (#1045)
- [eda8efdf](https://github.com/kubedb/apimachinery/commit/eda8efdf) Make the conditions uniform across database opsRequests (#1052)
- [eb1b7f21](https://github.com/kubedb/apimachinery/commit/eb1b7f21) Change schema-manager constants type (#1051)
- [a763fb6b](https://github.com/kubedb/apimachinery/commit/a763fb6b) Use updated kmapi Conditions (#1049)
- [ebc00ae2](https://github.com/kubedb/apimachinery/commit/ebc00ae2) Add AsOwner() utility for dbs (#1046)
- [224dd567](https://github.com/kubedb/apimachinery/commit/224dd567) Add Custom Configuration spec for Kafka (#1041)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.20.0](https://github.com/kubedb/autoscaler/releases/tag/v0.20.0)

- [02970fe1](https://github.com/kubedb/autoscaler/commit/02970fe1) Prepare for release v0.20.0 (#152)
- [bdd60f13](https://github.com/kubedb/autoscaler/commit/bdd60f13) Update dependencies (#151)
- [9fbd8bf6](https://github.com/kubedb/autoscaler/commit/9fbd8bf6) Update dependencies (#150)
- [8bc4f455](https://github.com/kubedb/autoscaler/commit/8bc4f455) Use new kmapi Condition (#149)
- [6ccd8cfc](https://github.com/kubedb/autoscaler/commit/6ccd8cfc) Update Makefile
- [23a3a0b1](https://github.com/kubedb/autoscaler/commit/23a3a0b1) Use restricted pod security label (#148)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.35.0](https://github.com/kubedb/cli/releases/tag/v0.35.0)

- [bbe4b2ef](https://github.com/kubedb/cli/commit/bbe4b2ef) Prepare for release v0.35.0 (#718)
- [be1a1198](https://github.com/kubedb/cli/commit/be1a1198) Update dependencies (#717)
- [6adaa37f](https://github.com/kubedb/cli/commit/6adaa37f) Add MongoDB data cli (#716)
- [95ef1341](https://github.com/kubedb/cli/commit/95ef1341) Add Elasticsearch CMD to insert, verify and drop data (#714)
- [196a75ca](https://github.com/kubedb/cli/commit/196a75ca) Added Postgres Data insert verify drop through Kubedb CLI (#712)
- [9953efb7](https://github.com/kubedb/cli/commit/9953efb7) Add Insert Verify Drop for MariaDB in KubeDB CLI (#715)
- [41139e49](https://github.com/kubedb/cli/commit/41139e49) Add Insert Verify Drop for MySQL in KubeDB CLI (#713)
- [cf49e9aa](https://github.com/kubedb/cli/commit/cf49e9aa) Add Redis CMD for data insert (#709)
- [3a14bd72](https://github.com/kubedb/cli/commit/3a14bd72) Use svcName in exec instead of static primary (#711)
- [af0c5734](https://github.com/kubedb/cli/commit/af0c5734) Update dependencies (#710)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.11.0](https://github.com/kubedb/dashboard/releases/tag/v0.11.0)

- [40e20d8](https://github.com/kubedb/dashboard/commit/40e20d8) Prepare for release v0.11.0 (#81)
- [9b390b6](https://github.com/kubedb/dashboard/commit/9b390b6) Update dependencies (#80)
- [2db4453](https://github.com/kubedb/dashboard/commit/2db4453) Update dependencies (#79)
- [49332b4](https://github.com/kubedb/dashboard/commit/49332b4) Update client-go for GET and PATCH call issue fix (#77)
- [711eba9](https://github.com/kubedb/dashboard/commit/711eba9) Use new kmapi Condition (#78)
- [4d8f1d9](https://github.com/kubedb/dashboard/commit/4d8f1d9) Update Makefile
- [cf67a4d](https://github.com/kubedb/dashboard/commit/cf67a4d) Use restricted pod security label (#76)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.35.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.35.0)

- [bc8df643](https://github.com/kubedb/elasticsearch/commit/bc8df6435) Prepare for release v0.35.0 (#662)
- [e70d2d10](https://github.com/kubedb/elasticsearch/commit/e70d2d108) Update dependencies (#661)
- [8775e15d](https://github.com/kubedb/elasticsearch/commit/8775e15d0) Confirm the db has been paused before ops continue (#660)
- [b717a900](https://github.com/kubedb/elasticsearch/commit/b717a900c) Update nightly
- [096538e8](https://github.com/kubedb/elasticsearch/commit/096538e8d) Update dependencies (#659)
- [9b5de295](https://github.com/kubedb/elasticsearch/commit/9b5de295f) update nightly test profile to provisioner (#658)
- [c25227e3](https://github.com/kubedb/elasticsearch/commit/c25227e39) Add opensearch-2.5.0 in nightly tests
- [36705273](https://github.com/kubedb/elasticsearch/commit/367052731) Fix Disable Security failing Builtin User cred synchronization Issue (#654)
- [7102b622](https://github.com/kubedb/elasticsearch/commit/7102b622d) Add inputs to nightly workflow
- [93eda557](https://github.com/kubedb/elasticsearch/commit/93eda557e) Fix GET and PATCH call issue (#648)
- [af7e4c23](https://github.com/kubedb/elasticsearch/commit/af7e4c237) Fix nightly (#651)
- [0e1de49b](https://github.com/kubedb/elasticsearch/commit/0e1de49b5) Use KIND v0.20.0 (#652)
- [ffc5d7f6](https://github.com/kubedb/elasticsearch/commit/ffc5d7f6d) Use master branch with nightly.yml
- [ec391b1e](https://github.com/kubedb/elasticsearch/commit/ec391b1e2) Update nightly.yml
- [179aa150](https://github.com/kubedb/elasticsearch/commit/179aa1507) Update nightly test matrix
- [b6a094db](https://github.com/kubedb/elasticsearch/commit/b6a094db2) Run e2e tests nightly (#650)
- [e09e1e70](https://github.com/kubedb/elasticsearch/commit/e09e1e700) Use new kmapi Condition (#649)
- [ce0cac63](https://github.com/kubedb/elasticsearch/commit/ce0cac63a) Update Makefile
- [f0b570d4](https://github.com/kubedb/elasticsearch/commit/f0b570d4d) Use restricted pod security label (#647)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2023.08.18](https://github.com/kubedb/installer/releases/tag/v2023.08.18)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.6.0](https://github.com/kubedb/kafka/releases/tag/v0.6.0)

- [1b83b3b](https://github.com/kubedb/kafka/commit/1b83b3b) Prepare for release v0.6.0 (#35)
- [4eb30cd](https://github.com/kubedb/kafka/commit/4eb30cd) Add Support for Cruise Control (#33)
- [1414470](https://github.com/kubedb/kafka/commit/1414470) Add custom configuration (#28)
- [5a5537b](https://github.com/kubedb/kafka/commit/5a5537b) Run nightly tests against master
- [d973665](https://github.com/kubedb/kafka/commit/d973665) Update nightly.yml
- [1cbdccd](https://github.com/kubedb/kafka/commit/1cbdccd) Run e2e tests nightly (#34)
- [987235c](https://github.com/kubedb/kafka/commit/987235c) Use new kmapi Condition (#32)
- [dbbbc7f](https://github.com/kubedb/kafka/commit/dbbbc7f) Update Makefile
- [f8adcfe](https://github.com/kubedb/kafka/commit/f8adcfe) Use restricted pod security label (#31)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.19.0](https://github.com/kubedb/mariadb/releases/tag/v0.19.0)

- [f98e4730](https://github.com/kubedb/mariadb/commit/f98e4730) Prepare for release v0.19.0 (#228)
- [5d47fbb2](https://github.com/kubedb/mariadb/commit/5d47fbb2) Update dependencies (#227)
- [40a21a5e](https://github.com/kubedb/mariadb/commit/40a21a5e) Confirm the db has been paused before ops continue (#226)
- [97924029](https://github.com/kubedb/mariadb/commit/97924029) Update dependencies (#225)
- [edc232c8](https://github.com/kubedb/mariadb/commit/edc232c8) update nightly test profile to provisioner (#224)
- [0087bcfa](https://github.com/kubedb/mariadb/commit/0087bcfa) Add inputs fields to manual trigger ci file (#222)
- [ca265d0f](https://github.com/kubedb/mariadb/commit/ca265d0f) reduce get/patch api calls (#218)
- [ec7a0a79](https://github.com/kubedb/mariadb/commit/ec7a0a79) fix nightly test workflow (#221)
- [e145c47d](https://github.com/kubedb/mariadb/commit/e145c47d) Use KIND v0.20.0 (#220)
- [bc1cb72d](https://github.com/kubedb/mariadb/commit/bc1cb72d) Run nightly tests against master
- [c8f6dab2](https://github.com/kubedb/mariadb/commit/c8f6dab2) Update nightly.yml
- [f577fa72](https://github.com/kubedb/mariadb/commit/f577fa72) Run e2e tests nightly (#219)
- [256ae22a](https://github.com/kubedb/mariadb/commit/256ae22a) Use new kmapi Condition (#217)
- [37bbf08e](https://github.com/kubedb/mariadb/commit/37bbf08e) Use restricted pod security label (#216)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.15.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.15.0)

- [0d67bc6d](https://github.com/kubedb/mariadb-coordinator/commit/0d67bc6d) Prepare for release v0.15.0 (#89)
- [49c68129](https://github.com/kubedb/mariadb-coordinator/commit/49c68129) Update dependencies (#88)
- [e9c737c5](https://github.com/kubedb/mariadb-coordinator/commit/e9c737c5) Update dependencies (#87)
- [77b5b854](https://github.com/kubedb/mariadb-coordinator/commit/77b5b854) Reduce get/patch api calls (#86)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.28.0](https://github.com/kubedb/memcached/releases/tag/v0.28.0)

- [fd40e37e](https://github.com/kubedb/memcached/commit/fd40e37e) Prepare for release v0.28.0 (#402)
- [f759a6d6](https://github.com/kubedb/memcached/commit/f759a6d6) Update dependencies (#401)
- [4a82561c](https://github.com/kubedb/memcached/commit/4a82561c) Update dependencies (#400)
- [29b39605](https://github.com/kubedb/memcached/commit/29b39605) Fix e2e and nightly workflows
- [1c77d33f](https://github.com/kubedb/memcached/commit/1c77d33f) Use KIND v0.20.0 (#399)
- [bfc480a7](https://github.com/kubedb/memcached/commit/bfc480a7) Run nightly tests against master
- [7acbe89e](https://github.com/kubedb/memcached/commit/7acbe89e) Update nightly.yml
- [08adb133](https://github.com/kubedb/memcached/commit/08adb133) Run e2e tests nightly (#398)
- [3d10ada3](https://github.com/kubedb/memcached/commit/3d10ada3) Use new kmapi Condition (#397)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.28.0](https://github.com/kubedb/mongodb/releases/tag/v0.28.0)

- [df494f03](https://github.com/kubedb/mongodb/commit/df494f03) Prepare for release v0.28.0 (#568)
- [d8705c76](https://github.com/kubedb/mongodb/commit/d8705c76) Update dependencies (#567)
- [b84464c3](https://github.com/kubedb/mongodb/commit/b84464c3) Confirm the db has been paused before ops continue (#566)
- [b0e8a237](https://github.com/kubedb/mongodb/commit/b0e8a237) Update dependencies (#565)
- [ecd154fb](https://github.com/kubedb/mongodb/commit/ecd154fb) add test input (#564)
- [96fac12b](https://github.com/kubedb/mongodb/commit/96fac12b) Reduce get/patch api calls (#557)
- [568f3e28](https://github.com/kubedb/mongodb/commit/568f3e28) Fix stash installation (#563)
- [d905767f](https://github.com/kubedb/mongodb/commit/d905767f) Run only general profile tests
- [9692d296](https://github.com/kubedb/mongodb/commit/9692d296) Use KIND v0.20.0 (#562)
- [30fe37a7](https://github.com/kubedb/mongodb/commit/30fe37a7) Use --bind_ip to fix 3.4.* CrashLoopbackOff issue (#559)
- [f658d023](https://github.com/kubedb/mongodb/commit/f658d023) Run nightly.yml against master branch
- [af990bc2](https://github.com/kubedb/mongodb/commit/af990bc2) Run nightly tests against master branch
- [83abcb97](https://github.com/kubedb/mongodb/commit/83abcb97) Run e2e test nightly (#560)
- [35bb3970](https://github.com/kubedb/mongodb/commit/35bb3970) Use new kmapi Condition (#558)
- [6c5e8551](https://github.com/kubedb/mongodb/commit/6c5e8551) Update Makefile
- [02269ae8](https://github.com/kubedb/mongodb/commit/02269ae8) Use restricted pod security level (#556)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.28.0](https://github.com/kubedb/mysql/releases/tag/v0.28.0)




## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.13.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.13.0)

- [11ad765](https://github.com/kubedb/mysql-coordinator/commit/11ad765) Prepare for release v0.13.0 (#85)
- [12b4608](https://github.com/kubedb/mysql-coordinator/commit/12b4608) Update dependencies (#84)
- [9cd6e03](https://github.com/kubedb/mysql-coordinator/commit/9cd6e03) Update dependencies (#83)
- [4587ab8](https://github.com/kubedb/mysql-coordinator/commit/4587ab8) reduce k8s api calls for get and patch (#82)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.13.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.13.0)

- [2a59ae1](https://github.com/kubedb/mysql-router-init/commit/2a59ae1) Update dependencies (#36)
- [a4f4318](https://github.com/kubedb/mysql-router-init/commit/a4f4318) Update dependencies (#35)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.22.0](https://github.com/kubedb/ops-manager/releases/tag/v0.22.0)




## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.22.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.22.0)

- [6d522b6c](https://github.com/kubedb/percona-xtradb/commit/6d522b6c) Prepare for release v0.22.0 (#327)
- [9ba97882](https://github.com/kubedb/percona-xtradb/commit/9ba97882) Update dependencies (#326)
- [408be6e9](https://github.com/kubedb/percona-xtradb/commit/408be6e9) Confirm the db has been paused before ops continue (#325)
- [b314569f](https://github.com/kubedb/percona-xtradb/commit/b314569f) Update dependencies (#324)
- [ff7e5e09](https://github.com/kubedb/percona-xtradb/commit/ff7e5e09) Update nightly.yml
- [f1ddeb07](https://github.com/kubedb/percona-xtradb/commit/f1ddeb07) reduce get/patch api calls (#320)
- [b3d3564c](https://github.com/kubedb/percona-xtradb/commit/b3d3564c) Create nightly.yml
- [29f6ab80](https://github.com/kubedb/percona-xtradb/commit/29f6ab80) Remove nightly workflow
- [6c47d97f](https://github.com/kubedb/percona-xtradb/commit/6c47d97f) Merge pull request #323 from kubedb/fix-nightly
- [c8d2e630](https://github.com/kubedb/percona-xtradb/commit/c8d2e630) Fix nightly
- [c2854017](https://github.com/kubedb/percona-xtradb/commit/c2854017) Use KIND v0.20.0 (#322)
- [ff4d7c11](https://github.com/kubedb/percona-xtradb/commit/ff4d7c11) Run nightly tests against master
- [0328b6ad](https://github.com/kubedb/percona-xtradb/commit/0328b6ad) Update nightly.yml
- [eb533938](https://github.com/kubedb/percona-xtradb/commit/eb533938) Run e2e tests nightly (#321)
- [6be27644](https://github.com/kubedb/percona-xtradb/commit/6be27644) Use new kmapi Condition (#319)
- [33571e97](https://github.com/kubedb/percona-xtradb/commit/33571e97) Use restricted pod security label (#318)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.8.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.8.0)

- [2bc9a17](https://github.com/kubedb/percona-xtradb-coordinator/commit/2bc9a17) Prepare for release v0.8.0 (#46)
- [b886ff2](https://github.com/kubedb/percona-xtradb-coordinator/commit/b886ff2) Update dependencies (#45)
- [9d5feb9](https://github.com/kubedb/percona-xtradb-coordinator/commit/9d5feb9) Update dependencies (#44)
- [2c8983d](https://github.com/kubedb/percona-xtradb-coordinator/commit/2c8983d) reduce get/patch api calls (#43)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.19.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.19.0)

- [a8ee999d](https://github.com/kubedb/pg-coordinator/commit/a8ee999d) Prepare for release v0.19.0 (#129)
- [1434fdfc](https://github.com/kubedb/pg-coordinator/commit/1434fdfc) Update dependencies (#128)
- [36ceccc8](https://github.com/kubedb/pg-coordinator/commit/36ceccc8) Use cached client (#127)
- [190a4880](https://github.com/kubedb/pg-coordinator/commit/190a4880) Update dependencies (#126)
- [8aad969e](https://github.com/kubedb/pg-coordinator/commit/8aad969e) fix failover and standby sync issue (#125)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.22.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.22.0)

- [fe943791](https://github.com/kubedb/pgbouncer/commit/fe943791) Prepare for release v0.22.0 (#289)
- [489949d7](https://github.com/kubedb/pgbouncer/commit/489949d7) Update dependencies (#288)
- [ea7e4b3c](https://github.com/kubedb/pgbouncer/commit/ea7e4b3c) Update dependencies (#287)
- [8ddf699c](https://github.com/kubedb/pgbouncer/commit/8ddf699c) Fix: get and patch call issue (#285)
- [81fa0fb3](https://github.com/kubedb/pgbouncer/commit/81fa0fb3) Use KIND v0.20.0 (#286)
- [6bc9e12b](https://github.com/kubedb/pgbouncer/commit/6bc9e12b) Run nightly tests against master
- [655e1d06](https://github.com/kubedb/pgbouncer/commit/655e1d06) Update nightly.yml
- [2d1bc4e5](https://github.com/kubedb/pgbouncer/commit/2d1bc4e5) Update nightly.yml
- [94419822](https://github.com/kubedb/pgbouncer/commit/94419822) Run e2e tests nightly
- [c41aa109](https://github.com/kubedb/pgbouncer/commit/c41aa109) Use new kmapi Condition (#284)
- [e62cbde9](https://github.com/kubedb/pgbouncer/commit/e62cbde9) Update Makefile
- [6734feba](https://github.com/kubedb/pgbouncer/commit/6734feba) Use restricted pod security label (#283)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.35.0](https://github.com/kubedb/postgres/releases/tag/v0.35.0)

- [8e62ebef](https://github.com/kubedb/postgres/commit/8e62ebef5) Prepare for release v0.35.0 (#662)
- [1b23e335](https://github.com/kubedb/postgres/commit/1b23e3352) Update dependencies (#661)
- [92a455d7](https://github.com/kubedb/postgres/commit/92a455d74) Confirm the db has been paused before ops continue (#660)
- [e642b565](https://github.com/kubedb/postgres/commit/e642b5655) add pod watch permision (#659)
- [192be10e](https://github.com/kubedb/postgres/commit/192be10e5) fix client (#658)
- [b2bff6fe](https://github.com/kubedb/postgres/commit/b2bff6fe2) close client engine (#656)
- [df65982e](https://github.com/kubedb/postgres/commit/df65982ea) Update dependencies (#657)
- [63185866](https://github.com/kubedb/postgres/commit/631858669) Check all the replica's are connected to the primary (#654)
- [c0c7689b](https://github.com/kubedb/postgres/commit/c0c7689ba) fix get and patch call issue (#649)
- [cc5c1468](https://github.com/kubedb/postgres/commit/cc5c1468e) Merge pull request #653 from kubedb/fix-nightly
- [5e4a7a19](https://github.com/kubedb/postgres/commit/5e4a7a196) Fixed nightly yaml.
- [6bf8ea0b](https://github.com/kubedb/postgres/commit/6bf8ea0be) Use KIND v0.20.0 (#652)
- [5c3dc9c8](https://github.com/kubedb/postgres/commit/5c3dc9c8e) Run nightly tests against master
- [bcec1cfb](https://github.com/kubedb/postgres/commit/bcec1cfbc) Update nightly.yml
- [dc4ae6ad](https://github.com/kubedb/postgres/commit/dc4ae6ad2) Run e2e tests nightly (#651)
- [b958109a](https://github.com/kubedb/postgres/commit/b958109aa) Use new kmapi Condition (#650)
- [ca9f77af](https://github.com/kubedb/postgres/commit/ca9f77af0) Update Makefile
- [b36a06f2](https://github.com/kubedb/postgres/commit/b36a06f2d) Use restricted pod security label (#648)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.35.0](https://github.com/kubedb/provisioner/releases/tag/v0.35.0)

- [47f0fc82](https://github.com/kubedb/provisioner/commit/47f0fc823) Prepare for release v0.35.0 (#54)
- [8b716999](https://github.com/kubedb/provisioner/commit/8b716999e) Update dependencies (#53)
- [81dd67a3](https://github.com/kubedb/provisioner/commit/81dd67a31) Update dependencies (#52)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.22.0](https://github.com/kubedb/proxysql/releases/tag/v0.22.0)

- [8b568349](https://github.com/kubedb/proxysql/commit/8b568349) Prepare for release v0.22.0 (#310)
- [fdfd9943](https://github.com/kubedb/proxysql/commit/fdfd9943) Update dependencies (#309)
- [9dc1c3fc](https://github.com/kubedb/proxysql/commit/9dc1c3fc) Confirm the db has been paused before ops continue (#308)
- [e600efb7](https://github.com/kubedb/proxysql/commit/e600efb7) Update dependencies (#307)
- [63e65342](https://github.com/kubedb/proxysql/commit/63e65342) Add inputs fields to manual trigger ci file (#306)
- [800c10ae](https://github.com/kubedb/proxysql/commit/800c10ae) Update nightly.yml
- [b1816f1a](https://github.com/kubedb/proxysql/commit/b1816f1a) reduce get/patch api calls (#303)
- [21d71bc5](https://github.com/kubedb/proxysql/commit/21d71bc5) Merge pull request #305 from kubedb/fix_nightly
- [91a019e1](https://github.com/kubedb/proxysql/commit/91a019e1) Nightly fix
- [09492c1d](https://github.com/kubedb/proxysql/commit/09492c1d) Run nightly tests against master
- [37547f7b](https://github.com/kubedb/proxysql/commit/37547f7b) Update nightly.yml
- [3e31ed14](https://github.com/kubedb/proxysql/commit/3e31ed14) Update nightly.yml
- [491ad083](https://github.com/kubedb/proxysql/commit/491ad083) RUn e2e tests nightly (#304)
- [b6151aeb](https://github.com/kubedb/proxysql/commit/b6151aeb) Use new kmapi Condition (#302)
- [7d5756b1](https://github.com/kubedb/proxysql/commit/7d5756b1) Use restricted pod security label (#301)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.28.0](https://github.com/kubedb/redis/releases/tag/v0.28.0)

- [ea9ebdf2](https://github.com/kubedb/redis/commit/ea9ebdf2) Prepare for release v0.28.0 (#484)
- [63b32c43](https://github.com/kubedb/redis/commit/63b32c43) Update dependencies (#483)
- [7a2df42c](https://github.com/kubedb/redis/commit/7a2df42c) Confirm the db has been paused before ops continue (#482)
- [ca81ca3f](https://github.com/kubedb/redis/commit/ca81ca3f) Update dependencies (#481)
- [b13517b1](https://github.com/kubedb/redis/commit/b13517b1) update nightly test profile to provisioner (#480)
- [c00dab7c](https://github.com/kubedb/redis/commit/c00dab7c) Add inputs to nightly workflow (#479)
- [06ca0ad2](https://github.com/kubedb/redis/commit/06ca0ad2) Fix nightly (#477)
- [33ee7af4](https://github.com/kubedb/redis/commit/33ee7af4) Fix Redis nightly test workflow
- [5647852a](https://github.com/kubedb/redis/commit/5647852a) Use KIND v0.20.0 (#476)
- [5ef88f14](https://github.com/kubedb/redis/commit/5ef88f14) Run nightly tests against master
- [083a9124](https://github.com/kubedb/redis/commit/083a9124) Update nightly.yml
- [aa9b75ae](https://github.com/kubedb/redis/commit/aa9b75ae) Run e2e tests nightly (#473)
- [b4f312f4](https://github.com/kubedb/redis/commit/b4f312f4) Reduce get/patch api calls (#471)
- [ebd30b79](https://github.com/kubedb/redis/commit/ebd30b79) Use new kmapi Condition (#472)
- [5d191bf8](https://github.com/kubedb/redis/commit/5d191bf8) Update Makefile
- [aaf4a815](https://github.com/kubedb/redis/commit/aaf4a815) Use restricted pod security label (#470)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.14.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.14.0)

- [136eb79](https://github.com/kubedb/redis-coordinator/commit/136eb79) Prepare for release v0.14.0 (#77)
- [cc749e6](https://github.com/kubedb/redis-coordinator/commit/cc749e6) Update dependencies (#76)
- [fbc75dd](https://github.com/kubedb/redis-coordinator/commit/fbc75dd) Update dependencies (#75)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.22.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.22.0)

- [11106cf6](https://github.com/kubedb/replication-mode-detector/commit/11106cf6) Prepare for release v0.22.0 (#241)
- [5a7a2a75](https://github.com/kubedb/replication-mode-detector/commit/5a7a2a75) Update dependencies (#240)
- [914e86dc](https://github.com/kubedb/replication-mode-detector/commit/914e86dc) Update dependencies (#239)
- [f30374d2](https://github.com/kubedb/replication-mode-detector/commit/f30374d2) Use new kmapi Condition (#238)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.11.0](https://github.com/kubedb/schema-manager/releases/tag/v0.11.0)

- [65459bc6](https://github.com/kubedb/schema-manager/commit/65459bc6) Prepare for release v0.11.0 (#81)
- [30dd907b](https://github.com/kubedb/schema-manager/commit/30dd907b) Update dependencies (#80)
- [472e7496](https://github.com/kubedb/schema-manager/commit/472e7496) Update dependencies (#79)
- [1c4a60a8](https://github.com/kubedb/schema-manager/commit/1c4a60a8) Use new kmapi Condition (#78)
- [bec3c7b8](https://github.com/kubedb/schema-manager/commit/bec3c7b8) Update Makefile
- [3df964b1](https://github.com/kubedb/schema-manager/commit/3df964b1) Use restricted pod security label (#77)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.20.0](https://github.com/kubedb/tests/releases/tag/v0.20.0)

- [ea935103](https://github.com/kubedb/tests/commit/ea935103) Prepare for release v0.20.0 (#242)
- [bc927923](https://github.com/kubedb/tests/commit/bc927923) Update dependencies (#241)
- [319bf4a2](https://github.com/kubedb/tests/commit/319bf4a2) Fix mg termination_policy & env_variables (#237)
- [5424bdbe](https://github.com/kubedb/tests/commit/5424bdbe) update vertical scaling constant (#239)
- [40229b3d](https://github.com/kubedb/tests/commit/40229b3d) Update dependencies (#238)
- [68deeafb](https://github.com/kubedb/tests/commit/68deeafb) Exclude volume expansion (#235)
- [7a364367](https://github.com/kubedb/tests/commit/7a364367) Fix test for ES & OS with disabled security (#232)
- [313151de](https://github.com/kubedb/tests/commit/313151de) fix mariadb test (#234)
- [c5d9911e](https://github.com/kubedb/tests/commit/c5d9911e) Update tests by test profile (#231)
- [b2a5f384](https://github.com/kubedb/tests/commit/b2a5f384) Fix general tests (#230)
- [a13b9095](https://github.com/kubedb/tests/commit/a13b9095) Use new kmapi Condition (#229)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.11.0](https://github.com/kubedb/ui-server/releases/tag/v0.11.0)

- [7ccfc49c](https://github.com/kubedb/ui-server/commit/7ccfc49c) Prepare for release v0.11.0 (#89)
- [04edecf9](https://github.com/kubedb/ui-server/commit/04edecf9) Update dependencies (#88)
- [8d1f7b4b](https://github.com/kubedb/ui-server/commit/8d1f7b4b) Update dependencies (#87)
- [a0fd42d8](https://github.com/kubedb/ui-server/commit/a0fd42d8) Use new kmapi Condition (#86)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.11.0](https://github.com/kubedb/webhook-server/releases/tag/v0.11.0)

- [26e96671](https://github.com/kubedb/webhook-server/commit/26e96671) Prepare for release (#66)
- [d446d877](https://github.com/kubedb/webhook-server/commit/d446d877) Update dependencies (#65)
- [278f450b](https://github.com/kubedb/webhook-server/commit/278f450b) Use KIND v0.20.0 (#64)
- [6ba4191e](https://github.com/kubedb/webhook-server/commit/6ba4191e) Use new kmapi Condition (#63)
- [5cc21a08](https://github.com/kubedb/webhook-server/commit/5cc21a08) Update Makefile
- [0edd0610](https://github.com/kubedb/webhook-server/commit/0edd0610) Use restricted pod security label (#62)




