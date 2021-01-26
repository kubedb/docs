---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2021.01.26
    name: Changelog-v2021.01.26
    parent: welcome
    weight: 20210126
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2021.01.26/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2021.01.26/
---

# KubeDB v2021.01.26 (2021-01-26)


## [appscode/kubedb-autoscaler](https://github.com/appscode/kubedb-autoscaler)

### [v0.1.2](https://github.com/appscode/kubedb-autoscaler/releases/tag/v0.1.2)

- [8a42374](https://github.com/appscode/kubedb-autoscaler/commit/8a42374) Prepare for release v0.1.2 (#15)
- [75e0b0e](https://github.com/appscode/kubedb-autoscaler/commit/75e0b0e) Update repository config (#13)
- [bf1487e](https://github.com/appscode/kubedb-autoscaler/commit/bf1487e) Fix Elasticsearch storage autoscaler (#12)
- [b23280c](https://github.com/appscode/kubedb-autoscaler/commit/b23280c) Update readme
- [d320045](https://github.com/appscode/kubedb-autoscaler/commit/d320045) Fix Elasticsearch Autoscaler (#11)



## [appscode/kubedb-enterprise](https://github.com/appscode/kubedb-enterprise)

### [v0.3.2](https://github.com/appscode/kubedb-enterprise/releases/tag/v0.3.2)

- [d235a3ec](https://github.com/appscode/kubedb-enterprise/commit/d235a3ec) Prepare for release v0.3.2 (#132)
- [98ac77be](https://github.com/appscode/kubedb-enterprise/commit/98ac77be) Delete operator generated owned certificate secrets before creating new ones (#131)
- [a8e699f9](https://github.com/appscode/kubedb-enterprise/commit/a8e699f9) Ingore paused DB events from enterprise operator too (#130)
- [fcaf1b8b](https://github.com/appscode/kubedb-enterprise/commit/fcaf1b8b) Fix scale up and scale down (#124)
- [7d37df14](https://github.com/appscode/kubedb-enterprise/commit/7d37df14) Update Kubernetes v1.18.9 dependencies (#114)
- [ff12ad3c](https://github.com/appscode/kubedb-enterprise/commit/ff12ad3c) Update reconfigureTLS for Elasticsearch (#125)
- [0e9e15c6](https://github.com/appscode/kubedb-enterprise/commit/0e9e15c6) Use `NewSpecStatusChangeHandler` for Ops Requests (#129)
- [00c41590](https://github.com/appscode/kubedb-enterprise/commit/00c41590) Change `DBSizeDiffPercentage` to `ObjectsCountDiffPercentage` (#128)
- [4bfcacad](https://github.com/appscode/kubedb-enterprise/commit/4bfcacad) Update repository config (#127)
- [f0570d8b](https://github.com/appscode/kubedb-enterprise/commit/f0570d8b) Update repository config (#126)
- [ddf7ca41](https://github.com/appscode/kubedb-enterprise/commit/ddf7ca41) Check readiness gates for IsPodReady (#123)



## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.16.2](https://github.com/kubedb/apimachinery/releases/tag/v0.16.2)

- [7eb1fdda](https://github.com/kubedb/apimachinery/commit/7eb1fdda) Update Kubernetes v1.18.9 dependencies (#692)
- [ed484da9](https://github.com/kubedb/apimachinery/commit/ed484da9) Don't add default subject to certificate if already exists (#689)
- [d3b5b50e](https://github.com/kubedb/apimachinery/commit/d3b5b50e) Change `DBSizeDiffPercentage` to `ObjectsCountDiffPercentage` (#690)
- [63e27a25](https://github.com/kubedb/apimachinery/commit/63e27a25) Update for release Stash@v2021.01.21 (#691)
- [459684a5](https://github.com/kubedb/apimachinery/commit/459684a5) Update repository config (#688)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.16.2](https://github.com/kubedb/cli/releases/tag/v0.16.2)

- [ada47bf8](https://github.com/kubedb/cli/commit/ada47bf8) Prepare for release v0.16.2 (#584)
- [ff1a7aac](https://github.com/kubedb/cli/commit/ff1a7aac) Update Kubernetes v1.18.9 dependencies (#583)
- [664f1b1c](https://github.com/kubedb/cli/commit/664f1b1c) Update for release Stash@v2021.01.21 (#582)
- [7a07edfd](https://github.com/kubedb/cli/commit/7a07edfd) Update repository config (#581)
- [2ddea9f5](https://github.com/kubedb/cli/commit/2ddea9f5) Update repository config (#580)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.16.2](https://github.com/kubedb/elasticsearch/releases/tag/v0.16.2)

- [7787d2a6](https://github.com/kubedb/elasticsearch/commit/7787d2a6) Prepare for release v0.16.2 (#463)
- [29e4198a](https://github.com/kubedb/elasticsearch/commit/29e4198a) Add nodeDNs to configuration even when enableSSL is false (#458)
- [4a76db12](https://github.com/kubedb/elasticsearch/commit/4a76db12) Update Kubernetes v1.18.9 dependencies (#462)
- [42680118](https://github.com/kubedb/elasticsearch/commit/42680118) Update for release Stash@v2021.01.21 (#461)
- [27525afb](https://github.com/kubedb/elasticsearch/commit/27525afb) Update repository config (#460)
- [02d0fb3f](https://github.com/kubedb/elasticsearch/commit/02d0fb3f) Update repository config (#459)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v0.16.2](https://github.com/kubedb/installer/releases/tag/v0.16.2)

- [61bbb19](https://github.com/kubedb/installer/commit/61bbb19) Prepare for release v0.16.2 (#227)
- [091665f](https://github.com/kubedb/installer/commit/091665f) Revert "Update Percona MongoDB Server Images (#219)"
- [9736ad8](https://github.com/kubedb/installer/commit/9736ad8) Add permission to add finalizers on custom resoures (#226)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.9.2](https://github.com/kubedb/memcached/releases/tag/v0.9.2)

- [2a1e2e7c](https://github.com/kubedb/memcached/commit/2a1e2e7c) Prepare for release v0.9.2 (#277)
- [dd5f19d6](https://github.com/kubedb/memcached/commit/dd5f19d6) Update Kubernetes v1.18.9 dependencies (#276)
- [2dfc00ee](https://github.com/kubedb/memcached/commit/2dfc00ee) Update repository config (#275)
- [a4278122](https://github.com/kubedb/memcached/commit/a4278122) Update repository config (#274)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.9.2](https://github.com/kubedb/mongodb/releases/tag/v0.9.2)

- [9ecb8b3f](https://github.com/kubedb/mongodb/commit/9ecb8b3f) Prepare for release v0.9.2 (#362)
- [6ff0b2ab](https://github.com/kubedb/mongodb/commit/6ff0b2ab) Return error when catalog doesn't exist (#361)
- [70559218](https://github.com/kubedb/mongodb/commit/70559218) Update Kubernetes v1.18.9 dependencies (#360)
- [e46daaf7](https://github.com/kubedb/mongodb/commit/e46daaf7) Update for release Stash@v2021.01.21 (#359)
- [dd4c2fcf](https://github.com/kubedb/mongodb/commit/dd4c2fcf) Update repository config (#358)
- [f8ab57cb](https://github.com/kubedb/mongodb/commit/f8ab57cb) Update repository config (#357)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.9.2](https://github.com/kubedb/mysql/releases/tag/v0.9.2)

- [5f7dfd8c](https://github.com/kubedb/mysql/commit/5f7dfd8c) Prepare for release v0.9.2 (#351)
- [26ef56cb](https://github.com/kubedb/mysql/commit/26ef56cb) Configure innodb buffer pool and group repl cache size (#350)
- [6562cf8e](https://github.com/kubedb/mysql/commit/6562cf8e) Fix Health-checker for standalone (#345)
- [f20f5763](https://github.com/kubedb/mysql/commit/f20f5763) Update Kubernetes v1.18.9 dependencies (#349)
- [e11bea0b](https://github.com/kubedb/mysql/commit/e11bea0b) Update for release Stash@v2021.01.21 (#348)
- [5cdc3424](https://github.com/kubedb/mysql/commit/5cdc3424) Update repository config (#347)
- [0438f075](https://github.com/kubedb/mysql/commit/0438f075) Update repository config (#346)



## [kubedb/operator](https://github.com/kubedb/operator)

### [v0.16.2](https://github.com/kubedb/operator/releases/tag/v0.16.2)

- [92baf160](https://github.com/kubedb/operator/commit/92baf160) Prepare for release v0.16.2 (#385)
- [aa818921](https://github.com/kubedb/operator/commit/aa818921) Update Kubernetes v1.18.9 dependencies (#384)
- [8344e056](https://github.com/kubedb/operator/commit/8344e056) Update repository config (#383)
- [242bae58](https://github.com/kubedb/operator/commit/242bae58) Update repository config (#382)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.3.2](https://github.com/kubedb/percona-xtradb/releases/tag/v0.3.2)

- [875dbcfb](https://github.com/kubedb/percona-xtradb/commit/875dbcfb) Prepare for release v0.3.2 (#173)
- [afcb2e37](https://github.com/kubedb/percona-xtradb/commit/afcb2e37) Update Kubernetes v1.18.9 dependencies (#172)
- [48aa03cc](https://github.com/kubedb/percona-xtradb/commit/48aa03cc) Update for release Stash@v2021.01.21 (#171)
- [2bd07624](https://github.com/kubedb/percona-xtradb/commit/2bd07624) Update repository config (#170)
- [d10fccc5](https://github.com/kubedb/percona-xtradb/commit/d10fccc5) Update repository config (#169)



## [kubedb/pg-leader-election](https://github.com/kubedb/pg-leader-election)

### [v0.4.2](https://github.com/kubedb/pg-leader-election/releases/tag/v0.4.2)

- [4162fc7](https://github.com/kubedb/pg-leader-election/commit/4162fc7) Update Kubernetes v1.18.9 dependencies (#47)
- [5f1ec75](https://github.com/kubedb/pg-leader-election/commit/5f1ec75) Update repository config (#46)
- [6f29932](https://github.com/kubedb/pg-leader-election/commit/6f29932) Update repository config (#45)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.3.2](https://github.com/kubedb/pgbouncer/releases/tag/v0.3.2)

- [1c1f20bf](https://github.com/kubedb/pgbouncer/commit/1c1f20bf) Prepare for release v0.3.2 (#138)
- [7dc88cc4](https://github.com/kubedb/pgbouncer/commit/7dc88cc4) Update Kubernetes v1.18.9 dependencies (#137)
- [c2574a34](https://github.com/kubedb/pgbouncer/commit/c2574a34) Update repository config (#136)
- [d5baad1f](https://github.com/kubedb/pgbouncer/commit/d5baad1f) Update repository config (#135)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.16.2](https://github.com/kubedb/postgres/releases/tag/v0.16.2)

- [a5e81da7](https://github.com/kubedb/postgres/commit/a5e81da7) Prepare for release v0.16.2 (#462)
- [f5bcfb66](https://github.com/kubedb/postgres/commit/f5bcfb66) Update Kubernetes v1.18.9 dependencies (#461)
- [c8e4da8b](https://github.com/kubedb/postgres/commit/c8e4da8b) Update for release Stash@v2021.01.21 (#460)
- [d0d9c090](https://github.com/kubedb/postgres/commit/d0d9c090) Update repository config (#459)
- [9323c043](https://github.com/kubedb/postgres/commit/9323c043) Update repository config (#458)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.3.2](https://github.com/kubedb/proxysql/releases/tag/v0.3.2)

- [928bac65](https://github.com/kubedb/proxysql/commit/928bac65) Prepare for release v0.3.2 (#154)
- [49a9a9f6](https://github.com/kubedb/proxysql/commit/49a9a9f6) Update Kubernetes v1.18.9 dependencies (#153)
- [830eb7c6](https://github.com/kubedb/proxysql/commit/830eb7c6) Update for release Stash@v2021.01.21 (#152)
- [aa856424](https://github.com/kubedb/proxysql/commit/aa856424) Update repository config (#151)
- [6b16f30c](https://github.com/kubedb/proxysql/commit/6b16f30c) Update repository config (#150)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.9.2](https://github.com/kubedb/redis/releases/tag/v0.9.2)

- [a94faf53](https://github.com/kubedb/redis/commit/a94faf53) Prepare for release v0.9.2 (#299)
- [cfcbb855](https://github.com/kubedb/redis/commit/cfcbb855) Update Kubernetes v1.18.9 dependencies (#298)
- [76b9b70c](https://github.com/kubedb/redis/commit/76b9b70c) Update repository config (#297)
- [0cb62a27](https://github.com/kubedb/redis/commit/0cb62a27) Update repository config (#296)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.3.2](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.3.2)

- [a2e3ff5](https://github.com/kubedb/replication-mode-detector/commit/a2e3ff5) Prepare for release v0.3.2 (#123)
- [1b43ee1](https://github.com/kubedb/replication-mode-detector/commit/1b43ee1) Update Kubernetes v1.18.9 dependencies (#122)
- [a0e0fc0](https://github.com/kubedb/replication-mode-detector/commit/a0e0fc0) Update repository config (#121)
- [84155f6](https://github.com/kubedb/replication-mode-detector/commit/84155f6) Update repository config (#120)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.1.2](https://github.com/kubedb/tests/releases/tag/v0.1.2)

- [6b6c030](https://github.com/kubedb/tests/commit/6b6c030) Prepare for release v0.1.2 (#95)
- [3456495](https://github.com/kubedb/tests/commit/3456495) Update Kubernetes v1.18.9 dependencies (#92)
- [e335294](https://github.com/kubedb/tests/commit/e335294) Update repository config (#91)
- [9d82b07](https://github.com/kubedb/tests/commit/9d82b07) Update repository config (#90)




