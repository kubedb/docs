---
title: Postgres Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: pg-storage-auto-scaling-overview
    name: Overview
    parent: pg-storage-auto-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Postgres Storage Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database storage using `postgresautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Postgres](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)

## How Storage Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `Postgres` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Storage Auto Scaling process of Postgres" src="/docs/images/postgres/pg-storage-process.svg">
<figcaption align="center">Fig: Storage Auto Scaling process of Postgres</figcaption>
</figure>


The Auto Scaling process consists of the following steps:

1. At first, a user creates a `Postgres` Custom Resource (CR).

2. `KubeDB` Provisioner  operator watches the `Postgres` CR.

3. When the operator finds a `Postgres` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

- Each PetSet creates a Persistent Volume according to the Volume Claim Template provided in the petset configuration.

4. Then, in order to set up storage autoscaling of the `Postgres` database the user creates a `PostgresAutoscaler` CRO with desired configuration.

5. `KubeDB` Autoscaler operator watches the `PostgresAutoscaler` CRO.

6. `KubeDB` Autoscaler operator continuously watches persistent volumes of the databases to check if it exceeds the specified usage threshold.
- If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `PostgresOpsRequest` to expand the storage of the database. 
   
7. `KubeDB` Ops-manager operator watches the `PostgresOpsRequest` CRO.

8. Then the `KubeDB` Ops-manager operator will expand the storage of the database component as specified on the `PostgresOpsRequest` CRO.

In the next docs, we are going to show a step by step guide on Autoscaling storage of various Postgres database components using `PostgresAutoscaler` CRD.
