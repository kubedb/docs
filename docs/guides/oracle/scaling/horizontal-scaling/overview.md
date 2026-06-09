---
title: Oracle Horizontal Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: oracle-horizontal-scaling-overview
    name: Overview
    parent: oracle-horizontal-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Oracle Horizontal Scaling

This guide will give an overview of how KubeDB Ops-manager scales the number of nodes in a `Oracle` database cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)

## How Horizontal Scaling Works

The Horizontal Scaling process consists of the following steps:

1. At first, a user creates a `Oracle` CR.

2. `KubeDB-Provisioner` operator watches the `Oracle` CR.

3. When the operator finds a `Oracle` CR, it creates a `StatefulSet` with the specified number of node replicas, along with related necessary stuff like secrets, services, etc.

4. Then, in order to scale the number of nodes in the `Oracle` cluster, the user creates a `OracleOpsRequest` CR with the desired node count.

5. `KubeDB` Ops-manager operator watches the `OracleOpsRequest` CR.

6. When it finds a `OracleOpsRequest` CR, it pauses the `Oracle` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Oracle` during the scaling process.

7. Then the `KubeDB` Ops-manager operator scales the `StatefulSet` to the desired number of replicas.

8. After the successful scaling of the `StatefulSet`, the `KubeDB` Ops-manager updates the replica count in the `Oracle` object to reflect the updated state.

9. After the successful Horizontal Scaling, the `KubeDB` Ops-manager resumes the `Oracle` object so that the `KubeDB-Provisioner` resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on Horizontal Scaling of a Oracle database using `OracleOpsRequest` CRD.
