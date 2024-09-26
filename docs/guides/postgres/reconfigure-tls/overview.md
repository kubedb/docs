---
title: Reconfiguring TLS of Postgres Database
menu:
  docs_{{ .version }}:
    identifier: pg-reconfigure-tls-overview
    name: Overview
    parent: pg-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring TLS of Postgres Database

This guide will give an overview on how KubeDB Ops-manager operator reconfigures TLS configuration i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates of a `Postgres` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Postgres](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)

## How Reconfiguring Postgres TLS Configuration Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures TLS of a `Postgres` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring TLS process of Postgres" src="/docs/images/day-2-operation/postgres/pg-reconfigure-tls.svg">
<figcaption align="center">Fig: Reconfiguring TLS process of Postgres</figcaption>
</figure>

The Reconfiguring Postgres TLS process consists of the following steps:

1. At first, a user creates a `Postgres` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `Postgres` CRO.

3. When the operator finds a `Postgres` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the TLS configuration of the `Postgres` database the user creates a `PostgresOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `PostgresOpsRequest` CR.

6. When it finds a `PostgresOpsRequest` CR, it pauses the `Postgres` object which is referred from the `PostgresOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Postgres` object during the reconfiguring TLS process.  

7. Then the `KubeDB` Ops-manager operator will add, remove, update or rotate TLS configuration based on the Ops Request yaml.

8. Then the `KubeDB` Ops-manager operator will restart all the Pods of the database so that they restart with the new TLS configuration defined in the `PostgresOpsRequest` CR.

9. After the successful reconfiguring of the `Postgres` TLS, the `KubeDB` Ops-manager operator resumes the `Postgres` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on reconfiguring TLS configuration of a Postgres database using `PostgresOpsRequest` CRD.