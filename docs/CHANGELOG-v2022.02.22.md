---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2022.02.22
    name: Changelog-v2022.02.22
    parent: welcome
    weight: 20220222
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2022.02.22/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2022.02.22/
---

# KubeDB v2022.02.22 (2022-02-18)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.25.0](https://github.com/kubedb/apimachinery/releases/tag/v0.25.0)

- [4ed35401](https://github.com/kubedb/apimachinery/commit/4ed35401) Add Elasticsearch dashboard helper methods and constants (#860)
- [8c9cb4b7](https://github.com/kubedb/apimachinery/commit/8c9cb4b7) fix mysqldatabase validator webhook (#870)
- [6be2000e](https://github.com/kubedb/apimachinery/commit/6be2000e) Add Schema Manager for Postgres (#854)
- [fa5b5267](https://github.com/kubedb/apimachinery/commit/fa5b5267) Add helper method for MySQL (#871)
- [161fcef7](https://github.com/kubedb/apimachinery/commit/161fcef7) Remove RedisDatabase crd (#869)
- [f7217890](https://github.com/kubedb/apimachinery/commit/f7217890) Use admission/v1 api types (#868)
- [02c89901](https://github.com/kubedb/apimachinery/commit/02c89901) Cancel concurrent CI runs for same pr/commit (#867)
- [c6db524e](https://github.com/kubedb/apimachinery/commit/c6db524e) Cancel concurrent CI runs for same pr/commit (#866)
- [f1d3fa44](https://github.com/kubedb/apimachinery/commit/f1d3fa44) Remove Enable***Webhook fields from common Config (#865)
- [f0d84187](https://github.com/kubedb/apimachinery/commit/f0d84187) Add ES constants: ElasticsearchJavaOptsEnv (#864)
- [90877a3d](https://github.com/kubedb/apimachinery/commit/90877a3d) Add disableAuth Support in Redis (#863)
- [da0fea34](https://github.com/kubedb/apimachinery/commit/da0fea34) Add support to configure JVM heap in term of percentage (#861)
- [5d665ff1](https://github.com/kubedb/apimachinery/commit/5d665ff1) Add doubleOptIn helpers; Change 'Successful' to 'Current' (#856)
- [1ff8a60c](https://github.com/kubedb/apimachinery/commit/1ff8a60c) Fix dashboard api and webhook helper function (#852)
- [fdad1ab2](https://github.com/kubedb/apimachinery/commit/fdad1ab2) Convert configmap for redis
- [b2887180](https://github.com/kubedb/apimachinery/commit/b2887180) Update repository config (#855)
- [713bb229](https://github.com/kubedb/apimachinery/commit/713bb229) Make dashboard & dashboardInitContainer fields optional (#853)
- [fc35bc33](https://github.com/kubedb/apimachinery/commit/fc35bc33) Add Constants for MariaDB ApplyConfig OpsReq (#851)
- [18a94e28](https://github.com/kubedb/apimachinery/commit/18a94e28) Update constants for Elasticsearch horizontal scaling (#849)
- [c327bc75](https://github.com/kubedb/apimachinery/commit/c327bc75) Add helper method for Mysql Read Replica (#848)
- [967a2137](https://github.com/kubedb/apimachinery/commit/967a2137) Add common condition-related constants & GetPhase function (#845)
- [e6e6d092](https://github.com/kubedb/apimachinery/commit/e6e6d092) Add dashboard image in ElasticsearchVersion (#824)
- [13b91fde](https://github.com/kubedb/apimachinery/commit/13b91fde) Add helper method for  MySQL Read Replica (#847)
- [dfb7dd5c](https://github.com/kubedb/apimachinery/commit/dfb7dd5c) Add `ApplyConfig` on MariaDB Reconfigure Ops Request (#846)
- [449b6d64](https://github.com/kubedb/apimachinery/commit/449b6d64) Use lower case letters
- [ee91f91a](https://github.com/kubedb/apimachinery/commit/ee91f91a) Fix typo in package name (#844)
- [8a260d9a](https://github.com/kubedb/apimachinery/commit/8a260d9a) Add Config Generator for Reconfigure (#835)
- [55f68a75](https://github.com/kubedb/apimachinery/commit/55f68a75) Add support for MySQL Read Only Replica (#827)
- [c2d563c4](https://github.com/kubedb/apimachinery/commit/c2d563c4) Fix linter error
- [de55a914](https://github.com/kubedb/apimachinery/commit/de55a914) Add Timeout on MySQLOpsRequestSpec (#825)
- [6894aa4d](https://github.com/kubedb/apimachinery/commit/6894aa4d) Update Volume Expansion Mode Name in Storage Autoscaler (#843)
- [4834d9c2](https://github.com/kubedb/apimachinery/commit/4834d9c2) Add dashboard and schema-manager apis (#841)
- [1bfbd8a8](https://github.com/kubedb/apimachinery/commit/1bfbd8a8) Add VolumeExpansion Mode in PostgresOpsRequest (#842)
- [9831faf3](https://github.com/kubedb/apimachinery/commit/9831faf3) Add UpdateVersion ops request type (#838)
- [113972a8](https://github.com/kubedb/apimachinery/commit/113972a8) Add EnforceFsGroup field in Postgres Spec (#839)
- [ff687f61](https://github.com/kubedb/apimachinery/commit/ff687f61) Add Changes for MariaDB Offline Volume Expansion and MariaDB AutoScaler (#834)
- [e438a2d8](https://github.com/kubedb/apimachinery/commit/e438a2d8) Fix spelling
- [a90c72c7](https://github.com/kubedb/apimachinery/commit/a90c72c7) Rename ***Overview api types to ***Insight (#840)
- [f3a216bf](https://github.com/kubedb/apimachinery/commit/f3a216bf) Add support of offline volume expansion for Elasticsearch (#826)
- [77708728](https://github.com/kubedb/apimachinery/commit/77708728) Update repository config (#837)
- [de96ed84](https://github.com/kubedb/apimachinery/commit/de96ed84) Add mongodb reprovision ops request (#829)
- [d35fa391](https://github.com/kubedb/apimachinery/commit/d35fa391) Add EphemerStorage in MongoDB (#828)
- [24b06131](https://github.com/kubedb/apimachinery/commit/24b06131) Add constant for mongodb `configuration.js` (#830)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.10.0](https://github.com/kubedb/autoscaler/releases/tag/v0.10.0)

- [5b5cab07](https://github.com/kubedb/autoscaler/commit/5b5cab07) Prepare for release v0.10.0 (#75)
- [1da577b7](https://github.com/kubedb/autoscaler/commit/1da577b7) Add MariaDB Autoscaler | Add expandMode field on Autoscaler (#58)
- [292b4a17](https://github.com/kubedb/autoscaler/commit/292b4a17) Fix typo (#74)
- [f3a518ce](https://github.com/kubedb/autoscaler/commit/f3a518ce) Add suffix to webhook resource (#73)
- [e5683679](https://github.com/kubedb/autoscaler/commit/e5683679) Allow partially installing webhook server (#72)
- [dc1e1a19](https://github.com/kubedb/autoscaler/commit/dc1e1a19) Fix AdmissionReview api version (#71)
- [8baf503a](https://github.com/kubedb/autoscaler/commit/8baf503a) Update dependencies
- [11935336](https://github.com/kubedb/autoscaler/commit/11935336) Add make uninstall & purge targets
- [e6bd0c08](https://github.com/kubedb/autoscaler/commit/e6bd0c08) Fix commands (#69)
- [11aab741](https://github.com/kubedb/autoscaler/commit/11aab741) Cancel concurrent CI runs for same pr/commit (#68)
- [a38279f3](https://github.com/kubedb/autoscaler/commit/a38279f3) Fix linter error (#67)
- [749cca26](https://github.com/kubedb/autoscaler/commit/749cca26) Update dependencies (#66)
- [10510c7f](https://github.com/kubedb/autoscaler/commit/10510c7f) Cancel concurrent CI runs for same pr/commit (#65)
- [f372ee10](https://github.com/kubedb/autoscaler/commit/f372ee10) Introduce separate commands for operator and webhook (#64)
- [5a8b7e36](https://github.com/kubedb/autoscaler/commit/5a8b7e36) Use stash.appscode.dev/apimachinery@v0.18.0 (#63)
- [0252232d](https://github.com/kubedb/autoscaler/commit/0252232d) Update UID generation for GenericResource (#62)
- [8bf7600b](https://github.com/kubedb/autoscaler/commit/8bf7600b) Fix mongodb inMemory shard Autoscaler (#61)
- [4d1c2222](https://github.com/kubedb/autoscaler/commit/4d1c2222) Update SiteInfo (#60)
- [93c5cbaf](https://github.com/kubedb/autoscaler/commit/93c5cbaf) Generate GenericResource
- [21fed0b2](https://github.com/kubedb/autoscaler/commit/21fed0b2) Publish GenericResource (#59)
- [4b57f902](https://github.com/kubedb/autoscaler/commit/4b57f902) Recover from panic in reconcilers (#57)
- [678eab33](https://github.com/kubedb/autoscaler/commit/678eab33) Use Go 1.17 module format
- [bff8d517](https://github.com/kubedb/autoscaler/commit/bff8d517) Update package module path



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.25.0](https://github.com/kubedb/cli/releases/tag/v0.25.0)

- [d22f9b86](https://github.com/kubedb/cli/commit/d22f9b86) Prepare for release v0.25.0 (#654)
- [829e5d49](https://github.com/kubedb/cli/commit/829e5d49) Cancel concurrent CI runs for same pr/commit (#653)
- [2366e0fc](https://github.com/kubedb/cli/commit/2366e0fc) Update dependencies (#652)
- [3a4e8d6a](https://github.com/kubedb/cli/commit/3a4e8d6a) Cancel concurrent CI runs for same pr/commit (#651)
- [21d910b6](https://github.com/kubedb/cli/commit/21d910b6) Use GO 1.17 module format (#650)
- [3972a064](https://github.com/kubedb/cli/commit/3972a064) Use stash.appscode.dev/apimachinery@v0.18.0 (#649)
- [287d32bc](https://github.com/kubedb/cli/commit/287d32bc) Update SiteInfo (#648)
- [7aacbc4e](https://github.com/kubedb/cli/commit/7aacbc4e) Publish GenericResource (#647)
- [926af73f](https://github.com/kubedb/cli/commit/926af73f) Release cli for darwin/arm64 (#646)
- [f575f520](https://github.com/kubedb/cli/commit/f575f520) Recover from panic in reconcilers (#645)
- [5ebd64b6](https://github.com/kubedb/cli/commit/5ebd64b6) Update for release Stash@v2021.11.24 (#644)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.1.0](https://github.com/kubedb/dashboard/releases/tag/v0.1.0)

- [dc2c5cd](https://github.com/kubedb/dashboard/commit/dc2c5cd) Prepare for release v0.1.0 (#15)
- [9444404](https://github.com/kubedb/dashboard/commit/9444404) Cancel concurrent CI runs for same pr/commit (#14)
- [19a0cc3](https://github.com/kubedb/dashboard/commit/19a0cc3) Add support  for config-merger init container (#13)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.25.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.25.0)

- [c5725973](https://github.com/kubedb/elasticsearch/commit/c5725973) Prepare for release v0.25.0 (#565)
- [d7535c40](https://github.com/kubedb/elasticsearch/commit/d7535c40) Add support for Elasticsearch:5.6.16-searchguard (#564)
- [0c348e3d](https://github.com/kubedb/elasticsearch/commit/0c348e3d) Add suffix to webhook resource (#563)
- [72a19921](https://github.com/kubedb/elasticsearch/commit/72a19921) Allow partially installing webhook server (#562)
- [fc6fd671](https://github.com/kubedb/elasticsearch/commit/fc6fd671) Fix AdmissionReview api version (#561)
- [99cac224](https://github.com/kubedb/elasticsearch/commit/99cac224) Fix commands (#559)
- [db3a0ef1](https://github.com/kubedb/elasticsearch/commit/db3a0ef1) Cancel concurrent CI runs for same pr/commit (#558)
- [6fd7c0df](https://github.com/kubedb/elasticsearch/commit/6fd7c0df) Update dependencies (#557)
- [847fe9c4](https://github.com/kubedb/elasticsearch/commit/847fe9c4) Cancel concurrent CI runs for same pr/commit (#555)
- [96a85825](https://github.com/kubedb/elasticsearch/commit/96a85825) Introduce separate commands for operator and webhook (#554)
- [7781b596](https://github.com/kubedb/elasticsearch/commit/7781b596) Use stash.appscode.dev/apimachinery@v0.18.0 (#553)
- [e8d411b4](https://github.com/kubedb/elasticsearch/commit/e8d411b4) Update UID generation for GenericResource (#552)
- [eb2f7d24](https://github.com/kubedb/elasticsearch/commit/eb2f7d24) Add support for JVM heap size in term of percentage (#551)
- [bbb0d8c5](https://github.com/kubedb/elasticsearch/commit/bbb0d8c5) Update SiteInfo (#550)
- [feaf7f2a](https://github.com/kubedb/elasticsearch/commit/feaf7f2a) Generate GenericResource
- [ef1cc55e](https://github.com/kubedb/elasticsearch/commit/ef1cc55e) Publish GenericResource (#549)
- [f1b3203b](https://github.com/kubedb/elasticsearch/commit/f1b3203b) Revert PRODUCT_NAME in makefile (#548)
- [30c80c26](https://github.com/kubedb/elasticsearch/commit/30c80c26) Fix resource patching issue in upsertContainer func (#547)
- [5ab262d3](https://github.com/kubedb/elasticsearch/commit/5ab262d3) Fix service ExternalTrafficPolicy repetitive patch issue (#546)
- [85231bcc](https://github.com/kubedb/elasticsearch/commit/85231bcc) Recover from panic in reconcilers (#545)



## [kubedb/enterprise](https://github.com/kubedb/enterprise)

### [v0.12.0](https://github.com/kubedb/enterprise/releases/tag/v0.12.0)




## [kubedb/installer](https://github.com/kubedb/installer)

### [v2022.02.22](https://github.com/kubedb/installer/releases/tag/v2022.02.22)




## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.9.0](https://github.com/kubedb/mariadb/releases/tag/v0.9.0)

- [6a778dd9](https://github.com/kubedb/mariadb/commit/6a778dd9) Prepare for release v0.9.0 (#135)
- [b4a30c99](https://github.com/kubedb/mariadb/commit/b4a30c99) added-all
- [c6a5cc23](https://github.com/kubedb/mariadb/commit/c6a5cc23) Update validator webhook gvr
- [7d2d3c91](https://github.com/kubedb/mariadb/commit/7d2d3c91) Add suffix to webhook resource (#134)
- [0636c331](https://github.com/kubedb/mariadb/commit/0636c331) Allow partially installing webhook server (#133)
- [a735bbfb](https://github.com/kubedb/mariadb/commit/a735bbfb) Fix AdmissionReview api version (#132)
- [d29518d2](https://github.com/kubedb/mariadb/commit/d29518d2) Fix commands (#130)
- [d04f9e9a](https://github.com/kubedb/mariadb/commit/d04f9e9a) Cancel concurrent CI runs for same pr/commit (#129)
- [8c266208](https://github.com/kubedb/mariadb/commit/8c266208) Update dependencies (#128)
- [4871392b](https://github.com/kubedb/mariadb/commit/4871392b) Cancel concurrent CI runs for same pr/commit (#126)
- [34ff3c21](https://github.com/kubedb/mariadb/commit/34ff3c21) Introduce separate commands for operator and webhook (#125)
- [09286968](https://github.com/kubedb/mariadb/commit/09286968) Use stash.appscode.dev/apimachinery@v0.18.0 (#124)
- [c4f75cfd](https://github.com/kubedb/mariadb/commit/c4f75cfd) Update UID generation for GenericResource (#123)
- [bbdd36d6](https://github.com/kubedb/mariadb/commit/bbdd36d6) Update SiteInfo (#121)
- [10ff7827](https://github.com/kubedb/mariadb/commit/10ff7827) Generate GenericResource
- [cf8ac7fe](https://github.com/kubedb/mariadb/commit/cf8ac7fe) Publish GenericResource (#120)
- [a8d68263](https://github.com/kubedb/mariadb/commit/a8d68263) Allow database service account to get DB object from coordinator (#117)
- [5710c30c](https://github.com/kubedb/mariadb/commit/5710c30c) Revert product name on Makefile (#119)
- [62ec6717](https://github.com/kubedb/mariadb/commit/62ec6717) Update kubedb-community chart name to kubedb-provisioner (#118)
- [e909bd19](https://github.com/kubedb/mariadb/commit/e909bd19) Recover from panic in reconcilers (#116)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.5.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.5.0)

- [c22dc23](https://github.com/kubedb/mariadb-coordinator/commit/c22dc23) Prepare for release v0.5.0 (#35)
- [c55747c](https://github.com/kubedb/mariadb-coordinator/commit/c55747c) Cancel concurrent CI runs for same pr/commit (#34)
- [aff152f](https://github.com/kubedb/mariadb-coordinator/commit/aff152f) Update dependencies (#33)
- [3ee0cd0](https://github.com/kubedb/mariadb-coordinator/commit/3ee0cd0) Cancel concurrent CI runs for same pr/commit (#32)
- [9e0e94c](https://github.com/kubedb/mariadb-coordinator/commit/9e0e94c) Update SiteInfo (#31)
- [f4a25a9](https://github.com/kubedb/mariadb-coordinator/commit/f4a25a9) Publish GenericResource (#30)
- [b086548](https://github.com/kubedb/mariadb-coordinator/commit/b086548) Get ReplicaCount from DB object when StatefulSet isNotFound (#29)
- [548249c](https://github.com/kubedb/mariadb-coordinator/commit/548249c) Recover from panic in reconcilers (#28)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.18.0](https://github.com/kubedb/memcached/releases/tag/v0.18.0)

- [134b2f2d](https://github.com/kubedb/memcached/commit/134b2f2d) Prepare for release v0.18.0 (#345)
- [f0f5c8b4](https://github.com/kubedb/memcached/commit/f0f5c8b4) Add suffix to webhook resource (#344)
- [87d70155](https://github.com/kubedb/memcached/commit/87d70155) Allow partially installing webhook server (#343)
- [5949b5d6](https://github.com/kubedb/memcached/commit/5949b5d6) Fix AdmissionReview api version (#342)
- [31f19773](https://github.com/kubedb/memcached/commit/31f19773) Fix commands (#340)
- [bc8f831e](https://github.com/kubedb/memcached/commit/bc8f831e) Cancel concurrent CI runs for same pr/commit (#339)
- [dd6ce3a7](https://github.com/kubedb/memcached/commit/dd6ce3a7) Update dependencies (#338)
- [0fb58107](https://github.com/kubedb/memcached/commit/0fb58107) Cancel concurrent CI runs for same pr/commit (#337)
- [54c0f656](https://github.com/kubedb/memcached/commit/54c0f656) Introduce separate commands for operator and webhook (#336)
- [930593a8](https://github.com/kubedb/memcached/commit/930593a8) Use stash.appscode.dev/apimachinery@v0.18.0 (#335)
- [61742791](https://github.com/kubedb/memcached/commit/61742791) Update UID generation for GenericResource (#334)
- [bdba4e60](https://github.com/kubedb/memcached/commit/bdba4e60) Update SiteInfo (#333)
- [7fc04444](https://github.com/kubedb/memcached/commit/7fc04444) Generate GenericResource
- [8eb438f6](https://github.com/kubedb/memcached/commit/8eb438f6) Publish GenericResource (#332)
- [247d2c99](https://github.com/kubedb/memcached/commit/247d2c99) Recover from panic in reconcilers (#331)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.18.0](https://github.com/kubedb/mongodb/releases/tag/v0.18.0)

- [7a354d6c](https://github.com/kubedb/mongodb/commit/7a354d6c) Prepare for release v0.18.0 (#462)
- [46b5f2a7](https://github.com/kubedb/mongodb/commit/46b5f2a7) Add suffix to webhook resource (#461)
- [8db1061d](https://github.com/kubedb/mongodb/commit/8db1061d) Allow partially installing webhook server (#460)
- [d21b4c46](https://github.com/kubedb/mongodb/commit/d21b4c46) Fix AdmissionReview api version (#458)
- [85ae88c1](https://github.com/kubedb/mongodb/commit/85ae88c1) Fix commands (#456)
- [e528b327](https://github.com/kubedb/mongodb/commit/e528b327) Cancel concurrent CI runs for same pr/commit (#455)
- [111d3d88](https://github.com/kubedb/mongodb/commit/111d3d88) Update dependencies (#454)
- [417ca61e](https://github.com/kubedb/mongodb/commit/417ca61e) Cancel concurrent CI runs for same pr/commit (#453)
- [2aacc8b2](https://github.com/kubedb/mongodb/commit/2aacc8b2) Introduce separate commands for operator and webhook (#452)
- [aaa48967](https://github.com/kubedb/mongodb/commit/aaa48967) Use stash.appscode.dev/apimachinery@v0.18.0 (#451)
- [b4f039b7](https://github.com/kubedb/mongodb/commit/b4f039b7) Update UID generation for GenericResource (#450)
- [537f1d2a](https://github.com/kubedb/mongodb/commit/537f1d2a) Fix shard health check (#448)
- [787cd3b0](https://github.com/kubedb/mongodb/commit/787cd3b0) Update SiteInfo (#447)
- [2d7b4b0e](https://github.com/kubedb/mongodb/commit/2d7b4b0e) Generate GenericResource
- [7df41e17](https://github.com/kubedb/mongodb/commit/7df41e17) Publish GenericResource (#446)
- [4eaa7aa2](https://github.com/kubedb/mongodb/commit/4eaa7aa2) Add configuration for ephemeral storage (#442)
- [28c8967d](https://github.com/kubedb/mongodb/commit/28c8967d) Add read/write health check (#443)
- [921451d7](https://github.com/kubedb/mongodb/commit/921451d7) Add support to apply `configuration.js` (#445)
- [027e307a](https://github.com/kubedb/mongodb/commit/027e307a) Update Reconcile method (#444)
- [9d609162](https://github.com/kubedb/mongodb/commit/9d609162) Recover from panic in reconcilers (#441)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.18.0](https://github.com/kubedb/mysql/releases/tag/v0.18.0)

- [12b89a3e](https://github.com/kubedb/mysql/commit/12b89a3e) Prepare for release v0.18.0 (#455)
- [03df7640](https://github.com/kubedb/mysql/commit/03df7640) Add Support for MySQL Read Replica (#439)
- [abfa3adc](https://github.com/kubedb/mysql/commit/abfa3adc) Use component specific webhook install command
- [f42185e5](https://github.com/kubedb/mysql/commit/f42185e5) Add suffix to webhook resource (#454)
- [b8c47c15](https://github.com/kubedb/mysql/commit/b8c47c15) Fix AdmissionReview api version (#453)
- [e987420a](https://github.com/kubedb/mysql/commit/e987420a) Fix commands (#451)
- [db3a06de](https://github.com/kubedb/mysql/commit/db3a06de) Cancel concurrent CI runs for same pr/commit (#450)
- [4a4f156e](https://github.com/kubedb/mysql/commit/4a4f156e) Update dependencies (#449)
- [b7209b20](https://github.com/kubedb/mysql/commit/b7209b20) Cancel concurrent CI runs for same pr/commit (#448)
- [f6214514](https://github.com/kubedb/mysql/commit/f6214514) Introduce separate commands for operator and webhook (#447)
- [97ba973b](https://github.com/kubedb/mysql/commit/97ba973b) Use stash.appscode.dev/apimachinery@v0.18.0 (#446)
- [668f50ff](https://github.com/kubedb/mysql/commit/668f50ff) Update UID generation for GenericResource (#445)
- [ac411e95](https://github.com/kubedb/mysql/commit/ac411e95) Remove coordinator container for stand alone instance. (#443)
- [99441193](https://github.com/kubedb/mysql/commit/99441193) Update SiteInfo (#444)
- [a248a8e2](https://github.com/kubedb/mysql/commit/a248a8e2) Generate GenericResource
- [1e7e681b](https://github.com/kubedb/mysql/commit/1e7e681b) Publish GenericResource (#442)
- [2cca63b8](https://github.com/kubedb/mysql/commit/2cca63b8) Rename MySQLClusterModeGroupReplication to MySQLModeGroupReplication (#441)
- [a25b9a4c](https://github.com/kubedb/mysql/commit/a25b9a4c) Pass --set-gtid-purged=off to stash for innodb cluster. (#437)
- [e2533c3a](https://github.com/kubedb/mysql/commit/e2533c3a) Recover from panic in reconcilers (#436)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.3.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.3.0)

- [5784a32](https://github.com/kubedb/mysql-coordinator/commit/5784a32) Prepare for release v0.3.0 (#29)
- [9d7d210](https://github.com/kubedb/mysql-coordinator/commit/9d7d210) Cancel concurrent CI runs for same pr/commit (#28)
- [8c0afbd](https://github.com/kubedb/mysql-coordinator/commit/8c0afbd) Update dependencies (#27)
- [99284f0](https://github.com/kubedb/mysql-coordinator/commit/99284f0) Cancel concurrent CI runs for same pr/commit (#26)
- [bccd960](https://github.com/kubedb/mysql-coordinator/commit/bccd960) Update SiteInfo (#25)
- [d7c0d30](https://github.com/kubedb/mysql-coordinator/commit/d7c0d30) Publish GenericResource (#24)
- [4ffaf21](https://github.com/kubedb/mysql-coordinator/commit/4ffaf21) Recover from panic in reconcilers (#22)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.3.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.3.0)

- [32d8ac8](https://github.com/kubedb/mysql-router-init/commit/32d8ac8) Cancel concurrent CI runs for same pr/commit (#16)
- [c01384d](https://github.com/kubedb/mysql-router-init/commit/c01384d) Cancel concurrent CI runs for same pr/commit (#15)
- [febef73](https://github.com/kubedb/mysql-router-init/commit/febef73) Publish GenericResource (#14)



## [kubedb/operator](https://github.com/kubedb/operator)

### [v0.25.0](https://github.com/kubedb/operator/releases/tag/v0.25.0)




## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.12.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.12.0)

- [d01b2914](https://github.com/kubedb/percona-xtradb/commit/d01b2914) Prepare for release v0.12.0 (#249)
- [6f821254](https://github.com/kubedb/percona-xtradb/commit/6f821254) Add suffix to webhook resource (#248)
- [0be0d287](https://github.com/kubedb/percona-xtradb/commit/0be0d287) Allow partially installing webhook server (#247)
- [a867d68c](https://github.com/kubedb/percona-xtradb/commit/a867d68c) Fix AdmissionReview api version (#246)
- [15c1e045](https://github.com/kubedb/percona-xtradb/commit/15c1e045) Fix commands (#244)
- [3c12213b](https://github.com/kubedb/percona-xtradb/commit/3c12213b) Cancel concurrent CI runs for same pr/commit (#243)
- [1458bd2b](https://github.com/kubedb/percona-xtradb/commit/1458bd2b) Update dependencies (#242)
- [675cd747](https://github.com/kubedb/percona-xtradb/commit/675cd747) Cancel concurrent CI runs for same pr/commit (#241)
- [3c5f5df0](https://github.com/kubedb/percona-xtradb/commit/3c5f5df0) Introduce separate commands for operator and webhook (#240)
- [cb4dd867](https://github.com/kubedb/percona-xtradb/commit/cb4dd867) Use stash.appscode.dev/apimachinery@v0.18.0 (#239)
- [b8bd01a9](https://github.com/kubedb/percona-xtradb/commit/b8bd01a9) Update UID generation for GenericResource (#238)
- [e6b35455](https://github.com/kubedb/percona-xtradb/commit/e6b35455) Update SiteInfo (#236)
- [473b9ba6](https://github.com/kubedb/percona-xtradb/commit/473b9ba6) Generate GenericResource
- [28321621](https://github.com/kubedb/percona-xtradb/commit/28321621) Publish GenericResource (#235)
- [94984a32](https://github.com/kubedb/percona-xtradb/commit/94984a32) Recover from panic in reconcilers (#234)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.9.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.9.0)

- [eb50dc2](https://github.com/kubedb/pg-coordinator/commit/eb50dc2) Prepare for release v0.9.0 (#66)
- [d27428b](https://github.com/kubedb/pg-coordinator/commit/d27428b) Cancel concurrent CI runs for same pr/commit (#65)
- [7beba31](https://github.com/kubedb/pg-coordinator/commit/7beba31) Update dependencies (#64)
- [feed8e5](https://github.com/kubedb/pg-coordinator/commit/feed8e5) Cancel concurrent CI runs for same pr/commit (#63)
- [d509ec3](https://github.com/kubedb/pg-coordinator/commit/d509ec3) Update SiteInfo (#62)
- [dfa09ba](https://github.com/kubedb/pg-coordinator/commit/dfa09ba) Publish GenericResource (#61)
- [3a850da](https://github.com/kubedb/pg-coordinator/commit/3a850da) Fix custom Auth secret issues (#60)
- [5cdea8c](https://github.com/kubedb/pg-coordinator/commit/5cdea8c) Use Postgres CR to get replica count (#59)
- [1070903](https://github.com/kubedb/pg-coordinator/commit/1070903) Recover from panic in reconcilers (#58)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.12.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.12.0)

- [a244100d](https://github.com/kubedb/pgbouncer/commit/a244100d) Prepare for release v0.12.0 (#208)
- [3571411a](https://github.com/kubedb/pgbouncer/commit/3571411a) Add suffix to webhook resource (#207)
- [8d13a7bc](https://github.com/kubedb/pgbouncer/commit/8d13a7bc) Allow partially installing webhook server (#206)
- [05098834](https://github.com/kubedb/pgbouncer/commit/05098834) Fix AdmissionReview api version (#205)
- [117c33a7](https://github.com/kubedb/pgbouncer/commit/117c33a7) Fix commands (#203)
- [876c86d6](https://github.com/kubedb/pgbouncer/commit/876c86d6) Cancel concurrent CI runs for same pr/commit (#202)
- [d23c8939](https://github.com/kubedb/pgbouncer/commit/d23c8939) Update dependencies (#201)
- [3e1ed897](https://github.com/kubedb/pgbouncer/commit/3e1ed897) Cancel concurrent CI runs for same pr/commit (#200)
- [6ab49fde](https://github.com/kubedb/pgbouncer/commit/6ab49fde) Introduce separate commands for operator and webhook (#199)
- [aa1e2c7f](https://github.com/kubedb/pgbouncer/commit/aa1e2c7f) Use stash.appscode.dev/apimachinery@v0.18.0 (#198)
- [b602f703](https://github.com/kubedb/pgbouncer/commit/b602f703) Update UID generation for GenericResource (#197)
- [7acd55f4](https://github.com/kubedb/pgbouncer/commit/7acd55f4) Update SiteInfo (#196)
- [504f39d7](https://github.com/kubedb/pgbouncer/commit/504f39d7) Generate GenericResource
- [e4aaec6c](https://github.com/kubedb/pgbouncer/commit/e4aaec6c) Publish GenericResource (#195)
- [fe1b6138](https://github.com/kubedb/pgbouncer/commit/fe1b6138) Recover from panic in reconcilers (#194)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.25.0](https://github.com/kubedb/postgres/releases/tag/v0.25.0)

- [7b764c0b](https://github.com/kubedb/postgres/commit/7b764c0b) Prepare for release v0.25.0 (#562)
- [3cc55bf0](https://github.com/kubedb/postgres/commit/3cc55bf0) Add suffix to webhook resource (#561)
- [59393ddb](https://github.com/kubedb/postgres/commit/59393ddb) Allow partially installing webhook server (#560)
- [a4eaa7af](https://github.com/kubedb/postgres/commit/a4eaa7af) Fix AdmissionReview api version (#559)
- [bc82ff36](https://github.com/kubedb/postgres/commit/bc82ff36) Fix commands (#557)
- [b4eaa521](https://github.com/kubedb/postgres/commit/b4eaa521) Cancel concurrent CI runs for same pr/commit (#556)
- [b43419f3](https://github.com/kubedb/postgres/commit/b43419f3) Update dependencies (#555)
- [3212a076](https://github.com/kubedb/postgres/commit/3212a076) Cancel concurrent CI runs for same pr/commit (#554)
- [578f48f1](https://github.com/kubedb/postgres/commit/578f48f1) Introduce separate commands for operator and webhook (#552)
- [124489ce](https://github.com/kubedb/postgres/commit/124489ce) Use stash.appscode.dev/apimachinery@v0.18.0 (#553)
- [6af28e8f](https://github.com/kubedb/postgres/commit/6af28e8f) Update UID generation for GenericResource (#551)
- [824b4a89](https://github.com/kubedb/postgres/commit/824b4a89) Update SiteInfo (#550)
- [2d8e23ed](https://github.com/kubedb/postgres/commit/2d8e23ed) Generate GenericResource
- [a933d0fb](https://github.com/kubedb/postgres/commit/a933d0fb) Publish GenericResource (#549)
- [2fbb7c8b](https://github.com/kubedb/postgres/commit/2fbb7c8b) Enforce FsGroup and add permission to get Postgres CR from coordinator (#547)
- [cdf23fcb](https://github.com/kubedb/postgres/commit/cdf23fcb) Fix: remove func SetDefaultResourceLimits call (#548)
- [adf84055](https://github.com/kubedb/postgres/commit/adf84055) Recover from panic in reconcilers (#545)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.12.0](https://github.com/kubedb/proxysql/releases/tag/v0.12.0)

- [0bff10e8](https://github.com/kubedb/proxysql/commit/0bff10e8) Prepare for release v0.12.0 (#224)
- [4781caf4](https://github.com/kubedb/proxysql/commit/4781caf4) Fix AdmissionReview api version
- [b9d175c3](https://github.com/kubedb/proxysql/commit/b9d175c3) Add suffix to webhook resource (#223)
- [9935ddf2](https://github.com/kubedb/proxysql/commit/9935ddf2) Allow partially installing webhook server (#222)
- [31e15e52](https://github.com/kubedb/proxysql/commit/31e15e52) Create namespace if not present in install commands
- [15139595](https://github.com/kubedb/proxysql/commit/15139595) Fix commands (#220)
- [dbbf3ba2](https://github.com/kubedb/proxysql/commit/dbbf3ba2) Cancel concurrent CI runs for same pr/commit (#219)
- [85c46c87](https://github.com/kubedb/proxysql/commit/85c46c87) Update dependencies (#218)
- [ee41ced8](https://github.com/kubedb/proxysql/commit/ee41ced8) Cancel concurrent CI runs for same pr/commit (#217)
- [635f6b9b](https://github.com/kubedb/proxysql/commit/635f6b9b) Introduce separate commands for operator and webhook (#216)
- [056cfac6](https://github.com/kubedb/proxysql/commit/056cfac6) Use stash.appscode.dev/apimachinery@v0.18.0 (#215)
- [335460c7](https://github.com/kubedb/proxysql/commit/335460c7) Update UID generation for GenericResource (#214)
- [2148e1d7](https://github.com/kubedb/proxysql/commit/2148e1d7) Update SiteInfo (#213)
- [1a903feb](https://github.com/kubedb/proxysql/commit/1a903feb) Generate GenericResource
- [d7ec8b90](https://github.com/kubedb/proxysql/commit/d7ec8b90) Publish GenericResource (#212)
- [62769ef2](https://github.com/kubedb/proxysql/commit/62769ef2) Recover from panic in reconcilers (#211)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.18.0](https://github.com/kubedb/redis/releases/tag/v0.18.0)

- [f506b8ad](https://github.com/kubedb/redis/commit/f506b8ad) Prepare for release v0.18.0 (#386)
- [ca77cd30](https://github.com/kubedb/redis/commit/ca77cd30) Fix: Multiple Redis Cluster with same name Monitored by Sentinel (#385)
- [ee2c5d31](https://github.com/kubedb/redis/commit/ee2c5d31) Fix AdmissionReview api version (#384)
- [02db9598](https://github.com/kubedb/redis/commit/02db9598) Add DisableAuth Support For Redis and Sentinel (#372)
- [64751b05](https://github.com/kubedb/redis/commit/64751b05) Fix: health checker for Redis Cluster Mode (#363)
- [10be0855](https://github.com/kubedb/redis/commit/10be0855) Add suffix to webhook resource (#383)
- [c5a2e86f](https://github.com/kubedb/redis/commit/c5a2e86f) Allow partially installing webhook server (#382)
- [46216979](https://github.com/kubedb/redis/commit/46216979) Change command name
- [dd1afb75](https://github.com/kubedb/redis/commit/dd1afb75) Fix admission api alias
- [ed61d9fa](https://github.com/kubedb/redis/commit/ed61d9fa) Fix commands (#379)
- [066c65a5](https://github.com/kubedb/redis/commit/066c65a5) Cancel concurrent CI runs for same pr/commit (#380)
- [63f58773](https://github.com/kubedb/redis/commit/63f58773) Install webhook server chart (#378)
- [4a3be0c8](https://github.com/kubedb/redis/commit/4a3be0c8) Update dependencies (#377)
- [4340c91e](https://github.com/kubedb/redis/commit/4340c91e) Cancel concurrent CI runs for same pr/commit (#376)
- [1719bf95](https://github.com/kubedb/redis/commit/1719bf95) Introduce separate commands for operator and webhook (#375)
- [aceab546](https://github.com/kubedb/redis/commit/aceab546) Use stash.appscode.dev/apimachinery@v0.18.0 (#374)
- [73283002](https://github.com/kubedb/redis/commit/73283002) Update UID generation for GenericResource (#373)
- [2c23c89b](https://github.com/kubedb/redis/commit/2c23c89b) Update SiteInfo (#371)
- [efd0041f](https://github.com/kubedb/redis/commit/efd0041f) Generate GenericResource
- [0e9a3244](https://github.com/kubedb/redis/commit/0e9a3244) Publish GenericResource (#370)
- [b4deca3e](https://github.com/kubedb/redis/commit/b4deca3e) Fix: Volume-Exp Permission Issue from Validator (#369)
- [ae1384d8](https://github.com/kubedb/redis/commit/ae1384d8) Add Container name when exec into pod for clustering (#368)
- [83dcec6d](https://github.com/kubedb/redis/commit/83dcec6d) Recover from panic in reconcilers (#367)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.4.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.4.0)

- [a2adbd9](https://github.com/kubedb/redis-coordinator/commit/a2adbd9) Prepare for release v0.4.0 (#25)
- [7ab65d2](https://github.com/kubedb/redis-coordinator/commit/7ab65d2) Fix: Multiple Redis cluster with same name for Sentinel Monitoring (#24)
- [94043db](https://github.com/kubedb/redis-coordinator/commit/94043db) Disable redis auth (#20)
- [4a5c2e6](https://github.com/kubedb/redis-coordinator/commit/4a5c2e6) Cancel concurrent CI runs for same pr/commit (#23)
- [a207e38](https://github.com/kubedb/redis-coordinator/commit/a207e38) Update dependencies (#22)
- [cedef27](https://github.com/kubedb/redis-coordinator/commit/cedef27) Use Go 1.17 module format
- [335b4f6](https://github.com/kubedb/redis-coordinator/commit/335b4f6) Cancel concurrent CI runs for same pr/commit (#21)
- [17a7a07](https://github.com/kubedb/redis-coordinator/commit/17a7a07) Update SiteInfo (#19)
- [6f6013d](https://github.com/kubedb/redis-coordinator/commit/6f6013d) Publish GenericResource (#18)
- [3785029](https://github.com/kubedb/redis-coordinator/commit/3785029) Recover from panic in reconcilers (#17)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.12.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.12.0)

- [86ec5d2a](https://github.com/kubedb/replication-mode-detector/commit/86ec5d2a) Prepare for release v0.12.0 (#184)
- [21cf5fe5](https://github.com/kubedb/replication-mode-detector/commit/21cf5fe5) Cancel concurrent CI runs for same pr/commit (#183)
- [c8a693ba](https://github.com/kubedb/replication-mode-detector/commit/c8a693ba) Update dependencies (#182)
- [31268557](https://github.com/kubedb/replication-mode-detector/commit/31268557) Cancel concurrent CI runs for same pr/commit (#181)
- [c471f782](https://github.com/kubedb/replication-mode-detector/commit/c471f782) Update SiteInfo (#180)
- [301a0b0c](https://github.com/kubedb/replication-mode-detector/commit/301a0b0c) Publish GenericResource (#179)
- [157723f2](https://github.com/kubedb/replication-mode-detector/commit/157723f2) Recover from panic in reconcilers (#178)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.1.0](https://github.com/kubedb/schema-manager/releases/tag/v0.1.0)

- [27cbd85f](https://github.com/kubedb/schema-manager/commit/27cbd85f) Prepare for release v0.1.0 (#21)
- [e599be46](https://github.com/kubedb/schema-manager/commit/e599be46) Add Schema-Manager support for PostgreSQL (#12)
- [c0b7b037](https://github.com/kubedb/schema-manager/commit/c0b7b037) Reflect stash-v2022.02.22 related changes for MongoDB (#13)
- [cb194e15](https://github.com/kubedb/schema-manager/commit/cb194e15) Add stash support for mysqldatabase (#19)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.10.0](https://github.com/kubedb/tests/releases/tag/v0.10.0)

- [72008dce](https://github.com/kubedb/tests/commit/72008dce) Prepare for release v0.10.0 (#169)
- [9f48d54c](https://github.com/kubedb/tests/commit/9f48d54c) Cancel concurrent CI runs for same pr/commit (#168)
- [39fb2faa](https://github.com/kubedb/tests/commit/39fb2faa) Update dependencies (#167)
- [82bef4de](https://github.com/kubedb/tests/commit/82bef4de) Update dependencies (#166)
- [3de40073](https://github.com/kubedb/tests/commit/3de40073) Use stash.appscode.dev/apimachinery@v0.18.0 (#165)
- [02695e02](https://github.com/kubedb/tests/commit/02695e02) Update SiteInfo (#164)
- [917f979c](https://github.com/kubedb/tests/commit/917f979c) Update dependencies (#163)
- [bd430a5e](https://github.com/kubedb/tests/commit/bd430a5e) Recover from panic in reconcilers (#159)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.1.0](https://github.com/kubedb/ui-server/releases/tag/v0.1.0)

- [62f41a7](https://github.com/kubedb/ui-server/commit/62f41a7) Prepare for release v0.1.0 (#25)
- [004104c](https://github.com/kubedb/ui-server/commit/004104c) Cancel concurrent CI runs for same pr/commit (#24)
- [757f36a](https://github.com/kubedb/ui-server/commit/757f36a) Update uid generation (#23)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.1.0](https://github.com/kubedb/webhook-server/releases/tag/v0.1.0)

- [70336af](https://github.com/kubedb/webhook-server/commit/70336af) Prepare for release v0.1.0 (#11)




