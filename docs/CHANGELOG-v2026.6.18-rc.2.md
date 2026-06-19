---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2026.6.18-rc.2
    name: Changelog-v2026.6.18-rc.2
    parent: welcome
    weight: 20260618
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.6.18-rc.2/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.6.18-rc.2/
---

# KubeDB v2026.6.18-rc.2 (2026-06-18)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.65.0-rc.2](https://github.com/kubedb/apimachinery/releases/tag/v0.65.0-rc.2)

- [2309f632](https://github.com/kubedb/apimachinery/commit/2309f632f) Update for release KubeStash@v2026.6.18-rc.2 (#1771)
- [20f0dc78](https://github.com/kubedb/apimachinery/commit/20f0dc781) Fix Lint (#1769)
- [d9be48e0](https://github.com/kubedb/apimachinery/commit/d9be48e09) Add gitops Hazelcast type (#1730)
- [e7402f25](https://github.com/kubedb/apimachinery/commit/e7402f256) documentdb-reconfigure (#1767)
- [d06f5f04](https://github.com/kubedb/apimachinery/commit/d06f5f04d) Add gitops HanaDB type (#1731)
- [81617de4](https://github.com/kubedb/apimachinery/commit/81617de4b) Add gitops Cassandra type (#1732)
- [2b0cbe95](https://github.com/kubedb/apimachinery/commit/2b0cbe95a) feat: add GitSyncer and Init *InitSpec fields for git-sync support (#1728)
- [371e3875](https://github.com/kubedb/apimachinery/commit/371e3875d) Add migration TLS (#1768)
- [352eb38c](https://github.com/kubedb/apimachinery/commit/352eb38cf) Add gitops Oracle type (#1737)
- [613ec382](https://github.com/kubedb/apimachinery/commit/613ec3825) Add gitops Weaviate type (#1739)
- [b0d2d686](https://github.com/kubedb/apimachinery/commit/b0d2d6861) Add gitops Ignite type (#1740)
- [b242fc47](https://github.com/kubedb/apimachinery/commit/b242fc478) Add gitops Milvus type (#1735)
- [550837ab](https://github.com/kubedb/apimachinery/commit/550837abf) Add gitops Neo4j type (#1734)
- [29d18fef](https://github.com/kubedb/apimachinery/commit/29d18fef2) Add gitops Qdrant type (#1738)
- [120dee06](https://github.com/kubedb/apimachinery/commit/120dee060) Add oracle wallet configurationsecret api (#1764)
- [6837ec90](https://github.com/kubedb/apimachinery/commit/6837ec90a) fix singlestore standalone tls (#1760)
- [75c375a3](https://github.com/kubedb/apimachinery/commit/75c375a3c) Exclude virtual auth secrets from persistent secret tracking (#1762)
- [0fb5640b](https://github.com/kubedb/apimachinery/commit/0fb5640ba) wv-ops (#1743)
- [376078d8](https://github.com/kubedb/apimachinery/commit/376078d84) Update documentdb raft client port (#1759)
- [57cbbd07](https://github.com/kubedb/apimachinery/commit/57cbbd07f) Add gitops ClickHouse type (#1733)
- [4ea06c51](https://github.com/kubedb/apimachinery/commit/4ea06c513) Allow init-script for percona-xtradb cluster (#1749)
- [b60722d5](https://github.com/kubedb/apimachinery/commit/b60722d59) Migrator cli api update for mysql (#1754)
- [9ae2bd64](https://github.com/kubedb/apimachinery/commit/9ae2bd644) Update ops api, add coordinator defaulting for HanaDB (#1723)
- [0f38d07e](https://github.com/kubedb/apimachinery/commit/0f38d07ee) updated documentdb constant (#1758)
- [c313668b](https://github.com/kubedb/apimachinery/commit/c313668bb) added halted and config (#1757)
- [d90741a1](https://github.com/kubedb/apimachinery/commit/d90741a1a) update autoscaler api for missing database (#1752)
- [0c9dd395](https://github.com/kubedb/apimachinery/commit/0c9dd395b) documentdb helpers (#1756)
- [c0556262](https://github.com/kubedb/apimachinery/commit/c05562622) Documentdb-adminSecret (#1729)
- [6fead405](https://github.com/kubedb/apimachinery/commit/6fead4056) Add Milvus Autoscaler Webhook Validation (#1747)
- [4b212df0](https://github.com/kubedb/apimachinery/commit/4b212df05) Fix swagger.json
- [7c98b972](https://github.com/kubedb/apimachinery/commit/7c98b972c) Remove Etcd support (#1751)
- [74ea704f](https://github.com/kubedb/apimachinery/commit/74ea704fe) Fix nil pointer dereferences in storage migration webhook validators (#1748)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.50.0-rc.2](https://github.com/kubedb/autoscaler/releases/tag/v0.50.0-rc.2)

- [c7169bb5](https://github.com/kubedb/autoscaler/commit/c7169bb5) Prepare for release v0.50.0-rc.2 (#308)
- [001c9856](https://github.com/kubedb/autoscaler/commit/001c9856) feat: add Weaviate compute and storage autoscaler (#301)
- [a51c5af2](https://github.com/kubedb/autoscaler/commit/a51c5af2) Add Milvus Autoscaler (#306)
- [edcd29ac](https://github.com/kubedb/autoscaler/commit/edcd29ac) Add support for Neo4j (#298)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.18.0-rc.2](https://github.com/kubedb/cassandra/releases/tag/v0.18.0-rc.2)

- [a14d2587](https://github.com/kubedb/cassandra/commit/a14d2587) Prepare for release v0.18.0-rc.2 (#89)
- [5f55678b](https://github.com/kubedb/cassandra/commit/5f55678b) Honor user-provided renewBefore in TLS certificate ops (#85)
- [602bdf49](https://github.com/kubedb/cassandra/commit/602bdf49) Add StorageMigration OpsRequest support for Cassandra (#81)
- [ef021693](https://github.com/kubedb/cassandra/commit/ef021693) Add NetworkPolicyFlavor support for cilium (#88)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.12.0-rc.2](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.12.0-rc.2)

- [2584476c](https://github.com/kubedb/cassandra-medusa-plugin/commit/2584476c) Prepare for release v0.12.0-rc.2 (#40)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.20.0-rc.2](https://github.com/kubedb/clickhouse/releases/tag/v0.20.0-rc.2)

- [6f1712d6](https://github.com/kubedb/clickhouse/commit/6f1712d6) Prepare for release v0.20.0-rc.2 (#114)
- [610759d3](https://github.com/kubedb/clickhouse/commit/610759d3) Honor user-provided renewBefore in TLS certificate ops (#110)
- [b8b9b4d5](https://github.com/kubedb/clickhouse/commit/b8b9b4d5) Add NetworkPolicyFlavor support (#113)



## [kubedb/clickhouse-backup-plugin](https://github.com/kubedb/clickhouse-backup-plugin)

### [v0.2.0-rc.2](https://github.com/kubedb/clickhouse-backup-plugin/releases/tag/v0.2.0-rc.2)

- [5d481e4c](https://github.com/kubedb/clickhouse-backup-plugin/commit/5d481e4c) Prepare for release v0.2.0-rc.2 (#26)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.20.0-rc.2](https://github.com/kubedb/crd-manager/releases/tag/v0.20.0-rc.2)

- [d9cd2e04](https://github.com/kubedb/crd-manager/commit/d9cd2e04) Prepare for release v0.20.0-rc.2 (#140)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.23.0-rc.2](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.23.0-rc.2)

- [cdd75695](https://github.com/kubedb/dashboard-restic-plugin/commit/cdd75695) Prepare for release v0.23.0-rc.2 (#79)
- [c3afe4cd](https://github.com/kubedb/dashboard-restic-plugin/commit/c3afe4cd) Add restic backup progress streaming (#78)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.20.0-rc.2](https://github.com/kubedb/db-client-go/releases/tag/v0.20.0-rc.2)

- [ff2e61c8](https://github.com/kubedb/db-client-go/commit/ff2e61c8) Prepare for release v0.20.0-rc.2 (#249)
- [b0c28745](https://github.com/kubedb/db-client-go/commit/b0c28745) Add HanaDB TLS (#234)
- [d97925e7](https://github.com/kubedb/db-client-go/commit/d97925e7) Add New Func to Neo4j (#241)



## [kubedb/db2](https://github.com/kubedb/db2)

### [v0.6.0-rc.2](https://github.com/kubedb/db2/releases/tag/v0.6.0-rc.2)

- [99854604](https://github.com/kubedb/db2/commit/99854604) Prepare for release v0.6.0-rc.2 (#32)
- [f2fe477e](https://github.com/kubedb/db2/commit/f2fe477e) Fix Db2 deletion (#31)
- [e263fee1](https://github.com/kubedb/db2/commit/e263fee1) Add NetworkPolicyFlavor support (#30)



## [kubedb/db2-coordinator](https://github.com/kubedb/db2-coordinator)

### [v0.6.0-rc.2](https://github.com/kubedb/db2-coordinator/releases/tag/v0.6.0-rc.2)

- [9f4c409](https://github.com/kubedb/db2-coordinator/commit/9f4c409) Prepare for release v0.6.0-rc.2 (#13)



## [kubedb/documentdb](https://github.com/kubedb/documentdb)

### [v0.2.0-rc.2](https://github.com/kubedb/documentdb/releases/tag/v0.2.0-rc.2)

- [153e7117](https://github.com/kubedb/documentdb/commit/153e7117) Prepare for release v0.2.0-rc.2 (#27)
- [d5322e67](https://github.com/kubedb/documentdb/commit/d5322e67) documentdb-reconfigure (#25)
- [148c6fb0](https://github.com/kubedb/documentdb/commit/148c6fb0) Add OpsRequest support for DocumentDB ported from Postgres (#22)
- [999aef78](https://github.com/kubedb/documentdb/commit/999aef78) bring reverted changes (#24)
- [0195ac68](https://github.com/kubedb/documentdb/commit/0195ac68) Update apimachinery (#23)
- [846bca49](https://github.com/kubedb/documentdb/commit/846bca49) Update apimachinery (#21)
- [dbeb9d3c](https://github.com/kubedb/documentdb/commit/dbeb9d3c) Clustering  (#7)
- [f34712d9](https://github.com/kubedb/documentdb/commit/f34712d9) Add NetworkPolicyFlavor support (#19)



## [kubedb/documentdb-coordinator](https://github.com/kubedb/documentdb-coordinator)

### [v0.1.0-rc.2](https://github.com/kubedb/documentdb-coordinator/releases/tag/v0.1.0-rc.2)

- [ee91f72](https://github.com/kubedb/documentdb-coordinator/commit/ee91f72) Use gh cli instead of old hub cli
- [05a2c8b](https://github.com/kubedb/documentdb-coordinator/commit/05a2c8b) Prepare for release v0.1.0-rc.2 (#4)
- [020634d](https://github.com/kubedb/documentdb-coordinator/commit/020634d) Fix DocumentDBCoordinatorClientPort (2389 → 2379) (#3)
- [4b8cbae](https://github.com/kubedb/documentdb-coordinator/commit/4b8cbae) Up deps on apimachinery (#2)
- [d2f2e7b](https://github.com/kubedb/documentdb-coordinator/commit/d2f2e7b) Bootstrap (#1)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.20.0-rc.2](https://github.com/kubedb/druid/releases/tag/v0.20.0-rc.2)

- [7ea60517](https://github.com/kubedb/druid/commit/7ea60517) Prepare for release v0.20.0-rc.2 (#140)
- [89371af1](https://github.com/kubedb/druid/commit/89371af1) Honor user-provided renewBefore in TLS certificate ops (#135)
- [f4f4627f](https://github.com/kubedb/druid/commit/f4f4627f) Fix Panic Issue For ExternallyManaged authSecret (#139)
- [e89ff10a](https://github.com/kubedb/druid/commit/e89ff10a) Add NetworkPolicyFlavor support (#138)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.65.0-rc.2](https://github.com/kubedb/elasticsearch/releases/tag/v0.65.0-rc.2)

- [b0421f8d](https://github.com/kubedb/elasticsearch/commit/b0421f8d3) Prepare for release v0.65.0-rc.2 (#818)
- [7a09aa83](https://github.com/kubedb/elasticsearch/commit/7a09aa83c) Honor user-provided renewBefore in TLS certificate ops (#815)
- [1c3cb34f](https://github.com/kubedb/elasticsearch/commit/1c3cb34f1) feat: implement git-sync init container for Elasticsearch (#814)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.28.0-rc.2](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.28.0-rc.2)

- [ac051dbf](https://github.com/kubedb/elasticsearch-restic-plugin/commit/ac051dbf) Prepare for release v0.28.0-rc.2 (#102)
- [021ea97a](https://github.com/kubedb/elasticsearch-restic-plugin/commit/021ea97a) Add restic backup progress streaming (#101)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.13.0-rc.2](https://github.com/kubedb/gitops/releases/tag/v0.13.0-rc.2)

- [32a65cd4](https://github.com/kubedb/gitops/commit/32a65cd4) Prepare for release v0.13.0-rc.2 (#80)
- [0131a5f7](https://github.com/kubedb/gitops/commit/0131a5f7) Add GitOps support for RabbitMQ (#72)
- [a7f78221](https://github.com/kubedb/gitops/commit/a7f78221) Add gitops support for Neo4j (#62)
- [cba199d7](https://github.com/kubedb/gitops/commit/cba199d7) Skip OpsCreation If Any Same Type Ops InProgress (#79)



## [kubedb/hanadb](https://github.com/kubedb/hanadb)

### [v0.6.0-rc.2](https://github.com/kubedb/hanadb/releases/tag/v0.6.0-rc.2)

- [b3f6ef5b](https://github.com/kubedb/hanadb/commit/b3f6ef5b) Prepare for release v0.6.0-rc.2 (#45)
- [e38a3433](https://github.com/kubedb/hanadb/commit/e38a3433) Add StorageMigration OpsRequest support (#34)
- [c08ac12e](https://github.com/kubedb/hanadb/commit/c08ac12e) Add tls, reconfigure tls, vertical scaling, rotate auth, volume expantion ops (#38)
- [696ffce2](https://github.com/kubedb/hanadb/commit/696ffce2) Add NetworkPolicyFlavor support (#44)



## [kubedb/hanadb-coordinator](https://github.com/kubedb/hanadb-coordinator)

### [v0.5.0-rc.2](https://github.com/kubedb/hanadb-coordinator/releases/tag/v0.5.0-rc.2)

- [51502038](https://github.com/kubedb/hanadb-coordinator/commit/51502038) Prepare for release v0.5.0-rc.2 (#16)
- [5371f1fa](https://github.com/kubedb/hanadb-coordinator/commit/5371f1fa) Disable automatic backups and fix failover handling (#15)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.11.0-rc.2](https://github.com/kubedb/hazelcast/releases/tag/v0.11.0-rc.2)

- [841ee968](https://github.com/kubedb/hazelcast/commit/841ee968) Prepare for release v0.11.0-rc.2 (#52)
- [13deaa26](https://github.com/kubedb/hazelcast/commit/13deaa26) Add NetworkPolicyFlavor support (#50)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.12.0-rc.2](https://github.com/kubedb/ignite/releases/tag/v0.12.0-rc.2)

- [abaa3f61](https://github.com/kubedb/ignite/commit/abaa3f61) Prepare for release v0.12.0-rc.2 (#61)
- [0e4fa9e6](https://github.com/kubedb/ignite/commit/0e4fa9e6) Fix Ignite deletion (#60)
- [1ccf34ce](https://github.com/kubedb/ignite/commit/1ccf34ce) Add NetworkPolicyFlavor support for cilium (#59)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2026.6.18-rc.2](https://github.com/kubedb/installer/releases/tag/v2026.6.18-rc.2)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.36.0-rc.2](https://github.com/kubedb/kafka/releases/tag/v0.36.0-rc.2)

- [da458b15](https://github.com/kubedb/kafka/commit/da458b15) Prepare for release v0.36.0-rc.2 (#202)
- [d95e5cbc](https://github.com/kubedb/kafka/commit/d95e5cbc) Honor user-provided renewBefore in TLS certificate ops (#198)
- [8ac2cd7b](https://github.com/kubedb/kafka/commit/8ac2cd7b) Add NetworkPolicyFlavor support for cilium (#201)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.41.0-rc.2](https://github.com/kubedb/kibana/releases/tag/v0.41.0-rc.2)

- [73bafa3b](https://github.com/kubedb/kibana/commit/73bafa3b) Prepare for release v0.41.0-rc.2 (#185)
- [d1ec131e](https://github.com/kubedb/kibana/commit/d1ec131e) Add network-policy-flavor flag for cilium support (#184)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.28.0-rc.2](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.28.0-rc.2)

- [0ebaf789](https://github.com/kubedb/kubedb-manifest-plugin/commit/0ebaf789) Prepare for release v0.28.0-rc.2 (#135)
- [834ef012](https://github.com/kubedb/kubedb-manifest-plugin/commit/834ef012) Add restic backup progress streaming (#134)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.16.0-rc.2](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.16.0-rc.2)

- [fe877028](https://github.com/kubedb/kubedb-verifier/commit/fe877028) Prepare for release v0.16.0-rc.2 (#53)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.49.0-rc.2](https://github.com/kubedb/mariadb/releases/tag/v0.49.0-rc.2)

- [dece23fb](https://github.com/kubedb/mariadb/commit/dece23fb3) Prepare for release v0.49.0-rc.2 (#409)
- [5249f7ac](https://github.com/kubedb/mariadb/commit/5249f7ac9) Honor user-provided renewBefore in TLS certificate ops (#404)
- [b1a16bc6](https://github.com/kubedb/mariadb/commit/b1a16bc61) Use endpoint from ResticStats.Summary (#407)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.25.0-rc.2](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.25.0-rc.2)

- [8cbbe471](https://github.com/kubedb/mariadb-archiver/commit/8cbbe471) Prepare for release v0.25.0-rc.2 (#95)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.45.0-rc.2](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.45.0-rc.2)

- [11fc976d](https://github.com/kubedb/mariadb-coordinator/commit/11fc976d) Prepare for release v0.45.0-rc.2 (#181)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.25.0-rc.2](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.25.0-rc.2)

- [730ee57c](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/730ee57c) Prepare for release v0.25.0-rc.2 (#80)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.23.0-rc.2](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.23.0-rc.2)

- [03ee3ee4](https://github.com/kubedb/mariadb-restic-plugin/commit/03ee3ee4) Prepare for release v0.23.0-rc.2 (#94)
- [ab0be84b](https://github.com/kubedb/mariadb-restic-plugin/commit/ab0be84b) Add restic backup progress streaming (#92)
- [d3bc1259](https://github.com/kubedb/mariadb-restic-plugin/commit/d3bc1259) Update Backup Job Name for Distributed (#93)



## [kubedb/migrator-cli](https://github.com/kubedb/migrator-cli)

### [v0.5.0-rc.2](https://github.com/kubedb/migrator-cli/releases/tag/v0.5.0-rc.2)

- [fe2606c9](https://github.com/kubedb/migrator-cli/commit/fe2606c9) Prepare for release v0.5.0-rc.2 (#25)
- [b7088fad](https://github.com/kubedb/migrator-cli/commit/b7088fad) Mysql migration init (#11)
- [06810c85](https://github.com/kubedb/migrator-cli/commit/06810c85) Update README.md



## [kubedb/migrator-operator](https://github.com/kubedb/migrator-operator)

### [v0.5.0-rc.2](https://github.com/kubedb/migrator-operator/releases/tag/v0.5.0-rc.2)

- [a091c07](https://github.com/kubedb/migrator-operator/commit/a091c07) Prepare for release v0.5.0-rc.2 (#23)
- [23f8959](https://github.com/kubedb/migrator-operator/commit/23f8959) Add mysql support (#19)
- [3a9acef](https://github.com/kubedb/migrator-operator/commit/3a9acef) fixed extraconfig overwrite issue (#21)



## [kubedb/milvus](https://github.com/kubedb/milvus)

### [v0.6.0-rc.2](https://github.com/kubedb/milvus/releases/tag/v0.6.0-rc.2)

- [fa57da9b](https://github.com/kubedb/milvus/commit/fa57da9b) Prepare for release v0.6.0-rc.2 (#48)
- [85d63146](https://github.com/kubedb/milvus/commit/85d63146) Add HorizontalScaling OpsRequest support (#45)
- [41e84651](https://github.com/kubedb/milvus/commit/41e84651) Honor user-provided renewBefore in TLS certificate ops (#40)
- [d364f887](https://github.com/kubedb/milvus/commit/d364f887) Fix Milvus deletion (#44)
- [1ae0debe](https://github.com/kubedb/milvus/commit/1ae0debe) Add RotateAuth OpsRequest support for Milvus (#46)
- [79cba14c](https://github.com/kubedb/milvus/commit/79cba14c) Add NetworkPolicyFlavor support for cilium (#43)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.58.0-rc.2](https://github.com/kubedb/mongodb/releases/tag/v0.58.0-rc.2)

- [b2085f89](https://github.com/kubedb/mongodb/commit/b2085f897) Prepare for release v0.58.0-rc.2 (#767)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.26.0-rc.2](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.26.0-rc.2)

- [fe132cd7](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/fe132cd7) Prepare for release v0.26.0-rc.2 (#85)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.28.0-rc.2](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.28.0-rc.2)

- [d6b5fa11](https://github.com/kubedb/mongodb-restic-plugin/commit/d6b5fa11) Prepare for release v0.28.0-rc.2 (#131)
- [58aef2eb](https://github.com/kubedb/mongodb-restic-plugin/commit/58aef2eb) Add Backup Progress Streaming Support in Snapshot Status (#130)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.20.0-rc.2](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.20.0-rc.2)

- [2716be6e](https://github.com/kubedb/mssql-coordinator/commit/2716be6e) Prepare for release v0.20.0-rc.2 (#73)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.20.0-rc.2](https://github.com/kubedb/mssqlserver/releases/tag/v0.20.0-rc.2)

- [418c160c](https://github.com/kubedb/mssqlserver/commit/418c160c) Prepare for release v0.20.0-rc.2 (#141)
- [2fadba5a](https://github.com/kubedb/mssqlserver/commit/2fadba5a) Add NetworkPolicyFlavor support for cilium (#139)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.19.0-rc.2](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.19.0-rc.2)

- [df59193](https://github.com/kubedb/mssqlserver-archiver/commit/df59193) Prepare for release v0.19.0-rc.2 (#30)



## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.19.0-rc.2](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.19.0-rc.2)

- [1a44640](https://github.com/kubedb/mssqlserver-walg-plugin/commit/1a44640) Prepare for release v0.19.0-rc.2 (#62)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.26.0-rc.2](https://github.com/kubedb/mysql-archiver/releases/tag/v0.26.0-rc.2)

- [530f2468](https://github.com/kubedb/mysql-archiver/commit/530f2468) Prepare for release v0.26.0-rc.2 (#109)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.43.0-rc.2](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.43.0-rc.2)

- [67143c6d](https://github.com/kubedb/mysql-coordinator/commit/67143c6d) Prepare for release v0.43.0-rc.2 (#183)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.26.0-rc.2](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.26.0-rc.2)

- [3fd6f8d9](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/3fd6f8d9) Prepare for release v0.26.0-rc.2 (#81)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.28.0-rc.2](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.28.0-rc.2)

- [85ba443a](https://github.com/kubedb/mysql-restic-plugin/commit/85ba443a) Prepare for release v0.28.0-rc.2 (#116)
- [edb738be](https://github.com/kubedb/mysql-restic-plugin/commit/edb738be) Add Backup Progress Streaming Support in Snapshot Status (#113)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.43.0-rc.2](https://github.com/kubedb/mysql-router-init/releases/tag/v0.43.0-rc.2)

- [b0ab3a5](https://github.com/kubedb/mysql-router-init/commit/b0ab3a5) Prepare for release v0.43.0-rc.2 (#64)



## [kubedb/neo4j](https://github.com/kubedb/neo4j)

### [v0.6.0-rc.2](https://github.com/kubedb/neo4j/releases/tag/v0.6.0-rc.2)

- [2639a89b](https://github.com/kubedb/neo4j/commit/2639a89b) Prepare for release v0.6.0-rc.2 (#42)
- [f742558f](https://github.com/kubedb/neo4j/commit/f742558f) Fix passing credential as a literal env value in the PetSet (#41)
- [c465d3c8](https://github.com/kubedb/neo4j/commit/c465d3c8) feat: implement git-sync init container for Neo4j (#36)
- [4f787bca](https://github.com/kubedb/neo4j/commit/4f787bca) Fix Neo4j Deletion (#40)
- [e6ab6aa3](https://github.com/kubedb/neo4j/commit/e6ab6aa3) Add NetworkPolicyFlavor support (#39)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.52.0-rc.2](https://github.com/kubedb/ops-manager/releases/tag/v0.52.0-rc.2)

- [9c2b9c23](https://github.com/kubedb/ops-manager/commit/9c2b9c231) Prepare for release v0.52.0-rc.2 (#876)
- [f9d41073](https://github.com/kubedb/ops-manager/commit/f9d410734) Add HanaDB TLS, reconfigure TLS (#847)
- [a7d0dcbb](https://github.com/kubedb/ops-manager/commit/a7d0dcbb6) dcoumendb-registered (#875)
- [880772d0](https://github.com/kubedb/ops-manager/commit/880772d0e) Add oracle ops reconfigure (#850)
- [82dc17d5](https://github.com/kubedb/ops-manager/commit/82dc17d53) added weaviate ops (#848)
- [58e5d73c](https://github.com/kubedb/ops-manager/commit/58e5d73c1) Add Recommendation Engine support for Milvus (#870)
- [c015013f](https://github.com/kubedb/ops-manager/commit/c015013f8) Add Recommendation Engine support for RabbitMQ (#861)



## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.11.0-rc.2](https://github.com/kubedb/oracle/releases/tag/v0.11.0-rc.2)

- [91f689e2](https://github.com/kubedb/oracle/commit/91f689e2) Prepare for release v0.11.0-rc.2 (#60)
- [666716b1](https://github.com/kubedb/oracle/commit/666716b1) Implement VolumeExpansion ops for Oracle (#46)
- [6c4afce5](https://github.com/kubedb/oracle/commit/6c4afce5) Add VerticalScaling OpsRequest implementation (#47)
- [5bf3fe65](https://github.com/kubedb/oracle/commit/5bf3fe65) Implement RotateAuthentication for Oracle (#48)
- [4ffeaede](https://github.com/kubedb/oracle/commit/4ffeaede) add oracle ops restart reconfigure and appbinding (#42)
- [44e872ea](https://github.com/kubedb/oracle/commit/44e872ea) add appbinding for oracle (#51)
- [1cc52267](https://github.com/kubedb/oracle/commit/1cc52267) Add NetworkPolicyFlavor support (#59)
- [df02ec98](https://github.com/kubedb/oracle/commit/df02ec98) Honor user-provided renewBefore in TLS certificate ops (#56)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.11.0-rc.2](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.11.0-rc.2)

- [5c50fdd](https://github.com/kubedb/oracle-coordinator/commit/5c50fdd) Prepare for release v0.11.0-rc.2 (#38)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.52.0-rc.2](https://github.com/kubedb/percona-xtradb/releases/tag/v0.52.0-rc.2)

- [446a97f4](https://github.com/kubedb/percona-xtradb/commit/446a97f45) Prepare for release v0.52.0-rc.2 (#460)
- [0b5518e3](https://github.com/kubedb/percona-xtradb/commit/0b5518e33) feat: implement git-sync init container for PerconaXtraDB (#456)
- [10329f89](https://github.com/kubedb/percona-xtradb/commit/10329f895) Honor user-provided renewBefore in TLS certificate ops (#457)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.38.0-rc.2](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.38.0-rc.2)

- [b33e29b9](https://github.com/kubedb/percona-xtradb-coordinator/commit/b33e29b9) Prepare for release v0.38.0-rc.2 (#129)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.49.0-rc.2](https://github.com/kubedb/pg-coordinator/releases/tag/v0.49.0-rc.2)

- [c2213131](https://github.com/kubedb/pg-coordinator/commit/c2213131) Prepare for release v0.49.0-rc.2 (#254)
- [46d694d9](https://github.com/kubedb/pg-coordinator/commit/46d694d9) Document pod injection role in AGENTS.md (#253)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.52.0-rc.2](https://github.com/kubedb/pgbouncer/releases/tag/v0.52.0-rc.2)

- [cc7ae286](https://github.com/kubedb/pgbouncer/commit/cc7ae2860) Prepare for release v0.52.0-rc.2 (#417)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.20.0-rc.2](https://github.com/kubedb/pgpool/releases/tag/v0.20.0-rc.2)

- [a361fefb](https://github.com/kubedb/pgpool/commit/a361fefb) Prepare for release v0.20.0-rc.2 (#125)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.65.0-rc.2](https://github.com/kubedb/postgres/releases/tag/v0.65.0-rc.2)

- [98471e23](https://github.com/kubedb/postgres/commit/98471e234) Prepare for release v0.65.0-rc.2 (#898)
- [ae9804e8](https://github.com/kubedb/postgres/commit/ae9804e8b) Use endpoint from ResticStats.Summary (#895)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.26.0-rc.2](https://github.com/kubedb/postgres-archiver/releases/tag/v0.26.0-rc.2)

- [43553c97](https://github.com/kubedb/postgres-archiver/commit/43553c97) Prepare for release v0.26.0-rc.2 (#110)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.26.0-rc.2](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.26.0-rc.2)

- [597ce968](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/597ce968) Prepare for release v0.26.0-rc.2 (#91)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.28.0-rc.2](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.28.0-rc.2)

- [8b893309](https://github.com/kubedb/postgres-restic-plugin/commit/8b893309) Prepare for release v0.28.0-rc.2 (#113)
- [4702b178](https://github.com/kubedb/postgres-restic-plugin/commit/4702b178) Add restic backup progress streaming (#112)
- [78b32801](https://github.com/kubedb/postgres-restic-plugin/commit/78b32801) Bump postgres 16 image from 16.1 to 16.4
- [6984c384](https://github.com/kubedb/postgres-restic-plugin/commit/6984c384) Fix WaitForDBConnection not logging the actual connection error (#111)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.26.0-rc.2](https://github.com/kubedb/provider-aws/releases/tag/v0.26.0-rc.2)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.26.0-rc.2](https://github.com/kubedb/provider-azure/releases/tag/v0.26.0-rc.2)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.26.0-rc.2](https://github.com/kubedb/provider-gcp/releases/tag/v0.26.0-rc.2)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.65.0-rc.2](https://github.com/kubedb/provisioner/releases/tag/v0.65.0-rc.2)

- [e5bf679f](https://github.com/kubedb/provisioner/commit/e5bf679f5) Prepare for release v0.65.0-rc.2 (#214)
- [9fd231eb](https://github.com/kubedb/provisioner/commit/9fd231eb4) Add missing DBs for cilium support (#212)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.52.0-rc.2](https://github.com/kubedb/proxysql/releases/tag/v0.52.0-rc.2)

- [fb2e8d61](https://github.com/kubedb/proxysql/commit/fb2e8d616) Prepare for release v0.52.0-rc.2 (#439)
- [75fcf154](https://github.com/kubedb/proxysql/commit/75fcf1540) Honor user-provided renewBefore in TLS certificate ops (#436)
- [3c4cd633](https://github.com/kubedb/proxysql/commit/3c4cd6331) Implement RotateAuthentication for ProxySQL (#432)



## [kubedb/qdrant](https://github.com/kubedb/qdrant)

### [v0.6.0-rc.2](https://github.com/kubedb/qdrant/releases/tag/v0.6.0-rc.2)

- [46a46c95](https://github.com/kubedb/qdrant/commit/46a46c95) Prepare for release v0.6.0-rc.2 (#48)
- [2448f7e5](https://github.com/kubedb/qdrant/commit/2448f7e5) Honor user-provided renewBefore in TLS certificate ops (#44)
- [f9d88701](https://github.com/kubedb/qdrant/commit/f9d88701) Fix Qdrant deletion (#47)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.20.0-rc.2](https://github.com/kubedb/rabbitmq/releases/tag/v0.20.0-rc.2)

- [c981cd46](https://github.com/kubedb/rabbitmq/commit/c981cd46) Prepare for release v0.20.0-rc.2 (#138)
- [1ae8fd38](https://github.com/kubedb/rabbitmq/commit/1ae8fd38) Add support for cilium network  policy (#136)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.58.0-rc.2](https://github.com/kubedb/redis/releases/tag/v0.58.0-rc.2)

- [38804ba3](https://github.com/kubedb/redis/commit/38804ba31) Prepare for release v0.58.0-rc.2 (#651)
- [05f9d6d5](https://github.com/kubedb/redis/commit/05f9d6d50) Honor user-provided renewBefore in TLS certificate ops (#647)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.44.0-rc.2](https://github.com/kubedb/redis-coordinator/releases/tag/v0.44.0-rc.2)

- [8cc88c9c](https://github.com/kubedb/redis-coordinator/commit/8cc88c9c) Prepare for release v0.44.0-rc.2 (#162)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.28.0-rc.2](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.28.0-rc.2)

- [c9a8d01b](https://github.com/kubedb/redis-restic-plugin/commit/c9a8d01b) Prepare for release v0.28.0-rc.2 (#109)
- [eaeb0c7a](https://github.com/kubedb/redis-restic-plugin/commit/eaeb0c7a) Add restic backup progress streaming (#108)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.52.0-rc.2](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.52.0-rc.2)

- [fe05bb5e](https://github.com/kubedb/replication-mode-detector/commit/fe05bb5e) Prepare for release v0.52.0-rc.2 (#325)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.41.0-rc.2](https://github.com/kubedb/schema-manager/releases/tag/v0.41.0-rc.2)

- [380ec889](https://github.com/kubedb/schema-manager/commit/380ec889) Prepare for release v0.41.0-rc.2 (#172)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.20.0-rc.2](https://github.com/kubedb/singlestore/releases/tag/v0.20.0-rc.2)

- [5be14e4d](https://github.com/kubedb/singlestore/commit/5be14e4d) Prepare for release v0.20.0-rc.2 (#129)
- [98d3f455](https://github.com/kubedb/singlestore/commit/98d3f455) Fix SingleStore deletion (#128)
- [5c8a7b45](https://github.com/kubedb/singlestore/commit/5c8a7b45) Honor user-provided renewBefore in TLS certificate ops (#125)
- [d6a82975](https://github.com/kubedb/singlestore/commit/d6a82975) Add RotateAuth Ops Request Support (#122)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.23.0-rc.2](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.23.0-rc.2)

- [1cbde534](https://github.com/kubedb/singlestore-restic-plugin/commit/1cbde534) Prepare for release v0.23.0-rc.2 (#88)
- [0ac6f752](https://github.com/kubedb/singlestore-restic-plugin/commit/0ac6f752) Add restic backup progress streaming (#87)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.20.0-rc.2](https://github.com/kubedb/solr/releases/tag/v0.20.0-rc.2)

- [1c955c70](https://github.com/kubedb/solr/commit/1c955c70) Prepare for release v0.20.0-rc.2 (#136)
- [443bf0ad](https://github.com/kubedb/solr/commit/443bf0ad) Fix Solr deletion (#134)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.50.0-rc.2](https://github.com/kubedb/tests/releases/tag/v0.50.0-rc.2)

- [5829f980](https://github.com/kubedb/tests/commit/5829f9808) Prepare for release v0.50.0-rc.2 (#541)
- [66234e49](https://github.com/kubedb/tests/commit/66234e492) Update kubedb apimachinery vendor and fix AppBinding pointer type (#540)
- [957fa4e5](https://github.com/kubedb/tests/commit/957fa4e5d) Add E2E tests for ClickHouse (#525)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.41.0-rc.2](https://github.com/kubedb/ui-server/releases/tag/v0.41.0-rc.2)

- [2d9cc487](https://github.com/kubedb/ui-server/commit/2d9cc4875) Prepare for release v0.41.0-rc.2 (#209)



## [kubedb/weaviate](https://github.com/kubedb/weaviate)

### [v0.6.0-rc.2](https://github.com/kubedb/weaviate/releases/tag/v0.6.0-rc.2)

- [7226c70c](https://github.com/kubedb/weaviate/commit/7226c70c) Prepare for release v0.6.0-rc.2 (#43)
- [9d3b6c57](https://github.com/kubedb/weaviate/commit/9d3b6c57) weaviate restart ops-request added (#12)
- [14d862e3](https://github.com/kubedb/weaviate/commit/14d862e3) Fix Weaviate deletion (#41)
- [690c3ad0](https://github.com/kubedb/weaviate/commit/690c3ad0) Add NetworkPolicyFlavor support (#40)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.41.0-rc.2](https://github.com/kubedb/webhook-server/releases/tag/v0.41.0-rc.2)

- [2db5df91](https://github.com/kubedb/webhook-server/commit/2db5df913) Prepare for release v0.41.0-rc.2 (#224)
- [3d375b73](https://github.com/kubedb/webhook-server/commit/3d375b730) documentdb-ops-webhook (#223)
- [cc9b766e](https://github.com/kubedb/webhook-server/commit/cc9b766ed) Add Cassandra & ClickHouse ops and Neo4j autoscaler webhook registrations (#222)
- [e341c482](https://github.com/kubedb/webhook-server/commit/e341c482f) create oracle ops reconfigure (#211)
- [65259ba2](https://github.com/kubedb/webhook-server/commit/65259ba23) Add HanaDB ops (#216)
- [4fd5117d](https://github.com/kubedb/webhook-server/commit/4fd5117db) Add Milvus Autoscaler Validation (#221)
- [0b56463a](https://github.com/kubedb/webhook-server/commit/0b56463a7) Add weaviate ops (#218)



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.13.0-rc.2](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.13.0-rc.2)

- [7405795f](https://github.com/kubedb/xtrabackup-restic-plugin/commit/7405795f) Prepare for release v0.13.0-rc.2 (#57)
- [3c66fa34](https://github.com/kubedb/xtrabackup-restic-plugin/commit/3c66fa34) Add restic backup progress streaming (#56)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.20.0-rc.2](https://github.com/kubedb/zookeeper/releases/tag/v0.20.0-rc.2)

- [23412a6b](https://github.com/kubedb/zookeeper/commit/23412a6b) Prepare for release v0.20.0-rc.2 (#127)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.20.0-rc.2](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.20.0-rc.2)

- [74c137ba](https://github.com/kubedb/zookeeper-restic-plugin/commit/74c137ba) Prepare for release v0.20.0-rc.2 (#73)
- [308c991d](https://github.com/kubedb/zookeeper-restic-plugin/commit/308c991d) Add restic backup progress streaming (#72)




