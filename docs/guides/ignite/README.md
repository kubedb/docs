---
title: Ignite
menu:
  docs_{{ .version }}:
    identifier: ig-readme-ignite
    name: Ignite
    parent: ignite-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/ignite/
aliases:
  - /docs/{{ .version }}/guides/ignite/README/
---
> New to KubeDB? Please start [here](/docs/README.md).
## Supported Ignite Features

| Features                                          | Availability |
| ------------------------------------------------- | :----------: |
| Clustering                                        |   &#10003;   |
| Persistent Volume                                 |   &#10003;   |
| Multiple Ignite Versions                          |   &#10003;   |
| Custom Configuration                              |   &#10003;   |
| Externally manageable Auth Secret                 |   &#10003;   |
| Reconfigurable Health Checker                     |   &#10003;   |
| Using Custom docker image                         |   &#10003;   |
| Builtin Prometheus Discovery                      |   &#10003;   |
| Using Prometheus operator                         |   &#10003;   |
| Automated Horizontal Scaling                      |   &#10003;   |
| Automatic Vertical Scaling                        |   &#10003;   |
| Automated Volume Expansion                        |   &#10003;   |
| Autoscaling (compute, storage)                    |   &#10003;   |
| Reconfigure                                       |   &#10003;   |
| TLS: Add, Remove, Update, Rotate ( Cert Manager ) |   &#10003;   |
| Restart                                           |   &#10003;   |
| Custom RBAC                                       |   &#10003;   |



## Life Cycle of a Ignite Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/ignite/ignite-lifecycle.png">
</p>

## User Guide
- [Quickstart Ignite](/docs/guides/ignite/quickstart/quickstart.md) with KubeDB Operator.
- Monitor your Ignite server with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/ignite/monitoring/using-prometheus-operator.md).
- Monitor your Ignite server with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/ignite/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/ignite/private-registry/using-private-registry.md) to deploy Ignite with KubeDB.
- Use [Custom Configuration](/docs/guides/ignite/custom-configuration/using-config-file.md) to configure Ignite.
- Use [Custom RBAC](/docs/guides/ignite/custom-rbac/using-custom-rbac.md) to run Ignite with custom RBAC resources.
- [Horizontal Scale](/docs/guides/ignite/scaling/horizontal-scaling/horizontal-scaling.md) your Ignite cluster with KubeDB Ops Manager.
- [Vertical Scale](/docs/guides/ignite/scaling/vertical-scaling/vertical-scaling.md) your Ignite cluster with KubeDB Ops Manager.
- [Volume Expansion](/docs/guides/ignite/volume-expansion/volume-expansion.md) of your Ignite cluster with KubeDB Ops Manager.
- [Reconfigure](/docs/guides/ignite/reconfigure/reconfigure.md) your Ignite cluster with KubeDB Ops Manager.
- [Reconfigure TLS/SSL](/docs/guides/ignite/reconfigure-tls/reconfigure-tls.md) of your Ignite cluster with KubeDB Ops Manager.
- [Autoscale Compute Resources](/docs/guides/ignite/autoscaler/compute/compute-autoscale.md) of your Ignite cluster with KubeDB Autoscaler.
- [Autoscale Storage](/docs/guides/ignite/autoscaler/storage/storage-autoscale.md) of your Ignite cluster with KubeDB Autoscaler.
- [Restart](/docs/guides/ignite/restart/restart.md) your Ignite cluster with KubeDB Ops Manager.
- Detail concepts of [Ignite object](/docs/guides/ignite/concepts/ignite.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).