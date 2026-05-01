---
title: Oracle Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: oracle-vertical-scaling-overview
    name: Overview
    parent: oracle-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Oracle Vertical Scaling

This guide will give an overview of how KubeDB Ops-manager updates the CPU and memory resources of `Oracle` database nodes.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)

## How Vertical Scaling Works

The Vertical Scaling process consists of the following steps:

1. At first, a user creates a `Oracle` CR.

2. `KubeDB-Provisioner` operator watches the `Oracle` CR.

3. When the operator finds a `Oracle` CR, it creates a `StatefulSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the CPU and memory resources of the `Oracle` database nodes, the user creates a `OracleOpsRequest` CR with the desired resource specifications.

5. `KubeDB` Ops-manager operator watches the `OracleOpsRequest` CR.

6. When it finds a `OracleOpsRequest` CR, it pauses the `Oracle` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Oracle` during the scaling process.

7. Then the `KubeDB` Ops-manager operator updates the resources of the `StatefulSet` pods to the desired values defined in the `OracleOpsRequest` CR.

8. After the successful resource update of the pods, the `KubeDB` Ops-manager updates the resource specifications in the `Oracle` object to reflect the updated state.

9. After the successful Vertical Scaling, the `KubeDB` Ops-manager resumes the `Oracle` object so that the `KubeDB-Provisioner` resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on Vertical Scaling of a Oracle database using `OracleOpsRequest` CRD.
