---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2026.4.13-rc.0
    name: Changelog-v2026.4.13-rc.0
    parent: welcome
    weight: 20260413
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.4.13-rc.0/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.4.13-rc.0/
---

# KubeDB v2026.4.13-rc.0 (2026-04-17)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.64.0-rc.0](https://github.com/kubedb/apimachinery/releases/tag/v0.64.0-rc.0)

- [0dd1d315](https://github.com/kubedb/apimachinery/commit/0dd1d3155) Update for release KubeStash@v2026.4.13-rc.0 (#1647)
- [d0174436](https://github.com/kubedb/apimachinery/commit/d0174436f) Update cassandra webhook (#1645)
- [d64cb9b9](https://github.com/kubedb/apimachinery/commit/d64cb9b9e) change-ConfigSecretName-func (#1642)
- [e7cf45b2](https://github.com/kubedb/apimachinery/commit/e7cf45b22) Add type constants for init backup check (#1641)
- [f029669d](https://github.com/kubedb/apimachinery/commit/f029669d5) add-oracle-config-validator (#1639)
- [12491135](https://github.com/kubedb/apimachinery/commit/124911359) Add support for documentdb (#1604)
- [c8d2f8a5](https://github.com/kubedb/apimachinery/commit/c8d2f8a5f) Add Neo4j Opsreq Api (#1615)
- [93a538c9](https://github.com/kubedb/apimachinery/commit/93a538c94) Pgpool RemoveCustomConfig on LoadBalancingSpec (#1636)
- [8e275216](https://github.com/kubedb/apimachinery/commit/8e275216c) Qdrant Ops Request  (#1631)
- [d585bc1d](https://github.com/kubedb/apimachinery/commit/d585bc1d6) Make mssql tls field required, Refactor PredicateFuncs (#1624)
- [b20e1ee7](https://github.com/kubedb/apimachinery/commit/b20e1ee7c) Cleanup CVEs (#1638)
- [4863df4e](https://github.com/kubedb/apimachinery/commit/4863df4ef) Add double-optin utils; Add GetPredicateFuncsForSelf() (#1637)
- [48f4c200](https://github.com/kubedb/apimachinery/commit/48f4c200a) Add HanaDB monitoring (#1583)
- [7c1e745e](https://github.com/kubedb/apimachinery/commit/7c1e745e1) Add Generic Archiver to Database Mapping Functions (#1634)
- [f57d49cd](https://github.com/kubedb/apimachinery/commit/f57d49cde) Add Milvus Monitoring Support (#1590)
- [08ca9f42](https://github.com/kubedb/apimachinery/commit/08ca9f428) Add mssql arbiter vertical scaling api (#1626)
- [f129d2fc](https://github.com/kubedb/apimachinery/commit/f129d2fc0) Allow parallel processing of ops with the same name in diff namespaces (#1617)
- [d700721a](https://github.com/kubedb/apimachinery/commit/d700721a6) Add oracle init spec (#1599)
- [1832c036](https://github.com/kubedb/apimachinery/commit/1832c036f) Fix MySQL VolumeExpansion (#1633)
- [b790503c](https://github.com/kubedb/apimachinery/commit/b790503c8) Fix DB Container Resource Limit, Request (#1630)
- [0fc180f7](https://github.com/kubedb/apimachinery/commit/0fc180f72) Add separate api fields for load balancing specification (#1618)
- [c166dda1](https://github.com/kubedb/apimachinery/commit/c166dda1d) Fix oracle observer (#1627)
- [317d979b](https://github.com/kubedb/apimachinery/commit/317d979b6) fix pgbouncer webhook (#1616)
- [52687dc9](https://github.com/kubedb/apimachinery/commit/52687dc9f) remove validation for initcontainer (#1625)
- [2e3d3bce](https://github.com/kubedb/apimachinery/commit/2e3d3bce5) Mark archiver in PostgresVersion as optional (#1621)
- [60eb421b](https://github.com/kubedb/apimachinery/commit/60eb421b3) Use cert-manager v1.19.4 (#1619)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.49.0-rc.0](https://github.com/kubedb/autoscaler/releases/tag/v0.49.0-rc.0)

- [fee22931](https://github.com/kubedb/autoscaler/commit/fee22931) Prepare for release v0.49.0-rc.0 (#286)
- [a24c29b8](https://github.com/kubedb/autoscaler/commit/a24c29b8) Add operator sharding support (#276)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.17.0-rc.0](https://github.com/kubedb/cassandra/releases/tag/v0.17.0-rc.0)

- [d74a7a9d](https://github.com/kubedb/cassandra/commit/d74a7a9d) Prepare for release v0.17.0-rc.0 (#73)
- [693210e7](https://github.com/kubedb/cassandra/commit/693210e7) Add Sharding Facility for Ops-Request (#70)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.11.0-rc.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.11.0-rc.0)

- [7121c38b](https://github.com/kubedb/cassandra-medusa-plugin/commit/7121c38b) Prepare for release v0.11.0-rc.0 (#31)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.64.0-rc.0](https://github.com/kubedb/cli/releases/tag/v0.64.0-rc.0)

- [35686912](https://github.com/kubedb/cli/commit/35686912c) Prepare for release v0.64.0-rc.0 (#818)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.19.0-rc.0](https://github.com/kubedb/clickhouse/releases/tag/v0.19.0-rc.0)

- [764c306d](https://github.com/kubedb/clickhouse/commit/764c306d) Prepare for release v0.19.0-rc.0 (#97)
- [cb3c3c9f](https://github.com/kubedb/clickhouse/commit/cb3c3c9f) Add shard ops support (#84)
- [124f5124](https://github.com/kubedb/clickhouse/commit/124f5124) Update Cluster Health Check (#95)
- [ed130007](https://github.com/kubedb/clickhouse/commit/ed130007) Ops-Req fix (#92)
- [fa3541ab](https://github.com/kubedb/clickhouse/commit/fa3541ab) Fix multiple ops request in progressing state (#88)



## [kubedb/clickhouse-backup-plugin](https://github.com/kubedb/clickhouse-backup-plugin)

### [v0.1.0-rc.0](https://github.com/kubedb/clickhouse-backup-plugin/releases/tag/v0.1.0-rc.0)

- [b51a354e](https://github.com/kubedb/clickhouse-backup-plugin/commit/b51a354e) Prepare for release v0.1.0-rc.0 (#13)
- [002d44d3](https://github.com/kubedb/clickhouse-backup-plugin/commit/002d44d3) Update component id
- [54251996](https://github.com/kubedb/clickhouse-backup-plugin/commit/54251996) update deps and remove wrong replace (#12)
- [8aea7054](https://github.com/kubedb/clickhouse-backup-plugin/commit/8aea7054) Use debin base image (#11)
- [01c33fb7](https://github.com/kubedb/clickhouse-backup-plugin/commit/01c33fb7) Add ClickHouse Backup with native backup CLI
- [8f3469e3](https://github.com/kubedb/clickhouse-backup-plugin/commit/8f3469e3) Fix makefile indentation (#5)
- [d5cd3ff9](https://github.com/kubedb/clickhouse-backup-plugin/commit/d5cd3ff9) Publish Image for Redhat software certification (#1)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.19.0-rc.0](https://github.com/kubedb/crd-manager/releases/tag/v0.19.0-rc.0)

- [cbedf109](https://github.com/kubedb/crd-manager/commit/cbedf109) Prepare for release v0.19.0-rc.0 (#125)
- [ba0b1210](https://github.com/kubedb/crd-manager/commit/ba0b1210) Add documentdb (#115)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.22.0-rc.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.22.0-rc.0)

- [157f4431](https://github.com/kubedb/dashboard-restic-plugin/commit/157f4431) Prepare for release v0.22.0-rc.0 (#68)
- [7069528b](https://github.com/kubedb/dashboard-restic-plugin/commit/7069528b) Incorporate changes for the AWS credless feature (#67)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.19.0-rc.0](https://github.com/kubedb/db-client-go/releases/tag/v0.19.0-rc.0)

- [f3e2d74c](https://github.com/kubedb/db-client-go/commit/f3e2d74c) Prepare for release v0.19.0-rc.0 (#235)
- [2bd0b01b](https://github.com/kubedb/db-client-go/commit/2bd0b01b) Add Neo4j Ops Req Function (#229)
- [d32ab714](https://github.com/kubedb/db-client-go/commit/d32ab714) Qdrant HTTP client moved from apimachinery (#231)



## [kubedb/db2](https://github.com/kubedb/db2)

### [v0.5.0-rc.0](https://github.com/kubedb/db2/releases/tag/v0.5.0-rc.0)

- [1c894a64](https://github.com/kubedb/db2/commit/1c894a64) Prepare for release v0.5.0-rc.0 (#20)
- [bf277e9a](https://github.com/kubedb/db2/commit/bf277e9a) fix predicate (#19)



## [kubedb/db2-coordinator](https://github.com/kubedb/db2-coordinator)

### [v0.5.0-rc.0](https://github.com/kubedb/db2-coordinator/releases/tag/v0.5.0-rc.0)




## [kubedb/documentdb](https://github.com/kubedb/documentdb)

### [v0.1.0-rc.0](https://github.com/kubedb/documentdb/releases/tag/v0.1.0-rc.0)

- [61eb614](https://github.com/kubedb/documentdb/commit/61eb614) Prepare for release v0.1.0-rc.0 (#6)
- [876a2eb](https://github.com/kubedb/documentdb/commit/876a2eb) Support documentdb standalone (#5)
- [f68930a](https://github.com/kubedb/documentdb/commit/f68930a) oprator ready to test



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.19.0-rc.0](https://github.com/kubedb/druid/releases/tag/v0.19.0-rc.0)

- [0c9f169a](https://github.com/kubedb/druid/commit/0c9f169a) Prepare for release v0.19.0-rc.0 (#125)
- [323de9c4](https://github.com/kubedb/druid/commit/323de9c4) Add sharding facility for ops-request (#122)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.64.0-rc.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.64.0-rc.0)

- [ab8ad887](https://github.com/kubedb/elasticsearch/commit/ab8ad8879) Prepare for release v0.64.0-rc.0 (#803)
- [41039a44](https://github.com/kubedb/elasticsearch/commit/41039a44a) Add Ops manager sharding (#791)
- [3ab32aaf](https://github.com/kubedb/elasticsearch/commit/3ab32aaf2) Increase Ops Parallel Timeout (#799)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.27.0-rc.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.27.0-rc.0)

- [cd3259d0](https://github.com/kubedb/elasticsearch-restic-plugin/commit/cd3259d0) Prepare for release v0.27.0-rc.0 (#91)
- [6bbe9c64](https://github.com/kubedb/elasticsearch-restic-plugin/commit/6bbe9c64) Incorporate changes for the AWS credless feature (#90)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.19.0-rc.0](https://github.com/kubedb/ferretdb/releases/tag/v0.19.0-rc.0)

- [f3d696bc](https://github.com/kubedb/ferretdb/commit/f3d696bc) Prepare for release v0.19.0-rc.0 (#113)
- [a21800b2](https://github.com/kubedb/ferretdb/commit/a21800b2) Add sharding facility for Ops-Requests (#102)
- [e8b49242](https://github.com/kubedb/ferretdb/commit/e8b49242) Ops-Req Fix (#109)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.12.0-rc.0](https://github.com/kubedb/gitops/releases/tag/v0.12.0-rc.0)

- [c0f14b75](https://github.com/kubedb/gitops/commit/c0f14b75) Prepare for release v0.12.0-rc.0 (#50)
- [15a023ae](https://github.com/kubedb/gitops/commit/15a023ae) Fix gitops (#47)



## [kubedb/hanadb](https://github.com/kubedb/hanadb)

### [v0.5.0-rc.0](https://github.com/kubedb/hanadb/releases/tag/v0.5.0-rc.0)

- [51c5be87](https://github.com/kubedb/hanadb/commit/51c5be87) Prepare for release v0.5.0-rc.0 (#28)
- [7e98f3ba](https://github.com/kubedb/hanadb/commit/7e98f3ba) fix predicate (#26)
- [d384fbbe](https://github.com/kubedb/hanadb/commit/d384fbbe) Add monitoring (#23)



## [kubedb/hanadb-coordinator](https://github.com/kubedb/hanadb-coordinator)

### [v0.4.0-rc.0](https://github.com/kubedb/hanadb-coordinator/releases/tag/v0.4.0-rc.0)

- [7e87d731](https://github.com/kubedb/hanadb-coordinator/commit/7e87d731) Prepare for release v0.4.0-rc.0 (#7)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.10.0-rc.0](https://github.com/kubedb/hazelcast/releases/tag/v0.10.0-rc.0)

- [f2fba09f](https://github.com/kubedb/hazelcast/commit/f2fba09f) Prepare for release v0.10.0-rc.0 (#38)
- [9eecbdcf](https://github.com/kubedb/hazelcast/commit/9eecbdcf) Add Sharding facility for ops-request (#35)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.11.0-rc.0](https://github.com/kubedb/ignite/releases/tag/v0.11.0-rc.0)

- [1cc1291f](https://github.com/kubedb/ignite/commit/1cc1291f) Prepare for release v0.11.0-rc.0 (#47)
- [956eae3f](https://github.com/kubedb/ignite/commit/956eae3f) Add Sharding Facility for ops-request (#44)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2026.4.13-rc.0](https://github.com/kubedb/installer/releases/tag/v2026.4.13-rc.0)

- [4732c499](https://github.com/kubedb/installer/commit/4732c499d) Prepare for release v2026.4.13-rc.0 (#2249)
- [978ec3fc](https://github.com/kubedb/installer/commit/978ec3fc3) Use LOCALBIN for downloading all binaries if needed (#2248)
- [a73ae56d](https://github.com/kubedb/installer/commit/a73ae56d0) Update mongo wal-g version to v2026.3.30 (#2243)
- [dbde57b1](https://github.com/kubedb/installer/commit/dbde57b1a) Add ClickHouse Backup (#2209)
- [d36570c0](https://github.com/kubedb/installer/commit/d36570c07) Add documentdb (#2124)
- [cd423cca](https://github.com/kubedb/installer/commit/cd423cca3) Update crds for kubedb/apimachinery@d64cb9b9 (#2242)
- [8ed118d0](https://github.com/kubedb/installer/commit/8ed118d0c) Add New Druid Version (#2221)
- [94aa6276](https://github.com/kubedb/installer/commit/94aa6276e) Add kubedb-autoscaler sharding, Update mssql exporter image (#2227)
- [2f85da66](https://github.com/kubedb/installer/commit/2f85da66f) Add Qdrant Ops Validating Webhook (#2224)
- [bb1f496b](https://github.com/kubedb/installer/commit/bb1f496be) Add HanaDB monitoring resources (#2156)
- [38e2f250](https://github.com/kubedb/installer/commit/38e2f250f) Add ClickHouse Version 26.2.6 (#2214)
- [7b5012cb](https://github.com/kubedb/installer/commit/7b5012cb3) Add cassandraversions (#2220)
- [15f5031d](https://github.com/kubedb/installer/commit/15f5031d7) Update MySQL Init Image Tag (#2233)
- [585ea550](https://github.com/kubedb/installer/commit/585ea5505) Update ui chart tags (#2203)
- [d6559e1b](https://github.com/kubedb/installer/commit/d6559e1b7) Update cve report (#2229)
- [9a0c34e0](https://github.com/kubedb/installer/commit/9a0c34e0b) Update crds for kubedb/apimachinery@d700721a (#2216)
- [33e7cc70](https://github.com/kubedb/installer/commit/33e7cc704) Update cve report (#2228)
- [84f55c85](https://github.com/kubedb/installer/commit/84f55c854) Update cve report (#2226)
- [36c6c750](https://github.com/kubedb/installer/commit/36c6c7505) Update cve report (#2225)
- [532618b4](https://github.com/kubedb/installer/commit/532618b48) Update cve report (#2223)
- [603a4f2e](https://github.com/kubedb/installer/commit/603a4f2ec) Update crds for kubedb/apimachinery@f57d49cd (#2222)
- [819fe55f](https://github.com/kubedb/installer/commit/819fe55fb) Update crds for kubedb/apimachinery@08ca9f42 (#2218)
- [7ccf7a86](https://github.com/kubedb/installer/commit/7ccf7a866) Update crds for kubedb/apimachinery@f129d2fc (#2217)
- [1fe988db](https://github.com/kubedb/installer/commit/1fe988db4) Add Milvus Kubedb-Metrics (#2154)
- [83fa4f38](https://github.com/kubedb/installer/commit/83fa4f388) Update cve report (#2219)
- [617449a2](https://github.com/kubedb/installer/commit/617449a20) Fix MySQL Versions (#2210)
- [3770b140](https://github.com/kubedb/installer/commit/3770b1402) Add multiple shardkeys for proxysql support (#2215)
- [919ef65e](https://github.com/kubedb/installer/commit/919ef65e7) Update cve report (#2213)
- [af4bc358](https://github.com/kubedb/installer/commit/af4bc358e) Update crds for kubedb/apimachinery@0fc180f7 (#2208)
- [83e52559](https://github.com/kubedb/installer/commit/83e52559e) Update crds for kubedb/apimachinery@b790503c (#2212)
- [089b4c7d](https://github.com/kubedb/installer/commit/089b4c7d0) Update cve report (#2211)
- [23a8d0f7](https://github.com/kubedb/installer/commit/23a8d0f78) Update cve report (#2207)
- [49d47374](https://github.com/kubedb/installer/commit/49d473745) Update cve report (#2206)
- [23fc6506](https://github.com/kubedb/installer/commit/23fc65068) Update cve report (#2204)
- [830a9915](https://github.com/kubedb/installer/commit/830a99154) Update cve report (#2202)
- [e525d00d](https://github.com/kubedb/installer/commit/e525d00d4) Update cve report (#2201)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.35.0-rc.0](https://github.com/kubedb/kafka/releases/tag/v0.35.0-rc.0)

- [75023940](https://github.com/kubedb/kafka/commit/75023940) Prepare for release v0.35.0-rc.0 (#188)
- [bacc2f8b](https://github.com/kubedb/kafka/commit/bacc2f8b) Add Sharding Facility for Ops-Request (#185)
- [4cbe05d9](https://github.com/kubedb/kafka/commit/4cbe05d9) Increase Ops Parallel Timeout (#183)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.40.0-rc.0](https://github.com/kubedb/kibana/releases/tag/v0.40.0-rc.0)

- [dc91ad31](https://github.com/kubedb/kibana/commit/dc91ad31) Prepare for release v0.40.0-rc.0 (#174)
- [f37fab91](https://github.com/kubedb/kibana/commit/f37fab91) fix predicate (#173)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.27.0-rc.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.27.0-rc.0)

- [d21843b1](https://github.com/kubedb/kubedb-manifest-plugin/commit/d21843b1) Prepare for release v0.27.0-rc.0 (#123)
- [14870c71](https://github.com/kubedb/kubedb-manifest-plugin/commit/14870c71) Incorporate changes for the AWS credless feature (#122)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.15.0-rc.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.15.0-rc.0)

- [3549908d](https://github.com/kubedb/kubedb-verifier/commit/3549908d) Prepare for release v0.15.0-rc.0 (#44)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.48.0-rc.0](https://github.com/kubedb/mariadb/releases/tag/v0.48.0-rc.0)

- [5d5c0b7d](https://github.com/kubedb/mariadb/commit/5d5c0b7d8) Prepare for release v0.48.0-rc.0 (#388)
- [263e8d00](https://github.com/kubedb/mariadb/commit/263e8d005) added env for binlog cleanup (#385)
- [59deb42a](https://github.com/kubedb/mariadb/commit/59deb42af) Add Binlog File Prefix env (#384)
- [d8205512](https://github.com/kubedb/mariadb/commit/d8205512a) Fix MariaDB Standalone Volume Expansion (#382)
- [3baa4a87](https://github.com/kubedb/mariadb/commit/3baa4a87d) Add sharding facility for Ops-Request (#371)
- [7fd16637](https://github.com/kubedb/mariadb/commit/7fd16637e) Increase timeout period for Mariadb ops (#380)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.24.0-rc.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.24.0-rc.0)

- [11b821ab](https://github.com/kubedb/mariadb-archiver/commit/11b821ab) Prepare for release v0.24.0-rc.0 (#86)
- [f34db6cb](https://github.com/kubedb/mariadb-archiver/commit/f34db6cb) added binlog cleanup feature (#84)
- [f7972591](https://github.com/kubedb/mariadb-archiver/commit/f7972591) Update binlog file name (#83)
- [185293f2](https://github.com/kubedb/mariadb-archiver/commit/185293f2) Keep old log stats when sidekick pod restarts (#82)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.44.0-rc.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.44.0-rc.0)

- [53447a78](https://github.com/kubedb/mariadb-coordinator/commit/53447a78) Prepare for release v0.44.0-rc.0 (#171)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.24.0-rc.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.24.0-rc.0)

- [925ae43a](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/925ae43a) Prepare for release v0.24.0-rc.0 (#72)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.22.0-rc.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.22.0-rc.0)

- [1bfa01dd](https://github.com/kubedb/mariadb-restic-plugin/commit/1bfa01dd) Prepare for release v0.22.0-rc.0 (#81)
- [9faf6576](https://github.com/kubedb/mariadb-restic-plugin/commit/9faf6576) Incorporate changes for the AWS credless feature (#80)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.57.0-rc.0](https://github.com/kubedb/memcached/releases/tag/v0.57.0-rc.0)

- [78caf0db](https://github.com/kubedb/memcached/commit/78caf0db1) Prepare for release v0.57.0-rc.0 (#533)
- [bb514094](https://github.com/kubedb/memcached/commit/bb5140941) Shard ops support (#530)



## [kubedb/migrator-cli](https://github.com/kubedb/migrator-cli)

### [v0.4.0-rc.0](https://github.com/kubedb/migrator-cli/releases/tag/v0.4.0-rc.0)

- [7954e97](https://github.com/kubedb/migrator-cli/commit/7954e97) Prepare for release v0.4.0-rc.0 (#13)



## [kubedb/migrator-operator](https://github.com/kubedb/migrator-operator)

### [v0.4.0-rc.0](https://github.com/kubedb/migrator-operator/releases/tag/v0.4.0-rc.0)

- [1e9351f](https://github.com/kubedb/migrator-operator/commit/1e9351f) Prepare for release v0.4.0-rc.0 (#11)



## [kubedb/milvus](https://github.com/kubedb/milvus)

### [v0.5.0-rc.0](https://github.com/kubedb/milvus/releases/tag/v0.5.0-rc.0)

- [e09593ae](https://github.com/kubedb/milvus/commit/e09593ae) Prepare for release v0.5.0-rc.0 (#29)
- [12334ec8](https://github.com/kubedb/milvus/commit/12334ec8) Update predicate funcs (#28)
- [408ee09d](https://github.com/kubedb/milvus/commit/408ee09d) Sync with Apimachinery (#26)
- [00be19c5](https://github.com/kubedb/milvus/commit/00be19c5) Add Milvus Monitoring (#15)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.57.0-rc.0](https://github.com/kubedb/mongodb/releases/tag/v0.57.0-rc.0)

- [8d3533f2](https://github.com/kubedb/mongodb/commit/8d3533f2f) Prepare for release v0.57.0-rc.0 (#750)
- [4eb07bb4](https://github.com/kubedb/mongodb/commit/4eb07bb40) Fix repo name & oplog-restore job name (#748)
- [5ba17db6](https://github.com/kubedb/mongodb/commit/5ba17db61) Add Ops manager sharding (#737)
- [5234f14d](https://github.com/kubedb/mongodb/commit/5234f14d0) Increase Ops Parallel Timeout (#745)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.25.0-rc.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.25.0-rc.0)

- [009ec633](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/009ec633) Prepare for release v0.25.0-rc.0 (#76)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.27.0-rc.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.27.0-rc.0)

- [82226b2e](https://github.com/kubedb/mongodb-restic-plugin/commit/82226b2e) Prepare for release v0.27.0-rc.0 (#119)
- [9722d2ee](https://github.com/kubedb/mongodb-restic-plugin/commit/9722d2ee) Incorporate changes for the AWS credless feature (#113)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.19.0-rc.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.19.0-rc.0)

- [5856f18c](https://github.com/kubedb/mssql-coordinator/commit/5856f18c) Prepare for release v0.19.0-rc.0 (#62)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.19.0-rc.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.19.0-rc.0)

- [d997b056](https://github.com/kubedb/mssqlserver/commit/d997b056) Prepare for release v0.19.0-rc.0 (#123)
- [a72de37f](https://github.com/kubedb/mssqlserver/commit/a72de37f) Fix error handling for NewMSSQLServerReconcileState() (#122)
- [704443fb](https://github.com/kubedb/mssqlserver/commit/704443fb) Add Shard ops Support (#117)
- [ad7d0b01](https://github.com/kubedb/mssqlserver/commit/ad7d0b01) Add vertical scaling support for coordinator, exporter, arbiter components (#118)
- [615b8c1a](https://github.com/kubedb/mssqlserver/commit/615b8c1a) Reconcile DB while refered archiver update (#120)
- [8656b80f](https://github.com/kubedb/mssqlserver/commit/8656b80f) Fix multiple Ops Request Goes in Progressing issue (#115)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.18.0-rc.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.18.0-rc.0)




## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.18.0-rc.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.18.0-rc.0)

- [028bad5](https://github.com/kubedb/mssqlserver-walg-plugin/commit/028bad5) Prepare for release v0.18.0-rc.0 (#52)
- [911601f](https://github.com/kubedb/mssqlserver-walg-plugin/commit/911601f) Register scheme for endpointslice
- [fb3d7bd](https://github.com/kubedb/mssqlserver-walg-plugin/commit/fb3d7bd) Register scheme for endpointslice



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.57.0-rc.0](https://github.com/kubedb/mysql/releases/tag/v0.57.0-rc.0)

- [34a01ae2](https://github.com/kubedb/mysql/commit/34a01ae21) Prepare for release v0.57.0-rc.0 (#739)
- [bfa85452](https://github.com/kubedb/mysql/commit/bfa854526) Ensure cloud annotations to SA before sidekick creation (#736)
- [a003be8d](https://github.com/kubedb/mysql/commit/a003be8d3) Chaos Test: Update Health Check Query, Run Parallel Timeout Log (#735)
- [74eae52f](https://github.com/kubedb/mysql/commit/74eae52fd) Fix Azure Provider awsCAExist Panic (#734)
- [cb2e10e8](https://github.com/kubedb/mysql/commit/cb2e10e8e) Add sharding facility for Ops-Request (#723)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.25.0-rc.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.25.0-rc.0)

- [a98f9c0a](https://github.com/kubedb/mysql-archiver/commit/a98f9c0a) Prepare for release v0.25.0-rc.0 (#97)
- [1fbae6ea](https://github.com/kubedb/mysql-archiver/commit/1fbae6ea) Update Wal-G Version to v2026.3.30 (#95)
- [0cf2431c](https://github.com/kubedb/mysql-archiver/commit/0cf2431c) Keep old log stats when sidekick pod restarts (#91)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.42.0-rc.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.42.0-rc.0)

- [e0d65d33](https://github.com/kubedb/mysql-coordinator/commit/e0d65d33) Prepare for release v0.42.0-rc.0 (#172)
- [e5924b69](https://github.com/kubedb/mysql-coordinator/commit/e5924b69) Chaos Test: Add Network Partition, Auto Heal Support



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.25.0-rc.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.25.0-rc.0)

- [cd90f512](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/cd90f512) Prepare for release v0.25.0-rc.0 (#73)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.27.0-rc.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.27.0-rc.0)

- [09cda23f](https://github.com/kubedb/mysql-restic-plugin/commit/09cda23f) Prepare for release v0.27.0-rc.0 (#103)
- [888a4720](https://github.com/kubedb/mysql-restic-plugin/commit/888a4720) Incorporate changes for the AWS credless feature (#102)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.42.0-rc.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.42.0-rc.0)




## [kubedb/neo4j](https://github.com/kubedb/neo4j)

### [v0.5.0-rc.0](https://github.com/kubedb/neo4j/releases/tag/v0.5.0-rc.0)

- [9e097389](https://github.com/kubedb/neo4j/commit/9e097389) Prepare for release v0.5.0-rc.0 (#27)
- [2adf9978](https://github.com/kubedb/neo4j/commit/2adf9978) Add Neo4j Ops req (#22)
- [176ddf86](https://github.com/kubedb/neo4j/commit/176ddf86) Add Sharding Facility for Ops-Request (#23)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.51.0-rc.0](https://github.com/kubedb/ops-manager/releases/tag/v0.51.0-rc.0)




## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.10.0-rc.0](https://github.com/kubedb/oracle/releases/tag/v0.10.0-rc.0)

- [9fead433](https://github.com/kubedb/oracle/commit/9fead433) Prepare for release v0.10.0-rc.0 (#39)
- [22c38c19](https://github.com/kubedb/oracle/commit/22c38c19) fix conf-secret-creation (#38)
- [f861e5b2](https://github.com/kubedb/oracle/commit/f861e5b2) Add oracle conf (#27)
- [0c5da9d2](https://github.com/kubedb/oracle/commit/0c5da9d2) Add Sharding facility for Ops-Request (#35)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.10.0-rc.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.10.0-rc.0)

- [d2f36b4](https://github.com/kubedb/oracle-coordinator/commit/d2f36b4) Prepare for release v0.10.0-rc.0 (#28)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.51.0-rc.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.51.0-rc.0)

- [618d161c](https://github.com/kubedb/percona-xtradb/commit/618d161cb) Prepare for release v0.51.0-rc.0 (#445)
- [d1e40962](https://github.com/kubedb/percona-xtradb/commit/d1e409623) Add Sharding Facility for ops-request (#442)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.37.0-rc.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.37.0-rc.0)

- [7262ddc8](https://github.com/kubedb/percona-xtradb-coordinator/commit/7262ddc8) Prepare for release v0.37.0-rc.0 (#121)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.48.0-rc.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.48.0-rc.0)

- [2f9085df](https://github.com/kubedb/pg-coordinator/commit/2f9085df) Prepare for release v0.48.0-rc.0 (#241)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.51.0-rc.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.51.0-rc.0)

- [24c3141a](https://github.com/kubedb/pgbouncer/commit/24c3141a4) Prepare for release v0.51.0-rc.0 (#406)
- [cd3c35d3](https://github.com/kubedb/pgbouncer/commit/cd3c35d33) Add Sharding Facility for Ops-Request (#403)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.19.0-rc.0](https://github.com/kubedb/pgpool/releases/tag/v0.19.0-rc.0)

- [cedd0f8e](https://github.com/kubedb/pgpool/commit/cedd0f8e) Prepare for release v0.19.0-rc.0 (#114)
- [05502740](https://github.com/kubedb/pgpool/commit/05502740) Add Sharding Facility for Ops-Request (#111)
- [3a3dbca1](https://github.com/kubedb/pgpool/commit/3a3dbca1) Add Pgpool Load Balancing Support & Reconfigure Bug Fix (#109)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.64.0-rc.0](https://github.com/kubedb/postgres/releases/tag/v0.64.0-rc.0)

- [433bacda](https://github.com/kubedb/postgres/commit/433bacdae) Prepare for release v0.64.0-rc.0 (#875)
- [17ccf004](https://github.com/kubedb/postgres/commit/17ccf004f) Fix archiver restore for VS driver (#874)
- [7c718f65](https://github.com/kubedb/postgres/commit/7c718f65c) Add sidekick credless support for s3 provider (#872)
- [a7196def](https://github.com/kubedb/postgres/commit/a7196def6) fix read replica halt issue (#869)
- [15041daf](https://github.com/kubedb/postgres/commit/15041daff) Ignore Read Replica While Moving To HA (#868)
- [e61a0166](https://github.com/kubedb/postgres/commit/e61a01666) Change label patch procedure for database pods (#866)
- [8b171999](https://github.com/kubedb/postgres/commit/8b171999b) Fix Standalone to HA Scaling (#865)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.25.0-rc.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.25.0-rc.0)

- [2036b224](https://github.com/kubedb/postgres-archiver/commit/2036b224) Prepare for release v0.25.0-rc.0 (#100)
- [943a8921](https://github.com/kubedb/postgres-archiver/commit/943a8921) Bump github.com/aws/aws-sdk-go-v2/service/s3 from 1.78.2 to 1.97.3 (#97)
- [f8f0f633](https://github.com/kubedb/postgres-archiver/commit/f8f0f633) Bump go.opentelemetry.io/otel/sdk from 1.40.0 to 1.43.0 (#99)
- [b01bd9c9](https://github.com/kubedb/postgres-archiver/commit/b01bd9c9) update Go version and wal-g new version (#98)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.25.0-rc.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.25.0-rc.0)

- [b0dd01e1](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/b0dd01e1) Prepare for release v0.25.0-rc.0 (#83)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.27.0-rc.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.27.0-rc.0)

- [26862c3b](https://github.com/kubedb/postgres-restic-plugin/commit/26862c3b) Prepare for release v0.27.0-rc.0 (#102)
- [d5c1805f](https://github.com/kubedb/postgres-restic-plugin/commit/d5c1805f) Incorporate changes for the AWS credless feature (#101)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.25.0-rc.0](https://github.com/kubedb/provider-aws/releases/tag/v0.25.0-rc.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.25.0-rc.0](https://github.com/kubedb/provider-azure/releases/tag/v0.25.0-rc.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.25.0-rc.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.25.0-rc.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.64.0-rc.0](https://github.com/kubedb/provisioner/releases/tag/v0.64.0-rc.0)




## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.51.0-rc.0](https://github.com/kubedb/proxysql/releases/tag/v0.51.0-rc.0)

- [47946ab8](https://github.com/kubedb/proxysql/commit/47946ab8c) Prepare for release v0.51.0-rc.0 (#426)
- [c8839200](https://github.com/kubedb/proxysql/commit/c8839200c) Add Sharding Facility for Ops-Request (#423)
- [ff797db4](https://github.com/kubedb/proxysql/commit/ff797db40) Fix multiple Ops Request Goes in Progressing issue (#421)



## [kubedb/qdrant](https://github.com/kubedb/qdrant)

### [v0.5.0-rc.0](https://github.com/kubedb/qdrant/releases/tag/v0.5.0-rc.0)

- [b6158e90](https://github.com/kubedb/qdrant/commit/b6158e90) Prepare for release v0.5.0-rc.0 (#31)
- [64feaac8](https://github.com/kubedb/qdrant/commit/64feaac8) update deps (#30)
- [a7446a3a](https://github.com/kubedb/qdrant/commit/a7446a3a) Add Ops Request Support (#28)
- [b547a2bb](https://github.com/kubedb/qdrant/commit/b547a2bb) Add Sharding Facility for Ops-Request (#25)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.19.0-rc.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.19.0-rc.0)

- [8b47cc9f](https://github.com/kubedb/rabbitmq/commit/8b47cc9f) Prepare for release v0.19.0-rc.0 (#125)
- [34c289d5](https://github.com/kubedb/rabbitmq/commit/34c289d5) Add Sharding Facility for Ops-Request (#122)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.57.0-rc.0](https://github.com/kubedb/redis/releases/tag/v0.57.0-rc.0)

- [2df711f4](https://github.com/kubedb/redis/commit/2df711f4a) Prepare for release v0.57.0-rc.0 (#634)
- [7d51add0](https://github.com/kubedb/redis/commit/7d51add0f) Add Ops manager sharding (#622)
- [1cdb85c4](https://github.com/kubedb/redis/commit/1cdb85c49) Increase Ops Parallel Timeout (#630)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.43.0-rc.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.43.0-rc.0)

- [01f86fc3](https://github.com/kubedb/redis-coordinator/commit/01f86fc3) Prepare for release v0.43.0-rc.0 (#154)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.27.0-rc.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.27.0-rc.0)

- [52cc7c5f](https://github.com/kubedb/redis-restic-plugin/commit/52cc7c5f) Prepare for release v0.27.0-rc.0 (#98)
- [fd1a9ec0](https://github.com/kubedb/redis-restic-plugin/commit/fd1a9ec0) Incorporate changes for the AWS credless feature (#97)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.51.0-rc.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.51.0-rc.0)

- [c00452fb](https://github.com/kubedb/replication-mode-detector/commit/c00452fb) Prepare for release v0.51.0-rc.0 (#317)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.40.0-rc.0](https://github.com/kubedb/schema-manager/releases/tag/v0.40.0-rc.0)

- [cbe0c2d6](https://github.com/kubedb/schema-manager/commit/cbe0c2d6) Merge pull request #163 from kubedb/v2026.4.13-rc.0-master
- [9db522be](https://github.com/kubedb/schema-manager/commit/9db522be) Prepare for release v0.40.0-rc.0



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.19.0-rc.0](https://github.com/kubedb/singlestore/releases/tag/v0.19.0-rc.0)

- [e5334dc3](https://github.com/kubedb/singlestore/commit/e5334dc3) Prepare for release v0.19.0-rc.0 (#112)
- [9fb196da](https://github.com/kubedb/singlestore/commit/9fb196da) Add Sharding Facility for Ops-Request (#109)
- [533a210a](https://github.com/kubedb/singlestore/commit/533a210a) Fix ops kind (#107)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.19.0-rc.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.19.0-rc.0)

- [7c86dbb2](https://github.com/kubedb/singlestore-coordinator/commit/7c86dbb2) Prepare for release v0.19.0-rc.0 (#65)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.22.0-rc.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.22.0-rc.0)

- [8f533385](https://github.com/kubedb/singlestore-restic-plugin/commit/8f533385) Prepare for release v0.22.0-rc.0 (#77)
- [ed2bdcd7](https://github.com/kubedb/singlestore-restic-plugin/commit/ed2bdcd7) Incorporate changes for the AWS credless feature (#76)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.19.0-rc.0](https://github.com/kubedb/solr/releases/tag/v0.19.0-rc.0)

- [709184eb](https://github.com/kubedb/solr/commit/709184eb) Prepare for release v0.19.0-rc.0 (#123)
- [7fabe142](https://github.com/kubedb/solr/commit/7fabe142) Add sharding facility for Ops-Requests (#112)
- [67fe597a](https://github.com/kubedb/solr/commit/67fe597a) Ops-Req Fix (#119)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.49.0-rc.0](https://github.com/kubedb/tests/releases/tag/v0.49.0-rc.0)

- [805c200f](https://github.com/kubedb/tests/commit/805c200f3) Prepare for release v0.49.0-rc.0 (#518)
- [76f7731c](https://github.com/kubedb/tests/commit/76f7731c5) Add Postgres ReadReplica Test (#514)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.40.0-rc.0](https://github.com/kubedb/ui-server/releases/tag/v0.40.0-rc.0)

- [656c975b](https://github.com/kubedb/ui-server/commit/656c975b) Merge pull request #197 from kubedb/v2026.4.13-rc.0-master
- [d2c18c07](https://github.com/kubedb/ui-server/commit/d2c18c07) Prepare for release v0.40.0-rc.0
- [34fae698](https://github.com/kubedb/ui-server/commit/34fae698) Return full list if preset's available is empty (#196)
- [ee6e0343](https://github.com/kubedb/ui-server/commit/ee6e0343) Pgpool Configuration API fix (#194)



## [kubedb/weaviate](https://github.com/kubedb/weaviate)

### [v0.5.0-rc.0](https://github.com/kubedb/weaviate/releases/tag/v0.5.0-rc.0)

- [4faa2d55](https://github.com/kubedb/weaviate/commit/4faa2d55) Prepare for release v0.5.0-rc.0 (#26)
- [b1e926d7](https://github.com/kubedb/weaviate/commit/b1e926d7) fix predicate funcs (#25)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.40.0-rc.0](https://github.com/kubedb/webhook-server/releases/tag/v0.40.0-rc.0)




## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.12.0-rc.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.12.0-rc.0)

- [2b54d4cd](https://github.com/kubedb/xtrabackup-restic-plugin/commit/2b54d4cd) Prepare for release v0.12.0-rc.0 (#45)
- [9bd5434f](https://github.com/kubedb/xtrabackup-restic-plugin/commit/9bd5434f) Incorporate changes for the AWS credless feature (#44)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.19.0-rc.0](https://github.com/kubedb/zookeeper/releases/tag/v0.19.0-rc.0)

- [f6bc123d](https://github.com/kubedb/zookeeper/commit/f6bc123d) Prepare for release v0.19.0-rc.0 (#114)
- [11a11a14](https://github.com/kubedb/zookeeper/commit/11a11a14) Add Sharding Facility for Ops-Request (#111)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.19.0-rc.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.19.0-rc.0)

- [efcb15ba](https://github.com/kubedb/zookeeper-restic-plugin/commit/efcb15ba) Prepare for release v0.19.0-rc.0 (#61)
- [4dc926fc](https://github.com/kubedb/zookeeper-restic-plugin/commit/4dc926fc) Incorporate changes for the AWS credless feature (#60)




