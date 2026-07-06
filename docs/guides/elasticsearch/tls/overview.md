---
title: Elasticsearch TLS/SSL Encryption Overview
menu:
  docs_{{ .version }}:
    identifier: es-tls-overview
    name: Overview
    parent: es-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch TLS/SSL Encryption

**Prerequisite :** To configure TLS/SSL in `Elasticsearch`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

To issue a certificate, the following crds of `cert-manager` are used:

- `Issuer/ClusterIssuer`: Issuers, and ClusterIssuers represent certificate authorities (CAs) that are able to generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer that is in a ready condition to attempt to honor the request. You can learn more details [here](https://cert-manager.io/docs/concepts/issuer/).

- `Certificate`: `cert-manager` has the concept of Certificates that define a desired x509 certificate which will be renewed and kept up to date. You can learn more details [here](https://cert-manager.io/docs/concepts/certificate/).

**Elasticsearch CRD Specification :**

KubeDB uses the following crd fields to enable SSL/TLS encryption in `Elasticsearch`.

- `spec:`
  - `enableSSL`
  - `tls:`
    - `issuerRef`
    - `certificates`

Read about the fields in details from [elasticsearch concept](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).

When `enableSSL` is set to `true`, the users must specify the `tls.issuerRef` field. `KubeDB` uses the `issuer` or `clusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificates` to generate certificate secrets using `Issuer/ClusterIssuer` specification. These certificate secrets including `ca.crt`, `tls.crt` and `tls.key` are used to configure both transport-layer (node-to-node) and HTTP-layer (client-to-node) TLS in `Elasticsearch`.

## How TLS/SSL configures in Elasticsearch

The following figure shows how `KubeDB` enterprise used to configure TLS/SSL in Elasticsearch. Open the image in a new tab to see the enlarged version.

<figure align="center">
<img alt="Deploy Elasticsearch with TLS/SSL" src="/docs/images/elasticsearch/es-tls.svg">
<figcaption align="center">Fig: Deploy Elasticsearch with TLS/SSL</figcaption>
</figure>

Deploying Elasticsearch with TLS/SSL configuration process consists of the following steps:

1. At first, a user creates an `Issuer/ClusterIssuer` cr.

2. Then the user creates an `Elasticsearch` CR with `enableSSL: true` which refers to the `Issuer/ClusterIssuer` CR that the user created in the previous step.

3. `KubeDB` Provisioner operator watches for the `Elasticsearch` cr.

4. When it finds one, it creates `Secret`, `Service`, etc. for the `Elasticsearch` cluster.

5. `KubeDB` Ops-manager operator watches for `Elasticsearch`(5c), `Issuer/ClusterIssuer`(5b), `Secret` and `Service`(5a).

6. When it finds all the resources(`Elasticsearch`, `Issuer/ClusterIssuer`, `Secret`, `Service`), it creates `Certificates` by using `tls.issuerRef` and `tls.certificates` field specification from `Elasticsearch` cr.

7. `cert-manager` watches for certificates.

8. When it finds one, it creates certificate secrets `tls-secrets` (transport and HTTP secrets etc.) that hold the actual certificate signed by the CA, containing `ca.crt`, `tls.crt` and `tls.key`.

9. `KubeDB` Provisioner operator watches for the Certificate secrets `tls-secrets`.

10. When it finds all the tls-secrets, it creates the related `PetSets` so that Elasticsearch can be configured with TLS/SSL for both transport and HTTP layers.

In the next docs, we are going to show a step-by-step guide on how to configure an `Elasticsearch` cluster with TLS/SSL.
