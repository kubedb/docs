---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2021.04.16
    name: Changelog-v2021.04.16
    parent: welcome
    weight: 20210416
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2021.04.16/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2021.04.16/
---

# KubeDB v2021.04.16 (2021-04-16)


## [appscode/kubedb-autoscaler](https://github.com/appscode/kubedb-autoscaler)

### [v0.3.0](https://github.com/appscode/kubedb-autoscaler/releases/tag/v0.3.0)

- [335c87a](https://github.com/appscode/kubedb-autoscaler/commit/335c87a) Prepare for release v0.3.0 (#20)
- [e90c615](https://github.com/appscode/kubedb-autoscaler/commit/e90c615) Use license-verifier v0.8.1
- [432ea8d](https://github.com/appscode/kubedb-autoscaler/commit/432ea8d) Use license verifier v0.8.0
- [e6293b0](https://github.com/appscode/kubedb-autoscaler/commit/e6293b0) Update license verifier
- [573e940](https://github.com/appscode/kubedb-autoscaler/commit/573e940) Fix spelling



## [appscode/kubedb-enterprise](https://github.com/appscode/kubedb-enterprise)

### [v0.5.0](https://github.com/appscode/kubedb-enterprise/releases/tag/v0.5.0)

- [a6eccd35](https://github.com/appscode/kubedb-enterprise/commit/a6eccd35) Prepare for release v0.5.0 (#175)
- [a4af7e22](https://github.com/appscode/kubedb-enterprise/commit/a4af7e22) Fix wait for backup logic (#172)
- [5ef4fd8c](https://github.com/appscode/kubedb-enterprise/commit/5ef4fd8c) Fix nil pointer exception while updating MongoDB configSecret (#173)
- [b9ee5297](https://github.com/appscode/kubedb-enterprise/commit/b9ee5297) Pause `BackupConfiguration` and Wait for `BackupSession` & `RestoreSession` to complete (#168)
- [7064e346](https://github.com/appscode/kubedb-enterprise/commit/7064e346) Fix various issues for MongoDBOpsRequest (#169)
- [adf174e7](https://github.com/appscode/kubedb-enterprise/commit/adf174e7) Add Ops Request Phase `Pending` (#166)
- [355d1b1e](https://github.com/appscode/kubedb-enterprise/commit/355d1b1e) Fix panic for MongoDB (#167)
- [df672de0](https://github.com/appscode/kubedb-enterprise/commit/df672de0) Add HostNetwork and DNSPolicy to new StatefulSet (#171)
- [7a279d7b](https://github.com/appscode/kubedb-enterprise/commit/7a279d7b) Add Elsticsearch statefulSet reconciler (#161)
- [0fbd67b6](https://github.com/appscode/kubedb-enterprise/commit/0fbd67b6) Updated MustCertSecretName to GetCertSecretName (#162)
- [7e6a0d78](https://github.com/appscode/kubedb-enterprise/commit/7e6a0d78) Remove panic from Postgres (#170)
- [ae7f27bb](https://github.com/appscode/kubedb-enterprise/commit/ae7f27bb) Use license-verifier v0.8.1
- [e3ff9160](https://github.com/appscode/kubedb-enterprise/commit/e3ff9160) Elasticsearch: Return default certificate secret name if missing (#165)
- [824a2d80](https://github.com/appscode/kubedb-enterprise/commit/824a2d80) Use license verifier v0.8.0
- [40ec97e9](https://github.com/appscode/kubedb-enterprise/commit/40ec97e9) Update license verifier
- [c2757fb3](https://github.com/appscode/kubedb-enterprise/commit/c2757fb3) Fix spelling
- [4c41bc1e](https://github.com/appscode/kubedb-enterprise/commit/4c41bc1e) Don't activate namespace validator



## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.18.0](https://github.com/kubedb/apimachinery/releases/tag/v0.18.0)

- [fdf2681b](https://github.com/kubedb/apimachinery/commit/fdf2681b) Remove some panics from MongoDB (#734)
- [4bc52e52](https://github.com/kubedb/apimachinery/commit/4bc52e52) Add Ops Request Phase `Pending` (#736)
- [62e6b324](https://github.com/kubedb/apimachinery/commit/62e6b324) Add timeout for MongoDBOpsrequest (#738)
- [15a029cc](https://github.com/kubedb/apimachinery/commit/15a029cc) Check for all Stash CRDs before starting Start controller (#740)
- [b9da5117](https://github.com/kubedb/apimachinery/commit/b9da5117) Add backup condition constants for ops requests (#737)
- [989d8200](https://github.com/kubedb/apimachinery/commit/989d8200) Remove Panic from MariaDB (#731)
- [b7b2a28c](https://github.com/kubedb/apimachinery/commit/b7b2a28c) Remove panic from Postgres (#739)
- [264c0872](https://github.com/kubedb/apimachinery/commit/264c0872) Add IsIP helper
- [feebf1d8](https://github.com/kubedb/apimachinery/commit/feebf1d8) Add pod identity for cluster configuration (#729)
- [1b58e82b](https://github.com/kubedb/apimachinery/commit/1b58e82b) Add SecurityContext to ElasticsearchVersion CRD (#733)
- [b3c20afc](https://github.com/kubedb/apimachinery/commit/b3c20afc) Update for release Stash@v2021.04.07 (#735)
- [ced2341f](https://github.com/kubedb/apimachinery/commit/ced2341f) Return default cert-secret name if missing (#730)
- [ecf77001](https://github.com/kubedb/apimachinery/commit/ecf77001) Rename Features to SecurityContext in Postgres Version Spec (#732)
- [e5917a15](https://github.com/kubedb/apimachinery/commit/e5917a15) Rename RunAsAny to RunAsAnyNonRoot in PostgresVersion
- [5060058c](https://github.com/kubedb/apimachinery/commit/5060058c) Add Custom UID Options for Postgres (#728)
- [ed221fe1](https://github.com/kubedb/apimachinery/commit/ed221fe1) Fix spelling



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.18.0](https://github.com/kubedb/cli/releases/tag/v0.18.0)

- [a1a424ba](https://github.com/kubedb/cli/commit/a1a424ba) Prepare for release v0.18.0 (#598)
- [c8bec973](https://github.com/kubedb/cli/commit/c8bec973) Fix spelling



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.18.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.18.0)

- [2da042ce](https://github.com/kubedb/elasticsearch/commit/2da042ce) Prepare for release v0.18.0 (#490)
- [e398bd53](https://github.com/kubedb/elasticsearch/commit/e398bd53) Add statefulSet reconciler (#488)
- [2944f5e5](https://github.com/kubedb/elasticsearch/commit/2944f5e5) Use license-verifier v0.8.1
- [8d7177ab](https://github.com/kubedb/elasticsearch/commit/8d7177ab) Add support for custom UID for Elasticsearch (#489)
- [71d1fac3](https://github.com/kubedb/elasticsearch/commit/71d1fac3) Use license verifier v0.8.0
- [fb5ee170](https://github.com/kubedb/elasticsearch/commit/fb5ee170) Update license verifier
- [f166bf3a](https://github.com/kubedb/elasticsearch/commit/f166bf3a) Update stash make targets (#487)
- [39d5be0f](https://github.com/kubedb/elasticsearch/commit/39d5be0f) Fix spelling



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2021.04.16](https://github.com/kubedb/installer/releases/tag/v2021.04.16)

- [a5f3c63](https://github.com/kubedb/installer/commit/a5f3c63) Prepare for release v2021.04.16 (#299)
- [6aa0d8b](https://github.com/kubedb/installer/commit/6aa0d8b) Update MongoDB init container image (#298)
- [7407a80](https://github.com/kubedb/installer/commit/7407a80) Add `poddisruptionbudgets` and backup permissions for KubeDB Enterprise (#297)
- [c12845b](https://github.com/kubedb/installer/commit/c12845b) Add mariadb-init-docker Image (#296)
- [39c7a94](https://github.com/kubedb/installer/commit/39c7a94) Add MySQL init container images to catalog (#291)
- [2272f27](https://github.com/kubedb/installer/commit/2272f27) Add support for Elasticsearch v7.12.0 (#295)
- [eebcbac](https://github.com/kubedb/installer/commit/eebcbac) Update installer schema
- [8619834](https://github.com/kubedb/installer/commit/8619834) Allow passing registry fqdn (#294)
- [4435651](https://github.com/kubedb/installer/commit/4435651) Custom UID for Postgres (#293)
- [1658fb2](https://github.com/kubedb/installer/commit/1658fb2) Fix spelling



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.2.0](https://github.com/kubedb/mariadb/releases/tag/v0.2.0)

- [db6efa46](https://github.com/kubedb/mariadb/commit/db6efa46) Prepare for release v0.2.0 (#65)
- [01575e35](https://github.com/kubedb/mariadb/commit/01575e35) Updated validator for requireSSL field. (#61)
- [585f1873](https://github.com/kubedb/mariadb/commit/585f1873) Introduced MariaDB init-container (#62)
- [821c3688](https://github.com/kubedb/mariadb/commit/821c3688) Updated MustCertSecretName to GetCertSecretName (#64)
- [5d41c58a](https://github.com/kubedb/mariadb/commit/5d41c58a) Add POD_IP env variable (#63)
- [11e56c19](https://github.com/kubedb/mariadb/commit/11e56c19) Use license-verifier v0.8.1
- [f7d6c516](https://github.com/kubedb/mariadb/commit/f7d6c516) Use license verifier v0.8.0
- [3cfc4979](https://github.com/kubedb/mariadb/commit/3cfc4979) Update license verifier
- [60e8e7a3](https://github.com/kubedb/mariadb/commit/60e8e7a3) Update stash make targets (#60)
- [9424f4be](https://github.com/kubedb/mariadb/commit/9424f4be) Fix spelling



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.11.0](https://github.com/kubedb/memcached/releases/tag/v0.11.0)

- [4f391d75](https://github.com/kubedb/memcached/commit/4f391d75) Prepare for release v0.11.0 (#293)
- [294fb730](https://github.com/kubedb/memcached/commit/294fb730) Use license-verifier v0.8.1
- [717a5c06](https://github.com/kubedb/memcached/commit/717a5c06) Use license verifier v0.8.0
- [37f7bba6](https://github.com/kubedb/memcached/commit/37f7bba6) Update license verifier
- [4a6fea4d](https://github.com/kubedb/memcached/commit/4a6fea4d) Fix spelling



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.11.0](https://github.com/kubedb/mongodb/releases/tag/v0.11.0)

- [c4d7444c](https://github.com/kubedb/mongodb/commit/c4d7444c) Prepare for release v0.11.0 (#390)
- [26fbc21b](https://github.com/kubedb/mongodb/commit/26fbc21b) Use `IPv6EnabledInKernel` (#389)
- [0221339a](https://github.com/kubedb/mongodb/commit/0221339a) Selectively enable binding IPv6 address (#388)
- [7a53a0bc](https://github.com/kubedb/mongodb/commit/7a53a0bc) Remove panic (#387)
- [87623d58](https://github.com/kubedb/mongodb/commit/87623d58) Introduce NodeReconciler (#384)
- [891aac47](https://github.com/kubedb/mongodb/commit/891aac47) Use license-verifier v0.8.1
- [0722ad6d](https://github.com/kubedb/mongodb/commit/0722ad6d) Use license verifier v0.8.0
- [f8522304](https://github.com/kubedb/mongodb/commit/f8522304) Update license verifier
- [dab6babc](https://github.com/kubedb/mongodb/commit/dab6babc) Update stash make targets (#386)
- [b18cbac5](https://github.com/kubedb/mongodb/commit/b18cbac5) Fix spelling



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.11.0](https://github.com/kubedb/mysql/releases/tag/v0.11.0)

- [d7e33a0c](https://github.com/kubedb/mysql/commit/d7e33a0c) Prepare for release v0.11.0 (#382)
- [782b3613](https://github.com/kubedb/mysql/commit/782b3613) Add podIP to pod env (#381)
- [f0166dbe](https://github.com/kubedb/mysql/commit/f0166dbe) Always pass -address-type to peer-finder (#380)
- [df770ff2](https://github.com/kubedb/mysql/commit/df770ff2) Use license-verifier v0.8.1
- [d610fddc](https://github.com/kubedb/mysql/commit/d610fddc) Add support for using official mysql image (#377)
- [a99510f2](https://github.com/kubedb/mysql/commit/a99510f2) Use license verifier v0.8.0
- [100dd336](https://github.com/kubedb/mysql/commit/100dd336) Update license verifier
- [1bdbe4ed](https://github.com/kubedb/mysql/commit/1bdbe4ed) Update stash make targets (#379)
- [40f9a2f2](https://github.com/kubedb/mysql/commit/40f9a2f2) Fix spelling



## [kubedb/operator](https://github.com/kubedb/operator)

### [v0.18.0](https://github.com/kubedb/operator/releases/tag/v0.18.0)

- [5c2cb8b2](https://github.com/kubedb/operator/commit/5c2cb8b2) Prepare for release v0.18.0 (#402)
- [5ca8913b](https://github.com/kubedb/operator/commit/5ca8913b) Use license-verifier v0.8.1



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.5.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.5.0)

- [06e1c3fb](https://github.com/kubedb/percona-xtradb/commit/06e1c3fb) Prepare for release v0.5.0 (#194)
- [ecba4c64](https://github.com/kubedb/percona-xtradb/commit/ecba4c64) Use license-verifier v0.8.1
- [9d59d002](https://github.com/kubedb/percona-xtradb/commit/9d59d002) Use license verifier v0.8.0
- [6f924248](https://github.com/kubedb/percona-xtradb/commit/6f924248) Update license verifier
- [e1055e9b](https://github.com/kubedb/percona-xtradb/commit/e1055e9b) Update stash make targets (#193)
- [febdf8de](https://github.com/kubedb/percona-xtradb/commit/febdf8de) Fix spelling



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.2.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.2.0)

- [b5311e9](https://github.com/kubedb/pg-coordinator/commit/b5311e9) Prepare for release v0.2.0 (#16)
- [db687a1](https://github.com/kubedb/pg-coordinator/commit/db687a1) Add Support for Custom UID (#15)
- [1f923a4](https://github.com/kubedb/pg-coordinator/commit/1f923a4) Fix spelling



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.5.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.5.0)

- [1b32fbbb](https://github.com/kubedb/pgbouncer/commit/1b32fbbb) Prepare for release v0.5.0 (#154)
- [d5102b66](https://github.com/kubedb/pgbouncer/commit/d5102b66) Use license-verifier v0.8.1
- [30e3e2f9](https://github.com/kubedb/pgbouncer/commit/30e3e2f9) Use license verifier v0.8.0
- [3c2833db](https://github.com/kubedb/pgbouncer/commit/3c2833db) Update license verifier
- [06463c97](https://github.com/kubedb/pgbouncer/commit/06463c97) Fix spelling



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.18.0](https://github.com/kubedb/postgres/releases/tag/v0.18.0)

- [e0b70d83](https://github.com/kubedb/postgres/commit/e0b70d83) Prepare for release v0.18.0 (#489)
- [a3ffea16](https://github.com/kubedb/postgres/commit/a3ffea16) remove panic from postgres (#488)
- [09adb390](https://github.com/kubedb/postgres/commit/09adb390) Remove wait-group from postgres operator (#487)
- [ae8b87da](https://github.com/kubedb/postgres/commit/ae8b87da) Use license-verifier v0.8.1
- [77f220b8](https://github.com/kubedb/postgres/commit/77f220b8) Update KubeDB api (#486)
- [b0234c4b](https://github.com/kubedb/postgres/commit/b0234c4b) Add Custom-UID Support for Debian Images (#485)
- [fdf4d2df](https://github.com/kubedb/postgres/commit/fdf4d2df) Use license verifier v0.8.0
- [dd59f9b1](https://github.com/kubedb/postgres/commit/dd59f9b1) Update license verifier
- [43fd0c33](https://github.com/kubedb/postgres/commit/43fd0c33) Update stash make targets (#484)
- [8632e4c5](https://github.com/kubedb/postgres/commit/8632e4c5) Fix spelling



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.5.0](https://github.com/kubedb/proxysql/releases/tag/v0.5.0)

- [8cd69666](https://github.com/kubedb/proxysql/commit/8cd69666) Prepare for release v0.5.0 (#172)
- [7cc0781a](https://github.com/kubedb/proxysql/commit/7cc0781a) Use license-verifier v0.8.1
- [296e14f0](https://github.com/kubedb/proxysql/commit/296e14f0) Use license verifier v0.8.0
- [2fd9f4e5](https://github.com/kubedb/proxysql/commit/2fd9f4e5) Update license verifier
- [7fb0a67f](https://github.com/kubedb/proxysql/commit/7fb0a67f) Fix spelling



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.11.0](https://github.com/kubedb/redis/releases/tag/v0.11.0)

- [f08b8987](https://github.com/kubedb/redis/commit/f08b8987) Prepare for release v0.11.0 (#316)
- [02347918](https://github.com/kubedb/redis/commit/02347918) Use license-verifier v0.8.1
- [fc33c657](https://github.com/kubedb/redis/commit/fc33c657) Use license verifier v0.8.0
- [1cd12234](https://github.com/kubedb/redis/commit/1cd12234) Update license verifier
- [5ba20810](https://github.com/kubedb/redis/commit/5ba20810) Fix spelling



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.5.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.5.0)

- [ee224f1](https://github.com/kubedb/replication-mode-detector/commit/ee224f1) Prepare for release v0.5.0 (#135)
- [8293c27](https://github.com/kubedb/replication-mode-detector/commit/8293c27) Add comparing host with podIP or DNS for MySQL (#134)
- [f608626](https://github.com/kubedb/replication-mode-detector/commit/f608626) Fix mysql query for getting primary member "ONLINE" (#124)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.3.0](https://github.com/kubedb/tests/releases/tag/v0.3.0)

- [1d230e5](https://github.com/kubedb/tests/commit/1d230e5) Prepare for release v0.3.0 (#115)
- [a2148b0](https://github.com/kubedb/tests/commit/a2148b0) Rename `MustCertSecretName` to `GetCertSecretName` (#113)




