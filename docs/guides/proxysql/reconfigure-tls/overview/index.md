---
title: Reconfiguring TLS of ProxySQL
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-reconfigure-tls-overview
    name: Overview
    parent: guides-proxysql-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Reconfiguring TLS of ProxySQL

This guide will give an overview on how KubeDB Enterprise operator reconfigures TLS configuration i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates of a `ProxySQL`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql)
  - [ProxySQLOpsRequest](/docs/guides/proxysql/concepts/opsrequest)

## How Reconfiguring ProxySQL TLS Configuration Process Works

The following diagram shows how KubeDB Enterprise operator reconfigures TLS of a `ProxySQL`. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring TLS process of ProxySQL" src="/docs/guides/proxysql/reconfigure-tls/overview/images/reconfigure-tls.png">
<figcaption align="center">Fig: Reconfiguring TLS process of ProxySQL</figcaption>
</figure>

The Reconfiguring ProxySQL TLS process consists of the following steps:

1. At first, a user creates a `ProxySQL` Custom Resource Object (CRO).

2. `KubeDB` Community operator watches the `ProxySQL` CRO.

3. When the operator finds a `ProxySQL` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the TLS configuration of the `ProxySQL`, the user creates a `ProxySQLOpsRequest` CR with desired information.

5. `KubeDB` Enterprise operator watches the `ProxySQLOpsRequest` CR.

6. When it finds a `ProxySQLOpsRequest` CR, it pauses the `ProxySQL` object which is referred from the `ProxySQLOpsRequest`. So, the `KubeDB` Community operator doesn't perform any operations on the `ProxySQL` object during the reconfiguring TLS process.  

7. Then the `KubeDB` Enterprise operator will add, remove, update or rotate TLS configuration based on the Ops Request yaml.

8. Then the `KubeDB` Enterprise operator will restart all the Pods of the server so that they restart with the new TLS configuration defined in the `ProxySQLOpsRequest` CR.

9. After the successful reconfiguring of the `ProxySQL` TLS, the `KubeDB` Enterprise operator resumes the `ProxySQL` object so that the `KubeDB` Community operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on reconfiguring TLS configuration of a ProxySQL using `ProxySQLOpsRequest` CRD.