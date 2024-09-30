---
title: Kafka Migration Overview
description: Kafka Migration Overview
menu:
  docs_{{ .version }}:
    identifier: kf-migration-overview
    name: Overview
    parent: kf-migration-kafka
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/guides/).

# Kafka Migration

This guide will help you to migrate your existing Kafka cluster to KubeDB or within KubeDB Kafka cluster.

## Before You Begin

- You should have familiar with the following `KubeDB` concepts:
    - [KubeDB Kafka](/docs/guides/kafka/concepts/kafka.md)
    - [ConnectCluster](/docs/guides/kafka/concepts/connectcluster.md)
    - [Connector](/docs/guides/kafka/concepts/connector.md)

Migration of Kafka cluster with minimal downtime is a challenging task. Proper planning and execution are required to ensure a smooth migration. The following things should be considered before starting the migration:

1. Evaluate the existing Kafka cluster and its dependencies.

2. Evaluate the existing Kafka cluster's data volume and its growth rate.

3. Validate network connectivity between the existing Kafka cluster and the new Kafka cluster.

4. Validate the compatibility of the existing Kafka cluster with the new Kafka cluster. Try upgrading the existing Kafka cluster to the recent version before starting the migration.

5. List down the Kafka topics and consumer groups that need to be migrated.

## Migration Scenarios

The following are the possible scenarios for Kafka migration:

1. User has an existing Kafka cluster and wants to migrate it to KubeDB Kafka cluster.

2. User creates a KubeDB Kafka cluster using `Kafka` Custom Resource(CR).

3. Setup security, monitoring, and users/ACLs in the new Kafka cluster.

4. Create `ConnectCluster` with reference to the new Kafka cluster.

5. Create `mirror-source`, `checkpoint` and `hearbeat` connectors to replicate data from the existing Kafka cluster to the new Kafka cluster.

6. Validate the data replication between the existing Kafka cluster and the new Kafka cluster.

7. Switch the application(produces/consumers) to use the new Kafka cluster.

8. Validate applications are functional with the new Kafka cluster.

9. Test load, security and performance of the new Kafka cluster.

In the next docs, we are going to show a step-by-step guide on how to migrate an existing Kafka cluster to KubeDB Kafka cluster.