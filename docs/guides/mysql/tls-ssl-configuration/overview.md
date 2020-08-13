---
title: MySQL TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: my-tls-overview
    name: Overview
    parent: my-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> :warning: **This doc is only for KubeDB Enterprise**: You need to be an enterprise user!

# MySQL TLS/SSL Encryption

**Prerequisite :** To configure TLS/SSL in `MySQL`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

To issue a certificate, the following crd of `cert-manager` is used:

- `Issuer/ClusterIssuer` : Issuers, and ClusterIssuers represent certificate authorities (CAs) that are able to generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer that is in a ready condition to attempt to honor the request. You can learn more details [here](https://cert-manager.io/docs/concepts/issuer/).

- `Certificate` : `cert-manager` has the concept of Certificates that define a desired x509 certificate which will be renewed and kept up to date. You can learn more details [here](https://cert-manager.io/docs/concepts/certificate/).

**MySQL CRD Specification :**

KubeDB uses following crd fields to enable SSL/TLS encryption in `MySQL`.

- `spec:`
  - `requireSSL`
  - `tls:`
    - `issuerRef`
    - `certificates`

Read about the fields in details from [mysql concept](/docs/concepts/databases/mysql.md#),

When, `requireSSL` is set, the users must specify the `tls.issuerRef` field. `KubeDB` uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets using `Issuer/ClusterIssuers` specification. These certificates secrets including `ca.crt`, `tls.crt` and `tls.key` etc. are used to configure `MySQL` server, exporter etc. respectively.

## How TLS/SSL configures in MySQL

The following figure shows how `KubeDB` enterprise used to configure TLS/SSL in MySQL. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Stash Backup Flow" src="/docs/images/day-2-operation/mysql/my-tls-ssl.png">
<figcaption align="center">Fig: Deploy MySQL with TLS/SSL</figcaption>
</figure>

Deploying MySQL with TLS/SSL configuration process consists of the following steps:

1. At first, a user creates a `Issuer/ClusterIssuer` cr.

2. Then the use creates a `MySQL` cr.

3. `KubeDB` community operator watches for the `MySQL` cr.

4. When it finds one, it creates `Secret`, `Service`, etc. for the `MySQL` database.

5. `KubeDB` enterprise operator watches for `MySQL`(5c), `Issuer/ClusterIssuer`(5b), `Secret` and `Service`(5a).

6. When it finds all the resources(`MySQL`, `Issuer/ClusterIssuer`, `Secret`, `Service`), it creates `Certificates` by using `tls.issuerRef` and `tls.certificates` field specification from `MySQL` cr.

7. `cert-manager` watches for certificates.

8. When it finds one, it creates certificate secrets `tls-secrets`(server, client, exporter secrets etc.) that holds the actual self-signed certificate.

9. `KubeDB` community operator watches for the Certificate secrets `tls-secrets`.

10. When it finds all the tls-secret, it creates a `StatefulSet` so that MySQL server is configured with TLS/SSL.

In the next doc, we are going to show a step by step guide on how to configure a `MySQL` database with TLS/SSL.