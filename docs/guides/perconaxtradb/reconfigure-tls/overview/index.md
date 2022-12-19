---
title: Reconfiguring TLS of PerconaXtraDB Database
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-reconfigure-tls-overview
    name: Overview
    parent: guides-perconaxtradb-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Reconfiguring TLS of PerconaXtraDB Database

This guide will give an overview on how KubeDB Enterprise operator reconfigures TLS configuration i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates of a `PerconaXtraDB` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [PerconaXtraDB](/docs/guides/perconaxtradb/concepts/perconaxtradb)
  - [PerconaXtraDBOpsRequest](/docs/guides/perconaxtradb/concepts/opsrequest)

## How Reconfiguring PerconaXtraDB TLS Configuration Process Works

The following diagram shows how KubeDB Enterprise operator reconfigures TLS of a `PerconaXtraDB` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring TLS process of PerconaXtraDB" src="/docs/guides/perconaxtradb/reconfigure-tls/overview/images/reconfigure-tls.jpeg">
<figcaption align="center">Fig: Reconfiguring TLS process of PerconaXtraDB</figcaption>
</figure>

The Reconfiguring PerconaXtraDB TLS process consists of the following steps:

1. At first, a user creates a `PerconaXtraDB` Custom Resource Object (CRO).

2. `KubeDB` Community operator watches the `PerconaXtraDB` CRO.

3. When the operator finds a `PerconaXtraDB` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the TLS configuration of the `PerconaXtraDB` database the user creates a `PerconaXtraDBOpsRequest` CR with desired information.

5. `KubeDB` Enterprise operator watches the `PerconaXtraDBOpsRequest` CR.

6. When it finds a `PerconaXtraDBOpsRequest` CR, it pauses the `PerconaXtraDB` object which is referred from the `PerconaXtraDBOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `PerconaXtraDB` object during the reconfiguring TLS process.  

7. Then the `KubeDB` Enterprise operator will add, remove, update or rotate TLS configuration based on the Ops Request yaml.

8. Then the `KubeDB` Enterprise operator will restart all the Pods of the database so that they restart with the new TLS configuration defined in the `PerconaXtraDBOpsRequest` CR.

9. After the successful reconfiguring of the `PerconaXtraDB` TLS, the `KubeDB` Enterprise operator resumes the `PerconaXtraDB` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on reconfiguring TLS configuration of a PerconaXtraDB database using `PerconaXtraDBOpsRequest` CRD.