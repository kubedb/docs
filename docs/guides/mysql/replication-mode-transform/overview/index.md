---
title: MySQL Replication Mode Transform Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-replication-mode-transform-overview
    name: Overview
    parent: guides-mysql-mode-transform
    weight: 11
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQL Replication Mode Transform

This guide will give an overview on how KubeDB Ops Manager transform replication mode of `MySQL`. Currently, you can transform `remote replica` to `group replication`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [MySQL](/docs/guides/mysql/concepts/mysqldatabase)
    - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest)

## How Replication Mode Transform Process Works

The following diagram shows how KubeDB Ops Manager transform replication mode of `MySQL` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Volume Expansion process of MySQL" src="/docs/guides/mysql/replication-mode-transform/overview/images/replication-mode-transform.svg">
<figcaption align="center">Fig: Replication Mode Transform process of MySQL</figcaption>
</figure>

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `MySQL` Custom Resource (CR).

2. `KubeDB` provisioner operator watches the `MySQL` CR.

3. When the operator finds a `MySQL` CR, it creates required `PetSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to transform replication mode of the `MySQL` database the user creates a `MySQLOpsRequest` CR with desired information.

5. `KubeDB` ops-manager operator watches the `MySQLOpsRequest` CR.

6. When it finds a `MySQLOpsRequest` CR, it pauses the `MySQL` object which is referred from the `MySQLOpsRequest`. So, the `KubeDB` provisioner operator doesn't perform any operations on the `MySQL` object during the mode transform process.

7. Then the `KubeDB` ops-request operator will transform replication mode to reach the expected replication mode defined in the `MySQLOpsRequest` CR.

8. After the successful transformation of replication mode of the related PetSet Pods, the `KubeDB` ops-request operator updates the new replication mode in the `MySQL` object to reflect the updated state.

9. After the successful transformation of replication mode of the `MySQL`, the `KubeDB` ops-request operator resumes the `MySQL` object so that the `KubeDB` provisioner operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on transform replication mode of various MySQL database using `MySQLOpsRequest` CRD.
