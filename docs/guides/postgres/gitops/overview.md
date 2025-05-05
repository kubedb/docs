---
title: PostgreSQL GitOps Overview
description: PostgreSQL GitOps Overview
menu:
  docs_{{ .version }}:
    identifier: pg-gitops-overview
    name: Overview
    parent: pg-gitops-postgres
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Overview for PostgreSQL

This guide will give you an overview of how KubeDB gitops operator works with PostgreSQL databases using the `gitops.kubedb.com/v1alpha1` API. It will help you understand the GitOps workflow for managing PostgreSQL databases in Kubernetes.

What is GitOps for PostgreSQL?

GitOps is a practice that uses Git as the single source of truth for managing PostgreSQL databases in Kubernetes. With KubeDB’s GitOps support, you define your PostgreSQL database configuration in a Git repository, and a GitOps operator automatically applies and maintains it in your cluster.

How It Works





Define PostgreSQL Configuration:





You create a Custom Resource (CR) of kind PostgresGitOps using the gitops.kubedb.com/v1alpha1 API.



This CR mirrors the specs of a standard KubeDB PostgreSQL CR (e.g., version, replicas, storage).



Store in Git:





Commit the PostgresGitOps CR to a Git repository managed by a GitOps tool like ArgoCD or FluxCD.



Automated Deployment:





The KubeDB GitOps operator detects the PostgresGitOps CR in Git.



It creates a corresponding KubeDB Postgres CR in the Kubernetes cluster to deploy the database.



Handle Updates:





When you update the PostgresGitOps CR (e.g., change version, resources, or configs), the operator notices the change.



Instead of directly modifying the Postgres CR, it generates an Ops Request to safely apply the update, ensuring stability.



Continuous Sync:





The GitOps tool (ArgoCD/FluxCD) continuously monitors the Git repository.



It ensures the cluster’s state matches the desired state in Git, correcting any drift.

Benefits





Simplified Management: Update databases by editing Git files—no manual Ops Requests needed.



Automation: The GitOps operator handles CR creation and Ops Requests automatically.



Consistency: Git ensures versioned, auditable changes.



Team Collaboration: Pull requests enable reviews and teamwork.

This flow makes managing PostgreSQL databases efficient, reliable, and fully integrated with GitOps practices.