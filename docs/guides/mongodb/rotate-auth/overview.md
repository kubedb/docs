---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: mg-rotate-auth-overview
    name: Overview
    parent: mg-rotate-auth
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of MongoDB

This guide will give an overview on how KubeDB Ops-manager operator Rotate Authentication configuration.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
    - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)

## How Rotate MongoDB Authentication Configuration Process Works

[//]: # (The following diagram shows how KubeDB Ops-manager operator Rotate Authentication of a `MongoDB`. Open the image in a new tab to see the enlarged version.)

[//]: # ()
[//]: # (<figure align="center">)

[//]: # (  <img alt="Rotate Authentication process of MongoDB" src="/docs/images/day-2-operation/MongoDB/kf-rotate-auth.svg">)

[//]: # (<figcaption align="center">Fig: Rotate Auth process of MongoDB</figcaption>)

[//]: # (</figure>)

The authentication rotation process for MongoDB using KubeDB involves the following steps:

1. A user first creates a `MongoDB` Custom Resource Object (CRO).

2. The `KubeDB Provisioner operator` continuously watches for `MongoDB` CROs.

3. When the operator detects a `MongoDB` CR, it provisions the required `PetSets`, along with related resources such as secrets, services, and other dependencies.

4. To initiate authentication rotation, the user creates a `MongoDBOpsRequest` CR with the desired configuration.

5. The `KubeDB Ops-manager` operator watches for `MongoDBOpsRequest` CRs.

6. Upon detecting a `MongoDBOpsRequest`, the operator pauses the referenced `MongoDB` object, ensuring that the Provisioner
   operator does not perform any operations during the authentication rotation process.

7. The `Ops-manager` operator then updates the necessary configuration (such as credentials) based on the provided `MongoDBOpsRequest` specification.

8. After applying the updated configuration, the operator restarts all `MongoDB` Pods so they come up with the new authentication environment variables and settings.

9. Once the authentication rotation is completed successfully, the operator resumes the `MongoDB` object, allowing the Provisioner operator to continue its usual operations.

In the next section, we will walk you through a step-by-step guide to rotating MongoDB authentication using the `MongoDBOpsRequest` CRD.
