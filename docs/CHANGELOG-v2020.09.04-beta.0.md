---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2020.09.04-beta.0
    name: Changelog-v2020.09.04-beta.0
    parent: welcome
    weight: 20200904
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2020.09.04-beta.0/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2020.09.04-beta.0/
---

# KubeDB v2020.09.04-beta.0 (2020-09-04)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.14.0-beta.2](https://github.com/kubedb/apimachinery/releases/tag/v0.14.0-beta.2)

- [76ac9bc0](https://github.com/kubedb/apimachinery/commit/76ac9bc0) Remove CertManagerClient client
- [b99048f4](https://github.com/kubedb/apimachinery/commit/b99048f4) Remove unused constants for ProxySQL
- [152cef57](https://github.com/kubedb/apimachinery/commit/152cef57) Update Kubernetes v1.18.3 dependencies (#578)
- [24c5e829](https://github.com/kubedb/apimachinery/commit/24c5e829) Update redis constants (#575)
- [7075b38d](https://github.com/kubedb/apimachinery/commit/7075b38d) Remove spec.updateStrategy field (#577)
- [dfd11955](https://github.com/kubedb/apimachinery/commit/dfd11955) Remove description from CRD yamls (#576)
- [2d1b5878](https://github.com/kubedb/apimachinery/commit/2d1b5878) Add autoscaling crds (#554)
- [68ed8127](https://github.com/kubedb/apimachinery/commit/68ed8127) Fix build
- [63d18f0d](https://github.com/kubedb/apimachinery/commit/63d18f0d) Rename PgBouncer archiver to client
- [a219c251](https://github.com/kubedb/apimachinery/commit/a219c251) Handle shard scenario for MongoDB cert names (#574)
- [d2c80e55](https://github.com/kubedb/apimachinery/commit/d2c80e55) Add MongoDB Custom Config Spec (#562)
- [1e69fb02](https://github.com/kubedb/apimachinery/commit/1e69fb02) Support multiple certificates per DB (#555)
- [9bbed3d1](https://github.com/kubedb/apimachinery/commit/9bbed3d1) Update Kubernetes v1.18.3 dependencies (#573)
- [7df78c7a](https://github.com/kubedb/apimachinery/commit/7df78c7a) Update CRD yamls
- [406d895d](https://github.com/kubedb/apimachinery/commit/406d895d) Implement ServiceMonitorAdditionalLabels method (#572)
- [cfe4374a](https://github.com/kubedb/apimachinery/commit/cfe4374a) Make ServiceMonitor name same as stats service (#563)
- [d2ed6b4a](https://github.com/kubedb/apimachinery/commit/d2ed6b4a) Update for release Stash@v2020.08.27 (#571)
- [749b9084](https://github.com/kubedb/apimachinery/commit/749b9084) Update for release Stash@v2020.08.27-rc.0 (#570)
- [5d8bf42c](https://github.com/kubedb/apimachinery/commit/5d8bf42c) Update for release Stash@v2020.08.26-rc.1 (#569)
- [6edc4782](https://github.com/kubedb/apimachinery/commit/6edc4782) Update for release Stash@v2020.08.26-rc.0 (#568)
- [c451ff3a](https://github.com/kubedb/apimachinery/commit/c451ff3a) Update Kubernetes v1.18.3 dependencies (#565)
- [fdc6e2d6](https://github.com/kubedb/apimachinery/commit/fdc6e2d6) Update Kubernetes v1.18.3 dependencies (#564)
- [2f509c26](https://github.com/kubedb/apimachinery/commit/2f509c26) Update Kubernetes v1.18.3 dependencies (#561)
- [da655afe](https://github.com/kubedb/apimachinery/commit/da655afe) Update Kubernetes v1.18.3 dependencies (#560)
- [9c2c06a9](https://github.com/kubedb/apimachinery/commit/9c2c06a9) Fix MySQL enterprise condition's  constant (#559)
- [81ed2724](https://github.com/kubedb/apimachinery/commit/81ed2724) Update Kubernetes v1.18.3 dependencies (#558)
- [738b7ade](https://github.com/kubedb/apimachinery/commit/738b7ade) Update Kubernetes v1.18.3 dependencies (#557)
- [93f0af4b](https://github.com/kubedb/apimachinery/commit/93f0af4b) Add MySQL Constants (#553)
- [6049554d](https://github.com/kubedb/apimachinery/commit/6049554d) Add {Horizontal,Vertical}ScalingSpec for Redis (#534)
- [28552272](https://github.com/kubedb/apimachinery/commit/28552272) Enable TLS for Redis (#546)
- [68e00844](https://github.com/kubedb/apimachinery/commit/68e00844) Add Spec for MongoDB Volume Expansion (#548)
- [759a800a](https://github.com/kubedb/apimachinery/commit/759a800a) Add Subject spec for Certificate (#552)
- [b1552628](https://github.com/kubedb/apimachinery/commit/b1552628) Add email SANs for certificate (#551)
- [fdfad57e](https://github.com/kubedb/apimachinery/commit/fdfad57e) Update to cert-manager@v0.16.0 (#550)
- [3b5e9ece](https://github.com/kubedb/apimachinery/commit/3b5e9ece) Update to Kubernetes v1.18.3 (#549)
- [0c5a1e9b](https://github.com/kubedb/apimachinery/commit/0c5a1e9b) Make ElasticsearchVersion spec.tools optional (#526)
- [01a0b4b3](https://github.com/kubedb/apimachinery/commit/01a0b4b3) Add Conditions Constant for MongoDBOpsRequest (#535)
- [34a9ed61](https://github.com/kubedb/apimachinery/commit/34a9ed61) Update to Kubernetes v1.18.3 (#547)
- [6392f19e](https://github.com/kubedb/apimachinery/commit/6392f19e) Add Storage Engine Support for Percona Server MongoDB (#538)
- [02d205bc](https://github.com/kubedb/apimachinery/commit/02d205bc) Remove extra - from prefix/suffix (#543)
- [06158f51](https://github.com/kubedb/apimachinery/commit/06158f51) Update to Kubernetes v1.18.3 (#542)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.14.0-beta.2](https://github.com/kubedb/cli/releases/tag/v0.14.0-beta.2)

- [58b39094](https://github.com/kubedb/cli/commit/58b39094) Prepare for release v0.14.0-beta.2 (#484)
- [0f8819ce](https://github.com/kubedb/cli/commit/0f8819ce) Update Kubernetes v1.18.3 dependencies (#483)
- [86a92381](https://github.com/kubedb/cli/commit/86a92381) Update Kubernetes v1.18.3 dependencies (#482)
- [05e5cef2](https://github.com/kubedb/cli/commit/05e5cef2) Update for release Stash@v2020.08.27 (#481)
- [b1aa1dc2](https://github.com/kubedb/cli/commit/b1aa1dc2) Update for release Stash@v2020.08.27-rc.0 (#480)
- [36716efc](https://github.com/kubedb/cli/commit/36716efc) Update for release Stash@v2020.08.26-rc.1 (#479)
- [a30f21e0](https://github.com/kubedb/cli/commit/a30f21e0) Update for release Stash@v2020.08.26-rc.0 (#478)
- [836d6227](https://github.com/kubedb/cli/commit/836d6227) Update Kubernetes v1.18.3 dependencies (#477)
- [8a81d715](https://github.com/kubedb/cli/commit/8a81d715) Update Kubernetes v1.18.3 dependencies (#476)
- [7ce2101d](https://github.com/kubedb/cli/commit/7ce2101d) Update Kubernetes v1.18.3 dependencies (#475)
- [3c617e66](https://github.com/kubedb/cli/commit/3c617e66) Update Kubernetes v1.18.3 dependencies (#474)
- [f70b2ba4](https://github.com/kubedb/cli/commit/f70b2ba4) Update Kubernetes v1.18.3 dependencies (#473)
- [ba77ba2b](https://github.com/kubedb/cli/commit/ba77ba2b) Update Kubernetes v1.18.3 dependencies (#472)
- [b296035f](https://github.com/kubedb/cli/commit/b296035f) Use actions/upload-artifact@v2
- [7bb95619](https://github.com/kubedb/cli/commit/7bb95619) Update to Kubernetes v1.18.3 (#471)
- [6e5789a2](https://github.com/kubedb/cli/commit/6e5789a2) Update to Kubernetes v1.18.3 (#470)
- [9d550ebc](https://github.com/kubedb/cli/commit/9d550ebc) Update to Kubernetes v1.18.3 (#469)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.14.0-beta.2](https://github.com/kubedb/elasticsearch/releases/tag/v0.14.0-beta.2)

- [3b83c316](https://github.com/kubedb/elasticsearch/commit/3b83c316) Prepare for release v0.14.0-beta.2 (#339)
- [662823ae](https://github.com/kubedb/elasticsearch/commit/662823ae) Update release.yml
- [ada6c2d3](https://github.com/kubedb/elasticsearch/commit/ada6c2d3) Add support for Open-Distro-for-Elasticsearch (#303)
- [a9c7ba33](https://github.com/kubedb/elasticsearch/commit/a9c7ba33) Update Kubernetes v1.18.3 dependencies (#333)
- [c67b1290](https://github.com/kubedb/elasticsearch/commit/c67b1290) Update Kubernetes v1.18.3 dependencies (#332)
- [aa1d64ad](https://github.com/kubedb/elasticsearch/commit/aa1d64ad) Update Kubernetes v1.18.3 dependencies (#331)
- [3d6c3e91](https://github.com/kubedb/elasticsearch/commit/3d6c3e91) Update Kubernetes v1.18.3 dependencies (#330)
- [bb318e74](https://github.com/kubedb/elasticsearch/commit/bb318e74) Update Kubernetes v1.18.3 dependencies (#329)
- [6b6b4d2d](https://github.com/kubedb/elasticsearch/commit/6b6b4d2d) Update Kubernetes v1.18.3 dependencies (#328)
- [06cef782](https://github.com/kubedb/elasticsearch/commit/06cef782) Remove dependency on enterprise operator (#327)
- [20a2c7d4](https://github.com/kubedb/elasticsearch/commit/20a2c7d4) Update to cert-manager v0.16.0 (#326)
- [e767c356](https://github.com/kubedb/elasticsearch/commit/e767c356) Build images in e2e workflow (#325)
- [ae696dbe](https://github.com/kubedb/elasticsearch/commit/ae696dbe) Update to Kubernetes v1.18.3 (#324)
- [a511d8d6](https://github.com/kubedb/elasticsearch/commit/a511d8d6) Allow configuring k8s & db version in e2e tests (#323)
- [a50b503d](https://github.com/kubedb/elasticsearch/commit/a50b503d) Trigger e2e tests on /ok-to-test command (#322)
- [107faff2](https://github.com/kubedb/elasticsearch/commit/107faff2) Update to Kubernetes v1.18.3 (#321)
- [60fb6d9b](https://github.com/kubedb/elasticsearch/commit/60fb6d9b) Update to Kubernetes v1.18.3 (#320)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v0.14.0-beta.2](https://github.com/kubedb/installer/releases/tag/v0.14.0-beta.2)

- [cb0e278](https://github.com/kubedb/installer/commit/cb0e278) Prepare for release v0.14.0-beta.2 (#128)
- [b31ccbf](https://github.com/kubedb/installer/commit/b31ccbf) Update Kubernetes v1.18.3 dependencies (#127)
- [389ce6a](https://github.com/kubedb/installer/commit/389ce6a) Update Kubernetes v1.18.3 dependencies (#126)
- [db6f1e9](https://github.com/kubedb/installer/commit/db6f1e9) Update chart icons
- [9f41f2d](https://github.com/kubedb/installer/commit/9f41f2d) Update Kubernetes v1.18.3 dependencies (#124)
- [004373e](https://github.com/kubedb/installer/commit/004373e) Update Kubernetes v1.18.3 dependencies (#123)
- [e517626](https://github.com/kubedb/installer/commit/e517626) Prefix catalog files with non-patched versions deprecated- (#119)
- [2bf8715](https://github.com/kubedb/installer/commit/2bf8715) Update Kubernetes v1.18.3 dependencies (#121)
- [9a5cc7b](https://github.com/kubedb/installer/commit/9a5cc7b) Update Kubernetes v1.18.3 dependencies (#120)
- [e2f8ebd](https://github.com/kubedb/installer/commit/e2f8ebd) Add MySQL New catalog (#116)
- [72ad85e](https://github.com/kubedb/installer/commit/72ad85e) Update Kubernetes v1.18.3 dependencies (#118)
- [94ebcb2](https://github.com/kubedb/installer/commit/94ebcb2) Update Kubernetes v1.18.3 dependencies (#117)
- [5dc2808](https://github.com/kubedb/installer/commit/5dc2808) Remove excess permission (#115)
- [65b4443](https://github.com/kubedb/installer/commit/65b4443) Update redis exporter image tag
- [7191679](https://github.com/kubedb/installer/commit/7191679) Add support for Redis 6.0.6 (#99)
- [902f00e](https://github.com/kubedb/installer/commit/902f00e) Add Pod `exec` permission in ClusterRole (#102)
- [4a83599](https://github.com/kubedb/installer/commit/4a83599) Update to Kubernetes v1.18.3 (#114)
- [df8412a](https://github.com/kubedb/installer/commit/df8412a) Add Permissions for PVC (#112)
- [99d6e66](https://github.com/kubedb/installer/commit/99d6e66) Update elasticsearchversion crds (#111)
- [57561a3](https://github.com/kubedb/installer/commit/57561a3) Use `percona` as Suffix in MongoDBVersion Name (#110)
- [7706f93](https://github.com/kubedb/installer/commit/7706f93) Update to Kubernetes v1.18.3 (#109)
- [513db6d](https://github.com/kubedb/installer/commit/513db6d) Add Percona MongoDB Server Catalogs (#103)
- [2b10a12](https://github.com/kubedb/installer/commit/2b10a12) Update to Kubernetes v1.18.3 (#108)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.7.0-beta.2](https://github.com/kubedb/memcached/releases/tag/v0.7.0-beta.2)

- [b8fe927b](https://github.com/kubedb/memcached/commit/b8fe927b) Prepare for release v0.7.0-beta.2 (#177)
- [0f5014d2](https://github.com/kubedb/memcached/commit/0f5014d2) Update release.yml
- [1b627013](https://github.com/kubedb/memcached/commit/1b627013) Remove updateStrategy field (#176)
- [66f008d6](https://github.com/kubedb/memcached/commit/66f008d6) Update Kubernetes v1.18.3 dependencies (#175)
- [09ff8589](https://github.com/kubedb/memcached/commit/09ff8589) Update Kubernetes v1.18.3 dependencies (#174)
- [92e344d8](https://github.com/kubedb/memcached/commit/92e344d8) Update Kubernetes v1.18.3 dependencies (#173)
- [51e977f3](https://github.com/kubedb/memcached/commit/51e977f3) Update Kubernetes v1.18.3 dependencies (#172)
- [f32d7e9c](https://github.com/kubedb/memcached/commit/f32d7e9c) Update Kubernetes v1.18.3 dependencies (#171)
- [2cdba698](https://github.com/kubedb/memcached/commit/2cdba698) Update Kubernetes v1.18.3 dependencies (#170)
- [9486876e](https://github.com/kubedb/memcached/commit/9486876e) Update Kubernetes v1.18.3 dependencies (#169)
- [81648447](https://github.com/kubedb/memcached/commit/81648447) Update Kubernetes v1.18.3 dependencies (#168)
- [e9c3f98d](https://github.com/kubedb/memcached/commit/e9c3f98d) Fix install target
- [6dff8f7b](https://github.com/kubedb/memcached/commit/6dff8f7b) Remove dependency on enterprise operator (#167)
- [707d4d83](https://github.com/kubedb/memcached/commit/707d4d83) Build images in e2e workflow (#166)
- [ff1b144e](https://github.com/kubedb/memcached/commit/ff1b144e) Allow configuring k8s & db version in e2e tests (#165)
- [0b1699d8](https://github.com/kubedb/memcached/commit/0b1699d8) Update to Kubernetes v1.18.3 (#164)
- [b141122a](https://github.com/kubedb/memcached/commit/b141122a) Trigger e2e tests on /ok-to-test command (#163)
- [36b03266](https://github.com/kubedb/memcached/commit/36b03266) Update to Kubernetes v1.18.3 (#162)
- [3ede9dcc](https://github.com/kubedb/memcached/commit/3ede9dcc) Update to Kubernetes v1.18.3 (#161)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.7.0-beta.2](https://github.com/kubedb/mongodb/releases/tag/v0.7.0-beta.2)

- [8fd389de](https://github.com/kubedb/mongodb/commit/8fd389de) Prepare for release v0.7.0-beta.2 (#234)
- [3e4981ee](https://github.com/kubedb/mongodb/commit/3e4981ee) Update release.yml
- [c1d5cdb8](https://github.com/kubedb/mongodb/commit/c1d5cdb8) Always use OnDelete UpdateStrategy (#233)
- [a135b2c7](https://github.com/kubedb/mongodb/commit/a135b2c7) Fix build (#232)
- [cfb1788b](https://github.com/kubedb/mongodb/commit/cfb1788b) Use updated certificate spec (#221)
- [486e820a](https://github.com/kubedb/mongodb/commit/486e820a) Remove `storage` Validation Check (#231)
- [12e621ed](https://github.com/kubedb/mongodb/commit/12e621ed) Update Kubernetes v1.18.3 dependencies (#225)
- [0d7ea7d7](https://github.com/kubedb/mongodb/commit/0d7ea7d7) Update Kubernetes v1.18.3 dependencies (#224)
- [e79d1dfe](https://github.com/kubedb/mongodb/commit/e79d1dfe) Update Kubernetes v1.18.3 dependencies (#223)
- [d0ff5e1d](https://github.com/kubedb/mongodb/commit/d0ff5e1d) Update Kubernetes v1.18.3 dependencies (#222)
- [d22ade32](https://github.com/kubedb/mongodb/commit/d22ade32) Add `inMemory` Storage Engine Support for Percona MongoDB Server (#205)
- [90847996](https://github.com/kubedb/mongodb/commit/90847996) Update Kubernetes v1.18.3 dependencies (#220)
- [1098974f](https://github.com/kubedb/mongodb/commit/1098974f) Update Kubernetes v1.18.3 dependencies (#219)
- [e7d1407a](https://github.com/kubedb/mongodb/commit/e7d1407a) Fix install target
- [a5742d11](https://github.com/kubedb/mongodb/commit/a5742d11) Remove dependency on enterprise operator (#218)
- [1de4fbee](https://github.com/kubedb/mongodb/commit/1de4fbee) Build images in e2e workflow (#217)
- [b736c57e](https://github.com/kubedb/mongodb/commit/b736c57e) Update to Kubernetes v1.18.3 (#216)
- [180ae28d](https://github.com/kubedb/mongodb/commit/180ae28d) Allow configuring k8s & db version in e2e tests (#215)
- [c2f09a6f](https://github.com/kubedb/mongodb/commit/c2f09a6f) Trigger e2e tests on /ok-to-test command (#214)
- [c1c7fa39](https://github.com/kubedb/mongodb/commit/c1c7fa39) Update to Kubernetes v1.18.3 (#213)
- [8fb6cf78](https://github.com/kubedb/mongodb/commit/8fb6cf78) Update to Kubernetes v1.18.3 (#212)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.7.0-beta.2](https://github.com/kubedb/mysql/releases/tag/v0.7.0-beta.2)

- [6010c034](https://github.com/kubedb/mysql/commit/6010c034) Prepare for release v0.7.0-beta.2 (#224)
- [4b530066](https://github.com/kubedb/mysql/commit/4b530066) Update release.yml
- [184a6cbc](https://github.com/kubedb/mysql/commit/184a6cbc) Update dependencies (#223)
- [903b13b6](https://github.com/kubedb/mysql/commit/903b13b6) Always use OnDelete update strategy
- [1c10224a](https://github.com/kubedb/mysql/commit/1c10224a) Update Kubernetes v1.18.3 dependencies (#222)
- [4e9e5e44](https://github.com/kubedb/mysql/commit/4e9e5e44) Added TLS/SSL Configuration in MySQL Server (#204)
- [d08209b8](https://github.com/kubedb/mysql/commit/d08209b8) Use username/password constants from core/v1
- [87238c42](https://github.com/kubedb/mysql/commit/87238c42) Update MySQL vendor for changes of prometheus coreos operator (#216)
- [999005ed](https://github.com/kubedb/mysql/commit/999005ed) Update Kubernetes v1.18.3 dependencies (#215)
- [3eb5086e](https://github.com/kubedb/mysql/commit/3eb5086e) Update Kubernetes v1.18.3 dependencies (#214)
- [cd58f276](https://github.com/kubedb/mysql/commit/cd58f276) Update Kubernetes v1.18.3 dependencies (#213)
- [4dcfcd14](https://github.com/kubedb/mysql/commit/4dcfcd14) Update Kubernetes v1.18.3 dependencies (#212)
- [d41015c9](https://github.com/kubedb/mysql/commit/d41015c9) Update Kubernetes v1.18.3 dependencies (#211)
- [4350cb79](https://github.com/kubedb/mysql/commit/4350cb79) Update Kubernetes v1.18.3 dependencies (#210)
- [617af851](https://github.com/kubedb/mysql/commit/617af851) Fix install target
- [fc308cc3](https://github.com/kubedb/mysql/commit/fc308cc3) Remove dependency on enterprise operator (#209)
- [1b717aee](https://github.com/kubedb/mysql/commit/1b717aee) Detect primary pod in MySQL group replication (#190)
- [c3e516f4](https://github.com/kubedb/mysql/commit/c3e516f4) Support MySQL new version for group replication and standalone (#189)
- [8bedade3](https://github.com/kubedb/mysql/commit/8bedade3) Build images in e2e workflow (#208)
- [02c9434c](https://github.com/kubedb/mysql/commit/02c9434c) Allow configuring k8s & db version in e2e tests (#207)
- [ae5d757c](https://github.com/kubedb/mysql/commit/ae5d757c) Update to Kubernetes v1.18.3 (#206)
- [16bdc23f](https://github.com/kubedb/mysql/commit/16bdc23f) Trigger e2e tests on /ok-to-test command (#205)
- [7be13878](https://github.com/kubedb/mysql/commit/7be13878) Update to Kubernetes v1.18.3 (#203)
- [d69fe478](https://github.com/kubedb/mysql/commit/d69fe478) Update to Kubernetes v1.18.3 (#202)



## [kubedb/mysql-replication-mode-detector](https://github.com/kubedb/mysql-replication-mode-detector)

### [v0.1.0-beta.2](https://github.com/kubedb/mysql-replication-mode-detector/releases/tag/v0.1.0-beta.2)

- [eb878dc](https://github.com/kubedb/mysql-replication-mode-detector/commit/eb878dc) Prepare for release v0.1.0-beta.2 (#21)
- [6c214b8](https://github.com/kubedb/mysql-replication-mode-detector/commit/6c214b8) Update Kubernetes v1.18.3 dependencies (#19)
- [00800e8](https://github.com/kubedb/mysql-replication-mode-detector/commit/00800e8) Update Kubernetes v1.18.3 dependencies (#18)
- [373ab6d](https://github.com/kubedb/mysql-replication-mode-detector/commit/373ab6d) Update Kubernetes v1.18.3 dependencies (#17)
- [8b61313](https://github.com/kubedb/mysql-replication-mode-detector/commit/8b61313) Update Kubernetes v1.18.3 dependencies (#16)
- [f2a68e3](https://github.com/kubedb/mysql-replication-mode-detector/commit/f2a68e3) Update Kubernetes v1.18.3 dependencies (#15)
- [3bce396](https://github.com/kubedb/mysql-replication-mode-detector/commit/3bce396) Update Kubernetes v1.18.3 dependencies (#14)
- [32603a2](https://github.com/kubedb/mysql-replication-mode-detector/commit/32603a2) Don't push binary with release



## [kubedb/operator](https://github.com/kubedb/operator)

### [v0.14.0-beta.2](https://github.com/kubedb/operator/releases/tag/v0.14.0-beta.2)

- [a13ca48b](https://github.com/kubedb/operator/commit/a13ca48b) Prepare for release v0.14.0-beta.2 (#281)
- [fc6c1e9e](https://github.com/kubedb/operator/commit/fc6c1e9e) Update Kubernetes v1.18.3 dependencies (#280)
- [cd74716b](https://github.com/kubedb/operator/commit/cd74716b) Update Kubernetes v1.18.3 dependencies (#275)
- [5b3c76ed](https://github.com/kubedb/operator/commit/5b3c76ed) Update Kubernetes v1.18.3 dependencies (#274)
- [397a7e60](https://github.com/kubedb/operator/commit/397a7e60) Update Kubernetes v1.18.3 dependencies (#273)
- [616ea78d](https://github.com/kubedb/operator/commit/616ea78d) Update Kubernetes v1.18.3 dependencies (#272)
- [b7b0d2b9](https://github.com/kubedb/operator/commit/b7b0d2b9) Update Kubernetes v1.18.3 dependencies (#271)
- [3afadb7a](https://github.com/kubedb/operator/commit/3afadb7a) Update Kubernetes v1.18.3 dependencies (#270)
- [60b15632](https://github.com/kubedb/operator/commit/60b15632) Remove dependency on enterprise operator (#269)
- [b3648cde](https://github.com/kubedb/operator/commit/b3648cde) Build images in e2e workflow (#268)
- [73dee065](https://github.com/kubedb/operator/commit/73dee065) Update to Kubernetes v1.18.3 (#266)
- [a8a42ab8](https://github.com/kubedb/operator/commit/a8a42ab8) Allow configuring k8s in e2e tests (#267)
- [4b7d6ee3](https://github.com/kubedb/operator/commit/4b7d6ee3) Trigger e2e tests on /ok-to-test command (#265)
- [024fc40a](https://github.com/kubedb/operator/commit/024fc40a) Update to Kubernetes v1.18.3 (#264)
- [bd1da662](https://github.com/kubedb/operator/commit/bd1da662) Update to Kubernetes v1.18.3 (#263)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.1.0-beta.2](https://github.com/kubedb/percona-xtradb/releases/tag/v0.1.0-beta.2)

- [471b6def](https://github.com/kubedb/percona-xtradb/commit/471b6def) Prepare for release v0.1.0-beta.2 (#60)
- [9423a70f](https://github.com/kubedb/percona-xtradb/commit/9423a70f) Update release.yml
- [85d1d036](https://github.com/kubedb/percona-xtradb/commit/85d1d036) Use updated apis (#59)
- [6811b8dc](https://github.com/kubedb/percona-xtradb/commit/6811b8dc) Update Kubernetes v1.18.3 dependencies (#53)
- [4212d2a0](https://github.com/kubedb/percona-xtradb/commit/4212d2a0) Update Kubernetes v1.18.3 dependencies (#52)
- [659d646c](https://github.com/kubedb/percona-xtradb/commit/659d646c) Update Kubernetes v1.18.3 dependencies (#51)
- [a868e0c3](https://github.com/kubedb/percona-xtradb/commit/a868e0c3) Update Kubernetes v1.18.3 dependencies (#50)
- [162e6ca4](https://github.com/kubedb/percona-xtradb/commit/162e6ca4) Update Kubernetes v1.18.3 dependencies (#49)
- [a7fa1fbf](https://github.com/kubedb/percona-xtradb/commit/a7fa1fbf) Update Kubernetes v1.18.3 dependencies (#48)
- [b6a4583f](https://github.com/kubedb/percona-xtradb/commit/b6a4583f) Remove dependency on enterprise operator (#47)
- [a8909b38](https://github.com/kubedb/percona-xtradb/commit/a8909b38) Allow configuring k8s & db version in e2e tests (#46)
- [4d79d26e](https://github.com/kubedb/percona-xtradb/commit/4d79d26e) Update to Kubernetes v1.18.3 (#45)
- [189f3212](https://github.com/kubedb/percona-xtradb/commit/189f3212) Trigger e2e tests on /ok-to-test command (#44)
- [a037bd03](https://github.com/kubedb/percona-xtradb/commit/a037bd03) Update to Kubernetes v1.18.3 (#43)
- [33cabdf3](https://github.com/kubedb/percona-xtradb/commit/33cabdf3) Update to Kubernetes v1.18.3 (#42)



## [kubedb/pg-leader-election](https://github.com/kubedb/pg-leader-election)

### [v0.2.0-beta.2](https://github.com/kubedb/pg-leader-election/releases/tag/v0.2.0-beta.2)

- [f92f350](https://github.com/kubedb/pg-leader-election/commit/f92f350) Update Kubernetes v1.18.3 dependencies (#17)
- [65c551f](https://github.com/kubedb/pg-leader-election/commit/65c551f) Update Kubernetes v1.18.3 dependencies (#16)
- [c7b516d](https://github.com/kubedb/pg-leader-election/commit/c7b516d) Update Kubernetes v1.18.3 dependencies (#15)
- [8440ee3](https://github.com/kubedb/pg-leader-election/commit/8440ee3) Update Kubernetes v1.18.3 dependencies (#14)
- [33b175b](https://github.com/kubedb/pg-leader-election/commit/33b175b) Update Kubernetes v1.18.3 dependencies (#13)
- [102fbfa](https://github.com/kubedb/pg-leader-election/commit/102fbfa) Update Kubernetes v1.18.3 dependencies (#12)
- [d850da1](https://github.com/kubedb/pg-leader-election/commit/d850da1) Update Kubernetes v1.18.3 dependencies (#11)
- [0505eaf](https://github.com/kubedb/pg-leader-election/commit/0505eaf) Update Kubernetes v1.18.3 dependencies (#10)
- [d46e56c](https://github.com/kubedb/pg-leader-election/commit/d46e56c) Use actions/upload-artifact@v2
- [37fb860](https://github.com/kubedb/pg-leader-election/commit/37fb860) Update to Kubernetes v1.18.3 (#9)
- [7566bf3](https://github.com/kubedb/pg-leader-election/commit/7566bf3) Update to Kubernetes v1.18.3 (#8)
- [07c4965](https://github.com/kubedb/pg-leader-election/commit/07c4965) Update to Kubernetes v1.18.3 (#7)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.1.0-beta.2](https://github.com/kubedb/pgbouncer/releases/tag/v0.1.0-beta.2)

- [e083d55](https://github.com/kubedb/pgbouncer/commit/e083d55) Prepare for release v0.1.0-beta.2 (#41)
- [fe84790](https://github.com/kubedb/pgbouncer/commit/fe84790) Update release.yml
- [ddf5a85](https://github.com/kubedb/pgbouncer/commit/ddf5a85) Use updated certificate spec (#35)
- [d5cd5bf](https://github.com/kubedb/pgbouncer/commit/d5cd5bf) Update Kubernetes v1.18.3 dependencies (#39)
- [21693c7](https://github.com/kubedb/pgbouncer/commit/21693c7) Update Kubernetes v1.18.3 dependencies (#38)
- [39ad48d](https://github.com/kubedb/pgbouncer/commit/39ad48d) Update Kubernetes v1.18.3 dependencies (#37)
- [7f1ecc7](https://github.com/kubedb/pgbouncer/commit/7f1ecc7) Update Kubernetes v1.18.3 dependencies (#36)
- [8d9d379](https://github.com/kubedb/pgbouncer/commit/8d9d379) Update Kubernetes v1.18.3 dependencies (#34)
- [c9b8300](https://github.com/kubedb/pgbouncer/commit/c9b8300) Update Kubernetes v1.18.3 dependencies (#33)
- [66c72a4](https://github.com/kubedb/pgbouncer/commit/66c72a4) Remove dependency on enterprise operator (#32)
- [757dc10](https://github.com/kubedb/pgbouncer/commit/757dc10) Update to cert-manager v0.16.0 (#30)
- [0a183d1](https://github.com/kubedb/pgbouncer/commit/0a183d1) Build images in e2e workflow (#29)
- [ca61e88](https://github.com/kubedb/pgbouncer/commit/ca61e88) Allow configuring k8s & db version in e2e tests (#28)
- [a87278b](https://github.com/kubedb/pgbouncer/commit/a87278b) Update to Kubernetes v1.18.3 (#27)
- [5abe86f](https://github.com/kubedb/pgbouncer/commit/5abe86f) Fix formatting
- [845f7a3](https://github.com/kubedb/pgbouncer/commit/845f7a3) Trigger e2e tests on /ok-to-test command (#26)
- [2cc23c0](https://github.com/kubedb/pgbouncer/commit/2cc23c0) Fix cert-manager integration for PgBouncer (#25)
- [2a148c2](https://github.com/kubedb/pgbouncer/commit/2a148c2) Update to Kubernetes v1.18.3 (#24)
- [f6eb812](https://github.com/kubedb/pgbouncer/commit/f6eb812) Update Makefile.env



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.14.0-beta.2](https://github.com/kubedb/postgres/releases/tag/v0.14.0-beta.2)

- [6e6fe6fe](https://github.com/kubedb/postgres/commit/6e6fe6fe) Prepare for release v0.14.0-beta.2 (#345)
- [5ee33bb8](https://github.com/kubedb/postgres/commit/5ee33bb8) Update release.yml
- [9208f754](https://github.com/kubedb/postgres/commit/9208f754) Always use OnDelete update strategy
- [74367d01](https://github.com/kubedb/postgres/commit/74367d01) Update Kubernetes v1.18.3 dependencies (#344)
- [01843533](https://github.com/kubedb/postgres/commit/01843533) Update Kubernetes v1.18.3 dependencies (#343)
- [34a3a460](https://github.com/kubedb/postgres/commit/34a3a460) Update Kubernetes v1.18.3 dependencies (#338)
- [455bf56a](https://github.com/kubedb/postgres/commit/455bf56a) Update Kubernetes v1.18.3 dependencies (#337)
- [960d1efa](https://github.com/kubedb/postgres/commit/960d1efa) Update Kubernetes v1.18.3 dependencies (#336)
- [9b428745](https://github.com/kubedb/postgres/commit/9b428745) Update Kubernetes v1.18.3 dependencies (#335)
- [cc95c5f5](https://github.com/kubedb/postgres/commit/cc95c5f5) Update Kubernetes v1.18.3 dependencies (#334)
- [c0694d83](https://github.com/kubedb/postgres/commit/c0694d83) Update Kubernetes v1.18.3 dependencies (#333)
- [8d0977d3](https://github.com/kubedb/postgres/commit/8d0977d3) Remove dependency on enterprise operator (#332)
- [daa5b77c](https://github.com/kubedb/postgres/commit/daa5b77c) Build images in e2e workflow (#331)
- [197f1b2b](https://github.com/kubedb/postgres/commit/197f1b2b) Update to Kubernetes v1.18.3 (#329)
- [e732d319](https://github.com/kubedb/postgres/commit/e732d319) Allow configuring k8s & db version in e2e tests (#330)
- [f37180ec](https://github.com/kubedb/postgres/commit/f37180ec) Trigger e2e tests on /ok-to-test command (#328)
- [becb3e2c](https://github.com/kubedb/postgres/commit/becb3e2c) Update to Kubernetes v1.18.3 (#327)
- [91bf7440](https://github.com/kubedb/postgres/commit/91bf7440) Update to Kubernetes v1.18.3 (#326)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.1.0-beta.2](https://github.com/kubedb/proxysql/releases/tag/v0.1.0-beta.2)

- [f86bb6cd](https://github.com/kubedb/proxysql/commit/f86bb6cd) Prepare for release v0.1.0-beta.2 (#46)
- [e74f3803](https://github.com/kubedb/proxysql/commit/e74f3803) Update release.yml
- [7f5349cc](https://github.com/kubedb/proxysql/commit/7f5349cc) Use updated apis (#45)
- [27faefef](https://github.com/kubedb/proxysql/commit/27faefef) Update for release Stash@v2020.08.27 (#43)
- [65bc5bca](https://github.com/kubedb/proxysql/commit/65bc5bca) Update for release Stash@v2020.08.27-rc.0 (#42)
- [833ac78b](https://github.com/kubedb/proxysql/commit/833ac78b) Update for release Stash@v2020.08.26-rc.1 (#41)
- [fe13ce42](https://github.com/kubedb/proxysql/commit/fe13ce42) Update for release Stash@v2020.08.26-rc.0 (#40)
- [b1a72843](https://github.com/kubedb/proxysql/commit/b1a72843) Update Kubernetes v1.18.3 dependencies (#39)
- [a9c40618](https://github.com/kubedb/proxysql/commit/a9c40618) Update Kubernetes v1.18.3 dependencies (#38)
- [664c974a](https://github.com/kubedb/proxysql/commit/664c974a) Update Kubernetes v1.18.3 dependencies (#37)
- [69ed46d5](https://github.com/kubedb/proxysql/commit/69ed46d5) Update Kubernetes v1.18.3 dependencies (#36)
- [a93d80d4](https://github.com/kubedb/proxysql/commit/a93d80d4) Update Kubernetes v1.18.3 dependencies (#35)
- [84fc9e37](https://github.com/kubedb/proxysql/commit/84fc9e37) Update Kubernetes v1.18.3 dependencies (#34)
- [b09f89d0](https://github.com/kubedb/proxysql/commit/b09f89d0) Remove dependency on enterprise operator (#33)
- [78ad5a88](https://github.com/kubedb/proxysql/commit/78ad5a88) Build images in e2e workflow (#32)
- [6644058e](https://github.com/kubedb/proxysql/commit/6644058e) Update to Kubernetes v1.18.3 (#30)
- [2c03dadd](https://github.com/kubedb/proxysql/commit/2c03dadd) Allow configuring k8s & db version in e2e tests (#31)
- [2c6e04bc](https://github.com/kubedb/proxysql/commit/2c6e04bc) Trigger e2e tests on /ok-to-test command (#29)
- [c7830af8](https://github.com/kubedb/proxysql/commit/c7830af8) Update to Kubernetes v1.18.3 (#28)
- [f2da8746](https://github.com/kubedb/proxysql/commit/f2da8746) Update to Kubernetes v1.18.3 (#27)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.7.0-beta.2](https://github.com/kubedb/redis/releases/tag/v0.7.0-beta.2)

- [73cf267e](https://github.com/kubedb/redis/commit/73cf267e) Prepare for release v0.7.0-beta.2 (#192)
- [d2911ea9](https://github.com/kubedb/redis/commit/d2911ea9) Update release.yml
- [c76ee46e](https://github.com/kubedb/redis/commit/c76ee46e) Update dependencies (#191)
- [0b030534](https://github.com/kubedb/redis/commit/0b030534) Fix build
- [408216ab](https://github.com/kubedb/redis/commit/408216ab) Add support for Redis v6.0.6 and TLS (#180)
- [944327df](https://github.com/kubedb/redis/commit/944327df) Update Kubernetes v1.18.3 dependencies (#187)
- [40b7cde6](https://github.com/kubedb/redis/commit/40b7cde6) Update Kubernetes v1.18.3 dependencies (#186)
- [f2bf110d](https://github.com/kubedb/redis/commit/f2bf110d) Update Kubernetes v1.18.3 dependencies (#184)
- [61485cfa](https://github.com/kubedb/redis/commit/61485cfa) Update Kubernetes v1.18.3 dependencies (#183)
- [184ae35d](https://github.com/kubedb/redis/commit/184ae35d) Update Kubernetes v1.18.3 dependencies (#182)
- [bc72b51b](https://github.com/kubedb/redis/commit/bc72b51b) Update Kubernetes v1.18.3 dependencies (#181)
- [ca540560](https://github.com/kubedb/redis/commit/ca540560) Remove dependency on enterprise operator (#179)
- [09bade2e](https://github.com/kubedb/redis/commit/09bade2e) Allow configuring k8s & db version in e2e tests (#178)
- [2bafb114](https://github.com/kubedb/redis/commit/2bafb114) Update to Kubernetes v1.18.3 (#177)
- [b2fe59ef](https://github.com/kubedb/redis/commit/b2fe59ef) Trigger e2e tests on /ok-to-test command (#176)
- [df5131e1](https://github.com/kubedb/redis/commit/df5131e1) Update to Kubernetes v1.18.3 (#175)
- [a404ae08](https://github.com/kubedb/redis/commit/a404ae08) Update to Kubernetes v1.18.3 (#174)




