---
title: Reconfiguring Oracle
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-reconfigure-overview
    name: Overview
    parent: guides-oracle-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Oracle

This guide will give you an overview on how KubeDB Ops-manager operator reconfigures an `Oracle` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)

## How Reconfiguring Oracle Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures `Oracle` database components. Open the image in a new tab to see the enlarged version.

The Reconfiguring Oracle process consists of the following steps:

1. At first, a user creates an `Oracle` Custom Resource (CR).

2. `KubeDB` Provisioner operator watches the `Oracle` CR.

3. When the operator finds an `Oracle` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. The custom configuration of the Oracle database is provided through `spec.configuration`. The configuration is supplied through a `Secret` (whose key is `oracle.cnf`) and/or an `inline` map. Each `KEY = value` entry in the configuration is translated into an `ALTER SYSTEM SET KEY=value SCOPE=SPFILE;` statement that is applied to the database.

5. Then, in order to reconfigure the `Oracle` database the user creates an `OracleOpsRequest` CR with the desired configuration.

6. `KubeDB` Ops-manager operator watches the `OracleOpsRequest` CR.

7. When it finds an `OracleOpsRequest` CR, it pauses the `Oracle` object so that the `KubeDB` Provisioner operator doesn't perform any operations on the `Oracle` object during the reconfiguring process.

8. Then the `KubeDB` Ops-manager operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `OracleOpsRequest` CR.

9. Then the `KubeDB` Ops-manager operator will restart the related `PetSet` Pods (one at a time, in a reconciliation-safe rolling manner) so that they restart with the new configuration defined in the `OracleOpsRequest` CR.

10. After the successful reconfiguring of the `Oracle` components, the `KubeDB` Ops-manager operator updates the `Oracle` object to reflect the updated state.

In the next docs, we are going to show a step by step guide on reconfiguring of an Oracle database using `OracleOpsRequest` CRD.

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

> ## ⚠️ Legal Notice
>
> Oracle® and Oracle Database® are registered trademarks of Oracle Corporation.
> KubeDB is not affiliated with, endorsed by, or sponsored by Oracle Corporation.
>
> KubeDB provides only orchestration and management tooling for Kubernetes.
> It does not distribute, bundle, ship, or include any Oracle Database software or binaries.
>
> Users must provide their own Oracle container images and hold valid Oracle licenses.
> Users are solely responsible for compliance with Oracle’s licensing terms, including all rules regarding containers, Docker, and Kubernetes environments.
>
> KubeDB makes no representations or warranties regarding Oracle licensing compliance.
