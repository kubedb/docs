---
title: MySQL GitOps Overview
description: MySQL GitOps Overview
menu:
  docs_{{ .version }}:
    identifier: my-gitops-overview
    name: Overview
    parent: my-gitops-MySQL
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Overview for MySQL

This guide will give you an overview of how KubeDB `gitops` operator works with MySQL databases using the `gitops.kubedb.com/v1alpha1` API. It will help you understand the GitOps workflow for managing MySQL databases in Kubernetes.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [MySQL](/docs/guides/mysql/concepts/database/index.md)
    - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest/index.md)

## Workflow GitOps with MySQL

The following diagram shows how the `KubeDB` GitOps Operator used to sync with your database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="GitOps Flow" src="/docs/images/gitops/gitops.png">
<figcaption align="center">Fig: GitOps process of MySQL</figcaption>
</figure>

1. **Define GitOps MySQL**: Create Custom Resource (CR) of kind `MySQL` using the `gitops.kubedb.com/v1alpha1` API.
2. **Store in Git**: Push the CR to a Git repository.
3. **Automated Deployment**: Use a GitOps tool (like `ArgoCD` or `FluxCD`) to monitor the Git repository and synchronize the state of the Kubernetes cluster with the desired state defined in Git.
4. **Create Database**: The GitOps operator creates a corresponding KubeDB MySQL CR in the Kubernetes cluster to deploy the database.
5. **Handle Updates**: When you update the MySQLGitOps CR, the operator generates an Ops Request to safely apply the update(e.g. `VerticalScaling`, `HorizontalScaling`, `VolumeExapnsion`, `Reconfigure`, `RotateAuth`, `ReconfigureTLS`, `VersionUpdate`, ans `Restart`.

This flow makes managing MySQL databases efficient, reliable, and fully integrated with GitOps practices.

In the next doc, we are going to show a step by step guide on running MySQL using GitOps.
