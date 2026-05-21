---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2026.5.18-rc.0
    name: Changelog-v2026.5.18-rc.0
    parent: welcome
    weight: 20260518
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.5.18-rc.0/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.5.18-rc.0/
---

# KubeDB v2026.5.18-rc.0 (2026-05-21)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.65.0-rc.0](https://github.com/kubedb/apimachinery/releases/tag/v0.65.0-rc.0)

- [05817720](https://github.com/kubedb/apimachinery/commit/05817720d) Update for release KubeStash@v2026.5.18-rc.0 (#1718)
- [c5944a1e](https://github.com/kubedb/apimachinery/commit/c5944a1e2) Fix build (#1717)
- [6fc98997](https://github.com/kubedb/apimachinery/commit/6fc989976) hanadb tls (#1646)
- [58f38fc9](https://github.com/kubedb/apimachinery/commit/58f38fc9a) Extend StorageMigration specs with topology-aware per-component fields (#1715)
- [c2dfa3a8](https://github.com/kubedb/apimachinery/commit/c2dfa3a81) added mongodb migration fields (#1656)
- [fbdb3d8e](https://github.com/kubedb/apimachinery/commit/fbdb3d8ea) Set Ops Request Options Defaults in Autoscaler (#1669)
- [dfc3c09e](https://github.com/kubedb/apimachinery/commit/dfc3c09ef) Add Qdrant Autoscaler (#1648)
- [f48b0ac8](https://github.com/kubedb/apimachinery/commit/f48b0ac89) Add Aerospike (#1635)
- [3897eb73](https://github.com/kubedb/apimachinery/commit/3897eb732) Add ClickHouse Archiver (#1653)
- [cf2a76b9](https://github.com/kubedb/apimachinery/commit/cf2a76b9d) Documentdb Clusterring (#1673)
- [198401e0](https://github.com/kubedb/apimachinery/commit/198401e00) Add HanaDBAutoscaler types (#1709)
- [6d4b3682](https://github.com/kubedb/apimachinery/commit/6d4b36828) Add RotateAuth OpsRequest type for Weaviate (#1704)
- [9adb9ccd](https://github.com/kubedb/apimachinery/commit/9adb9ccd9) Add RotateAuth OpsRequest type for Hanadb (#1703)
- [bfd939a9](https://github.com/kubedb/apimachinery/commit/bfd939a9a) Add VerticalScaling type for WeaviateOpsRequest (#1701)
- [917d39e4](https://github.com/kubedb/apimachinery/commit/917d39e4d) Add WeaviateAutoscaler types and Weaviate ops scaling support (#1711)
- [56dd7891](https://github.com/kubedb/apimachinery/commit/56dd78914) Add VerticalScaling type for HanaDBOpsRequest
- [cddf6ff8](https://github.com/kubedb/apimachinery/commit/cddf6ff86) Add VolumeExpansion API to WeaviateOpsRequest (#1699)
- [c2e158d2](https://github.com/kubedb/apimachinery/commit/c2e158d26) Add VolumeExpansion API to HanaDBOpsRequest (#1697)
- [a53cf0a1](https://github.com/kubedb/apimachinery/commit/a53cf0a15) Add oracle ops reaconfigure (#1670)
- [53210b11](https://github.com/kubedb/apimachinery/commit/53210b114) Add DocumentDBOpsRequest API types and client bindings (#1706)
- [dfdd296c](https://github.com/kubedb/apimachinery/commit/dfdd296c9) Add Neo4jAutoscaler types (#1707)
- [0083a77b](https://github.com/kubedb/apimachinery/commit/0083a77b2) Add DocumentDBAutoscaler API types (#1710)
- [562fac68](https://github.com/kubedb/apimachinery/commit/562fac689) Add StorageMigration OpsRequest type for Druid (#1713)
- [28c3244b](https://github.com/kubedb/apimachinery/commit/28c3244bb) Add StorageMigration to HanaDB ops API (#1695)
- [6610d121](https://github.com/kubedb/apimachinery/commit/6610d121c) Add StorageMigration op type for Cassandra (#1688)
- [fb019130](https://github.com/kubedb/apimachinery/commit/fb0191306) Add StorageMigration op type for ClickHouse (#1691)
- [eca9d5c2](https://github.com/kubedb/apimachinery/commit/eca9d5c2c) Add StorageMigration op type for Ignite (#1690)
- [4f3a5c8a](https://github.com/kubedb/apimachinery/commit/4f3a5c8ad) Add StorageMigration op type for Hazelcast (#1687)
- [0807f185](https://github.com/kubedb/apimachinery/commit/0807f1850) Add StorageMigration OpsRequest support for RabbitMQ (#1692)
- [d6cb56c0](https://github.com/kubedb/apimachinery/commit/d6cb56c08) Add StorageMigration OpsRequest support for Solr (#1685)
- [b0dac243](https://github.com/kubedb/apimachinery/commit/b0dac243e) Add StorageMigration OpsRequest type for MSSQLServer (#1684)
- [72fdc722](https://github.com/kubedb/apimachinery/commit/72fdc722d) Add StorageMigration OpsRequest type for Singlestore (#1683)
- [c0312b1a](https://github.com/kubedb/apimachinery/commit/c0312b1a1) Add StorageMigration OpsRequest support for Redis (#1682)
- [fee11548](https://github.com/kubedb/apimachinery/commit/fee115489) Add migration api for mysql and mariadb (#1628)
- [fdf5544c](https://github.com/kubedb/apimachinery/commit/fdf5544cf) Add StorageMigration OpsRequest support for PerconaXtraDB (#1681)
- [cf027932](https://github.com/kubedb/apimachinery/commit/cf0279326) Add StorageMigration op type for MariaDB (#1680)
- [e49a2e61](https://github.com/kubedb/apimachinery/commit/e49a2e615) Add StorageMigration OpsRequest support for Qdrant (#1686)
- [9c39cb2c](https://github.com/kubedb/apimachinery/commit/9c39cb2c4) Add StorageMigration OpsRequest support for Neo4j (#1689)
- [0dcd1c39](https://github.com/kubedb/apimachinery/commit/0dcd1c390) Add StorageMigration op type for Kafka (#1679)
- [12f28802](https://github.com/kubedb/apimachinery/commit/12f28802b) Add StorageMigration op type for Elasticsearch (#1677)
- [50082601](https://github.com/kubedb/apimachinery/commit/500826015) Add StorageMigration op type for MongoDB (#1678)
- [840ffde3](https://github.com/kubedb/apimachinery/commit/840ffde38) Add Milvus TLS Support (#1632)
- [ec2f05e6](https://github.com/kubedb/apimachinery/commit/ec2f05e6d) Add AGENTS.md for AI coding agents
- [7b1cb929](https://github.com/kubedb/apimachinery/commit/7b1cb9292) Pin git user to 1gtm in update-crds/update-docs workflows (#1675)
- [479e8aa1](https://github.com/kubedb/apimachinery/commit/479e8aa1d) Harden CI workflows (#1668)
- [0e06cafc](https://github.com/kubedb/apimachinery/commit/0e06cafc8) Reconfig Merger bug fix (#1674)
- [0f3ce7e2](https://github.com/kubedb/apimachinery/commit/0f3ce7e28) Add MySQL InnodDB variables (#1644)
- [8bc390ba](https://github.com/kubedb/apimachinery/commit/8bc390baf) Ignore secret copy for local provider (#1664)
- [014f6503](https://github.com/kubedb/apimachinery/commit/014f65032) Do not create secret for credless mode (#1663)
- [e7d97b2a](https://github.com/kubedb/apimachinery/commit/e7d97b2a5) Change default resource for Cassandra db container (#1661)
- [ba60ac23](https://github.com/kubedb/apimachinery/commit/ba60ac23d) Add utility function for storage cred secret (#1662)
- [c3eb083e](https://github.com/kubedb/apimachinery/commit/c3eb083ed) Delete metrics-exporter-config secret on wipeout for elasticsearch (#1659)
- [c3ac2f62](https://github.com/kubedb/apimachinery/commit/c3ac2f620) Use kubestash.dev/apimachinery v0.27.0 (#1660)
- [c1023291](https://github.com/kubedb/apimachinery/commit/c10232911) Redis ACL update stuck (#1658)
- [36d3d4d1](https://github.com/kubedb/apimachinery/commit/36d3d4d1a) Fix volume expansion validation issue for all databases (#1655)
- [9485c5e7](https://github.com/kubedb/apimachinery/commit/9485c5e71) Delete metrics-exporter-config secret on wipeout (#1657)
- [f4dbf42b](https://github.com/kubedb/apimachinery/commit/f4dbf42b5) Add Neo4j update version, volume Expansion Ops Api (#1643)
- [aa07ae21](https://github.com/kubedb/apimachinery/commit/aa07ae212) Configure dependabot refresh schedule (#1652)
- [22d42057](https://github.com/kubedb/apimachinery/commit/22d420571) Test against k8s 1.35 (#1651)
- [faed84a4](https://github.com/kubedb/apimachinery/commit/faed84a4c) Update documentdb short name code (#1650)
- [3a8dc30b](https://github.com/kubedb/apimachinery/commit/3a8dc30be) fix: register Ignite autoscaler (#1649)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.50.0-rc.0](https://github.com/kubedb/autoscaler/releases/tag/v0.50.0-rc.0)




## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.18.0-rc.0](https://github.com/kubedb/cassandra/releases/tag/v0.18.0-rc.0)

- [a728b7f6](https://github.com/kubedb/cassandra/commit/a728b7f6) Prepare for release v0.18.0-rc.0 (#82)
- [0363c990](https://github.com/kubedb/cassandra/commit/0363c990) Tighten CI/release workflow secrets, perms, and release notes
- [995e9f8a](https://github.com/kubedb/cassandra/commit/995e9f8a) Harden release and release-tracker workflows
- [eaa2a39b](https://github.com/kubedb/cassandra/commit/eaa2a39b) Add CLAUDE.md pointing to AGENTS.md
- [cbc64d19](https://github.com/kubedb/cassandra/commit/cbc64d19) Add AGENTS.md for AI coding agents
- [fcc9a127](https://github.com/kubedb/cassandra/commit/fcc9a127) Harden CI workflows (#79)
- [990b84e6](https://github.com/kubedb/cassandra/commit/990b84e6) Prepare for release v0.17.0 (#78)
- [581dd6c7](https://github.com/kubedb/cassandra/commit/581dd6c7) fix-volume-expansion-edit (#77)
- [c27307bb](https://github.com/kubedb/cassandra/commit/c27307bb) fix-volume-expansion (#76)
- [b0cc6961](https://github.com/kubedb/cassandra/commit/b0cc6961) Configure dependabot refresh schedule (#75)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.12.0-rc.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.12.0-rc.0)

- [29127ccd](https://github.com/kubedb/cassandra-medusa-plugin/commit/29127ccd) Prepare for release v0.12.0-rc.0 (#38)
- [972cc3ec](https://github.com/kubedb/cassandra-medusa-plugin/commit/972cc3ec) Harden release and release-tracker workflows
- [f5566ac1](https://github.com/kubedb/cassandra-medusa-plugin/commit/f5566ac1) Add AGENTS.md for AI coding agents
- [494bac88](https://github.com/kubedb/cassandra-medusa-plugin/commit/494bac88) Harden CI workflows (#36)
- [99ef1c22](https://github.com/kubedb/cassandra-medusa-plugin/commit/99ef1c22) Prepare for release v0.11.0 (#35)
- [ba3a4b55](https://github.com/kubedb/cassandra-medusa-plugin/commit/ba3a4b55) Configure dependabot refresh schedule (#34)
- [c64ab302](https://github.com/kubedb/cassandra-medusa-plugin/commit/c64ab302) Configure dependabot refresh schedule (#33)
- [1c52789e](https://github.com/kubedb/cassandra-medusa-plugin/commit/1c52789e) Test against k8s 1.35 (#32)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.65.0-rc.0](https://github.com/kubedb/cli/releases/tag/v0.65.0-rc.0)

- [ba043d95](https://github.com/kubedb/cli/commit/ba043d952) Prepare for release v0.65.0-rc.0 (#827)
- [a936042e](https://github.com/kubedb/cli/commit/a936042e3) Tighten CI/release workflow secrets, perms, and release notes
- [570188ed](https://github.com/kubedb/cli/commit/570188ed7) Add CLAUDE.md pointing to AGENTS.md
- [6cdec339](https://github.com/kubedb/cli/commit/6cdec339b) Harden CI workflows (#825)
- [0cbe37e4](https://github.com/kubedb/cli/commit/0cbe37e49) Add AGENTS.md for AI coding agents (#826)
- [2be81847](https://github.com/kubedb/cli/commit/2be81847d) Prepare for release v0.64.0 (#823)
- [7132dd00](https://github.com/kubedb/cli/commit/7132dd00a) Configure dependabot refresh schedule (#822)
- [f718b38e](https://github.com/kubedb/cli/commit/f718b38ed) Configure dependabot refresh schedule (#821)
- [5667872c](https://github.com/kubedb/cli/commit/5667872c0) Test against k8s 1.35 (#820)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.20.0-rc.0](https://github.com/kubedb/clickhouse/releases/tag/v0.20.0-rc.0)

- [42932dad](https://github.com/kubedb/clickhouse/commit/42932dad) Prepare for release v0.20.0-rc.0 (#107)
- [5c3c401c](https://github.com/kubedb/clickhouse/commit/5c3c401c) Tighten CI/release workflow secrets, perms, and release notes
- [a0a93b59](https://github.com/kubedb/clickhouse/commit/a0a93b59) Harden release and release-tracker workflows
- [2c864737](https://github.com/kubedb/clickhouse/commit/2c864737) Add CLAUDE.md pointing to AGENTS.md
- [001c11d6](https://github.com/kubedb/clickhouse/commit/001c11d6) Add AGENTS.md for AI coding agents
- [f4a08fd1](https://github.com/kubedb/clickhouse/commit/f4a08fd1) Restrict /ok-to-test to org members
- [6bc3a028](https://github.com/kubedb/clickhouse/commit/6bc3a028) Prepare for release v0.19.0 (#102)
- [49cb281b](https://github.com/kubedb/clickhouse/commit/49cb281b) Offline Volume Expansion Fix (#101)
- [723034be](https://github.com/kubedb/clickhouse/commit/723034be) Configure dependabot refresh schedule (#100)



## [kubedb/clickhouse-backup-plugin](https://github.com/kubedb/clickhouse-backup-plugin)

### [v0.2.0-rc.0](https://github.com/kubedb/clickhouse-backup-plugin/releases/tag/v0.2.0-rc.0)

- [8a2fd2fa](https://github.com/kubedb/clickhouse-backup-plugin/commit/8a2fd2fa) Prepare for release v0.2.0-rc.0 (#23)
- [a4fdf4d8](https://github.com/kubedb/clickhouse-backup-plugin/commit/a4fdf4d8) Add AGENTS.md for AI coding agents
- [0d0bca21](https://github.com/kubedb/clickhouse-backup-plugin/commit/0d0bca21) Use GitHub App token for release tracker comments
- [6596ce77](https://github.com/kubedb/clickhouse-backup-plugin/commit/6596ce77) Fix DB Ready Check (#18)
- [fcd4d572](https://github.com/kubedb/clickhouse-backup-plugin/commit/fcd4d572) Prepare for release v0.1.0 (#17)
- [2401ac44](https://github.com/kubedb/clickhouse-backup-plugin/commit/2401ac44) Fix time format parsing issue (#16)
- [672e7e6c](https://github.com/kubedb/clickhouse-backup-plugin/commit/672e7e6c) Configure dependabot refresh schedule (#15)
- [532e2552](https://github.com/kubedb/clickhouse-backup-plugin/commit/532e2552) Configure dependabot refresh schedule (#14)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.20.0-rc.0](https://github.com/kubedb/crd-manager/releases/tag/v0.20.0-rc.0)




## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.23.0-rc.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.23.0-rc.0)

- [87c14d95](https://github.com/kubedb/dashboard-restic-plugin/commit/87c14d95) Prepare for release v0.23.0-rc.0 (#76)
- [f26f8aab](https://github.com/kubedb/dashboard-restic-plugin/commit/f26f8aab) Harden release and release-tracker workflows
- [da5af73f](https://github.com/kubedb/dashboard-restic-plugin/commit/da5af73f) Add AGENTS.md for AI coding agents
- [8e800912](https://github.com/kubedb/dashboard-restic-plugin/commit/8e800912) Harden CI workflows (#74)
- [9d5f47a3](https://github.com/kubedb/dashboard-restic-plugin/commit/9d5f47a3) Prepare for release v0.22.0 (#73)
- [44ccb689](https://github.com/kubedb/dashboard-restic-plugin/commit/44ccb689) Bump RESTIC_VERSION to 0.18.1-20260421 (#72)
- [5eb81552](https://github.com/kubedb/dashboard-restic-plugin/commit/5eb81552) Configure dependabot refresh schedule (#71)
- [5d7dff71](https://github.com/kubedb/dashboard-restic-plugin/commit/5d7dff71) Configure dependabot refresh schedule (#70)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.20.0-rc.0](https://github.com/kubedb/db-client-go/releases/tag/v0.20.0-rc.0)

- [3caab860](https://github.com/kubedb/db-client-go/commit/3caab860) Prepare for release v0.20.0-rc.0 (#244)
- [2b9e043c](https://github.com/kubedb/db-client-go/commit/2b9e043c) Harden release and release-tracker workflows
- [462a2a68](https://github.com/kubedb/db-client-go/commit/462a2a68) Qdrant HTTP Client TLS (#239)
- [3460fe00](https://github.com/kubedb/db-client-go/commit/3460fe00) Update for distributed postgres (#243)
- [df0e92a0](https://github.com/kubedb/db-client-go/commit/df0e92a0) Add Milvus TLS (#232)
- [e122e57f](https://github.com/kubedb/db-client-go/commit/e122e57f) Add AGENTS.md for AI coding agents
- [465442b7](https://github.com/kubedb/db-client-go/commit/465442b7) Harden CI workflows (#240)
- [a169f968](https://github.com/kubedb/db-client-go/commit/a169f968) Prepare for release v0.19.0 (#238)
- [52036006](https://github.com/kubedb/db-client-go/commit/52036006) Configure dependabot refresh schedule (#237)
- [19607748](https://github.com/kubedb/db-client-go/commit/19607748) Configure dependabot refresh schedule (#236)



## [kubedb/db2](https://github.com/kubedb/db2)

### [v0.6.0-rc.0](https://github.com/kubedb/db2/releases/tag/v0.6.0-rc.0)

- [83b7b475](https://github.com/kubedb/db2/commit/83b7b475) Prepare for release v0.6.0-rc.0 (#25)
- [ea12a946](https://github.com/kubedb/db2/commit/ea12a946) Tighten CI/release workflow secrets, perms, and release notes
- [2151e287](https://github.com/kubedb/db2/commit/2151e287) Harden release and release-tracker workflows
- [0493419f](https://github.com/kubedb/db2/commit/0493419f) Add CLAUDE.md pointing to AGENTS.md
- [ee433a7e](https://github.com/kubedb/db2/commit/ee433a7e) Add AGENTS.md for AI coding agents
- [14a0f81b](https://github.com/kubedb/db2/commit/14a0f81b) Harden CI workflows (#23)
- [f160bc0a](https://github.com/kubedb/db2/commit/f160bc0a) Prepare for release v0.5.0 (#22)
- [e1001311](https://github.com/kubedb/db2/commit/e1001311) Configure dependabot refresh schedule (#21)



## [kubedb/db2-coordinator](https://github.com/kubedb/db2-coordinator)

### [v0.6.0-rc.0](https://github.com/kubedb/db2-coordinator/releases/tag/v0.6.0-rc.0)

- [a1e57d7](https://github.com/kubedb/db2-coordinator/commit/a1e57d7) Tighten CI/release workflow secrets, perms, and release notes
- [d6b3d18](https://github.com/kubedb/db2-coordinator/commit/d6b3d18) Harden release and release-tracker workflows
- [ac8ef66](https://github.com/kubedb/db2-coordinator/commit/ac8ef66) Add AGENTS.md for AI coding agents
- [45f3f3d](https://github.com/kubedb/db2-coordinator/commit/45f3f3d) Harden CI workflows (#10)
- [bafbaed](https://github.com/kubedb/db2-coordinator/commit/bafbaed) image rebuild (#9)
- [5cbc34b](https://github.com/kubedb/db2-coordinator/commit/5cbc34b) Configure dependabot refresh schedule (#8)
- [382966b](https://github.com/kubedb/db2-coordinator/commit/382966b) Configure dependabot refresh schedule (#7)



## [kubedb/documentdb](https://github.com/kubedb/documentdb)

### [v0.2.0-rc.0](https://github.com/kubedb/documentdb/releases/tag/v0.2.0-rc.0)

- [8ff7cc98](https://github.com/kubedb/documentdb/commit/8ff7cc98) Prepare for release v0.2.0-rc.0 (#15)
- [aa27c166](https://github.com/kubedb/documentdb/commit/aa27c166) Tighten CI/release workflow secrets, perms, and release notes
- [9c8eb095](https://github.com/kubedb/documentdb/commit/9c8eb095) removed default password (#14)
- [76f64d15](https://github.com/kubedb/documentdb/commit/76f64d15) Harden release and release-tracker workflows
- [56ea3118](https://github.com/kubedb/documentdb/commit/56ea3118) Add AGENTS.md for AI coding agents
- [2e08c01c](https://github.com/kubedb/documentdb/commit/2e08c01c) Harden CI workflows (#11)
- [35bad366](https://github.com/kubedb/documentdb/commit/35bad366) Prepare for release v0.1.0 (#10)
- [d2a7d975](https://github.com/kubedb/documentdb/commit/d2a7d975) Configure dependabot refresh schedule (#9)
- [d5aef4d2](https://github.com/kubedb/documentdb/commit/d5aef4d2) Configure dependabot refresh schedule (#8)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.20.0-rc.0](https://github.com/kubedb/druid/releases/tag/v0.20.0-rc.0)

- [e35bc8a6](https://github.com/kubedb/druid/commit/e35bc8a6) Prepare for release v0.20.0-rc.0 (#132)
- [d8f0c94f](https://github.com/kubedb/druid/commit/d8f0c94f) Tighten CI/release workflow secrets, perms, and release notes
- [81bde5dd](https://github.com/kubedb/druid/commit/81bde5dd) Implement StorageMigration OpsRequest for Druid (#131)
- [fc0a7a2b](https://github.com/kubedb/druid/commit/fc0a7a2b) Harden release and release-tracker workflows
- [bb6a95d0](https://github.com/kubedb/druid/commit/bb6a95d0) Add CLAUDE.md pointing to AGENTS.md
- [07b19896](https://github.com/kubedb/druid/commit/07b19896) Add AGENTS.md for AI coding agents
- [5d02b3ca](https://github.com/kubedb/druid/commit/5d02b3ca) Harden CI workflows (#129)
- [b304b260](https://github.com/kubedb/druid/commit/b304b260) Prepare for release v0.19.0 (#128)
- [67598e74](https://github.com/kubedb/druid/commit/67598e74) Fix Offline VolumeExpansion (#127)
- [8ba485de](https://github.com/kubedb/druid/commit/8ba485de) Configure dependabot refresh schedule (#126)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.65.0-rc.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.65.0-rc.0)

- [ecddd1f6](https://github.com/kubedb/elasticsearch/commit/ecddd1f68) Prepare for release v0.65.0-rc.0 (#812)
- [fc42205f](https://github.com/kubedb/elasticsearch/commit/fc42205fa) Tighten CI/release workflow secrets, perms, and release notes
- [df2dae4a](https://github.com/kubedb/elasticsearch/commit/df2dae4a0) Harden release and release-tracker workflows
- [a134aed4](https://github.com/kubedb/elasticsearch/commit/a134aed4a) Run Ops Request Locally (#811)
- [02b16b73](https://github.com/kubedb/elasticsearch/commit/02b16b73b) Add CLAUDE.md pointing to AGENTS.md
- [65db23b7](https://github.com/kubedb/elasticsearch/commit/65db23b70) Add AGENTS.md for AI coding agents
- [ebf149db](https://github.com/kubedb/elasticsearch/commit/ebf149dba) Harden CI workflows (#808)
- [295aa682](https://github.com/kubedb/elasticsearch/commit/295aa682d) Prepare for release v0.64.0 (#807)
- [1d4e0ea3](https://github.com/kubedb/elasticsearch/commit/1d4e0ea30) Fix Volume Expansion Opsreq bug (#806)
- [aaa9bf10](https://github.com/kubedb/elasticsearch/commit/aaa9bf109) Configure dependabot refresh schedule (#805)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.28.0-rc.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.28.0-rc.0)

- [f7bce7aa](https://github.com/kubedb/elasticsearch-restic-plugin/commit/f7bce7aa) Prepare for release v0.28.0-rc.0 (#99)
- [b38a1c55](https://github.com/kubedb/elasticsearch-restic-plugin/commit/b38a1c55) Harden release and release-tracker workflows
- [853d30a7](https://github.com/kubedb/elasticsearch-restic-plugin/commit/853d30a7) Add AGENTS.md for AI coding agents
- [8b4b3ac0](https://github.com/kubedb/elasticsearch-restic-plugin/commit/8b4b3ac0) Harden CI workflows (#97)
- [f649744f](https://github.com/kubedb/elasticsearch-restic-plugin/commit/f649744f) Prepare for release v0.27.0 (#96)
- [43b4eb14](https://github.com/kubedb/elasticsearch-restic-plugin/commit/43b4eb14) Bump RESTIC_VERSION to 0.18.1-20260421 (#95)
- [779030ba](https://github.com/kubedb/elasticsearch-restic-plugin/commit/779030ba) Configure dependabot refresh schedule (#94)
- [b903f26d](https://github.com/kubedb/elasticsearch-restic-plugin/commit/b903f26d) Configure dependabot refresh schedule (#93)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.20.0-rc.0](https://github.com/kubedb/ferretdb/releases/tag/v0.20.0-rc.0)

- [d675099b](https://github.com/kubedb/ferretdb/commit/d675099b) Prepare for release v0.20.0-rc.0 (#119)
- [3a79f27d](https://github.com/kubedb/ferretdb/commit/3a79f27d) Tighten CI/release workflow secrets, perms, and release notes
- [2f9b7581](https://github.com/kubedb/ferretdb/commit/2f9b7581) Harden release and release-tracker workflows
- [50fc01b7](https://github.com/kubedb/ferretdb/commit/50fc01b7) Add AGENTS.md for AI coding agents
- [cab0a64d](https://github.com/kubedb/ferretdb/commit/cab0a64d) Harden CI workflows (#117)
- [805d066b](https://github.com/kubedb/ferretdb/commit/805d066b) Prepare for release v0.19.0 (#116)
- [6650292a](https://github.com/kubedb/ferretdb/commit/6650292a) Configure dependabot refresh schedule (#115)
- [04040043](https://github.com/kubedb/ferretdb/commit/04040043) Configure dependabot refresh schedule (#114)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.13.0-rc.0](https://github.com/kubedb/gitops/releases/tag/v0.13.0-rc.0)

- [a5a97ff5](https://github.com/kubedb/gitops/commit/a5a97ff5) Prepare for release v0.13.0-rc.0 (#57)
- [da0e1032](https://github.com/kubedb/gitops/commit/da0e1032) Tighten CI/release workflow secrets, perms, and release notes
- [baab2fac](https://github.com/kubedb/gitops/commit/baab2fac) Fix Recurring Ops Creation for ReconfigeTLS (#55)
- [40969d95](https://github.com/kubedb/gitops/commit/40969d95) Harden release and release-tracker workflows
- [e72a0e12](https://github.com/kubedb/gitops/commit/e72a0e12) Add CLAUDE.md pointing to AGENTS.md
- [7152e2d5](https://github.com/kubedb/gitops/commit/7152e2d5) Add AGENTS.md for AI coding agents
- [c7a5c666](https://github.com/kubedb/gitops/commit/c7a5c666) Harden CI workflows (#54)
- [9cbab643](https://github.com/kubedb/gitops/commit/9cbab643) Prepare for release v0.12.0 (#53)
- [4447f92c](https://github.com/kubedb/gitops/commit/4447f92c) Configure dependabot refresh schedule (#52)
- [64f35282](https://github.com/kubedb/gitops/commit/64f35282) Configure dependabot refresh schedule (#51)



## [kubedb/hanadb](https://github.com/kubedb/hanadb)

### [v0.6.0-rc.0](https://github.com/kubedb/hanadb/releases/tag/v0.6.0-rc.0)

- [5f480c10](https://github.com/kubedb/hanadb/commit/5f480c10) Prepare for release v0.6.0-rc.0 (#39)
- [28694dd1](https://github.com/kubedb/hanadb/commit/28694dd1) Tighten CI/release workflow secrets, perms, and release notes
- [9fb48705](https://github.com/kubedb/hanadb/commit/9fb48705) Harden release and release-tracker workflows
- [d530b4af](https://github.com/kubedb/hanadb/commit/d530b4af) Add AGENTS.md for AI coding agents
- [52612af6](https://github.com/kubedb/hanadb/commit/52612af6) Harden CI workflows (#32)
- [2f15f4f9](https://github.com/kubedb/hanadb/commit/2f15f4f9) Prepare for release v0.5.0 (#31)
- [6dce67f7](https://github.com/kubedb/hanadb/commit/6dce67f7) Configure dependabot refresh schedule (#30)
- [7b2be4f9](https://github.com/kubedb/hanadb/commit/7b2be4f9) Configure dependabot refresh schedule (#29)



## [kubedb/hanadb-coordinator](https://github.com/kubedb/hanadb-coordinator)

### [v0.5.0-rc.0](https://github.com/kubedb/hanadb-coordinator/releases/tag/v0.5.0-rc.0)

- [8d68a1ab](https://github.com/kubedb/hanadb-coordinator/commit/8d68a1ab) Prepare for release v0.5.0-rc.0 (#13)
- [a1df343c](https://github.com/kubedb/hanadb-coordinator/commit/a1df343c) Tighten CI/release workflow secrets, perms, and release notes
- [de35968d](https://github.com/kubedb/hanadb-coordinator/commit/de35968d) Harden release and release-tracker workflows
- [0a5612a4](https://github.com/kubedb/hanadb-coordinator/commit/0a5612a4) Add AGENTS.md for AI coding agents
- [f8cd4851](https://github.com/kubedb/hanadb-coordinator/commit/f8cd4851) Harden CI workflows (#11)
- [831fad13](https://github.com/kubedb/hanadb-coordinator/commit/831fad13) Prepare for release v0.4.0 (#10)
- [57516959](https://github.com/kubedb/hanadb-coordinator/commit/57516959) Configure dependabot refresh schedule (#9)
- [f581db41](https://github.com/kubedb/hanadb-coordinator/commit/f581db41) Configure dependabot refresh schedule (#8)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.11.0-rc.0](https://github.com/kubedb/hazelcast/releases/tag/v0.11.0-rc.0)

- [8cfb374f](https://github.com/kubedb/hazelcast/commit/8cfb374f) Prepare for release v0.11.0-rc.0 (#46)
- [198c8a50](https://github.com/kubedb/hazelcast/commit/198c8a50) Tighten CI/release workflow secrets, perms, and release notes
- [f31ab7cc](https://github.com/kubedb/hazelcast/commit/f31ab7cc) Harden release and release-tracker workflows
- [791b67dd](https://github.com/kubedb/hazelcast/commit/791b67dd) Add CLAUDE.md pointing to AGENTS.md
- [289a07ef](https://github.com/kubedb/hazelcast/commit/289a07ef) Add AGENTS.md for AI coding agents
- [f8e2fc26](https://github.com/kubedb/hazelcast/commit/f8e2fc26) Harden CI workflows (#43)
- [a49ffac7](https://github.com/kubedb/hazelcast/commit/a49ffac7) Prepare for release v0.10.0 (#42)
- [1bf1cda9](https://github.com/kubedb/hazelcast/commit/1bf1cda9) Fixed Offline Volume Expansion (#41)
- [d8c06262](https://github.com/kubedb/hazelcast/commit/d8c06262) Configure dependabot refresh schedule (#40)
- [d063a0ec](https://github.com/kubedb/hazelcast/commit/d063a0ec) Configure dependabot refresh schedule (#39)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.12.0-rc.0](https://github.com/kubedb/ignite/releases/tag/v0.12.0-rc.0)

- [2ae50f33](https://github.com/kubedb/ignite/commit/2ae50f33) Prepare for release v0.12.0-rc.0 (#54)
- [b04edbc9](https://github.com/kubedb/ignite/commit/b04edbc9) Tighten CI/release workflow secrets, perms, and release notes
- [b966bfea](https://github.com/kubedb/ignite/commit/b966bfea) Harden release and release-tracker workflows
- [2b92c194](https://github.com/kubedb/ignite/commit/2b92c194) Add CLAUDE.md pointing to AGENTS.md
- [a44aad7e](https://github.com/kubedb/ignite/commit/a44aad7e) Add AGENTS.md for AI coding agents
- [9848ac9f](https://github.com/kubedb/ignite/commit/9848ac9f) Harden CI workflows (#51)
- [dc7c8a79](https://github.com/kubedb/ignite/commit/dc7c8a79) Prepare for release v0.11.0 (#50)
- [5a4750da](https://github.com/kubedb/ignite/commit/5a4750da) Offline Volume Expansion Fix (#49)
- [f6e606b1](https://github.com/kubedb/ignite/commit/f6e606b1) Configure dependabot refresh schedule (#48)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2026.5.18-rc.0](https://github.com/kubedb/installer/releases/tag/v2026.5.18-rc.0)

- [164a534e](https://github.com/kubedb/installer/commit/164a534e0) Prepare for release v2026.5.18-rc.0 (#2321)
- [2a63f9c8](https://github.com/kubedb/installer/commit/2a63f9c87) Update crds for kubedb/apimachinery@05817720 (#2315)
- [f9ec8682](https://github.com/kubedb/installer/commit/f9ec8682d) Add CLAUDE.md pointing to AGENTS.md



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.36.0-rc.0](https://github.com/kubedb/kafka/releases/tag/v0.36.0-rc.0)

- [af590f7a](https://github.com/kubedb/kafka/commit/af590f7a) Prepare for release v0.36.0-rc.0 (#196)
- [1182e416](https://github.com/kubedb/kafka/commit/1182e416) Tighten CI/release workflow secrets, perms, and release notes
- [bc7be0ce](https://github.com/kubedb/kafka/commit/bc7be0ce) Harden release and release-tracker workflows
- [2bfcd4c6](https://github.com/kubedb/kafka/commit/2bfcd4c6) Run Ops Request Locally (#195)
- [7cbac0ff](https://github.com/kubedb/kafka/commit/7cbac0ff) Add CLAUDE.md pointing to AGENTS.md
- [96456b8e](https://github.com/kubedb/kafka/commit/96456b8e) Add AGENTS.md for AI coding agents
- [8b49617d](https://github.com/kubedb/kafka/commit/8b49617d) Harden CI workflows (#192)
- [6ce1397a](https://github.com/kubedb/kafka/commit/6ce1397a) Prepare for release v0.35.0 (#191)
- [832e9e7f](https://github.com/kubedb/kafka/commit/832e9e7f) Offline Volume Expansion Fix (#190)
- [e62fe0c7](https://github.com/kubedb/kafka/commit/e62fe0c7) Configure dependabot refresh schedule (#189)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.41.0-rc.0](https://github.com/kubedb/kibana/releases/tag/v0.41.0-rc.0)

- [7f848668](https://github.com/kubedb/kibana/commit/7f848668) Prepare for release v0.41.0-rc.0 (#182)
- [4a984afd](https://github.com/kubedb/kibana/commit/4a984afd) Tighten CI/release workflow secrets, perms, and release notes
- [aee66c89](https://github.com/kubedb/kibana/commit/aee66c89) Harden release and release-tracker workflows
- [3e24fcd8](https://github.com/kubedb/kibana/commit/3e24fcd8) Add CLAUDE.md pointing to AGENTS.md
- [ea78001c](https://github.com/kubedb/kibana/commit/ea78001c) Fix release tracker workflow
- [b4fb97cd](https://github.com/kubedb/kibana/commit/b4fb97cd) Use GitHub App token for release tracker comments (#179)
- [0420d005](https://github.com/kubedb/kibana/commit/0420d005) Merge branch 'master' into use-app-token-2284
- [70f324fc](https://github.com/kubedb/kibana/commit/70f324fc) Add AGENTS.md for AI coding agents
- [15282d6c](https://github.com/kubedb/kibana/commit/15282d6c) Pin actions to commit SHAs
- [27d597c7](https://github.com/kubedb/kibana/commit/27d597c7) Harden CI workflows (#180)
- [402312ef](https://github.com/kubedb/kibana/commit/402312ef) Use GitHub App token for release tracker comments
- [a125f71e](https://github.com/kubedb/kibana/commit/a125f71e) Prepare for release v0.40.0 (#177)
- [af12545d](https://github.com/kubedb/kibana/commit/af12545d) Configure dependabot refresh schedule (#176)
- [aa8a3fa5](https://github.com/kubedb/kibana/commit/aa8a3fa5) Configure dependabot refresh schedule (#175)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.28.0-rc.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.28.0-rc.0)

- [1874f594](https://github.com/kubedb/kubedb-manifest-plugin/commit/1874f594) Prepare for release v0.28.0-rc.0 (#132)
- [a8c857c8](https://github.com/kubedb/kubedb-manifest-plugin/commit/a8c857c8) Harden release and release-tracker workflows
- [4390340d](https://github.com/kubedb/kubedb-manifest-plugin/commit/4390340d) Add AGENTS.md for AI coding agents
- [88a6d9b5](https://github.com/kubedb/kubedb-manifest-plugin/commit/88a6d9b5) Harden CI workflows (#130)
- [f07e155d](https://github.com/kubedb/kubedb-manifest-plugin/commit/f07e155d) Prepare for release v0.27.0 (#128)
- [b22aa0b9](https://github.com/kubedb/kubedb-manifest-plugin/commit/b22aa0b9) Bump RESTIC_VERSION to 0.18.1-20260421 (#127)
- [c4e0a75c](https://github.com/kubedb/kubedb-manifest-plugin/commit/c4e0a75c) Configure dependabot refresh schedule (#126)
- [554ca50b](https://github.com/kubedb/kubedb-manifest-plugin/commit/554ca50b) Configure dependabot refresh schedule (#125)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.16.0-rc.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.16.0-rc.0)

- [e78f35ea](https://github.com/kubedb/kubedb-verifier/commit/e78f35ea) Prepare for release v0.16.0-rc.0 (#51)
- [fa4b05f8](https://github.com/kubedb/kubedb-verifier/commit/fa4b05f8) Harden release and release-tracker workflows
- [f13312e1](https://github.com/kubedb/kubedb-verifier/commit/f13312e1) Add AGENTS.md for AI coding agents
- [8d1ab507](https://github.com/kubedb/kubedb-verifier/commit/8d1ab507) Harden CI workflows (#49)
- [8a0602f4](https://github.com/kubedb/kubedb-verifier/commit/8a0602f4) Prepare for release v0.15.0 (#48)
- [323d90e5](https://github.com/kubedb/kubedb-verifier/commit/323d90e5) Configure dependabot refresh schedule (#47)
- [669fa6c3](https://github.com/kubedb/kubedb-verifier/commit/669fa6c3) Configure dependabot refresh schedule (#46)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.49.0-rc.0](https://github.com/kubedb/mariadb/releases/tag/v0.49.0-rc.0)

- [59e4024a](https://github.com/kubedb/mariadb/commit/59e4024aa) Prepare for release v0.49.0-rc.0 (#402)
- [38769956](https://github.com/kubedb/mariadb/commit/387699562) Tighten CI/release workflow secrets, perms, and release notes
- [ff6c15a4](https://github.com/kubedb/mariadb/commit/ff6c15a42) Add wal backup support for azure credless mode (#401)
- [04e3bd52](https://github.com/kubedb/mariadb/commit/04e3bd52a) Harden release and release-tracker workflows
- [27edbbd2](https://github.com/kubedb/mariadb/commit/27edbbd28) fix distributed reconfig (#400)
- [83c8573a](https://github.com/kubedb/mariadb/commit/83c8573ac) Run Ops Request Locally (#399)
- [6f736842](https://github.com/kubedb/mariadb/commit/6f736842b) Add CLAUDE.md pointing to AGENTS.md
- [8eb47e99](https://github.com/kubedb/mariadb/commit/8eb47e992) Fix Label on Health Check (#396)
- [109bf50f](https://github.com/kubedb/mariadb/commit/109bf50f3) Add AGENTS.md for AI coding agents
- [78c6159b](https://github.com/kubedb/mariadb/commit/78c6159b7) Harden CI workflows (#395)
- [2771f497](https://github.com/kubedb/mariadb/commit/2771f497d) sidekick leader selection fix; storage secret sync up fix (#394)
- [95c40d9d](https://github.com/kubedb/mariadb/commit/95c40d9d2) Prepare for release v0.48.0 (#393)
- [14558b15](https://github.com/kubedb/mariadb/commit/14558b159) Delete metrics-exporter-config secret (#392)
- [0dba0b88](https://github.com/kubedb/mariadb/commit/0dba0b886) Ensure cloud annotations to SA before sidekick creation (#386)
- [688a5667](https://github.com/kubedb/mariadb/commit/688a56674) Configure dependabot refresh schedule (#391)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.25.0-rc.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.25.0-rc.0)

- [0c208981](https://github.com/kubedb/mariadb-archiver/commit/0c208981) Prepare for release v0.25.0-rc.0 (#92)
- [e0dc262d](https://github.com/kubedb/mariadb-archiver/commit/e0dc262d) Tighten CI/release workflow secrets, perms, and release notes
- [b10e53a2](https://github.com/kubedb/mariadb-archiver/commit/b10e53a2) Harden release and release-tracker workflows
- [2a326a8b](https://github.com/kubedb/mariadb-archiver/commit/2a326a8b) Add AGENTS.md for AI coding agents
- [0eae12a8](https://github.com/kubedb/mariadb-archiver/commit/0eae12a8) Harden CI workflows (#90)
- [13e2fce5](https://github.com/kubedb/mariadb-archiver/commit/13e2fce5) Prepare for release v0.24.0 (#88)
- [40106456](https://github.com/kubedb/mariadb-archiver/commit/40106456) Update Wal-G version for AWS credless mode (#85)
- [b3eb743f](https://github.com/kubedb/mariadb-archiver/commit/b3eb743f) Configure dependabot refresh schedule (#87)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.45.0-rc.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.45.0-rc.0)

- [da2d60f4](https://github.com/kubedb/mariadb-coordinator/commit/da2d60f4) Prepare for release v0.45.0-rc.0 (#179)
- [77d610ef](https://github.com/kubedb/mariadb-coordinator/commit/77d610ef) Tighten CI/release workflow secrets, perms, and release notes
- [492ee0ac](https://github.com/kubedb/mariadb-coordinator/commit/492ee0ac) Chaos Test: Fix Disaster Recovery (#174)
- [7b81cb8c](https://github.com/kubedb/mariadb-coordinator/commit/7b81cb8c) Harden release and release-tracker workflows
- [43a65240](https://github.com/kubedb/mariadb-coordinator/commit/43a65240) Add AGENTS.md for AI coding agents
- [0af4eea5](https://github.com/kubedb/mariadb-coordinator/commit/0af4eea5) Harden CI workflows (#177)
- [8eeb0bbb](https://github.com/kubedb/mariadb-coordinator/commit/8eeb0bbb) Prepare for release v0.44.0 (#175)
- [4e9fe98d](https://github.com/kubedb/mariadb-coordinator/commit/4e9fe98d) Configure dependabot refresh schedule (#173)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.25.0-rc.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.25.0-rc.0)

- [f432211e](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/f432211e) Prepare for release v0.25.0-rc.0 (#78)
- [b75d52e3](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/b75d52e3) Tighten CI/release workflow secrets, perms, and release notes
- [d622d7f3](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/d622d7f3) Harden release and release-tracker workflows
- [748e02a6](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/748e02a6) Add AGENTS.md for AI coding agents
- [cccdc734](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/cccdc734) Harden CI workflows (#76)
- [976621bf](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/976621bf) Prepare for release v0.24.0 (#75)
- [5c212903](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/5c212903) Configure dependabot refresh schedule (#74)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.23.0-rc.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.23.0-rc.0)

- [7ffb72a7](https://github.com/kubedb/mariadb-restic-plugin/commit/7ffb72a7) Prepare for release v0.23.0-rc.0 (#90)
- [9e4c9505](https://github.com/kubedb/mariadb-restic-plugin/commit/9e4c9505) Harden release and release-tracker workflows
- [3e6d811d](https://github.com/kubedb/mariadb-restic-plugin/commit/3e6d811d) Add AGENTS.md for AI coding agents
- [df88fa1b](https://github.com/kubedb/mariadb-restic-plugin/commit/df88fa1b) Harden CI workflows (#88)
- [be3b7385](https://github.com/kubedb/mariadb-restic-plugin/commit/be3b7385) Prepare for release v0.22.0 (#86)
- [f68e059d](https://github.com/kubedb/mariadb-restic-plugin/commit/f68e059d) Bump RESTIC_VERSION to 0.18.1-20260421 (#85)
- [14c31f3b](https://github.com/kubedb/mariadb-restic-plugin/commit/14c31f3b) Configure dependabot refresh schedule (#84)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.58.0-rc.0](https://github.com/kubedb/memcached/releases/tag/v0.58.0-rc.0)

- [52bc0492](https://github.com/kubedb/memcached/commit/52bc04927) Prepare for release v0.58.0-rc.0 (#539)
- [8d50dcc5](https://github.com/kubedb/memcached/commit/8d50dcc52) Tighten CI/release workflow secrets, perms, and release notes
- [29e17104](https://github.com/kubedb/memcached/commit/29e17104a) Harden release and release-tracker workflows
- [b640b6a5](https://github.com/kubedb/memcached/commit/b640b6a5f) Run Ops Request Locally (#538)
- [8fa0e190](https://github.com/kubedb/memcached/commit/8fa0e1903) Add CLAUDE.md pointing to AGENTS.md
- [628d4100](https://github.com/kubedb/memcached/commit/628d4100f) Add AGENTS.md for AI coding agents
- [ea061558](https://github.com/kubedb/memcached/commit/ea0615585) Harden CI workflows (#536)
- [9fb644d2](https://github.com/kubedb/memcached/commit/9fb644d26) Prepare for release v0.57.0 (#535)
- [a9f93a3c](https://github.com/kubedb/memcached/commit/a9f93a3cc) Configure dependabot refresh schedule (#534)



## [kubedb/migrator-cli](https://github.com/kubedb/migrator-cli)

### [v0.5.0-rc.0](https://github.com/kubedb/migrator-cli/releases/tag/v0.5.0-rc.0)

- [cfecdf1](https://github.com/kubedb/migrator-cli/commit/cfecdf1) Prepare for release v0.5.0-rc.0 (#21)
- [1615cc3](https://github.com/kubedb/migrator-cli/commit/1615cc3) Tighten CI/release workflow secrets, perms, and release notes
- [bbba784](https://github.com/kubedb/migrator-cli/commit/bbba784) Added MongoDB migration (#16)
- [27f1c91](https://github.com/kubedb/migrator-cli/commit/27f1c91) Harden release and release-tracker workflows
- [d21c274](https://github.com/kubedb/migrator-cli/commit/d21c274) Add AGENTS.md for AI coding agents
- [7d5ab3d](https://github.com/kubedb/migrator-cli/commit/7d5ab3d) Harden CI workflows (#19)
- [8a21ae4](https://github.com/kubedb/migrator-cli/commit/8a21ae4) Separate dockerfile for each databases (#17)
- [e1fea3b](https://github.com/kubedb/migrator-cli/commit/e1fea3b) Prepare for release v0.4.0 (#18)
- [f8e80b9](https://github.com/kubedb/migrator-cli/commit/f8e80b9) Configure dependabot refresh schedule (#15)



## [kubedb/migrator-operator](https://github.com/kubedb/migrator-operator)

### [v0.5.0-rc.0](https://github.com/kubedb/migrator-operator/releases/tag/v0.5.0-rc.0)

- [9167b80](https://github.com/kubedb/migrator-operator/commit/9167b80) Prepare for release v0.5.0-rc.0 (#17)
- [dd9aacc](https://github.com/kubedb/migrator-operator/commit/dd9aacc) Tighten CI/release workflow secrets, perms, and release notes
- [ad706b6](https://github.com/kubedb/migrator-operator/commit/ad706b6) Changes for mongodb migration (#15)
- [e8e8b07](https://github.com/kubedb/migrator-operator/commit/e8e8b07) Harden release and release-tracker workflows
- [638a996](https://github.com/kubedb/migrator-operator/commit/638a996) Add AGENTS.md for AI coding agents
- [ecaa185](https://github.com/kubedb/migrator-operator/commit/ecaa185) Harden CI workflows (#14)
- [ff0cdd8](https://github.com/kubedb/migrator-operator/commit/ff0cdd8) Prepare for release v0.4.0 (#13)
- [d23a639](https://github.com/kubedb/migrator-operator/commit/d23a639) Configure dependabot refresh schedule (#12)



## [kubedb/milvus](https://github.com/kubedb/milvus)

### [v0.6.0-rc.0](https://github.com/kubedb/milvus/releases/tag/v0.6.0-rc.0)

- [062b7eab](https://github.com/kubedb/milvus/commit/062b7eab) Prepare for release v0.6.0-rc.0 (#37)
- [984bf9be](https://github.com/kubedb/milvus/commit/984bf9be) Tighten CI/release workflow secrets, perms, and release notes
- [63fee186](https://github.com/kubedb/milvus/commit/63fee186) Harden release and release-tracker workflows
- [ceb74c51](https://github.com/kubedb/milvus/commit/ceb74c51) Add CLAUDE.md pointing to AGENTS.md
- [b69a946b](https://github.com/kubedb/milvus/commit/b69a946b) Add Milvus Tls (#25)
- [f319f086](https://github.com/kubedb/milvus/commit/f319f086) Add AGENTS.md for AI coding agents
- [10e389a6](https://github.com/kubedb/milvus/commit/10e389a6) Harden CI workflows (#34)
- [649cc2ae](https://github.com/kubedb/milvus/commit/649cc2ae) Prepare for release v0.5.0 (#32)
- [03934e0b](https://github.com/kubedb/milvus/commit/03934e0b) Configure dependabot refresh schedule (#31)
- [c04453e9](https://github.com/kubedb/milvus/commit/c04453e9) Configure dependabot refresh schedule (#30)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.58.0-rc.0](https://github.com/kubedb/mongodb/releases/tag/v0.58.0-rc.0)

- [0f32df04](https://github.com/kubedb/mongodb/commit/0f32df042) Prepare for release v0.58.0-rc.0 (#761)
- [d1f66646](https://github.com/kubedb/mongodb/commit/d1f66646a) Tighten CI/release workflow secrets, perms, and release notes
- [71ddfc66](https://github.com/kubedb/mongodb/commit/71ddfc668) Harden release and release-tracker workflows
- [b6c2d289](https://github.com/kubedb/mongodb/commit/b6c2d2898) Run Ops Request Locally (#760)
- [93f12e1a](https://github.com/kubedb/mongodb/commit/93f12e1a9) Add StorageMigration OpsRequest support (#759)
- [c895f662](https://github.com/kubedb/mongodb/commit/c895f662f) Add CLAUDE.md pointing to AGENTS.md
- [a80e04c8](https://github.com/kubedb/mongodb/commit/a80e04c8a) Add AGENTS.md for AI coding agents
- [30e4b394](https://github.com/kubedb/mongodb/commit/30e4b394c) Harden CI workflows (#757)
- [56b32245](https://github.com/kubedb/mongodb/commit/56b322456) Fix Sidekick issue; Fix storage cred secret sync issue (#756)
- [5369fa5a](https://github.com/kubedb/mongodb/commit/5369fa5a4) Prepare for release v0.57.0 (#755)
- [5d951fbc](https://github.com/kubedb/mongodb/commit/5d951fbcb) Configure dependabot refresh schedule (#754)
- [7a809347](https://github.com/kubedb/mongodb/commit/7a8093470) Configure dependabot refresh schedule (#753)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.26.0-rc.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.26.0-rc.0)

- [154dedfb](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/154dedfb) Prepare for release v0.26.0-rc.0 (#83)
- [86a4f09c](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/86a4f09c) Tighten CI/release workflow secrets, perms, and release notes
- [c445729e](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/c445729e) Harden release and release-tracker workflows
- [6f77d33a](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/6f77d33a) Add AGENTS.md for AI coding agents
- [fd4c25de](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/fd4c25de) Harden CI workflows (#81)
- [85824be3](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/85824be3) Prepare for release v0.25.0 (#79)
- [fdd5d683](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/fdd5d683) Configure dependabot refresh schedule (#78)
- [446ffe40](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/446ffe40) Configure dependabot refresh schedule (#77)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.28.0-rc.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.28.0-rc.0)

- [70c2a000](https://github.com/kubedb/mongodb-restic-plugin/commit/70c2a000) Prepare for release v0.28.0-rc.0 (#128)
- [88c32465](https://github.com/kubedb/mongodb-restic-plugin/commit/88c32465) Harden release and release-tracker workflows
- [4c6631af](https://github.com/kubedb/mongodb-restic-plugin/commit/4c6631af) Add uri connection string for mongodump and mongorestore (#124)
- [fcb9af9b](https://github.com/kubedb/mongodb-restic-plugin/commit/fcb9af9b) Add AGENTS.md for AI coding agents
- [76f8adb4](https://github.com/kubedb/mongodb-restic-plugin/commit/76f8adb4) Harden CI workflows (#125)
- [47c763b6](https://github.com/kubedb/mongodb-restic-plugin/commit/47c763b6) Prepare for release v0.27.0 (#123)
- [4737d045](https://github.com/kubedb/mongodb-restic-plugin/commit/4737d045) Bump RESTIC_VERSION to 0.18.1-20260421 (#122)
- [d9f373dc](https://github.com/kubedb/mongodb-restic-plugin/commit/d9f373dc) Configure dependabot refresh schedule (#121)
- [af12c273](https://github.com/kubedb/mongodb-restic-plugin/commit/af12c273) Configure dependabot refresh schedule (#120)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.20.0-rc.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.20.0-rc.0)

- [6293f384](https://github.com/kubedb/mssql-coordinator/commit/6293f384) Prepare for release v0.20.0-rc.0 (#70)
- [cb31d9b3](https://github.com/kubedb/mssql-coordinator/commit/cb31d9b3) Tighten CI/release workflow secrets, perms, and release notes
- [26c8ec07](https://github.com/kubedb/mssql-coordinator/commit/26c8ec07) Add AGENTS.md for AI coding agents
- [29b5b49e](https://github.com/kubedb/mssql-coordinator/commit/29b5b49e) Use GitHub App token for release tracker comments (#68)
- [3432bcf1](https://github.com/kubedb/mssql-coordinator/commit/3432bcf1) Prepare for release v0.19.0 (#66)
- [3ce7a06d](https://github.com/kubedb/mssql-coordinator/commit/3ce7a06d) Configure dependabot refresh schedule (#65)
- [af4c8acc](https://github.com/kubedb/mssql-coordinator/commit/af4c8acc) Configure dependabot refresh schedule (#64)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.20.0-rc.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.20.0-rc.0)

- [787c5056](https://github.com/kubedb/mssqlserver/commit/787c5056) Prepare for release v0.20.0-rc.0 (#133)
- [76e95035](https://github.com/kubedb/mssqlserver/commit/76e95035) Tighten CI/release workflow secrets, perms, and release notes
- [8afbd426](https://github.com/kubedb/mssqlserver/commit/8afbd426) Harden release and release-tracker workflows
- [c7de0e00](https://github.com/kubedb/mssqlserver/commit/c7de0e00) Add CLAUDE.md pointing to AGENTS.md
- [ccef8128](https://github.com/kubedb/mssqlserver/commit/ccef8128) Add AGENTS.md for AI coding agents
- [a8ffa607](https://github.com/kubedb/mssqlserver/commit/a8ffa607) Harden CI workflows (#130)
- [a4e311d8](https://github.com/kubedb/mssqlserver/commit/a4e311d8) Update sidekick leader selection labels; storage sync is not implemented (#129)
- [b90a9ea2](https://github.com/kubedb/mssqlserver/commit/b90a9ea2) Prepare for release v0.19.0 (#128)
- [4bed5e0e](https://github.com/kubedb/mssqlserver/commit/4bed5e0e) Offline Volume Expansion Fix (#127)
- [d72370a5](https://github.com/kubedb/mssqlserver/commit/d72370a5) Configure dependabot refresh schedule (#126)
- [1e84f01f](https://github.com/kubedb/mssqlserver/commit/1e84f01f) Configure dependabot refresh schedule (#125)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.19.0-rc.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.19.0-rc.0)

- [24678f5](https://github.com/kubedb/mssqlserver-archiver/commit/24678f5) Tighten CI/release workflow secrets, perms, and release notes
- [958106a](https://github.com/kubedb/mssqlserver-archiver/commit/958106a) Harden release and release-tracker workflows
- [ea87799](https://github.com/kubedb/mssqlserver-archiver/commit/ea87799) Add AGENTS.md for AI coding agents
- [71f8c47](https://github.com/kubedb/mssqlserver-archiver/commit/71f8c47) Harden CI workflows (#26)
- [942ca43](https://github.com/kubedb/mssqlserver-archiver/commit/942ca43) Harden CI workflows (#25)
- [1ae0349](https://github.com/kubedb/mssqlserver-archiver/commit/1ae0349) Configure dependabot refresh schedule (#24)
- [df15bc6](https://github.com/kubedb/mssqlserver-archiver/commit/df15bc6) Configure dependabot refresh schedule (#23)



## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.19.0-rc.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.19.0-rc.0)

- [5cc5025](https://github.com/kubedb/mssqlserver-walg-plugin/commit/5cc5025) Prepare for release v0.19.0-rc.0 (#59)
- [5433238](https://github.com/kubedb/mssqlserver-walg-plugin/commit/5433238) Add AGENTS.md for AI coding agents
- [0e152a6](https://github.com/kubedb/mssqlserver-walg-plugin/commit/0e152a6) Use GitHub App token for release tracker comments (#57)
- [feef31e](https://github.com/kubedb/mssqlserver-walg-plugin/commit/feef31e) Prepare for release v0.18.0 (#55)
- [bc91aed](https://github.com/kubedb/mssqlserver-walg-plugin/commit/bc91aed) Configure dependabot refresh schedule (#54)
- [64f286f](https://github.com/kubedb/mssqlserver-walg-plugin/commit/64f286f) Configure dependabot refresh schedule (#53)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.58.0-rc.0](https://github.com/kubedb/mysql/releases/tag/v0.58.0-rc.0)

- [3ed6eff3](https://github.com/kubedb/mysql/commit/3ed6eff34) Prepare for release v0.58.0-rc.0 (#752)
- [e75778ac](https://github.com/kubedb/mysql/commit/e75778ac4) Tighten CI/release workflow secrets, perms, and release notes
- [34ba4f36](https://github.com/kubedb/mysql/commit/34ba4f36c) Add wal backup support for azure credless mode (#751)
- [1acc2da1](https://github.com/kubedb/mysql/commit/1acc2da1c) Harden release and release-tracker workflows
- [22b2e3ec](https://github.com/kubedb/mysql/commit/22b2e3ec4) Version Upgrade InnoDB Cluster (#748)
- [1597abec](https://github.com/kubedb/mysql/commit/1597abec9) Add AGENTS.md for AI coding agents
- [37d17545](https://github.com/kubedb/mysql/commit/37d17545f) Harden CI workflows (#749)
- [7a04db05](https://github.com/kubedb/mysql/commit/7a04db05c) Add InnoDB Cluster Support for 8.4+ (#738)
- [1cc5ada4](https://github.com/kubedb/mysql/commit/1cc5ada40) Fix sidekick issue: remove copy func (#747)
- [00d531b1](https://github.com/kubedb/mysql/commit/00d531b15) Fix sidekick issue (#746)
- [66db8c4b](https://github.com/kubedb/mysql/commit/66db8c4b8) Prepare for release v0.57.0 (#745)
- [ce857001](https://github.com/kubedb/mysql/commit/ce857001a) Delete metrics-exporter-config secret (#744)
- [9faeb2a0](https://github.com/kubedb/mysql/commit/9faeb2a08) Configure dependabot refresh schedule (#743)
- [449c5542](https://github.com/kubedb/mysql/commit/449c5542c) Configure dependabot refresh schedule (#742)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.26.0-rc.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.26.0-rc.0)

- [eb55d4ad](https://github.com/kubedb/mysql-archiver/commit/eb55d4ad) Prepare for release v0.26.0-rc.0 (#106)
- [d0ba1b24](https://github.com/kubedb/mysql-archiver/commit/d0ba1b24) Tighten CI/release workflow secrets, perms, and release notes
- [ae11ecbd](https://github.com/kubedb/mysql-archiver/commit/ae11ecbd) Harden release and release-tracker workflows
- [53fbffa9](https://github.com/kubedb/mysql-archiver/commit/53fbffa9) Add AGENTS.md for AI coding agents
- [6448ddf1](https://github.com/kubedb/mysql-archiver/commit/6448ddf1) Harden CI workflows (#104)
- [1098c78a](https://github.com/kubedb/mysql-archiver/commit/1098c78a) Fix binlog reply (#103)
- [50815f7d](https://github.com/kubedb/mysql-archiver/commit/50815f7d) Prepare for release v0.25.0 (#102)
- [f6bc1ac1](https://github.com/kubedb/mysql-archiver/commit/f6bc1ac1) Configure dependabot refresh schedule (#101)
- [23d32e5d](https://github.com/kubedb/mysql-archiver/commit/23d32e5d) Configure dependabot refresh schedule (#100)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.43.0-rc.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.43.0-rc.0)

- [700dcc95](https://github.com/kubedb/mysql-coordinator/commit/700dcc95) Prepare for release v0.43.0-rc.0 (#180)
- [7550f622](https://github.com/kubedb/mysql-coordinator/commit/7550f622) Tighten CI/release workflow secrets, perms, and release notes
- [8772c753](https://github.com/kubedb/mysql-coordinator/commit/8772c753) Harden release and release-tracker workflows
- [1a20ce7c](https://github.com/kubedb/mysql-coordinator/commit/1a20ce7c) Fix innodb cluster support for 8.4+ (#171)
- [21bcc293](https://github.com/kubedb/mysql-coordinator/commit/21bcc293) Add AGENTS.md for AI coding agents
- [92a93955](https://github.com/kubedb/mysql-coordinator/commit/92a93955) Harden CI workflows (#178)
- [fb729a7e](https://github.com/kubedb/mysql-coordinator/commit/fb729a7e) Fix FullRecovery Acknowlegdement Process (#177)
- [f094a593](https://github.com/kubedb/mysql-coordinator/commit/f094a593) Prepare for release v0.42.0 (#176)
- [4dfe4dbe](https://github.com/kubedb/mysql-coordinator/commit/4dfe4dbe) Configure dependabot refresh schedule (#175)
- [432033b2](https://github.com/kubedb/mysql-coordinator/commit/432033b2) Configure dependabot refresh schedule (#174)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.26.0-rc.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.26.0-rc.0)

- [02b24bac](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/02b24bac) Prepare for release v0.26.0-rc.0 (#79)
- [b98f8382](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/b98f8382) Tighten CI/release workflow secrets, perms, and release notes
- [aeb71151](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/aeb71151) Harden release and release-tracker workflows
- [960c3538](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/960c3538) Add AGENTS.md for AI coding agents
- [06b90e07](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/06b90e07) Harden CI workflows (#77)
- [8cef25be](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/8cef25be) Prepare for release v0.25.0 (#76)
- [fde57433](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/fde57433) Configure dependabot refresh schedule (#75)
- [a400c0f6](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/a400c0f6) Configure dependabot refresh schedule (#74)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.28.0-rc.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.28.0-rc.0)

- [dd1b9eae](https://github.com/kubedb/mysql-restic-plugin/commit/dd1b9eae) Prepare for release v0.28.0-rc.0 (#114)
- [879cbd56](https://github.com/kubedb/mysql-restic-plugin/commit/879cbd56) Harden release and release-tracker workflows
- [3a29e9c8](https://github.com/kubedb/mysql-restic-plugin/commit/3a29e9c8) Add AGENTS.md for AI coding agents
- [a5955a10](https://github.com/kubedb/mysql-restic-plugin/commit/a5955a10) Harden CI workflows (#111)
- [3a49671f](https://github.com/kubedb/mysql-restic-plugin/commit/3a49671f) Add New Version Support, Innodb Cluster Support (#110)
- [ba400cb7](https://github.com/kubedb/mysql-restic-plugin/commit/ba400cb7) Prepare for release v0.27.0 (#108)
- [372401ba](https://github.com/kubedb/mysql-restic-plugin/commit/372401ba) Bump RESTIC_VERSION to 0.18.1-20260421 (#107)
- [1f5e0d96](https://github.com/kubedb/mysql-restic-plugin/commit/1f5e0d96) Configure dependabot refresh schedule (#106)
- [0cf94d94](https://github.com/kubedb/mysql-restic-plugin/commit/0cf94d94) Configure dependabot refresh schedule (#105)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.43.0-rc.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.43.0-rc.0)

- [1fe79e5](https://github.com/kubedb/mysql-router-init/commit/1fe79e5) Tighten CI/release workflow secrets, perms, and release notes
- [399bf84](https://github.com/kubedb/mysql-router-init/commit/399bf84) Add AGENTS.md for AI coding agents
- [e1d55d1](https://github.com/kubedb/mysql-router-init/commit/e1d55d1) Merge pull request #59 from kubedb/use-app-token-2284
- [aacf7fa](https://github.com/kubedb/mysql-router-init/commit/aacf7fa) Configure dependabot refresh schedule (#57)
- [37d1624](https://github.com/kubedb/mysql-router-init/commit/37d1624) Configure dependabot refresh schedule (#56)



## [kubedb/neo4j](https://github.com/kubedb/neo4j)

### [v0.6.0-rc.0](https://github.com/kubedb/neo4j/releases/tag/v0.6.0-rc.0)

- [63a96dad](https://github.com/kubedb/neo4j/commit/63a96dad) Prepare for release v0.6.0-rc.0 (#34)
- [e215c4c8](https://github.com/kubedb/neo4j/commit/e215c4c8) Tighten CI/release workflow secrets, perms, and release notes
- [98283581](https://github.com/kubedb/neo4j/commit/98283581) Harden release and release-tracker workflows
- [77704502](https://github.com/kubedb/neo4j/commit/77704502) Add StorageMigration OpsRequest support for Neo4j (#33)
- [8dc5edbf](https://github.com/kubedb/neo4j/commit/8dc5edbf) Add CLAUDE.md pointing to AGENTS.md
- [c8f95c51](https://github.com/kubedb/neo4j/commit/c8f95c51) Add AGENTS.md for AI coding agents
- [beb02ed6](https://github.com/kubedb/neo4j/commit/beb02ed6) Harden CI workflows (#30)
- [8a7b1696](https://github.com/kubedb/neo4j/commit/8a7b1696) Prepare for release v0.5.0 (#29)
- [8d7a6886](https://github.com/kubedb/neo4j/commit/8d7a6886) Configure dependabot refresh schedule (#28)
- [c57e8561](https://github.com/kubedb/neo4j/commit/c57e8561) Add Neo4j Ops Req  updVersion, volumeExpansion (#26)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.52.0-rc.0](https://github.com/kubedb/ops-manager/releases/tag/v0.52.0-rc.0)




## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.11.0-rc.0](https://github.com/kubedb/oracle/releases/tag/v0.11.0-rc.0)

- [22259d8b](https://github.com/kubedb/oracle/commit/22259d8b) Prepare for release v0.11.0-rc.0 (#52)
- [ef7a90cb](https://github.com/kubedb/oracle/commit/ef7a90cb) Tighten CI/release workflow secrets, perms, and release notes
- [7484f753](https://github.com/kubedb/oracle/commit/7484f753) fix-petset-GetObjectMeta (#49)
- [a2a4e973](https://github.com/kubedb/oracle/commit/a2a4e973) Harden release and release-tracker workflows
- [449beab1](https://github.com/kubedb/oracle/commit/449beab1) Add CLAUDE.md pointing to AGENTS.md
- [ae706c75](https://github.com/kubedb/oracle/commit/ae706c75) Add AGENTS.md for AI coding agents
- [47addbe9](https://github.com/kubedb/oracle/commit/47addbe9) Harden CI workflows (#43)
- [75cab6ba](https://github.com/kubedb/oracle/commit/75cab6ba) Prepare for release v0.10.0 (#41)
- [4e495dcd](https://github.com/kubedb/oracle/commit/4e495dcd) Configure dependabot refresh schedule (#40)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.11.0-rc.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.11.0-rc.0)

- [eebe9e3](https://github.com/kubedb/oracle-coordinator/commit/eebe9e3) Prepare for release v0.11.0-rc.0 (#35)
- [40f851c](https://github.com/kubedb/oracle-coordinator/commit/40f851c) Tighten CI/release workflow secrets, perms, and release notes
- [5d987d9](https://github.com/kubedb/oracle-coordinator/commit/5d987d9) Add AGENTS.md for AI coding agents
- [9c163c8](https://github.com/kubedb/oracle-coordinator/commit/9c163c8) Use GitHub App token for release tracker comments (#33)
- [f9f4dd8](https://github.com/kubedb/oracle-coordinator/commit/f9f4dd8) Prepare for release v0.10.0 (#31)
- [b2a1a40](https://github.com/kubedb/oracle-coordinator/commit/b2a1a40) Configure dependabot refresh schedule (#30)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.52.0-rc.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.52.0-rc.0)

- [a03b66b9](https://github.com/kubedb/percona-xtradb/commit/a03b66b9c) Prepare for release v0.52.0-rc.0 (#454)
- [2f3fbc70](https://github.com/kubedb/percona-xtradb/commit/2f3fbc700) Tighten CI/release workflow secrets, perms, and release notes
- [cd18b8f0](https://github.com/kubedb/percona-xtradb/commit/cd18b8f00) Harden release and release-tracker workflows
- [06ee2cbf](https://github.com/kubedb/percona-xtradb/commit/06ee2cbfb) Run Ops Request Locally (#453)
- [887fa9a0](https://github.com/kubedb/percona-xtradb/commit/887fa9a01) Add CLAUDE.md pointing to AGENTS.md
- [bc309e6c](https://github.com/kubedb/percona-xtradb/commit/bc309e6ce) Add AGENTS.md for AI coding agents
- [52dc6b56](https://github.com/kubedb/percona-xtradb/commit/52dc6b56c) Harden CI workflows (#450)
- [b10cd4ce](https://github.com/kubedb/percona-xtradb/commit/b10cd4ceb) Prepare for release v0.51.0 (#449)
- [84446e80](https://github.com/kubedb/percona-xtradb/commit/84446e805) Delete metrics exposter config secret (#448)
- [4285acd4](https://github.com/kubedb/percona-xtradb/commit/4285acd48) Configure dependabot refresh schedule (#447)
- [9ab6296a](https://github.com/kubedb/percona-xtradb/commit/9ab6296ae) Configure dependabot refresh schedule (#446)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.38.0-rc.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.38.0-rc.0)

- [d71a8b78](https://github.com/kubedb/percona-xtradb-coordinator/commit/d71a8b78) Prepare for release v0.38.0-rc.0 (#127)
- [dcf12fa2](https://github.com/kubedb/percona-xtradb-coordinator/commit/dcf12fa2) Tighten CI/release workflow secrets, perms, and release notes
- [98fc7cc6](https://github.com/kubedb/percona-xtradb-coordinator/commit/98fc7cc6) Harden release and release-tracker workflows
- [37eb5997](https://github.com/kubedb/percona-xtradb-coordinator/commit/37eb5997) Add AGENTS.md for AI coding agents
- [1881477c](https://github.com/kubedb/percona-xtradb-coordinator/commit/1881477c) Harden CI workflows (#125)
- [e1bf59f7](https://github.com/kubedb/percona-xtradb-coordinator/commit/e1bf59f7) Prepare for release v0.37.0 (#124)
- [9b44e907](https://github.com/kubedb/percona-xtradb-coordinator/commit/9b44e907) Configure dependabot refresh schedule (#123)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.49.0-rc.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.49.0-rc.0)

- [e1bf909c](https://github.com/kubedb/pg-coordinator/commit/e1bf909c) Prepare for release v0.49.0-rc.0 (#251)
- [9bda4b8a](https://github.com/kubedb/pg-coordinator/commit/9bda4b8a) Tighten CI/release workflow secrets, perms, and release notes
- [960efc2e](https://github.com/kubedb/pg-coordinator/commit/960efc2e) Harden release and release-tracker workflows
- [14041cfd](https://github.com/kubedb/pg-coordinator/commit/14041cfd) Update ci (#250)
- [be5cb705](https://github.com/kubedb/pg-coordinator/commit/be5cb705) Update ci (#249)
- [8e884ef3](https://github.com/kubedb/pg-coordinator/commit/8e884ef3) Update for distributed postgres (#248)
- [16abe9e8](https://github.com/kubedb/pg-coordinator/commit/16abe9e8) Add AGENTS.md for AI coding agents
- [80340f14](https://github.com/kubedb/pg-coordinator/commit/80340f14) Harden CI workflows (#246)
- [65e4a089](https://github.com/kubedb/pg-coordinator/commit/65e4a089) Add missing postgres client library package on ubi dockerfile (#237)
- [9b36ab39](https://github.com/kubedb/pg-coordinator/commit/9b36ab39) Prepare for release v0.48.0 (#245)
- [c9285b86](https://github.com/kubedb/pg-coordinator/commit/c9285b86) Configure dependabot refresh schedule (#244)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.52.0-rc.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.52.0-rc.0)

- [700a235c](https://github.com/kubedb/pgbouncer/commit/700a235cf) Prepare for release v0.52.0-rc.0 (#413)
- [84b1a19e](https://github.com/kubedb/pgbouncer/commit/84b1a19ea) Tighten CI/release workflow secrets, perms, and release notes
- [c3e4b4e1](https://github.com/kubedb/pgbouncer/commit/c3e4b4e10) Harden release and release-tracker workflows
- [e0746413](https://github.com/kubedb/pgbouncer/commit/e07464138) Run Ops Request Locally (#412)
- [ce2cad78](https://github.com/kubedb/pgbouncer/commit/ce2cad789) Add AGENTS.md for AI coding agents
- [1215a247](https://github.com/kubedb/pgbouncer/commit/1215a247e) Harden CI workflows (#410)
- [23024d0c](https://github.com/kubedb/pgbouncer/commit/23024d0c7) Prepare for release v0.51.0 (#409)
- [7b5ce364](https://github.com/kubedb/pgbouncer/commit/7b5ce3642) Configure dependabot refresh schedule (#408)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.20.0-rc.0](https://github.com/kubedb/pgpool/releases/tag/v0.20.0-rc.0)

- [9dd15b6e](https://github.com/kubedb/pgpool/commit/9dd15b6e) Prepare for release v0.20.0-rc.0 (#120)
- [dbd10776](https://github.com/kubedb/pgpool/commit/dbd10776) Tighten CI/release workflow secrets, perms, and release notes
- [12f83ac2](https://github.com/kubedb/pgpool/commit/12f83ac2) Harden release and release-tracker workflows
- [23a0bf8f](https://github.com/kubedb/pgpool/commit/23a0bf8f) Add AGENTS.md for AI coding agents
- [efab69a4](https://github.com/kubedb/pgpool/commit/efab69a4) Harden CI workflows (#118)
- [e34a9780](https://github.com/kubedb/pgpool/commit/e34a9780) Prepare for release v0.19.0 (#117)
- [927c85f7](https://github.com/kubedb/pgpool/commit/927c85f7) Configure dependabot refresh schedule (#116)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.65.0-rc.0](https://github.com/kubedb/postgres/releases/tag/v0.65.0-rc.0)

- [4aca95ba](https://github.com/kubedb/postgres/commit/4aca95ba7) Prepare for release v0.65.0-rc.0 (#889)
- [cd25a1fc](https://github.com/kubedb/postgres/commit/cd25a1fca) Tighten CI/release workflow secrets, perms, and release notes
- [f0d40da9](https://github.com/kubedb/postgres/commit/f0d40da9b) Add wal backup support for azure credless mode (#880)
- [557c11f9](https://github.com/kubedb/postgres/commit/557c11f95) Harden release and release-tracker workflows
- [9b22d795](https://github.com/kubedb/postgres/commit/9b22d7959) Run Ops Request Locally (#887)
- [36945c39](https://github.com/kubedb/postgres/commit/36945c392) Add CLAUDE.md pointing to AGENTS.md
- [44a337ef](https://github.com/kubedb/postgres/commit/44a337ef5) Add AGENTS.md for AI coding agents
- [c4d57a65](https://github.com/kubedb/postgres/commit/c4d57a65a) Use docker/login-action; drop redundant docker hub steps (#884)
- [723bdc55](https://github.com/kubedb/postgres/commit/723bdc55a) Harden CI workflows (#882)
- [791061f6](https://github.com/kubedb/postgres/commit/791061f6b) Fix Sidekick issue; Fix storage cred secret sync issue (#881)
- [261287ee](https://github.com/kubedb/postgres/commit/261287eee) Prepare for release v0.64.0 (#879)
- [0ab9077a](https://github.com/kubedb/postgres/commit/0ab9077ae) Configure dependabot refresh schedule (#878)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.26.0-rc.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.26.0-rc.0)

- [32139be4](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/32139be4) Prepare for release v0.26.0-rc.0 (#89)
- [e1be14bc](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/e1be14bc) Tighten CI/release workflow secrets, perms, and release notes
- [7590aa73](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/7590aa73) Harden release and release-tracker workflows
- [56eabbb2](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/56eabbb2) Add AGENTS.md for AI coding agents
- [7ce77805](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/7ce77805) Use GitHub App token for release tracker comments (#87)
- [5cdf2c50](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/5cdf2c50) Harden CI workflows (#86)
- [a1bac671](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/a1bac671) Prepare for release v0.25.0 (#85)
- [c20cd235](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/c20cd235) Configure dependabot refresh schedule (#84)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.28.0-rc.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.28.0-rc.0)

- [dad161fc](https://github.com/kubedb/postgres-restic-plugin/commit/dad161fc) Prepare for release v0.28.0-rc.0 (#109)
- [e69d3dbc](https://github.com/kubedb/postgres-restic-plugin/commit/e69d3dbc) Harden release and release-tracker workflows
- [341a4452](https://github.com/kubedb/postgres-restic-plugin/commit/341a4452) Add AGENTS.md for AI coding agents
- [6b3dcd51](https://github.com/kubedb/postgres-restic-plugin/commit/6b3dcd51) Harden CI workflows (#107)
- [b5399717](https://github.com/kubedb/postgres-restic-plugin/commit/b5399717) Prepare for release v0.27.0 (#106)
- [ff32b63e](https://github.com/kubedb/postgres-restic-plugin/commit/ff32b63e) Bump RESTIC_VERSION to 0.18.1-20260421 (#105)
- [deeb99ee](https://github.com/kubedb/postgres-restic-plugin/commit/deeb99ee) Configure dependabot refresh schedule (#104)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.26.0-rc.0](https://github.com/kubedb/provider-aws/releases/tag/v0.26.0-rc.0)

- [2b8fb0f](https://github.com/kubedb/provider-aws/commit/2b8fb0f) Tighten CI/release workflow secrets, perms, and release notes
- [2d7e8ba](https://github.com/kubedb/provider-aws/commit/2d7e8ba) Harden release and release-tracker workflows
- [c984af3](https://github.com/kubedb/provider-aws/commit/c984af3) Add AGENTS.md for AI coding agents (#44)
- [d88e77b](https://github.com/kubedb/provider-aws/commit/d88e77b) Harden CI workflows (#43)
- [a619973](https://github.com/kubedb/provider-aws/commit/a619973) Configure dependabot refresh schedule (#42)



## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.26.0-rc.0](https://github.com/kubedb/provider-azure/releases/tag/v0.26.0-rc.0)

- [630e43c](https://github.com/kubedb/provider-azure/commit/630e43c) Tighten CI/release workflow secrets, perms, and release notes
- [3d78f9c](https://github.com/kubedb/provider-azure/commit/3d78f9c) Add AGENTS.md for AI coding agents (#30)
- [7f7b570](https://github.com/kubedb/provider-azure/commit/7f7b570) Restrict /ok-to-test to org members (#29)
- [2923d77](https://github.com/kubedb/provider-azure/commit/2923d77) Configure dependabot refresh schedule (#27)



## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.26.0-rc.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.26.0-rc.0)

- [e8f212a](https://github.com/kubedb/provider-gcp/commit/e8f212a) Tighten CI/release workflow secrets, perms, and release notes
- [4f19b3b](https://github.com/kubedb/provider-gcp/commit/4f19b3b) Harden release and release-tracker workflows
- [ca1476e](https://github.com/kubedb/provider-gcp/commit/ca1476e) Add AGENTS.md for AI coding agents (#29)
- [4c70d45](https://github.com/kubedb/provider-gcp/commit/4c70d45) Harden CI workflows (#28)
- [a8dc1d6](https://github.com/kubedb/provider-gcp/commit/a8dc1d6) Configure dependabot refresh schedule (#27)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.65.0-rc.0](https://github.com/kubedb/provisioner/releases/tag/v0.65.0-rc.0)




## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.52.0-rc.0](https://github.com/kubedb/proxysql/releases/tag/v0.52.0-rc.0)

- [d5d1d882](https://github.com/kubedb/proxysql/commit/d5d1d882e) Prepare for release v0.52.0-rc.0 (#434)
- [bb7e268e](https://github.com/kubedb/proxysql/commit/bb7e268e4) Tighten CI/release workflow secrets, perms, and release notes
- [9c59d49c](https://github.com/kubedb/proxysql/commit/9c59d49ce) Harden release and release-tracker workflows
- [e29d1669](https://github.com/kubedb/proxysql/commit/e29d16693) Run Ops Request Locally (#433)
- [b47e8c2d](https://github.com/kubedb/proxysql/commit/b47e8c2d3) Add CLAUDE.md pointing to AGENTS.md
- [681acb3c](https://github.com/kubedb/proxysql/commit/681acb3c4) Add AGENTS.md for AI coding agents
- [3b6a4061](https://github.com/kubedb/proxysql/commit/3b6a40616) Harden CI workflows (#430)
- [34985f15](https://github.com/kubedb/proxysql/commit/34985f159) Prepare for release v0.51.0 (#429)
- [32439267](https://github.com/kubedb/proxysql/commit/324392678) Configure dependabot refresh schedule (#428)
- [05c70796](https://github.com/kubedb/proxysql/commit/05c707961) Configure dependabot refresh schedule (#427)



## [kubedb/qdrant](https://github.com/kubedb/qdrant)

### [v0.6.0-rc.0](https://github.com/kubedb/qdrant/releases/tag/v0.6.0-rc.0)

- [7f7dd3e5](https://github.com/kubedb/qdrant/commit/7f7dd3e5) Prepare for release v0.6.0-rc.0 (#41)
- [a79ec8c3](https://github.com/kubedb/qdrant/commit/a79ec8c3) Tighten CI/release workflow secrets, perms, and release notes
- [6715626c](https://github.com/kubedb/qdrant/commit/6715626c) Harden release and release-tracker workflows
- [784adaa7](https://github.com/kubedb/qdrant/commit/784adaa7) Add CLAUDE.md pointing to AGENTS.md
- [4ef3db03](https://github.com/kubedb/qdrant/commit/4ef3db03) Add AGENTS.md for AI coding agents
- [e3aaca10](https://github.com/kubedb/qdrant/commit/e3aaca10) Harden CI workflows (#38)
- [bcf12c03](https://github.com/kubedb/qdrant/commit/bcf12c03) Prepare for release v0.5.0 (#36)
- [0bc2e4d5](https://github.com/kubedb/qdrant/commit/0bc2e4d5) Bug Fix (#35)
- [783dd192](https://github.com/kubedb/qdrant/commit/783dd192) Offline Volume Expansion Fix (#34)
- [5c011201](https://github.com/kubedb/qdrant/commit/5c011201) Configure dependabot refresh schedule (#33)
- [f18cd566](https://github.com/kubedb/qdrant/commit/f18cd566) Configure dependabot refresh schedule (#32)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.20.0-rc.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.20.0-rc.0)

- [2826d51c](https://github.com/kubedb/rabbitmq/commit/2826d51c) Prepare for release v0.20.0-rc.0 (#133)
- [7b849665](https://github.com/kubedb/rabbitmq/commit/7b849665) Tighten CI/release workflow secrets, perms, and release notes
- [7c952783](https://github.com/kubedb/rabbitmq/commit/7c952783) Harden release and release-tracker workflows
- [8b547651](https://github.com/kubedb/rabbitmq/commit/8b547651) Add CLAUDE.md pointing to AGENTS.md
- [866a6b3c](https://github.com/kubedb/rabbitmq/commit/866a6b3c) Add AGENTS.md for AI coding agents
- [2a733db2](https://github.com/kubedb/rabbitmq/commit/2a733db2) Harden CI workflows (#130)
- [cd286bab](https://github.com/kubedb/rabbitmq/commit/cd286bab) Prepare for release v0.19.0 (#129)
- [3dff0452](https://github.com/kubedb/rabbitmq/commit/3dff0452) Fix Offline Volume Expansion (#128)
- [c80c9679](https://github.com/kubedb/rabbitmq/commit/c80c9679) Configure dependabot refresh schedule (#127)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.58.0-rc.0](https://github.com/kubedb/redis/releases/tag/v0.58.0-rc.0)

- [3afe56ea](https://github.com/kubedb/redis/commit/3afe56eab) Prepare for release v0.58.0-rc.0 (#646)
- [c961cdb4](https://github.com/kubedb/redis/commit/c961cdb42) Tighten CI/release workflow secrets, perms, and release notes
- [a09720d5](https://github.com/kubedb/redis/commit/a09720d5b) Harden release and release-tracker workflows
- [7d00abef](https://github.com/kubedb/redis/commit/7d00abef9) Run Ops Request Locally (#645)
- [740c5c96](https://github.com/kubedb/redis/commit/740c5c968) Add CLAUDE.md pointing to AGENTS.md
- [d23b442c](https://github.com/kubedb/redis/commit/d23b442c3) Add AGENTS.md for AI coding agents
- [d5c294b0](https://github.com/kubedb/redis/commit/d5c294b03) Harden CI workflows (#640)
- [34c6e5d5](https://github.com/kubedb/redis/commit/34c6e5d56) Add governing svc name in cert (#639)
- [680eaa9e](https://github.com/kubedb/redis/commit/680eaa9ec) Prepare for release v0.57.0 (#638)
- [7d78dfd9](https://github.com/kubedb/redis/commit/7d78dfd9e) Configure dependabot refresh schedule (#637)
- [fc44407f](https://github.com/kubedb/redis/commit/fc44407fc) Configure dependabot refresh schedule (#636)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.44.0-rc.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.44.0-rc.0)

- [367048d8](https://github.com/kubedb/redis-coordinator/commit/367048d8) Prepare for release v0.44.0-rc.0 (#160)
- [a083e153](https://github.com/kubedb/redis-coordinator/commit/a083e153) Tighten CI/release workflow secrets, perms, and release notes
- [25c46735](https://github.com/kubedb/redis-coordinator/commit/25c46735) Harden release and release-tracker workflows
- [928c964e](https://github.com/kubedb/redis-coordinator/commit/928c964e) Add AGENTS.md for AI coding agents
- [90c797de](https://github.com/kubedb/redis-coordinator/commit/90c797de) Harden CI workflows (#158)
- [f39b5c5d](https://github.com/kubedb/redis-coordinator/commit/f39b5c5d) Prepare for release v0.43.0 (#157)
- [f508e9dc](https://github.com/kubedb/redis-coordinator/commit/f508e9dc) Configure dependabot refresh schedule (#156)
- [2edf641a](https://github.com/kubedb/redis-coordinator/commit/2edf641a) Configure dependabot refresh schedule (#155)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.28.0-rc.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.28.0-rc.0)

- [96e95cee](https://github.com/kubedb/redis-restic-plugin/commit/96e95cee) Prepare for release v0.28.0-rc.0 (#106)
- [044d20f6](https://github.com/kubedb/redis-restic-plugin/commit/044d20f6) Harden release and release-tracker workflows
- [2c524ff2](https://github.com/kubedb/redis-restic-plugin/commit/2c524ff2) Add AGENTS.md for AI coding agents
- [e29aaa32](https://github.com/kubedb/redis-restic-plugin/commit/e29aaa32) Harden CI workflows (#104)
- [df86bb19](https://github.com/kubedb/redis-restic-plugin/commit/df86bb19) Prepare for release v0.27.0 (#103)
- [ca23d967](https://github.com/kubedb/redis-restic-plugin/commit/ca23d967) Bump RESTIC_VERSION to 0.18.1-20260421 (#102)
- [b043e083](https://github.com/kubedb/redis-restic-plugin/commit/b043e083) Configure dependabot refresh schedule (#101)
- [f301cd48](https://github.com/kubedb/redis-restic-plugin/commit/f301cd48) Configure dependabot refresh schedule (#100)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.52.0-rc.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.52.0-rc.0)

- [c2ca4a03](https://github.com/kubedb/replication-mode-detector/commit/c2ca4a03) Prepare for release v0.52.0-rc.0 (#323)
- [a22986ef](https://github.com/kubedb/replication-mode-detector/commit/a22986ef) Tighten CI/release workflow secrets, perms, and release notes
- [7f22be9e](https://github.com/kubedb/replication-mode-detector/commit/7f22be9e) Harden release and release-tracker workflows
- [e8f81f8e](https://github.com/kubedb/replication-mode-detector/commit/e8f81f8e) Add AGENTS.md for AI coding agents
- [3c2fd6f5](https://github.com/kubedb/replication-mode-detector/commit/3c2fd6f5) Harden CI workflows (#321)
- [bdc776c1](https://github.com/kubedb/replication-mode-detector/commit/bdc776c1) Prepare for release v0.51.0 (#320)
- [38d8e512](https://github.com/kubedb/replication-mode-detector/commit/38d8e512) Configure dependabot refresh schedule (#319)
- [dc13533a](https://github.com/kubedb/replication-mode-detector/commit/dc13533a) Configure dependabot refresh schedule (#318)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.41.0-rc.0](https://github.com/kubedb/schema-manager/releases/tag/v0.41.0-rc.0)

- [c31cfb97](https://github.com/kubedb/schema-manager/commit/c31cfb97) Prepare for release v0.41.0-rc.0 (#170)
- [7f62c921](https://github.com/kubedb/schema-manager/commit/7f62c921) Tighten CI/release workflow secrets, perms, and release notes
- [bba504bd](https://github.com/kubedb/schema-manager/commit/bba504bd) Harden release and release-tracker workflows
- [95b80652](https://github.com/kubedb/schema-manager/commit/95b80652) Add AGENTS.md for AI coding agents
- [c6495b95](https://github.com/kubedb/schema-manager/commit/c6495b95) Harden CI workflows (#168)
- [8590f8e5](https://github.com/kubedb/schema-manager/commit/8590f8e5) Prepare for release v0.40.0 (#167)
- [5a81c014](https://github.com/kubedb/schema-manager/commit/5a81c014) Configure dependabot refresh schedule (#166)
- [cedf8375](https://github.com/kubedb/schema-manager/commit/cedf8375) Configure dependabot refresh schedule (#165)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.20.0-rc.0](https://github.com/kubedb/singlestore/releases/tag/v0.20.0-rc.0)

- [8d5c9087](https://github.com/kubedb/singlestore/commit/8d5c9087) Prepare for release v0.20.0-rc.0 (#121)
- [26b67dc6](https://github.com/kubedb/singlestore/commit/26b67dc6) Tighten CI/release workflow secrets, perms, and release notes
- [01a4c59f](https://github.com/kubedb/singlestore/commit/01a4c59f) Harden release and release-tracker workflows
- [d663617d](https://github.com/kubedb/singlestore/commit/d663617d) Add CLAUDE.md pointing to AGENTS.md
- [e4255529](https://github.com/kubedb/singlestore/commit/e4255529) Add AGENTS.md for AI coding agents
- [90f8831e](https://github.com/kubedb/singlestore/commit/90f8831e) Harden CI workflows (#118)
- [be212d59](https://github.com/kubedb/singlestore/commit/be212d59) Prepare for release v0.19.0 (#117)
- [38cf8233](https://github.com/kubedb/singlestore/commit/38cf8233) fix volume expansion bug (#116)
- [66594528](https://github.com/kubedb/singlestore/commit/66594528) offline volume expansion fix (#115)
- [742f9ca3](https://github.com/kubedb/singlestore/commit/742f9ca3) Configure dependabot refresh schedule (#114)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.20.0-rc.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.20.0-rc.0)

- [d287dabd](https://github.com/kubedb/singlestore-coordinator/commit/d287dabd) Prepare for release v0.20.0-rc.0 (#72)
- [fcb2528e](https://github.com/kubedb/singlestore-coordinator/commit/fcb2528e) Tighten CI/release workflow secrets, perms, and release notes
- [0f78329e](https://github.com/kubedb/singlestore-coordinator/commit/0f78329e) Harden release and release-tracker workflows
- [0ff7dc5b](https://github.com/kubedb/singlestore-coordinator/commit/0ff7dc5b) Add AGENTS.md for AI coding agents
- [357d9a67](https://github.com/kubedb/singlestore-coordinator/commit/357d9a67) Harden CI workflows (#70)
- [89cacaf6](https://github.com/kubedb/singlestore-coordinator/commit/89cacaf6) Prepare for release v0.19.0 (#69)
- [4b0e55b2](https://github.com/kubedb/singlestore-coordinator/commit/4b0e55b2) Configure dependabot refresh schedule (#68)
- [a7242e7c](https://github.com/kubedb/singlestore-coordinator/commit/a7242e7c) Configure dependabot refresh schedule (#67)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.23.0-rc.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.23.0-rc.0)

- [ef6b233c](https://github.com/kubedb/singlestore-restic-plugin/commit/ef6b233c) Prepare for release v0.23.0-rc.0 (#85)
- [ecfc7fc6](https://github.com/kubedb/singlestore-restic-plugin/commit/ecfc7fc6) Harden release and release-tracker workflows
- [88b88a43](https://github.com/kubedb/singlestore-restic-plugin/commit/88b88a43) Add AGENTS.md for AI coding agents
- [9e76f60c](https://github.com/kubedb/singlestore-restic-plugin/commit/9e76f60c) Harden CI workflows (#83)
- [66fee03a](https://github.com/kubedb/singlestore-restic-plugin/commit/66fee03a) Prepare for release v0.22.0 (#82)
- [4a9f961c](https://github.com/kubedb/singlestore-restic-plugin/commit/4a9f961c) Bump RESTIC_VERSION to 0.18.1-20260421 (#81)
- [fc7ad966](https://github.com/kubedb/singlestore-restic-plugin/commit/fc7ad966) Configure dependabot refresh schedule (#80)
- [536de2a0](https://github.com/kubedb/singlestore-restic-plugin/commit/536de2a0) Configure dependabot refresh schedule (#79)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.20.0-rc.0](https://github.com/kubedb/solr/releases/tag/v0.20.0-rc.0)

- [d6f12bb7](https://github.com/kubedb/solr/commit/d6f12bb7) Prepare for release v0.20.0-rc.0 (#130)
- [b947dd5c](https://github.com/kubedb/solr/commit/b947dd5c) Tighten CI/release workflow secrets, perms, and release notes
- [5a339830](https://github.com/kubedb/solr/commit/5a339830) Harden release and release-tracker workflows
- [59801485](https://github.com/kubedb/solr/commit/59801485) Add CLAUDE.md pointing to AGENTS.md
- [8e55bf7a](https://github.com/kubedb/solr/commit/8e55bf7a) Add AGENTS.md for AI coding agents
- [a91a2ae7](https://github.com/kubedb/solr/commit/a91a2ae7) Harden CI workflows (#127)
- [d4d7b833](https://github.com/kubedb/solr/commit/d4d7b833) Prepare for release v0.19.0 (#126)
- [301b697f](https://github.com/kubedb/solr/commit/301b697f) Offline Volume Expansion Fix (#125)
- [d255112f](https://github.com/kubedb/solr/commit/d255112f) Configure dependabot refresh schedule (#124)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.50.0-rc.0](https://github.com/kubedb/tests/releases/tag/v0.50.0-rc.0)

- [7d4ad629](https://github.com/kubedb/tests/commit/7d4ad629a) Prepare for release v0.50.0-rc.0 (#536)
- [2ad57dc7](https://github.com/kubedb/tests/commit/2ad57dc71) Harden release and release-tracker workflows
- [27074f61](https://github.com/kubedb/tests/commit/27074f613) Add CLAUDE.md pointing to AGENTS.md
- [e5e62b8a](https://github.com/kubedb/tests/commit/e5e62b8a0) Add AGENTS.md for AI coding agents
- [9ee1c9bc](https://github.com/kubedb/tests/commit/9ee1c9bc7) Harden CI workflows (#522)
- [af6fa871](https://github.com/kubedb/tests/commit/af6fa8712) Prepare for release v0.49.0 (#521)
- [7086f3e0](https://github.com/kubedb/tests/commit/7086f3e0e) Configure dependabot refresh schedule (#520)
- [dd86a543](https://github.com/kubedb/tests/commit/dd86a5437) Configure dependabot refresh schedule (#519)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.41.0-rc.0](https://github.com/kubedb/ui-server/releases/tag/v0.41.0-rc.0)

- [b0cae9e1](https://github.com/kubedb/ui-server/commit/b0cae9e1) Prepare for release v0.41.0-rc.0 (#203)
- [8633e481](https://github.com/kubedb/ui-server/commit/8633e481) Tighten CI/release workflow secrets, perms, and release notes
- [ed171989](https://github.com/kubedb/ui-server/commit/ed171989) Harden release and release-tracker workflows
- [1709ef94](https://github.com/kubedb/ui-server/commit/1709ef94) Pass componenetName field & Refactor (#201)
- [d171e2c8](https://github.com/kubedb/ui-server/commit/d171e2c8) Add AGENTS.md for AI coding agents
- [9e3b8aed](https://github.com/kubedb/ui-server/commit/9e3b8aed) Harden CI workflows (#200)
- [09c4dcc4](https://github.com/kubedb/ui-server/commit/09c4dcc4) Prepare for release v0.40.0 (#199)
- [996b2c98](https://github.com/kubedb/ui-server/commit/996b2c98) Configure dependabot refresh schedule (#198)



## [kubedb/weaviate](https://github.com/kubedb/weaviate)

### [v0.6.0-rc.0](https://github.com/kubedb/weaviate/releases/tag/v0.6.0-rc.0)

- [bb3c8918](https://github.com/kubedb/weaviate/commit/bb3c8918) Prepare for release v0.6.0-rc.0 (#35)
- [a77224f3](https://github.com/kubedb/weaviate/commit/a77224f3) Tighten CI/release workflow secrets, perms, and release notes
- [716556bd](https://github.com/kubedb/weaviate/commit/716556bd) Harden release and release-tracker workflows
- [7131fe7e](https://github.com/kubedb/weaviate/commit/7131fe7e) Add CLAUDE.md pointing to AGENTS.md
- [bb95826e](https://github.com/kubedb/weaviate/commit/bb95826e) Add AGENTS.md for AI coding agents
- [3dfc4fe3](https://github.com/kubedb/weaviate/commit/3dfc4fe3) Harden CI workflows (#29)
- [31218a35](https://github.com/kubedb/weaviate/commit/31218a35) Prepare for release v0.5.0 (#28)
- [4fc6dac2](https://github.com/kubedb/weaviate/commit/4fc6dac2) Configure dependabot refresh schedule (#27)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.41.0-rc.0](https://github.com/kubedb/webhook-server/releases/tag/v0.41.0-rc.0)




## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.13.0-rc.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.13.0-rc.0)

- [fd5f2106](https://github.com/kubedb/xtrabackup-restic-plugin/commit/fd5f2106) Prepare for release v0.13.0-rc.0 (#53)
- [1cf78111](https://github.com/kubedb/xtrabackup-restic-plugin/commit/1cf78111) Add AGENTS.md for AI coding agents
- [14f542a3](https://github.com/kubedb/xtrabackup-restic-plugin/commit/14f542a3) Use GitHub App token for release tracker comments (#51)
- [1c67d63e](https://github.com/kubedb/xtrabackup-restic-plugin/commit/1c67d63e) Prepare for release v0.12.0 (#49)
- [6be593b0](https://github.com/kubedb/xtrabackup-restic-plugin/commit/6be593b0) Bump RESTIC_VERSION to 0.18.1-20260421 (#48)
- [6199440e](https://github.com/kubedb/xtrabackup-restic-plugin/commit/6199440e) Configure dependabot refresh schedule (#47)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.20.0-rc.0](https://github.com/kubedb/zookeeper/releases/tag/v0.20.0-rc.0)

- [a572da63](https://github.com/kubedb/zookeeper/commit/a572da63) Prepare for release v0.20.0-rc.0 (#121)
- [f583941f](https://github.com/kubedb/zookeeper/commit/f583941f) Tighten CI/release workflow secrets, perms, and release notes
- [d4475edf](https://github.com/kubedb/zookeeper/commit/d4475edf) Harden release and release-tracker workflows
- [7a658949](https://github.com/kubedb/zookeeper/commit/7a658949) Add CLAUDE.md pointing to AGENTS.md
- [d06d8429](https://github.com/kubedb/zookeeper/commit/d06d8429) Add AGENTS.md for AI coding agents
- [36e65048](https://github.com/kubedb/zookeeper/commit/36e65048) Harden CI workflows (#119)
- [bd934f48](https://github.com/kubedb/zookeeper/commit/bd934f48) Prepare for release v0.19.0 (#118)
- [88cc9518](https://github.com/kubedb/zookeeper/commit/88cc9518) fix vol exp (#117)
- [bba031e5](https://github.com/kubedb/zookeeper/commit/bba031e5) Configure dependabot refresh schedule (#116)
- [f745f1a2](https://github.com/kubedb/zookeeper/commit/f745f1a2) Configure dependabot refresh schedule (#115)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.20.0-rc.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.20.0-rc.0)

- [5124f6a7](https://github.com/kubedb/zookeeper-restic-plugin/commit/5124f6a7) Prepare for release v0.20.0-rc.0 (#69)
- [f01b3b87](https://github.com/kubedb/zookeeper-restic-plugin/commit/f01b3b87) Tighten CI/release workflow secrets, perms, and release notes
- [85b763fa](https://github.com/kubedb/zookeeper-restic-plugin/commit/85b763fa) Harden release and release-tracker workflows
- [0cc56540](https://github.com/kubedb/zookeeper-restic-plugin/commit/0cc56540) Add AGENTS.md for AI coding agents
- [391b6eb7](https://github.com/kubedb/zookeeper-restic-plugin/commit/391b6eb7) Harden CI workflows (#67)
- [9df74bd0](https://github.com/kubedb/zookeeper-restic-plugin/commit/9df74bd0) Prepare for release v0.19.0 (#66)
- [e7662aa2](https://github.com/kubedb/zookeeper-restic-plugin/commit/e7662aa2) Bump RESTIC_VERSION to 0.18.1-20260421 (#65)
- [0ce9d404](https://github.com/kubedb/zookeeper-restic-plugin/commit/0ce9d404) Configure dependabot refresh schedule (#64)
- [5e490249](https://github.com/kubedb/zookeeper-restic-plugin/commit/5e490249) Configure dependabot refresh schedule (#63)




