---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2024.1.31
    name: Changelog-v2024.1.31
    parent: welcome
    weight: 20240131
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2024.1.31/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2024.1.31/
---

# KubeDB v2024.1.31 (2024-02-02)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.41.0](https://github.com/kubedb/apimachinery/releases/tag/v0.41.0)

- [447a890a](https://github.com/kubedb/apimachinery/commit/447a890af) Update kubestash
- [a81b9dc2](https://github.com/kubedb/apimachinery/commit/a81b9dc28) Increase CPU resource for mongo versions >= 6 (#1140)
- [c711b3ab](https://github.com/kubedb/apimachinery/commit/c711b3abb) ferretdb apm fix (#1138)
- [02bd64df](https://github.com/kubedb/apimachinery/commit/02bd64dfa) Update security context based on version (#1137)
- [2a63b8b1](https://github.com/kubedb/apimachinery/commit/2a63b8b1a) Update deps
- [2293744e](https://github.com/kubedb/apimachinery/commit/2293744e3) Update deps
- [32a0f294](https://github.com/kubedb/apimachinery/commit/32a0f2944) Update deps
- [c389dcb1](https://github.com/kubedb/apimachinery/commit/c389dcb17) Add Singlestore Config Type (#1136)
- [ef7f62fb](https://github.com/kubedb/apimachinery/commit/ef7f62fbd) Defaulting RunAsGroup (#1134)
- [e08f63ba](https://github.com/kubedb/apimachinery/commit/e08f63ba0) Minox fixes in rlease (#1135)
- [760f1c55](https://github.com/kubedb/apimachinery/commit/760f1c554) Ferretdb webhook and apis updated (#1132)
- [958de8ec](https://github.com/kubedb/apimachinery/commit/958de8ec3) Fix spelling mistakes in dashboard. (#1133)
- [f614ab97](https://github.com/kubedb/apimachinery/commit/f614ab976) Fix release issues and add version 28.0.1 (#1131)
- [df53756a](https://github.com/kubedb/apimachinery/commit/df53756a3) Fix dashboard config merger command. (#1126)
- [4b8a46ab](https://github.com/kubedb/apimachinery/commit/4b8a46ab1) Add kafka connector webhook (#1128)
- [3e06dc03](https://github.com/kubedb/apimachinery/commit/3e06dc03a) Update Rabbitmq helpers and webhooks (#1130)
- [23153f41](https://github.com/kubedb/apimachinery/commit/23153f41f) Add ZooKeeper Standalone Mode (#1129)
- [650406ba](https://github.com/kubedb/apimachinery/commit/650406ba8) Remove replica condition for Pgpool (#1127)
- [dbd8e067](https://github.com/kubedb/apimachinery/commit/dbd8e0679) Update docker/docker
- [a28b2662](https://github.com/kubedb/apimachinery/commit/a28b2662e) Add validator to check negative number of replicas. (#1124)
- [cc189c3c](https://github.com/kubedb/apimachinery/commit/cc189c3c8) Add utilities to extract databaseInfo (#1123)
- [ceef191e](https://github.com/kubedb/apimachinery/commit/ceef191e0) Fix short name for FerretDBVersion
- [ef49cbfa](https://github.com/kubedb/apimachinery/commit/ef49cbfa8) Update deps
- [f85d1410](https://github.com/kubedb/apimachinery/commit/f85d14100) Without non-root (#1122)
- [79fd675a](https://github.com/kubedb/apimachinery/commit/79fd675a0) Add `PausedBackups` field into `OpsRequestStatus` (#1114)
- [778a1af2](https://github.com/kubedb/apimachinery/commit/778a1af25) Add FerretDB Apis (#1119)
- [329083aa](https://github.com/kubedb/apimachinery/commit/329083aa6) Add missing entries while ignoring openapi schema (#1121)
- [0f8ac911](https://github.com/kubedb/apimachinery/commit/0f8ac9110) Fix API for new Databases (#1120)
- [b625c64c](https://github.com/kubedb/apimachinery/commit/b625c64c5) Fix issues with Pgpool HealthChecker field and version check in webhook (#1118)
- [e78c6ff7](https://github.com/kubedb/apimachinery/commit/e78c6ff74) Remove unnecessary apis for singlestore (#1117)
- [6e98cd41](https://github.com/kubedb/apimachinery/commit/6e98cd41c) Add Rabbitmq API (#1109)
- [e7a088fa](https://github.com/kubedb/apimachinery/commit/e7a088faf) Remove api call from Solr setDefaults. (#1116)
- [a73a825b](https://github.com/kubedb/apimachinery/commit/a73a825b7) Add Solr API (#1110)
- [9d687049](https://github.com/kubedb/apimachinery/commit/9d6870498) Pgpool Backend Set to Required (#1113)
- [72d44aef](https://github.com/kubedb/apimachinery/commit/72d44aef7) Fix ElasticsearchDashboard constants
- [0c40a769](https://github.com/kubedb/apimachinery/commit/0c40a7698) Change dashboard api group to elasticsearch (#1112)
- [85e4ae23](https://github.com/kubedb/apimachinery/commit/85e4ae232) Add ZooKeeper API (#1104)
- [ee446682](https://github.com/kubedb/apimachinery/commit/ee446682d) Add Pgpool apis (#1103)
- [4995ebf3](https://github.com/kubedb/apimachinery/commit/4995ebf3d) Add Druid API (#1111)
- [556a36df](https://github.com/kubedb/apimachinery/commit/556a36dfe) Add SingleStore APIS (#1108)
- [a72bb1ff](https://github.com/kubedb/apimachinery/commit/a72bb1ffc) Add runAsGroup field in mgVersion api (#1107)
- [1ee5ee41](https://github.com/kubedb/apimachinery/commit/1ee5ee41d) Add Kafka Connect Cluster and Connector APIs (#1066)
- [2fd99ee8](https://github.com/kubedb/apimachinery/commit/2fd99ee82) Fix replica count for arbiter & hidden node (#1106)
- [4e194f0a](https://github.com/kubedb/apimachinery/commit/4e194f0a2) Implement validator for autoscalers (#1105)
- [6a454592](https://github.com/kubedb/apimachinery/commit/6a4545928) Add kubestash controller for changing kubeDB phase (#1096)
- [44757753](https://github.com/kubedb/apimachinery/commit/447577539) Ignore validators.autoscaling.kubedb.com webhook handlers
- [45cbf75e](https://github.com/kubedb/apimachinery/commit/45cbf75e3) Update deps
- [dc224c1a](https://github.com/kubedb/apimachinery/commit/dc224c1a1) Remove crd informer (#1102)
- [87c402a1](https://github.com/kubedb/apimachinery/commit/87c402a1a) Remove discovery.ResourceMapper (#1101)
- [a1d475ce](https://github.com/kubedb/apimachinery/commit/a1d475ceb) Replace deprecated PollImmediate (#1100)
- [75db4a37](https://github.com/kubedb/apimachinery/commit/75db4a378) Add ConfigureOpenAPI helper (#1099)
- [83be295b](https://github.com/kubedb/apimachinery/commit/83be295b0) update sidekick deps
- [032b2721](https://github.com/kubedb/apimachinery/commit/032b27211) Fix linter
- [389a934c](https://github.com/kubedb/apimachinery/commit/389a934c7) Use k8s 1.29 client libs (#1093)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.26.0](https://github.com/kubedb/autoscaler/releases/tag/v0.26.0)




## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.41.0](https://github.com/kubedb/cli/releases/tag/v0.41.0)

- [8ff0608c](https://github.com/kubedb/cli/commit/8ff0608c) Prepare for release v0.41.0 (#755)
- [7aeaa861](https://github.com/kubedb/cli/commit/7aeaa861) Monitor CLI added for check-connection and aggregate all monitor CLI (#754)
- [a67dadc9](https://github.com/kubedb/cli/commit/a67dadc9) Prepare for release v0.41.0-rc.1 (#753)
- [d1b206ee](https://github.com/kubedb/cli/commit/d1b206ee) Update deps (#752)
- [50f15d19](https://github.com/kubedb/cli/commit/50f15d19) Update deps (#751)
- [64ad0b63](https://github.com/kubedb/cli/commit/64ad0b63) Prepare for release v0.41.0-rc.0 (#749)
- [d188eae6](https://github.com/kubedb/cli/commit/d188eae6) Grafana dashboard's metric checking CLI (#740)
- [234b7051](https://github.com/kubedb/cli/commit/234b7051) Prepare for release v0.41.0-beta.1 (#748)
- [1ebdd532](https://github.com/kubedb/cli/commit/1ebdd532) Update deps
- [c0165e83](https://github.com/kubedb/cli/commit/c0165e83) Prepare for release v0.41.0-beta.0 (#747)
- [d9c905e5](https://github.com/kubedb/cli/commit/d9c905e5) Update deps (#746)
- [bc415a1d](https://github.com/kubedb/cli/commit/bc415a1d) Update deps (#745)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.0.4](https://github.com/kubedb/crd-manager/releases/tag/v0.0.4)

- [a45ec91](https://github.com/kubedb/crd-manager/commit/a45ec91) Prepare for release v0.0.4 (#13)
- [39cec60](https://github.com/kubedb/crd-manager/commit/39cec60) Fix deploy-to-kind make target



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.17.0](https://github.com/kubedb/dashboard/releases/tag/v0.17.0)




## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.0.10](https://github.com/kubedb/db-client-go/releases/tag/v0.0.10)

- [902c39a0](https://github.com/kubedb/db-client-go/commit/902c39a0) Prepare for release v0.0.10 (#86)
- [377f8591](https://github.com/kubedb/db-client-go/commit/377f8591) Update deps
- [67567b71](https://github.com/kubedb/db-client-go/commit/67567b71) Update deps (#85)
- [4e2471e3](https://github.com/kubedb/db-client-go/commit/4e2471e3) Update deps (#84)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.0.4](https://github.com/kubedb/druid/releases/tag/v0.0.4)

- [8d4fdb6](https://github.com/kubedb/druid/commit/8d4fdb6) Prepare for release v0.0.4 (#8)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.41.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.41.0)

- [3d0feb70](https://github.com/kubedb/elasticsearch/commit/3d0feb70f) Prepare for release v0.41.0 (#704)
- [f287ef9f](https://github.com/kubedb/elasticsearch/commit/f287ef9f1) Prepare for release v0.41.0-rc.1 (#703)
- [fcabe6ba](https://github.com/kubedb/elasticsearch/commit/fcabe6bae) Update deps (#702)
- [861d01f3](https://github.com/kubedb/elasticsearch/commit/861d01f30) Update deps (#701)
- [69735e9e](https://github.com/kubedb/elasticsearch/commit/69735e9e1) Prepare for release v0.41.0-rc.0 (#700)
- [c410b39f](https://github.com/kubedb/elasticsearch/commit/c410b39f5) Prepare for release v0.41.0-beta.1 (#699)
- [3394f1d1](https://github.com/kubedb/elasticsearch/commit/3394f1d13) Use ptr.Deref(); Update deps
- [f00ee052](https://github.com/kubedb/elasticsearch/commit/f00ee052e) Update ci & makefile for crd-manager (#698)
- [e37e6d63](https://github.com/kubedb/elasticsearch/commit/e37e6d631) Add catalog client in scheme. (#697)
- [a46bfd41](https://github.com/kubedb/elasticsearch/commit/a46bfd41b) Add Support for DB phase change for restoring using KubeStash (#696)
- [9cbac2fc](https://github.com/kubedb/elasticsearch/commit/9cbac2fc4) Update makefile for dynamic crd installer (#695)
- [3ab4d77d](https://github.com/kubedb/elasticsearch/commit/3ab4d77d2) Prepare for release v0.41.0-beta.0 (#694)
- [c38c61cb](https://github.com/kubedb/elasticsearch/commit/c38c61cbc) Dynamically start crd controller (#693)
- [6a798d30](https://github.com/kubedb/elasticsearch/commit/6a798d309) Update deps (#692)
- [bdf034a4](https://github.com/kubedb/elasticsearch/commit/bdf034a49) Update deps (#691)
- [ea22eecb](https://github.com/kubedb/elasticsearch/commit/ea22eecb2) Add openapi configuration for webhook server (#690)
- [b97636cd](https://github.com/kubedb/elasticsearch/commit/b97636cd1) Update lint command
- [0221ac14](https://github.com/kubedb/elasticsearch/commit/0221ac14e) Update deps
- [b4cb8d60](https://github.com/kubedb/elasticsearch/commit/b4cb8d603) Use k8s 1.29 client libs (#689)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.4.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.4.0)

- [11a8a76](https://github.com/kubedb/elasticsearch-restic-plugin/commit/11a8a76) Prepare for release v0.4.0 (#19)
- [675caf7](https://github.com/kubedb/elasticsearch-restic-plugin/commit/675caf7) Prepare for release v0.4.0-rc.1 (#18)
- [18ea6da](https://github.com/kubedb/elasticsearch-restic-plugin/commit/18ea6da) Prepare for release v0.4.0-rc.0 (#17)
- [584dfd9](https://github.com/kubedb/elasticsearch-restic-plugin/commit/584dfd9) Prepare for release v0.4.0-beta.1 (#16)
- [5e9aef5](https://github.com/kubedb/elasticsearch-restic-plugin/commit/5e9aef5) Prepare for release v0.4.0-beta.0 (#15)
- [2fdcafa](https://github.com/kubedb/elasticsearch-restic-plugin/commit/2fdcafa) Use k8s 1.29 client libs (#14)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.0.4](https://github.com/kubedb/ferretdb/releases/tag/v0.0.4)

- [19ac254](https://github.com/kubedb/ferretdb/commit/19ac254) Prepare for release v0.0.4 (#9)
- [1135ae9](https://github.com/kubedb/ferretdb/commit/1135ae9) Auth secret name change and codebase Refactor (#8)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2024.1.31](https://github.com/kubedb/installer/releases/tag/v2024.1.31)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.12.0](https://github.com/kubedb/kafka/releases/tag/v0.12.0)

- [8ffe42e6](https://github.com/kubedb/kafka/commit/8ffe42e6) Prepare for release v0.12.0 (#76)
- [511914c2](https://github.com/kubedb/kafka/commit/511914c2) Prepare for release v0.12.0-rc.1 (#75)
- [fb908cf2](https://github.com/kubedb/kafka/commit/fb908cf2) Update deps (#74)
- [cccaf86c](https://github.com/kubedb/kafka/commit/cccaf86c) Update deps (#73)
- [9d73e3ce](https://github.com/kubedb/kafka/commit/9d73e3ce) Prepare for release v0.12.0-rc.0 (#71)
- [c1d08f75](https://github.com/kubedb/kafka/commit/c1d08f75) Remove cassandra, clickhouse, etcd flags
- [e7283583](https://github.com/kubedb/kafka/commit/e7283583) Fix podtemplate containers reference isuue (#70)
- [6d04bf0f](https://github.com/kubedb/kafka/commit/6d04bf0f) Add termination policy for kafka and connect cluster (#69)
- [34f4967f](https://github.com/kubedb/kafka/commit/34f4967f) Prepare for release v0.12.0-beta.1 (#68)
- [7176931c](https://github.com/kubedb/kafka/commit/7176931c) Move Kafka Podtemplate to ofshoot-api v2 (#66)
- [9454adf6](https://github.com/kubedb/kafka/commit/9454adf6) Update ci & makefile for crd-manager (#67)
- [fda770d8](https://github.com/kubedb/kafka/commit/fda770d8) Add kafka connector controller (#65)
- [6ed0ccd4](https://github.com/kubedb/kafka/commit/6ed0ccd4) Add Kafka connect  controller (#44)
- [18e9a45c](https://github.com/kubedb/kafka/commit/18e9a45c) update deps (#64)
- [a7dfb409](https://github.com/kubedb/kafka/commit/a7dfb409) Update makefile for dynamic crd installer (#63)
- [f9350578](https://github.com/kubedb/kafka/commit/f9350578) Prepare for release v0.12.0-beta.0 (#62)
- [692f2bef](https://github.com/kubedb/kafka/commit/692f2bef) Dynamically start crd controller (#61)
- [a50dc8b4](https://github.com/kubedb/kafka/commit/a50dc8b4) Update deps (#60)
- [7ff28ed7](https://github.com/kubedb/kafka/commit/7ff28ed7) Update deps (#59)
- [16130571](https://github.com/kubedb/kafka/commit/16130571) Add openapi configuration for webhook server (#58)
- [cc465de9](https://github.com/kubedb/kafka/commit/cc465de9) Use k8s 1.29 client libs (#57)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.4.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.4.0)

- [7d51761](https://github.com/kubedb/kubedb-manifest-plugin/commit/7d51761) Prepare for release v0.4.0 (#40)
- [5daf9ce](https://github.com/kubedb/kubedb-manifest-plugin/commit/5daf9ce) Prepare for release v0.4.0-rc.1 (#39)
- [b7ec4a4](https://github.com/kubedb/kubedb-manifest-plugin/commit/b7ec4a4) Prepare for release v0.4.0-rc.0 (#38)
- [c77b4ae](https://github.com/kubedb/kubedb-manifest-plugin/commit/c77b4ae) Prepare for release v0.4.0-beta.1 (#37)
- [6a8a822](https://github.com/kubedb/kubedb-manifest-plugin/commit/6a8a822) Update component name (#35)
- [c315615](https://github.com/kubedb/kubedb-manifest-plugin/commit/c315615) Prepare for release v0.4.0-beta.0 (#36)
- [5ce328d](https://github.com/kubedb/kubedb-manifest-plugin/commit/5ce328d) Use k8s 1.29 client libs (#34)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.25.0](https://github.com/kubedb/mariadb/releases/tag/v0.25.0)

- [bcf4484c](https://github.com/kubedb/mariadb/commit/bcf4484c9) Fix mariadb health check (#256)
- [fcfa0e96](https://github.com/kubedb/mariadb/commit/fcfa0e966) Prepare for release v0.25.0 (#257)
- [2ca44131](https://github.com/kubedb/mariadb/commit/2ca441314) Prepare for release v0.25.0-rc.1 (#255)
- [dad5b837](https://github.com/kubedb/mariadb/commit/dad5b837a) Update deps (#254)
- [a210c867](https://github.com/kubedb/mariadb/commit/a210c8675) Fix appbinding scheme (#251)
- [81f985cd](https://github.com/kubedb/mariadb/commit/81f985cd7) Update deps (#253)
- [4bdcd6cc](https://github.com/kubedb/mariadb/commit/4bdcd6cca) Prepare for release v0.25.0-rc.0 (#252)
- [c4d4942f](https://github.com/kubedb/mariadb/commit/c4d4942f8) Prepare for release v0.25.0-beta.1 (#250)
- [25fe3917](https://github.com/kubedb/mariadb/commit/25fe39177) Use ptr.Deref(); Update deps
- [c76704cc](https://github.com/kubedb/mariadb/commit/c76704cc8) Fix ci & makefile for crd-manager (#249)
- [67396abb](https://github.com/kubedb/mariadb/commit/67396abb9) Incorporate with apimachinery package name change from `stash` to `restore` (#248)
- [b93ddce3](https://github.com/kubedb/mariadb/commit/b93ddce3d) Prepare for release v0.25.0-beta.0 (#247)
- [8099af6d](https://github.com/kubedb/mariadb/commit/8099af6d9) Dynamically start crd controller (#246)
- [0a9dd9e0](https://github.com/kubedb/mariadb/commit/0a9dd9e03) Update deps (#245)
- [5c548629](https://github.com/kubedb/mariadb/commit/5c548629e) Update deps (#244)
- [0f9ea4f2](https://github.com/kubedb/mariadb/commit/0f9ea4f20) Update deps
- [89641d3c](https://github.com/kubedb/mariadb/commit/89641d3c7) Use k8s 1.29 client libs (#242)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.1.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.1.0)

- [bd6798f](https://github.com/kubedb/mariadb-archiver/commit/bd6798f) Prepare for release v0.1.0 (#8)
- [8743316](https://github.com/kubedb/mariadb-archiver/commit/8743316) Prepare for release v0.1.0-rc.1 (#7)
- [90b9d66](https://github.com/kubedb/mariadb-archiver/commit/90b9d66) Prepare for release v0.1.0-rc.0 (#6)
- [e8564fe](https://github.com/kubedb/mariadb-archiver/commit/e8564fe) Prepare for release v0.1.0-beta.1 (#5)
- [e5e8945](https://github.com/kubedb/mariadb-archiver/commit/e5e8945) Don't use fail-fast
- [8c8e09a](https://github.com/kubedb/mariadb-archiver/commit/8c8e09a) Prepare for release v0.1.0-beta.0 (#4)
- [90ae04c](https://github.com/kubedb/mariadb-archiver/commit/90ae04c) Use k8s 1.29 client libs (#3)
- [b3067c8](https://github.com/kubedb/mariadb-archiver/commit/b3067c8) Fix binlog command
- [5cc0b6a](https://github.com/kubedb/mariadb-archiver/commit/5cc0b6a) Fix release workflow
- [910b7ce](https://github.com/kubedb/mariadb-archiver/commit/910b7ce) Prepare for release v0.1.0 (#1)
- [3801668](https://github.com/kubedb/mariadb-archiver/commit/3801668) mysql -> mariadb
- [4e905fb](https://github.com/kubedb/mariadb-archiver/commit/4e905fb) Implemenet new algorithm for archiver and restorer (#5)
- [22701c8](https://github.com/kubedb/mariadb-archiver/commit/22701c8) Fix 5.7.x build
- [6da2b1c](https://github.com/kubedb/mariadb-archiver/commit/6da2b1c) Update build matrix
- [e2f6244](https://github.com/kubedb/mariadb-archiver/commit/e2f6244) Use separate dockerfile per mysql version (#9)
- [e800623](https://github.com/kubedb/mariadb-archiver/commit/e800623) Prepare for release v0.2.0 (#8)
- [b9f6ec5](https://github.com/kubedb/mariadb-archiver/commit/b9f6ec5) Install mysqlbinlog (#7)
- [c46d991](https://github.com/kubedb/mariadb-archiver/commit/c46d991) Use appscode-images as base image (#6)
- [721eaa8](https://github.com/kubedb/mariadb-archiver/commit/721eaa8) Prepare for release v0.1.0 (#4)
- [8c65d14](https://github.com/kubedb/mariadb-archiver/commit/8c65d14) Prepare for release v0.1.0-rc.1 (#3)
- [f79286a](https://github.com/kubedb/mariadb-archiver/commit/f79286a) Prepare for release v0.1.0-rc.0 (#2)
- [dcd2e30](https://github.com/kubedb/mariadb-archiver/commit/dcd2e30) Fix wal-g binary
- [6c20a4a](https://github.com/kubedb/mariadb-archiver/commit/6c20a4a) Fix build
- [f034e7b](https://github.com/kubedb/mariadb-archiver/commit/f034e7b) Add build script (#1)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.21.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.21.0)

- [6e2b4dee](https://github.com/kubedb/mariadb-coordinator/commit/6e2b4dee) Prepare for release v0.21.0 (#107)
- [e0e9c489](https://github.com/kubedb/mariadb-coordinator/commit/e0e9c489) Fix MariaDB Health Check (#106)
- [4289bcd1](https://github.com/kubedb/mariadb-coordinator/commit/4289bcd1) Prepare for release v0.21.0-rc.1 (#105)
- [34f610f7](https://github.com/kubedb/mariadb-coordinator/commit/34f610f7) Update deps (#104)
- [13dbe3f7](https://github.com/kubedb/mariadb-coordinator/commit/13dbe3f7) Update deps (#103)
- [15a83758](https://github.com/kubedb/mariadb-coordinator/commit/15a83758) Prepare for release v0.21.0-rc.0 (#102)
- [1c30e710](https://github.com/kubedb/mariadb-coordinator/commit/1c30e710) Prepare for release v0.21.0-beta.1 (#101)
- [28677618](https://github.com/kubedb/mariadb-coordinator/commit/28677618) Prepare for release v0.21.0-beta.0 (#100)
- [655a2c66](https://github.com/kubedb/mariadb-coordinator/commit/655a2c66) Update deps (#99)
- [ef206cfe](https://github.com/kubedb/mariadb-coordinator/commit/ef206cfe) Update deps (#98)
- [ef72c98b](https://github.com/kubedb/mariadb-coordinator/commit/ef72c98b) Use k8s 1.29 client libs (#97)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.1.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.1.0)

- [d28c59f](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/d28c59f) Prepare for release v0.1.0 (#11)
- [299687e](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/299687e) Update README.md (#9)
- [00e8552](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/00e8552) Update deps (#8)
- [ac2caaf](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/ac2caaf) Update pause and resume queries (#1)
- [066b41c](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/066b41c) Prepare for release v0.1.0-rc.1 (#7)
- [ebd73c7](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/ebd73c7) Prepare for release v0.1.0-rc.0 (#6)
- [adac38d](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/adac38d) Prepare for release v0.1.0-beta.1 (#5)
- [09f68b7](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/09f68b7) Prepare for release v0.1.0-beta.0 (#4)
- [7407444](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/7407444) Use k8s 1.29 client libs (#3)
- [933e138](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/933e138) Prepare for release v0.1.0 (#2)
- [5d38f94](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/5d38f94) Enable GH actions
- [2a97178](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/2a97178) Replace mysql with mariadb



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.34.0](https://github.com/kubedb/memcached/releases/tag/v0.34.0)

- [c43923ba](https://github.com/kubedb/memcached/commit/c43923ba) Prepare for release v0.34.0 (#423)
- [928f00e2](https://github.com/kubedb/memcached/commit/928f00e2) Prepare for release v0.34.0-rc.1 (#422)
- [42aa5dcc](https://github.com/kubedb/memcached/commit/42aa5dcc) Update deps (#421)
- [d9fd6358](https://github.com/kubedb/memcached/commit/d9fd6358) Update deps (#420)
- [3ae5739b](https://github.com/kubedb/memcached/commit/3ae5739b) Prepare for release v0.34.0-rc.0 (#419)
- [754ba398](https://github.com/kubedb/memcached/commit/754ba398) Prepare for release v0.34.0-beta.1 (#418)
- [abd9dbb6](https://github.com/kubedb/memcached/commit/abd9dbb6) Incorporate with apimachinery package name change from stash to restore (#417)
- [6fe1686a](https://github.com/kubedb/memcached/commit/6fe1686a) Prepare for release v0.34.0-beta.0 (#416)
- [1cfb0544](https://github.com/kubedb/memcached/commit/1cfb0544) Dynamically start crd controller (#415)
- [171faff2](https://github.com/kubedb/memcached/commit/171faff2) Update deps (#414)
- [639495c7](https://github.com/kubedb/memcached/commit/639495c7) Update deps (#413)
- [223d295a](https://github.com/kubedb/memcached/commit/223d295a) Use k8s 1.29 client libs (#412)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.34.0](https://github.com/kubedb/mongodb/releases/tag/v0.34.0)

- [92691f3f](https://github.com/kubedb/mongodb/commit/92691f3f0) Prepare for release v0.34.0 (#611)
- [2e96a733](https://github.com/kubedb/mongodb/commit/2e96a7338) Prepare for release v0.34.0-rc.1 (#610)
- [9a7e16bd](https://github.com/kubedb/mongodb/commit/9a7e16bdb) Update deps (#609)
- [369b51e8](https://github.com/kubedb/mongodb/commit/369b51e8a) Update deps (#608)
- [278ce846](https://github.com/kubedb/mongodb/commit/278ce846b) Prepare for release v0.34.0-rc.0 (#607)
- [c0c58448](https://github.com/kubedb/mongodb/commit/c0c58448b) Prepare for release v0.34.0-beta.1 (#606)
- [5df39d09](https://github.com/kubedb/mongodb/commit/5df39d09f) Update ci mgVersion;  Fix pointer dereference issue (#605)
- [e2781eae](https://github.com/kubedb/mongodb/commit/e2781eaea) Run ci with specific crd-manager branch (#604)
- [b57bc47a](https://github.com/kubedb/mongodb/commit/b57bc47ae) Add kubestash for health check (#603)
- [62cb9c81](https://github.com/kubedb/mongodb/commit/62cb9c816) Install crd-manager specifiying DATABASE (#602)
- [6bf45fe7](https://github.com/kubedb/mongodb/commit/6bf45fe72) 7.0.4 -> 7.0.5; update deps
- [e5b9841e](https://github.com/kubedb/mongodb/commit/e5b9841e5) Fix oplog backup directory (#601)
- [452b785f](https://github.com/kubedb/mongodb/commit/452b785f0) Add  Support for DB phase change for restoring using `KubeStash` (#586)
- [35d93d0b](https://github.com/kubedb/mongodb/commit/35d93d0bc) add ssl/tls args command (#595)
- [7ff67238](https://github.com/kubedb/mongodb/commit/7ff672382) Prepare for release v0.34.0-beta.0 (#600)
- [beca63a4](https://github.com/kubedb/mongodb/commit/beca63a48) Dynamically start crd controller (#599)
- [17d90616](https://github.com/kubedb/mongodb/commit/17d90616d) Update deps (#598)
- [bc25ca00](https://github.com/kubedb/mongodb/commit/bc25ca001) Update deps (#597)
- [4ce5a94a](https://github.com/kubedb/mongodb/commit/4ce5a94a4) Configure openapi for webhook server (#596)
- [8d8206db](https://github.com/kubedb/mongodb/commit/8d8206db3) Update ci versions
- [bfdd519f](https://github.com/kubedb/mongodb/commit/bfdd519fc) Update deps
- [01a7c268](https://github.com/kubedb/mongodb/commit/01a7c2685) Use k8s 1.29 client libs (#594)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.2.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.2.0)

- [06a2057](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/06a2057) Prepare for release v0.2.0 (#15)
- [5b28353](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/5b28353) Prepare for release v0.2.0-rc.1 (#14)
- [afd4fdb](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/afd4fdb) Prepare for release v0.2.0-rc.0 (#13)
- [5680265](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/5680265) Prepare for release v0.2.0-beta.1 (#12)
- [72693c8](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/72693c8) Fix component driver status (#11)
- [0ea73ee](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/0ea73ee) Update deps (#10)
- [ef74421](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/ef74421) Prepare for release v0.2.0-beta.0 (#9)
- [c2c9bd4](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/c2c9bd4) Use k8s 1.29 client libs (#8)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.4.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.4.0)

- [26a2566](https://github.com/kubedb/mongodb-restic-plugin/commit/26a2566) Prepare for release v0.4.0 (#26)
- [3fa6286](https://github.com/kubedb/mongodb-restic-plugin/commit/3fa6286) Prepare for release v0.4.0-rc.1 (#25)
- [bff5aa4](https://github.com/kubedb/mongodb-restic-plugin/commit/bff5aa4) Prepare for release v0.4.0-rc.0 (#24)
- [6ae8ae2](https://github.com/kubedb/mongodb-restic-plugin/commit/6ae8ae2) Prepare for release v0.4.0-beta.1 (#23)
- [d8e1636](https://github.com/kubedb/mongodb-restic-plugin/commit/d8e1636) Reorder the execution of cleanup funcs (#22)
- [4f0b021](https://github.com/kubedb/mongodb-restic-plugin/commit/4f0b021) Prepare for release v0.4.0-beta.0 (#20)
- [91ee7c0](https://github.com/kubedb/mongodb-restic-plugin/commit/91ee7c0) Use k8s 1.29 client libs (#19)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.34.0](https://github.com/kubedb/mysql/releases/tag/v0.34.0)

- [ed8d2fcd](https://github.com/kubedb/mysql/commit/ed8d2fcd8) Prepare for release v0.34.0 (#609)
- [6088a1f9](https://github.com/kubedb/mysql/commit/6088a1f9d) Add azure, gcs support for backup and restore archiver (#595)
- [f3133290](https://github.com/kubedb/mysql/commit/f31332908) Prepare for release v0.34.0-rc.1 (#607)
- [0f3ddf23](https://github.com/kubedb/mysql/commit/0f3ddf233) Update deps (#606)
- [1a5d7bd1](https://github.com/kubedb/mysql/commit/1a5d7bd13) Fix appbinding scheme (#603)
- [c643ded9](https://github.com/kubedb/mysql/commit/c643ded94) Update deps (#605)
- [aaaf3aad](https://github.com/kubedb/mysql/commit/aaaf3aad0) Prepare for release v0.34.0-rc.0 (#604)
- [d2f2eba7](https://github.com/kubedb/mysql/commit/d2f2eba7d) Refactor (#602)
- [fa00fc42](https://github.com/kubedb/mysql/commit/fa00fc424) Fix provider env in sidekick (#601)
- [e75f6e26](https://github.com/kubedb/mysql/commit/e75f6e26e) Fix restore service selector (#600)
- [e9dbf269](https://github.com/kubedb/mysql/commit/e9dbf269c) Prepare for release v0.34.0-beta.1 (#599)
- [44eda2d2](https://github.com/kubedb/mysql/commit/44eda2d25) Prepare for release v0.34.0-beta.1 (#598)
- [16dd4637](https://github.com/kubedb/mysql/commit/16dd46377) Fix pointer dereference issue (#597)
- [334c1a1d](https://github.com/kubedb/mysql/commit/334c1a1dd) Update ci & makefile for crd-manager (#596)
- [edb9b1a1](https://github.com/kubedb/mysql/commit/edb9b1a11) Fix binlog backup directory (#587)
- [fc6d7030](https://github.com/kubedb/mysql/commit/fc6d70303) Add Support for DB phase change for restoring using KubeStash (#594)
- [354f6f3e](https://github.com/kubedb/mysql/commit/354f6f3e1) Prepare for release v0.34.0-beta.0 (#593)
- [01498d02](https://github.com/kubedb/mysql/commit/01498d025) Dynamically start crd controller (#592)
- [e68015cf](https://github.com/kubedb/mysql/commit/e68015cfd) Update deps (#591)
- [67029acc](https://github.com/kubedb/mysql/commit/67029acc9) Update deps (#590)
- [87d2de4a](https://github.com/kubedb/mysql/commit/87d2de4a1) Include kubestash catalog chart in makefile (#588)
- [e5874ffb](https://github.com/kubedb/mysql/commit/e5874ffb7) Add openapi configuration for webhook server (#589)
- [977d3cd3](https://github.com/kubedb/mysql/commit/977d3cd38) Update deps
- [3df86853](https://github.com/kubedb/mysql/commit/3df868533) Use k8s 1.29 client libs (#586)
- [d159ad05](https://github.com/kubedb/mysql/commit/d159ad052) Ensure MySQLArchiver crd (#585)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.2.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.2.0)

- [ed748c7](https://github.com/kubedb/mysql-archiver/commit/ed748c7) Prepare for release v0.2.0 (#20)
- [4d1c341](https://github.com/kubedb/mysql-archiver/commit/4d1c341) Remove example files (#17)
- [9c0545e](https://github.com/kubedb/mysql-archiver/commit/9c0545e) Add azure, gcs support (#13)
- [d0fbef5](https://github.com/kubedb/mysql-archiver/commit/d0fbef5) Prepare for release v0.2.0-rc.1 (#19)
- [a6fdf50](https://github.com/kubedb/mysql-archiver/commit/a6fdf50) Prepare for release v0.2.0-rc.0 (#18)
- [718511e](https://github.com/kubedb/mysql-archiver/commit/718511e) Remove obsolete files (#16)
- [07fc1eb](https://github.com/kubedb/mysql-archiver/commit/07fc1eb) Fix mysql-community-common version in docker file
- [e5bdae3](https://github.com/kubedb/mysql-archiver/commit/e5bdae3) Prepare for release v0.2.0-beta.1 (#15)
- [7ef752c](https://github.com/kubedb/mysql-archiver/commit/7ef752c) Refactor + Cleanup wal-g example files (#14)
- [5857a8d](https://github.com/kubedb/mysql-archiver/commit/5857a8d) Don't use fail-fast
- [5833776](https://github.com/kubedb/mysql-archiver/commit/5833776) Prepare for release v0.2.0-beta.0 (#12)
- [f3e68b2](https://github.com/kubedb/mysql-archiver/commit/f3e68b2) Use k8s 1.29 client libs (#11)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.19.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.19.0)

- [ef84bd37](https://github.com/kubedb/mysql-coordinator/commit/ef84bd37) Prepare for release v0.19.0 (#103)
- [40a2b6eb](https://github.com/kubedb/mysql-coordinator/commit/40a2b6eb) Prepare for release v0.19.0-rc.1 (#102)
- [8b80958d](https://github.com/kubedb/mysql-coordinator/commit/8b80958d) Update deps (#101)
- [df438239](https://github.com/kubedb/mysql-coordinator/commit/df438239) Update deps (#100)
- [1bc71d04](https://github.com/kubedb/mysql-coordinator/commit/1bc71d04) Prepare for release v0.19.0-rc.0 (#99)
- [59a11671](https://github.com/kubedb/mysql-coordinator/commit/59a11671) Prepare for release v0.19.0-beta.1 (#98)
- [e0cc149f](https://github.com/kubedb/mysql-coordinator/commit/e0cc149f) Prepare for release v0.19.0-beta.0 (#97)
- [67aeb229](https://github.com/kubedb/mysql-coordinator/commit/67aeb229) Update deps (#96)
- [2fa4423f](https://github.com/kubedb/mysql-coordinator/commit/2fa4423f) Update deps (#95)
- [b0735769](https://github.com/kubedb/mysql-coordinator/commit/b0735769) Use k8s 1.29 client libs (#94)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.2.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.2.0)

- [61d40c2](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/61d40c2) Prepare for release v0.2.0 (#8)
- [ac60bf4](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/ac60bf4) Prepare for release v0.2.0-rc.1 (#7)
- [21e9470](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/21e9470) Prepare for release v0.2.0-rc.0 (#6)
- [d5771cf](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/d5771cf) Prepare for release v0.2.0-beta.1 (#5)
- [b4ffc6f](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/b4ffc6f) Fix component driver status & Update deps (#3)
- [d285eff](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/d285eff) Prepare for release v0.2.0-beta.0 (#4)
- [7a46441](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/7a46441) Use k8s 1.29 client libs (#2)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.4.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.4.0)

- [416d3cd](https://github.com/kubedb/mysql-restic-plugin/commit/416d3cd) Prepare for release v0.4.0 (#24)
- [55897ab](https://github.com/kubedb/mysql-restic-plugin/commit/55897ab) Prepare for release v0.4.0-rc.1 (#23)
- [eedf2e7](https://github.com/kubedb/mysql-restic-plugin/commit/eedf2e7) Prepare for release v0.4.0-rc.0 (#22)
- [105888a](https://github.com/kubedb/mysql-restic-plugin/commit/105888a) Prepare for release v0.4.0-beta.1 (#21)
- [b42d0cf](https://github.com/kubedb/mysql-restic-plugin/commit/b42d0cf) Removed `--all-databases` flag for restoring (#20)
- [742d2ce](https://github.com/kubedb/mysql-restic-plugin/commit/742d2ce) Prepare for release v0.4.0-beta.0 (#19)
- [0402847](https://github.com/kubedb/mysql-restic-plugin/commit/0402847) Use k8s 1.29 client libs (#18)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.19.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.19.0)

- [6a5deed](https://github.com/kubedb/mysql-router-init/commit/6a5deed) Update deps (#40)
- [0078f09](https://github.com/kubedb/mysql-router-init/commit/0078f09) Update deps (#39)
- [85f8c6f](https://github.com/kubedb/mysql-router-init/commit/85f8c6f) Update deps (#38)
- [7dd201c](https://github.com/kubedb/mysql-router-init/commit/7dd201c) Use k8s 1.29 client libs (#37)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.28.0](https://github.com/kubedb/ops-manager/releases/tag/v0.28.0)




## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.28.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.28.0)

- [279330e0](https://github.com/kubedb/percona-xtradb/commit/279330e09) Prepare for release v0.28.0 (#354)
- [b567db53](https://github.com/kubedb/percona-xtradb/commit/b567db53a) Prepare for release v0.28.0-rc.1 (#353)
- [c0ddb330](https://github.com/kubedb/percona-xtradb/commit/c0ddb330b) Update deps (#352)
- [d461df3e](https://github.com/kubedb/percona-xtradb/commit/d461df3ed) Fix appbinding scheme (#349)
- [2752f7e3](https://github.com/kubedb/percona-xtradb/commit/2752f7e36) Update deps (#351)
- [80cd3a03](https://github.com/kubedb/percona-xtradb/commit/80cd3a030) Prepare for release v0.28.0-rc.0 (#350)
- [475a5e32](https://github.com/kubedb/percona-xtradb/commit/475a5e328) Prepare for release v0.28.0-beta.1 (#348)
- [4c1380ab](https://github.com/kubedb/percona-xtradb/commit/4c1380ab7) Incorporate with apimachinery package name change from `stash` to `restore` (#347)
- [0ceb3028](https://github.com/kubedb/percona-xtradb/commit/0ceb30284) Prepare for release v0.28.0-beta.0 (#346)
- [e7d35606](https://github.com/kubedb/percona-xtradb/commit/e7d356062) Dynamically start crd controller (#345)
- [5d07b565](https://github.com/kubedb/percona-xtradb/commit/5d07b5655) Update deps (#344)
- [1a639f84](https://github.com/kubedb/percona-xtradb/commit/1a639f840) Update deps (#343)
- [4f8b24ab](https://github.com/kubedb/percona-xtradb/commit/4f8b24aba) Update deps
- [e5254020](https://github.com/kubedb/percona-xtradb/commit/e52540202) Use k8s 1.29 client libs (#341)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.14.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.14.0)

- [6fd3b3cd](https://github.com/kubedb/percona-xtradb-coordinator/commit/6fd3b3cd) Prepare for release v0.14.0 (#63)
- [da619fa3](https://github.com/kubedb/percona-xtradb-coordinator/commit/da619fa3) Prepare for release v0.14.0-rc.1 (#62)
- [c39daf56](https://github.com/kubedb/percona-xtradb-coordinator/commit/c39daf56) Update deps (#61)
- [42dc1a95](https://github.com/kubedb/percona-xtradb-coordinator/commit/42dc1a95) Update deps (#60)
- [7581630e](https://github.com/kubedb/percona-xtradb-coordinator/commit/7581630e) Prepare for release v0.14.0-rc.0 (#59)
- [560bc5c3](https://github.com/kubedb/percona-xtradb-coordinator/commit/560bc5c3) Prepare for release v0.14.0-beta.1 (#58)
- [963756eb](https://github.com/kubedb/percona-xtradb-coordinator/commit/963756eb) Prepare for release v0.14.0-beta.0 (#57)
- [5489bb8c](https://github.com/kubedb/percona-xtradb-coordinator/commit/5489bb8c) Update deps (#56)
- [a8424e18](https://github.com/kubedb/percona-xtradb-coordinator/commit/a8424e18) Update deps (#55)
- [ee4add86](https://github.com/kubedb/percona-xtradb-coordinator/commit/ee4add86) Use k8s 1.29 client libs (#54)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.25.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.25.0)

- [7e7d32d9](https://github.com/kubedb/pg-coordinator/commit/7e7d32d9) Prepare for release v0.25.0 (#154)
- [9a720273](https://github.com/kubedb/pg-coordinator/commit/9a720273) Prepare for release v0.25.0-rc.1 (#153)
- [f103f1fc](https://github.com/kubedb/pg-coordinator/commit/f103f1fc) Update deps (#152)
- [84b92d89](https://github.com/kubedb/pg-coordinator/commit/84b92d89) Update deps (#151)
- [41cc97b6](https://github.com/kubedb/pg-coordinator/commit/41cc97b6) Prepare for release v0.25.0-rc.0 (#150)
- [5298a177](https://github.com/kubedb/pg-coordinator/commit/5298a177) Fixed (#149)
- [bc296307](https://github.com/kubedb/pg-coordinator/commit/bc296307) Prepare for release v0.25.0-beta.1 (#148)
- [30973540](https://github.com/kubedb/pg-coordinator/commit/30973540) Prepare for release v0.25.0-beta.0 (#147)
- [7b84e198](https://github.com/kubedb/pg-coordinator/commit/7b84e198) Update deps (#146)
- [f1bfe818](https://github.com/kubedb/pg-coordinator/commit/f1bfe818) Update deps (#145)
- [1de05a6e](https://github.com/kubedb/pg-coordinator/commit/1de05a6e) Use k8s 1.29 client libs (#144)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.28.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.28.0)

- [f79abcdd](https://github.com/kubedb/pgbouncer/commit/f79abcdd) Prepare for release v0.28.0 (#317)
- [0ca01e53](https://github.com/kubedb/pgbouncer/commit/0ca01e53) Prepare for release v0.28.0-rc.1 (#316)
- [4b76d2cb](https://github.com/kubedb/pgbouncer/commit/4b76d2cb) Update deps (#315)
- [f32676bc](https://github.com/kubedb/pgbouncer/commit/f32676bc) Update deps (#314)
- [e69aa743](https://github.com/kubedb/pgbouncer/commit/e69aa743) Prepare for release v0.28.0-rc.0 (#313)
- [55c248d5](https://github.com/kubedb/pgbouncer/commit/55c248d5) Prepare for release v0.28.0-beta.1 (#312)
- [1b86664a](https://github.com/kubedb/pgbouncer/commit/1b86664a) Incorporate with apimachinery package name change from stash to restore (#311)
- [3c6bc335](https://github.com/kubedb/pgbouncer/commit/3c6bc335) Prepare for release v0.28.0-beta.0 (#310)
- [73c5f6fb](https://github.com/kubedb/pgbouncer/commit/73c5f6fb) Dynamically start crd controller (#309)
- [f9edc2cd](https://github.com/kubedb/pgbouncer/commit/f9edc2cd) Update deps (#308)
- [d54251c0](https://github.com/kubedb/pgbouncer/commit/d54251c0) Update deps (#307)
- [de40a35e](https://github.com/kubedb/pgbouncer/commit/de40a35e) Update deps
- [8c325577](https://github.com/kubedb/pgbouncer/commit/8c325577) Use k8s 1.29 client libs (#305)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.0.4](https://github.com/kubedb/pgpool/releases/tag/v0.0.4)

- [b6546d3](https://github.com/kubedb/pgpool/commit/b6546d3) Prepare for release v0.0.4 (#12)
- [6f7ebca](https://github.com/kubedb/pgpool/commit/6f7ebca) Add daily to workflows (#10)
- [18c06cd](https://github.com/kubedb/pgpool/commit/18c06cd) Fix InitConfiguration issue (#9)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.41.0](https://github.com/kubedb/postgres/releases/tag/v0.41.0)

- [cf1b2726](https://github.com/kubedb/postgres/commit/cf1b27268) Prepare for release v0.41.0 (#715)
- [29aaa191](https://github.com/kubedb/postgres/commit/29aaa191b) Add sub path for sidekick (#714)
- [9f487d98](https://github.com/kubedb/postgres/commit/9f487d984) Add postmaster arguments for exporter image (#713)
- [071b2645](https://github.com/kubedb/postgres/commit/071b26455) Prepare for release v0.41.0-rc.1 (#712)
- [c9f1d5b6](https://github.com/kubedb/postgres/commit/c9f1d5b68) Update deps (#711)
- [723fb80c](https://github.com/kubedb/postgres/commit/723fb80c0) Update deps (#710)
- [8135d351](https://github.com/kubedb/postgres/commit/8135d3511) Prepare for release v0.41.0-rc.0 (#709)
- [72a1ee29](https://github.com/kubedb/postgres/commit/72a1ee294) Prepare for release v0.41.0-beta.1 (#708)
- [026598f4](https://github.com/kubedb/postgres/commit/026598f44) Prepare for release v0.41.0-beta.1 (#707)
- [8af305aa](https://github.com/kubedb/postgres/commit/8af305aa4) Use ptr.Deref(); Update deps
- [c7c0652d](https://github.com/kubedb/postgres/commit/c7c0652dc) Update ci & makefile for crd-manager (#706)
- [d468bdb3](https://github.com/kubedb/postgres/commit/d468bdb34) Fix wal backup directory (#705)
- [c6992bed](https://github.com/kubedb/postgres/commit/c6992bed8) Add Support for DB phase change for restoring using KubeStash (#704)
- [d1bd909b](https://github.com/kubedb/postgres/commit/d1bd909ba) Prepare for release v0.41.0-beta.0 (#703)
- [5e8101e3](https://github.com/kubedb/postgres/commit/5e8101e39) Dynamically start crd controller (#702)
- [47dbbff5](https://github.com/kubedb/postgres/commit/47dbbff53) Update deps (#701)
- [84f99c58](https://github.com/kubedb/postgres/commit/84f99c58b) Disable fairness api
- [a715765d](https://github.com/kubedb/postgres/commit/a715765dc) Set --restricted=false for ci tests (#700)
- [fe9af597](https://github.com/kubedb/postgres/commit/fe9af5977) Add Postgres test fix (#699)
- [8bae8886](https://github.com/kubedb/postgres/commit/8bae88860) Configure openapi for webhook server (#698)
- [9ce2efce](https://github.com/kubedb/postgres/commit/9ce2efce5) Update deps
- [24e4e9ca](https://github.com/kubedb/postgres/commit/24e4e9ca5) Use k8s 1.29 client libs (#697)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.2.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.2.0)

- [a55baa8](https://github.com/kubedb/postgres-archiver/commit/a55baa8) Prepare for release v0.2.0 (#21)
- [8f66790](https://github.com/kubedb/postgres-archiver/commit/8f66790) Prepare for release v0.2.0-rc.1 (#20)
- [bff75cb](https://github.com/kubedb/postgres-archiver/commit/bff75cb) Prepare for release v0.2.0-rc.0 (#19)
- [bb8c342](https://github.com/kubedb/postgres-archiver/commit/bb8c342) Create directory for wal-backup (#18)
- [c4405c1](https://github.com/kubedb/postgres-archiver/commit/c4405c1) Prepare for release v0.2.0-beta.1 (#17)
- [c353dcd](https://github.com/kubedb/postgres-archiver/commit/c353dcd) Don't use fail-fast
- [a9cbe08](https://github.com/kubedb/postgres-archiver/commit/a9cbe08) Prepare for release v0.2.0-beta.0 (#16)
- [183e97c](https://github.com/kubedb/postgres-archiver/commit/183e97c) Use k8s 1.29 client libs (#15)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.2.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.2.0)

- [5e0031f](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/5e0031f) Prepare for release v0.2.0 (#18)
- [369c9a3](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/369c9a3) Prepare for release v0.2.0-rc.1 (#17)
- [87240d8](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/87240d8) Prepare for release v0.2.0-rc.0 (#16)
- [dc4f85e](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/dc4f85e) Prepare for release v0.2.0-beta.1 (#15)
- [098365a](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/098365a) Update README.md (#14)
- [5ef571f](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/5ef571f) Update deps (#13)
- [f0e546a](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/f0e546a) Prepare for release v0.2.0-beta.0 (#12)
- [aae7294](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/aae7294) Use k8s 1.29 client libs (#11)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.4.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.4.0)




## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.3.0](https://github.com/kubedb/provider-aws/releases/tag/v0.3.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.3.0](https://github.com/kubedb/provider-azure/releases/tag/v0.3.0)

- [ebba4fa](https://github.com/kubedb/provider-azure/commit/ebba4fa) Checkout fake release branch for release workflow



## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.3.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.3.0)

- [82f52c3](https://github.com/kubedb/provider-gcp/commit/82f52c3) Checkout fake release branch for release workflow



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.41.0](https://github.com/kubedb/provisioner/releases/tag/v0.41.0)




## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.28.0](https://github.com/kubedb/proxysql/releases/tag/v0.28.0)

- [b0d0e92c](https://github.com/kubedb/proxysql/commit/b0d0e92cf) Prepare for release v0.28.0 (#335)
- [0ff6f90d](https://github.com/kubedb/proxysql/commit/0ff6f90d2) Prepare for release v0.28.0-rc.1 (#334)
- [382d3283](https://github.com/kubedb/proxysql/commit/382d3283e) Update deps (#333)
- [0b4da810](https://github.com/kubedb/proxysql/commit/0b4da8101) Update deps (#332)
- [2fa5679d](https://github.com/kubedb/proxysql/commit/2fa5679d7) Prepare for release v0.28.0-rc.0 (#331)
- [2cc59016](https://github.com/kubedb/proxysql/commit/2cc590165) Update ci & makefile for crd-manager (#326)
- [79e29efd](https://github.com/kubedb/proxysql/commit/79e29efdb) Handle MySQL URL Parsing (#330)
- [b3372a53](https://github.com/kubedb/proxysql/commit/b3372a53d) Fix MySQL Client and sync_user (#328)
- [213ebfc4](https://github.com/kubedb/proxysql/commit/213ebfc43) Prepare for release v0.28.0-beta.1 (#327)
- [8427158e](https://github.com/kubedb/proxysql/commit/8427158ec) Incorporate with apimachinery package name change from stash to restore (#325)
- [c0805050](https://github.com/kubedb/proxysql/commit/c0805050e) Prepare for release v0.28.0-beta.0 (#324)
- [88ef1f1d](https://github.com/kubedb/proxysql/commit/88ef1f1de) Dynamically start crd controller (#323)
- [8c0a96ac](https://github.com/kubedb/proxysql/commit/8c0a96ac7) Update deps (#322)
- [e96797e4](https://github.com/kubedb/proxysql/commit/e96797e48) Update deps (#321)
- [e8fd529b](https://github.com/kubedb/proxysql/commit/e8fd529b2) Update deps
- [b2e9a1df](https://github.com/kubedb/proxysql/commit/b2e9a1df8) Use k8s 1.29 client libs (#319)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.0.4](https://github.com/kubedb/rabbitmq/releases/tag/v0.0.4)

- [cbcc9132](https://github.com/kubedb/rabbitmq/commit/cbcc9132) Prepare for release v0.0.4 (#9)
- [89636ce7](https://github.com/kubedb/rabbitmq/commit/89636ce7) Remove standby service and fix init container security context (#8)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.34.0](https://github.com/kubedb/redis/releases/tag/v0.34.0)

- [bd9d152b](https://github.com/kubedb/redis/commit/bd9d152b) Prepare for release v0.34.0 (#523)
- [5e171587](https://github.com/kubedb/redis/commit/5e171587) Prepare for release v0.34.0-rc.1 (#522)
- [71665b9b](https://github.com/kubedb/redis/commit/71665b9b) Update deps (#521)
- [302f1f19](https://github.com/kubedb/redis/commit/302f1f19) Update deps (#520)
- [0703a513](https://github.com/kubedb/redis/commit/0703a513) Prepare for release v0.34.0-rc.0 (#519)
- [b1a296b7](https://github.com/kubedb/redis/commit/b1a296b7) Init sentinel before secret watcher (#518)
- [01290634](https://github.com/kubedb/redis/commit/01290634) Prepare for release v0.34.0-beta.1 (#517)
- [e51f93e1](https://github.com/kubedb/redis/commit/e51f93e1) Fix panic (#516)
- [dc75c163](https://github.com/kubedb/redis/commit/dc75c163) Update ci & makefile for crd-manager (#515)
- [09688f35](https://github.com/kubedb/redis/commit/09688f35) Add Support for DB phase change for restoring using KubeStash (#514)
- [7e844ab1](https://github.com/kubedb/redis/commit/7e844ab1) Prepare for release v0.34.0-beta.0 (#513)
- [6318d04f](https://github.com/kubedb/redis/commit/6318d04f) Dynamically start crd controller (#512)
- [92b8a3a9](https://github.com/kubedb/redis/commit/92b8a3a9) Update deps (#511)
- [f0fb4c69](https://github.com/kubedb/redis/commit/f0fb4c69) Update deps (#510)
- [c99d9498](https://github.com/kubedb/redis/commit/c99d9498) Update deps
- [90299544](https://github.com/kubedb/redis/commit/90299544) Use k8s 1.29 client libs (#508)
- [fced7010](https://github.com/kubedb/redis/commit/fced7010) Update redis versions in nightly tests (#507)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.20.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.20.0)

- [cd9b64c3](https://github.com/kubedb/redis-coordinator/commit/cd9b64c3) Prepare for release v0.20.0 (#94)
- [055ceaf1](https://github.com/kubedb/redis-coordinator/commit/055ceaf1) Prepare for release v0.20.0-rc.1 (#93)
- [79575d26](https://github.com/kubedb/redis-coordinator/commit/79575d26) Update deps (#92)
- [a5b4c4b4](https://github.com/kubedb/redis-coordinator/commit/a5b4c4b4) Update deps (#91)
- [f09062c4](https://github.com/kubedb/redis-coordinator/commit/f09062c4) Prepare for release v0.20.0-rc.0 (#90)
- [fd3b2112](https://github.com/kubedb/redis-coordinator/commit/fd3b2112) Prepare for release v0.20.0-beta.1 (#89)
- [4c36accd](https://github.com/kubedb/redis-coordinator/commit/4c36accd) Prepare for release v0.20.0-beta.0 (#88)
- [c8658380](https://github.com/kubedb/redis-coordinator/commit/c8658380) Update deps (#87)
- [c99c2e9b](https://github.com/kubedb/redis-coordinator/commit/c99c2e9b) Update deps (#86)
- [22c7beb4](https://github.com/kubedb/redis-coordinator/commit/22c7beb4) Use k8s 1.29 client libs (#85)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.4.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.4.0)

- [5436cb6](https://github.com/kubedb/redis-restic-plugin/commit/5436cb6) Prepare for release v0.4.0 (#20)
- [67a8942](https://github.com/kubedb/redis-restic-plugin/commit/67a8942) Prepare for release v0.4.0-rc.1 (#19)
- [968da13](https://github.com/kubedb/redis-restic-plugin/commit/968da13) Prepare for release v0.4.0-rc.0 (#18)
- [fac6226](https://github.com/kubedb/redis-restic-plugin/commit/fac6226) Prepare for release v0.4.0-beta.1 (#17)
- [da2796a](https://github.com/kubedb/redis-restic-plugin/commit/da2796a) Prepare for release v0.4.0-beta.0 (#16)
- [0553c6f](https://github.com/kubedb/redis-restic-plugin/commit/0553c6f) Use k8s 1.29 client libs (#15)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.28.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.28.0)

- [4cce748f](https://github.com/kubedb/replication-mode-detector/commit/4cce748f) Prepare for release v0.28.0 (#258)
- [de39974e](https://github.com/kubedb/replication-mode-detector/commit/de39974e) Prepare for release v0.28.0-rc.1 (#257)
- [e1ef5191](https://github.com/kubedb/replication-mode-detector/commit/e1ef5191) Update deps (#256)
- [7b4e4149](https://github.com/kubedb/replication-mode-detector/commit/7b4e4149) Update deps (#255)
- [d55f7e69](https://github.com/kubedb/replication-mode-detector/commit/d55f7e69) Prepare for release v0.28.0-rc.0 (#254)
- [f948a650](https://github.com/kubedb/replication-mode-detector/commit/f948a650) Prepare for release v0.28.0-beta.1 (#253)
- [572668c8](https://github.com/kubedb/replication-mode-detector/commit/572668c8) Prepare for release v0.28.0-beta.0 (#252)
- [39ba3ce0](https://github.com/kubedb/replication-mode-detector/commit/39ba3ce0) Update deps (#251)
- [d3d2ad96](https://github.com/kubedb/replication-mode-detector/commit/d3d2ad96) Update deps (#250)
- [633d7b76](https://github.com/kubedb/replication-mode-detector/commit/633d7b76) Use k8s 1.29 client libs (#249)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.17.0](https://github.com/kubedb/schema-manager/releases/tag/v0.17.0)




## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.0.4](https://github.com/kubedb/singlestore/releases/tag/v0.0.4)

- [e8cf66f](https://github.com/kubedb/singlestore/commit/e8cf66f) Prepare for release v0.0.4 (#11)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.0.4](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.0.4)

- [b451944](https://github.com/kubedb/singlestore-coordinator/commit/b451944) Prepare for release v0.0.4 (#5)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.0.4](https://github.com/kubedb/solr/releases/tag/v0.0.4)

- [8be74c4](https://github.com/kubedb/solr/commit/8be74c4) Prepare for release v0.0.4 (#12)
- [0647ccf](https://github.com/kubedb/solr/commit/0647ccf) Remove overseer discovery service. (#10)
- [4901117](https://github.com/kubedb/solr/commit/4901117) Add daily yml. (#9)
- [3abb79b](https://github.com/kubedb/solr/commit/3abb79b) Add auth secret reference in appbinding. (#8)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.26.0](https://github.com/kubedb/tests/releases/tag/v0.26.0)

- [16543a0f](https://github.com/kubedb/tests/commit/16543a0f) Prepare for release v0.26.0 (#304)
- [92607278](https://github.com/kubedb/tests/commit/92607278) Add dependencies flag (#301)
- [17bbf43c](https://github.com/kubedb/tests/commit/17bbf43c) Add Reconfigure with Vertical Scaling (#300)
- [f3e3fba1](https://github.com/kubedb/tests/commit/f3e3fba1) Add Singlestore Provisioning Test (#287)
- [5a527051](https://github.com/kubedb/tests/commit/5a527051) Prepare for release v0.26.0-rc.1 (#299)
- [03d71b6d](https://github.com/kubedb/tests/commit/03d71b6d) Update deps (#298)
- [2d928008](https://github.com/kubedb/tests/commit/2d928008) Update deps (#297)
- [1730fd31](https://github.com/kubedb/tests/commit/1730fd31) Prepare for release v0.26.0-rc.0 (#296)
- [d1805668](https://github.com/kubedb/tests/commit/d1805668) Add ZooKeeper Tests (#294)
- [4c27754c](https://github.com/kubedb/tests/commit/4c27754c) Fix kafka env-variable tests (#293)
- [3cfc1212](https://github.com/kubedb/tests/commit/3cfc1212) Prepare for release v0.26.0-beta.1 (#292)
- [b810e690](https://github.com/kubedb/tests/commit/b810e690) increase cpu limit for vertical scaling (#289)
- [c43985ba](https://github.com/kubedb/tests/commit/c43985ba) Change dashboard api group (#291)
- [1b96881e](https://github.com/kubedb/tests/commit/1b96881e) Fix error logging
- [33f78143](https://github.com/kubedb/tests/commit/33f78143) forceCleanup PVCs for mongo (#288)
- [0dcd3e38](https://github.com/kubedb/tests/commit/0dcd3e38) Add PostgreSQL logical replication tests  (#202)
- [2f403c85](https://github.com/kubedb/tests/commit/2f403c85) Find profiles in array, Don't match with string (#286)
- [5aca2293](https://github.com/kubedb/tests/commit/5aca2293) Give time to PDB status to be updated (#285)
- [5f3fabd7](https://github.com/kubedb/tests/commit/5f3fabd7) Prepare for release v0.26.0-beta.0 (#284)
- [27a24dff](https://github.com/kubedb/tests/commit/27a24dff) Update deps (#283)
- [b9021186](https://github.com/kubedb/tests/commit/b9021186) Update deps (#282)
- [589ca51c](https://github.com/kubedb/tests/commit/589ca51c) mongodb vertical scaling fix (#281)
- [feaa0f6a](https://github.com/kubedb/tests/commit/feaa0f6a) Add `--restricted` flag (#280)
- [2423ee38](https://github.com/kubedb/tests/commit/2423ee38) Fix linter errors
- [dcd64c7c](https://github.com/kubedb/tests/commit/dcd64c7c) Update lint command
- [c3ef1fa4](https://github.com/kubedb/tests/commit/c3ef1fa4) Use k8s 1.29 client libs (#279)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.17.0](https://github.com/kubedb/ui-server/releases/tag/v0.17.0)

- [7a1d7c5e](https://github.com/kubedb/ui-server/commit/7a1d7c5e) Prepare for release v0.17.0 (#110)
- [ed2c04e7](https://github.com/kubedb/ui-server/commit/ed2c04e7) Prepare for release v0.17.0-rc.1 (#109)
- [645c4ac2](https://github.com/kubedb/ui-server/commit/645c4ac2) Update deps (#108)
- [e75f0f9e](https://github.com/kubedb/ui-server/commit/e75f0f9e) Update deps (#107)
- [3046f685](https://github.com/kubedb/ui-server/commit/3046f685) Prepare for release v0.17.0-rc.0 (#106)
- [98c1a6dd](https://github.com/kubedb/ui-server/commit/98c1a6dd) Prepare for release v0.17.0-beta.1 (#105)
- [8173cfc2](https://github.com/kubedb/ui-server/commit/8173cfc2) Implement SingularNameProvider
- [6e8f80dc](https://github.com/kubedb/ui-server/commit/6e8f80dc) Prepare for release v0.17.0-beta.0 (#104)
- [6a05721f](https://github.com/kubedb/ui-server/commit/6a05721f) Update deps (#103)
- [3c24fd5e](https://github.com/kubedb/ui-server/commit/3c24fd5e) Update deps (#102)
- [25e29443](https://github.com/kubedb/ui-server/commit/25e29443) Use k8s 1.29 client libs (#101)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.17.0](https://github.com/kubedb/webhook-server/releases/tag/v0.17.0)

- [93116fb5](https://github.com/kubedb/webhook-server/commit/93116fb5) Prepare for release v0.17.0 (#95)
- [a49ecca7](https://github.com/kubedb/webhook-server/commit/a49ecca7) Prepare for release v0.17.0-rc.1 (#94)
- [5f8de57b](https://github.com/kubedb/webhook-server/commit/5f8de57b) Update deps (#93)
- [8c22ce2d](https://github.com/kubedb/webhook-server/commit/8c22ce2d) Update deps (#92)
- [f9cf0b11](https://github.com/kubedb/webhook-server/commit/f9cf0b11) Prepare for release v0.17.0-rc.0 (#91)
- [98914ade](https://github.com/kubedb/webhook-server/commit/98914ade) Add kafka connector webhook apitypes (#90)
- [1184db7a](https://github.com/kubedb/webhook-server/commit/1184db7a) Fix solr webhook
- [2a84cedb](https://github.com/kubedb/webhook-server/commit/2a84cedb) Prepare for release v0.17.0-beta.1 (#89)
- [bb4a5c22](https://github.com/kubedb/webhook-server/commit/bb4a5c22) Add kafka connect-cluster (#87)
- [c46c6662](https://github.com/kubedb/webhook-server/commit/c46c6662) Add new Database support (#88)
- [c6387e9e](https://github.com/kubedb/webhook-server/commit/c6387e9e) Set default kubebuilder client for autoscaler (#86)
- [14c07899](https://github.com/kubedb/webhook-server/commit/14c07899) Incorporate apimachinery (#85)
- [266c79a0](https://github.com/kubedb/webhook-server/commit/266c79a0) Add kafka ops request validator (#84)
- [528b8463](https://github.com/kubedb/webhook-server/commit/528b8463) Fix webhook handlers (#83)
- [dfdeb6c3](https://github.com/kubedb/webhook-server/commit/dfdeb6c3) Prepare for release v0.17.0-beta.0 (#82)
- [bf54df2a](https://github.com/kubedb/webhook-server/commit/bf54df2a) Update deps (#81)
- [c7d17faa](https://github.com/kubedb/webhook-server/commit/c7d17faa) Update deps (#79)
- [170573b1](https://github.com/kubedb/webhook-server/commit/170573b1) Use k8s 1.29 client libs (#78)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.0.4](https://github.com/kubedb/zookeeper/releases/tag/v0.0.4)

- [7347527](https://github.com/kubedb/zookeeper/commit/7347527) Prepare for release v0.0.4 (#8)




