---
title: Pgpool TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: pp-tls-overview
    name: Overview
    parent: pp-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Pgpool TLS/SSL Encryption

**Prerequisite :** To configure TLS/SSL in `Pgpool`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

To issue a certificate, the following crd of `cert-manager` is used:

- `Issuer/ClusterIssuer`: Issuers, and ClusterIssuers represent certificate authorities (CAs) that are able to generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer that is in a ready condition to attempt to honor the request. You can learn more details [here](https://cert-manager.io/docs/concepts/issuer/).

- `Certificate`: `cert-manager` has the concept of Certificates that define a desired x509 certificate which will be renewed and kept up to date. You can learn more details [here](https://cert-manager.io/docs/concepts/certificate/).

**Pgpool CRD Specification :**

KubeDB uses following crd fields to enable SSL/TLS encryption in `Pgpool`.

- `spec:`
  - `sslMode`
  - `tls:`
    - `issuerRef`
    - `certificates`
  - `cientAuthMode`
Read about the fields in details from [pgpool concept](/docs/guides/pgpool/concepts/pgpool.md),

When, `sslMode` is set to `require`, the users must specify the `tls.issuerRef` field. `KubeDB` uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets using `Issuer/ClusterIssuers` specification. These certificates secrets including `ca.crt`, `tls.crt` and `tls.key` etc. are used to configure `Pgpool` server, exporter etc. respectively.

## How TLS/SSL configures in Pgpool

The following figure shows how `KubeDB` enterprise used to configure TLS/SSL in Pgpool. Open the image in a new tab to see the enlarged version.

<figure align="center">
<img alt="Deploy Pgpool with TLS/SSL" src="/docs/images/day-2-operation/pgpool/pgpool-tls.svg">
<figcaption align="center">Fig: Deploy Pgpool with TLS/SSL</figcaption>
</figure>

Deploying Pgpool with TLS/SSL configuration process consists of the following steps:

1. At first, a user creates a `Issuer/ClusterIssuer` cr.

2. Then the user creates a `Pgpool` cr which refers to the `Issuer/ClusterIssuer` cr that the user created in the previous step.

3. `KubeDB` Provisioner  operator watches for the `Pgpool` cr.

4. When it finds one, it creates `Secret`, `Service`, etc. for the `Pgpool` database.

5. `KubeDB` Ops-manager operator watches for `Pgpool`(5c), `Issuer/ClusterIssuer`(5b), `Secret` and `Service`(5a).

6. When it finds all the resources(`Pgpool`, `Issuer/ClusterIssuer`, `Secret`, `Service`), it creates `Certificates` by using `tls.issuerRef` and `tls.certificates` field specification from `Pgpool` cr.

7. `cert-manager` watches for certificates.

8. When it finds one, it creates certificate secrets `tls-secrets`(server, client, exporter secrets etc.) that holds the actual certificate signed by the CA.

9. `KubeDB` Provisioner  operator watches for the Certificate secrets `tls-secrets`.

10. When it finds all the tls-secret, it creates the related `StatefulSets` so that Pgpool database can be configured with TLS/SSL.

In the next doc, we are going to show a step by step guide on how to configure a `Pgpool` database with TLS/SSL.