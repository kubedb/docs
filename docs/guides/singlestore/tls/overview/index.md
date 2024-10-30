---
title: SingleStore TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-tls-overview
    name: Overview
    parent: guides-sdb-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SingleStore TLS/SSL Encryption

**Prerequisite :** To configure TLS/SSL in `SingleStore`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

To issue a certificate, the following cr of `cert-manager` is used:

- `Issuer/ClusterIssuer`: Issuers and ClusterIssuers represent certificate authorities (CAs) that are able to generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer that is in a ready condition to attempt to honor the request. You can learn more details [here](https://cert-manager.io/docs/concepts/issuer/).

- `Certificate`: `cert-manager` has the concept of Certificates that define the desired x509 certificate which will be renewed and kept up to date. You can learn more details [here](https://cert-manager.io/docs/concepts/certificate/).

**SingleStore CRD Specification:**

KubeDB uses the following cr fields to enable SSL/TLS encryption in `SingleStore`.

- `spec:`
  - `tls:`
    - `issuerRef`
    - `certificates`

Read about the fields in details from [singlestore concept](/docs/guides/singlestore/concepts/singlestore.md#spectls),

`KubeDB` uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificate` to generate certificate secrets using `Issuer/ClusterIssuers` specification. These certificates secrets including `ca.crt`, `tls.crt` and `tls.key` etc. are used to configure `SingleStore` server, studio, exporter etc. respectively.

## How TLS/SSL configures in SingleStore

The following figure shows how `KubeDB` enterprise is used to configure TLS/SSL in SingleStore. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Stash Backup Flow" src="/docs/guides/singlestore/tls/overview/images/sdb-tls.svg">
<figcaption align="center">Fig: Deploy SingleStore with TLS/SSL</figcaption>
</figure>

Deploying SingleStore with TLS/SSL configuration process consists of the following steps:

1. At first, a user creates an `Issuer/ClusterIssuer` cr.

2. Then the user creates a `SingleStore` cr.

3. `KubeDB` Provisioner operator watches for the `SingleStore` cr.

4. When it finds one, it creates `Secret`, `Service`, etc. for the `SingleStore` database.

5. `KubeDB` Ops Manager watches for `SingleStore`(5c), `Issuer/ClusterIssuer`(5b), `Secret` and `Service`(5a).

6. When it finds all the resources(`SingleStore`, `Issuer/ClusterIssuer`, `Secret`, `Service`), it creates `Certificates` by using `tls.issuerRef` and `tls.certificates` field specification from `SingleStore` cr.

7. `cert-manager` watches for certificates.

8. When it finds one, it creates certificate secrets `tls-secrets`(server, client, secrets, etc.) that hold the actual self-signed certificate.

9. `KubeDB` Provisioner operator watches for the Certificate secrets `tls-secrets`.

10. When it finds all the tls-secret, it creates a `PetSet` so that SingleStore server is configured with TLS/SSL.

In the next doc, we are going to show a step by step guide on how to configure a `SingleStore` database with TLS/SSL.
