---
title: GitOps overview Microsoft SQL Server
menu:
  docs_{{ .version }}:
    identifier: ms-overview
    name: Gitops overview Microsoft SQL Server
    parent: mssqlserver-gitops
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Overview for MSSQLServer

This guide will give you an overview of how KubeDB `gitops` operator works with MSSQLServer databases using the `gitops.kubedb.com/v1alpha1` API. It will help you understand the GitOps workflow for
managing MSSQLServer databases in Kubernetes.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
    - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)



## Workflow GitOps with MSSQLServer

The following diagram shows how the `KubeDB` GitOps Operator used to sync with your database. Open the image in a new tab to see the enlarged version.


<figure align="center">

  <img alt="GitOps Flow" src="/docs/images/gitops/gitops.png">

<figcaption align="center">Fig: GitOps process of MSSQLServer</figcaption>

</figure>

1. **Define GitOps MSSQLServer**: Create Custom Resource (CR) of kind `MSSQLServer` using the `gitops.kubedb.com/v1alpha1` API.
2. **Store in Git**: Push the CR to a Git repository.
3. **Automated Deployment**: Use a GitOps tool (like `ArgoCD` or `FluxCD`) to monitor the Git repository and synchronize the state of the Kubernetes cluster with the desired state defined in Git.
4. **Create Database**: The GitOps operator creates a corresponding KubeDB MSSQLServer CR in the Kubernetes cluster to deploy the database.
5. **Handle Updates**: When you update the MSSqlGitOps CR, the operator generates an Ops Request to safely apply the update(e.g. `VerticalScaling`, `HorizontalScaling`, `VolumeExapnsion`, `Reconfigure`, `RotateAuth`, `ReconfigureTLS`, `VersionUpdate`, ans `Restart`).

This flow makes managing MSSQLServer databases efficient, reliable, and fully integrated with GitOps practices.

In the next doc, we are going to show a step by step guide on running MSSQLServer using GitOps.
