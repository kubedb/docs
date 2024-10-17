---
title: Reconfiguring TLS of SingleStore Database
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-reconfigure-tls-overview
    name: Overview
    parent: guides-sdb-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring TLS of SingleStore Database

This guide will give an overview on how KubeDB Ops Manager reconfigures TLS configuration i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates of a `SingleStore` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
  - [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)

## How Reconfiguring SingleStore TLS Configuration Process Works

The following diagram shows how KubeDB Ops Manager reconfigures TLS of a `SingleStore` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring TLS process of SingleStore" src="/docs/guides/singlestore/reconfigure-tls/overview/images/reconfigure-tls.svg">
<figcaption align="center">Fig: Reconfiguring TLS process of SingleStore</figcaption>
</figure>

The Reconfiguring SingleStore TLS process consists of the following steps:

1. At first, a user creates a `SingleStore` Custom Resource Object (CRO).

2. `KubeDB` Provisioner operator watches the `SingleStore` CRO.

3. When the operator finds a `SingleStore` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the TLS configuration of the `SingleStore` database the user creates a `SingleStoreOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `SingleStoreOpsRequest` CR.

6. When it finds a `SingleStoreOpsRequest` CR, it pauses the `SingleStore` object which is referred from the `SingleStoreOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `SingleStore` object during the reconfiguring TLS process.  

7. Then the `KubeDB` Enterprise operator will add, remove, update or rotate TLS configuration based on the Ops Request yaml.

8. Then the `KubeDB` Enterprise operator will restart all the Pods of the database so that they restart with the new TLS configuration defined in the `SingleStoreOpsRequest` CR.

9. After the successful reconfiguring of the `SingleStore` TLS, the `KubeDB` Enterprise operator resumes the `SingleStore` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on reconfiguring TLS configuration of a SingleStore database using `SingleStoreOpsRequest` CRD.