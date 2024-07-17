---
title: MariaDB Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-volume-expansion-overview
    name: Overview
    parent: guides-mariadb-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MariaDB Volume Expansion

This guide will give an overview on how KubeDB Ops Manager expand the volume of `MariaDB`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Ops Manager expand the volumes of `MariaDB` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Volume Expansion process of MariaDB" src="/docs/guides/mariadb/volume-expansion/overview/images/volume-expansion.jpeg">
<figcaption align="center">Fig: Volume Expansion process of MariaDB</figcaption>
</figure>

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `MariaDB` Custom Resource (CR).

2. `KubeDB` Community operator watches the `MariaDB` CR.

3. When the operator finds a `MariaDB` CR, it creates required `PetSet` and related necessary stuff like secrets, services, etc.

4. The petSet creates Persistent Volumes according to the Volume Claim Template provided in the petset configuration. This Persistent Volume will be expanded by the `KubeDB` Enterprise operator.

5. Then, in order to expand the volume of the `MariaDB` database the user creates a `MariaDBOpsRequest` CR with desired information.

6. `KubeDB` Enterprise operator watches the `MariaDBOpsRequest` CR.

7. When it finds a `MariaDBOpsRequest` CR, it pauses the `MariaDB` object which is referred from the `MariaDBOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `MariaDB` object during the volume expansion process.

8. Then the `KubeDB` Enterprise operator will expand the persistent volume to reach the expected size defined in the `MariaDBOpsRequest` CR.

9. After the successfully expansion of the volume of the related PetSet Pods, the `KubeDB` Enterprise operator updates the new volume size in the `MariaDB` object to reflect the updated state.

10. After the successful Volume Expansion of the `MariaDB`, the `KubeDB` Enterprise operator resumes the `MariaDB` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on Volume Expansion of various MariaDB database using `MariaDBOpsRequest` CRD.
