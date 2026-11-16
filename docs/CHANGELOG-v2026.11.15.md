---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2026.11.15
    name: Changelog-v2026.11.15
    parent: welcome
    weight: 20261115
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2026.11.15/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2026.11.15/
---

# KubeDB v2026.11.15 (2026-11-14)


## [kubedb/aerospike](https://github.com/kubedb/aerospike)

### [v0.5.0](https://github.com/kubedb/aerospike/releases/tag/v0.5.0)

- [1dc8cfc7](https://github.com/kubedb/aerospike/commit/1dc8cfc7) Prepare for release v0.5.0 (#20)



## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.69.0](https://github.com/kubedb/apimachinery/releases/tag/v0.69.0)

- [cb8b0c69](https://github.com/kubedb/apimachinery/commit/cb8b0c69e) Cleanup deps
- [f42a5fad](https://github.com/kubedb/apimachinery/commit/f42a5fadb) Add Postgres MajorUpgradeInProgress raft key and conditions (#1954)
- [dbfac26f](https://github.com/kubedb/apimachinery/commit/dbfac26f5) Cache the OpenShift check per RESTMapper (#1956)
- [b89bffca](https://github.com/kubedb/apimachinery/commit/b89bffca5) OracleOSBWS Constants (#1953)
- [d24add47](https://github.com/kubedb/apimachinery/commit/d24add471) Default database UID/GID/fsGroup from OpenShift namespace SCC ranges (#1952)
- [30b10799](https://github.com/kubedb/apimachinery/commit/30b10799b) Add spec.objectStorage.tls to Milvus for a custom object storage CA (#1949)
- [0e4a9a41](https://github.com/kubedb/apimachinery/commit/0e4a9a412) Milvus: add spec.autoOps, GPU guard for UpdateVersion, StreamingNode storage default (#1947)
- [6cd3915e](https://github.com/kubedb/apimachinery/commit/6cd3915e3) Milvus webhook: allow the archiver on Milvus 3.x (#1945)
- [f07dcf5a](https://github.com/kubedb/apimachinery/commit/f07dcf5a4) Add Qdrant PITR Support (#1671)
- [bf4b2936](https://github.com/kubedb/apimachinery/commit/bf4b2936a) add mysql auto reconfigure feature (#1951)
- [c6b3e417](https://github.com/kubedb/apimachinery/commit/c6b3e4178) Restrict shared service aliases; add per-DB alias types and Kafka broker alias (#1931)
- [b537910b](https://github.com/kubedb/apimachinery/commit/b537910ba) [MySQL] Add Archiver Restore Ops (#1854)
- [b62975e8](https://github.com/kubedb/apimachinery/commit/b62975e85) Add Neo4j archiver CRD helper (#1927)
- [4606c7f8](https://github.com/kubedb/apimachinery/commit/4606c7f88) Add RecoverMember ops request type for Etcd (#1930)
- [9e71e6a1](https://github.com/kubedb/apimachinery/commit/9e71e6a1a) Accept a Milvus VerticalScaling that only targets node groups (#1929)
- [c5629ed3](https://github.com/kubedb/apimachinery/commit/c5629ed38) Milvus: external-etcd client auth/TLS, etcd root path, restricted SCC knob (#1928)
- [5af9abb5](https://github.com/kubedb/apimachinery/commit/5af9abb56) Fix lint issues



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.54.0](https://github.com/kubedb/autoscaler/releases/tag/v0.54.0)

- [3deb5ab5](https://github.com/kubedb/autoscaler/commit/3deb5ab5c) Prepare for release v0.54.0 (#322)



## [kubedb/cassandra](https://github.com/kubedb/cassandra)

### [v0.22.0](https://github.com/kubedb/cassandra/releases/tag/v0.22.0)

- [acf93ea7](https://github.com/kubedb/cassandra/commit/acf93ea7) Prepare for release v0.22.0 (#112)
- [112d65ce](https://github.com/kubedb/cassandra/commit/112d65ce) Default security IDs from the OpenShift namespace in the operator (#111)



## [kubedb/cassandra-medusa-plugin](https://github.com/kubedb/cassandra-medusa-plugin)

### [v0.16.0](https://github.com/kubedb/cassandra-medusa-plugin/releases/tag/v0.16.0)

- [c000e161](https://github.com/kubedb/cassandra-medusa-plugin/commit/c000e161) Prepare for release v0.16.0 (#50)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.69.0](https://github.com/kubedb/cli/releases/tag/v0.69.0)

- [efba2f31](https://github.com/kubedb/cli/commit/efba2f311) Prepare for release v0.69.0 (#848)
- [2fea7014](https://github.com/kubedb/cli/commit/2fea7014f) Merge pull request #844 from kubedb/dc-dr-failover-group



## [kubedb/clickhouse](https://github.com/kubedb/clickhouse)

### [v0.24.0](https://github.com/kubedb/clickhouse/releases/tag/v0.24.0)

- [2a903f01](https://github.com/kubedb/clickhouse/commit/2a903f01) Prepare for release v0.24.0 (#144)
- [c45846db](https://github.com/kubedb/clickhouse/commit/c45846db) Disaster Recovery For Shard Replica (#138)
- [c36c65c8](https://github.com/kubedb/clickhouse/commit/c36c65c8) Stop leaking primary Service nodePorts into governing Service (#142)



## [kubedb/clickhouse-backup-plugin](https://github.com/kubedb/clickhouse-backup-plugin)

### [v0.6.0](https://github.com/kubedb/clickhouse-backup-plugin/releases/tag/v0.6.0)

- [a0042e48](https://github.com/kubedb/clickhouse-backup-plugin/commit/a0042e48) Prepare for release v0.6.0 (#38)



## [kubedb/courier](https://github.com/kubedb/courier)

### [v0.9.0](https://github.com/kubedb/courier/releases/tag/v0.9.0)

- [b1182a9](https://github.com/kubedb/courier/commit/b1182a9) Prepare for release v0.9.0 (#65)
- [93b716f](https://github.com/kubedb/courier/commit/93b716f) courier: add Microsoft SQL Server Branch engine (#35)
- [466a5c0](https://github.com/kubedb/courier/commit/466a5c0) courier: add Kafka Branch engine (distributed) (#41)
- [6da045b](https://github.com/kubedb/courier/commit/6da045b) courier: add Neo4j Branch engine (#42)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.24.0](https://github.com/kubedb/crd-manager/releases/tag/v0.24.0)

- [c39f6de1](https://github.com/kubedb/crd-manager/commit/c39f6de1) Prepare for release v0.24.0 (#162)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.27.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.27.0)

- [3b2b41a0](https://github.com/kubedb/dashboard-restic-plugin/commit/3b2b41a0) Prepare for release v0.27.0 (#90)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.24.0](https://github.com/kubedb/db-client-go/releases/tag/v0.24.0)

- [dc07eaf4](https://github.com/kubedb/db-client-go/commit/dc07eaf4) Prepare for release v0.24.0 (#275)
- [ef134e34](https://github.com/kubedb/db-client-go/commit/ef134e34) postgres: bound GetPostgresXormClient's SELECT 1 by the caller context; close engine on failure (#269)
- [af39558a](https://github.com/kubedb/db-client-go/commit/af39558a) etcd: start clients on round_robin to stop canceled-dial warnings (#272)
- [a39d2ab9](https://github.com/kubedb/db-client-go/commit/a39d2ab9) etcd: dial learners anonymously, bound per-member Status probes (#267)



## [kubedb/db2](https://github.com/kubedb/db2)

### [v0.10.0](https://github.com/kubedb/db2/releases/tag/v0.10.0)

- [1a0c7022](https://github.com/kubedb/db2/commit/1a0c7022) Prepare for release v0.10.0 (#48)
- [1f64e540](https://github.com/kubedb/db2/commit/1f64e540) Default security IDs from the OpenShift namespace in the operator (#47)



## [kubedb/db2-coordinator](https://github.com/kubedb/db2-coordinator)

### [v0.10.0](https://github.com/kubedb/db2-coordinator/releases/tag/v0.10.0)

- [4487d38](https://github.com/kubedb/db2-coordinator/commit/4487d38) Prepare for release v0.10.0 (#25)



## [kubedb/documentdb](https://github.com/kubedb/documentdb)

### [v0.6.0](https://github.com/kubedb/documentdb/releases/tag/v0.6.0)

- [574a85ed](https://github.com/kubedb/documentdb/commit/574a85ed) Prepare for release v0.6.0 (#56)
- [fef315fc](https://github.com/kubedb/documentdb/commit/fef315fc) Add documentdb basebackup restore & PITR (#51)
- [7878763d](https://github.com/kubedb/documentdb/commit/7878763d) Default security IDs from the OpenShift namespace in the operator (#55)



## [kubedb/documentdb-coordinator](https://github.com/kubedb/documentdb-coordinator)

### [v0.5.0](https://github.com/kubedb/documentdb-coordinator/releases/tag/v0.5.0)

- [3ed03d8](https://github.com/kubedb/documentdb-coordinator/commit/3ed03d8) Prepare for release v0.5.0 (#17)



## [kubedb/documentdb-csi-snapshotter-plugin](https://github.com/kubedb/documentdb-csi-snapshotter-plugin)

### [v0.3.0](https://github.com/kubedb/documentdb-csi-snapshotter-plugin/releases/tag/v0.3.0)

- [84daceb](https://github.com/kubedb/documentdb-csi-snapshotter-plugin/commit/84daceb) Prepare for release v0.3.0 (#3)
- [178d225](https://github.com/kubedb/documentdb-csi-snapshotter-plugin/commit/178d225) Use Go 1.27, k8s 1.37 & debian:13 (#2)
- [7a1ce6e](https://github.com/kubedb/documentdb-csi-snapshotter-plugin/commit/7a1ce6e) Add documentdb csi snapshotter plugin (#1)
- [458b87f](https://github.com/kubedb/documentdb-csi-snapshotter-plugin/commit/458b87f) Add ci files and gitignore



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.24.0](https://github.com/kubedb/druid/releases/tag/v0.24.0)

- [376b2284](https://github.com/kubedb/druid/commit/376b2284) Prepare for release v0.24.0 (#161)
- [414c272b](https://github.com/kubedb/druid/commit/414c272b) Default security IDs from the OpenShift namespace in the operator (#160)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.69.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.69.0)

- [69737852](https://github.com/kubedb/elasticsearch/commit/697378523) Prepare for release v0.69.0 (#844)
- [f690eb03](https://github.com/kubedb/elasticsearch/commit/f690eb039) Default security IDs from the OpenShift namespace in the operator (#843)
- [fbfd719e](https://github.com/kubedb/elasticsearch/commit/fbfd719ea) Apply Elasticsearch reconfiguration without restart when possible (#841)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.32.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.32.0)

- [77a200ff](https://github.com/kubedb/elasticsearch-restic-plugin/commit/77a200ff) Prepare for release v0.32.0 (#113)



## [kubedb/etcd](https://github.com/kubedb/etcd)

### [v0.3.0](https://github.com/kubedb/etcd/releases/tag/v0.3.0)

- [ebbcfede](https://github.com/kubedb/etcd/commit/ebbcfede) Prepare for release v0.3.0 (#13)
- [c04a28ec](https://github.com/kubedb/etcd/commit/c04a28ec) Default security IDs from the OpenShift namespace in the operator (#12)
- [85bbc15a](https://github.com/kubedb/etcd/commit/85bbc15a) Run restore jobs as the etcd container's uid/gid (#11)
- [b938e665](https://github.com/kubedb/etcd/commit/b938e665) Enable etcd authentication once the cluster is Ready (#10)



## [kubedb/etcd-restic-plugin](https://github.com/kubedb/etcd-restic-plugin)

### [v0.3.0](https://github.com/kubedb/etcd-restic-plugin/releases/tag/v0.3.0)

- [8f96766](https://github.com/kubedb/etcd-restic-plugin/commit/8f96766) Prepare for release v0.3.0 (#5)



## [kubedb/gitops](https://github.com/kubedb/gitops)

### [v0.17.0](https://github.com/kubedb/gitops/releases/tag/v0.17.0)

- [56d39d66](https://github.com/kubedb/gitops/commit/56d39d66) Prepare for release v0.17.0 (#100)
- [c952cdfd](https://github.com/kubedb/gitops/commit/c952cdfd) Add HanaDB horizontal scaling, update version GitOps support (#87)
- [fddb0df6](https://github.com/kubedb/gitops/commit/fddb0df6) Add GitOps support for DocumentDB (#94)



## [kubedb/hanadb](https://github.com/kubedb/hanadb)

### [v0.10.0](https://github.com/kubedb/hanadb/releases/tag/v0.10.0)

- [c3db55db](https://github.com/kubedb/hanadb/commit/c3db55db) Prepare for release v0.10.0 (#69)
- [9a1e05cf](https://github.com/kubedb/hanadb/commit/9a1e05cf) Default security IDs from the OpenShift namespace in the operator (#68)



## [kubedb/hanadb-coordinator](https://github.com/kubedb/hanadb-coordinator)

### [v0.9.0](https://github.com/kubedb/hanadb-coordinator/releases/tag/v0.9.0)

- [6697127d](https://github.com/kubedb/hanadb-coordinator/commit/6697127d) Prepare for release v0.9.0 (#27)



## [kubedb/hazelcast](https://github.com/kubedb/hazelcast)

### [v0.15.0](https://github.com/kubedb/hazelcast/releases/tag/v0.15.0)

- [22c72bfd](https://github.com/kubedb/hazelcast/commit/22c72bfd) Prepare for release v0.15.0 (#73)
- [a6faaaef](https://github.com/kubedb/hazelcast/commit/a6faaaef) Default security IDs from the OpenShift namespace in the operator (#72)



## [kubedb/ignite](https://github.com/kubedb/ignite)

### [v0.16.0](https://github.com/kubedb/ignite/releases/tag/v0.16.0)

- [1f1b5b19](https://github.com/kubedb/ignite/commit/1f1b5b19) Prepare for release v0.16.0 (#81)
- [6f153323](https://github.com/kubedb/ignite/commit/6f153323) Default security IDs from the OpenShift namespace in the operator (#80)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2026.11.15](https://github.com/kubedb/installer/releases/tag/v2026.11.15)

- [c2c45cbc](https://github.com/kubedb/installer/commit/c2c45cbcc) Prepare for release v2026.11.15 (#2488)
- [8fec90a9](https://github.com/kubedb/installer/commit/8fec90a9a) Coordinate Milvus object-storage TLS backup worker rollout (#2482)
- [9d13743c](https://github.com/kubedb/installer/commit/9d13743c9) Etcd: set runAsAnyNonRoot to true in EtcdVersion catalog (#2487)
- [96d57e4a](https://github.com/kubedb/installer/commit/96d57e4ac) Milvus: archiver addon and walg image for 2.6.15 and 3.0.0 (#2477)
- [af59a6d5](https://github.com/kubedb/installer/commit/af59a6d51) Enable open-source databases in DBaaS licenses (#2485)
- [ce31cc78](https://github.com/kubedb/installer/commit/ce31cc783) remove recoverToLatest flag (#2444)
- [c5d472e7](https://github.com/kubedb/installer/commit/c5d472e7b) Update MySQL Init Docker (#2475)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.40.0](https://github.com/kubedb/kafka/releases/tag/v0.40.0)

- [2e8e2700](https://github.com/kubedb/kafka/commit/2e8e2700) Prepare for release v0.40.0 (#227)
- [ec01de7f](https://github.com/kubedb/kafka/commit/ec01de7f) Default security IDs from the OpenShift namespace in the operator (#226)
- [b0cd83e8](https://github.com/kubedb/kafka/commit/b0cd83e8) Add branched mode for KubeDB Courier Branch (#207)
- [65ebad4a](https://github.com/kubedb/kafka/commit/65ebad4a) Add external access for Kafka brokers via per-broker LoadBalancer Services (#225)



## [kubedb/kibana](https://github.com/kubedb/kibana)

### [v0.45.0](https://github.com/kubedb/kibana/releases/tag/v0.45.0)

- [8f0b8c26](https://github.com/kubedb/kibana/commit/8f0b8c26) Prepare for release v0.45.0 (#194)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.32.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.32.0)

- [171d67a7](https://github.com/kubedb/kubedb-manifest-plugin/commit/171d67a7) Prepare for release v0.32.0 (#149)
- [862244f5](https://github.com/kubedb/kubedb-manifest-plugin/commit/862244f5) Add documentdb manifest backup&restore (#145)



## [kubedb/kubedb-verifier](https://github.com/kubedb/kubedb-verifier)

### [v0.20.0](https://github.com/kubedb/kubedb-verifier/releases/tag/v0.20.0)

- [62fb2d4b](https://github.com/kubedb/kubedb-verifier/commit/62fb2d4b) Prepare for release v0.20.0 (#64)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.53.0](https://github.com/kubedb/mariadb/releases/tag/v0.53.0)

- [144cdc9c](https://github.com/kubedb/mariadb/commit/144cdc9c5) Prepare for release v0.53.0 (#439)
- [9c9a6613](https://github.com/kubedb/mariadb/commit/9c9a6613b) Default security IDs from the OpenShift namespace in the operator (#437)
- [782d4384](https://github.com/kubedb/mariadb/commit/782d43848) Allow skipping the primary service template for MaxScale or MariaDB (#438)
- [d07615c7](https://github.com/kubedb/mariadb/commit/d07615c7f) add mariadb auto reconfigure feature (#436)
- [b79ab6ad](https://github.com/kubedb/mariadb/commit/b79ab6adf) Default exporter resources in vertical scaling (#432)
- [09048e77](https://github.com/kubedb/mariadb/commit/09048e77b) Add the ArchiverRestore ops request and in-place restore path (#428)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.29.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.29.0)

- [ef202691](https://github.com/kubedb/mariadb-archiver/commit/ef202691) Prepare for release v0.29.0 (#108)
- [e3330cdd](https://github.com/kubedb/mariadb-archiver/commit/e3330cdd) Use Go 1.27, k8s 1.37 & debian:13 (#107)
- [10e8e9b7](https://github.com/kubedb/mariadb-archiver/commit/10e8e9b7) Stop the replay reporting success over binlogs it never applied (#102)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.49.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.49.0)

- [a815cddf](https://github.com/kubedb/mariadb-coordinator/commit/a815cddf) Prepare for release v0.49.0 (#193)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.29.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.29.0)

- [19db0b59](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/19db0b59) Prepare for release v0.29.0 (#90)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.27.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.27.0)

- [21c5284c](https://github.com/kubedb/mariadb-restic-plugin/commit/21c5284c) Prepare for release v0.27.0 (#107)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.63.0](https://github.com/kubedb/memcached/releases/tag/v0.63.0)

- [2a4bfd1c](https://github.com/kubedb/memcached/commit/2a4bfd1cc) Prepare for release v0.63.0 (#562)
- [59c850d7](https://github.com/kubedb/memcached/commit/59c850d77) Default security IDs from the OpenShift namespace in the operator (#561)



## [kubedb/migrator](https://github.com/kubedb/migrator)

### [v0.9.0](https://github.com/kubedb/migrator/releases/tag/v0.9.0)

- [70540922](https://github.com/kubedb/migrator/commit/70540922) Prepare for release v0.9.0 (#58)
- [2fbbdb62](https://github.com/kubedb/migrator/commit/2fbbdb62) Use Go 1.27, k8s 1.37 & debian:13 (#57)



## [kubedb/milvus](https://github.com/kubedb/milvus)

### [v0.10.0](https://github.com/kubedb/milvus/releases/tag/v0.10.0)

- [6b48c2ae](https://github.com/kubedb/milvus/commit/6b48c2ae) Prepare for release v0.10.0 (#90)
- [55162b5d](https://github.com/kubedb/milvus/commit/55162b5d) Default security IDs from the OpenShift namespace in the operator (#89)
- [da38092c](https://github.com/kubedb/milvus/commit/da38092c) Fix RotateAuth lockout, false InPlace successes, and Halt/WipeOut of the meta etcd (#82)
- [1b21b402](https://github.com/kubedb/milvus/commit/1b21b402) Mark the Milvus incremental snapshot Succeeded when the archiver is removed (#83)
- [f0dbd153](https://github.com/kubedb/milvus/commit/f0dbd153) Add spec.objectStorage.tls for an object storage with a private CA (#87)
- [00955f64](https://github.com/kubedb/milvus/commit/00955f64) Stop a new Milvus on a used object storage path (#86)
- [64c0b3ff](https://github.com/kubedb/milvus/commit/64c0b3ff) Propagate podTemplate securityContext to existing PetSets (#88)
- [42c05e0f](https://github.com/kubedb/milvus/commit/42c05e0f) Start meta etcd clients on round_robin (#81)



## [kubedb/milvus-archiver](https://github.com/kubedb/milvus-archiver)

### [v0.3.0](https://github.com/kubedb/milvus-archiver/releases/tag/v0.3.0)

- [ab07a1b](https://github.com/kubedb/milvus-archiver/commit/ab07a1b) Prepare for release v0.3.0 (#9)
- [1229059](https://github.com/kubedb/milvus-archiver/commit/1229059) Trust customer object-storage CAs independently in Milvus backup workers (#8)
- [d04e29b](https://github.com/kubedb/milvus-archiver/commit/d04e29b) archive: keep the WAL component Running while archiving (#5)
- [6934abf](https://github.com/kubedb/milvus-archiver/commit/6934abf) Add storage-check to report a used object storage path (#7)



## [kubedb/milvus-backup-plugin](https://github.com/kubedb/milvus-backup-plugin)

### [v0.2.0](https://github.com/kubedb/milvus-backup-plugin/releases/tag/v0.2.0)

- [3b42051](https://github.com/kubedb/milvus-backup-plugin/commit/3b42051) Prepare for release v0.2.0 (#8)
- [4e884ae](https://github.com/kubedb/milvus-backup-plugin/commit/4e884ae) Support independent object-storage TLS in Milvus backup and restore (#7)



## [kubedb/milvus-csi-snapshotter-plugin](https://github.com/kubedb/milvus-csi-snapshotter-plugin)

### [v0.3.0](https://github.com/kubedb/milvus-csi-snapshotter-plugin/releases/tag/v0.3.0)

- [52d8da3](https://github.com/kubedb/milvus-csi-snapshotter-plugin/commit/52d8da3) Prepare for release v0.3.0 (#5)
- [29d792e](https://github.com/kubedb/milvus-csi-snapshotter-plugin/commit/29d792e) Use independent object-storage trust in Milvus CSI backups (#4)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.62.0](https://github.com/kubedb/mongodb/releases/tag/v0.62.0)

- [e0cfa3f0](https://github.com/kubedb/mongodb/commit/e0cfa3f04) Prepare for release v0.62.0 (#793)
- [65aea857](https://github.com/kubedb/mongodb/commit/65aea8572) Default security IDs from the OpenShift namespace in the operator (#792)
- [8dd3669c](https://github.com/kubedb/mongodb/commit/8dd3669cc) Apply document-valued setParameter entries as one parameter on Reconfigure (#791)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.30.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.30.0)

- [6801138e](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/6801138e) Prepare for release v0.30.0 (#95)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.32.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.32.0)

- [555bceeb](https://github.com/kubedb/mongodb-restic-plugin/commit/555bceeb) Prepare for release v0.32.0 (#142)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.24.0](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.24.0)

- [02010040](https://github.com/kubedb/mssql-coordinator/commit/02010040) Prepare for release v0.24.0 (#88)
- [62b8f436](https://github.com/kubedb/mssql-coordinator/commit/62b8f436) Reset the cloned Availability Group on a Courier branch (#87)



## [kubedb/mssqlserver](https://github.com/kubedb/mssqlserver)

### [v0.24.0](https://github.com/kubedb/mssqlserver/releases/tag/v0.24.0)

- [2f8b8345](https://github.com/kubedb/mssqlserver/commit/2f8b8345) Prepare for release v0.24.0 (#166)
- [1bc17f10](https://github.com/kubedb/mssqlserver/commit/1bc17f10) Default security IDs from the OpenShift namespace in the operator (#164)
- [c847dd9b](https://github.com/kubedb/mssqlserver/commit/c847dd9b) Add branched mode for KubeDB Courier Branch (#148)



## [kubedb/mssqlserver-archiver](https://github.com/kubedb/mssqlserver-archiver)

### [v0.23.0](https://github.com/kubedb/mssqlserver-archiver/releases/tag/v0.23.0)

- [770a6a3](https://github.com/kubedb/mssqlserver-archiver/commit/770a6a3) Prepare for release v0.23.0 (#40)



## [kubedb/mssqlserver-walg-plugin](https://github.com/kubedb/mssqlserver-walg-plugin)

### [v0.23.0](https://github.com/kubedb/mssqlserver-walg-plugin/releases/tag/v0.23.0)

- [6e60a28b](https://github.com/kubedb/mssqlserver-walg-plugin/commit/6e60a28b) Prepare for release v0.23.0 (#73)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.62.0](https://github.com/kubedb/mysql/releases/tag/v0.62.0)

- [63743fc3](https://github.com/kubedb/mysql/commit/63743fc33) Prepare for release v0.62.0 (#789)
- [2746998f](https://github.com/kubedb/mysql/commit/2746998fc) Default security IDs from the OpenShift namespace in the operator (#788)
- [465cc44a](https://github.com/kubedb/mysql/commit/465cc44a5) add mysql auto reconfigure feature (#787)
- [3ba4b7b6](https://github.com/kubedb/mysql/commit/3ba4b7b6a) Fix Enterprise Reconfig (#785)
- [f12272b9](https://github.com/kubedb/mysql/commit/f12272b95) Add Archiver Restore Ops (#776)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.30.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.30.0)

- [28e3e4fe](https://github.com/kubedb/mysql-archiver/commit/28e3e4fe) Prepare for release v0.30.0 (#120)
- [507e6e59](https://github.com/kubedb/mysql-archiver/commit/507e6e59) Rotate the binlog on an interval so there is something to archive (#114)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.47.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.47.0)

- [04085d70](https://github.com/kubedb/mysql-coordinator/commit/04085d70) Prepare for release v0.47.0 (#197)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.30.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.30.0)

- [189d70ba](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/189d70ba) Prepare for release v0.30.0 (#92)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.32.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.32.0)

- [44527f7f](https://github.com/kubedb/mysql-restic-plugin/commit/44527f7f) Prepare for release v0.32.0 (#128)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.47.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.47.0)

- [9f4eae9](https://github.com/kubedb/mysql-router-init/commit/9f4eae9) Prepare for release v0.47.0 (#74)



## [kubedb/neo4j](https://github.com/kubedb/neo4j)

### [v0.10.0](https://github.com/kubedb/neo4j/releases/tag/v0.10.0)

- [6496f635](https://github.com/kubedb/neo4j/commit/6496f635) Prepare for release v0.10.0 (#73)
- [8d8dc4a7](https://github.com/kubedb/neo4j/commit/8d8dc4a7) Default security IDs from the OpenShift namespace in the operator (#72)
- [1f93acbb](https://github.com/kubedb/neo4j/commit/1f93acbb) Fix issuer-change restart ordering; apply Reconfigure without restart when possible (#71)
- [af15f502](https://github.com/kubedb/neo4j/commit/af15f502) Add branched mode for KubeDB Courier Branch (#49)



## [kubedb/neo4j-backup-plugin](https://github.com/kubedb/neo4j-backup-plugin)

### [v0.5.0](https://github.com/kubedb/neo4j-backup-plugin/releases/tag/v0.5.0)

- [a4b1071](https://github.com/kubedb/neo4j-backup-plugin/commit/a4b1071) Prepare for release v0.5.0 (#18)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.56.0](https://github.com/kubedb/ops-manager/releases/tag/v0.56.0)

- [086ebd0a](https://github.com/kubedb/ops-manager/commit/086ebd0a2) Prepare for release v0.56.0 (#913)
- [0012efcc](https://github.com/kubedb/ops-manager/commit/0012efcc5) Register the Milvus recommendation engine and keep its version updates on the same GPU/CPU flavor (#907)
- [85f2b4da](https://github.com/kubedb/ops-manager/commit/85f2b4da7) Bump apimachinery and milvus to commits reachable from master (#912)
- [d1a942db](https://github.com/kubedb/ops-manager/commit/d1a942db0) add SetupOracleOpsRequestWebhookWithManager (#892)
- [416e7f60](https://github.com/kubedb/ops-manager/commit/416e7f604) Update milvus: stop a new Milvus on a used object storage path (#909)



## [kubedb/oracle](https://github.com/kubedb/oracle)

### [v0.15.0](https://github.com/kubedb/oracle/releases/tag/v0.15.0)

- [654503e5](https://github.com/kubedb/oracle/commit/654503e5) Prepare for release v0.15.0 (#93)
- [80a72951](https://github.com/kubedb/oracle/commit/80a72951) Default security IDs from the OpenShift namespace in the operator (#91)
- [b11e2250](https://github.com/kubedb/oracle/commit/b11e2250) Add minio (#87)



## [kubedb/oracle-backup-plugin](https://github.com/kubedb/oracle-backup-plugin)

### [v0.3.0](https://github.com/kubedb/oracle-backup-plugin/releases/tag/v0.3.0)

- [8c66b47](https://github.com/kubedb/oracle-backup-plugin/commit/8c66b47) Prepare for release v0.3.0 (#12)
- [04745c2](https://github.com/kubedb/oracle-backup-plugin/commit/04745c2) Remove oracle restore flags (#4)
- [8452de4](https://github.com/kubedb/oracle-backup-plugin/commit/8452de4) add minio support (#6)



## [kubedb/oracle-coordinator](https://github.com/kubedb/oracle-coordinator)

### [v0.15.0](https://github.com/kubedb/oracle-coordinator/releases/tag/v0.15.0)

- [7ca76cad](https://github.com/kubedb/oracle-coordinator/commit/7ca76cad) Prepare for release v0.15.0 (#50)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.56.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.56.0)

- [7657afe3](https://github.com/kubedb/percona-xtradb/commit/7657afe33) Prepare for release v0.56.0 (#481)
- [2a83b5fc](https://github.com/kubedb/percona-xtradb/commit/2a83b5fc9) Default security IDs from the OpenShift namespace in the operator (#480)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.42.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.42.0)

- [ebe7cc57](https://github.com/kubedb/percona-xtradb-coordinator/commit/ebe7cc57) Prepare for release v0.42.0 (#141)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.53.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.53.0)




## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.56.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.56.0)

- [21d5f4db](https://github.com/kubedb/pgbouncer/commit/21d5f4dbd) Prepare for release v0.56.0 (#436)
- [5c619eb0](https://github.com/kubedb/pgbouncer/commit/5c619eb0b) Default security IDs from the OpenShift namespace in the operator (#435)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.24.0](https://github.com/kubedb/pgpool/releases/tag/v0.24.0)

- [87428010](https://github.com/kubedb/pgpool/commit/87428010) Prepare for release v0.24.0 (#157)
- [b5d8aab4](https://github.com/kubedb/pgpool/commit/b5d8aab4) Default security IDs from the OpenShift namespace in the operator (#156)
- [d01086ed](https://github.com/kubedb/pgpool/commit/d01086ed) Mirror cross-namespace backend auth secret into Pgpool namespace (#155)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.69.0](https://github.com/kubedb/postgres/releases/tag/v0.69.0)

- [d0e732d0](https://github.com/kubedb/postgres/commit/d0e732d01) Prepare for release v0.69.0 (#968)
- [d73b5f62](https://github.com/kubedb/postgres/commit/d73b5f62d) Find run.sh pid via /proc on RHEL-family images (#935)
- [f5f92e6c](https://github.com/kubedb/postgres/commit/f5f92e6cd) Ignore Issuer events that arrive before the Postgres watchers are initialised (#963)
- [23dc1723](https://github.com/kubedb/postgres/commit/23dc1723f) Default Postgres UID/GID/fsGroup from OpenShift namespace SCC ranges (#961)
- [15f12a76](https://github.com/kubedb/postgres/commit/15f12a760) Set a MajorUpgradeInProgress raft key for the length of a major upgrade (#964)
- [af89100a](https://github.com/kubedb/postgres/commit/af89100a4) Reconfigure: make restart: auto decide, verify reloads through PostgreSQL (#901)
- [f22cdc67](https://github.com/kubedb/postgres/commit/f22cdc67c) Opt-in service account token hardening and coordinator API token/TLS (#955)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.30.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.30.0)

- [b81900f8](https://github.com/kubedb/postgres-archiver/commit/b81900f8) Prepare for release v0.30.0 (#128)
- [d163d666](https://github.com/kubedb/postgres-archiver/commit/d163d666) Add documentdb PITR (#124)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.30.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.30.0)

- [712c418b](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/712c418b) Prepare for release v0.30.0 (#102)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.32.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.32.0)

- [9d84047a](https://github.com/kubedb/postgres-restic-plugin/commit/9d84047a) Prepare for release v0.32.0 (#136)
- [5e485b7b](https://github.com/kubedb/postgres-restic-plugin/commit/5e485b7b) Add documentdb baseBackup (#130)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.30.0](https://github.com/kubedb/provider-aws/releases/tag/v0.30.0)




## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.30.0](https://github.com/kubedb/provider-azure/releases/tag/v0.30.0)




## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.30.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.30.0)




## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.69.0](https://github.com/kubedb/provisioner/releases/tag/v0.69.0)

- [29f2708d](https://github.com/kubedb/provisioner/commit/29f2708df) Prepare for release v0.69.0 (#247)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.56.0](https://github.com/kubedb/proxysql/releases/tag/v0.56.0)

- [da0e7d7d](https://github.com/kubedb/proxysql/commit/da0e7d7d9) Prepare for release v0.56.0 (#458)



## [kubedb/qdrant](https://github.com/kubedb/qdrant)

### [v0.10.0](https://github.com/kubedb/qdrant/releases/tag/v0.10.0)

- [f30d7f24](https://github.com/kubedb/qdrant/commit/f30d7f24) Prepare for release v0.10.0 (#67)



## [kubedb/qdrant-restic-plugin](https://github.com/kubedb/qdrant-restic-plugin)

### [v0.5.0](https://github.com/kubedb/qdrant-restic-plugin/releases/tag/v0.5.0)

- [8d1ccbf](https://github.com/kubedb/qdrant-restic-plugin/commit/8d1ccbf) Prepare for release v0.5.0 (#16)
- [a392d41](https://github.com/kubedb/qdrant-restic-plugin/commit/a392d41) Fix Total Size Issue in Backup Stats (#15)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.24.0](https://github.com/kubedb/rabbitmq/releases/tag/v0.24.0)

- [3b13c743](https://github.com/kubedb/rabbitmq/commit/3b13c743) Prepare for release v0.24.0 (#160)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.62.0](https://github.com/kubedb/redis/releases/tag/v0.62.0)

- [a8b092f4](https://github.com/kubedb/redis/commit/a8b092f4a) Prepare for release v0.62.0 (#686)
- [fdc22f84](https://github.com/kubedb/redis/commit/fdc22f84d) Default security IDs from the OpenShift namespace in the operator (#685)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.48.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.48.0)

- [bb9bfa60](https://github.com/kubedb/redis-coordinator/commit/bb9bfa60) Prepare for release v0.48.0 (#173)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.32.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.32.0)

- [356ac6b2](https://github.com/kubedb/redis-restic-plugin/commit/356ac6b2) Prepare for release v0.32.0 (#120)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.56.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.56.0)

- [47b217d3](https://github.com/kubedb/replication-mode-detector/commit/47b217d3) Prepare for release v0.56.0 (#335)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.45.0](https://github.com/kubedb/schema-manager/releases/tag/v0.45.0)

- [45955851](https://github.com/kubedb/schema-manager/commit/45955851) Prepare for release v0.45.0 (#181)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.24.0](https://github.com/kubedb/singlestore/releases/tag/v0.24.0)

- [94ff056d](https://github.com/kubedb/singlestore/commit/94ff056d) Prepare for release v0.24.0 (#151)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.24.0](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.24.0)

- [eeefe830](https://github.com/kubedb/singlestore-coordinator/commit/eeefe830) Prepare for release v0.24.0 (#85)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.27.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.27.0)

- [8937fd4a](https://github.com/kubedb/singlestore-restic-plugin/commit/8937fd4a) Prepare for release v0.27.0 (#100)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.24.0](https://github.com/kubedb/solr/releases/tag/v0.24.0)

- [b6a0132d](https://github.com/kubedb/solr/commit/b6a0132d) Prepare for release v0.24.0 (#163)



## [kubedb/solr-backup-plugin](https://github.com/kubedb/solr-backup-plugin)

### [v0.3.0](https://github.com/kubedb/solr-backup-plugin/releases/tag/v0.3.0)

- [a8bdafff](https://github.com/kubedb/solr-backup-plugin/commit/a8bdafff) Prepare for release v0.3.0 (#26)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.54.0](https://github.com/kubedb/tests/releases/tag/v0.54.0)

- [baa84ccb](https://github.com/kubedb/tests/commit/baa84ccbc) Prepare for release v0.54.0 (#561)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.45.0](https://github.com/kubedb/ui-server/releases/tag/v0.45.0)

- [877952b5](https://github.com/kubedb/ui-server/commit/877952b55) Prepare for release v0.45.0 (#233)



## [kubedb/weaviate](https://github.com/kubedb/weaviate)

### [v0.10.0](https://github.com/kubedb/weaviate/releases/tag/v0.10.0)

- [d7b47bad](https://github.com/kubedb/weaviate/commit/d7b47bad) Prepare for release v0.10.0 (#66)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.45.0](https://github.com/kubedb/webhook-server/releases/tag/v0.45.0)

- [98434ca4](https://github.com/kubedb/webhook-server/commit/98434ca48) Prepare for release v0.45.0 (#250)
- [c77a27f0](https://github.com/kubedb/webhook-server/commit/c77a27f03) Register the Etcd webhook (#249)



## [kubedb/xtrabackup-restic-plugin](https://github.com/kubedb/xtrabackup-restic-plugin)

### [v0.17.0](https://github.com/kubedb/xtrabackup-restic-plugin/releases/tag/v0.17.0)

- [d3ced9d6](https://github.com/kubedb/xtrabackup-restic-plugin/commit/d3ced9d6) Prepare for release v0.17.0 (#69)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.24.0](https://github.com/kubedb/zookeeper/releases/tag/v0.24.0)

- [5497b6df](https://github.com/kubedb/zookeeper/commit/5497b6df) Prepare for release v0.24.0 (#146)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.24.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.24.0)

- [d5a2c83b](https://github.com/kubedb/zookeeper-restic-plugin/commit/d5a2c83b) Prepare for release v0.24.0 (#84)




