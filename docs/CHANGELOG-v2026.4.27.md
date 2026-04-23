---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2026.4.27
    name: Changelog-v2026.4.27
    parent: welcome
    weight: 20260427
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.4.27/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.4.27/
---

# KubeDB v2026.4.27 (2026-04-23)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.64.0](https://github.com/kubedb/apimachinery/releases/tag/v0.64.0)

- [c3ac2f62](https://github.com/kubedb/apimachinery/commit/c3ac2f620) Use kubestash.dev/apimachinery v0.27.0 (#1660)
- [c1023291](https://github.com/kubedb/apimachinery/commit/c10232911) Redis ACL update stuck (#1658)
- [36d3d4d1](https://github.com/kubedb/apimachinery/commit/36d3d4d1a) Fix volume expansion validation issue for all databases (#1655)
- [9485c5e7](https://github.com/kubedb/apimachinery/commit/9485c5e71) Delete metrics-exporter-config secret on wipeout (#1657)
- [f4dbf42b](https://github.com/kubedb/apimachinery/commit/f4dbf42b5) Add Neo4j update version, volume Expansion Ops Api (#1643)
- [aa07ae21](https://github.com/kubedb/apimachinery/commit/aa07ae212) Configure dependabot refresh schedule (#1652)
- [22d42057](https://github.com/kubedb/apimachinery/commit/22d420571) Test against k8s 1.35 (#1651)
- [faed84a4](https://github.com/kubedb/apimachinery/commit/faed84a4c) Update documentdb short name code (#1650)
- [3a8dc30b](https://github.com/kubedb/apimachinery/commit/3a8dc30be) fix: register Ignite autoscaler (#1649)
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

### [v0.49.0](https://github.com/kubedb/autoscaler/releases/tag/v0.49.0)

- [4954e41b](https://github.com/kubedb/autoscaler/commit/4954e41b) Prepare for release v0.49.0 (#291)
- [411b251e](https://github.com/kubedb/autoscaler/commit/411b251e) Configure dependabot refresh schedule (#289)
- [fee22931](https://github.com/kubedb/autoscaler/commit/fee22931) Prepare for release v0.49.0-rc.0 (#286)
- [a24c29b8](https://github.com/kubedb/autoscaler/commit/a24c29b8) Add operator sharding support (#276)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.17.0](https://github.com/kubedb/cassandra/releases/tag/v0.17.0)

- [990b84e6](https://github.com/kubedb/cassandra/commit/990b84e6) Prepare for release v0.17.0 (#78)
- [581dd6c7](https://github.com/kubedb/cassandra/commit/581dd6c7) fix-volume-expansion-edit (#77)
- [c27307bb](https://github.com/kubedb/cassandra/commit/c27307bb) fix-volume-expansion (#76)
- [b0cc6961](https://github.com/kubedb/cassandra/commit/b0cc6961) Configure dependabot refresh schedule (#75)
- [d74a7a9d](https://github.com/kubedb/cassandra/commit/d74a7a9d) Prepare for release v0.17.0-rc.0 (#73)
- [693210e7](https://github.com/kubedb/cassandra/commit/693210e7) Add Sharding Facility for Ops-Request (#70)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.11.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.11.0)

- [99ef1c22](https://github.com/kubedb/cassandra-medusa-plugin/commit/99ef1c22) Prepare for release v0.11.0 (#35)
- [ba3a4b55](https://github.com/kubedb/cassandra-medusa-plugin/commit/ba3a4b55) Configure dependabot refresh schedule (#34)
- [c64ab302](https://github.com/kubedb/cassandra-medusa-plugin/commit/c64ab302) Configure dependabot refresh schedule (#33)
- [1c52789e](https://github.com/kubedb/cassandra-medusa-plugin/commit/1c52789e) Test against k8s 1.35 (#32)
- [7121c38b](https://github.com/kubedb/cassandra-medusa-plugin/commit/7121c38b) Prepare for release v0.11.0-rc.0 (#31)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.64.0](https://github.com/kubedb/cli/releases/tag/v0.64.0)

- [2be81847](https://github.com/kubedb/cli/commit/2be81847d) Prepare for release v0.64.0 (#823)
- [7132dd00](https://github.com/kubedb/cli/commit/7132dd00a) Configure dependabot refresh schedule (#822)
- [f718b38e](https://github.com/kubedb/cli/commit/f718b38ed) Configure dependabot refresh schedule (#821)
- [5667872c](https://github.com/kubedb/cli/commit/5667872c0) Test against k8s 1.35 (#820)
- [35686912](https://github.com/kubedb/cli/commit/35686912c) Prepare for release v0.64.0-rc.0 (#818)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.19.0](https://github.com/kubedb/clickhouse/releases/tag/v0.19.0)

- [6bc3a028](https://github.com/kubedb/clickhouse/commit/6bc3a028) Prepare for release v0.19.0 (#102)
- [49cb281b](https://github.com/kubedb/clickhouse/commit/49cb281b) Offline Volume Expansion Fix (#101)
- [723034be](https://github.com/kubedb/clickhouse/commit/723034be) Configure dependabot refresh schedule (#100)
- [764c306d](https://github.com/kubedb/clickhouse/commit/764c306d) Prepare for release v0.19.0-rc.0 (#97)
- [cb3c3c9f](https://github.com/kubedb/clickhouse/commit/cb3c3c9f) Add shard ops support (#84)
- [124f5124](https://github.com/kubedb/clickhouse/commit/124f5124) Update Cluster Health Check (#95)
- [ed130007](https://github.com/kubedb/clickhouse/commit/ed130007) Ops-Req fix (#92)
- [fa3541ab](https://github.com/kubedb/clickhouse/commit/fa3541ab) Fix multiple ops request in progressing state (#88)



## [kubedb/clickhouse-backup-plugin](https://github.com/kubedb/clickhouse-backup-plugin)

### [v0.1.0](https://github.com/kubedb/clickhouse-backup-plugin/releases/tag/v0.1.0)

- [fcd4d572](https://github.com/kubedb/clickhouse-backup-plugin/commit/fcd4d572) Prepare for release v0.1.0 (#17)
- [2401ac44](https://github.com/kubedb/clickhouse-backup-plugin/commit/2401ac44) Fix time format parsing issue (#16)
- [672e7e6c](https://github.com/kubedb/clickhouse-backup-plugin/commit/672e7e6c) Configure dependabot refresh schedule (#15)
- [532e2552](https://github.com/kubedb/clickhouse-backup-plugin/commit/532e2552) Configure dependabot refresh schedule (#14)
- [b51a354e](https://github.com/kubedb/clickhouse-backup-plugin/commit/b51a354e) Prepare for release v0.1.0-rc.0 (#13)
- [002d44d3](https://github.com/kubedb/clickhouse-backup-plugin/commit/002d44d3) Update component id
- [54251996](https://github.com/kubedb/clickhouse-backup-plugin/commit/54251996) update deps and remove wrong replace (#12)
- [8aea7054](https://github.com/kubedb/clickhouse-backup-plugin/commit/8aea7054) Use debin base image (#11)
- [01c33fb7](https://github.com/kubedb/clickhouse-backup-plugin/commit/01c33fb7) Add ClickHouse Backup with native backup CLI
- [8f3469e3](https://github.com/kubedb/clickhouse-backup-plugin/commit/8f3469e3) Fix makefile indentation (#5)
- [d5cd3ff9](https://github.com/kubedb/clickhouse-backup-plugin/commit/d5cd3ff9) Publish Image for Redhat software certification (#1)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.19.0](https://github.com/kubedb/crd-manager/releases/tag/v0.19.0)

- [d9566bb3](https://github.com/kubedb/crd-manager/commit/d9566bb3) Prepare for release v0.19.0 (#130)
- [2ed40e64](https://github.com/kubedb/crd-manager/commit/2ed40e64) Configure dependabot refresh schedule (#129)
- [9245b746](https://github.com/kubedb/crd-manager/commit/9245b746) Configure dependabot refresh schedule (#128)
- [89863915](https://github.com/kubedb/crd-manager/commit/89863915) Add missing CRDs: IgniteAutoscaler, HazelcastAutoscaler, Neo4jOpsRequest, QdrantOpsRequest (#126)
- [cbedf109](https://github.com/kubedb/crd-manager/commit/cbedf109) Prepare for release v0.19.0-rc.0 (#125)
- [ba0b1210](https://github.com/kubedb/crd-manager/commit/ba0b1210) Add documentdb (#115)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.22.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.22.0)

- [9d5f47a3](https://github.com/kubedb/dashboard-restic-plugin/commit/9d5f47a3) Prepare for release v0.22.0 (#73)
- [44ccb689](https://github.com/kubedb/dashboard-restic-plugin/commit/44ccb689) Bump RESTIC_VERSION to 0.18.1-20260421 (#72)
- [5eb81552](https://github.com/kubedb/dashboard-restic-plugin/commit/5eb81552) Configure dependabot refresh schedule (#71)
- [5d7dff71](https://github.com/kubedb/dashboard-restic-plugin/commit/5d7dff71) Configure dependabot refresh schedule (#70)
- [157f4431](https://github.com/kubedb/dashboard-restic-plugin/commit/157f4431) Prepare for release v0.22.0-rc.0 (#68)
- [7069528b](https://github.com/kubedb/dashboard-restic-plugin/commit/7069528b) Incorporate changes for the AWS credless feature (#67)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.19.0](https://github.com/kubedb/db-client-go/releases/tag/v0.19.0)

- [a169f968](https://github.com/kubedb/db-client-go/commit/a169f968) Prepare for release v0.19.0 (#238)
- [52036006](https://github.com/kubedb/db-client-go/commit/52036006) Configure dependabot refresh schedule (#237)
- [19607748](https://github.com/kubedb/db-client-go/commit/19607748) Configure dependabot refresh schedule (#236)
- [f3e2d74c](https://github.com/kubedb/db-client-go/commit/f3e2d74c) Prepare for release v0.19.0-rc.0 (#235)
- [2bd0b01b](https://github.com/kubedb/db-client-go/commit/2bd0b01b) Add Neo4j Ops Req Function (#229)
- [d32ab714](https://github.com/kubedb/db-client-go/commit/d32ab714) Qdrant HTTP client moved from apimachinery (#231)



## [kubedb/db2](https://github.com/kubedb/db2)

### [v0.5.0](https://github.com/kubedb/db2/releases/tag/v0.5.0)

- [f160bc0a](https://github.com/kubedb/db2/commit/f160bc0a) Prepare for release v0.5.0 (#22)
- [e1001311](https://github.com/kubedb/db2/commit/e1001311) Configure dependabot refresh schedule (#21)
- [1c894a64](https://github.com/kubedb/db2/commit/1c894a64) Prepare for release v0.5.0-rc.0 (#20)
- [bf277e9a](https://github.com/kubedb/db2/commit/bf277e9a) fix predicate (#19)



## [kubedb/db2-coordinator](https://github.com/kubedb/db2-coordinator)

### [v0.5.0](https://github.com/kubedb/db2-coordinator/releases/tag/v0.5.0)

- [5cbc34b](https://github.com/kubedb/db2-coordinator/commit/5cbc34b) Configure dependabot refresh schedule (#8)
- [382966b](https://github.com/kubedb/db2-coordinator/commit/382966b) Configure dependabot refresh schedule (#7)



## [kubedb/documentdb](https://github.com/kubedb/documentdb)

### [v0.1.0](https://github.com/kubedb/documentdb/releases/tag/v0.1.0)

- [35bad36](https://github.com/kubedb/documentdb/commit/35bad36) Prepare for release v0.1.0 (#10)
- [d2a7d97](https://github.com/kubedb/documentdb/commit/d2a7d97) Configure dependabot refresh schedule (#9)
- [d5aef4d](https://github.com/kubedb/documentdb/commit/d5aef4d) Configure dependabot refresh schedule (#8)
- [61eb614](https://github.com/kubedb/documentdb/commit/61eb614) Prepare for release v0.1.0-rc.0 (#6)
- [876a2eb](https://github.com/kubedb/documentdb/commit/876a2eb) Support documentdb standalone (#5)
- [f68930a](https://github.com/kubedb/documentdb/commit/f68930a) oprator ready to test



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.19.0](https://github.com/kubedb/druid/releases/tag/v0.19.0)

- [b304b260](https://github.com/kubedb/druid/commit/b304b260) Prepare for release v0.19.0 (#128)
- [67598e74](https://github.com/kubedb/druid/commit/67598e74) Fix Offline VolumeExpansion (#127)
- [8ba485de](https://github.com/kubedb/druid/commit/8ba485de) Configure dependabot refresh schedule (#126)
- [0c9f169a](https://github.com/kubedb/druid/commit/0c9f169a) Prepare for release v0.19.0-rc.0 (#125)
- [323de9c4](https://github.com/kubedb/druid/commit/323de9c4) Add sharding facility for ops-request (#122)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.64.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.64.0)

- [295aa682](https://github.com/kubedb/elasticsearch/commit/295aa682d) Prepare for release v0.64.0 (#807)
- [1d4e0ea3](https://github.com/kubedb/elasticsearch/commit/1d4e0ea30) Fix Volume Expansion Opsreq bug (#806)
- [aaa9bf10](https://github.com/kubedb/elasticsearch/commit/aaa9bf109) Configure dependabot refresh schedule (#805)
- [ab8ad887](https://github.com/kubedb/elasticsearch/commit/ab8ad8879) Prepare for release v0.64.0-rc.0 (#803)
- [41039a44](https://github.com/kubedb/elasticsearch/commit/41039a44a) Add Ops manager sharding (#791)
- [3ab32aaf](https://github.com/kubedb/elasticsearch/commit/3ab32aaf2) Increase Ops Parallel Timeout (#799)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.27.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.27.0)

- [f649744f](https://github.com/kubedb/elasticsearch-restic-plugin/commit/f649744f) Prepare for release v0.27.0 (#96)
- [43b4eb14](https://github.com/kubedb/elasticsearch-restic-plugin/commit/43b4eb14) Bump RESTIC_VERSION to 0.18.1-20260421 (#95)
- [779030ba](https://github.com/kubedb/elasticsearch-restic-plugin/commit/779030ba) Configure dependabot refresh schedule (#94)
- [b903f26d](https://github.com/kubedb/elasticsearch-restic-plugin/commit/b903f26d) Configure dependabot refresh schedule (#93)
- [cd3259d0](https://github.com/kubedb/elasticsearch-restic-plugin/commit/cd3259d0) Prepare for release v0.27.0-rc.0 (#91)
- [6bbe9c64](https://github.com/kubedb/elasticsearch-restic-plugin/commit/6bbe9c64) Incorporate changes for the AWS credless feature (#90)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.19.0](https://github.com/kubedb/ferretdb/releases/tag/v0.19.0)

- [805d066b](https://github.com/kubedb/ferretdb/commit/805d066b) Prepare for release v0.19.0 (#116)
- [6650292a](https://github.com/kubedb/ferretdb/commit/6650292a) Configure dependabot refresh schedule (#115)
- [04040043](https://github.com/kubedb/ferretdb/commit/04040043) Configure dependabot refresh schedule (#114)
- [f3d696bc](https://github.com/kubedb/ferretdb/commit/f3d696bc) Prepare for release v0.19.0-rc.0 (#113)
- [a21800b2](https://github.com/kubedb/ferretdb/commit/a21800b2) Add sharding facility for Ops-Requests (#102)
- [e8b49242](https://github.com/kubedb/ferretdb/commit/e8b49242) Ops-Req Fix (#109)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.12.0](https://github.com/kubedb/gitops/releases/tag/v0.12.0)

- [9cbab643](https://github.com/kubedb/gitops/commit/9cbab643) Prepare for release v0.12.0 (#53)
- [4447f92c](https://github.com/kubedb/gitops/commit/4447f92c) Configure dependabot refresh schedule (#52)
- [64f35282](https://github.com/kubedb/gitops/commit/64f35282) Configure dependabot refresh schedule (#51)
- [c0f14b75](https://github.com/kubedb/gitops/commit/c0f14b75) Prepare for release v0.12.0-rc.0 (#50)
- [15a023ae](https://github.com/kubedb/gitops/commit/15a023ae) Fix gitops (#47)



## [kubedb/hanadb](https://github.com/kubedb/hanadb)

### [v0.5.0](https://github.com/kubedb/hanadb/releases/tag/v0.5.0)

- [2f15f4f9](https://github.com/kubedb/hanadb/commit/2f15f4f9) Prepare for release v0.5.0 (#31)
- [6dce67f7](https://github.com/kubedb/hanadb/commit/6dce67f7) Configure dependabot refresh schedule (#30)
- [7b2be4f9](https://github.com/kubedb/hanadb/commit/7b2be4f9) Configure dependabot refresh schedule (#29)
- [51c5be87](https://github.com/kubedb/hanadb/commit/51c5be87) Prepare for release v0.5.0-rc.0 (#28)
- [7e98f3ba](https://github.com/kubedb/hanadb/commit/7e98f3ba) fix predicate (#26)
- [d384fbbe](https://github.com/kubedb/hanadb/commit/d384fbbe) Add monitoring (#23)



## [kubedb/hanadb-coordinator](https://github.com/kubedb/hanadb-coordinator)

### [v0.4.0](https://github.com/kubedb/hanadb-coordinator/releases/tag/v0.4.0)

- [831fad13](https://github.com/kubedb/hanadb-coordinator/commit/831fad13) Prepare for release v0.4.0 (#10)
- [57516959](https://github.com/kubedb/hanadb-coordinator/commit/57516959) Configure dependabot refresh schedule (#9)
- [f581db41](https://github.com/kubedb/hanadb-coordinator/commit/f581db41) Configure dependabot refresh schedule (#8)
- [7e87d731](https://github.com/kubedb/hanadb-coordinator/commit/7e87d731) Prepare for release v0.4.0-rc.0 (#7)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.10.0](https://github.com/kubedb/hazelcast/releases/tag/v0.10.0)

- [a49ffac7](https://github.com/kubedb/hazelcast/commit/a49ffac7) Prepare for release v0.10.0 (#42)
- [1bf1cda9](https://github.com/kubedb/hazelcast/commit/1bf1cda9) Fixed Offline Volume Expansion (#41)
- [d8c06262](https://github.com/kubedb/hazelcast/commit/d8c06262) Configure dependabot refresh schedule (#40)
- [d063a0ec](https://github.com/kubedb/hazelcast/commit/d063a0ec) Configure dependabot refresh schedule (#39)
- [f2fba09f](https://github.com/kubedb/hazelcast/commit/f2fba09f) Prepare for release v0.10.0-rc.0 (#38)
- [9eecbdcf](https://github.com/kubedb/hazelcast/commit/9eecbdcf) Add Sharding facility for ops-request (#35)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.11.0](https://github.com/kubedb/ignite/releases/tag/v0.11.0)

- [dc7c8a79](https://github.com/kubedb/ignite/commit/dc7c8a79) Prepare for release v0.11.0 (#50)
- [5a4750da](https://github.com/kubedb/ignite/commit/5a4750da) Offline Volume Expansion Fix (#49)
- [f6e606b1](https://github.com/kubedb/ignite/commit/f6e606b1) Configure dependabot refresh schedule (#48)
- [1cc1291f](https://github.com/kubedb/ignite/commit/1cc1291f) Prepare for release v0.11.0-rc.0 (#47)
- [956eae3f](https://github.com/kubedb/ignite/commit/956eae3f) Add Sharding Facility for ops-request (#44)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2026.4.27](https://github.com/kubedb/installer/releases/tag/v2026.4.27)

- [9f6e9cb1](https://github.com/kubedb/installer/commit/9f6e9cb17) Prepare for release v2026.4.27 (#2258)
- [5877cba9](https://github.com/kubedb/installer/commit/5877cba96) Update cve report (#2256)
- [d6ae7335](https://github.com/kubedb/installer/commit/d6ae73350) Update crds for kubedb/apimachinery@f4dbf42b (#2255)
- [062cb0e4](https://github.com/kubedb/installer/commit/062cb0e41) Configure dependabot refresh schedule (#2254)
- [2d19e16e](https://github.com/kubedb/installer/commit/2d19e16eb) Configure dependabot refresh schedule (#2253)
- [1f39e89a](https://github.com/kubedb/installer/commit/1f39e89af) Update cve report (#2252)
- [34f9e721](https://github.com/kubedb/installer/commit/34f9e7219) Update crds for kubedb/apimachinery@faed84a4 (#2251)
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

### [v0.35.0](https://github.com/kubedb/kafka/releases/tag/v0.35.0)

- [6ce1397a](https://github.com/kubedb/kafka/commit/6ce1397a) Prepare for release v0.35.0 (#191)
- [832e9e7f](https://github.com/kubedb/kafka/commit/832e9e7f) Offline Volume Expansion Fix (#190)
- [e62fe0c7](https://github.com/kubedb/kafka/commit/e62fe0c7) Configure dependabot refresh schedule (#189)
- [75023940](https://github.com/kubedb/kafka/commit/75023940) Prepare for release v0.35.0-rc.0 (#188)
- [bacc2f8b](https://github.com/kubedb/kafka/commit/bacc2f8b) Add Sharding Facility for Ops-Request (#185)
- [4cbe05d9](https://github.com/kubedb/kafka/commit/4cbe05d9) Increase Ops Parallel Timeout (#183)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.40.0](https://github.com/kubedb/kibana/releases/tag/v0.40.0)

- [a125f71e](https://github.com/kubedb/kibana/commit/a125f71e) Prepare for release v0.40.0 (#177)
- [af12545d](https://github.com/kubedb/kibana/commit/af12545d) Configure dependabot refresh schedule (#176)
- [aa8a3fa5](https://github.com/kubedb/kibana/commit/aa8a3fa5) Configure dependabot refresh schedule (#175)
- [dc91ad31](https://github.com/kubedb/kibana/commit/dc91ad31) Prepare for release v0.40.0-rc.0 (#174)
- [f37fab91](https://github.com/kubedb/kibana/commit/f37fab91) fix predicate (#173)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.27.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.27.0)

- [f07e155d](https://github.com/kubedb/kubedb-manifest-plugin/commit/f07e155d) Prepare for release v0.27.0 (#128)
- [b22aa0b9](https://github.com/kubedb/kubedb-manifest-plugin/commit/b22aa0b9) Bump RESTIC_VERSION to 0.18.1-20260421 (#127)
- [c4e0a75c](https://github.com/kubedb/kubedb-manifest-plugin/commit/c4e0a75c) Configure dependabot refresh schedule (#126)
- [554ca50b](https://github.com/kubedb/kubedb-manifest-plugin/commit/554ca50b) Configure dependabot refresh schedule (#125)
- [d21843b1](https://github.com/kubedb/kubedb-manifest-plugin/commit/d21843b1) Prepare for release v0.27.0-rc.0 (#123)
- [14870c71](https://github.com/kubedb/kubedb-manifest-plugin/commit/14870c71) Incorporate changes for the AWS credless feature (#122)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.15.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.15.0)

- [8a0602f4](https://github.com/kubedb/kubedb-verifier/commit/8a0602f4) Prepare for release v0.15.0 (#48)
- [323d90e5](https://github.com/kubedb/kubedb-verifier/commit/323d90e5) Configure dependabot refresh schedule (#47)
- [669fa6c3](https://github.com/kubedb/kubedb-verifier/commit/669fa6c3) Configure dependabot refresh schedule (#46)
- [3549908d](https://github.com/kubedb/kubedb-verifier/commit/3549908d) Prepare for release v0.15.0-rc.0 (#44)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.48.0](https://github.com/kubedb/mariadb/releases/tag/v0.48.0)

- [95c40d9d](https://github.com/kubedb/mariadb/commit/95c40d9d2) Prepare for release v0.48.0 (#393)
- [14558b15](https://github.com/kubedb/mariadb/commit/14558b159) Delete metrics-exporter-config secret (#392)
- [0dba0b88](https://github.com/kubedb/mariadb/commit/0dba0b886) Ensure cloud annotations to SA before sidekick creation (#386)
- [688a5667](https://github.com/kubedb/mariadb/commit/688a56674) Configure dependabot refresh schedule (#391)
- [5d5c0b7d](https://github.com/kubedb/mariadb/commit/5d5c0b7d8) Prepare for release v0.48.0-rc.0 (#388)
- [263e8d00](https://github.com/kubedb/mariadb/commit/263e8d005) added env for binlog cleanup (#385)
- [59deb42a](https://github.com/kubedb/mariadb/commit/59deb42af) Add Binlog File Prefix env (#384)
- [d8205512](https://github.com/kubedb/mariadb/commit/d8205512a) Fix MariaDB Standalone Volume Expansion (#382)
- [3baa4a87](https://github.com/kubedb/mariadb/commit/3baa4a87d) Add sharding facility for Ops-Request (#371)
- [7fd16637](https://github.com/kubedb/mariadb/commit/7fd16637e) Increase timeout period for Mariadb ops (#380)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.24.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.24.0)

- [13e2fce5](https://github.com/kubedb/mariadb-archiver/commit/13e2fce5) Prepare for release v0.24.0 (#88)
- [40106456](https://github.com/kubedb/mariadb-archiver/commit/40106456) Update Wal-G version for AWS credless mode (#85)
- [b3eb743f](https://github.com/kubedb/mariadb-archiver/commit/b3eb743f) Configure dependabot refresh schedule (#87)
- [11b821ab](https://github.com/kubedb/mariadb-archiver/commit/11b821ab) Prepare for release v0.24.0-rc.0 (#86)
- [f34db6cb](https://github.com/kubedb/mariadb-archiver/commit/f34db6cb) added binlog cleanup feature (#84)
- [f7972591](https://github.com/kubedb/mariadb-archiver/commit/f7972591) Update binlog file name (#83)
- [185293f2](https://github.com/kubedb/mariadb-archiver/commit/185293f2) Keep old log stats when sidekick pod restarts (#82)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.44.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.44.0)

- [8eeb0bbb](https://github.com/kubedb/mariadb-coordinator/commit/8eeb0bbb) Prepare for release v0.44.0 (#175)
- [4e9fe98d](https://github.com/kubedb/mariadb-coordinator/commit/4e9fe98d) Configure dependabot refresh schedule (#173)
- [53447a78](https://github.com/kubedb/mariadb-coordinator/commit/53447a78) Prepare for release v0.44.0-rc.0 (#171)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.24.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.24.0)

- [976621bf](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/976621bf) Prepare for release v0.24.0 (#75)
- [5c212903](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/5c212903) Configure dependabot refresh schedule (#74)
- [925ae43a](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/925ae43a) Prepare for release v0.24.0-rc.0 (#72)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.22.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.22.0)

- [be3b7385](https://github.com/kubedb/mariadb-restic-plugin/commit/be3b7385) Prepare for release v0.22.0 (#86)
- [f68e059d](https://github.com/kubedb/mariadb-restic-plugin/commit/f68e059d) Bump RESTIC_VERSION to 0.18.1-20260421 (#85)
- [14c31f3b](https://github.com/kubedb/mariadb-restic-plugin/commit/14c31f3b) Configure dependabot refresh schedule (#84)
- [1bfa01dd](https://github.com/kubedb/mariadb-restic-plugin/commit/1bfa01dd) Prepare for release v0.22.0-rc.0 (#81)
- [9faf6576](https://github.com/kubedb/mariadb-restic-plugin/commit/9faf6576) Incorporate changes for the AWS credless feature (#80)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.57.0](https://github.com/kubedb/memcached/releases/tag/v0.57.0)

- [9fb644d2](https://github.com/kubedb/memcached/commit/9fb644d26) Prepare for release v0.57.0 (#535)
- [a9f93a3c](https://github.com/kubedb/memcached/commit/a9f93a3cc) Configure dependabot refresh schedule (#534)
- [78caf0db](https://github.com/kubedb/memcached/commit/78caf0db1) Prepare for release v0.57.0-rc.0 (#533)
- [bb514094](https://github.com/kubedb/memcached/commit/bb5140941) Shard ops support (#530)



## [kubedb/migrator-cli](https://github.com/kubedb/migrator-cli)

### [v0.4.0](https://github.com/kubedb/migrator-cli/releases/tag/v0.4.0)

- [e1fea3b](https://github.com/kubedb/migrator-cli/commit/e1fea3b) Prepare for release v0.4.0 (#18)
- [f8e80b9](https://github.com/kubedb/migrator-cli/commit/f8e80b9) Configure dependabot refresh schedule (#15)
- [7954e97](https://github.com/kubedb/migrator-cli/commit/7954e97) Prepare for release v0.4.0-rc.0 (#13)



## [kubedb/migrator-operator](https://github.com/kubedb/migrator-operator)

### [v0.4.0](https://github.com/kubedb/migrator-operator/releases/tag/v0.4.0)

- [ff0cdd8](https://github.com/kubedb/migrator-operator/commit/ff0cdd8) Prepare for release v0.4.0 (#13)
- [d23a639](https://github.com/kubedb/migrator-operator/commit/d23a639) Configure dependabot refresh schedule (#12)
- [1e9351f](https://github.com/kubedb/migrator-operator/commit/1e9351f) Prepare for release v0.4.0-rc.0 (#11)



## [kubedb/milvus](https://github.com/kubedb/milvus)

### [v0.5.0](https://github.com/kubedb/milvus/releases/tag/v0.5.0)

- [649cc2ae](https://github.com/kubedb/milvus/commit/649cc2ae) Prepare for release v0.5.0 (#32)
- [03934e0b](https://github.com/kubedb/milvus/commit/03934e0b) Configure dependabot refresh schedule (#31)
- [c04453e9](https://github.com/kubedb/milvus/commit/c04453e9) Configure dependabot refresh schedule (#30)
- [e09593ae](https://github.com/kubedb/milvus/commit/e09593ae) Prepare for release v0.5.0-rc.0 (#29)
- [12334ec8](https://github.com/kubedb/milvus/commit/12334ec8) Update predicate funcs (#28)
- [408ee09d](https://github.com/kubedb/milvus/commit/408ee09d) Sync with Apimachinery (#26)
- [00be19c5](https://github.com/kubedb/milvus/commit/00be19c5) Add Milvus Monitoring (#15)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.57.0](https://github.com/kubedb/mongodb/releases/tag/v0.57.0)

- [5369fa5a](https://github.com/kubedb/mongodb/commit/5369fa5a4) Prepare for release v0.57.0 (#755)
- [5d951fbc](https://github.com/kubedb/mongodb/commit/5d951fbcb) Configure dependabot refresh schedule (#754)
- [7a809347](https://github.com/kubedb/mongodb/commit/7a8093470) Configure dependabot refresh schedule (#753)
- [8d3533f2](https://github.com/kubedb/mongodb/commit/8d3533f2f) Prepare for release v0.57.0-rc.0 (#750)
- [4eb07bb4](https://github.com/kubedb/mongodb/commit/4eb07bb40) Fix repo name & oplog-restore job name (#748)
- [5ba17db6](https://github.com/kubedb/mongodb/commit/5ba17db61) Add Ops manager sharding (#737)
- [5234f14d](https://github.com/kubedb/mongodb/commit/5234f14d0) Increase Ops Parallel Timeout (#745)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.25.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.25.0)

- [85824be3](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/85824be3) Prepare for release v0.25.0 (#79)
- [fdd5d683](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/fdd5d683) Configure dependabot refresh schedule (#78)
- [446ffe40](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/446ffe40) Configure dependabot refresh schedule (#77)
- [009ec633](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/009ec633) Prepare for release v0.25.0-rc.0 (#76)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.27.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.27.0)

- [47c763b6](https://github.com/kubedb/mongodb-restic-plugin/commit/47c763b6) Prepare for release v0.27.0 (#123)
- [4737d045](https://github.com/kubedb/mongodb-restic-plugin/commit/4737d045) Bump RESTIC_VERSION to 0.18.1-20260421 (#122)
- [d9f373dc](https://github.com/kubedb/mongodb-restic-plugin/commit/d9f373dc) Configure dependabot refresh schedule (#121)
- [af12c273](https://github.com/kubedb/mongodb-restic-plugin/commit/af12c273) Configure dependabot refresh schedule (#120)
- [82226b2e](https://github.com/kubedb/mongodb-restic-plugin/commit/82226b2e) Prepare for release v0.27.0-rc.0 (#119)
- [9722d2ee](https://github.com/kubedb/mongodb-restic-plugin/commit/9722d2ee) Incorporate changes for the AWS credless feature (#113)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.19.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.19.0)

- [3432bcf1](https://github.com/kubedb/mssql-coordinator/commit/3432bcf1) Prepare for release v0.19.0 (#66)
- [3ce7a06d](https://github.com/kubedb/mssql-coordinator/commit/3ce7a06d) Configure dependabot refresh schedule (#65)
- [af4c8acc](https://github.com/kubedb/mssql-coordinator/commit/af4c8acc) Configure dependabot refresh schedule (#64)
- [5856f18c](https://github.com/kubedb/mssql-coordinator/commit/5856f18c) Prepare for release v0.19.0-rc.0 (#62)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.19.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.19.0)

- [b90a9ea2](https://github.com/kubedb/mssqlserver/commit/b90a9ea2) Prepare for release v0.19.0 (#128)
- [4bed5e0e](https://github.com/kubedb/mssqlserver/commit/4bed5e0e) Offline Volume Expansion Fix (#127)
- [d72370a5](https://github.com/kubedb/mssqlserver/commit/d72370a5) Configure dependabot refresh schedule (#126)
- [1e84f01f](https://github.com/kubedb/mssqlserver/commit/1e84f01f) Configure dependabot refresh schedule (#125)
- [d997b056](https://github.com/kubedb/mssqlserver/commit/d997b056) Prepare for release v0.19.0-rc.0 (#123)
- [a72de37f](https://github.com/kubedb/mssqlserver/commit/a72de37f) Fix error handling for NewMSSQLServerReconcileState() (#122)
- [704443fb](https://github.com/kubedb/mssqlserver/commit/704443fb) Add Shard ops Support (#117)
- [ad7d0b01](https://github.com/kubedb/mssqlserver/commit/ad7d0b01) Add vertical scaling support for coordinator, exporter, arbiter components (#118)
- [615b8c1a](https://github.com/kubedb/mssqlserver/commit/615b8c1a) Reconcile DB while refered archiver update (#120)
- [8656b80f](https://github.com/kubedb/mssqlserver/commit/8656b80f) Fix multiple Ops Request Goes in Progressing issue (#115)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.18.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.18.0)

- [1ae0349](https://github.com/kubedb/mssqlserver-archiver/commit/1ae0349) Configure dependabot refresh schedule (#24)
- [df15bc6](https://github.com/kubedb/mssqlserver-archiver/commit/df15bc6) Configure dependabot refresh schedule (#23)



## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.18.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.18.0)

- [feef31e](https://github.com/kubedb/mssqlserver-walg-plugin/commit/feef31e) Prepare for release v0.18.0 (#55)
- [bc91aed](https://github.com/kubedb/mssqlserver-walg-plugin/commit/bc91aed) Configure dependabot refresh schedule (#54)
- [64f286f](https://github.com/kubedb/mssqlserver-walg-plugin/commit/64f286f) Configure dependabot refresh schedule (#53)
- [028bad5](https://github.com/kubedb/mssqlserver-walg-plugin/commit/028bad5) Prepare for release v0.18.0-rc.0 (#52)
- [911601f](https://github.com/kubedb/mssqlserver-walg-plugin/commit/911601f) Register scheme for endpointslice
- [fb3d7bd](https://github.com/kubedb/mssqlserver-walg-plugin/commit/fb3d7bd) Register scheme for endpointslice



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.57.0](https://github.com/kubedb/mysql/releases/tag/v0.57.0)

- [66db8c4b](https://github.com/kubedb/mysql/commit/66db8c4b8) Prepare for release v0.57.0 (#745)
- [ce857001](https://github.com/kubedb/mysql/commit/ce857001a) Delete metrics-exporter-config secret (#744)
- [9faeb2a0](https://github.com/kubedb/mysql/commit/9faeb2a08) Configure dependabot refresh schedule (#743)
- [449c5542](https://github.com/kubedb/mysql/commit/449c5542c) Configure dependabot refresh schedule (#742)
- [34a01ae2](https://github.com/kubedb/mysql/commit/34a01ae21) Prepare for release v0.57.0-rc.0 (#739)
- [bfa85452](https://github.com/kubedb/mysql/commit/bfa854526) Ensure cloud annotations to SA before sidekick creation (#736)
- [a003be8d](https://github.com/kubedb/mysql/commit/a003be8d3) Chaos Test: Update Health Check Query, Run Parallel Timeout Log (#735)
- [74eae52f](https://github.com/kubedb/mysql/commit/74eae52fd) Fix Azure Provider awsCAExist Panic (#734)
- [cb2e10e8](https://github.com/kubedb/mysql/commit/cb2e10e8e) Add sharding facility for Ops-Request (#723)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.25.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.25.0)

- [50815f7d](https://github.com/kubedb/mysql-archiver/commit/50815f7d) Prepare for release v0.25.0 (#102)
- [f6bc1ac1](https://github.com/kubedb/mysql-archiver/commit/f6bc1ac1) Configure dependabot refresh schedule (#101)
- [23d32e5d](https://github.com/kubedb/mysql-archiver/commit/23d32e5d) Configure dependabot refresh schedule (#100)
- [e87da317](https://github.com/kubedb/mysql-archiver/commit/e87da317) update go version for release ci fix (#99)
- [a98f9c0a](https://github.com/kubedb/mysql-archiver/commit/a98f9c0a) Prepare for release v0.25.0-rc.0 (#97)
- [1fbae6ea](https://github.com/kubedb/mysql-archiver/commit/1fbae6ea) Update Wal-G Version to v2026.3.30 (#95)
- [0cf2431c](https://github.com/kubedb/mysql-archiver/commit/0cf2431c) Keep old log stats when sidekick pod restarts (#91)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.42.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.42.0)

- [f094a593](https://github.com/kubedb/mysql-coordinator/commit/f094a593) Prepare for release v0.42.0 (#176)
- [4dfe4dbe](https://github.com/kubedb/mysql-coordinator/commit/4dfe4dbe) Configure dependabot refresh schedule (#175)
- [432033b2](https://github.com/kubedb/mysql-coordinator/commit/432033b2) Configure dependabot refresh schedule (#174)
- [e0d65d33](https://github.com/kubedb/mysql-coordinator/commit/e0d65d33) Prepare for release v0.42.0-rc.0 (#172)
- [e5924b69](https://github.com/kubedb/mysql-coordinator/commit/e5924b69) Chaos Test: Add Network Partition, Auto Heal Support



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.25.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.25.0)

- [8cef25be](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/8cef25be) Prepare for release v0.25.0 (#76)
- [fde57433](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/fde57433) Configure dependabot refresh schedule (#75)
- [a400c0f6](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/a400c0f6) Configure dependabot refresh schedule (#74)
- [cd90f512](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/cd90f512) Prepare for release v0.25.0-rc.0 (#73)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.27.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.27.0)

- [ba400cb7](https://github.com/kubedb/mysql-restic-plugin/commit/ba400cb7) Prepare for release v0.27.0 (#108)
- [372401ba](https://github.com/kubedb/mysql-restic-plugin/commit/372401ba) Bump RESTIC_VERSION to 0.18.1-20260421 (#107)
- [1f5e0d96](https://github.com/kubedb/mysql-restic-plugin/commit/1f5e0d96) Configure dependabot refresh schedule (#106)
- [0cf94d94](https://github.com/kubedb/mysql-restic-plugin/commit/0cf94d94) Configure dependabot refresh schedule (#105)
- [09cda23f](https://github.com/kubedb/mysql-restic-plugin/commit/09cda23f) Prepare for release v0.27.0-rc.0 (#103)
- [888a4720](https://github.com/kubedb/mysql-restic-plugin/commit/888a4720) Incorporate changes for the AWS credless feature (#102)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.42.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.42.0)

- [aacf7fa](https://github.com/kubedb/mysql-router-init/commit/aacf7fa) Configure dependabot refresh schedule (#57)
- [37d1624](https://github.com/kubedb/mysql-router-init/commit/37d1624) Configure dependabot refresh schedule (#56)



## [kubedb/neo4j](https://github.com/kubedb/neo4j)

### [v0.5.0](https://github.com/kubedb/neo4j/releases/tag/v0.5.0)

- [8a7b1696](https://github.com/kubedb/neo4j/commit/8a7b1696) Prepare for release v0.5.0 (#29)
- [8d7a6886](https://github.com/kubedb/neo4j/commit/8d7a6886) Configure dependabot refresh schedule (#28)
- [c57e8561](https://github.com/kubedb/neo4j/commit/c57e8561) Add Neo4j Ops Req  updVersion, volumeExpansion (#26)
- [9e097389](https://github.com/kubedb/neo4j/commit/9e097389) Prepare for release v0.5.0-rc.0 (#27)
- [2adf9978](https://github.com/kubedb/neo4j/commit/2adf9978) Add Neo4j Ops req (#22)
- [176ddf86](https://github.com/kubedb/neo4j/commit/176ddf86) Add Sharding Facility for Ops-Request (#23)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.51.0](https://github.com/kubedb/ops-manager/releases/tag/v0.51.0)




## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.10.0](https://github.com/kubedb/oracle/releases/tag/v0.10.0)

- [75cab6ba](https://github.com/kubedb/oracle/commit/75cab6ba) Prepare for release v0.10.0 (#41)
- [4e495dcd](https://github.com/kubedb/oracle/commit/4e495dcd) Configure dependabot refresh schedule (#40)
- [9fead433](https://github.com/kubedb/oracle/commit/9fead433) Prepare for release v0.10.0-rc.0 (#39)
- [22c38c19](https://github.com/kubedb/oracle/commit/22c38c19) fix conf-secret-creation (#38)
- [f861e5b2](https://github.com/kubedb/oracle/commit/f861e5b2) Add oracle conf (#27)
- [0c5da9d2](https://github.com/kubedb/oracle/commit/0c5da9d2) Add Sharding facility for Ops-Request (#35)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.10.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.10.0)

- [f9f4dd8](https://github.com/kubedb/oracle-coordinator/commit/f9f4dd8) Prepare for release v0.10.0 (#31)
- [b2a1a40](https://github.com/kubedb/oracle-coordinator/commit/b2a1a40) Configure dependabot refresh schedule (#30)
- [d2f36b4](https://github.com/kubedb/oracle-coordinator/commit/d2f36b4) Prepare for release v0.10.0-rc.0 (#28)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.51.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.51.0)

- [b10cd4ce](https://github.com/kubedb/percona-xtradb/commit/b10cd4ceb) Prepare for release v0.51.0 (#449)
- [84446e80](https://github.com/kubedb/percona-xtradb/commit/84446e805) Delete metrics exposter config secret (#448)
- [4285acd4](https://github.com/kubedb/percona-xtradb/commit/4285acd48) Configure dependabot refresh schedule (#447)
- [9ab6296a](https://github.com/kubedb/percona-xtradb/commit/9ab6296ae) Configure dependabot refresh schedule (#446)
- [618d161c](https://github.com/kubedb/percona-xtradb/commit/618d161cb) Prepare for release v0.51.0-rc.0 (#445)
- [d1e40962](https://github.com/kubedb/percona-xtradb/commit/d1e409623) Add Sharding Facility for ops-request (#442)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.37.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.37.0)

- [e1bf59f7](https://github.com/kubedb/percona-xtradb-coordinator/commit/e1bf59f7) Prepare for release v0.37.0 (#124)
- [9b44e907](https://github.com/kubedb/percona-xtradb-coordinator/commit/9b44e907) Configure dependabot refresh schedule (#123)
- [7262ddc8](https://github.com/kubedb/percona-xtradb-coordinator/commit/7262ddc8) Prepare for release v0.37.0-rc.0 (#121)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.48.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.48.0)

- [9b36ab39](https://github.com/kubedb/pg-coordinator/commit/9b36ab39) Prepare for release v0.48.0 (#245)
- [c9285b86](https://github.com/kubedb/pg-coordinator/commit/c9285b86) Configure dependabot refresh schedule (#244)
- [2f9085df](https://github.com/kubedb/pg-coordinator/commit/2f9085df) Prepare for release v0.48.0-rc.0 (#241)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.51.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.51.0)

- [23024d0c](https://github.com/kubedb/pgbouncer/commit/23024d0c7) Prepare for release v0.51.0 (#409)
- [7b5ce364](https://github.com/kubedb/pgbouncer/commit/7b5ce3642) Configure dependabot refresh schedule (#408)
- [24c3141a](https://github.com/kubedb/pgbouncer/commit/24c3141a4) Prepare for release v0.51.0-rc.0 (#406)
- [cd3c35d3](https://github.com/kubedb/pgbouncer/commit/cd3c35d33) Add Sharding Facility for Ops-Request (#403)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.19.0](https://github.com/kubedb/pgpool/releases/tag/v0.19.0)

- [e34a9780](https://github.com/kubedb/pgpool/commit/e34a9780) Prepare for release v0.19.0 (#117)
- [927c85f7](https://github.com/kubedb/pgpool/commit/927c85f7) Configure dependabot refresh schedule (#116)
- [cedd0f8e](https://github.com/kubedb/pgpool/commit/cedd0f8e) Prepare for release v0.19.0-rc.0 (#114)
- [05502740](https://github.com/kubedb/pgpool/commit/05502740) Add Sharding Facility for Ops-Request (#111)
- [3a3dbca1](https://github.com/kubedb/pgpool/commit/3a3dbca1) Add Pgpool Load Balancing Support & Reconfigure Bug Fix (#109)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.64.0](https://github.com/kubedb/postgres/releases/tag/v0.64.0)

- [261287ee](https://github.com/kubedb/postgres/commit/261287eee) Prepare for release v0.64.0 (#879)
- [0ab9077a](https://github.com/kubedb/postgres/commit/0ab9077ae) Configure dependabot refresh schedule (#878)
- [433bacda](https://github.com/kubedb/postgres/commit/433bacdae) Prepare for release v0.64.0-rc.0 (#875)
- [17ccf004](https://github.com/kubedb/postgres/commit/17ccf004f) Fix archiver restore for VS driver (#874)
- [7c718f65](https://github.com/kubedb/postgres/commit/7c718f65c) Add sidekick credless support for s3 provider (#872)
- [a7196def](https://github.com/kubedb/postgres/commit/a7196def6) fix read replica halt issue (#869)
- [15041daf](https://github.com/kubedb/postgres/commit/15041daff) Ignore Read Replica While Moving To HA (#868)
- [e61a0166](https://github.com/kubedb/postgres/commit/e61a01666) Change label patch procedure for database pods (#866)
- [8b171999](https://github.com/kubedb/postgres/commit/8b171999b) Fix Standalone to HA Scaling (#865)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.25.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.25.0)

- [9f8fbdc3](https://github.com/kubedb/postgres-archiver/commit/9f8fbdc3) Prepare for release v0.25.0 (#104)
- [e66cdfb3](https://github.com/kubedb/postgres-archiver/commit/e66cdfb3) Configure dependabot refresh schedule (#103)
- [3bfd085f](https://github.com/kubedb/postgres-archiver/commit/3bfd085f) Update Go version (#101)
- [2036b224](https://github.com/kubedb/postgres-archiver/commit/2036b224) Prepare for release v0.25.0-rc.0 (#100)
- [943a8921](https://github.com/kubedb/postgres-archiver/commit/943a8921) Bump github.com/aws/aws-sdk-go-v2/service/s3 from 1.78.2 to 1.97.3 (#97)
- [f8f0f633](https://github.com/kubedb/postgres-archiver/commit/f8f0f633) Bump go.opentelemetry.io/otel/sdk from 1.40.0 to 1.43.0 (#99)
- [b01bd9c9](https://github.com/kubedb/postgres-archiver/commit/b01bd9c9) update Go version and wal-g new version (#98)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.25.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.25.0)

- [a1bac671](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/a1bac671) Prepare for release v0.25.0 (#85)
- [c20cd235](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/c20cd235) Configure dependabot refresh schedule (#84)
- [b0dd01e1](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/b0dd01e1) Prepare for release v0.25.0-rc.0 (#83)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.27.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.27.0)

- [b5399717](https://github.com/kubedb/postgres-restic-plugin/commit/b5399717) Prepare for release v0.27.0 (#106)
- [ff32b63e](https://github.com/kubedb/postgres-restic-plugin/commit/ff32b63e) Bump RESTIC_VERSION to 0.18.1-20260421 (#105)
- [deeb99ee](https://github.com/kubedb/postgres-restic-plugin/commit/deeb99ee) Configure dependabot refresh schedule (#104)
- [26862c3b](https://github.com/kubedb/postgres-restic-plugin/commit/26862c3b) Prepare for release v0.27.0-rc.0 (#102)
- [d5c1805f](https://github.com/kubedb/postgres-restic-plugin/commit/d5c1805f) Incorporate changes for the AWS credless feature (#101)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.25.0](https://github.com/kubedb/provider-aws/releases/tag/v0.25.0)

- [a619973](https://github.com/kubedb/provider-aws/commit/a619973) Configure dependabot refresh schedule (#42)



## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.25.0](https://github.com/kubedb/provider-azure/releases/tag/v0.25.0)

- [2923d77](https://github.com/kubedb/provider-azure/commit/2923d77) Configure dependabot refresh schedule (#27)



## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.25.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.25.0)

- [a8dc1d6](https://github.com/kubedb/provider-gcp/commit/a8dc1d6) Configure dependabot refresh schedule (#27)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.64.0](https://github.com/kubedb/provisioner/releases/tag/v0.64.0)




## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.51.0](https://github.com/kubedb/proxysql/releases/tag/v0.51.0)

- [34985f15](https://github.com/kubedb/proxysql/commit/34985f159) Prepare for release v0.51.0 (#429)
- [32439267](https://github.com/kubedb/proxysql/commit/324392678) Configure dependabot refresh schedule (#428)
- [05c70796](https://github.com/kubedb/proxysql/commit/05c707961) Configure dependabot refresh schedule (#427)
- [47946ab8](https://github.com/kubedb/proxysql/commit/47946ab8c) Prepare for release v0.51.0-rc.0 (#426)
- [c8839200](https://github.com/kubedb/proxysql/commit/c8839200c) Add Sharding Facility for Ops-Request (#423)
- [ff797db4](https://github.com/kubedb/proxysql/commit/ff797db40) Fix multiple Ops Request Goes in Progressing issue (#421)



## [kubedb/qdrant](https://github.com/kubedb/qdrant)

### [v0.5.0](https://github.com/kubedb/qdrant/releases/tag/v0.5.0)

- [bcf12c03](https://github.com/kubedb/qdrant/commit/bcf12c03) Prepare for release v0.5.0 (#36)
- [0bc2e4d5](https://github.com/kubedb/qdrant/commit/0bc2e4d5) Bug Fix (#35)
- [783dd192](https://github.com/kubedb/qdrant/commit/783dd192) Offline Volume Expansion Fix (#34)
- [5c011201](https://github.com/kubedb/qdrant/commit/5c011201) Configure dependabot refresh schedule (#33)
- [f18cd566](https://github.com/kubedb/qdrant/commit/f18cd566) Configure dependabot refresh schedule (#32)
- [b6158e90](https://github.com/kubedb/qdrant/commit/b6158e90) Prepare for release v0.5.0-rc.0 (#31)
- [64feaac8](https://github.com/kubedb/qdrant/commit/64feaac8) update deps (#30)
- [a7446a3a](https://github.com/kubedb/qdrant/commit/a7446a3a) Add Ops Request Support (#28)
- [b547a2bb](https://github.com/kubedb/qdrant/commit/b547a2bb) Add Sharding Facility for Ops-Request (#25)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.19.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.19.0)

- [cd286bab](https://github.com/kubedb/rabbitmq/commit/cd286bab) Prepare for release v0.19.0 (#129)
- [3dff0452](https://github.com/kubedb/rabbitmq/commit/3dff0452) Fix Offline Volume Expansion (#128)
- [c80c9679](https://github.com/kubedb/rabbitmq/commit/c80c9679) Configure dependabot refresh schedule (#127)
- [8b47cc9f](https://github.com/kubedb/rabbitmq/commit/8b47cc9f) Prepare for release v0.19.0-rc.0 (#125)
- [34c289d5](https://github.com/kubedb/rabbitmq/commit/34c289d5) Add Sharding Facility for Ops-Request (#122)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.57.0](https://github.com/kubedb/redis/releases/tag/v0.57.0)

- [680eaa9e](https://github.com/kubedb/redis/commit/680eaa9ec) Prepare for release v0.57.0 (#638)
- [7d78dfd9](https://github.com/kubedb/redis/commit/7d78dfd9e) Configure dependabot refresh schedule (#637)
- [fc44407f](https://github.com/kubedb/redis/commit/fc44407fc) Configure dependabot refresh schedule (#636)
- [2df711f4](https://github.com/kubedb/redis/commit/2df711f4a) Prepare for release v0.57.0-rc.0 (#634)
- [7d51add0](https://github.com/kubedb/redis/commit/7d51add0f) Add Ops manager sharding (#622)
- [1cdb85c4](https://github.com/kubedb/redis/commit/1cdb85c49) Increase Ops Parallel Timeout (#630)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.43.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.43.0)

- [f39b5c5d](https://github.com/kubedb/redis-coordinator/commit/f39b5c5d) Prepare for release v0.43.0 (#157)
- [f508e9dc](https://github.com/kubedb/redis-coordinator/commit/f508e9dc) Configure dependabot refresh schedule (#156)
- [2edf641a](https://github.com/kubedb/redis-coordinator/commit/2edf641a) Configure dependabot refresh schedule (#155)
- [01f86fc3](https://github.com/kubedb/redis-coordinator/commit/01f86fc3) Prepare for release v0.43.0-rc.0 (#154)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.27.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.27.0)

- [df86bb19](https://github.com/kubedb/redis-restic-plugin/commit/df86bb19) Prepare for release v0.27.0 (#103)
- [ca23d967](https://github.com/kubedb/redis-restic-plugin/commit/ca23d967) Bump RESTIC_VERSION to 0.18.1-20260421 (#102)
- [b043e083](https://github.com/kubedb/redis-restic-plugin/commit/b043e083) Configure dependabot refresh schedule (#101)
- [f301cd48](https://github.com/kubedb/redis-restic-plugin/commit/f301cd48) Configure dependabot refresh schedule (#100)
- [52cc7c5f](https://github.com/kubedb/redis-restic-plugin/commit/52cc7c5f) Prepare for release v0.27.0-rc.0 (#98)
- [fd1a9ec0](https://github.com/kubedb/redis-restic-plugin/commit/fd1a9ec0) Incorporate changes for the AWS credless feature (#97)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.51.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.51.0)

- [bdc776c1](https://github.com/kubedb/replication-mode-detector/commit/bdc776c1) Prepare for release v0.51.0 (#320)
- [38d8e512](https://github.com/kubedb/replication-mode-detector/commit/38d8e512) Configure dependabot refresh schedule (#319)
- [dc13533a](https://github.com/kubedb/replication-mode-detector/commit/dc13533a) Configure dependabot refresh schedule (#318)
- [c00452fb](https://github.com/kubedb/replication-mode-detector/commit/c00452fb) Prepare for release v0.51.0-rc.0 (#317)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.40.0](https://github.com/kubedb/schema-manager/releases/tag/v0.40.0)

- [8590f8e5](https://github.com/kubedb/schema-manager/commit/8590f8e5) Prepare for release v0.40.0 (#167)
- [5a81c014](https://github.com/kubedb/schema-manager/commit/5a81c014) Configure dependabot refresh schedule (#166)
- [cedf8375](https://github.com/kubedb/schema-manager/commit/cedf8375) Configure dependabot refresh schedule (#165)
- [cbe0c2d6](https://github.com/kubedb/schema-manager/commit/cbe0c2d6) Merge pull request #163 from kubedb/v2026.4.13-rc.0-master
- [9db522be](https://github.com/kubedb/schema-manager/commit/9db522be) Prepare for release v0.40.0-rc.0



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.19.0](https://github.com/kubedb/singlestore/releases/tag/v0.19.0)

- [be212d59](https://github.com/kubedb/singlestore/commit/be212d59) Prepare for release v0.19.0 (#117)
- [38cf8233](https://github.com/kubedb/singlestore/commit/38cf8233) fix volume expansion bug (#116)
- [66594528](https://github.com/kubedb/singlestore/commit/66594528) offline volume expansion fix (#115)
- [742f9ca3](https://github.com/kubedb/singlestore/commit/742f9ca3) Configure dependabot refresh schedule (#114)
- [e5334dc3](https://github.com/kubedb/singlestore/commit/e5334dc3) Prepare for release v0.19.0-rc.0 (#112)
- [9fb196da](https://github.com/kubedb/singlestore/commit/9fb196da) Add Sharding Facility for Ops-Request (#109)
- [533a210a](https://github.com/kubedb/singlestore/commit/533a210a) Fix ops kind (#107)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.19.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.19.0)

- [89cacaf6](https://github.com/kubedb/singlestore-coordinator/commit/89cacaf6) Prepare for release v0.19.0 (#69)
- [4b0e55b2](https://github.com/kubedb/singlestore-coordinator/commit/4b0e55b2) Configure dependabot refresh schedule (#68)
- [a7242e7c](https://github.com/kubedb/singlestore-coordinator/commit/a7242e7c) Configure dependabot refresh schedule (#67)
- [7c86dbb2](https://github.com/kubedb/singlestore-coordinator/commit/7c86dbb2) Prepare for release v0.19.0-rc.0 (#65)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.22.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.22.0)

- [66fee03a](https://github.com/kubedb/singlestore-restic-plugin/commit/66fee03a) Prepare for release v0.22.0 (#82)
- [4a9f961c](https://github.com/kubedb/singlestore-restic-plugin/commit/4a9f961c) Bump RESTIC_VERSION to 0.18.1-20260421 (#81)
- [fc7ad966](https://github.com/kubedb/singlestore-restic-plugin/commit/fc7ad966) Configure dependabot refresh schedule (#80)
- [536de2a0](https://github.com/kubedb/singlestore-restic-plugin/commit/536de2a0) Configure dependabot refresh schedule (#79)
- [8f533385](https://github.com/kubedb/singlestore-restic-plugin/commit/8f533385) Prepare for release v0.22.0-rc.0 (#77)
- [ed2bdcd7](https://github.com/kubedb/singlestore-restic-plugin/commit/ed2bdcd7) Incorporate changes for the AWS credless feature (#76)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.19.0](https://github.com/kubedb/solr/releases/tag/v0.19.0)

- [d4d7b833](https://github.com/kubedb/solr/commit/d4d7b833) Prepare for release v0.19.0 (#126)
- [301b697f](https://github.com/kubedb/solr/commit/301b697f) Offline Volume Expansion Fix (#125)
- [d255112f](https://github.com/kubedb/solr/commit/d255112f) Configure dependabot refresh schedule (#124)
- [709184eb](https://github.com/kubedb/solr/commit/709184eb) Prepare for release v0.19.0-rc.0 (#123)
- [7fabe142](https://github.com/kubedb/solr/commit/7fabe142) Add sharding facility for Ops-Requests (#112)
- [67fe597a](https://github.com/kubedb/solr/commit/67fe597a) Ops-Req Fix (#119)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.49.0](https://github.com/kubedb/tests/releases/tag/v0.49.0)

- [af6fa871](https://github.com/kubedb/tests/commit/af6fa8712) Prepare for release v0.49.0 (#521)
- [7086f3e0](https://github.com/kubedb/tests/commit/7086f3e0e) Configure dependabot refresh schedule (#520)
- [dd86a543](https://github.com/kubedb/tests/commit/dd86a5437) Configure dependabot refresh schedule (#519)
- [805c200f](https://github.com/kubedb/tests/commit/805c200f3) Prepare for release v0.49.0-rc.0 (#518)
- [76f7731c](https://github.com/kubedb/tests/commit/76f7731c5) Add Postgres ReadReplica Test (#514)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.40.0](https://github.com/kubedb/ui-server/releases/tag/v0.40.0)

- [09c4dcc4](https://github.com/kubedb/ui-server/commit/09c4dcc4) Prepare for release v0.40.0 (#199)
- [996b2c98](https://github.com/kubedb/ui-server/commit/996b2c98) Configure dependabot refresh schedule (#198)
- [656c975b](https://github.com/kubedb/ui-server/commit/656c975b) Merge pull request #197 from kubedb/v2026.4.13-rc.0-master
- [d2c18c07](https://github.com/kubedb/ui-server/commit/d2c18c07) Prepare for release v0.40.0-rc.0
- [34fae698](https://github.com/kubedb/ui-server/commit/34fae698) Return full list if preset's available is empty (#196)
- [ee6e0343](https://github.com/kubedb/ui-server/commit/ee6e0343) Pgpool Configuration API fix (#194)



## [kubedb/weaviate](https://github.com/kubedb/weaviate)

### [v0.5.0](https://github.com/kubedb/weaviate/releases/tag/v0.5.0)

- [31218a35](https://github.com/kubedb/weaviate/commit/31218a35) Prepare for release v0.5.0 (#28)
- [4fc6dac2](https://github.com/kubedb/weaviate/commit/4fc6dac2) Configure dependabot refresh schedule (#27)
- [4faa2d55](https://github.com/kubedb/weaviate/commit/4faa2d55) Prepare for release v0.5.0-rc.0 (#26)
- [b1e926d7](https://github.com/kubedb/weaviate/commit/b1e926d7) fix predicate funcs (#25)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.40.0](https://github.com/kubedb/webhook-server/releases/tag/v0.40.0)




## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.12.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.12.0)

- [1c67d63e](https://github.com/kubedb/xtrabackup-restic-plugin/commit/1c67d63e) Prepare for release v0.12.0 (#49)
- [6be593b0](https://github.com/kubedb/xtrabackup-restic-plugin/commit/6be593b0) Bump RESTIC_VERSION to 0.18.1-20260421 (#48)
- [6199440e](https://github.com/kubedb/xtrabackup-restic-plugin/commit/6199440e) Configure dependabot refresh schedule (#47)
- [2b54d4cd](https://github.com/kubedb/xtrabackup-restic-plugin/commit/2b54d4cd) Prepare for release v0.12.0-rc.0 (#45)
- [9bd5434f](https://github.com/kubedb/xtrabackup-restic-plugin/commit/9bd5434f) Incorporate changes for the AWS credless feature (#44)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.19.0](https://github.com/kubedb/zookeeper/releases/tag/v0.19.0)

- [bd934f48](https://github.com/kubedb/zookeeper/commit/bd934f48) Prepare for release v0.19.0 (#118)
- [88cc9518](https://github.com/kubedb/zookeeper/commit/88cc9518) fix vol exp (#117)
- [bba031e5](https://github.com/kubedb/zookeeper/commit/bba031e5) Configure dependabot refresh schedule (#116)
- [f745f1a2](https://github.com/kubedb/zookeeper/commit/f745f1a2) Configure dependabot refresh schedule (#115)
- [f6bc123d](https://github.com/kubedb/zookeeper/commit/f6bc123d) Prepare for release v0.19.0-rc.0 (#114)
- [11a11a14](https://github.com/kubedb/zookeeper/commit/11a11a14) Add Sharding Facility for Ops-Request (#111)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.19.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.19.0)

- [9df74bd0](https://github.com/kubedb/zookeeper-restic-plugin/commit/9df74bd0) Prepare for release v0.19.0 (#66)
- [e7662aa2](https://github.com/kubedb/zookeeper-restic-plugin/commit/e7662aa2) Bump RESTIC_VERSION to 0.18.1-20260421 (#65)
- [0ce9d404](https://github.com/kubedb/zookeeper-restic-plugin/commit/0ce9d404) Configure dependabot refresh schedule (#64)
- [5e490249](https://github.com/kubedb/zookeeper-restic-plugin/commit/5e490249) Configure dependabot refresh schedule (#63)
- [efcb15ba](https://github.com/kubedb/zookeeper-restic-plugin/commit/efcb15ba) Prepare for release v0.19.0-rc.0 (#61)
- [4dc926fc](https://github.com/kubedb/zookeeper-restic-plugin/commit/4dc926fc) Incorporate changes for the AWS credless feature (#60)




