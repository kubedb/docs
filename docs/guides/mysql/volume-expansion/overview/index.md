---
title: MySQL Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-volume-expansion-overview
    name: Overview
    parent: guides-mysql-volume-expansion
    weight: 11
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# MySQL Volume Expansion

This guide will give an overview on how KubeDB Enterprise operator expand the volume of `MySQL`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [MySQL](/docs/guides/mysql/concepts/mysqldatabase)
    - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Enterprise operator expand the volumes of `MySQL` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Volume Expansion process of MySQL" src="/docs/guides/mysql/volume-expansion/overview/images/volume-expansion.jpg">
<figcaption align="center">Fig: Volume Expansion process of MySQL</figcaption>
</figure>

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `MySQL` Custom Resource (CR).

2. `KubeDB` Community operator watches the `MySQL` CR.

3. When the operator finds a `MySQL` CR, it creates required `StatefulSet` and related necessary stuff like secrets, services, etc.

4. The statefulSet creates Persistent Volumes according to the Volume Claim Template provided in the statefulset configuration. This Persistent Volume will be expanded by the `KubeDB` Enterprise operator.

5. Then, in order to expand the volume of the `MySQL` database the user creates a `MySQLOpsRequest` CR with desired information.

6. `KubeDB` Enterprise operator watches the `MySQLOpsRequest` CR.

7. When it finds a `MySQLOpsRequest` CR, it pauses the `MySQL` object which is referred from the `MySQLOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `MySQL` object during the volume expansion process.

8. Then the `KubeDB` Enterprise operator will expand the persistent volume to reach the expected size defined in the `MySQLOpsRequest` CR.

9. After the successfully expansion of the volume of the related StatefulSet Pods, the `KubeDB` Enterprise operator updates the new volume size in the `MySQL` object to reflect the updated state.

10. After the successful Volume Expansion of the `MySQL`, the `KubeDB` Enterprise operator resumes the `MySQL` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on Volume Expansion of various MySQL database using `MySQLOpsRequest` CRD.
