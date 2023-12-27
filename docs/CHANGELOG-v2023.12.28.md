---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2023.12.28
    name: Changelog-v2023.12.28
    parent: welcome
    weight: 20231228
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2023.12.28/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2023.12.28/
---

# KubeDB v2023.12.28 (2023-12-27)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.40.0](https://github.com/kubedb/apimachinery/releases/tag/v0.40.0)

- [000dfa1a](https://github.com/kubedb/apimachinery/commit/000dfa1a6) Use kubestash v0.3.0
- [541ddfd4](https://github.com/kubedb/apimachinery/commit/541ddfd45) Update client-go deps
- [b6912d25](https://github.com/kubedb/apimachinery/commit/b6912d25a) Defaulting compute autoscaler fields (#1097)
- [61b590f7](https://github.com/kubedb/apimachinery/commit/61b590f7b) Add mysql archiver apis (#1086)
- [750f6385](https://github.com/kubedb/apimachinery/commit/750f6385b) Add scaleUp & scaleDown diffPercentage fields to autoscaler (#1092)
- [0922ff18](https://github.com/kubedb/apimachinery/commit/0922ff18c) Add default resource for initContainer (#1094)
- [da96ad5f](https://github.com/kubedb/apimachinery/commit/da96ad5fe) Revert "Add kubestash controller for changing kubeDB phase (#1076)"
- [d6368a16](https://github.com/kubedb/apimachinery/commit/d6368a16f) Add kubestash controller for changing kubeDB phase (#1076)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.25.0](https://github.com/kubedb/autoscaler/releases/tag/v0.25.0)

- [557a5503](https://github.com/kubedb/autoscaler/commit/557a5503) Prepare for release v0.25.0 (#169)
- [346ed2f0](https://github.com/kubedb/autoscaler/commit/346ed2f0) Implement nodePool jumping when topology specified  (#168)
- [5d464e14](https://github.com/kubedb/autoscaler/commit/5d464e14) Add Dockerfile for dbg; Use go 21 (#167)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.40.0](https://github.com/kubedb/cli/releases/tag/v0.40.0)

- [94aedfc9](https://github.com/kubedb/cli/commit/94aedfc9) Prepare for release v0.40.0 (#742)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.16.0](https://github.com/kubedb/dashboard/releases/tag/v0.16.0)

- [67058d0b](https://github.com/kubedb/dashboard/commit/67058d0b) Prepare for release v0.16.0 (#90)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.40.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.40.0)

- [745c7022](https://github.com/kubedb/elasticsearch/commit/745c70225) Prepare for release v0.40.0 (#687)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.3.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.3.0)

- [231c402](https://github.com/kubedb/elasticsearch-restic-plugin/commit/231c402) Prepare for release v0.3.0 (#13)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2023.12.28](https://github.com/kubedb/installer/releases/tag/v2023.12.28)

- [8c5db03d](https://github.com/kubedb/installer/commit/8c5db03d) Prepare for release v2023.12.28 (#758)
- [2e323fab](https://github.com/kubedb/installer/commit/2e323fab) mongodb-csisnapshotter -> mongodb-csi-snapshotter
- [c8226e7d](https://github.com/kubedb/installer/commit/c8226e7d) postgres-csisnapshotter -> postgres-csi-snapshotter
- [e17deeb2](https://github.com/kubedb/installer/commit/e17deeb2) Add NodeTopology crd to autoscaler chart (#757)
- [84673762](https://github.com/kubedb/installer/commit/84673762) Add mysql archiver specs (#755)
- [76713c53](https://github.com/kubedb/installer/commit/76713c53) Templatize wal-g images (#756)
- [74e57f6e](https://github.com/kubedb/installer/commit/74e57f6e) Add image.* templates to kubestash catalog chart
- [66efd1f3](https://github.com/kubedb/installer/commit/66efd1f3) Update crds for kubedb/apimachinery@61b590f7 (#754)
- [dc051893](https://github.com/kubedb/installer/commit/dc051893) Update crds for kubedb/apimachinery@750f6385 (#753)



## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.11.0](https://github.com/kubedb/kafka/releases/tag/v0.11.0)

- [65e61f0](https://github.com/kubedb/kafka/commit/65e61f0) Prepare for release v0.11.0 (#56)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.3.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.3.0)

- [c664d92](https://github.com/kubedb/kubedb-manifest-plugin/commit/c664d92) Prepare for release v0.3.0 (#33)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.24.0](https://github.com/kubedb/mariadb/releases/tag/v0.24.0)

- [94f03b1b](https://github.com/kubedb/mariadb/commit/94f03b1b) Prepare for release v0.24.0 (#241)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.1.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.1.0)

- [910b7ce](https://github.com/kubedb/mariadb-archiver/commit/910b7ce) Prepare for release v0.1.0 (#1)
- [3801668](https://github.com/kubedb/mariadb-archiver/commit/3801668) mysql -> mariadb
- [4e905fb](https://github.com/kubedb/mariadb-archiver/commit/4e905fb) Implemenet new algorithm for archiver and restorer (#5)
- [22701c8](https://github.com/kubedb/mariadb-archiver/commit/22701c8) Fix 5.7.x build
- [6da2b1c](https://github.com/kubedb/mariadb-archiver/commit/6da2b1c) Update build matrix
- [e2f6244](https://github.com/kubedb/mariadb-archiver/commit/e2f6244) Use separate dockerfile per mysql version (#9)
- [e800623](https://github.com/kubedb/mariadb-archiver/commit/e800623) Prepare for release v0.2.0 (#8)
- [b9f6ec5](https://github.com/kubedb/mariadb-archiver/commit/b9f6ec5) Install mysqlbinlog (#7)
- [c46d991](https://github.com/kubedb/mariadb-archiver/commit/c46d991) Use appscode-images as base image (#6)
- [721eaa8](https://github.com/kubedb/mariadb-archiver/commit/721eaa8) Prepare for release v0.1.0 (#4)
- [8c65d14](https://github.com/kubedb/mariadb-archiver/commit/8c65d14) Prepare for release v0.1.0-rc.1 (#3)
- [f79286a](https://github.com/kubedb/mariadb-archiver/commit/f79286a) Prepare for release v0.1.0-rc.0 (#2)
- [dcd2e30](https://github.com/kubedb/mariadb-archiver/commit/dcd2e30) Fix wal-g binary
- [6c20a4a](https://github.com/kubedb/mariadb-archiver/commit/6c20a4a) Fix build
- [f034e7b](https://github.com/kubedb/mariadb-archiver/commit/f034e7b) Add build script (#1)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.20.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.20.0)

- [ff2b45fc](https://github.com/kubedb/mariadb-coordinator/commit/ff2b45fc) Prepare for release v0.20.0 (#96)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.1.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.1.0)

- [933e138](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/933e138) Prepare for release v0.1.0 (#2)
- [5d38f94](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/5d38f94) Enable GH actions
- [2a97178](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/2a97178) Replace mysql with mariadb



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.33.0](https://github.com/kubedb/memcached/releases/tag/v0.33.0)

- [bf3329f4](https://github.com/kubedb/memcached/commit/bf3329f4) Prepare for release v0.33.0 (#411)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.33.0](https://github.com/kubedb/mongodb/releases/tag/v0.33.0)

- [30a34a1c](https://github.com/kubedb/mongodb/commit/30a34a1c) Prepare for release v0.33.0 (#592)
- [71c092df](https://github.com/kubedb/mongodb/commit/71c092df) Trigger backupSession once while backupConfig created (#591)
- [57c8a367](https://github.com/kubedb/mongodb/commit/57c8a367) Set Default initContainer resource (#590)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.1.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.1.0)

- [fc233a7](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/fc233a7) Prepare for release v0.1.0 (#7)
- [2bad72d](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/2bad72d) Prepare for release v0.2.0 (#6)
- [c2fcb4f](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/c2fcb4f) Prepare for release v0.1.0 (#5)
- [92b28e8](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/92b28e8) Prepare for release v0.1.0-rc.1 (#4)
- [f06d344](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/f06d344) Prepare for release v0.1.0-rc.0 (#3)
- [df1a966](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/df1a966) Update flags + Refactor (#2)
- [7eb7cea](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/7eb7cea) Fix issues
- [1f189b4](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/1f189b4) Test against K8s 1.27.0 (#1)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.3.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.3.0)

- [efac8ef](https://github.com/kubedb/mongodb-restic-plugin/commit/efac8ef) Prepare for release v0.3.0 (#18)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.33.0](https://github.com/kubedb/mysql/releases/tag/v0.33.0)

- [74a63d01](https://github.com/kubedb/mysql/commit/74a63d01) Prepare for release v0.33.0 (#583)
- [272eb81c](https://github.com/kubedb/mysql/commit/272eb81c) fix typos and secondlast snapshot time (#582)
- [7efda78a](https://github.com/kubedb/mysql/commit/7efda78a) Fix error return
- [a6cd4fe9](https://github.com/kubedb/mysql/commit/a6cd4fe9) Add support for archiver (#577)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.1.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.1.0)

- [c956cb9](https://github.com/kubedb/mysql-archiver/commit/c956cb9) Prepare for release v0.1.0 (#10)
- [4e905fb](https://github.com/kubedb/mysql-archiver/commit/4e905fb) Implemenet new algorithm for archiver and restorer (#5)
- [22701c8](https://github.com/kubedb/mysql-archiver/commit/22701c8) Fix 5.7.x build
- [6da2b1c](https://github.com/kubedb/mysql-archiver/commit/6da2b1c) Update build matrix
- [e2f6244](https://github.com/kubedb/mysql-archiver/commit/e2f6244) Use separate dockerfile per mysql version (#9)
- [e800623](https://github.com/kubedb/mysql-archiver/commit/e800623) Prepare for release v0.2.0 (#8)
- [b9f6ec5](https://github.com/kubedb/mysql-archiver/commit/b9f6ec5) Install mysqlbinlog (#7)
- [c46d991](https://github.com/kubedb/mysql-archiver/commit/c46d991) Use appscode-images as base image (#6)
- [721eaa8](https://github.com/kubedb/mysql-archiver/commit/721eaa8) Prepare for release v0.1.0 (#4)
- [8c65d14](https://github.com/kubedb/mysql-archiver/commit/8c65d14) Prepare for release v0.1.0-rc.1 (#3)
- [f79286a](https://github.com/kubedb/mysql-archiver/commit/f79286a) Prepare for release v0.1.0-rc.0 (#2)
- [dcd2e30](https://github.com/kubedb/mysql-archiver/commit/dcd2e30) Fix wal-g binary
- [6c20a4a](https://github.com/kubedb/mysql-archiver/commit/6c20a4a) Fix build
- [f034e7b](https://github.com/kubedb/mysql-archiver/commit/f034e7b) Add build script (#1)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.18.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.18.0)

- [8d6d3073](https://github.com/kubedb/mysql-coordinator/commit/8d6d3073) Prepare for release v0.18.0 (#93)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.1.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.1.0)

- [34bd9fd](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/34bd9fd) Prepare for release v0.1.0 (#1)
- [a0ddb4a](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/a0ddb4a) Enable GH actions



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.3.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.3.0)

- [a364862](https://github.com/kubedb/mysql-restic-plugin/commit/a364862) Prepare for release v0.3.0 (#17)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.18.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.18.0)




## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.27.0](https://github.com/kubedb/ops-manager/releases/tag/v0.27.0)

- [abbbea47](https://github.com/kubedb/ops-manager/commit/abbbea47) Prepare for release v0.27.0 (#505)
- [5955ffd7](https://github.com/kubedb/ops-manager/commit/5955ffd7) Delete pod if it was found in CrashLoopBackOff while restarting (#504)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.27.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.27.0)

- [117ce794](https://github.com/kubedb/percona-xtradb/commit/117ce794) Prepare for release v0.27.0 (#340)
- [56a4b354](https://github.com/kubedb/percona-xtradb/commit/56a4b354) Fix initContainer resource (#339)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.13.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.13.0)

- [8ce147f](https://github.com/kubedb/percona-xtradb-coordinator/commit/8ce147f) Prepare for release v0.13.0 (#53)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.24.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.24.0)

- [e3f8df76](https://github.com/kubedb/pg-coordinator/commit/e3f8df76) Prepare for release v0.24.0 (#143)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.27.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.27.0)

- [0abff0c8](https://github.com/kubedb/pgbouncer/commit/0abff0c8) Prepare for release v0.27.0 (#304)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.40.0](https://github.com/kubedb/postgres/releases/tag/v0.40.0)

- [17d39368](https://github.com/kubedb/postgres/commit/17d393689) Prepare for release v0.40.0 (#695)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.1.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.1.0)

- [12cb5f0](https://github.com/kubedb/postgres-archiver/commit/12cb5f0) Prepare for release v0.1.0 (#14)
- [91c52a5](https://github.com/kubedb/postgres-archiver/commit/91c52a5) Add tls support for connection string (#13)
- [c4f7e11](https://github.com/kubedb/postgres-archiver/commit/c4f7e11) Fix formatting
- [1feeaeb](https://github.com/kubedb/postgres-archiver/commit/1feeaeb) Fix wal-g version
- [c86ede7](https://github.com/kubedb/postgres-archiver/commit/c86ede7) Update readme
- [f5b4fb3](https://github.com/kubedb/postgres-archiver/commit/f5b4fb3) Rename to postgres-archiver
- [302fbc1](https://github.com/kubedb/postgres-archiver/commit/302fbc1) Merge pull request #12 from kubedb/cleanup
- [020b817](https://github.com/kubedb/postgres-archiver/commit/020b817) clean up
- [5ae6dee](https://github.com/kubedb/postgres-archiver/commit/5ae6dee) Add ca-certificates into docker image
- [2a9e7b5](https://github.com/kubedb/postgres-archiver/commit/2a9e7b5) Build images parallelly
- [ec05751](https://github.com/kubedb/postgres-archiver/commit/ec05751) Build bookwork images
- [1ed24d1](https://github.com/kubedb/postgres-archiver/commit/1ed24d1) Build multi version docker images
- [57dd7e5](https://github.com/kubedb/postgres-archiver/commit/57dd7e5) Format repo (#11)
- [adc5e71](https://github.com/kubedb/postgres-archiver/commit/adc5e71) Implement archiver command (#7)
- [7d0adba](https://github.com/kubedb/postgres-archiver/commit/7d0adba) Test against K8s 1.27.0 (#10)
- [9b2a242](https://github.com/kubedb/postgres-archiver/commit/9b2a242) Update Makefile
- [cbbe124](https://github.com/kubedb/postgres-archiver/commit/cbbe124) Use ghcr.io for appscode/golang-dev (#9)
- [03877ad](https://github.com/kubedb/postgres-archiver/commit/03877ad) Update wrokflows (Go 1.20, k8s 1.26) (#8)
- [ad607ec](https://github.com/kubedb/postgres-archiver/commit/ad607ec) Use Go 1.18 (#5)
- [32d1866](https://github.com/kubedb/postgres-archiver/commit/32d1866) Use Go 1.18 (#4)
- [42ae1cb](https://github.com/kubedb/postgres-archiver/commit/42ae1cb) make fmt (#3)
- [1a6fe8d](https://github.com/kubedb/postgres-archiver/commit/1a6fe8d) Update repository config (#2)
- [8100920](https://github.com/kubedb/postgres-archiver/commit/8100920) Update repository config (#1)
- [8e3c29d](https://github.com/kubedb/postgres-archiver/commit/8e3c29d) Add License and Makefile
- [0097568](https://github.com/kubedb/postgres-archiver/commit/0097568) fix: added proper wal-handler need to do: fix kill container
- [5ebca34](https://github.com/kubedb/postgres-archiver/commit/5ebca34) added check for primary Signed-off-by: Emon46 <emon@appscode.com>
- [0829021](https://github.com/kubedb/postgres-archiver/commit/0829021) added: Different wal dir for different base-backup
- [2e66200](https://github.com/kubedb/postgres-archiver/commit/2e66200) added: basebackup handler update: intial listing func
- [b9c938f](https://github.com/kubedb/postgres-archiver/commit/b9c938f) update: added walg base-backup in bucket storage Signed-off-by: Emon46 <emon@appscode.com>
- [3c8c8da](https://github.com/kubedb/postgres-archiver/commit/3c8c8da) fix: fix go routine fr bucket listing update: added ticker in go routine update: combined two go-routine for listing file and bucket queue
- [21c1076](https://github.com/kubedb/postgres-archiver/commit/21c1076) added: wal-g push need to fix : filter list is not working
- [bc41b06](https://github.com/kubedb/postgres-archiver/commit/bc41b06) added - getExistingWalFiles func() updated - UpdateBucketPushQueue() Signed-off-by: Emon46 <emon@appscode.com>
- [b93fa5b](https://github.com/kubedb/postgres-archiver/commit/b93fa5b) added functions Signed-off-by: Emon331046 <emon@appscode.com>
- [7ded016](https://github.com/kubedb/postgres-archiver/commit/7ded016) added-func-name
- [cc7544d](https://github.com/kubedb/postgres-archiver/commit/cc7544d) file watcher local postgres watching Signed-off-by: Emon331046 <emon@appscode.com>



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.1.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.1.0)

- [b141665](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/b141665) Prepare for release v0.1.0 (#10)
- [bce9779](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/bce9779) Prepare for release v0.2.0 (#9)
- [31f5fc5](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/31f5fc5) Prepare for release v0.1.0 (#8)
- [57a7bdf](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/57a7bdf) Prepare for release v0.1.0-rc.1 (#7)
- [02a45da](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/02a45da) Prepare for release v0.1.0-rc.0 (#6)
- [1a6457c](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/1a6457c) Update flags and deps + Refactor (#5)
- [f32b56b](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/f32b56b) Delete .idea folder
- [e7f8135](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/e7f8135) clean up (#4)
- [06e7e70](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/06e7e70) clean up (#3)
- [b23dd63](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/b23dd63) Add build scripts
- [2e1dff2](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/2e1dff2) Add Postgres backup plugin (#1)
- [d0d156b](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/d0d156b) Test against K8s 1.27.0 (#2)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.3.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.3.0)

- [4a0356a](https://github.com/kubedb/postgres-restic-plugin/commit/4a0356a) Prepare for release v0.3.0 (#10)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.2.0](https://github.com/kubedb/provider-aws/releases/tag/v0.2.0)

- [ec4459c](https://github.com/kubedb/provider-aws/commit/ec4459c) Add dynamically start crd reconciler (#9)



## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.2.0](https://github.com/kubedb/provider-azure/releases/tag/v0.2.0)

- [0d449ff](https://github.com/kubedb/provider-azure/commit/0d449ff) Add dynamically start crd reconciler (#3)



## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.2.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.2.0)

- [a3de663](https://github.com/kubedb/provider-gcp/commit/a3de663) Add dynamically start crd reconciler (#3)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.40.0](https://github.com/kubedb/provisioner/releases/tag/v0.40.0)

- [715f4be8](https://github.com/kubedb/provisioner/commit/715f4be87) Prepare for release v0.40.0 (#66)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.27.0](https://github.com/kubedb/proxysql/releases/tag/v0.27.0)

- [1abe8cd0](https://github.com/kubedb/proxysql/commit/1abe8cd0) Prepare for release v0.27.0 (#318)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.33.0](https://github.com/kubedb/redis/releases/tag/v0.33.0)

- [9e36ab06](https://github.com/kubedb/redis/commit/9e36ab06) Prepare for release v0.33.0 (#506)
- [58b47ecb](https://github.com/kubedb/redis/commit/58b47ecb) Fix initContainer resources (#505)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.19.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.19.0)

- [c4d1d8b7](https://github.com/kubedb/redis-coordinator/commit/c4d1d8b7) Prepare for release v0.19.0 (#84)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.3.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.3.0)

- [c7105ef](https://github.com/kubedb/redis-restic-plugin/commit/c7105ef) Prepare for release v0.3.0 (#14)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.27.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.27.0)

- [125a1972](https://github.com/kubedb/replication-mode-detector/commit/125a1972) Prepare for release v0.27.0 (#248)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.16.0](https://github.com/kubedb/schema-manager/releases/tag/v0.16.0)

- [4aef1f64](https://github.com/kubedb/schema-manager/commit/4aef1f64) Prepare for release v0.16.0 (#91)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.25.0](https://github.com/kubedb/tests/releases/tag/v0.25.0)

- [a8a640dd](https://github.com/kubedb/tests/commit/a8a640dd) Prepare for release v0.25.0 (#278)
- [7149cc0c](https://github.com/kubedb/tests/commit/7149cc0c) Fix elasticsearch vertical scalinig tests. (#277)
- [3c71eea9](https://github.com/kubedb/tests/commit/3c71eea9) Fix build for autosclaer & verticalOps breaking api-changes (#276)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.16.0](https://github.com/kubedb/ui-server/releases/tag/v0.16.0)

- [4e1c32a2](https://github.com/kubedb/ui-server/commit/4e1c32a2) Prepare for release v0.16.0 (#100)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.16.0](https://github.com/kubedb/webhook-server/releases/tag/v0.16.0)

- [17659ce8](https://github.com/kubedb/webhook-server/commit/17659ce8) Prepare for release v0.16.0 (#77)




