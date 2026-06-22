---
title: Oracle TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-tls-overview
    name: Overview
    parent: guides-oracle-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Oracle TLS/SSL Encryption

**Prerequisite :** To configure TLS/SSL in `Oracle`, `KubeDB` uses [cert-manager](https://cert-manager.io/) to issue certificates. So, first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

To issue a certificate, the following number of CRDs of `cert-manager` is used:

- `Issuer/ClusterIssuer`: The issuer or cluster issuer refers to the certificate authority (CA) that signs the certificates. KubeDB uses the issuer referenced through `spec.tcpsConfig.tls.issuerRef` to generate the certificates required for the Oracle database.
- `Certificate`: The KubeDB operator creates `Certificate` objects for the database. cert-manager then creates the corresponding TLS `Secret` containing `tls.crt`, `tls.key`, and `ca.crt`.

## How TLS/SSL configures in Oracle

Oracle uses **TCPS** (TCP with SSL/TLS — Oracle Net over TLS) to encrypt client/server traffic. When TLS is enabled, the plaintext SQL\*Net listener stays on port `1521` and an additional encrypted **TCPS listener** is exposed on port `2484`. The KubeDB operator turns the cert-manager issued certificates into an Oracle **auto-login wallet** that the database and clients use to establish the TLS handshake.

The following figure shows how the KubeDB operator configures TLS/SSL on an Oracle database.

The steps the operator performs are:

1. **Users create an `Issuer`/`ClusterIssuer`** (backed by a CA secret) that will sign the Oracle certificates.

2. **Users deploy an `Oracle` CR** with `spec.tcpsConfig` set, referencing the issuer through `spec.tcpsConfig.tls.issuerRef` and (optionally) a TCPS listener port through `spec.tcpsConfig.tcpsListener.port` (defaults to `2484`).

3. **The KubeDB operator watches the `Oracle` CR**. When it finds `spec.tcpsConfig`, it creates three cert-manager `Certificate` objects:
    - a **server** certificate (`<db-name>-server-cert`) used by the database listener,
    - a **client** certificate (`<db-name>-client-cert`, common name `sys`) used for mutual TLS,
    - a **metrics-exporter** certificate (`<db-name>-metrics-exporter-cert`) used by the monitoring exporter.

4. **cert-manager issues the certificates** and stores them in Kubernetes `Secret`s, which the operator mounts into the database pod.

5. **Inside the pod**, the bootstrap scripts build an Oracle auto-login wallet from the certificates, configure `sqlnet.ora`/`listener.ora`/`tnsnames.ora` for the TCPS listener on port `2484` (with `SSL_VERSION=1.2` and mutual TLS), and then publish the wallet as a Kubernetes `Secret` named `<db-name>-tls-wallet`. Clients mount this wallet secret to connect over TCPS.

> Note: If the referenced `Issuer`/`ClusterIssuer` is not present (or not `Ready`), the Oracle database will stay in the `Provisioning` phase until the issuer becomes available.

### Oracle CRD Specification for TLS

The relevant portion of the `Oracle` CRD that controls TLS/SSL is `spec.tcpsConfig`:

```yaml
spec:
  tcpsConfig:
    tls:
      issuerRef:
        apiGroup: cert-manager.io
        kind: Issuer
        name: oracle-ca-issuer
    tcpsListener:
      port: 2484
```

Here,

- `spec.tcpsConfig.tls.issuerRef` is a reference to the `Issuer` or `ClusterIssuer` used to issue the database certificates. It has the following fields:
    - `apiGroup` — the group name of the resource being referenced. The value for `Issuer` or `ClusterIssuer` is `cert-manager.io`.
    - `kind` — the type of resource being referenced. KubeDB supports both `Issuer` and `ClusterIssuer`.
    - `name` — the name of the referenced `Issuer`/`ClusterIssuer`.
- `spec.tcpsConfig.tcpsListener.port` is the port the encrypted TCPS listener binds to. It defaults to `2484`.

In the next doc, we are going to show a step by step guide on how to configure a TLS/SSL enabled Oracle database using KubeDB.

## Next Steps

- Deploy a TLS/SSL secured [Oracle database with KubeDB](/docs/guides/oracle/tls/configure/index.md).
- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
