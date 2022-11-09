---
title: ProxySQL TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-tls-overview
    name: Overview
    parent: guides-proxysql-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# ProxySQL TLS/SSL Encryption

**Prerequisite :** To configure TLS/SSL in `ProxySQL`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

To issue a certificate, the following cr of `cert-manager` is used:

- `Issuer/ClusterIssuer`: Issuers and ClusterIssuers represent certificate authorities (CAs) that are able to generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer that is in a ready condition to attempt to honor the request. You can learn more details [here](https://cert-manager.io/docs/concepts/issuer/).

- `Certificate`: `cert-manager` has the concept of Certificates that define the desired x509 certificate which will be renewed and kept up to date. You can learn more details [here](https://cert-manager.io/docs/concepts/certificate/).

**ProxySQL CRD Specification:**

KubeDB uses the following cr fields to enable SSL/TLS encryption in `ProxySQL`.

- `spec:`
  - `tls:`
    - `issuerRef`
    - `certificates`

Read about the fields in details from [proxysql concept](/docs/guides/proxysql/concepts/proxysql/#spectls),

`KubeDB` uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets using `Issuer/ClusterIssuers` specification. These certificates secrets including `ca.crt`, `tls.crt` and `tls.key` etc. are used to configure `ProxySQL` server, exporter etc. respectively.

## How TLS/SSL configures in ProxySQL

The following figure shows how `KubeDB` enterprise is used to configure TLS/SSL in ProxySQL. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Stash Backup Flow" src="/docs/guides/proxysql/tls/overview/images/proxy-tls-ssl.png">
<figcaption align="center">Fig: Deploy ProxySQL with TLS/SSL</figcaption>
</figure>

Deploying ProxySQL with TLS/SSL configuration process consists of the following steps:

1. At first, a user creates an `Issuer/ClusterIssuer` cr.

2. Then the user creates a `ProxySQL` cr.

3. `KubeDB` community operator watches for the `ProxySQL` cr.

4. When it finds one, it creates `Secret`, `Service`, etc. for the `ProxySQL` server.

5. `KubeDB` enterprise operator watches for `ProxySQL`(5c), `Issuer/ClusterIssuer`(5b), `Secret` and `Service`(5a).

6. When it finds all the resources(`ProxySQL`, `Issuer/ClusterIssuer`, `Secret`, `Service`), it creates `Certificates` by using `tls.issuerRef` and `tls.certificates` field specification from `ProxySQL` cr.

7. `cert-manager` watches for certificates.

8. When it finds one, it creates certificate secrets `tls-secrets`(server, client, exporter secrets, etc.) that hold the actual self-signed certificate.

9. `KubeDB` community operator watches for the Certificate secrets `tls-secrets`.

10. When it finds all the tls-secret, it creates a `StatefulSet` so that ProxySQL server is configured with TLS/SSL.

In the next doc, we are going to show a step by step guide on how to configure a `ProxySQL` server with TLS/SSL.
