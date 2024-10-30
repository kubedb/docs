---
title: SingleStore Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-volume-expansion-overview
    name: Overview
    parent: guides-sdb-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SingleStore Volume Expansion

This guide will give an overview on how KubeDB Ops Manager expand the volume of `SingleStore`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
  - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Ops Manager expand the volumes of `SingleStore` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Volume Expansion process of SingleStore" src="/docs/guides/singlestore/volume-expansion/overview/images/volume-expansion.svg">
<figcaption align="center">Fig: Volume Expansion process of SingleStore</figcaption>
</figure>

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `SingleStore` Custom Resource (CR).

2. `KubeDB` Provisioner operator watches the `SingleStore` CR.

3. When the operator finds a `SingleStore` CR, it creates required `PetSet` and related necessary stuff like secrets, services, etc.

4. The petSet creates Persistent Volumes according to the Volume Claim Template provided in the petset configuration. This Persistent Volume will be expanded by the `KubeDB` Ops-manager operator.

5. Then, in order to expand the volume of the `SingleStore` database the user creates a `SingleStoreOpsRequest` CR with desired information.

6. `KubeDB` Ops-manager operator watches the `SingleStoreOpsRequest` CR.

7. When it finds a `SingleStoreOpsRequest` CR, it pauses the `SingleStore` object which is referred from the `SingleStoreOpsRequest`. So, the `KubeDB` Provisioner operator doesn't perform any operations on the `SingleStore` object during the volume expansion process.

8. Then the `KubeDB` Ops-manager operator will expand the persistent volume to reach the expected size defined in the `SingleStoreOpsRequest` CR.

9. After the successfully expansion of the volume of the related PetSet Pods, the `KubeDB` Ops-manager operator updates the new volume size in the `SingleStore` object to reflect the updated state.

10. After the successful Volume Expansion of the `SingleStore`, the `KubeDB` Ops-manager operator resumes the `SingleStore` object so that the `KubeDB` Provisioner operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on Volume Expansion of various SingleStore database using `SingleStoreOpsRequest` CRD.
