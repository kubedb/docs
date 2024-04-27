---
title: Changelog | KubeDB
description: Changelog
menu:
  docs_{{.version}}:
    identifier: changelog-kubedb-v2024.4.27
    name: Changelog-v2024.4.27
    parent: welcome
    weight: 20240427
product_name: kubedb
menu_name: docs_{{.version}}
section_menu_id: welcome
url: /docs/{{.version}}/welcome/changelog-v2024.4.27/
aliases:
  - /docs/{{.version}}/CHANGELOG-v2024.4.27/
---

# KubeDB v2024.4.27 (2024-04-27)


## [kubedb/apimachinery](https://github.com/kubedb/apimachinery)

### [v0.45.0](https://github.com/kubedb/apimachinery/releases/tag/v0.45.0)

- [104a8209](https://github.com/kubedb/apimachinery/commit/104a82097) MSSQL -> MSSQLServer (#1202)
- [f93640ff](https://github.com/kubedb/apimachinery/commit/f93640fff) Add RabbitMQ Autoscaler API (#1194)
- [06ac2b3b](https://github.com/kubedb/apimachinery/commit/06ac2b3b2) Add RabbitMQ OpsRequest API (#1183)
- [7e031847](https://github.com/kubedb/apimachinery/commit/7e031847c) Add Memcached Health Check api (#1200)
- [c3b34d02](https://github.com/kubedb/apimachinery/commit/c3b34d027) Update offshoot-api (#1201)
- [7fb3d561](https://github.com/kubedb/apimachinery/commit/7fb3d5619) Add SingleStore TLS (#1196)
- [84398fe4](https://github.com/kubedb/apimachinery/commit/84398fe4e) Add Pgpool TLS (#1199)
- [6efef42b](https://github.com/kubedb/apimachinery/commit/6efef42b8) Add MS SQL Server APIs  (#1174)
- [1767aac0](https://github.com/kubedb/apimachinery/commit/1767aac04) Update druid exporter port (#1191)
- [837bbc4d](https://github.com/kubedb/apimachinery/commit/837bbc4d8) Use `applyConfig` in reconfigure opsReqs; Remove inlineConfig (#1144)
- [1955e7f0](https://github.com/kubedb/apimachinery/commit/1955e7f0a) Update deps
- [3bdcf426](https://github.com/kubedb/apimachinery/commit/3bdcf426d) Use Go 1.22 (#1192)
- [7386252c](https://github.com/kubedb/apimachinery/commit/7386252ce) Add RabbitMQ shovel and federation plugin constants (#1190)
- [39554487](https://github.com/kubedb/apimachinery/commit/39554487f) Update deps
- [0b85ea37](https://github.com/kubedb/apimachinery/commit/0b85ea377) Add Port Field to NamedURL (#1189)
- [adf39bff](https://github.com/kubedb/apimachinery/commit/adf39bff0) Use gateway port in db status (#1188)
- [3768ead9](https://github.com/kubedb/apimachinery/commit/3768ead95) Fix Stash Restore Target Issue (#1185)
- [4c457ff1](https://github.com/kubedb/apimachinery/commit/4c457ff1a) Mutate prometheus exporter. (#1187)
- [c9359d20](https://github.com/kubedb/apimachinery/commit/c9359d205) Fix Druid resource issue (#1186)
- [eca0297e](https://github.com/kubedb/apimachinery/commit/eca0297e1) Add metrics emitter path for Druid Monitoring (#1182)
- [cfd5a061](https://github.com/kubedb/apimachinery/commit/cfd5a0615) Use Go 1.22 (#1179)



## [kubedb/autoscaler](https://github.com/kubedb/autoscaler)

### [v0.30.0](https://github.com/kubedb/autoscaler/releases/tag/v0.30.0)

- [b5d7a386](https://github.com/kubedb/autoscaler/commit/b5d7a386) Prepare for release v0.30.0 (#203)
- [e830ed82](https://github.com/kubedb/autoscaler/commit/e830ed82) Trigger VolumeExpansion according to scalingRules (#202)
- [a5c6c179](https://github.com/kubedb/autoscaler/commit/a5c6c179) Add Autoscaler for RabbitMQ (#200)
- [ab11f0f4](https://github.com/kubedb/autoscaler/commit/ab11f0f4) Fetch from petsets also (#201)
- [8f2727f2](https://github.com/kubedb/autoscaler/commit/8f2727f2) Use applyConfig (#199)



## [kubedb/cli](https://github.com/kubedb/cli)

### [v0.45.0](https://github.com/kubedb/cli/releases/tag/v0.45.0)

- [f93894ac](https://github.com/kubedb/cli/commit/f93894ac) Prepare for release v0.45.0 (#765)
- [5fc61456](https://github.com/kubedb/cli/commit/5fc61456) Prepare for release v0.45.0 (#764)
- [5827f93f](https://github.com/kubedb/cli/commit/5827f93f) Use Go 1.22 (#763)



## [kubedb/crd-manager](https://github.com/kubedb/crd-manager)

### [v0.0.9](https://github.com/kubedb/crd-manager/releases/tag/v0.0.9)

- [ebd4310](https://github.com/kubedb/crd-manager/commit/ebd4310) Prepare for release v0.0.9 (#26)
- [6b411dc](https://github.com/kubedb/crd-manager/commit/6b411dc) Prepare for release v0.0.9 (#25)
- [a5dc013](https://github.com/kubedb/crd-manager/commit/a5dc013) Ensure RabbitMQ Opsreq and autoscaler CRDs (#24)
- [3ebea21](https://github.com/kubedb/crd-manager/commit/3ebea21) Update MSSQLServer api (#23)
- [6f87254](https://github.com/kubedb/crd-manager/commit/6f87254) Add Support for MsSQL (#20)
- [065bad4](https://github.com/kubedb/crd-manager/commit/065bad4) MicrosoftSQLServer -> MSSQLServer
- [8cf13f0](https://github.com/kubedb/crd-manager/commit/8cf13f0) Use deps (#22)
- [9f08ee7](https://github.com/kubedb/crd-manager/commit/9f08ee7) Use Go 1.22 (#21)
- [c138cd8](https://github.com/kubedb/crd-manager/commit/c138cd8) Update license header
- [ef190af](https://github.com/kubedb/crd-manager/commit/ef190af) Update db list



## [kubedb/dashboard](https://github.com/kubedb/dashboard)

### [v0.21.0](https://github.com/kubedb/dashboard/releases/tag/v0.21.0)

- [b1dc72c1](https://github.com/kubedb/dashboard/commit/b1dc72c1) Prepare for release v0.21.0 (#113)
- [a0ce7e6c](https://github.com/kubedb/dashboard/commit/a0ce7e6c) Use deps (#112)
- [fb3f18ba](https://github.com/kubedb/dashboard/commit/fb3f18ba) Use Go 1.22 (#111)



## [kubedb/dashboard-restic-plugin](https://github.com/kubedb/dashboard-restic-plugin)

### [v0.3.0](https://github.com/kubedb/dashboard-restic-plugin/releases/tag/v0.3.0)

- [54be40d](https://github.com/kubedb/dashboard-restic-plugin/commit/54be40d) Add support for create space for dashboard restore (#7)
- [b74d46e](https://github.com/kubedb/dashboard-restic-plugin/commit/b74d46e) Prepare for release v0.3.0 (#8)
- [81e3440](https://github.com/kubedb/dashboard-restic-plugin/commit/81e3440) Use restic 0.16.4 (#6)
- [11108d4](https://github.com/kubedb/dashboard-restic-plugin/commit/11108d4) Use Go 1.22 (#5)



## [kubedb/db-client-go](https://github.com/kubedb/db-client-go)

### [v0.0.15](https://github.com/kubedb/db-client-go/releases/tag/v0.0.15)

- [d3a1eb1c](https://github.com/kubedb/db-client-go/commit/d3a1eb1c) Prepare for release v0.0.15 (#104)
- [f377b7fd](https://github.com/kubedb/db-client-go/commit/f377b7fd) Fix driver name (#103)
- [31518cdb](https://github.com/kubedb/db-client-go/commit/31518cdb) MSSQL -> MSSQLServer (#102)
- [8638376f](https://github.com/kubedb/db-client-go/commit/8638376f) Add tls for Pgpool client (#99)
- [4edbe630](https://github.com/kubedb/db-client-go/commit/4edbe630) Add SingleStore TLS (#97)
- [8c73c085](https://github.com/kubedb/db-client-go/commit/8c73c085) Add MS SQL Server DB Client (#88)
- [4b4e47e1](https://github.com/kubedb/db-client-go/commit/4b4e47e1) Add WithCred() & WithAuthDatabase() builder utility (#98)
- [deb90f66](https://github.com/kubedb/db-client-go/commit/deb90f66) Add AMQP client Getter for RabbitMQ (#96)
- [f4fab3d2](https://github.com/kubedb/db-client-go/commit/f4fab3d2) Add create space method for dashboard (#95)
- [e01cb9ea](https://github.com/kubedb/db-client-go/commit/e01cb9ea) Update deps



## [kubedb/druid](https://github.com/kubedb/druid)

### [v0.0.9](https://github.com/kubedb/druid/releases/tag/v0.0.9)

- [5901334](https://github.com/kubedb/druid/commit/5901334) Prepare for release v0.0.9 (#22)
- [c77e449](https://github.com/kubedb/druid/commit/c77e449) Prepare for release v0.0.9 (#21)
- [831f543](https://github.com/kubedb/druid/commit/831f543) Set Defaults to db on reconcile (#20)
- [b1ca09c](https://github.com/kubedb/druid/commit/b1ca09c) Use deps (#19)
- [05dd9ca](https://github.com/kubedb/druid/commit/05dd9ca) Use Go 1.22 (#18)
- [6c77255](https://github.com/kubedb/druid/commit/6c77255) Add support for monitoring (#13)
- [a834fc5](https://github.com/kubedb/druid/commit/a834fc5) Remove license check for webhook-server (#17)



## [kubedb/elasticsearch](https://github.com/kubedb/elasticsearch)

### [v0.45.0](https://github.com/kubedb/elasticsearch/releases/tag/v0.45.0)

- [744a4329](https://github.com/kubedb/elasticsearch/commit/744a4329f) Prepare for release v0.45.0 (#718)
- [6198c23e](https://github.com/kubedb/elasticsearch/commit/6198c23ee) Prepare for release v0.45.0 (#717)
- [ea61d011](https://github.com/kubedb/elasticsearch/commit/ea61d011d) Remove redundant log for monitoring agent not found (#716)
- [04ad15dc](https://github.com/kubedb/elasticsearch/commit/04ad15dcf) Use restic 0.16.4 (#715)
- [4bdd974f](https://github.com/kubedb/elasticsearch/commit/4bdd974f4) Use Go 1.22 (#714)
- [98dab6a7](https://github.com/kubedb/elasticsearch/commit/98dab6a76) Remove License Check from Webhook Server (#713)



## [kubedb/elasticsearch-restic-plugin](https://github.com/kubedb/elasticsearch-restic-plugin)

### [v0.8.0](https://github.com/kubedb/elasticsearch-restic-plugin/releases/tag/v0.8.0)

- [aee2a2a](https://github.com/kubedb/elasticsearch-restic-plugin/commit/aee2a2a) Prepare for release v0.8.0 (#30)
- [4a06148](https://github.com/kubedb/elasticsearch-restic-plugin/commit/4a06148) Prepare for release v0.8.0 (#29)
- [10f442c](https://github.com/kubedb/elasticsearch-restic-plugin/commit/10f442c) Use restic 0.16.4 (#28)
- [8f2be78](https://github.com/kubedb/elasticsearch-restic-plugin/commit/8f2be78) Use Go 1.22 (#27)



## [kubedb/ferretdb](https://github.com/kubedb/ferretdb)

### [v0.0.9](https://github.com/kubedb/ferretdb/releases/tag/v0.0.9)

- [b8b2db43](https://github.com/kubedb/ferretdb/commit/b8b2db43) Prepare for release v0.0.9 (#22)
- [321300f9](https://github.com/kubedb/ferretdb/commit/321300f9) Prepare for release v0.0.9 (#21)
- [a3552e21](https://github.com/kubedb/ferretdb/commit/a3552e21) Update api (#20)
- [b15a0032](https://github.com/kubedb/ferretdb/commit/b15a0032) Set Defaults to db on reconcile (#19)
- [9276d3c6](https://github.com/kubedb/ferretdb/commit/9276d3c6) Use deps (#18)
- [c38d6ea1](https://github.com/kubedb/ferretdb/commit/c38d6ea1) Use Go 1.22 (#17)
- [35188b72](https://github.com/kubedb/ferretdb/commit/35188b72) Remove: license check for webhook server (#16)



## [kubedb/installer](https://github.com/kubedb/installer)

### [v2024.4.27](https://github.com/kubedb/installer/releases/tag/v2024.4.27)




## [kubedb/kafka](https://github.com/kubedb/kafka)

### [v0.16.0](https://github.com/kubedb/kafka/releases/tag/v0.16.0)

- [088b36f0](https://github.com/kubedb/kafka/commit/088b36f0) Prepare for release v0.16.0 (#89)
- [9151e670](https://github.com/kubedb/kafka/commit/9151e670) Prepare for release v0.16.0 (#88)
- [dec0a20c](https://github.com/kubedb/kafka/commit/dec0a20c) Set Defaults to db on reconcile (#87)
- [2b88de4e](https://github.com/kubedb/kafka/commit/2b88de4e) Add Kafka and Connect Cluster PDB (#84)
- [847c8fe2](https://github.com/kubedb/kafka/commit/847c8fe2) Use deps (#86)
- [6880da47](https://github.com/kubedb/kafka/commit/6880da47) Use Go 1.22 (#85)
- [9242c91d](https://github.com/kubedb/kafka/commit/9242c91d) Remove license check for webhook-server (#83)



## [kubedb/kubedb-manifest-plugin](https://github.com/kubedb/kubedb-manifest-plugin)

### [v0.8.0](https://github.com/kubedb/kubedb-manifest-plugin/releases/tag/v0.8.0)

- [feeb3e9](https://github.com/kubedb/kubedb-manifest-plugin/commit/feeb3e9) Prepare for release v0.8.0 (#51)
- [8ec129f](https://github.com/kubedb/kubedb-manifest-plugin/commit/8ec129f) Prepare for release v0.8.0 (#50)
- [b3c08dd](https://github.com/kubedb/kubedb-manifest-plugin/commit/b3c08dd) Use restic 0.16.4 (#49)
- [4edccaa](https://github.com/kubedb/kubedb-manifest-plugin/commit/4edccaa) Use Go 1.22 (#48)



## [kubedb/mariadb](https://github.com/kubedb/mariadb)

### [v0.29.0](https://github.com/kubedb/mariadb/releases/tag/v0.29.0)

- [cb49215e](https://github.com/kubedb/mariadb/commit/cb49215ee) Prepare for release v0.29.0 (#267)
- [f709a65e](https://github.com/kubedb/mariadb/commit/f709a65e4) Prepare for release v0.29.0 (#266)
- [08936bee](https://github.com/kubedb/mariadb/commit/08936bee1) Use deps (#265)
- [bfdbc178](https://github.com/kubedb/mariadb/commit/bfdbc178d) Use Go 1.22 (#264)



## [kubedb/mariadb-archiver](https://github.com/kubedb/mariadb-archiver)

### [v0.5.0](https://github.com/kubedb/mariadb-archiver/releases/tag/v0.5.0)

- [2255429](https://github.com/kubedb/mariadb-archiver/commit/2255429) Prepare for release v0.5.0 (#16)
- [e638a7b](https://github.com/kubedb/mariadb-archiver/commit/e638a7b) Prepare for release v0.5.0 (#15)
- [ac5b6d6](https://github.com/kubedb/mariadb-archiver/commit/ac5b6d6) Use Go 1.22 (#14)



## [kubedb/mariadb-coordinator](https://github.com/kubedb/mariadb-coordinator)

### [v0.25.0](https://github.com/kubedb/mariadb-coordinator/releases/tag/v0.25.0)

- [ebc40587](https://github.com/kubedb/mariadb-coordinator/commit/ebc40587) Prepare for release v0.25.0 (#115)
- [653202e2](https://github.com/kubedb/mariadb-coordinator/commit/653202e2) Prepare for release v0.25.0 (#114)
- [15fd803a](https://github.com/kubedb/mariadb-coordinator/commit/15fd803a) Use Go 1.22 (#113)



## [kubedb/mariadb-csi-snapshotter-plugin](https://github.com/kubedb/mariadb-csi-snapshotter-plugin)

### [v0.5.0](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/releases/tag/v0.5.0)

- [f9cae9d](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/f9cae9d) Prepare for release v0.5.0 (#19)
- [3741c24](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/3741c24) Prepare for release v0.5.0 (#18)
- [edbee83](https://github.com/kubedb/mariadb-csi-snapshotter-plugin/commit/edbee83) Use Go 1.22 (#17)



## [kubedb/mariadb-restic-plugin](https://github.com/kubedb/mariadb-restic-plugin)

### [v0.3.0](https://github.com/kubedb/mariadb-restic-plugin/releases/tag/v0.3.0)

- [38a2bea](https://github.com/kubedb/mariadb-restic-plugin/commit/38a2bea) Prepare for release v0.3.0 (#9)
- [0e030ae](https://github.com/kubedb/mariadb-restic-plugin/commit/0e030ae) Prepare for release v0.3.0 (#8)
- [eedd422](https://github.com/kubedb/mariadb-restic-plugin/commit/eedd422) Use restic 0.16.4 (#7)
- [59287df](https://github.com/kubedb/mariadb-restic-plugin/commit/59287df) Use Go 1.22 (#6)



## [kubedb/memcached](https://github.com/kubedb/memcached)

### [v0.38.0](https://github.com/kubedb/memcached/releases/tag/v0.38.0)

- [c60a1ada](https://github.com/kubedb/memcached/commit/c60a1ada) Prepare for release v0.38.0 (#440)
- [2abd9694](https://github.com/kubedb/memcached/commit/2abd9694) Prepare for release v0.38.0 (#439)
- [b019d4a5](https://github.com/kubedb/memcached/commit/b019d4a5) Added Memcached Health (#437)
- [032d1be9](https://github.com/kubedb/memcached/commit/032d1be9) Remove redundant log for monitoring agent not found (#431)
- [2018e90d](https://github.com/kubedb/memcached/commit/2018e90d) Use deps (#430)
- [6aadd80d](https://github.com/kubedb/memcached/commit/6aadd80d) Use Go 1.22 (#429)
- [1521e538](https://github.com/kubedb/memcached/commit/1521e538) Fix webhook server (#428)



## [kubedb/mongodb](https://github.com/kubedb/mongodb)

### [v0.38.0](https://github.com/kubedb/mongodb/releases/tag/v0.38.0)

- [53dee180](https://github.com/kubedb/mongodb/commit/53dee1806) Prepare for release v0.38.0 (#628)
- [d4e6f1f2](https://github.com/kubedb/mongodb/commit/d4e6f1f2b) Prepare for release v0.38.0 (#627)
- [4aa876c0](https://github.com/kubedb/mongodb/commit/4aa876c07) Remove redundant log for monitoring agent not found (#625)
- [94111489](https://github.com/kubedb/mongodb/commit/94111489c) Use deps (#624)
- [5ae7a268](https://github.com/kubedb/mongodb/commit/5ae7a2686) Use Go 1.22 (#623)
- [e35a5838](https://github.com/kubedb/mongodb/commit/e35a58388) Skip license check for webhook-server (#622)



## [kubedb/mongodb-csi-snapshotter-plugin](https://github.com/kubedb/mongodb-csi-snapshotter-plugin)

### [v0.6.0](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/releases/tag/v0.6.0)

- [8119429](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/8119429) Prepare for release v0.6.0 (#24)
- [27e13d9](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/27e13d9) Prepare for release v0.6.0 (#23)
- [519cec3](https://github.com/kubedb/mongodb-csi-snapshotter-plugin/commit/519cec3) Use Go 1.22 (#22)



## [kubedb/mongodb-restic-plugin](https://github.com/kubedb/mongodb-restic-plugin)

### [v0.8.0](https://github.com/kubedb/mongodb-restic-plugin/releases/tag/v0.8.0)

- [9b2f456](https://github.com/kubedb/mongodb-restic-plugin/commit/9b2f456) Prepare for release v0.8.0 (#42)
- [9c629ed](https://github.com/kubedb/mongodb-restic-plugin/commit/9c629ed) Prepare for release v0.8.0 (#41)
- [7ed30f3](https://github.com/kubedb/mongodb-restic-plugin/commit/7ed30f3) fix: restoresession running phase (#38)
- [390b27a](https://github.com/kubedb/mongodb-restic-plugin/commit/390b27a) Use restic 0.16.4 (#40)
- [e702a1c](https://github.com/kubedb/mongodb-restic-plugin/commit/e702a1c) Use Go 1.22 (#39)



## [kubedb/mssql](https://github.com/kubedb/mssql)

### [v0.0.1](https://github.com/kubedb/mssql/releases/tag/v0.0.1)

- [5ec0f33](https://github.com/kubedb/mssql/commit/5ec0f33) Prepare for release v0.0.1 (#6)
- [7b3e1a9](https://github.com/kubedb/mssql/commit/7b3e1a9) Add MS SQL Server Provisioner Operator (#5)



## [kubedb/mssql-coordinator](https://github.com/kubedb/mssql-coordinator)

### [v0.0.1](https://github.com/kubedb/mssql-coordinator/releases/tag/v0.0.1)

- [d85fb389](https://github.com/kubedb/mssql-coordinator/commit/d85fb389) Prepare for release v0.0.1 (#3)
- [ca78acac](https://github.com/kubedb/mssql-coordinator/commit/ca78acac) Prepare for release v0.0.1 (#2)
- [09ac9814](https://github.com/kubedb/mssql-coordinator/commit/09ac9814) Provision MS SQL Server with Raft (#1)
- [39120f71](https://github.com/kubedb/mssql-coordinator/commit/39120f71) pg-coordinator -> mssql-coordinator
- [aba0d83f](https://github.com/kubedb/mssql-coordinator/commit/aba0d83f) Prepare for release v0.26.0 (#155)
- [e9c0a986](https://github.com/kubedb/mssql-coordinator/commit/e9c0a986) Prepare for release v0.25.0 (#154)
- [cbdb8698](https://github.com/kubedb/mssql-coordinator/commit/cbdb8698) Prepare for release v0.25.0-rc.1 (#153)
- [e4042bfd](https://github.com/kubedb/mssql-coordinator/commit/e4042bfd) Update deps (#152)
- [b9b84db3](https://github.com/kubedb/mssql-coordinator/commit/b9b84db3) Update deps (#151)
- [e35b556c](https://github.com/kubedb/mssql-coordinator/commit/e35b556c) Prepare for release v0.25.0-rc.0 (#150)
- [fa77ab4e](https://github.com/kubedb/mssql-coordinator/commit/fa77ab4e) Fixed (#149)
- [aef2df3c](https://github.com/kubedb/mssql-coordinator/commit/aef2df3c) Prepare for release v0.25.0-beta.1 (#148)
- [f00e1837](https://github.com/kubedb/mssql-coordinator/commit/f00e1837) Prepare for release v0.25.0-beta.0 (#147)
- [00ef3cb7](https://github.com/kubedb/mssql-coordinator/commit/00ef3cb7) Update deps (#146)
- [30783ca6](https://github.com/kubedb/mssql-coordinator/commit/30783ca6) Update deps (#145)
- [656f9a82](https://github.com/kubedb/mssql-coordinator/commit/656f9a82) Use k8s 1.29 client libs (#144)
- [8e4c36ec](https://github.com/kubedb/mssql-coordinator/commit/8e4c36ec) Prepare for release v0.24.0 (#143)
- [0b8239a1](https://github.com/kubedb/mssql-coordinator/commit/0b8239a1) Prepare for release v0.23.0 (#142)
- [c14f61d1](https://github.com/kubedb/mssql-coordinator/commit/c14f61d1) Prepare for release v0.22.0 (#141)
- [56e5b1de](https://github.com/kubedb/mssql-coordinator/commit/56e5b1de) Prepare for release v0.22.0-rc.1 (#140)
- [3379cb47](https://github.com/kubedb/mssql-coordinator/commit/3379cb47) Prepare for release v0.22.0-rc.0 (#139)
- [1b9d3b30](https://github.com/kubedb/mssql-coordinator/commit/1b9d3b30) Add support for arbiter (#136)
- [63690f56](https://github.com/kubedb/mssql-coordinator/commit/63690f56) added postgres 16.0 support (#137)
- [63637c30](https://github.com/kubedb/mssql-coordinator/commit/63637c30) Added & modified logs (#134)
- [c90d2edc](https://github.com/kubedb/mssql-coordinator/commit/c90d2edc) Prepare for release v0.21.0 (#135)
- [060d7da9](https://github.com/kubedb/mssql-coordinator/commit/060d7da9) Prepare for release v0.20.0 (#133)
- [867a5a36](https://github.com/kubedb/mssql-coordinator/commit/867a5a36) Fix standby labeling (#132)
- [676e2957](https://github.com/kubedb/mssql-coordinator/commit/676e2957) Only cache sibling pods using selector (#131)
- [9732ad99](https://github.com/kubedb/mssql-coordinator/commit/9732ad99) fix linter issue (#130)
- [8bf809a2](https://github.com/kubedb/mssql-coordinator/commit/8bf809a2) Prepare for release v0.19.0 (#129)
- [dd87bd6a](https://github.com/kubedb/mssql-coordinator/commit/dd87bd6a) Update dependencies (#128)
- [ccd18e6d](https://github.com/kubedb/mssql-coordinator/commit/ccd18e6d) Use cached client (#127)
- [368c11d9](https://github.com/kubedb/mssql-coordinator/commit/368c11d9) Update dependencies (#126)
- [36094ae6](https://github.com/kubedb/mssql-coordinator/commit/36094ae6) fix failover and standby sync issue (#125)
- [8ff1d1a2](https://github.com/kubedb/mssql-coordinator/commit/8ff1d1a2) Prepare for release v0.18.0 (#124)
- [70423ac2](https://github.com/kubedb/mssql-coordinator/commit/70423ac2) Prepare for release v0.18.0-rc.0 (#123)
- [6e3f71a9](https://github.com/kubedb/mssql-coordinator/commit/6e3f71a9) Update license verifier (#122)
- [5a20e0a8](https://github.com/kubedb/mssql-coordinator/commit/5a20e0a8) Update license verifier (#121)
- [bb781f89](https://github.com/kubedb/mssql-coordinator/commit/bb781f89) Add enableServiceLinks to PodSpec (#120)
- [9e313041](https://github.com/kubedb/mssql-coordinator/commit/9e313041) Test against K8s 1.27.0 (#119)
- [abf0822f](https://github.com/kubedb/mssql-coordinator/commit/abf0822f) Prepare for release v0.17.0 (#118)
- [f6b27009](https://github.com/kubedb/mssql-coordinator/commit/f6b27009) Cleanup CI
- [c359680e](https://github.com/kubedb/mssql-coordinator/commit/c359680e) Use ghcr.io for appscode/golang-dev (#117)
- [53afc3f5](https://github.com/kubedb/mssql-coordinator/commit/53afc3f5) Dynamically select runner type
- [ce64e41a](https://github.com/kubedb/mssql-coordinator/commit/ce64e41a) Update workflows (Go 1.20, k8s 1.26) (#116)
- [754d2dad](https://github.com/kubedb/mssql-coordinator/commit/754d2dad) Test against Kubernetes 1.26.0 (#114)
- [773c1e37](https://github.com/kubedb/mssql-coordinator/commit/773c1e37) Prepare for release v0.16.0 (#113)
- [e160b1c6](https://github.com/kubedb/mssql-coordinator/commit/e160b1c6) Update sidekick dependency (#112)
- [2835261c](https://github.com/kubedb/mssql-coordinator/commit/2835261c) Read imge pull secret from operator flags (#111)
- [916f0c37](https://github.com/kubedb/mssql-coordinator/commit/916f0c37) Fix `waiting for the target to be leader` issue (#110)
- [e28d7b5c](https://github.com/kubedb/mssql-coordinator/commit/e28d7b5c) Prepare for release v0.14.1 (#109)
- [32cf0205](https://github.com/kubedb/mssql-coordinator/commit/32cf0205) Prepare for release v0.14.0 (#108)
- [706969bd](https://github.com/kubedb/mssql-coordinator/commit/706969bd) Update dependencies (#107)
- [02340b1c](https://github.com/kubedb/mssql-coordinator/commit/02340b1c) Prepare for release v0.14.0-rc.1 (#106)
- [51942bbd](https://github.com/kubedb/mssql-coordinator/commit/51942bbd) Prepare for release v0.14.0-rc.0 (#105)
- [c14a0d4c](https://github.com/kubedb/mssql-coordinator/commit/c14a0d4c) Update deps (#104)
- [5e6730a5](https://github.com/kubedb/mssql-coordinator/commit/5e6730a5) Merge pull request #102 from kubedb/leader-switch
- [e25599f0](https://github.com/kubedb/mssql-coordinator/commit/e25599f0) Merge branch 'master' into leader-switch
- [2708d012](https://github.com/kubedb/mssql-coordinator/commit/2708d012) Add PG Reset Wal for Single user mode failed #101
- [4d8bd4c3](https://github.com/kubedb/mssql-coordinator/commit/4d8bd4c3) retry eviction of pod and delete pod if fails
- [0e09b7ec](https://github.com/kubedb/mssql-coordinator/commit/0e09b7ec) Update deps
- [404ebc14](https://github.com/kubedb/mssql-coordinator/commit/404ebc14) Refined
- [c0eaf986](https://github.com/kubedb/mssql-coordinator/commit/c0eaf986) Fix: Transfer Leadership issue fix with pod delete
- [3ed11903](https://github.com/kubedb/mssql-coordinator/commit/3ed11903) Add PG Reset Wal for Single user mode failed
- [194c1bcb](https://github.com/kubedb/mssql-coordinator/commit/194c1bcb) Run GH actions on ubuntu-20.04 (#103)
- [8df98fa1](https://github.com/kubedb/mssql-coordinator/commit/8df98fa1) Prepare for release v0.13.0 (#100)
- [34c72fda](https://github.com/kubedb/mssql-coordinator/commit/34c72fda) Prepare for release v0.13.0-rc.0 (#99)
- [0295aeca](https://github.com/kubedb/mssql-coordinator/commit/0295aeca) Update dependencies (#98)
- [9b371d85](https://github.com/kubedb/mssql-coordinator/commit/9b371d85) Test against Kubernetes 1.25.0 (#97)
- [1a27ef77](https://github.com/kubedb/mssql-coordinator/commit/1a27ef77) Check for PDB version only once (#95)
- [04c780bd](https://github.com/kubedb/mssql-coordinator/commit/04c780bd) Handle status conversion for CronJob/VolumeSnapshot (#94)
- [b71a1118](https://github.com/kubedb/mssql-coordinator/commit/b71a1118) Use Go 1.19 (#93)
- [6270e67d](https://github.com/kubedb/mssql-coordinator/commit/6270e67d) Use k8s 1.25.1 libs (#92)
- [84787e88](https://github.com/kubedb/mssql-coordinator/commit/84787e88) Stop using removed apis in Kubernetes 1.25 (#91)
- [7550a387](https://github.com/kubedb/mssql-coordinator/commit/7550a387) Use health checker types from kmodules (#90)
- [c53a50ad](https://github.com/kubedb/mssql-coordinator/commit/c53a50ad) Prepare for release v0.12.0 (#89)
- [ee679735](https://github.com/kubedb/mssql-coordinator/commit/ee679735) Prepare for release v0.12.0-rc.1 (#88)
- [ca070943](https://github.com/kubedb/mssql-coordinator/commit/ca070943) Update health checker (#86)
- [c38bf66b](https://github.com/kubedb/mssql-coordinator/commit/c38bf66b) Prepare for release v0.12.0-rc.0 (#85)
- [ee946044](https://github.com/kubedb/mssql-coordinator/commit/ee946044) Acquire license from license-proxyserver if available (#83)
- [321e6a8f](https://github.com/kubedb/mssql-coordinator/commit/321e6a8f) Remove role scripts from the coordinator. (#82)
- [6d679b22](https://github.com/kubedb/mssql-coordinator/commit/6d679b22) Update to k8s 1.24 toolchain (#81)
- [cde3602d](https://github.com/kubedb/mssql-coordinator/commit/cde3602d) Prepare for release v0.11.0 (#80)
- [d20975b6](https://github.com/kubedb/mssql-coordinator/commit/d20975b6) Update dependencies (#79)
- [cb516fbc](https://github.com/kubedb/mssql-coordinator/commit/cb516fbc) Add Raft Metrics And graceful shutdown of Postgres (#74)
- [e78a2c5b](https://github.com/kubedb/mssql-coordinator/commit/e78a2c5b) Update dependencies(nats client, mongo-driver) (#78)
- [c4af9ba0](https://github.com/kubedb/mssql-coordinator/commit/c4af9ba0) Fix: Fast Shut-down Postgres server to avoid single-user mode shutdown failure (#73)
- [a9661653](https://github.com/kubedb/mssql-coordinator/commit/a9661653) Prepare for release v0.10.0 (#72)
- [3774d5c8](https://github.com/kubedb/mssql-coordinator/commit/3774d5c8) Update dependencies (#71)
- [f72f0a31](https://github.com/kubedb/mssql-coordinator/commit/f72f0a31) Avoid ExacIntoPod to fix memory leak (#70)
- [b1d68e3e](https://github.com/kubedb/mssql-coordinator/commit/b1d68e3e) Use Go 1.18 (#68)
- [dfecf3c7](https://github.com/kubedb/mssql-coordinator/commit/dfecf3c7) make fmt (#67)
- [389e47ea](https://github.com/kubedb/mssql-coordinator/commit/389e47ea) Prepare for release v0.9.0 (#66)
- [d6c1b92b](https://github.com/kubedb/mssql-coordinator/commit/d6c1b92b) Cancel concurrent CI runs for same pr/commit (#65)
- [122e1291](https://github.com/kubedb/mssql-coordinator/commit/122e1291) Update dependencies (#64)
- [8fab8e7a](https://github.com/kubedb/mssql-coordinator/commit/8fab8e7a) Cancel concurrent CI runs for same pr/commit (#63)
- [749e8501](https://github.com/kubedb/mssql-coordinator/commit/749e8501) Update SiteInfo (#62)
- [1b11bc19](https://github.com/kubedb/mssql-coordinator/commit/1b11bc19) Publish GenericResource (#61)
- [9a256f15](https://github.com/kubedb/mssql-coordinator/commit/9a256f15) Fix custom Auth secret issues (#60)
- [2b3434ac](https://github.com/kubedb/mssql-coordinator/commit/2b3434ac) Use Postgres CR to get replica count (#59)
- [8ded7410](https://github.com/kubedb/mssql-coordinator/commit/8ded7410) Recover from panic in reconcilers (#58)
- [917270b8](https://github.com/kubedb/mssql-coordinator/commit/917270b8) Prepare for release v0.8.0 (#57)
- [f84a1dc0](https://github.com/kubedb/mssql-coordinator/commit/f84a1dc0) Update dependencies (#56)
- [076b8ede](https://github.com/kubedb/mssql-coordinator/commit/076b8ede) Check if pods are controlled by kubedb statefulset (#55)
- [4c55eed6](https://github.com/kubedb/mssql-coordinator/commit/4c55eed6) Prepare for release v0.7.0 (#54)
- [f98e8947](https://github.com/kubedb/mssql-coordinator/commit/f98e8947) Update kmodules.xyz/monitoring-agent-api (#53)
- [6f59d29f](https://github.com/kubedb/mssql-coordinator/commit/6f59d29f) Update repository config (#52)
- [5ff5f4b8](https://github.com/kubedb/mssql-coordinator/commit/5ff5f4b8) Fix: Raft log corrupted issue (#51)
- [314fd357](https://github.com/kubedb/mssql-coordinator/commit/314fd357) Use DisableAnalytics flag from license (#50)
- [88b19fcc](https://github.com/kubedb/mssql-coordinator/commit/88b19fcc) Update license-verifier (#49)
- [06248438](https://github.com/kubedb/mssql-coordinator/commit/06248438) Support custom pod and controller labels (#48)
- [633bfcac](https://github.com/kubedb/mssql-coordinator/commit/633bfcac) Postgres Server Restart If Sig-Killed (#44)
- [5047efeb](https://github.com/kubedb/mssql-coordinator/commit/5047efeb) Print logs at Debug level
- [7d42c4bd](https://github.com/kubedb/mssql-coordinator/commit/7d42c4bd) Log timestamp from zap logger used in raft (#47)
- [137015d0](https://github.com/kubedb/mssql-coordinator/commit/137015d0) Update xorm dependency (#46)
- [81f78270](https://github.com/kubedb/mssql-coordinator/commit/81f78270) Fix satori/go.uuid security vulnerability (#45)
- [b3685227](https://github.com/kubedb/mssql-coordinator/commit/b3685227) Fix jwt-go security vulnerability (#43)
- [9bfed99c](https://github.com/kubedb/mssql-coordinator/commit/9bfed99c) Fix: Postgres server single user mode start for bullseye image (#42)
- [f55f8d6e](https://github.com/kubedb/mssql-coordinator/commit/f55f8d6e) Update dependencies to publish SiteInfo (#40)
- [b82a3d77](https://github.com/kubedb/mssql-coordinator/commit/b82a3d77) Add support for Postgres version v14.0 (#41)
- [158105b8](https://github.com/kubedb/mssql-coordinator/commit/158105b8) Prepare for release v0.6.0 (#39)
- [4b7e8593](https://github.com/kubedb/mssql-coordinator/commit/4b7e8593) Log warning if Community License is used with non-demo namespace (#38)
- [f87aee15](https://github.com/kubedb/mssql-coordinator/commit/f87aee15) Prepare for release v0.5.0 (#37)
- [4357d0b7](https://github.com/kubedb/mssql-coordinator/commit/4357d0b7) Update dependencies (#36)
- [a14448cb](https://github.com/kubedb/mssql-coordinator/commit/a14448cb) Fix Rewind And Memory leak Issues (#35)
- [245fcde9](https://github.com/kubedb/mssql-coordinator/commit/245fcde9) Update repository config (#34)
- [066da089](https://github.com/kubedb/mssql-coordinator/commit/066da089) Update dependencies (#33)
- [d2c2ef68](https://github.com/kubedb/mssql-coordinator/commit/d2c2ef68) Prepare for release v0.4.0 (#32)
- [0f61b36f](https://github.com/kubedb/mssql-coordinator/commit/0f61b36f) Update dependencies (#31)
- [e98ad26a](https://github.com/kubedb/mssql-coordinator/commit/e98ad26a) Update dependencies (#30)
- [4421b2ab](https://github.com/kubedb/mssql-coordinator/commit/4421b2ab) Update repository config (#29)
- [9d4483f5](https://github.com/kubedb/mssql-coordinator/commit/9d4483f5) Update repository config (#28)
- [adb9f368](https://github.com/kubedb/mssql-coordinator/commit/adb9f368) Update dependencies (#27)
- [1728412c](https://github.com/kubedb/mssql-coordinator/commit/1728412c) Prepare for release v0.3.0 (#26)
- [b6d007ff](https://github.com/kubedb/mssql-coordinator/commit/b6d007ff) Prepare for release v0.3.0-rc.0 (#25)
- [58216f8a](https://github.com/kubedb/mssql-coordinator/commit/58216f8a) Update Client TLS Path for Postgres (#24)
- [a07d255e](https://github.com/kubedb/mssql-coordinator/commit/a07d255e) Raft Version Update And Ops Request Fix (#23)
- [1855cecf](https://github.com/kubedb/mssql-coordinator/commit/1855cecf) Use klog/v2 (#19)
- [727dcc87](https://github.com/kubedb/mssql-coordinator/commit/727dcc87) Use klog/v2
- [eff1c8f0](https://github.com/kubedb/mssql-coordinator/commit/eff1c8f0) Prepare for release v0.2.0 (#16)
- [ffb6ee5e](https://github.com/kubedb/mssql-coordinator/commit/ffb6ee5e) Add Support for Custom UID (#15)
- [72d46511](https://github.com/kubedb/mssql-coordinator/commit/72d46511) Fix spelling
- [6067a826](https://github.com/kubedb/mssql-coordinator/commit/6067a826) Prepare for release v0.1.1 (#14)
- [2fd91dc0](https://github.com/kubedb/mssql-coordinator/commit/2fd91dc0) Prepare for release v0.1.0 (#13)
- [d0eb5419](https://github.com/kubedb/mssql-coordinator/commit/d0eb5419) fix: added basic auth client for raft client http
- [acb1a2e3](https://github.com/kubedb/mssql-coordinator/commit/acb1a2e3) Update KubeDB api (#12)
- [cc923be0](https://github.com/kubedb/mssql-coordinator/commit/cc923be0) fix: added basic auth for http client
- [48204a3d](https://github.com/kubedb/mssql-coordinator/commit/48204a3d) fix: cleanup logs
- [f0cd88d3](https://github.com/kubedb/mssql-coordinator/commit/f0cd88d3) Use alpine as base for prod docker image (#11)
- [98837f30](https://github.com/kubedb/mssql-coordinator/commit/98837f30) fix: revendor with api-mechinary
- [12779aa6](https://github.com/kubedb/mssql-coordinator/commit/12779aa6) fix: updated docker file
- [414905da](https://github.com/kubedb/mssql-coordinator/commit/414905da) fix: replace rsync with cp
- [2f1f7994](https://github.com/kubedb/mssql-coordinator/commit/2f1f7994) fix: pg_rewind failed
- [173bbf94](https://github.com/kubedb/mssql-coordinator/commit/173bbf94) fix: updated server's port 2379 (client),2380 (peer)
- [ab30d53e](https://github.com/kubedb/mssql-coordinator/commit/ab30d53e) fix: license added
- [d0bc0a50](https://github.com/kubedb/mssql-coordinator/commit/d0bc0a50) fix: make lint with constant updated
- [e6dc704b](https://github.com/kubedb/mssql-coordinator/commit/e6dc704b) fix: updated with api-mechinary constant name
- [80a13a5b](https://github.com/kubedb/mssql-coordinator/commit/80a13a5b) fix: primary & replica have different cluster id
- [4345dd34](https://github.com/kubedb/mssql-coordinator/commit/4345dd34) fix: initial access denied before final initialize complete
- [9a68d868](https://github.com/kubedb/mssql-coordinator/commit/9a68d868) fix: working on http basic auth (on progress)
- [22b37d3e](https://github.com/kubedb/mssql-coordinator/commit/22b37d3e) fix: last leader same as new one
- [b798ba4d](https://github.com/kubedb/mssql-coordinator/commit/b798ba4d) fix: single user mode read buffer
- [b7e1f678](https://github.com/kubedb/mssql-coordinator/commit/b7e1f678) fix: make lint
- [8c106f06](https://github.com/kubedb/mssql-coordinator/commit/8c106f06) fix: make gen fmt
- [71108b31](https://github.com/kubedb/mssql-coordinator/commit/71108b31) fix: go mod updated
- [5cdc4e2b](https://github.com/kubedb/mssql-coordinator/commit/5cdc4e2b) fix: error handling fix : log handled
- [3c57af41](https://github.com/kubedb/mssql-coordinator/commit/3c57af41) fix: working on log and error handling
- [1de624df](https://github.com/kubedb/mssql-coordinator/commit/1de624df) fix: updated api-machinery
- [8eb48dbf](https://github.com/kubedb/mssql-coordinator/commit/8eb48dbf) updated sslMode=$SSL_MODE
- [9c745042](https://github.com/kubedb/mssql-coordinator/commit/9c745042) fix: added scram-sha-256
- [4a8817e2](https://github.com/kubedb/mssql-coordinator/commit/4a8817e2) fix: typo
- [f7f3625e](https://github.com/kubedb/mssql-coordinator/commit/f7f3625e) fix: sslMode check done
- [eae73f6d](https://github.com/kubedb/mssql-coordinator/commit/eae73f6d) fix: CLIENT_AUTH_MODE value updated
- [a4c8a7fe](https://github.com/kubedb/mssql-coordinator/commit/a4c8a7fe) fix: auth and sslmode
- [f7e96f1e](https://github.com/kubedb/mssql-coordinator/commit/f7e96f1e) fix ca.crt for client Signed-off-by: Emon46 <emon@appscode.com>
- [daab4f78](https://github.com/kubedb/mssql-coordinator/commit/daab4f78) fix: set connection string to verify-fulll fix: remove restore.sh from all version
- [b085140b](https://github.com/kubedb/mssql-coordinator/commit/b085140b) fix: clean up logs
- [cd1904a3](https://github.com/kubedb/mssql-coordinator/commit/cd1904a3) fix: handle DoReinitialization func returning error in failover
- [1d188de5](https://github.com/kubedb/mssql-coordinator/commit/1d188de5) fix: api mechinary updated for  removing archiver spec
- [cc5f7482](https://github.com/kubedb/mssql-coordinator/commit/cc5f7482) fix: added demote immediate function
- [887eddca](https://github.com/kubedb/mssql-coordinator/commit/887eddca) fix: added pgREcovery check for leader test
- [beb56969](https://github.com/kubedb/mssql-coordinator/commit/beb56969) fix: removed unnecessary functions fix: scripts Signed-off-by: Emon46 <emon@appscode.com>
- [24e5e2aa](https://github.com/kubedb/mssql-coordinator/commit/24e5e2aa) fix: script modified
- [49637489](https://github.com/kubedb/mssql-coordinator/commit/49637489) fix: timeline issue
- [c418bc0c](https://github.com/kubedb/mssql-coordinator/commit/c418bc0c) Remove .idea folder
- [75662379](https://github.com/kubedb/mssql-coordinator/commit/75662379) Add license and Makefile (#8)
- [898f76f0](https://github.com/kubedb/mssql-coordinator/commit/898f76f0) fix: added pass auth for users in pg_hba.conf
- [50e0ddcd](https://github.com/kubedb/mssql-coordinator/commit/50e0ddcd) working with 9,10,11,12,13
- [d1318c92](https://github.com/kubedb/mssql-coordinator/commit/d1318c92) added script for 9,10,11,12,13 fix : syntex error issue in 9 fix : updated pg_rewind error message check
- [f13ea0fd](https://github.com/kubedb/mssql-coordinator/commit/f13ea0fd) fix: working with version 12
- [6d87091c](https://github.com/kubedb/mssql-coordinator/commit/6d87091c) fixing for updated version 12 and 13
- [a95e5d97](https://github.com/kubedb/mssql-coordinator/commit/a95e5d97) fix: config recovery after basebackup in normal flow
- [de14b8b3](https://github.com/kubedb/mssql-coordinator/commit/de14b8b3) update: variable check SSL_MODE
- [b19426c1](https://github.com/kubedb/mssql-coordinator/commit/b19426c1) added field for hearttick and leadertick
- [ca1c44af](https://github.com/kubedb/mssql-coordinator/commit/ca1c44af) fix: revondoring done. label updated
- [e5d59292](https://github.com/kubedb/mssql-coordinator/commit/e5d59292) fix: final working version. fixed: raft vendoring . added vendor from appscode fix: error handled for pg_rewind, pg_dump, start, stop, single user mode
- [f027542a](https://github.com/kubedb/mssql-coordinator/commit/f027542a) Signed-off-by: Emon331046 <emon@appscode.com>
- [db071c27](https://github.com/kubedb/mssql-coordinator/commit/db071c27) fix : working version with pg rewind check and other stuffs. note: this is the final one
- [2f03442a](https://github.com/kubedb/mssql-coordinator/commit/2f03442a) added process check
- [4cca7bf4](https://github.com/kubedb/mssql-coordinator/commit/4cca7bf4) fix all crush prev-link isssue
- [dd92164f](https://github.com/kubedb/mssql-coordinator/commit/dd92164f) fixed single user mode start
- [5980f861](https://github.com/kubedb/mssql-coordinator/commit/5980f861) working with doing always pg_rewind note: needPgRewind is always true here
- [ca6dabd7](https://github.com/kubedb/mssql-coordinator/commit/ca6dabd7) fixing primary after recover prev-link invalid wal Signed-off-by: Emon46 <emon@appscode.com>
- [8e678581](https://github.com/kubedb/mssql-coordinator/commit/8e678581) fixing prev-link record error
- [c5518602](https://github.com/kubedb/mssql-coordinator/commit/c5518602) fix ... Signed-off-by: Emon331046 <emon@appscode.com>
- [6c5c31ae](https://github.com/kubedb/mssql-coordinator/commit/6c5c31ae) fix some pg_rewind issue
- [ef9cdb17](https://github.com/kubedb/mssql-coordinator/commit/ef9cdb17) fix recover demote follower
- [8345e968](https://github.com/kubedb/mssql-coordinator/commit/8345e968) pg_rewind and status and stop start func fix Signed-off-by: Emon331046 <emon@appscode.com>
- [8851de41](https://github.com/kubedb/mssql-coordinator/commit/8851de41) debuging codebase Signed-off-by: Emon46 <emon@appscode.com>
- [657b5c03](https://github.com/kubedb/mssql-coordinator/commit/657b5c03) fix : always delete pod problem
- [a2e0c15e](https://github.com/kubedb/mssql-coordinator/commit/a2e0c15e) working on pg rewind Signed-off-by: Emon331046 <emon@appscode.com>
- [692a6f8f](https://github.com/kubedb/mssql-coordinator/commit/692a6f8f) working on do pg rewind Signed-off-by: Emon331046 <emon@appscode.com>
- [50f34767](https://github.com/kubedb/mssql-coordinator/commit/50f34767) added pg_rewind check
- [fcaf8261](https://github.com/kubedb/mssql-coordinator/commit/fcaf8261) unifnished : pg_dump and Pg_rewind Signed-off-by: Emon46 <emon@appscode.com>
- [2c7005ce](https://github.com/kubedb/mssql-coordinator/commit/2c7005ce) adding og_rewind
- [62372a23](https://github.com/kubedb/mssql-coordinator/commit/62372a23) need a fix
- [c0459bf4](https://github.com/kubedb/mssql-coordinator/commit/c0459bf4) fix: leader transfer api added Signed-off-by: Emon331046 <emon@appscode.com>
- [21f016d0](https://github.com/kubedb/mssql-coordinator/commit/21f016d0) need to fix Signed-off-by: Emon331046 <emon@appscode.com>
- [afea498c](https://github.com/kubedb/mssql-coordinator/commit/afea498c) clean up codebase
- [3166aa2d](https://github.com/kubedb/mssql-coordinator/commit/3166aa2d) change pg_hba.conf for localhost
- [af5719eb](https://github.com/kubedb/mssql-coordinator/commit/af5719eb)  need to fix change pg_hba.conf
- [e91b1214](https://github.com/kubedb/mssql-coordinator/commit/e91b1214) change pghb from root to postgres for certs
- [97547ff1](https://github.com/kubedb/mssql-coordinator/commit/97547ff1) change owner from root to postgres for certs
- [72101b53](https://github.com/kubedb/mssql-coordinator/commit/72101b53) need to fix
- [46119298](https://github.com/kubedb/mssql-coordinator/commit/46119298) fix: vendor updated
- [5456ed6f](https://github.com/kubedb/mssql-coordinator/commit/5456ed6f) fix: svc for standby
- [b700c1fe](https://github.com/kubedb/mssql-coordinator/commit/b700c1fe) fix: svc for standby
- [9299ea6c](https://github.com/kubedb/mssql-coordinator/commit/9299ea6c) fix: vendoring apimachinery for constants
- [2bee5907](https://github.com/kubedb/mssql-coordinator/commit/2bee5907) fix: added support for v9
- [5049d829](https://github.com/kubedb/mssql-coordinator/commit/5049d829) fix: added support for v9
- [273d3ca9](https://github.com/kubedb/mssql-coordinator/commit/273d3ca9) fix: added client for converting a node learner or candidate
- [cbe5c267](https://github.com/kubedb/mssql-coordinator/commit/cbe5c267) added makeLearner func
- [b78e4245](https://github.com/kubedb/mssql-coordinator/commit/b78e4245) fix: the trigger-file issue
- [1b32281c](https://github.com/kubedb/mssql-coordinator/commit/1b32281c) fix: old leader restart fixed
- [6e79582b](https://github.com/kubedb/mssql-coordinator/commit/6e79582b) fix: crashing after running script fixed
- [2049963c](https://github.com/kubedb/mssql-coordinator/commit/2049963c) need to fix the script part
- [6d7985c7](https://github.com/kubedb/mssql-coordinator/commit/6d7985c7) added scripts in leader-election container
- [df37d25c](https://github.com/kubedb/mssql-coordinator/commit/df37d25c) added leader port
- [72de772f](https://github.com/kubedb/mssql-coordinator/commit/72de772f) added leader port
- [746914ff](https://github.com/kubedb/mssql-coordinator/commit/746914ff) leader test
- [12a635d1](https://github.com/kubedb/mssql-coordinator/commit/12a635d1) leader test
- [ce3787db](https://github.com/kubedb/mssql-coordinator/commit/ce3787db) leader test
- [02677281](https://github.com/kubedb/mssql-coordinator/commit/02677281) leader test
- [b64c6e31](https://github.com/kubedb/mssql-coordinator/commit/b64c6e31) leader test
- [dd970354](https://github.com/kubedb/mssql-coordinator/commit/dd970354) leader test
- [84686fc5](https://github.com/kubedb/mssql-coordinator/commit/84686fc5) leader test
- [57fb4982](https://github.com/kubedb/mssql-coordinator/commit/57fb4982) leader test
- [eeebe8ff](https://github.com/kubedb/mssql-coordinator/commit/eeebe8ff) leader test



## [kubedb/mysql](https://github.com/kubedb/mysql)

### [v0.38.0](https://github.com/kubedb/mysql/releases/tag/v0.38.0)

- [cf8f4d83](https://github.com/kubedb/mysql/commit/cf8f4d83a) Prepare for release v0.38.0 (#622)
- [26f4d72b](https://github.com/kubedb/mysql/commit/26f4d72b7) Prepare for release v0.38.0 (#621)
- [6bcb3ac4](https://github.com/kubedb/mysql/commit/6bcb3ac48) Remove redundant log for monitoring agent not found (#620)
- [c1593bf8](https://github.com/kubedb/mysql/commit/c1593bf83) Use deps (#619)
- [7278772d](https://github.com/kubedb/mysql/commit/7278772d6) Use Go 1.22 (#618)



## [kubedb/mysql-archiver](https://github.com/kubedb/mysql-archiver)

### [v0.6.0](https://github.com/kubedb/mysql-archiver/releases/tag/v0.6.0)

- [971d27c](https://github.com/kubedb/mysql-archiver/commit/971d27c) Prepare for release v0.6.0 (#29)
- [9a6f2d1](https://github.com/kubedb/mysql-archiver/commit/9a6f2d1) Prepare for release v0.6.0 (#28)
- [3309740](https://github.com/kubedb/mysql-archiver/commit/3309740) Use Go 1.22 (#27)



## [kubedb/mysql-coordinator](https://github.com/kubedb/mysql-coordinator)

### [v0.23.0](https://github.com/kubedb/mysql-coordinator/releases/tag/v0.23.0)

- [91dbda5e](https://github.com/kubedb/mysql-coordinator/commit/91dbda5e) Prepare for release v0.23.0 (#110)
- [8f19e568](https://github.com/kubedb/mysql-coordinator/commit/8f19e568) Use Go 1.22 (#108)



## [kubedb/mysql-csi-snapshotter-plugin](https://github.com/kubedb/mysql-csi-snapshotter-plugin)

### [v0.6.0](https://github.com/kubedb/mysql-csi-snapshotter-plugin/releases/tag/v0.6.0)

- [145c9e6](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/145c9e6) Prepare for release v0.6.0 (#17)
- [804abd7](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/804abd7) Prepare for release v0.6.0 (#16)
- [5dbc6c8](https://github.com/kubedb/mysql-csi-snapshotter-plugin/commit/5dbc6c8) Use Go 1.22 (#15)



## [kubedb/mysql-restic-plugin](https://github.com/kubedb/mysql-restic-plugin)

### [v0.8.0](https://github.com/kubedb/mysql-restic-plugin/releases/tag/v0.8.0)

- [3b2d338](https://github.com/kubedb/mysql-restic-plugin/commit/3b2d338) Prepare for release v0.8.0 (#36)
- [c684322](https://github.com/kubedb/mysql-restic-plugin/commit/c684322) Prepare for release v0.8.0 (#35)
- [3cb8489](https://github.com/kubedb/mysql-restic-plugin/commit/3cb8489) Use restic 0.16.4 (#33)
- [cec2dc5](https://github.com/kubedb/mysql-restic-plugin/commit/cec2dc5) Use Go 1.22 (#32)



## [kubedb/mysql-router-init](https://github.com/kubedb/mysql-router-init)

### [v0.23.0](https://github.com/kubedb/mysql-router-init/releases/tag/v0.23.0)

- [2e1631a](https://github.com/kubedb/mysql-router-init/commit/2e1631a) Use Go 1.22 (#42)



## [kubedb/ops-manager](https://github.com/kubedb/ops-manager)

### [v0.32.0](https://github.com/kubedb/ops-manager/releases/tag/v0.32.0)

- [dcea6a22](https://github.com/kubedb/ops-manager/commit/dcea6a223) Prepare for release v0.32.0 (#561)
- [6ebc03dc](https://github.com/kubedb/ops-manager/commit/6ebc03dc8) Add vertical scaling and volume expansion opsrequest (#560)
- [722f44d4](https://github.com/kubedb/ops-manager/commit/722f44d44) Resume DB even if opsReq failed (#559)
- [9db28273](https://github.com/kubedb/ops-manager/commit/9db28273f) Add SingleStore TLS (#555)
- [b5e759da](https://github.com/kubedb/ops-manager/commit/b5e759da6) Add Pgpool TLS (#558)
- [31804581](https://github.com/kubedb/ops-manager/commit/318045819) Remove inlineConfig; Use applyConfig only (#556)



## [kubedb/percona-xtradb](https://github.com/kubedb/percona-xtradb)

### [v0.32.0](https://github.com/kubedb/percona-xtradb/releases/tag/v0.32.0)

- [45ed6cba](https://github.com/kubedb/percona-xtradb/commit/45ed6cba2) Prepare for release v0.32.0 (#365)
- [12ffd594](https://github.com/kubedb/percona-xtradb/commit/12ffd594d) Prepare for release v0.32.0 (#364)
- [495e6172](https://github.com/kubedb/percona-xtradb/commit/495e61720) Use deps (#363)
- [0b9f1955](https://github.com/kubedb/percona-xtradb/commit/0b9f1955a) Use Go 1.22 (#362)



## [kubedb/percona-xtradb-coordinator](https://github.com/kubedb/percona-xtradb-coordinator)

### [v0.18.0](https://github.com/kubedb/percona-xtradb-coordinator/releases/tag/v0.18.0)

- [f20706df](https://github.com/kubedb/percona-xtradb-coordinator/commit/f20706df) Prepare for release v0.18.0 (#70)
- [2bb57896](https://github.com/kubedb/percona-xtradb-coordinator/commit/2bb57896) Prepare for release v0.18.0 (#69)
- [d5739deb](https://github.com/kubedb/percona-xtradb-coordinator/commit/d5739deb) Use Go 1.22 (#68)



## [kubedb/pg-coordinator](https://github.com/kubedb/pg-coordinator)

### [v0.29.0](https://github.com/kubedb/pg-coordinator/releases/tag/v0.29.0)

- [33a82733](https://github.com/kubedb/pg-coordinator/commit/33a82733) Prepare for release v0.29.0 (#163)
- [2deeaed4](https://github.com/kubedb/pg-coordinator/commit/2deeaed4) Prepare for release v0.29.0 (#162)
- [15139ecf](https://github.com/kubedb/pg-coordinator/commit/15139ecf) Use Go 1.22 (#161)



## [kubedb/pgbouncer](https://github.com/kubedb/pgbouncer)

### [v0.32.0](https://github.com/kubedb/pgbouncer/releases/tag/v0.32.0)

- [8ccd7b74](https://github.com/kubedb/pgbouncer/commit/8ccd7b74) Prepare for release v0.32.0 (#326)
- [2c84ca8c](https://github.com/kubedb/pgbouncer/commit/2c84ca8c) Use deps (#324)
- [4bc53222](https://github.com/kubedb/pgbouncer/commit/4bc53222) Use Go 1.22 (#323)



## [kubedb/pgpool](https://github.com/kubedb/pgpool)

### [v0.0.9](https://github.com/kubedb/pgpool/releases/tag/v0.0.9)

- [71982d7](https://github.com/kubedb/pgpool/commit/71982d7) Prepare for release v0.0.9 (#27)
- [544327d](https://github.com/kubedb/pgpool/commit/544327d) Prepare for release v0.0.9 (#26)
- [dd8144f](https://github.com/kubedb/pgpool/commit/dd8144f) Add TLS support (#25)
- [cabc3db](https://github.com/kubedb/pgpool/commit/cabc3db) Set Defaults to db on reconcile (#24)
- [8fb6dfe](https://github.com/kubedb/pgpool/commit/8fb6dfe) Add PetSet in daily workflow and makefile (#11)
- [a510bbf](https://github.com/kubedb/pgpool/commit/a510bbf) Use deps (#23)
- [89285b1](https://github.com/kubedb/pgpool/commit/89285b1) Use Go 1.22 (#22)
- [279978f](https://github.com/kubedb/pgpool/commit/279978f) Remove license check for webhook-server (#21)



## [kubedb/postgres](https://github.com/kubedb/postgres)

### [v0.45.0](https://github.com/kubedb/postgres/releases/tag/v0.45.0)

- [78cec238](https://github.com/kubedb/postgres/commit/78cec238a) Prepare for release v0.45.0 (#731)
- [b939cb7f](https://github.com/kubedb/postgres/commit/b939cb7f8) Prepare for release v0.45.0 (#730)
- [22617a14](https://github.com/kubedb/postgres/commit/22617a148) Remove unnecessary log (#729)
- [6627c462](https://github.com/kubedb/postgres/commit/6627c4628) Use deps (#728)
- [3fcdc899](https://github.com/kubedb/postgres/commit/3fcdc899e) Use Go 1.22 (#727)
- [e0627ad8](https://github.com/kubedb/postgres/commit/e0627ad89) Fix WebHook (#726)



## [kubedb/postgres-archiver](https://github.com/kubedb/postgres-archiver)

### [v0.6.0](https://github.com/kubedb/postgres-archiver/releases/tag/v0.6.0)

- [db6968c](https://github.com/kubedb/postgres-archiver/commit/db6968c) Prepare for release v0.6.0 (#28)
- [a05987c](https://github.com/kubedb/postgres-archiver/commit/a05987c) Prepare for release v0.6.0 (#27)
- [b8cb6d3](https://github.com/kubedb/postgres-archiver/commit/b8cb6d3) Use Go 1.22 (#26)



## [kubedb/postgres-csi-snapshotter-plugin](https://github.com/kubedb/postgres-csi-snapshotter-plugin)

### [v0.6.0](https://github.com/kubedb/postgres-csi-snapshotter-plugin/releases/tag/v0.6.0)

- [a4916b1](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/a4916b1) Prepare for release v0.6.0 (#26)
- [1bf318a](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/1bf318a) Prepare for release v0.6.0 (#25)
- [b3921a5](https://github.com/kubedb/postgres-csi-snapshotter-plugin/commit/b3921a5) Use Go 1.22 (#24)



## [kubedb/postgres-restic-plugin](https://github.com/kubedb/postgres-restic-plugin)

### [v0.8.0](https://github.com/kubedb/postgres-restic-plugin/releases/tag/v0.8.0)

- [8addbb2](https://github.com/kubedb/postgres-restic-plugin/commit/8addbb2) Prepare for release v0.8.0 (#32)
- [965f18c](https://github.com/kubedb/postgres-restic-plugin/commit/965f18c) Prepare for release v0.8.0 (#31)
- [48f21b1](https://github.com/kubedb/postgres-restic-plugin/commit/48f21b1) Add ClientAuthMode MD5 support for TLS (#29)
- [ef14301](https://github.com/kubedb/postgres-restic-plugin/commit/ef14301) Refactor (#30)
- [764c43b](https://github.com/kubedb/postgres-restic-plugin/commit/764c43b) Tls Fix (#28)
- [c717c8c](https://github.com/kubedb/postgres-restic-plugin/commit/c717c8c) Use restic 0.16.4 (#27)
- [42b6af6](https://github.com/kubedb/postgres-restic-plugin/commit/42b6af6) Use Go 1.22 (#26)



## [kubedb/provider-aws](https://github.com/kubedb/provider-aws)

### [v0.7.0](https://github.com/kubedb/provider-aws/releases/tag/v0.7.0)

- [f102958](https://github.com/kubedb/provider-aws/commit/f102958) Use Go 1.22 (#14)



## [kubedb/provider-azure](https://github.com/kubedb/provider-azure)

### [v0.7.0](https://github.com/kubedb/provider-azure/releases/tag/v0.7.0)

- [387e3f2](https://github.com/kubedb/provider-azure/commit/387e3f2) Use Go 1.22 (#5)



## [kubedb/provider-gcp](https://github.com/kubedb/provider-gcp)

### [v0.7.0](https://github.com/kubedb/provider-gcp/releases/tag/v0.7.0)

- [40683f8](https://github.com/kubedb/provider-gcp/commit/40683f8) Use Go 1.22 (#5)



## [kubedb/provisioner](https://github.com/kubedb/provisioner)

### [v0.45.0](https://github.com/kubedb/provisioner/releases/tag/v0.45.0)

- [c73f1e11](https://github.com/kubedb/provisioner/commit/c73f1e111) Prepare for release v0.45.0 (#93)
- [564a8ca2](https://github.com/kubedb/provisioner/commit/564a8ca20) Register MSSQLServer (#92)
- [ddc05af4](https://github.com/kubedb/provisioner/commit/ddc05af4e) Initialize default client (#90)



## [kubedb/proxysql](https://github.com/kubedb/proxysql)

### [v0.32.0](https://github.com/kubedb/proxysql/releases/tag/v0.32.0)

- [859048d6](https://github.com/kubedb/proxysql/commit/859048d6b) Prepare for release v0.32.0 (#342)
- [c3348d30](https://github.com/kubedb/proxysql/commit/c3348d307) Use deps (#341)
- [c8409775](https://github.com/kubedb/proxysql/commit/c8409775c) Use Go 1.22 (#340)



## [kubedb/rabbitmq](https://github.com/kubedb/rabbitmq)

### [v0.0.11](https://github.com/kubedb/rabbitmq/releases/tag/v0.0.11)

- [b529602](https://github.com/kubedb/rabbitmq/commit/b529602) Prepare for release v0.0.11 (#28)
- [26b5e6f](https://github.com/kubedb/rabbitmq/commit/26b5e6f) Prepare for release v0.0.11 (#27)
- [630c0ef](https://github.com/kubedb/rabbitmq/commit/630c0ef) Patch AuthsecretRef to DB yaml (#26)
- [9248270](https://github.com/kubedb/rabbitmq/commit/9248270) Set Defaults to db on reconcile (#24)
- [ef00643](https://github.com/kubedb/rabbitmq/commit/ef00643) Add support for PDB and extend default plugin supports (#21)
- [bf35fcd](https://github.com/kubedb/rabbitmq/commit/bf35fcd) Use deps (#23)
- [048dbbe](https://github.com/kubedb/rabbitmq/commit/048dbbe) Use Go 1.22 (#22)



## [kubedb/redis](https://github.com/kubedb/redis)

### [v0.38.0](https://github.com/kubedb/redis/releases/tag/v0.38.0)

- [6b29c6ab](https://github.com/kubedb/redis/commit/6b29c6ab2) Prepare for release v0.38.0 (#538)
- [5403c046](https://github.com/kubedb/redis/commit/5403c0460) Prepare for release v0.38.0 (#537)
- [b9a7cbc7](https://github.com/kubedb/redis/commit/b9a7cbc70) Remove redundant log for monitoring agent not found (#536)
- [56a9e21f](https://github.com/kubedb/redis/commit/56a9e21f3) Use deps (#535)
- [37c0f40d](https://github.com/kubedb/redis/commit/37c0f40d7) Use Go 1.22 (#534)
- [5645fb9b](https://github.com/kubedb/redis/commit/5645fb9bd) Remove License Check from Webhook Server (#533)



## [kubedb/redis-coordinator](https://github.com/kubedb/redis-coordinator)

### [v0.24.0](https://github.com/kubedb/redis-coordinator/releases/tag/v0.24.0)

- [7542fe79](https://github.com/kubedb/redis-coordinator/commit/7542fe79) Prepare for release v0.24.0 (#101)
- [7901b594](https://github.com/kubedb/redis-coordinator/commit/7901b594) Prepare for release v0.24.0 (#100)
- [c7809cc0](https://github.com/kubedb/redis-coordinator/commit/c7809cc0) Use Go 1.22 (#99)



## [kubedb/redis-restic-plugin](https://github.com/kubedb/redis-restic-plugin)

### [v0.8.0](https://github.com/kubedb/redis-restic-plugin/releases/tag/v0.8.0)

- [76a22e9](https://github.com/kubedb/redis-restic-plugin/commit/76a22e9) Prepare for release v0.8.0 (#31)
- [de1eb07](https://github.com/kubedb/redis-restic-plugin/commit/de1eb07) Prepare for release v0.8.0 (#30)
- [117aae0](https://github.com/kubedb/redis-restic-plugin/commit/117aae0) Use restic 0.16.4 (#29)
- [e855f79](https://github.com/kubedb/redis-restic-plugin/commit/e855f79) Use Go 1.22 (#28)



## [kubedb/replication-mode-detector](https://github.com/kubedb/replication-mode-detector)

### [v0.32.0](https://github.com/kubedb/replication-mode-detector/releases/tag/v0.32.0)

- [ead0dae3](https://github.com/kubedb/replication-mode-detector/commit/ead0dae3) Prepare for release v0.32.0 (#267)
- [803fe52b](https://github.com/kubedb/replication-mode-detector/commit/803fe52b) Use Go 1.22 (#265)



## [kubedb/schema-manager](https://github.com/kubedb/schema-manager)

### [v0.21.0](https://github.com/kubedb/schema-manager/releases/tag/v0.21.0)

- [217af272](https://github.com/kubedb/schema-manager/commit/217af272) Prepare for release v0.21.0 (#110)



## [kubedb/singlestore](https://github.com/kubedb/singlestore)

### [v0.0.9](https://github.com/kubedb/singlestore/releases/tag/v0.0.9)

- [499dfa6](https://github.com/kubedb/singlestore/commit/499dfa6) Prepare for release v0.0.9 (#27)
- [4082540](https://github.com/kubedb/singlestore/commit/4082540) Prepare for release v0.0.9 (#26)
- [99b143b](https://github.com/kubedb/singlestore/commit/99b143b) Add SingleStore TLS (#25)
- [ae86b87](https://github.com/kubedb/singlestore/commit/ae86b87) Set Defaults to db on reconcile (#23)
- [3c28add](https://github.com/kubedb/singlestore/commit/3c28add) Use deps (#22)
- [d9cce0e](https://github.com/kubedb/singlestore/commit/d9cce0e) Use Go 1.22 (#21)
- [d113503](https://github.com/kubedb/singlestore/commit/d113503) Remove License check from Webhook-Server (#20)



## [kubedb/singlestore-coordinator](https://github.com/kubedb/singlestore-coordinator)

### [v0.0.8](https://github.com/kubedb/singlestore-coordinator/releases/tag/v0.0.8)

- [f85b392](https://github.com/kubedb/singlestore-coordinator/commit/f85b392) Prepare for release v0.0.8 (#15)
- [4aa8378](https://github.com/kubedb/singlestore-coordinator/commit/4aa8378) Prepare for release v0.0.8 (#14)
- [265105a](https://github.com/kubedb/singlestore-coordinator/commit/265105a) Add SingleStore TLS (#13)
- [123d55b](https://github.com/kubedb/singlestore-coordinator/commit/123d55b) Use Go 1.22 (#12)



## [kubedb/singlestore-restic-plugin](https://github.com/kubedb/singlestore-restic-plugin)

### [v0.3.0](https://github.com/kubedb/singlestore-restic-plugin/releases/tag/v0.3.0)

- [032d33e](https://github.com/kubedb/singlestore-restic-plugin/commit/032d33e) Prepare for release v0.3.0 (#11)
- [e7cbc0a](https://github.com/kubedb/singlestore-restic-plugin/commit/e7cbc0a) Prepare for release v0.3.0 (#10)
- [cf1f07e](https://github.com/kubedb/singlestore-restic-plugin/commit/cf1f07e) Add SingleStore TLS (#9)
- [df11e7f](https://github.com/kubedb/singlestore-restic-plugin/commit/df11e7f) Use deps (#7)
- [846c188](https://github.com/kubedb/singlestore-restic-plugin/commit/846c188) Use Go 1.22 (#6)



## [kubedb/solr](https://github.com/kubedb/solr)

### [v0.0.11](https://github.com/kubedb/solr/releases/tag/v0.0.11)

- [3030c02](https://github.com/kubedb/solr/commit/3030c02) Prepare for release v0.0.11 (#25)
- [e2d4c78](https://github.com/kubedb/solr/commit/e2d4c78) Set Defaults to db on reconcile (#23)
- [4566b2c](https://github.com/kubedb/solr/commit/4566b2c) Use deps (#22)
- [22e3924](https://github.com/kubedb/solr/commit/22e3924) Use Go 1.22 (#21)



## [kubedb/tests](https://github.com/kubedb/tests)

### [v0.30.0](https://github.com/kubedb/tests/releases/tag/v0.30.0)

- [34a5843d](https://github.com/kubedb/tests/commit/34a5843d) Prepare for release v0.30.0 (#320)
- [975c6ab2](https://github.com/kubedb/tests/commit/975c6ab2) Prepare for release v0.30.0 (#319)
- [6a301a03](https://github.com/kubedb/tests/commit/6a301a03) Remove inlineConfig (#318)
- [6101e637](https://github.com/kubedb/tests/commit/6101e637) Fix certstore invoker. (#317)
- [1a0556ce](https://github.com/kubedb/tests/commit/1a0556ce) Use Go 1.22 (#316)
- [886fcdf7](https://github.com/kubedb/tests/commit/886fcdf7) add storage-class in pg-logical-replication (#315)
- [a6250835](https://github.com/kubedb/tests/commit/a6250835) Add Version Upgrade Test Case for PG (#295)
- [34822e19](https://github.com/kubedb/tests/commit/34822e19) Add Pgpool tests (#302)
- [ed294bf0](https://github.com/kubedb/tests/commit/ed294bf0) Fix build (#314)



## [kubedb/ui-server](https://github.com/kubedb/ui-server)

### [v0.21.0](https://github.com/kubedb/ui-server/releases/tag/v0.21.0)

- [bd2a1ab4](https://github.com/kubedb/ui-server/commit/bd2a1ab4) Prepare for release v0.21.0 (#117)
- [21449110](https://github.com/kubedb/ui-server/commit/21449110) Vendor automaxprocs
- [c5d30b0c](https://github.com/kubedb/ui-server/commit/c5d30b0c) Use Go 1.22 (#116)



## [kubedb/webhook-server](https://github.com/kubedb/webhook-server)

### [v0.21.0](https://github.com/kubedb/webhook-server/releases/tag/v0.21.0)

- [bf41d483](https://github.com/kubedb/webhook-server/commit/bf41d483) Prepare for release v0.21.0 (#105)
- [ebb428b7](https://github.com/kubedb/webhook-server/commit/ebb428b7) Register RabbitMQ autoscaler & MSSQLServer (#104)
- [77e54fc3](https://github.com/kubedb/webhook-server/commit/77e54fc3) Use deps (#103)
- [3c0d0c2e](https://github.com/kubedb/webhook-server/commit/3c0d0c2e) Use Go 1.22 (#102)



## [kubedb/zookeeper](https://github.com/kubedb/zookeeper)

### [v0.0.10](https://github.com/kubedb/zookeeper/releases/tag/v0.0.10)

- [ce8f982](https://github.com/kubedb/zookeeper/commit/ce8f982) Prepare for release v0.0.10 (#22)
- [4f111a3](https://github.com/kubedb/zookeeper/commit/4f111a3) Set Default From Operator (#21)
- [1b3ea42](https://github.com/kubedb/zookeeper/commit/1b3ea42) Use deps (#20)
- [5094c26](https://github.com/kubedb/zookeeper/commit/5094c26) Use Go 1.22 (#19)



## [kubedb/zookeeper-restic-plugin](https://github.com/kubedb/zookeeper-restic-plugin)

### [v0.1.0](https://github.com/kubedb/zookeeper-restic-plugin/releases/tag/v0.1.0)

- [0a3b8a6](https://github.com/kubedb/zookeeper-restic-plugin/commit/0a3b8a6) Prepare for release v0.1.0 (#7)
- [cc4e071](https://github.com/kubedb/zookeeper-restic-plugin/commit/cc4e071) Prepare for release v0.1.0 (#6)




