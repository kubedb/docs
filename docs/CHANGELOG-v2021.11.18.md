---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2021.11.18
    name: Changelog-v2021.11.18
    parent: welcome
    weight: 20211118
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2021.11.18/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2021.11.18/
---

# KubeDB v2021.11.18 (2021-11-17)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.23.0](https://github.com/kubedb/apimachinery/releases/tag/v0.23.0)

- [ff3a4175](https://github.com/kubedb/apimachinery/commit/ff3a4175) Update repository config (#819)
- [1969d04c](https://github.com/kubedb/apimachinery/commit/1969d04c) Remove EnableAnalytics (#818)
- [1222a1d6](https://github.com/kubedb/apimachinery/commit/1222a1d6) Add pod and workload controller label support (#817)
- [1cef6837](https://github.com/kubedb/apimachinery/commit/1cef6837) Allow vertical scaling Coordinator (#816)
- [90d46474](https://github.com/kubedb/apimachinery/commit/90d46474) Add distribution tags for KubeDB (#815)
- [24d44217](https://github.com/kubedb/apimachinery/commit/24d44217) Update default resource for pg-coordinator (#813)
- [807280ce](https://github.com/kubedb/apimachinery/commit/807280ce) Add `applyConfig` in MongoDBOpsRequest for custom configuration (#811)
- [6f31cb6a](https://github.com/kubedb/apimachinery/commit/6f31cb6a) Add support for OpenSearch (#810)
- [b9f7eadd](https://github.com/kubedb/apimachinery/commit/b9f7eadd) Stop using storage.k8s.io/v1beta1 deprecated in k8s 1.22 (#814)
- [73f09b6b](https://github.com/kubedb/apimachinery/commit/73f09b6b) Add support for reconfigure Elasticsearch (#793)
- [997836d3](https://github.com/kubedb/apimachinery/commit/997836d3) Add Redis Constants for Config files (#812)
- [d0176524](https://github.com/kubedb/apimachinery/commit/d0176524) Remove statefulSetOrdinal from MySQL ops reuqest (#809)
- [ad8c1f78](https://github.com/kubedb/apimachinery/commit/ad8c1f78) Update MySQLClusterMode constants (#808)
- [8ce33b18](https://github.com/kubedb/apimachinery/commit/8ce33b18) Update deps
- [d8ea50ce](https://github.com/kubedb/apimachinery/commit/d8ea50ce) Update for release Stash@v2021.10.11 (#807)
- [4937544e](https://github.com/kubedb/apimachinery/commit/4937544e) Add MySQL Router constant (#805)
- [ee093b4d](https://github.com/kubedb/apimachinery/commit/ee093b4d) Add new distribution values (#806)
- [47b42be0](https://github.com/kubedb/apimachinery/commit/47b42be0) Add support for MySQL coordinator (#803)
- [d1f74f0b](https://github.com/kubedb/apimachinery/commit/d1f74f0b) Update repository config (#804)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.8.0](https://github.com/kubedb/autoscaler/releases/tag/v0.8.0)

- [85440f50](https://github.com/kubedb/autoscaler/commit/85440f50) Prepare for release v0.8.0 (#52)
- [5f59bb99](https://github.com/kubedb/autoscaler/commit/5f59bb99) Update kmodules.xyz/monitoring-agent-api (#51)
- [dcaa9d9d](https://github.com/kubedb/autoscaler/commit/dcaa9d9d) Update repository config (#50)
- [a9c755f1](https://github.com/kubedb/autoscaler/commit/a9c755f1) Use DisableAnalytics flag from license (#49)
- [a97521e7](https://github.com/kubedb/autoscaler/commit/a97521e7) Update license-verifier (#48)
- [7cda0e3b](https://github.com/kubedb/autoscaler/commit/7cda0e3b) Support custom pod and controller labels (#47)
- [99b2710c](https://github.com/kubedb/autoscaler/commit/99b2710c) Fix mongodb shard autoscaling issue (#46)
- [3302e496](https://github.com/kubedb/autoscaler/commit/3302e496) Merge recommended resource with current resource (#45)
- [7f6e3994](https://github.com/kubedb/autoscaler/commit/7f6e3994) Update dependencies (#44)
- [0ab54377](https://github.com/kubedb/autoscaler/commit/0ab54377) Fix satori/go.uuid security vulnerability (#43)
- [898e4497](https://github.com/kubedb/autoscaler/commit/898e4497) Fix jwt-go security vulnerability (#42)
- [39e647ff](https://github.com/kubedb/autoscaler/commit/39e647ff) Fix jwt-go security vulnerability (#41)
- [e898b195](https://github.com/kubedb/autoscaler/commit/e898b195) Use nats.go v1.13.0 (#39)
- [dc0b7b32](https://github.com/kubedb/autoscaler/commit/dc0b7b32) Setup SiteInfo publisher (#40)
- [de10d221](https://github.com/kubedb/autoscaler/commit/de10d221) Update dependencies to publish SiteInfo (#38)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.23.0](https://github.com/kubedb/cli/releases/tag/v0.23.0)

- [176feb8f](https://github.com/kubedb/cli/commit/176feb8f) Prepare for release v0.23.0 (#641)
- [c10cf297](https://github.com/kubedb/cli/commit/c10cf297) Update kmodules.xyz/monitoring-agent-api (#640)
- [6e5e3e57](https://github.com/kubedb/cli/commit/6e5e3e57) Use DisableAnalytics flag from license (#639)
- [3ca8fbf6](https://github.com/kubedb/cli/commit/3ca8fbf6) Update license-verifier (#638)
- [9ba88756](https://github.com/kubedb/cli/commit/9ba88756) Support custom pod and controller labels (#637)
- [67ae9ed4](https://github.com/kubedb/cli/commit/67ae9ed4) Update dependencies (#636)
- [58159a19](https://github.com/kubedb/cli/commit/58159a19) Fix satori/go.uuid security vulnerability (#635)
- [1350b7f4](https://github.com/kubedb/cli/commit/1350b7f4) Fix jwt-go security vulnerability (#634)
- [1e783e44](https://github.com/kubedb/cli/commit/1e783e44) Fix jwt-go security vulnerability (#633)
- [fe1a9aeb](https://github.com/kubedb/cli/commit/fe1a9aeb) Fix jwt-go security vulnerability (#632)
- [a79d2705](https://github.com/kubedb/cli/commit/a79d2705) Update dependencies to publish SiteInfo (#631)
- [6e9be4c3](https://github.com/kubedb/cli/commit/6e9be4c3) Update dependencies to publish SiteInfo (#630)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.23.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.23.0)

- [d653718f](https://github.com/kubedb/elasticsearch/commit/d653718f) Prepare for release v0.23.0 (#540)
- [a15c804d](https://github.com/kubedb/elasticsearch/commit/a15c804d) Update kmodules.xyz/monitoring-agent-api (#539)
- [0778ff4f](https://github.com/kubedb/elasticsearch/commit/0778ff4f) Remove global variable for preconditions (#538)
- [bd084ade](https://github.com/kubedb/elasticsearch/commit/bd084ade) Update repository config (#537)
- [180f2c47](https://github.com/kubedb/elasticsearch/commit/180f2c47) Remove docs folder
- [a97b7131](https://github.com/kubedb/elasticsearch/commit/a97b7131) Update docs
- [16895e92](https://github.com/kubedb/elasticsearch/commit/16895e92) Use DisableAnalytics flag from license (#536)
- [62c2d28f](https://github.com/kubedb/elasticsearch/commit/62c2d28f) Update license-verifier (#535)
- [c70caee3](https://github.com/kubedb/elasticsearch/commit/c70caee3) Add pod, services and workload-controller(sts) label support (#532)
- [0e5aeb59](https://github.com/kubedb/elasticsearch/commit/0e5aeb59) Add support for OpenSearch (#529)
- [fd527ccb](https://github.com/kubedb/elasticsearch/commit/fd527ccb) Update dependencies (#531)
- [b2cf3e9f](https://github.com/kubedb/elasticsearch/commit/b2cf3e9f) Always create admin certs if the cluster security is enabled (#516)
- [aae6bc29](https://github.com/kubedb/elasticsearch/commit/aae6bc29) Fix satori/go.uuid security vulnerability (#530)
- [50aa1c9e](https://github.com/kubedb/elasticsearch/commit/50aa1c9e) Fix jwt-go security vulnerability (#528)
- [4b6ebc0c](https://github.com/kubedb/elasticsearch/commit/4b6ebc0c) Fix jwt-go security vulnerability (#527)
- [9d87c0a7](https://github.com/kubedb/elasticsearch/commit/9d87c0a7) Use nats.go v1.13.0 (#526)
- [cc4811ef](https://github.com/kubedb/elasticsearch/commit/cc4811ef) Setup SiteInfo publisher (#525)
- [00feb65a](https://github.com/kubedb/elasticsearch/commit/00feb65a) Update dependencies to publish SiteInfo (#524)
- [a7f4137f](https://github.com/kubedb/elasticsearch/commit/a7f4137f) Update dependencies to publish SiteInfo (#523)
- [7e3c63fd](https://github.com/kubedb/elasticsearch/commit/7e3c63fd) Collect metrics from all type of Elasticsearch nodes (#521)



## [kubedb/enterprise](https://github.com/kubedb/enterprise)

### [v0.10.0](https://github.com/kubedb/enterprise/releases/tag/v0.10.0)

- [3214618e](https://github.com/kubedb/enterprise/commit/3214618e) Prepare for release v0.10.0 (#251)
- [41c5b619](https://github.com/kubedb/enterprise/commit/41c5b619) Update kmodules.xyz/monitoring-agent-api (#250)
- [64252ddb](https://github.com/kubedb/enterprise/commit/64252ddb) Remove global variable for preconditions (#249)
- [aa368d82](https://github.com/kubedb/enterprise/commit/aa368d82) Update repository config (#248)
- [9bcfee5a](https://github.com/kubedb/enterprise/commit/9bcfee5a) Fix semver checking. (#247)
- [91ea48f1](https://github.com/kubedb/enterprise/commit/91ea48f1) Update docs
- [34560f43](https://github.com/kubedb/enterprise/commit/34560f43) Use DisableAnalytics flag from license (#246)
- [fac8c82b](https://github.com/kubedb/enterprise/commit/fac8c82b) Update license-verifier (#245)
- [2d3839af](https://github.com/kubedb/enterprise/commit/2d3839af) Support custom pod and controller labels (#244)
- [83488e1c](https://github.com/kubedb/enterprise/commit/83488e1c) Add backup permission for mysql replication user (#243)
- [0dd46f8f](https://github.com/kubedb/enterprise/commit/0dd46f8f) Add support for reconfigure Elasticsearch (#220)
- [2a980832](https://github.com/kubedb/enterprise/commit/2a980832) Use `kubedb.dev/db-client-go` for mongodb (#241)
- [69d59b2c](https://github.com/kubedb/enterprise/commit/69d59b2c) Update mongodb vertical scaling logic (#240)
- [e9b227b7](https://github.com/kubedb/enterprise/commit/e9b227b7) Update Redis Reconfigure Ops Request (#236)
- [9342c052](https://github.com/kubedb/enterprise/commit/9342c052) Add support for mongodb reconfigure replicaSet config (#235)
- [330c3be1](https://github.com/kubedb/enterprise/commit/330c3be1) Fix upgrade opsrequest for mysql coordinator (#229)
- [f64f7d29](https://github.com/kubedb/enterprise/commit/f64f7d29) Update dependencies (#239)
- [60b2d128](https://github.com/kubedb/enterprise/commit/60b2d128) Update xorm dependency (#238)
- [fdcd91d3](https://github.com/kubedb/enterprise/commit/fdcd91d3) Fix satori/go.uuid security vulnerability (#237)
- [62cb9918](https://github.com/kubedb/enterprise/commit/62cb9918) Fix jwt-go security vulnerability (#234)
- [6af2b5a5](https://github.com/kubedb/enterprise/commit/6af2b5a5) Fix jwt-go security vulnerability (#233)
- [527eb0e4](https://github.com/kubedb/enterprise/commit/527eb0e4) Fix: major and minor Upgrade issue for Postgres Debian images (#232)
- [e8b5d6ae](https://github.com/kubedb/enterprise/commit/e8b5d6ae) Use nats.go v1.13.0 (#231)
- [80c6c4ec](https://github.com/kubedb/enterprise/commit/80c6c4ec) Setup SiteInfo publisher (#230)
- [b9d9d37f](https://github.com/kubedb/enterprise/commit/b9d9d37f) Update dependencies to publish SiteInfo (#228)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2021.11.18](https://github.com/kubedb/installer/releases/tag/v2021.11.18)

- [3fe4c869](https://github.com/kubedb/installer/commit/3fe4c869) Prepare for release v2021.11.18 (#395)
- [b7321269](https://github.com/kubedb/installer/commit/b7321269) Use mysqld Exporter Image with custom query support (#390)
- [5372ab67](https://github.com/kubedb/installer/commit/5372ab67) Add new Postgres versions in catalog (#394)
- [09621a10](https://github.com/kubedb/installer/commit/09621a10) Update kmodules.xyz/monitoring-agent-api (#393)
- [23f4d5c1](https://github.com/kubedb/installer/commit/23f4d5c1) Update repository config (#392)
- [3fa119af](https://github.com/kubedb/installer/commit/3fa119af) Add labels to license related rolebindings & secrets (#391)
- [9e18ab72](https://github.com/kubedb/installer/commit/9e18ab72) Add MySQL 5.7.36, 8.0.17 8.0.27 (#378)
- [5c8d2854](https://github.com/kubedb/installer/commit/5c8d2854) Remove --enable-analytics flag (#389)
- [589c93f8](https://github.com/kubedb/installer/commit/589c93f8) Update license-verifier (#388)
- [f1f19f47](https://github.com/kubedb/installer/commit/f1f19f47) Fix regression in #386 (#387)
- [ddd3e9f3](https://github.com/kubedb/installer/commit/ddd3e9f3) Change installer namespace to kubedb (#386)
- [64e95827](https://github.com/kubedb/installer/commit/64e95827) Support OpenSearch 1.1.0 and update elasticsearch-exporter image to 1.3.0 (#384)
- [6cf27bc1](https://github.com/kubedb/installer/commit/6cf27bc1) Update Postgres-init Image Version to v0.4.0 (#382)
- [d5db17f8](https://github.com/kubedb/installer/commit/d5db17f8) Update crds
- [0dbc20cb](https://github.com/kubedb/installer/commit/0dbc20cb) Update crds
- [c11c29f6](https://github.com/kubedb/installer/commit/c11c29f6) Update dependencies (#381)
- [1ef34de3](https://github.com/kubedb/installer/commit/1ef34de3) Update Redis Init Image version for Custom Config fixes (#380)
- [cb48c46b](https://github.com/kubedb/installer/commit/cb48c46b) Fix satori/go.uuid security vulnerability (#379)
- [f39255c3](https://github.com/kubedb/installer/commit/f39255c3) Add innodb and coordiantor support (#371)
- [53ab0b1b](https://github.com/kubedb/installer/commit/53ab0b1b) Fix jwt-go security vulnerability (#377)
- [01ae1087](https://github.com/kubedb/installer/commit/01ae1087) Fix jwt-go security vulnerability (#376)
- [b17e22a0](https://github.com/kubedb/installer/commit/b17e22a0) Add fields to MySQL Metrics (#375)
- [b4e4d317](https://github.com/kubedb/installer/commit/b4e4d317) Add New Postgres versions (#374)
- [19094b7a](https://github.com/kubedb/installer/commit/19094b7a) Update crds
- [85e02d5b](https://github.com/kubedb/installer/commit/85e02d5b) Add mongodb `5.0.3` (#372)
- [e055d79a](https://github.com/kubedb/installer/commit/e055d79a) Mark versions using Official docker images as Official Distro (#373)
- [3e45fb54](https://github.com/kubedb/installer/commit/3e45fb54) Add SiteInfo publisher permission
- [df913373](https://github.com/kubedb/installer/commit/df913373) Update dependencies to publish SiteInfo (#369)
- [a1cc057f](https://github.com/kubedb/installer/commit/a1cc057f) Add fields to redis-metrics (#366)
- [bf243169](https://github.com/kubedb/installer/commit/bf243169) Add fields to MariaDB Metrics (#370)
- [b53d5cc3](https://github.com/kubedb/installer/commit/b53d5cc3) Add v14.0 in Postgres catalog (#368)
- [e3ec7f67](https://github.com/kubedb/installer/commit/e3ec7f67) Add redis sentinel metrics configuration (#367)
- [1881967c](https://github.com/kubedb/installer/commit/1881967c) Update various kubedb metrics and metric labels (#364)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.7.0](https://github.com/kubedb/mariadb/releases/tag/v0.7.0)

- [05707163](https://github.com/kubedb/mariadb/commit/05707163) Prepare for release v0.7.0 (#112)
- [2818eb2b](https://github.com/kubedb/mariadb/commit/2818eb2b) Update kmodules.xyz/monitoring-agent-api (#111)
- [4580ebd5](https://github.com/kubedb/mariadb/commit/4580ebd5) Remove global variable for preconditions (#110)
- [8223c352](https://github.com/kubedb/mariadb/commit/8223c352) Update repository config (#109)
- [8be974a6](https://github.com/kubedb/mariadb/commit/8be974a6) Remove docs
- [0279fa08](https://github.com/kubedb/mariadb/commit/0279fa08) Update docs
- [45cbdb9e](https://github.com/kubedb/mariadb/commit/45cbdb9e) Use DisableAnalytics flag from license (#108)
- [0d4ae537](https://github.com/kubedb/mariadb/commit/0d4ae537) Update license-verifier (#107)
- [92626beb](https://github.com/kubedb/mariadb/commit/92626beb) Support custom pod, service, and controller(sts) labels (#105)
- [afd25e04](https://github.com/kubedb/mariadb/commit/afd25e04) Update dependencies (#104)
- [297c7cdb](https://github.com/kubedb/mariadb/commit/297c7cdb) Update xorm dependency (#103)
- [fc99578b](https://github.com/kubedb/mariadb/commit/fc99578b) Fix satori/go.uuid security vulnerability (#102)
- [43236638](https://github.com/kubedb/mariadb/commit/43236638) Fix jwt-go security vulnerability (#101)
- [247e1413](https://github.com/kubedb/mariadb/commit/247e1413) Fix jwt-go security vulnerability (#100)
- [1ef0690d](https://github.com/kubedb/mariadb/commit/1ef0690d) Use nats.go v1.13.0 (#99)
- [2a067c0b](https://github.com/kubedb/mariadb/commit/2a067c0b) Setup SiteInfo publisher (#98)
- [72c93bb2](https://github.com/kubedb/mariadb/commit/72c93bb2) Update dependencies to publish SiteInfo (#97)
- [2e17fbb6](https://github.com/kubedb/mariadb/commit/2e17fbb6) Update dependencies to publish SiteInfo (#96)
- [8b091ae3](https://github.com/kubedb/mariadb/commit/8b091ae3) Update repository config (#95)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.3.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.3.0)

- [d229ad1](https://github.com/kubedb/mariadb-coordinator/commit/d229ad1) Prepare for release v0.3.0 (#24)
- [04ac158](https://github.com/kubedb/mariadb-coordinator/commit/04ac158) Update kmodules.xyz/monitoring-agent-api (#23)
- [b1836cd](https://github.com/kubedb/mariadb-coordinator/commit/b1836cd) Update repository config (#22)
- [670cce7](https://github.com/kubedb/mariadb-coordinator/commit/670cce7) Use DisableAnalytics flag from license (#21)
- [b2149b3](https://github.com/kubedb/mariadb-coordinator/commit/b2149b3) Update license-verifier (#20)
- [43e2907](https://github.com/kubedb/mariadb-coordinator/commit/43e2907) Support custom pod and controller labels (#19)
- [054ad28](https://github.com/kubedb/mariadb-coordinator/commit/054ad28) Update dependencies (#18)
- [73b094a](https://github.com/kubedb/mariadb-coordinator/commit/73b094a) Update xorm dependency (#17)
- [d401ce6](https://github.com/kubedb/mariadb-coordinator/commit/d401ce6) Fix satori/go.uuid security vulnerability (#16)
- [fbbec4b](https://github.com/kubedb/mariadb-coordinator/commit/fbbec4b) Fix jwt-go security vulnerability (#15)
- [bf9222c](https://github.com/kubedb/mariadb-coordinator/commit/bf9222c) Fix jwt-go security vulnerability (#14)
- [dbac458](https://github.com/kubedb/mariadb-coordinator/commit/dbac458) Use nats.go v1.13.0 (#13)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.16.0](https://github.com/kubedb/memcached/releases/tag/v0.16.0)

- [f1131b24](https://github.com/kubedb/memcached/commit/f1131b24) Prepare for release v0.16.0 (#327)
- [9a48dfb4](https://github.com/kubedb/memcached/commit/9a48dfb4) Update kmodules.xyz/monitoring-agent-api (#326)
- [eedff52b](https://github.com/kubedb/memcached/commit/eedff52b) Remove global variable for preconditions (#325)
- [7e9aa7cb](https://github.com/kubedb/memcached/commit/7e9aa7cb) Update repository config (#324)
- [83d8990b](https://github.com/kubedb/memcached/commit/83d8990b) Remove docs
- [75c6aaae](https://github.com/kubedb/memcached/commit/75c6aaae) Update docs
- [d44def4b](https://github.com/kubedb/memcached/commit/d44def4b) Use DisableAnalytics flag from license (#323)
- [f1ac7471](https://github.com/kubedb/memcached/commit/f1ac7471) Update license-verifier (#322)
- [7c395019](https://github.com/kubedb/memcached/commit/7c395019) Support custom pod, service, and controller labels (#321)
- [b138b898](https://github.com/kubedb/memcached/commit/b138b898) Update dependencies (#320)
- [789dd6f7](https://github.com/kubedb/memcached/commit/789dd6f7) Fix satori/go.uuid security vulnerability (#319)
- [37d03918](https://github.com/kubedb/memcached/commit/37d03918) Fix jwt-go security vulnerability (#318)
- [27e097a3](https://github.com/kubedb/memcached/commit/27e097a3) Fix jwt-go security vulnerability (#317)
- [8fe76024](https://github.com/kubedb/memcached/commit/8fe76024) Use nats.go v1.13.0 (#316)
- [1e1443e0](https://github.com/kubedb/memcached/commit/1e1443e0) Update dependencies to publish SiteInfo (#315)
- [5c4569d2](https://github.com/kubedb/memcached/commit/5c4569d2) Update dependencies to publish SiteInfo (#314)
- [912ec127](https://github.com/kubedb/memcached/commit/912ec127) Update repository config (#313)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.16.0](https://github.com/kubedb/mongodb/releases/tag/v0.16.0)

- [c72e7335](https://github.com/kubedb/mongodb/commit/c72e7335) Prepare for release v0.16.0 (#437)
- [43ac7699](https://github.com/kubedb/mongodb/commit/43ac7699) Update kmodules.xyz/monitoring-agent-api (#436)
- [4ad8f28c](https://github.com/kubedb/mongodb/commit/4ad8f28c) Remove global variable for preconditions (#435)
- [e009f4ec](https://github.com/kubedb/mongodb/commit/e009f4ec) Update repository config (#434)
- [02cc1e50](https://github.com/kubedb/mongodb/commit/02cc1e50) Remove docs
- [e24969a1](https://github.com/kubedb/mongodb/commit/e24969a1) Use DisableAnalytics flag from license (#433)
- [8dc342e6](https://github.com/kubedb/mongodb/commit/8dc342e6) Update license-verifier (#432)
- [ecfb1583](https://github.com/kubedb/mongodb/commit/ecfb1583) Support custom pod and controller labels (#431)
- [a0550a93](https://github.com/kubedb/mongodb/commit/a0550a93) Add pod, statefulSet and service labels support (#430)
- [6ac1a182](https://github.com/kubedb/mongodb/commit/6ac1a182) Use `kubedb.dev/db-client-go` (#429)
- [8b2ed1c6](https://github.com/kubedb/mongodb/commit/8b2ed1c6) Add support for ReplicaSet configuration (#426)
- [07a2f120](https://github.com/kubedb/mongodb/commit/07a2f120) Update dependencies (#428)
- [f3f206f8](https://github.com/kubedb/mongodb/commit/f3f206f8) Fix satori/go.uuid security vulnerability (#427)
- [5c5c669b](https://github.com/kubedb/mongodb/commit/5c5c669b) Set owner reference to the secrets created by the operator (#425)
- [17ea4294](https://github.com/kubedb/mongodb/commit/17ea4294) Fix jwt-go security vulnerability (#424)
- [6a0dccf3](https://github.com/kubedb/mongodb/commit/6a0dccf3) Fix jwt-go security vulnerability (#423)
- [db40027d](https://github.com/kubedb/mongodb/commit/db40027d) Use nats.go v1.13.0 (#422)
- [473928f4](https://github.com/kubedb/mongodb/commit/473928f4) Setup SiteInfo publisher (#421)
- [b9ce138a](https://github.com/kubedb/mongodb/commit/b9ce138a) Update dependencies to publish SiteInfo (#420)
- [fff26a96](https://github.com/kubedb/mongodb/commit/fff26a96) Update dependencies to publish SiteInfo (#419)
- [41f3ccd9](https://github.com/kubedb/mongodb/commit/41f3ccd9) Update repository config (#418)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.16.0](https://github.com/kubedb/mysql/releases/tag/v0.16.0)

- [0680eeb3](https://github.com/kubedb/mysql/commit/0680eeb3) Prepare for release v0.16.0 (#429)
- [375760f3](https://github.com/kubedb/mysql/commit/375760f3) Export Group Replication stats in Exporter Container (#425)
- [2b5af248](https://github.com/kubedb/mysql/commit/2b5af248) Update kmodules.xyz/monitoring-agent-api (#428)
- [57f7cf60](https://github.com/kubedb/mysql/commit/57f7cf60) Remove global variable for preconditions (#427)
- [d47d0e39](https://github.com/kubedb/mysql/commit/d47d0e39) Update repository config (#426)
- [8847d166](https://github.com/kubedb/mysql/commit/8847d166) Update dependencies
- [646da2c8](https://github.com/kubedb/mysql/commit/646da2c8) Remove docs
- [eca0cfd5](https://github.com/kubedb/mysql/commit/eca0cfd5) Use DisableAnalytics flag from license (#424)
- [86d7a80d](https://github.com/kubedb/mysql/commit/86d7a80d) Update license-verifier (#423)
- [de8696fc](https://github.com/kubedb/mysql/commit/de8696fc) Add support for custom pod, service, and controller(sts) labels (#420)
- [87e3ea31](https://github.com/kubedb/mysql/commit/87e3ea31) Update entry point command for mysql router. (#422)
- [73178faa](https://github.com/kubedb/mysql/commit/73178faa) Add support for MySQL Coordinator (#406)
- [0075cf98](https://github.com/kubedb/mysql/commit/0075cf98) Update dependencies (#418)
- [4188a194](https://github.com/kubedb/mysql/commit/4188a194) Fix satori/go.uuid security vulnerability (#417)
- [569e220f](https://github.com/kubedb/mysql/commit/569e220f) Fix jwt-go security vulnerability (#416)
- [be5be397](https://github.com/kubedb/mysql/commit/be5be397) Restrict group replicas for size 2 in Validator (#402)
- [4dbb18f3](https://github.com/kubedb/mysql/commit/4dbb18f3) Fix jwt-go security vulnerability (#414)
- [f4b0bb43](https://github.com/kubedb/mysql/commit/f4b0bb43) Use nats.go v1.13.0 (#413)
- [c5eefa7e](https://github.com/kubedb/mysql/commit/c5eefa7e) Setup SiteInfo publisher (#412)
- [8157ec8f](https://github.com/kubedb/mysql/commit/8157ec8f) Update dependencies to publish SiteInfo (#411)
- [808dbd85](https://github.com/kubedb/mysql/commit/808dbd85) Update dependencies to publish SiteInfo (#410)
- [a949af00](https://github.com/kubedb/mysql/commit/a949af00) Update repository config (#409)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.1.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.1.0)

- [51cf61d](https://github.com/kubedb/mysql-coordinator/commit/51cf61d) Prepare for release v0.1.0 (#18)
- [104431b](https://github.com/kubedb/mysql-coordinator/commit/104431b) Prepare for release v0.1.0 (#17)
- [1cd379d](https://github.com/kubedb/mysql-coordinator/commit/1cd379d) Update kmodules.xyz/monitoring-agent-api (#16)
- [e85255b](https://github.com/kubedb/mysql-coordinator/commit/e85255b) Update repository config (#15)
- [d7f6193](https://github.com/kubedb/mysql-coordinator/commit/d7f6193) Use DisableAnalytics flag from license (#14)
- [c0a51bb](https://github.com/kubedb/mysql-coordinator/commit/c0a51bb) Update license-verifier (#13)
- [d624835](https://github.com/kubedb/mysql-coordinator/commit/d624835) Support custom pod and controller labels (#12)
- [d3bc5ba](https://github.com/kubedb/mysql-coordinator/commit/d3bc5ba) Add sleep for now to avoid the joining problem (#11)
- [e44f9d1](https://github.com/kubedb/mysql-coordinator/commit/e44f9d1) Update dependencies (#10)
- [653e357](https://github.com/kubedb/mysql-coordinator/commit/653e357) Update xorm.io dependency (#9)
- [772ebbd](https://github.com/kubedb/mysql-coordinator/commit/772ebbd) Fix satori/go.uuid security vulnerability (#8)
- [50fdeee](https://github.com/kubedb/mysql-coordinator/commit/50fdeee) Fix jwt-go security vulnerability (#7)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.1.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.1.0)

- [e16b07e](https://github.com/kubedb/mysql-router-init/commit/e16b07e) Update repository config (#13)
- [1d14631](https://github.com/kubedb/mysql-router-init/commit/1d14631) Monitor mysql router process id  and restart it if closed. (#11)
- [e36615e](https://github.com/kubedb/mysql-router-init/commit/e36615e) Support custom pod and controller labels (#12)
- [48829ef](https://github.com/kubedb/mysql-router-init/commit/48829ef) Fix satori/go.uuid security vulnerability (#10)
- [5f363e8](https://github.com/kubedb/mysql-router-init/commit/5f363e8) Fix jwt-go security vulnerability (#9)
- [41b0fb7](https://github.com/kubedb/mysql-router-init/commit/41b0fb7) Update deps
- [51fc22e](https://github.com/kubedb/mysql-router-init/commit/51fc22e) Fix jwt-go security vulnerability (#8)



## [kubedb/operator](https://github.com/kubedb/operator)

### [v0.23.0](https://github.com/kubedb/operator/releases/tag/v0.23.0)




## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.10.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.10.0)

- [99ac8dca](https://github.com/kubedb/percona-xtradb/commit/99ac8dca) Prepare for release v0.10.0 (#230)
- [5b90ae92](https://github.com/kubedb/percona-xtradb/commit/5b90ae92) Update kmodules.xyz/monitoring-agent-api (#229)
- [13edd56c](https://github.com/kubedb/percona-xtradb/commit/13edd56c) Remove global variable for preconditions (#228)
- [29b4a103](https://github.com/kubedb/percona-xtradb/commit/29b4a103) Update repository config (#227)
- [56b7d005](https://github.com/kubedb/percona-xtradb/commit/56b7d005) Remove docs
- [87f94bb7](https://github.com/kubedb/percona-xtradb/commit/87f94bb7) Use DisableAnalytics flag from license (#226)
- [2f92a7d0](https://github.com/kubedb/percona-xtradb/commit/2f92a7d0) Update license-verifier (#225)
- [11db9761](https://github.com/kubedb/percona-xtradb/commit/11db9761) Update audit and license-verifier version (#223)
- [4026e363](https://github.com/kubedb/percona-xtradb/commit/4026e363) Add pod, statefulSet and service labels support (#224)
- [eb09a518](https://github.com/kubedb/percona-xtradb/commit/eb09a518) Fix satori/go.uuid security vulnerability (#222)
- [0b6063c4](https://github.com/kubedb/percona-xtradb/commit/0b6063c4) Fix jwt-go security vulnerability (#221)
- [ba344a97](https://github.com/kubedb/percona-xtradb/commit/ba344a97) Fix jwt-go security vulnerability (#220)
- [9d3c6e65](https://github.com/kubedb/percona-xtradb/commit/9d3c6e65) Use nats.go v1.13.0 (#219)
- [7dbb955f](https://github.com/kubedb/percona-xtradb/commit/7dbb955f) Setup SiteInfo publisher (#218)
- [eab16c22](https://github.com/kubedb/percona-xtradb/commit/eab16c22) Update dependencies to publish SiteInfo (#217)
- [31e773dd](https://github.com/kubedb/percona-xtradb/commit/31e773dd) Update dependencies to publish SiteInfo (#216)
- [5a2ff511](https://github.com/kubedb/percona-xtradb/commit/5a2ff511) Update repository config (#215)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.7.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.7.0)

- [e81fa81](https://github.com/kubedb/pg-coordinator/commit/e81fa81) Prepare for release v0.7.0 (#54)
- [7c49a84](https://github.com/kubedb/pg-coordinator/commit/7c49a84) Update kmodules.xyz/monitoring-agent-api (#53)
- [aed68ec](https://github.com/kubedb/pg-coordinator/commit/aed68ec) Update repository config (#52)
- [b052255](https://github.com/kubedb/pg-coordinator/commit/b052255) Fix: Raft log corrupted issue (#51)
- [9413347](https://github.com/kubedb/pg-coordinator/commit/9413347) Use DisableAnalytics flag from license (#50)
- [2fe1bfc](https://github.com/kubedb/pg-coordinator/commit/2fe1bfc) Update license-verifier (#49)
- [d6f9afd](https://github.com/kubedb/pg-coordinator/commit/d6f9afd) Support custom pod and controller labels (#48)
- [fb2b48c](https://github.com/kubedb/pg-coordinator/commit/fb2b48c) Postgres Server Restart If Sig-Killed (#44)
- [ab85e39](https://github.com/kubedb/pg-coordinator/commit/ab85e39) Print logs at Debug level
- [9b65232](https://github.com/kubedb/pg-coordinator/commit/9b65232) Log timestamp from zap logger used in raft (#47)
- [6d3eb77](https://github.com/kubedb/pg-coordinator/commit/6d3eb77) Update xorm dependency (#46)
- [b77df43](https://github.com/kubedb/pg-coordinator/commit/b77df43) Fix satori/go.uuid security vulnerability (#45)
- [3cd9cc4](https://github.com/kubedb/pg-coordinator/commit/3cd9cc4) Fix jwt-go security vulnerability (#43)
- [bd2356d](https://github.com/kubedb/pg-coordinator/commit/bd2356d) Fix: Postgres server single user mode start for bullseye image (#42)
- [0c8c18d](https://github.com/kubedb/pg-coordinator/commit/0c8c18d) Update dependencies to publish SiteInfo (#40)
- [06ee14c](https://github.com/kubedb/pg-coordinator/commit/06ee14c) Add support for Postgres version v14.0 (#41)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.10.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.10.0)

- [e12cc8a9](https://github.com/kubedb/pgbouncer/commit/e12cc8a9) Prepare for release v0.10.0 (#190)
- [1e7d783e](https://github.com/kubedb/pgbouncer/commit/1e7d783e) Update kmodules.xyz/monitoring-agent-api (#189)
- [6e08b78b](https://github.com/kubedb/pgbouncer/commit/6e08b78b) Update repository config (#187)
- [ecd28729](https://github.com/kubedb/pgbouncer/commit/ecd28729) Remove global variable for preconditions (#188)
- [e8ad1227](https://github.com/kubedb/pgbouncer/commit/e8ad1227) Remove docs
- [3a2a4143](https://github.com/kubedb/pgbouncer/commit/3a2a4143) Use DisableAnalytics flag from license (#186)
- [308c521f](https://github.com/kubedb/pgbouncer/commit/308c521f) Update license-verifier (#185)
- [a3eb245d](https://github.com/kubedb/pgbouncer/commit/a3eb245d) Update audit and license-verifier version (#184)
- [236cec3c](https://github.com/kubedb/pgbouncer/commit/236cec3c) Support custom pod, service and controller(sts) labels (#183)
- [8a935075](https://github.com/kubedb/pgbouncer/commit/8a935075) Stop using beta apis
- [6f2bce67](https://github.com/kubedb/pgbouncer/commit/6f2bce67) Fix satori/go.uuid security vulnerability (#182)
- [51676d8e](https://github.com/kubedb/pgbouncer/commit/51676d8e) Fix jwt-go security vulnerability (#181)
- [ac2bbd35](https://github.com/kubedb/pgbouncer/commit/ac2bbd35) Fix jwt-go security vulnerability (#180)
- [01c0adc9](https://github.com/kubedb/pgbouncer/commit/01c0adc9) Use nats.go v1.13.0 (#179)
- [3260a07d](https://github.com/kubedb/pgbouncer/commit/3260a07d) Setup SiteInfo publisher (#178)
- [36353a42](https://github.com/kubedb/pgbouncer/commit/36353a42) Update dependencies to publish SiteInfo (#176)
- [ce4fdfc1](https://github.com/kubedb/pgbouncer/commit/ce4fdfc1) Update repository config (#175)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.23.0](https://github.com/kubedb/postgres/releases/tag/v0.23.0)

- [b9b2521a](https://github.com/kubedb/postgres/commit/b9b2521a) Prepare for release v0.23.0 (#540)
- [6f98f884](https://github.com/kubedb/postgres/commit/6f98f884) Update kmodules.xyz/monitoring-agent-api (#539)
- [015dd315](https://github.com/kubedb/postgres/commit/015dd315) Update repository config (#537)
- [1ce33dd4](https://github.com/kubedb/postgres/commit/1ce33dd4) Remove global variable for preconditions (#538)
- [967b1bd5](https://github.com/kubedb/postgres/commit/967b1bd5) Remove docs
- [63585d4d](https://github.com/kubedb/postgres/commit/63585d4d) Use DisableAnalytics flag from license (#536)
- [8030b449](https://github.com/kubedb/postgres/commit/8030b449) Update license-verifier (#535)
- [30407273](https://github.com/kubedb/postgres/commit/30407273) Add pod, services, and pod-controller(sts) labels support (#533)
- [55c626a2](https://github.com/kubedb/postgres/commit/55c626a2) Add Raft client Port In Primary Service (#530)
- [a1a4bdb3](https://github.com/kubedb/postgres/commit/a1a4bdb3) Stop using beta api
- [e0e2a3e4](https://github.com/kubedb/postgres/commit/e0e2a3e4) Update xorm.io/xorm dependency (#532)
- [e6aacd05](https://github.com/kubedb/postgres/commit/e6aacd05) Fix satori/go.uuid security vulnerability (#531)
- [140226f7](https://github.com/kubedb/postgres/commit/140226f7) Fix jwt-go security vulnerability (#529)
- [31e9df33](https://github.com/kubedb/postgres/commit/31e9df33) Fix jwt-go security vulnerability (#528)
- [70fb383a](https://github.com/kubedb/postgres/commit/70fb383a) Use nats.go v1.13.0 (#527)
- [77d43f95](https://github.com/kubedb/postgres/commit/77d43f95) Setup SiteInfo publisher (#526)
- [8755bde2](https://github.com/kubedb/postgres/commit/8755bde2) Update dependencies to publish SiteInfo (#525)
- [feb81410](https://github.com/kubedb/postgres/commit/feb81410) Update repository config (#524)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.10.0](https://github.com/kubedb/proxysql/releases/tag/v0.10.0)

- [88940863](https://github.com/kubedb/proxysql/commit/88940863) Prepare for release v0.10.0 (#207)
- [66ce0801](https://github.com/kubedb/proxysql/commit/66ce0801) Update kmodules.xyz/monitoring-agent-api (#206)
- [21b59886](https://github.com/kubedb/proxysql/commit/21b59886) Remove global variable for preconditions (#205)
- [884e3915](https://github.com/kubedb/proxysql/commit/884e3915) Update repository config (#204)
- [81c11592](https://github.com/kubedb/proxysql/commit/81c11592) Remove docs
- [271bc5af](https://github.com/kubedb/proxysql/commit/271bc5af) Use DisableAnalytics flag from license (#203)
- [4710a672](https://github.com/kubedb/proxysql/commit/4710a672) Update license-verifier (#202)
- [229ba8c7](https://github.com/kubedb/proxysql/commit/229ba8c7) Support custom pod, service and controller(sts) labels (#201)
- [3c915f61](https://github.com/kubedb/proxysql/commit/3c915f61) Update dependencies (#200)
- [7ce88a70](https://github.com/kubedb/proxysql/commit/7ce88a70) Fix jwt-go security vulnerability (#199)
- [bb2c78e8](https://github.com/kubedb/proxysql/commit/bb2c78e8) Fix jwt-go security vulnerability (#198)
- [2764f4c7](https://github.com/kubedb/proxysql/commit/2764f4c7) Use nats.go v1.13.0 (#197)
- [b06f614b](https://github.com/kubedb/proxysql/commit/b06f614b) Update dependencies to publish SiteInfo (#196)
- [6a067416](https://github.com/kubedb/proxysql/commit/6a067416) Update dependencies to publish SiteInfo (#195)
- [5f1ce0f2](https://github.com/kubedb/proxysql/commit/5f1ce0f2) Update repository config (#194)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.16.0](https://github.com/kubedb/redis/releases/tag/v0.16.0)

- [c3986f47](https://github.com/kubedb/redis/commit/c3986f47) Prepare for release v0.16.0 (#362)
- [158af05f](https://github.com/kubedb/redis/commit/158af05f) Update kmodules.xyz/monitoring-agent-api (#361)
- [4cc13143](https://github.com/kubedb/redis/commit/4cc13143) Remove global variable for preconditions (#360)
- [16011733](https://github.com/kubedb/redis/commit/16011733) Update repository config (#359)
- [eea15b8a](https://github.com/kubedb/redis/commit/eea15b8a) Fix: Sentinel and Redis In Different Namespaces (#358)
- [38b28c4e](https://github.com/kubedb/redis/commit/38b28c4e) Remove docs
- [32b2565a](https://github.com/kubedb/redis/commit/32b2565a) Use DisableAnalytics flag from license (#357)
- [27d7f428](https://github.com/kubedb/redis/commit/27d7f428) Update license-verifier (#356)
- [c00f72bf](https://github.com/kubedb/redis/commit/c00f72bf) Update audit and license-verifier version (#354)
- [7aebec13](https://github.com/kubedb/redis/commit/7aebec13) Add pod, statefulSet and service labels support (#355)
- [2f09ae66](https://github.com/kubedb/redis/commit/2f09ae66) Fix: resolve panic issue when sentinelRef is Null or empty (#353)
- [016fc0ff](https://github.com/kubedb/redis/commit/016fc0ff) Redis Custom Config issue (#351)
- [09d750ac](https://github.com/kubedb/redis/commit/09d750ac) Update dependencies (#352)
- [4ac3e812](https://github.com/kubedb/redis/commit/4ac3e812) Fix jwt-go security vulnerability (#350)
- [4f7fd873](https://github.com/kubedb/redis/commit/4f7fd873) Fix jwt-go security vulnerability (#349)
- [f86d4fb1](https://github.com/kubedb/redis/commit/f86d4fb1) Fix: Redis Panic issue for sentinel (#348)
- [7b1c53a6](https://github.com/kubedb/redis/commit/7b1c53a6) Use nats.go v1.13.0 (#347)
- [7d9017e8](https://github.com/kubedb/redis/commit/7d9017e8) Setup SiteInfo publisher (#346)
- [34c98fc3](https://github.com/kubedb/redis/commit/34c98fc3) Update dependencies to publish SiteInfo (#345)
- [1831c5b7](https://github.com/kubedb/redis/commit/1831c5b7) Update dependencies to publish SiteInfo (#344)
- [4798de2d](https://github.com/kubedb/redis/commit/4798de2d) Update repository config (#343)
- [a14cd630](https://github.com/kubedb/redis/commit/a14cd630) Fix: Redis monitoring port (#342)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.2.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.2.0)

- [8b9d7eb](https://github.com/kubedb/redis-coordinator/commit/8b9d7eb) Prepare for release v0.2.0 (#13)
- [3399280](https://github.com/kubedb/redis-coordinator/commit/3399280) Update kmodules.xyz/monitoring-agent-api (#12)
- [eb51783](https://github.com/kubedb/redis-coordinator/commit/eb51783) Update repository config (#11)
- [fff31b5](https://github.com/kubedb/redis-coordinator/commit/fff31b5) Use DisableAnalytics flag from license (#10)
- [f2b347c](https://github.com/kubedb/redis-coordinator/commit/f2b347c) Update license-verifier (#9)
- [361e3f7](https://github.com/kubedb/redis-coordinator/commit/361e3f7) Support custom pod and controller labels (#8)
- [ad486b9](https://github.com/kubedb/redis-coordinator/commit/ad486b9) Update dependencies (#7)
- [560e04d](https://github.com/kubedb/redis-coordinator/commit/560e04d) Fix satori/go.uuid security vulnerability (#6)
- [a0bd03b](https://github.com/kubedb/redis-coordinator/commit/a0bd03b) Fix jwt-go security vulnerability (#5)
- [6a1f913](https://github.com/kubedb/redis-coordinator/commit/6a1f913) Fix jwt-go security vulnerability (#4)
- [8f1418e](https://github.com/kubedb/redis-coordinator/commit/8f1418e) Use nats.go v1.13.0 (#3)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.10.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.10.0)

- [bde31cc8](https://github.com/kubedb/replication-mode-detector/commit/bde31cc8) Prepare for release v0.10.0 (#175)
- [93abeed6](https://github.com/kubedb/replication-mode-detector/commit/93abeed6) Update kmodules.xyz/monitoring-agent-api (#174)
- [78bff385](https://github.com/kubedb/replication-mode-detector/commit/78bff385) Update repository config (#173)
- [6578cc86](https://github.com/kubedb/replication-mode-detector/commit/6578cc86) Use DisableAnalytics flag from license (#172)
- [b99f779b](https://github.com/kubedb/replication-mode-detector/commit/b99f779b) Update license-verifier (#171)
- [a62adbd0](https://github.com/kubedb/replication-mode-detector/commit/a62adbd0) Support custom pod and controller labels (#170)
- [afb2bfd9](https://github.com/kubedb/replication-mode-detector/commit/afb2bfd9) Update dependencies (#169)
- [9b65b2c5](https://github.com/kubedb/replication-mode-detector/commit/9b65b2c5) Update xorm dependency (#168)
- [a2427a67](https://github.com/kubedb/replication-mode-detector/commit/a2427a67) Fix satori/go.uuid security vulnerability (#167)
- [0a0163ca](https://github.com/kubedb/replication-mode-detector/commit/0a0163ca) Fix jwt-go security vulnerability (#166)
- [4f69c8c3](https://github.com/kubedb/replication-mode-detector/commit/4f69c8c3) Fix jwt-go security vulnerability (#165)
- [1005f1a2](https://github.com/kubedb/replication-mode-detector/commit/1005f1a2) Update dependencies to publish SiteInfo (#164)
- [9ac9d09e](https://github.com/kubedb/replication-mode-detector/commit/9ac9d09e) Update dependencies to publish SiteInfo (#163)
- [c55ae055](https://github.com/kubedb/replication-mode-detector/commit/c55ae055) Update repository config (#162)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.8.0](https://github.com/kubedb/tests/releases/tag/v0.8.0)

- [50f414a9](https://github.com/kubedb/tests/commit/50f414a9) Prepare for release v0.8.0 (#156)
- [25e5c5b6](https://github.com/kubedb/tests/commit/25e5c5b6) Update kmodules.xyz/monitoring-agent-api (#155)
- [951d17ca](https://github.com/kubedb/tests/commit/951d17ca) Update repository config (#154)
- [d0988abf](https://github.com/kubedb/tests/commit/d0988abf) Use DisableAnalytics flag from license (#153)
- [7aea8907](https://github.com/kubedb/tests/commit/7aea8907) Update license-verifier (#152)
- [637ae1a0](https://github.com/kubedb/tests/commit/637ae1a0) Support custom pod and controller labels (#151)
- [45223290](https://github.com/kubedb/tests/commit/45223290) Update dependencies (#150)
- [fcef9222](https://github.com/kubedb/tests/commit/fcef9222) Fix satori/go.uuid security vulnerability (#149)
- [1d308fc7](https://github.com/kubedb/tests/commit/1d308fc7) Fix jwt-go security vulnerability (#148)
- [9c764d48](https://github.com/kubedb/tests/commit/9c764d48) Fix jwt-go security vulnerability (#147)
- [cef34499](https://github.com/kubedb/tests/commit/cef34499) Fix jwt-go security vulnerability (#146)
- [2c1d6094](https://github.com/kubedb/tests/commit/2c1d6094) Update dependencies to publish SiteInfo (#145)
- [443c8390](https://github.com/kubedb/tests/commit/443c8390) Update dependencies to publish SiteInfo (#144)
- [5f71478f](https://github.com/kubedb/tests/commit/5f71478f) Update repository config (#143)




