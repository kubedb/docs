---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2026.6.19
    name: Changelog-v2026.6.19
    parent: welcome
    weight: 20260619
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.6.19/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.6.19/
---

# KubeDB v2026.6.19 (2026-06-20)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.65.0](https://github.com/kubedb/apimachinery/releases/tag/v0.65.0)

- [374e7dd6](https://github.com/kubedb/apimachinery/commit/374e7dd65) Update for release KubeStash@v2026.6.19 (#1780)
- [e726a79a](https://github.com/kubedb/apimachinery/commit/e726a79ab) register aerospike version (#1766)
- [63558553](https://github.com/kubedb/apimachinery/commit/635585536) Fix Milvus StorageMigration Webhook (#1779)
- [e3a3063f](https://github.com/kubedb/apimachinery/commit/e3a3063fd) Add cilium backup ingress policy and DNS egress for backup jobs (#1778)
- [d72310c7](https://github.com/kubedb/apimachinery/commit/d72310c73) Add weaviate TLS (#1772)
- [9b0689a2](https://github.com/kubedb/apimachinery/commit/9b0689a26) use value receiver (#1777)
- [fc958327](https://github.com/kubedb/apimachinery/commit/fc958327a) Add aerospike, db2, documentdb API types (#1774)
- [314ace89](https://github.com/kubedb/apimachinery/commit/314ace897) Add Aerospike feature gate and register AerospikeVersion in scheme (#1776)
- [42091fca](https://github.com/kubedb/apimachinery/commit/42091fca8) Add gitsyncer option in singlestore version api (#1773)
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
- [23c81b3b](https://github.com/kubedb/apimachinery/commit/23c81b3b6) Fix go.mod
- [007ef3ea](https://github.com/kubedb/apimachinery/commit/007ef3ea6) Remove FerretDB support (#1750)
- [101341c2](https://github.com/kubedb/apimachinery/commit/101341c29) Added wv-ops-validation and autoscaling webhook (#1722)
- [fb439fff](https://github.com/kubedb/apimachinery/commit/fb439fff4) Register Neo4j Autoscaler (#1746)
- [62f69e28](https://github.com/kubedb/apimachinery/commit/62f69e286) Add CiliumNetworkPolicy flavor support (#1714)
- [9b1a220b](https://github.com/kubedb/apimachinery/commit/9b1a220b4) Bump go.bytebuilders.dev/audit to v0.0.52 (#1745)
- [dd3d0e3d](https://github.com/kubedb/apimachinery/commit/dd3d0e3d3) Update go.bytebuilders.dev/audit to v0.0.51 (#1744)
- [0e48a4da](https://github.com/kubedb/apimachinery/commit/0e48a4da6) Add common configuration for milvus storage migration (#1741)
- [ea56f8dd](https://github.com/kubedb/apimachinery/commit/ea56f8dd2) Add pkg/secret helpers for dual-path auth secret access (#1726)
- [ef57da77](https://github.com/kubedb/apimachinery/commit/ef57da777) Fix DatabaseConfiguration resource name casing and remove stale CRDs (#1727)
- [9bdcc3b7](https://github.com/kubedb/apimachinery/commit/9bdcc3b7e) DatabaseInfo -> DatabaseConfiguration (#1725)
- [66b31784](https://github.com/kubedb/apimachinery/commit/66b31784e) Introduce summary api (#1724)
- [59b4eaa3](https://github.com/kubedb/apimachinery/commit/59b4eaa36) Add weaviate ops helpers
- [f7c72195](https://github.com/kubedb/apimachinery/commit/f7c721957) Review autoscaler api (#1719)
- [97db127e](https://github.com/kubedb/apimachinery/commit/97db127e6) Update Elasticsearch StorageMigration Ops Api (#1721)
- [7b38d914](https://github.com/kubedb/apimachinery/commit/7b38d9144) Add Milvus Ops Request (#1666)
- [46001228](https://github.com/kubedb/apimachinery/commit/46001228f) add weaviate ops_req (#1716)
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



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.50.0](https://github.com/kubedb/autoscaler/releases/tag/v0.50.0)

- [f6ba9830](https://github.com/kubedb/autoscaler/commit/f6ba9830) Prepare for release v0.50.0 (#309)
- [764b0a49](https://github.com/kubedb/autoscaler/commit/764b0a49) Add HanaDB autoscaler support (#300)
- [be2cf3bc](https://github.com/kubedb/autoscaler/commit/be2cf3bc) Add Autoscaling support for DocumentDB (#299)
- [416c01f8](https://github.com/kubedb/autoscaler/commit/416c01f8) Add support for Oracle (autoscaler) (#302)
- [c7169bb5](https://github.com/kubedb/autoscaler/commit/c7169bb5) Prepare for release v0.50.0-rc.2 (#308)
- [001c9856](https://github.com/kubedb/autoscaler/commit/001c9856) feat: add Weaviate compute and storage autoscaler (#301)
- [a51c5af2](https://github.com/kubedb/autoscaler/commit/a51c5af2) Add Milvus Autoscaler (#306)
- [edcd29ac](https://github.com/kubedb/autoscaler/commit/edcd29ac) Add support for Neo4j (#298)
- [9d983c07](https://github.com/kubedb/autoscaler/commit/9d983c07) Prepare for release v0.50.0-rc.1 (#305)
- [9d7909f8](https://github.com/kubedb/autoscaler/commit/9d7909f8) Remove FerretDB support (#304)
- [9366ad79](https://github.com/kubedb/autoscaler/commit/9366ad79) Set Ops Request Options Default (#293)
- [587219ce](https://github.com/kubedb/autoscaler/commit/587219ce) Add Qdrant Autoscaler (#285)
- [0dc951b2](https://github.com/kubedb/autoscaler/commit/0dc951b2) Prepare for release v0.50.0-rc.0 (#303)
- [9ca97e5f](https://github.com/kubedb/autoscaler/commit/9ca97e5f) Tighten CI/release workflow secrets, perms, and release notes
- [1e275efc](https://github.com/kubedb/autoscaler/commit/1e275efc) Harden release and release-tracker workflows
- [f296dee8](https://github.com/kubedb/autoscaler/commit/f296dee8) feat: read PVC storage metrics from custom metrics API
- [3542cf19](https://github.com/kubedb/autoscaler/commit/3542cf19) Add AGENTS.md for AI coding agents
- [3f8899bf](https://github.com/kubedb/autoscaler/commit/3f8899bf) Use GitHub App token for release tracker comments (#295)
- [cf6b3e5f](https://github.com/kubedb/autoscaler/commit/cf6b3e5f) Harden CI workflows (#292)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.18.0](https://github.com/kubedb/cassandra/releases/tag/v0.18.0)

- [e5733596](https://github.com/kubedb/cassandra/commit/e5733596) Prepare for release v0.18.0 (#90)
- [a14d2587](https://github.com/kubedb/cassandra/commit/a14d2587) Prepare for release v0.18.0-rc.2 (#89)
- [5f55678b](https://github.com/kubedb/cassandra/commit/5f55678b) Honor user-provided renewBefore in TLS certificate ops (#85)
- [602bdf49](https://github.com/kubedb/cassandra/commit/602bdf49) Add StorageMigration OpsRequest support for Cassandra (#81)
- [ef021693](https://github.com/kubedb/cassandra/commit/ef021693) Add NetworkPolicyFlavor support for cilium (#88)
- [4855098e](https://github.com/kubedb/cassandra/commit/4855098e) Prepare for release v0.18.0-rc.1 (#87)
- [a728b7f6](https://github.com/kubedb/cassandra/commit/a728b7f6) Prepare for release v0.18.0-rc.0 (#82)
- [0363c990](https://github.com/kubedb/cassandra/commit/0363c990) Tighten CI/release workflow secrets, perms, and release notes
- [995e9f8a](https://github.com/kubedb/cassandra/commit/995e9f8a) Harden release and release-tracker workflows
- [eaa2a39b](https://github.com/kubedb/cassandra/commit/eaa2a39b) Add CLAUDE.md pointing to AGENTS.md
- [cbc64d19](https://github.com/kubedb/cassandra/commit/cbc64d19) Add AGENTS.md for AI coding agents
- [fcc9a127](https://github.com/kubedb/cassandra/commit/fcc9a127) Harden CI workflows (#79)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.12.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.12.0)

- [eba85ff4](https://github.com/kubedb/cassandra-medusa-plugin/commit/eba85ff4) Prepare for release v0.12.0 (#41)
- [2584476c](https://github.com/kubedb/cassandra-medusa-plugin/commit/2584476c) Prepare for release v0.12.0-rc.2 (#40)
- [aefdc390](https://github.com/kubedb/cassandra-medusa-plugin/commit/aefdc390) Prepare for release v0.12.0-rc.1 (#39)
- [29127ccd](https://github.com/kubedb/cassandra-medusa-plugin/commit/29127ccd) Prepare for release v0.12.0-rc.0 (#38)
- [972cc3ec](https://github.com/kubedb/cassandra-medusa-plugin/commit/972cc3ec) Harden release and release-tracker workflows
- [f5566ac1](https://github.com/kubedb/cassandra-medusa-plugin/commit/f5566ac1) Add AGENTS.md for AI coding agents
- [494bac88](https://github.com/kubedb/cassandra-medusa-plugin/commit/494bac88) Harden CI workflows (#36)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.20.0](https://github.com/kubedb/clickhouse/releases/tag/v0.20.0)

- [d347611d](https://github.com/kubedb/clickhouse/commit/d347611d) Prepare for release v0.20.0 (#116)
- [01ed2962](https://github.com/kubedb/clickhouse/commit/01ed2962) Update github.com/moby/spdystream to v0.5.1 (#115)
- [d5a8072e](https://github.com/kubedb/clickhouse/commit/d5a8072e) feat: implement git-sync init container for ClickHouse (#109)
- [85f38620](https://github.com/kubedb/clickhouse/commit/85f38620) Add StorageMigration OpsRequest support for ClickHouse (#106)
- [6f1712d6](https://github.com/kubedb/clickhouse/commit/6f1712d6) Prepare for release v0.20.0-rc.2 (#114)
- [610759d3](https://github.com/kubedb/clickhouse/commit/610759d3) Honor user-provided renewBefore in TLS certificate ops (#110)
- [b8b9b4d5](https://github.com/kubedb/clickhouse/commit/b8b9b4d5) Add NetworkPolicyFlavor support (#113)
- [18935285](https://github.com/kubedb/clickhouse/commit/18935285) Prepare for release v0.20.0-rc.1 (#112)
- [42932dad](https://github.com/kubedb/clickhouse/commit/42932dad) Prepare for release v0.20.0-rc.0 (#107)
- [5c3c401c](https://github.com/kubedb/clickhouse/commit/5c3c401c) Tighten CI/release workflow secrets, perms, and release notes
- [a0a93b59](https://github.com/kubedb/clickhouse/commit/a0a93b59) Harden release and release-tracker workflows
- [2c864737](https://github.com/kubedb/clickhouse/commit/2c864737) Add CLAUDE.md pointing to AGENTS.md
- [001c11d6](https://github.com/kubedb/clickhouse/commit/001c11d6) Add AGENTS.md for AI coding agents
- [f4a08fd1](https://github.com/kubedb/clickhouse/commit/f4a08fd1) Restrict /ok-to-test to org members



## [kubedb/clickhouse-backup-plugin](https://github.com/kubedb/clickhouse-backup-plugin)

### [v0.2.0](https://github.com/kubedb/clickhouse-backup-plugin/releases/tag/v0.2.0)

- [e566d6a8](https://github.com/kubedb/clickhouse-backup-plugin/commit/e566d6a8) Prepare for release v0.2.0 (#28)
- [bd2b5438](https://github.com/kubedb/clickhouse-backup-plugin/commit/bd2b5438) Update github.com/moby/spdystream to v0.5.1 (#27)
- [5d481e4c](https://github.com/kubedb/clickhouse-backup-plugin/commit/5d481e4c) Prepare for release v0.2.0-rc.2 (#26)
- [8ee159c7](https://github.com/kubedb/clickhouse-backup-plugin/commit/8ee159c7) Prepare for release v0.2.0-rc.1 (#25)
- [0a880684](https://github.com/kubedb/clickhouse-backup-plugin/commit/0a880684) Fix CI hardening: use app token in release-tracker, add packages:write
- [8a2fd2fa](https://github.com/kubedb/clickhouse-backup-plugin/commit/8a2fd2fa) Prepare for release v0.2.0-rc.0 (#23)
- [a4fdf4d8](https://github.com/kubedb/clickhouse-backup-plugin/commit/a4fdf4d8) Add AGENTS.md for AI coding agents
- [0d0bca21](https://github.com/kubedb/clickhouse-backup-plugin/commit/0d0bca21) Use GitHub App token for release tracker comments



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.20.0](https://github.com/kubedb/crd-manager/releases/tag/v0.20.0)

- [9588532f](https://github.com/kubedb/crd-manager/commit/9588532f) Prepare for release v0.20.0 (#144)
- [7bad63d9](https://github.com/kubedb/crd-manager/commit/7bad63d9) Remove Duplicate Cassandra (#143)
- [bbd8d7db](https://github.com/kubedb/crd-manager/commit/bbd8d7db) Add Aerospike CRDs, complete DB2 ops/autoscaling, wire GitOps for all DBs (#142)
- [74b20efd](https://github.com/kubedb/crd-manager/commit/74b20efd) Add Aerospike database support (#141)
- [d9cd2e04](https://github.com/kubedb/crd-manager/commit/d9cd2e04) Prepare for release v0.20.0-rc.2 (#140)
- [f63e3c4d](https://github.com/kubedb/crd-manager/commit/f63e3c4d) Prepare for release v0.20.0-rc.1 (#139)
- [33485516](https://github.com/kubedb/crd-manager/commit/33485516) Remove FerretDB support (#138)
- [f4f684e9](https://github.com/kubedb/crd-manager/commit/f4f684e9) Add all missing CRDs (#137)
- [2fb5a23a](https://github.com/kubedb/crd-manager/commit/2fb5a23a) Tighten CI/release workflow secrets, perms, and release notes (#136)
- [2d75e6b5](https://github.com/kubedb/crd-manager/commit/2d75e6b5) Prepare for release v0.20.0-rc.0 (#135)
- [464e548c](https://github.com/kubedb/crd-manager/commit/464e548c) Harden release and release-tracker workflows
- [0e5c52c6](https://github.com/kubedb/crd-manager/commit/0e5c52c6) Add clickhouse archiver CR (#134)
- [3362d975](https://github.com/kubedb/crd-manager/commit/3362d975) add qdrant autoscaler crd (#131)
- [ffe49332](https://github.com/kubedb/crd-manager/commit/ffe49332) Add AGENTS.md for AI coding agents
- [02872ff1](https://github.com/kubedb/crd-manager/commit/02872ff1) Harden CI workflows (#132)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.23.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.23.0)

- [df998786](https://github.com/kubedb/dashboard-restic-plugin/commit/df998786) Prepare for release v0.23.0 (#80)
- [cdd75695](https://github.com/kubedb/dashboard-restic-plugin/commit/cdd75695) Prepare for release v0.23.0-rc.2 (#79)
- [c3afe4cd](https://github.com/kubedb/dashboard-restic-plugin/commit/c3afe4cd) Add restic backup progress streaming (#78)
- [dbe58a1d](https://github.com/kubedb/dashboard-restic-plugin/commit/dbe58a1d) Prepare for release v0.23.0-rc.1 (#77)
- [87c14d95](https://github.com/kubedb/dashboard-restic-plugin/commit/87c14d95) Prepare for release v0.23.0-rc.0 (#76)
- [f26f8aab](https://github.com/kubedb/dashboard-restic-plugin/commit/f26f8aab) Harden release and release-tracker workflows
- [da5af73f](https://github.com/kubedb/dashboard-restic-plugin/commit/da5af73f) Add AGENTS.md for AI coding agents
- [8e800912](https://github.com/kubedb/dashboard-restic-plugin/commit/8e800912) Harden CI workflows (#74)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.20.0](https://github.com/kubedb/db-client-go/releases/tag/v0.20.0)

- [39d47608](https://github.com/kubedb/db-client-go/commit/39d47608) Prepare for release v0.20.0 (#250)
- [34229270](https://github.com/kubedb/db-client-go/commit/34229270) Add weaviate tls (#248)
- [ff2e61c8](https://github.com/kubedb/db-client-go/commit/ff2e61c8) Prepare for release v0.20.0-rc.2 (#249)
- [b0c28745](https://github.com/kubedb/db-client-go/commit/b0c28745) Add HanaDB TLS (#234)
- [d97925e7](https://github.com/kubedb/db-client-go/commit/d97925e7) Add New Func to Neo4j (#241)
- [a24ee1cd](https://github.com/kubedb/db-client-go/commit/a24ee1cd) Prepare for release v0.20.0-rc.1 (#247)
- [0efcbce1](https://github.com/kubedb/db-client-go/commit/0efcbce1) Use shared pkg/secret helpers for dual-path auth secret access (#245)
- [0878c362](https://github.com/kubedb/db-client-go/commit/0878c362) Bump kubedb.dev/apimachinery to drop FerretDB (#246)
- [523e0304](https://github.com/kubedb/db-client-go/commit/523e0304) Tighten CI/release workflow secrets, perms, and release notes
- [3caab860](https://github.com/kubedb/db-client-go/commit/3caab860) Prepare for release v0.20.0-rc.0 (#244)
- [2b9e043c](https://github.com/kubedb/db-client-go/commit/2b9e043c) Harden release and release-tracker workflows
- [462a2a68](https://github.com/kubedb/db-client-go/commit/462a2a68) Qdrant HTTP Client TLS (#239)
- [3460fe00](https://github.com/kubedb/db-client-go/commit/3460fe00) Update for distributed postgres (#243)
- [df0e92a0](https://github.com/kubedb/db-client-go/commit/df0e92a0) Add Milvus TLS (#232)
- [e122e57f](https://github.com/kubedb/db-client-go/commit/e122e57f) Add AGENTS.md for AI coding agents
- [465442b7](https://github.com/kubedb/db-client-go/commit/465442b7) Harden CI workflows (#240)



## [kubedb/db2](https://github.com/kubedb/db2)

### [v0.6.0](https://github.com/kubedb/db2/releases/tag/v0.6.0)

- [675c40a6](https://github.com/kubedb/db2/commit/675c40a6) Prepare for release v0.6.0 (#33)
- [99854604](https://github.com/kubedb/db2/commit/99854604) Prepare for release v0.6.0-rc.2 (#32)
- [f2fe477e](https://github.com/kubedb/db2/commit/f2fe477e) Fix Db2 deletion (#31)
- [e263fee1](https://github.com/kubedb/db2/commit/e263fee1) Add NetworkPolicyFlavor support (#30)
- [deec0ca0](https://github.com/kubedb/db2/commit/deec0ca0) Prepare for release v0.6.0-rc.1 (#29)
- [83b7b475](https://github.com/kubedb/db2/commit/83b7b475) Prepare for release v0.6.0-rc.0 (#25)
- [ea12a946](https://github.com/kubedb/db2/commit/ea12a946) Tighten CI/release workflow secrets, perms, and release notes
- [2151e287](https://github.com/kubedb/db2/commit/2151e287) Harden release and release-tracker workflows
- [0493419f](https://github.com/kubedb/db2/commit/0493419f) Add CLAUDE.md pointing to AGENTS.md
- [ee433a7e](https://github.com/kubedb/db2/commit/ee433a7e) Add AGENTS.md for AI coding agents
- [14a0f81b](https://github.com/kubedb/db2/commit/14a0f81b) Harden CI workflows (#23)



## [kubedb/db2-coordinator](https://github.com/kubedb/db2-coordinator)

### [v0.6.0](https://github.com/kubedb/db2-coordinator/releases/tag/v0.6.0)

- [b2f233d](https://github.com/kubedb/db2-coordinator/commit/b2f233d) Prepare for release v0.6.0 (#14)
- [9f4c409](https://github.com/kubedb/db2-coordinator/commit/9f4c409) Prepare for release v0.6.0-rc.2 (#13)
- [55e13e2](https://github.com/kubedb/db2-coordinator/commit/55e13e2) Fix version detection and import kubedb.dev/apimachinery (#12)
- [a1e57d7](https://github.com/kubedb/db2-coordinator/commit/a1e57d7) Tighten CI/release workflow secrets, perms, and release notes
- [d6b3d18](https://github.com/kubedb/db2-coordinator/commit/d6b3d18) Harden release and release-tracker workflows
- [ac8ef66](https://github.com/kubedb/db2-coordinator/commit/ac8ef66) Add AGENTS.md for AI coding agents
- [45f3f3d](https://github.com/kubedb/db2-coordinator/commit/45f3f3d) Harden CI workflows (#10)



## [kubedb/documentdb](https://github.com/kubedb/documentdb)

### [v0.2.0](https://github.com/kubedb/documentdb/releases/tag/v0.2.0)

- [f1c7d235](https://github.com/kubedb/documentdb/commit/f1c7d235) Prepare for release v0.2.0 (#29)
- [dd014eb4](https://github.com/kubedb/documentdb/commit/dd014eb4) Update github.com/moby/spdystream to v0.5.1 (#28)
- [153e7117](https://github.com/kubedb/documentdb/commit/153e7117) Prepare for release v0.2.0-rc.2 (#27)
- [d5322e67](https://github.com/kubedb/documentdb/commit/d5322e67) documentdb-reconfigure (#25)
- [148c6fb0](https://github.com/kubedb/documentdb/commit/148c6fb0) Add OpsRequest support for DocumentDB ported from Postgres (#22)
- [999aef78](https://github.com/kubedb/documentdb/commit/999aef78) bring reverted changes (#24)
- [0195ac68](https://github.com/kubedb/documentdb/commit/0195ac68) Update apimachinery (#23)
- [846bca49](https://github.com/kubedb/documentdb/commit/846bca49) Update apimachinery (#21)
- [dbeb9d3c](https://github.com/kubedb/documentdb/commit/dbeb9d3c) Clustering  (#7)
- [f34712d9](https://github.com/kubedb/documentdb/commit/f34712d9) Add NetworkPolicyFlavor support (#19)
- [481e035d](https://github.com/kubedb/documentdb/commit/481e035d) Prepare for release v0.2.0-rc.1 (#18)
- [8ff7cc98](https://github.com/kubedb/documentdb/commit/8ff7cc98) Prepare for release v0.2.0-rc.0 (#15)
- [aa27c166](https://github.com/kubedb/documentdb/commit/aa27c166) Tighten CI/release workflow secrets, perms, and release notes
- [9c8eb095](https://github.com/kubedb/documentdb/commit/9c8eb095) removed default password (#14)
- [76f64d15](https://github.com/kubedb/documentdb/commit/76f64d15) Harden release and release-tracker workflows
- [56ea3118](https://github.com/kubedb/documentdb/commit/56ea3118) Add AGENTS.md for AI coding agents
- [2e08c01c](https://github.com/kubedb/documentdb/commit/2e08c01c) Harden CI workflows (#11)



## [kubedb/documentdb-coordinator](https://github.com/kubedb/documentdb-coordinator)

### [v0.1.0](https://github.com/kubedb/documentdb-coordinator/releases/tag/v0.1.0)

- [41cc779](https://github.com/kubedb/documentdb-coordinator/commit/41cc779) Prepare for release v0.1.0 (#6)
- [bd5178a](https://github.com/kubedb/documentdb-coordinator/commit/bd5178a) Fix UBI Dockerfile labels and add bash/postgresql dependency (#5)
- [ee91f72](https://github.com/kubedb/documentdb-coordinator/commit/ee91f72) Use gh cli instead of old hub cli
- [05a2c8b](https://github.com/kubedb/documentdb-coordinator/commit/05a2c8b) Prepare for release v0.1.0-rc.2 (#4)
- [020634d](https://github.com/kubedb/documentdb-coordinator/commit/020634d) Fix DocumentDBCoordinatorClientPort (2389 → 2379) (#3)
- [4b8cbae](https://github.com/kubedb/documentdb-coordinator/commit/4b8cbae) Up deps on apimachinery (#2)
- [d2f2e7b](https://github.com/kubedb/documentdb-coordinator/commit/d2f2e7b) Bootstrap (#1)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.20.0](https://github.com/kubedb/druid/releases/tag/v0.20.0)

- [55331e3f](https://github.com/kubedb/druid/commit/55331e3f) Prepare for release v0.20.0 (#141)
- [7ea60517](https://github.com/kubedb/druid/commit/7ea60517) Prepare for release v0.20.0-rc.2 (#140)
- [89371af1](https://github.com/kubedb/druid/commit/89371af1) Honor user-provided renewBefore in TLS certificate ops (#135)
- [f4f4627f](https://github.com/kubedb/druid/commit/f4f4627f) Fix Panic Issue For ExternallyManaged authSecret (#139)
- [e89ff10a](https://github.com/kubedb/druid/commit/e89ff10a) Add NetworkPolicyFlavor support (#138)
- [7970cf8b](https://github.com/kubedb/druid/commit/7970cf8b) Prepare for release v0.20.0-rc.1 (#137)
- [e35bc8a6](https://github.com/kubedb/druid/commit/e35bc8a6) Prepare for release v0.20.0-rc.0 (#132)
- [d8f0c94f](https://github.com/kubedb/druid/commit/d8f0c94f) Tighten CI/release workflow secrets, perms, and release notes
- [81bde5dd](https://github.com/kubedb/druid/commit/81bde5dd) Implement StorageMigration OpsRequest for Druid (#131)
- [fc0a7a2b](https://github.com/kubedb/druid/commit/fc0a7a2b) Harden release and release-tracker workflows
- [bb6a95d0](https://github.com/kubedb/druid/commit/bb6a95d0) Add CLAUDE.md pointing to AGENTS.md
- [07b19896](https://github.com/kubedb/druid/commit/07b19896) Add AGENTS.md for AI coding agents
- [5d02b3ca](https://github.com/kubedb/druid/commit/5d02b3ca) Harden CI workflows (#129)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.65.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.65.0)

- [be47268d](https://github.com/kubedb/elasticsearch/commit/be47268d7) Prepare for release v0.65.0 (#820)
- [5f17098d](https://github.com/kubedb/elasticsearch/commit/5f17098d6) Update github.com/moby/spdystream to v0.5.1 (#819)
- [b0421f8d](https://github.com/kubedb/elasticsearch/commit/b0421f8d3) Prepare for release v0.65.0-rc.2 (#818)
- [7a09aa83](https://github.com/kubedb/elasticsearch/commit/7a09aa83c) Honor user-provided renewBefore in TLS certificate ops (#815)
- [1c3cb34f](https://github.com/kubedb/elasticsearch/commit/1c3cb34f1) feat: implement git-sync init container for Elasticsearch (#814)
- [890e13c1](https://github.com/kubedb/elasticsearch/commit/890e13c16) Prepare for release v0.65.0-rc.1 (#817)
- [613e034e](https://github.com/kubedb/elasticsearch/commit/613e034ea) Add StorageMigration OpsRequest support (#810)
- [ecddd1f6](https://github.com/kubedb/elasticsearch/commit/ecddd1f68) Prepare for release v0.65.0-rc.0 (#812)
- [fc42205f](https://github.com/kubedb/elasticsearch/commit/fc42205fa) Tighten CI/release workflow secrets, perms, and release notes
- [df2dae4a](https://github.com/kubedb/elasticsearch/commit/df2dae4a0) Harden release and release-tracker workflows
- [a134aed4](https://github.com/kubedb/elasticsearch/commit/a134aed4a) Run Ops Request Locally (#811)
- [02b16b73](https://github.com/kubedb/elasticsearch/commit/02b16b73b) Add CLAUDE.md pointing to AGENTS.md
- [65db23b7](https://github.com/kubedb/elasticsearch/commit/65db23b70) Add AGENTS.md for AI coding agents
- [ebf149db](https://github.com/kubedb/elasticsearch/commit/ebf149dba) Harden CI workflows (#808)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.28.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.28.0)

- [65522b4d](https://github.com/kubedb/elasticsearch-restic-plugin/commit/65522b4d) Prepare for release v0.28.0 (#103)
- [ac051dbf](https://github.com/kubedb/elasticsearch-restic-plugin/commit/ac051dbf) Prepare for release v0.28.0-rc.2 (#102)
- [021ea97a](https://github.com/kubedb/elasticsearch-restic-plugin/commit/021ea97a) Add restic backup progress streaming (#101)
- [650b7682](https://github.com/kubedb/elasticsearch-restic-plugin/commit/650b7682) Prepare for release v0.28.0-rc.1 (#100)
- [f7bce7aa](https://github.com/kubedb/elasticsearch-restic-plugin/commit/f7bce7aa) Prepare for release v0.28.0-rc.0 (#99)
- [b38a1c55](https://github.com/kubedb/elasticsearch-restic-plugin/commit/b38a1c55) Harden release and release-tracker workflows
- [853d30a7](https://github.com/kubedb/elasticsearch-restic-plugin/commit/853d30a7) Add AGENTS.md for AI coding agents
- [8b4b3ac0](https://github.com/kubedb/elasticsearch-restic-plugin/commit/8b4b3ac0) Harden CI workflows (#97)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.13.0](https://github.com/kubedb/gitops/releases/tag/v0.13.0)

- [a6540e22](https://github.com/kubedb/gitops/commit/a6540e22) Prepare for release v0.13.0 (#81)
- [5a3af661](https://github.com/kubedb/gitops/commit/5a3af661) Add GitOps support for Solr (#77)
- [f7a4e119](https://github.com/kubedb/gitops/commit/f7a4e119) Add GitOps support for Druid (#71)
- [126b3544](https://github.com/kubedb/gitops/commit/126b3544) Add gitops support for ClickHouse (#60)
- [3857b522](https://github.com/kubedb/gitops/commit/3857b522) Add GitOps support for PerconaXtraDB (#69)
- [98f16c54](https://github.com/kubedb/gitops/commit/98f16c54) Add GitOps support for Singlestore (#78)
- [32a65cd4](https://github.com/kubedb/gitops/commit/32a65cd4) Prepare for release v0.13.0-rc.2 (#80)
- [0131a5f7](https://github.com/kubedb/gitops/commit/0131a5f7) Add GitOps support for RabbitMQ (#72)
- [a7f78221](https://github.com/kubedb/gitops/commit/a7f78221) Add gitops support for Neo4j (#62)
- [cba199d7](https://github.com/kubedb/gitops/commit/cba199d7) Skip OpsCreation If Any Same Type Ops InProgress (#79)
- [0da911fa](https://github.com/kubedb/gitops/commit/0da911fa) Prepare for release v0.13.0-rc.1 (#68)
- [a5a97ff5](https://github.com/kubedb/gitops/commit/a5a97ff5) Prepare for release v0.13.0-rc.0 (#57)
- [da0e1032](https://github.com/kubedb/gitops/commit/da0e1032) Tighten CI/release workflow secrets, perms, and release notes
- [baab2fac](https://github.com/kubedb/gitops/commit/baab2fac) Fix Recurring Ops Creation for ReconfigeTLS (#55)
- [40969d95](https://github.com/kubedb/gitops/commit/40969d95) Harden release and release-tracker workflows
- [e72a0e12](https://github.com/kubedb/gitops/commit/e72a0e12) Add CLAUDE.md pointing to AGENTS.md
- [7152e2d5](https://github.com/kubedb/gitops/commit/7152e2d5) Add AGENTS.md for AI coding agents
- [c7a5c666](https://github.com/kubedb/gitops/commit/c7a5c666) Harden CI workflows (#54)



## [kubedb/hanadb](https://github.com/kubedb/hanadb)

### [v0.6.0](https://github.com/kubedb/hanadb/releases/tag/v0.6.0)

- [fe71d640](https://github.com/kubedb/hanadb/commit/fe71d640) Prepare for release v0.6.0 (#46)
- [b3f6ef5b](https://github.com/kubedb/hanadb/commit/b3f6ef5b) Prepare for release v0.6.0-rc.2 (#45)
- [e38a3433](https://github.com/kubedb/hanadb/commit/e38a3433) Add StorageMigration OpsRequest support (#34)
- [c08ac12e](https://github.com/kubedb/hanadb/commit/c08ac12e) Add tls, reconfigure tls, vertical scaling, rotate auth, volume expantion ops (#38)
- [696ffce2](https://github.com/kubedb/hanadb/commit/696ffce2) Add NetworkPolicyFlavor support (#44)
- [da2fb620](https://github.com/kubedb/hanadb/commit/da2fb620) Prepare for release v0.6.0-rc.1 (#43)
- [5f480c10](https://github.com/kubedb/hanadb/commit/5f480c10) Prepare for release v0.6.0-rc.0 (#39)
- [28694dd1](https://github.com/kubedb/hanadb/commit/28694dd1) Tighten CI/release workflow secrets, perms, and release notes
- [9fb48705](https://github.com/kubedb/hanadb/commit/9fb48705) Harden release and release-tracker workflows
- [d530b4af](https://github.com/kubedb/hanadb/commit/d530b4af) Add AGENTS.md for AI coding agents
- [52612af6](https://github.com/kubedb/hanadb/commit/52612af6) Harden CI workflows (#32)



## [kubedb/hanadb-coordinator](https://github.com/kubedb/hanadb-coordinator)

### [v0.5.0](https://github.com/kubedb/hanadb-coordinator/releases/tag/v0.5.0)

- [d38d2fdf](https://github.com/kubedb/hanadb-coordinator/commit/d38d2fdf) Prepare for release v0.5.0 (#18)
- [78c65a5d](https://github.com/kubedb/hanadb-coordinator/commit/78c65a5d) Update github.com/moby/spdystream to v0.5.1 (#17)
- [51502038](https://github.com/kubedb/hanadb-coordinator/commit/51502038) Prepare for release v0.5.0-rc.2 (#16)
- [5371f1fa](https://github.com/kubedb/hanadb-coordinator/commit/5371f1fa) Disable automatic backups and fix failover handling (#15)
- [97c79d78](https://github.com/kubedb/hanadb-coordinator/commit/97c79d78) Prepare for release v0.5.0-rc.1 (#14)
- [8d68a1ab](https://github.com/kubedb/hanadb-coordinator/commit/8d68a1ab) Prepare for release v0.5.0-rc.0 (#13)
- [a1df343c](https://github.com/kubedb/hanadb-coordinator/commit/a1df343c) Tighten CI/release workflow secrets, perms, and release notes
- [de35968d](https://github.com/kubedb/hanadb-coordinator/commit/de35968d) Harden release and release-tracker workflows
- [0a5612a4](https://github.com/kubedb/hanadb-coordinator/commit/0a5612a4) Add AGENTS.md for AI coding agents
- [f8cd4851](https://github.com/kubedb/hanadb-coordinator/commit/f8cd4851) Harden CI workflows (#11)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.11.0](https://github.com/kubedb/hazelcast/releases/tag/v0.11.0)

- [47da8055](https://github.com/kubedb/hazelcast/commit/47da8055) Prepare for release v0.11.0 (#53)
- [33821545](https://github.com/kubedb/hazelcast/commit/33821545) Fix Hazelcast deletion (#51)
- [841ee968](https://github.com/kubedb/hazelcast/commit/841ee968) Prepare for release v0.11.0-rc.2 (#52)
- [13deaa26](https://github.com/kubedb/hazelcast/commit/13deaa26) Add NetworkPolicyFlavor support (#50)
- [e45daa28](https://github.com/kubedb/hazelcast/commit/e45daa28) Prepare for release v0.11.0-rc.1 (#49)
- [8cfb374f](https://github.com/kubedb/hazelcast/commit/8cfb374f) Prepare for release v0.11.0-rc.0 (#46)
- [198c8a50](https://github.com/kubedb/hazelcast/commit/198c8a50) Tighten CI/release workflow secrets, perms, and release notes
- [f31ab7cc](https://github.com/kubedb/hazelcast/commit/f31ab7cc) Harden release and release-tracker workflows
- [791b67dd](https://github.com/kubedb/hazelcast/commit/791b67dd) Add CLAUDE.md pointing to AGENTS.md
- [289a07ef](https://github.com/kubedb/hazelcast/commit/289a07ef) Add AGENTS.md for AI coding agents
- [f8e2fc26](https://github.com/kubedb/hazelcast/commit/f8e2fc26) Harden CI workflows (#43)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.12.0](https://github.com/kubedb/ignite/releases/tag/v0.12.0)

- [1679ffc5](https://github.com/kubedb/ignite/commit/1679ffc5) Prepare for release v0.12.0 (#62)
- [db56d3e9](https://github.com/kubedb/ignite/commit/db56d3e9) Add StorageMigration OpsRequest support for Ignite (#53)
- [abaa3f61](https://github.com/kubedb/ignite/commit/abaa3f61) Prepare for release v0.12.0-rc.2 (#61)
- [0e4fa9e6](https://github.com/kubedb/ignite/commit/0e4fa9e6) Fix Ignite deletion (#60)
- [1ccf34ce](https://github.com/kubedb/ignite/commit/1ccf34ce) Add NetworkPolicyFlavor support for cilium (#59)
- [46fc9cb0](https://github.com/kubedb/ignite/commit/46fc9cb0) Prepare for release v0.12.0-rc.1 (#58)
- [2ae50f33](https://github.com/kubedb/ignite/commit/2ae50f33) Prepare for release v0.12.0-rc.0 (#54)
- [b04edbc9](https://github.com/kubedb/ignite/commit/b04edbc9) Tighten CI/release workflow secrets, perms, and release notes
- [b966bfea](https://github.com/kubedb/ignite/commit/b966bfea) Harden release and release-tracker workflows
- [2b92c194](https://github.com/kubedb/ignite/commit/2b92c194) Add CLAUDE.md pointing to AGENTS.md
- [a44aad7e](https://github.com/kubedb/ignite/commit/a44aad7e) Add AGENTS.md for AI coding agents
- [9848ac9f](https://github.com/kubedb/ignite/commit/9848ac9f) Harden CI workflows (#51)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2026.6.19](https://github.com/kubedb/installer/releases/tag/v2026.6.19)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.36.0](https://github.com/kubedb/kafka/releases/tag/v0.36.0)

- [d10d9665](https://github.com/kubedb/kafka/commit/d10d9665) Prepare for release v0.36.0 (#203)
- [da458b15](https://github.com/kubedb/kafka/commit/da458b15) Prepare for release v0.36.0-rc.2 (#202)
- [d95e5cbc](https://github.com/kubedb/kafka/commit/d95e5cbc) Honor user-provided renewBefore in TLS certificate ops (#198)
- [8ac2cd7b](https://github.com/kubedb/kafka/commit/8ac2cd7b) Add NetworkPolicyFlavor support for cilium (#201)
- [a1068e0b](https://github.com/kubedb/kafka/commit/a1068e0b) Prepare for release v0.36.0-rc.1 (#200)
- [c6a0dee3](https://github.com/kubedb/kafka/commit/c6a0dee3) Add StorageMigration OpsRequest support (#194)
- [af590f7a](https://github.com/kubedb/kafka/commit/af590f7a) Prepare for release v0.36.0-rc.0 (#196)
- [1182e416](https://github.com/kubedb/kafka/commit/1182e416) Tighten CI/release workflow secrets, perms, and release notes
- [bc7be0ce](https://github.com/kubedb/kafka/commit/bc7be0ce) Harden release and release-tracker workflows
- [2bfcd4c6](https://github.com/kubedb/kafka/commit/2bfcd4c6) Run Ops Request Locally (#195)
- [7cbac0ff](https://github.com/kubedb/kafka/commit/7cbac0ff) Add CLAUDE.md pointing to AGENTS.md
- [96456b8e](https://github.com/kubedb/kafka/commit/96456b8e) Add AGENTS.md for AI coding agents
- [8b49617d](https://github.com/kubedb/kafka/commit/8b49617d) Harden CI workflows (#192)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.41.0](https://github.com/kubedb/kibana/releases/tag/v0.41.0)

- [b042714b](https://github.com/kubedb/kibana/commit/b042714b) Prepare for release v0.41.0 (#186)
- [73bafa3b](https://github.com/kubedb/kibana/commit/73bafa3b) Prepare for release v0.41.0-rc.2 (#185)
- [d1ec131e](https://github.com/kubedb/kibana/commit/d1ec131e) Add network-policy-flavor flag for cilium support (#184)
- [7bec93e8](https://github.com/kubedb/kibana/commit/7bec93e8) Prepare for release v0.41.0-rc.1 (#183)
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



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.28.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.28.0)

- [bac18c12](https://github.com/kubedb/kubedb-manifest-plugin/commit/bac18c12) Prepare for release v0.28.0 (#136)
- [0ebaf789](https://github.com/kubedb/kubedb-manifest-plugin/commit/0ebaf789) Prepare for release v0.28.0-rc.2 (#135)
- [834ef012](https://github.com/kubedb/kubedb-manifest-plugin/commit/834ef012) Add restic backup progress streaming (#134)
- [5fe85111](https://github.com/kubedb/kubedb-manifest-plugin/commit/5fe85111) Prepare for release v0.28.0-rc.1 (#133)
- [1874f594](https://github.com/kubedb/kubedb-manifest-plugin/commit/1874f594) Prepare for release v0.28.0-rc.0 (#132)
- [a8c857c8](https://github.com/kubedb/kubedb-manifest-plugin/commit/a8c857c8) Harden release and release-tracker workflows
- [4390340d](https://github.com/kubedb/kubedb-manifest-plugin/commit/4390340d) Add AGENTS.md for AI coding agents
- [88a6d9b5](https://github.com/kubedb/kubedb-manifest-plugin/commit/88a6d9b5) Harden CI workflows (#130)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.16.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.16.0)

- [2af814fc](https://github.com/kubedb/kubedb-verifier/commit/2af814fc) Prepare for release v0.16.0 (#55)
- [783bc77f](https://github.com/kubedb/kubedb-verifier/commit/783bc77f) Update github.com/moby/spdystream to v0.5.1 (#54)
- [fe877028](https://github.com/kubedb/kubedb-verifier/commit/fe877028) Prepare for release v0.16.0-rc.2 (#53)
- [bf5f23ea](https://github.com/kubedb/kubedb-verifier/commit/bf5f23ea) Prepare for release v0.16.0-rc.1 (#52)
- [e78f35ea](https://github.com/kubedb/kubedb-verifier/commit/e78f35ea) Prepare for release v0.16.0-rc.0 (#51)
- [fa4b05f8](https://github.com/kubedb/kubedb-verifier/commit/fa4b05f8) Harden release and release-tracker workflows
- [f13312e1](https://github.com/kubedb/kubedb-verifier/commit/f13312e1) Add AGENTS.md for AI coding agents
- [8d1ab507](https://github.com/kubedb/kubedb-verifier/commit/8d1ab507) Harden CI workflows (#49)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.49.0](https://github.com/kubedb/mariadb/releases/tag/v0.49.0)

- [0a91d48a](https://github.com/kubedb/mariadb/commit/0a91d48a9) Prepare for release v0.49.0 (#411)
- [a07d6b7f](https://github.com/kubedb/mariadb/commit/a07d6b7f2) Update github.com/moby/spdystream to v0.5.1 (#410)
- [dbc68587](https://github.com/kubedb/mariadb/commit/dbc685874) Inc Snapshot Update for Distributed (#408)
- [dece23fb](https://github.com/kubedb/mariadb/commit/dece23fb3) Prepare for release v0.49.0-rc.2 (#409)
- [5249f7ac](https://github.com/kubedb/mariadb/commit/5249f7ac9) Honor user-provided renewBefore in TLS certificate ops (#404)
- [b1a16bc6](https://github.com/kubedb/mariadb/commit/b1a16bc61) Use endpoint from ResticStats.Summary (#407)
- [f0f7b7e0](https://github.com/kubedb/mariadb/commit/f0f7b7e0f) Prepare for release v0.49.0-rc.1 (#406)
- [1bb8a02a](https://github.com/kubedb/mariadb/commit/1bb8a02ac) Add StorageMigration OpsRequest support (#398)
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



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.25.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.25.0)

- [c365ef8c](https://github.com/kubedb/mariadb-archiver/commit/c365ef8c) Prepare for release v0.25.0 (#97)
- [104a6420](https://github.com/kubedb/mariadb-archiver/commit/104a6420) Pass ci (#96)
- [8cbbe471](https://github.com/kubedb/mariadb-archiver/commit/8cbbe471) Prepare for release v0.25.0-rc.2 (#95)
- [06784371](https://github.com/kubedb/mariadb-archiver/commit/06784371) Prepare for release v0.25.0-rc.1 (#93)
- [0c208981](https://github.com/kubedb/mariadb-archiver/commit/0c208981) Prepare for release v0.25.0-rc.0 (#92)
- [e0dc262d](https://github.com/kubedb/mariadb-archiver/commit/e0dc262d) Tighten CI/release workflow secrets, perms, and release notes
- [b10e53a2](https://github.com/kubedb/mariadb-archiver/commit/b10e53a2) Harden release and release-tracker workflows
- [2a326a8b](https://github.com/kubedb/mariadb-archiver/commit/2a326a8b) Add AGENTS.md for AI coding agents
- [0eae12a8](https://github.com/kubedb/mariadb-archiver/commit/0eae12a8) Harden CI workflows (#90)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.45.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.45.0)

- [16a7f616](https://github.com/kubedb/mariadb-coordinator/commit/16a7f616) Prepare for release v0.45.0 (#183)
- [87044a42](https://github.com/kubedb/mariadb-coordinator/commit/87044a42) Update github.com/moby/spdystream to v0.5.1 (#182)
- [11fc976d](https://github.com/kubedb/mariadb-coordinator/commit/11fc976d) Prepare for release v0.45.0-rc.2 (#181)
- [1358936a](https://github.com/kubedb/mariadb-coordinator/commit/1358936a) Prepare for release v0.45.0-rc.1 (#180)
- [da2d60f4](https://github.com/kubedb/mariadb-coordinator/commit/da2d60f4) Prepare for release v0.45.0-rc.0 (#179)
- [77d610ef](https://github.com/kubedb/mariadb-coordinator/commit/77d610ef) Tighten CI/release workflow secrets, perms, and release notes
- [492ee0ac](https://github.com/kubedb/mariadb-coordinator/commit/492ee0ac) Chaos Test: Fix Disaster Recovery (#174)
- [7b81cb8c](https://github.com/kubedb/mariadb-coordinator/commit/7b81cb8c) Harden release and release-tracker workflows
- [43a65240](https://github.com/kubedb/mariadb-coordinator/commit/43a65240) Add AGENTS.md for AI coding agents
- [0af4eea5](https://github.com/kubedb/mariadb-coordinator/commit/0af4eea5) Harden CI workflows (#177)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.25.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.25.0)

- [de9f69df](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/de9f69df) Prepare for release v0.25.0 (#82)
- [0c58c412](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/0c58c412) Update github.com/moby/spdystream to v0.5.1 (#81)
- [730ee57c](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/730ee57c) Prepare for release v0.25.0-rc.2 (#80)
- [70b5cdd2](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/70b5cdd2) Prepare for release v0.25.0-rc.1 (#79)
- [f432211e](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/f432211e) Prepare for release v0.25.0-rc.0 (#78)
- [b75d52e3](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/b75d52e3) Tighten CI/release workflow secrets, perms, and release notes
- [d622d7f3](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/d622d7f3) Harden release and release-tracker workflows
- [748e02a6](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/748e02a6) Add AGENTS.md for AI coding agents
- [cccdc734](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/cccdc734) Harden CI workflows (#76)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.23.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.23.0)

- [ad641e56](https://github.com/kubedb/mariadb-restic-plugin/commit/ad641e56) Prepare for release v0.23.0 (#96)
- [0226aa69](https://github.com/kubedb/mariadb-restic-plugin/commit/0226aa69) Update github.com/moby/spdystream to v0.5.1 (#95)
- [03ee3ee4](https://github.com/kubedb/mariadb-restic-plugin/commit/03ee3ee4) Prepare for release v0.23.0-rc.2 (#94)
- [ab0be84b](https://github.com/kubedb/mariadb-restic-plugin/commit/ab0be84b) Add restic backup progress streaming (#92)
- [d3bc1259](https://github.com/kubedb/mariadb-restic-plugin/commit/d3bc1259) Update Backup Job Name for Distributed (#93)
- [3b08f7c6](https://github.com/kubedb/mariadb-restic-plugin/commit/3b08f7c6) Prepare for release v0.23.0-rc.1 (#91)
- [7ffb72a7](https://github.com/kubedb/mariadb-restic-plugin/commit/7ffb72a7) Prepare for release v0.23.0-rc.0 (#90)
- [9e4c9505](https://github.com/kubedb/mariadb-restic-plugin/commit/9e4c9505) Harden release and release-tracker workflows
- [3e6d811d](https://github.com/kubedb/mariadb-restic-plugin/commit/3e6d811d) Add AGENTS.md for AI coding agents
- [df88fa1b](https://github.com/kubedb/mariadb-restic-plugin/commit/df88fa1b) Harden CI workflows (#88)



## [kubedb/migrator-cli](https://github.com/kubedb/migrator-cli)

### [v0.5.0](https://github.com/kubedb/migrator-cli/releases/tag/v0.5.0)

- [eb2fe1d4](https://github.com/kubedb/migrator-cli/commit/eb2fe1d4) Prepare for release v0.5.0 (#29)
- [77456208](https://github.com/kubedb/migrator-cli/commit/77456208) Merge pull request #28 from kubedb/mongoshakePathChange
- [c8134f14](https://github.com/kubedb/migrator-cli/commit/c8134f14) Merge branch 'master' into mongoshakePathChange
- [09a10537](https://github.com/kubedb/migrator-cli/commit/09a10537) changed mongoshake branch
- [1c853d60](https://github.com/kubedb/migrator-cli/commit/1c853d60) Add tls for postgres (#26)
- [5e531a13](https://github.com/kubedb/migrator-cli/commit/5e531a13) changed mongoshake source from alibaba to kubedb
- [3f035485](https://github.com/kubedb/migrator-cli/commit/3f035485) Update github.com/jackc/pgx/v5 to v5.9.2 (#27)
- [fe2606c9](https://github.com/kubedb/migrator-cli/commit/fe2606c9) Prepare for release v0.5.0-rc.2 (#25)
- [b7088fad](https://github.com/kubedb/migrator-cli/commit/b7088fad) Mysql migration init (#11)
- [06810c85](https://github.com/kubedb/migrator-cli/commit/06810c85) Update README.md
- [f64c95cd](https://github.com/kubedb/migrator-cli/commit/f64c95cd) Prepare for release v0.5.0-rc.1 (#23)
- [cfecdf1a](https://github.com/kubedb/migrator-cli/commit/cfecdf1a) Prepare for release v0.5.0-rc.0 (#21)
- [1615cc31](https://github.com/kubedb/migrator-cli/commit/1615cc31) Tighten CI/release workflow secrets, perms, and release notes
- [bbba7844](https://github.com/kubedb/migrator-cli/commit/bbba7844) Added MongoDB migration (#16)
- [27f1c91c](https://github.com/kubedb/migrator-cli/commit/27f1c91c) Harden release and release-tracker workflows
- [d21c2740](https://github.com/kubedb/migrator-cli/commit/d21c2740) Add AGENTS.md for AI coding agents
- [7d5ab3d4](https://github.com/kubedb/migrator-cli/commit/7d5ab3d4) Harden CI workflows (#19)
- [8a21ae49](https://github.com/kubedb/migrator-cli/commit/8a21ae49) Separate dockerfile for each databases (#17)



## [kubedb/migrator-operator](https://github.com/kubedb/migrator-operator)

### [v0.5.0](https://github.com/kubedb/migrator-operator/releases/tag/v0.5.0)




## [kubedb/milvus](https://github.com/kubedb/milvus)

### [v0.6.0](https://github.com/kubedb/milvus/releases/tag/v0.6.0)

- [0fdf9a63](https://github.com/kubedb/milvus/commit/0fdf9a63) Prepare for release v0.6.0 (#49)
- [611a5818](https://github.com/kubedb/milvus/commit/611a5818) Fix mTLS issue (#47)
- [fa57da9b](https://github.com/kubedb/milvus/commit/fa57da9b) Prepare for release v0.6.0-rc.2 (#48)
- [85d63146](https://github.com/kubedb/milvus/commit/85d63146) Add HorizontalScaling OpsRequest support (#45)
- [41e84651](https://github.com/kubedb/milvus/commit/41e84651) Honor user-provided renewBefore in TLS certificate ops (#40)
- [d364f887](https://github.com/kubedb/milvus/commit/d364f887) Fix Milvus deletion (#44)
- [1ae0debe](https://github.com/kubedb/milvus/commit/1ae0debe) Add RotateAuth OpsRequest support for Milvus (#46)
- [79cba14c](https://github.com/kubedb/milvus/commit/79cba14c) Add NetworkPolicyFlavor support for cilium (#43)
- [703de898](https://github.com/kubedb/milvus/commit/703de898) Prepare for release v0.6.0-rc.1 (#42)
- [83890286](https://github.com/kubedb/milvus/commit/83890286) Add StorageMigration Ops-Request Support (#39)
- [718988e1](https://github.com/kubedb/milvus/commit/718988e1) Add Milvus Ops Request (#33)
- [062b7eab](https://github.com/kubedb/milvus/commit/062b7eab) Prepare for release v0.6.0-rc.0 (#37)
- [984bf9be](https://github.com/kubedb/milvus/commit/984bf9be) Tighten CI/release workflow secrets, perms, and release notes
- [63fee186](https://github.com/kubedb/milvus/commit/63fee186) Harden release and release-tracker workflows
- [ceb74c51](https://github.com/kubedb/milvus/commit/ceb74c51) Add CLAUDE.md pointing to AGENTS.md
- [b69a946b](https://github.com/kubedb/milvus/commit/b69a946b) Add Milvus Tls (#25)
- [f319f086](https://github.com/kubedb/milvus/commit/f319f086) Add AGENTS.md for AI coding agents
- [10e389a6](https://github.com/kubedb/milvus/commit/10e389a6) Harden CI workflows (#34)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.58.0](https://github.com/kubedb/mongodb/releases/tag/v0.58.0)

- [becd088e](https://github.com/kubedb/mongodb/commit/becd088e0) Prepare for release v0.58.0 (#769)
- [29e1f9b3](https://github.com/kubedb/mongodb/commit/29e1f9b32) Update github.com/moby/spdystream to v0.5.1 (#768)
- [b2085f89](https://github.com/kubedb/mongodb/commit/b2085f897) Prepare for release v0.58.0-rc.2 (#767)
- [4591f101](https://github.com/kubedb/mongodb/commit/4591f1011) Prepare for release v0.58.0-rc.1 (#765)
- [0ec52968](https://github.com/kubedb/mongodb/commit/0ec529684) Add virtual secret support (#762)
- [2946c711](https://github.com/kubedb/mongodb/commit/2946c7119) Honor user-provided renewBefore in TLS certificate ops (#764)
- [75fb23b2](https://github.com/kubedb/mongodb/commit/75fb23b2d) Implement cilium networkpolicy (#763)
- [0f32df04](https://github.com/kubedb/mongodb/commit/0f32df042) Prepare for release v0.58.0-rc.0 (#761)
- [d1f66646](https://github.com/kubedb/mongodb/commit/d1f66646a) Tighten CI/release workflow secrets, perms, and release notes
- [71ddfc66](https://github.com/kubedb/mongodb/commit/71ddfc668) Harden release and release-tracker workflows
- [b6c2d289](https://github.com/kubedb/mongodb/commit/b6c2d2898) Run Ops Request Locally (#760)
- [93f12e1a](https://github.com/kubedb/mongodb/commit/93f12e1a9) Add StorageMigration OpsRequest support (#759)
- [c895f662](https://github.com/kubedb/mongodb/commit/c895f662f) Add CLAUDE.md pointing to AGENTS.md
- [a80e04c8](https://github.com/kubedb/mongodb/commit/a80e04c8a) Add AGENTS.md for AI coding agents
- [30e4b394](https://github.com/kubedb/mongodb/commit/30e4b394c) Harden CI workflows (#757)
- [56b32245](https://github.com/kubedb/mongodb/commit/56b322456) Fix Sidekick issue; Fix storage cred secret sync issue (#756)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.26.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.26.0)

- [68a2ebeb](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/68a2ebeb) Prepare for release v0.26.0 (#86)
- [fe132cd7](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/fe132cd7) Prepare for release v0.26.0-rc.2 (#85)
- [ec3b81a8](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/ec3b81a8) Prepare for release v0.26.0-rc.1 (#84)
- [154dedfb](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/154dedfb) Prepare for release v0.26.0-rc.0 (#83)
- [86a4f09c](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/86a4f09c) Tighten CI/release workflow secrets, perms, and release notes
- [c445729e](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/c445729e) Harden release and release-tracker workflows
- [6f77d33a](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/6f77d33a) Add AGENTS.md for AI coding agents
- [fd4c25de](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/fd4c25de) Harden CI workflows (#81)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.28.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.28.0)

- [90025a3f](https://github.com/kubedb/mongodb-restic-plugin/commit/90025a3f) Prepare for release v0.28.0 (#132)
- [d6b5fa11](https://github.com/kubedb/mongodb-restic-plugin/commit/d6b5fa11) Prepare for release v0.28.0-rc.2 (#131)
- [58aef2eb](https://github.com/kubedb/mongodb-restic-plugin/commit/58aef2eb) Add Backup Progress Streaming Support in Snapshot Status (#130)
- [a8f1b59b](https://github.com/kubedb/mongodb-restic-plugin/commit/a8f1b59b) Prepare for release v0.28.0-rc.1 (#129)
- [70c2a000](https://github.com/kubedb/mongodb-restic-plugin/commit/70c2a000) Prepare for release v0.28.0-rc.0 (#128)
- [88c32465](https://github.com/kubedb/mongodb-restic-plugin/commit/88c32465) Harden release and release-tracker workflows
- [4c6631af](https://github.com/kubedb/mongodb-restic-plugin/commit/4c6631af) Add uri connection string for mongodump and mongorestore (#124)
- [fcb9af9b](https://github.com/kubedb/mongodb-restic-plugin/commit/fcb9af9b) Add AGENTS.md for AI coding agents
- [76f8adb4](https://github.com/kubedb/mongodb-restic-plugin/commit/76f8adb4) Harden CI workflows (#125)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.20.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.20.0)

- [601bb948](https://github.com/kubedb/mssql-coordinator/commit/601bb948) Prepare for release v0.20.0 (#75)
- [892acf04](https://github.com/kubedb/mssql-coordinator/commit/892acf04) Update github.com/moby/spdystream to v0.5.1 (#74)
- [2716be6e](https://github.com/kubedb/mssql-coordinator/commit/2716be6e) Prepare for release v0.20.0-rc.2 (#73)
- [acc28e43](https://github.com/kubedb/mssql-coordinator/commit/acc28e43) Prepare for release v0.20.0-rc.1 (#72)
- [a7c9d240](https://github.com/kubedb/mssql-coordinator/commit/a7c9d240) Fix CI hardening: use app token in release-tracker, add packages:write
- [6293f384](https://github.com/kubedb/mssql-coordinator/commit/6293f384) Prepare for release v0.20.0-rc.0 (#70)
- [cb31d9b3](https://github.com/kubedb/mssql-coordinator/commit/cb31d9b3) Tighten CI/release workflow secrets, perms, and release notes
- [26c8ec07](https://github.com/kubedb/mssql-coordinator/commit/26c8ec07) Add AGENTS.md for AI coding agents
- [29b5b49e](https://github.com/kubedb/mssql-coordinator/commit/29b5b49e) Use GitHub App token for release tracker comments (#68)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.20.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.20.0)

- [8fd94742](https://github.com/kubedb/mssqlserver/commit/8fd94742) Prepare for release v0.20.0 (#143)
- [a47bf888](https://github.com/kubedb/mssqlserver/commit/a47bf888) Update github.com/moby/spdystream to v0.5.1 (#142)
- [418c160c](https://github.com/kubedb/mssqlserver/commit/418c160c) Prepare for release v0.20.0-rc.2 (#141)
- [2fadba5a](https://github.com/kubedb/mssqlserver/commit/2fadba5a) Add NetworkPolicyFlavor support for cilium (#139)
- [1ec9a0f1](https://github.com/kubedb/mssqlserver/commit/1ec9a0f1) Prepare for release v0.20.0-rc.1 (#138)
- [0eb8b1c8](https://github.com/kubedb/mssqlserver/commit/0eb8b1c8) Add StorageMigration OpsRequest support for MSSQLServer (#132)
- [787c5056](https://github.com/kubedb/mssqlserver/commit/787c5056) Prepare for release v0.20.0-rc.0 (#133)
- [76e95035](https://github.com/kubedb/mssqlserver/commit/76e95035) Tighten CI/release workflow secrets, perms, and release notes
- [8afbd426](https://github.com/kubedb/mssqlserver/commit/8afbd426) Harden release and release-tracker workflows
- [c7de0e00](https://github.com/kubedb/mssqlserver/commit/c7de0e00) Add CLAUDE.md pointing to AGENTS.md
- [ccef8128](https://github.com/kubedb/mssqlserver/commit/ccef8128) Add AGENTS.md for AI coding agents
- [a8ffa607](https://github.com/kubedb/mssqlserver/commit/a8ffa607) Harden CI workflows (#130)
- [a4e311d8](https://github.com/kubedb/mssqlserver/commit/a4e311d8) Update sidekick leader selection labels; storage sync is not implemented (#129)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.19.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.19.0)

- [327ebbf](https://github.com/kubedb/mssqlserver-archiver/commit/327ebbf) Prepare for release v0.19.0 (#31)
- [df59193](https://github.com/kubedb/mssqlserver-archiver/commit/df59193) Prepare for release v0.19.0-rc.2 (#30)
- [d3713a4](https://github.com/kubedb/mssqlserver-archiver/commit/d3713a4) Import kubedb.dev/apimachinery to track dependency (#29)
- [24678f5](https://github.com/kubedb/mssqlserver-archiver/commit/24678f5) Tighten CI/release workflow secrets, perms, and release notes
- [958106a](https://github.com/kubedb/mssqlserver-archiver/commit/958106a) Harden release and release-tracker workflows
- [ea87799](https://github.com/kubedb/mssqlserver-archiver/commit/ea87799) Add AGENTS.md for AI coding agents
- [71f8c47](https://github.com/kubedb/mssqlserver-archiver/commit/71f8c47) Harden CI workflows (#26)
- [942ca43](https://github.com/kubedb/mssqlserver-archiver/commit/942ca43) Harden CI workflows (#25)



## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.19.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.19.0)

- [9c7af14](https://github.com/kubedb/mssqlserver-walg-plugin/commit/9c7af14) Prepare for release v0.19.0 (#63)
- [1a44640](https://github.com/kubedb/mssqlserver-walg-plugin/commit/1a44640) Prepare for release v0.19.0-rc.2 (#62)
- [51ea4f2](https://github.com/kubedb/mssqlserver-walg-plugin/commit/51ea4f2) Prepare for release v0.19.0-rc.1 (#61)
- [1475bb0](https://github.com/kubedb/mssqlserver-walg-plugin/commit/1475bb0) Fix CI hardening: use app token in release-tracker, add packages:write
- [5cc5025](https://github.com/kubedb/mssqlserver-walg-plugin/commit/5cc5025) Prepare for release v0.19.0-rc.0 (#59)
- [5433238](https://github.com/kubedb/mssqlserver-walg-plugin/commit/5433238) Add AGENTS.md for AI coding agents
- [0e152a6](https://github.com/kubedb/mssqlserver-walg-plugin/commit/0e152a6) Use GitHub App token for release tracker comments (#57)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.26.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.26.0)

- [093b05d2](https://github.com/kubedb/mysql-archiver/commit/093b05d2) Prepare for release v0.26.0 (#110)
- [530f2468](https://github.com/kubedb/mysql-archiver/commit/530f2468) Prepare for release v0.26.0-rc.2 (#109)
- [33eaf828](https://github.com/kubedb/mysql-archiver/commit/33eaf828) Prepare for release v0.26.0-rc.1 (#108)
- [6a708251](https://github.com/kubedb/mysql-archiver/commit/6a708251) add permission for the release job (#107)
- [eb55d4ad](https://github.com/kubedb/mysql-archiver/commit/eb55d4ad) Prepare for release v0.26.0-rc.0 (#106)
- [d0ba1b24](https://github.com/kubedb/mysql-archiver/commit/d0ba1b24) Tighten CI/release workflow secrets, perms, and release notes
- [ae11ecbd](https://github.com/kubedb/mysql-archiver/commit/ae11ecbd) Harden release and release-tracker workflows
- [53fbffa9](https://github.com/kubedb/mysql-archiver/commit/53fbffa9) Add AGENTS.md for AI coding agents
- [6448ddf1](https://github.com/kubedb/mysql-archiver/commit/6448ddf1) Harden CI workflows (#104)
- [1098c78a](https://github.com/kubedb/mysql-archiver/commit/1098c78a) Fix binlog reply (#103)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.43.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.43.0)

- [4cb2b084](https://github.com/kubedb/mysql-coordinator/commit/4cb2b084) Prepare for release v0.43.0 (#185)
- [eea38e29](https://github.com/kubedb/mysql-coordinator/commit/eea38e29) Update github.com/moby/spdystream to v0.5.1 (#184)
- [67143c6d](https://github.com/kubedb/mysql-coordinator/commit/67143c6d) Prepare for release v0.43.0-rc.2 (#183)
- [c813bf0e](https://github.com/kubedb/mysql-coordinator/commit/c813bf0e) Prepare for release v0.43.0-rc.1 (#181)
- [700dcc95](https://github.com/kubedb/mysql-coordinator/commit/700dcc95) Prepare for release v0.43.0-rc.0 (#180)
- [7550f622](https://github.com/kubedb/mysql-coordinator/commit/7550f622) Tighten CI/release workflow secrets, perms, and release notes
- [8772c753](https://github.com/kubedb/mysql-coordinator/commit/8772c753) Harden release and release-tracker workflows
- [1a20ce7c](https://github.com/kubedb/mysql-coordinator/commit/1a20ce7c) Fix innodb cluster support for 8.4+ (#171)
- [21bcc293](https://github.com/kubedb/mysql-coordinator/commit/21bcc293) Add AGENTS.md for AI coding agents
- [92a93955](https://github.com/kubedb/mysql-coordinator/commit/92a93955) Harden CI workflows (#178)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.26.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.26.0)

- [aa454ed7](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/aa454ed7) Prepare for release v0.26.0 (#82)
- [3fd6f8d9](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/3fd6f8d9) Prepare for release v0.26.0-rc.2 (#81)
- [5c73df2a](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/5c73df2a) Prepare for release v0.26.0-rc.1 (#80)
- [02b24bac](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/02b24bac) Prepare for release v0.26.0-rc.0 (#79)
- [b98f8382](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/b98f8382) Tighten CI/release workflow secrets, perms, and release notes
- [aeb71151](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/aeb71151) Harden release and release-tracker workflows
- [960c3538](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/960c3538) Add AGENTS.md for AI coding agents
- [06b90e07](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/06b90e07) Harden CI workflows (#77)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.28.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.28.0)

- [58d3cd4c](https://github.com/kubedb/mysql-restic-plugin/commit/58d3cd4c) Prepare for release v0.28.0 (#117)
- [85ba443a](https://github.com/kubedb/mysql-restic-plugin/commit/85ba443a) Prepare for release v0.28.0-rc.2 (#116)
- [edb738be](https://github.com/kubedb/mysql-restic-plugin/commit/edb738be) Add Backup Progress Streaming Support in Snapshot Status (#113)
- [8751f570](https://github.com/kubedb/mysql-restic-plugin/commit/8751f570) Prepare for release v0.28.0-rc.1 (#115)
- [dd1b9eae](https://github.com/kubedb/mysql-restic-plugin/commit/dd1b9eae) Prepare for release v0.28.0-rc.0 (#114)
- [879cbd56](https://github.com/kubedb/mysql-restic-plugin/commit/879cbd56) Harden release and release-tracker workflows
- [3a29e9c8](https://github.com/kubedb/mysql-restic-plugin/commit/3a29e9c8) Add AGENTS.md for AI coding agents
- [a5955a10](https://github.com/kubedb/mysql-restic-plugin/commit/a5955a10) Harden CI workflows (#111)
- [3a49671f](https://github.com/kubedb/mysql-restic-plugin/commit/3a49671f) Add New Version Support, Innodb Cluster Support (#110)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.43.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.43.0)

- [d65df0c](https://github.com/kubedb/mysql-router-init/commit/d65df0c) Prepare for release v0.43.0 (#65)
- [b0ab3a5](https://github.com/kubedb/mysql-router-init/commit/b0ab3a5) Prepare for release v0.43.0-rc.2 (#64)
- [9169dc3](https://github.com/kubedb/mysql-router-init/commit/9169dc3) Import kubedb.dev/apimachinery to track dependency (#63)
- [e15fe6a](https://github.com/kubedb/mysql-router-init/commit/e15fe6a) Fix CI hardening: use app token in release-tracker, add packages (#61)
- [1fe79e5](https://github.com/kubedb/mysql-router-init/commit/1fe79e5) Tighten CI/release workflow secrets, perms, and release notes
- [399bf84](https://github.com/kubedb/mysql-router-init/commit/399bf84) Add AGENTS.md for AI coding agents
- [e1d55d1](https://github.com/kubedb/mysql-router-init/commit/e1d55d1) Merge pull request #59 from kubedb/use-app-token-2284



## [kubedb/neo4j](https://github.com/kubedb/neo4j)

### [v0.6.0](https://github.com/kubedb/neo4j/releases/tag/v0.6.0)

- [534f93d1](https://github.com/kubedb/neo4j/commit/534f93d1) Prepare for release v0.6.0 (#44)
- [5956ff9d](https://github.com/kubedb/neo4j/commit/5956ff9d) CustomConfig VolumeMount fix (#43)
- [2639a89b](https://github.com/kubedb/neo4j/commit/2639a89b) Prepare for release v0.6.0-rc.2 (#42)
- [f742558f](https://github.com/kubedb/neo4j/commit/f742558f) Fix passing credential as a literal env value in the PetSet (#41)
- [c465d3c8](https://github.com/kubedb/neo4j/commit/c465d3c8) feat: implement git-sync init container for Neo4j (#36)
- [4f787bca](https://github.com/kubedb/neo4j/commit/4f787bca) Fix Neo4j Deletion (#40)
- [e6ab6aa3](https://github.com/kubedb/neo4j/commit/e6ab6aa3) Add NetworkPolicyFlavor support (#39)
- [c5dcc324](https://github.com/kubedb/neo4j/commit/c5dcc324) Prepare for release v0.6.0-rc.1 (#38)
- [0bf9da80](https://github.com/kubedb/neo4j/commit/0bf9da80) Add Backup port in primary service (#31)
- [63a96dad](https://github.com/kubedb/neo4j/commit/63a96dad) Prepare for release v0.6.0-rc.0 (#34)
- [e215c4c8](https://github.com/kubedb/neo4j/commit/e215c4c8) Tighten CI/release workflow secrets, perms, and release notes
- [98283581](https://github.com/kubedb/neo4j/commit/98283581) Harden release and release-tracker workflows
- [77704502](https://github.com/kubedb/neo4j/commit/77704502) Add StorageMigration OpsRequest support for Neo4j (#33)
- [8dc5edbf](https://github.com/kubedb/neo4j/commit/8dc5edbf) Add CLAUDE.md pointing to AGENTS.md
- [c8f95c51](https://github.com/kubedb/neo4j/commit/c8f95c51) Add AGENTS.md for AI coding agents
- [beb02ed6](https://github.com/kubedb/neo4j/commit/beb02ed6) Harden CI workflows (#30)



## [kubedb/neo4j-backup-plugin](https://github.com/kubedb/neo4j-backup-plugin)

### [v0.1.0](https://github.com/kubedb/neo4j-backup-plugin/releases/tag/v0.1.0)

- [6059399](https://github.com/kubedb/neo4j-backup-plugin/commit/6059399) Use gh cli instead of hub cli
- [3c30045](https://github.com/kubedb/neo4j-backup-plugin/commit/3c30045) Prepare for release v0.1.0 (#2)
- [3342b9f](https://github.com/kubedb/neo4j-backup-plugin/commit/3342b9f) Add Neo4j Backup Plugin (#1)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.52.0](https://github.com/kubedb/ops-manager/releases/tag/v0.52.0)




## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.11.0](https://github.com/kubedb/oracle/releases/tag/v0.11.0)

- [ec2b76b3](https://github.com/kubedb/oracle/commit/ec2b76b3) Prepare for release v0.11.0 (#62)
- [4f6c5ef7](https://github.com/kubedb/oracle/commit/4f6c5ef7) Update github.com/moby/spdystream to v0.5.1 (#61)
- [91f689e2](https://github.com/kubedb/oracle/commit/91f689e2) Prepare for release v0.11.0-rc.2 (#60)
- [666716b1](https://github.com/kubedb/oracle/commit/666716b1) Implement VolumeExpansion ops for Oracle (#46)
- [6c4afce5](https://github.com/kubedb/oracle/commit/6c4afce5) Add VerticalScaling OpsRequest implementation (#47)
- [5bf3fe65](https://github.com/kubedb/oracle/commit/5bf3fe65) Implement RotateAuthentication for Oracle (#48)
- [4ffeaede](https://github.com/kubedb/oracle/commit/4ffeaede) add oracle ops restart reconfigure and appbinding (#42)
- [44e872ea](https://github.com/kubedb/oracle/commit/44e872ea) add appbinding for oracle (#51)
- [1cc52267](https://github.com/kubedb/oracle/commit/1cc52267) Add NetworkPolicyFlavor support (#59)
- [df02ec98](https://github.com/kubedb/oracle/commit/df02ec98) Honor user-provided renewBefore in TLS certificate ops (#56)
- [789d3804](https://github.com/kubedb/oracle/commit/789d3804) Prepare for release v0.11.0-rc.1 (#58)
- [13c00d3e](https://github.com/kubedb/oracle/commit/13c00d3e) add ImagePullSecret for observer (#54)
- [22259d8b](https://github.com/kubedb/oracle/commit/22259d8b) Prepare for release v0.11.0-rc.0 (#52)
- [ef7a90cb](https://github.com/kubedb/oracle/commit/ef7a90cb) Tighten CI/release workflow secrets, perms, and release notes
- [7484f753](https://github.com/kubedb/oracle/commit/7484f753) fix-petset-GetObjectMeta (#49)
- [a2a4e973](https://github.com/kubedb/oracle/commit/a2a4e973) Harden release and release-tracker workflows
- [449beab1](https://github.com/kubedb/oracle/commit/449beab1) Add CLAUDE.md pointing to AGENTS.md
- [ae706c75](https://github.com/kubedb/oracle/commit/ae706c75) Add AGENTS.md for AI coding agents
- [47addbe9](https://github.com/kubedb/oracle/commit/47addbe9) Harden CI workflows (#43)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.11.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.11.0)

- [81f5255](https://github.com/kubedb/oracle-coordinator/commit/81f5255) Prepare for release v0.11.0 (#40)
- [f16a847](https://github.com/kubedb/oracle-coordinator/commit/f16a847) Update github.com/moby/spdystream to v0.5.1 (#39)
- [5c50fdd](https://github.com/kubedb/oracle-coordinator/commit/5c50fdd) Prepare for release v0.11.0-rc.2 (#38)
- [1a66329](https://github.com/kubedb/oracle-coordinator/commit/1a66329) Prepare for release v0.11.0-rc.1 (#37)
- [cde0783](https://github.com/kubedb/oracle-coordinator/commit/cde0783) Fix CI hardening: add kodiak.toml, use app token in release-tracker
- [eebe9e3](https://github.com/kubedb/oracle-coordinator/commit/eebe9e3) Prepare for release v0.11.0-rc.0 (#35)
- [40f851c](https://github.com/kubedb/oracle-coordinator/commit/40f851c) Tighten CI/release workflow secrets, perms, and release notes
- [5d987d9](https://github.com/kubedb/oracle-coordinator/commit/5d987d9) Add AGENTS.md for AI coding agents
- [9c163c8](https://github.com/kubedb/oracle-coordinator/commit/9c163c8) Use GitHub App token for release tracker comments (#33)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.52.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.52.0)

- [973e597d](https://github.com/kubedb/percona-xtradb/commit/973e597d4) Prepare for release v0.52.0 (#461)
- [446a97f4](https://github.com/kubedb/percona-xtradb/commit/446a97f45) Prepare for release v0.52.0-rc.2 (#460)
- [0b5518e3](https://github.com/kubedb/percona-xtradb/commit/0b5518e33) feat: implement git-sync init container for PerconaXtraDB (#456)
- [10329f89](https://github.com/kubedb/percona-xtradb/commit/10329f895) Honor user-provided renewBefore in TLS certificate ops (#457)
- [db4f1d80](https://github.com/kubedb/percona-xtradb/commit/db4f1d809) Prepare for release v0.52.0-rc.1 (#459)
- [023944e1](https://github.com/kubedb/percona-xtradb/commit/023944e11) Add StorageMigration OpsRequest support for PerconaXtraDB (#452)
- [a03b66b9](https://github.com/kubedb/percona-xtradb/commit/a03b66b9c) Prepare for release v0.52.0-rc.0 (#454)
- [2f3fbc70](https://github.com/kubedb/percona-xtradb/commit/2f3fbc700) Tighten CI/release workflow secrets, perms, and release notes
- [cd18b8f0](https://github.com/kubedb/percona-xtradb/commit/cd18b8f00) Harden release and release-tracker workflows
- [06ee2cbf](https://github.com/kubedb/percona-xtradb/commit/06ee2cbfb) Run Ops Request Locally (#453)
- [887fa9a0](https://github.com/kubedb/percona-xtradb/commit/887fa9a01) Add CLAUDE.md pointing to AGENTS.md
- [bc309e6c](https://github.com/kubedb/percona-xtradb/commit/bc309e6ce) Add AGENTS.md for AI coding agents
- [52dc6b56](https://github.com/kubedb/percona-xtradb/commit/52dc6b56c) Harden CI workflows (#450)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.38.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.38.0)

- [c14587a3](https://github.com/kubedb/percona-xtradb-coordinator/commit/c14587a3) Prepare for release v0.38.0 (#131)
- [10be320f](https://github.com/kubedb/percona-xtradb-coordinator/commit/10be320f) Update github.com/moby/spdystream to v0.5.1 (#130)
- [b33e29b9](https://github.com/kubedb/percona-xtradb-coordinator/commit/b33e29b9) Prepare for release v0.38.0-rc.2 (#129)
- [a0b5004e](https://github.com/kubedb/percona-xtradb-coordinator/commit/a0b5004e) Prepare for release v0.38.0-rc.1 (#128)
- [d71a8b78](https://github.com/kubedb/percona-xtradb-coordinator/commit/d71a8b78) Prepare for release v0.38.0-rc.0 (#127)
- [dcf12fa2](https://github.com/kubedb/percona-xtradb-coordinator/commit/dcf12fa2) Tighten CI/release workflow secrets, perms, and release notes
- [98fc7cc6](https://github.com/kubedb/percona-xtradb-coordinator/commit/98fc7cc6) Harden release and release-tracker workflows
- [37eb5997](https://github.com/kubedb/percona-xtradb-coordinator/commit/37eb5997) Add AGENTS.md for AI coding agents
- [1881477c](https://github.com/kubedb/percona-xtradb-coordinator/commit/1881477c) Harden CI workflows (#125)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.49.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.49.0)

- [784ef83c](https://github.com/kubedb/pg-coordinator/commit/784ef83c) Prepare for release v0.49.0 (#256)
- [6ab0ea56](https://github.com/kubedb/pg-coordinator/commit/6ab0ea56) Update vulnerable dependencies (#255)
- [c2213131](https://github.com/kubedb/pg-coordinator/commit/c2213131) Prepare for release v0.49.0-rc.2 (#254)
- [46d694d9](https://github.com/kubedb/pg-coordinator/commit/46d694d9) Document pod injection role in AGENTS.md (#253)
- [412aedcf](https://github.com/kubedb/pg-coordinator/commit/412aedcf) Prepare for release v0.49.0-rc.1 (#252)
- [e1bf909c](https://github.com/kubedb/pg-coordinator/commit/e1bf909c) Prepare for release v0.49.0-rc.0 (#251)
- [9bda4b8a](https://github.com/kubedb/pg-coordinator/commit/9bda4b8a) Tighten CI/release workflow secrets, perms, and release notes
- [960efc2e](https://github.com/kubedb/pg-coordinator/commit/960efc2e) Harden release and release-tracker workflows



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.52.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.52.0)

- [11585f77](https://github.com/kubedb/pgbouncer/commit/11585f77d) Prepare for release v0.52.0 (#419)
- [aa44932f](https://github.com/kubedb/pgbouncer/commit/aa44932ff) Update github.com/moby/spdystream to v0.5.1 (#418)
- [cc7ae286](https://github.com/kubedb/pgbouncer/commit/cc7ae2860) Prepare for release v0.52.0-rc.2 (#417)
- [8e5905a4](https://github.com/kubedb/pgbouncer/commit/8e5905a4b) Prepare for release v0.52.0-rc.1 (#416)
- [c5e7a81e](https://github.com/kubedb/pgbouncer/commit/c5e7a81e0) Add --network-policy-flavor flag with cilium support (#415)
- [700a235c](https://github.com/kubedb/pgbouncer/commit/700a235cf) Prepare for release v0.52.0-rc.0 (#413)
- [84b1a19e](https://github.com/kubedb/pgbouncer/commit/84b1a19ea) Tighten CI/release workflow secrets, perms, and release notes
- [c3e4b4e1](https://github.com/kubedb/pgbouncer/commit/c3e4b4e10) Harden release and release-tracker workflows
- [e0746413](https://github.com/kubedb/pgbouncer/commit/e07464138) Run Ops Request Locally (#412)
- [ce2cad78](https://github.com/kubedb/pgbouncer/commit/ce2cad789) Add AGENTS.md for AI coding agents
- [1215a247](https://github.com/kubedb/pgbouncer/commit/1215a247e) Harden CI workflows (#410)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.20.0](https://github.com/kubedb/pgpool/releases/tag/v0.20.0)

- [295950cd](https://github.com/kubedb/pgpool/commit/295950cd) Prepare for release v0.20.0 (#127)
- [1300a83b](https://github.com/kubedb/pgpool/commit/1300a83b) Update github.com/moby/spdystream to v0.5.1 (#126)
- [a361fefb](https://github.com/kubedb/pgpool/commit/a361fefb) Prepare for release v0.20.0-rc.2 (#125)
- [1750b479](https://github.com/kubedb/pgpool/commit/1750b479) Prepare for release v0.20.0-rc.1 (#123)
- [eff93373](https://github.com/kubedb/pgpool/commit/eff93373) Add --network-policy-flavor flag with cilium support (#122)
- [9dd15b6e](https://github.com/kubedb/pgpool/commit/9dd15b6e) Prepare for release v0.20.0-rc.0 (#120)
- [dbd10776](https://github.com/kubedb/pgpool/commit/dbd10776) Tighten CI/release workflow secrets, perms, and release notes
- [12f83ac2](https://github.com/kubedb/pgpool/commit/12f83ac2) Harden release and release-tracker workflows
- [23a0bf8f](https://github.com/kubedb/pgpool/commit/23a0bf8f) Add AGENTS.md for AI coding agents
- [efab69a4](https://github.com/kubedb/pgpool/commit/efab69a4) Harden CI workflows (#118)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.65.0](https://github.com/kubedb/postgres/releases/tag/v0.65.0)

- [679f511d](https://github.com/kubedb/postgres/commit/679f511d8) Prepare for release v0.65.0 (#900)
- [46d91698](https://github.com/kubedb/postgres/commit/46d916986) Export PgQueue field in Reconciler struct (#897)
- [a5351479](https://github.com/kubedb/postgres/commit/a53514794) Update github.com/moby/spdystream to v0.5.1 (#899)
- [98471e23](https://github.com/kubedb/postgres/commit/98471e234) Prepare for release v0.65.0-rc.2 (#898)
- [ae9804e8](https://github.com/kubedb/postgres/commit/ae9804e8b) Use endpoint from ResticStats.Summary (#895)
- [da3dca05](https://github.com/kubedb/postgres/commit/da3dca059) Prepare for release v0.65.0-rc.1 (#894)
- [f203a907](https://github.com/kubedb/postgres/commit/f203a9079) Add --network-policy-flavor flag with cilium support (#886)
- [aa3c1fd5](https://github.com/kubedb/postgres/commit/aa3c1fd54) Honor user-provided renewBefore in TLS certificate ops (#893)
- [5d3d25d5](https://github.com/kubedb/postgres/commit/5d3d25d55) (Skip coordinator + Fix health check) for Remote Replica (#891)
- [4f7c4133](https://github.com/kubedb/postgres/commit/4f7c41333) Update cluster.local -> slice.local (#890)
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



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.26.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.26.0)

- [3c9fae1b](https://github.com/kubedb/postgres-archiver/commit/3c9fae1b) Prepare for release v0.26.0 (#112)
- [be08e519](https://github.com/kubedb/postgres-archiver/commit/be08e519) Update github.com/moby/spdystream to v0.5.1 (#111)
- [43553c97](https://github.com/kubedb/postgres-archiver/commit/43553c97) Prepare for release v0.26.0-rc.2 (#110)
- [10de04a8](https://github.com/kubedb/postgres-archiver/commit/10de04a8) Prepare for release v0.26.0-rc.1 (#109)
- [8246614a](https://github.com/kubedb/postgres-archiver/commit/8246614a) Add write permission (#108)
- [77f8cd11](https://github.com/kubedb/postgres-archiver/commit/77f8cd11) Prepare for release v0.26.0-rc.0 (#107)
- [07418083](https://github.com/kubedb/postgres-archiver/commit/07418083) Tighten CI/release workflow secrets, perms, and release notes
- [ba7d5601](https://github.com/kubedb/postgres-archiver/commit/ba7d5601) Harden release and release-tracker workflows
- [08fddab4](https://github.com/kubedb/postgres-archiver/commit/08fddab4) Add AGENTS.md for AI coding agents
- [1806aa73](https://github.com/kubedb/postgres-archiver/commit/1806aa73) Harden CI workflows (#105)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.26.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.26.0)

- [c6ae859c](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/c6ae859c) Prepare for release v0.26.0 (#92)
- [597ce968](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/597ce968) Prepare for release v0.26.0-rc.2 (#91)
- [a578dc28](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/a578dc28) Prepare for release v0.26.0-rc.1 (#90)
- [32139be4](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/32139be4) Prepare for release v0.26.0-rc.0 (#89)
- [e1be14bc](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/e1be14bc) Tighten CI/release workflow secrets, perms, and release notes
- [7590aa73](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/7590aa73) Harden release and release-tracker workflows
- [56eabbb2](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/56eabbb2) Add AGENTS.md for AI coding agents
- [7ce77805](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/7ce77805) Use GitHub App token for release tracker comments (#87)
- [5cdf2c50](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/5cdf2c50) Harden CI workflows (#86)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.28.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.28.0)

- [b97a07f6](https://github.com/kubedb/postgres-restic-plugin/commit/b97a07f6) Prepare for release v0.28.0 (#114)
- [8b893309](https://github.com/kubedb/postgres-restic-plugin/commit/8b893309) Prepare for release v0.28.0-rc.2 (#113)
- [4702b178](https://github.com/kubedb/postgres-restic-plugin/commit/4702b178) Add restic backup progress streaming (#112)
- [78b32801](https://github.com/kubedb/postgres-restic-plugin/commit/78b32801) Bump postgres 16 image from 16.1 to 16.4
- [6984c384](https://github.com/kubedb/postgres-restic-plugin/commit/6984c384) Fix WaitForDBConnection not logging the actual connection error (#111)
- [229651a4](https://github.com/kubedb/postgres-restic-plugin/commit/229651a4) Prepare for release v0.28.0-rc.1 (#110)
- [dad161fc](https://github.com/kubedb/postgres-restic-plugin/commit/dad161fc) Prepare for release v0.28.0-rc.0 (#109)
- [e69d3dbc](https://github.com/kubedb/postgres-restic-plugin/commit/e69d3dbc) Harden release and release-tracker workflows
- [341a4452](https://github.com/kubedb/postgres-restic-plugin/commit/341a4452) Add AGENTS.md for AI coding agents
- [6b3dcd51](https://github.com/kubedb/postgres-restic-plugin/commit/6b3dcd51) Harden CI workflows (#107)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.26.0](https://github.com/kubedb/provider-aws/releases/tag/v0.26.0)

- [2b8fb0f](https://github.com/kubedb/provider-aws/commit/2b8fb0f) Tighten CI/release workflow secrets, perms, and release notes
- [2d7e8ba](https://github.com/kubedb/provider-aws/commit/2d7e8ba) Harden release and release-tracker workflows
- [c984af3](https://github.com/kubedb/provider-aws/commit/c984af3) Add AGENTS.md for AI coding agents (#44)
- [d88e77b](https://github.com/kubedb/provider-aws/commit/d88e77b) Harden CI workflows (#43)



## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.26.0](https://github.com/kubedb/provider-azure/releases/tag/v0.26.0)

- [630e43c](https://github.com/kubedb/provider-azure/commit/630e43c) Tighten CI/release workflow secrets, perms, and release notes
- [3d78f9c](https://github.com/kubedb/provider-azure/commit/3d78f9c) Add AGENTS.md for AI coding agents (#30)
- [7f7b570](https://github.com/kubedb/provider-azure/commit/7f7b570) Restrict /ok-to-test to org members (#29)



## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.26.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.26.0)

- [e8f212a](https://github.com/kubedb/provider-gcp/commit/e8f212a) Tighten CI/release workflow secrets, perms, and release notes
- [4f19b3b](https://github.com/kubedb/provider-gcp/commit/4f19b3b) Harden release and release-tracker workflows
- [ca1476e](https://github.com/kubedb/provider-gcp/commit/ca1476e) Add AGENTS.md for AI coding agents (#29)
- [4c70d45](https://github.com/kubedb/provider-gcp/commit/4c70d45) Harden CI workflows (#28)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.65.0](https://github.com/kubedb/provisioner/releases/tag/v0.65.0)




## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.52.0](https://github.com/kubedb/proxysql/releases/tag/v0.52.0)

- [420e5c0e](https://github.com/kubedb/proxysql/commit/420e5c0e5) Prepare for release v0.52.0 (#440)
- [fb2e8d61](https://github.com/kubedb/proxysql/commit/fb2e8d616) Prepare for release v0.52.0-rc.2 (#439)
- [75fcf154](https://github.com/kubedb/proxysql/commit/75fcf1540) Honor user-provided renewBefore in TLS certificate ops (#436)
- [3c4cd633](https://github.com/kubedb/proxysql/commit/3c4cd6331) Implement RotateAuthentication for ProxySQL (#432)
- [fe85fcd0](https://github.com/kubedb/proxysql/commit/fe85fcd09) Prepare for release v0.52.0-rc.1 (#438)
- [d5d1d882](https://github.com/kubedb/proxysql/commit/d5d1d882e) Prepare for release v0.52.0-rc.0 (#434)
- [bb7e268e](https://github.com/kubedb/proxysql/commit/bb7e268e4) Tighten CI/release workflow secrets, perms, and release notes
- [9c59d49c](https://github.com/kubedb/proxysql/commit/9c59d49ce) Harden release and release-tracker workflows
- [e29d1669](https://github.com/kubedb/proxysql/commit/e29d16693) Run Ops Request Locally (#433)
- [b47e8c2d](https://github.com/kubedb/proxysql/commit/b47e8c2d3) Add CLAUDE.md pointing to AGENTS.md
- [681acb3c](https://github.com/kubedb/proxysql/commit/681acb3c4) Add AGENTS.md for AI coding agents
- [3b6a4061](https://github.com/kubedb/proxysql/commit/3b6a40616) Harden CI workflows (#430)



## [kubedb/qdrant](https://github.com/kubedb/qdrant)

### [v0.6.0](https://github.com/kubedb/qdrant/releases/tag/v0.6.0)

- [8023163e](https://github.com/kubedb/qdrant/commit/8023163e) Prepare for release v0.6.0 (#49)
- [46a46c95](https://github.com/kubedb/qdrant/commit/46a46c95) Prepare for release v0.6.0-rc.2 (#48)
- [2448f7e5](https://github.com/kubedb/qdrant/commit/2448f7e5) Honor user-provided renewBefore in TLS certificate ops (#44)
- [f9d88701](https://github.com/kubedb/qdrant/commit/f9d88701) Fix Qdrant deletion (#47)
- [6ea63e53](https://github.com/kubedb/qdrant/commit/6ea63e53) Prepare for release v0.6.0-rc.1 (#46)
- [88c4b011](https://github.com/kubedb/qdrant/commit/88c4b011) Add --network-policy-flavor flag with cilium support (#45)
- [928e3450](https://github.com/kubedb/qdrant/commit/928e3450) Add StorageMigration OpsRequest support for Qdrant (#40)
- [b144b7b6](https://github.com/kubedb/qdrant/commit/b144b7b6) Add Reconfigure TLS (#37)
- [7f7dd3e5](https://github.com/kubedb/qdrant/commit/7f7dd3e5) Prepare for release v0.6.0-rc.0 (#41)
- [a79ec8c3](https://github.com/kubedb/qdrant/commit/a79ec8c3) Tighten CI/release workflow secrets, perms, and release notes
- [6715626c](https://github.com/kubedb/qdrant/commit/6715626c) Harden release and release-tracker workflows
- [784adaa7](https://github.com/kubedb/qdrant/commit/784adaa7) Add CLAUDE.md pointing to AGENTS.md
- [4ef3db03](https://github.com/kubedb/qdrant/commit/4ef3db03) Add AGENTS.md for AI coding agents
- [e3aaca10](https://github.com/kubedb/qdrant/commit/e3aaca10) Harden CI workflows (#38)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.20.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.20.0)

- [b6ec3253](https://github.com/kubedb/rabbitmq/commit/b6ec3253) Prepare for release v0.20.0 (#140)
- [7ec5ac18](https://github.com/kubedb/rabbitmq/commit/7ec5ac18) Update github.com/moby/spdystream to v0.5.1 (#139)
- [c981cd46](https://github.com/kubedb/rabbitmq/commit/c981cd46) Prepare for release v0.20.0-rc.2 (#138)
- [1ae8fd38](https://github.com/kubedb/rabbitmq/commit/1ae8fd38) Add support for cilium network  policy (#136)
- [d1511a11](https://github.com/kubedb/rabbitmq/commit/d1511a11) Prepare for release v0.20.0-rc.1 (#137)
- [b2d406fb](https://github.com/kubedb/rabbitmq/commit/b2d406fb) Add StorageMigration OpsRequest support for RabbitMQ (#132)
- [2826d51c](https://github.com/kubedb/rabbitmq/commit/2826d51c) Prepare for release v0.20.0-rc.0 (#133)
- [7b849665](https://github.com/kubedb/rabbitmq/commit/7b849665) Tighten CI/release workflow secrets, perms, and release notes
- [7c952783](https://github.com/kubedb/rabbitmq/commit/7c952783) Harden release and release-tracker workflows
- [8b547651](https://github.com/kubedb/rabbitmq/commit/8b547651) Add CLAUDE.md pointing to AGENTS.md
- [866a6b3c](https://github.com/kubedb/rabbitmq/commit/866a6b3c) Add AGENTS.md for AI coding agents
- [2a733db2](https://github.com/kubedb/rabbitmq/commit/2a733db2) Harden CI workflows (#130)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.58.0](https://github.com/kubedb/redis/releases/tag/v0.58.0)

- [d18bffe8](https://github.com/kubedb/redis/commit/d18bffe83) Prepare for release v0.58.0 (#653)
- [551bdca0](https://github.com/kubedb/redis/commit/551bdca01) vnpay empty acl bug fix (#650)
- [c811b5e4](https://github.com/kubedb/redis/commit/c811b5e4c) Update github.com/moby/spdystream to v0.5.1 (#652)
- [38804ba3](https://github.com/kubedb/redis/commit/38804ba31) Prepare for release v0.58.0-rc.2 (#651)
- [05f9d6d5](https://github.com/kubedb/redis/commit/05f9d6d50) Honor user-provided renewBefore in TLS certificate ops (#647)
- [ca8050ef](https://github.com/kubedb/redis/commit/ca8050ef7) Prepare for release v0.58.0-rc.1 (#649)
- [82e9701e](https://github.com/kubedb/redis/commit/82e9701e5) Add --network-policy-flavor flag with cilium support (#648)
- [7f56a8b9](https://github.com/kubedb/redis/commit/7f56a8b98) Health Check updated (#641)
- [7452ec78](https://github.com/kubedb/redis/commit/7452ec78a) Add StorageMigration OpsRequest support for Redis (#644)
- [21037f13](https://github.com/kubedb/redis/commit/21037f135) Merge ACL in reconfigure merger (#642)
- [3afe56ea](https://github.com/kubedb/redis/commit/3afe56eab) Prepare for release v0.58.0-rc.0 (#646)
- [c961cdb4](https://github.com/kubedb/redis/commit/c961cdb42) Tighten CI/release workflow secrets, perms, and release notes
- [a09720d5](https://github.com/kubedb/redis/commit/a09720d5b) Harden release and release-tracker workflows
- [7d00abef](https://github.com/kubedb/redis/commit/7d00abef9) Run Ops Request Locally (#645)
- [740c5c96](https://github.com/kubedb/redis/commit/740c5c968) Add CLAUDE.md pointing to AGENTS.md
- [d23b442c](https://github.com/kubedb/redis/commit/d23b442c3) Add AGENTS.md for AI coding agents
- [d5c294b0](https://github.com/kubedb/redis/commit/d5c294b03) Harden CI workflows (#640)
- [34c6e5d5](https://github.com/kubedb/redis/commit/34c6e5d56) Add governing svc name in cert (#639)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.44.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.44.0)

- [bf16a4c7](https://github.com/kubedb/redis-coordinator/commit/bf16a4c7) Prepare for release v0.44.0 (#163)
- [8cc88c9c](https://github.com/kubedb/redis-coordinator/commit/8cc88c9c) Prepare for release v0.44.0-rc.2 (#162)
- [b0fc9339](https://github.com/kubedb/redis-coordinator/commit/b0fc9339) Prepare for release v0.44.0-rc.1 (#161)
- [367048d8](https://github.com/kubedb/redis-coordinator/commit/367048d8) Prepare for release v0.44.0-rc.0 (#160)
- [a083e153](https://github.com/kubedb/redis-coordinator/commit/a083e153) Tighten CI/release workflow secrets, perms, and release notes
- [25c46735](https://github.com/kubedb/redis-coordinator/commit/25c46735) Harden release and release-tracker workflows
- [928c964e](https://github.com/kubedb/redis-coordinator/commit/928c964e) Add AGENTS.md for AI coding agents
- [90c797de](https://github.com/kubedb/redis-coordinator/commit/90c797de) Harden CI workflows (#158)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.28.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.28.0)

- [982e8bd5](https://github.com/kubedb/redis-restic-plugin/commit/982e8bd5) Prepare for release v0.28.0 (#110)
- [c9a8d01b](https://github.com/kubedb/redis-restic-plugin/commit/c9a8d01b) Prepare for release v0.28.0-rc.2 (#109)
- [eaeb0c7a](https://github.com/kubedb/redis-restic-plugin/commit/eaeb0c7a) Add restic backup progress streaming (#108)
- [b2769079](https://github.com/kubedb/redis-restic-plugin/commit/b2769079) Prepare for release v0.28.0-rc.1 (#107)
- [96e95cee](https://github.com/kubedb/redis-restic-plugin/commit/96e95cee) Prepare for release v0.28.0-rc.0 (#106)
- [044d20f6](https://github.com/kubedb/redis-restic-plugin/commit/044d20f6) Harden release and release-tracker workflows
- [2c524ff2](https://github.com/kubedb/redis-restic-plugin/commit/2c524ff2) Add AGENTS.md for AI coding agents
- [e29aaa32](https://github.com/kubedb/redis-restic-plugin/commit/e29aaa32) Harden CI workflows (#104)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.52.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.52.0)

- [93e34492](https://github.com/kubedb/replication-mode-detector/commit/93e34492) Prepare for release v0.52.0 (#326)
- [fe05bb5e](https://github.com/kubedb/replication-mode-detector/commit/fe05bb5e) Prepare for release v0.52.0-rc.2 (#325)
- [adc96bf9](https://github.com/kubedb/replication-mode-detector/commit/adc96bf9) Prepare for release v0.52.0-rc.1 (#324)
- [c2ca4a03](https://github.com/kubedb/replication-mode-detector/commit/c2ca4a03) Prepare for release v0.52.0-rc.0 (#323)
- [a22986ef](https://github.com/kubedb/replication-mode-detector/commit/a22986ef) Tighten CI/release workflow secrets, perms, and release notes
- [7f22be9e](https://github.com/kubedb/replication-mode-detector/commit/7f22be9e) Harden release and release-tracker workflows
- [e8f81f8e](https://github.com/kubedb/replication-mode-detector/commit/e8f81f8e) Add AGENTS.md for AI coding agents
- [3c2fd6f5](https://github.com/kubedb/replication-mode-detector/commit/3c2fd6f5) Harden CI workflows (#321)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.41.0](https://github.com/kubedb/schema-manager/releases/tag/v0.41.0)

- [e9c20a65](https://github.com/kubedb/schema-manager/commit/e9c20a65) Prepare for release v0.41.0 (#174)
- [5444ca3e](https://github.com/kubedb/schema-manager/commit/5444ca3e) Update github.com/moby/spdystream to v0.5.1 (#173)
- [380ec889](https://github.com/kubedb/schema-manager/commit/380ec889) Prepare for release v0.41.0-rc.2 (#172)
- [3d0da9a2](https://github.com/kubedb/schema-manager/commit/3d0da9a2) Prepare for release v0.41.0-rc.1 (#171)
- [c31cfb97](https://github.com/kubedb/schema-manager/commit/c31cfb97) Prepare for release v0.41.0-rc.0 (#170)
- [7f62c921](https://github.com/kubedb/schema-manager/commit/7f62c921) Tighten CI/release workflow secrets, perms, and release notes
- [bba504bd](https://github.com/kubedb/schema-manager/commit/bba504bd) Harden release and release-tracker workflows
- [95b80652](https://github.com/kubedb/schema-manager/commit/95b80652) Add AGENTS.md for AI coding agents
- [c6495b95](https://github.com/kubedb/schema-manager/commit/c6495b95) Harden CI workflows (#168)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.20.0](https://github.com/kubedb/singlestore/releases/tag/v0.20.0)

- [f1d36604](https://github.com/kubedb/singlestore/commit/f1d36604) Prepare for release v0.20.0 (#131)
- [a4098c82](https://github.com/kubedb/singlestore/commit/a4098c82) Update github.com/moby/spdystream to v0.5.1 (#130)
- [f08e3952](https://github.com/kubedb/singlestore/commit/f08e3952) feat: implement git-sync init container for Singlestore (#124)
- [5be14e4d](https://github.com/kubedb/singlestore/commit/5be14e4d) Prepare for release v0.20.0-rc.2 (#129)
- [98d3f455](https://github.com/kubedb/singlestore/commit/98d3f455) Fix SingleStore deletion (#128)
- [5c8a7b45](https://github.com/kubedb/singlestore/commit/5c8a7b45) Honor user-provided renewBefore in TLS certificate ops (#125)
- [d6a82975](https://github.com/kubedb/singlestore/commit/d6a82975) Add RotateAuth Ops Request Support (#122)
- [0ac62594](https://github.com/kubedb/singlestore/commit/0ac62594) Prepare for release v0.20.0-rc.1 (#127)
- [9860ed42](https://github.com/kubedb/singlestore/commit/9860ed42) Add --network-policy-flavor flag with cilium support (#126)
- [9b843f68](https://github.com/kubedb/singlestore/commit/9b843f68) Add StorageMigration OpsRequest support for Singlestore (#120)
- [8d5c9087](https://github.com/kubedb/singlestore/commit/8d5c9087) Prepare for release v0.20.0-rc.0 (#121)
- [26b67dc6](https://github.com/kubedb/singlestore/commit/26b67dc6) Tighten CI/release workflow secrets, perms, and release notes
- [01a4c59f](https://github.com/kubedb/singlestore/commit/01a4c59f) Harden release and release-tracker workflows
- [d663617d](https://github.com/kubedb/singlestore/commit/d663617d) Add CLAUDE.md pointing to AGENTS.md
- [e4255529](https://github.com/kubedb/singlestore/commit/e4255529) Add AGENTS.md for AI coding agents
- [90f8831e](https://github.com/kubedb/singlestore/commit/90f8831e) Harden CI workflows (#118)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.20.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.20.0)

- [87d71d04](https://github.com/kubedb/singlestore-coordinator/commit/87d71d04) Prepare for release v0.20.0 (#76)
- [5da3ad74](https://github.com/kubedb/singlestore-coordinator/commit/5da3ad74) Update github.com/moby/spdystream to v0.5.1 (#75)
- [7f87bc81](https://github.com/kubedb/singlestore-coordinator/commit/7f87bc81) Prepare for release v0.20.0-rc.2 (#74)
- [ae11d990](https://github.com/kubedb/singlestore-coordinator/commit/ae11d990) Prepare for release v0.20.0-rc.1 (#73)
- [d287dabd](https://github.com/kubedb/singlestore-coordinator/commit/d287dabd) Prepare for release v0.20.0-rc.0 (#72)
- [fcb2528e](https://github.com/kubedb/singlestore-coordinator/commit/fcb2528e) Tighten CI/release workflow secrets, perms, and release notes
- [0f78329e](https://github.com/kubedb/singlestore-coordinator/commit/0f78329e) Harden release and release-tracker workflows
- [0ff7dc5b](https://github.com/kubedb/singlestore-coordinator/commit/0ff7dc5b) Add AGENTS.md for AI coding agents
- [357d9a67](https://github.com/kubedb/singlestore-coordinator/commit/357d9a67) Harden CI workflows (#70)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.23.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.23.0)

- [c26da542](https://github.com/kubedb/singlestore-restic-plugin/commit/c26da542) Prepare for release v0.23.0 (#89)
- [1cbde534](https://github.com/kubedb/singlestore-restic-plugin/commit/1cbde534) Prepare for release v0.23.0-rc.2 (#88)
- [0ac6f752](https://github.com/kubedb/singlestore-restic-plugin/commit/0ac6f752) Add restic backup progress streaming (#87)
- [d6832394](https://github.com/kubedb/singlestore-restic-plugin/commit/d6832394) Prepare for release v0.23.0-rc.1 (#86)
- [ef6b233c](https://github.com/kubedb/singlestore-restic-plugin/commit/ef6b233c) Prepare for release v0.23.0-rc.0 (#85)
- [ecfc7fc6](https://github.com/kubedb/singlestore-restic-plugin/commit/ecfc7fc6) Harden release and release-tracker workflows
- [88b88a43](https://github.com/kubedb/singlestore-restic-plugin/commit/88b88a43) Add AGENTS.md for AI coding agents
- [9e76f60c](https://github.com/kubedb/singlestore-restic-plugin/commit/9e76f60c) Harden CI workflows (#83)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.20.0](https://github.com/kubedb/solr/releases/tag/v0.20.0)

- [c210af00](https://github.com/kubedb/solr/commit/c210af00) Prepare for release v0.20.0 (#137)
- [5748039b](https://github.com/kubedb/solr/commit/5748039b) Add StorageMigration OpsRequest support (#129)
- [ace6b48c](https://github.com/kubedb/solr/commit/ace6b48c) Fix AuthSecret (#135)
- [1c955c70](https://github.com/kubedb/solr/commit/1c955c70) Prepare for release v0.20.0-rc.2 (#136)
- [443bf0ad](https://github.com/kubedb/solr/commit/443bf0ad) Fix Solr deletion (#134)
- [0d6417ef](https://github.com/kubedb/solr/commit/0d6417ef) Prepare for release v0.20.0-rc.1 (#133)
- [4f32f195](https://github.com/kubedb/solr/commit/4f32f195) Add --network-policy-flavor flag with cilium support (#132)
- [d6f12bb7](https://github.com/kubedb/solr/commit/d6f12bb7) Prepare for release v0.20.0-rc.0 (#130)
- [b947dd5c](https://github.com/kubedb/solr/commit/b947dd5c) Tighten CI/release workflow secrets, perms, and release notes
- [5a339830](https://github.com/kubedb/solr/commit/5a339830) Harden release and release-tracker workflows
- [59801485](https://github.com/kubedb/solr/commit/59801485) Add CLAUDE.md pointing to AGENTS.md
- [8e55bf7a](https://github.com/kubedb/solr/commit/8e55bf7a) Add AGENTS.md for AI coding agents
- [a91a2ae7](https://github.com/kubedb/solr/commit/a91a2ae7) Harden CI workflows (#127)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.50.0](https://github.com/kubedb/tests/releases/tag/v0.50.0)

- [df07cb41](https://github.com/kubedb/tests/commit/df07cb414) Prepare for release v0.50.0 (#543)
- [e2c2cc9e](https://github.com/kubedb/tests/commit/e2c2cc9eb) Update github.com/moby/spdystream to v0.5.1 (#542)
- [5829f980](https://github.com/kubedb/tests/commit/5829f9808) Prepare for release v0.50.0-rc.2 (#541)
- [66234e49](https://github.com/kubedb/tests/commit/66234e492) Update kubedb apimachinery vendor and fix AppBinding pointer type (#540)
- [957fa4e5](https://github.com/kubedb/tests/commit/957fa4e5d) Add E2E tests for ClickHouse (#525)
- [59d384bc](https://github.com/kubedb/tests/commit/59d384bc4) Prepare for release v0.50.0-rc.1 (#538)
- [d5a306ba](https://github.com/kubedb/tests/commit/d5a306baf) Remove FerretDB support (#537)
- [7d4ad629](https://github.com/kubedb/tests/commit/7d4ad629a) Prepare for release v0.50.0-rc.0 (#536)
- [2ad57dc7](https://github.com/kubedb/tests/commit/2ad57dc71) Harden release and release-tracker workflows
- [27074f61](https://github.com/kubedb/tests/commit/27074f613) Add CLAUDE.md pointing to AGENTS.md
- [e5e62b8a](https://github.com/kubedb/tests/commit/e5e62b8a0) Add AGENTS.md for AI coding agents
- [9ee1c9bc](https://github.com/kubedb/tests/commit/9ee1c9bc7) Harden CI workflows (#522)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.41.0](https://github.com/kubedb/ui-server/releases/tag/v0.41.0)

- [aa12c7c3](https://github.com/kubedb/ui-server/commit/aa12c7c3a) Prepare for release v0.41.0 (#210)
- [2d9cc487](https://github.com/kubedb/ui-server/commit/2d9cc4875) Prepare for release v0.41.0-rc.2 (#209)
- [62d7f163](https://github.com/kubedb/ui-server/commit/62d7f163b) Prepare for release v0.41.0-rc.1 (#208)
- [5fa49a75](https://github.com/kubedb/ui-server/commit/5fa49a752) Remove FerretDB support (#207)
- [194e5646](https://github.com/kubedb/ui-server/commit/194e56464) Update api (#206)
- [b61ac13e](https://github.com/kubedb/ui-server/commit/b61ac13e1) DatabaseInfo -> DatabaseConfiguration (#205)
- [90f6cf68](https://github.com/kubedb/ui-server/commit/90f6cf683) Implement summary api (#204)
- [b0cae9e1](https://github.com/kubedb/ui-server/commit/b0cae9e1b) Prepare for release v0.41.0-rc.0 (#203)
- [8633e481](https://github.com/kubedb/ui-server/commit/8633e4813) Tighten CI/release workflow secrets, perms, and release notes
- [ed171989](https://github.com/kubedb/ui-server/commit/ed1719891) Harden release and release-tracker workflows
- [1709ef94](https://github.com/kubedb/ui-server/commit/1709ef941) Pass componenetName field & Refactor (#201)
- [d171e2c8](https://github.com/kubedb/ui-server/commit/d171e2c8c) Add AGENTS.md for AI coding agents
- [9e3b8aed](https://github.com/kubedb/ui-server/commit/9e3b8aed8) Harden CI workflows (#200)



## [kubedb/weaviate](https://github.com/kubedb/weaviate)

### [v0.6.0](https://github.com/kubedb/weaviate/releases/tag/v0.6.0)

- [32d0bbd6](https://github.com/kubedb/weaviate/commit/32d0bbd6) Prepare for release v0.6.0 (#45)
- [ef1c2dc0](https://github.com/kubedb/weaviate/commit/ef1c2dc0) TLS (#44)
- [7226c70c](https://github.com/kubedb/weaviate/commit/7226c70c) Prepare for release v0.6.0-rc.2 (#43)
- [9d3b6c57](https://github.com/kubedb/weaviate/commit/9d3b6c57) weaviate restart ops-request added (#12)
- [14d862e3](https://github.com/kubedb/weaviate/commit/14d862e3) Fix Weaviate deletion (#41)
- [690c3ad0](https://github.com/kubedb/weaviate/commit/690c3ad0) Add NetworkPolicyFlavor support (#40)
- [513e4bf4](https://github.com/kubedb/weaviate/commit/513e4bf4) Prepare for release v0.6.0-rc.1 (#39)
- [bb3c8918](https://github.com/kubedb/weaviate/commit/bb3c8918) Prepare for release v0.6.0-rc.0 (#35)
- [a77224f3](https://github.com/kubedb/weaviate/commit/a77224f3) Tighten CI/release workflow secrets, perms, and release notes
- [716556bd](https://github.com/kubedb/weaviate/commit/716556bd) Harden release and release-tracker workflows
- [7131fe7e](https://github.com/kubedb/weaviate/commit/7131fe7e) Add CLAUDE.md pointing to AGENTS.md
- [bb95826e](https://github.com/kubedb/weaviate/commit/bb95826e) Add AGENTS.md for AI coding agents
- [3dfc4fe3](https://github.com/kubedb/weaviate/commit/3dfc4fe3) Harden CI workflows (#29)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.41.0](https://github.com/kubedb/webhook-server/releases/tag/v0.41.0)

- [9309bb35](https://github.com/kubedb/webhook-server/commit/9309bb354) Prepare for release v0.41.0 (#226)
- [9cd37a06](https://github.com/kubedb/webhook-server/commit/9cd37a063) Add Aerospike webhook server (#225)
- [2db5df91](https://github.com/kubedb/webhook-server/commit/2db5df913) Prepare for release v0.41.0-rc.2 (#224)
- [3d375b73](https://github.com/kubedb/webhook-server/commit/3d375b730) documentdb-ops-webhook (#223)
- [cc9b766e](https://github.com/kubedb/webhook-server/commit/cc9b766ed) Add Cassandra & ClickHouse ops and Neo4j autoscaler webhook registrations (#222)
- [e341c482](https://github.com/kubedb/webhook-server/commit/e341c482f) create oracle ops reconfigure (#211)
- [65259ba2](https://github.com/kubedb/webhook-server/commit/65259ba23) Add HanaDB ops (#216)
- [4fd5117d](https://github.com/kubedb/webhook-server/commit/4fd5117db) Add Milvus Autoscaler Validation (#221)
- [0b56463a](https://github.com/kubedb/webhook-server/commit/0b56463a7) Add weaviate ops (#218)
- [29608ec8](https://github.com/kubedb/webhook-server/commit/29608ec8b) Prepare for release v0.41.0-rc.1 (#220)
- [9baa5eb5](https://github.com/kubedb/webhook-server/commit/9baa5eb56) Remove FerretDB support (#219)
- [0f68066f](https://github.com/kubedb/webhook-server/commit/0f68066fa) Setup Qdrant Autoscaler Webhook (#209)
- [fe08c7c4](https://github.com/kubedb/webhook-server/commit/fe08c7c4d) Add Milvus OPS Webhook (#217)
- [8009a9ef](https://github.com/kubedb/webhook-server/commit/8009a9ef8) Fix CI hardening: use app token in release-tracker, add packages: (#215)
- [367f75a7](https://github.com/kubedb/webhook-server/commit/367f75a72) Prepare for release v0.41.0-rc.0 (#214)
- [79e9b2bd](https://github.com/kubedb/webhook-server/commit/79e9b2bd4) Tighten CI/release workflow secrets, perms, and release notes
- [6e9a0640](https://github.com/kubedb/webhook-server/commit/6e9a06404) Add AGENTS.md for AI coding agents
- [dddcda1f](https://github.com/kubedb/webhook-server/commit/dddcda1f0) Restrict /ok-to-test to org members (#212)



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.13.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.13.0)

- [5dd2507e](https://github.com/kubedb/xtrabackup-restic-plugin/commit/5dd2507e) Prepare for release v0.13.0 (#58)
- [7405795f](https://github.com/kubedb/xtrabackup-restic-plugin/commit/7405795f) Prepare for release v0.13.0-rc.2 (#57)
- [3c66fa34](https://github.com/kubedb/xtrabackup-restic-plugin/commit/3c66fa34) Add restic backup progress streaming (#56)
- [36817455](https://github.com/kubedb/xtrabackup-restic-plugin/commit/36817455) Prepare for release v0.13.0-rc.1 (#55)
- [71c71b57](https://github.com/kubedb/xtrabackup-restic-plugin/commit/71c71b57) use package: write and fetch-depth: 0 (#54)
- [fd5f2106](https://github.com/kubedb/xtrabackup-restic-plugin/commit/fd5f2106) Prepare for release v0.13.0-rc.0 (#53)
- [1cf78111](https://github.com/kubedb/xtrabackup-restic-plugin/commit/1cf78111) Add AGENTS.md for AI coding agents
- [14f542a3](https://github.com/kubedb/xtrabackup-restic-plugin/commit/14f542a3) Use GitHub App token for release tracker comments (#51)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.20.0](https://github.com/kubedb/zookeeper/releases/tag/v0.20.0)

- [9cbd2e6d](https://github.com/kubedb/zookeeper/commit/9cbd2e6d) Prepare for release v0.20.0 (#128)
- [23412a6b](https://github.com/kubedb/zookeeper/commit/23412a6b) Prepare for release v0.20.0-rc.2 (#127)
- [b3e8b477](https://github.com/kubedb/zookeeper/commit/b3e8b477) Prepare for release v0.20.0-rc.1 (#125)
- [37106c9c](https://github.com/kubedb/zookeeper/commit/37106c9c) Add --network-policy-flavor flag with cilium support (#124)
- [dca38f13](https://github.com/kubedb/zookeeper/commit/dca38f13) Use PatchStatus instead of CreateOrPatch to avoid timing issues on db deletion (#122)
- [a572da63](https://github.com/kubedb/zookeeper/commit/a572da63) Prepare for release v0.20.0-rc.0 (#121)
- [f583941f](https://github.com/kubedb/zookeeper/commit/f583941f) Tighten CI/release workflow secrets, perms, and release notes
- [d4475edf](https://github.com/kubedb/zookeeper/commit/d4475edf) Harden release and release-tracker workflows
- [7a658949](https://github.com/kubedb/zookeeper/commit/7a658949) Add CLAUDE.md pointing to AGENTS.md
- [d06d8429](https://github.com/kubedb/zookeeper/commit/d06d8429) Add AGENTS.md for AI coding agents
- [36e65048](https://github.com/kubedb/zookeeper/commit/36e65048) Harden CI workflows (#119)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.20.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.20.0)

- [2f27830e](https://github.com/kubedb/zookeeper-restic-plugin/commit/2f27830e) Prepare for release v0.20.0 (#74)
- [74c137ba](https://github.com/kubedb/zookeeper-restic-plugin/commit/74c137ba) Prepare for release v0.20.0-rc.2 (#73)
- [308c991d](https://github.com/kubedb/zookeeper-restic-plugin/commit/308c991d) Add restic backup progress streaming (#72)
- [b5ff2162](https://github.com/kubedb/zookeeper-restic-plugin/commit/b5ff2162) Prepare for release v0.20.0-rc.1 (#71)
- [a6878a44](https://github.com/kubedb/zookeeper-restic-plugin/commit/a6878a44) Fix release workflow regressions from CI hardening (#70)
- [5124f6a7](https://github.com/kubedb/zookeeper-restic-plugin/commit/5124f6a7) Prepare for release v0.20.0-rc.0 (#69)
- [f01b3b87](https://github.com/kubedb/zookeeper-restic-plugin/commit/f01b3b87) Tighten CI/release workflow secrets, perms, and release notes
- [85b763fa](https://github.com/kubedb/zookeeper-restic-plugin/commit/85b763fa) Harden release and release-tracker workflows
- [0cc56540](https://github.com/kubedb/zookeeper-restic-plugin/commit/0cc56540) Add AGENTS.md for AI coding agents
- [391b6eb7](https://github.com/kubedb/zookeeper-restic-plugin/commit/391b6eb7) Harden CI workflows (#67)




