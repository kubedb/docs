---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2025.12.9-rc.0
    name: Changelog-v2025.12.9-rc.0
    parent: welcome
    weight: 20251209
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2025.12.9-rc.0/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2025.12.9-rc.0/
---

# KubeDB v2025.12.9-rc.0 (2025-12-12)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.60.0-rc.0](https://github.com/kubedb/apimachinery/releases/tag/v0.60.0-rc.0)

- [bd999704](https://github.com/kubedb/apimachinery/commit/bd9997041) Update deps
- [fac813c7](https://github.com/kubedb/apimachinery/commit/fac813c78) Update deps
- [e58fe916](https://github.com/kubedb/apimachinery/commit/e58fe916a) Update license libraries
- [0562e8d1](https://github.com/kubedb/apimachinery/commit/0562e8d15) Use golangci-lint 2.x (#1549)
- [3d10db60](https://github.com/kubedb/apimachinery/commit/3d10db600) Implement upgradability helpers (#1536)
- [9131dc31](https://github.com/kubedb/apimachinery/commit/9131dc317) Add milvus apis (#1533)
- [7d56b3e6](https://github.com/kubedb/apimachinery/commit/7d56b3e6c) Virtual secret for pgbouncer pgpool redis (#1516)
- [e1143b59](https://github.com/kubedb/apimachinery/commit/e1143b591) Add MaxRetries api (#1547)
- [b8b8a903](https://github.com/kubedb/apimachinery/commit/b8b8a9030) Fix webhook: postgres volume exp ops, memcached auth secret spec (#1548)
- [56ba059d](https://github.com/kubedb/apimachinery/commit/56ba059d8) Add oracle tls api (#1545)
- [0c8e327b](https://github.com/kubedb/apimachinery/commit/0c8e327b1) Allow Force Failover with data loss | add tuning api (#1542)
- [e34d1088](https://github.com/kubedb/apimachinery/commit/e34d1088a) Add Weaviate APIs (#1528)
- [8059bb8b](https://github.com/kubedb/apimachinery/commit/8059bb8b7) Add DB2 APIs (#1530)
- [285acd8b](https://github.com/kubedb/apimachinery/commit/285acd8bd) Add Neo4j API (#1544)
- [7e98c3f0](https://github.com/kubedb/apimachinery/commit/7e98c3f00) Update skipper and webhook for config merger (#1535)
- [e3c8ce90](https://github.com/kubedb/apimachinery/commit/e3c8ce90f) Add Hanadb (#1527)
- [f56bbc3a](https://github.com/kubedb/apimachinery/commit/f56bbc3a4) Fix MySQL reconfigure apply config (#1534)
- [2ed241d2](https://github.com/kubedb/apimachinery/commit/2ed241d26) Add hazelcast binding implementation (#1540)
- [2bbafe54](https://github.com/kubedb/apimachinery/commit/2bbafe54f) Add Qdrant (#1524)
- [4de23d6d](https://github.com/kubedb/apimachinery/commit/4de23d6d3) Update vulnerable deps (#1541)
- [832e8ff4](https://github.com/kubedb/apimachinery/commit/832e8ff48) Move lib pkg from ops-manager (#1532)
- [bf32e8fc](https://github.com/kubedb/apimachinery/commit/bf32e8fca) Remove redundant print column `Type` in DB CRs.  (#1531)
- [bbc01408](https://github.com/kubedb/apimachinery/commit/bbc014087) make gen fmt (#1529)
- [9ae5deaa](https://github.com/kubedb/apimachinery/commit/9ae5deaaa) Update deps



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.45.0-rc.0](https://github.com/kubedb/autoscaler/releases/tag/v0.45.0-rc.0)

- [7765b291](https://github.com/kubedb/autoscaler/commit/7765b291) Prepare for release v0.45.0-rc.0 (#268)
- [9a7f1096](https://github.com/kubedb/autoscaler/commit/9a7f1096) Skip controller activation in certification mode (#267)
- [d8119820](https://github.com/kubedb/autoscaler/commit/d8119820) Use golangci-lint 2.x (#266)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.13.0-rc.0](https://github.com/kubedb/cassandra/releases/tag/v0.13.0-rc.0)

- [78213d47](https://github.com/kubedb/cassandra/commit/78213d47) Prepare for release v0.13.0-rc.0 (#53)
- [8f8458f1](https://github.com/kubedb/cassandra/commit/8f8458f1) Fix multiple restart issue by introducing parallelismController (#52)
- [5563cf24](https://github.com/kubedb/cassandra/commit/5563cf24) Moved ops code to dev repo (#51)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.7.0-rc.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.7.0-rc.0)

- [dd6cbb1](https://github.com/kubedb/cassandra-medusa-plugin/commit/dd6cbb1) Prepare for release v0.7.0-rc.0 (#15)
- [b2132d1](https://github.com/kubedb/cassandra-medusa-plugin/commit/b2132d1) update deps (#14)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.60.0-rc.0](https://github.com/kubedb/cli/releases/tag/v0.60.0-rc.0)

- [fe2da356](https://github.com/kubedb/cli/commit/fe2da3569) Prepare for release v0.60.0-rc.0 (#807)
- [3fc4051c](https://github.com/kubedb/cli/commit/3fc4051ca) Add gitops CLI (#804)
- [8749544b](https://github.com/kubedb/cli/commit/8749544bc) Update Lint (#806)
- [41a390a6](https://github.com/kubedb/cli/commit/41a390a60) Make the debug cli common (#805)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.15.0-rc.0](https://github.com/kubedb/clickhouse/releases/tag/v0.15.0-rc.0)

- [49c77055](https://github.com/kubedb/clickhouse/commit/49c77055) Prepare for release v0.15.0-rc.0 (#74)
- [876fb0be](https://github.com/kubedb/clickhouse/commit/876fb0be) Fix go version (#73)
- [2bcf4461](https://github.com/kubedb/clickhouse/commit/2bcf4461) Fix multiple restart issue by introducing parallelismController (#72)
- [d3ebb6df](https://github.com/kubedb/clickhouse/commit/d3ebb6df) Move in Ops-manager code to ClickHouse (#71)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.15.0-rc.0](https://github.com/kubedb/crd-manager/releases/tag/v0.15.0-rc.0)

- [d4dd9b8c](https://github.com/kubedb/crd-manager/commit/d4dd9b8c) Prepare for release v0.15.0-rc.0 (#101)
- [8347e307](https://github.com/kubedb/crd-manager/commit/8347e307) replace go1.25.5 to go1.25 (#100)
- [34895de8](https://github.com/kubedb/crd-manager/commit/34895de8) Add neo4j
- [c89068f3](https://github.com/kubedb/crd-manager/commit/c89068f3) HanaDB (#90)
- [1245ee00](https://github.com/kubedb/crd-manager/commit/1245ee00) Add milvus crd-manager (#94)
- [da01d86e](https://github.com/kubedb/crd-manager/commit/da01d86e) Add Weaviate Crds (#97)
- [521fbd57](https://github.com/kubedb/crd-manager/commit/521fbd57) Add Db2 Crds (#96)
- [1cf9a782](https://github.com/kubedb/crd-manager/commit/1cf9a782) Add Neo4j Crds (#98)
- [80e2bc36](https://github.com/kubedb/crd-manager/commit/80e2bc36) Add Qdrant CRD (#92)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.18.0-rc.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.18.0-rc.0)

- [c970a47](https://github.com/kubedb/dashboard-restic-plugin/commit/c970a47) Prepare for release v0.18.0-rc.0 (#51)
- [b42b5f6](https://github.com/kubedb/dashboard-restic-plugin/commit/b42b5f6) change to 1.25 (#50)
- [c1811e1](https://github.com/kubedb/dashboard-restic-plugin/commit/c1811e1) Update Dependency (#49)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.15.0-rc.0](https://github.com/kubedb/db-client-go/releases/tag/v0.15.0-rc.0)

- [238d08e2](https://github.com/kubedb/db-client-go/commit/238d08e2) Prepare for release v0.15.0-rc.0 (#209)
- [3789532d](https://github.com/kubedb/db-client-go/commit/3789532d) Fix Go version (#208)
- [eaca6fc1](https://github.com/kubedb/db-client-go/commit/eaca6fc1) Add Weaviate Client Go (#202)
- [1826245f](https://github.com/kubedb/db-client-go/commit/1826245f) Add hanadb client-go (#201)
- [fbf24826](https://github.com/kubedb/db-client-go/commit/fbf24826) Add milvus db-client-go (#198)
- [2e876291](https://github.com/kubedb/db-client-go/commit/2e876291) Update go.mod file
- [e3d3ecca](https://github.com/kubedb/db-client-go/commit/e3d3ecca) Use golangci-lint 2.x (#207)
- [5beb8789](https://github.com/kubedb/db-client-go/commit/5beb8789) Add Oracle tls support (#205)
- [0daf0012](https://github.com/kubedb/db-client-go/commit/0daf0012) Add Neo4j Client (#204)
- [8f852044](https://github.com/kubedb/db-client-go/commit/8f852044) Add db2 client go (#206)
- [0b613041](https://github.com/kubedb/db-client-go/commit/0b613041) Add Qdrant (#197)
- [f77e6314](https://github.com/kubedb/db-client-go/commit/f77e6314) Add Virtual Secret Pgbouncer Pgpool Redis Valkey (#199)
- [b824c01d](https://github.com/kubedb/db-client-go/commit/b824c01d) move go_es file to db-client (#200)



## [kubedb/db2](https://github.com/kubedb/db2)

### [v0.1.0-rc.0](https://github.com/kubedb/db2/releases/tag/v0.1.0-rc.0)

- [c61ee118](https://github.com/kubedb/db2/commit/c61ee118) Prepare for release v0.1.0-rc.0 (#3)
- [e5dd65e7](https://github.com/kubedb/db2/commit/e5dd65e7) Implement db2 controller
- [cbe78ad6](https://github.com/kubedb/db2/commit/cbe78ad6) Add vendor



## [kubedb/db2-coordinator](https://github.com/kubedb/db2-coordinator)

### [v0.1.0-rc.0](https://github.com/kubedb/db2-coordinator/releases/tag/v0.1.0-rc.0)

- [e44da80](https://github.com/kubedb/db2-coordinator/commit/e44da80) Use golangci-lint 2.x (#1)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.15.0-rc.0](https://github.com/kubedb/druid/releases/tag/v0.15.0-rc.0)

- [d1e1b108](https://github.com/kubedb/druid/commit/d1e1b108) Prepare for release v0.15.0-rc.0 (#105)
- [5632cb2f](https://github.com/kubedb/druid/commit/5632cb2f) Fix multiple restart issue by introducing parallelismController (#104)
- [5f237f3c](https://github.com/kubedb/druid/commit/5f237f3c) fixed some mistake (#103)
- [23ad9f70](https://github.com/kubedb/druid/commit/23ad9f70) move ops code to db repo (#102)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.60.0-rc.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.60.0-rc.0)

- [f75c70d0](https://github.com/kubedb/elasticsearch/commit/f75c70d09) Prepare for release v0.60.0-rc.0 (#780)
- [da290110](https://github.com/kubedb/elasticsearch/commit/da2901102) move go_es file to db-client-go (#779)
- [1de4f810](https://github.com/kubedb/elasticsearch/commit/1de4f8105) move ops code to db repo (#778)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.23.0-rc.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.23.0-rc.0)

- [da511bd8](https://github.com/kubedb/elasticsearch-restic-plugin/commit/da511bd8) Prepare for release v0.23.0-rc.0 (#75)
- [f8bc69f1](https://github.com/kubedb/elasticsearch-restic-plugin/commit/f8bc69f1) change to 1.25 (#74)
- [e5a1b21d](https://github.com/kubedb/elasticsearch-restic-plugin/commit/e5a1b21d) update-deps (#73)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.15.0-rc.0](https://github.com/kubedb/ferretdb/releases/tag/v0.15.0-rc.0)

- [7b572633](https://github.com/kubedb/ferretdb/commit/7b572633) Prepare for release v0.15.0-rc.0 (#93)
- [b05a9345](https://github.com/kubedb/ferretdb/commit/b05a9345) Fix go version (#92)
- [e5d1ee2d](https://github.com/kubedb/ferretdb/commit/e5d1ee2d) Fix multiple restart issue by introducing parallelismController (#91)
- [719326aa](https://github.com/kubedb/ferretdb/commit/719326aa) Move in FerretDb Ops Manager Code (#90)
- [6662b891](https://github.com/kubedb/ferretdb/commit/6662b891) Test against k8s 1.34 (#89)
- [3c73ec67](https://github.com/kubedb/ferretdb/commit/3c73ec67) No need to check backend's owner ref



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.8.0-rc.0](https://github.com/kubedb/gitops/releases/tag/v0.8.0-rc.0)

- [2e506be8](https://github.com/kubedb/gitops/commit/2e506be8) Prepare for release v0.8.0-rc.0 (#31)
- [9de37946](https://github.com/kubedb/gitops/commit/9de37946) Fix vertical scaling ops creation for storage resouce changes (#30)



## [kubedb/hanadb](https://github.com/kubedb/hanadb)

### [v0.1.0-rc.0](https://github.com/kubedb/hanadb/releases/tag/v0.1.0-rc.0)

- [3fc76f7e](https://github.com/kubedb/hanadb/commit/3fc76f7e) Prepare for release v0.1.0-rc.0 (#5)
- [67f88948](https://github.com/kubedb/hanadb/commit/67f88948) Modify go.mod (#4)
- [c440319c](https://github.com/kubedb/hanadb/commit/c440319c) Modify Go version (#3)
- [aa961389](https://github.com/kubedb/hanadb/commit/aa961389) Test against k8s 1.34 (#2)
- [a5a5049f](https://github.com/kubedb/hanadb/commit/a5a5049f) Add HanaDB provisioner (#1)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.6.0-rc.0](https://github.com/kubedb/hazelcast/releases/tag/v0.6.0-rc.0)

- [37277ce1](https://github.com/kubedb/hazelcast/commit/37277ce1) Prepare for release v0.6.0-rc.0 (#18)
- [dbfe579c](https://github.com/kubedb/hazelcast/commit/dbfe579c) Fix multiple restart issue by introducing parallelismController (#17)
- [da33d710](https://github.com/kubedb/hazelcast/commit/da33d710) move ops code to db repo (#16)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.7.0-rc.0](https://github.com/kubedb/ignite/releases/tag/v0.7.0-rc.0)

- [35277e58](https://github.com/kubedb/ignite/commit/35277e58) Prepare for release v0.7.0-rc.0 (#27)
- [c642d15e](https://github.com/kubedb/ignite/commit/c642d15e) Fix multiple restart issue by introducing parallelismController (#26)
- [371c83df](https://github.com/kubedb/ignite/commit/371c83df) Move ops to db repo (#24)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2025.12.9-rc.0](https://github.com/kubedb/installer/releases/tag/v2025.12.9-rc.0)

- [679a5fec](https://github.com/kubedb/installer/commit/679a5feca) Prepare for release v2025.12.9-rc.0 (#1963)
- [66929a1b](https://github.com/kubedb/installer/commit/66929a1b5) Update es init tags: update feature-gates list (#1962)
- [7abea93d](https://github.com/kubedb/installer/commit/7abea93d0) Review the webhookconfiguration files (#1959)
- [b423856b](https://github.com/kubedb/installer/commit/b423856b0) Prepare chart for redhat certification (#1955)
- [1ad0e641](https://github.com/kubedb/installer/commit/1ad0e6414) Add Neo4jVersion (#1932)
- [b8a31dab](https://github.com/kubedb/installer/commit/b8a31dab5) Update postgres init image tag (#1958)
- [16fd66b2](https://github.com/kubedb/installer/commit/16fd66b26) Add oracle tls (#1956)
- [d824a4d5](https://github.com/kubedb/installer/commit/d824a4d55) HanaDB (#1881)
- [3f149961](https://github.com/kubedb/installer/commit/3f1499618) Add Qdrant (#1890)
- [f7d128e7](https://github.com/kubedb/installer/commit/f7d128e7d) Add milvus installer (#1900)
- [1123c189](https://github.com/kubedb/installer/commit/1123c1892) Add Db2 Version (#1904)
- [29dcbe33](https://github.com/kubedb/installer/commit/29dcbe33f) Update cve report 2025-12-07 (#1954)
- [793d82b5](https://github.com/kubedb/installer/commit/793d82b5a) Add mariadb physical backup image (#1939)
- [817c44ca](https://github.com/kubedb/installer/commit/817c44ca6) Add Weaviate Catalogs (#1911)
- [0365dbdd](https://github.com/kubedb/installer/commit/0365dbdda) Update Go version and Redis Init image Tag update (#1957)
- [bf9df2ed](https://github.com/kubedb/installer/commit/bf9df2edc) Add ubi mode (#1927)
- [01c32e1e](https://github.com/kubedb/installer/commit/01c32e1e9) Update cve report (#1952)
- [6c6a3b84](https://github.com/kubedb/installer/commit/6c6a3b84e) Use golangci-lint 2.x (#1953)
- [4a195ef3](https://github.com/kubedb/installer/commit/4a195ef33) Update crds for kubedb/apimachinery@9131dc31 (#1950)
- [61aa39a6](https://github.com/kubedb/installer/commit/61aa39a6a) Update crds for kubedb/apimachinery@e1143b59 (#1948)
- [59565868](https://github.com/kubedb/installer/commit/595658685) Update crds for kubedb/apimachinery@0c8e327b (#1945)
- [25ca5d6f](https://github.com/kubedb/installer/commit/25ca5d6f4) Update crds for kubedb/apimachinery@e34d1088 (#1944)
- [18e7b523](https://github.com/kubedb/installer/commit/18e7b523c) Update cve report (#1943)
- [2c113fc7](https://github.com/kubedb/installer/commit/2c113fc77) Update crds for kubedb/apimachinery@7e98c3f0 (#1941)
- [bbdb57ea](https://github.com/kubedb/installer/commit/bbdb57ead) Update cve report (#1938)
- [455f0788](https://github.com/kubedb/installer/commit/455f07880) Update cve report (#1937)
- [7923b3fa](https://github.com/kubedb/installer/commit/7923b3fa3) Update cve report (#1936)
- [4c96365b](https://github.com/kubedb/installer/commit/4c96365b8) Update cve report (#1935)
- [ea790c88](https://github.com/kubedb/installer/commit/ea790c887) Update cve report (#1934)
- [fcb1ea18](https://github.com/kubedb/installer/commit/fcb1ea183) Update cve report (#1933)
- [0c5c749a](https://github.com/kubedb/installer/commit/0c5c749ac) Update cve report (#1931)
- [1ac66018](https://github.com/kubedb/installer/commit/1ac660181) Update cve report (#1930)
- [ffed437e](https://github.com/kubedb/installer/commit/ffed437e2) Update cve report (#1929)
- [0c4e8c21](https://github.com/kubedb/installer/commit/0c4e8c219) Update cve report (#1928)
- [991a4702](https://github.com/kubedb/installer/commit/991a47022) Update crds for kubedb/apimachinery@2bbafe54 (#1924)
- [83567b9b](https://github.com/kubedb/installer/commit/83567b9be) Update cve report (#1923)
- [314786ea](https://github.com/kubedb/installer/commit/314786ea7) Update cve report (#1920)
- [2c8b0c11](https://github.com/kubedb/installer/commit/2c8b0c113) Update vulnerable deps (#1922)
- [7cda2f68](https://github.com/kubedb/installer/commit/7cda2f686) Update cve report (#1918)
- [0b99a9ae](https://github.com/kubedb/installer/commit/0b99a9aeb) Update cve report (#1917)
- [382818e4](https://github.com/kubedb/installer/commit/382818e48) Update cve report (#1916)
- [6fc3cd4f](https://github.com/kubedb/installer/commit/6fc3cd4f3) Update cve report (#1915)
- [4d3da89e](https://github.com/kubedb/installer/commit/4d3da89e1) Update cve report (#1914)
- [395bed20](https://github.com/kubedb/installer/commit/395bed209) Update cve report (#1913)
- [3552b77a](https://github.com/kubedb/installer/commit/3552b77ab) Update cve report (#1912)
- [e7d87c6d](https://github.com/kubedb/installer/commit/e7d87c6de) Update cve report (#1910)
- [650ad03e](https://github.com/kubedb/installer/commit/650ad03e1) Update cve report (#1909)
- [4c38a043](https://github.com/kubedb/installer/commit/4c38a0436) Update cve report (#1908)
- [001a720f](https://github.com/kubedb/installer/commit/001a720f7) Update cve report (#1907)
- [f95d048a](https://github.com/kubedb/installer/commit/f95d048a8) Update cve report (#1906)
- [e0c9772e](https://github.com/kubedb/installer/commit/e0c9772e9) Update cve report (#1905)
- [da490298](https://github.com/kubedb/installer/commit/da4902985) Update cve report (#1903)
- [17a30804](https://github.com/kubedb/installer/commit/17a308045) Update cve report (#1902)
- [003822e2](https://github.com/kubedb/installer/commit/003822e25) Update gitops operator container name (#1901)
- [d1e28390](https://github.com/kubedb/installer/commit/d1e28390f) Update cve report (#1899)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.31.0-rc.0](https://github.com/kubedb/kafka/releases/tag/v0.31.0-rc.0)

- [1003c37f](https://github.com/kubedb/kafka/commit/1003c37f) Prepare for release v0.31.0-rc.0 (#167)
- [e270d3e6](https://github.com/kubedb/kafka/commit/e270d3e6) Fix multiple restart issue by introducing parallelismController (#166)
- [3f7bf918](https://github.com/kubedb/kafka/commit/3f7bf918) Move ops-manager code to base repo (#165)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.36.0-rc.0](https://github.com/kubedb/kibana/releases/tag/v0.36.0-rc.0)

- [92ca7ab3](https://github.com/kubedb/kibana/commit/92ca7ab3) Prepare for release v0.36.0-rc.0 (#161)
- [529d7301](https://github.com/kubedb/kibana/commit/529d7301) update deps (#160)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.23.0-rc.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.23.0-rc.0)

- [1a10c291](https://github.com/kubedb/kubedb-manifest-plugin/commit/1a10c291) Prepare for release v0.23.0-rc.0 (#106)
- [75d723c8](https://github.com/kubedb/kubedb-manifest-plugin/commit/75d723c8) update-dep (#105)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.11.0-rc.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.11.0-rc.0)

- [616c0b5d](https://github.com/kubedb/kubedb-verifier/commit/616c0b5d) Prepare for release v0.11.0-rc.0 (#28)
- [78b37027](https://github.com/kubedb/kubedb-verifier/commit/78b37027) Use golangci-lint 2.x (#27)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.44.0-rc.0](https://github.com/kubedb/mariadb/releases/tag/v0.44.0-rc.0)

- [ea4d4a69](https://github.com/kubedb/mariadb/commit/ea4d4a69a) Prepare for release v0.44.0-rc.0 (#358)
- [d9257a6e](https://github.com/kubedb/mariadb/commit/d9257a6ea) Change Makefile GO Version (#357)
- [4f419cbe](https://github.com/kubedb/mariadb/commit/4f419cbec) Add golangci (#356)
- [96a12452](https://github.com/kubedb/mariadb/commit/96a12452c) Fix multiple restart issue by introducing parallelismController (#355)
- [c1fccdd7](https://github.com/kubedb/mariadb/commit/c1fccdd74) Fi Archiver Version Compatibility (#354)
- [b95f58e6](https://github.com/kubedb/mariadb/commit/b95f58e69) move ops code to db repo (#353)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.20.0-rc.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.20.0-rc.0)

- [5a4450ec](https://github.com/kubedb/mariadb-archiver/commit/5a4450ec) Prepare for release v0.20.0-rc.0 (#64)
- [51b288e1](https://github.com/kubedb/mariadb-archiver/commit/51b288e1) Change Makefile GO Version (#63)
- [12dd22cc](https://github.com/kubedb/mariadb-archiver/commit/12dd22cc) Add golangci (#62)
- [4364d82b](https://github.com/kubedb/mariadb-archiver/commit/4364d82b) Build and push ubi image (#61)
- [cd57feb0](https://github.com/kubedb/mariadb-archiver/commit/cd57feb0) Fix none mode slave not running issue (#60)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.40.0-rc.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.40.0-rc.0)

- [e007307f](https://github.com/kubedb/mariadb-coordinator/commit/e007307f) Prepare for release v0.40.0-rc.0 (#156)
- [67015617](https://github.com/kubedb/mariadb-coordinator/commit/67015617) Simplify code
- [8a015658](https://github.com/kubedb/mariadb-coordinator/commit/8a015658) format grpc package
- [810d7607](https://github.com/kubedb/mariadb-coordinator/commit/810d7607) Build ubi image (#155)
- [0b95dc57](https://github.com/kubedb/mariadb-coordinator/commit/0b95dc57) Use golangci-lint 2.x (#154)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.20.0-rc.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.20.0-rc.0)

- [615724cf](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/615724cf) Prepare for release v0.20.0-rc.0 (#59)
- [f44d64b1](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/f44d64b1) Change Makefile GO Version (#58)
- [38df4729](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/38df4729) Add golangci (#57)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.18.0-rc.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.18.0-rc.0)

- [3951117](https://github.com/kubedb/mariadb-restic-plugin/commit/3951117) Prepare for release v0.18.0-rc.0 (#62)
- [8ddff19](https://github.com/kubedb/mariadb-restic-plugin/commit/8ddff19) Fix  Host Missing, Restore Issue (#61)
- [ddb84e4](https://github.com/kubedb/mariadb-restic-plugin/commit/ddb84e4) Change Makefile GO Version (#60)
- [8db3131](https://github.com/kubedb/mariadb-restic-plugin/commit/8db3131) Add golangci (#59)
- [47f4fd3](https://github.com/kubedb/mariadb-restic-plugin/commit/47f4fd3) push images for physical backup (#58)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.53.0-rc.0](https://github.com/kubedb/memcached/releases/tag/v0.53.0-rc.0)

- [dc1dc137](https://github.com/kubedb/memcached/commit/dc1dc137f) Prepare for release v0.53.0-rc.0 (#514)
- [aa8446d9](https://github.com/kubedb/memcached/commit/aa8446d99) Fix multiple restart issue by introducing parallelismController (#513)
- [1e26cc41](https://github.com/kubedb/memcached/commit/1e26cc416) Move ops to db repo (#512)



## [kubedb/milvus](https://github.com/kubedb/milvus)

### [v0.1.0-rc.0](https://github.com/kubedb/milvus/releases/tag/v0.1.0-rc.0)

- [cc2ca53d](https://github.com/kubedb/milvus/commit/cc2ca53d) Prepare for release v0.1.0-rc.0 (#6)
- [796f5386](https://github.com/kubedb/milvus/commit/796f5386) Downgrade k8s version (#5)
- [d1cf5ee0](https://github.com/kubedb/milvus/commit/d1cf5ee0) goversion modified (#3)
- [9317ea0f](https://github.com/kubedb/milvus/commit/9317ea0f) Test against k8s 1.34 (#2)
- [88031d6a](https://github.com/kubedb/milvus/commit/88031d6a) Add milvus provisioner (standalone) (#1)
- [9f7923b8](https://github.com/kubedb/milvus/commit/9f7923b8) Update .gitignore



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.53.0-rc.0](https://github.com/kubedb/mongodb/releases/tag/v0.53.0-rc.0)

- [f73a7159](https://github.com/kubedb/mongodb/commit/f73a7159f) Prepare for release v0.53.0-rc.0 (#724)
- [4a27317f](https://github.com/kubedb/mongodb/commit/4a27317f1) Update golang version
- [3b46da51](https://github.com/kubedb/mongodb/commit/3b46da516) Fix linter (#723)
- [75045740](https://github.com/kubedb/mongodb/commit/75045740a) Fix multiple restart issue by introducing parallelismController (#722)
- [d42b514f](https://github.com/kubedb/mongodb/commit/d42b514f8) Move ops code to db repo (#721)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.21.0-rc.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.21.0-rc.0)

- [75b334b9](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/75b334b9) Prepare for release v0.21.0-rc.0 (#62)
- [2e658828](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/2e658828) Fix linter (#61)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.23.0-rc.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.23.0-rc.0)

- [ecc63fec](https://github.com/kubedb/mongodb-restic-plugin/commit/ecc63fec) Prepare for release v0.23.0-rc.0 (#95)
- [821c386c](https://github.com/kubedb/mongodb-restic-plugin/commit/821c386c) Fix makefile for ubi images (#94)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.15.0-rc.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.15.0-rc.0)

- [e859024e](https://github.com/kubedb/mssql-coordinator/commit/e859024e) Prepare for release v0.15.0-rc.0 (#47)
- [1b3d5a4d](https://github.com/kubedb/mssql-coordinator/commit/1b3d5a4d) Fix container build command
- [a3e461c3](https://github.com/kubedb/mssql-coordinator/commit/a3e461c3) Build ubi image (#46)
- [9268580a](https://github.com/kubedb/mssql-coordinator/commit/9268580a) Use golangci-lint 2.x (#45)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.15.0-rc.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.15.0-rc.0)

- [ac4cbed6](https://github.com/kubedb/mssqlserver/commit/ac4cbed6) Prepare for release v0.15.0-rc.0 (#98)
- [0c591eba](https://github.com/kubedb/mssqlserver/commit/0c591eba) Add ReconfigureOps merger (#96)
- [e3ad8e20](https://github.com/kubedb/mssqlserver/commit/e3ad8e20) Improve codebase by refactoring (#93)
- [01ab5c30](https://github.com/kubedb/mssqlserver/commit/01ab5c30) Fix multiple restart issue by introducing parallelismController (#97)
- [2f9976d3](https://github.com/kubedb/mssqlserver/commit/2f9976d3) Fix archiver issue for TLS secure Minio (#94)
- [b02aeeb3](https://github.com/kubedb/mssqlserver/commit/b02aeeb3) Move ops to DB repo (#95)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.14.0-rc.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.14.0-rc.0)

- [c65e2ae](https://github.com/kubedb/mssqlserver-archiver/commit/c65e2ae) Use golangci-lint 2.x (#15)



## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.14.0-rc.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.14.0-rc.0)

- [93cd25f](https://github.com/kubedb/mssqlserver-walg-plugin/commit/93cd25f) Prepare for release v0.14.0-rc.0 (#37)
- [677dc46](https://github.com/kubedb/mssqlserver-walg-plugin/commit/677dc46) Use golangci-lint 2.x (#36)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.53.0-rc.0](https://github.com/kubedb/mysql/releases/tag/v0.53.0-rc.0)

- [613bbbe1](https://github.com/kubedb/mysql/commit/613bbbe19) Prepare for release v0.53.0-rc.0 (#707)
- [6f1a6f0e](https://github.com/kubedb/mysql/commit/6f1a6f0ea) Change Makefile GO Version (#706)
- [1f037c8e](https://github.com/kubedb/mysql/commit/1f037c8e4) Use golangci-lint 2.x (#705)
- [6b649dec](https://github.com/kubedb/mysql/commit/6b649dec9) Fix multiple restart issue by introducing parallelismController (#704)
- [57998a04](https://github.com/kubedb/mysql/commit/57998a04b) move ops code to db repo (#703)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.21.0-rc.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.21.0-rc.0)

- [da732228](https://github.com/kubedb/mysql-archiver/commit/da732228) Prepare for release v0.21.0-rc.0 (#71)
- [e3a552e8](https://github.com/kubedb/mysql-archiver/commit/e3a552e8) Change Makefile GO Version (#70)
- [ad8911b6](https://github.com/kubedb/mysql-archiver/commit/ad8911b6) Add golangci (#69)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.38.0-rc.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.38.0-rc.0)

- [4f0b448e](https://github.com/kubedb/mysql-coordinator/commit/4f0b448e) Prepare for release v0.38.0-rc.0 (#154)
- [78c11d72](https://github.com/kubedb/mysql-coordinator/commit/78c11d72) Build ubi image (#153)
- [c4ae8e49](https://github.com/kubedb/mysql-coordinator/commit/c4ae8e49) Use golangci-lint 2.x (#152)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.21.0-rc.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.21.0-rc.0)

- [64893d8b](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/64893d8b) Prepare for release v0.21.0-rc.0 (#59)
- [48a23ce5](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/48a23ce5) Change Makefile GO Version (#58)
- [d1732152](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/d1732152) Add golangci (#57)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.23.0-rc.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.23.0-rc.0)

- [4552f5f7](https://github.com/kubedb/mysql-restic-plugin/commit/4552f5f7) Prepare for release v0.23.0-rc.0 (#85)
- [1358b920](https://github.com/kubedb/mysql-restic-plugin/commit/1358b920) Fix makefile for ubi images (#84)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.38.0-rc.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.38.0-rc.0)

- [57d25a8](https://github.com/kubedb/mysql-router-init/commit/57d25a8) Use k8s 1.32 client libs (#53)



## [kubedb/neo4j](https://github.com/kubedb/neo4j)

### [v0.1.0-rc.0](https://github.com/kubedb/neo4j/releases/tag/v0.1.0-rc.0)

- [669b8b79](https://github.com/kubedb/neo4j/commit/669b8b79) Use k8s 1.32 client libraries (#6)
- [3d0c2f6f](https://github.com/kubedb/neo4j/commit/3d0c2f6f) Prepare for release v0.1.0-rc.0 (#5)
- [b2f7e8f0](https://github.com/kubedb/neo4j/commit/b2f7e8f0) Fix Readme
- [63ee24ca](https://github.com/kubedb/neo4j/commit/63ee24ca) replace go1.25.5 to go1.25 (#3)
- [78e8ae36](https://github.com/kubedb/neo4j/commit/78e8ae36) Neo4j Operator



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.47.0-rc.0](https://github.com/kubedb/ops-manager/releases/tag/v0.47.0-rc.0)

- [cb3b4b33](https://github.com/kubedb/ops-manager/commit/cb3b4b332) Prepare for release v0.47.0-rc.0 (#809)
- [5bdccf30](https://github.com/kubedb/ops-manager/commit/5bdccf30f) Skip controller activation in certification mode (#808)
- [3eb8519f](https://github.com/kubedb/ops-manager/commit/3eb8519f1) Fix linter warning
- [4f15899d](https://github.com/kubedb/ops-manager/commit/4f15899d1) Use golangci-lint 2.x (#807)



## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.6.0-rc.0](https://github.com/kubedb/oracle/releases/tag/v0.6.0-rc.0)

- [f8d1e749](https://github.com/kubedb/oracle/commit/f8d1e749) Prepare for release v0.6.0-rc.0 (#20)
- [0aaf27b2](https://github.com/kubedb/oracle/commit/0aaf27b2) Oracle tls complete after review (#18)
- [06611a2f](https://github.com/kubedb/oracle/commit/06611a2f) Moved ops code in oracle operator (#19)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.6.0-rc.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.6.0-rc.0)

- [6ca661e](https://github.com/kubedb/oracle-coordinator/commit/6ca661e) Prepare for release v0.6.0-rc.0 (#15)
- [f3d4752](https://github.com/kubedb/oracle-coordinator/commit/f3d4752) Build ubi image (#14)
- [c141fa7](https://github.com/kubedb/oracle-coordinator/commit/c141fa7) Use golangci-lint 2.x (#13)
- [f0b4cbb](https://github.com/kubedb/oracle-coordinator/commit/f0b4cbb) Support TLS (#12)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.47.0-rc.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.47.0-rc.0)

- [528043c1](https://github.com/kubedb/percona-xtradb/commit/528043c17) Prepare for release v0.47.0-rc.0 (#424)
- [c3cb6155](https://github.com/kubedb/percona-xtradb/commit/c3cb6155c) update go version in makefile (#423)
- [654c6464](https://github.com/kubedb/percona-xtradb/commit/654c6464e) Use golangci-lint 2.x (#422)
- [0ec47977](https://github.com/kubedb/percona-xtradb/commit/0ec479776) fix multiple restart issue by introducing parallelism controller (#421)
- [4208a065](https://github.com/kubedb/percona-xtradb/commit/4208a065f) move ops code to db repo (#420)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.33.0-rc.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.33.0-rc.0)

- [d0d71987](https://github.com/kubedb/percona-xtradb-coordinator/commit/d0d71987) Prepare for release v0.33.0-rc.0 (#105)
- [f3648936](https://github.com/kubedb/percona-xtradb-coordinator/commit/f3648936) Update image source and name in Dockerfile.ubi
- [9e9945b7](https://github.com/kubedb/percona-xtradb-coordinator/commit/9e9945b7) Build ubi image (#104)
- [437be99c](https://github.com/kubedb/percona-xtradb-coordinator/commit/437be99c) Use golangci-lint 2.x (#103)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.44.0-rc.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.44.0-rc.0)

- [452097bc](https://github.com/kubedb/pg-coordinator/commit/452097bc) Prepare for release v0.44.0-rc.0 (#219)
- [9c1eeec7](https://github.com/kubedb/pg-coordinator/commit/9c1eeec7) Increase High availabilty of cluster (#211)
- [5ea8f8d1](https://github.com/kubedb/pg-coordinator/commit/5ea8f8d1) Fix linter warning
- [b3b8080a](https://github.com/kubedb/pg-coordinator/commit/b3b8080a) Fix linter warnings (#218)
- [0d13c636](https://github.com/kubedb/pg-coordinator/commit/0d13c636) Fix container build command
- [97da5fdf](https://github.com/kubedb/pg-coordinator/commit/97da5fdf) Build ubi image (#217)
- [618be9b7](https://github.com/kubedb/pg-coordinator/commit/618be9b7) Use golangci-lint 2.x (#216)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.47.0-rc.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.47.0-rc.0)

- [f90f38f4](https://github.com/kubedb/pgbouncer/commit/f90f38f40) Prepare for release v0.47.0-rc.0 (#388)
- [d8410281](https://github.com/kubedb/pgbouncer/commit/d84102818) Add golnagci (#387)
- [2b58993f](https://github.com/kubedb/pgbouncer/commit/2b58993f3) Add virtual secret and update Lint (#383)
- [6396e383](https://github.com/kubedb/pgbouncer/commit/6396e383b) Move ops code to db repo (#386)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.15.0-rc.0](https://github.com/kubedb/pgpool/releases/tag/v0.15.0-rc.0)

- [536c4368](https://github.com/kubedb/pgpool/commit/536c4368) Prepare for release v0.15.0-rc.0 (#90)
- [ce38a201](https://github.com/kubedb/pgpool/commit/ce38a201) Add golangci (#89)
- [92bf292e](https://github.com/kubedb/pgpool/commit/92bf292e) add virtual secret and update lint (#88)
- [66cfd1c1](https://github.com/kubedb/pgpool/commit/66cfd1c1) Move ops code to db repo (#87)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.60.0-rc.0](https://github.com/kubedb/postgres/releases/tag/v0.60.0-rc.0)

- [43ab7076](https://github.com/kubedb/postgres/commit/43ab70768) Prepare for release v0.60.0-rc.0 (#842)
- [1f118af2](https://github.com/kubedb/postgres/commit/1f118af2f) Fix postgres split brain | Add auto config tuning support |   (#841)
- [73418b2f](https://github.com/kubedb/postgres/commit/73418b2f8) Virtual Secret further bug fix (#836)
- [57e8114e](https://github.com/kubedb/postgres/commit/57e8114ed) Move in Ops-manager code to Postgres (#837)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.21.0-rc.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.21.0-rc.0)

- [1658ccf6](https://github.com/kubedb/postgres-archiver/commit/1658ccf6) Prepare for release v0.21.0-rc.0 (#72)
- [fb8f8428](https://github.com/kubedb/postgres-archiver/commit/fb8f8428) Update Go version (#71)
- [d0df1d7b](https://github.com/kubedb/postgres-archiver/commit/d0df1d7b) Merge pull request #70 from kubedb/lint
- [26d3ba73](https://github.com/kubedb/postgres-archiver/commit/26d3ba73) update linter
- [4f806c00](https://github.com/kubedb/postgres-archiver/commit/4f806c00) Build and push ubi image (#69)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.21.0-rc.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.21.0-rc.0)

- [0b38fae9](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/0b38fae9) Prepare for release v0.21.0-rc.0 (#69)
- [6a3af86c](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/6a3af86c) Update golang version (#68)
- [3ad4fd33](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/3ad4fd33) update lint (#67)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.23.0-rc.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.23.0-rc.0)

- [397396b](https://github.com/kubedb/postgres-restic-plugin/commit/397396b) Prepare for release v0.23.0-rc.0 (#83)
- [b53cd1b](https://github.com/kubedb/postgres-restic-plugin/commit/b53cd1b) Update Go version (#82)
- [cfe4796](https://github.com/kubedb/postgres-restic-plugin/commit/cfe4796) Update linter (#81)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.21.0-rc.0](https://github.com/kubedb/provider-aws/releases/tag/v0.21.0-rc.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.21.0-rc.0](https://github.com/kubedb/provider-azure/releases/tag/v0.21.0-rc.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.21.0-rc.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.21.0-rc.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.60.0-rc.0](https://github.com/kubedb/provisioner/releases/tag/v0.60.0-rc.0)

- [8bae571d](https://github.com/kubedb/provisioner/commit/8bae571d5) Prepare for release v0.60.0-rc.0 (#177)
- [47838996](https://github.com/kubedb/provisioner/commit/478389968) Add new DBs (#176)
- [029d1c60](https://github.com/kubedb/provisioner/commit/029d1c60c) added milvus provisioner (#171)
- [00ec61f1](https://github.com/kubedb/provisioner/commit/00ec61f1c) Add Qdrant (#170)
- [b8abc67f](https://github.com/kubedb/provisioner/commit/b8abc67f1) Skip controller activation in certification mode (#174)
- [b2ecb2f8](https://github.com/kubedb/provisioner/commit/b2ecb2f89) Update deps (#173)
- [f83d0339](https://github.com/kubedb/provisioner/commit/f83d0339b) Use golangci-lint 2.x (#172)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.47.0-rc.0](https://github.com/kubedb/proxysql/releases/tag/v0.47.0-rc.0)

- [1f965f4f](https://github.com/kubedb/proxysql/commit/1f965f4fd) Prepare for release v0.47.0-rc.0 (#408)
- [4ed80eed](https://github.com/kubedb/proxysql/commit/4ed80eed3) Fix multiple restart issue by introducing parallelismController (#407)
- [111aa2aa](https://github.com/kubedb/proxysql/commit/111aa2aa5) Move ops to db repo (#406)



## [kubedb/qdrant](https://github.com/kubedb/qdrant)

### [v0.1.0-rc.0](https://github.com/kubedb/qdrant/releases/tag/v0.1.0-rc.0)

- [61f7198f](https://github.com/kubedb/qdrant/commit/61f7198f) Prepare for release v0.1.0-rc.0 (#7)
- [55cd4a5c](https://github.com/kubedb/qdrant/commit/55cd4a5c) k8s downgrade (#6)
- [23165061](https://github.com/kubedb/qdrant/commit/23165061) makefile go 1.25 (#4)
- [a3c6b74e](https://github.com/kubedb/qdrant/commit/a3c6b74e) Add Qdrant Provisioner
- [251a144d](https://github.com/kubedb/qdrant/commit/251a144d) Update .gitignore



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.15.0-rc.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.15.0-rc.0)

- [8a124e34](https://github.com/kubedb/rabbitmq/commit/8a124e34) Prepare for release v0.15.0-rc.0 (#102)
- [4b3e6a49](https://github.com/kubedb/rabbitmq/commit/4b3e6a49) Fix multiple restart issue by introducing parallelismController (#101)
- [f05a5396](https://github.com/kubedb/rabbitmq/commit/f05a5396) Move ops manager repo to base (#100)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.53.0-rc.0](https://github.com/kubedb/redis/releases/tag/v0.53.0-rc.0)

- [acecd5f2](https://github.com/kubedb/redis/commit/acecd5f21) Prepare for release v0.53.0-rc.0 (#611)
- [fc0fe37b](https://github.com/kubedb/redis/commit/fc0fe37bd) add Virtual Secret and update lint (#610)
- [994b34ec](https://github.com/kubedb/redis/commit/994b34ece) Move ops code to db repo (#609)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.39.0-rc.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.39.0-rc.0)

- [75f9f5d9](https://github.com/kubedb/redis-coordinator/commit/75f9f5d9) Prepare for release v0.39.0-rc.0 (#139)
- [5f7ce82b](https://github.com/kubedb/redis-coordinator/commit/5f7ce82b) Virtual Secret added and Make Lint update (#138)
- [520f67f2](https://github.com/kubedb/redis-coordinator/commit/520f67f2) Update Dockerfile label for Redis Coordinator
- [a190c965](https://github.com/kubedb/redis-coordinator/commit/a190c965) Build ubi image (#137)
- [1d673b73](https://github.com/kubedb/redis-coordinator/commit/1d673b73) Use golangci-lint 2.x (#136)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.23.0-rc.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.23.0-rc.0)

- [7dfce486](https://github.com/kubedb/redis-restic-plugin/commit/7dfce486) Prepare for release v0.23.0-rc.0 (#78)
- [fec65710](https://github.com/kubedb/redis-restic-plugin/commit/fec65710) Update linter (#77)
- [927601ab](https://github.com/kubedb/redis-restic-plugin/commit/927601ab) Update make lint and Add Virtual Secret (#76)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.47.0-rc.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.47.0-rc.0)

- [7c77bd18](https://github.com/kubedb/replication-mode-detector/commit/7c77bd18) Prepare for release v0.47.0-rc.0 (#302)
- [5a586bb3](https://github.com/kubedb/replication-mode-detector/commit/5a586bb3) Update Dockerfile label for application name
- [09bc4ae6](https://github.com/kubedb/replication-mode-detector/commit/09bc4ae6) Build ubi image (#301)
- [fd4f2bb5](https://github.com/kubedb/replication-mode-detector/commit/fd4f2bb5) Use golangci-lint 2.x (#300)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.36.0-rc.0](https://github.com/kubedb/schema-manager/releases/tag/v0.36.0-rc.0)

- [1ebf2d05](https://github.com/kubedb/schema-manager/commit/1ebf2d05) Prepare for release v0.36.0-rc.0 (#149)
- [f58a1650](https://github.com/kubedb/schema-manager/commit/f58a1650) Fix linter (#148)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.15.0-rc.0](https://github.com/kubedb/singlestore/releases/tag/v0.15.0-rc.0)

- [f3529d24](https://github.com/kubedb/singlestore/commit/f3529d24) Prepare for release v0.15.0-rc.0 (#91)
- [5378a876](https://github.com/kubedb/singlestore/commit/5378a876) update go version in makefile (#90)
- [57f14b26](https://github.com/kubedb/singlestore/commit/57f14b26) Use golangci-lint 2.x (#89)
- [d2c074b5](https://github.com/kubedb/singlestore/commit/d2c074b5) fix multiple restart issue by introducing parallelism controller (#87)
- [4fe9d994](https://github.com/kubedb/singlestore/commit/4fe9d994) Fix frequent db patch and requeue issue (#88)
- [d771a53e](https://github.com/kubedb/singlestore/commit/d771a53e) move ops code to db repo (#86)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.15.0-rc.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.15.0-rc.0)

- [0fe5c82](https://github.com/kubedb/singlestore-coordinator/commit/0fe5c82) Prepare for release v0.15.0-rc.0 (#52)
- [4fd8963](https://github.com/kubedb/singlestore-coordinator/commit/4fd8963) Update Dockerfile label from 'External DNS Operator' to 'Singlestore Coordinator'
- [2115237](https://github.com/kubedb/singlestore-coordinator/commit/2115237) Build ubi image (#51)
- [a2c1fa5](https://github.com/kubedb/singlestore-coordinator/commit/a2c1fa5) Use golangci-lint 2.x (#50)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.18.0-rc.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.18.0-rc.0)

- [88634e7](https://github.com/kubedb/singlestore-restic-plugin/commit/88634e7) Prepare for release v0.18.0-rc.0 (#58)
- [285e464](https://github.com/kubedb/singlestore-restic-plugin/commit/285e464) update go version in makefile (#57)
- [25bff27](https://github.com/kubedb/singlestore-restic-plugin/commit/25bff27) Use golangci-lint 2.x (#56)
- [25f1e62](https://github.com/kubedb/singlestore-restic-plugin/commit/25f1e62) Fix makefile for ubi images (#55)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.15.0-rc.0](https://github.com/kubedb/solr/releases/tag/v0.15.0-rc.0)

- [090077d5](https://github.com/kubedb/solr/commit/090077d5) Prepare for release v0.15.0-rc.0 (#102)
- [cb654575](https://github.com/kubedb/solr/commit/cb654575) Fix multiple restart issue by introducing parallelismController (#101)
- [8919aabd](https://github.com/kubedb/solr/commit/8919aabd) Move in Ops-manager code to Solr (#100)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.45.0-rc.0](https://github.com/kubedb/tests/releases/tag/v0.45.0-rc.0)

- [2bd7d59f](https://github.com/kubedb/tests/commit/2bd7d59f) Prepare for release v0.45.0-rc.0 (#495)
- [c91842f7](https://github.com/kubedb/tests/commit/c91842f7) Update go version (#494)
- [e2d20207](https://github.com/kubedb/tests/commit/e2d20207) Release Change: Go- 1.25.5, Linter Fix (#493)
- [ec7c52a1](https://github.com/kubedb/tests/commit/ec7c52a1) MariaDB MaxScale Volume Expansion (#492)
- [2096100e](https://github.com/kubedb/tests/commit/2096100e) MariaDB MaxScale Scaling (Ops) (#488)
- [82246e84](https://github.com/kubedb/tests/commit/82246e84) Add restic, Volumesnapshotter to CI (#490)
- [08bba9ff](https://github.com/kubedb/tests/commit/08bba9ff) Add Base-Backup Mode for Postgres (#484)
- [42ba4e21](https://github.com/kubedb/tests/commit/42ba4e21) Test against k8s 1.34 (#489)
- [16eed29b](https://github.com/kubedb/tests/commit/16eed29b) MySQL Archiver CI  (restic+volumesnapshotter mode) (#473)
- [a635b1eb](https://github.com/kubedb/tests/commit/a635b1eb) add rotate-auth for semi-sync, innodb (#487)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.36.0-rc.0](https://github.com/kubedb/ui-server/releases/tag/v0.36.0-rc.0)

- [f5d4d79e](https://github.com/kubedb/ui-server/commit/f5d4d79e) Prepare for release v0.36.0-rc.0 (#178)
- [3563951b](https://github.com/kubedb/ui-server/commit/3563951b) Update Lint (#177)
- [072090b3](https://github.com/kubedb/ui-server/commit/072090b3) make lint update (#176)



## [kubedb/weaviate](https://github.com/kubedb/weaviate)

### [v0.1.0-rc.0](https://github.com/kubedb/weaviate/releases/tag/v0.1.0-rc.0)

- [8bb614d](https://github.com/kubedb/weaviate/commit/8bb614d) Prepare for release v0.1.0-rc.0 (#4)
- [d53d53e](https://github.com/kubedb/weaviate/commit/d53d53e) Add Weaviate DB Support
- [1b27e93](https://github.com/kubedb/weaviate/commit/1b27e93) Add vendor



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.36.0-rc.0](https://github.com/kubedb/webhook-server/releases/tag/v0.36.0-rc.0)

- [4f47d8c3](https://github.com/kubedb/webhook-server/commit/4f47d8c3) Prepare for release v0.36.0-rc.0 (#180)
- [31fbaa00](https://github.com/kubedb/webhook-server/commit/31fbaa00) Fix build
- [6e256028](https://github.com/kubedb/webhook-server/commit/6e256028) Add missing new dbs in client-setup
- [b90632db](https://github.com/kubedb/webhook-server/commit/b90632db) added milvus webhook-server (#177)
- [9d58f0c3](https://github.com/kubedb/webhook-server/commit/9d58f0c3) Add Qdrant (#176)
- [0ce55e0a](https://github.com/kubedb/webhook-server/commit/0ce55e0a) Add HanaDB (#178)



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.8.0-rc.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.8.0-rc.0)

- [074b334](https://github.com/kubedb/xtrabackup-restic-plugin/commit/074b334) Prepare for release v0.8.0-rc.0 (#26)
- [84dae3a](https://github.com/kubedb/xtrabackup-restic-plugin/commit/84dae3a) update go version in makefile (#25)
- [37d546d](https://github.com/kubedb/xtrabackup-restic-plugin/commit/37d546d) Use golangci-lint 2.x (#24)
- [4255c2f](https://github.com/kubedb/xtrabackup-restic-plugin/commit/4255c2f) Fix makefile for ubi images (#23)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.15.0-rc.0](https://github.com/kubedb/zookeeper/releases/tag/v0.15.0-rc.0)

- [4be45687](https://github.com/kubedb/zookeeper/commit/4be45687) Prepare for release v0.15.0-rc.0 (#94)
- [c1da0e5f](https://github.com/kubedb/zookeeper/commit/c1da0e5f) replace go1.25.5 to go1.25 (#93)
- [55fe861b](https://github.com/kubedb/zookeeper/commit/55fe861b) Fix multiple restart issue by introducing parallelismController (#91)
- [c75acf90](https://github.com/kubedb/zookeeper/commit/c75acf90) Make Update go version and lint fix (#92)
- [23219262](https://github.com/kubedb/zookeeper/commit/23219262) Move ops code to db repo (#90)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.16.0-rc.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.16.0-rc.0)

- [8970e4d](https://github.com/kubedb/zookeeper-restic-plugin/commit/8970e4d) Prepare for release v0.16.0-rc.0 (#47)
- [079fdca](https://github.com/kubedb/zookeeper-restic-plugin/commit/079fdca) Go version update (#46)
- [4c5f81a](https://github.com/kubedb/zookeeper-restic-plugin/commit/4c5f81a) linter fix (#45)




