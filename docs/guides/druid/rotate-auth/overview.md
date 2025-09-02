---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: druid-rotate-auth-overview
    name: Overview
    parent: guides-druid-rotate-auth
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Druid

This guide will give an overview on how KubeDB Ops-manager operator Rotate Authentication configuration.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Druid](/docs/guides/druid/concepts/druid.md)
    - [DruidOpsRequest](/docs/guides/druid/concepts/druidopsrequest.md)

## How Rotate Druid Authentication Configuration Process Works

[//]: # (The following diagram shows how KubeDB Ops-manager operator Rotate Authentication of a `Druid`. Open the image in a new tab to see the enlarged version.)

[//]: # ()
[//]: # (<figure align="center">)

[//]: # (  <img alt="Rotate Authentication process of Druid" src="/docs/images/day-2-operation/Druid/kf-rotate-auth.svg">)

[//]: # (<figcaption align="center">Fig: Rotate Auth process of Druid</figcaption>)

[//]: # (</figure>)

The authentication rotation process for Druid using KubeDB involves the following steps:

1. A user first creates a `Druid` Custom Resource Object (CRO).

2. The `KubeDB Provisioner operator` continuously watches for `Druid` CROs.

3. When the operator detects a `Druid` CR, it provisions the required `PetSets`, along with related resources such as secrets, services, and other dependencies.

4. To initiate authentication rotation, the user creates a `DruidOpsRequest` CR with the desired configuration.

5. The `KubeDB Ops-manager` operator watches for `DruidOpsRequest` CRs.

6. Upon detecting a `DruidOpsRequest`, the operator pauses the referenced `Druid` object, ensuring that the Provisioner
   operator does not perform any operations during the authentication rotation process.

7. The `Ops-manager` operator then updates the necessary configuration (such as credentials) based on the provided `DruidOpsRequest` specification.

8. After applying the updated configuration, the operator restarts all `Druid` Pods so they come up with the new authentication environment variables and settings.

9. Once the authentication rotation is completed successfully, the operator resumes the `Druid` object, allowing the Provisioner operator to continue its usual operations.

In the next section, we will walk you through a step-by-step guide to rotating Druid authentication using the `DruidOpsRequest` CRD.
