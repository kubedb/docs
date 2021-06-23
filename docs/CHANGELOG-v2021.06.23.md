---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2021.06.23
    name: Changelog-v2021.06.23
    parent: welcome
    weight: 20210623
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2021.06.23/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2021.06.23/
---

# KubeDB v2021.06.23 (2021-06-23)


## [appscode/kubedb-autoscaler](https://github.com/appscode/kubedb-autoscaler)

### [v0.4.0](https://github.com/appscode/kubedb-autoscaler/releases/tag/v0.4.0)

- [93e27c4](https://github.com/appscode/kubedb-autoscaler/commit/93e27c4) Prepare for release v0.4.0 (#27)
- [e7e6c98](https://github.com/appscode/kubedb-autoscaler/commit/e7e6c98) Disable api priority and fairness feature for webhook server (#26)
- [fe37e94](https://github.com/appscode/kubedb-autoscaler/commit/fe37e94) Prepare for release v0.4.0-rc.0 (#25)
- [f81237c](https://github.com/appscode/kubedb-autoscaler/commit/f81237c) Update audit lib (#24)
- [ed8c87b](https://github.com/appscode/kubedb-autoscaler/commit/ed8c87b) Send audit events if analytics enabled
- [c0a03d5](https://github.com/appscode/kubedb-autoscaler/commit/c0a03d5) Create auditor if license file is provided (#23)
- [2775227](https://github.com/appscode/kubedb-autoscaler/commit/2775227) Publish audit events (#22)
- [636d3a7](https://github.com/appscode/kubedb-autoscaler/commit/636d3a7) Use kglog helper
- [6a64bb1](https://github.com/appscode/kubedb-autoscaler/commit/6a64bb1) Use k8s 1.21.0 toolchain (#21)



## [appscode/kubedb-enterprise](https://github.com/appscode/kubedb-enterprise)

### [v0.6.0](https://github.com/appscode/kubedb-enterprise/releases/tag/v0.6.0)

- [4b2195b7](https://github.com/appscode/kubedb-enterprise/commit/4b2195b7) Prepare for release v0.6.0 (#202)
- [cde1a54c](https://github.com/appscode/kubedb-enterprise/commit/cde1a54c) Improve Elasticsearch version upgrade with reconciler (#201)
- [c4c2d3c9](https://github.com/appscode/kubedb-enterprise/commit/c4c2d3c9) Use NSS_Wrapper for Pg_Upgrade Command (#200)
- [97749287](https://github.com/appscode/kubedb-enterprise/commit/97749287) Prepare for release v0.6.0-rc.0 (#199)
- [401cfc86](https://github.com/appscode/kubedb-enterprise/commit/401cfc86) Update audit lib (#197)
- [4378d35a](https://github.com/appscode/kubedb-enterprise/commit/4378d35a) Add MariaDB OpsReq [Restart, Upgrade, Scaling, Volume Expansion, Reconfigure Custom Config] (#179)
- [f879e934](https://github.com/appscode/kubedb-enterprise/commit/f879e934) Postgres Ops Req (Upgrade, Horizontal, Vertical, Volume Expansion, Reconfigure, Reconfigure TLS, Restart) (#193)
- [79b51d25](https://github.com/appscode/kubedb-enterprise/commit/79b51d25) Skip stash checks if stash CRD doesn't exist (#196)
- [3efc4ee8](https://github.com/appscode/kubedb-enterprise/commit/3efc4ee8) Refactor MongoDB Scale Down Shard (#189)
- [64962f36](https://github.com/appscode/kubedb-enterprise/commit/64962f36) Add timeout for Elasticsearch ops request (#183)
- [4ed736b8](https://github.com/appscode/kubedb-enterprise/commit/4ed736b8) Send audit events if analytics enabled
- [498ef67b](https://github.com/appscode/kubedb-enterprise/commit/498ef67b) Create auditor if license file is provided (#195)
- [a61965cc](https://github.com/appscode/kubedb-enterprise/commit/a61965cc) Publish audit events (#194)
- [cdc0ee37](https://github.com/appscode/kubedb-enterprise/commit/cdc0ee37) Fix log level issue with klog (#187)
- [356c6965](https://github.com/appscode/kubedb-enterprise/commit/356c6965) Use kglog helper
- [d7248cfd](https://github.com/appscode/kubedb-enterprise/commit/d7248cfd) Update Kubernetes toolchain to v1.21.0 (#181)
- [b8493083](https://github.com/appscode/kubedb-enterprise/commit/b8493083) Only restart the changed pods while VerticalScaling Elasticsearch (#174)



## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.19.0](https://github.com/kubedb/apimachinery/releases/tag/v0.19.0)

- [7cecea8e](https://github.com/kubedb/apimachinery/commit/7cecea8e) Add docs badge
- [c885fc2d](https://github.com/kubedb/apimachinery/commit/c885fc2d) Postgres DB Container's RunAsGroup As FSGroup (#769)
- [29cb0260](https://github.com/kubedb/apimachinery/commit/29cb0260) Add fixes to helper method (#768)
- [b20b40c2](https://github.com/kubedb/apimachinery/commit/b20b40c2) Use Stash v2021.06.23
- [e98fb31f](https://github.com/kubedb/apimachinery/commit/e98fb31f) Update audit event publisher (#767)
- [81e26637](https://github.com/kubedb/apimachinery/commit/81e26637) Add MariaDB Constants (#766)
- [532b6982](https://github.com/kubedb/apimachinery/commit/532b6982) Update Elasticsearch API to support various node roles including hot-warm-cold (#764)
- [a9979e15](https://github.com/kubedb/apimachinery/commit/a9979e15) Update for release Stash@v2021.6.18 (#765)
- [d20c46a2](https://github.com/kubedb/apimachinery/commit/d20c46a2) Fix locking in ResourceMapper
- [3a597982](https://github.com/kubedb/apimachinery/commit/3a597982) Send audit events if analytics enabled
- [27cc118e](https://github.com/kubedb/apimachinery/commit/27cc118e) Add auditor to shared Controller (#761)
- [eb13a94f](https://github.com/kubedb/apimachinery/commit/eb13a94f) Rename TimeoutSeconds to Timeout in MongoDBOpsRequest (#759)
- [29627ec6](https://github.com/kubedb/apimachinery/commit/29627ec6) Add timeout for each step of ES ops request (#742)
- [cc6b9690](https://github.com/kubedb/apimachinery/commit/cc6b9690) Add MariaDB OpsRequest Types (#743)
- [6fb2646e](https://github.com/kubedb/apimachinery/commit/6fb2646e) Update default resource limits for databases (#755)
- [161b3fe3](https://github.com/kubedb/apimachinery/commit/161b3fe3) Add UpdateMariaDBOpsRequestStatus function (#727)
- [98cd75f0](https://github.com/kubedb/apimachinery/commit/98cd75f0) Add Fields, Constant, Func  For Ops Request Postgres (#758)
- [722656b7](https://github.com/kubedb/apimachinery/commit/722656b7) Add Innodb Group Replication Mode (#750)
- [eb8e5883](https://github.com/kubedb/apimachinery/commit/eb8e5883) Replace go-bindata with //go:embed (#753)
- [df570f7b](https://github.com/kubedb/apimachinery/commit/df570f7b) Add HealthCheckInterval constant (#752)
- [e982e590](https://github.com/kubedb/apimachinery/commit/e982e590) Use kglog helper
- [e725873d](https://github.com/kubedb/apimachinery/commit/e725873d) Fix tests (#749)
- [11d1c306](https://github.com/kubedb/apimachinery/commit/11d1c306) Cleanup dependencies
- [7030bd8f](https://github.com/kubedb/apimachinery/commit/7030bd8f) Update crds
- [766fa11f](https://github.com/kubedb/apimachinery/commit/766fa11f) Update Kubernetes toolchain to v1.21.0 (#746)
- [12014667](https://github.com/kubedb/apimachinery/commit/12014667) Add Elasticsearch vertical scaling constants (#741)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.19.0](https://github.com/kubedb/cli/releases/tag/v0.19.0)

- [2b394bba](https://github.com/kubedb/cli/commit/2b394bba) Prepare for release v0.19.0 (#610)
- [b367d2a5](https://github.com/kubedb/cli/commit/b367d2a5) Prepare for release v0.19.0-rc.0 (#609)
- [b9214d6a](https://github.com/kubedb/cli/commit/b9214d6a) Use Kubernetes 1.21.1 toolchain (#608)
- [36866cf5](https://github.com/kubedb/cli/commit/36866cf5) Use kglog helper
- [e4ee9973](https://github.com/kubedb/cli/commit/e4ee9973) Cleanup dependencies (#607)
- [07999fc2](https://github.com/kubedb/cli/commit/07999fc2) Use Kubernetes v1.21.0 toolchain (#606)
- [05e3b7e5](https://github.com/kubedb/cli/commit/05e3b7e5) Use Kubernetes v1.21.0 toolchain (#605)
- [44f4188e](https://github.com/kubedb/cli/commit/44f4188e) Use Kubernetes v1.21.0 toolchain (#604)
- [82cd8399](https://github.com/kubedb/cli/commit/82cd8399) Use Kubernetes v1.21.0 toolchain (#603)
- [998506cd](https://github.com/kubedb/cli/commit/998506cd) Use Kubernetes v1.21.0 toolchain (#602)
- [4ff64f94](https://github.com/kubedb/cli/commit/4ff64f94) Use Kubernetes v1.21.0 toolchain (#601)
- [19b257f1](https://github.com/kubedb/cli/commit/19b257f1) Update Kubernetes toolchain to v1.21.0 (#600)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.19.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.19.0)

- [a38490f9](https://github.com/kubedb/elasticsearch/commit/a38490f9) Prepare for release v0.19.0 (#503)
- [aed0fcb4](https://github.com/kubedb/elasticsearch/commit/aed0fcb4) Prepare for release v0.19.0-rc.0 (#502)
- [630d6940](https://github.com/kubedb/elasticsearch/commit/630d6940) Update audit lib (#501)
- [df4c9a0d](https://github.com/kubedb/elasticsearch/commit/df4c9a0d) Do not create user credentials when security is disabled (#500)
- [3b656b57](https://github.com/kubedb/elasticsearch/commit/3b656b57) Add support for various node roles for ElasticStack (#499)
- [64133cb6](https://github.com/kubedb/elasticsearch/commit/64133cb6) Send audit events if analytics enabled
- [21caa38f](https://github.com/kubedb/elasticsearch/commit/21caa38f) Create auditor if license file is provided (#498)
- [8319ba70](https://github.com/kubedb/elasticsearch/commit/8319ba70) Publish audit events (#497)
- [5f08d1b2](https://github.com/kubedb/elasticsearch/commit/5f08d1b2) Skip health check for halted DB (#494)
- [6a23d464](https://github.com/kubedb/elasticsearch/commit/6a23d464) Disable flow control if api is not enabled (#495)
- [a23c5481](https://github.com/kubedb/elasticsearch/commit/a23c5481) Fix log level issue with klog (#496)
- [38dbddda](https://github.com/kubedb/elasticsearch/commit/38dbddda) Limit health checker go-routine for specific DB object (#491)
- [0aefd5f7](https://github.com/kubedb/elasticsearch/commit/0aefd5f7) Use kglog helper
- [03255078](https://github.com/kubedb/elasticsearch/commit/03255078) Cleanup glog dependency
- [57bb1bf1](https://github.com/kubedb/elasticsearch/commit/57bb1bf1) Update dependencies
- [69fdfde7](https://github.com/kubedb/elasticsearch/commit/69fdfde7) Update Kubernetes toolchain to v1.21.0 (#492)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2021.06.23](https://github.com/kubedb/installer/releases/tag/v2021.06.23)

- [334e8b4](https://github.com/kubedb/installer/commit/334e8b4) Prepare for release v2021.06.23 (#317)
- [823feb3](https://github.com/kubedb/installer/commit/823feb3) Prepare for release v2021.06.21-rc.0 (#315)
- [946dc13](https://github.com/kubedb/installer/commit/946dc13) Use Stash v2021.06.23
- [77a54a1](https://github.com/kubedb/installer/commit/77a54a1) Use Kubernetes 1.21.1 toolchain (#314)
- [2b15157](https://github.com/kubedb/installer/commit/2b15157) Add support for Elasticsearch v7.13.2 (#313)
- [a11d7d0](https://github.com/kubedb/installer/commit/a11d7d0) Support MongoDB Version 4.4.6 (#312)
- [4c79e1a](https://github.com/kubedb/installer/commit/4c79e1a) Update Elasticsearch versions to support various node roles (#308)
- [8e52114](https://github.com/kubedb/installer/commit/8e52114) Update for release Stash@v2021.6.18 (#311)
- [95aa010](https://github.com/kubedb/installer/commit/95aa010) Update to MariaDB init docker version 0.2.0 (#310)
- [1659b91](https://github.com/kubedb/installer/commit/1659b91) Fix: Update Ops Request yaml for Reconfigure TLS in Postgres (#307)
- [b2a806b](https://github.com/kubedb/installer/commit/b2a806b) Use mongodb-exporter v0.20.4 (#305)
- [12e720a](https://github.com/kubedb/installer/commit/12e720a) Update Kubernetes toolchain to v1.21.0 (#302)
- [3ff3bc3](https://github.com/kubedb/installer/commit/3ff3bc3) Add monitoring values to global chart (#301)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.3.0](https://github.com/kubedb/mariadb/releases/tag/v0.3.0)

- [189cc352](https://github.com/kubedb/mariadb/commit/189cc352) Prepare for release v0.3.0 (#78)
- [9b982f74](https://github.com/kubedb/mariadb/commit/9b982f74) Prepare for release v0.3.0-rc.0 (#77)
- [0ad0022c](https://github.com/kubedb/mariadb/commit/0ad0022c) Update audit lib (#75)
- [501a2e61](https://github.com/kubedb/mariadb/commit/501a2e61) Update custom config mount path for MariaDB Cluster (#59)
- [d00cf65b](https://github.com/kubedb/mariadb/commit/d00cf65b) Separate Reconcile functionality in a new function ReconcileNode (#68)
- [e9239d4f](https://github.com/kubedb/mariadb/commit/e9239d4f) Limit Go routines in Health Checker (#73)
- [d695adf1](https://github.com/kubedb/mariadb/commit/d695adf1) Send audit events if analytics enabled (#74)
- [070a0f79](https://github.com/kubedb/mariadb/commit/070a0f79) Create auditor if license file is provided (#72)
- [fc9046c3](https://github.com/kubedb/mariadb/commit/fc9046c3) Publish audit events (#71)
- [3a1f08a9](https://github.com/kubedb/mariadb/commit/3a1f08a9) Fix log level issue with klog for MariaDB (#70)
- [b6075e5d](https://github.com/kubedb/mariadb/commit/b6075e5d) Use kglog helper
- [f510e375](https://github.com/kubedb/mariadb/commit/f510e375) Use klog/v2
- [c009905e](https://github.com/kubedb/mariadb/commit/c009905e) Update Kubernetes toolchain to v1.21.0 (#66)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.12.0](https://github.com/kubedb/memcached/releases/tag/v0.12.0)

- [9c2c58d7](https://github.com/kubedb/memcached/commit/9c2c58d7) Prepare for release v0.12.0 (#301)
- [604d95db](https://github.com/kubedb/memcached/commit/604d95db) Disable api priority and fairness feature for webhook server (#300)
- [99ab26b5](https://github.com/kubedb/memcached/commit/99ab26b5) Prepare for release v0.12.0-rc.0 (#299)
- [213807d5](https://github.com/kubedb/memcached/commit/213807d5) Update audit lib (#298)
- [29054b5b](https://github.com/kubedb/memcached/commit/29054b5b) Send audit events if analytics enabled (#297)
- [a4888446](https://github.com/kubedb/memcached/commit/a4888446) Publish audit events (#296)
- [236d6108](https://github.com/kubedb/memcached/commit/236d6108) Use kglog helper
- [7ffe5c73](https://github.com/kubedb/memcached/commit/7ffe5c73) Use klog/v2
- [fb34645b](https://github.com/kubedb/memcached/commit/fb34645b) Update Kubernetes toolchain to v1.21.0 (#294)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.12.0](https://github.com/kubedb/mongodb/releases/tag/v0.12.0)

- [06b04a8c](https://github.com/kubedb/mongodb/commit/06b04a8c) Prepare for release v0.12.0 (#402)
- [ae4e0cd1](https://github.com/kubedb/mongodb/commit/ae4e0cd1) Fix mongodb exporter error (#401)
- [11eb6ee8](https://github.com/kubedb/mongodb/commit/11eb6ee8) Prepare for release v0.12.0-rc.0 (#400)
- [dbf5cd16](https://github.com/kubedb/mongodb/commit/dbf5cd16) Update audit lib (#399)
- [a55bf1d5](https://github.com/kubedb/mongodb/commit/a55bf1d5) Limit go routine in health check (#394)
- [0a61c733](https://github.com/kubedb/mongodb/commit/0a61c733) Update TLS args for Exporter (#395)
- [80d3fec2](https://github.com/kubedb/mongodb/commit/80d3fec2) Send audit events if analytics enabled (#398)
- [8ac51d7e](https://github.com/kubedb/mongodb/commit/8ac51d7e) Create auditor if license file is provided (#397)
- [c6c4b380](https://github.com/kubedb/mongodb/commit/c6c4b380) Publish audit events (#396)
- [e261937a](https://github.com/kubedb/mongodb/commit/e261937a) Fix log level issue with klog (#393)
- [426afbfc](https://github.com/kubedb/mongodb/commit/426afbfc) Use kglog helper
- [24b7976c](https://github.com/kubedb/mongodb/commit/24b7976c) Use klog/v2
- [0ace005d](https://github.com/kubedb/mongodb/commit/0ace005d) Update Kubernetes toolchain to v1.21.0 (#391)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.12.0](https://github.com/kubedb/mysql/releases/tag/v0.12.0)

- [5fb0bf79](https://github.com/kubedb/mysql/commit/5fb0bf79) Prepare for release v0.12.0 (#393)
- [9533c528](https://github.com/kubedb/mysql/commit/9533c528) Prepare for release v0.12.0-rc.0 (#392)
- [f0313b17](https://github.com/kubedb/mysql/commit/f0313b17) Limit Health Checker goroutines (#385)
- [ab601a28](https://github.com/kubedb/mysql/commit/ab601a28) Use gomodules.xyz/password-generator v0.2.7
- [782362db](https://github.com/kubedb/mysql/commit/782362db) Update audit library (#390)
- [1d36bacb](https://github.com/kubedb/mysql/commit/1d36bacb) Send audit events if analytics enabled (#389)
- [55a903a3](https://github.com/kubedb/mysql/commit/55a903a3) Create auditor if license file is provided (#388)
- [dc6f6ea5](https://github.com/kubedb/mysql/commit/dc6f6ea5) Publish audit events (#387)
- [75bd1a1c](https://github.com/kubedb/mysql/commit/75bd1a1c) Fix log level issue with klog for mysql (#386)
- [1014a393](https://github.com/kubedb/mysql/commit/1014a393) Use kglog helper
- [728fa299](https://github.com/kubedb/mysql/commit/728fa299) Use klog/v2
- [80581df4](https://github.com/kubedb/mysql/commit/80581df4) Update Kubernetes toolchain to v1.21.0 (#383)



## [kubedb/operator](https://github.com/kubedb/operator)

### [v0.19.0](https://github.com/kubedb/operator/releases/tag/v0.19.0)

- [cd2b14ca](https://github.com/kubedb/operator/commit/cd2b14ca) Prepare for release v0.19.0 (#409)
- [e48d8929](https://github.com/kubedb/operator/commit/e48d8929) Disable api priority and fairness feature for webhook server (#408)
- [08daa22a](https://github.com/kubedb/operator/commit/08daa22a) Prepare for release v0.19.0-rc.0 (#407)
- [203ffa38](https://github.com/kubedb/operator/commit/203ffa38) Update audit lib (#406)
- [704a774f](https://github.com/kubedb/operator/commit/704a774f) Send audit events if analytics enabled (#405)
- [7e8f1be0](https://github.com/kubedb/operator/commit/7e8f1be0) Stop using gomodules.xyz/version
- [49d7d7f2](https://github.com/kubedb/operator/commit/49d7d7f2) Publish audit events (#404)
- [820d7372](https://github.com/kubedb/operator/commit/820d7372) Use kglog helper
- [396ae75f](https://github.com/kubedb/operator/commit/396ae75f) Update Kubernetes toolchain to v1.21.0 (#403)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.6.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.6.0)

- [318d259d](https://github.com/kubedb/percona-xtradb/commit/318d259d) Prepare for release v0.6.0 (#203)
- [181a7d71](https://github.com/kubedb/percona-xtradb/commit/181a7d71) Disable api priority and fairness feature for webhook server (#202)
- [870e08df](https://github.com/kubedb/percona-xtradb/commit/870e08df) Prepare for release v0.6.0-rc.0 (#201)
- [f163f637](https://github.com/kubedb/percona-xtradb/commit/f163f637) Update audit lib (#200)
- [c42c3401](https://github.com/kubedb/percona-xtradb/commit/c42c3401) Send audit events if analytics enabled (#199)
- [e2ce3664](https://github.com/kubedb/percona-xtradb/commit/e2ce3664) Create auditor if license file is provided (#198)
- [3e85edb2](https://github.com/kubedb/percona-xtradb/commit/3e85edb2) Publish audit events (#197)
- [6f23031c](https://github.com/kubedb/percona-xtradb/commit/6f23031c) Use kglog helper
- [cc0e270a](https://github.com/kubedb/percona-xtradb/commit/cc0e270a) Use klog/v2
- [a44e3347](https://github.com/kubedb/percona-xtradb/commit/a44e3347) Update Kubernetes toolchain to v1.21.0 (#195)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.3.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.3.0)

- [d0c24fa](https://github.com/kubedb/pg-coordinator/commit/d0c24fa) Prepare for release v0.3.0 (#26)
- [3ca5f67](https://github.com/kubedb/pg-coordinator/commit/3ca5f67) Prepare for release v0.3.0-rc.0 (#25)
- [4ef7d95](https://github.com/kubedb/pg-coordinator/commit/4ef7d95) Update Client TLS Path for Postgres (#24)
- [7208199](https://github.com/kubedb/pg-coordinator/commit/7208199) Raft Version Update And Ops Request Fix (#23)
- [5adb304](https://github.com/kubedb/pg-coordinator/commit/5adb304) Use klog/v2 (#19)
- [a9b3f16](https://github.com/kubedb/pg-coordinator/commit/a9b3f16) Use klog/v2



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.6.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.6.0)

- [3ab2f55a](https://github.com/kubedb/pgbouncer/commit/3ab2f55a) Prepare for release v0.6.0 (#163)
- [ae89b9a6](https://github.com/kubedb/pgbouncer/commit/ae89b9a6) Disable api priority and fairness feature for webhook server (#162)
- [cbba6969](https://github.com/kubedb/pgbouncer/commit/cbba6969) Prepare for release v0.6.0-rc.0 (#161)
- [bc6428cd](https://github.com/kubedb/pgbouncer/commit/bc6428cd) Update audit lib (#160)
- [442f0635](https://github.com/kubedb/pgbouncer/commit/442f0635) Send audit events if analytics enabled (#159)
- [2ebaf4bb](https://github.com/kubedb/pgbouncer/commit/2ebaf4bb) Create auditor if license file is provided (#158)
- [4e3f115d](https://github.com/kubedb/pgbouncer/commit/4e3f115d) Publish audit events (#157)
- [1ed2f883](https://github.com/kubedb/pgbouncer/commit/1ed2f883) Use kglog helper
- [870cf108](https://github.com/kubedb/pgbouncer/commit/870cf108) Use klog/v2
- [11c2ac03](https://github.com/kubedb/pgbouncer/commit/11c2ac03) Update Kubernetes toolchain to v1.21.0 (#155)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.19.0](https://github.com/kubedb/postgres/releases/tag/v0.19.0)

- [d10c5e40](https://github.com/kubedb/postgres/commit/d10c5e40) Prepare for release v0.19.0 (#509)
- [06fcab6e](https://github.com/kubedb/postgres/commit/06fcab6e) Prepare for release v0.19.0-rc.0 (#508)
- [5c0e0fa2](https://github.com/kubedb/postgres/commit/5c0e0fa2) Run All DB Pod's Container with Custom-UID (#507)
- [9496dadf](https://github.com/kubedb/postgres/commit/9496dadf) Update audit lib (#506)
- [d51cdfdd](https://github.com/kubedb/postgres/commit/d51cdfdd) Limit Health Check for Postgres (#504)
- [24851ba8](https://github.com/kubedb/postgres/commit/24851ba8) Send audit events if analytics enabled (#505)
- [faecf01d](https://github.com/kubedb/postgres/commit/faecf01d) Create auditor if license file is provided (#503)
- [8d4bf26b](https://github.com/kubedb/postgres/commit/8d4bf26b) Stop using gomodules.xyz/version (#501)
- [906c678e](https://github.com/kubedb/postgres/commit/906c678e) Publish audit events (#500)
- [c6afe209](https://github.com/kubedb/postgres/commit/c6afe209) Fix: Log Level Issue with klog (#496)
- [2a910034](https://github.com/kubedb/postgres/commit/2a910034) Use kglog helper
- [a4e685d6](https://github.com/kubedb/postgres/commit/a4e685d6) Use klog/v2
- [ee9a9d15](https://github.com/kubedb/postgres/commit/ee9a9d15) Update Kubernetes toolchain to v1.21.0 (#492)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.6.0](https://github.com/kubedb/proxysql/releases/tag/v0.6.0)

- [08e892e2](https://github.com/kubedb/proxysql/commit/08e892e2) Prepare for release v0.6.0 (#181)
- [ecd32aea](https://github.com/kubedb/proxysql/commit/ecd32aea) Disable api priority and fairness feature for webhook server (#180)
- [ba5ec48b](https://github.com/kubedb/proxysql/commit/ba5ec48b) Prepare for release v0.6.0-rc.0 (#179)
- [9770fa0d](https://github.com/kubedb/proxysql/commit/9770fa0d) Update audit lib (#178)
- [3e307411](https://github.com/kubedb/proxysql/commit/3e307411) Send audit events if analytics enabled (#177)
- [790b57ed](https://github.com/kubedb/proxysql/commit/790b57ed) Create auditor if license file is provided (#176)
- [6e6c9ba1](https://github.com/kubedb/proxysql/commit/6e6c9ba1) Publish audit events (#175)
- [df2937ed](https://github.com/kubedb/proxysql/commit/df2937ed) Use kglog helper
- [2ca12e48](https://github.com/kubedb/proxysql/commit/2ca12e48) Use klog/v2
- [3796f730](https://github.com/kubedb/proxysql/commit/3796f730) Update Kubernetes toolchain to v1.21.0 (#173)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.12.0](https://github.com/kubedb/redis/releases/tag/v0.12.0)

- [a29ff99d](https://github.com/kubedb/redis/commit/a29ff99d) Prepare for release v0.12.0 (#326)
- [a1392dee](https://github.com/kubedb/redis/commit/a1392dee) Disable api priority and fairness feature for webhook server (#325)
- [0c15054c](https://github.com/kubedb/redis/commit/0c15054c) Prepare for release v0.12.0-rc.0 (#324)
- [5a5ec318](https://github.com/kubedb/redis/commit/5a5ec318) Update audit lib (#323)
- [6673f940](https://github.com/kubedb/redis/commit/6673f940) Limit Health Check go-routine Redis (#321)
- [e945029e](https://github.com/kubedb/redis/commit/e945029e) Send audit events if analytics enabled (#322)
- [3715ff10](https://github.com/kubedb/redis/commit/3715ff10) Create auditor if license file is provided (#320)
- [9d5d90a9](https://github.com/kubedb/redis/commit/9d5d90a9) Add auditor handler
- [5004f56c](https://github.com/kubedb/redis/commit/5004f56c) Publish audit events (#319)
- [146b3863](https://github.com/kubedb/redis/commit/146b3863) Use kglog helper
- [71d8ced8](https://github.com/kubedb/redis/commit/71d8ced8) Use klog/v2
- [4900a564](https://github.com/kubedb/redis/commit/4900a564) Update Kubernetes toolchain to v1.21.0 (#317)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.6.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.6.0)

- [c1af00f](https://github.com/kubedb/replication-mode-detector/commit/c1af00f) Prepare for release v0.6.0 (#144)
- [1382382](https://github.com/kubedb/replication-mode-detector/commit/1382382) Prepare for release v0.6.0-rc.0 (#143)
- [feba070](https://github.com/kubedb/replication-mode-detector/commit/feba070) Remove glog dependency
- [fd757b4](https://github.com/kubedb/replication-mode-detector/commit/fd757b4) Use kglog helper
- [8ba20a3](https://github.com/kubedb/replication-mode-detector/commit/8ba20a3) Update repository config (#142)
- [eece885](https://github.com/kubedb/replication-mode-detector/commit/eece885) Use klog/v2
- [e30c050](https://github.com/kubedb/replication-mode-detector/commit/e30c050) Use Kubernetes v1.21.0 toolchain (#140)
- [8e7b7c2](https://github.com/kubedb/replication-mode-detector/commit/8e7b7c2) Use Kubernetes v1.21.0 toolchain (#139)
- [6bceb2f](https://github.com/kubedb/replication-mode-detector/commit/6bceb2f) Use Kubernetes v1.21.0 toolchain (#138)
- [0fe720e](https://github.com/kubedb/replication-mode-detector/commit/0fe720e) Use Kubernetes v1.21.0 toolchain (#137)
- [8c54b2a](https://github.com/kubedb/replication-mode-detector/commit/8c54b2a) Update Kubernetes toolchain to v1.21.0 (#136)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.4.0](https://github.com/kubedb/tests/releases/tag/v0.4.0)

- [c6f1adc](https://github.com/kubedb/tests/commit/c6f1adc) Prepare for release v0.4.0 (#125)
- [b6b4be3](https://github.com/kubedb/tests/commit/b6b4be3) Prepare for release v0.4.0-rc.0 (#124)
- [62e6b50](https://github.com/kubedb/tests/commit/62e6b50) Fix locking in ResourceMapper (#123)
- [a855fab](https://github.com/kubedb/tests/commit/a855fab) Update dependencies (#122)
- [7d5b1a4](https://github.com/kubedb/tests/commit/7d5b1a4) Use kglog helper
- [a08eee4](https://github.com/kubedb/tests/commit/a08eee4) Use klog/v2
- [ed1afd4](https://github.com/kubedb/tests/commit/ed1afd4) Use Kubernetes v1.21.0 toolchain (#120)
- [ccb54f1](https://github.com/kubedb/tests/commit/ccb54f1) Use Kubernetes v1.21.0 toolchain (#119)
- [2a6f06d](https://github.com/kubedb/tests/commit/2a6f06d) Use Kubernetes v1.21.0 toolchain (#118)
- [7fb99f7](https://github.com/kubedb/tests/commit/7fb99f7) Use Kubernetes v1.21.0 toolchain (#117)
- [aaa0647](https://github.com/kubedb/tests/commit/aaa0647) Update Kubernetes toolchain to v1.21.0 (#116)
- [79d815d](https://github.com/kubedb/tests/commit/79d815d) Fix Elasticsearch status check while creating the client (#114)




