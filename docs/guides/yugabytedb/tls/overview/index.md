---
title: YugabyteDB TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-tls-overview
    name: Overview
    parent: yugabytedb-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB TLS/SSL Encryption

KubeDB uses [cert-manager](https://cert-manager.io/) to issue the certificates of a YugabyteDB universe. TLS is configured with `spec.tls`:

- `spec.tls.issuerRef` references an `Issuer` or `ClusterIssuer`.
- `spec.tls.nodeToNode` (default `true`) encrypts the traffic between masters and tservers.
- `spec.tls.clientToNode` (default `true`) encrypts the traffic between clients and tservers (YSQL and YCQL).

## What KubeDB creates

For a universe named `yb-demo` with TLS, KubeDB creates three cert-manager `Certificate` objects and a CA bundle:

| Secret | Content |
|--------|---------|
| `yb-demo-master-server-cert` | server key pair for the masters, with the pod FQDNs, the governing Service and `yb-demo.demo.svc` in the SANs |
| `yb-demo-tserver-server-cert` | the same for the tservers |
| `yb-demo-client-cert` | client key pair (common name `yugabyte`), used by the operator and available to your clients |
| `yb-demo-ca-bundle` | the trusted CAs, projected into the pods as `ca.crt` |

The pods mount the files at `/opt/certs/yugabyte` (`node.yugabyte.crt`, `node.yugabyte.key`, `ca.crt`).

## Behavior you must know

1. **Plaintext clients are refused** once client-to-node TLS is on. A plaintext `ysqlsh` fails with `no pg_hba.conf entry ... no encryption`.
2. **Rotating certificates and changing the issuer are online.** The pods reload the certificates without a restart. The measured results are in [Reconfigure TLS](/docs/guides/yugabytedb/reconfigure-tls/reconfigure-tls.md).
3. **Enabling or disabling TLS on a running cluster is NOT zero downtime** on this YugabyteDB version. A TLS server cannot connect to a plaintext server, so while TLS and plaintext servers are mixed the masters cannot reach the tservers, the leaders cannot be drained, and clients see connection failures and a stall of about 20 seconds. Plan a maintenance window, or enable TLS at creation time.
4. With TLS from the start, an RF3 universe is `Ready` in about 80 to 90 seconds.

## Next Steps

- [Configure TLS](/docs/guides/yugabytedb/tls/configure/)
- [Reconfigure TLS](/docs/guides/yugabytedb/reconfigure-tls/reconfigure-tls.md)
