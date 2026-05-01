---
title: Reconfiguring Oracle TLS
menu:
  docs_{{ .version }}:
    identifier: oracle-reconfigure-tls-overview
    name: Overview
    parent: oracle-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring TLS for Oracle

This guide will give an overview of how KubeDB Ops-manager reconfigures TLS for a `Oracle` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)
- Use the example files from `docs/examples/oracle/quickstart/distributed.yaml` and `docs/examples/oracle/reconfigure-tls/ops-request.yaml`.

```bash
kubectl create ns demo
```

## Deploy Oracle

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/quickstart/distributed.yaml
kubectl get oracle -n demo oracle-sample -w
```

## How Reconfigure TLS Works

The Reconfigure TLS process consists of the following steps:

1. At first, a user creates a `Oracle` CR.

2. `KubeDB-Provisioner` operator watches the `Oracle` CR.

3. When the operator finds a `Oracle` CR, it creates a `StatefulSet` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure TLS of the `Oracle` database, the user creates a `OracleOpsRequest` CR specifying the desired TLS configuration. The user can add TLS to an existing non-TLS database, rotate the existing certificates, or remove TLS entirely.

5. `KubeDB` Ops-manager operator watches the `OracleOpsRequest` CR.

6. When it finds a `OracleOpsRequest` CR, it pauses the `Oracle` object so that the `KubeDB-Provisioner` operator doesn't perform any operations on the `Oracle` during the TLS reconfiguration process.

7. Then the `KubeDB` Ops-manager operator updates the TLS secrets and restarts the pods in a rolling fashion with the new TLS configuration.

8. After the successful TLS reconfiguration, the `KubeDB` Ops-manager updates the `Oracle` object to reflect the updated TLS state.

9. After the successful Reconfigure TLS, the `KubeDB` Ops-manager resumes the `Oracle` object so that the `KubeDB-Provisioner` resumes its usual operations.

In the next doc, we are going to show a step-by-step guide on reconfiguring TLS for a Oracle database using `OracleOpsRequest` CRD.
