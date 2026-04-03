---
title: MongoDB Gitops Overview
description: MongoDB Gitops Overview
menu:
  docs_{{ .version }}:
    identifier: es-gitops-overview
    name: Overview
    parent: es-gitops
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Overview for MongoDB

This guide will give you an overview of how KubeDB `gitops` operator works with MongoDB databases using the `gitops.kubedb.com/v1alpha1` API. It will help you understand the GitOps workflow for
managing MongoDB databases in Kubernetes.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [MongoDB](/docs/guides/MongoDB/concepts/MongoDB/index.md)
    - [ElasticsearchOpsRequest](/docs/guides/MongoDB/concepts/MongoDB-ops-request/index.md)



## Workflow GitOps with MongoDB

The following diagram shows how the `KubeDB` GitOps Operator used to sync with your database. Open the image in a new tab to see the enlarged version.


<figure align="center">

  <img alt="GitOps Flow" src="/docs/images/gitops/gitops.png">

<figcaption align="center">Fig: GitOps process of MongoDB</figcaption>

</figure>

1. **Define GitOps MongoDB**: Create Custom Resource (CR) of kind `MongoDB` using the `gitops.kubedb.com/v1alpha1` API.
2. **Store in Git**: Push the CR to a Git repository.
3. **Automated Deployment**: Use a GitOps tool (like `ArgoCD` or `FluxCD`) to monitor the Git repository and synchronize the state of the Kubernetes cluster with the desired state defined in Git.
4. **Create Database**: The GitOps operator creates a corresponding KubeDB MongoDB CR in the Kubernetes cluster to deploy the database.
5. **Handle Updates**: When you update the ElasticsearchGitOps CR, the operator generates an Ops Request to safely apply the update(e.g. `VerticalScaling`, `HorizontalScaling`, `VolumeExapnsion`, `Reconfigure`, `RotateAuth`, `ReconfigureTLS`, `VersionUpdate`, ans `Restart`.

This flow makes managing MongoDB databases efficient, reliable, and fully integrated with GitOps practices.

In the next doc, we are going to show a step by step guide on running MongoDB using GitOps.
