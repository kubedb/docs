---
title: Distributed MariaDB Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-distributed-autoscaling-storage-overview
    name: Overview
    parent: guides-mariadb-distributed-autoscaling-storage
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Distributed MariaDB Storage Autoscaling

This guide will give an overview on how KubeDB Autoscaler operator autoscales the database storage of a **distributed** MariaDB cluster using `mariadbautoscaler` crd.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [MariaDB](/docs/guides/mariadb/concepts/mariadb)
  - [MariaDBAutoscaler](/docs/guides/mariadb/concepts/autoscaler)
  - [MariaDBOpsRequest](/docs/guides/mariadb/concepts/opsrequest)
  - [Distributed MariaDB Overview](/docs/guides/mariadb/distributed/overview)

## How Storage Autoscaling Works

The following diagram shows how KubeDB Autoscaler operator autoscales the resources of `MariaDB` database components. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Storage Autoscaling process of MariaDB" src="/docs/guides/mariadb/autoscaler/storage/overview/images/mdas-storage.jpeg">
<figcaption align="center">Fig: Storage Autoscaling process of MariaDB</figcaption>
</figure>

The Auto Scaling process consists of the following steps:

1. At first, the user creates a `PlacementPolicy` Custom Resource (CR) with `monitoring.prometheus.url` configured for each spoke cluster. This allows the autoscaler to monitor storage usage across all clusters where MariaDB pods are running.

2. The user creates a `MariaDB` Custom Resource (CR) with `spec.distributed: true` and a reference to the `PlacementPolicy`.

3. `KubeDB` Community operator watches the `MariaDB` CR.

4. When the operator finds a `MariaDB` CR, it creates required number of `PetSets` and distributes them across spoke clusters as defined by the `PlacementPolicy`.

5. Each PetSet creates a Persistent Volume according to the Volume Claim Template provided in the petset configuration. This Persistent Volume will be expanded by the `KubeDB` Enterprise operator.

6. Then, in order to set up storage autoscaling of the `MariaDB` database the user creates a `MariaDBAutoscaler` CRO with desired configuration.

7. `KubeDB` Autoscaler operator watches the `MariaDBAutoscaler` CRO.

8. `KubeDB` Autoscaler operator continuously watches persistent volumes of the databases across all spoke clusters to check if storage usage exceeds the specified threshold. It queries the Prometheus endpoints configured per cluster in the `PlacementPolicy` to collect storage metrics.

9. If the usage exceeds the specified usage threshold, then `KubeDB` Autoscaler operator creates a `MariaDBOpsRequest` to expand the storage of the database.

10. `KubeDB` Enterprise operator watches the `MariaDBOpsRequest` CRO.

11. Then the `KubeDB` Enterprise operator will expand the storage of the database component as specified on the `MariaDBOpsRequest` CRO.

> **Key Difference from Non-Distributed Autoscaling**: For distributed MariaDB, the `PlacementPolicy` must include a `monitoring.prometheus.url` for each spoke cluster's `distributionRules` entry. The autoscaler uses these Prometheus endpoints to collect storage usage metrics from pods running across multiple Kubernetes clusters.

In the next docs, we are going to show a step by step guide on Autoscaling storage of a Distributed MariaDB database using `MariaDBAutoscaler` CRD.
