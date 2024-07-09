---
title: PerconaXtraDB Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-volume-expansion-overview
    name: Overview
    parent: guides-perconaxtradb-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PerconaXtraDB Volume Expansion

This guide will give an overview on how KubeDB Ops Manager expand the volume of `PerconaXtraDB`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb)
  - [PerconaXtraDBOpsRequest](/docs/guides/percona-xtradb/concepts/opsrequest)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Ops Manager expand the volumes of `PerconaXtraDB` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Volume Expansion process of PerconaXtraDB" src="/docs/guides/percona-xtradb/volume-expansion/overview/images/volume-expansion.jpeg">
<figcaption align="center">Fig: Volume Expansion process of PerconaXtraDB</figcaption>
</figure>

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `PerconaXtraDB` Custom Resource (CR).

2. `KubeDB` Community operator watches the `PerconaXtraDB` CR.

3. When the operator finds a `PerconaXtraDB` CR, it creates required `StatefulSet` and related necessary stuff like secrets, services, etc.

4. The petSet creates Persistent Volumes according to the Volume Claim Template provided in the statefulset configuration. This Persistent Volume will be expanded by the `KubeDB` Enterprise operator.

5. Then, in order to expand the volume of the `PerconaXtraDB` database the user creates a `PerconaXtraDBOpsRequest` CR with desired information.

6. `KubeDB` Enterprise operator watches the `PerconaXtraDBOpsRequest` CR.

7. When it finds a `PerconaXtraDBOpsRequest` CR, it pauses the `PerconaXtraDB` object which is referred from the `PerconaXtraDBOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `PerconaXtraDB` object during the volume expansion process.

8. Then the `KubeDB` Enterprise operator will expand the persistent volume to reach the expected size defined in the `PerconaXtraDBOpsRequest` CR.

9. After the successfully expansion of the volume of the related StatefulSet Pods, the `KubeDB` Enterprise operator updates the new volume size in the `PerconaXtraDB` object to reflect the updated state.

10. After the successful Volume Expansion of the `PerconaXtraDB`, the `KubeDB` Enterprise operator resumes the `PerconaXtraDB` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on Volume Expansion of various PerconaXtraDB database using `PerconaXtraDBOpsRequest` CRD.
