---
title: MySQL TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: my-tls-overview
    name: Overview
    parent: my-tls
    weight: 20
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
    - `certificate`

Read about the fields in details in [mysql concept](/docs/concepts/databases/mysql.md),

When, `requireSSL` is set, the users must specify the `tls.issuerRef` field. `KubeDB` uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets using `Issuer/ClusterIssuers` specification. These certificate secrets including `ca.crt`, `server.pem` and `exporter.pem` etc. are used to configure `MySQL` server, exporter respectively.

The subject of `client.pem` certificate is added as `root` user in mysql database. So, user can use this client certificate for `MySQL-X509` `authenticationMechanism`.

## How TLS/SSL configures in MySQL

The following figure shows how to configure TLS/SSL in MySQL using `KubeDB` enterprise. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Stash Backup Flow" src="/docs/images/day-2-operation/tls-tls.png">
<figcaption align="center">Fig: Deploy MySQL with TLS/SSL</figcaption>
</figure>

Deploying MySQL with TLS/SSL configuration process consists of the following steps:

1. At first, a user creates a `Issuer/ClusterIssuer` crd.

2. Then the use creates a `MySQL` crd.

3. `KubeDB` community operator watches for `MySQL` crd.

4. When it finds one, it creates `Secret`, `Service`, etc. and watch for the `certificate Secrets` which will be used to create a StatefulSet so that MySQL server TLS/SSL is configured.

5. `KubeDB` enterprise operator watches for `MySQL`(5c), `Issuer/ClusterIssuer`(5b), `Secret` and `Service`(5a).

6. When it finds all resources(`MySQL`, `Issuer/ClusterIssuer`, `Secret`, `Service`), it creates `Certificates` using `Issuer` specification.

7. `cert-manager` watches for certificates.

8. When it finds one, it creates certificate secret `tls-secret` that holds the actual self-signed certificate.

9. When the `KubeDB` community operator finds the certificate secrets then will create a `StatefulSet` so that MySQL server is configured to TLS/SSL.