---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2023.04.10
    name: Changelog-v2023.04.10
    parent: welcome
    weight: 20230410
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2023.04.10/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2023.04.10/
---

# KubeDB v2023.04.10 (2023-04-07)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.33.0](https://github.com/kubedb/apimachinery/releases/tag/v0.33.0)

- [16319573](https://github.com/kubedb/apimachinery/commit/16319573) Cleanup ci files
- [8a9b762e](https://github.com/kubedb/apimachinery/commit/8a9b762e) Rename UpgradeConstraints to UpdateConstraints in catalogs (#1035)
- [1f9d8cb4](https://github.com/kubedb/apimachinery/commit/1f9d8cb4) Add support for mongo version 6 (#1034)
- [c787eb94](https://github.com/kubedb/apimachinery/commit/c787eb94) Add Kafka monitor API (#1014)
- [3f1adae7](https://github.com/kubedb/apimachinery/commit/3f1adae7) Use enum generator for ops types (#1031)
- [d08e21e3](https://github.com/kubedb/apimachinery/commit/d08e21e3) Use ghcr.io for appscode/golang-dev (#1032)
- [b51ef1ea](https://github.com/kubedb/apimachinery/commit/b51ef1ea) Change return type of GetRequestType() func (#1030)
- [2e1cc0ab](https://github.com/kubedb/apimachinery/commit/2e1cc0ab) Use UpdateVersion instead of Upgrade in ops-manager (#1028)
- [b02d8800](https://github.com/kubedb/apimachinery/commit/b02d8800) Update for release Stash@v2023.03.13 (#1029)
- [03af3f01](https://github.com/kubedb/apimachinery/commit/03af3f01) Update workflows (Go 1.20, k8s 1.26) (#1027)
- [a5bd3816](https://github.com/kubedb/apimachinery/commit/a5bd3816) Refect Monitoring Agent StatAccessor API Update (#1024)
- [e867759c](https://github.com/kubedb/apimachinery/commit/e867759c) Update wrokflows (Go 1.20, k8s 1.26) (#1026)
- [854c4fa4](https://github.com/kubedb/apimachinery/commit/854c4fa4) Test against Kubernetes 1.26.0 (#1025)
- [1a3cbc58](https://github.com/kubedb/apimachinery/commit/1a3cbc58) Update for release Stash@v2023.02.28 (#1023)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.18.0](https://github.com/kubedb/autoscaler/releases/tag/v0.18.0)

- [311d6970](https://github.com/kubedb/autoscaler/commit/311d6970) Prepare for release v0.18.0 (#141)
- [ef53b5a7](https://github.com/kubedb/autoscaler/commit/ef53b5a7) Use ghcr.io
- [7ce66405](https://github.com/kubedb/autoscaler/commit/7ce66405) Use Homebrew in CI
- [a4880d45](https://github.com/kubedb/autoscaler/commit/a4880d45) Stop publishing to docker hub
- [a5e4e870](https://github.com/kubedb/autoscaler/commit/a5e4e870) Update package label in Docker files
- [61be362d](https://github.com/kubedb/autoscaler/commit/61be362d) Dynamically select runner type
- [572d61ab](https://github.com/kubedb/autoscaler/commit/572d61ab) Use ghcr.io for appscode/golang-dev (#140)
- [1bf5f5ef](https://github.com/kubedb/autoscaler/commit/1bf5f5ef) Update workflows (Go 1.20, k8s 1.26) (#139)
- [901be09a](https://github.com/kubedb/autoscaler/commit/901be09a) Update wrokflows (Go 1.20, k8s 1.26) (#138)
- [12f24074](https://github.com/kubedb/autoscaler/commit/12f24074) Test against Kubernetes 1.26.0 (#137)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.33.0](https://github.com/kubedb/cli/releases/tag/v0.33.0)

- [4849d48b](https://github.com/kubedb/cli/commit/4849d48b) Prepare for release v0.33.0 (#704)
- [7f680b1e](https://github.com/kubedb/cli/commit/7f680b1e) Cleanup CI
- [2c607eef](https://github.com/kubedb/cli/commit/2c607eef) Use ghcr.io for appscode/golang-dev (#703)
- [4867dac1](https://github.com/kubedb/cli/commit/4867dac1) Update workflows (Go 1.20, k8s 1.26) (#702)
- [3ed34cca](https://github.com/kubedb/cli/commit/3ed34cca) Update wrokflows (Go 1.20, k8s 1.26) (#701)
- [26fa3901](https://github.com/kubedb/cli/commit/26fa3901) Test against Kubernetes 1.26.0 (#700)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.9.0](https://github.com/kubedb/dashboard/releases/tag/v0.9.0)

- [a796df9](https://github.com/kubedb/dashboard/commit/a796df9) Prepare for release v0.9.0 (#69)
- [396b3c9](https://github.com/kubedb/dashboard/commit/396b3c9) Use ghcr.io
- [0014ad3](https://github.com/kubedb/dashboard/commit/0014ad3) Stop publishing to docker hub
- [6f5ca8e](https://github.com/kubedb/dashboard/commit/6f5ca8e) Dynamically select runner type
- [55c6d4b](https://github.com/kubedb/dashboard/commit/55c6d4b) Use ghcr.io for appscode/golang-dev (#68)
- [d6c42ca](https://github.com/kubedb/dashboard/commit/d6c42ca) Update workflows (Go 1.20, k8s 1.26) (#67)
- [1367b94](https://github.com/kubedb/dashboard/commit/1367b94) Update wrokflows (Go 1.20, k8s 1.26) (#66)
- [8085731](https://github.com/kubedb/dashboard/commit/8085731) Test against Kubernetes 1.26.0 (#65)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.33.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.33.0)

- [fc4188e9](https://github.com/kubedb/elasticsearch/commit/fc4188e94) Prepare for release v0.33.0 (#635)
- [3bfc204d](https://github.com/kubedb/elasticsearch/commit/3bfc204de) Update e2e workflow
- [62b43e46](https://github.com/kubedb/elasticsearch/commit/62b43e46c) Use ghcr.io
- [3fee0f94](https://github.com/kubedb/elasticsearch/commit/3fee0f947) Update workflows
- [d8dad30e](https://github.com/kubedb/elasticsearch/commit/d8dad30e9) Dynamically select runner type
- [845dd87c](https://github.com/kubedb/elasticsearch/commit/845dd87cb) Use ghcr.io for appscode/golang-dev (#633)
- [579d778d](https://github.com/kubedb/elasticsearch/commit/579d778de) Update workflows (Go 1.20, k8s 1.26) (#632)
- [c0aa077f](https://github.com/kubedb/elasticsearch/commit/c0aa077f9) Update wrokflows (Go 1.20, k8s 1.26) (#631)
- [9269ed75](https://github.com/kubedb/elasticsearch/commit/9269ed75c) Test against Kubernetes 1.26.0 (#630)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2023.04.10](https://github.com/kubedb/installer/releases/tag/v2023.04.10)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.4.0](https://github.com/kubedb/kafka/releases/tag/v0.4.0)

- [6d43376](https://github.com/kubedb/kafka/commit/6d43376) Prepare for release v0.4.0 (#22)
- [9d9beea](https://github.com/kubedb/kafka/commit/9d9beea) Update e2e workflow
- [1fc0185](https://github.com/kubedb/kafka/commit/1fc0185) Cleanup Makefile
- [c4b3450](https://github.com/kubedb/kafka/commit/c4b3450) Update workflows - Stop publishing to docker hub - Enable e2e tests - Use homebrew to install tools
- [50c408e](https://github.com/kubedb/kafka/commit/50c408e) Remove Kafka advertised.listeners config from controller node (#21)
- [2e80fc8](https://github.com/kubedb/kafka/commit/2e80fc8) Add support for monitoring (#19)
- [58894bb](https://github.com/kubedb/kafka/commit/58894bb) Use ghcr.io for appscode/golang-dev (#20)
- [c7b1158](https://github.com/kubedb/kafka/commit/c7b1158) Dynamically select runner type
- [d5f02d5](https://github.com/kubedb/kafka/commit/d5f02d5) Update workflows (Go 1.20, k8s 1.26) (#18)
- [842361d](https://github.com/kubedb/kafka/commit/842361d) Test against Kubernetes 1.26.0 (#16)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.17.0](https://github.com/kubedb/mariadb/releases/tag/v0.17.0)

- [d0ab53a5](https://github.com/kubedb/mariadb/commit/d0ab53a5) Prepare for release v0.17.0 (#205)
- [a7cb3789](https://github.com/kubedb/mariadb/commit/a7cb3789) Update e2e workflow
- [03182a15](https://github.com/kubedb/mariadb/commit/03182a15) Use ghcr.io
- [48a0ae24](https://github.com/kubedb/mariadb/commit/48a0ae24) Update workflows - Stop publishing to docker hub - Enable e2e tests - Use homebrew to install tools
- [b5fc163d](https://github.com/kubedb/mariadb/commit/b5fc163d) Use ghcr.io for appscode/golang-dev (#204)
- [ffd17645](https://github.com/kubedb/mariadb/commit/ffd17645) Dynamically select runner type
- [9e18fbf6](https://github.com/kubedb/mariadb/commit/9e18fbf6) Update workflows (Go 1.20, k8s 1.26) (#203)
- [02c9169d](https://github.com/kubedb/mariadb/commit/02c9169d) Update wrokflows (Go 1.20, k8s 1.26) (#202)
- [ccddab5f](https://github.com/kubedb/mariadb/commit/ccddab5f) Test against Kubernetes 1.26.0 (#201)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.13.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.13.0)

- [838c879](https://github.com/kubedb/mariadb-coordinator/commit/838c879) Prepare for release v0.13.0 (#78)
- [1242437](https://github.com/kubedb/mariadb-coordinator/commit/1242437) Update CI
- [561fd55](https://github.com/kubedb/mariadb-coordinator/commit/561fd55) Use ghcr.io for appscode/golang-dev (#77)
- [0f67cb9](https://github.com/kubedb/mariadb-coordinator/commit/0f67cb9) DYnamically select runner type
- [adbb2d3](https://github.com/kubedb/mariadb-coordinator/commit/adbb2d3) Update workflows (Go 1.20, k8s 1.26) (#76)
- [e27c8f6](https://github.com/kubedb/mariadb-coordinator/commit/e27c8f6) Update wrokflows (Go 1.20, k8s 1.26) (#75)
- [633dc41](https://github.com/kubedb/mariadb-coordinator/commit/633dc41) Test against Kubernetes 1.26.0 (#74)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.26.0](https://github.com/kubedb/memcached/releases/tag/v0.26.0)

- [f7975e7b](https://github.com/kubedb/memcached/commit/f7975e7b) Prepare for release v0.26.0 (#391)
- [81e92a08](https://github.com/kubedb/memcached/commit/81e92a08) Update e2e workflow
- [dc8ccbf4](https://github.com/kubedb/memcached/commit/dc8ccbf4) Cleanup CI
- [318b3f14](https://github.com/kubedb/memcached/commit/318b3f14) Update workflows - Stop publishing to docker hub - Enable e2e tests - Use homebrew to install tools
- [2ca568aa](https://github.com/kubedb/memcached/commit/2ca568aa) Use ghcr.io for appscode/golang-dev (#390)
- [d85f38d2](https://github.com/kubedb/memcached/commit/d85f38d2) Dynamically select runner type
- [962f3daa](https://github.com/kubedb/memcached/commit/962f3daa) Update workflows (Go 1.20, k8s 1.26) (#389)
- [c1eb2df8](https://github.com/kubedb/memcached/commit/c1eb2df8) Update wrokflows (Go 1.20, k8s 1.26) (#388)
- [48b784c8](https://github.com/kubedb/memcached/commit/48b784c8) Test against Kubernetes 1.26.0 (#387)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.26.0](https://github.com/kubedb/mongodb/releases/tag/v0.26.0)

- [9e52ecf2](https://github.com/kubedb/mongodb/commit/9e52ecf2) Prepare for release v0.26.0 (#545)
- [987a14ba](https://github.com/kubedb/mongodb/commit/987a14ba) Replace Mongos bootstrap container with postStart hook (#542)
- [8f2437cf](https://github.com/kubedb/mongodb/commit/8f2437cf) Use --timeout=24h for e2e tests
- [41c3a6e9](https://github.com/kubedb/mongodb/commit/41c3a6e9) Rename ref flag to rest for e2e workflows
- [e7e3203b](https://github.com/kubedb/mongodb/commit/e7e3203b) Customize installer ref for e2e workflows
- [d57e6f31](https://github.com/kubedb/mongodb/commit/d57e6f31) Issue license key for kubedb
- [bcb712e7](https://github.com/kubedb/mongodb/commit/bcb712e7) Cleanup CI
- [3bd0d6cc](https://github.com/kubedb/mongodb/commit/3bd0d6cc) Stop publishing to docker hub
- [1c5327b5](https://github.com/kubedb/mongodb/commit/1c5327b5) Update db versions for e2e tests
- [7a8e9b25](https://github.com/kubedb/mongodb/commit/7a8e9b25) Speed up e2e tests
- [471e3259](https://github.com/kubedb/mongodb/commit/471e3259) Use brew to install tools
- [71574901](https://github.com/kubedb/mongodb/commit/71574901) Use fircracker vms for e2e tests
- [220b1b14](https://github.com/kubedb/mongodb/commit/220b1b14) Update e2e workflow
- [5bf24d7b](https://github.com/kubedb/mongodb/commit/5bf24d7b) Use ghcr.io for appscode/golang-dev (#541)
- [553dc5b5](https://github.com/kubedb/mongodb/commit/553dc5b5) Dynamically select runner type
- [0e94ca5a](https://github.com/kubedb/mongodb/commit/0e94ca5a) Update workflows (Go 1.20, k8s 1.26) (#539)
- [c8858e12](https://github.com/kubedb/mongodb/commit/c8858e12) Update wrokflows (Go 1.20, k8s 1.26) (#538)
- [b9e36634](https://github.com/kubedb/mongodb/commit/b9e36634) Test against Kubernetes 1.26.0 (#537)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.26.0](https://github.com/kubedb/mysql/releases/tag/v0.26.0)

- [f4897ade](https://github.com/kubedb/mysql/commit/f4897ade) Prepare for release v0.26.0 (#530)
- [942f3675](https://github.com/kubedb/mysql/commit/942f3675) Add MySQL-5 MaxLen Check (#528)
- [d3dcd00e](https://github.com/kubedb/mysql/commit/d3dcd00e) Update e2e workflows
- [ec1ee2a3](https://github.com/kubedb/mysql/commit/ec1ee2a3) Cleanup CI
- [9330ee56](https://github.com/kubedb/mysql/commit/9330ee56) Update workflows - Stop publishing to docker hub - Enable e2e tests - Use homebrew to install tools
- [4382bb9a](https://github.com/kubedb/mysql/commit/4382bb9a) Use ghcr.io for appscode/golang-dev (#527)
- [dbeadab3](https://github.com/kubedb/mysql/commit/dbeadab3) Dynamically select runner type
- [a693489b](https://github.com/kubedb/mysql/commit/a693489b) Update workflows (Go 1.20, k8s 1.26) (#526)
- [67c8a8f0](https://github.com/kubedb/mysql/commit/67c8a8f0) Update wrokflows (Go 1.20, k8s 1.26) (#525)
- [4bd25e73](https://github.com/kubedb/mysql/commit/4bd25e73) Test against Kubernetes 1.26.0 (#524)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.11.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.11.0)

- [16c77c8](https://github.com/kubedb/mysql-coordinator/commit/16c77c8) Prepare for release v0.11.0 (#76)
- [50ad81b](https://github.com/kubedb/mysql-coordinator/commit/50ad81b) Cleanup CI
- [b920384](https://github.com/kubedb/mysql-coordinator/commit/b920384) Use ghcr.io for appscode/golang-dev (#75)
- [f1de5ed](https://github.com/kubedb/mysql-coordinator/commit/f1de5ed) Dynamically select runner type
- [dc31944](https://github.com/kubedb/mysql-coordinator/commit/dc31944) Update workflows (Go 1.20, k8s 1.26) (#74)
- [244154d](https://github.com/kubedb/mysql-coordinator/commit/244154d) Update workflows (Go 1.20, k8s 1.26) (#73)
- [505ecfa](https://github.com/kubedb/mysql-coordinator/commit/505ecfa) Test against Kubernetes 1.26.0 (#72)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.11.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.11.0)

- [710fc9f](https://github.com/kubedb/mysql-router-init/commit/710fc9f) Cleanup CI
- [2fe6586](https://github.com/kubedb/mysql-router-init/commit/2fe6586) Use ghcr.io for appscode/golang-dev (#34)
- [989bb29](https://github.com/kubedb/mysql-router-init/commit/989bb29) Dynamically select runner type
- [2d00c02](https://github.com/kubedb/mysql-router-init/commit/2d00c02) Update workflows (Go 1.20, k8s 1.26) (#33)
- [1a70e0c](https://github.com/kubedb/mysql-router-init/commit/1a70e0c) Update wrokflows (Go 1.20, k8s 1.26) (#32)
- [a68b30e](https://github.com/kubedb/mysql-router-init/commit/a68b30e) Test against Kubernetes 1.26.0 (#31)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.20.0](https://github.com/kubedb/ops-manager/releases/tag/v0.20.0)

- [e1e3a251](https://github.com/kubedb/ops-manager/commit/e1e3a251) Prepare for release v0.20.0 (#434)
- [972654d8](https://github.com/kubedb/ops-manager/commit/972654d8) Rename UpgradeConstraints to UpdateConstrints in catalogs (#432)
- [f2a545b2](https://github.com/kubedb/ops-manager/commit/f2a545b2) Fix mongodb Upgrade (#431)
- [ce8027f0](https://github.com/kubedb/ops-manager/commit/ce8027f0) Cleanup CI
- [64bef08d](https://github.com/kubedb/ops-manager/commit/64bef08d) Add cve report to version upgrade Recommendation (#395)
- [34a8838e](https://github.com/kubedb/ops-manager/commit/34a8838e) Use ghcr.io for appscode/golang-dev (#430)
- [fcd0704d](https://github.com/kubedb/ops-manager/commit/fcd0704d) Use Redis and Sentinel Client from db-client-go (#429)
- [55c6e16f](https://github.com/kubedb/ops-manager/commit/55c6e16f) Use UpdateVersion instead of Upgrade (#427)
- [51554040](https://github.com/kubedb/ops-manager/commit/51554040) Auto detect runs-on label (#428)
- [a41c70fb](https://github.com/kubedb/ops-manager/commit/a41c70fb) Use self-hosted runners
- [bb60637d](https://github.com/kubedb/ops-manager/commit/bb60637d) Add Postgres UpdateVersion support (#425)
- [ac890091](https://github.com/kubedb/ops-manager/commit/ac890091) Update workflows (Go 1.20, k8s 1.26) (#424)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.20.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.20.0)

- [0783793f](https://github.com/kubedb/percona-xtradb/commit/0783793f) Prepare for release v0.20.0 (#307)
- [281c5eac](https://github.com/kubedb/percona-xtradb/commit/281c5eac) Update e2e workflow
- [d16efe09](https://github.com/kubedb/percona-xtradb/commit/d16efe09) Cleanup CI
- [f533248c](https://github.com/kubedb/percona-xtradb/commit/f533248c) Update workflows - Stop publishing to docker hub - Enable e2e tests - Use homebrew to install tools
- [c6e08088](https://github.com/kubedb/percona-xtradb/commit/c6e08088) Use ghcr.io for appscode/golang-dev (#305)
- [227526f0](https://github.com/kubedb/percona-xtradb/commit/227526f0) Dynamically select runner type
- [cd6321b2](https://github.com/kubedb/percona-xtradb/commit/cd6321b2) Update workflows (Go 1.20, k8s 1.26) (#304)
- [9840d83a](https://github.com/kubedb/percona-xtradb/commit/9840d83a) Update wrokflows (Go 1.20, k8s 1.26) (#303)
- [33c6116b](https://github.com/kubedb/percona-xtradb/commit/33c6116b) Test against Kubernetes 1.26.0 (#302)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.6.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.6.0)

- [cc06348](https://github.com/kubedb/percona-xtradb-coordinator/commit/cc06348) Prepare for release v0.6.0 (#35)
- [566cdac](https://github.com/kubedb/percona-xtradb-coordinator/commit/566cdac) Cleanup CI
- [82bab80](https://github.com/kubedb/percona-xtradb-coordinator/commit/82bab80) Use ghcr.io for appscode/golang-dev (#34)
- [2e9c1f5](https://github.com/kubedb/percona-xtradb-coordinator/commit/2e9c1f5) Dynamically select runner type
- [b21a83d](https://github.com/kubedb/percona-xtradb-coordinator/commit/b21a83d) Update workflows (Go 1.20, k8s 1.26) (#33)
- [a55967f](https://github.com/kubedb/percona-xtradb-coordinator/commit/a55967f) Update wrokflows (Go 1.20, k8s 1.26) (#32)
- [a8eaa03](https://github.com/kubedb/percona-xtradb-coordinator/commit/a8eaa03) Test against Kubernetes 1.26.0 (#31)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.17.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.17.0)

- [c958b2b6](https://github.com/kubedb/pg-coordinator/commit/c958b2b6) Prepare for release v0.17.0 (#118)
- [8faaf376](https://github.com/kubedb/pg-coordinator/commit/8faaf376) Cleanup CI
- [cbc12702](https://github.com/kubedb/pg-coordinator/commit/cbc12702) Use ghcr.io for appscode/golang-dev (#117)
- [cbfef2aa](https://github.com/kubedb/pg-coordinator/commit/cbfef2aa) Dynamically select runner type
- [57f2ad58](https://github.com/kubedb/pg-coordinator/commit/57f2ad58) Update workflows (Go 1.20, k8s 1.26) (#116)
- [fb81176a](https://github.com/kubedb/pg-coordinator/commit/fb81176a) Test against Kubernetes 1.26.0 (#114)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.20.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.20.0)

- [f3f89b3e](https://github.com/kubedb/pgbouncer/commit/f3f89b3e) Prepare for release v0.20.0 (#273)
- [7b1391e1](https://github.com/kubedb/pgbouncer/commit/7b1391e1) Update e2e workflows
- [3ae31397](https://github.com/kubedb/pgbouncer/commit/3ae31397) Cleanup CI
- [e8fe48b3](https://github.com/kubedb/pgbouncer/commit/e8fe48b3) Update workflows - Stop publishing to docker hub - Enable e2e tests - Use homebrew to install tools
- [153effe0](https://github.com/kubedb/pgbouncer/commit/153effe0) Use ghcr.io for appscode/golang-dev (#271)
- [d141e211](https://github.com/kubedb/pgbouncer/commit/d141e211) Dynamically select runner type



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.33.0](https://github.com/kubedb/postgres/releases/tag/v0.33.0)

- [c95f2b59](https://github.com/kubedb/postgres/commit/c95f2b59d) Prepare for release v0.33.0 (#636)
- [8d53d60a](https://github.com/kubedb/postgres/commit/8d53d60a2) Update e2e workflows
- [fa31d6e3](https://github.com/kubedb/postgres/commit/fa31d6e33) Cleanup CI
- [40cf94c4](https://github.com/kubedb/postgres/commit/40cf94c4f) Update workflows - Stop publishing to docker hub - Enable e2e tests - Use homebrew to install tools
- [78b13fe0](https://github.com/kubedb/postgres/commit/78b13fe00) Use ghcr.io for appscode/golang-dev (#634)
- [c5ca8a99](https://github.com/kubedb/postgres/commit/c5ca8a99d) Dynamically select runner type
- [6005fce1](https://github.com/kubedb/postgres/commit/6005fce14) Update workflows (Go 1.20, k8s 1.26) (#633)
- [26751826](https://github.com/kubedb/postgres/commit/267518268) Update wrokflows (Go 1.20, k8s 1.26) (#632)
- [aad6863b](https://github.com/kubedb/postgres/commit/aad6863b7) Test against Kubernetes 1.26.0 (#631)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.33.0](https://github.com/kubedb/provisioner/releases/tag/v0.33.0)

- [9e5ad46a](https://github.com/kubedb/provisioner/commit/9e5ad46a1) Prepare for release v0.33.0 (#44)
- [e1a68944](https://github.com/kubedb/provisioner/commit/e1a689443) Update e2e workflow
- [0c177001](https://github.com/kubedb/provisioner/commit/0c1770015) Update workflows - Stop publishing to docker hub - Enable e2e tests - Use homebrew to install tools
- [f30f3b60](https://github.com/kubedb/provisioner/commit/f30f3b60c) Use ghcr.io for appscode/golang-dev (#43)
- [1ef0787b](https://github.com/kubedb/provisioner/commit/1ef0787b8) Dynamically select runner type



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.20.0](https://github.com/kubedb/proxysql/releases/tag/v0.20.0)

- [b49d3af7](https://github.com/kubedb/proxysql/commit/b49d3af7) Prepare for release v0.20.0 (#288)
- [61044810](https://github.com/kubedb/proxysql/commit/61044810) Update e2e workflow
- [32ddb10e](https://github.com/kubedb/proxysql/commit/32ddb10e) Cleanup CI
- [2f16bec6](https://github.com/kubedb/proxysql/commit/2f16bec6) Update workflows - Stop publishing to docker hub - Enable e2e tests - Use homebrew to install tools
- [0724ef23](https://github.com/kubedb/proxysql/commit/0724ef23) Use ghcr.io for appscode/golang-dev (#287)
- [73541f5c](https://github.com/kubedb/proxysql/commit/73541f5c) Dynamically select runner type
- [b28209ab](https://github.com/kubedb/proxysql/commit/b28209ab) Update workflows (Go 1.20, k8s 1.26) (#286)
- [7f682fe8](https://github.com/kubedb/proxysql/commit/7f682fe8) Update wrokflows (Go 1.20, k8s 1.26) (#285)
- [7daac3a4](https://github.com/kubedb/proxysql/commit/7daac3a4) Test against Kubernetes 1.26.0 (#284)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.26.0](https://github.com/kubedb/redis/releases/tag/v0.26.0)

- [a52ff8e8](https://github.com/kubedb/redis/commit/a52ff8e8) Prepare for release v0.26.0 (#458)
- [f6b74025](https://github.com/kubedb/redis/commit/f6b74025) Update e2e workflow
- [7799c78b](https://github.com/kubedb/redis/commit/7799c78b) Cleanup CI
- [a5abe6d2](https://github.com/kubedb/redis/commit/a5abe6d2) Update workflows - Stop publishing to docker hub - Enable e2e tests - Use homebrew to install tools
- [b4a13c26](https://github.com/kubedb/redis/commit/b4a13c26) Use ghcr.io for appscode/golang-dev (#456)
- [a6df02d8](https://github.com/kubedb/redis/commit/a6df02d8) Dynamically select runner type
- [11d13a42](https://github.com/kubedb/redis/commit/11d13a42) Update workflows (Go 1.20, k8s 1.26) (#455)
- [2366e2db](https://github.com/kubedb/redis/commit/2366e2db) Update wrokflows (Go 1.20, k8s 1.26) (#454)
- [4ffdad8b](https://github.com/kubedb/redis/commit/4ffdad8b) Test against Kubernetes 1.26.0 (#453)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.12.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.12.0)

- [d04d7c3](https://github.com/kubedb/redis-coordinator/commit/d04d7c3) Prepare for release v0.12.0 (#68)
- [a28d24b](https://github.com/kubedb/redis-coordinator/commit/a28d24b) Cleanup CI
- [79ee2a2](https://github.com/kubedb/redis-coordinator/commit/79ee2a2) Use ghcr.io for appscode/golang-dev (#67)
- [45e832e](https://github.com/kubedb/redis-coordinator/commit/45e832e) Dynamically select runner type
- [9d96b05](https://github.com/kubedb/redis-coordinator/commit/9d96b05) Update workflows (Go 1.20, k8s 1.26) (#66)
- [983ef3b](https://github.com/kubedb/redis-coordinator/commit/983ef3b) Update wrokflows (Go 1.20, k8s 1.26) (#65)
- [e12472f](https://github.com/kubedb/redis-coordinator/commit/e12472f) Test against Kubernetes 1.26.0 (#64)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.20.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.20.0)

- [181606ec](https://github.com/kubedb/replication-mode-detector/commit/181606ec) Prepare for release v0.20.0 (#230)
- [90e75258](https://github.com/kubedb/replication-mode-detector/commit/90e75258) Cleanup CI
- [9b1ffb20](https://github.com/kubedb/replication-mode-detector/commit/9b1ffb20) Use ghcr.io for appscode/golang-dev (#229)
- [83e4656c](https://github.com/kubedb/replication-mode-detector/commit/83e4656c) Dynamically select runner type
- [160bd418](https://github.com/kubedb/replication-mode-detector/commit/160bd418) Update workflows (Go 1.20, k8s 1.26) (#228)
- [306b14ac](https://github.com/kubedb/replication-mode-detector/commit/306b14ac) Test against Kubernetes 1.26.0 (#227)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.9.0](https://github.com/kubedb/schema-manager/releases/tag/v0.9.0)

- [e0f28fd4](https://github.com/kubedb/schema-manager/commit/e0f28fd4) Prepare for release v0.9.0 (#70)
- [d633359a](https://github.com/kubedb/schema-manager/commit/d633359a) Cleanup CI
- [5386ed59](https://github.com/kubedb/schema-manager/commit/5386ed59) Use ghcr.io for appscode/golang-dev (#69)
- [8e517e3d](https://github.com/kubedb/schema-manager/commit/8e517e3d) Dynamically select runner type
- [12abe459](https://github.com/kubedb/schema-manager/commit/12abe459) Update workflows (Go 1.20, k8s 1.26) (#68)
- [6b7412b4](https://github.com/kubedb/schema-manager/commit/6b7412b4) Update wrokflows (Go 1.20, k8s 1.26) (#67)
- [b1734e39](https://github.com/kubedb/schema-manager/commit/b1734e39) Test against Kubernetes 1.26.0 (#66)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.18.0](https://github.com/kubedb/tests/releases/tag/v0.18.0)

- [655e8669](https://github.com/kubedb/tests/commit/655e8669) Prepare for release v0.18.0 (#225)
- [27b5548f](https://github.com/kubedb/tests/commit/27b5548f) Use UpdateVersion instead of Upgrade (#224)
- [3a57f668](https://github.com/kubedb/tests/commit/3a57f668) Fix mongo e2e test (#223)
- [76a8abdd](https://github.com/kubedb/tests/commit/76a8abdd) Update deps
- [0bb20a34](https://github.com/kubedb/tests/commit/0bb20a34) Replace deprecated CurrentGinkgoTestDescription() with CurrentSpecReport()
- [4adb3f61](https://github.com/kubedb/tests/commit/4adb3f61) Cleanup CI
- [c3b1a205](https://github.com/kubedb/tests/commit/c3b1a205) Add MongoDB Hidden-node (#205)
- [1d7f62bb](https://github.com/kubedb/tests/commit/1d7f62bb) Use ghcr.io for appscode/golang-dev (#222)
- [bf3df4dc](https://github.com/kubedb/tests/commit/bf3df4dc) Dynamically select runner type
- [1a5a1d04](https://github.com/kubedb/tests/commit/1a5a1d04) Update workflows (Go 1.20, k8s 1.26) (#221)
- [016687ad](https://github.com/kubedb/tests/commit/016687ad) Update workflows (Go 1.20, k8s 1.26) (#220)
- [3155a749](https://github.com/kubedb/tests/commit/3155a749) Test against Kubernetes 1.26.0 (#219)
- [a58933cd](https://github.com/kubedb/tests/commit/a58933cd) Fix typo in  MySQL tests (#218)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.9.0](https://github.com/kubedb/ui-server/releases/tag/v0.9.0)

- [49a09f28](https://github.com/kubedb/ui-server/commit/49a09f28) Prepare for release v0.9.0 (#73)
- [eaba6a4d](https://github.com/kubedb/ui-server/commit/eaba6a4d) Cleanup CI
- [779d75a5](https://github.com/kubedb/ui-server/commit/779d75a5) Use ghcr.io for appscode/golang-dev (#72)
- [9a94bf21](https://github.com/kubedb/ui-server/commit/9a94bf21) Dynamically select runner type
- [3c053ff1](https://github.com/kubedb/ui-server/commit/3c053ff1) Update workflows (Go 1.20, k8s 1.26) (#71)
- [7adf8e99](https://github.com/kubedb/ui-server/commit/7adf8e99) Update wrokflows (Go 1.20, k8s 1.26) (#70)
- [97704312](https://github.com/kubedb/ui-server/commit/97704312) Test against Kubernetes 1.26.0 (#69)
- [8fa7286a](https://github.com/kubedb/ui-server/commit/8fa7286a) Add context to Redis Client Builder (#67)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.9.0](https://github.com/kubedb/webhook-server/releases/tag/v0.9.0)

- [a50a54aa](https://github.com/kubedb/webhook-server/commit/a50a54aa) Prepare for release v0.9.0 (#57)
- [34807440](https://github.com/kubedb/webhook-server/commit/34807440) Update e2e workflow
- [38c71e46](https://github.com/kubedb/webhook-server/commit/38c71e46) Update workflows - Stop publishing to docker hub - Enable e2e tests - Use homebrew to install tools
- [1b32b482](https://github.com/kubedb/webhook-server/commit/1b32b482) Use ghcr.io for appscode/golang-dev (#56)
- [6a15e2e4](https://github.com/kubedb/webhook-server/commit/6a15e2e4) Dynamically select runner type
- [4d08d51b](https://github.com/kubedb/webhook-server/commit/4d08d51b) Update workflows (Go 1.20, k8s 1.26) (#55)
- [3ea6558a](https://github.com/kubedb/webhook-server/commit/3ea6558a) Update wrokflows (Go 1.20, k8s 1.26) (#54)
- [b0edce4a](https://github.com/kubedb/webhook-server/commit/b0edce4a) Test against Kubernetes 1.26.0 (#53)




