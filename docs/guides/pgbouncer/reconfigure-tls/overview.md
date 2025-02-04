---
title: Reconfiguring TLS of PgBouncer
menu:
  docs_{{ .version }}:
    identifier: pb-reconfigure-tls-overview
    name: Overview
    parent: pb-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring TLS of PgBouncer

This guide will give an overview on how KubeDB Ops-manager operator reconfigures TLS configuration i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates of a `PgBouncer`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [PgBouncer](/docs/guides/pgbouncer/concepts/pgbouncer.md)
  - [PgBouncerOpsRequest](/docs/guides/pgbouncer/concepts/opsrequest.md)

## How Reconfiguring PgBouncer TLS Configuration Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures TLS of a `PgBouncer`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring TLS process of PgBouncer" src="/docs/images/day-2-operation/pgbouncer/pb-reconfigure-tls.png">
<figcaption align="center">Fig: Reconfiguring TLS process of PgBouncer</figcaption>
</figure>

The Reconfiguring PgBouncer TLS process consists of the following steps:

1. At first, a user creates a `PgBouncer` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `PgBouncer` CRO.

3. When the operator finds a `PgBouncer` CR, it creates `PetSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the TLS configuration of the `PgBouncer` the user creates a `PgBouncerOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `PgBouncerOpsRequest` CR.

6. When it finds a `PgBouncerOpsRequest` CR, it pauses the `PgBouncer` object which is referred from the `PgBouncerOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `PgBouncer` object during the reconfiguring TLS process.  

7. Then the `KubeDB` Ops-manager operator will add, remove, update or rotate TLS configuration based on the Ops Request yaml.

8. Then the `KubeDB` Ops-manager operator will restart all the Pods of the pgbouncer so that they restart with the new TLS configuration defined in the `PgBouncerOpsRequest` CR.

9. After the successful reconfiguring of the `PgBouncer` TLS, the `KubeDB` Ops-manager operator resumes the `PgBouncer` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on reconfiguring TLS configuration of a PgBouncer using `PgBouncerOpsRequest` CRD.