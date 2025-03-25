---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2025.3.24
    name: Changelog-v2025.3.24
    parent: welcome
    weight: 20250324
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2025.3.24/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2025.3.24/
---

# KubeDB v2025.3.24 (2025-03-25)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.53.0](https://github.com/kubedb/apimachinery/releases/tag/v0.53.0)

- [1f19088e](https://github.com/kubedb/apimachinery/commit/1f19088e5) Update for release KubeStash@v2025.3.24 (#1437)
- [c3966a55](https://github.com/kubedb/apimachinery/commit/c3966a553) Dont use selector while setting up restoresession reconciler (#1436)
- [b7bdda58](https://github.com/kubedb/apimachinery/commit/b7bdda585) Configure restore session controller name by db kind
- [d15c5892](https://github.com/kubedb/apimachinery/commit/d15c58928) Udate license verifier
- [525f2962](https://github.com/kubedb/apimachinery/commit/525f29629) Generate release notes (#1435)
- [26f19746](https://github.com/kubedb/apimachinery/commit/26f197466) Configure restoresession controller name (#1434)
- [ceb4a0bc](https://github.com/kubedb/apimachinery/commit/ceb4a0bcc) Update deps
- [7b4a7dd0](https://github.com/kubedb/apimachinery/commit/7b4a7dd05) Fix MariaDB Webhook (#1432)
- [12201f9f](https://github.com/kubedb/apimachinery/commit/12201f9fa) Add const for zookeeper-ready  (#1433)
- [749f062a](https://github.com/kubedb/apimachinery/commit/749f062a9) Add postgres version in FerretDBVersion (#1429)
- [c1c97c5d](https://github.com/kubedb/apimachinery/commit/c1c97c5d7) Update deps
- [e7fb1125](https://github.com/kubedb/apimachinery/commit/e7fb1125f) Move & convert oldDB validators to new-webhook signature (#1428)
- [d82a3f17](https://github.com/kubedb/apimachinery/commit/d82a3f17c) Register cassandra autoscaler types
- [4d0febcc](https://github.com/kubedb/apimachinery/commit/4d0febcc5) Rename SecretSource -> SecretStore
- [f80d5dc7](https://github.com/kubedb/apimachinery/commit/f80d5dc74) Move webhooks folder inside pkg
- [5e31d682](https://github.com/kubedb/apimachinery/commit/5e31d6825) update ferretdb autoscaler api (#1427)
- [5d295e37](https://github.com/kubedb/apimachinery/commit/5d295e374) Move & convert opsRequest validator files to new webhook style (#1425)
- [051711da](https://github.com/kubedb/apimachinery/commit/051711da2) Update deps
- [70c28b34](https://github.com/kubedb/apimachinery/commit/70c28b34c) Update autoscaler webhooks (#1424)
- [a9caf6ac](https://github.com/kubedb/apimachinery/commit/a9caf6acf) Update deps
- [999c0b80](https://github.com/kubedb/apimachinery/commit/999c0b803) Update deps
- [ee518661](https://github.com/kubedb/apimachinery/commit/ee5186618) Fix build (#1423)
- [28f8a1d6](https://github.com/kubedb/apimachinery/commit/28f8a1d6d) Add FerretDB replication support (#1420)
- [4fcb10d1](https://github.com/kubedb/apimachinery/commit/4fcb10d1c) Integrate virtual-secret in `authSecret` (#1414)
- [7b3b2d57](https://github.com/kubedb/apimachinery/commit/7b3b2d57a) Remove defaultclient from kafka api group (#1419)
- [3b730ac8](https://github.com/kubedb/apimachinery/commit/3b730ac81) Add Engine-type in Redis-Version to distinguish valkey version (#1410)
- [f5d0d929](https://github.com/kubedb/apimachinery/commit/f5d0d929f) Fix ChangeRequest Status type (#1411)
- [9493bd16](https://github.com/kubedb/apimachinery/commit/9493bd16d) Update schema manager to new webhook style (#1421)
- [dc851d6f](https://github.com/kubedb/apimachinery/commit/dc851d6fb) Increate MSSQL CPU Request (#1412)
- [38bb7a72](https://github.com/kubedb/apimachinery/commit/38bb7a72c) Add Operator Sharding Support (#1415)
- [5db581f5](https://github.com/kubedb/apimachinery/commit/5db581f54) Use k8s 1.32 client libs (#1405)
- [64f26616](https://github.com/kubedb/apimachinery/commit/64f266165) Set securityContext for exporter containers (#1418)
- [8673786a](https://github.com/kubedb/apimachinery/commit/8673786aa) Add IsCluster inside Replication (#1416)
- [d57315cd](https://github.com/kubedb/apimachinery/commit/d57315cdf) Add regex to kubedb object names (#1417)
- [c3871c5d](https://github.com/kubedb/apimachinery/commit/c3871c5de) Test against k8s 1.32 (#1409)
- [5065873e](https://github.com/kubedb/apimachinery/commit/5065873e2) Test against k8s 1.32 (#1407)
- [93b6bea8](https://github.com/kubedb/apimachinery/commit/93b6bea8a) Fix CI (#1408)
- [d685cb86](https://github.com/kubedb/apimachinery/commit/d685cb862) Add gitops apis (#1406)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.38.0](https://github.com/kubedb/autoscaler/releases/tag/v0.38.0)




## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.6.0](https://github.com/kubedb/cassandra/releases/tag/v0.6.0)

- [dfcd0bec](https://github.com/kubedb/cassandra/commit/dfcd0bec) Prepare for release v0.6.0 (#31)
- [9288c65b](https://github.com/kubedb/cassandra/commit/9288c65b) Prepare for release v0.6.0-rc.1 (#30)
- [47f7e4fe](https://github.com/kubedb/cassandra/commit/47f7e4fe) Prepare for release v0.6.0-rc.0 (#29)
- [458feaa9](https://github.com/kubedb/cassandra/commit/458feaa9) Update webhook path (#28)
- [b592f3d3](https://github.com/kubedb/cassandra/commit/b592f3d3) Install ace-user-roles and Add Operator Sharding (#25)
- [a931ee8b](https://github.com/kubedb/cassandra/commit/a931ee8b) Update to k8s v1.32 (#27)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.53.0](https://github.com/kubedb/cli/releases/tag/v0.53.0)

- [b037e812](https://github.com/kubedb/cli/commit/b037e812b) Prepare for release v0.53.0 (#794)
- [3085c9cd](https://github.com/kubedb/cli/commit/3085c9cdd) Prepare for release v0.53.0-rc.1 (#793)
- [6a2800b1](https://github.com/kubedb/cli/commit/6a2800b19) Prepare for release v0.53.0-rc.0 (#792)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.8.0](https://github.com/kubedb/clickhouse/releases/tag/v0.8.0)

- [5ebe6b91](https://github.com/kubedb/clickhouse/commit/5ebe6b91) Prepare for release v0.8.0 (#47)
- [a1a9f411](https://github.com/kubedb/clickhouse/commit/a1a9f411) Update webhook setup (#46)
- [db836fa8](https://github.com/kubedb/clickhouse/commit/db836fa8) Prepare for release v0.8.0-rc.1 (#45)
- [e58549bc](https://github.com/kubedb/clickhouse/commit/e58549bc) Prepare for release v0.8.0-rc.0 (#43)
- [3f65aac1](https://github.com/kubedb/clickhouse/commit/3f65aac1) update webhook path (#42)
- [ad3d79cf](https://github.com/kubedb/clickhouse/commit/ad3d79cf) Add Operator Sharding Support (#41)
- [205feb39](https://github.com/kubedb/clickhouse/commit/205feb39) Update to k8s v32 (#40)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.8.0](https://github.com/kubedb/crd-manager/releases/tag/v0.8.0)

- [ed428dd4](https://github.com/kubedb/crd-manager/commit/ed428dd4) Prepare for release v0.8.0 (#73)
- [4782903b](https://github.com/kubedb/crd-manager/commit/4782903b) Fix for ferret crd changes (#71)
- [e5b58c7d](https://github.com/kubedb/crd-manager/commit/e5b58c7d) Prepare for release v0.8.0-rc.1 (#69)
- [e3384a94](https://github.com/kubedb/crd-manager/commit/e3384a94) Add gitops CRDs (#70)
- [f6bf57e7](https://github.com/kubedb/crd-manager/commit/f6bf57e7) Prepare for release v0.8.0-rc.0 (#68)
- [235b872d](https://github.com/kubedb/crd-manager/commit/235b872d) Update to k8s v0.32 (#67)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.11.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.11.0)

- [21936ca](https://github.com/kubedb/dashboard-restic-plugin/commit/21936ca) Prepare for release v0.11.0 (#37)
- [e30202c](https://github.com/kubedb/dashboard-restic-plugin/commit/e30202c) Prepare for release v0.11.0-rc.1 (#36)
- [bdd5aa5](https://github.com/kubedb/dashboard-restic-plugin/commit/bdd5aa5) Prepare for release v0.11.0-rc.0 (#35)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.8.0](https://github.com/kubedb/db-client-go/releases/tag/v0.8.0)

- [a9aa5a3e](https://github.com/kubedb/db-client-go/commit/a9aa5a3e) Prepare for release v0.8.0 (#169)
- [ca0ff72d](https://github.com/kubedb/db-client-go/commit/ca0ff72d) Prepare for release v0.8.0-rc.1 (#168)
- [2beb4691](https://github.com/kubedb/db-client-go/commit/2beb4691) Prepare for release v0.8.0-rc.0 (#167)
- [214a0339](https://github.com/kubedb/db-client-go/commit/214a0339) Remove decoding of data (#166)
- [0303ed1c](https://github.com/kubedb/db-client-go/commit/0303ed1c) Updates for integrating `Virtual-Secrets` (#165)
- [056c3ddf](https://github.com/kubedb/db-client-go/commit/056c3ddf) Fix TLS Handshake Issue for MySQL 5.7.x (#164)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.8.0](https://github.com/kubedb/druid/releases/tag/v0.8.0)

- [cece32a5](https://github.com/kubedb/druid/commit/cece32a5) Prepare for release v0.8.0 (#83)
- [3cdf6fc4](https://github.com/kubedb/druid/commit/3cdf6fc4) Configure restore session controller (#82)
- [6d86f999](https://github.com/kubedb/druid/commit/6d86f999) Update Webhook Setup (#81)
- [795a8bb0](https://github.com/kubedb/druid/commit/795a8bb0) Prepare for release v0.8.0-rc.1 (#80)
- [bb2ae901](https://github.com/kubedb/druid/commit/bb2ae901) Prepare for release v0.8.0-rc.0 (#79)
- [eb2421c5](https://github.com/kubedb/druid/commit/eb2421c5) Update webhook path (#78)
- [2a8160d1](https://github.com/kubedb/druid/commit/2a8160d1) Add Operator Sharding Support (#77)
- [1408db1d](https://github.com/kubedb/druid/commit/1408db1d) Fix webhook call
- [f6b65aeb](https://github.com/kubedb/druid/commit/f6b65aeb) Setup new webhook style; Update to k8s v1.32 (#76)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.53.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.53.0)

- [93f4f6f5](https://github.com/kubedb/elasticsearch/commit/93f4f6f56) Prepare for release v0.53.0 (#763)
- [869d96ff](https://github.com/kubedb/elasticsearch/commit/869d96ffc) Cleanup restoresession controller (#762)
- [373f9b40](https://github.com/kubedb/elasticsearch/commit/373f9b404) Fix restore session controller name
- [b5e9c2e5](https://github.com/kubedb/elasticsearch/commit/b5e9c2e5b) Prepare for release v0.53.0-rc.1 (#761)
- [0884a60d](https://github.com/kubedb/elasticsearch/commit/0884a60de) Prepare for release v0.53.0-rc.0 (#760)
- [3790689f](https://github.com/kubedb/elasticsearch/commit/3790689f4) Restructure elasticsearch webhook (#759)
- [bc877df4](https://github.com/kubedb/elasticsearch/commit/bc877df46) Add operator sharding for Elasticsearch (#758)
- [b8937d5b](https://github.com/kubedb/elasticsearch/commit/b8937d5b2) Webhook changes for k8s 1.32 (#757)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.16.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.16.0)

- [c734c701](https://github.com/kubedb/elasticsearch-restic-plugin/commit/c734c701) Prepare for release v0.16.0 (#61)
- [1e2569f5](https://github.com/kubedb/elasticsearch-restic-plugin/commit/1e2569f5) Prepare for release v0.16.0-rc.1 (#60)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.8.0](https://github.com/kubedb/ferretdb/releases/tag/v0.8.0)

- [58965013](https://github.com/kubedb/ferretdb/commit/58965013) Prepare for release v0.8.0 (#72)
- [065e8623](https://github.com/kubedb/ferretdb/commit/065e8623) Update webhook setup (#71)
- [621c6430](https://github.com/kubedb/ferretdb/commit/621c6430) Remove backend externally managed feature (#70)
- [5e59d781](https://github.com/kubedb/ferretdb/commit/5e59d781) Prepare for release v0.8.0-rc.1 (#69)
- [178a987c](https://github.com/kubedb/ferretdb/commit/178a987c) Prepare for release v0.8.0-rc.0 (#68)
- [aa8bb0a9](https://github.com/kubedb/ferretdb/commit/aa8bb0a9) Fix backend init scripts and labels-selectors (#67)
- [9a5064dc](https://github.com/kubedb/ferretdb/commit/9a5064dc) Add FerretDB replication support (#66)
- [6b72cb54](https://github.com/kubedb/ferretdb/commit/6b72cb54) Fix security context for backend pg and add operator sharding (#60)
- [cfd63de6](https://github.com/kubedb/ferretdb/commit/cfd63de6) Setup new webhook style; Update to k8s v1.32 (#65)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.1.0](https://github.com/kubedb/gitops/releases/tag/v0.1.0)

- [b90aee61](https://github.com/kubedb/gitops/commit/b90aee61) Prepare for release v0.1.0 (#11)
- [59222995](https://github.com/kubedb/gitops/commit/59222995) Standardize the codebase (#10)
- [b9208aa9](https://github.com/kubedb/gitops/commit/b9208aa9) Prepare for release v0.1.0-rc.1 (#9)
- [643d32d5](https://github.com/kubedb/gitops/commit/643d32d5) Prepare for release v0.1.0-rc.0 (#8)
- [79c495f8](https://github.com/kubedb/gitops/commit/79c495f8) Add Initial Setup for Dbs (#6)
- [466aad87](https://github.com/kubedb/gitops/commit/466aad87) Use firecracker runners for ci
- [741059fa](https://github.com/kubedb/gitops/commit/741059fa) Add license header



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2025.3.24](https://github.com/kubedb/installer/releases/tag/v2025.3.24)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.24.0](https://github.com/kubedb/kafka/releases/tag/v0.24.0)

- [6d23084d](https://github.com/kubedb/kafka/commit/6d23084d) Prepare for release v0.24.0 (#148)
- [a5a7cb40](https://github.com/kubedb/kafka/commit/a5a7cb40) Restructure admission webhook (#147)
- [9f2b6c69](https://github.com/kubedb/kafka/commit/9f2b6c69) Prepare for release v0.24.0-rc.1 (#146)
- [e8195ac4](https://github.com/kubedb/kafka/commit/e8195ac4) Update webhook dependency (#145)
- [1714ea67](https://github.com/kubedb/kafka/commit/1714ea67) Add Kafka Operator Shading Support (#144)
- [a5d76b95](https://github.com/kubedb/kafka/commit/a5d76b95) Setup new webhook style; Update to k8s v1.32 (#143)
- [bb875226](https://github.com/kubedb/kafka/commit/bb875226) Install ace-user-roles; Run tests twice a week (#142)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.29.0](https://github.com/kubedb/kibana/releases/tag/v0.29.0)

- [1763629f](https://github.com/kubedb/kibana/commit/1763629f) Prepare for release v0.29.0 (#150)
- [b6d66c62](https://github.com/kubedb/kibana/commit/b6d66c62) Configure restoresession controller (#149)
- [5647fe5c](https://github.com/kubedb/kibana/commit/5647fe5c) Update webhook setup (#148)
- [2ca1a410](https://github.com/kubedb/kibana/commit/2ca1a410) Prepare for release v0.29.0-rc.1 (#147)
- [027a3bd0](https://github.com/kubedb/kibana/commit/027a3bd0) Prepare for release v0.29.0-rc.0 (#146)
- [10a89f00](https://github.com/kubedb/kibana/commit/10a89f00) Fix dependency for kibana (#145)
- [a9c6e233](https://github.com/kubedb/kibana/commit/a9c6e233) Add operator sharding support for kibana (#144)
- [71b978a9](https://github.com/kubedb/kibana/commit/71b978a9) Webhook changes for k8s 1.32 (#143)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.16.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.16.0)

- [e14d1611](https://github.com/kubedb/kubedb-manifest-plugin/commit/e14d1611) Prepare for release v0.16.0 (#92)
- [8dbc92d0](https://github.com/kubedb/kubedb-manifest-plugin/commit/8dbc92d0) Prepare for release v0.16.0-rc.1 (#91)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.4.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.4.0)

- [92cdf9b](https://github.com/kubedb/kubedb-verifier/commit/92cdf9b) Prepare for release v0.4.0 (#17)
- [cb903bb](https://github.com/kubedb/kubedb-verifier/commit/cb903bb) Prepare for release v0.4.0-rc.1 (#16)
- [8b3a5e6](https://github.com/kubedb/kubedb-verifier/commit/8b3a5e6) Prepare for release v0.4.0-rc.0 (#15)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.37.0](https://github.com/kubedb/mariadb/releases/tag/v0.37.0)

- [45a9741e](https://github.com/kubedb/mariadb/commit/45a9741e3) Prepare for release v0.37.0 (#326)
- [d8aa6c5c](https://github.com/kubedb/mariadb/commit/d8aa6c5cc) Cleanup restoresession reconciler (#325)
- [90add00e](https://github.com/kubedb/mariadb/commit/90add00e1) Fix Standalone Mode, Init Container Security Context for Maxscale (#324)
- [f6bad15c](https://github.com/kubedb/mariadb/commit/f6bad15c0) Prepare for release v0.37.0-rc.1 (#323)
- [12653892](https://github.com/kubedb/mariadb/commit/126538929) Restructure admission webhook (#322)
- [c858218f](https://github.com/kubedb/mariadb/commit/c858218f2) Run Maxscale as Non Root User (#321)
- [9aaa179a](https://github.com/kubedb/mariadb/commit/9aaa179ad) Add MaxScale Cluster Support (#318)
- [84f8d44c](https://github.com/kubedb/mariadb/commit/84f8d44c8) Update IRSA annotations (#320)
- [510aeb7a](https://github.com/kubedb/mariadb/commit/510aeb7a8) ShardConfiguration, Install ace-user-roles, Run tests twice a week,  (#316)
- [b4fc4ddc](https://github.com/kubedb/mariadb/commit/b4fc4ddc7) Update to k8s v32 (#319)
- [08554f66](https://github.com/kubedb/mariadb/commit/08554f660) Add MariaDB Replication Support (#303)
- [e2212de1](https://github.com/kubedb/mariadb/commit/e2212de15) Lower Sidekick Default Resource (#315)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.13.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.13.0)

- [e31a2586](https://github.com/kubedb/mariadb-archiver/commit/e31a2586) Prepare for release v0.13.0 (#47)
- [ea393fb4](https://github.com/kubedb/mariadb-archiver/commit/ea393fb4) Prepare for release v0.13.0-rc.1 (#46)
- [a7e37d89](https://github.com/kubedb/mariadb-archiver/commit/a7e37d89) Prepare for release v0.13.0-rc.0 (#45)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.33.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.33.0)

- [2e94e104](https://github.com/kubedb/mariadb-coordinator/commit/2e94e104) Prepare for release v0.33.0 (#141)
- [fe289242](https://github.com/kubedb/mariadb-coordinator/commit/fe289242) Prepare for release v0.33.0-rc.1 (#140)
- [a19d30bf](https://github.com/kubedb/mariadb-coordinator/commit/a19d30bf) Add Support for MariaDB Replication (#134)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.13.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.13.0)

- [9754cf52](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/9754cf52) Prepare for release v0.13.0 (#45)
- [87587225](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/87587225) Prepare for release v0.13.0-rc.1 (#44)
- [6d1a6a2e](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/6d1a6a2e) Prepare for release v0.13.0-rc.0 (#43)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.11.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.11.0)

- [a1ac3ce](https://github.com/kubedb/mariadb-restic-plugin/commit/a1ac3ce) Prepare for release v0.11.0 (#44)
- [d525ec1](https://github.com/kubedb/mariadb-restic-plugin/commit/d525ec1) Prepare for release v0.11.0-rc.1 (#43)
- [bc5977d](https://github.com/kubedb/mariadb-restic-plugin/commit/bc5977d) Prepare for release v0.11.0-rc.0 (#42)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.46.0](https://github.com/kubedb/memcached/releases/tag/v0.46.0)

- [2c6dff64](https://github.com/kubedb/memcached/commit/2c6dff64e) Prepare for release v0.46.0 (#496)
- [ca7ab019](https://github.com/kubedb/memcached/commit/ca7ab0190) Prepare for release v0.46.0-rc.1 (#494)
- [b5ddb698](https://github.com/kubedb/memcached/commit/b5ddb698e) Prepare for release v0.46.0-rc.0 (#493)
- [816d7472](https://github.com/kubedb/memcached/commit/816d7472e) Restructure admission webhook (#492)
- [5ee33f9c](https://github.com/kubedb/memcached/commit/5ee33f9c3) Add operator sharding support; Update ci & makefile (#490)
- [93a42e71](https://github.com/kubedb/memcached/commit/93a42e718) Update to k8s v32 (#491)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.46.0](https://github.com/kubedb/mongodb/releases/tag/v0.46.0)

- [24070d37](https://github.com/kubedb/mongodb/commit/24070d373) Prepare for release v0.46.0 (#700)
- [3d9eeb7d](https://github.com/kubedb/mongodb/commit/3d9eeb7d3) Cleanup restoresession reconciler (#699)
- [53363ae4](https://github.com/kubedb/mongodb/commit/53363ae46) Configure restore session controller name
- [6da8b44f](https://github.com/kubedb/mongodb/commit/6da8b44fd) Prepare for release v0.46.0-rc.1 (#698)
- [d81d312a](https://github.com/kubedb/mongodb/commit/d81d312a8) Prepare for release v0.46.0-rc.0 (#697)
- [ce969181](https://github.com/kubedb/mongodb/commit/ce969181d) Restructure admission webhook (#696)
- [ce4b97c8](https://github.com/kubedb/mongodb/commit/ce4b97c87) Implement operator-sharding; Add IRSA annotation in sideick (#694)
- [68a81a75](https://github.com/kubedb/mongodb/commit/68a81a756) Update to k8s v32 (#692)
- [f9fe924e](https://github.com/kubedb/mongodb/commit/f9fe924e7) Merge pull request #690 from kubedb/mongo-ci-fix
- [77483bab](https://github.com/kubedb/mongodb/commit/77483bab5) change manual trigger, default case
- [ae1291b1](https://github.com/kubedb/mongodb/commit/ae1291b1b) add dropdown options to input sections
- [6a0a313a](https://github.com/kubedb/mongodb/commit/6a0a313a9) check after matrix profile remove
- [e80ca10f](https://github.com/kubedb/mongodb/commit/e80ca10f7) review changes done
- [3c14acc9](https://github.com/kubedb/mongodb/commit/3c14acc97) separate dynamic profiles
- [2aa206d8](https://github.com/kubedb/mongodb/commit/2aa206d8b) nhasdg
- [30555b40](https://github.com/kubedb/mongodb/commit/30555b400) fix checkout repo
- [0374303a](https://github.com/kubedb/mongodb/commit/0374303ae) checkout from test to working repo
- [47616663](https://github.com/kubedb/mongodb/commit/47616663a) add db_mode
- [5a261d5b](https://github.com/kubedb/mongodb/commit/5a261d5bf) wip
- [fad5ac51](https://github.com/kubedb/mongodb/commit/fad5ac51d) Set default resources on the walg container
- [5d6bedb4](https://github.com/kubedb/mongodb/commit/5d6bedb47) Install ace-user-roles; Run tests twice a week (#691)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.14.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.14.0)

- [d944f36c](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/d944f36c) Prepare for release v0.14.0 (#49)
- [f23ef219](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/f23ef219) Prepare for release v0.14.0-rc.1 (#48)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.16.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.16.0)

- [7f87dc0](https://github.com/kubedb/mongodb-restic-plugin/commit/7f87dc0) Prepare for release v0.16.0 (#82)
- [849d2e4](https://github.com/kubedb/mongodb-restic-plugin/commit/849d2e4) Prepare for release v0.16.0-rc.1 (#81)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.8.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.8.0)

- [d420a4b3](https://github.com/kubedb/mssql-coordinator/commit/d420a4b3) Prepare for release v0.8.0 (#34)
- [d2e1774a](https://github.com/kubedb/mssql-coordinator/commit/d2e1774a) Prepare for release v0.8.0-rc.1 (#33)
- [cbb75c6a](https://github.com/kubedb/mssql-coordinator/commit/cbb75c6a) Prepare for release v0.8.0-rc.0 (#32)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.8.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.8.0)

- [3c2620e3](https://github.com/kubedb/mssqlserver/commit/3c2620e3) Prepare for release v0.8.0 (#76)
- [eaa936d6](https://github.com/kubedb/mssqlserver/commit/eaa936d6) Update webhook setup (#74)
- [8fe96c25](https://github.com/kubedb/mssqlserver/commit/8fe96c25) Configure restore session controller name
- [9798e329](https://github.com/kubedb/mssqlserver/commit/9798e329) Prepare for release v0.8.0-rc.1 (#73)
- [e499126b](https://github.com/kubedb/mssqlserver/commit/e499126b) Prepare for release v0.8.0-rc.0 (#72)
- [cc54e87d](https://github.com/kubedb/mssqlserver/commit/cc54e87d) Update deps (#71)
- [de2304b7](https://github.com/kubedb/mssqlserver/commit/de2304b7) Add Operator Sharding Support; Add IRSA annotation in Sidekick (#70)
- [4bf06b68](https://github.com/kubedb/mssqlserver/commit/4bf06b68) Setup new webhook style; Update to k8s v1.32 (#69)
- [04686b68](https://github.com/kubedb/mssqlserver/commit/04686b68) Don't PublishNotReadyAddresses for non Headless services (#68)
- [b9da2f24](https://github.com/kubedb/mssqlserver/commit/b9da2f24) Install ace-user-roles; Run tests twice a week (#67)
- [b9d4c52b](https://github.com/kubedb/mssqlserver/commit/b9d4c52b) Set default resources on the walg container



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.7.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.7.0)




## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.7.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.7.0)

- [35c239b](https://github.com/kubedb/mssqlserver-walg-plugin/commit/35c239b) Prepare for release v0.7.0 (#23)
- [86c049a](https://github.com/kubedb/mssqlserver-walg-plugin/commit/86c049a) Prepare for release v0.7.0-rc.1 (#22)
- [2d7f303](https://github.com/kubedb/mssqlserver-walg-plugin/commit/2d7f303) Prepare for release v0.7.0-rc.0 (#21)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.46.0](https://github.com/kubedb/mysql/releases/tag/v0.46.0)

- [7fd4cf1a](https://github.com/kubedb/mysql/commit/7fd4cf1a3) Prepare for release v0.46.0 (#681)
- [f94b3045](https://github.com/kubedb/mysql/commit/f94b30458) Cleanup restoresession reconciler (#679)
- [6a5b7dec](https://github.com/kubedb/mysql/commit/6a5b7deca) Configure restore session controller name
- [7d3e6468](https://github.com/kubedb/mysql/commit/7d3e64685) Prepare for release v0.46.0-rc.1 (#678)
- [aa15539f](https://github.com/kubedb/mysql/commit/aa15539f5) Prepare for release v0.46.0-rc.0 (#677)
- [7e9aa056](https://github.com/kubedb/mysql/commit/7e9aa0564) Restructure admission webhook (#676)
- [04e0c7f0](https://github.com/kubedb/mysql/commit/04e0c7f07) Install ace-user-roles; Run tests twice a week (#674)
- [d107c361](https://github.com/kubedb/mysql/commit/d107c3614) Update to k8s v32 (#675)
- [8f36148e](https://github.com/kubedb/mysql/commit/8f36148eb) Lower Sidekick Default Resource (#673)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.14.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.14.0)

- [0226646b](https://github.com/kubedb/mysql-archiver/commit/0226646b) Prepare for release v0.14.0 (#57)
- [ffeaa873](https://github.com/kubedb/mysql-archiver/commit/ffeaa873) Prepare for release v0.14.0-rc.1 (#56)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.31.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.31.0)

- [fecf1748](https://github.com/kubedb/mysql-coordinator/commit/fecf1748) Prepare for release v0.31.0 (#140)
- [d4bc9ecb](https://github.com/kubedb/mysql-coordinator/commit/d4bc9ecb) Prepare for release v0.31.0-rc.1 (#139)
- [8750fb07](https://github.com/kubedb/mysql-coordinator/commit/8750fb07) Prepare for release v0.31.0-rc.0 (#138)
- [026cf350](https://github.com/kubedb/mysql-coordinator/commit/026cf350) Fix TLS Handshake Issue for MySQL 5.7.x (#137)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.14.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.14.0)

- [9070f090](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/9070f090) Prepare for release v0.14.0 (#45)
- [1215a3d1](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/1215a3d1) Prepare for release v0.14.0-rc.1 (#44)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.16.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.16.0)

- [24a5160](https://github.com/kubedb/mysql-restic-plugin/commit/24a5160) Prepare for release v0.16.0 (#72)
- [a1a92a6](https://github.com/kubedb/mysql-restic-plugin/commit/a1a92a6) Prepare for release v0.16.0-rc.1 (#71)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.31.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.31.0)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.40.0](https://github.com/kubedb/ops-manager/releases/tag/v0.40.0)




## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.40.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.40.0)

- [5a3c9251](https://github.com/kubedb/percona-xtradb/commit/5a3c92513) Prepare for release v0.40.0 (#405)
- [b9354d4d](https://github.com/kubedb/percona-xtradb/commit/b9354d4d7) Prepare for release v0.40.0-rc.1 (#404)
- [3cfd53c5](https://github.com/kubedb/percona-xtradb/commit/3cfd53c55) Restructure admission webhook (#403)
- [d6174a69](https://github.com/kubedb/percona-xtradb/commit/d6174a69c) Add operator-sharding; Install ace-user-roles (#401)
- [fd5c0af1](https://github.com/kubedb/percona-xtradb/commit/fd5c0af1f) Update to k8s v32 (#402)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.26.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.26.0)

- [c8c2b7f0](https://github.com/kubedb/percona-xtradb-coordinator/commit/c8c2b7f0) Prepare for release v0.26.0 (#94)
- [20479694](https://github.com/kubedb/percona-xtradb-coordinator/commit/20479694) Prepare for release v0.26.0-rc.1 (#93)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.37.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.37.0)

- [7edc6d33](https://github.com/kubedb/pg-coordinator/commit/7edc6d33) Prepare for release v0.37.0 (#198)
- [f676b187](https://github.com/kubedb/pg-coordinator/commit/f676b187) Prepare for release v0.37.0-rc.1 (#195)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.40.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.40.0)

- [c74520b2](https://github.com/kubedb/pgbouncer/commit/c74520b2) Prepare for release v0.40.0 (#369)
- [45eac53c](https://github.com/kubedb/pgbouncer/commit/45eac53c) Prepare for release v0.40.0-rc.1 (#368)
- [eb30f5ef](https://github.com/kubedb/pgbouncer/commit/eb30f5ef) Prepare for release v0.40.0-rc.0 (#367)
- [60375a15](https://github.com/kubedb/pgbouncer/commit/60375a15) Update deps
- [39fe5cea](https://github.com/kubedb/pgbouncer/commit/39fe5cea) Restructure admission webhook (#366)
- [5d7615b6](https://github.com/kubedb/pgbouncer/commit/5d7615b6) Install ace-user-roles; Run tests twice a week & operator sharding (#364)
- [dee41bca](https://github.com/kubedb/pgbouncer/commit/dee41bca) Update to k8s v32 (#365)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.8.0](https://github.com/kubedb/pgpool/releases/tag/v0.8.0)

- [4faa72dd](https://github.com/kubedb/pgpool/commit/4faa72dd) Prepare for release v0.8.0 (#72)
- [23eb8dbb](https://github.com/kubedb/pgpool/commit/23eb8dbb) Prepare for release v0.8.0-rc.1 (#71)
- [f6639b78](https://github.com/kubedb/pgpool/commit/f6639b78) Prepare for release v0.8.0-rc.0 (#70)
- [683a8a1f](https://github.com/kubedb/pgpool/commit/683a8a1f) update webhook path (#69)
- [f8a989de](https://github.com/kubedb/pgpool/commit/f8a989de) Implement operator-sharding; Install ace-user-roles (#67)
- [6aee4ccc](https://github.com/kubedb/pgpool/commit/6aee4ccc) Setup new webhook style; Update to k8s v1.32 (#68)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.53.0](https://github.com/kubedb/postgres/releases/tag/v0.53.0)

- [cc5a9f3d](https://github.com/kubedb/postgres/commit/cc5a9f3df) Prepare for release v0.53.0 (#808)
- [c9947237](https://github.com/kubedb/postgres/commit/c99472374) Update Restoresession deps (#807)
- [87c23963](https://github.com/kubedb/postgres/commit/87c239632) Configure restore session controller name
- [79de1743](https://github.com/kubedb/postgres/commit/79de17436) Prepare for release v0.53.0-rc.1 (#806)
- [d794ce54](https://github.com/kubedb/postgres/commit/d794ce54c) Prepare for release v0.53.0-rc.0 (#805)
- [4c1923e2](https://github.com/kubedb/postgres/commit/4c1923e25) Refactor Postgres Webhook (#804)
- [4d1e1989](https://github.com/kubedb/postgres/commit/4d1e19894) Integrate virtual-secret in postgres (#799)
- [e22f77c4](https://github.com/kubedb/postgres/commit/e22f77c42) Update IRSA annotations (#802)
- [bc42a794](https://github.com/kubedb/postgres/commit/bc42a7947) Update Postgres With ShardConfiguration Related Changes (#800)
- [aca0e41e](https://github.com/kubedb/postgres/commit/aca0e41e7) support for kubernetes v1.32.x (#801)
- [3bdacec0](https://github.com/kubedb/postgres/commit/3bdacec07) Run Provisioner Tests On Monday + Friday, Fix Makefile (#797)
- [a9165b82](https://github.com/kubedb/postgres/commit/a9165b828) Lower Sidekick Default Resource (#798)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.14.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.14.0)

- [50b71c6c](https://github.com/kubedb/postgres-archiver/commit/50b71c6c) Prepare for release v0.14.0 (#59)
- [96f5fde6](https://github.com/kubedb/postgres-archiver/commit/96f5fde6) Prepare for release v0.14.0-rc.1 (#58)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.14.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.14.0)

- [3810d222](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/3810d222) Prepare for release v0.14.0 (#55)
- [c910645a](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/c910645a) Prepare for release v0.14.0-rc.1 (#54)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.16.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.16.0)

- [80232ae](https://github.com/kubedb/postgres-restic-plugin/commit/80232ae) Prepare for release v0.16.0 (#69)
- [a047ea7](https://github.com/kubedb/postgres-restic-plugin/commit/a047ea7) Prepare for release v0.16.0-rc.1 (#68)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.53.0](https://github.com/kubedb/provisioner/releases/tag/v0.53.0)




## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.40.0](https://github.com/kubedb/proxysql/releases/tag/v0.40.0)

- [46f7f34f](https://github.com/kubedb/proxysql/commit/46f7f34f7) Prepare for release v0.40.0 (#390)
- [7f5bb28a](https://github.com/kubedb/proxysql/commit/7f5bb28aa) Prepare for release v0.40.0-rc.1 (#389)
- [12932712](https://github.com/kubedb/proxysql/commit/12932712e) Prepare for release v0.40.0-rc.0 (#388)
- [ca0c5be1](https://github.com/kubedb/proxysql/commit/ca0c5be1c) Update deps
- [f20352b2](https://github.com/kubedb/proxysql/commit/f20352b2e) Restructure admission webhook (#387)
- [d39f73d5](https://github.com/kubedb/proxysql/commit/d39f73d59) Add Operator Sharding Support (#385)
- [37380c51](https://github.com/kubedb/proxysql/commit/37380c51e) Update to k8s v32 (#384)
- [27edacd7](https://github.com/kubedb/proxysql/commit/27edacd70) Fix TLS Handshake Issue for MySQL 5.7.x (#382)
- [e6ae715f](https://github.com/kubedb/proxysql/commit/e6ae715fe) Install ace-user-roles; Run tests twice a week;Don't ins all CRDs (#381)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.8.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.8.0)

- [a79d50ed](https://github.com/kubedb/rabbitmq/commit/a79d50ed) Prepare for release v0.8.0 (#81)
- [0bdf2587](https://github.com/kubedb/rabbitmq/commit/0bdf2587) Update Webhook Setup (#80)
- [74769b6b](https://github.com/kubedb/rabbitmq/commit/74769b6b) Prepare for release v0.8.0-rc.1 (#79)
- [271e6a67](https://github.com/kubedb/rabbitmq/commit/271e6a67) Prepare for release v0.8.0-rc.0 (#78)
- [b07d95d7](https://github.com/kubedb/rabbitmq/commit/b07d95d7) Update webhook dependency (#77)
- [e93d9ac2](https://github.com/kubedb/rabbitmq/commit/e93d9ac2) Add Operator Sharding Support For RabbitMQ (#75)
- [a5da5f57](https://github.com/kubedb/rabbitmq/commit/a5da5f57) Setup new webhook style; Update to k8s v1.32 (#76)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.46.0](https://github.com/kubedb/redis/releases/tag/v0.46.0)

- [2f94dfdf](https://github.com/kubedb/redis/commit/2f94dfdf4) Prepare for release v0.46.0 (#590)
- [4d36462e](https://github.com/kubedb/redis/commit/4d36462ee) Cleanup restoresession reconciler
- [9d66d6d5](https://github.com/kubedb/redis/commit/9d66d6d5b) Configure restore session controller name
- [0385fa3c](https://github.com/kubedb/redis/commit/0385fa3cc) Prepare for release v0.46.0-rc.1 (#589)
- [abbcc896](https://github.com/kubedb/redis/commit/abbcc8962) Prepare for release v0.46.0-rc.0 (#588)
- [7f66d703](https://github.com/kubedb/redis/commit/7f66d703b) Restructure admission webhook (#587)
- [5bcfeff7](https://github.com/kubedb/redis/commit/5bcfeff71) Add Operator Sharding Support (#586)
- [03f9ebd5](https://github.com/kubedb/redis/commit/03f9ebd54) Update to k8s v32 (#585)
- [b5a67c90](https://github.com/kubedb/redis/commit/b5a67c907) Install ace-user-roles; Run tests twice a week (#583)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.32.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.32.0)

- [37e763bc](https://github.com/kubedb/redis-coordinator/commit/37e763bc) Prepare for release v0.32.0 (#126)
- [0b85ddde](https://github.com/kubedb/redis-coordinator/commit/0b85ddde) Prepare for release v0.32.0-rc.1 (#125)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.16.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.16.0)

- [5d59ea6](https://github.com/kubedb/redis-restic-plugin/commit/5d59ea6) Prepare for release v0.16.0 (#64)
- [a535e54](https://github.com/kubedb/redis-restic-plugin/commit/a535e54) Prepare for release v0.16.0-rc.1 (#63)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.40.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.40.0)

- [8c13a5ab](https://github.com/kubedb/replication-mode-detector/commit/8c13a5ab) Prepare for release v0.40.0 (#291)
- [e7943024](https://github.com/kubedb/replication-mode-detector/commit/e7943024) Prepare for release v0.40.0-rc.1 (#290)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.29.0](https://github.com/kubedb/schema-manager/releases/tag/v0.29.0)

- [77cdc6c5](https://github.com/kubedb/schema-manager/commit/77cdc6c5) Prepare for release v0.29.0 (#137)
- [b7094850](https://github.com/kubedb/schema-manager/commit/b7094850) Prepare for release v0.29.0-rc.1 (#136)
- [eccaede6](https://github.com/kubedb/schema-manager/commit/eccaede6) Prepare for release v0.29.0-rc.0 (#135)
- [a328302c](https://github.com/kubedb/schema-manager/commit/a328302c) Update deps
- [46d62873](https://github.com/kubedb/schema-manager/commit/46d62873) k8s 1.32 and webhook related changes (#134)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.8.0](https://github.com/kubedb/singlestore/releases/tag/v0.8.0)

- [f190f8ac](https://github.com/kubedb/singlestore/commit/f190f8ac) Prepare for release v0.8.0 (#69)
- [66f1745a](https://github.com/kubedb/singlestore/commit/66f1745a) Cleanup restoresession reconciler
- [7fecb62a](https://github.com/kubedb/singlestore/commit/7fecb62a) Configure restore session controller name
- [0272fc4d](https://github.com/kubedb/singlestore/commit/0272fc4d) Prepare for release v0.8.0-rc.1 (#68)
- [78123100](https://github.com/kubedb/singlestore/commit/78123100) Prepare for release v0.8.0-rc.0 (#67)
- [4b4ca182](https://github.com/kubedb/singlestore/commit/4b4ca182) update deps (#66)
- [52126e70](https://github.com/kubedb/singlestore/commit/52126e70) Install ace-user-roles and Add Operator Sharding (#64)
- [9d9e89a6](https://github.com/kubedb/singlestore/commit/9d9e89a6) Update to k8s v1.32 (#65)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.8.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.8.0)

- [c18ec07](https://github.com/kubedb/singlestore-coordinator/commit/c18ec07) Prepare for release v0.8.0 (#41)
- [821d775](https://github.com/kubedb/singlestore-coordinator/commit/821d775) Prepare for release v0.8.0-rc.1 (#40)
- [cfedcc0](https://github.com/kubedb/singlestore-coordinator/commit/cfedcc0) Prepare for release v0.8.0-rc.0 (#39)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.11.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.11.0)

- [35285d9](https://github.com/kubedb/singlestore-restic-plugin/commit/35285d9) Prepare for release v0.11.0 (#41)
- [ae4805f](https://github.com/kubedb/singlestore-restic-plugin/commit/ae4805f) Prepare for release v0.11.0-rc.1 (#40)
- [e939734](https://github.com/kubedb/singlestore-restic-plugin/commit/e939734) Prepare for release v0.11.0-rc.0 (#39)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.8.0](https://github.com/kubedb/solr/releases/tag/v0.8.0)

- [4d35f4c4](https://github.com/kubedb/solr/commit/4d35f4c4) Prepare for release v0.8.0 (#82)
- [f6f0d2db](https://github.com/kubedb/solr/commit/f6f0d2db) Fix webhook configs (#81)
- [3598e015](https://github.com/kubedb/solr/commit/3598e015) create petset after zookeeper ready (#80)
- [a2f3abd6](https://github.com/kubedb/solr/commit/a2f3abd6) Prepare for release v0.8.0-rc.1 (#79)
- [a2a12869](https://github.com/kubedb/solr/commit/a2a12869) Prepare for release v0.8.0-rc.0 (#78)
- [a68d5e84](https://github.com/kubedb/solr/commit/a68d5e84) Fix dependency for webhook (#77)
- [2322338c](https://github.com/kubedb/solr/commit/2322338c) Add operator sharding support for solr (#76)
- [792ceb13](https://github.com/kubedb/solr/commit/792ceb13) Webhook changes for k8s 1.32 (#75)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.38.0](https://github.com/kubedb/tests/releases/tag/v0.38.0)

- [7be6ac4d](https://github.com/kubedb/tests/commit/7be6ac4d) Prepare for release v0.38.0 (#453)
- [ca56e5d1](https://github.com/kubedb/tests/commit/ca56e5d1) Remove externally managed backend from ferretdb (#452)
- [8edab841](https://github.com/kubedb/tests/commit/8edab841) Divide MariaDB Test Profiles (#446)
- [b4e5e71b](https://github.com/kubedb/tests/commit/b4e5e71b) Divide Test Profiles for Postgres (#447)
- [065abf8a](https://github.com/kubedb/tests/commit/065abf8a) Prepare for release v0.38.0-rc.1 (#451)
- [d3aec073](https://github.com/kubedb/tests/commit/d3aec073) Update go.mod with k8s v0.32.2 (#449)
- [7181a2ab](https://github.com/kubedb/tests/commit/7181a2ab) Fix MongoDB restic backup CI (#438)
- [8780cccd](https://github.com/kubedb/tests/commit/8780cccd) MongoDB Test Profile Divide (#442)
- [1576dbf8](https://github.com/kubedb/tests/commit/1576dbf8) rabbitmq test log fix (#445)
- [86e04cc1](https://github.com/kubedb/tests/commit/86e04cc1) Test against k8s 1.32 (#444)
- [ac643c07](https://github.com/kubedb/tests/commit/ac643c07) Add Postgres CI (restic-backup) (#434)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.29.0](https://github.com/kubedb/ui-server/releases/tag/v0.29.0)

- [e5d9a3ec](https://github.com/kubedb/ui-server/commit/e5d9a3ec) Prepare for release v0.29.0 (#157)
- [683a6c9d](https://github.com/kubedb/ui-server/commit/683a6c9d) Prepare for release v0.29.0-rc.1 (#156)
- [8b3fbf94](https://github.com/kubedb/ui-server/commit/8b3fbf94) Prepare for release v0.29.0-rc.0 (#155)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.29.0](https://github.com/kubedb/webhook-server/releases/tag/v0.29.0)

- [8988f1e4](https://github.com/kubedb/webhook-server/commit/8988f1e4) Prepare for release v0.29.0 (#152)
- [29ca05cc](https://github.com/kubedb/webhook-server/commit/29ca05cc) Fix for sl,md (#151)
- [cf06ba55](https://github.com/kubedb/webhook-server/commit/cf06ba55) Set PerconaXtraDB Default Replica to 3 (#150)
- [bc3f46fa](https://github.com/kubedb/webhook-server/commit/bc3f46fa) Use port 443 instead of 9443
- [32299e12](https://github.com/kubedb/webhook-server/commit/32299e12) Prepare for release v0.29.0-rc.1 (#149)
- [a18927dc](https://github.com/kubedb/webhook-server/commit/a18927dc) Remove old-db & ops-manager deps for webhooks (#148)
- [4eab5c8e](https://github.com/kubedb/webhook-server/commit/4eab5c8e) Rewrite webhook calls; Update to k8s v32 (#147)



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.2.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.2.0)




## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.8.0](https://github.com/kubedb/zookeeper/releases/tag/v0.8.0)

- [3b71bca3](https://github.com/kubedb/zookeeper/commit/3b71bca3) Prepare for release v0.8.0 (#72)
- [66198a06](https://github.com/kubedb/zookeeper/commit/66198a06) Prepare for release v0.8.0-rc.1 (#70)
- [e26b5b6b](https://github.com/kubedb/zookeeper/commit/e26b5b6b) Prepare for release v0.8.0-rc.0 (#69)
- [30850f54](https://github.com/kubedb/zookeeper/commit/30850f54) update webhook path (#68)
- [aef92815](https://github.com/kubedb/zookeeper/commit/aef92815) add operator sharding support (#67)
- [9fe3244d](https://github.com/kubedb/zookeeper/commit/9fe3244d) Update to k8s v1.32 (#66)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.9.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.9.0)

- [8a45cbd](https://github.com/kubedb/zookeeper-restic-plugin/commit/8a45cbd) Prepare for release v0.9.0 (#33)
- [3d99e28](https://github.com/kubedb/zookeeper-restic-plugin/commit/3d99e28) Prepare for release v0.9.0-rc.1 (#32)
- [deb7075](https://github.com/kubedb/zookeeper-restic-plugin/commit/deb7075) Prepare for release v0.9.0-rc.0 (#31)




