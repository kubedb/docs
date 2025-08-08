---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: cas-rotate-auth-overview
    name: Overview
    parent: cas-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of Cassandra

This guide will give an overview on how KubeDB Ops-manager operator Rotate Authentication configuration.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
    - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)

## How Rotate Cassandra Authentication Configuration Process Works

The Rotate Cassandra Authentication process consists of the following steps:

1. At first, a user creates a `Cassandra` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `Cassandra` CRO.

3. When the operator finds a `Cassandra` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to rotate the authentication configuration of the `Cassandra`, the user creates a `CassandraOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `CassandraOpsRequest` CR.

6. When it finds a `CassandraOpsRequest` CR, it pauses the `Cassandra` object which is referred from the `CassandraOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Cassandra` object during the rotating Authentication process.

7. Then the `KubeDB` Ops-manager operator will update necessary configuration based on the Ops Request yaml to update credentials.

8. Then the `KubeDB` Ops-manager operator will restart all the Pods of the database so that they restart with the new authentication `ENVs` or other configuration defined in the `CassandraOpsRequest` CR.

9. After the successful rotating of the `Cassandra` Authentication, the `KubeDB` Ops-manager operator resumes the `Cassandra` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on rotating Authentication configuration of a Cassandra using `CassandraOpsRequest` CRD.
