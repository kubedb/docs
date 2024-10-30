---
title: SingleStore Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: sdn-storage-auto-scaling-overview
    name: Overview
    parent: sdb-storage-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SingleStore Vertical Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database storage using `singlestoreautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
    - [SingleStoreAutoscaler](/docs/guides/singlestore/concepts/autoscaler.md)
    - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)

## How Storage Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `SingleStore` cluster components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Storage Auto Scaling process of SingleStore" src="/docs/images/singlestore/storage-autoscaling.svg">
<figcaption align="center">Fig: Storage Auto Scaling process of SingleStore</figcaption>
</figure>


The Auto Scaling process consists of the following steps:

1. At first, a user creates a `SingleStore` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `SingleStore` CR.

3. When the operator finds a `SingleStore` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

- Each PetSet creates a Persistent Volume according to the Volume Claim Template provided in the petset configuration.

4. Then, in order to set up storage autoscaling of the various components (ie. Aggregator, Leaf, Standalone.) of the `singlestore` cluster, the user creates a `SingleStoreAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `SingleStoreAutoscaler` CRO.

6. `KubeDB` Autoscaler operator continuously watches persistent volumes of the clusters to check if it exceeds the specified usage threshold.
- If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `SinglestoreOpsRequest` to expand the storage of the database.

7. `KubeDB` Ops-manager operator watches the `SinglestoreOpsRequest` CRO.

8. Then the `KubeDB` Ops-manager operator will expand the storage of the cluster component as specified on the `SinglestoreOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling storage of various Kafka cluster components using `SinglestoreAutoscaler` CRD.
