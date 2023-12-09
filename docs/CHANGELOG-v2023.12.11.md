---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2023.12.11
    name: Changelog-v2023.12.11
    parent: welcome
    weight: 20231211
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2023.12.11/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2023.12.11/
---

# KubeDB v2023.12.11 (2023-12-08)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.38.0](https://github.com/kubedb/apimachinery/releases/tag/v0.38.0)

- [566c617f](https://github.com/kubedb/apimachinery/commit/566c617f) Update kafka webhook mutating verb (#1084)
- [c6ac8def](https://github.com/kubedb/apimachinery/commit/c6ac8def) Add IPS_LOCK and SYS_RESOURCE (#1083)
- [96238937](https://github.com/kubedb/apimachinery/commit/96238937) Add Postgres arbiter spec (#1082)
- [24013ada](https://github.com/kubedb/apimachinery/commit/24013ada) Fix update-crds wf
- [de0bb4e2](https://github.com/kubedb/apimachinery/commit/de0bb4e2) Update kubestash apimachienry
- [545731a9](https://github.com/kubedb/apimachinery/commit/545731a9) Add default KubeBuilder client (#1081)
- [f260aa8e](https://github.com/kubedb/apimachinery/commit/f260aa8e) Add SecurityContext field in catalogs; Set default accordingly (#1080)
- [e070a3ae](https://github.com/kubedb/apimachinery/commit/e070a3ae) Do not default the seccompProfile (#1079)
- [29c96031](https://github.com/kubedb/apimachinery/commit/29c96031) Set Default Security Context for MariaDB (#1077)
- [fc35d376](https://github.com/kubedb/apimachinery/commit/fc35d376) Set default SecurityContext for mysql (#1070)
- [ee71aca0](https://github.com/kubedb/apimachinery/commit/ee71aca0) Update dependencies
- [93b5ba51](https://github.com/kubedb/apimachinery/commit/93b5ba51) add encriptSecret to postgresAchiver (#1078)
- [2b06b6e5](https://github.com/kubedb/apimachinery/commit/2b06b6e5) Add mongodb & postgres archiver (#1016)
- [47793c9a](https://github.com/kubedb/apimachinery/commit/47793c9a) Set default  SecurityContext for Elasticsearch. (#1072)
- [90567b46](https://github.com/kubedb/apimachinery/commit/90567b46) Set default SecurityContext for Kafka (#1068)
- [449a4e00](https://github.com/kubedb/apimachinery/commit/449a4e00) Remove redundant helper functions for Kafka and Update constants (#1074)
- [b28463f4](https://github.com/kubedb/apimachinery/commit/b28463f4) Set fsGroup to 999 to avoid mountedPath's files permission issue in different storageClass (#1075)
- [8e497b92](https://github.com/kubedb/apimachinery/commit/8e497b92) Set Default Security Context for Redis (#1073)
- [88ab93c7](https://github.com/kubedb/apimachinery/commit/88ab93c7) Set default SecurityContext for mongodb (#1067)
- [e7ac5d2e](https://github.com/kubedb/apimachinery/commit/e7ac5d2e) Set default for security Context for postgres (#1069)
- [f5de4a28](https://github.com/kubedb/apimachinery/commit/f5de4a28) Add support for init with git-sync; Add const (#1065)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.23.0](https://github.com/kubedb/autoscaler/releases/tag/v0.23.0)

- [d7c1af24](https://github.com/kubedb/autoscaler/commit/d7c1af24) Prepare for release v0.23.0 (#160)
- [193fb07b](https://github.com/kubedb/autoscaler/commit/193fb07b) Prepare for release v0.23.0-rc.1 (#159)
- [a406fbda](https://github.com/kubedb/autoscaler/commit/a406fbda) Prepare for release v0.23.0-rc.0 (#158)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.38.0](https://github.com/kubedb/cli/releases/tag/v0.38.0)

- [8c968939](https://github.com/kubedb/cli/commit/8c968939) Prepare for release v0.38.0 (#739)
- [a99b2857](https://github.com/kubedb/cli/commit/a99b2857) Prepare for release v0.38.0-rc.1 (#738)
- [3a4dcc47](https://github.com/kubedb/cli/commit/3a4dcc47) Prepare for release v0.38.0-rc.0 (#737)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.14.0](https://github.com/kubedb/dashboard/releases/tag/v0.14.0)

- [7741d24d](https://github.com/kubedb/dashboard/commit/7741d24d) Prepare for release v0.14.0 (#87)
- [7031fb23](https://github.com/kubedb/dashboard/commit/7031fb23) Prepare for release v0.14.0-rc.1 (#86)
- [c2982e93](https://github.com/kubedb/dashboard/commit/c2982e93) Prepare for release v0.14.0-rc.0 (#85)
- [9a9e6cd9](https://github.com/kubedb/dashboard/commit/9a9e6cd9) Add container security context for elasticsearch dashboard. (#84)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.38.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.38.0)

- [da1e77ef](https://github.com/kubedb/elasticsearch/commit/da1e77ef4) Prepare for release v0.38.0 (#681)
- [aec25e8a](https://github.com/kubedb/elasticsearch/commit/aec25e8a9) Add new version in elasticsearch yaml. (#679)
- [bd0fd357](https://github.com/kubedb/elasticsearch/commit/bd0fd357e) Prepare for release v0.38.0-rc.1 (#680)
- [6b2943f1](https://github.com/kubedb/elasticsearch/commit/6b2943f19) Prepare for release v0.38.0-rc.0 (#678)
- [7f1a37e1](https://github.com/kubedb/elasticsearch/commit/7f1a37e1a) Add prepare cluster installer before test runner (#677)
- [1d49f16d](https://github.com/kubedb/elasticsearch/commit/1d49f16d2) Remove `init-sysctl` container and add default containerSecurityContext (#676)
- [4bb15e48](https://github.com/kubedb/elasticsearch/commit/4bb15e48b) Update daily-opensearch workflow to provision v1.3.13



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.1.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.1.0)

- [1d1abdd](https://github.com/kubedb/elasticsearch-restic-plugin/commit/1d1abdd) Prepare for release v0.1.0 (#10)
- [f6a9e4c](https://github.com/kubedb/elasticsearch-restic-plugin/commit/f6a9e4c) Prepare for release v0.1.0-rc.1 (#9)
- [eb95c84](https://github.com/kubedb/elasticsearch-restic-plugin/commit/eb95c84) Prepare for release v0.1.0-rc.0 (#8)
- [fe82e1b](https://github.com/kubedb/elasticsearch-restic-plugin/commit/fe82e1b) Update component name (#7)
- [c155643](https://github.com/kubedb/elasticsearch-restic-plugin/commit/c155643) Update snapshot time (#6)
- [7093d5a](https://github.com/kubedb/elasticsearch-restic-plugin/commit/7093d5a) Move to kubedb org
- [a3a079e](https://github.com/kubedb/elasticsearch-restic-plugin/commit/a3a079e) Update deps (#5)
- [7a0fd38](https://github.com/kubedb/elasticsearch-restic-plugin/commit/7a0fd38) Refactor (#4)
- [b262635](https://github.com/kubedb/elasticsearch-restic-plugin/commit/b262635) Add support for backup and restore (#1)
- [50bde7e](https://github.com/kubedb/elasticsearch-restic-plugin/commit/50bde7e) Fix build
- [b9686b7](https://github.com/kubedb/elasticsearch-restic-plugin/commit/b9686b7) Prepare for release v0.1.0-rc.0 (#3)
- [ba0c0ed](https://github.com/kubedb/elasticsearch-restic-plugin/commit/ba0c0ed) Fix binary name
- [b0aa991](https://github.com/kubedb/elasticsearch-restic-plugin/commit/b0aa991) Use firecracker runner
- [a621400](https://github.com/kubedb/elasticsearch-restic-plugin/commit/a621400) Use Go 1.21 and restic 0.16.0
- [f08e4e8](https://github.com/kubedb/elasticsearch-restic-plugin/commit/f08e4e8) Use github runner to push docker image



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2023.12.11](https://github.com/kubedb/installer/releases/tag/v2023.12.11)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.9.0](https://github.com/kubedb/kafka/releases/tag/v0.9.0)

- [9c62eb1](https://github.com/kubedb/kafka/commit/9c62eb1) Prepare for release v0.9.0 (#52)
- [8ddb2b8](https://github.com/kubedb/kafka/commit/8ddb2b8) Remove hardcoded fsgroup from statefulset (#51)
- [0516c18](https://github.com/kubedb/kafka/commit/0516c18) Prepare for release v0.9.0-rc.1 (#50)
- [6554778](https://github.com/kubedb/kafka/commit/6554778) Set default KubeBuilder client (#49)
- [0770fff](https://github.com/kubedb/kafka/commit/0770fff) Prepare for release v0.9.0-rc.0 (#48)
- [ee3dcf5](https://github.com/kubedb/kafka/commit/ee3dcf5) Add condition for ssl.properties file (#47)
- [4bd632b](https://github.com/kubedb/kafka/commit/4bd632b) Reconfigure kafka for updated config properties (#45)
- [cc9795b](https://github.com/kubedb/kafka/commit/cc9795b) Upsert Init Containers with Kafka podtemplate.spec and update default test-profile (#43)
- [76e743c](https://github.com/kubedb/kafka/commit/76e743c) Update daily e2e tests yml (#42)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.1.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.1.0)

- [2dd0a52](https://github.com/kubedb/kubedb-manifest-plugin/commit/2dd0a52) Prepare for release v0.1.0 (#30)
- [4bd44b8](https://github.com/kubedb/kubedb-manifest-plugin/commit/4bd44b8) Prepare for release v0.1.0-rc.1 (#29)
- [bef777c](https://github.com/kubedb/kubedb-manifest-plugin/commit/bef777c) Prepare for release v0.1.0-rc.0 (#28)
- [46ad967](https://github.com/kubedb/kubedb-manifest-plugin/commit/46ad967) Remove redundancy (#27)
- [4eaf765](https://github.com/kubedb/kubedb-manifest-plugin/commit/4eaf765) Update snapshot time (#26)
- [e8ace42](https://github.com/kubedb/kubedb-manifest-plugin/commit/e8ace42) Fix plugin binary name
- [d4e3c34](https://github.com/kubedb/kubedb-manifest-plugin/commit/d4e3c34) Move to kubedb org
- [15770b2](https://github.com/kubedb/kubedb-manifest-plugin/commit/15770b2) Update deps (#25)
- [f50a3af](https://github.com/kubedb/kubedb-manifest-plugin/commit/f50a3af) Fix directory cleanup (#24)
- [d41eba7](https://github.com/kubedb/kubedb-manifest-plugin/commit/d41eba7) Refactor
- [0e154e7](https://github.com/kubedb/kubedb-manifest-plugin/commit/0e154e7) Fix release workflow
- [35c6b95](https://github.com/kubedb/kubedb-manifest-plugin/commit/35c6b95) Prepare for release v0.2.0-rc.0 (#22)
- [da97d9a](https://github.com/kubedb/kubedb-manifest-plugin/commit/da97d9a) Use gh runner token to publish image
- [592c51f](https://github.com/kubedb/kubedb-manifest-plugin/commit/592c51f) Use firecracker runner
- [008042d](https://github.com/kubedb/kubedb-manifest-plugin/commit/008042d) Use Go 1.21
- [985bcab](https://github.com/kubedb/kubedb-manifest-plugin/commit/985bcab) Set snapshot time after snapshot completed (#21)
- [6a8c682](https://github.com/kubedb/kubedb-manifest-plugin/commit/6a8c682) Refactor code (#20)
- [bcb944d](https://github.com/kubedb/kubedb-manifest-plugin/commit/bcb944d) Remove manifest option flags (#19)
- [5a47722](https://github.com/kubedb/kubedb-manifest-plugin/commit/5a47722) Fix secret restore issue (#18)
- [3ced8b7](https://github.com/kubedb/kubedb-manifest-plugin/commit/3ced8b7) Update `kmodules.xyz/client-go` version to `v0.25.27` (#17)
- [2ee1314](https://github.com/kubedb/kubedb-manifest-plugin/commit/2ee1314) Update Readme (#16)
- [42d0e52](https://github.com/kubedb/kubedb-manifest-plugin/commit/42d0e52) Set initial component status prior to backup and restore (#15)
- [31a64d6](https://github.com/kubedb/kubedb-manifest-plugin/commit/31a64d6) Remove redundant flags (#14)
- [a804ba8](https://github.com/kubedb/kubedb-manifest-plugin/commit/a804ba8) Pass Snapshot name for restore
- [99ca49f](https://github.com/kubedb/kubedb-manifest-plugin/commit/99ca49f) Set snapshot time, integrity and size (#12)
- [384bbb6](https://github.com/kubedb/kubedb-manifest-plugin/commit/384bbb6) Set backup error in component status + Refactor codebase (#11)
- [513eef5](https://github.com/kubedb/kubedb-manifest-plugin/commit/513eef5) Update for snapshot and restoresession API changes (#10)
- [4fb8f52](https://github.com/kubedb/kubedb-manifest-plugin/commit/4fb8f52) Add options for issuerref (#9)
- [2931d9e](https://github.com/kubedb/kubedb-manifest-plugin/commit/2931d9e) Update restic modules (#7)
- [3422ddf](https://github.com/kubedb/kubedb-manifest-plugin/commit/3422ddf) Fix bugs + Sync with updated snapshot api (#6)
- [b1a69b5](https://github.com/kubedb/kubedb-manifest-plugin/commit/b1a69b5) Prepare for release v0.1.0 (#5)
- [5344e9f](https://github.com/kubedb/kubedb-manifest-plugin/commit/5344e9f) Update modules (#4)
- [14b2797](https://github.com/kubedb/kubedb-manifest-plugin/commit/14b2797) Add CI badge
- [969eeda](https://github.com/kubedb/kubedb-manifest-plugin/commit/969eeda) Organize code structure (#3)
- [9fc3cbe](https://github.com/kubedb/kubedb-manifest-plugin/commit/9fc3cbe) Postgres manifest (#2)
- [8e2a56f](https://github.com/kubedb/kubedb-manifest-plugin/commit/8e2a56f) Merge pull request #1 from kubestash/mongodb-manifest
- [e80c1d0](https://github.com/kubedb/kubedb-manifest-plugin/commit/e80c1d0) update flag names.
- [80d3908](https://github.com/kubedb/kubedb-manifest-plugin/commit/80d3908) Add options for changing name in the restored files.
- [e7da42d](https://github.com/kubedb/kubedb-manifest-plugin/commit/e7da42d) Fix error.
- [70a0267](https://github.com/kubedb/kubedb-manifest-plugin/commit/70a0267) Sync with updated snapshot api
- [9d747d8](https://github.com/kubedb/kubedb-manifest-plugin/commit/9d747d8) Merge branch 'mongodb-manifest' of github.com:stashed/kubedb-manifest into mongodb-manifest
- [90e00e3](https://github.com/kubedb/kubedb-manifest-plugin/commit/90e00e3) Fix bugs.
- [9c3fc1e](https://github.com/kubedb/kubedb-manifest-plugin/commit/9c3fc1e) Sync with updated snapshot api
- [c321013](https://github.com/kubedb/kubedb-manifest-plugin/commit/c321013) update component path.
- [7f4bd17](https://github.com/kubedb/kubedb-manifest-plugin/commit/7f4bd17) Refactor.
- [2b61ff0](https://github.com/kubedb/kubedb-manifest-plugin/commit/2b61ff0) Specify component directory
- [6264cdf](https://github.com/kubedb/kubedb-manifest-plugin/commit/6264cdf) Support restoring particular mongo component.
- [0008570](https://github.com/kubedb/kubedb-manifest-plugin/commit/0008570) Fix restore component phase updating.
- [8bd4c95](https://github.com/kubedb/kubedb-manifest-plugin/commit/8bd4c95) Fix restore manifests.
- [7eda9f9](https://github.com/kubedb/kubedb-manifest-plugin/commit/7eda9f9) Update Snapshot phase calculation.
- [a2b52d2](https://github.com/kubedb/kubedb-manifest-plugin/commit/a2b52d2) Add core to runtime scheme.
- [9bd6bd5](https://github.com/kubedb/kubedb-manifest-plugin/commit/9bd6bd5) Fix bugs.
- [9e08774](https://github.com/kubedb/kubedb-manifest-plugin/commit/9e08774) Fix build
- [01225c6](https://github.com/kubedb/kubedb-manifest-plugin/commit/01225c6) Update module path
- [45d0e45](https://github.com/kubedb/kubedb-manifest-plugin/commit/45d0e45) updated flags.
- [fb0282f](https://github.com/kubedb/kubedb-manifest-plugin/commit/fb0282f) update docker file.
- [ad4c004](https://github.com/kubedb/kubedb-manifest-plugin/commit/ad4c004) refactor.
- [8f71d3a](https://github.com/kubedb/kubedb-manifest-plugin/commit/8f71d3a) Fix build
- [115ef23](https://github.com/kubedb/kubedb-manifest-plugin/commit/115ef23) update makefile.
- [a274690](https://github.com/kubedb/kubedb-manifest-plugin/commit/a274690) update backup and restore.
- [cff449f](https://github.com/kubedb/kubedb-manifest-plugin/commit/cff449f) Use yaml pkg from k8s.io.
- [dcbb399](https://github.com/kubedb/kubedb-manifest-plugin/commit/dcbb399) Use restic package from KubeStash.
- [596a498](https://github.com/kubedb/kubedb-manifest-plugin/commit/596a498) fix restore implementation.
- [6ebc19b](https://github.com/kubedb/kubedb-manifest-plugin/commit/6ebc19b) Implement restore.
- [3e8a869](https://github.com/kubedb/kubedb-manifest-plugin/commit/3e8a869) Start implementing restore.
- [e841113](https://github.com/kubedb/kubedb-manifest-plugin/commit/e841113) Add backup methods for mongodb.
- [b5961f7](https://github.com/kubedb/kubedb-manifest-plugin/commit/b5961f7) Continue implementing backup.
- [d943f6a](https://github.com/kubedb/kubedb-manifest-plugin/commit/d943f6a) Implement manifest backup for MongoDB.
- [e644c67](https://github.com/kubedb/kubedb-manifest-plugin/commit/e644c67) Implement kubedb-manifest plugin to MongoDB manifests.



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.22.0](https://github.com/kubedb/mariadb/releases/tag/v0.22.0)

- [b6995945](https://github.com/kubedb/mariadb/commit/b6995945) Prepare for release v0.22.0 (#237)
- [25018ad7](https://github.com/kubedb/mariadb/commit/25018ad7) Fix Statefulset Security Context Assign (#236)
- [9c157c66](https://github.com/kubedb/mariadb/commit/9c157c66) Prepare for release v0.22.0-rc.1 (#235)
- [1d0c2579](https://github.com/kubedb/mariadb/commit/1d0c2579) Pass version in SetDefaults func (#234)
- [e360fd82](https://github.com/kubedb/mariadb/commit/e360fd82) Prepare for release v0.22.0-rc.0 (#233)
- [3956f18c](https://github.com/kubedb/mariadb/commit/3956f18c) Set Default Security Context for MariaDB (#232)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.1.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.1.0)

- [a014ffc](https://github.com/kubedb/mariadb-archiver/commit/a014ffc) Prepare for release v0.1.0 (#5)
- [a2afbc9](https://github.com/kubedb/mariadb-archiver/commit/a2afbc9) Prepare for release v0.1.0-rc.1 (#4)
- [65fd6bf](https://github.com/kubedb/mariadb-archiver/commit/65fd6bf) Prepare for release v0.1.0-rc.0 (#3)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.18.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.18.0)

- [ec9782e7](https://github.com/kubedb/mariadb-coordinator/commit/ec9782e7) Prepare for release v0.18.0 (#94)
- [118bcda4](https://github.com/kubedb/mariadb-coordinator/commit/118bcda4) Prepare for release v0.18.0-rc.1 (#93)
- [bf515bfa](https://github.com/kubedb/mariadb-coordinator/commit/bf515bfa) Prepare for release v0.18.0-rc.0 (#92)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.31.0](https://github.com/kubedb/memcached/releases/tag/v0.31.0)

- [da52b0a5](https://github.com/kubedb/memcached/commit/da52b0a5) Prepare for release v0.31.0 (#407)
- [fab2a879](https://github.com/kubedb/memcached/commit/fab2a879) Prepare for release v0.31.0-rc.1 (#406)
- [e44be0a6](https://github.com/kubedb/memcached/commit/e44be0a6) Prepare for release v0.31.0-rc.0 (#405)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.31.0](https://github.com/kubedb/mongodb/releases/tag/v0.31.0)

- [32ab5a6a](https://github.com/kubedb/mongodb/commit/32ab5a6a) Prepare for release v0.31.0 (#584)
- [1a79be25](https://github.com/kubedb/mongodb/commit/1a79be25) Use args instead of cmd to work with latest walg image (#583)
- [de48eeb7](https://github.com/kubedb/mongodb/commit/de48eeb7) Prepare for release v0.31.0-rc.1 (#582)
- [c368ec94](https://github.com/kubedb/mongodb/commit/c368ec94) Prepare for release v0.31.0-rc.0 (#581)
- [020d5599](https://github.com/kubedb/mongodb/commit/020d5599) Set manifest component in restoreSession (#579)
- [95103a47](https://github.com/kubedb/mongodb/commit/95103a47) Implement mongodb archiver (#534)
- [fb01b593](https://github.com/kubedb/mongodb/commit/fb01b593) Update apimachinery deps for fsgroup defaulting (#578)
- [22a5bb29](https://github.com/kubedb/mongodb/commit/22a5bb29) Make changes to run containers as non-root user (#576)
- [8667f411](https://github.com/kubedb/mongodb/commit/8667f411) Rearrange the daily CI (#577)
- [7024a3ca](https://github.com/kubedb/mongodb/commit/7024a3ca) Add support for initialization with git-sync (#575)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.1.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.1.0)




## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.1.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.1.0)

- [93b29cd](https://github.com/kubedb/mongodb-restic-plugin/commit/93b29cd) Prepare for release v0.1.0 (#15)
- [1daa490](https://github.com/kubedb/mongodb-restic-plugin/commit/1daa490) Prepare for release v0.1.0-rc.1 (#14)
- [745f5cb](https://github.com/kubedb/mongodb-restic-plugin/commit/745f5cb) Prepare for release v0.1.0-rc.0 (#13)
- [2c381ee](https://github.com/kubedb/mongodb-restic-plugin/commit/2c381ee) Rename `max-Concurrency` flag name to `max-concurrency` (#12)
- [769bb27](https://github.com/kubedb/mongodb-restic-plugin/commit/769bb27) Set DB version from env if empty (#11)
- [7f51333](https://github.com/kubedb/mongodb-restic-plugin/commit/7f51333) Update snapshot time (#10)
- [e5972d1](https://github.com/kubedb/mongodb-restic-plugin/commit/e5972d1) Move to kubedb org
- [004ef7e](https://github.com/kubedb/mongodb-restic-plugin/commit/004ef7e) Update deps (#9)
- [e54bc9b](https://github.com/kubedb/mongodb-restic-plugin/commit/e54bc9b) Remove version prefix from files (#8)
- [2ab94f7](https://github.com/kubedb/mongodb-restic-plugin/commit/2ab94f7) Add db version flag (#6)
- [d3e752d](https://github.com/kubedb/mongodb-restic-plugin/commit/d3e752d) Prepare for release v0.1.0-rc.0 (#7)
- [e0872f9](https://github.com/kubedb/mongodb-restic-plugin/commit/e0872f9) Use firecracker runners
- [a2e18e9](https://github.com/kubedb/mongodb-restic-plugin/commit/a2e18e9) Use github runner to push docker image
- [b32ebb2](https://github.com/kubedb/mongodb-restic-plugin/commit/b32ebb2) Build docker images for each db version (#5)
- [bc3219d](https://github.com/kubedb/mongodb-restic-plugin/commit/bc3219d) Update deps
- [8040cc0](https://github.com/kubedb/mongodb-restic-plugin/commit/8040cc0) MongoDB backup and restore addon (#2)
- [d9cd315](https://github.com/kubedb/mongodb-restic-plugin/commit/d9cd315) Update Readme and license (#1)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.31.0](https://github.com/kubedb/mysql/releases/tag/v0.31.0)

- [9094f699](https://github.com/kubedb/mysql/commit/9094f699) Prepare for release v0.31.0 (#576)
- [fd0ebe09](https://github.com/kubedb/mysql/commit/fd0ebe09) Fix Statefulset Security Context Assign (#575)
- [79cb58c1](https://github.com/kubedb/mysql/commit/79cb58c1) Prepare for release v0.31.0-rc.1 (#574)
- [e5b37c00](https://github.com/kubedb/mysql/commit/e5b37c00) Pass version in SetDefaults func (#573)
- [3c005b51](https://github.com/kubedb/mysql/commit/3c005b51) Prepare for release v0.31.0-rc.0 (#572)
- [bcdfaf4a](https://github.com/kubedb/mysql/commit/bcdfaf4a) Set Default Security Context for MySQL (#571)
- [9009bcac](https://github.com/kubedb/mysql/commit/9009bcac) Add git sync constants from apimachinery (#570)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.1.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.1.0)

- [721eaa8](https://github.com/kubedb/mysql-archiver/commit/721eaa8) Prepare for release v0.1.0 (#4)
- [8c65d14](https://github.com/kubedb/mysql-archiver/commit/8c65d14) Prepare for release v0.1.0-rc.1 (#3)
- [f79286a](https://github.com/kubedb/mysql-archiver/commit/f79286a) Prepare for release v0.1.0-rc.0 (#2)
- [dcd2e30](https://github.com/kubedb/mysql-archiver/commit/dcd2e30) Fix wal-g binary



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.16.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.16.0)

- [c93152ea](https://github.com/kubedb/mysql-coordinator/commit/c93152ea) Prepare for release v0.16.0 (#91)
- [63cb0a33](https://github.com/kubedb/mysql-coordinator/commit/63cb0a33) Prepare for release v0.16.0-rc.1 (#90)
- [b5e481fc](https://github.com/kubedb/mysql-coordinator/commit/b5e481fc) Prepare for release v0.16.0-rc.0 (#89)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.1.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.1.0)

- [9ed9b45](https://github.com/kubedb/mysql-restic-plugin/commit/9ed9b45) Prepare for release v0.1.0 (#14)
- [f77476b](https://github.com/kubedb/mysql-restic-plugin/commit/f77476b) Prepare for release v0.1.0-rc.1 (#13)
- [81ceb55](https://github.com/kubedb/mysql-restic-plugin/commit/81ceb55) Add `databases` flag (#12)
- [b255e47](https://github.com/kubedb/mysql-restic-plugin/commit/b255e47) Prepare for release v0.1.0-rc.0 (#11)
- [9a17360](https://github.com/kubedb/mysql-restic-plugin/commit/9a17360) Set DB version from env if empty (#10)
- [c67ba7c](https://github.com/kubedb/mysql-restic-plugin/commit/c67ba7c) Update snapshot time (#9)
- [abef89e](https://github.com/kubedb/mysql-restic-plugin/commit/abef89e) Fix binary name
- [db1bbbf](https://github.com/kubedb/mysql-restic-plugin/commit/db1bbbf) Move to kubedb org
- [746d13e](https://github.com/kubedb/mysql-restic-plugin/commit/746d13e) Update deps (#8)
- [569533a](https://github.com/kubedb/mysql-restic-plugin/commit/569533a) Add version flag + Refactor (#6)
- [f0abd94](https://github.com/kubedb/mysql-restic-plugin/commit/f0abd94) Prepare for release v0.1.0-rc.0 (#7)
- [01bff62](https://github.com/kubedb/mysql-restic-plugin/commit/01bff62) Remove arm64 image support
- [277fda8](https://github.com/kubedb/mysql-restic-plugin/commit/277fda8) Build docker images for each db version (#5)
- [94f000d](https://github.com/kubedb/mysql-restic-plugin/commit/94f000d) Use Go 1.21
- [2e4f30d](https://github.com/kubedb/mysql-restic-plugin/commit/2e4f30d) Update Readme (#4)
- [272c8f9](https://github.com/kubedb/mysql-restic-plugin/commit/272c8f9) Add support for mysql backup and restore (#1)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.16.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.16.0)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.25.0](https://github.com/kubedb/ops-manager/releases/tag/v0.25.0)

- [63d69118](https://github.com/kubedb/ops-manager/commit/63d69118) Prepare for release v0.25.0 (#497)
- [96f06e76](https://github.com/kubedb/ops-manager/commit/96f06e76) Update deps
- [387bd0b0](https://github.com/kubedb/ops-manager/commit/387bd0b0) Modify update version logic for mongo to run chown (#496)
- [3a1ee06c](https://github.com/kubedb/ops-manager/commit/3a1ee06c) Add support for arbiter vertical scaling & volume expansion (#495)
- [98dbd6c0](https://github.com/kubedb/ops-manager/commit/98dbd6c0) Prepare for release v0.25.0-rc.1 (#494)
- [640fe280](https://github.com/kubedb/ops-manager/commit/640fe280) Prepare for release v0.25.0-rc.0 (#492)
- [9714e841](https://github.com/kubedb/ops-manager/commit/9714e841) Add kafka version 3.6.0 to daily test (#491)
- [dd18b17c](https://github.com/kubedb/ops-manager/commit/dd18b17c) postgres arbiter related changes and bug fixes (#483)
- [de52bda7](https://github.com/kubedb/ops-manager/commit/de52bda7) Remove default configuration and restart kafka with new config (#490)
- [f7850172](https://github.com/kubedb/ops-manager/commit/f7850172) Add prepare cluster installer before test runners (#489)
- [79e646ef](https://github.com/kubedb/ops-manager/commit/79e646ef) Update ServiceDNS for kafka (#488)
- [18851802](https://github.com/kubedb/ops-manager/commit/18851802) added daily postgres (#487)
- [0c2bdda1](https://github.com/kubedb/ops-manager/commit/0c2bdda1) added daily-postgres.yml (#486)
- [5cb75965](https://github.com/kubedb/ops-manager/commit/5cb75965) Fixed BUG in postgres reconfigureTLS opsreq (#485)
- [145e08d5](https://github.com/kubedb/ops-manager/commit/145e08d5) Failover before restarting primary on restart ops (#481)
- [e53a72ce](https://github.com/kubedb/ops-manager/commit/e53a72ce) Add Kafka daily yml (#475)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.25.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.25.0)

- [2780eea8](https://github.com/kubedb/percona-xtradb/commit/2780eea8) Prepare for release v0.25.0 (#335)
- [3a15a15e](https://github.com/kubedb/percona-xtradb/commit/3a15a15e) Fix Statefulset Security Context Assign (#334)
- [bad0b334](https://github.com/kubedb/percona-xtradb/commit/bad0b334) Prepare for release v0.25.0-rc.1 (#333)
- [b8447936](https://github.com/kubedb/percona-xtradb/commit/b8447936) Pass version in SetDefaults func (#332)
- [d374a542](https://github.com/kubedb/percona-xtradb/commit/d374a542) Prepare for release v0.25.0-rc.0 (#331)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.11.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.11.0)

- [44e0fea](https://github.com/kubedb/percona-xtradb-coordinator/commit/44e0fea) Prepare for release v0.11.0 (#51)
- [7a66da3](https://github.com/kubedb/percona-xtradb-coordinator/commit/7a66da3) Prepare for release v0.11.0-rc.1 (#50)
- [69e7d1e](https://github.com/kubedb/percona-xtradb-coordinator/commit/69e7d1e) Prepare for release v0.11.0-rc.0 (#49)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.22.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.22.0)

- [6eec739b](https://github.com/kubedb/pg-coordinator/commit/6eec739b) Prepare for release v0.22.0 (#141)
- [9f9614e6](https://github.com/kubedb/pg-coordinator/commit/9f9614e6) Prepare for release v0.22.0-rc.1 (#140)
- [e4efa4db](https://github.com/kubedb/pg-coordinator/commit/e4efa4db) Prepare for release v0.22.0-rc.0 (#139)
- [7c862bcd](https://github.com/kubedb/pg-coordinator/commit/7c862bcd) Add support for arbiter (#136)
- [53ba32a9](https://github.com/kubedb/pg-coordinator/commit/53ba32a9) added postgres 16.0 support (#137)
- [24445f9b](https://github.com/kubedb/pg-coordinator/commit/24445f9b) Added & modified logs (#134)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.25.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.25.0)

- [6e148ed2](https://github.com/kubedb/pgbouncer/commit/6e148ed2) Prepare for release v0.25.0 (#299)
- [efa76519](https://github.com/kubedb/pgbouncer/commit/efa76519) Fix Statefulset Security Context Assign (#298)
- [e3e9f84d](https://github.com/kubedb/pgbouncer/commit/e3e9f84d) Prepare for release v0.25.0-rc.1 (#297)
- [21ba9f0f](https://github.com/kubedb/pgbouncer/commit/21ba9f0f) Prepare for release v0.25.0-rc.0 (#296)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.38.0](https://github.com/kubedb/postgres/releases/tag/v0.38.0)

- [8fe3fe0e](https://github.com/kubedb/postgres/commit/8fe3fe0e1) Prepare for release v0.38.0 (#687)
- [7a853e00](https://github.com/kubedb/postgres/commit/7a853e001) Add postgres arbiter custom size limit (#686)
- [0e493a1c](https://github.com/kubedb/postgres/commit/0e493a1cd) Prepare for release v0.38.0-rc.1 (#685)
- [8738ad73](https://github.com/kubedb/postgres/commit/8738ad73e) Prepare for release v0.38.0-rc.0 (#684)
- [adb69b02](https://github.com/kubedb/postgres/commit/adb69b02e) Implement PostgreSQL archiver (#628)
- [668e15dd](https://github.com/kubedb/postgres/commit/668e15dd4) Remove test directory (#683)
- [d857c354](https://github.com/kubedb/postgres/commit/d857c354a) added postgres arbiter support (#677)
- [8fc98e8e](https://github.com/kubedb/postgres/commit/8fc98e8ed) Fixed a bug for init container (#681)
- [a2b408ff](https://github.com/kubedb/postgres/commit/a2b408ffb) Bugfix for security context (#680)
- [fb14015e](https://github.com/kubedb/postgres/commit/fb14015e9) added nightly yml for postgres (#679)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.1.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.1.0)




## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.1.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.1.0)

- [31f5fc5](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/31f5fc5) Prepare for release v0.1.0 (#8)
- [57a7bdf](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/57a7bdf) Prepare for release v0.1.0-rc.1 (#7)
- [02a45da](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/02a45da) Prepare for release v0.1.0-rc.0 (#6)
- [1a6457c](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/1a6457c) Update flags and deps + Refactor (#5)
- [f32b56b](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/f32b56b) Delete .idea folder
- [e7f8135](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/e7f8135) clean up (#4)
- [06e7e70](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/06e7e70) clean up (#3)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.1.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.1.0)

- [d5524a7](https://github.com/kubedb/postgres-restic-plugin/commit/d5524a7) Prepare for release v0.1.0 (#7)
- [584bbad](https://github.com/kubedb/postgres-restic-plugin/commit/584bbad) Prepare for release v0.1.0-rc.1 (#6)
- [da1ecd7](https://github.com/kubedb/postgres-restic-plugin/commit/da1ecd7) Refactor (#5)
- [8208814](https://github.com/kubedb/postgres-restic-plugin/commit/8208814) Prepare for release v0.1.0-rc.0 (#4)
- [a56fcfa](https://github.com/kubedb/postgres-restic-plugin/commit/a56fcfa) Move to kubedb org (#3)
- [e8928c7](https://github.com/kubedb/postgres-restic-plugin/commit/e8928c7) Added postgres addon for kubestash (#2)
- [7c55105](https://github.com/kubedb/postgres-restic-plugin/commit/7c55105) Prepare for release v0.1.0-rc.0 (#1)
- [19eff67](https://github.com/kubedb/postgres-restic-plugin/commit/19eff67) Use gh runner token to publish docker image
- [6a71410](https://github.com/kubedb/postgres-restic-plugin/commit/6a71410) Use firecracker runner
- [e278d71](https://github.com/kubedb/postgres-restic-plugin/commit/e278d71) Use Go 1.21
- [4899879](https://github.com/kubedb/postgres-restic-plugin/commit/4899879) Update readme + cleanup



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.38.0](https://github.com/kubedb/provisioner/releases/tag/v0.38.0)

- [284eef89](https://github.com/kubedb/provisioner/commit/284eef89b) Prepare for release v0.38.0 (#63)
- [4396cf5d](https://github.com/kubedb/provisioner/commit/4396cf5d5) Add storage, archiver, kubestash scheme (#62)
- [086300d9](https://github.com/kubedb/provisioner/commit/086300d90) Prepare for release v0.38.0-rc.1 (#61)
- [0dfe3742](https://github.com/kubedb/provisioner/commit/0dfe37425) Ensure archiver CRDs (#60)
- [7e6099e0](https://github.com/kubedb/provisioner/commit/7e6099e0e) Prepare for release v0.38.0-rc.0 (#59)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.25.0](https://github.com/kubedb/proxysql/releases/tag/v0.25.0)

- [69d82b19](https://github.com/kubedb/proxysql/commit/69d82b19) Prepare for release v0.25.0 (#315)
- [1f87cbc5](https://github.com/kubedb/proxysql/commit/1f87cbc5) Prepare for release v0.25.0-rc.1 (#314)
- [c4775bf7](https://github.com/kubedb/proxysql/commit/c4775bf7) Prepare for release v0.25.0-rc.0 (#313)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.31.0](https://github.com/kubedb/redis/releases/tag/v0.31.0)

- [de7b9f50](https://github.com/kubedb/redis/commit/de7b9f50) Prepare for release v0.31.0 (#500)
- [ffe5982e](https://github.com/kubedb/redis/commit/ffe5982e) Fix DB update from version for RedisSentinel (#499)
- [9f4f26ac](https://github.com/kubedb/redis/commit/9f4f26ac) Fix Statefulset Security Context Assign (#498)
- [a3d4b7b8](https://github.com/kubedb/redis/commit/a3d4b7b8) Prepare for release v0.31.0-rc.1 (#497)
- [bb101b6a](https://github.com/kubedb/redis/commit/bb101b6a) Pass version in SetDefaults func (#496)
- [966f14ca](https://github.com/kubedb/redis/commit/966f14ca) Prepare for release v0.31.0-rc.0 (#495)
- [b72d8319](https://github.com/kubedb/redis/commit/b72d8319) Run Redis and RedisSentinel as non root (#494)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.17.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.17.0)

- [34f65113](https://github.com/kubedb/redis-coordinator/commit/34f65113) Prepare for release v0.17.0 (#82)
- [7e5fbf31](https://github.com/kubedb/redis-coordinator/commit/7e5fbf31) Prepare for release v0.17.0-rc.1 (#81)
- [9f724e43](https://github.com/kubedb/redis-coordinator/commit/9f724e43) Prepare for release v0.17.0-rc.0 (#80)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.1.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.1.0)

- [79d23fd](https://github.com/kubedb/redis-restic-plugin/commit/79d23fd) Prepare for release v0.1.0 (#11)
- [8cae5ef](https://github.com/kubedb/redis-restic-plugin/commit/8cae5ef) Prepare for release v0.1.0-rc.1 (#10)
- [f8de18b](https://github.com/kubedb/redis-restic-plugin/commit/f8de18b) Prepare for release v0.1.0-rc.0 (#9)
- [a4c03d9](https://github.com/kubedb/redis-restic-plugin/commit/a4c03d9) Update snapshot time (#8)
- [404447d](https://github.com/kubedb/redis-restic-plugin/commit/404447d) Fix binary name
- [4dbc58b](https://github.com/kubedb/redis-restic-plugin/commit/4dbc58b) Move to kubedb org
- [e4a6fb2](https://github.com/kubedb/redis-restic-plugin/commit/e4a6fb2) Update deps (#7)
- [1b28954](https://github.com/kubedb/redis-restic-plugin/commit/1b28954) Remove maxConcurrency variable (#6)
- [4d13ee5](https://github.com/kubedb/redis-restic-plugin/commit/4d13ee5) Remove addon implementer + Refactor (#5)
- [44ac2c7](https://github.com/kubedb/redis-restic-plugin/commit/44ac2c7) Prepare for release v0.1.0-rc.0 (#4)
- [ce275bd](https://github.com/kubedb/redis-restic-plugin/commit/ce275bd) Use firecracker runner
- [bf39971](https://github.com/kubedb/redis-restic-plugin/commit/bf39971) Update deps
- [ef24891](https://github.com/kubedb/redis-restic-plugin/commit/ef24891) Use github runner to push docker image
- [6a6f6d6](https://github.com/kubedb/redis-restic-plugin/commit/6a6f6d6) Add support for redis backup and restore (#1)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.25.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.25.0)

- [5189e007](https://github.com/kubedb/replication-mode-detector/commit/5189e007) Prepare for release v0.25.0 (#246)
- [758906fe](https://github.com/kubedb/replication-mode-detector/commit/758906fe) Prepare for release v0.25.0-rc.1 (#245)
- [77886a28](https://github.com/kubedb/replication-mode-detector/commit/77886a28) Prepare for release v0.25.0-rc.0 (#244)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.14.0](https://github.com/kubedb/schema-manager/releases/tag/v0.14.0)

- [3707ea4f](https://github.com/kubedb/schema-manager/commit/3707ea4f) Prepare for release v0.14.0 (#87)
- [f7e384b2](https://github.com/kubedb/schema-manager/commit/f7e384b2) Prepare for release v0.14.0-rc.1 (#86)
- [893fe8d9](https://github.com/kubedb/schema-manager/commit/893fe8d9) Prepare for release v0.14.0-rc.0 (#85)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.23.0](https://github.com/kubedb/tests/releases/tag/v0.23.0)

- [3c1ea68e](https://github.com/kubedb/tests/commit/3c1ea68e) Prepare for release v0.23.0 (#274)
- [ea46166e](https://github.com/kubedb/tests/commit/ea46166e) Fix postgres tls test cases (#273)
- [f76a34b2](https://github.com/kubedb/tests/commit/f76a34b2) Arbiter related test cases added (#268)
- [8f82cf9a](https://github.com/kubedb/tests/commit/8f82cf9a) Prepare for release v0.23.0-rc.1 (#272)
- [0bfcc3b6](https://github.com/kubedb/tests/commit/0bfcc3b6) Fix kafka restart-pods test (#271)
- [bfd1ec79](https://github.com/kubedb/tests/commit/bfd1ec79) Prepare for release v0.23.0-rc.0 (#270)
- [fab75dd1](https://github.com/kubedb/tests/commit/fab75dd1) Add disableDefault while deploying elasticsearch. (#269)
- [009399c7](https://github.com/kubedb/tests/commit/009399c7) Run tests in restriced PodSecurityStandard (#266)
- [4be89382](https://github.com/kubedb/tests/commit/4be89382) Fixed stash test and Innodb issues in MySQL (#250)
- [f007f5f5](https://github.com/kubedb/tests/commit/f007f5f5) Added test for Standalone to HA scalin (#267)
- [017546ec](https://github.com/kubedb/tests/commit/017546ec) Add Postgres e2e tests (#233)
- [fbd16c88](https://github.com/kubedb/tests/commit/fbd16c88) Add kafka e2e tests (#254)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.14.0](https://github.com/kubedb/ui-server/releases/tag/v0.14.0)

- [bfe213eb](https://github.com/kubedb/ui-server/commit/bfe213eb) Prepare for release v0.14.0 (#96)
- [82f78763](https://github.com/kubedb/ui-server/commit/82f78763) Prepare for release v0.14.0-rc.1 (#95)
- [b59415fd](https://github.com/kubedb/ui-server/commit/b59415fd) Prepare for release v0.14.0-rc.0 (#94)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.14.0](https://github.com/kubedb/webhook-server/releases/tag/v0.14.0)

- [9ff121c7](https://github.com/kubedb/webhook-server/commit/9ff121c7) Prepare for release v0.14.0 (#73)
- [01d13baa](https://github.com/kubedb/webhook-server/commit/01d13baa) Prepare for release v0.14.0-rc.1 (#72)
- [e869d0ce](https://github.com/kubedb/webhook-server/commit/e869d0ce) Initialize default KubeBuilder client (#71)
- [c36d61e5](https://github.com/kubedb/webhook-server/commit/c36d61e5) Prepare for release v0.14.0-rc.0 (#70)




