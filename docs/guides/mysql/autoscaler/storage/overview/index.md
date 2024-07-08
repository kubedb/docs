---
title: MySQL Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: mguides-mysql-autoscaling-storage-overview
    name: Overview
    parent: guides-mysql-autoscaling-storage
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQL Vertical Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database storage using `mysqlautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MySQL](/docs/guides/mysql/concepts/mysqldatabase)
  - [MySQLAutoscaler](/docs/guides/mysql/concepts/autoscaler)
  - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest)

## How Storage Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `MySQL` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Storage Autoscaling process of MySQL" src="/docs/guides/mysql/autoscaler/storage/overview/images/storage-autoscaling.jpg">
<figcaption align="center">Fig: Storage Autoscaling process of MySQL</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `MySQL` Custom Resource (CR).

2. `KubeDB` Community operator watches the `MySQL` CR.

3. When the operator finds a `MySQL` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Each PetSet creates a Persistent Volume according to the Volume Claim Template provided in the statefulset configuration. This Persistent Volume will be expanded by the `KubeDB` Enterprise operator.

5. Then, in order to set up storage autoscaling of the `MySQL` database the user creates a `MySQLAutoscaler` CRO with desired configuration.

6. `KubeDB` Autoscaler operator watches the `MySQLAutoscaler` CRO.

7. `KubeDB` Autoscaler operator continuously watches persistent volumes of the databases to check if it exceeds the specified usage threshold.

8. If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `MySQLOpsRequest` to expand the storage of the database.
9. `KubeDB` Enterprise operator watches the `MySQLOpsRequest` CRO.
10. Then the `KubeDB` Enterprise operator will expand the storage of the database component as specified on the `MySQLOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling storage of various MySQL database components using `MySQLAutoscaler` CRD.
