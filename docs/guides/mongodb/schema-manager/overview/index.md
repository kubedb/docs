---
title: MongoDB Schema Manager Overview
menu:
  docs_{{ .version }}:
    identifier: schema-manager-overview
    name: Overview
    parent: mg-schema-manager
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}


## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [MongoDBDatabase](/docs/guides/mongodb/concepts/mongodbdatabase.md)


## What is Schema Manager

`Schema Manager` is a Kubernetes operator developed by AppsCode that implements multi-tenancy inside KubeDB provisioned database servers like MySQL, MariaDB, PosgreSQL and MongoDB etc. With `Schema Manager` one can create database into specific database server. An user will also be created with KubeVault and assigned to that database. Using the newly created user credential one can access the database and run operations into it. One may pass the database server reference, configuration, user access policy through a single yaml and `Schema Manager` will do all the task above mentioned. `Schema Manager` also allows initializing the database and restore snapshot while bootstrap.


## How MongoDB Schema Manager Process Works

The following diagram shows how MongoDB Schema Manager process worked. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="MongoDB Schema Mananger Diagram" src="/docs/guides/mongodb/schema-manager/overview/images/mongodb-schema-manager-diagram.svg">
<figcaption align="center">Fig: Process of MongoDB Schema Manager</figcaption>
</figure>

The process consists of the following steps:

1. At first the user will deploy a `MongoDBDatabase` object.

2. Once a `MongoDBDatabase` object is deployed to the cluster, the `Schema Manager` operator first verifies if it has the required permission to be able to interact with the referred database-server by checking `Double-OptIn`. After the `Double-OptIn` verification `Schema Manager` operator checks in the `MongoDB` server if the target database is already present or not. If the database already present there, then the `MongoDBDatabase` object will be immediately denied. 

3. Once everything is ok in the `MongoDB` server side, then the target database will be created and an entry for that will be entered in the `kubedb_system` database.

4. Then `Schema Manager` operator creates a `MongoDB Role`.

5. `Vault` operator always watches for a Database `Role`.

6. Once `Vault` operator finds a Database `Role`, it creates a `Secret` for that `Role`.

7. After this process, the `Vault` operator creates a `User` in the `MongoDB` server. The user gets all the privileges on our target database and its credentials are served with the `Secret`. The user credentials secret reference is patched with the `MongoDBDatabase` object yaml in the `.status.authSecret.name` field.

8. If there is any `init script` associated with the `MongoDBDatabase` object, it will be executed in this step with the `Schema Manager` operator. 

9. The user can also provide a `snapshot` reference for initialization. In that case `Schema Manager` operator fetches necessary `appbinding`, `secrets`, `repository`. 

10. `Stash` operator watches for a `Restoresession`.

11. Once `Stash` operator finds a `Restoresession`, it Restores the targeted database with the `Snapshot`.

In the next doc, we are going to show a step by step guide of using MongoDB Schema Manager with KubeDB.