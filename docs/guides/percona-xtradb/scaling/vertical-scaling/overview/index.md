---
title: PerconaXtraDB Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-scaling-vertical-overview
    name: Overview
    parent: guides-perconaxtradb-scaling-vertical
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PerconaXtraDB Vertical Scaling

This guide will give an overview on how KubeDB Enterprise operator vertically scales up `PerconaXtraDB`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb/)
  - [PerconaXtraDBOpsRequest](/docs/guides/percona-xtradb/concepts/opsrequest/)

## How Horizontal Scaling Process Works

The following diagram shows how KubeDB Enterprise operator scales up or down `PerconaXtraDB` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of PerconaXtraDB" src="/docs/guides/percona-xtradb/scaling/vertical-scaling/overview/images/vertical-scaling.jpg">
<figcaption align="center">Fig: Vertical scaling process of PerconaXtraDB</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `PerconaXtraDB` Custom Resource (CR).

2. `KubeDB` Community operator watches the `PerconaXtraDB` CR.

3. When the operator finds a `PerconaXtraDB` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `PerconaXtraDB` database the user creates a `PerconaXtraDBOpsRequest` CR with desired information.

5. `KubeDB` Enterprise operator watches the `PerconaXtraDBOpsRequest` CR.

6. When it finds a `PerconaXtraDBOpsRequest` CR, it halts the `PerconaXtraDB` object which is referred from the `PerconaXtraDBOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `PerconaXtraDB` object during the vertical scaling process.  

7. Then the `KubeDB` Enterprise operator will update resources of the StatefulSet Pods to reach desired state.

8. After the successful update of the resources of the StatefulSet's replica, the `KubeDB` Enterprise operator updates the `PerconaXtraDB` object to reflect the updated state.

9. After the successful update  of the `PerconaXtraDB` resources, the `KubeDB` Enterprise operator resumes the `PerconaXtraDB` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on updating resources of PerconaXtraDB database using `PerconaXtraDBOpsRequest` CRD.