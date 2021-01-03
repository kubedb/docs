---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2021.01.02-rc.0
    name: Changelog-v2021.01.02-rc.0
    parent: welcome
    weight: 20210102
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2021.01.02-rc.0/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2021.01.02-rc.0/
---

# KubeDB v2021.01.02-rc.0 (2021-01-03)


## [appscode/kubedb-autoscaler](https://github.com/appscode/kubedb-autoscaler)

### [v0.1.0-rc.0](https://github.com/appscode/kubedb-autoscaler/releases/tag/v0.1.0-rc.0)

- [f346d5e](https://github.com/appscode/kubedb-autoscaler/commit/f346d5e) Prepare for release v0.0.1-rc.0 (#5)
- [bd5dbd9](https://github.com/appscode/kubedb-autoscaler/commit/bd5dbd9) Remove extra informers (#4)
- [9b461a5](https://github.com/appscode/kubedb-autoscaler/commit/9b461a5) Enable GitHub Actions (#6)
- [de39ed0](https://github.com/appscode/kubedb-autoscaler/commit/de39ed0) Update license header (#7)
- [5518680](https://github.com/appscode/kubedb-autoscaler/commit/5518680) Remove validators and enable ES autoscaler (#3)
- [c0d65f4](https://github.com/appscode/kubedb-autoscaler/commit/c0d65f4) Add `inMemory` configuration in vertical scaling (#2)
- [088777c](https://github.com/appscode/kubedb-autoscaler/commit/088777c) Add Elasticsearch Autoscaler Controller (#1)
- [779a2d2](https://github.com/appscode/kubedb-autoscaler/commit/779a2d2) Add Conditions
- [cce0828](https://github.com/appscode/kubedb-autoscaler/commit/cce0828) Update Makefile for install and uninstall
- [04c9f28](https://github.com/appscode/kubedb-autoscaler/commit/04c9f28) Remove some prometheus flags
- [118284a](https://github.com/appscode/kubedb-autoscaler/commit/118284a) Refactor some common code
- [bdf8d89](https://github.com/appscode/kubedb-autoscaler/commit/bdf8d89) Fix Webhook
- [2934025](https://github.com/appscode/kubedb-autoscaler/commit/2934025) Handle empty prometheus vector
- [c718118](https://github.com/appscode/kubedb-autoscaler/commit/c718118) Fix Trigger
- [b795a24](https://github.com/appscode/kubedb-autoscaler/commit/b795a24) Update Prometheus Client
- [20c69c1](https://github.com/appscode/kubedb-autoscaler/commit/20c69c1) Add MongoDBAutoscaler CRD
- [6c2c2be](https://github.com/appscode/kubedb-autoscaler/commit/6c2c2be) Add Storage Auto Scaler



## [appscode/kubedb-enterprise](https://github.com/appscode/kubedb-enterprise)

### [v0.3.0-rc.0](https://github.com/appscode/kubedb-enterprise/releases/tag/v0.3.0-rc.0)

- [a9ed2e6a](https://github.com/appscode/kubedb-enterprise/commit/a9ed2e6a) Prepare for release v0.3.0-rc.0 (#109)
- [d62bdf40](https://github.com/appscode/kubedb-enterprise/commit/d62bdf40) Change offshoot selector labels to standard k8s app labels (#96)
- [137b1d11](https://github.com/appscode/kubedb-enterprise/commit/137b1d11) Add evict pods in MongoDB (#106)



## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.16.0-rc.0](https://github.com/kubedb/apimachinery/releases/tag/v0.16.0-rc.0)

- [e4cb7ef9](https://github.com/kubedb/apimachinery/commit/e4cb7ef9) MySQL primary service dns helper (#677)
- [2469f17e](https://github.com/kubedb/apimachinery/commit/2469f17e) Add constants for Elasticsearch TLS reconfiguration (#672)
- [2c61fb41](https://github.com/kubedb/apimachinery/commit/2c61fb41) Add MongoDB constants (#676)
- [31584e58](https://github.com/kubedb/apimachinery/commit/31584e58) Add DB constants and tls-reconfigure checker func (#657)
- [4d67bea1](https://github.com/kubedb/apimachinery/commit/4d67bea1) Add MongoDB & Elasticsearch Autoscaler CRDs (#659)
- [fb88afcf](https://github.com/kubedb/apimachinery/commit/fb88afcf) Update Kubernetes v1.18.9 dependencies (#675)
- [56a61c7f](https://github.com/kubedb/apimachinery/commit/56a61c7f) Change default resource limits to 1Gi ram and 500m cpu (#674)
- [a36050ca](https://github.com/kubedb/apimachinery/commit/a36050ca) Invoke update handler on labels or annotations change
- [37c68bd0](https://github.com/kubedb/apimachinery/commit/37c68bd0) Change offshoot selector labels to standard k8s app labels (#673)
- [83fb66c2](https://github.com/kubedb/apimachinery/commit/83fb66c2) Add redis constants and an address function (#663)
- [2c0e6319](https://github.com/kubedb/apimachinery/commit/2c0e6319) Add support for Elasticsearch volume expansion (#666)
- [d16f40aa](https://github.com/kubedb/apimachinery/commit/d16f40aa) Add changes to Elasticsearch vertical scaling spec (#662)
- [938147c4](https://github.com/kubedb/apimachinery/commit/938147c4) Add Elasticsearch scaling constants (#658)
- [b1641bdf](https://github.com/kubedb/apimachinery/commit/b1641bdf) Update for release Stash@v2020.12.17 (#671)
- [d37718a2](https://github.com/kubedb/apimachinery/commit/d37718a2) Remove doNotPause logic from namespace validator (#669)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.16.0-rc.0](https://github.com/kubedb/cli/releases/tag/v0.16.0-rc.0)

- [2a3bc5a8](https://github.com/kubedb/cli/commit/2a3bc5a8) Prepare for release v0.16.0-rc.0 (#575)
- [500b142a](https://github.com/kubedb/cli/commit/500b142a) Update KubeDB api (#574)
- [8208fcf1](https://github.com/kubedb/cli/commit/8208fcf1) Update KubeDB api (#573)
- [59ac94e7](https://github.com/kubedb/cli/commit/59ac94e7) Update Kubernetes v1.18.9 dependencies (#572)
- [1ebd0633](https://github.com/kubedb/cli/commit/1ebd0633) Update KubeDB api (#571)
- [0ccba4d1](https://github.com/kubedb/cli/commit/0ccba4d1) Update KubeDB api (#570)
- [770f94be](https://github.com/kubedb/cli/commit/770f94be) Update KubeDB api (#569)
- [fbdcce08](https://github.com/kubedb/cli/commit/fbdcce08) Update KubeDB api (#568)
- [93b038e9](https://github.com/kubedb/cli/commit/93b038e9) Update KubeDB api (#567)
- [ef758783](https://github.com/kubedb/cli/commit/ef758783) Update for release Stash@v2020.12.17 (#566)
- [07fa4a7e](https://github.com/kubedb/cli/commit/07fa4a7e) Update KubeDB api (#565)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.16.0-rc.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.16.0-rc.0)

- [9961f623](https://github.com/kubedb/elasticsearch/commit/9961f623) Prepare for release v0.16.0-rc.0 (#450)
- [e7d84a5f](https://github.com/kubedb/elasticsearch/commit/e7d84a5f) Update KubeDB api (#449)
- [7a40f5a5](https://github.com/kubedb/elasticsearch/commit/7a40f5a5) Update KubeDB api (#448)
- [c680498d](https://github.com/kubedb/elasticsearch/commit/c680498d) Update Kubernetes v1.18.9 dependencies (#447)
- [e28277d8](https://github.com/kubedb/elasticsearch/commit/e28277d8) Update KubeDB api (#446)
- [21f98151](https://github.com/kubedb/elasticsearch/commit/21f98151) Fix annotations passing to AppBinding (#445)
- [6c7ff056](https://github.com/kubedb/elasticsearch/commit/6c7ff056) Use StatefulSet naming methods (#430)
- [23a53309](https://github.com/kubedb/elasticsearch/commit/23a53309) Update KubeDB api (#444)
- [a4217edf](https://github.com/kubedb/elasticsearch/commit/a4217edf) Change offshoot selector labels to standard k8s app labels (#442)
- [6535adff](https://github.com/kubedb/elasticsearch/commit/6535adff) Delete tests moved to tests repo (#443)
- [ca2b5be5](https://github.com/kubedb/elasticsearch/commit/ca2b5be5) Update KubeDB api (#441)
- [ce19a83e](https://github.com/kubedb/elasticsearch/commit/ce19a83e) Update KubeDB api (#440)
- [662902a9](https://github.com/kubedb/elasticsearch/commit/662902a9) Update immutable field list (#435)
- [efe804c9](https://github.com/kubedb/elasticsearch/commit/efe804c9) Update KubeDB api (#438)
- [6ac3eb02](https://github.com/kubedb/elasticsearch/commit/6ac3eb02) Update for release Stash@v2020.12.17 (#437)
- [1da53ab9](https://github.com/kubedb/elasticsearch/commit/1da53ab9) Update KubeDB api (#436)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v0.16.0-rc.0](https://github.com/kubedb/installer/releases/tag/v0.16.0-rc.0)

- [feb4a3f](https://github.com/kubedb/installer/commit/feb4a3f) Prepare for release v0.16.0-rc.0 (#218)
- [7e17d4d](https://github.com/kubedb/installer/commit/7e17d4d) Add kubedb-autoscaler chart (#137)
- [fe87336](https://github.com/kubedb/installer/commit/fe87336) Rename gerbage-collector-rbac.yaml to garbage-collector-rbac.yaml
- [5630a5e](https://github.com/kubedb/installer/commit/5630a5e) Use kmodules.xyz/schema-checker to validate values schema (#217)
- [e22e67e](https://github.com/kubedb/installer/commit/e22e67e) Update repository config (#215)
- [3ded17a](https://github.com/kubedb/installer/commit/3ded17a) Update Kubernetes v1.18.9 dependencies (#214)
- [cb9a295](https://github.com/kubedb/installer/commit/cb9a295) Add enforceTerminationPolicy (#212)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.9.0-rc.0](https://github.com/kubedb/memcached/releases/tag/v0.9.0-rc.0)

- [33752041](https://github.com/kubedb/memcached/commit/33752041) Prepare for release v0.9.0-rc.0 (#269)
- [9cf96e13](https://github.com/kubedb/memcached/commit/9cf96e13) Update KubeDB api (#268)
- [0bfe24df](https://github.com/kubedb/memcached/commit/0bfe24df) Update KubeDB api (#267)
- [29fc8f33](https://github.com/kubedb/memcached/commit/29fc8f33) Update Kubernetes v1.18.9 dependencies (#266)
- [c9dfe14c](https://github.com/kubedb/memcached/commit/c9dfe14c) Update KubeDB api (#265)
- [f75073c9](https://github.com/kubedb/memcached/commit/f75073c9) Fix annotations passing to AppBinding (#264)
- [28cdfdfd](https://github.com/kubedb/memcached/commit/28cdfdfd) Initialize mapper
- [6a9243ab](https://github.com/kubedb/memcached/commit/6a9243ab) Change offshoot selector labels to standard k8s app labels (#263)
- [e838aec4](https://github.com/kubedb/memcached/commit/e838aec4) Update KubeDB api (#262)
- [88654cdd](https://github.com/kubedb/memcached/commit/88654cdd) Update KubeDB api (#261)
- [c2fb7c2f](https://github.com/kubedb/memcached/commit/c2fb7c2f) Update KubeDB api (#260)
- [5cc2cf17](https://github.com/kubedb/memcached/commit/5cc2cf17) Update KubeDB api (#259)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.9.0-rc.0](https://github.com/kubedb/mongodb/releases/tag/v0.9.0-rc.0)

- [ee410983](https://github.com/kubedb/mongodb/commit/ee410983) Prepare for release v0.9.0-rc.0 (#348)
- [b39b664b](https://github.com/kubedb/mongodb/commit/b39b664b) Update KubeDB api (#347)
- [84e007fe](https://github.com/kubedb/mongodb/commit/84e007fe) Update KubeDB api (#346)
- [e8aa1f8a](https://github.com/kubedb/mongodb/commit/e8aa1f8a) Close connections when operation completes (#338)
- [1ec2a2c7](https://github.com/kubedb/mongodb/commit/1ec2a2c7) Update Kubernetes v1.18.9 dependencies (#345)
- [7306fb26](https://github.com/kubedb/mongodb/commit/7306fb26) Update KubeDB api (#344)
- [efa62a85](https://github.com/kubedb/mongodb/commit/efa62a85) Fix annotations passing to AppBinding (#342)
- [9d88e69e](https://github.com/kubedb/mongodb/commit/9d88e69e) Remove `inMemory` setting from Config Server (#343)
- [32b96d12](https://github.com/kubedb/mongodb/commit/32b96d12) Change offshoot selector labels to standard k8s app labels (#341)
- [67fcdbf4](https://github.com/kubedb/mongodb/commit/67fcdbf4) Update KubeDB api (#340)
- [cf2c0778](https://github.com/kubedb/mongodb/commit/cf2c0778) Update KubeDB api (#339)
- [232a4a00](https://github.com/kubedb/mongodb/commit/232a4a00) Update KubeDB api (#337)
- [0a1307e7](https://github.com/kubedb/mongodb/commit/0a1307e7) Update for release Stash@v2020.12.17 (#336)
- [89b4e4fc](https://github.com/kubedb/mongodb/commit/89b4e4fc) Update KubeDB api (#335)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.9.0-rc.0](https://github.com/kubedb/mysql/releases/tag/v0.9.0-rc.0)

- [ad9d9879](https://github.com/kubedb/mysql/commit/ad9d9879) Prepare for release v0.9.0-rc.0 (#337)
- [a9e9d1f7](https://github.com/kubedb/mysql/commit/a9e9d1f7) Fix args for TLS (#336)
- [9dd89572](https://github.com/kubedb/mysql/commit/9dd89572) Update KubeDB api (#335)
- [29ff2c57](https://github.com/kubedb/mysql/commit/29ff2c57) Fixes DB Health Checker and StatefulSet Patch (#322)
- [47470895](https://github.com/kubedb/mysql/commit/47470895) Remove unnecessary StatefulSet waitloop (#331)
- [3aec8f59](https://github.com/kubedb/mysql/commit/3aec8f59) Update Kubernetes v1.18.9 dependencies (#334)
- [c1ca980d](https://github.com/kubedb/mysql/commit/c1ca980d) Update KubeDB api (#333)
- [96f4b59c](https://github.com/kubedb/mysql/commit/96f4b59c) Fix annotations passing to AppBinding (#332)
- [76f371a2](https://github.com/kubedb/mysql/commit/76f371a2) Change offshoot selector labels to standard k8s app labels (#329)
- [aa3d6b6f](https://github.com/kubedb/mysql/commit/aa3d6b6f) Delete tests moved to tests repo (#330)
- [6c544d2c](https://github.com/kubedb/mysql/commit/6c544d2c) Update KubeDB api (#328)
- [fe03a36c](https://github.com/kubedb/mysql/commit/fe03a36c) Update KubeDB api (#327)
- [29fd7474](https://github.com/kubedb/mysql/commit/29fd7474) Use basic-auth secret type for auth secret (#326)
- [90457549](https://github.com/kubedb/mysql/commit/90457549) Update KubeDB api (#325)
- [1487f15e](https://github.com/kubedb/mysql/commit/1487f15e) Update for release Stash@v2020.12.17 (#324)
- [2d7fa549](https://github.com/kubedb/mysql/commit/2d7fa549) Update KubeDB api (#323)



## [kubedb/operator](https://github.com/kubedb/operator)

### [v0.16.0-rc.0](https://github.com/kubedb/operator/releases/tag/v0.16.0-rc.0)

- [3ee052dc](https://github.com/kubedb/operator/commit/3ee052dc) Prepare for release v0.16.0-rc.0 (#376)
- [dbb5195b](https://github.com/kubedb/operator/commit/dbb5195b) Update KubeDB api (#375)
- [4b162e08](https://github.com/kubedb/operator/commit/4b162e08) Update KubeDB api (#374)
- [39762b0f](https://github.com/kubedb/operator/commit/39762b0f) Update KubeDB api (#373)
- [d6a2cf27](https://github.com/kubedb/operator/commit/d6a2cf27) Change offshoot selector labels to standard k8s app labels (#372)
- [36a8ab6f](https://github.com/kubedb/operator/commit/36a8ab6f) Update Kubernetes v1.18.9 dependencies (#371)
- [554638e0](https://github.com/kubedb/operator/commit/554638e0) Update KubeDB api (#369)
- [8c7ef91d](https://github.com/kubedb/operator/commit/8c7ef91d) Update KubeDB api (#368)
- [dd96574e](https://github.com/kubedb/operator/commit/dd96574e) Update KubeDB api (#367)
- [eef04de1](https://github.com/kubedb/operator/commit/eef04de1) Update KubeDB api (#366)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.3.0-rc.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.3.0-rc.0)

- [f545beb4](https://github.com/kubedb/percona-xtradb/commit/f545beb4) Prepare for release v0.3.0-rc.0 (#166)
- [c5d0c826](https://github.com/kubedb/percona-xtradb/commit/c5d0c826) Update KubeDB api (#164)
- [b3da5757](https://github.com/kubedb/percona-xtradb/commit/b3da5757) Fix annotations passing to AppBinding (#163)
- [7aeaee74](https://github.com/kubedb/percona-xtradb/commit/7aeaee74) Change offshoot selector labels to standard k8s app labels (#161)
- [a36ffa87](https://github.com/kubedb/percona-xtradb/commit/a36ffa87) Update Kubernetes v1.18.9 dependencies (#162)
- [fa3a2a9d](https://github.com/kubedb/percona-xtradb/commit/fa3a2a9d) Update KubeDB api (#160)
- [a1db6821](https://github.com/kubedb/percona-xtradb/commit/a1db6821) Update KubeDB api (#159)
- [4357b18a](https://github.com/kubedb/percona-xtradb/commit/4357b18a) Use basic-auth secret type for auth secret (#158)
- [f9ccfc4e](https://github.com/kubedb/percona-xtradb/commit/f9ccfc4e) Update KubeDB api (#157)
- [11739165](https://github.com/kubedb/percona-xtradb/commit/11739165) Update for release Stash@v2020.12.17 (#156)
- [80bf041c](https://github.com/kubedb/percona-xtradb/commit/80bf041c) Update KubeDB api (#155)



## [kubedb/pg-leader-election](https://github.com/kubedb/pg-leader-election)

### [v0.4.0-rc.0](https://github.com/kubedb/pg-leader-election/releases/tag/v0.4.0-rc.0)

- [31050c1](https://github.com/kubedb/pg-leader-election/commit/31050c1) Update KubeDB api (#44)
- [dc786b7](https://github.com/kubedb/pg-leader-election/commit/dc786b7) Update KubeDB api (#43)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.3.0-rc.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.3.0-rc.0)

- [51c8fee2](https://github.com/kubedb/pgbouncer/commit/51c8fee2) Prepare for release v0.3.0-rc.0 (#132)
- [fded227a](https://github.com/kubedb/pgbouncer/commit/fded227a) Update KubeDB api (#130)
- [7702e10a](https://github.com/kubedb/pgbouncer/commit/7702e10a) Change offshoot selector labels to standard k8s app labels (#128)
- [2ba5284c](https://github.com/kubedb/pgbouncer/commit/2ba5284c) Update Kubernetes v1.18.9 dependencies (#129)
- [3507a96c](https://github.com/kubedb/pgbouncer/commit/3507a96c) Update KubeDB api (#127)
- [fc8330e4](https://github.com/kubedb/pgbouncer/commit/fc8330e4) Update KubeDB api (#126)
- [3e9b4e77](https://github.com/kubedb/pgbouncer/commit/3e9b4e77) Update KubeDB api (#125)
- [6c85ca6a](https://github.com/kubedb/pgbouncer/commit/6c85ca6a) Update KubeDB api (#124)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.16.0-rc.0](https://github.com/kubedb/postgres/releases/tag/v0.16.0-rc.0)

- [c7b618f5](https://github.com/kubedb/postgres/commit/c7b618f5) Prepare for release v0.16.0-rc.0 (#452)
- [be060733](https://github.com/kubedb/postgres/commit/be060733) Update KubeDB api (#451)
- [d2d2f32c](https://github.com/kubedb/postgres/commit/d2d2f32c) Update KubeDB api (#450)
- [ed375b2b](https://github.com/kubedb/postgres/commit/ed375b2b) Update KubeDB api (#449)
- [a3940790](https://github.com/kubedb/postgres/commit/a3940790) Fix annotations passing to AppBinding (#448)
- [f0b5a9dd](https://github.com/kubedb/postgres/commit/f0b5a9dd) Change offshoot selector labels to standard k8s app labels (#447)
- [eb4f80ab](https://github.com/kubedb/postgres/commit/eb4f80ab) Update KubeDB api (#446)
- [c9075b5a](https://github.com/kubedb/postgres/commit/c9075b5a) Update KubeDB api (#445)
- [a04891e1](https://github.com/kubedb/postgres/commit/a04891e1) Use basic-auth secret type for auth secret (#444)
- [e7503eec](https://github.com/kubedb/postgres/commit/e7503eec) Update KubeDB api (#443)
- [0eb3a1b9](https://github.com/kubedb/postgres/commit/0eb3a1b9) Update for release Stash@v2020.12.17 (#442)
- [c3ea786d](https://github.com/kubedb/postgres/commit/c3ea786d) Update KubeDB api (#441)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.3.0-rc.0](https://github.com/kubedb/proxysql/releases/tag/v0.3.0-rc.0)

- [1ae8aed1](https://github.com/kubedb/proxysql/commit/1ae8aed1) Prepare for release v0.3.0-rc.0 (#147)
- [0e60bddf](https://github.com/kubedb/proxysql/commit/0e60bddf) Update KubeDB api (#145)
- [df11880c](https://github.com/kubedb/proxysql/commit/df11880c) Change offshoot selector labels to standard k8s app labels (#143)
- [540bdea2](https://github.com/kubedb/proxysql/commit/540bdea2) Update Kubernetes v1.18.9 dependencies (#144)
- [52907cb4](https://github.com/kubedb/proxysql/commit/52907cb4) Update KubeDB api (#142)
- [d1686708](https://github.com/kubedb/proxysql/commit/d1686708) Update KubeDB api (#141)
- [e5e2a798](https://github.com/kubedb/proxysql/commit/e5e2a798) Use basic-auth secret type for auth secret (#140)
- [8cf2a9e4](https://github.com/kubedb/proxysql/commit/8cf2a9e4) Update KubeDB api (#139)
- [7b0cdb0f](https://github.com/kubedb/proxysql/commit/7b0cdb0f) Update for release Stash@v2020.12.17 (#138)
- [ce7136a1](https://github.com/kubedb/proxysql/commit/ce7136a1) Update KubeDB api (#137)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.9.0-rc.0](https://github.com/kubedb/redis/releases/tag/v0.9.0-rc.0)

- [b416a016](https://github.com/kubedb/redis/commit/b416a016) Prepare for release v0.9.0-rc.0 (#290)
- [751b8f6b](https://github.com/kubedb/redis/commit/751b8f6b) Update KubeDB api (#289)
- [0affafe9](https://github.com/kubedb/redis/commit/0affafe9) Update KubeDB api (#287)
- [665d6b4f](https://github.com/kubedb/redis/commit/665d6b4f) Remove tests moved to kubedb/tests (#288)
- [6c254e3b](https://github.com/kubedb/redis/commit/6c254e3b) Update KubeDB api (#286)
- [1b73def3](https://github.com/kubedb/redis/commit/1b73def3) Fix annotations passing to AppBinding (#285)
- [dc349058](https://github.com/kubedb/redis/commit/dc349058) Update KubeDB api (#283)
- [7d47e506](https://github.com/kubedb/redis/commit/7d47e506) Change offshoot selector labels to standard k8s app labels (#282)
- [f8f7570f](https://github.com/kubedb/redis/commit/f8f7570f) Update Kubernetes v1.18.9 dependencies (#284)
- [63cb769d](https://github.com/kubedb/redis/commit/63cb769d) Update KubeDB api (#281)
- [19ec4460](https://github.com/kubedb/redis/commit/19ec4460) Update KubeDB api (#280)
- [af67e190](https://github.com/kubedb/redis/commit/af67e190) Update KubeDB api (#279)
- [4b89034c](https://github.com/kubedb/redis/commit/4b89034c) Update KubeDB api (#278)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.3.0-rc.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.3.0-rc.0)

- [179e153](https://github.com/kubedb/replication-mode-detector/commit/179e153) Prepare for release v0.3.0-rc.0 (#115)
- [d47023b](https://github.com/kubedb/replication-mode-detector/commit/d47023b) Update KubeDB api (#114)
- [3e5db31](https://github.com/kubedb/replication-mode-detector/commit/3e5db31) Update KubeDB api (#113)
- [987f068](https://github.com/kubedb/replication-mode-detector/commit/987f068) Change offshoot selector labels to standard k8s app labels (#110)
- [21fc76f](https://github.com/kubedb/replication-mode-detector/commit/21fc76f) Update Kubernetes v1.18.9 dependencies (#112)
- [db85cbd](https://github.com/kubedb/replication-mode-detector/commit/db85cbd) Close database connection when operation completes (#107)
- [740d1d8](https://github.com/kubedb/replication-mode-detector/commit/740d1d8) Update Kubernetes v1.18.9 dependencies (#111)
- [6f228a5](https://github.com/kubedb/replication-mode-detector/commit/6f228a5) Update KubeDB api (#109)
- [256ea7a](https://github.com/kubedb/replication-mode-detector/commit/256ea7a) Update KubeDB api (#108)
- [7a9acc0](https://github.com/kubedb/replication-mode-detector/commit/7a9acc0) Update KubeDB api (#106)
- [21a18c2](https://github.com/kubedb/replication-mode-detector/commit/21a18c2) Update KubeDB api (#105)




