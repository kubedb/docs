---
title: SingleStore Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-scaling-vertical-overview
    name: Overview
    parent: guides-sdb-scaling-vertical
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SingleStore Vertical Scaling

This guide will give an overview on how KubeDB Ops Manager vertically scales up `SingleStore`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
  - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)

The following diagram shows how KubeDB Ops Manager scales up or down `SingleStore` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of SingleStore" src="/docs/guides/singlestore/scaling/vertical-scaling/overview/images/vertical-sacling.svg">
<figcaption align="center">Fig: Vertical scaling process of SingleStore</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `SingleStore` Custom Resource (CR).

2. `KubeDB` Provisioner operator watches the `SingleStore` CR.

3. When the operator finds a `SingleStore` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `SingleStore` database the user creates a `SingleStoreOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `SingleStoreOpsRequest` CR.

6. When it finds a `SingleStoreOpsRequest` CR, it halts the `SingleStore` object which is referred from the `SingleStoreOpsRequest`. So, the `KubeDB` Provisioner operator doesn't perform any operations on the `SingleStore` object during the vertical scaling process.  

7. Then the `KubeDB` Ops-manager operator will update resources of the PetSet Pods to reach desired state.

8. After the successful update of the resources of the PetSet's replica, the `KubeDB` Ops-manager operator updates the `SingleStore` object to reflect the updated state.

9. After the successful update  of the `SingleStore` resources, the `KubeDB` Ops-manager operator resumes the `SingleStore` object so that the `KubeDB` Provisioner operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on updating resources of SingleStore database using `SingleStoreOpsRequest` CRD.