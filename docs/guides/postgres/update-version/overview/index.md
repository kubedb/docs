---
title: Updating Postgres Overview
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-updating-overview
    name: Overview
    parent: guides-postgres-updating
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Updating Postgres version

This guide will give you an overview of how KubeDB ops manager updates the version of `Postgres` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Postgres](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)

## How update Process Works

The following diagram shows how KubeDB KubeDB ops manager used to update the version of `Postgres`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Postgres update Flow" src="/docs/guides/postgres/update-version/overview/images/pg-updating.png">
<figcaption align="center">Fig: updating Process of Postgres</figcaption>
</figure>

The updating process consists of the following steps:

1. At first, a user creates a `Postgres` cr.

2. `KubeDB-Provisioner` operator watches for the `Postgres` cr.

3. When it finds one, it creates a `StatefulSet` and related necessary stuff like secret, service, etc.

4. Then, in order to update the version of the `Postgres` database the user creates a `PostgresOpsRequest` cr with the desired version.

5. `KubeDB-ops-manager` operator watches for `PostgresOpsRequest`.

6. When it finds one, it Pauses the `Postgres` object so that the `KubeDB-Provisioner` operator doesn't perform any operation on the `Postgres` during the updating process.

7. By looking at the target version from `PostgresOpsRequest` cr, In case of major update `KubeDB-ops-manager` does some pre-update steps as we need old bin and lib files to update from current to target Postgres version. 
8. Then By looking at the target version from `PostgresOpsRequest` cr, `KubeDB-ops-manager` operator updates the images of the `StatefulSet` for updating versions.
  

9. After successful upgradation of the `StatefulSet` and its `Pod` images, the `KubeDB-ops-manager` updates the image of the `Postgres` object to reflect the updated cluster state.

10. After successful upgradation of `Postgres` object, the `KubeDB` ops manager resumes the `Postgres` object so that the `KubeDB-provisioner` can resume its usual operations.

In the next doc, we are going to show a step by step guide on updating of a Postgres database using update operation.