---
title: Reconfigure TLS Overview
menu:
  docs_{{ .version }}:
    identifier: dw-reconfigure-tls-overview
    name: Overview
    parent: dw-reconfigure-tls-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure TLS Overview

A `ReconfigureTLS` ops request changes the TLS configuration of a DurableWorkflow. KubeDB uses [cert-manager](https://cert-manager.io/docs/) to issue the certificates, see [TLS](/docs/guides/durableworkflow/tls/tls.md).

| Change                                                | Effect                                                                                                         | Downtime                              |
|-------------------------------------------------------|----------------------------------------------------------------------------------------------------------------|---------------------------------------|
| Rotate the certificates (`rotateCertificates: true`)   | The certificate Secrets are deleted and cert-manager issues new ones.                                           | none on a highly available cluster    |
| Update a certificate (`certificates`, same issuer)     | The entry of the alias is replaced and the certificate is issued again.                                         | none on a highly available cluster    |
| **Enable TLS** (`issuerRef` on a cluster without TLS)  | The certificates are issued and **all pods are replaced at once**.                                              | **yes**                               |
| **Remove TLS** (`remove: true`)                        | All pods are replaced at once, and the certificates are deleted.                                                | **yes**                               |
| **Change the issuer** (another `issuerRef`)            | All pods are replaced at once.                                                                                  | **yes**                               |

Enabling or removing TLS and changing the issuer cannot be rolled: a pod with TLS and a pod without it, or with certificates of another CA, cannot talk to each other. The webhook returns a warning for these three, for example `enabling TLS replaces every server pod at once, so the cluster is unavailable while they start`. Schedule them for a maintenance window.

## Rotation per engine

- **Temporal** reloads the certificates from disk, so the new certificates are used without restarting any pod. The request finishes when the new certificates exist.
- **Cadence** reads the certificates at start-up only. The pods are therefore [rolled one at a time](/docs/guides/durableworkflow/update-version/overview.md#how-the-pods-are-rolled) with the new certificates, which is not disruptive on a highly available cluster. Cadence needs a patched server build for TLS, see [TLS](/docs/guides/durableworkflow/tls/tls.md#cadence).

Certificate rotation was verified with 0 failed calls on both engines with TLS.

Cert-manager renews a certificate by itself before `renewBefore`. Because Cadence reads the certificates only at start-up, run a rotation request (or a Restart) for a Cadence cluster before the old certificate expires. The [recommendation engine](/docs/guides/durableworkflow/recommendation/recommendation.md) proposes a rotation when a certificate is close to its expiry.

## The request

- `tls.issuerRef` enables TLS or changes the issuer. It is required to enable TLS.
- `tls.certificates` updates certificates by `alias` (`internode`, `frontend`, `client`).
- `tls.rotateCertificates: true` reissues all the certificates.
- `tls.remove: true` removes TLS. It cannot be combined with the others, and it needs TLS to be on.

For Cadence, `spec.frontendClientAuth` must be `NotRequired` before TLS is enabled, and the version must have `publicClientTLS: true`.

## Next Steps

- Run a [Reconfigure TLS](/docs/guides/durableworkflow/reconfigure_tls/durableworkflow.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
