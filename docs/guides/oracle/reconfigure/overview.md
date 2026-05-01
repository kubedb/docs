---
title: Reconfiguring Oracle
menu:
  docs_{{ .version }}:
    identifier: oracle-reconfigure-overview
    name: Overview
    parent: oracle-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Oracle

This guide will give an overview of how KubeDB Ops-manager reconfigures a `Oracle` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)

## How Reconfiguration Works

The Reconfiguration process consists of the following steps:

1. At first, a user creates a `Oracle` CR.

2. `KubeDB-Provisioner` operator watches the `Oracle` CR.

3. When the operator finds a `Oracle` CR, it creates a `StatefulSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the `Oracle` database, the user creates a `OracleOpsRequest` CR with the new configuration. The user can provide the new configuration either via a new config secret, via `applyConfig`, or by removing the custom configuration (reverting to defaults).

5. `KubeDB` Ops-manager operator watches the `OracleOpsRequest` CR.

6. When it finds a `OracleOpsRequest` CR, it pauses the `Oracle` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Oracle` during the reconfiguration process.

7. Then the `KubeDB` Ops-manager operator updates the configuration secret and restarts the pods in a rolling fashion to apply the new configuration.

8. After the successful configuration update, the `KubeDB` Ops-manager updates the `Oracle` object to reflect the updated configuration state.

9. After the successful reconfiguration, the `KubeDB` Ops-manager resumes the `Oracle` object so that the `KubeDB-Provisioner` resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on reconfiguring a Oracle database using `OracleOpsRequest` CRD.
