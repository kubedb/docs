---
title: Overview Recommendation
menu:
  docs_{{ .version }}:
    identifier: overview-recommendation
    name: Overview
    parent: recommendation
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---


> New to KubeDB? Please start [here](/docs/README.md).

# Recommendation for KubeDB 
Databases on Kubernetes in production grade infrastructure often need to go through several administrative operations depending on specific resource requirements. Such operations include vertical scaling (cpu, memory) and storage expansion. Autoscaling support for KubeDB managed databases takes care of it. However, databases also need to go through some maintenance operations in order to ensure security, enhance performance, getting bug fixes and new features etc. Such operations mostly require organization's manual intervention. Even if these operations are automated, they need to be done in surveillance. KubeDB simplifies this by generating K8s Native Recommendations. 

## Overview

Recommendation is a custom resource definition (CRD) object which is created by KubeDB ops-manager controller and managed by supervisor. So, You need to have KubeDB and Supervisor installed first. You can simply install supervisor along with other KubeDB components using `--set supervisor.enabled=true` flag while installing KubeDB via helm chart.

<p align="center">
<img alt="Recommendation Generation"  src="/docs/operatormanual/recommendation/images/recommendation-generation.png">
</p>

KubeDB provisioner watches user provided database custom resource spec and creates/sync all the necessary DB resources. Once the Database is ready KubeDB Ops-manager watches the DB and creates Recommendation if it requires. KubeDB Supervisor then watches the Recommendation, updates status of the recommendation, creates recommended operation via OpsRequest if deadline reaches or manually triggered and watches the OpsRequest status to update accordingly in Recommendation custom resource.

KubeDB provides Three types of recommendation for KubeDB Databases:
1. [Version Update Recommendation](/docs/operatormanual/recommendation/version-update-recommendation.md)
2. [TLS Certificate Rotation Recommendation](/docs/operatormanual/recommendation/rotate-tls-recommendation.md)
3. [Authentication Secret Rotation Recommendation](/docs/operatormanual/recommendation/rotate-auth-recommendation.md)

## Recommendation Management

For detailed understanding of the recommendation system, refer to:

- [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md) - Complete field reference for Recommendation CRD
- [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) - Namespace-scoped scheduling for automatic operations
- [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md) - Cluster-wide default maintenance scheduling
- [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) - Linking maintenance windows to resources for automatic recommendation execution

The next pages describe these recommendations, how to approve/reject them, their generation mechanism and usability.

