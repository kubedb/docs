---
title: KubeDB Overview
menu:
  docs_{{ .version }}:
    identifier: overview-readme
    name: Readme
    parent: overview
    weight: -1
menu_name: docs_{{ .version }}
section_menu_id: overview
url: /docs/{{ .version }}/overview/
aliases:
  - /docs/{{ .version }}/overview/README/
---

# KubeDB by AppsCode

Kubernetes has emerged as the de-facto way to deploy modern containerized apps on cloud or on-premises. _"Despite all that growth on the application layer, the data layer hasn’t gotten as much traction with containerization"_ - [Google](https://cloud.google.com/blog/products/databases/to-run-or-not-to-run-a-database-on-kubernetes-what-to-consider). That’s not surprising, since handling things like state (the database), availability to other layers of the application, and redundancy for a database makes it challenging to run a database in a distributed environment like Kubernetes.

However, many developers want to treat data infrastructure the same as application stacks. Operators want to use the same tools for databases and applications and get the same benefits as the application layer in the data layer: rapid spin-up and repeatability across environments. This is where KubeDB by AppsCode comes as a solution.

KubeDB by AppsCode is a production-grade cloud-native database management solution for Kubernetes. KubeDB simplifies and automates routine database tasks such as provisioning, upgrading, patching, scaling, volume expansion, backup, recovery, failure detection, and repair for various popular databases on private and public clouds. It frees you to focus on your applications so you can give them the fast performance, high availability, security and compatibility they need.

KubeDB provides you with many familiar database engines to choose from, including **PostgreSQL**, **MySQL**, **MongoDB**, **MariaDB**, **Microsoft SQL Server**, **Elasticsearch**, **OpenSearch**, **Redis**, **Memcached**, **Percona XtraDB**, **Druid**, **FerretDB**, **Kafka**, **PgBouncer**, **Pgpool**, **ProxySQL**, **RabbitMQ**, **SingleStore**, **Solr** and **ZooKeeper** . KubeDB’s native integration with Kubernetes makes a unique solution compared to competitive solutions from cloud providers and database vendors.
