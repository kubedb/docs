---
title: Update Version Overview
menu:
  docs_{{ .version }}:
    identifier: kf-update-version-overview
    name: Overview
    parent: kf-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Kafka Update Version Overview

This guide will give you an overview on how KubeDB Ops-manager operator update the version of `Kafka`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [KafkaOpsRequest](/docs/guides/kafka/concepts/kafkaopsrequest.md)

## How update version Process Works

The following diagram shows how KubeDB Ops-manager operator used to update the version of `Kafka`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="updating Process of Kafka" src="/docs/images/day-2-operation/mongodb/mg-updating.svg">
<figcaption align="center">Fig: updating Process of Kafka</figcaption>
</figure>

The updating process consists of the following steps:

1. At first, a user creates a `Kafka` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Kafka` CR.

3. When the operator finds a `Kafka` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the version of the `Kafka` database the user creates a `KafkaOpsRequest` CR with the desired version.

5. `KubeDB` Ops-manager operator watches the `KafkaOpsRequest` CR.

6. When it finds a `KafkaOpsRequest` CR, it halts the `Kafka` object which is referred from the `KafkaOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Kafka` object during the updating process.

7. By looking at the target version from `KafkaOpsRequest` CR, `KubeDB` Ops-manager operator updates the images of all the `PetSets`.

8. After successfully updating the `PetSets` and their `Pods` images, the `KubeDB` Ops-manager operator updates the image of the `Kafka` object to reflect the updated state of the database.

9. After successfully updating of `Kafka` object, the `KubeDB` Ops-manager operator resumes the `Kafka` object so that the `KubeDB` Provisioner  operator can resume its usual operations.

In the next doc, we are going to show a step by step guide on updating of a Kafka database using updateVersion operation.