---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2024.1.19-beta.1
    name: Changelog-v2024.1.19-beta.1
    parent: welcome
    weight: 20240119
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2024.1.19-beta.1/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2024.1.19-beta.1/
---

# KubeDB v2024.1.19-beta.1 (2024-01-20)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.41.0-beta.1](https://github.com/kubedb/apimachinery/releases/tag/v0.41.0-beta.1)

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



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.26.0-beta.1](https://github.com/kubedb/autoscaler/releases/tag/v0.26.0-beta.1)

- [7cef99b3](https://github.com/kubedb/autoscaler/commit/7cef99b3) Prepare for release v0.26.0-beta.1 (#181)
- [621bf52c](https://github.com/kubedb/autoscaler/commit/621bf52c) Use RestMapper to check for crd availability (#180)
- [2ae4e01e](https://github.com/kubedb/autoscaler/commit/2ae4e01e) Initialize kubeuilder client for webhooks; cleanup (#179)
- [e536b856](https://github.com/kubedb/autoscaler/commit/e536b856) Conditionally check for vpa & checkpoints (#178)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.41.0-beta.1](https://github.com/kubedb/cli/releases/tag/v0.41.0-beta.1)

- [234b7051](https://github.com/kubedb/cli/commit/234b7051) Prepare for release v0.41.0-beta.1 (#748)
- [1ebdd532](https://github.com/kubedb/cli/commit/1ebdd532) Update deps



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.17.0-beta.1](https://github.com/kubedb/dashboard/releases/tag/v0.17.0-beta.1)

- [999f215f](https://github.com/kubedb/dashboard/commit/999f215f) Prepare for release v0.17.0-beta.1 (#100)
- [80780e17](https://github.com/kubedb/dashboard/commit/80780e17) Change dashboard api group to elasticsearch (#99)
- [b362ecb6](https://github.com/kubedb/dashboard/commit/b362ecb6) Use Go client from db-client-go lib (#98)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.0.1](https://github.com/kubedb/druid/releases/tag/v0.0.1)

- [46c4387](https://github.com/kubedb/druid/commit/46c4387) Prepare for release v0.0.1 (#2)
- [3a9e0dd](https://github.com/kubedb/druid/commit/3a9e0dd) Add Druid Controller (#1)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.41.0-beta.1](https://github.com/kubedb/elasticsearch/releases/tag/v0.41.0-beta.1)

- [c410b39f](https://github.com/kubedb/elasticsearch/commit/c410b39f5) Prepare for release v0.41.0-beta.1 (#699)
- [3394f1d1](https://github.com/kubedb/elasticsearch/commit/3394f1d13) Use ptr.Deref(); Update deps
- [f00ee052](https://github.com/kubedb/elasticsearch/commit/f00ee052e) Update ci & makefile for crd-manager (#698)
- [e37e6d63](https://github.com/kubedb/elasticsearch/commit/e37e6d631) Add catalog client in scheme. (#697)
- [a46bfd41](https://github.com/kubedb/elasticsearch/commit/a46bfd41b) Add Support for DB phase change for restoring using KubeStash (#696)
- [9cbac2fc](https://github.com/kubedb/elasticsearch/commit/9cbac2fc4) Update makefile for dynamic crd installer (#695)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.4.0-beta.1](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.4.0-beta.1)

- [584dfd9](https://github.com/kubedb/elasticsearch-restic-plugin/commit/584dfd9) Prepare for release v0.4.0-beta.1 (#16)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.0.1](https://github.com/kubedb/ferretdb/releases/tag/v0.0.1)

- [68618ec](https://github.com/kubedb/ferretdb/commit/68618ec) Prepare for release v0.0.1 (#4)
- [9443437](https://github.com/kubedb/ferretdb/commit/9443437) Add github workflow files (#3)
- [0287771](https://github.com/kubedb/ferretdb/commit/0287771) Add FerretDB Controller (#2)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2024.1.19-beta.1](https://github.com/kubedb/installer/releases/tag/v2024.1.19-beta.1)

- [a58a71f1](https://github.com/kubedb/installer/commit/a58a71f1) Prepare for release v2024.1.19-beta.1 (#813)
- [fad71f4d](https://github.com/kubedb/installer/commit/fad71f4d) Use appscode built opensearch images (#798)
- [016898c4](https://github.com/kubedb/installer/commit/016898c4) Update webhook values for solr. (#812)
- [238d29a9](https://github.com/kubedb/installer/commit/238d29a9) Add necessary cluster-role for kubestash (#811)
- [e476675f](https://github.com/kubedb/installer/commit/e476675f) Add Druid (#807)
- [ba594a40](https://github.com/kubedb/installer/commit/ba594a40) Add Pgpool (#809)
- [2fb21fa0](https://github.com/kubedb/installer/commit/2fb21fa0) Add ferretdb (#806)
- [1f285c40](https://github.com/kubedb/installer/commit/1f285c40) Update solr version crds. (#808)
- [588e078f](https://github.com/kubedb/installer/commit/588e078f) Revert "Update Solr webhook helm charts. (#796)"
- [f4db4314](https://github.com/kubedb/installer/commit/f4db4314) Update Solr webhook helm charts. (#796)
- [a33a050d](https://github.com/kubedb/installer/commit/a33a050d) Add Redis version 7.2.4 and 7.0.15 (#797)
- [9074c79c](https://github.com/kubedb/installer/commit/9074c79c) Add Singlestore  (#782)
- [eec84b67](https://github.com/kubedb/installer/commit/eec84b67) Add rabbitmq crd (#785)
- [06495dbc](https://github.com/kubedb/installer/commit/06495dbc) Update crds for kubedb/apimachinery@0f8ac911 (#805)
- [bb4786ea](https://github.com/kubedb/installer/commit/bb4786ea) Update crds for kubedb/apimachinery@e78c6ff7 (#804)
- [55ff929e](https://github.com/kubedb/installer/commit/55ff929e) Add ZooKeeper Versions (#776)
- [c3283eb4](https://github.com/kubedb/installer/commit/c3283eb4) Add mongodb perconaserver 7.0.4; Deprecate 4.2.7 & 4.4.10 (#802)
- [4ad98522](https://github.com/kubedb/installer/commit/4ad98522) Add percona versions for mongodb (#775)
- [0bbf1794](https://github.com/kubedb/installer/commit/0bbf1794) Update kubestash backup and restore task names (#766)
- [08e4002d](https://github.com/kubedb/installer/commit/08e4002d) Use kafka featureGate for kafkaConnector; Remove PSP (#792)
- [eafb83d0](https://github.com/kubedb/installer/commit/eafb83d0) Change dashboard api group to elasticsearch (#794)
- [625cc6e8](https://github.com/kubedb/installer/commit/625cc6e8) Remove kubedb-dashboard charts from the kubedb/kubedb-one chart (#793)
- [4b1a8f0c](https://github.com/kubedb/installer/commit/4b1a8f0c) Add if condition to ApiService creation for kafka (#786)
- [7cd2242d](https://github.com/kubedb/installer/commit/7cd2242d) Add Kafka connector (#784)
- [696850fc](https://github.com/kubedb/installer/commit/696850fc) Add runAsGroup; Mongo 7.0.4 -> 7.0.5 (#780)
- [34708816](https://github.com/kubedb/installer/commit/34708816) Update crds for kubedb/apimachinery@a72bb1ff (#781)
- [3bc7789a](https://github.com/kubedb/installer/commit/3bc7789a) Add mgversion for mongodb 7.0.4 (#763)
- [94e2d5b2](https://github.com/kubedb/installer/commit/94e2d5b2) Add MySQL 5.7.42-debian
- [04b2def1](https://github.com/kubedb/installer/commit/04b2def1) Add validator for autoscaler (#777)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.12.0-beta.1](https://github.com/kubedb/kafka/releases/tag/v0.12.0-beta.1)

- [34f4967f](https://github.com/kubedb/kafka/commit/34f4967f) Prepare for release v0.12.0-beta.1 (#68)
- [7176931c](https://github.com/kubedb/kafka/commit/7176931c) Move Kafka Podtemplate to ofshoot-api v2 (#66)
- [9454adf6](https://github.com/kubedb/kafka/commit/9454adf6) Update ci & makefile for crd-manager (#67)
- [fda770d8](https://github.com/kubedb/kafka/commit/fda770d8) Add kafka connector controller (#65)
- [6ed0ccd4](https://github.com/kubedb/kafka/commit/6ed0ccd4) Add Kafka connect  controller (#44)
- [18e9a45c](https://github.com/kubedb/kafka/commit/18e9a45c) update deps (#64)
- [a7dfb409](https://github.com/kubedb/kafka/commit/a7dfb409) Update makefile for dynamic crd installer (#63)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.4.0-beta.1](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.4.0-beta.1)

- [c77b4ae](https://github.com/kubedb/kubedb-manifest-plugin/commit/c77b4ae) Prepare for release v0.4.0-beta.1 (#37)
- [6a8a822](https://github.com/kubedb/kubedb-manifest-plugin/commit/6a8a822) Update component name (#35)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.25.0-beta.1](https://github.com/kubedb/mariadb/releases/tag/v0.25.0-beta.1)

- [c4d4942f](https://github.com/kubedb/mariadb/commit/c4d4942f8) Prepare for release v0.25.0-beta.1 (#250)
- [25fe3917](https://github.com/kubedb/mariadb/commit/25fe39177) Use ptr.Deref(); Update deps
- [c76704cc](https://github.com/kubedb/mariadb/commit/c76704cc8) Fix ci & makefile for crd-manager (#249)
- [67396abb](https://github.com/kubedb/mariadb/commit/67396abb9) Incorporate with apimachinery package name change from `stash` to `restore` (#248)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.1.0-beta.1](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.1.0-beta.1)

- [e8564fe](https://github.com/kubedb/mariadb-archiver/commit/e8564fe) Prepare for release v0.1.0-beta.1 (#5)
- [e5e8945](https://github.com/kubedb/mariadb-archiver/commit/e5e8945) Don't use fail-fast



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.21.0-beta.1](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.21.0-beta.1)

- [1c30e710](https://github.com/kubedb/mariadb-coordinator/commit/1c30e710) Prepare for release v0.21.0-beta.1 (#101)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.1.0-beta.1](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.1.0-beta.1)

- [adac38d](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/adac38d) Prepare for release v0.1.0-beta.1 (#5)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.34.0-beta.1](https://github.com/kubedb/memcached/releases/tag/v0.34.0-beta.1)

- [754ba398](https://github.com/kubedb/memcached/commit/754ba398) Prepare for release v0.34.0-beta.1 (#418)
- [abd9dbb6](https://github.com/kubedb/memcached/commit/abd9dbb6) Incorporate with apimachinery package name change from stash to restore (#417)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.34.0-beta.1](https://github.com/kubedb/mongodb/releases/tag/v0.34.0-beta.1)

- [c0c58448](https://github.com/kubedb/mongodb/commit/c0c58448b) Prepare for release v0.34.0-beta.1 (#606)
- [5df39d09](https://github.com/kubedb/mongodb/commit/5df39d09f) Update ci mgVersion;  Fix pointer dereference issue (#605)
- [e2781eae](https://github.com/kubedb/mongodb/commit/e2781eaea) Run ci with specific crd-manager branch (#604)
- [b57bc47a](https://github.com/kubedb/mongodb/commit/b57bc47ae) Add kubestash for health check (#603)
- [62cb9c81](https://github.com/kubedb/mongodb/commit/62cb9c816) Install crd-manager specifiying DATABASE (#602)
- [6bf45fe7](https://github.com/kubedb/mongodb/commit/6bf45fe72) 7.0.4 -> 7.0.5; update deps
- [e5b9841e](https://github.com/kubedb/mongodb/commit/e5b9841e5) Fix oplog backup directory (#601)
- [452b785f](https://github.com/kubedb/mongodb/commit/452b785f0) Add  Support for DB phase change for restoring using `KubeStash` (#586)
- [35d93d0b](https://github.com/kubedb/mongodb/commit/35d93d0bc) add ssl/tls args command (#595)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.2.0-beta.1](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.2.0-beta.1)

- [5680265](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/5680265) Prepare for release v0.2.0-beta.1 (#12)
- [72693c8](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/72693c8) Fix component driver status (#11)
- [0ea73ee](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/0ea73ee) Update deps (#10)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.4.0-beta.1](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.4.0-beta.1)

- [6ae8ae2](https://github.com/kubedb/mongodb-restic-plugin/commit/6ae8ae2) Prepare for release v0.4.0-beta.1 (#23)
- [d8e1636](https://github.com/kubedb/mongodb-restic-plugin/commit/d8e1636) Reorder the execution of cleanup funcs (#22)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.34.0-beta.1](https://github.com/kubedb/mysql/releases/tag/v0.34.0-beta.1)

- [e9dbf269](https://github.com/kubedb/mysql/commit/e9dbf269c) Prepare for release v0.34.0-beta.1 (#599)
- [44eda2d2](https://github.com/kubedb/mysql/commit/44eda2d25) Prepare for release v0.34.0-beta.1 (#598)
- [16dd4637](https://github.com/kubedb/mysql/commit/16dd46377) Fix pointer dereference issue (#597)
- [334c1a1d](https://github.com/kubedb/mysql/commit/334c1a1dd) Update ci & makefile for crd-manager (#596)
- [edb9b1a1](https://github.com/kubedb/mysql/commit/edb9b1a11) Fix binlog backup directory (#587)
- [fc6d7030](https://github.com/kubedb/mysql/commit/fc6d70303) Add Support for DB phase change for restoring using KubeStash (#594)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.2.0-beta.1](https://github.com/kubedb/mysql-archiver/releases/tag/v0.2.0-beta.1)

- [e5bdae3](https://github.com/kubedb/mysql-archiver/commit/e5bdae3) Prepare for release v0.2.0-beta.1 (#15)
- [7ef752c](https://github.com/kubedb/mysql-archiver/commit/7ef752c) Refactor + Cleanup wal-g example files (#14)
- [5857a8d](https://github.com/kubedb/mysql-archiver/commit/5857a8d) Don't use fail-fast



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.19.0-beta.1](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.19.0-beta.1)

- [59a11671](https://github.com/kubedb/mysql-coordinator/commit/59a11671) Prepare for release v0.19.0-beta.1 (#98)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.2.0-beta.1](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.2.0-beta.1)

- [d5771cf](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/d5771cf) Prepare for release v0.2.0-beta.1 (#5)
- [b4ffc6f](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/b4ffc6f) Fix component driver status & Update deps (#3)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.4.0-beta.1](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.4.0-beta.1)

- [105888a](https://github.com/kubedb/mysql-restic-plugin/commit/105888a) Prepare for release v0.4.0-beta.1 (#21)
- [b42d0cf](https://github.com/kubedb/mysql-restic-plugin/commit/b42d0cf) Removed `--all-databases` flag for restoring (#20)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.19.0-beta.1](https://github.com/kubedb/mysql-router-init/releases/tag/v0.19.0-beta.1)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.28.0-beta.1](https://github.com/kubedb/ops-manager/releases/tag/v0.28.0-beta.1)

- [5976d8ed](https://github.com/kubedb/ops-manager/commit/5976d8ed0) Prepare for release v0.28.0-beta.1 (#529)
- [90e4c315](https://github.com/kubedb/ops-manager/commit/90e4c3159) Update deps; Add license
- [d6c0e148](https://github.com/kubedb/ops-manager/commit/d6c0e1487) Add backupConfiguration `Pause` & `Resume` support for Kubestash (#528)
- [e9b4bfea](https://github.com/kubedb/ops-manager/commit/e9b4bfea0) Fix kafka vertical scaling ops request for ofshoot api v2 (#527)
- [b230d6bb](https://github.com/kubedb/ops-manager/commit/b230d6bb6) Made crd-manager non required
- [439031ae](https://github.com/kubedb/ops-manager/commit/439031aea) Fix operator installation in ci (#526)
- [88014501](https://github.com/kubedb/ops-manager/commit/88014501f) Seperate mongo ci according to profiles; Change `daily`'s schedule (#525)
- [335a3e49](https://github.com/kubedb/ops-manager/commit/335a3e49f) Add TLS support for Kafka Connect Cluster (#518)
- [69de3f3e](https://github.com/kubedb/ops-manager/commit/69de3f3e8) Run new mongo versions to ci (#524)
- [e5fbed83](https://github.com/kubedb/ops-manager/commit/e5fbed839) Incorporate with apimachinery package name change from stash to restore (#523)
- [6320384b](https://github.com/kubedb/ops-manager/commit/6320384bf) Reorganize recommendation pkg
- [8f2b36d7](https://github.com/kubedb/ops-manager/commit/8f2b36d72) Update wait condition in makefile (#522)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.28.0-beta.1](https://github.com/kubedb/percona-xtradb/releases/tag/v0.28.0-beta.1)

- [475a5e32](https://github.com/kubedb/percona-xtradb/commit/475a5e328) Prepare for release v0.28.0-beta.1 (#348)
- [4c1380ab](https://github.com/kubedb/percona-xtradb/commit/4c1380ab7) Incorporate with apimachinery package name change from `stash` to `restore` (#347)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.14.0-beta.1](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.14.0-beta.1)

- [560bc5c3](https://github.com/kubedb/percona-xtradb-coordinator/commit/560bc5c3) Prepare for release v0.14.0-beta.1 (#58)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.25.0-beta.1](https://github.com/kubedb/pg-coordinator/releases/tag/v0.25.0-beta.1)

- [bc296307](https://github.com/kubedb/pg-coordinator/commit/bc296307) Prepare for release v0.25.0-beta.1 (#148)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.28.0-beta.1](https://github.com/kubedb/pgbouncer/releases/tag/v0.28.0-beta.1)

- [55c248d5](https://github.com/kubedb/pgbouncer/commit/55c248d5) Prepare for release v0.28.0-beta.1 (#312)
- [1b86664a](https://github.com/kubedb/pgbouncer/commit/1b86664a) Incorporate with apimachinery package name change from stash to restore (#311)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.0.1](https://github.com/kubedb/pgpool/releases/tag/v0.0.1)

- [dbb333b](https://github.com/kubedb/pgpool/commit/dbb333b) Prepare for release v0.0.1 (#3)
- [b9c96e2](https://github.com/kubedb/pgpool/commit/b9c96e2) Pgpool operator (#2)
- [7c878e7](https://github.com/kubedb/pgpool/commit/7c878e7) C1:bootstrap Initialization project and basic api design
- [c437da3](https://github.com/kubedb/pgpool/commit/c437da3) C1:bootstrap Initialization project and basic api design



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.41.0-beta.1](https://github.com/kubedb/postgres/releases/tag/v0.41.0-beta.1)

- [72a1ee29](https://github.com/kubedb/postgres/commit/72a1ee294) Prepare for release v0.41.0-beta.1 (#708)
- [026598f4](https://github.com/kubedb/postgres/commit/026598f44) Prepare for release v0.41.0-beta.1 (#707)
- [8af305aa](https://github.com/kubedb/postgres/commit/8af305aa4) Use ptr.Deref(); Update deps
- [c7c0652d](https://github.com/kubedb/postgres/commit/c7c0652dc) Update ci & makefile for crd-manager (#706)
- [d468bdb3](https://github.com/kubedb/postgres/commit/d468bdb34) Fix wal backup directory (#705)
- [c6992bed](https://github.com/kubedb/postgres/commit/c6992bed8) Add Support for DB phase change for restoring using KubeStash (#704)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.2.0-beta.1](https://github.com/kubedb/postgres-archiver/releases/tag/v0.2.0-beta.1)

- [c4405c1](https://github.com/kubedb/postgres-archiver/commit/c4405c1) Prepare for release v0.2.0-beta.1 (#17)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.2.0-beta.1](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.2.0-beta.1)

- [dc4f85e](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/dc4f85e) Prepare for release v0.2.0-beta.1 (#15)
- [098365a](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/098365a) Update README.md (#14)
- [5ef571f](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/5ef571f) Update deps (#13)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.4.0-beta.1](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.4.0-beta.1)

- [4ed2b4a](https://github.com/kubedb/postgres-restic-plugin/commit/4ed2b4a) Prepare for release v0.4.0-beta.1 (#14)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.3.0-beta.1](https://github.com/kubedb/provider-aws/releases/tag/v0.3.0-beta.1)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.3.0-beta.1](https://github.com/kubedb/provider-azure/releases/tag/v0.3.0-beta.1)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.3.0-beta.1](https://github.com/kubedb/provider-gcp/releases/tag/v0.3.0-beta.1)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.41.0-beta.1](https://github.com/kubedb/provisioner/releases/tag/v0.41.0-beta.1)

- [52cb0fa9](https://github.com/kubedb/provisioner/commit/52cb0fa9c) Prepare for release v0.41.0-beta.1 (#75)
- [92f05e8e](https://github.com/kubedb/provisioner/commit/92f05e8e7) Add New Database support (#74)
- [514709fc](https://github.com/kubedb/provisioner/commit/514709fc9) Add ElasticsearchDashboard controllers (#73)
- [b826a5f1](https://github.com/kubedb/provisioner/commit/b826a5f1e) Add Support for DB phase change for restoring using KubeStash (#72)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.28.0-beta.1](https://github.com/kubedb/proxysql/releases/tag/v0.28.0-beta.1)

- [213ebfc4](https://github.com/kubedb/proxysql/commit/213ebfc43) Prepare for release v0.28.0-beta.1 (#327)
- [8427158e](https://github.com/kubedb/proxysql/commit/8427158ec) Incorporate with apimachinery package name change from stash to restore (#325)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.0.1](https://github.com/kubedb/rabbitmq/releases/tag/v0.0.1)

- [48d2ec95](https://github.com/kubedb/rabbitmq/commit/48d2ec95) Prepare for release v0.0.1 (#2)
- [d9dcec0f](https://github.com/kubedb/rabbitmq/commit/d9dcec0f) Add Rabbitmq controller (#1)
- [6844a9cf](https://github.com/kubedb/rabbitmq/commit/6844a9cf) Add Appscode Community license and release workflows



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.34.0-beta.1](https://github.com/kubedb/redis/releases/tag/v0.34.0-beta.1)

- [01290634](https://github.com/kubedb/redis/commit/01290634) Prepare for release v0.34.0-beta.1 (#517)
- [e51f93e1](https://github.com/kubedb/redis/commit/e51f93e1) Fix panic (#516)
- [dc75c163](https://github.com/kubedb/redis/commit/dc75c163) Update ci & makefile for crd-manager (#515)
- [09688f35](https://github.com/kubedb/redis/commit/09688f35) Add Support for DB phase change for restoring using KubeStash (#514)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.20.0-beta.1](https://github.com/kubedb/redis-coordinator/releases/tag/v0.20.0-beta.1)

- [fd3b2112](https://github.com/kubedb/redis-coordinator/commit/fd3b2112) Prepare for release v0.20.0-beta.1 (#89)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.4.0-beta.1](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.4.0-beta.1)

- [fac6226](https://github.com/kubedb/redis-restic-plugin/commit/fac6226) Prepare for release v0.4.0-beta.1 (#17)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.28.0-beta.1](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.28.0-beta.1)

- [f948a650](https://github.com/kubedb/replication-mode-detector/commit/f948a650) Prepare for release v0.28.0-beta.1 (#253)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.17.0-beta.1](https://github.com/kubedb/schema-manager/releases/tag/v0.17.0-beta.1)

- [f14516a9](https://github.com/kubedb/schema-manager/commit/f14516a9) Prepare for release v0.17.0-beta.1 (#97)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.0.1](https://github.com/kubedb/singlestore/releases/tag/v0.0.1)

- [8feeb79](https://github.com/kubedb/singlestore/commit/8feeb79) Prepare for release v0.0.1 (#5)
- [fb79ff9](https://github.com/kubedb/singlestore/commit/fb79ff9) Add Singlestore Operator (#4)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.0.1](https://github.com/kubedb/solr/releases/tag/v0.0.1)

- [58fb5b4](https://github.com/kubedb/solr/commit/58fb5b4) Prepare for release v0.0.1 (#1)
- [6b7c3ef](https://github.com/kubedb/solr/commit/6b7c3ef) Add release workflows
- [9db6c84](https://github.com/kubedb/solr/commit/9db6c84) Disable ferret db in catalog helm command. (#5)
- [19553e7](https://github.com/kubedb/solr/commit/19553e7) Add solr operator. (#3)
- [ff4b9ae](https://github.com/kubedb/solr/commit/ff4b9ae) Reset master (#4)
- [7804b0a](https://github.com/kubedb/solr/commit/7804b0a) Add initial controller implementation (#2)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.26.0-beta.1](https://github.com/kubedb/tests/releases/tag/v0.26.0-beta.1)

- [3cfc1212](https://github.com/kubedb/tests/commit/3cfc1212) Prepare for release v0.26.0-beta.1 (#292)
- [b810e690](https://github.com/kubedb/tests/commit/b810e690) increase cpu limit for vertical scaling (#289)
- [c43985ba](https://github.com/kubedb/tests/commit/c43985ba) Change dashboard api group (#291)
- [1b96881e](https://github.com/kubedb/tests/commit/1b96881e) Fix error logging
- [33f78143](https://github.com/kubedb/tests/commit/33f78143) forceCleanup PVCs for mongo (#288)
- [0dcd3e38](https://github.com/kubedb/tests/commit/0dcd3e38) Add PostgreSQL logical replication tests  (#202)
- [2f403c85](https://github.com/kubedb/tests/commit/2f403c85) Find profiles in array, Don't match with string (#286)
- [5aca2293](https://github.com/kubedb/tests/commit/5aca2293) Give time to PDB status to be updated (#285)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.17.0-beta.1](https://github.com/kubedb/ui-server/releases/tag/v0.17.0-beta.1)

- [98c1a6dd](https://github.com/kubedb/ui-server/commit/98c1a6dd) Prepare for release v0.17.0-beta.1 (#105)
- [8173cfc2](https://github.com/kubedb/ui-server/commit/8173cfc2) Implement SingularNameProvider



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.17.0-beta.1](https://github.com/kubedb/webhook-server/releases/tag/v0.17.0-beta.1)

- [2a84cedb](https://github.com/kubedb/webhook-server/commit/2a84cedb) Prepare for release v0.17.0-beta.1 (#89)
- [bb4a5c22](https://github.com/kubedb/webhook-server/commit/bb4a5c22) Add kafka connect-cluster (#87)
- [c46c6662](https://github.com/kubedb/webhook-server/commit/c46c6662) Add new Database support (#88)
- [c6387e9e](https://github.com/kubedb/webhook-server/commit/c6387e9e) Set default kubebuilder client for autoscaler (#86)
- [14c07899](https://github.com/kubedb/webhook-server/commit/14c07899) Incorporate apimachinery (#85)
- [266c79a0](https://github.com/kubedb/webhook-server/commit/266c79a0) Add kafka ops request validator (#84)




