---
title: Virtual Secret Guide Postgres
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-virtual-secret-guide
    name: Virtual Secret  Guide                         
    parent: guides-postgres-virtual-secret
    weight: 10
menu_name: docs_{{ .version }}
---

[//]: # (> New to KubeDB? Please start [here]&#40;/docs/README.md&#41;.)

[//]: # ()
[//]: # (# Postgres Virtual Secret)

[//]: # ()
[//]: # ()
[//]: # (## A Guide to KubeDB's High Availability and Auto-Failover)

[//]: # ()
[//]: # (In today's data-driven world, database downtime is not just an inconvenience; it can be a critical business failure. For teams running stateful applications on Kubernetes, ensuring the resilience of their databases is paramount. This is where KubeDB steps in, offering a robust, cloud-native way to manage PostgreSQL on Kubernetes.)

[//]: # ()
[//]: # (One of KubeDB's most powerful features is its built-in support for High)

[//]: # (Availability &#40;HA&#41; and automated failover. The KubeDB operator continuously)

[//]: # (monitors the health of your PostgreSQL cluster and along with the db sidecar injected)

[//]: # (for maintaining failover, it can automatically)

[//]: # (respond to failures, ensuring your database remains available with)

[//]: # (minimal disruption.)

[//]: # ()
[//]: # (This article will guide you through KubeDB's automated failover capabilities for PostgreSQL. We will set up an HA cluster and then simulate a leader failure to see KubeDB's auto-recovery mechanism in action.)

[//]: # ()
[//]: # (> You will see how fast the failover happens when it's truly necessary. Failover in KubeDB-managed PostgreSQL will generally happen within 2–10 seconds depending on your cluster networking. There is an exception scenario that we discussed later in this doc where failover might take a bit longer up to 45 seconds. But that is a bit rare though.)

[//]: # ()
[//]: # (### Before You Start)

[//]: # ()
[//]: # (To follow along with this tutorial, you will need:)

[//]: # ()
[//]: # (1. A running Kubernetes cluster.)

[//]: # (2. KubeDB [installed]&#40;https://kubedb.com/docs/v2025.5.30/setup/install/kubedb/&#41; in your cluster.)

[//]: # (3. kubectl command-line tool configured to communicate with your cluster.)

[//]: # ()
