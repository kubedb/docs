---
title: Postgres TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-tls-overview
    name: Overview
    parent: guides-postgres-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Postgres TLS/SSL Encryption

**Prerequisite :** To configure TLS/SSL in `Postgres`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

To issue a certificate, the following cr of `cert-manager` is used:

- `Issuer/ClusterIssuer`: Issuers and ClusterIssuers represent certificate authorities (CAs) that are able to generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer that is in a ready condition to attempt to honor the request. You can learn more details [here](https://cert-manager.io/docs/concepts/issuer/).

- `Certificate`: `cert-manager` has the concept of Certificates that define the desired x509 certificate which will be renewed and kept up to date. You can learn more details [here](https://cert-manager.io/docs/concepts/certificate/).

**Postgres CRD Specification:**

KubeDB uses the following cr fields to enable SSL/TLS encryption in `Postgres`.

- `spec:`
  - `sslMode`
  - `tls:`
    - `issuerRef`
    - `certificates`

Read about the fields in details from [postgres concept](/docs/guides/postgres/concepts/postgres.md#),

- `sslMode` supported values are [`disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full`]
  - `disable:`  It ensures that the server does not use TLS/SSL
  - `allow:` you don't care about security, but I will pay the overhead of encryption if the server insists on it.
  - `prefer:` you don't care about encryption, but you wish to pay the overhead of encryption if the server supports it.
  - `require:`  you want your data to be encrypted, and you accept the overhead. you want to be sure that you connect to a server that you trust.
  - `verify-ca:` you want your data to be encrypted, and you accept the overhead. you want to be sure that you connect to a server you trust, and that it's the one you specify.
   
When, `sslMode` is set and the value is not `disable` then, the users must specify the `tls.issuerRef` field. `KubeDB` uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets using `Issuer/ClusterIssuers` specification. These certificates secrets including `ca.crt`, `tls.crt` and `tls.key` etc. are used to configure `Postgres` server, exporter etc. respectively.

## How TLS/SSL configures in Postgres

The following figure shows how `KubeDB` enterprise is used to configure TLS/SSL in Postgres. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Postgres with TLS/SSL Flow" src="/docs/guides/postgres/tls/overview/images/pg-tls-ssl.png">
<figcaption align="center">Fig: Deploy Postgres with TLS/SSL</figcaption>
</figure>

Deploying Postgres with TLS/SSL configuration process consists of the following steps:

1. At first, a user creates an `Issuer/ClusterIssuer` cr.

2. Then the user creates a `Postgres` cr.

3. `KubeDB` community operator watches for the `Postgres` cr.

4. When it finds one, it creates `Secret`, `Service`, etc. for the `Postgres` database.

5. `KubeDB` enterprise operator watches for `Postgres`(5c), `Issuer/ClusterIssuer`(5b), `Secret` and `Service`(5a).

6. When it finds all the resources(`Postgres`, `Issuer/ClusterIssuer`, `Secret`, `Service`), it creates `Certificates` by using `tls.issuerRef` and `tls.certificates` field specification from `Postgres` cr.

7. `cert-manager` watches for certificates.

8. When it finds one, it creates certificate secrets `cert-secrets`(server, client, exporter secrets, etc.) that hold the actual self-signed certificate.

9. `KubeDB` community operator watches for the Certificate secrets `tls-secrets`.

10. When it finds all the tls-secret, it creates a `StatefulSet` so that Postgres server is configured with TLS/SSL.

In the next doc, we are going to show a step by step guide on how to configure a `Postgres` database with TLS/SSL.