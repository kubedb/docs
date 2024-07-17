---
title: MongoDB Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: mg-vertical-scaling-overview
    name: Overview
    parent: mg-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Vertical Scaling

This guide will give an overview on how KubeDB Ops-manager operator updates the resources(for example CPU and Memory etc.) of the `MongoDB` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)

## How Vertical Scaling Process Works

The following diagram shows how KubeDB Ops-manager operator updates the resources of the `MongoDB` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Vertical scaling process of MongoDB" src="/docs/images/day-2-operation/mongodb/mg-vertical-scaling.svg">
<figcaption align="center">Fig: Vertical scaling process of MongoDB</figcaption>
</figure>

The vertical scaling process consists of the following steps:

1. At first, a user creates a `MongoDB` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `MongoDB` CR.

3. When the operator finds a `MongoDB` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the resources(for example `CPU`, `Memory` etc.) of the `MongoDB` database the user creates a `MongoDBOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `MongoDBOpsRequest` CR.

6. When it finds a `MongoDBOpsRequest` CR, it halts the `MongoDB` object which is referred from the `MongoDBOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `MongoDB` object during the vertical scaling process.  

7. Then the `KubeDB` Ops-manager operator will update resources of the PetSet Pods to reach desired state.

8. After the successful update of the resources of the PetSet's replica, the `KubeDB` Ops-manager operator updates the `MongoDB` object to reflect the updated state.

9. After the successful update  of the `MongoDB` resources, the `KubeDB` Ops-manager operator resumes the `MongoDB` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on updating resources of MongoDB database using `MongoDBOpsRequest` CRD.