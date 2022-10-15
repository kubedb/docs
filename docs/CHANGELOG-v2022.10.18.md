---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2022.10.18
    name: Changelog-v2022.10.18
    parent: welcome
    weight: 20221018
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2022.10.18/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2022.10.18/
---

# KubeDB v2022.10.18 (2022-10-15)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.29.0](https://github.com/kubedb/apimachinery/releases/tag/v0.29.0)

- [daeafa99](https://github.com/kubedb/apimachinery/commit/daeafa99) Update crds properly
- [fc357a90](https://github.com/kubedb/apimachinery/commit/fc357a90) Make exporter optional in ProxySQL catalog (#992)
- [7d01a527](https://github.com/kubedb/apimachinery/commit/7d01a527) Add conditions for postgres logical replication. (#990)
- [197a2568](https://github.com/kubedb/apimachinery/commit/197a2568) Remove storage autoscaler from Sentinel spec (#991)
- [2dae1fa1](https://github.com/kubedb/apimachinery/commit/2dae1fa1) Add GetSystemUserSecret Heplers on PerconaXtraDB (#989)
- [35e1d5e5](https://github.com/kubedb/apimachinery/commit/35e1d5e5) Make OpsRequestType specific to databases (#988)
- [e7310243](https://github.com/kubedb/apimachinery/commit/e7310243) Add Redis Sentinel Ops Requests APIs (#958)
- [b937b3dc](https://github.com/kubedb/apimachinery/commit/b937b3dc) Update digest.go
- [1b1732a9](https://github.com/kubedb/apimachinery/commit/1b1732a9) Change ProxySQL backend to a local obj ref (#987)
- [31c66a34](https://github.com/kubedb/apimachinery/commit/31c66a34) Include Arbiter & hidden nodes in MongoAutoscaler (#979)
- [3c2f4a7a](https://github.com/kubedb/apimachinery/commit/3c2f4a7a) Add autoscaler types for Postgres (#969)
- [9f60ebbe](https://github.com/kubedb/apimachinery/commit/9f60ebbe) Add GetAuthSecretName() helper (#986)
- [b48d0118](https://github.com/kubedb/apimachinery/commit/b48d0118) Ignore TLS certificate validation when using private domains (#984)
- [11a09d52](https://github.com/kubedb/apimachinery/commit/11a09d52) Use stash.appscode.dev/apimachinery@v0.23.0 (#983)
- [cb611290](https://github.com/kubedb/apimachinery/commit/cb611290) Remove duplicate short name from redis sentinel (#982)
- [f5eabfc2](https://github.com/kubedb/apimachinery/commit/f5eabfc2) Fix typo 'SuccessfullyRestatedStatefulSet' (#980)
- [4f6d7eac](https://github.com/kubedb/apimachinery/commit/4f6d7eac) Test against Kubernetes 1.25.0 (#981)
- [c0388bc2](https://github.com/kubedb/apimachinery/commit/c0388bc2) Use authSecret.externallyManaged field (#978)
- [7f39736a](https://github.com/kubedb/apimachinery/commit/7f39736a) Remove default values from authSecret (#977)
- [2d9abdb4](https://github.com/kubedb/apimachinery/commit/2d9abdb4) Support different types of secrets and password rotation (#976)
- [f01cf5b9](https://github.com/kubedb/apimachinery/commit/f01cf5b9) Using opsRequestOpts for elastic,maria & percona (#970)
- [e26f6417](https://github.com/kubedb/apimachinery/commit/e26f6417) Fix typos of Postgres Logical Replication CRDs. (#974)
- [d43f454e](https://github.com/kubedb/apimachinery/commit/d43f454e) Check for PDB version only once (#975)
- [fb5283cd](https://github.com/kubedb/apimachinery/commit/fb5283cd) Handle status conversion for PDB (#973)
- [7263b503](https://github.com/kubedb/apimachinery/commit/7263b503) Update kutil
- [5c643b97](https://github.com/kubedb/apimachinery/commit/5c643b97) Use Go 1.19
- [a0b96812](https://github.com/kubedb/apimachinery/commit/a0b96812) Fix mergo dependency
- [b7b93597](https://github.com/kubedb/apimachinery/commit/b7b93597) Use k8s 1.25.1 libs (#971)
- [c1f407b0](https://github.com/kubedb/apimachinery/commit/c1f407b0) Add MySQLAutoscaler support (#968)
- [693f5243](https://github.com/kubedb/apimachinery/commit/693f5243) Add MongoDB HiddenNode support (#956)
- [0b3be441](https://github.com/kubedb/apimachinery/commit/0b3be441) Add Postgres Publisher & Subscriber CRDs (#967)
- [71947dec](https://github.com/kubedb/apimachinery/commit/71947dec) Update README.md
- [818f48fa](https://github.com/kubedb/apimachinery/commit/818f48fa) Add redis-sentinel autoscaler types (#965)
- [011938c4](https://github.com/kubedb/apimachinery/commit/011938c4) Add PerconaXtraDB OpsReq and Autoscaler APIs (#953)
- [b57e7099](https://github.com/kubedb/apimachinery/commit/b57e7099) Add RedisAutoscaler support (#963)
- [2ccea895](https://github.com/kubedb/apimachinery/commit/2ccea895) Remove `DisableScaleDown` field from autoscaler (#966)
- [02b47709](https://github.com/kubedb/apimachinery/commit/02b47709) Support PDB v1 or v1beta1 api based on k8s version (#964)
- [e2d0bb4f](https://github.com/kubedb/apimachinery/commit/e2d0bb4f) Stop using removed apis in Kubernetes 1.25 (#962)
- [722a1bc1](https://github.com/kubedb/apimachinery/commit/722a1bc1) Use health checker types from kmodules (#961)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.14.0](https://github.com/kubedb/autoscaler/releases/tag/v0.14.0)

- [e798cfae](https://github.com/kubedb/autoscaler/commit/e798cfae) Prepare for release v0.14.0 (#120)
- [defb6306](https://github.com/kubedb/autoscaler/commit/defb6306) Use password-generator@v0.2.9 (#119)
- [99ffb7a4](https://github.com/kubedb/autoscaler/commit/99ffb7a4) Prepare for release v0.14.0-rc.0 (#118)
- [deec8a47](https://github.com/kubedb/autoscaler/commit/deec8a47) Update dependencies (#117)
- [c06eff58](https://github.com/kubedb/autoscaler/commit/c06eff58) Support mongo arbiter & hidden nodes (#115)
- [513b5fb4](https://github.com/kubedb/autoscaler/commit/513b5fb4) Add support for Postgres Autoscaler (#112)
- [87dd17fe](https://github.com/kubedb/autoscaler/commit/87dd17fe) Using opsRequestOpts on storageAutoscalers to satisfy cmp.Equal() (#116)
- [a8afe242](https://github.com/kubedb/autoscaler/commit/a8afe242) Test against Kubernetes 1.25.0 (#114)
- [65d3869c](https://github.com/kubedb/autoscaler/commit/65d3869c) Test against Kubernetes 1.25.0 (#113)
- [bc069f48](https://github.com/kubedb/autoscaler/commit/bc069f48) Add MySQL Autoscaler support (#106)
- [88c985a0](https://github.com/kubedb/autoscaler/commit/88c985a0) Check for PDB version only once (#110)
- [6f5f9ae2](https://github.com/kubedb/autoscaler/commit/6f5f9ae2) Handle status conversion for CronJob/VolumeSnapshot (#109)
- [46b925c0](https://github.com/kubedb/autoscaler/commit/46b925c0) Use Go 1.19 (#108)
- [674e3b7a](https://github.com/kubedb/autoscaler/commit/674e3b7a) Use k8s 1.25.1 libs (#107)
- [6a5d4274](https://github.com/kubedb/autoscaler/commit/6a5d4274) Improve internal API; using milliValue (#105)
- [757cdfed](https://github.com/kubedb/autoscaler/commit/757cdfed) Add support for RedisSentinel autoscaler (#104)
- [56b92c66](https://github.com/kubedb/autoscaler/commit/56b92c66) Update README.md
- [f3b9904f](https://github.com/kubedb/autoscaler/commit/f3b9904f) Add PerconaXtraDB Autoscaler Support (#103)
- [7ac495d9](https://github.com/kubedb/autoscaler/commit/7ac495d9) Implement redisAutoscaler feature (#102)
- [997180f5](https://github.com/kubedb/autoscaler/commit/997180f5) Stop using removed apis in Kubernetes 1.25 (#101)
- [490a6b69](https://github.com/kubedb/autoscaler/commit/490a6b69) Use health checker types from kmodules (#100)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.29.0](https://github.com/kubedb/cli/releases/tag/v0.29.0)

- [64ed984d](https://github.com/kubedb/cli/commit/64ed984d) Prepare for release v0.29.0 (#686)
- [a3228690](https://github.com/kubedb/cli/commit/a3228690) Use password-generator@v0.2.9 (#685)
- [8033e31b](https://github.com/kubedb/cli/commit/8033e31b) Prepare for release v0.29.0-rc.0 (#684)
- [b021f761](https://github.com/kubedb/cli/commit/b021f761) Update dependencies (#683)
- [792efd14](https://github.com/kubedb/cli/commit/792efd14) Support externally managed secrets (#681)
- [7ec2adbc](https://github.com/kubedb/cli/commit/7ec2adbc) Test against Kubernetes 1.25.0 (#682)
- [fc9b63c7](https://github.com/kubedb/cli/commit/fc9b63c7) Check for PDB version only once (#680)
- [81199060](https://github.com/kubedb/cli/commit/81199060) Handle status conversion for CronJob/VolumeSnapshot (#679)
- [17c6e94d](https://github.com/kubedb/cli/commit/17c6e94d) Use Go 1.19 (#678)
- [31c24f80](https://github.com/kubedb/cli/commit/31c24f80) Use k8s 1.25.1 libs (#677)
- [68e9ada6](https://github.com/kubedb/cli/commit/68e9ada6) Update README.md
- [4202bc84](https://github.com/kubedb/cli/commit/4202bc84) Stop using removed apis in Kubernetes 1.25 (#676)
- [eb922b19](https://github.com/kubedb/cli/commit/eb922b19) Use health checker types from kmodules (#675)



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.5.0](https://github.com/kubedb/dashboard/releases/tag/v0.5.0)

- [903e551](https://github.com/kubedb/dashboard/commit/903e551) Prepare for release v0.5.0 (#48)
- [b06a4cf](https://github.com/kubedb/dashboard/commit/b06a4cf) Use password-generator@v0.2.9 (#47)
- [fd8f1bc](https://github.com/kubedb/dashboard/commit/fd8f1bc) Prepare for release v0.5.0-rc.0 (#46)
- [4b093a9](https://github.com/kubedb/dashboard/commit/4b093a9) Update dependencies (#45)
- [9804a55](https://github.com/kubedb/dashboard/commit/9804a55) Test against Kubernetes 1.25.0 (#44)
- [5f9caec](https://github.com/kubedb/dashboard/commit/5f9caec) Check for PDB version only once (#42)
- [91b256c](https://github.com/kubedb/dashboard/commit/91b256c) Handle status conversion for CronJob/VolumeSnapshot (#41)
- [11445c2](https://github.com/kubedb/dashboard/commit/11445c2) Use Go 1.19 (#40)
- [858bced](https://github.com/kubedb/dashboard/commit/858bced) Use k8s 1.25.1 libs (#39)
- [ebaaade](https://github.com/kubedb/dashboard/commit/ebaaade) Stop using removed apis in Kubernetes 1.25 (#38)
- [51d4f7f](https://github.com/kubedb/dashboard/commit/51d4f7f) Use health checker types from kmodules (#37)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.29.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.29.0)

- [fe996350](https://github.com/kubedb/elasticsearch/commit/fe9963505) Prepare for release v0.29.0 (#613)
- [13a36665](https://github.com/kubedb/elasticsearch/commit/13a36665d) Use password-generator@v0.2.9 (#612)
- [1e715c1a](https://github.com/kubedb/elasticsearch/commit/1e715c1a9) Prepare for release v0.29.0-rc.0 (#611)
- [4ab0de97](https://github.com/kubedb/elasticsearch/commit/4ab0de973) Update dependencies (#610)
- [1803a407](https://github.com/kubedb/elasticsearch/commit/1803a4078) Add support for Externally Managed secret (#609)
- [c2fb96e2](https://github.com/kubedb/elasticsearch/commit/c2fb96e2b) Test against Kubernetes 1.25.0 (#608)
- [96bbc6a8](https://github.com/kubedb/elasticsearch/commit/96bbc6a85) Check for PDB version only once (#606)
- [38099062](https://github.com/kubedb/elasticsearch/commit/380990623) Handle status conversion for CronJob/VolumeSnapshot (#605)
- [6e86f853](https://github.com/kubedb/elasticsearch/commit/6e86f853a) Use Go 1.19 (#604)
- [838ab6ae](https://github.com/kubedb/elasticsearch/commit/838ab6aec) Use k8s 1.25.1 libs (#603)
- [ce6877b5](https://github.com/kubedb/elasticsearch/commit/ce6877b58) Update README.md
- [297c6004](https://github.com/kubedb/elasticsearch/commit/297c60040) Stop using removed apis in Kubernetes 1.25 (#602)
- [7f9ef6bf](https://github.com/kubedb/elasticsearch/commit/7f9ef6bf1) Use health checker types from kmodules (#601)
- [baf9b9c1](https://github.com/kubedb/elasticsearch/commit/baf9b9c1b) Fix ClientCreated counter increment issue in healthchecker (#600)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2022.10.18](https://github.com/kubedb/installer/releases/tag/v2022.10.18)

- [333e9724](https://github.com/kubedb/installer/commit/333e9724) Prepare for release v2022.10.18 (#556)
- [e8a05842](https://github.com/kubedb/installer/commit/e8a05842) Update crds for kubedb/apimachinery@daeafa99 (#555)
- [1bb1d84b](https://github.com/kubedb/installer/commit/1bb1d84b) Update metricsconfig crd
- [46c2b8f2](https://github.com/kubedb/installer/commit/46c2b8f2) Add postgres crds to charts
- [0d5ff889](https://github.com/kubedb/installer/commit/0d5ff889) Prepare for release v2022.10.12-rc.0 (#553)
- [f0410def](https://github.com/kubedb/installer/commit/f0410def) Fix backend name for ProxySQL (#554)
- [545d7326](https://github.com/kubedb/installer/commit/545d7326) Add support for mysql 8.0.31 (#552)
- [0ecf6b7a](https://github.com/kubedb/installer/commit/0ecf6b7a) Update crds
- [3c80588f](https://github.com/kubedb/installer/commit/3c80588f) Add ProxySQL-2.3.2-debian/centos-v2 (#549)
- [edb50a92](https://github.com/kubedb/installer/commit/edb50a92) Add ProxySQL MetricsConfiguration (#545)
- [78961127](https://github.com/kubedb/installer/commit/78961127) Update Redis Init Container Image (#551)
- [e266fe95](https://github.com/kubedb/installer/commit/e266fe95) Update Percona XtraDB init container image (#550)
- [c2b9f93b](https://github.com/kubedb/installer/commit/c2b9f93b) Update mongodb init container image (#548)
- [cb4d226a](https://github.com/kubedb/installer/commit/cb4d226a) Add Redis Sentinel Ops Requests changes (#533)
- [f970eac3](https://github.com/kubedb/installer/commit/f970eac3) Fix missing docker images (#547)
- [d34e3363](https://github.com/kubedb/installer/commit/d34e3363) Add mutating webhook for postgresAutoscaler (#544)
- [bb0ae0de](https://github.com/kubedb/installer/commit/bb0ae0de) Fix valuePath for app_namespace key (#546)
- [862d034e](https://github.com/kubedb/installer/commit/862d034e) Add Subscriber apiservice (#543)
- [88d1225e](https://github.com/kubedb/installer/commit/88d1225e) Use k8s 1.25 client libs (#228)
- [46641e26](https://github.com/kubedb/installer/commit/46641e26) Add proxysql new version 2.4.4 (#539)
- [f498f5ae](https://github.com/kubedb/installer/commit/f498f5ae) Add Percona XtraDB 8.0.28 (#529)
- [24519580](https://github.com/kubedb/installer/commit/24519580) Add PerconaXtraDB Metrics (#532)
- [a1f8ac75](https://github.com/kubedb/installer/commit/a1f8ac75) Update crds (#541)
- [4b500533](https://github.com/kubedb/installer/commit/4b500533) Add Postgres Logical Replication rbac and validators (#534)
- [753f60c4](https://github.com/kubedb/installer/commit/753f60c4) Use k8s 1.25.2
- [700dacb1](https://github.com/kubedb/installer/commit/700dacb1) Test against Kubernetes 1.25.0 (#540)
- [92069bc7](https://github.com/kubedb/installer/commit/92069bc7) Test against k8s 1.25.0 (#537)
- [b944c0b0](https://github.com/kubedb/installer/commit/b944c0b0) Don't create PSP object in k8s >= 1.25 (#536)
- [d35f8aec](https://github.com/kubedb/installer/commit/d35f8aec) Use Go 1.19 (#535)
- [59a10600](https://github.com/kubedb/installer/commit/59a10600) Add all db-types in autoscaler mutatingwebhookConfiguration (#531)
- [3010e3e4](https://github.com/kubedb/installer/commit/3010e3e4) Update README.md
- [2763ae75](https://github.com/kubedb/installer/commit/2763ae75) Add exclusion for health index in Elasticsearch (#530)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.13.0](https://github.com/kubedb/mariadb/releases/tag/v0.13.0)

- [c9d28b74](https://github.com/kubedb/mariadb/commit/c9d28b74) Prepare for release v0.13.0 (#182)
- [132af6e6](https://github.com/kubedb/mariadb/commit/132af6e6) Use password-generator@v0.2.9 (#181)
- [ae53b273](https://github.com/kubedb/mariadb/commit/ae53b273) Not wait for Exporter Config Secret (#180)
- [b13d62cf](https://github.com/kubedb/mariadb/commit/b13d62cf) Prepare for release v0.13.0-rc.0 (#179)
- [5a8b0877](https://github.com/kubedb/mariadb/commit/5a8b0877) Add TLS Secret on Appbinding (#178)
- [a7f976f6](https://github.com/kubedb/mariadb/commit/a7f976f6) Add AppRef on AppBinding and Add Exporter Config Secret (#177)
- [a3d17697](https://github.com/kubedb/mariadb/commit/a3d17697) Update dependencies (#176)
- [8c666da8](https://github.com/kubedb/mariadb/commit/8c666da8) Add Externally Manage Secret Support (#175)
- [b14391f2](https://github.com/kubedb/mariadb/commit/b14391f2) Test against Kubernetes 1.25.0 (#174)
- [a07bbf68](https://github.com/kubedb/mariadb/commit/a07bbf68) Check for PDB version only once (#172)
- [8a316b93](https://github.com/kubedb/mariadb/commit/8a316b93) Handle status conversion for CronJob/VolumeSnapshot (#171)
- [56b6cd33](https://github.com/kubedb/mariadb/commit/56b6cd33) Use Go 1.19 (#170)
- [c666db48](https://github.com/kubedb/mariadb/commit/c666db48) Use k8s 1.25.1 libs (#169)
- [665d7f2a](https://github.com/kubedb/mariadb/commit/665d7f2a) Fix health check issue (#166)
- [c089e057](https://github.com/kubedb/mariadb/commit/c089e057) Update README.md
- [c0efefeb](https://github.com/kubedb/mariadb/commit/c0efefeb) Stop using removed apis in Kubernetes 1.25 (#168)
- [e3ef008e](https://github.com/kubedb/mariadb/commit/e3ef008e) Use health checker types from kmodules (#167)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.9.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.9.0)

- [923c65e](https://github.com/kubedb/mariadb-coordinator/commit/923c65e) Prepare for release v0.9.0 (#63)
- [b16d49d](https://github.com/kubedb/mariadb-coordinator/commit/b16d49d) Prepare for release v0.9.0-rc.0 (#62)
- [a5e1a7b](https://github.com/kubedb/mariadb-coordinator/commit/a5e1a7b) Update dependencies (#61)
- [119956c](https://github.com/kubedb/mariadb-coordinator/commit/119956c) Test against Kubernetes 1.25.0 (#60)
- [4950880](https://github.com/kubedb/mariadb-coordinator/commit/4950880) Check for PDB version only once (#58)
- [8e89509](https://github.com/kubedb/mariadb-coordinator/commit/8e89509) Handle status conversion for CronJob/VolumeSnapshot (#57)
- [79dc72c](https://github.com/kubedb/mariadb-coordinator/commit/79dc72c) Use Go 1.19 (#56)
- [5a57951](https://github.com/kubedb/mariadb-coordinator/commit/5a57951) Use k8s 1.25.1 libs (#55)
- [101e71a](https://github.com/kubedb/mariadb-coordinator/commit/101e71a) Stop using removed apis in Kubernetes 1.25 (#54)
- [61c60ed](https://github.com/kubedb/mariadb-coordinator/commit/61c60ed) Use health checker types from kmodules (#53)
- [fe8a57e](https://github.com/kubedb/mariadb-coordinator/commit/fe8a57e) Prepare for release v0.8.0-rc.1 (#52)
- [9c3c47f](https://github.com/kubedb/mariadb-coordinator/commit/9c3c47f) Update health checker (#51)
- [82bad04](https://github.com/kubedb/mariadb-coordinator/commit/82bad04) Prepare for release v0.8.0-rc.0 (#50)
- [487fdbb](https://github.com/kubedb/mariadb-coordinator/commit/487fdbb) Acquire license from license-proxyserver if available (#49)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.22.0](https://github.com/kubedb/memcached/releases/tag/v0.22.0)

- [05b13b2f](https://github.com/kubedb/memcached/commit/05b13b2f) Prepare for release v0.22.0 (#375)
- [9849bbb8](https://github.com/kubedb/memcached/commit/9849bbb8) Use password-generator@v0.2.9 (#374)
- [255abab3](https://github.com/kubedb/memcached/commit/255abab3) Prepare for release v0.22.0-rc.0 (#373)
- [2cbc373f](https://github.com/kubedb/memcached/commit/2cbc373f) Update dependencies (#372)
- [6995e546](https://github.com/kubedb/memcached/commit/6995e546) Test against Kubernetes 1.25.0 (#371)
- [2974948e](https://github.com/kubedb/memcached/commit/2974948e) Check for PDB version only once (#369)
- [f2662305](https://github.com/kubedb/memcached/commit/f2662305) Handle status conversion for CronJob/VolumeSnapshot (#368)
- [a79d8ed9](https://github.com/kubedb/memcached/commit/a79d8ed9) Use Go 1.19 (#367)
- [e2a89736](https://github.com/kubedb/memcached/commit/e2a89736) Use k8s 1.25.1 libs (#366)
- [15ba567f](https://github.com/kubedb/memcached/commit/15ba567f) Stop using removed apis in Kubernetes 1.25 (#365)
- [12204d85](https://github.com/kubedb/memcached/commit/12204d85) Use health checker types from kmodules (#364)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.22.0](https://github.com/kubedb/mongodb/releases/tag/v0.22.0)

- [fe689daa](https://github.com/kubedb/mongodb/commit/fe689daa) Prepare for release v0.22.0 (#519)
- [e6ce6f0e](https://github.com/kubedb/mongodb/commit/e6ce6f0e) Use password-generator@v0.2.9 (#518)
- [b9e03cc5](https://github.com/kubedb/mongodb/commit/b9e03cc5) Prepare for release v0.22.0-rc.0 (#517)
- [2f0c8b65](https://github.com/kubedb/mongodb/commit/2f0c8b65) Set TLSSecret name (#516)
- [ffb021ea](https://github.com/kubedb/mongodb/commit/ffb021ea) Configure AppRef in appbinding (#515)
- [2c9eb87b](https://github.com/kubedb/mongodb/commit/2c9eb87b) Add support for externally-managed authSecret (#514)
- [f4789ab7](https://github.com/kubedb/mongodb/commit/f4789ab7) Test against Kubernetes 1.25.0 (#513)
- [9ad4c219](https://github.com/kubedb/mongodb/commit/9ad4c219) Change operator name in event (#511)
- [dbb7ff10](https://github.com/kubedb/mongodb/commit/dbb7ff10) Check for PDB version only once (#510)
- [79d53b0a](https://github.com/kubedb/mongodb/commit/79d53b0a) Handle status conversion for CronJob/VolumeSnapshot (#509)
- [37521202](https://github.com/kubedb/mongodb/commit/37521202) Use Go 1.19 (#508)
- [d1a2d55a](https://github.com/kubedb/mongodb/commit/d1a2d55a) Use k8s 1.25.1 libs (#507)
- [43399906](https://github.com/kubedb/mongodb/commit/43399906) Add support for Hidden node (#503)
- [91acbffc](https://github.com/kubedb/mongodb/commit/91acbffc) Update README.md
- [b053290c](https://github.com/kubedb/mongodb/commit/b053290c) Stop using removed apis in Kubernetes 1.25 (#506)
- [79b99580](https://github.com/kubedb/mongodb/commit/79b99580) Use health checker types from kmodules (#505)
- [ff39883d](https://github.com/kubedb/mongodb/commit/ff39883d) Fix health check issue (#504)



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.22.0](https://github.com/kubedb/mysql/releases/tag/v0.22.0)

- [2dc839c0](https://github.com/kubedb/mysql/commit/2dc839c0) Prepare for release v0.22.0 (#507)
- [1c55ac97](https://github.com/kubedb/mysql/commit/1c55ac97) Add return statement if client engine is not created (#506)
- [4d402bf3](https://github.com/kubedb/mysql/commit/4d402bf3) Use password-generator@v0.2.9 (#505)
- [8386ed4c](https://github.com/kubedb/mysql/commit/8386ed4c) Prepare for release v0.22.0-rc.0 (#504)
- [8d58bbd8](https://github.com/kubedb/mysql/commit/8d58bbd8) Add cluster role for watching mysqlversion in coordinator (#503)
- [53b207b3](https://github.com/kubedb/mysql/commit/53b207b3) Add TLS Secret Name in appbinding (#501)
- [541e9f5e](https://github.com/kubedb/mysql/commit/541e9f5e) Update dependencies (#502)
- [e51a494c](https://github.com/kubedb/mysql/commit/e51a494c) Fix innodb router issues (#500)
- [c4f78c1f](https://github.com/kubedb/mysql/commit/c4f78c1f) Wait for externaly managed auth secret (#499)
- [90f337a2](https://github.com/kubedb/mysql/commit/90f337a2) Test against Kubernetes 1.25.0 (#498)
- [af6d6654](https://github.com/kubedb/mysql/commit/af6d6654) Check for PDB version only once (#496)
- [27611133](https://github.com/kubedb/mysql/commit/27611133) Handle status conversion for CronJob/VolumeSnapshot (#495)
- [a662b10d](https://github.com/kubedb/mysql/commit/a662b10d) Use Go 1.19 (#494)
- [07ce8211](https://github.com/kubedb/mysql/commit/07ce8211) Use k8s 1.25.1 libs (#493)
- [fac38c31](https://github.com/kubedb/mysql/commit/fac38c31) Update README.md
- [9676f388](https://github.com/kubedb/mysql/commit/9676f388) Stop using removed apis in Kubernetes 1.25 (#492)
- [db176142](https://github.com/kubedb/mysql/commit/db176142) Use health checker types from kmodules (#491)
- [3c9835b0](https://github.com/kubedb/mysql/commit/3c9835b0) Fix health check issue (#489)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.7.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.7.0)

- [650bb92](https://github.com/kubedb/mysql-coordinator/commit/650bb92) Prepare for release v0.7.0 (#59)
- [0930361](https://github.com/kubedb/mysql-coordinator/commit/0930361) Add version check for MySQL 5.0 (#57)
- [b1d9ecf](https://github.com/kubedb/mysql-coordinator/commit/b1d9ecf) Prepare for release v0.7.0-rc.0 (#58)
- [88d01ef](https://github.com/kubedb/mysql-coordinator/commit/88d01ef) Update dependencies (#56)
- [cbb3504](https://github.com/kubedb/mysql-coordinator/commit/cbb3504) fix group_replication extra transcions jonning issue (#49)
- [8939e89](https://github.com/kubedb/mysql-coordinator/commit/8939e89) Test against Kubernetes 1.25.0 (#55)
- [0ba243d](https://github.com/kubedb/mysql-coordinator/commit/0ba243d) Check for PDB version only once (#53)
- [dac7227](https://github.com/kubedb/mysql-coordinator/commit/dac7227) Handle status conversion for CronJob/VolumeSnapshot (#52)
- [100f268](https://github.com/kubedb/mysql-coordinator/commit/100f268) Use Go 1.19 (#51)
- [07fc1af](https://github.com/kubedb/mysql-coordinator/commit/07fc1af) Use k8s 1.25.1 libs (#50)
- [71fe729](https://github.com/kubedb/mysql-coordinator/commit/71fe729) Stop using removed apis in Kubernetes 1.25 (#48)
- [f968206](https://github.com/kubedb/mysql-coordinator/commit/f968206) Use health checker types from kmodules (#47)
- [fa7ad1c](https://github.com/kubedb/mysql-coordinator/commit/fa7ad1c) Prepare for release v0.6.0-rc.1 (#46)
- [2c3615b](https://github.com/kubedb/mysql-coordinator/commit/2c3615b) update labels (#45)
- [38a4f88](https://github.com/kubedb/mysql-coordinator/commit/38a4f88) Update health checker (#43)
- [7c79e5f](https://github.com/kubedb/mysql-coordinator/commit/7c79e5f) Prepare for release v0.6.0-rc.0 (#42)
- [2eb313d](https://github.com/kubedb/mysql-coordinator/commit/2eb313d) Acquire license from license-proxyserver if available (#40)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.7.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.7.0)

- [e5eba9e](https://github.com/kubedb/mysql-router-init/commit/e5eba9e) Test against Kubernetes 1.25.0 (#26)
- [f0bdfdd](https://github.com/kubedb/mysql-router-init/commit/f0bdfdd) Use Go 1.19 (#25)
- [5631a3c](https://github.com/kubedb/mysql-router-init/commit/5631a3c) Use k8s 1.25.1 libs (#24)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.16.0](https://github.com/kubedb/ops-manager/releases/tag/v0.16.0)

- [121610ce](https://github.com/kubedb/ops-manager/commit/121610ce) Prepare for release v0.16.0 (#377)
- [b7f7b559](https://github.com/kubedb/ops-manager/commit/b7f7b559) Fix remove TLS and version upgrade for mysql (#375)
- [4a527621](https://github.com/kubedb/ops-manager/commit/4a527621) Fix exporter config secret cleanup for MariaDB and Percona XtraDB (#376)
- [2df29e1d](https://github.com/kubedb/ops-manager/commit/2df29e1d) Use password-generator@v0.2.9 (#374)
- [698d4f84](https://github.com/kubedb/ops-manager/commit/698d4f84) Prepare for release v0.16.0-rc.0 (#373)
- [f85a6048](https://github.com/kubedb/ops-manager/commit/f85a6048) Handle private registry with self-signed certs (#372)
- [205b8e3c](https://github.com/kubedb/ops-manager/commit/205b8e3c) Fix replication user update password (#371)
- [6680a32c](https://github.com/kubedb/ops-manager/commit/6680a32c) Fix HS Ops Request (#370)
- [b96f4592](https://github.com/kubedb/ops-manager/commit/b96f4592) Add PostgreSQL Logical Replication (#353)
- [2ca9b5f8](https://github.com/kubedb/ops-manager/commit/2ca9b5f8) ProxySQL Ops-requests (#368)
- [c2d6b85e](https://github.com/kubedb/ops-manager/commit/c2d6b85e) Remove ensureExporterSecretForTLSConfig for MariaDB and PXC (#369)
- [06b69609](https://github.com/kubedb/ops-manager/commit/06b69609) Add PerconaXtraDB OpsReq (#367)
- [891a2288](https://github.com/kubedb/ops-manager/commit/891a2288) Make opsReqType specific to databases (#366)
- [82d960b0](https://github.com/kubedb/ops-manager/commit/82d960b0) MySQL ops request fix for Innodb (#365)
- [13401a96](https://github.com/kubedb/ops-manager/commit/13401a96) Add Redis Sentinel Ops Request (#328)
- [8ee68b62](https://github.com/kubedb/ops-manager/commit/8ee68b62) Modify reconfigureTLS to support arbiter & hidden enabled mongo (#364)
- [805f8bba](https://github.com/kubedb/ops-manager/commit/805f8bba) Test against Kubernetes 1.25.0 (#363)
- [787f7bea](https://github.com/kubedb/ops-manager/commit/787f7bea) Fix MariaDB Upgrade OpsReq Image name issue (#361)
- [e676ea51](https://github.com/kubedb/ops-manager/commit/e676ea51) Fix podnames & selectors for Mongo volumeExpansion (#358)
- [7a5e34b1](https://github.com/kubedb/ops-manager/commit/7a5e34b1) Check for PDB version only once (#357)
- [3fb148a5](https://github.com/kubedb/ops-manager/commit/3fb148a5) Handle status conversion for CronJob/VolumeSnapshot (#356)
- [9f058091](https://github.com/kubedb/ops-manager/commit/9f058091) Use Go 1.19 (#355)
- [25febfcb](https://github.com/kubedb/ops-manager/commit/25febfcb) Update .kodiak.toml
- [eb0f3792](https://github.com/kubedb/ops-manager/commit/eb0f3792) Use k8s 1.25.1 libs (#354)
- [d09da904](https://github.com/kubedb/ops-manager/commit/d09da904) Add opsRequests for mongo hidden-node (#347)
- [9114f329](https://github.com/kubedb/ops-manager/commit/9114f329) Rework Mongo verticalScaling; Fix arbiter & exporter-related issues (#346)
- [74f4831c](https://github.com/kubedb/ops-manager/commit/74f4831c) Update README.md
- [2277c28f](https://github.com/kubedb/ops-manager/commit/2277c28f) Skip Image Digest for Dev Builds (#350)
- [c7f3cf07](https://github.com/kubedb/ops-manager/commit/c7f3cf07) Stop using removed apis in Kubernetes 1.25 (#352)
- [0fbc4d57](https://github.com/kubedb/ops-manager/commit/0fbc4d57) Use health checker types from kmodules (#351)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.16.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.16.0)

- [78c63bf7](https://github.com/kubedb/percona-xtradb/commit/78c63bf7) Prepare for release v0.16.0 (#285)
- [d47bd015](https://github.com/kubedb/percona-xtradb/commit/d47bd015) Use password-generator@v0.2.9 (#284)
- [f3278c09](https://github.com/kubedb/percona-xtradb/commit/f3278c09) Not wait for Exporter Config Secret (#283)
- [030e063d](https://github.com/kubedb/percona-xtradb/commit/030e063d) Prepare for release v0.16.0-rc.0 (#282)
- [47345de1](https://github.com/kubedb/percona-xtradb/commit/47345de1) Add TLS Secret on AppBinding (#281)
- [0aa33548](https://github.com/kubedb/percona-xtradb/commit/0aa33548) Add AppRef on AppBinding and Add Exporter Config Secret (#280)
- [82685157](https://github.com/kubedb/percona-xtradb/commit/82685157) Merge pull request #269 from kubedb/add-px-ops
- [f7f1898e](https://github.com/kubedb/percona-xtradb/commit/f7f1898e) Add Externally Managed Secret Support on PerconaXtraDB
- [43dcc76d](https://github.com/kubedb/percona-xtradb/commit/43dcc76d) Test against Kubernetes 1.25.0 (#278)
- [bc5c97db](https://github.com/kubedb/percona-xtradb/commit/bc5c97db) Check for PDB version only once (#276)
- [13a57a32](https://github.com/kubedb/percona-xtradb/commit/13a57a32) Handle status conversion for CronJob/VolumeSnapshot (#275)
- [6013a92e](https://github.com/kubedb/percona-xtradb/commit/6013a92e) Use Go 1.19 (#274)
- [45c413b9](https://github.com/kubedb/percona-xtradb/commit/45c413b9) Use k8s 1.25.1 libs (#273)
- [fd7d238a](https://github.com/kubedb/percona-xtradb/commit/fd7d238a) Update README.md
- [13da58d6](https://github.com/kubedb/percona-xtradb/commit/13da58d6) Stop using removed apis in Kubernetes 1.25 (#272)
- [6941e6d6](https://github.com/kubedb/percona-xtradb/commit/6941e6d6) Use health checker types from kmodules (#271)
- [9f813287](https://github.com/kubedb/percona-xtradb/commit/9f813287) Fix health check issue (#270)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.2.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.2.0)

- [9a28a21](https://github.com/kubedb/percona-xtradb-coordinator/commit/9a28a21) Prepare for release v0.2.0 (#20)
- [cf6c54c](https://github.com/kubedb/percona-xtradb-coordinator/commit/cf6c54c) Prepare for release v0.2.0-rc.0 (#19)
- [a71e01d](https://github.com/kubedb/percona-xtradb-coordinator/commit/a71e01d) Update dependencies (#18)
- [0b51751](https://github.com/kubedb/percona-xtradb-coordinator/commit/0b51751) Test against Kubernetes 1.25.0 (#17)
- [1f2b1a5](https://github.com/kubedb/percona-xtradb-coordinator/commit/1f2b1a5) Check for PDB version only once (#15)
- [03125ba](https://github.com/kubedb/percona-xtradb-coordinator/commit/03125ba) Handle status conversion for CronJob/VolumeSnapshot (#14)
- [06a2634](https://github.com/kubedb/percona-xtradb-coordinator/commit/06a2634) Use Go 1.19 (#13)
- [1a8a90b](https://github.com/kubedb/percona-xtradb-coordinator/commit/1a8a90b) Use k8s 1.25.1 libs (#12)
- [f33c751](https://github.com/kubedb/percona-xtradb-coordinator/commit/f33c751) Stop using removed apis in Kubernetes 1.25 (#11)
- [91495bf](https://github.com/kubedb/percona-xtradb-coordinator/commit/91495bf) Use health checker types from kmodules (#10)
- [290e281](https://github.com/kubedb/percona-xtradb-coordinator/commit/290e281) Prepare for release v0.1.0-rc.1 (#9)
- [c57449c](https://github.com/kubedb/percona-xtradb-coordinator/commit/c57449c) Update health checker (#8)
- [adad8b5](https://github.com/kubedb/percona-xtradb-coordinator/commit/adad8b5) Acquire license from license-proxyserver if available (#7)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.13.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.13.0)

- [43cd452f](https://github.com/kubedb/pg-coordinator/commit/43cd452f) Prepare for release v0.13.0 (#100)
- [85fb61bb](https://github.com/kubedb/pg-coordinator/commit/85fb61bb) Prepare for release v0.13.0-rc.0 (#99)
- [58720b10](https://github.com/kubedb/pg-coordinator/commit/58720b10) Update dependencies (#98)
- [5a9dcc5f](https://github.com/kubedb/pg-coordinator/commit/5a9dcc5f) Test against Kubernetes 1.25.0 (#97)
- [eb45fd8e](https://github.com/kubedb/pg-coordinator/commit/eb45fd8e) Check for PDB version only once (#95)
- [a66884fb](https://github.com/kubedb/pg-coordinator/commit/a66884fb) Handle status conversion for CronJob/VolumeSnapshot (#94)
- [db150c63](https://github.com/kubedb/pg-coordinator/commit/db150c63) Use Go 1.19 (#93)
- [8bd4fcc5](https://github.com/kubedb/pg-coordinator/commit/8bd4fcc5) Use k8s 1.25.1 libs (#92)
- [4a510768](https://github.com/kubedb/pg-coordinator/commit/4a510768) Stop using removed apis in Kubernetes 1.25 (#91)
- [3b26263c](https://github.com/kubedb/pg-coordinator/commit/3b26263c) Use health checker types from kmodules (#90)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.16.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.16.0)

- [4ba55b3e](https://github.com/kubedb/pgbouncer/commit/4ba55b3e) Prepare for release v0.16.0 (#247)
- [a4e978e9](https://github.com/kubedb/pgbouncer/commit/a4e978e9) Use password-generator@v0.2.9 (#246)
- [0d58567a](https://github.com/kubedb/pgbouncer/commit/0d58567a) Prepare for release v0.16.0-rc.0 (#245)
- [47329dfa](https://github.com/kubedb/pgbouncer/commit/47329dfa) Fix TLSSecret for appbinding. (#244)
- [3efec0cb](https://github.com/kubedb/pgbouncer/commit/3efec0cb) Update dependencies (#243)
- [8a1bd7b0](https://github.com/kubedb/pgbouncer/commit/8a1bd7b0) Fix health check issue (#234)
- [c20e87e5](https://github.com/kubedb/pgbouncer/commit/c20e87e5) Test against Kubernetes 1.25.0 (#242)
- [760fd8e3](https://github.com/kubedb/pgbouncer/commit/760fd8e3) Check for PDB version only once (#240)
- [8ba2692d](https://github.com/kubedb/pgbouncer/commit/8ba2692d) Handle status conversion for CronJob/VolumeSnapshot (#239)
- [ea1fc328](https://github.com/kubedb/pgbouncer/commit/ea1fc328) Use Go 1.19 (#238)
- [6a24f732](https://github.com/kubedb/pgbouncer/commit/6a24f732) Use k8s 1.25.1 libs (#237)
- [327242e1](https://github.com/kubedb/pgbouncer/commit/327242e1) Update README.md
- [c9754ecd](https://github.com/kubedb/pgbouncer/commit/c9754ecd) Stop using removed apis in Kubernetes 1.25 (#236)
- [bb7a3b6f](https://github.com/kubedb/pgbouncer/commit/bb7a3b6f) Use health checker types from kmodules (#235)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.29.0](https://github.com/kubedb/postgres/releases/tag/v0.29.0)

- [08afda20](https://github.com/kubedb/postgres/commit/08afda20) Prepare for release v0.29.0 (#607)
- [ad39a6cf](https://github.com/kubedb/postgres/commit/ad39a6cf) Update password generator hash (#606)
- [cd547a68](https://github.com/kubedb/postgres/commit/cd547a68) Prepare for release v0.29.0-rc.0 (#605)
- [9d98af14](https://github.com/kubedb/postgres/commit/9d98af14) Fix TlsSecret for AppBinding (#604)
- [7d73ce99](https://github.com/kubedb/postgres/commit/7d73ce99) Update dependencies (#603)
- [d8515a3f](https://github.com/kubedb/postgres/commit/d8515a3f) Configure appRef in AppBinding (#602)
- [69458e25](https://github.com/kubedb/postgres/commit/69458e25) Check auth secrets labels if key exists (#601)
- [3dd3563b](https://github.com/kubedb/postgres/commit/3dd3563b) Simplify ensureAuthSecret (#600)
- [67f3db64](https://github.com/kubedb/postgres/commit/67f3db64) Relax Postgres key detection for a secret (#599)
- [acdd2cda](https://github.com/kubedb/postgres/commit/acdd2cda) Add support for Externally Managed secret (#597)
- [5121a362](https://github.com/kubedb/postgres/commit/5121a362) Test against Kubernetes 1.25.0 (#598)
- [bfa46b08](https://github.com/kubedb/postgres/commit/bfa46b08) Check for PDB version only once (#594)
- [150fcf2c](https://github.com/kubedb/postgres/commit/150fcf2c) Handle status conversion for CronJob/VolumeSnapshot (#593)
- [86ff76e1](https://github.com/kubedb/postgres/commit/86ff76e1) Use Go 1.19 (#592)
- [7732e22b](https://github.com/kubedb/postgres/commit/7732e22b) Use k8s 1.25.1 libs (#591)
- [b4c7f426](https://github.com/kubedb/postgres/commit/b4c7f426) Update README.md
- [68b06e68](https://github.com/kubedb/postgres/commit/68b06e68) Stop using removed apis in Kubernetes 1.25 (#590)
- [51f600b9](https://github.com/kubedb/postgres/commit/51f600b9) Use health checker types from kmodules (#589)
- [2e45ad1b](https://github.com/kubedb/postgres/commit/2e45ad1b) Fix health check issue (#588)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.29.0](https://github.com/kubedb/provisioner/releases/tag/v0.29.0)

- [b499c076](https://github.com/kubedb/provisioner/commit/b499c0764) Prepare for release v0.29.0 (#23)
- [5f54d2b0](https://github.com/kubedb/provisioner/commit/5f54d2b03) Use password-generator@v0.2.9 (#22)
- [0fdb3106](https://github.com/kubedb/provisioner/commit/0fdb31068) Not wait for Exporter Config Secret mariadb/xtradb
- [5497cc1b](https://github.com/kubedb/provisioner/commit/5497cc1be) Prepare for release v0.29.0-rc.0 (#21)
- [26b43352](https://github.com/kubedb/provisioner/commit/26b43352c) Test against Kubernetes 1.25.0 (#19)
- [597518ce](https://github.com/kubedb/provisioner/commit/597518cea) Check for PDB version only once (#17)
- [a55613f6](https://github.com/kubedb/provisioner/commit/a55613f6e) Handle status conversion for CronJob/VolumeSnapshot (#16)
- [5ef0c78e](https://github.com/kubedb/provisioner/commit/5ef0c78ee) Use Go 1.19 (#15)
- [40fe839c](https://github.com/kubedb/provisioner/commit/40fe839c8) Use k8s 1.25.1 libs (#14)
- [444e527c](https://github.com/kubedb/provisioner/commit/444e527ca) Update README.md
- [dc895331](https://github.com/kubedb/provisioner/commit/dc8953315) Stop using removed apis in Kubernetes 1.25 (#13)
- [2910a39e](https://github.com/kubedb/provisioner/commit/2910a39e2) Use health checker types from kmodules (#12)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.16.0](https://github.com/kubedb/proxysql/releases/tag/v0.16.0)

- [060e14f6](https://github.com/kubedb/proxysql/commit/060e14f6) Prepare for release v0.16.0 (#261)
- [7598aa0b](https://github.com/kubedb/proxysql/commit/7598aa0b) Use password-generator@v0.2.9 (#260)
- [3dc6618f](https://github.com/kubedb/proxysql/commit/3dc6618f) Prepare for release v0.16.0-rc.0 (#259)
- [ec249ccf](https://github.com/kubedb/proxysql/commit/ec249ccf) Add External-backend support and changes for Ops-requests (#258)
- [fe9d736a](https://github.com/kubedb/proxysql/commit/fe9d736a) Fix health check issue (#247)
- [42a3dedf](https://github.com/kubedb/proxysql/commit/42a3dedf) Test against Kubernetes 1.25.0 (#256)
- [4677b6ab](https://github.com/kubedb/proxysql/commit/4677b6ab) Check for PDB version only once (#254)
- [8f3e6e64](https://github.com/kubedb/proxysql/commit/8f3e6e64) Handle status conversion for CronJob/VolumeSnapshot (#253)
- [19b856f4](https://github.com/kubedb/proxysql/commit/19b856f4) Use Go 1.19 (#252)
- [f8dd8297](https://github.com/kubedb/proxysql/commit/f8dd8297) Use k8s 1.25.1 libs (#251)
- [3a21c93a](https://github.com/kubedb/proxysql/commit/3a21c93a) Stop using removed apis in Kubernetes 1.25 (#249)
- [cb0a1efd](https://github.com/kubedb/proxysql/commit/cb0a1efd) Use health checker types from kmodules (#248)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.22.0](https://github.com/kubedb/redis/releases/tag/v0.22.0)

- [a949cd65](https://github.com/kubedb/redis/commit/a949cd65) Prepare for release v0.22.0 (#434)
- [b6d1e6dc](https://github.com/kubedb/redis/commit/b6d1e6dc) Use password-generator@v0.2.9 (#433)
- [24a961a8](https://github.com/kubedb/redis/commit/24a961a8) Prepare for release v0.22.0-rc.0 (#432)
- [586d92c6](https://github.com/kubedb/redis/commit/586d92c6) Add Client Cert to Appbinding (#431)
- [9931e951](https://github.com/kubedb/redis/commit/9931e951) Update dependencies (#430)
- [5a27f772](https://github.com/kubedb/redis/commit/5a27f772) Add Redis Sentinel Ops Request Changes (#421)
- [81ad08ab](https://github.com/kubedb/redis/commit/81ad08ab) Add Support for Externally Managed Secret (#428)
- [b16212e4](https://github.com/kubedb/redis/commit/b16212e4) Test against Kubernetes 1.25.0 (#429)
- [05a1b814](https://github.com/kubedb/redis/commit/05a1b814) Check for PDB version only once (#427)
- [bd41d16d](https://github.com/kubedb/redis/commit/bd41d16d) Handle status conversion for CronJob/VolumeSnapshot (#426)
- [e1746638](https://github.com/kubedb/redis/commit/e1746638) Use Go 1.19 (#425)
- [b220f611](https://github.com/kubedb/redis/commit/b220f611) Use k8s 1.25.1 libs (#424)
- [538e2539](https://github.com/kubedb/redis/commit/538e2539) Update README.md
- [1513ca9a](https://github.com/kubedb/redis/commit/1513ca9a) Stop using removed apis in Kubernetes 1.25 (#423)
- [c29f0f6b](https://github.com/kubedb/redis/commit/c29f0f6b) Use health checker types from kmodules (#422)
- [bda4de79](https://github.com/kubedb/redis/commit/bda4de79) Fix health check issue (#420)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.8.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.8.0)

- [140202b](https://github.com/kubedb/redis-coordinator/commit/140202b) Prepare for release v0.8.0 (#53)
- [21d63ea](https://github.com/kubedb/redis-coordinator/commit/21d63ea) Prepare for release v0.8.0-rc.0 (#52)
- [d7bcff0](https://github.com/kubedb/redis-coordinator/commit/d7bcff0) Update dependencies (#51)
- [db31014](https://github.com/kubedb/redis-coordinator/commit/db31014) Add Redis Sentinel Ops Requests Changes (#48)
- [3bc6a63](https://github.com/kubedb/redis-coordinator/commit/3bc6a63) Test against Kubernetes 1.25.0 (#50)
- [b144d17](https://github.com/kubedb/redis-coordinator/commit/b144d17) Check for PDB version only once (#47)
- [803f76a](https://github.com/kubedb/redis-coordinator/commit/803f76a) Handle status conversion for CronJob/VolumeSnapshot (#46)
- [a7cd5af](https://github.com/kubedb/redis-coordinator/commit/a7cd5af) Use Go 1.19 (#45)
- [f066d36](https://github.com/kubedb/redis-coordinator/commit/f066d36) Use k8s 1.25.1 libs (#44)
- [db04c50](https://github.com/kubedb/redis-coordinator/commit/db04c50) Stop using removed apis in Kubernetes 1.25 (#43)
- [10f1fb5](https://github.com/kubedb/redis-coordinator/commit/10f1fb5) Use health checker types from kmodules (#42)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.16.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.16.0)

- [866018be](https://github.com/kubedb/replication-mode-detector/commit/866018be) Prepare for release v0.16.0 (#215)
- [d051a8eb](https://github.com/kubedb/replication-mode-detector/commit/d051a8eb) Prepare for release v0.16.0-rc.0 (#214)
- [2d51c3f3](https://github.com/kubedb/replication-mode-detector/commit/2d51c3f3) Update dependencies (#213)
- [0a544cf9](https://github.com/kubedb/replication-mode-detector/commit/0a544cf9) Test against Kubernetes 1.25.0 (#212)
- [aa1635cf](https://github.com/kubedb/replication-mode-detector/commit/aa1635cf) Check for PDB version only once (#210)
- [6549acf6](https://github.com/kubedb/replication-mode-detector/commit/6549acf6) Handle status conversion for CronJob/VolumeSnapshot (#209)
- [fc7a68fd](https://github.com/kubedb/replication-mode-detector/commit/fc7a68fd) Use Go 1.19 (#208)
- [2f9a7435](https://github.com/kubedb/replication-mode-detector/commit/2f9a7435) Use k8s 1.25.1 libs (#207)
- [c831c08e](https://github.com/kubedb/replication-mode-detector/commit/c831c08e) Stop using removed apis in Kubernetes 1.25 (#206)
- [8c80e5b4](https://github.com/kubedb/replication-mode-detector/commit/8c80e5b4) Use health checker types from kmodules (#205)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.5.0](https://github.com/kubedb/schema-manager/releases/tag/v0.5.0)

- [e6fffe23](https://github.com/kubedb/schema-manager/commit/e6fffe23) Prepare for release v0.5.0 (#52)
- [56931e13](https://github.com/kubedb/schema-manager/commit/56931e13) Prepare for release v0.5.0-rc.0 (#51)
- [7a97cbbd](https://github.com/kubedb/schema-manager/commit/7a97cbbd) Add documentation for PostgreSql (#30)
- [786c9ebf](https://github.com/kubedb/schema-manager/commit/786c9ebf) Make packages according to db-types (#49)
- [b708e23e](https://github.com/kubedb/schema-manager/commit/b708e23e) Update dependencies (#50)
- [78c6b620](https://github.com/kubedb/schema-manager/commit/78c6b620) Test against Kubernetes 1.25.0 (#48)
- [a150a60c](https://github.com/kubedb/schema-manager/commit/a150a60c) Check for PDB version only once (#46)
- [627daf35](https://github.com/kubedb/schema-manager/commit/627daf35) Handle status conversion for CronJob/VolumeSnapshot (#45)
- [1663dd03](https://github.com/kubedb/schema-manager/commit/1663dd03) Use Go 1.19 (#44)
- [417b5ebf](https://github.com/kubedb/schema-manager/commit/417b5ebf) Use k8s 1.25.1 libs (#43)
- [19488002](https://github.com/kubedb/schema-manager/commit/19488002) Stop using removed apis in Kubernetes 1.25 (#42)
- [f1af7213](https://github.com/kubedb/schema-manager/commit/f1af7213) Use health checker types from kmodules (#41)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.14.0](https://github.com/kubedb/tests/releases/tag/v0.14.0)

- [1737b25f](https://github.com/kubedb/tests/commit/1737b25f) Prepare for release v0.14.0 (#203)
- [f272557a](https://github.com/kubedb/tests/commit/f272557a) Add MongoDB arbiter-related tests (#172)
- [e7c55a30](https://github.com/kubedb/tests/commit/e7c55a30) Use password-generator@v0.2.9 (#201)
- [a2d4d3ac](https://github.com/kubedb/tests/commit/a2d4d3ac) Prepare for release v0.14.0-rc.0 (#200)
- [03a028e7](https://github.com/kubedb/tests/commit/03a028e7) Update dependencies (#197)
- [b34253e7](https://github.com/kubedb/tests/commit/b34253e7) Test against Kubernetes 1.25.0 (#196)
- [b2c48e72](https://github.com/kubedb/tests/commit/b2c48e72) Check for PDB version only once (#194)
- [bd0b7f66](https://github.com/kubedb/tests/commit/bd0b7f66) Handle status conversion for CronJob/VolumeSnapshot (#193)
- [8e5103d0](https://github.com/kubedb/tests/commit/8e5103d0) Use Go 1.19 (#192)
- [096cfbf6](https://github.com/kubedb/tests/commit/096cfbf6) Use k8s 1.25.1 libs (#191)
- [6c45ea94](https://github.com/kubedb/tests/commit/6c45ea94) Migrate to GinkgoV2 (#188)
- [f89ab1c1](https://github.com/kubedb/tests/commit/f89ab1c1) Stop using removed apis in Kubernetes 1.25 (#190)
- [17954e8b](https://github.com/kubedb/tests/commit/17954e8b) Use health checker types from kmodules (#189)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.5.0](https://github.com/kubedb/ui-server/releases/tag/v0.5.0)

- [d205fcec](https://github.com/kubedb/ui-server/commit/d205fcec) Prepare for release v0.5.0 (#55)
- [9dc8acb9](https://github.com/kubedb/ui-server/commit/9dc8acb9) Prepare for release v0.5.0-rc.0 (#54)
- [7ccfe3ed](https://github.com/kubedb/ui-server/commit/7ccfe3ed) Update dependencies (#53)
- [55f85699](https://github.com/kubedb/ui-server/commit/55f85699) Use Go 1.19 (#52)
- [19c39ab1](https://github.com/kubedb/ui-server/commit/19c39ab1) Check for PDB version only once (#50)
- [c1d7c41f](https://github.com/kubedb/ui-server/commit/c1d7c41f) Handle status conversion for CronJob/VolumeSnapshot (#49)
- [96100e5f](https://github.com/kubedb/ui-server/commit/96100e5f) Use Go 1.19 (#48)
- [99bc4723](https://github.com/kubedb/ui-server/commit/99bc4723) Use k8s 1.25.1 libs (#47)
- [2c0ba4c1](https://github.com/kubedb/ui-server/commit/2c0ba4c1) Stop using removed apis in Kubernetes 1.25 (#46)
- [fc35287c](https://github.com/kubedb/ui-server/commit/fc35287c) Use health checker types from kmodules (#45)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.5.0](https://github.com/kubedb/webhook-server/releases/tag/v0.5.0)

- [41c41749](https://github.com/kubedb/webhook-server/commit/41c41749) Prepare for release v0.5.0 (#37)
- [eaa32942](https://github.com/kubedb/webhook-server/commit/eaa32942) Use password-generator@v0.2.9 (#36)
- [c06f6c42](https://github.com/kubedb/webhook-server/commit/c06f6c42) Register the missing types to webhook (#35)
- [59cb7fa0](https://github.com/kubedb/webhook-server/commit/59cb7fa0) Register pg sub/sub validators (#34)
- [1de1fe03](https://github.com/kubedb/webhook-server/commit/1de1fe03) Prepare for release v0.5.0-rc.0 (#33)
- [8f65154d](https://github.com/kubedb/webhook-server/commit/8f65154d) Test against Kubernetes 1.25.0 (#31)
- [ed6ba664](https://github.com/kubedb/webhook-server/commit/ed6ba664) Check for PDB version only once (#29)
- [ab4e44d0](https://github.com/kubedb/webhook-server/commit/ab4e44d0) Handle status conversion for CronJob/VolumeSnapshot (#28)
- [aef864b7](https://github.com/kubedb/webhook-server/commit/aef864b7) Use Go 1.19 (#27)




