---
title: Reconfiguring ProxySQL
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-reconfigure-overview
    name: Overview
    parent: guides-proxysql-reconfigure
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring ProxySQL

This guide will give an overview on how KubeDB Ops Manager reconfigures `ProxySQL`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [ProxySQL](/docs/guides/proxysql/concepts/proxysql)
  - [ProxySQLOpsRequest](/docs/guides/proxysql/concepts/opsrequest)

## How Reconfiguring ProxySQL Process Works

The following diagram shows how KubeDB Ops Manager reconfigures `ProxySQL` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Reconfiguring process of ProxySQL" src="/docs/guides/proxysql/reconfigure/overview/images/reconfigure.png">
<figcaption align="center">Fig: Reconfiguring process of ProxySQL</figcaption>
</figure>

The Reconfiguring ProxySQL process consists of the following steps:

1. At first, a user creates a `ProxySQL` Custom Resource (CR).

2. `KubeDB` Community operator watches the `ProxySQL` CR.

3. When the operator finds a `ProxySQL` CR, it creates required number of `StatefulSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the `ProxySQL` standalone or cluster the user creates a `ProxySQLOpsRequest` CR with desired information.

5. `KubeDB` Enterprise operator watches the `ProxySQLOpsRequest` CR.

6. Then the `KubeDB` Enterprise operator will replace the existing configuration with the new configuration provided or merge the new configuration with the existing configuration according to the `ProxySQLOpsRequest` CR.

In the next docs, we are going to show a step by step guide on reconfiguring ProxySQL database components using `ProxySQLOpsRequest` CRD.