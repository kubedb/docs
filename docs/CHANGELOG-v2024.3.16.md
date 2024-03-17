---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2024.3.16
    name: Changelog-v2024.3.16
    parent: welcome
    weight: 20240316
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2024.3.16/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2024.3.16/
---

# KubeDB v2024.3.16 (2024-03-17)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.44.0](https://github.com/kubedb/apimachinery/releases/tag/v0.44.0)

- [7c6eaeda](https://github.com/kubedb/apimachinery/commit/7c6eaeda0) Update kubestash api
- [c945620f](https://github.com/kubedb/apimachinery/commit/c945620f5) Register dbs in extractDatabaseInfo for petset (#1178)
- [ed0e97c8](https://github.com/kubedb/apimachinery/commit/ed0e97c8d) Add Monitoring support for Solr (#1151)
- [4a4c3980](https://github.com/kubedb/apimachinery/commit/4a4c3980b) Replace StatefulSet to PetSet (#1177)
- [e156d7af](https://github.com/kubedb/apimachinery/commit/e156d7afa) Add ReplicationSlot support for Postgres (#1142)
- [9c51f2dc](https://github.com/kubedb/apimachinery/commit/9c51f2dc0) Update SingleStore Default Memory Resources (#1175)
- [362c68b4](https://github.com/kubedb/apimachinery/commit/362c68b48) Add druid monitoring (#1156)
- [3d020832](https://github.com/kubedb/apimachinery/commit/3d0208322) Add petset support for all new-dbs (#1176)
- [84f90648](https://github.com/kubedb/apimachinery/commit/84f90648a) Remove cves
- [c50127cb](https://github.com/kubedb/apimachinery/commit/c50127cb2) Add ui support to gateway (#1173)
- [bfad33d2](https://github.com/kubedb/apimachinery/commit/bfad33d2e) Add MariaDB Archiver API (#1170)
- [403b7380](https://github.com/kubedb/apimachinery/commit/403b7380b) Add KafkaAutoscaler APIs (#1168)
- [64fea9fd](https://github.com/kubedb/apimachinery/commit/64fea9fd1) Add Pgpool monitoring (#1160)
- [9553f55b](https://github.com/kubedb/apimachinery/commit/9553f55ba) SingleStore Monitoring and UI (#1166)
- [ea1e9882](https://github.com/kubedb/apimachinery/commit/ea1e9882d) Add service gateway info to db status (#1157)
- [5b267a5c](https://github.com/kubedb/apimachinery/commit/5b267a5c4) Persist zookeeper digest credentials. (#1167)
- [f45cf6e6](https://github.com/kubedb/apimachinery/commit/f45cf6e65) Set db-specific exporter port
- [310f4e20](https://github.com/kubedb/apimachinery/commit/310f4e20d) Update client-go deps
- [a28e6433](https://github.com/kubedb/apimachinery/commit/a28e6433f) Set the metrics-port for rabbitmq & zookeeper (#1165)
- [429a2b10](https://github.com/kubedb/apimachinery/commit/429a2b10f) Fix solr helper condition. (#1164)
- [5344c20e](https://github.com/kubedb/apimachinery/commit/5344c20e4) Update APIs for external dependency (#1162)
- [40d63454](https://github.com/kubedb/apimachinery/commit/40d634543) Add Monitoring API (#1153)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.29.0](https://github.com/kubedb/autoscaler/releases/tag/v0.29.0)

- [b2b8afdd](https://github.com/kubedb/autoscaler/commit/b2b8afdd) Prepare for release v0.29.0 (#193)
- [2704a3a9](https://github.com/kubedb/autoscaler/commit/2704a3a9) Use Go 1.22 (#192)
- [8d58e694](https://github.com/kubedb/autoscaler/commit/8d58e694) Prepare for release v0.28.0-rc.0 (#191)
- [9ef5c754](https://github.com/kubedb/autoscaler/commit/9ef5c754) Add Kafka Autoscaler (#190)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.44.0](https://github.com/kubedb/cli/releases/tag/v0.44.0)

- [0a6bf8a8](https://github.com/kubedb/cli/commit/0a6bf8a8) Prepare for release v0.44.0 (#762)
- [23ac68f8](https://github.com/kubedb/cli/commit/23ac68f8) Use Go 1.22 (#761)
- [2d1a52e3](https://github.com/kubedb/cli/commit/2d1a52e3) Refer to remote dashboard files through `--url`
- [3410dd4f](https://github.com/kubedb/cli/commit/3410dd4f) Prepare for release v0.43.0-rc.0 (#760)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.0.8](https://github.com/kubedb/crd-manager/releases/tag/v0.0.8)

- [31204ad](https://github.com/kubedb/crd-manager/commit/31204ad) Prepare for release v0.0.8 (#19)
- [7c48391](https://github.com/kubedb/crd-manager/commit/7c48391) Use Go 1.22 (#18)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.20.0](https://github.com/kubedb/dashboard/releases/tag/v0.20.0)

- [38051a5c](https://github.com/kubedb/dashboard/commit/38051a5c) Prepare for release v0.20.0 (#110)
- [153091f1](https://github.com/kubedb/dashboard/commit/153091f1) Use Go 1.22 (#109)
- [35522dd1](https://github.com/kubedb/dashboard/commit/35522dd1) Prepare for release v0.19.0-rc.0 (#108)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.2.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.2.0)

- [99b5232](https://github.com/kubedb/dashboard-restic-plugin/commit/99b5232) Prepare for release v0.2.0 (#4)
- [124af63](https://github.com/kubedb/dashboard-restic-plugin/commit/124af63) Use Go 1.22 (#3)
- [3b2206a](https://github.com/kubedb/dashboard-restic-plugin/commit/3b2206a) Add Support for Cross Namespace Restore (#2)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.0.13](https://github.com/kubedb/db-client-go/releases/tag/v0.0.13)

- [ce0f5a8b](https://github.com/kubedb/db-client-go/commit/ce0f5a8b) Prepare for release v0.0.13 (#92)
- [7ae91ca4](https://github.com/kubedb/db-client-go/commit/7ae91ca4) Use Go 1.22 (#93)



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.0.8](https://github.com/kubedb/druid/releases/tag/v0.0.8)

- [4a724d1](https://github.com/kubedb/druid/commit/4a724d1) Prepare for release v0.0.8 (#16)
- [42f058d](https://github.com/kubedb/druid/commit/42f058d) Use Go 1.22 (#15)
- [79f1737](https://github.com/kubedb/druid/commit/79f1737) Replace Statefulset with Petset (#14)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.44.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.44.0)

- [e73de14f](https://github.com/kubedb/elasticsearch/commit/e73de14f7) Prepare for release v0.44.0 (#710)
- [2cef036e](https://github.com/kubedb/elasticsearch/commit/2cef036e8) Use Go 1.22 (#709)
- [0c68e7b1](https://github.com/kubedb/elasticsearch/commit/0c68e7b10) Prepare for release v0.43.0-rc.0 (#707)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.7.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.7.0)

- [ecf395f](https://github.com/kubedb/elasticsearch-restic-plugin/commit/ecf395f) Prepare for release v0.7.0 (#25)
- [5c039aa](https://github.com/kubedb/elasticsearch-restic-plugin/commit/5c039aa) Use Go 1.22 (#24)
- [19195b0](https://github.com/kubedb/elasticsearch-restic-plugin/commit/19195b0) Add Support for Cross Namespace Restore (#23)
- [de9ec11](https://github.com/kubedb/elasticsearch-restic-plugin/commit/de9ec11) Prepare for release v0.6.0-rc.0 (#22)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.0.8](https://github.com/kubedb/ferretdb/releases/tag/v0.0.8)

- [84daf7a0](https://github.com/kubedb/ferretdb/commit/84daf7a0) Prepare for release v0.0.8 (#15)
- [a85e325e](https://github.com/kubedb/ferretdb/commit/a85e325e) Use Go 1.22 (#14)
- [34f033dc](https://github.com/kubedb/ferretdb/commit/34f033dc) Add petset to ferretdb (#13)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2024.3.16](https://github.com/kubedb/installer/releases/tag/v2024.3.16)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.15.0](https://github.com/kubedb/kafka/releases/tag/v0.15.0)

- [b41b5d3e](https://github.com/kubedb/kafka/commit/b41b5d3e) Prepare for release v0.15.0 (#82)
- [44839f29](https://github.com/kubedb/kafka/commit/44839f29) Use Go 1.22 (#81)
- [e553c579](https://github.com/kubedb/kafka/commit/e553c579) Update daily tests to run with versions 3.5.2,3.6.1
- [c7038c02](https://github.com/kubedb/kafka/commit/c7038c02) Prepare for release v0.14.0-rc.0 (#80)
- [089c99b7](https://github.com/kubedb/kafka/commit/089c99b7) Fix Kafka node specific resource (#79)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.7.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.7.0)

- [b67ad6b](https://github.com/kubedb/kubedb-manifest-plugin/commit/b67ad6b) Prepare for release v0.7.0 (#47)
- [ec38640](https://github.com/kubedb/kubedb-manifest-plugin/commit/ec38640) Use Go 1.22 (#46)
- [8f8df05](https://github.com/kubedb/kubedb-manifest-plugin/commit/8f8df05) Prepare for release v0.6.0-rc.0 (#44)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.28.0](https://github.com/kubedb/mariadb/releases/tag/v0.28.0)

- [3eeabaee](https://github.com/kubedb/mariadb/commit/3eeabaeed) Prepare for release v0.28.0 (#262)
- [f40e7039](https://github.com/kubedb/mariadb/commit/f40e70398) Use Go 1.22 (#261)
- [08d7efaa](https://github.com/kubedb/mariadb/commit/08d7efaa0) Implement achiver backup and recovery (#183)
- [1389a6f8](https://github.com/kubedb/mariadb/commit/1389a6f85) Prepare for release v0.27.0-rc.0 (#260)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.4.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.4.0)

- [9f5b224](https://github.com/kubedb/mariadb-archiver/commit/9f5b224) Prepare for release v0.4.0 (#12)
- [acc612b](https://github.com/kubedb/mariadb-archiver/commit/acc612b) Use Go 1.22 (#11)
- [0ea457c](https://github.com/kubedb/mariadb-archiver/commit/0ea457c) Use GTID for binlog recovery (#2)
- [4e9bb56](https://github.com/kubedb/mariadb-archiver/commit/4e9bb56) Prepare for release v0.3.0-rc.0 (#10)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.24.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.24.0)

- [527b4c21](https://github.com/kubedb/mariadb-coordinator/commit/527b4c21) Prepare for release v0.24.0 (#112)
- [eea3b5b6](https://github.com/kubedb/mariadb-coordinator/commit/eea3b5b6) Use Go 1.22 (#111)
- [2380c6e5](https://github.com/kubedb/mariadb-coordinator/commit/2380c6e5) Prepare for release v0.23.0-rc.0 (#110)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.4.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.4.0)

- [4ac0a34](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/4ac0a34) Prepare for release v0.4.0 (#16)
- [a1bd3d7](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/a1bd3d7) Use Go 1.22 (#15)
- [d385351](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/d385351) Prepare for release v0.3.0-rc.0 (#14)
- [69c4631](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/69c4631) Set volumesnapshot time (#10)
- [be6eba4](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/be6eba4) Refactor (#13)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.2.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.2.0)

- [7aabe3e](https://github.com/kubedb/mariadb-restic-plugin/commit/7aabe3e) Prepare for release v0.2.0 (#4)
- [5a04032](https://github.com/kubedb/mariadb-restic-plugin/commit/5a04032) Use Go 1.22 (#3)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.37.0](https://github.com/kubedb/memcached/releases/tag/v0.37.0)

- [07fe6bc8](https://github.com/kubedb/memcached/commit/07fe6bc8) Prepare for release v0.37.0 (#427)
- [e57f9f2c](https://github.com/kubedb/memcached/commit/e57f9f2c) Use Go 1.22 (#426)
- [2a54bc53](https://github.com/kubedb/memcached/commit/2a54bc53) Prepare for release v0.36.0-rc.0 (#425)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.37.0](https://github.com/kubedb/mongodb/releases/tag/v0.37.0)

- [9e5db160](https://github.com/kubedb/mongodb/commit/9e5db160e) Prepare for release v0.37.0 (#620)
- [2ded92d1](https://github.com/kubedb/mongodb/commit/2ded92d15) Use Go 1.22 (#619)
- [58e27888](https://github.com/kubedb/mongodb/commit/58e278887) Prepare for release v0.36.0-rc.0 (#617)
- [1ead0453](https://github.com/kubedb/mongodb/commit/1ead04536) ignore mongos repl mode detector (#616)
- [c0797130](https://github.com/kubedb/mongodb/commit/c0797130c) Add replication mode detector container for shard (#614)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.5.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.5.0)

- [a2ca0f0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/a2ca0f0) Prepare for release v0.5.0 (#21)
- [e755ddd](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/e755ddd) Use Go 1.22 (#20)
- [83b0a1d](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/83b0a1d) Prepare for release v0.4.0-rc.0 (#19)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.7.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.7.0)

- [c2c3339](https://github.com/kubedb/mongodb-restic-plugin/commit/c2c3339) Prepare for release v0.7.0 (#36)
- [f997a0e](https://github.com/kubedb/mongodb-restic-plugin/commit/f997a0e) Use Go 1.22 (#35)
- [c867982](https://github.com/kubedb/mongodb-restic-plugin/commit/c867982) Fix wait for db ready (#34)
- [973824e](https://github.com/kubedb/mongodb-restic-plugin/commit/973824e) Add Support for Cross Namespace Restore (#32)
- [8f5c92d](https://github.com/kubedb/mongodb-restic-plugin/commit/8f5c92d) Prepare for release v0.6.0-rc.0 (#31)
- [e790456](https://github.com/kubedb/mongodb-restic-plugin/commit/e790456) fix order of lock and unlock secondary (#30)
- [a0281df](https://github.com/kubedb/mongodb-restic-plugin/commit/a0281df) Fix component-name issue for sharded cluster (#29)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.37.0](https://github.com/kubedb/mysql/releases/tag/v0.37.0)

- [04bb333c](https://github.com/kubedb/mysql/commit/04bb333cb) Prepare for release v0.37.0 (#616)
- [1dc0bff5](https://github.com/kubedb/mysql/commit/1dc0bff5d) Use Go 1.22 (#615)
- [7e213885](https://github.com/kubedb/mysql/commit/7e2138855) Prepare for release v0.36.0-rc.0 (#613)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.5.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.5.0)

- [3cab3f9](https://github.com/kubedb/mysql-archiver/commit/3cab3f9) Prepare for release v0.5.0 (#26)
- [199c6a9](https://github.com/kubedb/mysql-archiver/commit/199c6a9) Use Go 1.22 (#25)
- [00c0f0e](https://github.com/kubedb/mysql-archiver/commit/00c0f0e) Prepare for release v0.4.0-rc.0 (#24)
- [9fe9cd2](https://github.com/kubedb/mysql-archiver/commit/9fe9cd2) Fix Multiple Timeline not Restoring Issue (#23)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.22.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.22.0)

- [8a76734f](https://github.com/kubedb/mysql-coordinator/commit/8a76734f) Prepare for release v0.22.0 (#107)
- [0fd6ebc4](https://github.com/kubedb/mysql-coordinator/commit/0fd6ebc4) Use Go 1.22 (#106)
- [9da51c98](https://github.com/kubedb/mysql-coordinator/commit/9da51c98) Prepare for release v0.21.0-rc.0 (#105)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.5.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.5.0)

- [bafa2ae](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/bafa2ae) Prepare for release v0.5.0 (#14)
- [0afef2c](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/0afef2c) Use Go 1.22 (#13)
- [718dbc2](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/718dbc2) Prepare for release v0.4.0-rc.0 (#12)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.7.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.7.0)

- [200364a](https://github.com/kubedb/mysql-restic-plugin/commit/200364a) Prepare for release v0.7.0 (#31)
- [16a4bf6](https://github.com/kubedb/mysql-restic-plugin/commit/16a4bf6) Use Go 1.22 (#30)
- [21c0a09](https://github.com/kubedb/mysql-restic-plugin/commit/21c0a09) Add Support for Cross Namespace Restore (#29)
- [3dcb7b8](https://github.com/kubedb/mysql-restic-plugin/commit/3dcb7b8) Prepare for release v0.6.0-rc.0 (#28)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.22.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.22.0)

- [f1af6f6](https://github.com/kubedb/mysql-router-init/commit/f1af6f6) Use Go 1.22 (#41)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.31.0](https://github.com/kubedb/ops-manager/releases/tag/v0.31.0)

- [f6d894ce](https://github.com/kubedb/ops-manager/commit/f6d894ce3) Prepare for release v0.31.0 (#548)
- [18d77621](https://github.com/kubedb/ops-manager/commit/18d77621e) Update go.sum
- [ddc2a48c](https://github.com/kubedb/ops-manager/commit/ddc2a48c0) Use Go 1.22 (#547)
- [ba77c1fe](https://github.com/kubedb/ops-manager/commit/ba77c1feb) Update deps
- [f41286bd](https://github.com/kubedb/ops-manager/commit/f41286bd2) Drop replication slot on down scaling (#546)
- [11e19df1](https://github.com/kubedb/ops-manager/commit/11e19df16) Add scheme for scanner api (#545)
- [d038e38d](https://github.com/kubedb/ops-manager/commit/d038e38db) use mariadb-upgrade (#544)
- [062c3be7](https://github.com/kubedb/ops-manager/commit/062c3be72) Update tests to run Kafka versions 3.5.2,3.6.1
- [ac5f3271](https://github.com/kubedb/ops-manager/commit/ac5f32711) Fix Kafka Horizontal Scale Failing for Quorum voter config conflicts (#543)
- [f6c944e9](https://github.com/kubedb/ops-manager/commit/f6c944e92) Prepare for release v0.30.0-rc.0 (#542)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.31.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.31.0)

- [3ef2f278](https://github.com/kubedb/percona-xtradb/commit/3ef2f2783) Prepare for release v0.31.0 (#360)
- [c2eb50bd](https://github.com/kubedb/percona-xtradb/commit/c2eb50bd9) Use Go 1.22 (#359)
- [d69b6040](https://github.com/kubedb/percona-xtradb/commit/d69b6040c) Prepare for release v0.30.0-rc.0 (#357)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.17.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.17.0)

- [f5a482c0](https://github.com/kubedb/percona-xtradb-coordinator/commit/f5a482c0) Prepare for release v0.17.0 (#67)
- [55471a58](https://github.com/kubedb/percona-xtradb-coordinator/commit/55471a58) Use Go 1.22 (#66)
- [93f7f940](https://github.com/kubedb/percona-xtradb-coordinator/commit/93f7f940) Prepare for release v0.16.0-rc.0 (#65)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.28.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.28.0)

- [78ff38e8](https://github.com/kubedb/pg-coordinator/commit/78ff38e8) Prepare for release v0.28.0 (#160)
- [69bb82d7](https://github.com/kubedb/pg-coordinator/commit/69bb82d7) Use Go 1.22 (#159)
- [2f1ef049](https://github.com/kubedb/pg-coordinator/commit/2f1ef049) Add support for postgres replication slot (#158)
- [d1f6731b](https://github.com/kubedb/pg-coordinator/commit/d1f6731b) Prepare for release v0.27.0-rc.0 (#157)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.31.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.31.0)

- [cb3f3cb6](https://github.com/kubedb/pgbouncer/commit/cb3f3cb6) Prepare for release v0.31.0 (#321)
- [2daf8fc5](https://github.com/kubedb/pgbouncer/commit/2daf8fc5) Use Go 1.22 (#320)
- [b5e9f18b](https://github.com/kubedb/pgbouncer/commit/b5e9f18b) Prepare for release v0.30.0-rc.0 (#319)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.0.8](https://github.com/kubedb/pgpool/releases/tag/v0.0.8)

- [6b1ca42](https://github.com/kubedb/pgpool/commit/6b1ca42) Prepare for release v0.0.8 (#20)
- [7ab4f4b](https://github.com/kubedb/pgpool/commit/7ab4f4b) Use Go 1.22 (#19)
- [6d8d0ce](https://github.com/kubedb/pgpool/commit/6d8d0ce) Replace StatefulSet with PetSet (#18)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.44.0](https://github.com/kubedb/postgres/releases/tag/v0.44.0)

- [009d26a2](https://github.com/kubedb/postgres/commit/009d26a27) Prepare for release v0.44.0 (#724)
- [c811db17](https://github.com/kubedb/postgres/commit/c811db175) Use Go 1.22 (#723)
- [62d92776](https://github.com/kubedb/postgres/commit/62d92776c) Add Support for PG Replication Slot (#722)
- [6b772e80](https://github.com/kubedb/postgres/commit/6b772e809) Prepare for release v0.43.0-rc.0 (#720)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.5.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.5.0)

- [4598539](https://github.com/kubedb/postgres-archiver/commit/4598539) Prepare for release v0.5.0 (#25)
- [df6a88a](https://github.com/kubedb/postgres-archiver/commit/df6a88a) Use Go 1.22 (#24)
- [95a3f5a](https://github.com/kubedb/postgres-archiver/commit/95a3f5a) Prepare for release v0.4.0-rc.0 (#23)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.5.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.5.0)

- [c563225](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/c563225) Prepare for release v0.5.0 (#23)
- [682604d](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/682604d) Use Go 1.22 (#22)
- [7dfe005](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/7dfe005) Prepare for release v0.4.0-rc.0 (#21)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.7.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.7.0)

- [2d96c3e](https://github.com/kubedb/postgres-restic-plugin/commit/2d96c3e) Prepare for release v0.7.0 (#24)
- [4844eca](https://github.com/kubedb/postgres-restic-plugin/commit/4844eca) Use Go 1.22 (#23)
- [fbf6f61](https://github.com/kubedb/postgres-restic-plugin/commit/fbf6f61) Add Support for Cross Namespace Restore (#21)
- [ed4a89c](https://github.com/kubedb/postgres-restic-plugin/commit/ed4a89c) Prepare for release v0.6.0-rc.0 (#20)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.6.0](https://github.com/kubedb/provider-aws/releases/tag/v0.6.0)

- [2215aee](https://github.com/kubedb/provider-aws/commit/2215aee) Use Go 1.22 (#13)



## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.6.0](https://github.com/kubedb/provider-azure/releases/tag/v0.6.0)

- [d62697e](https://github.com/kubedb/provider-azure/commit/d62697e) Use Go 1.22 (#4)



## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.6.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.6.0)

- [c8f6d83](https://github.com/kubedb/provider-gcp/commit/c8f6d83) Use Go 1.22 (#4)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.44.0](https://github.com/kubedb/provisioner/releases/tag/v0.44.0)




## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.31.0](https://github.com/kubedb/proxysql/releases/tag/v0.31.0)

- [059103a0](https://github.com/kubedb/proxysql/commit/059103a09) Prepare for release v0.31.0 (#339)
- [96e0d84d](https://github.com/kubedb/proxysql/commit/96e0d84d3) Use Go 1.22 (#338)
- [b2e583b8](https://github.com/kubedb/proxysql/commit/b2e583b8a) Prepare for release v0.30.0-rc.0 (#337)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.0.9](https://github.com/kubedb/rabbitmq/releases/tag/v0.0.9)

- [21da11ea](https://github.com/kubedb/rabbitmq/commit/21da11ea) Prepare for release v0.0.9 (#16)
- [813c533c](https://github.com/kubedb/rabbitmq/commit/813c533c) Use Go 1.22 (#15)
- [8a6b0461](https://github.com/kubedb/rabbitmq/commit/8a6b0461) Update controller to create Petset replacing statefulset (#14)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.37.0](https://github.com/kubedb/redis/releases/tag/v0.37.0)

- [5c9adf10](https://github.com/kubedb/redis/commit/5c9adf105) Prepare for release v0.37.0 (#531)
- [9337d58e](https://github.com/kubedb/redis/commit/9337d58e1) Use Go 1.22 (#530)
- [62fb984d](https://github.com/kubedb/redis/commit/62fb984dc) Prepare for release v0.36.0-rc.0 (#528)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.23.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.23.0)

- [0acf3ba9](https://github.com/kubedb/redis-coordinator/commit/0acf3ba9) Prepare for release v0.23.0 (#98)
- [7cf38ea9](https://github.com/kubedb/redis-coordinator/commit/7cf38ea9) Use Go 1.22 (#97)
- [71f84047](https://github.com/kubedb/redis-coordinator/commit/71f84047) Prepare for release v0.22.0-rc.0 (#96)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.7.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.7.0)

- [86659d3](https://github.com/kubedb/redis-restic-plugin/commit/86659d3) Prepare for release v0.7.0 (#26)
- [7fe1b93](https://github.com/kubedb/redis-restic-plugin/commit/7fe1b93) Use Go 1.22 (#25)
- [2166c18](https://github.com/kubedb/redis-restic-plugin/commit/2166c18) Add Support for Cross Namespace Restore (#24)
- [0abdc75](https://github.com/kubedb/redis-restic-plugin/commit/0abdc75) Prepare for release v0.6.0-rc.0 (#23)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.31.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.31.0)

- [4c30ec3c](https://github.com/kubedb/replication-mode-detector/commit/4c30ec3c) Prepare for release v0.31.0 (#264)
- [17de6c25](https://github.com/kubedb/replication-mode-detector/commit/17de6c25) Use Go 1.22 (#263)
- [82431296](https://github.com/kubedb/replication-mode-detector/commit/82431296) Prepare for release v0.30.0-rc.0 (#261)
- [397f06c4](https://github.com/kubedb/replication-mode-detector/commit/397f06c4) Add label for shard (#260)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.20.0](https://github.com/kubedb/schema-manager/releases/tag/v0.20.0)

- [0dc3955f](https://github.com/kubedb/schema-manager/commit/0dc3955f) Prepare for release v0.20.0 (#106)
- [4585101f](https://github.com/kubedb/schema-manager/commit/4585101f) Use Go 1.22 (#105)
- [0b11c0ca](https://github.com/kubedb/schema-manager/commit/0b11c0ca) Prepare for release v0.19.0-rc.0 (#104)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.0.7](https://github.com/kubedb/singlestore/releases/tag/v0.0.7)

- [618ee09](https://github.com/kubedb/singlestore/commit/618ee09) Prepare for release v0.0.7 (#18)
- [2da9914](https://github.com/kubedb/singlestore/commit/2da9914) Use Go 1.22 (#17)
- [97f5098](https://github.com/kubedb/singlestore/commit/97f5098) Add PetSet (#16)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.0.7](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.0.7)

- [a61a41d](https://github.com/kubedb/singlestore-coordinator/commit/a61a41d) Prepare for release v0.0.7 (#11)
- [e21b3c5](https://github.com/kubedb/singlestore-coordinator/commit/e21b3c5) Use Go 1.22 (#10)
- [d7f3da6](https://github.com/kubedb/singlestore-coordinator/commit/d7f3da6) Add PetSet (#9)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.2.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.2.0)

- [7a50d9d](https://github.com/kubedb/singlestore-restic-plugin/commit/7a50d9d) Prepare for release v0.2.0 (#4)
- [c1e25b8](https://github.com/kubedb/singlestore-restic-plugin/commit/c1e25b8) Use Go 1.22 (#3)
- [36153bf](https://github.com/kubedb/singlestore-restic-plugin/commit/36153bf) Add Support for Cross Namespace Restore (#2)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.0.9](https://github.com/kubedb/solr/releases/tag/v0.0.9)

- [08ff522](https://github.com/kubedb/solr/commit/08ff522) Prepare for release v0.0.9 (#18)
- [a9995f0](https://github.com/kubedb/solr/commit/a9995f0) Use Go 1.22 (#17)
- [a8e2f71](https://github.com/kubedb/solr/commit/a8e2f71) Add monitoring and petset. (#11)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.29.0](https://github.com/kubedb/tests/releases/tag/v0.29.0)

- [81fec4d3](https://github.com/kubedb/tests/commit/81fec4d3) Prepare for release v0.29.0 (#313)
- [322dabb4](https://github.com/kubedb/tests/commit/322dabb4) Use Go 1.22 (#312)
- [f3de558d](https://github.com/kubedb/tests/commit/f3de558d) Add Kafka Connect Cluster tests (#290)
- [79f0cad6](https://github.com/kubedb/tests/commit/79f0cad6) Prepare for release v0.28.0-rc.0 (#310)
- [ed951ee1](https://github.com/kubedb/tests/commit/ed951ee1) MongoDB reprovision test (#309)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.20.0](https://github.com/kubedb/ui-server/releases/tag/v0.20.0)

- [0f1653d8](https://github.com/kubedb/ui-server/commit/0f1653d8) Prepare for release v0.20.0 (#115)
- [1da6ab0f](https://github.com/kubedb/ui-server/commit/1da6ab0f) Use Go 1.22 (#114)
- [f7a7ee62](https://github.com/kubedb/ui-server/commit/f7a7ee62) Prepare for release v0.19.0-rc.0 (#113)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.20.0](https://github.com/kubedb/webhook-server/releases/tag/v0.20.0)

- [65fd596e](https://github.com/kubedb/webhook-server/commit/65fd596e) Prepare for release v0.20.0 (#100)
- [98874df4](https://github.com/kubedb/webhook-server/commit/98874df4) Use Go 1.22 (#99)
- [3b6b6fc2](https://github.com/kubedb/webhook-server/commit/3b6b6fc2) Add Kafka Autoscaler Webhook (#98)
- [c9031073](https://github.com/kubedb/webhook-server/commit/c9031073) Prepare for release v0.19.0-rc.0 (#97)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.0.8](https://github.com/kubedb/zookeeper/releases/tag/v0.0.8)

- [ed34766](https://github.com/kubedb/zookeeper/commit/ed34766) Prepare for release v0.0.8 (#15)
- [7f507e7](https://github.com/kubedb/zookeeper/commit/7f507e7) Use Go 1.22 (#14)
- [bbe937a](https://github.com/kubedb/zookeeper/commit/bbe937a) Add petset support (#13)




