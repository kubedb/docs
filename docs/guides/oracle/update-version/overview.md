---
title: Updating Oracle Version
menu:
  docs_{{ .version }}:
    identifier: oracle-update-version-overview
    name: Overview
    parent: oracle-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Updating Oracle Version

This guide will give you an overview of how KubeDB Ops-manager updates the version of a `Oracle` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)

## How the Update Process Works

The updating process consists of the following steps:

1. At first, a user creates a `Oracle` CR.

2. `KubeDB-Provisioner` operator watches for the `Oracle` CR.

3. When it finds one, it creates a `StatefulSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to update the version of the `Oracle` database, the user creates a `OracleOpsRequest` CR with the desired target version.

5. `KubeDB-ops-manager` operator watches for `OracleOpsRequest`.

6. When it finds one, it pauses the `Oracle` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Oracle` during the updating process.

7. By looking at the target version from the `OracleOpsRequest` CR, the `KubeDB-ops-manager` operator updates the images of the `StatefulSet` for the new version.

8. After successful update of the `StatefulSet` and its Pod images, the `KubeDB-ops-manager` updates the image of the `Oracle` object to reflect the updated cluster state.

9. After successful update of the `Oracle` object, the `KubeDB` Ops-manager resumes the `Oracle` object so that the `KubeDB-Provisioner` can resume its usual operations.

In the next doc, we are going to show a step-by-step guide on updating a Oracle database using the `UpdateVersion` operation.
