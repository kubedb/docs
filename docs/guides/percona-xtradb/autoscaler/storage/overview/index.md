---
title: PerconaXtraDB Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-autoscaling-storage-overview
    name: Overview
    parent: guides-perconaxtradb-autoscaling-storage
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PerconaXtraDB Vertical Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database storage using `perconaxtradbautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/percona-xtradb/concepts/perconaxtradb)
  - [PerconaXtraDBAutoscaler](/docs/guides/percona-xtradb/concepts/autoscaler)
  - [PerconaXtraDBOpsRequest](/docs/guides/percona-xtradb/concepts/opsrequest)

## How Storage Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `PerconaXtraDB` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Storage Autoscaling process of PerconaXtraDB" src="/docs/guides/percona-xtradb/autoscaler/storage/overview/images/pxas-storage.jpeg">
<figcaption align="center">Fig: Storage Autoscaling process of PerconaXtraDB</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, a user creates a `PerconaXtraDB` Custom Resource (CR).

2. `KubeDB` Community operator watches the `PerconaXtraDB` CR.

3. When the operator finds a `PerconaXtraDB` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Each StatefulSet creates a Persistent Volume according to the Volume Claim Template provided in the statefulset configuration. This Persistent Volume will be expanded by the `KubeDB` Enterprise operator.

5. Then, in order to set up storage autoscaling of the `PerconaXtraDB` database the user creates a `PerconaXtraDBAutoscaler` CRO with desired configuration.

6. `KubeDB` Autoscaler operator watches the `PerconaXtraDBAutoscaler` CRO.

7. `KubeDB` Autoscaler operator continuously watches persistent volumes of the databases to check if it exceeds the specified usage threshold.

8. If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `PerconaXtraDBOpsRequest` to expand the storage of the database.
9. `KubeDB` Enterprise operator watches the `PerconaXtraDBOpsRequest` CRO.
10. Then the `KubeDB` Enterprise operator will expand the storage of the database component as specified on the `PerconaXtraDBOpsRequest` CRO.

In the next docs, we are going to show a step-by-step guide on Autoscaling storage of various PerconaXtraDB database components using `PerconaXtraDBAutoscaler` CRD.
