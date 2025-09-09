---
title: Hazelcast Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: hz-volume-expansion-overview
    name: Overview
    parent: hz-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Hazelcast Volume Expansion

This guide will give an overview on how KubeDB Ops-manager operator expand the volume of `Hazelcast` cluster members.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md)
  - [HazelcastOpsRequest](/docs/guides/hazelcast/concepts/hazelcast-opsrequest.md)

## How Volume Expansion Process Works

The following diagram shows how KubeDB Ops-manager operator expand the volumes of `Hazelcast` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
    <img alt="Volume Expansion process of Hazelcast" src="/docs/images/day-2-operation/hazelcast/hz-volume-expansion.svg">
<figcaption align="center">Fig: Volume Expansion process of Hazelcast</figcaption>
</figure>

The Volume Expansion process consists of the following steps:

1. At first, a user creates a `Hazelcast` Custom Resource (CR).

2. `KubeDB` Provisioner operator watches the `Hazelcast` CR.

3. When the operator finds a `Hazelcast` CR, it creates required number of `statefulsets` and related necessary stuff like secrets, services, etc.

4. Each statefulset creates a Persistent Volume according to the Volume Claim Template provided in the statefulset configuration. This Persistent Volume will be expanded by the `KubeDB` Ops-manager operator.

5. Then, in order to expand the volume of the `Hazelcast` database the user creates a `HazelcastOpsRequest` CR with desired information.

6. `KubeDB` Ops-manager operator watches the `HazelcastOpsRequest` CR.

7. When it finds a `HazelcastOpsRequest` CR, it pauses the `Hazelcast` object which is referred from the `HazelcastOpsRequest`. So, the `KubeDB` Provisioner operator doesn't perform any operations on the `Hazelcast` object during the volume expansion process.

8. Then the `KubeDB` Ops-manager operator will expand the related PersistentVolumeClaims to reach the expected size defined in the `HazelcastOpsRequest` CR.

9. After the successful expansion of the related PersistentVolumeClaims, the `KubeDB` Ops-manager operator updates the related statefulset so that the new volumes can be mounted.

10. After successfully updating statefulsets, the `KubeDB` Ops-manager operator resumes the `Hazelcast` object so that the `KubeDB` Provisioner operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on expanding volume of various Hazelcast cluster members using `HazelcastOpsRequest` CRD.
