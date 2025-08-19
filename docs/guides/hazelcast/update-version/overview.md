---
title: Updating Hazelcast Version Overview
menu:
  docs_{{ .version }}:
    identifier: hz-update-version-overview
    name: Overview
    parent: update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Updating Hazelcast version Overview

This guide will give you an overview on how KubeDB Ops Manager update the version of `Hazelcast` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md)
  - [HazelcastOpsRequest](/docs/guides/hazelcast/concepts/hazelcast-opsrequest.md)

## How Update Version Process Works

The following diagram shows how KubeDB Ops Manager used to update the version of `Hazelcast`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Update Version Process of Hazelcast" src="/docs/images/day-2-operation/hazelcast/hz-version-update.svg">
<figcaption align="center">Fig: Update Version Process of Hazelcast</figcaption>
</figure>

The updating process consists of the following steps:

1. At first, a user creates a `Hazelcast` Custom Resource (CR).

2. `KubeDB` Community operator watches the `Hazelcast` CR.

3. When the operator finds a `Hazelcast` CR, it creates required number of `Statefulsets` and related necessary stuff like appbinding, services, etc.

4. Then, in order to update the version of the `Hazelcast` database the user creates a `HazelcastOpsRequest` CR with the desired version.

5. `KubeDB` Enterprise operator watches the `HazelcastOpsRequest` CR.

6. When it finds a `HazelcastOpsRequest` CR, it halts the `Hazelcast` object which is referred from the `HazelcastOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `Hazelcast` object during the updating process.  

7. By looking at the target version from `HazelcastOpsRequest` CR, `KubeDB` Enterprise operator updates the images of all the `StatefulSets`.

8. After successfully updating the `StatefulSets` and their `Pods` images, the `KubeDB` Enterprise operator updates the image of the `Hazelcast` object to reflect the updated state of the database.

9. After successfully updating of `Hazelcast` object, the `KubeDB` Enterprise operator resumes the `Hazelcast` object so that the `KubeDB` Community operator can resume its usual operations.

In the next doc, we are going to show a step-by-step guide on updating of a Hazelcast database using update operation.
