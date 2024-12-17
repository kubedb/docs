---
title: Reconfiguring TLS of FerretDB
menu:
  docs_{{ .version }}:
    identifier: fr-reconfigure-tls-overview
    name: Overview
    parent: fr-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring TLS of FerretDB

This guide will give an overview on how KubeDB Ops-manager operator reconfigures TLS configuration i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates of a `FerretDB`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [FerretDB](/docs/guides/ferretdb/concepts/ferretdb.md)
    - [FerretDBOpsRequest](/docs/guides/ferretdb/concepts/opsrequest.md)

## How Reconfiguring FerretDB TLS Configuration Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures TLS of a `FerretDB`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring TLS process of FerretDB" src="/docs/images/ferretdb/fr-reconfigure-tls.svg">
<figcaption align="center">Fig: Reconfiguring TLS process of FerretDB</figcaption>
</figure>

The Reconfiguring FerretDB TLS process consists of the following steps:

1. At first, a user creates a `FerretDB` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `FerretDB` CRO.

3. When the operator finds a `FerretDB` CR, it creates `PetSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the TLS configuration of the `FerretDB` the user creates a `FerretDBOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `FerretDBOpsRequest` CR.

6. When it finds a `FerretDBOpsRequest` CR, it pauses the `FerretDB` object which is referred from the `FerretDBOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `FerretDB` object during the reconfiguring TLS process.

7. Then the `KubeDB` Ops-manager operator will add, remove, update or rotate TLS configuration based on the Ops Request yaml.

8. Then the `KubeDB` Ops-manager operator will restart all the Pods of the ferretdb so that they restart with the new TLS configuration defined in the `FerretDBOpsRequest` CR.

9. After the successful reconfiguring of the `FerretDB` TLS, the `KubeDB` Ops-manager operator resumes the `FerretDB` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step-by-step guide on reconfiguring TLS configuration of a FerretDB using `FerretDBOpsRequest` CRD.