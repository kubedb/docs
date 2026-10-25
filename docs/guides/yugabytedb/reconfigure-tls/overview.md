---
title: Reconfiguring YugabyteDB TLS
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-reconfigure-tls-overview
    name: Overview
    parent: yugabytedb-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring YugabyteDB TLS

A `YugabyteDBOpsRequest` with `spec.type: ReconfigureTLS` can add TLS, rotate the certificates, change the issuer, or remove TLS.

| Request | What KubeDB does | Zero downtime? |
|---------|------------------|----------------|
| `tls.rotateCertificates: true` | cert-manager issues new certificates; masters and tservers reload them (`ReloadCertificates`), no restart | yes |
| `tls.issuerRef` changed | a probe certificate gets the new CA; the CA bundle is extended with the old plus new CA, the certificates are re-issued, then the bundle is reduced to the new CA; no restart | yes |
| `tls.issuerRef` on a plaintext cluster (add TLS) | rolling round one with node-to-node TLS and insecure connections still allowed, then the secure flags are enforced and the tservers are rolled again for client-to-node TLS | **no** |
| `tls.remove: true` | the reverse of add: tservers first, then masters | **no** |

Why adding or removing TLS is not zero downtime: on this YugabyteDB version a TLS server cannot connect to a plaintext server. While TLS and plaintext servers are mixed, the masters cannot reach the tservers (or the other way round), so the tablet leaders cannot be drained, the usual safety gates cannot be used (KubeDB falls back to the universe health and a heartbeat gate and records `LeaderDrainTimedOut`), and clients see connection failures and a stall of about 20 seconds. Plaintext clients are also refused once client-to-node TLS is on.

## Next Step

[Reconfigure TLS](/docs/guides/yugabytedb/reconfigure-tls/reconfigure-tls.md).
