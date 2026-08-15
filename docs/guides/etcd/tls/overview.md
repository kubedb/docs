---
title: Etcd TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-tls-overview
    name: Overview
    parent: etcd-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd TLS/SSL Encryption

**Prerequisite :** To configure TLS/SSL in `Etcd`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster, follow the steps [here](https://cert-manager.io/docs/installation/).

To issue a certificate, the following CRDs of `cert-manager` are used:

- `Issuer/ClusterIssuer`: Issuers and ClusterIssuers represent certificate authorities (CAs) that are able to generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer that is in a ready condition to attempt to honor the request. You can learn more details [here](https://cert-manager.io/docs/concepts/issuer/).

- `Certificate`: `cert-manager` has the concept of Certificates that define a desired x509 certificate which will be renewed and kept up to date. You can learn more details [here](https://cert-manager.io/docs/concepts/certificate/).

**Etcd CRD Specification :**

KubeDB uses the following CRD fields to enable SSL/TLS encryption in `Etcd`.

- `spec:`
    - `tls:`
        - `issuerRef`
        - `certificates`

Read about the fields in detail in the [Etcd concept](/docs/guides/etcd/concepts/etcd.md#spectls).

There is **no separate `enableSSL` switch** for etcd. Setting `spec.tls` is what turns TLS on, and the webhook then requires `spec.tls.issuerRef` to be set. `KubeDB` uses the `Issuer` or `ClusterIssuer` referenced in `spec.tls.issuerRef`, together with the certificate specs provided in `spec.tls.certificates`, to generate certificate secrets. Each of those secrets holds `ca.crt`, `tls.crt` and `tls.key`, and is mounted into the etcd member pods.

## Which certificate secures what

etcd is not a single-listener database. A member listens on three separate sockets, and each of them has its own trust story. That is why KubeDB issues four certificates rather than one, addressed by `alias`:

| Alias              | Default secret name                    | Mount path                        | Secures                                                                                                |
|--------------------|----------------------------------------|-----------------------------------|--------------------------------------------------------------------------------------------------------|
| `server`           | `<db-name>-server-cert`                | `/var/run/etcd/tls/server`        | The **client API** listener on port `2379` — everything your applications and `etcdctl` talk to.        |
| `client`           | `<db-name>-client-cert`                | `/var/run/etcd/tls/client`        | KubeDB's **own** connections to etcd: the health checker, the ops-manager operator and the backup tooling. |
| `peer`             | `<db-name>-peer-cert`                  | `/var/run/etcd/tls/peer`          | **Member-to-member (Raft) traffic** on port `2380`.                                                      |
| `metrics-exporter` | `<db-name>-metrics-exporter-cert`      | `/var/run/etcd/tls/exporter`      | The Prometheus metrics endpoint. Only issued when `spec.monitor` is configured.                          |

A few properties follow from etcd's own flag semantics and are worth internalising before you enable TLS:

- **The client API is always mutually authenticated.** When `spec.tls` is set, KubeDB renders `--client-cert-auth=true` on every member. An anonymous TLS client — one that trusts the CA but presents no certificate of its own — is rejected. Anything that talks to port `2379` must present a certificate signed by the same CA.

- **Peer traffic is always mutual TLS.** KubeDB renders `--peer-client-cert-auth=true` alongside the peer certificate flags. etcd has no "encrypt peers but don't verify them" mode in this configuration, so there is no server-only option for the `peer` alias. This is deliberate: the Raft channel is the cluster's trust boundary, and a member that could join without proving its identity could join the quorum.

- **The `server` and `peer` certificates are issued with both `serverAuth` and `clientAuth` key usages.** Peers authenticate each other in both directions, and a member that is not the Raft leader forwards writes to the leader over the same channel — so both usages are required for either alias.

- **The `client` certificate is issued with `CN=root`.** etcd maps a client certificate's Common Name onto an etcd user when etcd RBAC is enabled, so KubeDB's operator certificate is issued for the `root` user. Do not reuse this certificate for your applications; issue your own from the same `Issuer` instead.

- **Enabling TLS changes every advertised URL from `http://` to `https://`.** That is a member-level restart, which is why turning TLS on or off on a *running* cluster is done through a [`ReconfigureTLS` EtcdOpsRequest](/docs/guides/etcd/reconfigure-tls/overview.md) rather than by editing the object in place.

- **The metrics listener stays plain HTTP.** KubeDB gives every member a dedicated metrics listener (`--listen-metrics-urls`) on port `2381`. Keeping it off the mutually authenticated client port is what allows kubelet's readiness probe — which has no client certificate to present — to keep working after TLS is enabled. See the [monitoring overview](/docs/guides/etcd/monitoring/overview.md).

## Subject alternative names

etcd validates the URL it dialled against the certificate the far end presents, and every member has its own ordinal DNS name behind the governing (headless) Service. KubeDB therefore issues the `server` and `peer` certificates with SANs covering both the load-balanced client Service and every member pod:

```
<db-name>
<db-name>.<namespace>
<db-name>.<namespace>.svc
<db-name>.<namespace>.svc.cluster.local
<db-name>-pods.<namespace>.svc
*.<db-name>-pods.<namespace>.svc
*.<db-name>-pods.<namespace>.svc.cluster.local
localhost
127.0.0.1
```

You do not need to add these by hand. If you need extra names — an external DNS record, for example — add them through `spec.tls.certificates[].dnsNames`; KubeDB merges your list with the ones above rather than replacing it.

## How TLS/SSL is configured in Etcd

Deploying an `Etcd` cluster with TLS/SSL enabled consists of the following steps:

1. At first, a user creates an `Issuer/ClusterIssuer` CR.

2. Then the user creates an `Etcd` CR which refers to the `Issuer/ClusterIssuer` CR that the user created in the previous step.

3. `KubeDB` Provisioner operator watches for the `Etcd` CR.

4. When it finds one, it creates the `Service`, `Secret`, RBAC objects, etc. for the `Etcd` cluster.

5. It then creates a cert-manager `Certificate` object for each alias etcd needs — `server`, `client`, `peer` and, when `spec.monitor` is set, `metrics-exporter` — using the `spec.tls.issuerRef` and `spec.tls.certificates` fields from the `Etcd` CR.

6. `cert-manager` watches for those `Certificate` objects.

7. When it finds them, it creates the certificate secrets that hold the actual certificates signed by the CA.

8. `KubeDB` Provisioner operator waits for every certificate secret to exist before it creates the `PetSet`, because etcd fails to start if a file named on its command line is missing.

9. Once all the secrets are present, the operator creates the `PetSet` with the TLS flags (`--cert-file`, `--key-file`, `--trusted-ca-file`, `--client-cert-auth`, `--peer-cert-file`, `--peer-key-file`, `--peer-trusted-ca-file`, `--peer-client-cert-auth`) and the matching volumes, so the etcd cluster comes up with TLS/SSL configured from the very first member.

In the [next](/docs/guides/etcd/tls/configure-ssl.md) doc, we are going to show a step-by-step guide on how to configure an `Etcd` cluster with TLS/SSL.
