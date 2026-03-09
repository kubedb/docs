---
title: Elasticsearch Recommendation
description: Elasticsearch Recommendation Overview
menu:
  docs_{{ .version }}:
    identifier: es-recommendation-overview
    name: Overview
    parent: es-recommendation-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Recommendation for KubeDB managed Elasticsearch and Opensearch

Databases on Kubernetes in production grade infrastructure often need to go through several administrative operations depending on specific resource requirements. Such operations include vertical scaling (cpu, memory) and storage expansion. Autoscaling support for KubeDB managed databases takes care of it. However, databases also need to go through some maintenance operations in order to ensure security, enhance performance, getting bug fixes and new features etc. Such operations mostly require organization's manual intervention. Even if these operations are automated, they need to be done in surveillance. KubeDB simplifies this by generating K8s Native Recommendations. 

## Overview

Recommendation is a custom resource definition (CRD) object which is created by KubeDB ops-manager controller and managed by supervisor. So, You need to have KubeDB and Supervisor installed first. You can simply install supervisor along with other KubeDB components using `--set supervisor.enabled=true` flag while installing KubeDB via helm chart.

<p align="center">
<img alt="Recommendation Generation"  src="/docs/guides/elasticsearch/recommendation/images/recommendation-generation.png">
</p>

KubeDB provisioner watches user provided database custom resource spec and creates/sync all the necessary DB resources. Once the Database is ready KubeDB Ops-manager watches the DB and creates Recommendation if it requires. KubeDB Supervisor then watches the Recommendation, updates status of the recommendation, creates recommended operation via OpsRequest if deadline reaches or manually triggered and watches the OpsRequest status to update accordingly in Recommendation custom resource.

KubeDB provides Three types of recommendation for Elasticsearch and Opensearch.

1. [Version Update Recommendation](/docs/guides/elasticsearch/recommendation/version-update-recommendation.md)
2. [TLS Certificate Rotation Recommendation](/docs/guides/elasticsearch/recommendation/rotate-tls-recommendation.md)
3. [Authentication Secret Rotation Recommendation](/docs/guides/elasticsearch/recommendation/rotate-auth-recommendation.md)

The next page describes these recommendations, how to approve/reject them, their generation mechanism and usability.

## Next Steps

- Learn about Elasticsearch [Version Update Recommendation](/docs/guides/elasticsearch/recommendation/version-update-recommendation.md).
- Learn about Elasticsearch [TLS Certificate Rotation Recommendation](/docs/guides/elasticsearch/recommendation/rotate-tls-recommendation.md)
- Learn about Elasticsearch [Authentication Secret Rotation Recommendation](/docs/guides/elasticsearch/recommendation/rotate-auth-recommendation.md)
