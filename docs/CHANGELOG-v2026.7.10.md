---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2026.7.10
    name: Changelog-v2026.7.10
    parent: welcome
    weight: 20260710
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.7.10/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.7.10/
---

# KubeDB v2026.7.10 (2026-07-12)


## [kubedb/aerospike](https://github.com/kubedb/aerospike)

### [v0.2.0](https://github.com/kubedb/aerospike/releases/tag/v0.2.0)

- [25703340](https://github.com/kubedb/aerospike/commit/25703340) Replace hub with gh in update-release-tracker.sh (#6)
- [48a43aca](https://github.com/kubedb/aerospike/commit/48a43aca) Prepare for release v0.2.0 (#5)
- [874a67cd](https://github.com/kubedb/aerospike/commit/874a67cd) Add Health check Signed-off-by: Hiranmoy <hiranmoy@appscode.com>
- [74741a4a](https://github.com/kubedb/aerospike/commit/74741a4a) changed into kubedb style (#3)
- [b18ebec9](https://github.com/kubedb/aerospike/commit/b18ebec9) Revert "Add pkg/controllers and pkg/cmds/server for provisioner integration"
- [695399fa](https://github.com/kubedb/aerospike/commit/695399fa) Revert "Clean up unused functions and imports in AerospikeReconciler"
- [faea0167](https://github.com/kubedb/aerospike/commit/faea0167) Clean up unused functions and imports in AerospikeReconciler
- [c33f4d9e](https://github.com/kubedb/aerospike/commit/c33f4d9e) Add pkg/controllers and pkg/cmds/server for provisioner integration
- [80b747bf](https://github.com/kubedb/aerospike/commit/80b747bf) Harden github actions
- [c0d0ece3](https://github.com/kubedb/aerospike/commit/c0d0ece3) Configure dependabot refresh schedule (#2)
- [7b0820a9](https://github.com/kubedb/aerospike/commit/7b0820a9) add service creation Signed-off-by: HiranmoyChowdhury <hiranmoy@appscode.com>
- [91c4a271](https://github.com/kubedb/aerospike/commit/91c4a271) now working Signed-off-by: HiranmoyChowdhury <hiranmoy@appscode.com>
- [584fd6a4](https://github.com/kubedb/aerospike/commit/584fd6a4) build fix Signed-off-by: HiranmoyChowdhury <hiranmoy@appscode.com>
- [0e5ab45d](https://github.com/kubedb/aerospike/commit/0e5ab45d) initialised Signed-off-by: HiranmoyChowdhury <hiranmoy@appscode.com>



## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.66.0](https://github.com/kubedb/apimachinery/releases/tag/v0.66.0)

- [e40ad552](https://github.com/kubedb/apimachinery/commit/e40ad5529) Modernize golangci-lint config (#1826)
- [e0b3004f](https://github.com/kubedb/apimachinery/commit/e0b3004fa) Update sidekick deps (#1827)
- [ca218a64](https://github.com/kubedb/apimachinery/commit/ca218a648) Update for release KubeStash@v2026.7.10 (#1825)
- [a9ed525b](https://github.com/kubedb/apimachinery/commit/a9ed525b6) run make fmt (#1823)
- [c85055d3](https://github.com/kubedb/apimachinery/commit/c85055d3a) ACL secret updation (#1822)
- [bad1baeb](https://github.com/kubedb/apimachinery/commit/bad1baeb0) Disallow InPlace vertical scaling mode for Neo4j (#1819)
- [9b75e234](https://github.com/kubedb/apimachinery/commit/9b75e2347) Fix Webhook for Virtual Secret (#1817)
- [5c452059](https://github.com/kubedb/apimachinery/commit/5c4520591) Don't allow updating version between distro (#1821)
- [fa18e712](https://github.com/kubedb/apimachinery/commit/fa18e712f) Restrict cross-baseOS version upgrades for Postgres (#1781)
- [0ea56fe9](https://github.com/kubedb/apimachinery/commit/0ea56fe94) Added app-binding in mssqlserver (#1820)
- [14397c47](https://github.com/kubedb/apimachinery/commit/14397c47b) fix etcd service name (#1818)
- [210cad0a](https://github.com/kubedb/apimachinery/commit/210cad0ad) Add Clickhouse Shard Scaling Support (#1784)
- [00be056e](https://github.com/kubedb/apimachinery/commit/00be056e2) Add ReconfigureTLS ops type and TLS spec to OracleOpsRequest (#1791)
- [21bd385d](https://github.com/kubedb/apimachinery/commit/21bd385dc) changed lederElection period (#1799)
- [233f90be](https://github.com/kubedb/apimachinery/commit/233f90be2) Add MongoDB distro option
- [75751efe](https://github.com/kubedb/apimachinery/commit/75751efe9) Add MilvusBind, QdrantBind, WeaviateBind wrappers (#1809)
- [cc1005a9](https://github.com/kubedb/apimachinery/commit/cc1005a99) Add VerticalScalingMode to all vertical scaling specs (#1808)
- [dd7736f0](https://github.com/kubedb/apimachinery/commit/dd7736f01) Add HanaDB volume permission option (#1802)
- [ac0d35e3](https://github.com/kubedb/apimachinery/commit/ac0d35e3d) Add PostgresSynchronousReplicationSpec for configurable sync replication (#1782)
- [c8af2eb6](https://github.com/kubedb/apimachinery/commit/c8af2eb6e) Fix Zookeeper Ops (#1788)
- [3e7d7558](https://github.com/kubedb/apimachinery/commit/3e7d75582) Set default container resizePolicy for all databases (#1789)
- [df643d84](https://github.com/kubedb/apimachinery/commit/df643d84b) Fix pgpool webhook build after ReconfigurationSpec embed change (#1816)
- [e14fe41d](https://github.com/kubedb/apimachinery/commit/e14fe41d9) Add Weaviate Monitoring Support (#1811)
- [d78e2c2e](https://github.com/kubedb/apimachinery/commit/d78e2c2e2) Improve Branch APIs and Status for human redable + fix duck typing ci (#1813)
- [779aee46](https://github.com/kubedb/apimachinery/commit/779aee469) Document make fmt requirement before opening PRs (#1814)
- [c3375004](https://github.com/kubedb/apimachinery/commit/c33750041) Register Migration and MigrationList to scheme
- [e0f21be8](https://github.com/kubedb/apimachinery/commit/e0f21be8f) Use db specific migration kind (#1812)
- [31ad8176](https://github.com/kubedb/apimachinery/commit/31ad81768) Add MSSQL Server migrator API types (#1742)
- [a5cc9682](https://github.com/kubedb/apimachinery/commit/a5cc9682d) courier: add Branch spec.target.issuerRef for branch TLS (#1810)
- [1476d578](https://github.com/kubedb/apimachinery/commit/1476d5788) Add courier.kubedb.com API group (rename migrator to courier) (#1807)
- [8cc4a11b](https://github.com/kubedb/apimachinery/commit/8cc4a11bd) Enable TLS in Backup Port (#1783)
- [79935675](https://github.com/kubedb/apimachinery/commit/799356756) changed lederElection period (#1785)
- [4cd0fc34](https://github.com/kubedb/apimachinery/commit/4cd0fc34c) Fix Cassandra VolumeExpansion Webhook (#1786)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.51.0](https://github.com/kubedb/autoscaler/releases/tag/v0.51.0)

- [4c1773f5](https://github.com/kubedb/autoscaler/commit/4c1773f5) Prepare for release v0.51.0 (#312)
- [b60fd249](https://github.com/kubedb/autoscaler/commit/b60fd249) Modernize golangci-lint config (#311)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.19.0](https://github.com/kubedb/cassandra/releases/tag/v0.19.0)

- [04610462](https://github.com/kubedb/cassandra/commit/04610462) Prepare for release v0.19.0
- [9623922c](https://github.com/kubedb/cassandra/commit/9623922c) Modernize golangci-lint config (#97)
- [d5df1520](https://github.com/kubedb/cassandra/commit/d5df1520) Add InPlace mode to vertical scaling (#96)
- [1061bc5b](https://github.com/kubedb/cassandra/commit/1061bc5b) Add backup pause/resume support for ops requests (#93)
- [7b1cbf0e](https://github.com/kubedb/cassandra/commit/7b1cbf0e) Fix Standalone OpsReq not Working (#92)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.13.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.13.0)

- [f5c7c8f9](https://github.com/kubedb/cassandra-medusa-plugin/commit/f5c7c8f9) Prepare for release v0.13.0 (#43)
- [5f75b9ac](https://github.com/kubedb/cassandra-medusa-plugin/commit/5f75b9ac) Modernize golangci-lint config (#42)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.66.0](https://github.com/kubedb/cli/releases/tag/v0.66.0)

- [e338453c](https://github.com/kubedb/cli/commit/e338453c5) Prepare for release v0.66.0 (#836)
- [29f2cf6c](https://github.com/kubedb/cli/commit/29f2cf6ca) Modernize golangci-lint config (#835)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.21.0](https://github.com/kubedb/clickhouse/releases/tag/v0.21.0)

- [2a420dd6](https://github.com/kubedb/clickhouse/commit/2a420dd6) Prepare for release v0.21.0 (#125)
- [fbc3c3d6](https://github.com/kubedb/clickhouse/commit/fbc3c3d6) Modernize golangci-lint config (#124)
- [972ddd22](https://github.com/kubedb/clickhouse/commit/972ddd22) Add InPlace mode to vertical scaling (#122)
- [fccd8b2b](https://github.com/kubedb/clickhouse/commit/fccd8b2b) Add Shard Scaling Support (#119)
- [40de00ae](https://github.com/kubedb/clickhouse/commit/40de00ae) as (#123)
- [3b4ca7b6](https://github.com/kubedb/clickhouse/commit/3b4ca7b6) Add Archiver Support (#99)



## [kubedb/clickhouse-backup-plugin](https://github.com/kubedb/clickhouse-backup-plugin)

### [v0.3.0](https://github.com/kubedb/clickhouse-backup-plugin/releases/tag/v0.3.0)

- [46d0a5cf](https://github.com/kubedb/clickhouse-backup-plugin/commit/46d0a5cf) Prepare for release v0.3.0 (#30)
- [1b314899](https://github.com/kubedb/clickhouse-backup-plugin/commit/1b314899) Add golangci-lint config (#29)
- [e06cf170](https://github.com/kubedb/clickhouse-backup-plugin/commit/e06cf170) Add Incremental backup (#19)



## [kubedb/courier](https://github.com/kubedb/courier)

### [v0.6.0](https://github.com/kubedb/courier/releases/tag/v0.6.0)

- [acda30f](https://github.com/kubedb/courier/commit/acda30f) Prepare for release v0.6.0 (#50)
- [a8910f6](https://github.com/kubedb/courier/commit/a8910f6) Modernize golangci-lint config (#49)
- [4b881f5](https://github.com/kubedb/courier/commit/4b881f5) Fix CI: Update deps
- [afab26c](https://github.com/kubedb/courier/commit/afab26c) added appbing support for mssqlserver (#47)
- [147d289](https://github.com/kubedb/courier/commit/147d289) Fix Mongoshake field access on MongoDB source (#46)
- [38ed45b](https://github.com/kubedb/courier/commit/38ed45b) Reconcile per-engine {DB}Migration CRDs via Migration duck type (#45)
- [4ba344b](https://github.com/kubedb/courier/commit/4ba344b) Rename to courier: module kubedb.dev/courier, binary kubedb-courier; scaffold Branch + manager (#27)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.21.0](https://github.com/kubedb/crd-manager/releases/tag/v0.21.0)

- [ef77e885](https://github.com/kubedb/crd-manager/commit/ef77e885) Prepare for release v0.21.0 (#148)
- [1257a1c4](https://github.com/kubedb/crd-manager/commit/1257a1c4) Rename migrator to courier CRDs (#147)
- [e3df0cbf](https://github.com/kubedb/crd-manager/commit/e3df0cbf) Modernize golangci-lint config (#146)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.24.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.24.0)

- [2d5d0a3c](https://github.com/kubedb/dashboard-restic-plugin/commit/2d5d0a3c) Prepare for release v0.24.0 (#83)
- [3e204216](https://github.com/kubedb/dashboard-restic-plugin/commit/3e204216) Modernize golangci-lint config (#82)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.21.0](https://github.com/kubedb/db-client-go/releases/tag/v0.21.0)

- [bc3a25ce](https://github.com/kubedb/db-client-go/commit/bc3a25ce) Prepare for release v0.21.0 (#255)
- [fec3c231](https://github.com/kubedb/db-client-go/commit/fec3c231) Modernize golangci-lint config (#254)
- [a3da77c5](https://github.com/kubedb/db-client-go/commit/a3da77c5) add aerospike client (#251)
- [144cd45a](https://github.com/kubedb/db-client-go/commit/144cd45a) Add New function For Neo4j (#252)
- [4b2ca834](https://github.com/kubedb/db-client-go/commit/4b2ca834) Add Qdrant Recover Snapshot Reader Function (#253)



## [kubedb/db2](https://github.com/kubedb/db2)

### [v0.7.0](https://github.com/kubedb/db2/releases/tag/v0.7.0)

- [78ad3099](https://github.com/kubedb/db2/commit/78ad3099) Prepare for release v0.7.0 (#35)
- [538b3685](https://github.com/kubedb/db2/commit/538b3685) Modernize golangci-lint config (#34)



## [kubedb/db2-coordinator](https://github.com/kubedb/db2-coordinator)

### [v0.7.0](https://github.com/kubedb/db2-coordinator/releases/tag/v0.7.0)

- [3a80ff3](https://github.com/kubedb/db2-coordinator/commit/3a80ff3) Prepare for release v0.7.0 (#16)
- [5dc64ab](https://github.com/kubedb/db2-coordinator/commit/5dc64ab) Modernize golangci-lint config (#15)



## [kubedb/documentdb](https://github.com/kubedb/documentdb)

### [v0.3.0](https://github.com/kubedb/documentdb/releases/tag/v0.3.0)

- [ba6ac93b](https://github.com/kubedb/documentdb/commit/ba6ac93b) Prepare for release v0.3.0 (#39)
- [4fc3b678](https://github.com/kubedb/documentdb/commit/4fc3b678) Modernize golangci-lint config (#38)
- [bd9d564e](https://github.com/kubedb/documentdb/commit/bd9d564e) Run gofmt with updated golang-dev toolchain (#37)
- [86fce11f](https://github.com/kubedb/documentdb/commit/86fce11f) Add InPlace mode to vertical scaling (#36)
- [3f92d8c2](https://github.com/kubedb/documentdb/commit/3f92d8c2) updated documentdb deps (#34)
- [9d7aee2c](https://github.com/kubedb/documentdb/commit/9d7aee2c) fixed termination time & health check issue (#30)



## [kubedb/documentdb-coordinator](https://github.com/kubedb/documentdb-coordinator)

### [v0.2.0](https://github.com/kubedb/documentdb-coordinator/releases/tag/v0.2.0)

- [2d44eda](https://github.com/kubedb/documentdb-coordinator/commit/2d44eda) Prepare for release v0.2.0
- [dcfbce6](https://github.com/kubedb/documentdb-coordinator/commit/dcfbce6) Add golangci-lint config (#9)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.21.0](https://github.com/kubedb/druid/releases/tag/v0.21.0)

- [a2a417cc](https://github.com/kubedb/druid/commit/a2a417cc) Prepare for release v0.21.0 (#147)
- [743a802d](https://github.com/kubedb/druid/commit/743a802d) Modernize golangci-lint config (#146)
- [8a8a5898](https://github.com/kubedb/druid/commit/8a8a5898) Run gofmt with updated golang-dev toolchain (#145)
- [2087b2a7](https://github.com/kubedb/druid/commit/2087b2a7) Merge pull request #144 from kubedb/inplace-vscale
- [3088dc57](https://github.com/kubedb/druid/commit/3088dc57) Merge branch 'master' into inplace-vscale
- [1730010c](https://github.com/kubedb/druid/commit/1730010c) update deps
- [c07155c2](https://github.com/kubedb/druid/commit/c07155c2) Add backup pause/resume support for ops requests (#143)
- [f89865da](https://github.com/kubedb/druid/commit/f89865da) Add InPlace mode to vertical scaling



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.66.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.66.0)

- [16ff98fa](https://github.com/kubedb/elasticsearch/commit/16ff98faf) Prepare for release v0.66.0
- [51fe3f9a](https://github.com/kubedb/elasticsearch/commit/51fe3f9ad) Add golangci-lint config (#827)
- [20b313e0](https://github.com/kubedb/elasticsearch/commit/20b313e04) Run gofmt with updated golang-dev toolchain (#826)
- [5a50d5eb](https://github.com/kubedb/elasticsearch/commit/5a50d5ebe) Merge pull request #825 from kubedb/inplace-vscale
- [c376b130](https://github.com/kubedb/elasticsearch/commit/c376b130f) update deps
- [97d7e676](https://github.com/kubedb/elasticsearch/commit/97d7e6769) Add InPlace mode to vertical scaling



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.29.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.29.0)

- [1b29f94d](https://github.com/kubedb/elasticsearch-restic-plugin/commit/1b29f94d) Prepare for release v0.29.0 (#106)
- [91b6fb9f](https://github.com/kubedb/elasticsearch-restic-plugin/commit/91b6fb9f) Modernize golangci-lint config (#105)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.14.0](https://github.com/kubedb/gitops/releases/tag/v0.14.0)

- [5df5c4c0](https://github.com/kubedb/gitops/commit/5df5c4c0) Prepare for release v0.14.0 (#85)
- [6cf3ed69](https://github.com/kubedb/gitops/commit/6cf3ed69) Add golangci-lint config (#84)
- [58c5dd1d](https://github.com/kubedb/gitops/commit/58c5dd1d) Add StorageClass Migration Support (#83)
- [bea9b77b](https://github.com/kubedb/gitops/commit/bea9b77b) Prepare AI Generated PR (#82)
- [0dd5a2c4](https://github.com/kubedb/gitops/commit/0dd5a2c4) Add gitops support for Cassandra (#58)
- [1e58fdeb](https://github.com/kubedb/gitops/commit/1e58fdeb) Add GitOps support for Memcached (#70)
- [2bb69ced](https://github.com/kubedb/gitops/commit/2bb69ced) Add gitops support for Weaviate (#67)
- [3e44a434](https://github.com/kubedb/gitops/commit/3e44a434) Add GitOps support for ProxySQL (#74)
- [5606d8b6](https://github.com/kubedb/gitops/commit/5606d8b6) Add gitops support for Milvus (#63)



## [kubedb/hanadb](https://github.com/kubedb/hanadb)

### [v0.7.0](https://github.com/kubedb/hanadb/releases/tag/v0.7.0)

- [c293b3af](https://github.com/kubedb/hanadb/commit/c293b3af) Prepare for release v0.7.0 (#53)
- [5309a61e](https://github.com/kubedb/hanadb/commit/5309a61e) Modernize golangci-lint config (#52)
- [038da9ea](https://github.com/kubedb/hanadb/commit/038da9ea) Run gofmt with updated golang-dev toolchain (#51)
- [95713bbd](https://github.com/kubedb/hanadb/commit/95713bbd) Add InPlace mode to vertical scaling (#50)
- [fbc510b2](https://github.com/kubedb/hanadb/commit/fbc510b2) Add backup pause/resume support for ops requests (#48)



## [kubedb/hanadb-coordinator](https://github.com/kubedb/hanadb-coordinator)

### [v0.6.0](https://github.com/kubedb/hanadb-coordinator/releases/tag/v0.6.0)

- [424def31](https://github.com/kubedb/hanadb-coordinator/commit/424def31) Prepare for release v0.6.0 (#20)
- [24d6bcd0](https://github.com/kubedb/hanadb-coordinator/commit/24d6bcd0) Modernize golangci-lint config (#19)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.12.0](https://github.com/kubedb/hazelcast/releases/tag/v0.12.0)

- [532c0572](https://github.com/kubedb/hazelcast/commit/532c0572) Prepare for release v0.12.0 (#60)
- [57d96750](https://github.com/kubedb/hazelcast/commit/57d96750) Modernize golangci-lint config (#59)
- [4dd0083f](https://github.com/kubedb/hazelcast/commit/4dd0083f) Run gofmt with updated golang-dev toolchain (#58)
- [2a05722e](https://github.com/kubedb/hazelcast/commit/2a05722e) Add InPlace mode to vertical scaling (#57)
- [d737d88b](https://github.com/kubedb/hazelcast/commit/d737d88b) Add backup pause/resume support for ops requests (#55)
- [63be50eb](https://github.com/kubedb/hazelcast/commit/63be50eb) [Fix VolumeExpansion] Patch DB before Ops Succeeded (#56)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.13.0](https://github.com/kubedb/ignite/releases/tag/v0.13.0)

- [cc26dcb4](https://github.com/kubedb/ignite/commit/cc26dcb4) Prepare for release v0.13.0 (#66)
- [7fc3f493](https://github.com/kubedb/ignite/commit/7fc3f493) Modernize golangci-lint config (#65)
- [eee09fbe](https://github.com/kubedb/ignite/commit/eee09fbe) Run gofmt with updated golang-dev toolchain (#64)
- [161f7c48](https://github.com/kubedb/ignite/commit/161f7c48) Add InPlace mode to vertical scaling (#63)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2026.7.10](https://github.com/kubedb/installer/releases/tag/v2026.7.10)

- [b5e02e3c](https://github.com/kubedb/installer/commit/b5e02e3cf) Prepare for release v2026.7.10 (#2383)
- [34316464](https://github.com/kubedb/installer/commit/34316464c) Add ClickHouse Archiver (#2369)
- [55c3d8a3](https://github.com/kubedb/installer/commit/55c3d8a30) Add Docker Hub login to CI to avoid image pull rate limits (#2381)
- [8efad7bf](https://github.com/kubedb/installer/commit/8efad7bf4) Build kubedb-autoscaler chart dependency before role-aggregator (#2379)
- [662ecbc6](https://github.com/kubedb/installer/commit/662ecbc6c) Modernize golangci-lint config (#2378)
- [e5e8906b](https://github.com/kubedb/installer/commit/e5e8906bf) Add storage-metrics-server as optional sub-chart of kubedb-autoscaler (#2371)
- [824f22db](https://github.com/kubedb/installer/commit/824f22db8) Grant ops-manager patch access to pods/resize (#2376)
- [0340a29c](https://github.com/kubedb/installer/commit/0340a29c3) Add Virtual Secret Permission to Ops (#2356)
- [2a338443](https://github.com/kubedb/installer/commit/2a3384439) Add postgres extension support for pgvector, pg_repack, postgis, pg_c… (#2375)
- [71c509bf](https://github.com/kubedb/installer/commit/71c509bff) Add Neo4j Version (#2374)
- [291da0fe](https://github.com/kubedb/installer/commit/291da0fe8) Add New Flag to Neo4j Backup Plugin (#2373)
- [895cccd2](https://github.com/kubedb/installer/commit/895cccd2e) Add Weaviate Kubedb-Metrics (#2366)
- [bae82448](https://github.com/kubedb/installer/commit/bae82448a) Add image to catalouge and scripts (#2372)
- [9577bbd0](https://github.com/kubedb/installer/commit/9577bbd0b) Add pvc permissions for custom metrics (#2370)
- [4ecc6402](https://github.com/kubedb/installer/commit/4ecc64023) Add Branch RBAC to kubedb-courier ClusterRole (#2367)
- [ff7bec44](https://github.com/kubedb/installer/commit/ff7bec44f) Add Ignite Version 2.18.0 (#2339)
- [b597c03d](https://github.com/kubedb/installer/commit/b597c03db) Add Qdrant Backup Restore Functions (#2359)
- [ea0cbaed](https://github.com/kubedb/installer/commit/ea0cbaed6) Split courier migrations CRD into per-engine {DB}Migration CRDs (#2368)
- [044e4872](https://github.com/kubedb/installer/commit/044e48722) Add ClickHouse Ops-manger metrics configuration (#2363)
- [11fd9f8a](https://github.com/kubedb/installer/commit/11fd9f8ad) Add notes
- [0cfddf26](https://github.com/kubedb/installer/commit/0cfddf260) Rename kubedb-migrator install artifacts to kubedb-courier (#2364)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.37.0](https://github.com/kubedb/kafka/releases/tag/v0.37.0)

- [bdd290d7](https://github.com/kubedb/kafka/commit/bdd290d7) Prepare for release v0.37.0
- [9c0a0375](https://github.com/kubedb/kafka/commit/9c0a0375) Modernize golangci-lint config (#211)
- [d80b95c5](https://github.com/kubedb/kafka/commit/d80b95c5) Run gofmt with updated golang-dev toolchain (#210)
- [7472c2e0](https://github.com/kubedb/kafka/commit/7472c2e0) Add InPlace mode to vertical scaling (#208)
- [4e11da51](https://github.com/kubedb/kafka/commit/4e11da51) Fix Lint (#209)
- [0b903f3b](https://github.com/kubedb/kafka/commit/0b903f3b) Add backup pause/resume support for ops requests (#205)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.42.0](https://github.com/kubedb/kibana/releases/tag/v0.42.0)

- [3a104a24](https://github.com/kubedb/kibana/commit/3a104a24) Prepare for release v0.42.0 (#188)
- [7bd02f57](https://github.com/kubedb/kibana/commit/7bd02f57) Modernize golangci-lint config (#187)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.29.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.29.0)

- [29af77f5](https://github.com/kubedb/kubedb-manifest-plugin/commit/29af77f5) Prepare for release v0.29.0 (#140)
- [42eaca97](https://github.com/kubedb/kubedb-manifest-plugin/commit/42eaca97) Modernize golangci-lint config (#139)
- [70ae44b0](https://github.com/kubedb/kubedb-manifest-plugin/commit/70ae44b0) Add ClickHouse Manifest (#129)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.17.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.17.0)

- [9ed0acba](https://github.com/kubedb/kubedb-verifier/commit/9ed0acba) Prepare for release v0.17.0 (#57)
- [11fc3fb9](https://github.com/kubedb/kubedb-verifier/commit/11fc3fb9) Modernize golangci-lint config (#56)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.50.0](https://github.com/kubedb/mariadb/releases/tag/v0.50.0)

- [d13a5dbc](https://github.com/kubedb/mariadb/commit/d13a5dbc4) Prepare for release v0.50.0
- [9b45199c](https://github.com/kubedb/mariadb/commit/9b45199c7) Modernize golangci-lint config (#419)
- [92fdd97b](https://github.com/kubedb/mariadb/commit/92fdd97be) Run gofmt with updated golang-dev toolchain (#418)
- [ebe378d2](https://github.com/kubedb/mariadb/commit/ebe378d2a) Merge pull request #417 from kubedb/inplace-vscale
- [a49fa865](https://github.com/kubedb/mariadb/commit/a49fa8655) fix build
- [9c552272](https://github.com/kubedb/mariadb/commit/9c5522724) Add InPlace mode to vertical scaling
- [7eb561f4](https://github.com/kubedb/mariadb/commit/7eb561f4a) Resume paused backups when ops request completes (#414)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.26.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.26.0)

- [6d836f63](https://github.com/kubedb/mariadb-archiver/commit/6d836f63) Prepare for release v0.26.0
- [00ed9470](https://github.com/kubedb/mariadb-archiver/commit/00ed9470) Modernize golangci-lint config (#98)
- [22666eed](https://github.com/kubedb/mariadb-archiver/commit/22666eed) Update Distributed Inc Snapshot (#94)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.46.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.46.0)

- [d3722e71](https://github.com/kubedb/mariadb-coordinator/commit/d3722e71) Prepare for release v0.46.0 (#186)
- [d00e0323](https://github.com/kubedb/mariadb-coordinator/commit/d00e0323) Modernize golangci-lint config (#185)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.26.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.26.0)

- [8bac9b0b](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/8bac9b0b) Prepare for release v0.26.0 (#84)
- [13b34333](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/13b34333) Modernize golangci-lint config (#83)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.24.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.24.0)

- [ac94025d](https://github.com/kubedb/mariadb-restic-plugin/commit/ac94025d) Prepare for release v0.24.0 (#99)
- [6b3c8e78](https://github.com/kubedb/mariadb-restic-plugin/commit/6b3c8e78) Modernize golangci-lint config (#98)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.59.0](https://github.com/kubedb/memcached/releases/tag/v0.59.0)

- [d7c6e509](https://github.com/kubedb/memcached/commit/d7c6e5099) Prepare for release v0.59.0 (#549)
- [c678218d](https://github.com/kubedb/memcached/commit/c678218d3) Modernize golangci-lint config (#548)
- [7b058c67](https://github.com/kubedb/memcached/commit/7b058c67b) Run gofmt with updated golang-dev toolchain (#547)
- [732e96e8](https://github.com/kubedb/memcached/commit/732e96e85) Add InPlace mode to vertical scaling (#546)



## [kubedb/migrator](https://github.com/kubedb/migrator)

### [v0.6.0](https://github.com/kubedb/migrator/releases/tag/v0.6.0)

- [323b3ba9](https://github.com/kubedb/migrator/commit/323b3ba9) Prepare for release v0.6.0 (#42)
- [52aebdd9](https://github.com/kubedb/migrator/commit/52aebdd9) Modernize golangci-lint config (#41)
- [b1c5905d](https://github.com/kubedb/migrator/commit/b1c5905d) Added progress for mssql in incr and updated docker (#40)
- [24013168](https://github.com/kubedb/migrator/commit/24013168) MSSQLServer no arm64 support (#39)
- [8490fd48](https://github.com/kubedb/migrator/commit/8490fd48) Implement overall progress for mysql,postgres (#33)
- [385b75a1](https://github.com/kubedb/migrator/commit/385b75a1) Fix CI: resolve staticcheck QF1008 by aliasing embedded engine source struct (#38)
- [8266bc2e](https://github.com/kubedb/migrator/commit/8266bc2e) Parse per-engine migration config; adopt per-engine {DB}Migration types (#37)
- [f7f8d4eb](https://github.com/kubedb/migrator/commit/f7f8d4eb) Add MSSQL Server migration support (#22)
- [5116be6d](https://github.com/kubedb/migrator/commit/5116be6d) Name per-db images kubedb-migrator-<db> (#36)
- [d20f452f](https://github.com/kubedb/migrator/commit/d20f452f) Rename module to kubedb.dev/migrator and binary to kubedb-migrator (#35)



## [kubedb/milvus](https://github.com/kubedb/milvus)

### [v0.7.0](https://github.com/kubedb/milvus/releases/tag/v0.7.0)

- [06c779bc](https://github.com/kubedb/milvus/commit/06c779bc) Prepare for release v0.7.0 (#55)
- [b0fe9379](https://github.com/kubedb/milvus/commit/b0fe9379) Add InPlace mode to vertical scaling (#52)
- [1985ebc3](https://github.com/kubedb/milvus/commit/1985ebc3) Modernize golangci-lint config (#54)
- [ead36da6](https://github.com/kubedb/milvus/commit/ead36da6) Add backup pause/resume support for ops requests (#50)
- [923a3221](https://github.com/kubedb/milvus/commit/923a3221) Fix etcd service name (#53)
- [3014ea40](https://github.com/kubedb/milvus/commit/3014ea40) Add Milvus Http Port (#51)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.59.0](https://github.com/kubedb/mongodb/releases/tag/v0.59.0)

- [2b61b507](https://github.com/kubedb/mongodb/commit/2b61b5070) Prepare for release v0.59.0
- [5612cb7a](https://github.com/kubedb/mongodb/commit/5612cb7a4) Modernize golangci-lint config (#776)
- [5e4ffce3](https://github.com/kubedb/mongodb/commit/5e4ffce33) Support InPlace mode for vertical scaling (#773)
- [0d548d47](https://github.com/kubedb/mongodb/commit/0d548d473) Skip duplicate --auth flag for mongodb-community-server images (#775)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.27.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.27.0)

- [a88a5c40](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/a88a5c40) Prepare for release v0.27.0 (#88)
- [d15fb79f](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/d15fb79f) Modernize golangci-lint config (#87)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.29.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.29.0)

- [758ae6b9](https://github.com/kubedb/mongodb-restic-plugin/commit/758ae6b9) Prepare for release v0.29.0 (#135)
- [1d683e17](https://github.com/kubedb/mongodb-restic-plugin/commit/1d683e17) Modernize golangci-lint config (#134)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.21.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.21.0)

- [5712b812](https://github.com/kubedb/mssql-coordinator/commit/5712b812) Prepare for release v0.21.0 (#79)
- [79515e15](https://github.com/kubedb/mssql-coordinator/commit/79515e15) Modernize golangci-lint config (#78)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.21.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.21.0)

- [b45fc246](https://github.com/kubedb/mssqlserver/commit/b45fc246) Prepare for release v0.21.0 (#152)
- [ce78720f](https://github.com/kubedb/mssqlserver/commit/ce78720f) Modernize golangci-lint config (#151)
- [ba6e13c6](https://github.com/kubedb/mssqlserver/commit/ba6e13c6) Run gofmt with updated golang-dev toolchain (#150)
- [c6a7e0bd](https://github.com/kubedb/mssqlserver/commit/c6a7e0bd) Add InPlace mode to vertical scaling (#149)
- [6af75d70](https://github.com/kubedb/mssqlserver/commit/6af75d70) Add backup pause/resume support for ops requests (#146)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.20.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.20.0)

- [a115fd5](https://github.com/kubedb/mssqlserver-archiver/commit/a115fd5) Prepare for release v0.20.0 (#33)
- [db4ea18](https://github.com/kubedb/mssqlserver-archiver/commit/db4ea18) Modernize golangci-lint config (#32)



## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.20.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.20.0)

- [6ca5fb8](https://github.com/kubedb/mssqlserver-walg-plugin/commit/6ca5fb8) Prepare for release v0.20.0 (#66)
- [cf72afd](https://github.com/kubedb/mssqlserver-walg-plugin/commit/cf72afd) Modernize golangci-lint config (#65)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.59.0](https://github.com/kubedb/mysql/releases/tag/v0.59.0)

- [8db7c5a1](https://github.com/kubedb/mysql/commit/8db7c5a12) Prepare for release v0.59.0 (#768)
- [a3227d2b](https://github.com/kubedb/mysql/commit/a3227d2b9) Modernize golangci-lint config (#767)
- [40dc0ca7](https://github.com/kubedb/mysql/commit/40dc0ca73) Virtual Secret Ops Support (#766)
- [14279e37](https://github.com/kubedb/mysql/commit/14279e375) Support InPlace mode for vertical scaling (#764)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.27.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.27.0)

- [0b5887b8](https://github.com/kubedb/mysql-archiver/commit/0b5887b8) Prepare for release v0.27.0 (#112)
- [633ec559](https://github.com/kubedb/mysql-archiver/commit/633ec559) Modernize golangci-lint config (#111)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.44.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.44.0)

- [a7f57425](https://github.com/kubedb/mysql-coordinator/commit/a7f57425) Prepare for release v0.44.0 (#188)
- [eb1460fa](https://github.com/kubedb/mysql-coordinator/commit/eb1460fa) Modernize golangci-lint config (#187)
- [db883280](https://github.com/kubedb/mysql-coordinator/commit/db883280) Add Virtual Secret (#182)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.27.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.27.0)

- [20f4ca6c](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/20f4ca6c) Prepare for release v0.27.0 (#84)
- [d87841f1](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/d87841f1) Modernize golangci-lint config (#83)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.29.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.29.0)

- [c13c0001](https://github.com/kubedb/mysql-restic-plugin/commit/c13c0001) Prepare for release v0.29.0 (#120)
- [a907f9a4](https://github.com/kubedb/mysql-restic-plugin/commit/a907f9a4) Modernize golangci-lint config (#119)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.44.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.44.0)

- [fa3f225](https://github.com/kubedb/mysql-router-init/commit/fa3f225) Prepare for release v0.44.0 (#67)
- [a848737](https://github.com/kubedb/mysql-router-init/commit/a848737) Modernize golangci-lint config (#66)



## [kubedb/neo4j](https://github.com/kubedb/neo4j)

### [v0.7.0](https://github.com/kubedb/neo4j/releases/tag/v0.7.0)

- [82145ee3](https://github.com/kubedb/neo4j/commit/82145ee3) Prepare for release v0.7.0 (#53)
- [75d9a446](https://github.com/kubedb/neo4j/commit/75d9a446) Modernize golangci-lint config (#52)
- [59d2adbc](https://github.com/kubedb/neo4j/commit/59d2adbc) Run gofmt with updated golang-dev toolchain (#51)
- [ce910cbd](https://github.com/kubedb/neo4j/commit/ce910cbd) Add InPlace mode to vertical scaling (#50)
- [a42ab5c8](https://github.com/kubedb/neo4j/commit/a42ab5c8) Add backup pause/resume support for ops requests (#47)
- [33b7be9b](https://github.com/kubedb/neo4j/commit/33b7be9b) Enable TLS in Backup Port (#46)



## [kubedb/neo4j-backup-plugin](https://github.com/kubedb/neo4j-backup-plugin)

### [v0.2.0](https://github.com/kubedb/neo4j-backup-plugin/releases/tag/v0.2.0)

- [1f6495f](https://github.com/kubedb/neo4j-backup-plugin/commit/1f6495f) Prepare for release v0.2.0 (#7)
- [bfa6380](https://github.com/kubedb/neo4j-backup-plugin/commit/bfa6380) Modernize golangci-lint config (#6)
- [d52defa](https://github.com/kubedb/neo4j-backup-plugin/commit/d52defa) Enable TLS on Backup Port (#4)
- [75f6051](https://github.com/kubedb/neo4j-backup-plugin/commit/75f6051) Add Neo4jAdminArg Flag for Restore (#3)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.53.0](https://github.com/kubedb/ops-manager/releases/tag/v0.53.0)




## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.12.0](https://github.com/kubedb/oracle/releases/tag/v0.12.0)

- [a480172f](https://github.com/kubedb/oracle/commit/a480172f) Prepare for release v0.12.0
- [38d23162](https://github.com/kubedb/oracle/commit/38d23162) Add InPlace mode to vertical scaling (#68)
- [de9d332a](https://github.com/kubedb/oracle/commit/de9d332a) Modernize golangci-lint config (#69)
- [3680ced6](https://github.com/kubedb/oracle/commit/3680ced6) Add backup pause/resume support for ops requests (#64)
- [498dfb0f](https://github.com/kubedb/oracle/commit/498dfb0f) Implement ReconfigureTLS ops for Oracle (#65)
- [7aabdc27](https://github.com/kubedb/oracle/commit/7aabdc27) add GOVERNING_SVC_FQDN env to maincontainer (#66)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.12.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.12.0)

- [225c88c](https://github.com/kubedb/oracle-coordinator/commit/225c88c) Prepare for release v0.12.0 (#43)
- [9d5c3e0](https://github.com/kubedb/oracle-coordinator/commit/9d5c3e0) Modernize golangci-lint config (#42)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.53.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.53.0)

- [97cf90f7](https://github.com/kubedb/percona-xtradb/commit/97cf90f73) Prepare for release v0.53.0 (#467)
- [3e42a7e6](https://github.com/kubedb/percona-xtradb/commit/3e42a7e6a) Modernize golangci-lint config (#466)
- [45244a17](https://github.com/kubedb/percona-xtradb/commit/45244a177) Run gofmt with updated golang-dev toolchain (#465)
- [f459b566](https://github.com/kubedb/percona-xtradb/commit/f459b5663) Merge pull request #464 from kubedb/inplace-vscale
- [c63f7c33](https://github.com/kubedb/percona-xtradb/commit/c63f7c33d) fix build
- [88f0f161](https://github.com/kubedb/percona-xtradb/commit/88f0f1612) Add InPlace mode to vertical scaling



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.39.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.39.0)

- [703e0b6f](https://github.com/kubedb/percona-xtradb-coordinator/commit/703e0b6f) Prepare for release v0.39.0 (#133)
- [e0f3fda2](https://github.com/kubedb/percona-xtradb-coordinator/commit/e0f3fda2) Modernize golangci-lint config (#132)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.50.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.50.0)

- [eb30eb7a](https://github.com/kubedb/pg-coordinator/commit/eb30eb7a) Prepare for release v0.50.0 (#263)
- [440db3d9](https://github.com/kubedb/pg-coordinator/commit/440db3d9) Modernize golangci-lint config (#262)
- [d34653a1](https://github.com/kubedb/pg-coordinator/commit/d34653a1) Fix formatlsn8 to zero-pad LSN low word to 8 hex digits (#260)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.21.0](https://github.com/kubedb/pgpool/releases/tag/v0.21.0)

- [4ed0ce7c](https://github.com/kubedb/pgpool/commit/4ed0ce7c) Prepare for release v0.21.0 (#134)
- [ca87c1ca](https://github.com/kubedb/pgpool/commit/ca87c1ca) Add InPlace mode to Pgpool vertical scaling (#132)
- [1fa8149d](https://github.com/kubedb/pgpool/commit/1fa8149d) Modernize golangci-lint config (#133)
- [29e0c111](https://github.com/kubedb/pgpool/commit/29e0c111) Fix build after apimachinery update and propagate serviceAccountName to PetSet (#130)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.66.0](https://github.com/kubedb/postgres/releases/tag/v0.66.0)

- [a793a776](https://github.com/kubedb/postgres/commit/a793a776e) Prepare for release v0.66.0
- [e4fdd364](https://github.com/kubedb/postgres/commit/e4fdd3649) Modernize golangci-lint config (#914)
- [e9c93433](https://github.com/kubedb/postgres/commit/e9c934336) Add InPlace mode to vertical scaling (#913)
- [cbd36f30](https://github.com/kubedb/postgres/commit/cbd36f30f) Pass SYNC_REPLICATION_MODE, NUM_SYNC_REPLICAS, SYNC_COMMIT_LEVEL env vars for sync replication (#904)
- [4bb191ed](https://github.com/kubedb/postgres/commit/4bb191ed7) controller: treat DoNotTerminate as Halt when webhook is bypassed (#903)
- [cd72e642](https://github.com/kubedb/postgres/commit/cd72e6426) Add preserve-on-halt and resource-policy: keep annotations for services (#909)
- [e3490048](https://github.com/kubedb/postgres/commit/e3490048d) Update service preserve annotation to kubedb.com/resource-policy: keep (#906)
- [bb8aab36](https://github.com/kubedb/postgres/commit/bb8aab36b) Do not Delete Service if preserveannotation is present (#905)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.27.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.27.0)

- [a11c9ce5](https://github.com/kubedb/postgres-archiver/commit/a11c9ce5) Prepare for release v0.27.0
- [f7ecc95c](https://github.com/kubedb/postgres-archiver/commit/f7ecc95c) Modernize golangci-lint config (#113)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.27.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.27.0)

- [5b3a81c2](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/5b3a81c2) Prepare for release v0.27.0
- [bab69004](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/bab69004) Modernize golangci-lint config (#93)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.29.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.29.0)

- [1903f349](https://github.com/kubedb/postgres-restic-plugin/commit/1903f349) Prepare for release v0.29.0 (#118)
- [4d655bbd](https://github.com/kubedb/postgres-restic-plugin/commit/4d655bbd) Modernize golangci-lint config (#117)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.27.0](https://github.com/kubedb/provider-aws/releases/tag/v0.27.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.27.0](https://github.com/kubedb/provider-azure/releases/tag/v0.27.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.27.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.27.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.66.0](https://github.com/kubedb/provisioner/releases/tag/v0.66.0)




## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.53.0](https://github.com/kubedb/proxysql/releases/tag/v0.53.0)

- [5abcbb4a](https://github.com/kubedb/proxysql/commit/5abcbb4ae) Prepare for release v0.53.0 (#445)
- [2eb0c23b](https://github.com/kubedb/proxysql/commit/2eb0c23bb) Modernize golangci-lint config (#444)
- [f133983c](https://github.com/kubedb/proxysql/commit/f133983c4) Run gofmt with updated golang-dev toolchain (#443)
- [2289f1fc](https://github.com/kubedb/proxysql/commit/2289f1fca) Merge pull request #442 from kubedb/inplace-vscale
- [6c983e5c](https://github.com/kubedb/proxysql/commit/6c983e5ca) fix build
- [3613a945](https://github.com/kubedb/proxysql/commit/3613a945e) Add InPlace mode to vertical scaling



## [kubedb/qdrant](https://github.com/kubedb/qdrant)

### [v0.7.0](https://github.com/kubedb/qdrant/releases/tag/v0.7.0)

- [5ea7efa0](https://github.com/kubedb/qdrant/commit/5ea7efa0) Prepare for release v0.7.0 (#55)
- [8ed0db3f](https://github.com/kubedb/qdrant/commit/8ed0db3f) Modernize golangci-lint config (#54)
- [b955a139](https://github.com/kubedb/qdrant/commit/b955a139) Run gofmt with updated golang-dev toolchain (#53)
- [74e7dcd3](https://github.com/kubedb/qdrant/commit/74e7dcd3) Add InPlace mode to vertical scaling (#52)
- [3618bbdf](https://github.com/kubedb/qdrant/commit/3618bbdf) Add backup pause/resume support for ops requests (#50)



## [kubedb/qdrant-restic-plugin](https://github.com/kubedb/qdrant-restic-plugin)

### [v0.2.0](https://github.com/kubedb/qdrant-restic-plugin/releases/tag/v0.2.0)

- [806455a](https://github.com/kubedb/qdrant-restic-plugin/commit/806455a) Prepare for release v0.2.0
- [b4a6cfa](https://github.com/kubedb/qdrant-restic-plugin/commit/b4a6cfa) Backup And Restore Qdrant With Kubestash (#3)
- [ea6ceae](https://github.com/kubedb/qdrant-restic-plugin/commit/ea6ceae) Bump RESTIC_VERSION to 0.18.1-20260421 (#6)
- [ce7f703](https://github.com/kubedb/qdrant-restic-plugin/commit/ce7f703) Configure dependabot refresh schedule (#5)
- [dd20ad3](https://github.com/kubedb/qdrant-restic-plugin/commit/dd20ad3) Configure dependabot refresh schedule (#4)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.21.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.21.0)

- [4827fb02](https://github.com/kubedb/rabbitmq/commit/4827fb02) Prepare for release v0.21.0 (#146)
- [eeedbe3b](https://github.com/kubedb/rabbitmq/commit/eeedbe3b) Modernize golangci-lint config (#145)
- [5d45e3ec](https://github.com/kubedb/rabbitmq/commit/5d45e3ec) Add backup pause/resume support for ops requests (#142)
- [7d8fc997](https://github.com/kubedb/rabbitmq/commit/7d8fc997) Support InPlace mode for vertical scaling (#144)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.59.0](https://github.com/kubedb/redis/releases/tag/v0.59.0)

- [96ab2c93](https://github.com/kubedb/redis/commit/96ab2c933) Prepare for release v0.59.0
- [2e2658c5](https://github.com/kubedb/redis/commit/2e2658c5b) Modernize golangci-lint config (#666)
- [514f4eb4](https://github.com/kubedb/redis/commit/514f4eb4a) Run gofmt with updated golang-dev toolchain (#665)
- [c0cceb86](https://github.com/kubedb/redis/commit/c0cceb867) Add InPlace mode to vertical scaling (#663)
- [ecdded87](https://github.com/kubedb/redis/commit/ecdded873) Add backup pause for ops requests (#657)
- [19f7e344](https://github.com/kubedb/redis/commit/19f7e3443) Vnpay issue solve for redis sentinel (#655)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.45.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.45.0)

- [f0b13174](https://github.com/kubedb/redis-coordinator/commit/f0b13174) Prepare for release v0.45.0 (#166)
- [7645c40a](https://github.com/kubedb/redis-coordinator/commit/7645c40a) Modernize golangci-lint config (#165)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.29.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.29.0)

- [e3ffb6da](https://github.com/kubedb/redis-restic-plugin/commit/e3ffb6da) Prepare for release v0.29.0 (#113)
- [48d518bb](https://github.com/kubedb/redis-restic-plugin/commit/48d518bb) Modernize golangci-lint config (#112)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.53.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.53.0)

- [56178f11](https://github.com/kubedb/replication-mode-detector/commit/56178f11) Prepare for release v0.53.0 (#328)
- [1d74ce2d](https://github.com/kubedb/replication-mode-detector/commit/1d74ce2d) Modernize golangci-lint config (#327)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.42.0](https://github.com/kubedb/schema-manager/releases/tag/v0.42.0)

- [105cce50](https://github.com/kubedb/schema-manager/commit/105cce50) Prepare for release v0.42.0 (#176)
- [f41aa349](https://github.com/kubedb/schema-manager/commit/f41aa349) Modernize golangci-lint config (#175)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.21.0](https://github.com/kubedb/singlestore/releases/tag/v0.21.0)

- [f51c87f9](https://github.com/kubedb/singlestore/commit/f51c87f9) Prepare for release v0.21.0 (#139)
- [8d7aeb36](https://github.com/kubedb/singlestore/commit/8d7aeb36) Modernize golangci-lint config (#138)
- [daf2d45d](https://github.com/kubedb/singlestore/commit/daf2d45d) Run gofmt with updated golang-dev toolchain (#137)
- [ea9437b9](https://github.com/kubedb/singlestore/commit/ea9437b9) Merge pull request #135 from kubedb/inplace-vscale
- [5a7d19bf](https://github.com/kubedb/singlestore/commit/5a7d19bf) Merge branch 'master' into inplace-vscale
- [aa92b2d3](https://github.com/kubedb/singlestore/commit/aa92b2d3) fix build
- [6086ab90](https://github.com/kubedb/singlestore/commit/6086ab90) Add backup pause/resume support for ops requests (#134)
- [44b37bf1](https://github.com/kubedb/singlestore/commit/44b37bf1) Add InPlace mode to vertical scaling



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.21.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.21.0)

- [332c70d2](https://github.com/kubedb/singlestore-coordinator/commit/332c70d2) Prepare for release v0.21.0 (#78)
- [2f951e42](https://github.com/kubedb/singlestore-coordinator/commit/2f951e42) Modernize golangci-lint config (#77)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.24.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.24.0)

- [a6b971a3](https://github.com/kubedb/singlestore-restic-plugin/commit/a6b971a3) Prepare for release v0.24.0
- [1533cce0](https://github.com/kubedb/singlestore-restic-plugin/commit/1533cce0) Modernize golangci-lint config (#91)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.21.0](https://github.com/kubedb/solr/releases/tag/v0.21.0)

- [0c84e83d](https://github.com/kubedb/solr/commit/0c84e83d) Prepare for release v0.21.0 (#145)
- [e3f0587d](https://github.com/kubedb/solr/commit/e3f0587d) Modernize golangci-lint config (#144)
- [824f8d6f](https://github.com/kubedb/solr/commit/824f8d6f) Run gofmt with updated golang-dev toolchain (#143)
- [1a81c48a](https://github.com/kubedb/solr/commit/1a81c48a) Add InPlace mode to vertical scaling (#142)
- [fd7d5de3](https://github.com/kubedb/solr/commit/fd7d5de3) Add backup pause/resume support for ops requests (#141)
- [b8e6684b](https://github.com/kubedb/solr/commit/b8e6684b) Add Solr Secret Fix (#139)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.51.0](https://github.com/kubedb/tests/releases/tag/v0.51.0)

- [56088833](https://github.com/kubedb/tests/commit/56088833b) Prepare for release v0.51.0 (#546)
- [83e22f2c](https://github.com/kubedb/tests/commit/83e22f2ce) Rename migrator to courier and add per-engine migration e2e tests (#545)
- [c1a2fec8](https://github.com/kubedb/tests/commit/c1a2fec82) Modernize golangci-lint config (#544)
- [8eaafd81](https://github.com/kubedb/tests/commit/8eaafd81f) Add E2E tests for Qdrant (#535)
- [795a53f6](https://github.com/kubedb/tests/commit/795a53f66) Add E2E tests for Oracle (#532)
- [4bacfea3](https://github.com/kubedb/tests/commit/4bacfea31) Add E2E tests for Neo4j (#531)
- [4e492b3d](https://github.com/kubedb/tests/commit/4e492b3d6) Add E2E tests for Cassandra (#524)
- [f42ef953](https://github.com/kubedb/tests/commit/f42ef9532) Implement Postgres Ops-Request Using GitOps (#517)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.42.0](https://github.com/kubedb/ui-server/releases/tag/v0.42.0)

- [5d905205](https://github.com/kubedb/ui-server/commit/5d9052055) Prepare for release v0.42.0 (#213)
- [1b25fc35](https://github.com/kubedb/ui-server/commit/1b25fc35a) Modernize golangci-lint config (#212)



## [kubedb/weaviate](https://github.com/kubedb/weaviate)

### [v0.7.0](https://github.com/kubedb/weaviate/releases/tag/v0.7.0)

- [6e7f0a95](https://github.com/kubedb/weaviate/commit/6e7f0a95) Prepare for release v0.7.0 (#53)
- [0e405cbf](https://github.com/kubedb/weaviate/commit/0e405cbf) Add InPlace mode to vertical scaling (#50)
- [3d0642c6](https://github.com/kubedb/weaviate/commit/3d0642c6) Modernize golangci-lint config (#52)
- [f8e0e4a7](https://github.com/kubedb/weaviate/commit/f8e0e4a7) Add Weaviate Monitoring (#49)
- [c06bfe8b](https://github.com/kubedb/weaviate/commit/c06bfe8b) Fix Lint (#51)
- [4ef1154c](https://github.com/kubedb/weaviate/commit/4ef1154c) Add backup pause/resume support for ops requests (#47)
- [0eb4f587](https://github.com/kubedb/weaviate/commit/0eb4f587) Fix rotate_auth (#46)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.42.0](https://github.com/kubedb/webhook-server/releases/tag/v0.42.0)




## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.14.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.14.0)

- [d1839680](https://github.com/kubedb/xtrabackup-restic-plugin/commit/d1839680) Prepare for release v0.14.0 (#61)
- [ec488ba5](https://github.com/kubedb/xtrabackup-restic-plugin/commit/ec488ba5) Modernize golangci-lint config (#60)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.21.0](https://github.com/kubedb/zookeeper/releases/tag/v0.21.0)

- [7f1c5b82](https://github.com/kubedb/zookeeper/commit/7f1c5b82) Prepare for release v0.21.0 (#134)
- [04b7e785](https://github.com/kubedb/zookeeper/commit/04b7e785) Modernize golangci-lint config (#133)
- [07570662](https://github.com/kubedb/zookeeper/commit/07570662) Run gofmt with updated golang-dev toolchain (#132)
- [cc5c3caf](https://github.com/kubedb/zookeeper/commit/cc5c3caf) Add InPlace mode to vertical scaling (#131)
- [0f00aa2b](https://github.com/kubedb/zookeeper/commit/0f00aa2b) Add backup pause/resume support for ops requests (#130)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.21.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.21.0)

- [41b6d8b2](https://github.com/kubedb/zookeeper-restic-plugin/commit/41b6d8b2) Prepare for release v0.21.0
- [14c471b5](https://github.com/kubedb/zookeeper-restic-plugin/commit/14c471b5) Modernize golangci-lint config (#76)




