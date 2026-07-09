---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: ignite-rotate-auth-overview
    name: Overview
    parent: guides-ignite-rotate-auth
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Ignite

This guide will give an overview on how KubeDB Ops-manager operator Rotate Authentication configuration.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Ignite](/docs/guides/ignite/concepts/ignite.md)
    - [IgniteOpsRequest](/docs/guides/ignite/concepts/opsrequest.md)

## How Rotate Ignite Authentication Configuration Process Works

The authentication rotation process for Ignite using KubeDB involves the following steps:

1. A user first creates an `Ignite` Custom Resource Object (CRO).

2. The `KubeDB Provisioner operator` continuously watches for `Ignite` CROs.

3. When the operator detects an `Ignite` CR, it provisions the required `PetSets`, along with related resources such as secrets, services, and other dependencies.

4. To initiate authentication rotation, the user creates an `IgniteOpsRequest` CR with the desired configuration.

5. The `KubeDB Ops-manager` operator watches for `IgniteOpsRequest` CRs.

6. Upon detecting an `IgniteOpsRequest`, the operator pauses the referenced `Ignite` object, ensuring that the Provisioner
   operator does not perform any operations during the authentication rotation process.

7. The `Ops-manager` operator then updates the necessary configuration (such as credentials) based on the provided `IgniteOpsRequest` specification.

8. After applying the updated configuration, the operator restarts all `Ignite` Pods so they come up with the new authentication environment variables and settings.

9. Once the authentication rotation is completed successfully, the operator resumes the `Ignite` object, allowing the Provisioner operator to continue its usual operations.

In the next section, we will walk you through a step-by-step guide to rotating Ignite authentication using the `IgniteOpsRequest` CRD.
