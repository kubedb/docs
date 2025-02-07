---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2025.2.6-rc.0
    name: Changelog-v2025.2.6-rc.0
    parent: welcome
    weight: 20250206
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2025.2.6-rc.0/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2025.2.6-rc.0/
---

# KubeDB v2025.2.6-rc.0 (2025-02-07)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.52.0-rc.0](https://github.com/kubedb/apimachinery/releases/tag/v0.52.0-rc.0)

- [d3ee4caa](https://github.com/kubedb/apimachinery/commit/d3ee4caa7) Update deps
- [f68a76c7](https://github.com/kubedb/apimachinery/commit/f68a76c7d) make fmt
- [1ba01a7d](https://github.com/kubedb/apimachinery/commit/1ba01a7de) Update deps
- [b05f9709](https://github.com/kubedb/apimachinery/commit/b05f97094) Add DisableSSLSessionResumption field in PostgresVersion CRD (#1390)
- [47fd90ca](https://github.com/kubedb/apimachinery/commit/47fd90caf) Update deps
- [f435670c](https://github.com/kubedb/apimachinery/commit/f435670c9) Add InitScript restore option into manifest restore (#1388)
- [e37dc985](https://github.com/kubedb/apimachinery/commit/e37dc985f) Set maxscale Field Optional (#1389)
- [e5366d52](https://github.com/kubedb/apimachinery/commit/e5366d526) Add Support for MariaDB Replication (#1383)
- [7ad43925](https://github.com/kubedb/apimachinery/commit/7ad439250) Add MS SQL Server Arbiter APIs (#1363)
- [affa6dc5](https://github.com/kubedb/apimachinery/commit/affa6dc53) Add RabbitMQ EnableAllFeatureFlags Constant (#1386)
- [7f9ee3e6](https://github.com/kubedb/apimachinery/commit/7f9ee3e6b) Add apis for archiver manifest recovery (#1387)
- [85c4437b](https://github.com/kubedb/apimachinery/commit/85c4437b1) Add Pgpool hba and pcp file name Constant (#1377)
- [2e163519](https://github.com/kubedb/apimachinery/commit/2e163519e) add reconfigure-tls for pgbouncer (#1384)
- [a7900520](https://github.com/kubedb/apimachinery/commit/a79005204) Add SchemaRegistry with RestProxy API (#1372)
- [bdae9dd4](https://github.com/kubedb/apimachinery/commit/bdae9dd4c) Move Redis mutator defaults to SetDefaults() (#1385)
- [ce659b3f](https://github.com/kubedb/apimachinery/commit/ce659b3fc) Add license restrictions (#1382)
- [051792b4](https://github.com/kubedb/apimachinery/commit/051792b4b) Implement DBBindInterface for generic bindings (#1381)
- [7dd4fbe3](https://github.com/kubedb/apimachinery/commit/7dd4fbe3f) Pass different ca for gateway & inCluster (#1380)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.37.0-rc.0](https://github.com/kubedb/autoscaler/releases/tag/v0.37.0-rc.0)

- [281f1a30](https://github.com/kubedb/autoscaler/commit/281f1a30) Prepare for release v0.37.0-rc.0 (#238)
- [750466be](https://github.com/kubedb/autoscaler/commit/750466be) Fix nats initialization (#237)
- [51b090f7](https://github.com/kubedb/autoscaler/commit/51b090f7) Disable image caching in setup-qemu action (#236)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.5.0-rc.0](https://github.com/kubedb/cassandra/releases/tag/v0.5.0-rc.0)

- [85d398f0](https://github.com/kubedb/cassandra/commit/85d398f0) Fix lister creation for client billing
- [e5b0f6ac](https://github.com/kubedb/cassandra/commit/e5b0f6ac) Prepare for release v0.5.0-rc.0 (#17)
- [75936904](https://github.com/kubedb/cassandra/commit/75936904) Disable image caching in setup-qemu action (#18)
- [282fb04e](https://github.com/kubedb/cassandra/commit/282fb04e) Add client billing support (#16)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.52.0-rc.0](https://github.com/kubedb/cli/releases/tag/v0.52.0-rc.0)

- [6682c895](https://github.com/kubedb/cli/commit/6682c895) Prepare for release v0.52.0-rc.0 (#789)
- [ae68ef49](https://github.com/kubedb/cli/commit/ae68ef49) Disable image caching in setup-qemu action (#788)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.7.0-rc.0](https://github.com/kubedb/clickhouse/releases/tag/v0.7.0-rc.0)

- [14822270](https://github.com/kubedb/clickhouse/commit/14822270) Fix lister creation for client billing
- [4ead52f5](https://github.com/kubedb/clickhouse/commit/4ead52f5) Prepare for release v0.7.0-rc.0 (#31)
- [62144b2f](https://github.com/kubedb/clickhouse/commit/62144b2f) Disable image caching in setup-qemu action (#32)
- [60c715bf](https://github.com/kubedb/clickhouse/commit/60c715bf) Add client billing event support (#30)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.7.0-rc.0](https://github.com/kubedb/crd-manager/releases/tag/v0.7.0-rc.0)

- [07576a67](https://github.com/kubedb/crd-manager/commit/07576a67) Prepare for release v0.7.0-rc.0 (#61)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.10.0-rc.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.10.0-rc.0)

- [3c2aa0a](https://github.com/kubedb/dashboard-restic-plugin/commit/3c2aa0a) Prepare for release v0.10.0-rc.0 (#32)
- [3a34720](https://github.com/kubedb/dashboard-restic-plugin/commit/3a34720) Disable image caching in setup-qemu action (#31)
- [f189381](https://github.com/kubedb/dashboard-restic-plugin/commit/f189381) Incorporate with go-sh leaf command execution (#29)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.7.0-rc.0](https://github.com/kubedb/db-client-go/releases/tag/v0.7.0-rc.0)

- [7deb14c1](https://github.com/kubedb/db-client-go/commit/7deb14c1) Prepare for release v0.7.0-rc.0 (#161)
- [4790577b](https://github.com/kubedb/db-client-go/commit/4790577b) cert remove from pgbouncer (#158)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.7.0-rc.0](https://github.com/kubedb/druid/releases/tag/v0.7.0-rc.0)

- [31e2f5d6](https://github.com/kubedb/druid/commit/31e2f5d6) Update listers for client billing
- [a2c96244](https://github.com/kubedb/druid/commit/a2c96244) Prepare for release v0.7.0-rc.0 (#69)
- [61806784](https://github.com/kubedb/druid/commit/61806784) Disable image caching in setup-qemu action (#68)
- [2f632333](https://github.com/kubedb/druid/commit/2f632333) Add client billing event support (#67)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.52.0-rc.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.52.0-rc.0)

- [7be7d728](https://github.com/kubedb/elasticsearch/commit/7be7d7282) Prepare for release v0.52.0-rc.0 (#749)
- [953e8b52](https://github.com/kubedb/elasticsearch/commit/953e8b529) Add restriction for client billing (#750)
- [c22f6099](https://github.com/kubedb/elasticsearch/commit/c22f60990) Disable image caching in setup-qemu action (#748)
- [39d72828](https://github.com/kubedb/elasticsearch/commit/39d728289) Add client billing event support (#747)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.15.0-rc.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.15.0-rc.0)

- [6a087679](https://github.com/kubedb/elasticsearch-restic-plugin/commit/6a087679) Prepare for release v0.15.0-rc.0 (#56)
- [e06c0877](https://github.com/kubedb/elasticsearch-restic-plugin/commit/e06c0877) Add Stdin Backup Leaf Command Support (#54)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.7.0-rc.0](https://github.com/kubedb/ferretdb/releases/tag/v0.7.0-rc.0)

- [84909469](https://github.com/kubedb/ferretdb/commit/84909469) Prepare for release v0.7.0-rc.0 (#58)
- [b158b789](https://github.com/kubedb/ferretdb/commit/b158b789) Disable image caching in setup-qemu action (#57)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2025.2.6-rc.0](https://github.com/kubedb/installer/releases/tag/v2025.2.6-rc.0)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.23.0-rc.0](https://github.com/kubedb/kafka/releases/tag/v0.23.0-rc.0)

- [70d1c5ba](https://github.com/kubedb/kafka/commit/70d1c5ba) Update listers for client billing (#136)
- [a9324d11](https://github.com/kubedb/kafka/commit/a9324d11) Prepare for release v0.23.0-rc.0 (#135)
- [b19380dd](https://github.com/kubedb/kafka/commit/b19380dd) Disable image caching in setup-qemu action (#134)
- [3b96d82d](https://github.com/kubedb/kafka/commit/3b96d82d) Add client billing event support (#132)
- [b948518b](https://github.com/kubedb/kafka/commit/b948518b) Add SchemaRegistry with Rest Proxy  (#128)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.28.0-rc.0](https://github.com/kubedb/kibana/releases/tag/v0.28.0-rc.0)

- [b447af9a](https://github.com/kubedb/kibana/commit/b447af9a) Prepare for release v0.28.0-rc.0 (#137)
- [e1c3a494](https://github.com/kubedb/kibana/commit/e1c3a494) Update reconcilestate and billing restrictions (#138)
- [82f37250](https://github.com/kubedb/kibana/commit/82f37250) Disable image caching in setup-qemu action (#136)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.15.0-rc.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.15.0-rc.0)

- [f85c4ee](https://github.com/kubedb/kubedb-manifest-plugin/commit/f85c4ee) Prepare for release v0.15.0-rc.0 (#88)
- [155aea8](https://github.com/kubedb/kubedb-manifest-plugin/commit/155aea8) Disable image caching in setup-qemu action (#87)
- [1c168b9](https://github.com/kubedb/kubedb-manifest-plugin/commit/1c168b9) Archiver and InitScript support added for backup and restore (#84)
- [e64809f](https://github.com/kubedb/kubedb-manifest-plugin/commit/e64809f) Incorporate with `go-sh` leaf command execution (#85)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.3.0-rc.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.3.0-rc.0)

- [3648382](https://github.com/kubedb/kubedb-verifier/commit/3648382) Prepare for release v0.3.0-rc.0 (#11)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.36.0-rc.0](https://github.com/kubedb/mariadb/releases/tag/v0.36.0-rc.0)

- [b4aa7696](https://github.com/kubedb/mariadb/commit/b4aa76961) Prepare for release v0.36.0-rc.0 (#308)
- [35f8872b](https://github.com/kubedb/mariadb/commit/35f8872bb) Disable image caching in setup-qemu action (#309)
- [1ae948a1](https://github.com/kubedb/mariadb/commit/1ae948a15) Add client billing event support (#306)
- [163d01c1](https://github.com/kubedb/mariadb/commit/163d01c11) Added archiver and init-script manifest restore support (#305)
- [c32d1af6](https://github.com/kubedb/mariadb/commit/c32d1af68) Fix AuthSecret Check (#296)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.12.0-rc.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.12.0-rc.0)

- [e9ac111d](https://github.com/kubedb/mariadb-archiver/commit/e9ac111d) Prepare for release v0.12.0-rc.0 (#42)
- [40d97244](https://github.com/kubedb/mariadb-archiver/commit/40d97244) Prepare for release v0.12.0-rc.0 (#41)
- [c156d207](https://github.com/kubedb/mariadb-archiver/commit/c156d207) Disable image caching in setup-qemu action (#40)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.32.0-rc.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.32.0-rc.0)

- [ff0a938a](https://github.com/kubedb/mariadb-coordinator/commit/ff0a938a) Prepare for release v0.32.0-rc.0 (#137)
- [efb48fb1](https://github.com/kubedb/mariadb-coordinator/commit/efb48fb1) Disable image caching in setup-qemu action (#136)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.12.0-rc.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.12.0-rc.0)

- [b94da98](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/b94da98) Prepare for release v0.12.0-rc.0 (#40)
- [2106e19](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/2106e19) Disable image caching in setup-qemu action (#39)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.10.0-rc.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.10.0-rc.0)

- [1838e52](https://github.com/kubedb/mariadb-restic-plugin/commit/1838e52) Prepare for release v0.10.0-rc.0 (#38)
- [ed072cc](https://github.com/kubedb/mariadb-restic-plugin/commit/ed072cc) Disable image caching in setup-qemu action (#39)
- [e7cfb82](https://github.com/kubedb/mariadb-restic-plugin/commit/e7cfb82) Add Stdin Backup Leaf Command support (#36)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.45.0-rc.0](https://github.com/kubedb/memcached/releases/tag/v0.45.0-rc.0)

- [1d14d49c](https://github.com/kubedb/memcached/commit/1d14d49c6) Prepare for release v0.45.0-rc.0 (#482)
- [1d3054df](https://github.com/kubedb/memcached/commit/1d3054dfa) Disable image caching in setup-qemu action (#483)
- [90fd0801](https://github.com/kubedb/memcached/commit/90fd08014) Add client billing event support (#480)
- [2e569244](https://github.com/kubedb/memcached/commit/2e5692449) Fix memcached config secret issue (#479)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.45.0-rc.0](https://github.com/kubedb/mongodb/releases/tag/v0.45.0-rc.0)

- [5fb8819c](https://github.com/kubedb/mongodb/commit/5fb8819ca) Prepare for release v0.45.0-rc.0 (#683)
- [38062196](https://github.com/kubedb/mongodb/commit/380621967) Disable image caching in setup-qemu action (#682)
- [9d9dfdae](https://github.com/kubedb/mongodb/commit/9d9dfdaea) Enable license restriction and client billing (#681)
- [9d6457c8](https://github.com/kubedb/mongodb/commit/9d6457c82) Added archiver and init-script manifest restore support (#679)
- [d295e56a](https://github.com/kubedb/mongodb/commit/d295e56a9) remove `replicas >= 1` validation from replicaset (#677)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.13.0-rc.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.13.0-rc.0)

- [9871552](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/9871552) Prepare for release v0.13.0-rc.0 (#45)
- [92999f5](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/92999f5) Disable image caching in setup-qemu action (#44)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.15.0-rc.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.15.0-rc.0)

- [c06b10a](https://github.com/kubedb/mongodb-restic-plugin/commit/c06b10a) Prepare for release v0.15.0-rc.0 (#78)
- [c1892f4](https://github.com/kubedb/mongodb-restic-plugin/commit/c1892f4) Disable image caching in setup-qemu action (#77)
- [c37543b](https://github.com/kubedb/mongodb-restic-plugin/commit/c37543b) Add Stdin Backup Leaf Command support (#75)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.7.0-rc.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.7.0-rc.0)

- [f5dd10a7](https://github.com/kubedb/mssql-coordinator/commit/f5dd10a7) Prepare for release v0.7.0-rc.0 (#29)
- [c570b028](https://github.com/kubedb/mssql-coordinator/commit/c570b028) Disable image caching in setup-qemu action (#28)
- [e285bb2c](https://github.com/kubedb/mssql-coordinator/commit/e285bb2c) Add Arbiter node support (#24)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.7.0-rc.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.7.0-rc.0)

- [25d9ee44](https://github.com/kubedb/mssqlserver/commit/25d9ee44) Prepare for release v0.7.0-rc.0 (#59)
- [3a9b02e7](https://github.com/kubedb/mssqlserver/commit/3a9b02e7) Update listers for client billing (#58)
- [4bb771b1](https://github.com/kubedb/mssqlserver/commit/4bb771b1) Disable image caching in setup-qemu action (#57)
- [bdf3175d](https://github.com/kubedb/mssqlserver/commit/bdf3175d) Enable license restriction and client billing (#55)
- [721f3cfe](https://github.com/kubedb/mssqlserver/commit/721f3cfe) Added archiver and init-script manifest restore support (#54)
- [fd1c8cfb](https://github.com/kubedb/mssqlserver/commit/fd1c8cfb) Add arbiter node support for quorum in even-sized clusters (#46)
- [d95108ae](https://github.com/kubedb/mssqlserver/commit/d95108ae) Set archiver ref into db while label set (#53)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.6.0-rc.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.6.0-rc.0)

- [cea73fd](https://github.com/kubedb/mssqlserver-archiver/commit/cea73fd) Disable image caching in setup-qemu action (#10)



## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.6.0-rc.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.6.0-rc.0)

- [41265e0](https://github.com/kubedb/mssqlserver-walg-plugin/commit/41265e0) Prepare for release v0.6.0-rc.0 (#18)
- [cc57300](https://github.com/kubedb/mssqlserver-walg-plugin/commit/cc57300) Disable image caching in setup-qemu action (#17)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.45.0-rc.0](https://github.com/kubedb/mysql/releases/tag/v0.45.0-rc.0)

- [c670388e](https://github.com/kubedb/mysql/commit/c670388e0) Prepare for release v0.45.0-rc.0 (#665)
- [fc288a35](https://github.com/kubedb/mysql/commit/fc288a35b) Disable image caching in setup-qemu action (#666)
- [4453af61](https://github.com/kubedb/mysql/commit/4453af610) Update BinLog Snapshot status (#659)
- [5cea0f13](https://github.com/kubedb/mysql/commit/5cea0f13f) Add client billing event support (#663)
- [e8c264e1](https://github.com/kubedb/mysql/commit/e8c264e1b) Added archiver and init-script manifest restore support (#661)
- [174fdd9e](https://github.com/kubedb/mysql/commit/174fdd9ed) Add --all-databases arg for v5.7.x (#662)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.13.0-rc.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.13.0-rc.0)

- [11c6e83c](https://github.com/kubedb/mysql-archiver/commit/11c6e83c) Prepare for release v0.13.0-rc.0 (#52)
- [364b3500](https://github.com/kubedb/mysql-archiver/commit/364b3500) Disable image caching in setup-qemu action (#53)
- [f7833baf](https://github.com/kubedb/mysql-archiver/commit/f7833baf) Update Inc Snapshot status (#49)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.30.0-rc.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.30.0-rc.0)

- [40e90068](https://github.com/kubedb/mysql-coordinator/commit/40e90068) Prepare for release v0.30.0-rc.0 (#133)
- [6b61ec27](https://github.com/kubedb/mysql-coordinator/commit/6b61ec27) Disable image caching in setup-qemu action (#134)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.13.0-rc.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.13.0-rc.0)

- [33f6918](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/33f6918) Prepare for release v0.13.0-rc.0 (#40)
- [9f91792](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/9f91792) Disable image caching in setup-qemu action (#41)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.15.0-rc.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.15.0-rc.0)

- [8acee3b](https://github.com/kubedb/mysql-restic-plugin/commit/8acee3b) Prepare for release v0.15.0-rc.0 (#67)
- [b2144fb](https://github.com/kubedb/mysql-restic-plugin/commit/b2144fb) Disable image caching in setup-qemu action (#68)
- [6830436](https://github.com/kubedb/mysql-restic-plugin/commit/6830436) Add Stdin Backup Leaf Command support (#61)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.30.0-rc.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.30.0-rc.0)

- [0ba8326](https://github.com/kubedb/mysql-router-init/commit/0ba8326) Disable image caching in setup-qemu action (#50)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.39.0-rc.0](https://github.com/kubedb/ops-manager/releases/tag/v0.39.0-rc.0)

- [f8bc59dd](https://github.com/kubedb/ops-manager/commit/f8bc59dde) Prepare for release v0.39.0-rc.0 (#708)
- [f6b05657](https://github.com/kubedb/ops-manager/commit/f6b05657e) Fix nats initialization
- [ad10ad5a](https://github.com/kubedb/ops-manager/commit/ad10ad5a1) Disable image caching in setup-qemu action (#707)
- [5545b2dc](https://github.com/kubedb/ops-manager/commit/5545b2dc6) Add Recommendation Improvement and Bug Fixes (#704)
- [81a08203](https://github.com/kubedb/ops-manager/commit/81a08203b) Allow Postgres Failover While Ops Request Progressing (#703)
- [f2d98a46](https://github.com/kubedb/ops-manager/commit/f2d98a46d) add reconfigure-tls to PgBouncer (#701)
- [b7300c6e](https://github.com/kubedb/ops-manager/commit/b7300c6e2) Add Arbiter Node Support, update restart for SQL Server (#649)
- [b513e9e1](https://github.com/kubedb/ops-manager/commit/b513e9e1d) Fix RabbitMQ Version Update (#702)
- [0abc0816](https://github.com/kubedb/ops-manager/commit/0abc08163) Bug Fix in Reconfiguration(pgpool) (#698)
- [1fd029ff](https://github.com/kubedb/ops-manager/commit/1fd029ff3) add pgpool rotate auth (#697)
- [6e732fc7](https://github.com/kubedb/ops-manager/commit/6e732fc7c) Add Rotate auth for pgbouncer (#693)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.39.0-rc.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.39.0-rc.0)

- [add342c6](https://github.com/kubedb/percona-xtradb/commit/add342c62) Prepare for release v0.39.0-rc.0 (#394)
- [bbe09606](https://github.com/kubedb/percona-xtradb/commit/bbe09606f) Fix AuthSecret Check (#386)
- [4f888fbf](https://github.com/kubedb/percona-xtradb/commit/4f888fbfe) Disable image caching in setup-qemu action (#393)
- [ebe98bef](https://github.com/kubedb/percona-xtradb/commit/ebe98bef2) Add client billing event support (#392)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.25.0-rc.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.25.0-rc.0)

- [3e09a9ad](https://github.com/kubedb/percona-xtradb-coordinator/commit/3e09a9ad) Prepare for release v0.25.0-rc.0 (#89)
- [4013787c](https://github.com/kubedb/percona-xtradb-coordinator/commit/4013787c) Disable image caching in setup-qemu action (#90)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.36.0-rc.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.36.0-rc.0)

- [0bc3ed8d](https://github.com/kubedb/pg-coordinator/commit/0bc3ed8d) Prepare for release v0.36.0-rc.0 (#187)
- [e5aabf68](https://github.com/kubedb/pg-coordinator/commit/e5aabf68) Disable image caching in setup-qemu action (#188)
- [a286040e](https://github.com/kubedb/pg-coordinator/commit/a286040e) Allow failover with data loss (#186)
- [8c0d12da](https://github.com/kubedb/pg-coordinator/commit/8c0d12da) Enhance Archive WAL Management and Improve Failover Stability (#184)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.39.0-rc.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.39.0-rc.0)

- [d2893e64](https://github.com/kubedb/pgbouncer/commit/d2893e64) Prepare for release v0.39.0-rc.0 (#359)
- [6b985c2c](https://github.com/kubedb/pgbouncer/commit/6b985c2c) Disable image caching in setup-qemu action (#358)
- [eb7d008e](https://github.com/kubedb/pgbouncer/commit/eb7d008e) Enable license restriction and client billing (#357)
- [2b4ee8f5](https://github.com/kubedb/pgbouncer/commit/2b4ee8f5) reconfigure-tls changes and cert remove (#356)
- [4cf2f0bc](https://github.com/kubedb/pgbouncer/commit/4cf2f0bc) Update AuthActiveAnnotation const in secret (#352)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.7.0-rc.0](https://github.com/kubedb/pgpool/releases/tag/v0.7.0-rc.0)

- [4ed69176](https://github.com/kubedb/pgpool/commit/4ed69176) Prepare for release v0.7.0-rc.0 (#60)
- [a4735c17](https://github.com/kubedb/pgpool/commit/a4735c17) Disable image caching in setup-qemu action (#59)
- [c1bf6d9a](https://github.com/kubedb/pgpool/commit/c1bf6d9a) Enable license restriction and client billing (#58)
- [0c326d14](https://github.com/kubedb/pgpool/commit/0c326d14) Add basic-auth-active-from annotation in auth secret & Config secret separation  & PID file location added (#56)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.52.0-rc.0](https://github.com/kubedb/postgres/releases/tag/v0.52.0-rc.0)

- [18f17d97](https://github.com/kubedb/postgres/commit/18f17d974) Prepare for release v0.52.0-rc.0 (#790)
- [db644a80](https://github.com/kubedb/postgres/commit/db644a803) Disable image caching in setup-qemu action (#789)
- [6ad420f9](https://github.com/kubedb/postgres/commit/6ad420f93) Add client billing event support (#783)
- [e056f480](https://github.com/kubedb/postgres/commit/e056f4800) Added archiver and init-script manifest restore support (#787)
- [8ed11d08](https://github.com/kubedb/postgres/commit/8ed11d084) Archive mode related changes and add startTime-endTime in incremental snapshot (#784)
- [fe0e2ac0](https://github.com/kubedb/postgres/commit/fe0e2ac0e) Parse license restrictions
- [eaf5b26c](https://github.com/kubedb/postgres/commit/eaf5b26c0) Add license restrictions (#782)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.13.0-rc.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.13.0-rc.0)

- [b7bee766](https://github.com/kubedb/postgres-archiver/commit/b7bee766) Prepare for release v0.13.0-rc.0 (#55)
- [7b34eeba](https://github.com/kubedb/postgres-archiver/commit/7b34eeba) Disable image caching in setup-qemu action (#54)
- [47ea5469](https://github.com/kubedb/postgres-archiver/commit/47ea5469) Update Logstats, update wal archive support for standby (#52)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.13.0-rc.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.13.0-rc.0)

- [bbe21c0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/bbe21c0) Prepare for release v0.13.0-rc.0 (#51)
- [f088e55](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/f088e55) Disable image caching in setup-qemu action (#50)
- [3db1026](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/3db1026) Update volumeSnapshotTime when volumeSnapshot is ready to use and add support for standalone volume snapshot (#43)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.15.0-rc.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.15.0-rc.0)

- [91d569b](https://github.com/kubedb/postgres-restic-plugin/commit/91d569b) Prepare for release v0.15.0-rc.0 (#65)
- [87c8e45](https://github.com/kubedb/postgres-restic-plugin/commit/87c8e45) Disable image caching in setup-qemu action (#64)
- [b287099](https://github.com/kubedb/postgres-restic-plugin/commit/b287099) Add Stdin Backup Leaf Command support (#62)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.14.0-rc.0](https://github.com/kubedb/provider-aws/releases/tag/v0.14.0-rc.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.14.0-rc.0](https://github.com/kubedb/provider-azure/releases/tag/v0.14.0-rc.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.14.0-rc.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.14.0-rc.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.52.0-rc.0](https://github.com/kubedb/provisioner/releases/tag/v0.52.0-rc.0)




## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.39.0-rc.0](https://github.com/kubedb/proxysql/releases/tag/v0.39.0-rc.0)

- [15e87ca7](https://github.com/kubedb/proxysql/commit/15e87ca70) Prepare for release v0.39.0-rc.0 (#375)
- [ade37da5](https://github.com/kubedb/proxysql/commit/ade37da53) Disable image caching in setup-qemu action (#374)
- [22a4b65a](https://github.com/kubedb/proxysql/commit/22a4b65af) Add client billing event support (#372)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.7.0-rc.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.7.0-rc.0)

- [7e6e1cca](https://github.com/kubedb/rabbitmq/commit/7e6e1cca) Fix lister creation for client billing
- [71d4a65d](https://github.com/kubedb/rabbitmq/commit/71d4a65d) Prepare for release v0.7.0-rc.0 (#66)
- [4dad3584](https://github.com/kubedb/rabbitmq/commit/4dad3584) Disable image caching in setup-qemu action (#67)
- [5ae024d8](https://github.com/kubedb/rabbitmq/commit/5ae024d8) Add client billing event support (#64)
- [a179e524](https://github.com/kubedb/rabbitmq/commit/a179e524) Fix RabbitMQ Deletion Issue (#63)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.45.0-rc.0](https://github.com/kubedb/redis/releases/tag/v0.45.0-rc.0)

- [57992fa9](https://github.com/kubedb/redis/commit/57992fa92) Prepare for release v0.45.0-rc.0 (#575)
- [01332fde](https://github.com/kubedb/redis/commit/01332fdeb) Disable image caching in setup-qemu action (#576)
- [67e0e092](https://github.com/kubedb/redis/commit/67e0e0923) Enable license restriction and client billing (#574)
- [d9140fd6](https://github.com/kubedb/redis/commit/d9140fd65) Enable license restriction and client billing (#572)
- [1e8f2763](https://github.com/kubedb/redis/commit/1e8f27634) Move mutator defaults to SetDefaults() (#571)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.31.0-rc.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.31.0-rc.0)

- [e56c4c25](https://github.com/kubedb/redis-coordinator/commit/e56c4c25) Prepare for release v0.31.0-rc.0 (#121)
- [2539e4b4](https://github.com/kubedb/redis-coordinator/commit/2539e4b4) Disable image caching in setup-qemu action (#120)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.15.0-rc.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.15.0-rc.0)

- [cd65bb3](https://github.com/kubedb/redis-restic-plugin/commit/cd65bb3) Prepare for release v0.15.0-rc.0 (#60)
- [a1e6dfa](https://github.com/kubedb/redis-restic-plugin/commit/a1e6dfa) Disable image caching in setup-qemu action (#59)
- [7cacd68](https://github.com/kubedb/redis-restic-plugin/commit/7cacd68) Add Stdin Backup Leaf Command support (#57)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.39.0-rc.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.39.0-rc.0)

- [43e19c32](https://github.com/kubedb/replication-mode-detector/commit/43e19c32) Prepare for release v0.39.0-rc.0 (#286)
- [fce9f58c](https://github.com/kubedb/replication-mode-detector/commit/fce9f58c) Disable image caching in setup-qemu action (#287)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.28.0-rc.0](https://github.com/kubedb/schema-manager/releases/tag/v0.28.0-rc.0)

- [c5649c28](https://github.com/kubedb/schema-manager/commit/c5649c28) Prepare for release v0.28.0-rc.0 (#131)
- [4f52fd34](https://github.com/kubedb/schema-manager/commit/4f52fd34) Fix nats initialization
- [49e83a00](https://github.com/kubedb/schema-manager/commit/49e83a00) Disable image caching in setup-qemu action (#130)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.7.0-rc.0](https://github.com/kubedb/singlestore/releases/tag/v0.7.0-rc.0)

- [c22e7c86](https://github.com/kubedb/singlestore/commit/c22e7c86) Prepare for release v0.7.0-rc.0 (#58)
- [ebb993be](https://github.com/kubedb/singlestore/commit/ebb993be) Disable image caching in setup-qemu action (#57)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.7.0-rc.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.7.0-rc.0)

- [3ab10f5](https://github.com/kubedb/singlestore-coordinator/commit/3ab10f5) Prepare for release v0.7.0-rc.0 (#36)
- [f7271dd](https://github.com/kubedb/singlestore-coordinator/commit/f7271dd) Disable image caching in setup-qemu action (#35)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.10.0-rc.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.10.0-rc.0)

- [1b36229](https://github.com/kubedb/singlestore-restic-plugin/commit/1b36229) Prepare for release v0.10.0-rc.0 (#36)
- [4c9a228](https://github.com/kubedb/singlestore-restic-plugin/commit/4c9a228) Disable image caching in setup-qemu action (#35)
- [98dfeae](https://github.com/kubedb/singlestore-restic-plugin/commit/98dfeae) Add Stdin Backup Leaf Command support (#33)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.7.0-rc.0](https://github.com/kubedb/solr/releases/tag/v0.7.0-rc.0)

- [700ce01b](https://github.com/kubedb/solr/commit/700ce01b) Prepare for release v0.7.0-rc.0 (#67)
- [9216caea](https://github.com/kubedb/solr/commit/9216caea) Fix lister creation for client billing (#66)
- [43fdf9b2](https://github.com/kubedb/solr/commit/43fdf9b2) Disable image caching in setup-qemu action (#65)
- [fe9b60d8](https://github.com/kubedb/solr/commit/fe9b60d8) Add client billing event support (#64)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.37.0-rc.0](https://github.com/kubedb/tests/releases/tag/v0.37.0-rc.0)

- [34cb745f](https://github.com/kubedb/tests/commit/34cb745f) Prepare for release v0.37.0-rc.0 (#437)
- [751cbfb3](https://github.com/kubedb/tests/commit/751cbfb3) Disable image caching in setup-qemu action (#436)
- [579aa5f0](https://github.com/kubedb/tests/commit/579aa5f0) Add Postgres Archiver Backup-Restore (#427)
- [d0da32f9](https://github.com/kubedb/tests/commit/d0da32f9) Fix TLS enabled secret part for Postgres (#432)
- [6aa80b8d](https://github.com/kubedb/tests/commit/6aa80b8d) MongoDB EndTime fix (Disaster Time Scenario) (#429)
- [9c288453](https://github.com/kubedb/tests/commit/9c288453) Add Solr OpsRequest Tests (#428)
- [3729dd68](https://github.com/kubedb/tests/commit/3729dd68) MariaDB version 11.6.2 check (all) (#431)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.28.0-rc.0](https://github.com/kubedb/ui-server/releases/tag/v0.28.0-rc.0)

- [b0032db9](https://github.com/kubedb/ui-server/commit/b0032db9) Prepare for release v0.28.0-rc.0 (#150)
- [61eff458](https://github.com/kubedb/ui-server/commit/61eff458) Disable image caching in setup-qemu action (#149)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.28.0-rc.0](https://github.com/kubedb/webhook-server/releases/tag/v0.28.0-rc.0)

- [b3b3c562](https://github.com/kubedb/webhook-server/commit/b3b3c562) Prepare for release v0.28.0-rc.0 (#143)
- [f5cfed8f](https://github.com/kubedb/webhook-server/commit/f5cfed8f) Disable image caching in setup-qemu action (#142)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.7.0-rc.0](https://github.com/kubedb/zookeeper/releases/tag/v0.7.0-rc.0)

- [ace3a6dd](https://github.com/kubedb/zookeeper/commit/ace3a6dd) Prepare for release v0.7.0-rc.0 (#57)
- [89699f8d](https://github.com/kubedb/zookeeper/commit/89699f8d) Disable image caching in setup-qemu action (#56)
- [9a81250c](https://github.com/kubedb/zookeeper/commit/9a81250c) Add client billing event support (#55)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.8.0-rc.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.8.0-rc.0)

- [52085d1](https://github.com/kubedb/zookeeper-restic-plugin/commit/52085d1) Prepare for release v0.8.0-rc.0 (#27)
- [7323969](https://github.com/kubedb/zookeeper-restic-plugin/commit/7323969) Disable image caching in setup-qemu action (#28)
- [096f163](https://github.com/kubedb/zookeeper-restic-plugin/commit/096f163) Add Stdin Backup Leaf Command support (#25)




