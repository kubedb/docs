---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2026.2.16-rc.0
    name: Changelog-v2026.2.16-rc.0
    parent: welcome
    weight: 20260216
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.2.16-rc.0/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.2.16-rc.0/
---

# KubeDB v2026.2.16-rc.0 (2026-02-18)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.61.0-rc.0](https://github.com/kubedb/apimachinery/releases/tag/v0.61.0-rc.0)

- [78da1b2d](https://github.com/kubedb/apimachinery/commit/78da1b2d6) Redis Config Secret Conversion fix (#1595)
- [f1ac6610](https://github.com/kubedb/apimachinery/commit/f1ac6610d) Standardize migrator pkg (#1594)
- [8c54988a](https://github.com/kubedb/apimachinery/commit/8c54988a4) Report k8s distro and storage provisioners (#1593)
- [3ab9a8c4](https://github.com/kubedb/apimachinery/commit/3ab9a8c46) Fix CI (#1592)
- [46868bd0](https://github.com/kubedb/apimachinery/commit/46868bd03) Add Distributed Archiver Support for MariaDB (#1582)
- [d390f703](https://github.com/kubedb/apimachinery/commit/d390f7030) Add Migrator Operator APIs and Configs for CLI (#1588)
- [71146793](https://github.com/kubedb/apimachinery/commit/711467936) Add Qdrant Monitoring Support (#1550)
- [8591d06f](https://github.com/kubedb/apimachinery/commit/8591d06f2) Add read replica support for Postgres (#1580)
- [09750ca1](https://github.com/kubedb/apimachinery/commit/09750ca1e) Add Milvus Cluster Apis (#1551)
- [eec0c01f](https://github.com/kubedb/apimachinery/commit/eec0c01f2) Introduce log RetentionPeriod for log cleanup (#1537)
- [05e2788c](https://github.com/kubedb/apimachinery/commit/05e2788ca) Add Kafka Tiered Storage APIs (#1378)
- [719ce3b0](https://github.com/kubedb/apimachinery/commit/719ce3b05) Update license header (#1591)
- [04bd7de6](https://github.com/kubedb/apimachinery/commit/04bd7de62) Add Constant For Opensearch/Elasticsearch (#1578)
- [7eb9e9a4](https://github.com/kubedb/apimachinery/commit/7eb9e9a4c) Add tolerations to pvc-mounter pod for storageclass migration (#1589)
- [81a671f5](https://github.com/kubedb/apimachinery/commit/81a671f52) Add relabel config support to ServiceMonitor (#1584)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.46.0-rc.0](https://github.com/kubedb/autoscaler/releases/tag/v0.46.0-rc.0)

- [381bc609](https://github.com/kubedb/autoscaler/commit/381bc609) Prepare for release v0.46.0-rc.0 (#278)
- [b5c6e303](https://github.com/kubedb/autoscaler/commit/b5c6e303) Merge pull request #277 from kubedb/update-deps
- [7e9e4708](https://github.com/kubedb/autoscaler/commit/7e9e4708) Prepare for release v0.45.0 (#274)
- [aadf7a1b](https://github.com/kubedb/autoscaler/commit/aadf7a1b) Update Mongo Reconfig API (#273)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.14.0-rc.0](https://github.com/kubedb/cassandra/releases/tag/v0.14.0-rc.0)

- [c60b9bfd](https://github.com/kubedb/cassandra/commit/c60b9bfd) Prepare for release v0.14.0-rc.0 (#62)
- [0c7fc6da](https://github.com/kubedb/cassandra/commit/0c7fc6da) move parallel-go to apimachinary (#60)
- [b3cd8596](https://github.com/kubedb/cassandra/commit/b3cd8596) Prepare for release v0.13.0 (#58)
- [70d0a9d0](https://github.com/kubedb/cassandra/commit/70d0a9d0) Improve and generalize configure-reconfigure process (#57)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.8.0-rc.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.8.0-rc.0)

- [c4f10d61](https://github.com/kubedb/cassandra-medusa-plugin/commit/c4f10d61) Prepare for release v0.8.0-rc.0 (#23)
- [52d43201](https://github.com/kubedb/cassandra-medusa-plugin/commit/52d43201) Prepare for release v0.7.0 (#21)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.61.0-rc.0](https://github.com/kubedb/cli/releases/tag/v0.61.0-rc.0)

- [ec20e01a](https://github.com/kubedb/cli/commit/ec20e01a3) Prepare for release v0.61.0-rc.0 (#812)
- [d0b02fca](https://github.com/kubedb/cli/commit/d0b02fca7) Prepare for release v0.60.0 (#810)



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.16.0-rc.0](https://github.com/kubedb/clickhouse/releases/tag/v0.16.0-rc.0)

- [64226cf9](https://github.com/kubedb/clickhouse/commit/64226cf9) Prepare for release v0.16.0-rc.0 (#85)
- [2b65eea3](https://github.com/kubedb/clickhouse/commit/2b65eea3) Integrate shouldProceed() utility on runParallel (#81)
- [36b98fc2](https://github.com/kubedb/clickhouse/commit/36b98fc2) Prepare for release v0.15.0 (#80)
- [41fe5c21](https://github.com/kubedb/clickhouse/commit/41fe5c21) Update Configure reconfigure process (#79)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.16.0-rc.0](https://github.com/kubedb/crd-manager/releases/tag/v0.16.0-rc.0)

- [2d2c0c9a](https://github.com/kubedb/crd-manager/commit/2d2c0c9a) Prepare for release v0.16.0-rc.0 (#110)
- [ef87879e](https://github.com/kubedb/crd-manager/commit/ef87879e) Update kafka crds (#108)
- [e9b0cd26](https://github.com/kubedb/crd-manager/commit/e9b0cd26) Prepare for release v0.15.0 (#107)
- [f16d4769](https://github.com/kubedb/crd-manager/commit/f16d4769) Add crd for cassandra, hazelcast, oracle (#106)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.19.0-rc.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.19.0-rc.0)

- [c151b9d](https://github.com/kubedb/dashboard-restic-plugin/commit/c151b9d) Prepare for release v0.19.0-rc.0 (#61)
- [bb05585](https://github.com/kubedb/dashboard-restic-plugin/commit/bb05585) Incorporate changes for restic standalone pkg (#60)
- [c6ab8f5](https://github.com/kubedb/dashboard-restic-plugin/commit/c6ab8f5) Prepare for release v0.18.0 (#58)
- [9ca336a](https://github.com/kubedb/dashboard-restic-plugin/commit/9ca336a) Use forked kubestash/restic (#57)
- [4825ed5](https://github.com/kubedb/dashboard-restic-plugin/commit/4825ed5) Use forked kubestash/restic (#56)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.16.0-rc.0](https://github.com/kubedb/db-client-go/releases/tag/v0.16.0-rc.0)

- [3cbe8746](https://github.com/kubedb/db-client-go/commit/3cbe8746) Prepare for release v0.16.0-rc.0 (#221)
- [496ef203](https://github.com/kubedb/db-client-go/commit/496ef203) Add Creating MariaDB Client without MariaDB Instance (#218)
- [2ecc2b55](https://github.com/kubedb/db-client-go/commit/2ecc2b55) Prepare for release v0.15.0 (#217)
- [9854141c](https://github.com/kubedb/db-client-go/commit/9854141c) fix opensearch-v3 health check issues (#216)
- [d08174b5](https://github.com/kubedb/db-client-go/commit/d08174b5) Add Qdrant TLS Support (#212)



## [kubedb/db2](https://github.com/kubedb/db2)

### [v0.2.0-rc.0](https://github.com/kubedb/db2/releases/tag/v0.2.0-rc.0)

- [083436d3](https://github.com/kubedb/db2/commit/083436d3) Prepare for release v0.2.0-rc.0 (#10)
- [aafe21ff](https://github.com/kubedb/db2/commit/aafe21ff) Prepare for release v0.1.0 (#8)



## [kubedb/db2-coordinator](https://github.com/kubedb/db2-coordinator)

### [v0.2.0-rc.0](https://github.com/kubedb/db2-coordinator/releases/tag/v0.2.0-rc.0)




## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.16.0-rc.0](https://github.com/kubedb/druid/releases/tag/v0.16.0-rc.0)

- [b04c676d](https://github.com/kubedb/druid/commit/b04c676d) Prepare for release v0.16.0-rc.0 (#114)
- [8c8432ef](https://github.com/kubedb/druid/commit/8c8432ef) fix-parallel (#112)
- [34a18a5a](https://github.com/kubedb/druid/commit/34a18a5a) Prepare for release v0.15.0 (#111)
- [6236221f](https://github.com/kubedb/druid/commit/6236221f) Reconfigure redesign (#110)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.61.0-rc.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.61.0-rc.0)

- [ea0990c0](https://github.com/kubedb/elasticsearch/commit/ea0990c0f) Prepare for release v0.61.0-rc.0 (#790)
- [9d920917](https://github.com/kubedb/elasticsearch/commit/9d920917b) Remove init Container from Opensearch (#787)
- [06726a8c](https://github.com/kubedb/elasticsearch/commit/06726a8ce) Merge pull request #788 from kubedb/prll-ctrl
- [bf671dfe](https://github.com/kubedb/elasticsearch/commit/bf671dfe6) Prepare for release v0.60.0 (#786)
- [3d9d54ad](https://github.com/kubedb/elasticsearch/commit/3d9d54ad0) Fix Opensearch health check issue for v3.0 (#785)
- [10542dbb](https://github.com/kubedb/elasticsearch/commit/10542dbb0) Improve and generalize configure-reconfigure process (#782)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.24.0-rc.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.24.0-rc.0)

- [21960b1b](https://github.com/kubedb/elasticsearch-restic-plugin/commit/21960b1b) Prepare for release v0.24.0-rc.0 (#84)
- [21875af7](https://github.com/kubedb/elasticsearch-restic-plugin/commit/21875af7) Incorporate changes for restic standalone pkg (#83)
- [a025d501](https://github.com/kubedb/elasticsearch-restic-plugin/commit/a025d501) Prepare for release v0.23.0 (#81)
- [63215a12](https://github.com/kubedb/elasticsearch-restic-plugin/commit/63215a12) Use forked kubestash/restic (#80)
- [085362c3](https://github.com/kubedb/elasticsearch-restic-plugin/commit/085362c3) Use forked kubestash/restic (#79)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.9.0-rc.0](https://github.com/kubedb/gitops/releases/tag/v0.9.0-rc.0)

- [1e454f0b](https://github.com/kubedb/gitops/commit/1e454f0b) Prepare for release v0.9.0-rc.0 (#39)
- [5948210d](https://github.com/kubedb/gitops/commit/5948210d) Prepare for release v0.8.0 (#37)
- [4dae1f47](https://github.com/kubedb/gitops/commit/4dae1f47) Update Configuration Changes (#36)



## [kubedb/hanadb](https://github.com/kubedb/hanadb)

### [v0.2.0-rc.0](https://github.com/kubedb/hanadb/releases/tag/v0.2.0-rc.0)

- [dd50ccfc](https://github.com/kubedb/hanadb/commit/dd50ccfc) Prepare for release v0.2.0-rc.0 (#15)
- [ae192632](https://github.com/kubedb/hanadb/commit/ae192632) fix auth issue (#12)
- [d29c0289](https://github.com/kubedb/hanadb/commit/d29c0289) Prepare for release v0.1.0 (#11)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.7.0-rc.0](https://github.com/kubedb/hazelcast/releases/tag/v0.7.0-rc.0)

- [8e683f77](https://github.com/kubedb/hazelcast/commit/8e683f77) Prepare for release v0.7.0-rc.0 (#27)
- [158438dc](https://github.com/kubedb/hazelcast/commit/158438dc) Merge pull request #25 from kubedb/proceed
- [03cc9a0b](https://github.com/kubedb/hazelcast/commit/03cc9a0b) Prepare for release v0.6.0 (#24)
- [9dc7cbac](https://github.com/kubedb/hazelcast/commit/9dc7cbac)  Improve and generalize configure-reconfigure process (#23)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.8.0-rc.0](https://github.com/kubedb/ignite/releases/tag/v0.8.0-rc.0)

- [b5e03fa2](https://github.com/kubedb/ignite/commit/b5e03fa2) Prepare for release v0.8.0-rc.0 (#36)
- [2f574436](https://github.com/kubedb/ignite/commit/2f574436) Merge pull request #34 from kubedb/proceed
- [f648ea48](https://github.com/kubedb/ignite/commit/f648ea48) Prepare for release v0.7.0 (#33)
- [50699ff3](https://github.com/kubedb/ignite/commit/50699ff3) Re-design Configuration process (#32)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2026.2.16-rc.0](https://github.com/kubedb/installer/releases/tag/v2026.2.16-rc.0)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.32.0-rc.0](https://github.com/kubedb/kafka/releases/tag/v0.32.0-rc.0)

- [ac302858](https://github.com/kubedb/kafka/commit/ac302858) Prepare for release v0.32.0-rc.0 (#176)
- [848ca394](https://github.com/kubedb/kafka/commit/848ca394) Add Kafka Tiered Storage Support (#130)
- [543ee275](https://github.com/kubedb/kafka/commit/543ee275) Merge pull request #174 from kubedb/proceed
- [3ccdb0fd](https://github.com/kubedb/kafka/commit/3ccdb0fd) Prepare for release v0.31.0 (#173)
- [8806da50](https://github.com/kubedb/kafka/commit/8806da50) Update Reconfigure/Configure (#172)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.37.0-rc.0](https://github.com/kubedb/kibana/releases/tag/v0.37.0-rc.0)

- [9c1f0d74](https://github.com/kubedb/kibana/commit/9c1f0d74) Prepare for release v0.37.0-rc.0 (#169)
- [2a0c9e0c](https://github.com/kubedb/kibana/commit/2a0c9e0c) Remove Init Container From es/os Dashboard (#168)
- [c454080c](https://github.com/kubedb/kibana/commit/c454080c) Prepare for release v0.36.0 (#167)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.24.0-rc.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.24.0-rc.0)

- [b3d1a34e](https://github.com/kubedb/kubedb-manifest-plugin/commit/b3d1a34e) Prepare for release v0.24.0-rc.0 (#116)
- [19fa3a25](https://github.com/kubedb/kubedb-manifest-plugin/commit/19fa3a25) Incorporate changes for restic standalone pkg (#115)
- [9aebbe1c](https://github.com/kubedb/kubedb-manifest-plugin/commit/9aebbe1c) Fix UBI image
- [39aabbc2](https://github.com/kubedb/kubedb-manifest-plugin/commit/39aabbc2) Prepare for release v0.23.0 (#113)
- [6ab7dfbc](https://github.com/kubedb/kubedb-manifest-plugin/commit/6ab7dfbc) Use forked kubestash/restic (#112)
- [839e93d7](https://github.com/kubedb/kubedb-manifest-plugin/commit/839e93d7) Use forked kubestash/restic (#111)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.12.0-rc.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.12.0-rc.0)

- [b07d581c](https://github.com/kubedb/kubedb-verifier/commit/b07d581c) Prepare for release v0.12.0-rc.0 (#38)
- [fc8b0048](https://github.com/kubedb/kubedb-verifier/commit/fc8b0048) Prepare for release v0.11.0 (#36)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.45.0-rc.0](https://github.com/kubedb/mariadb/releases/tag/v0.45.0-rc.0)

- [58f3723c](https://github.com/kubedb/mariadb/commit/58f3723cf) Prepare for release v0.45.0-rc.0 (#372)
- [106a9729](https://github.com/kubedb/mariadb/commit/106a97299) Distributed archiver support (#359)
- [f7a26a38](https://github.com/kubedb/mariadb/commit/f7a26a38d) Update for KubeStash API (#370)
- [1fdc4f97](https://github.com/kubedb/mariadb/commit/1fdc4f979) Update ops parallelism (#367)
- [b5e215a5](https://github.com/kubedb/mariadb/commit/b5e215a53) Prepare for release v0.44.0 (#366)
- [2b103c38](https://github.com/kubedb/mariadb/commit/2b103c386) Use Exporter Version v0.18.0 (#365)
- [0cfc4998](https://github.com/kubedb/mariadb/commit/0cfc49986) Improve and generalize configure-reconfigure (#363)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.21.0-rc.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.21.0-rc.0)

- [95ecbd4d](https://github.com/kubedb/mariadb-archiver/commit/95ecbd4d) Prepare for release v0.21.0-rc.0 (#75)
- [b6535426](https://github.com/kubedb/mariadb-archiver/commit/b6535426) Add Distributed Support (#65)
- [d11b907e](https://github.com/kubedb/mariadb-archiver/commit/d11b907e) Fix ubi build (#73)
- [40710dc0](https://github.com/kubedb/mariadb-archiver/commit/40710dc0) Prepare for release v0.20.0 (#72)
- [ce027593](https://github.com/kubedb/mariadb-archiver/commit/ce027593) Fix redhat catalog submission (#70)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.41.0-rc.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.41.0-rc.0)

- [b1ee0463](https://github.com/kubedb/mariadb-coordinator/commit/b1ee0463) Prepare for release v0.41.0-rc.0 (#163)
- [80432fdc](https://github.com/kubedb/mariadb-coordinator/commit/80432fdc) Distributed archiver support (#153)
- [08b426e5](https://github.com/kubedb/mariadb-coordinator/commit/08b426e5) Prepare for release v0.40.0 (#161)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.21.0-rc.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.21.0-rc.0)

- [80fb08c4](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/80fb08c4) Prepare for release v0.21.0-rc.0 (#67)
- [16d186de](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/16d186de) Add Distributed Support (#66)
- [26607e35](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/26607e35) Prepare for release v0.20.0 (#65)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.19.0-rc.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.19.0-rc.0)

- [0d5eb53](https://github.com/kubedb/mariadb-restic-plugin/commit/0d5eb53) Prepare for release v0.19.0-rc.0 (#73)
- [15c95b9](https://github.com/kubedb/mariadb-restic-plugin/commit/15c95b9) Incorporate changes for restic standalone pkg (#72)
- [9c46acc](https://github.com/kubedb/mariadb-restic-plugin/commit/9c46acc) Prepare for release v0.18.0 (#70)
- [0c011f7](https://github.com/kubedb/mariadb-restic-plugin/commit/0c011f7) Use forked kubestash/restic (#69)
- [7403f2d](https://github.com/kubedb/mariadb-restic-plugin/commit/7403f2d) Use forked kubestash/restic (#68)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.54.0-rc.0](https://github.com/kubedb/memcached/releases/tag/v0.54.0-rc.0)

- [f935bd5c](https://github.com/kubedb/memcached/commit/f935bd5cf) Prepare for release v0.54.0-rc.0 (#522)
- [9750dc58](https://github.com/kubedb/memcached/commit/9750dc583) Merge pull request #520 from kubedb/upd-prll
- [d9c49db8](https://github.com/kubedb/memcached/commit/d9c49db82) Prepare for release v0.53.0 (#519)
- [25b0aa41](https://github.com/kubedb/memcached/commit/25b0aa419) Update memcached reconfigure (#518)



## [kubedb/migrator-cli](https://github.com/kubedb/migrator-cli)

### [v0.1.0-rc.0](https://github.com/kubedb/migrator-cli/releases/tag/v0.1.0-rc.0)

- [6e3085b](https://github.com/kubedb/migrator-cli/commit/6e3085b) Prepare for release v0.1.0-rc.0 (#2)
- [4120491](https://github.com/kubedb/migrator-cli/commit/4120491) initial setup, integrate pg_dump and logical replication for postgres (#1)



## [kubedb/migrator-operator](https://github.com/kubedb/migrator-operator)

### [v0.1.0-rc.0](https://github.com/kubedb/migrator-operator/releases/tag/v0.1.0-rc.0)

- [41e08f9](https://github.com/kubedb/migrator-operator/commit/41e08f9) Fix CI (#5)
- [0355020](https://github.com/kubedb/migrator-operator/commit/0355020) Prepare for release v0.1.0-rc.0 (#4)
- [47b051e](https://github.com/kubedb/migrator-operator/commit/47b051e) Rewrite utils pkg to make it db agnostic (#2)
- [7e38d57](https://github.com/kubedb/migrator-operator/commit/7e38d57) Init operator for database migration to kubedb



## [kubedb/milvus](https://github.com/kubedb/milvus)

### [v0.2.0-rc.0](https://github.com/kubedb/milvus/releases/tag/v0.2.0-rc.0)

- [06366bab](https://github.com/kubedb/milvus/commit/06366bab) Prepare for release v0.2.0-rc.0 (#17)
- [b0081352](https://github.com/kubedb/milvus/commit/b0081352) Remove Dockerfile (#16)
- [8a73c4f7](https://github.com/kubedb/milvus/commit/8a73c4f7) Add Milvus Cluster (#12)
- [f801e32b](https://github.com/kubedb/milvus/commit/f801e32b) Fix Standalone Milvus Phase Update (#13)
- [6a714e39](https://github.com/kubedb/milvus/commit/6a714e39) Prepare for release v0.1.0 (#11)
- [e5782e5c](https://github.com/kubedb/milvus/commit/e5782e5c) Update Configuration process (#10)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.54.0-rc.0](https://github.com/kubedb/mongodb/releases/tag/v0.54.0-rc.0)

- [29774d97](https://github.com/kubedb/mongodb/commit/29774d97b) Prepare for release v0.54.0-rc.0 (#736)
- [26c98778](https://github.com/kubedb/mongodb/commit/26c98778d) update for kubestash api (#735)
- [aba284ca](https://github.com/kubedb/mongodb/commit/aba284ca0) Incorporate kubestash labels on snapshot (#733)
- [c440ff53](https://github.com/kubedb/mongodb/commit/c440ff53f) Merge pull request #731 from kubedb/proceed
- [5f538c7c](https://github.com/kubedb/mongodb/commit/5f538c7c6) Fix empty object issue on removeCustomConfig (#730)
- [4a9999a2](https://github.com/kubedb/mongodb/commit/4a9999a24) Prepare for release v0.53.0 (#729)
- [06a6f687](https://github.com/kubedb/mongodb/commit/06a6f687d) Re-design Configure reconfigure process (#726)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.22.0-rc.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.22.0-rc.0)

- [971d799d](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/971d799d) Prepare for release v0.22.0-rc.0 (#70)
- [e6fd5d5f](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/e6fd5d5f) Prepare for release v0.21.0 (#68)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.24.0-rc.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.24.0-rc.0)

- [b9e51294](https://github.com/kubedb/mongodb-restic-plugin/commit/b9e51294) Prepare for release v0.24.0-rc.0 (#106)
- [c1ba1aea](https://github.com/kubedb/mongodb-restic-plugin/commit/c1ba1aea) Incorporate changes for restic standalone pkg (#105)
- [3d9e87b1](https://github.com/kubedb/mongodb-restic-plugin/commit/3d9e87b1) Fix container build target
- [c02b2e73](https://github.com/kubedb/mongodb-restic-plugin/commit/c02b2e73) Fix for kubestash/restic (#103)
- [9ebb8b53](https://github.com/kubedb/mongodb-restic-plugin/commit/9ebb8b53) Prepare for release v0.23.0 (#102)
- [34d434d2](https://github.com/kubedb/mongodb-restic-plugin/commit/34d434d2) Use forked kubestash/restic (#100)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.16.0-rc.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.16.0-rc.0)

- [1bd94124](https://github.com/kubedb/mssql-coordinator/commit/1bd94124) Prepare for release v0.16.0-rc.0 (#55)
- [cbb8df3b](https://github.com/kubedb/mssql-coordinator/commit/cbb8df3b) Prepare for release v0.15.0 (#53)
- [a9d3248b](https://github.com/kubedb/mssql-coordinator/commit/a9d3248b) Correctly check if sqlservr process is running for new versions (#52)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.16.0-rc.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.16.0-rc.0)

- [b11cee95](https://github.com/kubedb/mssqlserver/commit/b11cee95) Prepare for release v0.16.0-rc.0 (#109)
- [ca90875f](https://github.com/kubedb/mssqlserver/commit/ca90875f) Update for KubeStash API (#108)
- [da9faff4](https://github.com/kubedb/mssqlserver/commit/da9faff4) Integrate shouldProceed() utility on runParallel (#105)
- [6567d1a5](https://github.com/kubedb/mssqlserver/commit/6567d1a5) Prepare for release v0.15.0 (#103)
- [2f64f6a0](https://github.com/kubedb/mssqlserver/commit/2f64f6a0) Improve and generalize configure-reconfigure process (#100)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.15.0-rc.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.15.0-rc.0)

- [720ce9b](https://github.com/kubedb/mssqlserver-archiver/commit/720ce9b) Update release wf (#21)



## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.15.0-rc.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.15.0-rc.0)

- [5d1c8e3](https://github.com/kubedb/mssqlserver-walg-plugin/commit/5d1c8e3) Prepare for release v0.15.0-rc.0 (#45)
- [91b8945](https://github.com/kubedb/mssqlserver-walg-plugin/commit/91b8945) Prepare for release v0.14.0 (#43)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.54.0-rc.0](https://github.com/kubedb/mysql/releases/tag/v0.54.0-rc.0)

- [05e392f5](https://github.com/kubedb/mysql/commit/05e392f58) Prepare for release v0.54.0-rc.0 (#722)
- [95604806](https://github.com/kubedb/mysql/commit/95604806d) added permission for binlog-cleanup (#719)
- [8959e3ea](https://github.com/kubedb/mysql/commit/8959e3ea5) Update for Kubestash API (#721)
- [865c8dd0](https://github.com/kubedb/mysql/commit/865c8dd0e) Add pod toleration during storageClass migration (#718)
- [5fed4bfc](https://github.com/kubedb/mysql/commit/5fed4bfc7) Update ops parallelism (#716)
- [62650d63](https://github.com/kubedb/mysql/commit/62650d638) Upgrade Exporter to v0.18.0 (#715)
- [55900db8](https://github.com/kubedb/mysql/commit/55900db8b) Prepare for release v0.53.0 (#714)
- [0a9fd37d](https://github.com/kubedb/mysql/commit/0a9fd37d4) Add Reconfig PodRestart based on User Requirement (#713)
- [b74becf0](https://github.com/kubedb/mysql/commit/b74becf00) Improve and generalize configure-reconfigure (#711)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.22.0-rc.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.22.0-rc.0)

- [2703fb78](https://github.com/kubedb/mysql-archiver/commit/2703fb78) Prepare for release v0.22.0-rc.0 (#81)
- [77cca4ad](https://github.com/kubedb/mysql-archiver/commit/77cca4ad) Clean up old binlogs (#75)
- [4b0fcbdd](https://github.com/kubedb/mysql-archiver/commit/4b0fcbdd) Remove ubi related stuffs (#79)
- [cc257ddb](https://github.com/kubedb/mysql-archiver/commit/cc257ddb) Update makefile for ubi (#68)
- [2e7232d0](https://github.com/kubedb/mysql-archiver/commit/2e7232d0) Fix ubi build (#78)
- [6b3becf0](https://github.com/kubedb/mysql-archiver/commit/6b3becf0) Prepare for release v0.21.0 (#77)
- [9a8fa9f6](https://github.com/kubedb/mysql-archiver/commit/9a8fa9f6) Submit to red hat catalog (#74)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.39.0-rc.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.39.0-rc.0)

- [926a94b4](https://github.com/kubedb/mysql-coordinator/commit/926a94b4) Prepare for release v0.39.0-rc.0 (#162)
- [64575b6f](https://github.com/kubedb/mysql-coordinator/commit/64575b6f) Prepare for release v0.38.0 (#160)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.22.0-rc.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.22.0-rc.0)

- [698c78a6](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/698c78a6) Prepare for release v0.22.0-rc.0 (#67)
- [be4af0c4](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/be4af0c4) Prepare for release v0.21.0 (#65)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.24.0-rc.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.24.0-rc.0)

- [c4651c69](https://github.com/kubedb/mysql-restic-plugin/commit/c4651c69) Prepare for release v0.24.0-rc.0 (#96)
- [09eb2747](https://github.com/kubedb/mysql-restic-plugin/commit/09eb2747) Incorporate changes for restic standalone pkg (#95)
- [29ef1ef6](https://github.com/kubedb/mysql-restic-plugin/commit/29ef1ef6) Fix Makefile (#93)
- [3dbb2b6f](https://github.com/kubedb/mysql-restic-plugin/commit/3dbb2b6f) Prepare for release v0.23.0 (#92)
- [13952311](https://github.com/kubedb/mysql-restic-plugin/commit/13952311) Use forked kubestash/restic (#90)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.39.0-rc.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.39.0-rc.0)




## [kubedb/neo4j](https://github.com/kubedb/neo4j)

### [v0.2.0-rc.0](https://github.com/kubedb/neo4j/releases/tag/v0.2.0-rc.0)

- [84994b32](https://github.com/kubedb/neo4j/commit/84994b32) Prepare for release v0.2.0-rc.0 (#14)
- [96ef4cff](https://github.com/kubedb/neo4j/commit/96ef4cff) bug fix and improved (#12)
- [bfe1d9d4](https://github.com/kubedb/neo4j/commit/bfe1d9d4) Prepare for release v0.1.0 (#11)
- [11567e45](https://github.com/kubedb/neo4j/commit/11567e45) Re-design Neo4j configuration field (#10)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.48.0-rc.0](https://github.com/kubedb/ops-manager/releases/tag/v0.48.0-rc.0)

- [c35f7f05](https://github.com/kubedb/ops-manager/commit/c35f7f051) Prepare for release v0.48.0-rc.0 (#825)
- [7c803e79](https://github.com/kubedb/ops-manager/commit/7c803e794) Redis VNPAY issue resolve (#824)
- [695c66f3](https://github.com/kubedb/ops-manager/commit/695c66f30) Fix clickhouse ops controller setu (#822)
- [6c5e356c](https://github.com/kubedb/ops-manager/commit/6c5e356c7) Update deps (#821)
- [c2a15c74](https://github.com/kubedb/ops-manager/commit/c2a15c744) Use Qdrnat CRD instead QdrantOpsRequest (#820)
- [0ffec481](https://github.com/kubedb/ops-manager/commit/0ffec4817) Prepare for release v0.47.0 (#819)
- [4a669596](https://github.com/kubedb/ops-manager/commit/4a6695963) Qdrant ops fix (#818)
- [f6367c98](https://github.com/kubedb/ops-manager/commit/f6367c98e) Add Qdrant TLS Support (#813)
- [a67e2b7c](https://github.com/kubedb/ops-manager/commit/a67e2b7ce) Oracle TLS (#806)
- [9ff1de6d](https://github.com/kubedb/ops-manager/commit/9ff1de6d1) Reconfigure redesign (#817)
- [8000ffef](https://github.com/kubedb/ops-manager/commit/8000ffefe) Add shard configuration support for ops manager (#812)



## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.7.0-rc.0](https://github.com/kubedb/oracle/releases/tag/v0.7.0-rc.0)

- [bb2b29c0](https://github.com/kubedb/oracle/commit/bb2b29c0) Prepare for release v0.7.0-rc.0 (#28)
- [30a93cc0](https://github.com/kubedb/oracle/commit/30a93cc0) Prepare for release v0.6.0 (#24)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.7.0-rc.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.7.0-rc.0)

- [8212e06](https://github.com/kubedb/oracle-coordinator/commit/8212e06) Prepare for release v0.7.0-rc.0 (#22)
- [7e1bc65](https://github.com/kubedb/oracle-coordinator/commit/7e1bc65) Prepare for release v0.6.0 (#20)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.48.0-rc.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.48.0-rc.0)

- [0a851c7b](https://github.com/kubedb/percona-xtradb/commit/0a851c7b1) Prepare for release v0.48.0-rc.0 (#433)
- [e2448241](https://github.com/kubedb/percona-xtradb/commit/e24482414) Update ops parallelism (#431)
- [c6b0c8cc](https://github.com/kubedb/percona-xtradb/commit/c6b0c8ccc) Prepare for release v0.47.0 (#430)
- [aa2c94b2](https://github.com/kubedb/percona-xtradb/commit/aa2c94b20) Use Exporter Version v0.18.0 (#429)
- [a607e1af](https://github.com/kubedb/percona-xtradb/commit/a607e1af5) Improve and generalize configure-reconfigure (#428)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.34.0-rc.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.34.0-rc.0)

- [593d9f8e](https://github.com/kubedb/percona-xtradb-coordinator/commit/593d9f8e) Prepare for release v0.34.0-rc.0 (#112)
- [6d891d1e](https://github.com/kubedb/percona-xtradb-coordinator/commit/6d891d1e) Prepare for release v0.33.0 (#110)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.45.0-rc.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.45.0-rc.0)

- [beb05456](https://github.com/kubedb/pg-coordinator/commit/beb05456) Prepare for release v0.45.0-rc.0 (#231)
- [f4f84f76](https://github.com/kubedb/pg-coordinator/commit/f4f84f76) Add Read Replica Support (#229)
- [b245d7b5](https://github.com/kubedb/pg-coordinator/commit/b245d7b5) Update timing (#227)
- [bfaefa78](https://github.com/kubedb/pg-coordinator/commit/bfaefa78) Prepare for release v0.44.0 (#226)
- [e3803360](https://github.com/kubedb/pg-coordinator/commit/e3803360) Add prev version compatibility (#220)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.48.0-rc.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.48.0-rc.0)

- [b7c6fedb](https://github.com/kubedb/pgbouncer/commit/b7c6fedbd) Prepare for release v0.48.0-rc.0 (#396)
- [31e2e720](https://github.com/kubedb/pgbouncer/commit/31e2e7207) Added Parallel Controller (#394)
- [0e9564ce](https://github.com/kubedb/pgbouncer/commit/0e9564ce6) Prepare for release v0.47.0 (#393)
- [2d669051](https://github.com/kubedb/pgbouncer/commit/2d669051a) Improve and generalize configure-reconfigure process for all dbs (#392)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.16.0-rc.0](https://github.com/kubedb/pgpool/releases/tag/v0.16.0-rc.0)

- [582b34f8](https://github.com/kubedb/pgpool/commit/582b34f8) Prepare for release v0.16.0-rc.0 (#99)
- [62699f51](https://github.com/kubedb/pgpool/commit/62699f51) Add Parallelism Controller (#97)
- [ccff7fd9](https://github.com/kubedb/pgpool/commit/ccff7fd9) Prepare for release v0.15.0 (#96)
- [a849809d](https://github.com/kubedb/pgpool/commit/a849809d) Improve and generalize configure-reconfigure process (#95)
- [e5f20e47](https://github.com/kubedb/pgpool/commit/e5f20e47) Fix resource watcher for vsecret (#92)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.61.0-rc.0](https://github.com/kubedb/postgres/releases/tag/v0.61.0-rc.0)

- [e9c1b88f](https://github.com/kubedb/postgres/commit/e9c1b88f9) Prepare for release v0.61.0-rc.0 (#856)
- [2564a529](https://github.com/kubedb/postgres/commit/2564a529a) added permissions for binlog-cloeanup (#855)
- [c16dae52](https://github.com/kubedb/postgres/commit/c16dae527) Add Read Replica Support  (#854)
- [b44e0d83](https://github.com/kubedb/postgres/commit/b44e0d839) Add pod toleration during storageClass migration (#853)
- [acc6ee97](https://github.com/kubedb/postgres/commit/acc6ee979) Incorporate new snapshot naming in postgres archiver (#851)
- [dde64cf3](https://github.com/kubedb/postgres/commit/dde64cf30) Update parallel processing logic (#849)
- [5086c260](https://github.com/kubedb/postgres/commit/5086c260c) Update Arbiter error code (#848)
- [a313afc9](https://github.com/kubedb/postgres/commit/a313afc9e) Prepare for release v0.60.0 (#847)
- [4d86169d](https://github.com/kubedb/postgres/commit/4d86169d7) Add sharding facility for Postgres Ops-Requests (#843)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.22.0-rc.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.22.0-rc.0)

- [6138276c](https://github.com/kubedb/postgres-archiver/commit/6138276c) Prepare for release v0.22.0-rc.0 (#85)
- [6f702e23](https://github.com/kubedb/postgres-archiver/commit/6f702e23) Test against k8s 1.35 (#84)
- [49c942eb](https://github.com/kubedb/postgres-archiver/commit/49c942eb) added wal-log clean feature (#83)
- [e92dc052](https://github.com/kubedb/postgres-archiver/commit/e92dc052) Fix ubi build (#81)
- [f45faf8c](https://github.com/kubedb/postgres-archiver/commit/f45faf8c) Prepare for release v0.21.0 (#80)
- [0c1f7850](https://github.com/kubedb/postgres-archiver/commit/0c1f7850) Fix redhat catalog submission (#78)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.22.0-rc.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.22.0-rc.0)

- [44414ace](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/44414ace) Prepare for release v0.22.0-rc.0 (#77)
- [ea6e3c50](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/ea6e3c50) Prepare for release v0.21.0 (#75)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.24.0-rc.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.24.0-rc.0)

- [c1a56b92](https://github.com/kubedb/postgres-restic-plugin/commit/c1a56b92) Prepare for release v0.24.0-rc.0 (#93)
- [2f72f231](https://github.com/kubedb/postgres-restic-plugin/commit/2f72f231) Incorporate changes for restic standalone pkg (#92)
- [ff45044b](https://github.com/kubedb/postgres-restic-plugin/commit/ff45044b) Prepare for release v0.23.0 (#90)
- [441378e5](https://github.com/kubedb/postgres-restic-plugin/commit/441378e5) Use forked kubestash/restic (#89)
- [f5f19a6f](https://github.com/kubedb/postgres-restic-plugin/commit/f5f19a6f) Use forked kubestash/restic (#88)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.22.0-rc.0](https://github.com/kubedb/provider-aws/releases/tag/v0.22.0-rc.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.22.0-rc.0](https://github.com/kubedb/provider-azure/releases/tag/v0.22.0-rc.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.22.0-rc.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.22.0-rc.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.61.0-rc.0](https://github.com/kubedb/provisioner/releases/tag/v0.61.0-rc.0)

- [f24e8311](https://github.com/kubedb/provisioner/commit/f24e8311c) Prepare for release v0.61.0-rc.0 (#192)
- [9f7d961f](https://github.com/kubedb/provisioner/commit/9f7d961f6) Updates for ferret, weaviate & snapshot changes (#190)
- [a94ed0b5](https://github.com/kubedb/provisioner/commit/a94ed0b58) Update deps for DBs (#187)
- [773dafc5](https://github.com/kubedb/provisioner/commit/773dafc59) add etcd sceme (#188)
- [81dedbc0](https://github.com/kubedb/provisioner/commit/81dedbc0b) Prepare for release v0.60.0 (#185)
- [6abadcb1](https://github.com/kubedb/provisioner/commit/6abadcb1a) Add clickhouse (#184)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.48.0-rc.0](https://github.com/kubedb/proxysql/releases/tag/v0.48.0-rc.0)

- [7a4da426](https://github.com/kubedb/proxysql/commit/7a4da426c) Prepare for release v0.48.0-rc.0 (#416)
- [835a5511](https://github.com/kubedb/proxysql/commit/835a55112) Merge pull request #414 from kubedb/run-parallel-imp
- [cdcee6ba](https://github.com/kubedb/proxysql/commit/cdcee6baf) Prepare for release v0.47.0 (#413)
- [96edac7b](https://github.com/kubedb/proxysql/commit/96edac7b7) Improve and generalize configure-reconfigure process (#412)



## [kubedb/qdrant](https://github.com/kubedb/qdrant)

### [v0.2.0-rc.0](https://github.com/kubedb/qdrant/releases/tag/v0.2.0-rc.0)

- [22c54605](https://github.com/kubedb/qdrant/commit/22c54605) Prepare for release v0.2.0-rc.0 (#18)
- [756abbb6](https://github.com/kubedb/qdrant/commit/756abbb6) TLS fixes (#16)
- [0ab05f74](https://github.com/kubedb/qdrant/commit/0ab05f74) Update deps (#15)
- [1ce521e9](https://github.com/kubedb/qdrant/commit/1ce521e9) Prepare for release v0.1.0 (#14)
- [051524f1](https://github.com/kubedb/qdrant/commit/051524f1) Re-design Configuration process (#12)
- [27031124](https://github.com/kubedb/qdrant/commit/27031124) Add TLS support (#9)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.16.0-rc.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.16.0-rc.0)

- [052636d5](https://github.com/kubedb/rabbitmq/commit/052636d5) Prepare for release v0.16.0-rc.0 (#113)
- [1215655f](https://github.com/kubedb/rabbitmq/commit/1215655f) Update Parallel Processing Logic (#111)
- [7a862fd3](https://github.com/kubedb/rabbitmq/commit/7a862fd3) Update Paraller Processing Logic (#110)
- [10445e04](https://github.com/kubedb/rabbitmq/commit/10445e04) update deletion Policy (#109)
- [34fefed0](https://github.com/kubedb/rabbitmq/commit/34fefed0) Prepare for release v0.15.0 (#108)
- [4de46792](https://github.com/kubedb/rabbitmq/commit/4de46792) Update configure-reconfigure process (#107)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.54.0-rc.0](https://github.com/kubedb/redis/releases/tag/v0.54.0-rc.0)

- [dd5032fa](https://github.com/kubedb/redis/commit/dd5032fa1) Prepare for release v0.54.0-rc.0 (#621)
- [a95277cc](https://github.com/kubedb/redis/commit/a95277ccf) VNPAY issue Resolve (#620)
- [3e886b96](https://github.com/kubedb/redis/commit/3e886b961) Update api (#619)
- [a9488ef7](https://github.com/kubedb/redis/commit/a9488ef7d) Redis Ops bug fix and Parallelism improvement (#617)
- [eeaf253e](https://github.com/kubedb/redis/commit/eeaf253e9) Prepare for release v0.53.0 (#616)
- [9d27adae](https://github.com/kubedb/redis/commit/9d27adaeb) Improve and generalize configure-reconfigure process (#615)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.40.0-rc.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.40.0-rc.0)

- [e6f57560](https://github.com/kubedb/redis-coordinator/commit/e6f57560) Prepare for release v0.40.0-rc.0 (#146)
- [0a121386](https://github.com/kubedb/redis-coordinator/commit/0a121386) Prepare for release v0.39.0 (#144)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.24.0-rc.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.24.0-rc.0)

- [e850d1c4](https://github.com/kubedb/redis-restic-plugin/commit/e850d1c4) Prepare for release v0.24.0-rc.0 (#90)
- [e0025b19](https://github.com/kubedb/redis-restic-plugin/commit/e0025b19) Incorporate changes for restic standalone pkg (#89)
- [a84cb45c](https://github.com/kubedb/redis-restic-plugin/commit/a84cb45c) Test against k8s 1.35 (#88)
- [c0283608](https://github.com/kubedb/redis-restic-plugin/commit/c0283608) Update DB ready condition check (#87)
- [acb3ec20](https://github.com/kubedb/redis-restic-plugin/commit/acb3ec20) Prepare for release v0.23.0 (#85)
- [095746db](https://github.com/kubedb/redis-restic-plugin/commit/095746db) Use forked kubestash/restic (#84)
- [ecb820f6](https://github.com/kubedb/redis-restic-plugin/commit/ecb820f6) Use forked kubestash/restic (#83)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.48.0-rc.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.48.0-rc.0)

- [02c69b56](https://github.com/kubedb/replication-mode-detector/commit/02c69b56) Prepare for release v0.48.0-rc.0 (#309)
- [e43601fb](https://github.com/kubedb/replication-mode-detector/commit/e43601fb) Prepare for release v0.47.0 (#307)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.37.0-rc.0](https://github.com/kubedb/schema-manager/releases/tag/v0.37.0-rc.0)

- [0404123c](https://github.com/kubedb/schema-manager/commit/0404123c) Prepare for release v0.37.0-rc.0 (#156)
- [c83b51e0](https://github.com/kubedb/schema-manager/commit/c83b51e0) Prepare for release v0.36.0 (#154)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.16.0-rc.0](https://github.com/kubedb/singlestore/releases/tag/v0.16.0-rc.0)

- [084e27aa](https://github.com/kubedb/singlestore/commit/084e27aa) Prepare for release v0.16.0-rc.0 (#99)
- [6a1e622c](https://github.com/kubedb/singlestore/commit/6a1e622c) Update ops parallelism (#97)
- [d655bfec](https://github.com/kubedb/singlestore/commit/d655bfec) Prepare for release v0.15.0 (#96)
- [08bcff61](https://github.com/kubedb/singlestore/commit/08bcff61) Improve and generalize configure-reconfigure (#95)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.16.0-rc.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.16.0-rc.0)

- [bc0e985f](https://github.com/kubedb/singlestore-coordinator/commit/bc0e985f) Prepare for release v0.16.0-rc.0 (#59)
- [49dc2d43](https://github.com/kubedb/singlestore-coordinator/commit/49dc2d43) Prepare for release v0.15.0 (#57)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.19.0-rc.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.19.0-rc.0)

- [7ca8197e](https://github.com/kubedb/singlestore-restic-plugin/commit/7ca8197e) Prepare for release v0.19.0-rc.0 (#70)
- [6f0e374a](https://github.com/kubedb/singlestore-restic-plugin/commit/6f0e374a) Incorporate changes for restic standalone pkg (#69)
- [52d3ae86](https://github.com/kubedb/singlestore-restic-plugin/commit/52d3ae86) Delete .golangci.yml.1 (#67)
- [e98ce73b](https://github.com/kubedb/singlestore-restic-plugin/commit/e98ce73b) Fix for restic/kubestash (#66)
- [c5f8dafa](https://github.com/kubedb/singlestore-restic-plugin/commit/c5f8dafa) Prepare for release v0.18.0 (#65)
- [c98bdded](https://github.com/kubedb/singlestore-restic-plugin/commit/c98bdded) Use forked kubestash/restic (#63)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.16.0-rc.0](https://github.com/kubedb/solr/releases/tag/v0.16.0-rc.0)

- [b646dfb1](https://github.com/kubedb/solr/commit/b646dfb1) Prepare for release v0.16.0-rc.0 (#111)
- [58f2dd61](https://github.com/kubedb/solr/commit/58f2dd61) Merge pull request #109 from kubedb/proceed
- [620d72f7](https://github.com/kubedb/solr/commit/620d72f7) Prepare for release v0.15.0 (#108)
- [f58a1f91](https://github.com/kubedb/solr/commit/f58a1f91) Update Configure reconfigure process (#107)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.46.0-rc.0](https://github.com/kubedb/tests/releases/tag/v0.46.0-rc.0)

- [b5255636](https://github.com/kubedb/tests/commit/b52556363) Prepare for release v0.46.0-rc.0 (#509)
- [6231c498](https://github.com/kubedb/tests/commit/6231c4982) Prepare for release v0.46.0-rc.0 (#508)
- [11c81d08](https://github.com/kubedb/tests/commit/11c81d087) Reconfigure Release Check (#505)
- [3516c7cb](https://github.com/kubedb/tests/commit/3516c7cb4) Prepare for release v0.45.0 (#504)
- [dab2b9be](https://github.com/kubedb/tests/commit/dab2b9be1) remove init_config form pgpool (#503)
- [0f520a67](https://github.com/kubedb/tests/commit/0f520a674) reconfigure change: druid, solr, pgpool, redis, (#502)
- [4dc2db5c](https://github.com/kubedb/tests/commit/4dc2db5c8) add apimachinery changes [reconfigurationSpec] (#501)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.37.0-rc.0](https://github.com/kubedb/ui-server/releases/tag/v0.37.0-rc.0)

- [15c16908](https://github.com/kubedb/ui-server/commit/15c16908) Prepare for release v0.37.0-rc.0 (#188)
- [50ce64af](https://github.com/kubedb/ui-server/commit/50ce64af) Prepare for release v0.36.0 (#186)



## [kubedb/weaviate](https://github.com/kubedb/weaviate)

### [v0.2.0-rc.0](https://github.com/kubedb/weaviate/releases/tag/v0.2.0-rc.0)

- [c0f5f858](https://github.com/kubedb/weaviate/commit/c0f5f858) Prepare for release v0.2.0-rc.0 (#15)
- [2cc31997](https://github.com/kubedb/weaviate/commit/2cc31997) update deletion Policy (#11)
- [cc2297c3](https://github.com/kubedb/weaviate/commit/cc2297c3) Prepare for release v0.1.0 (#10)
- [71791e5b](https://github.com/kubedb/weaviate/commit/71791e5b) Configuration redesign (#9)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.37.0-rc.0](https://github.com/kubedb/webhook-server/releases/tag/v0.37.0-rc.0)

- [3be1e859](https://github.com/kubedb/webhook-server/commit/3be1e859) Prepare for release v0.37.0-rc.0 (#194)
- [f7964f91](https://github.com/kubedb/webhook-server/commit/f7964f91) Update API for VNPAY volume expansion issue (#191)
- [ec1aeada](https://github.com/kubedb/webhook-server/commit/ec1aeada) Update Deps (#190)
- [121b23fa](https://github.com/kubedb/webhook-server/commit/121b23fa) Update Deps (#189)
- [aceff628](https://github.com/kubedb/webhook-server/commit/aceff628) Prepare for release v0.36.0 (#188)
- [60f8f40d](https://github.com/kubedb/webhook-server/commit/60f8f40d) Use Go version 1.25 in go.mod and Makefile



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.9.0-rc.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.9.0-rc.0)

- [416ddbe](https://github.com/kubedb/xtrabackup-restic-plugin/commit/416ddbe) Prepare for release v0.9.0-rc.0 (#37)
- [7c92d9a](https://github.com/kubedb/xtrabackup-restic-plugin/commit/7c92d9a) Incorporate changes for restic standalone pkg (#36)
- [66414d7](https://github.com/kubedb/xtrabackup-restic-plugin/commit/66414d7) Fix OS and ARC in makefile (#34)
- [6250c18](https://github.com/kubedb/xtrabackup-restic-plugin/commit/6250c18) Use forked kubestasgh/restic (#31)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.16.0-rc.0](https://github.com/kubedb/zookeeper/releases/tag/v0.16.0-rc.0)

- [111bc2a6](https://github.com/kubedb/zookeeper/commit/111bc2a6) Prepare for release v0.16.0-rc.0 (#103)
- [64f10ce2](https://github.com/kubedb/zookeeper/commit/64f10ce2) Integrate shouldProceed() utility on runParallel (#100)
- [e354f44a](https://github.com/kubedb/zookeeper/commit/e354f44a) Prepare for release v0.15.0 (#99)
- [7eeed01d](https://github.com/kubedb/zookeeper/commit/7eeed01d) Update Configure reconfigure process (#98)




