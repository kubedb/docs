---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2021.03.11
    name: Changelog-v2021.03.11
    parent: welcome
    weight: 20210311
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2021.03.11/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2021.03.11/
---

# KubeDB v2021.03.11 (2021-03-11)


## [appscode/kubedb-autoscaler](https://github.com/appscode/kubedb-autoscaler)

### [v0.2.0](https://github.com/appscode/kubedb-autoscaler/releases/tag/v0.2.0)

- [f93c060](https://github.com/appscode/kubedb-autoscaler/commit/f93c060) Prepare for release v0.2.0 (#18)
- [b86b36a](https://github.com/appscode/kubedb-autoscaler/commit/b86b36a) Update dependencies
- [3e50e33](https://github.com/appscode/kubedb-autoscaler/commit/3e50e33) Update repository config (#17)
- [efd6d82](https://github.com/appscode/kubedb-autoscaler/commit/efd6d82) Update repository config (#16)
- [eddccf2](https://github.com/appscode/kubedb-autoscaler/commit/eddccf2) Update repository config (#14)



## [appscode/kubedb-enterprise](https://github.com/appscode/kubedb-enterprise)

### [v0.4.0](https://github.com/appscode/kubedb-enterprise/releases/tag/v0.4.0)

- [6bcddad8](https://github.com/appscode/kubedb-enterprise/commit/6bcddad8) Prepare for release v0.4.0 (#154)
- [0785c36e](https://github.com/appscode/kubedb-enterprise/commit/0785c36e) Fix ConfigServer Horizontal Scaling Up (#153)
- [0784a195](https://github.com/appscode/kubedb-enterprise/commit/0784a195) Fix MySQL DB version patch (#152)
- [958b2390](https://github.com/appscode/kubedb-enterprise/commit/958b2390) Register CRD for MariaDB (#150)
- [32caa479](https://github.com/appscode/kubedb-enterprise/commit/32caa479) TLS support for PostgreSQL and pg-coordinator (#148)
- [03201b02](https://github.com/appscode/kubedb-enterprise/commit/03201b02) Add MariaDB TLS support (#110)
- [1ad3e7df](https://github.com/appscode/kubedb-enterprise/commit/1ad3e7df) Add redis & tls reconfigure, restart support (#98)
- [dad9c4cc](https://github.com/appscode/kubedb-enterprise/commit/dad9c4cc) Fix MongoDB Reconfigure TLS (#143)
- [f62fc9f4](https://github.com/appscode/kubedb-enterprise/commit/f62fc9f4) Update KubeDB api (#149)
- [c86c8d0c](https://github.com/appscode/kubedb-enterprise/commit/c86c8d0c) Update old env with the new one while upgrading ES version 6 to 7 (#147)
- [48af0d99](https://github.com/appscode/kubedb-enterprise/commit/48af0d99) Use Elasticsearch version from version CRD while creating client (#135)
- [70048682](https://github.com/appscode/kubedb-enterprise/commit/70048682) Fix MySQL Reconfigure TLS (#144)
- [7a45302b](https://github.com/appscode/kubedb-enterprise/commit/7a45302b) Fix MySQL major version upgrading (#134)
- [a5f76ab0](https://github.com/appscode/kubedb-enterprise/commit/a5f76ab0) Fix install command in Makefile (#145)
- [34ae3519](https://github.com/appscode/kubedb-enterprise/commit/34ae3519) Update repository config (#141)
- [f02f0007](https://github.com/appscode/kubedb-enterprise/commit/f02f0007) Update repository config (#139)
- [b1ea4c2e](https://github.com/appscode/kubedb-enterprise/commit/b1ea4c2e) Update Kubernetes v1.18.9 dependencies (#138)
- [341d79ae](https://github.com/appscode/kubedb-enterprise/commit/341d79ae) Update Kubernetes v1.18.9 dependencies (#137)
- [7b26337b](https://github.com/appscode/kubedb-enterprise/commit/7b26337b) Update Kubernetes v1.18.9 dependencies (#136)
- [e4455e82](https://github.com/appscode/kubedb-enterprise/commit/e4455e82) Update repository config (#133)



## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.17.0](https://github.com/kubedb/apimachinery/releases/tag/v0.17.0)

- [a550d467](https://github.com/kubedb/apimachinery/commit/a550d467) Removed constant MariaDBClusterRecommendedVersion (#722)
- [1029fc48](https://github.com/kubedb/apimachinery/commit/1029fc48) Add default monitoring configuration (#721)
- [6d0a8316](https://github.com/kubedb/apimachinery/commit/6d0a8316) Always run PostgreSQL container as 70 (#720)
- [0922b1ab](https://github.com/kubedb/apimachinery/commit/0922b1ab) Update for release Stash@v2021.03.08 (#719)
- [3cdca509](https://github.com/kubedb/apimachinery/commit/3cdca509) Merge ContainerTemplate into PodTemplate spec (#718)
- [8dedb762](https://github.com/kubedb/apimachinery/commit/8dedb762) Set default affinity rules for MariaDB (#717)
- [2bf1490e](https://github.com/kubedb/apimachinery/commit/2bf1490e) Default db container security context (#716)
- [8df88aaa](https://github.com/kubedb/apimachinery/commit/8df88aaa) Update container template (#715)
- [74561e0d](https://github.com/kubedb/apimachinery/commit/74561e0d) Use etcd ports for for pg coordinator (#714)
- [c2cd1993](https://github.com/kubedb/apimachinery/commit/c2cd1993) Update constant fields for pg-coordinator (#713)
- [77c4bd69](https://github.com/kubedb/apimachinery/commit/77c4bd69) Add Elasticsearch helper method for initial master nodes (#712)
- [5cc3309e](https://github.com/kubedb/apimachinery/commit/5cc3309e) Add distribution support to Postgres (#711)
- [5de49ea6](https://github.com/kubedb/apimachinery/commit/5de49ea6) Use import-crds.sh script (#710)
- [5e28f585](https://github.com/kubedb/apimachinery/commit/5e28f585) Use Es distro as ElasticStack
- [169675bb](https://github.com/kubedb/apimachinery/commit/169675bb) Remove spec.tools from EtcdVersion (#709)
- [f4aa5bcc](https://github.com/kubedb/apimachinery/commit/f4aa5bcc) Remove memberWeight from MySQLOpsRequest (#708)
- [201456e8](https://github.com/kubedb/apimachinery/commit/201456e8) Postgres : updated leader elector [ElectionTick,HeartbeatTick] (#700)
- [4cb3f571](https://github.com/kubedb/apimachinery/commit/4cb3f571) Add distribution support in catalog (#707)
- [eb98592d](https://github.com/kubedb/apimachinery/commit/eb98592d) MongoDB: Remove `OrganizationalUnit` and default `Organization` (#704)
- [14ba7e04](https://github.com/kubedb/apimachinery/commit/14ba7e04) Update dependencies
- [3075facf](https://github.com/kubedb/apimachinery/commit/3075facf) Update config types with stash addon config
- [1b8ec75a](https://github.com/kubedb/apimachinery/commit/1b8ec75a) Update catalog stash addon (#703)
- [4b451feb](https://github.com/kubedb/apimachinery/commit/4b451feb) Update repository config (#701)
- [6503a31c](https://github.com/kubedb/apimachinery/commit/6503a31c) Update repository config (#698)
- [612f7384](https://github.com/kubedb/apimachinery/commit/612f7384) Add Stash task refs to Catalog crds (#696)
- [bcb978c0](https://github.com/kubedb/apimachinery/commit/bcb978c0) Update Kubernetes v1.18.9 dependencies (#697)
- [5b44aa8c](https://github.com/kubedb/apimachinery/commit/5b44aa8c) Remove server-id from MySQL CR (#693)
- [4ab0a496](https://github.com/kubedb/apimachinery/commit/4ab0a496) Update Kubernetes v1.18.9 dependencies (#694)
- [8591d95d](https://github.com/kubedb/apimachinery/commit/8591d95d) Update crds via GitHub actions (#695)
- [5fc7c521](https://github.com/kubedb/apimachinery/commit/5fc7c521) Remove deprecated crd yamls
- [7a221977](https://github.com/kubedb/apimachinery/commit/7a221977) Add Mariadb support (#670)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.17.0](https://github.com/kubedb/cli/releases/tag/v0.17.0)

- [818df7f7](https://github.com/kubedb/cli/commit/818df7f7) Prepare for release v0.17.0 (#594)
- [235e88a0](https://github.com/kubedb/cli/commit/235e88a0) Update for release Stash@v2021.03.08 (#593)
- [755754a2](https://github.com/kubedb/cli/commit/755754a2) Update KubeDB api (#592)
- [2c13bea2](https://github.com/kubedb/cli/commit/2c13bea2) Update postgres cli (#591)
- [34b62534](https://github.com/kubedb/cli/commit/34b62534) Update repository config (#588)
- [1cda66c1](https://github.com/kubedb/cli/commit/1cda66c1) Update repository config (#587)
- [65b5d097](https://github.com/kubedb/cli/commit/65b5d097) Update Kubernetes v1.18.9 dependencies (#586)
- [10e2d9b2](https://github.com/kubedb/cli/commit/10e2d9b2) Update Kubernetes v1.18.9 dependencies (#585)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.17.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.17.0)

- [6df700d8](https://github.com/kubedb/elasticsearch/commit/6df700d8) Prepare for release v0.17.0 (#483)
- [58eb52eb](https://github.com/kubedb/elasticsearch/commit/58eb52eb) Update for release Stash@v2021.03.08 (#482)
- [11504552](https://github.com/kubedb/elasticsearch/commit/11504552) Update KubeDB api (#481)
- [d31d0364](https://github.com/kubedb/elasticsearch/commit/d31d0364) Update db container security context (#480)
- [e097ef82](https://github.com/kubedb/elasticsearch/commit/e097ef82) Update KubeDB api (#479)
- [03b16ef0](https://github.com/kubedb/elasticsearch/commit/03b16ef0) Use helper method for initial master nodes (#478)
- [b9785e29](https://github.com/kubedb/elasticsearch/commit/b9785e29) Fix appbinding type meta (#477)
- [fb6a25a8](https://github.com/kubedb/elasticsearch/commit/fb6a25a8) Fix install command in Makefile (#476)
- [8de7f729](https://github.com/kubedb/elasticsearch/commit/8de7f729) Update repository config (#475)
- [99a594c7](https://github.com/kubedb/elasticsearch/commit/99a594c7) Pass stash addon info to AppBinding (#474)
- [fe7603bb](https://github.com/kubedb/elasticsearch/commit/fe7603bb) Mount custom config files to Elasticsearch config directory (#466)
- [8e39688e](https://github.com/kubedb/elasticsearch/commit/8e39688e) Update repository config (#472)
- [1915aa8f](https://github.com/kubedb/elasticsearch/commit/1915aa8f) Update repository config (#471)
- [a0c0a92a](https://github.com/kubedb/elasticsearch/commit/a0c0a92a) Update Kubernetes v1.18.9 dependencies (#470)
- [5579736d](https://github.com/kubedb/elasticsearch/commit/5579736d) Update repository config (#469)
- [ff140030](https://github.com/kubedb/elasticsearch/commit/ff140030) Update Kubernetes v1.18.9 dependencies (#468)
- [95d848b5](https://github.com/kubedb/elasticsearch/commit/95d848b5) Update Kubernetes v1.18.9 dependencies (#467)
- [15ec7161](https://github.com/kubedb/elasticsearch/commit/15ec7161) Update repository config (#465)
- [005a8cc5](https://github.com/kubedb/elasticsearch/commit/005a8cc5) Update repository config (#464)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v0.17.0](https://github.com/kubedb/installer/releases/tag/v0.17.0)

- [c1770ad](https://github.com/kubedb/installer/commit/c1770ad) Prepare for release v0.17.0 (#279)
- [3ed3ee8](https://github.com/kubedb/installer/commit/3ed3ee8) Add global skipCleaner values field (#277)
- [0a14985](https://github.com/kubedb/installer/commit/0a14985) Update combined chart dependency (#276)
- [d4d9f3a](https://github.com/kubedb/installer/commit/d4d9f3a) Add open source images for TimescaleDB (#275)
- [d325aff](https://github.com/kubedb/installer/commit/d325aff) Use distro aware version sorting
- [c99847a](https://github.com/kubedb/installer/commit/c99847a) Add TimescaleDB in Postgres catalog (#274)
- [0cf8ceb](https://github.com/kubedb/installer/commit/0cf8ceb) Fail fmt command if formatting fails
- [605fa5a](https://github.com/kubedb/installer/commit/605fa5a) Update percona MongoDBVersion name (#273)
- [d9adea3](https://github.com/kubedb/installer/commit/d9adea3) Change Elasticsearch catalog naming format (#272)
- [98ce374](https://github.com/kubedb/installer/commit/98ce374) Handle non-semver db version names (#271)
- [291fca7](https://github.com/kubedb/installer/commit/291fca7) Auto download api repo to update crds (#270)
- [ca1b813](https://github.com/kubedb/installer/commit/ca1b813) Update MongoDB init container image (#268)
- [dad5f24](https://github.com/kubedb/installer/commit/dad5f24) Added official image for postgres (#269)
- [0724348](https://github.com/kubedb/installer/commit/0724348) Update for release Stash@v2021.03.08 (#267)
- [6ee56d9](https://github.com/kubedb/installer/commit/6ee56d9) Don't fail deleting namespace when license expires (#266)
- [833135b](https://github.com/kubedb/installer/commit/833135b) Add temporary volume for storing temporary certificates (#265)
- [1a52a95](https://github.com/kubedb/installer/commit/1a52a95) Added new postgres versions in kubedb-catalog (#259)
- [b8a0d0c](https://github.com/kubedb/installer/commit/b8a0d0c) Update MariaDB Image (#264)
- [2270ede](https://github.com/kubedb/installer/commit/2270ede) Fix Stash Addon params for ES SearchGuard & OpenDistro variant (#262)
- [7d361cb](https://github.com/kubedb/installer/commit/7d361cb) Fix build (#261)
- [6863b5a](https://github.com/kubedb/installer/commit/6863b5a) Create combined kubedb chart (#257)
- [a566b56](https://github.com/kubedb/installer/commit/a566b56) Format catalog chart with make fmt
- [fd67c67](https://github.com/kubedb/installer/commit/fd67c67) Add raw catalog yamls (#254)
- [6b283b4](https://github.com/kubedb/installer/commit/6b283b4) Add import-crds.sh script (#255)
- [9897427](https://github.com/kubedb/installer/commit/9897427) Update crds for kubedb/apimachinery@5e28f585 (#253)
- [f3ccbd9](https://github.com/kubedb/installer/commit/f3ccbd9) .Values.catalog.mongo -> .Values.catalog.mongodb (#252)
- [20538d3](https://github.com/kubedb/installer/commit/20538d3) Remove spec.tools from catalog (#250)
- [8bae1ac](https://github.com/kubedb/installer/commit/8bae1ac) Update crds for kubedb/apimachinery@169675bb (#251)
- [f5661b5](https://github.com/kubedb/installer/commit/f5661b5) Update crds for kubedb/apimachinery@f4aa5bcc (#249)
- [30e1a11](https://github.com/kubedb/installer/commit/30e1a11) Disable verify modules
- [6280dff](https://github.com/kubedb/installer/commit/6280dff) Add Stash addon info in MongoDB catalogs (#247)
- [af0b011](https://github.com/kubedb/installer/commit/af0b011) Update crds for kubedb/apimachinery@201456e8 (#248)
- [23f31da](https://github.com/kubedb/installer/commit/23f31da) Update crds for kubedb/apimachinery@1b8ec75a (#245)
- [a950d29](https://github.com/kubedb/installer/commit/a950d29) make ct (#242)
- [95176b0](https://github.com/kubedb/installer/commit/95176b0) Update repository config (#243)
- [9a63b89](https://github.com/kubedb/installer/commit/9a63b89) Remove unused template from chart
- [0da3eb1](https://github.com/kubedb/installer/commit/0da3eb1) Update repository config (#241)
- [cb559d8](https://github.com/kubedb/installer/commit/cb559d8) Update crds for kubedb/apimachinery@612f7384 (#240)
- [bbbd753](https://github.com/kubedb/installer/commit/bbbd753) Update crds for kubedb/apimachinery@5b44aa8c (#239)
- [25988b0](https://github.com/kubedb/installer/commit/25988b0) Add combined kubedb chart (#238)
- [d3bdf52](https://github.com/kubedb/installer/commit/d3bdf52) Rename kubedb chart to kubedb-community (#237)
- [154d542](https://github.com/kubedb/installer/commit/154d542) Add MariaDB Catalogs  (#208)
- [8682f3a](https://github.com/kubedb/installer/commit/8682f3a) Update MySQL catalogs (#235)
- [a32f766](https://github.com/kubedb/installer/commit/a32f766) Update Elasticsearch versions (#234)
- [435fc07](https://github.com/kubedb/installer/commit/435fc07) Update chart description
- [f7bebec](https://github.com/kubedb/installer/commit/f7bebec) Add kubedb-crds chart (#236)
- [26397fc](https://github.com/kubedb/installer/commit/26397fc) Skip generating YAMLs not needed for install command (#233)
- [5788701](https://github.com/kubedb/installer/commit/5788701) Update repository config (#232)
- [89b21eb](https://github.com/kubedb/installer/commit/89b21eb) Add statefulsets/finalizers to ClusterRole (#230)
- [d282443](https://github.com/kubedb/installer/commit/d282443) Cleanup CI workflow (#231)
- [cf01dd6](https://github.com/kubedb/installer/commit/cf01dd6) Update repository config (#229)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.1.0](https://github.com/kubedb/mariadb/releases/tag/v0.1.0)

- [146c9b87](https://github.com/kubedb/mariadb/commit/146c9b87) Prepare for release v0.1.0 (#56)
- [808ff4cd](https://github.com/kubedb/mariadb/commit/808ff4cd) Pass stash addon info to AppBinding (#55)
- [8e77c251](https://github.com/kubedb/mariadb/commit/8e77c251) Removed recommended version check from validator (#54)
- [4139a7b3](https://github.com/kubedb/mariadb/commit/4139a7b3) Update for release Stash@v2021.03.08 (#53)
- [fabdbce0](https://github.com/kubedb/mariadb/commit/fabdbce0) Update Makefile
- [9c4ee6e8](https://github.com/kubedb/mariadb/commit/9c4ee6e8) Implement MariaDB operator (#42)
- [2c3ad2c0](https://github.com/kubedb/mariadb/commit/2c3ad2c0) Fix install command in Makefile (#50)
- [2d57e20d](https://github.com/kubedb/mariadb/commit/2d57e20d) Update repository config (#48)
- [f2f8c646](https://github.com/kubedb/mariadb/commit/f2f8c646) Update repository config (#47)
- [3f120133](https://github.com/kubedb/mariadb/commit/3f120133) Update Kubernetes v1.18.9 dependencies (#46)
- [ede125ba](https://github.com/kubedb/mariadb/commit/ede125ba) Update repository config (#45)
- [d173d7a9](https://github.com/kubedb/mariadb/commit/d173d7a9) Update Kubernetes v1.18.9 dependencies (#44)
- [b22bc09e](https://github.com/kubedb/mariadb/commit/b22bc09e) Update Kubernetes v1.18.9 dependencies (#43)
- [3d63b2fa](https://github.com/kubedb/mariadb/commit/3d63b2fa) Update repository config (#40)
- [8440ebde](https://github.com/kubedb/mariadb/commit/8440ebde) Update repository config (#39)
- [24dfaf7d](https://github.com/kubedb/mariadb/commit/24dfaf7d) Update Kubernetes v1.18.9 dependencies (#37)
- [923f1e88](https://github.com/kubedb/mariadb/commit/923f1e88) Update for release Stash@v2021.01.21 (#36)
- [e5a1f271](https://github.com/kubedb/mariadb/commit/e5a1f271) Update repository config (#35)
- [b60c7fb6](https://github.com/kubedb/mariadb/commit/b60c7fb6) Update repository config (#34)
- [a9361a2f](https://github.com/kubedb/mariadb/commit/a9361a2f) Update KubeDB api (#33)
- [1b66913c](https://github.com/kubedb/mariadb/commit/1b66913c) Update KubeDB api (#32)
- [9b530888](https://github.com/kubedb/mariadb/commit/9b530888) Update KubeDB api (#30)
- [a5a2b4b8](https://github.com/kubedb/mariadb/commit/a5a2b4b8) Update KubeDB api (#29)
- [7642dfdf](https://github.com/kubedb/mariadb/commit/7642dfdf) Update KubeDB api (#28)
- [561e1da9](https://github.com/kubedb/mariadb/commit/561e1da9) Delete e2e tests moved to kubedb/test repo (#27)
- [1f772e07](https://github.com/kubedb/mariadb/commit/1f772e07) Update KubeDB api (#25)
- [7f18e249](https://github.com/kubedb/mariadb/commit/7f18e249) Fix annotations passing to AppBinding (#24)
- [40e5e1d6](https://github.com/kubedb/mariadb/commit/40e5e1d6) Initialize mapper
- [6b6be5d7](https://github.com/kubedb/mariadb/commit/6b6be5d7) Change offshoot selector labels to standard k8s app labels (#23)
- [8e88e863](https://github.com/kubedb/mariadb/commit/8e88e863) Update KubeDB api (#22)
- [ab55d3f3](https://github.com/kubedb/mariadb/commit/ab55d3f3) Update KubeDB api (#21)
- [d256ae83](https://github.com/kubedb/mariadb/commit/d256ae83) Use basic-auth secret type for auth secret (#20)
- [8988ecbe](https://github.com/kubedb/mariadb/commit/8988ecbe) Update KubeDB api (#19)
- [cb9264eb](https://github.com/kubedb/mariadb/commit/cb9264eb) Update for release Stash@v2020.12.17 (#18)
- [92a4a353](https://github.com/kubedb/mariadb/commit/92a4a353) Update KubeDB api (#17)
- [f95e6bd8](https://github.com/kubedb/mariadb/commit/f95e6bd8) Update KubeDB api (#16)
- [4ce3e1fe](https://github.com/kubedb/mariadb/commit/4ce3e1fe) Update KubeDB api (#15)
- [9c2d3e8f](https://github.com/kubedb/mariadb/commit/9c2d3e8f) Update Kubernetes v1.18.9 dependencies (#14)
- [57f57bcc](https://github.com/kubedb/mariadb/commit/57f57bcc) Update KubeDB api (#13)
- [a0e4100d](https://github.com/kubedb/mariadb/commit/a0e4100d) Update KubeDB api (#12)
- [6e3159b0](https://github.com/kubedb/mariadb/commit/6e3159b0) Update KubeDB api (#11)
- [04adaa56](https://github.com/kubedb/mariadb/commit/04adaa56) Update Kubernetes v1.18.9 dependencies (#10)
- [34c40bf6](https://github.com/kubedb/mariadb/commit/34c40bf6) Update e2e workflow (#9)
- [c95cb8e7](https://github.com/kubedb/mariadb/commit/c95cb8e7) Update KubeDB api (#8)
- [cefbd5e6](https://github.com/kubedb/mariadb/commit/cefbd5e6) Format shel scripts (#7)
- [3fbc312e](https://github.com/kubedb/mariadb/commit/3fbc312e) Update KubeDB api (#6)
- [6dfefd95](https://github.com/kubedb/mariadb/commit/6dfefd95) Update KubeDB api (#5)
- [b0e0bc48](https://github.com/kubedb/mariadb/commit/b0e0bc48) Update repository config (#4)
- [ddac5279](https://github.com/kubedb/mariadb/commit/ddac5279) Update readme
- [3aebb7b1](https://github.com/kubedb/mariadb/commit/3aebb7b1) Fix serviceTemplate inline json (#2)
- [bda7cb60](https://github.com/kubedb/mariadb/commit/bda7cb60) Rename to MariaDB (#3)
- [aa216cf5](https://github.com/kubedb/mariadb/commit/aa216cf5) Prepare for release v0.1.1 (#134)
- [d43b87a3](https://github.com/kubedb/mariadb/commit/d43b87a3) Update Kubernetes v1.18.9 dependencies (#133)
- [1a354dba](https://github.com/kubedb/mariadb/commit/1a354dba) Update KubeDB api (#132)
- [808366cc](https://github.com/kubedb/mariadb/commit/808366cc) Update Kubernetes v1.18.9 dependencies (#131)
- [adb44379](https://github.com/kubedb/mariadb/commit/adb44379) Update KubeDB api (#130)
- [6d6188de](https://github.com/kubedb/mariadb/commit/6d6188de) Update for release Stash@v2020.11.06 (#129)
- [8d3eaa37](https://github.com/kubedb/mariadb/commit/8d3eaa37) Update Kubernetes v1.18.9 dependencies (#128)
- [5f7253b6](https://github.com/kubedb/mariadb/commit/5f7253b6) Update KubeDB api (#126)
- [43f10d83](https://github.com/kubedb/mariadb/commit/43f10d83) Update KubeDB api (#125)
- [91940395](https://github.com/kubedb/mariadb/commit/91940395) Update for release Stash@v2020.10.30 (#124)
- [eba69286](https://github.com/kubedb/mariadb/commit/eba69286) Update KubeDB api (#123)
- [a4dd87ba](https://github.com/kubedb/mariadb/commit/a4dd87ba) Update for release Stash@v2020.10.29 (#122)
- [3b2593ce](https://github.com/kubedb/mariadb/commit/3b2593ce) Prepare for release v0.1.0 (#121)
- [ae82716f](https://github.com/kubedb/mariadb/commit/ae82716f) Prepare for release v0.1.0-rc.2 (#120)
- [4ac07f08](https://github.com/kubedb/mariadb/commit/4ac07f08) Prepare for release v0.1.0-rc.1 (#119)
- [397607a3](https://github.com/kubedb/mariadb/commit/397607a3) Prepare for release v0.1.0-beta.6 (#118)
- [a3b7642d](https://github.com/kubedb/mariadb/commit/a3b7642d) Create SRV records for governing service (#117)
- [9866a420](https://github.com/kubedb/mariadb/commit/9866a420) Prepare for release v0.1.0-beta.5 (#116)
- [f92081d1](https://github.com/kubedb/mariadb/commit/f92081d1) Create separate governing service for each database (#115)
- [6010b189](https://github.com/kubedb/mariadb/commit/6010b189) Update KubeDB api (#114)
- [95b57c72](https://github.com/kubedb/mariadb/commit/95b57c72) Update readme
- [14b2f1b2](https://github.com/kubedb/mariadb/commit/14b2f1b2) Prepare for release v0.1.0-beta.4 (#113)
- [eff1d265](https://github.com/kubedb/mariadb/commit/eff1d265) Update KubeDB api (#112)
- [a2878d4a](https://github.com/kubedb/mariadb/commit/a2878d4a) Update Kubernetes v1.18.9 dependencies (#111)
- [51f0d104](https://github.com/kubedb/mariadb/commit/51f0d104) Update KubeDB api (#110)
- [fcf5343b](https://github.com/kubedb/mariadb/commit/fcf5343b) Update for release Stash@v2020.10.21 (#109)
- [9fe68d43](https://github.com/kubedb/mariadb/commit/9fe68d43) Fix init validator (#107)
- [1c528cff](https://github.com/kubedb/mariadb/commit/1c528cff) Update KubeDB api (#108)
- [99d23f3d](https://github.com/kubedb/mariadb/commit/99d23f3d) Update KubeDB api (#106)
- [d0807640](https://github.com/kubedb/mariadb/commit/d0807640) Update Kubernetes v1.18.9 dependencies (#105)
- [bac7705b](https://github.com/kubedb/mariadb/commit/bac7705b) Update KubeDB api (#104)
- [475aabd5](https://github.com/kubedb/mariadb/commit/475aabd5) Update KubeDB api (#103)
- [60f7e5a9](https://github.com/kubedb/mariadb/commit/60f7e5a9) Update KubeDB api (#102)
- [84a97ced](https://github.com/kubedb/mariadb/commit/84a97ced) Update KubeDB api (#101)
- [d4a7b7c5](https://github.com/kubedb/mariadb/commit/d4a7b7c5) Update Kubernetes v1.18.9 dependencies (#100)
- [b818a4c5](https://github.com/kubedb/mariadb/commit/b818a4c5) Update KubeDB api (#99)
- [03df7739](https://github.com/kubedb/mariadb/commit/03df7739) Update KubeDB api (#98)
- [2f3ce0e6](https://github.com/kubedb/mariadb/commit/2f3ce0e6) Update KubeDB api (#96)
- [94e009e8](https://github.com/kubedb/mariadb/commit/94e009e8) Update repository config (#95)
- [fc61d440](https://github.com/kubedb/mariadb/commit/fc61d440) Update repository config (#94)
- [35f5b2bb](https://github.com/kubedb/mariadb/commit/35f5b2bb) Update repository config (#93)
- [d01e39dd](https://github.com/kubedb/mariadb/commit/d01e39dd) Initialize statefulset watcher from cmd/server/options.go (#92)
- [41bf932f](https://github.com/kubedb/mariadb/commit/41bf932f) Update KubeDB api (#91)
- [da92a1f3](https://github.com/kubedb/mariadb/commit/da92a1f3) Update Kubernetes v1.18.9 dependencies (#90)
- [554beafb](https://github.com/kubedb/mariadb/commit/554beafb) Publish docker images to ghcr.io (#89)
- [4c7031e1](https://github.com/kubedb/mariadb/commit/4c7031e1) Update KubeDB api (#88)
- [418c767a](https://github.com/kubedb/mariadb/commit/418c767a) Update KubeDB api (#87)
- [94eef91e](https://github.com/kubedb/mariadb/commit/94eef91e) Update KubeDB api (#86)
- [f3c2a360](https://github.com/kubedb/mariadb/commit/f3c2a360) Update KubeDB api (#85)
- [107bb6a6](https://github.com/kubedb/mariadb/commit/107bb6a6) Update repository config (#84)
- [938e64bc](https://github.com/kubedb/mariadb/commit/938e64bc) Cleanup monitoring spec api (#83)
- [deeaad8f](https://github.com/kubedb/mariadb/commit/deeaad8f) Use conditions to handle database initialization (#80)
- [798c3ddc](https://github.com/kubedb/mariadb/commit/798c3ddc) Update Kubernetes v1.18.9 dependencies (#82)
- [16c72ba6](https://github.com/kubedb/mariadb/commit/16c72ba6) Updated the exporter port and service (#81)
- [9314faf1](https://github.com/kubedb/mariadb/commit/9314faf1) Update for release Stash@v2020.09.29 (#79)
- [6cb53efc](https://github.com/kubedb/mariadb/commit/6cb53efc) Update Kubernetes v1.18.9 dependencies (#78)
- [fd2b8cdd](https://github.com/kubedb/mariadb/commit/fd2b8cdd) Update Kubernetes v1.18.9 dependencies (#76)
- [9d1038db](https://github.com/kubedb/mariadb/commit/9d1038db) Update repository config (#75)
- [41a05a44](https://github.com/kubedb/mariadb/commit/41a05a44) Update repository config (#74)
- [eccd2acd](https://github.com/kubedb/mariadb/commit/eccd2acd) Update Kubernetes v1.18.9 dependencies (#73)
- [27635f1c](https://github.com/kubedb/mariadb/commit/27635f1c) Update Kubernetes v1.18.3 dependencies (#72)
- [792326c7](https://github.com/kubedb/mariadb/commit/792326c7) Use common event recorder (#71)
- [0ff583b8](https://github.com/kubedb/mariadb/commit/0ff583b8) Prepare for release v0.1.0-beta.3 (#70)
- [627bc039](https://github.com/kubedb/mariadb/commit/627bc039) Use new `spec.init` section (#69)
- [f79e4771](https://github.com/kubedb/mariadb/commit/f79e4771) Update Kubernetes v1.18.3 dependencies (#68)
- [257954c2](https://github.com/kubedb/mariadb/commit/257954c2) Add license verifier (#67)
- [e06eec6b](https://github.com/kubedb/mariadb/commit/e06eec6b) Update for release Stash@v2020.09.16 (#66)
- [29901348](https://github.com/kubedb/mariadb/commit/29901348) Update Kubernetes v1.18.3 dependencies (#65)
- [02d5bfde](https://github.com/kubedb/mariadb/commit/02d5bfde) Use background deletion policy
- [6e6d8b5b](https://github.com/kubedb/mariadb/commit/6e6d8b5b) Update Kubernetes v1.18.3 dependencies (#63)
- [7601a237](https://github.com/kubedb/mariadb/commit/7601a237) Use AppsCode Community License (#62)
- [4d1a2424](https://github.com/kubedb/mariadb/commit/4d1a2424) Update Kubernetes v1.18.3 dependencies (#61)
- [471b6def](https://github.com/kubedb/mariadb/commit/471b6def) Prepare for release v0.1.0-beta.2 (#60)
- [9423a70f](https://github.com/kubedb/mariadb/commit/9423a70f) Update release.yml
- [85d1d036](https://github.com/kubedb/mariadb/commit/85d1d036) Use updated apis (#59)
- [6811b8dc](https://github.com/kubedb/mariadb/commit/6811b8dc) Update Kubernetes v1.18.3 dependencies (#53)
- [4212d2a0](https://github.com/kubedb/mariadb/commit/4212d2a0) Update Kubernetes v1.18.3 dependencies (#52)
- [659d646c](https://github.com/kubedb/mariadb/commit/659d646c) Update Kubernetes v1.18.3 dependencies (#51)
- [a868e0c3](https://github.com/kubedb/mariadb/commit/a868e0c3) Update Kubernetes v1.18.3 dependencies (#50)
- [162e6ca4](https://github.com/kubedb/mariadb/commit/162e6ca4) Update Kubernetes v1.18.3 dependencies (#49)
- [a7fa1fbf](https://github.com/kubedb/mariadb/commit/a7fa1fbf) Update Kubernetes v1.18.3 dependencies (#48)
- [b6a4583f](https://github.com/kubedb/mariadb/commit/b6a4583f) Remove dependency on enterprise operator (#47)
- [a8909b38](https://github.com/kubedb/mariadb/commit/a8909b38) Allow configuring k8s & db version in e2e tests (#46)
- [4d79d26e](https://github.com/kubedb/mariadb/commit/4d79d26e) Update to Kubernetes v1.18.3 (#45)
- [189f3212](https://github.com/kubedb/mariadb/commit/189f3212) Trigger e2e tests on /ok-to-test command (#44)
- [a037bd03](https://github.com/kubedb/mariadb/commit/a037bd03) Update to Kubernetes v1.18.3 (#43)
- [33cabdf3](https://github.com/kubedb/mariadb/commit/33cabdf3) Update to Kubernetes v1.18.3 (#42)
- [28b9fc0f](https://github.com/kubedb/mariadb/commit/28b9fc0f) Prepare for release v0.1.0-beta.1 (#41)
- [fb4f5444](https://github.com/kubedb/mariadb/commit/fb4f5444) Update for release Stash@v2020.07.09-beta.0 (#39)
- [ad221aa2](https://github.com/kubedb/mariadb/commit/ad221aa2) include Makefile.env
- [841ec855](https://github.com/kubedb/mariadb/commit/841ec855) Allow customizing chart registry (#38)
- [bb608980](https://github.com/kubedb/mariadb/commit/bb608980) Update License (#37)
- [cf8cd2fa](https://github.com/kubedb/mariadb/commit/cf8cd2fa) Update for release Stash@v2020.07.08-beta.0 (#36)
- [7b28c4b9](https://github.com/kubedb/mariadb/commit/7b28c4b9) Update to Kubernetes v1.18.3 (#35)
- [848ff94a](https://github.com/kubedb/mariadb/commit/848ff94a) Update ci.yml
- [d124dd6a](https://github.com/kubedb/mariadb/commit/d124dd6a) Load stash version from .env file for make (#34)
- [1de40e1d](https://github.com/kubedb/mariadb/commit/1de40e1d) Update update-release-tracker.sh
- [7a4503be](https://github.com/kubedb/mariadb/commit/7a4503be) Update update-release-tracker.sh
- [ad0dfaf8](https://github.com/kubedb/mariadb/commit/ad0dfaf8) Add script to update release tracker on pr merge (#33)
- [aaca6bd9](https://github.com/kubedb/mariadb/commit/aaca6bd9) Update .kodiak.toml
- [9a495724](https://github.com/kubedb/mariadb/commit/9a495724) Various fixes (#32)
- [9b6c9a53](https://github.com/kubedb/mariadb/commit/9b6c9a53) Update to Kubernetes v1.18.3 (#31)
- [67912547](https://github.com/kubedb/mariadb/commit/67912547) Update to Kubernetes v1.18.3
- [fc8ce4cc](https://github.com/kubedb/mariadb/commit/fc8ce4cc) Create .kodiak.toml
- [8aba5ef2](https://github.com/kubedb/mariadb/commit/8aba5ef2) Use CRD v1 for Kubernetes >= 1.16 (#30)
- [e81d2b4c](https://github.com/kubedb/mariadb/commit/e81d2b4c) Update to Kubernetes v1.18.3 (#29)
- [2a32730a](https://github.com/kubedb/mariadb/commit/2a32730a) Fix e2e tests (#28)
- [a79626d9](https://github.com/kubedb/mariadb/commit/a79626d9) Update stash install commands
- [52fc2059](https://github.com/kubedb/mariadb/commit/52fc2059) Use recommended kubernetes app labels (#27)
- [93dc10ec](https://github.com/kubedb/mariadb/commit/93dc10ec) Update crazy-max/ghaction-docker-buildx flag
- [ce5717e2](https://github.com/kubedb/mariadb/commit/ce5717e2) Revendor kubedb.dev/apimachinery@master (#26)
- [c1ca649d](https://github.com/kubedb/mariadb/commit/c1ca649d) Pass annotations from CRD to AppBinding (#25)
- [f327cc01](https://github.com/kubedb/mariadb/commit/f327cc01) Trigger the workflow on push or pull request
- [02432393](https://github.com/kubedb/mariadb/commit/02432393) Update CHANGELOG.md
- [a89dbc55](https://github.com/kubedb/mariadb/commit/a89dbc55) Use stash.appscode.dev/apimachinery@v0.9.0-rc.6 (#24)
- [e69742de](https://github.com/kubedb/mariadb/commit/e69742de) Update for percona-xtradb standalone restoresession (#23)
- [958877a1](https://github.com/kubedb/mariadb/commit/958877a1) Various fixes (#21)
- [fb0d7a35](https://github.com/kubedb/mariadb/commit/fb0d7a35) Update kubernetes client-go to 1.16.3 (#20)
- [293fe9a4](https://github.com/kubedb/mariadb/commit/293fe9a4) Fix default make command
- [39358e3b](https://github.com/kubedb/mariadb/commit/39358e3b) Use charts to install operator (#19)
- [6c5b3395](https://github.com/kubedb/mariadb/commit/6c5b3395) Several fixes and update tests (#18)
- [84ff139f](https://github.com/kubedb/mariadb/commit/84ff139f) Various Makefile improvements (#16)
- [e2737f65](https://github.com/kubedb/mariadb/commit/e2737f65) Remove EnableStatusSubresource (#17)
- [fb886b07](https://github.com/kubedb/mariadb/commit/fb886b07) Run e2e tests using GitHub actions (#12)
- [35b155d9](https://github.com/kubedb/mariadb/commit/35b155d9) Validate DBVersionSpecs and fixed broken build (#15)
- [67794bd9](https://github.com/kubedb/mariadb/commit/67794bd9) Update go.yml
- [f7666354](https://github.com/kubedb/mariadb/commit/f7666354) Various changes for Percona XtraDB (#13)
- [ceb7ba67](https://github.com/kubedb/mariadb/commit/ceb7ba67) Enable GitHub actions
- [f5a112af](https://github.com/kubedb/mariadb/commit/f5a112af) Refactor for ProxySQL Integration (#11)
- [26602049](https://github.com/kubedb/mariadb/commit/26602049) Revendor
- [71957d40](https://github.com/kubedb/mariadb/commit/71957d40) Rename from perconaxtradb to percona-xtradb (#10)
- [b526ccd8](https://github.com/kubedb/mariadb/commit/b526ccd8) Set database version in AppBinding (#7)
- [336e7203](https://github.com/kubedb/mariadb/commit/336e7203) Percona XtraDB Cluster support (#9)
- [71a42f7a](https://github.com/kubedb/mariadb/commit/71a42f7a) Don't set annotation to AppBinding (#8)
- [282298cb](https://github.com/kubedb/mariadb/commit/282298cb) Fix UpsertDatabaseAnnotation() function (#4)
- [2ab9dddf](https://github.com/kubedb/mariadb/commit/2ab9dddf) Add license header to Makefiles (#6)
- [df135c08](https://github.com/kubedb/mariadb/commit/df135c08) Add install, uninstall and purge command in Makefile (#3)
- [73d3a845](https://github.com/kubedb/mariadb/commit/73d3a845) Update .gitignore
- [59a4e754](https://github.com/kubedb/mariadb/commit/59a4e754) Add Makefile (#2)
- [f3551ddc](https://github.com/kubedb/mariadb/commit/f3551ddc) Rename package path (#1)
- [56a241d6](https://github.com/kubedb/mariadb/commit/56a241d6) Use explicit IP whitelist instead of automatic IP whitelist (#151)
- [9f0b5ca3](https://github.com/kubedb/mariadb/commit/9f0b5ca3) Update to k8s 1.14.0 client libraries using go.mod (#147)
- [73ad7c30](https://github.com/kubedb/mariadb/commit/73ad7c30) Update changelog
- [ccc36b5c](https://github.com/kubedb/mariadb/commit/ccc36b5c) Update README.md
- [9769e8e1](https://github.com/kubedb/mariadb/commit/9769e8e1) Start next dev cycle
- [a3fa468a](https://github.com/kubedb/mariadb/commit/a3fa468a) Prepare release 0.5.0
- [6d8862de](https://github.com/kubedb/mariadb/commit/6d8862de) Mysql Group Replication tests (#146)
- [49544e55](https://github.com/kubedb/mariadb/commit/49544e55) Mysql Group Replication (#144)
- [a85d4b44](https://github.com/kubedb/mariadb/commit/a85d4b44) Revendor dependencies
- [9c538460](https://github.com/kubedb/mariadb/commit/9c538460) Changed Role to exclude psp without name (#143)
- [6cace93b](https://github.com/kubedb/mariadb/commit/6cace93b) Modify mutator validator names (#142)
- [da0c19b9](https://github.com/kubedb/mariadb/commit/da0c19b9) Update changelog
- [b79c80d6](https://github.com/kubedb/mariadb/commit/b79c80d6) Start next dev cycle
- [838d9459](https://github.com/kubedb/mariadb/commit/838d9459) Prepare release 0.4.0
- [bf0f2c14](https://github.com/kubedb/mariadb/commit/bf0f2c14) Added PSP names and init container image in testing framework (#141)
- [3d227570](https://github.com/kubedb/mariadb/commit/3d227570) Added PSP support for mySQL (#137)
- [7b766657](https://github.com/kubedb/mariadb/commit/7b766657) Don't inherit app.kubernetes.io labels from CRD into offshoots (#140)
- [29e23470](https://github.com/kubedb/mariadb/commit/29e23470) Support for init container (#139)
- [3e1556f6](https://github.com/kubedb/mariadb/commit/3e1556f6) Add role label to stats service (#138)
- [ee078af9](https://github.com/kubedb/mariadb/commit/ee078af9) Update changelog
- [978f1139](https://github.com/kubedb/mariadb/commit/978f1139) Update Kubernetes client libraries to 1.13.0 release (#136)
- [821f23d1](https://github.com/kubedb/mariadb/commit/821f23d1) Start next dev cycle
- [678b26aa](https://github.com/kubedb/mariadb/commit/678b26aa) Prepare release 0.3.0
- [40ad7a23](https://github.com/kubedb/mariadb/commit/40ad7a23) Initial RBAC support: create and use K8s service account for MySQL (#134)
- [98f03387](https://github.com/kubedb/mariadb/commit/98f03387) Revendor dependencies (#135)
- [dfe92615](https://github.com/kubedb/mariadb/commit/dfe92615) Revendor dependencies : Retry Failed Scheduler Snapshot (#133)
- [71f8a350](https://github.com/kubedb/mariadb/commit/71f8a350) Added ephemeral StorageType support (#132)
- [0a6b6e46](https://github.com/kubedb/mariadb/commit/0a6b6e46) Added support of MySQL 8.0.14 (#131)
- [99e57a9e](https://github.com/kubedb/mariadb/commit/99e57a9e) Use PVC spec from snapshot if provided (#130)
- [61497be6](https://github.com/kubedb/mariadb/commit/61497be6) Revendored and updated tests for 'Prevent prefix matching of multiple snapshots' (#129)
- [7eafe088](https://github.com/kubedb/mariadb/commit/7eafe088) Add certificate health checker (#128)
- [973ec416](https://github.com/kubedb/mariadb/commit/973ec416) Update E2E test: Env update is not restricted anymore (#127)
- [339975ff](https://github.com/kubedb/mariadb/commit/339975ff) Fix AppBinding (#126)
- [62050a72](https://github.com/kubedb/mariadb/commit/62050a72) Update changelog
- [2d454043](https://github.com/kubedb/mariadb/commit/2d454043) Prepare release 0.2.0
- [6941ea59](https://github.com/kubedb/mariadb/commit/6941ea59) Reuse event recorder (#125)
- [b77e66c4](https://github.com/kubedb/mariadb/commit/b77e66c4) OSM binary upgraded in mysql-tools (#123)
- [c9228086](https://github.com/kubedb/mariadb/commit/c9228086) Revendor dependencies (#124)
- [97837120](https://github.com/kubedb/mariadb/commit/97837120) Test for faulty snapshot (#122)
- [c3e995b6](https://github.com/kubedb/mariadb/commit/c3e995b6) Start next dev cycle
- [8a4f3b13](https://github.com/kubedb/mariadb/commit/8a4f3b13) Prepare release 0.2.0-rc.2
- [79942191](https://github.com/kubedb/mariadb/commit/79942191) Upgrade database secret keys (#121)
- [1747fdf5](https://github.com/kubedb/mariadb/commit/1747fdf5) Ignore mutation of fields to default values during update (#120)
- [d902d588](https://github.com/kubedb/mariadb/commit/d902d588) Support configuration options for exporter sidecar (#119)
- [dd7c3f44](https://github.com/kubedb/mariadb/commit/dd7c3f44) Use flags.DumpAll (#118)
- [bc1ef05b](https://github.com/kubedb/mariadb/commit/bc1ef05b) Start next dev cycle
- [9d33c1a0](https://github.com/kubedb/mariadb/commit/9d33c1a0) Prepare release 0.2.0-rc.1
- [b076e141](https://github.com/kubedb/mariadb/commit/b076e141) Apply cleanup (#117)
- [7dc5641f](https://github.com/kubedb/mariadb/commit/7dc5641f) Set periodic analytics (#116)
- [90ea6acc](https://github.com/kubedb/mariadb/commit/90ea6acc) Introduce AppBinding support (#115)
- [a882d76a](https://github.com/kubedb/mariadb/commit/a882d76a) Fix Analytics (#114)
- [0961009c](https://github.com/kubedb/mariadb/commit/0961009c) Error out from cron job for deprecated dbversion (#113)
- [da1f4e27](https://github.com/kubedb/mariadb/commit/da1f4e27) Add CRDs without observation when operator starts (#112)
- [0a754d2f](https://github.com/kubedb/mariadb/commit/0a754d2f) Update changelog
- [b09bc6e1](https://github.com/kubedb/mariadb/commit/b09bc6e1) Start next dev cycle
- [0d467ccb](https://github.com/kubedb/mariadb/commit/0d467ccb) Prepare release 0.2.0-rc.0
- [c757007a](https://github.com/kubedb/mariadb/commit/c757007a) Merge commit 'cc6607a3589a79a5e61bb198d370ea0ae30b9d09'
- [ddfe4be1](https://github.com/kubedb/mariadb/commit/ddfe4be1) Support custom user passowrd for backup (#111)
- [8c84ba20](https://github.com/kubedb/mariadb/commit/8c84ba20) Support providing resources for monitoring container (#110)
- [7bcfbc48](https://github.com/kubedb/mariadb/commit/7bcfbc48) Update kubernetes client libraries to 1.12.0 (#109)
- [145bba2b](https://github.com/kubedb/mariadb/commit/145bba2b) Add validation webhook xray (#108)
- [6da1887f](https://github.com/kubedb/mariadb/commit/6da1887f) Various Fixes (#107)
- [111519e9](https://github.com/kubedb/mariadb/commit/111519e9) Merge ports from service template (#105)
- [38147ef1](https://github.com/kubedb/mariadb/commit/38147ef1) Replace doNotPause with TerminationPolicy = DoNotTerminate (#104)
- [e28ebc47](https://github.com/kubedb/mariadb/commit/e28ebc47) Pass resources to NamespaceValidator (#103)
- [aed12bf5](https://github.com/kubedb/mariadb/commit/aed12bf5) Various fixes (#102)
- [3d372ef6](https://github.com/kubedb/mariadb/commit/3d372ef6) Support Livecycle hook and container probes (#101)
- [b6ef6887](https://github.com/kubedb/mariadb/commit/b6ef6887) Check if Kubernetes version is supported before running operator (#100)
- [d89e7783](https://github.com/kubedb/mariadb/commit/d89e7783) Update package alias (#99)
- [f0b44b3a](https://github.com/kubedb/mariadb/commit/f0b44b3a) Start next dev cycle
- [a79ff03b](https://github.com/kubedb/mariadb/commit/a79ff03b) Prepare release 0.2.0-beta.1
- [0d8d3cca](https://github.com/kubedb/mariadb/commit/0d8d3cca) Revendor api (#98)
- [2f850243](https://github.com/kubedb/mariadb/commit/2f850243) Fix tests (#97)
- [4ced0bfe](https://github.com/kubedb/mariadb/commit/4ced0bfe) Revendor api for catalog apigroup (#96)
- [e7695400](https://github.com/kubedb/mariadb/commit/e7695400) Update chanelog
- [8e358aea](https://github.com/kubedb/mariadb/commit/8e358aea) Use --pull flag with docker build (#20) (#95)
- [d2a97d90](https://github.com/kubedb/mariadb/commit/d2a97d90) Merge commit '16c769ee4686576f172a6b79a10d25bfd79ca4a4'
- [d1fe8a8a](https://github.com/kubedb/mariadb/commit/d1fe8a8a) Start next dev cycle
- [04eb9bb5](https://github.com/kubedb/mariadb/commit/04eb9bb5) Prepare release 0.2.0-beta.0
- [9dfea960](https://github.com/kubedb/mariadb/commit/9dfea960) Pass extra args to tools.sh (#93)
- [47dd3cad](https://github.com/kubedb/mariadb/commit/47dd3cad) Don't try to wipe out Snapshot data for Local backend (#92)
- [9c4d485b](https://github.com/kubedb/mariadb/commit/9c4d485b) Add missing alt-tag docker folder mysql-tools images (#91)
- [be72f784](https://github.com/kubedb/mariadb/commit/be72f784) Use suffix for updated DBImage & Stop working for deprecated *Versions (#90)
- [05c8f14d](https://github.com/kubedb/mariadb/commit/05c8f14d) Search used secrets within same namespace of DB object (#89)
- [0d94c946](https://github.com/kubedb/mariadb/commit/0d94c946) Support Termination Policy (#88)
- [8775ddf7](https://github.com/kubedb/mariadb/commit/8775ddf7) Update builddeps.sh
- [796c93da](https://github.com/kubedb/mariadb/commit/796c93da) Revendor k8s.io/apiserver (#87)
- [5a1e3f57](https://github.com/kubedb/mariadb/commit/5a1e3f57) Revendor kubernetes-1.11.3 (#86)
- [809a3c49](https://github.com/kubedb/mariadb/commit/809a3c49) Support UpdateStrategy (#84)
- [372c52ef](https://github.com/kubedb/mariadb/commit/372c52ef) Add TerminationPolicy for databases (#83)
- [c01b55e8](https://github.com/kubedb/mariadb/commit/c01b55e8) Revendor api (#82)
- [5e196b95](https://github.com/kubedb/mariadb/commit/5e196b95) Use IntHash as status.observedGeneration (#81)
- [2da3bb1b](https://github.com/kubedb/mariadb/commit/2da3bb1b) fix github status (#80)
- [121d0a98](https://github.com/kubedb/mariadb/commit/121d0a98) Update pipeline (#79)
- [532e3137](https://github.com/kubedb/mariadb/commit/532e3137) Fix E2E test for minikube (#78)
- [0f107815](https://github.com/kubedb/mariadb/commit/0f107815) Update pipeline (#77)
- [851679e2](https://github.com/kubedb/mariadb/commit/851679e2) Migrate MySQL (#75)
- [0b997855](https://github.com/kubedb/mariadb/commit/0b997855) Use official exporter image (#74)
- [702d5736](https://github.com/kubedb/mariadb/commit/702d5736) Fix uninstall for concourse (#70)
- [9ee88bd2](https://github.com/kubedb/mariadb/commit/9ee88bd2) Update status.ObservedGeneration for failure phase (#73)
- [559cdb6a](https://github.com/kubedb/mariadb/commit/559cdb6a) Keep track of ObservedGenerationHash (#72)
- [61c8b898](https://github.com/kubedb/mariadb/commit/61c8b898) Use NewObservableHandler (#71)
- [421274dc](https://github.com/kubedb/mariadb/commit/421274dc) Merge commit '887037c7e36289e3135dda99346fccc7e2ce303b'
- [6a41d9bc](https://github.com/kubedb/mariadb/commit/6a41d9bc) Fix uninstall for concourse (#69)
- [f1af09db](https://github.com/kubedb/mariadb/commit/f1af09db) Update README.md
- [bf3f1823](https://github.com/kubedb/mariadb/commit/bf3f1823) Revise immutable spec fields (#68)
- [26adec3b](https://github.com/kubedb/mariadb/commit/26adec3b) Merge commit '5f83049fc01dc1d0709ac0014d6f3a0f74a39417'
- [31a97820](https://github.com/kubedb/mariadb/commit/31a97820) Support passing args via PodTemplate (#67)
- [60f4ee23](https://github.com/kubedb/mariadb/commit/60f4ee23) Introduce storageType : ephemeral (#66)
- [bfd3fcd6](https://github.com/kubedb/mariadb/commit/bfd3fcd6) Add support for running tests on cncf cluster (#63)
- [fba47b19](https://github.com/kubedb/mariadb/commit/fba47b19) Merge commit 'e010cbb302c8d59d4cf69dd77085b046ff423b78'
- [6be96ce0](https://github.com/kubedb/mariadb/commit/6be96ce0) Revendor api (#65)
- [0f629ab3](https://github.com/kubedb/mariadb/commit/0f629ab3) Keep track of observedGeneration in status (#64)
- [c9a9596f](https://github.com/kubedb/mariadb/commit/c9a9596f) Separate StatsService for monitoring (#62)
- [62854641](https://github.com/kubedb/mariadb/commit/62854641) Use MySQLVersion for MySQL images (#61)
- [3c170c56](https://github.com/kubedb/mariadb/commit/3c170c56) Use updated crd spec (#60)
- [873c285e](https://github.com/kubedb/mariadb/commit/873c285e) Rename OffshootLabels to OffshootSelectors (#59)
- [2fd02169](https://github.com/kubedb/mariadb/commit/2fd02169) Revendor api (#58)
- [a127d6cd](https://github.com/kubedb/mariadb/commit/a127d6cd) Use kmodules monitoring and objectstore api (#57)
- [2f79a038](https://github.com/kubedb/mariadb/commit/2f79a038) Support custom configuration (#52)
- [49c67f00](https://github.com/kubedb/mariadb/commit/49c67f00) Merge commit '44e6d4985d93556e39ddcc4677ada5437fc5be64'
- [fb28bc6c](https://github.com/kubedb/mariadb/commit/fb28bc6c) Refactor concourse scripts (#56)
- [4de4ced1](https://github.com/kubedb/mariadb/commit/4de4ced1) Fix command `./hack/make.py test e2e` (#55)
- [3082123e](https://github.com/kubedb/mariadb/commit/3082123e) Set generated binary name to my-operator (#54)
- [5698f314](https://github.com/kubedb/mariadb/commit/5698f314) Don't add admission/v1beta1 group as a prioritized version (#53)
- [696135d5](https://github.com/kubedb/mariadb/commit/696135d5) Fix travis build (#48)
- [c519ef89](https://github.com/kubedb/mariadb/commit/c519ef89) Format shell script (#51)
- [c93e2f40](https://github.com/kubedb/mariadb/commit/c93e2f40) Enable status subresource for crds (#50)
- [edd951ca](https://github.com/kubedb/mariadb/commit/edd951ca) Update client-go to v8.0.0 (#49)
- [520597a6](https://github.com/kubedb/mariadb/commit/520597a6) Merge commit '71850e2c90cda8fc588b7dedb340edf3d316baea'
- [f1549e95](https://github.com/kubedb/mariadb/commit/f1549e95) Support ENV variables in CRDs (#46)
- [67f37780](https://github.com/kubedb/mariadb/commit/67f37780) Updated osm version to 0.7.1 (#47)
- [10e309c0](https://github.com/kubedb/mariadb/commit/10e309c0) Prepare release 0.1.0
- [62a8fbbd](https://github.com/kubedb/mariadb/commit/62a8fbbd) Fixed missing error return (#45)
- [8c05bb83](https://github.com/kubedb/mariadb/commit/8c05bb83) Revendor dependencies (#44)
- [ca811a2e](https://github.com/kubedb/mariadb/commit/ca811a2e) Fix release script (#43)
- [b79541f6](https://github.com/kubedb/mariadb/commit/b79541f6) Add changelog (#42)
- [a2d13c82](https://github.com/kubedb/mariadb/commit/a2d13c82) Concourse (#41)
- [95b2186e](https://github.com/kubedb/mariadb/commit/95b2186e) Fixed kubeconfig plugin for Cloud Providers && Storage is required for MySQL (#40)
- [37762093](https://github.com/kubedb/mariadb/commit/37762093) Refactored E2E testing to support E2E testing with admission webhook in cloud (#38)
- [b6fe72ca](https://github.com/kubedb/mariadb/commit/b6fe72ca) Remove lost+found directory before initializing mysql (#39)
- [18ebb959](https://github.com/kubedb/mariadb/commit/18ebb959) Skip delete requests for empty resources (#37)
- [eeb7add0](https://github.com/kubedb/mariadb/commit/eeb7add0) Don't panic if admission options is nil (#36)
- [ccb59db0](https://github.com/kubedb/mariadb/commit/ccb59db0) Disable admission controllers for webhook server (#35)
- [b1c6c149](https://github.com/kubedb/mariadb/commit/b1c6c149) Separate ApiGroup for Mutating and Validating webhook && upgraded osm to 0.7.0 (#34)
- [b1890f7c](https://github.com/kubedb/mariadb/commit/b1890f7c) Update client-go to 7.0.0 (#33)
- [08c81726](https://github.com/kubedb/mariadb/commit/08c81726) Added update script for mysql-tools:8 (#32)
- [4bbe6c9f](https://github.com/kubedb/mariadb/commit/4bbe6c9f) Added support of mysql:5.7 (#31)
- [e657f512](https://github.com/kubedb/mariadb/commit/e657f512) Add support for one informer and N-eventHandler for snapshot, dromantDB and Job (#30)
- [bbcd48d6](https://github.com/kubedb/mariadb/commit/bbcd48d6) Use metrics from kube apiserver (#29)
- [1687e197](https://github.com/kubedb/mariadb/commit/1687e197) Bundle webhook server and Use SharedInformerFactory (#28)
- [cd0efc00](https://github.com/kubedb/mariadb/commit/cd0efc00) Move MySQL AdmissionWebhook packages into MySQL repository (#27)
- [46065e18](https://github.com/kubedb/mariadb/commit/46065e18) Use mysql:8.0.3 image as mysql:8.0 (#26)
- [1b73529f](https://github.com/kubedb/mariadb/commit/1b73529f) Update README.md
- [62eaa397](https://github.com/kubedb/mariadb/commit/62eaa397) Update README.md
- [c53704c7](https://github.com/kubedb/mariadb/commit/c53704c7) Remove Docker pull count
- [b9ec877e](https://github.com/kubedb/mariadb/commit/b9ec877e) Add travis yaml (#25)
- [ade3571c](https://github.com/kubedb/mariadb/commit/ade3571c) Start next dev cycle
- [b4b749df](https://github.com/kubedb/mariadb/commit/b4b749df) Prepare release 0.1.0-beta.2
- [4d46d95d](https://github.com/kubedb/mariadb/commit/4d46d95d) Migrating to apps/v1 (#23)
- [5ee1ac8c](https://github.com/kubedb/mariadb/commit/5ee1ac8c) Update validation (#22)
- [dd023c50](https://github.com/kubedb/mariadb/commit/dd023c50)  Fix dormantDB matching: pass same type to Equal method (#21)
- [37a1e4fd](https://github.com/kubedb/mariadb/commit/37a1e4fd) Use official code generator scripts (#20)
- [485d3d7c](https://github.com/kubedb/mariadb/commit/485d3d7c) Fixed dormantdb matching & Raised throttling time & Fixed MySQL version Checking (#19)
- [6db2ae8d](https://github.com/kubedb/mariadb/commit/6db2ae8d) Prepare release 0.1.0-beta.1
- [ebbfec2f](https://github.com/kubedb/mariadb/commit/ebbfec2f) converted to k8s 1.9 & Improved InitSpec in DormantDB & Added support for Job watcher & Improved Tests (#17)
- [a484e0e5](https://github.com/kubedb/mariadb/commit/a484e0e5) Fixed logger, analytics and removed rbac stuff (#16)
- [7aa2d1d2](https://github.com/kubedb/mariadb/commit/7aa2d1d2) Add rbac stuffs for mysql-exporter (#15)
- [078098c8](https://github.com/kubedb/mariadb/commit/078098c8)  Review Mysql docker images and Fixed monitring (#14)
- [6877108a](https://github.com/kubedb/mariadb/commit/6877108a) Update README.md
- [1f84a5da](https://github.com/kubedb/mariadb/commit/1f84a5da) Start next dev cycle
- [2f1e4b7d](https://github.com/kubedb/mariadb/commit/2f1e4b7d) Prepare release 0.1.0-beta.0
- [dce1e88e](https://github.com/kubedb/mariadb/commit/dce1e88e) Add release script
- [60ed55cb](https://github.com/kubedb/mariadb/commit/60ed55cb) Rename ms-operator to my-operator (#13)
- [5451d166](https://github.com/kubedb/mariadb/commit/5451d166) Fix Analytics and pass client-id as ENV to Snapshot Job (#12)
- [788ae178](https://github.com/kubedb/mariadb/commit/788ae178) update docker image validation (#11)
- [c966efd5](https://github.com/kubedb/mariadb/commit/c966efd5) Add docker-registry and WorkQueue  (#10)
- [be340103](https://github.com/kubedb/mariadb/commit/be340103) Set client id for analytics (#9)
- [ca11f683](https://github.com/kubedb/mariadb/commit/ca11f683) Fix CRD Registration (#8)
- [2f95c13d](https://github.com/kubedb/mariadb/commit/2f95c13d) Update issue repo link
- [6fffa713](https://github.com/kubedb/mariadb/commit/6fffa713) Update pkg paths to kubedb org (#7)
- [2d4d5c44](https://github.com/kubedb/mariadb/commit/2d4d5c44) Assign default Prometheus Monitoring Port (#6)
- [a7595613](https://github.com/kubedb/mariadb/commit/a7595613) Add Snapshot Backup, Restore and Backup-Scheduler (#4)
- [17a782c6](https://github.com/kubedb/mariadb/commit/17a782c6) Update Dockerfile
- [e92bfec9](https://github.com/kubedb/mariadb/commit/e92bfec9) Add mysql-util docker image (#5)
- [2a4b25ac](https://github.com/kubedb/mariadb/commit/2a4b25ac) Mysql db - Inititalizing  (#2)
- [cbfbc878](https://github.com/kubedb/mariadb/commit/cbfbc878) Update README.md
- [01cab651](https://github.com/kubedb/mariadb/commit/01cab651) Update README.md
- [0aa81cdf](https://github.com/kubedb/mariadb/commit/0aa81cdf) Use client-go 5.x
- [3de10d7f](https://github.com/kubedb/mariadb/commit/3de10d7f) Update ./hack folder (#3)
- [46f05b1f](https://github.com/kubedb/mariadb/commit/46f05b1f) Add skeleton for mysql (#1)
- [73147dba](https://github.com/kubedb/mariadb/commit/73147dba) Merge commit 'be70502b4993171bbad79d2ff89a9844f1c24caa' as 'hack/libbuild'



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.10.0](https://github.com/kubedb/memcached/releases/tag/v0.10.0)

- [58cdf64a](https://github.com/kubedb/memcached/commit/58cdf64a) Prepare for release v0.10.0 (#291)
- [13e5a1fb](https://github.com/kubedb/memcached/commit/13e5a1fb) Update KubeDB api (#290)
- [390739ad](https://github.com/kubedb/memcached/commit/390739ad) Update db container security context (#289)
- [7bf492f1](https://github.com/kubedb/memcached/commit/7bf492f1) Update KubeDB api (#288)
- [d074bf23](https://github.com/kubedb/memcached/commit/d074bf23) Fix make install (#287)
- [ee747948](https://github.com/kubedb/memcached/commit/ee747948) Update repository config (#285)
- [e19e53fe](https://github.com/kubedb/memcached/commit/e19e53fe) Update repository config (#284)
- [e7764dcf](https://github.com/kubedb/memcached/commit/e7764dcf) Update Kubernetes v1.18.9 dependencies (#283)
- [b491d1d9](https://github.com/kubedb/memcached/commit/b491d1d9) Update repository config (#282)
- [beaa42b1](https://github.com/kubedb/memcached/commit/beaa42b1) Update Kubernetes v1.18.9 dependencies (#281)
- [25e0c0a5](https://github.com/kubedb/memcached/commit/25e0c0a5) Update Kubernetes v1.18.9 dependencies (#280)
- [a4a6b2b8](https://github.com/kubedb/memcached/commit/a4a6b2b8) Update repository config (#279)
- [c3b1154b](https://github.com/kubedb/memcached/commit/c3b1154b) Update repository config (#278)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.10.0](https://github.com/kubedb/mongodb/releases/tag/v0.10.0)

- [e14d90b3](https://github.com/kubedb/mongodb/commit/e14d90b3) Prepare for release v0.10.0 (#382)
- [48e752f3](https://github.com/kubedb/mongodb/commit/48e752f3) Update for release Stash@v2021.03.08 (#381)
- [03c8ccb4](https://github.com/kubedb/mongodb/commit/03c8ccb4) Update KubeDB api (#380)
- [bda470f8](https://github.com/kubedb/mongodb/commit/bda470f8) Update db container security context (#379)
- [86bc54a8](https://github.com/kubedb/mongodb/commit/86bc54a8) Update KubeDB api (#378)
- [a406dfca](https://github.com/kubedb/mongodb/commit/a406dfca) Fix make install command (#377)
- [17121476](https://github.com/kubedb/mongodb/commit/17121476) Update install command (#376)
- [05fd7b77](https://github.com/kubedb/mongodb/commit/05fd7b77) Pass stash addon info to AppBinding (#374)
- [787861a3](https://github.com/kubedb/mongodb/commit/787861a3) Create TLS user in `$external` database (#366)
- [dc9cef47](https://github.com/kubedb/mongodb/commit/dc9cef47) Update Kubernetes v1.18.9 dependencies (#375)
- [e8081471](https://github.com/kubedb/mongodb/commit/e8081471) Update Kubernetes v1.18.9 dependencies (#373)
- [612f7350](https://github.com/kubedb/mongodb/commit/612f7350) Update repository config (#372)
- [94410f92](https://github.com/kubedb/mongodb/commit/94410f92) Update repository config (#371)
- [d10b9b03](https://github.com/kubedb/mongodb/commit/d10b9b03) Update Kubernetes v1.18.9 dependencies (#370)
- [132172b4](https://github.com/kubedb/mongodb/commit/132172b4) Update repository config (#369)
- [94fa1536](https://github.com/kubedb/mongodb/commit/94fa1536) #818 MongoDB IPv6 support (#365)
- [9614d777](https://github.com/kubedb/mongodb/commit/9614d777) Update Kubernetes v1.18.9 dependencies (#368)
- [054c7312](https://github.com/kubedb/mongodb/commit/054c7312) Update Kubernetes v1.18.9 dependencies (#367)
- [02bed305](https://github.com/kubedb/mongodb/commit/02bed305) Update repository config (#364)
- [ac0e9a51](https://github.com/kubedb/mongodb/commit/ac0e9a51) Update repository config (#363)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.10.0](https://github.com/kubedb/mysql/releases/tag/v0.10.0)

- [3c97ea11](https://github.com/kubedb/mysql/commit/3c97ea11) Prepare for release v0.10.0 (#375)
- [a7c9b3dc](https://github.com/kubedb/mysql/commit/a7c9b3dc) Inject "--set-gtid-purged=OFF" in backup Task params for clustered MySQL (#374)
- [4d1eba85](https://github.com/kubedb/mysql/commit/4d1eba85) Update for release Stash@v2021.03.08 (#373)
- [36d53c97](https://github.com/kubedb/mysql/commit/36d53c97) Fix default set binary log expire (#372)
- [9118c66e](https://github.com/kubedb/mysql/commit/9118c66e) Update KubeDB api (#371)
- [d323daae](https://github.com/kubedb/mysql/commit/d323daae) Update variable name
- [6fad227c](https://github.com/kubedb/mysql/commit/6fad227c) Update db container security context (#370)
- [9a570f2e](https://github.com/kubedb/mysql/commit/9a570f2e) Update KubeDB api (#369)
- [80e4b857](https://github.com/kubedb/mysql/commit/80e4b857) Fix appbinding type meta (#368)
- [6bc063b7](https://github.com/kubedb/mysql/commit/6bc063b7) Fix install command in Makefile (#367)
- [3400d8c0](https://github.com/kubedb/mysql/commit/3400d8c0) Pass stash addon info to AppBinding (#364)
- [2ddc20c4](https://github.com/kubedb/mysql/commit/2ddc20c4) Add ca bundle to AppBinding (#362)
- [fb55f0e4](https://github.com/kubedb/mysql/commit/fb55f0e4) Purge executed binary log after 3 days by default (#352)
- [92cd744c](https://github.com/kubedb/mysql/commit/92cd744c) Remove `baseServerID` from mysql cr (#356)
- [e99b4e51](https://github.com/kubedb/mysql/commit/e99b4e51) Fix updating mysql status condition when db is not online (#355)
- [d5527967](https://github.com/kubedb/mysql/commit/d5527967) Update repository config (#363)
- [970db7e8](https://github.com/kubedb/mysql/commit/970db7e8) Update repository config (#361)
- [077a4b44](https://github.com/kubedb/mysql/commit/077a4b44) Update Kubernetes v1.18.9 dependencies (#360)
- [7c577664](https://github.com/kubedb/mysql/commit/7c577664) Update repository config (#359)
- [1039210c](https://github.com/kubedb/mysql/commit/1039210c) Update Kubernetes v1.18.9 dependencies (#358)
- [27a7fab8](https://github.com/kubedb/mysql/commit/27a7fab8) Update Kubernetes v1.18.9 dependencies (#357)
- [b94283e9](https://github.com/kubedb/mysql/commit/b94283e9) Update repository config (#354)
- [78af88a4](https://github.com/kubedb/mysql/commit/78af88a4) Update repository config (#353)



## [kubedb/operator](https://github.com/kubedb/operator)

### [v0.17.0](https://github.com/kubedb/operator/releases/tag/v0.17.0)

- [fa0cb596](https://github.com/kubedb/operator/commit/fa0cb596) Prepare for release v0.17.0 (#399)
- [46576385](https://github.com/kubedb/operator/commit/46576385) Update KubeDB api (#397)
- [6f0c1887](https://github.com/kubedb/operator/commit/6f0c1887) Add MariaDB support to kubedb/operator (#396)
- [970e29d7](https://github.com/kubedb/operator/commit/970e29d7) Update repository config (#393)
- [728b320e](https://github.com/kubedb/operator/commit/728b320e) Update repository config (#392)
- [b0f2a1c3](https://github.com/kubedb/operator/commit/b0f2a1c3) Update Kubernetes v1.18.9 dependencies (#391)
- [8f31d09c](https://github.com/kubedb/operator/commit/8f31d09c) Update repository config (#390)
- [12dbdb2d](https://github.com/kubedb/operator/commit/12dbdb2d) Update Kubernetes v1.18.9 dependencies (#389)
- [e3a7e911](https://github.com/kubedb/operator/commit/e3a7e911) Update Kubernetes v1.18.9 dependencies (#388)
- [ebff29a4](https://github.com/kubedb/operator/commit/ebff29a4) Update repository config (#387)
- [65c6529f](https://github.com/kubedb/operator/commit/65c6529f) Update repository config (#386)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.4.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.4.0)

- [2fcad9f8](https://github.com/kubedb/percona-xtradb/commit/2fcad9f8) Prepare for release v0.4.0 (#190)
- [5e925447](https://github.com/kubedb/percona-xtradb/commit/5e925447) Update for release Stash@v2021.03.08 (#189)
- [43546c4a](https://github.com/kubedb/percona-xtradb/commit/43546c4a) Update KubeDB api (#188)
- [86cd32ae](https://github.com/kubedb/percona-xtradb/commit/86cd32ae) Update db container security context (#187)
- [efe459a3](https://github.com/kubedb/percona-xtradb/commit/efe459a3) Update KubeDB api (#186)
- [4cd31b92](https://github.com/kubedb/percona-xtradb/commit/4cd31b92) Fix make install (#185)
- [105b4ca5](https://github.com/kubedb/percona-xtradb/commit/105b4ca5) Fix install command in Makefile (#184)
- [be699bcb](https://github.com/kubedb/percona-xtradb/commit/be699bcb) Pass stash addon info to AppBinding (#182)
- [431bfad8](https://github.com/kubedb/percona-xtradb/commit/431bfad8) Update repository config (#181)
- [37953474](https://github.com/kubedb/percona-xtradb/commit/37953474) Update repository config (#180)
- [387795d2](https://github.com/kubedb/percona-xtradb/commit/387795d2) Update Kubernetes v1.18.9 dependencies (#179)
- [ccf8ee25](https://github.com/kubedb/percona-xtradb/commit/ccf8ee25) Update repository config (#178)
- [9f61328a](https://github.com/kubedb/percona-xtradb/commit/9f61328a) Update Kubernetes v1.18.9 dependencies (#177)
- [9241cd63](https://github.com/kubedb/percona-xtradb/commit/9241cd63) Update Kubernetes v1.18.9 dependencies (#176)
- [3687b603](https://github.com/kubedb/percona-xtradb/commit/3687b603) Update repository config (#175)
- [a8a83f93](https://github.com/kubedb/percona-xtradb/commit/a8a83f93) Update repository config (#174)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.1.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.1.0)

- [ebb5c70](https://github.com/kubedb/pg-coordinator/commit/ebb5c70) Prepare for release v0.1.0 (#13)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.4.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.4.0)

- [b68b46c7](https://github.com/kubedb/pgbouncer/commit/b68b46c7) Prepare for release v0.4.0 (#152)
- [efd337fe](https://github.com/kubedb/pgbouncer/commit/efd337fe) Update KubeDB api (#151)
- [c649bc9b](https://github.com/kubedb/pgbouncer/commit/c649bc9b) Update db container security context (#150)
- [6c7da627](https://github.com/kubedb/pgbouncer/commit/6c7da627) Update KubeDB api (#149)
- [a6254d15](https://github.com/kubedb/pgbouncer/commit/a6254d15) Fix make install (#148)
- [fcdcbe00](https://github.com/kubedb/pgbouncer/commit/fcdcbe00) Update repository config (#146)
- [7e1f30ef](https://github.com/kubedb/pgbouncer/commit/7e1f30ef) Update repository config (#145)
- [eed4411c](https://github.com/kubedb/pgbouncer/commit/eed4411c) Update Kubernetes v1.18.9 dependencies (#144)
- [2f3d4363](https://github.com/kubedb/pgbouncer/commit/2f3d4363) Update repository config (#143)
- [951bb00e](https://github.com/kubedb/pgbouncer/commit/951bb00e) Update Kubernetes v1.18.9 dependencies (#142)
- [13f63fe3](https://github.com/kubedb/pgbouncer/commit/13f63fe3) Update Kubernetes v1.18.9 dependencies (#141)
- [b80a350c](https://github.com/kubedb/pgbouncer/commit/b80a350c) Update repository config (#140)
- [1ae2b26c](https://github.com/kubedb/pgbouncer/commit/1ae2b26c) Update repository config (#139)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.17.0](https://github.com/kubedb/postgres/releases/tag/v0.17.0)

- [b47e7f4b](https://github.com/kubedb/postgres/commit/b47e7f4b) Prepare for release v0.17.0 (#481)
- [a73ac849](https://github.com/kubedb/postgres/commit/a73ac849) Pass stash addon info to AppBinding (#480)
- [d47d78ea](https://github.com/kubedb/postgres/commit/d47d78ea) Added supoort for TimescaleDB (#479)
- [6ac94ae6](https://github.com/kubedb/postgres/commit/6ac94ae6) Added support for Official Postgres Images (#478)
- [0506cb76](https://github.com/kubedb/postgres/commit/0506cb76) Update for release Stash@v2021.03.08 (#477)
- [5d004ff4](https://github.com/kubedb/postgres/commit/5d004ff4) Update Makefile
- [eb84fc88](https://github.com/kubedb/postgres/commit/eb84fc88) TLS support for postgres & Status condition update (#474)
- [a6a365dd](https://github.com/kubedb/postgres/commit/a6a365dd) Fix install command (#473)
- [004b2b8c](https://github.com/kubedb/postgres/commit/004b2b8c) Fix install command in Makefile (#472)
- [6c714c92](https://github.com/kubedb/postgres/commit/6c714c92) Fix install command
- [33eb6d74](https://github.com/kubedb/postgres/commit/33eb6d74) Update repository config (#470)
- [90f48417](https://github.com/kubedb/postgres/commit/90f48417) Update repository config (#469)
- [aa0f0760](https://github.com/kubedb/postgres/commit/aa0f0760) Update Kubernetes v1.18.9 dependencies (#468)
- [43f953d9](https://github.com/kubedb/postgres/commit/43f953d9) Update repository config (#467)
- [8247bcb6](https://github.com/kubedb/postgres/commit/8247bcb6) Update Kubernetes v1.18.9 dependencies (#466)
- [619c8903](https://github.com/kubedb/postgres/commit/619c8903) Update Kubernetes v1.18.9 dependencies (#465)
- [f2998147](https://github.com/kubedb/postgres/commit/f2998147) Update repository config (#464)
- [93d466be](https://github.com/kubedb/postgres/commit/93d466be) Update repository config (#463)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.4.0](https://github.com/kubedb/proxysql/releases/tag/v0.4.0)

- [6e8fb4a1](https://github.com/kubedb/proxysql/commit/6e8fb4a1) Prepare for release v0.4.0 (#169)
- [77bafd23](https://github.com/kubedb/proxysql/commit/77bafd23) Update for release Stash@v2021.03.08 (#168)
- [7bff702a](https://github.com/kubedb/proxysql/commit/7bff702a) Update KubeDB api (#167)
- [7fa81242](https://github.com/kubedb/proxysql/commit/7fa81242) Update db container security context (#166)
- [c218aa7e](https://github.com/kubedb/proxysql/commit/c218aa7e) Update KubeDB api (#165)
- [2705c4e0](https://github.com/kubedb/proxysql/commit/2705c4e0) Fix make install (#164)
- [8498b254](https://github.com/kubedb/proxysql/commit/8498b254) Update Kubernetes v1.18.9 dependencies (#163)
- [fa003df3](https://github.com/kubedb/proxysql/commit/fa003df3) Update repository config (#162)
- [eff1530f](https://github.com/kubedb/proxysql/commit/eff1530f) Update repository config (#161)
- [863abf38](https://github.com/kubedb/proxysql/commit/863abf38) Update Kubernetes v1.18.9 dependencies (#160)
- [70f8d51d](https://github.com/kubedb/proxysql/commit/70f8d51d) Update repository config (#159)
- [0641bc35](https://github.com/kubedb/proxysql/commit/0641bc35) Update Kubernetes v1.18.9 dependencies (#158)
- [a95d45e3](https://github.com/kubedb/proxysql/commit/a95d45e3) Update Kubernetes v1.18.9 dependencies (#157)
- [2229b43f](https://github.com/kubedb/proxysql/commit/2229b43f) Update repository config (#156)
- [a36856a6](https://github.com/kubedb/proxysql/commit/a36856a6) Update repository config (#155)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.10.0](https://github.com/kubedb/redis/releases/tag/v0.10.0)

- [fbb31240](https://github.com/kubedb/redis/commit/fbb31240) Prepare for release v0.10.0 (#314)
- [7ed160b4](https://github.com/kubedb/redis/commit/7ed160b4) Update KubeDB api (#313)
- [232d206e](https://github.com/kubedb/redis/commit/232d206e) Update db container security context (#312)
- [09084d0c](https://github.com/kubedb/redis/commit/09084d0c) Update KubeDB api (#311)
- [62f3cef7](https://github.com/kubedb/redis/commit/62f3cef7) Fix appbinding type meta (#310)
- [bed4e87d](https://github.com/kubedb/redis/commit/bed4e87d) Change redis config structure (#231)
- [3eb9a5b5](https://github.com/kubedb/redis/commit/3eb9a5b5) Update Redis Conditions (#250)
- [df65bfe8](https://github.com/kubedb/redis/commit/df65bfe8) Pass stash addon info to AppBinding (#308)
- [1a4b3fe2](https://github.com/kubedb/redis/commit/1a4b3fe2) Update repository config (#307)
- [fcab4120](https://github.com/kubedb/redis/commit/fcab4120) Update repository config (#306)
- [ffa4a9ba](https://github.com/kubedb/redis/commit/ffa4a9ba) Update Kubernetes v1.18.9 dependencies (#305)
- [5afb498e](https://github.com/kubedb/redis/commit/5afb498e) Update repository config (#304)
- [38e93cb9](https://github.com/kubedb/redis/commit/38e93cb9) Update Kubernetes v1.18.9 dependencies (#303)
- [f3083d8c](https://github.com/kubedb/redis/commit/f3083d8c) Update Kubernetes v1.18.9 dependencies (#302)
- [878b4f7e](https://github.com/kubedb/redis/commit/878b4f7e) Update repository config (#301)
- [d3a2e333](https://github.com/kubedb/redis/commit/d3a2e333) Update repository config (#300)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.4.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.4.0)

- [78195c3](https://github.com/kubedb/replication-mode-detector/commit/78195c3) Prepare for release v0.4.0 (#132)
- [3f1dc9c](https://github.com/kubedb/replication-mode-detector/commit/3f1dc9c) Update KubeDB api (#131)
- [31591ab](https://github.com/kubedb/replication-mode-detector/commit/31591ab) Update repository config (#128)
- [606a9ba](https://github.com/kubedb/replication-mode-detector/commit/606a9ba) Update repository config (#127)
- [69048ba](https://github.com/kubedb/replication-mode-detector/commit/69048ba) Update Kubernetes v1.18.9 dependencies (#126)
- [ae0857b](https://github.com/kubedb/replication-mode-detector/commit/ae0857b) Update Kubernetes v1.18.9 dependencies (#125)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.2.0](https://github.com/kubedb/tests/releases/tag/v0.2.0)

- [d515b35](https://github.com/kubedb/tests/commit/d515b35) Prepare for release v0.2.0 (#109)
- [8af80df](https://github.com/kubedb/tests/commit/8af80df) Remove parameters in clustered MySQL backup tests (#108)
- [34d9ed0](https://github.com/kubedb/tests/commit/34d9ed0) Add Stash backup tests for Elasticsearch (#86)
- [0ae044b](https://github.com/kubedb/tests/commit/0ae044b) Add e2e-tests for Elasticsearch Reconfigure TLS (#98)
- [0ecf747](https://github.com/kubedb/tests/commit/0ecf747) Add test for redis reconfiguration (#43)
- [051e74f](https://github.com/kubedb/tests/commit/051e74f) MariaDB Test with Backup Recovery (#96)
- [84f93a2](https://github.com/kubedb/tests/commit/84f93a2) Update KubeDB api (#107)
- [341b130](https://github.com/kubedb/tests/commit/341b130) Add `ElasticsearchAutoscaler` e2e test (#93)
- [b09219c](https://github.com/kubedb/tests/commit/b09219c) Add Stash Backup & Restore test for MySQL (#102)
- [d094303](https://github.com/kubedb/tests/commit/d094303) Test for MySQL Reconfigure, TLS-Reconfigure and VolumeExpansion (#62)
- [049cfb6](https://github.com/kubedb/tests/commit/049cfb6) Fix failed test for MySQL (#103)
- [f84b277](https://github.com/kubedb/tests/commit/f84b277) Update Kubernetes v1.18.9 dependencies (#105)
- [32e88cf](https://github.com/kubedb/tests/commit/32e88cf) Update repository config (#104)
- [b20b2d9](https://github.com/kubedb/tests/commit/b20b2d9) Update repository config (#101)
- [e720f80](https://github.com/kubedb/tests/commit/e720f80) Update Kubernetes v1.18.9 dependencies (#100)
- [25a6cdd](https://github.com/kubedb/tests/commit/25a6cdd) Add MongoDB ReconfigureTLS Test (#97)
- [1814c42](https://github.com/kubedb/tests/commit/1814c42) Update Elasticsearch go-client (#94)
- [9849e81](https://github.com/kubedb/tests/commit/9849e81) Update Kubernetes v1.18.9 dependencies (#99)




