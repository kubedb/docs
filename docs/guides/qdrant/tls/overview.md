---
title: Qdrant TLS Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-tls-overview
    name: Overview
    parent: qdrant-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant TLS Encryption

**Prerequisite:** To configure TLS/SSL in `Qdrant`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. Install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/).

To issue a certificate, the following CRDs of `cert-manager` are used:

- `Issuer/ClusterIssuer`: Issuers and ClusterIssuers represent certificate authorities (CAs) that are able to generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer that is in a ready condition to attempt to honor the request. You can learn more details [here](https://cert-manager.io/docs/concepts/issuer/).

- `Certificate`: cert-manager has the concept of Certificates that define a desired x509 certificate which will be renewed and kept up to date. You can learn more details [here](https://cert-manager.io/docs/concepts/certificate/).

**Qdrant CRD Specification:**

KubeDB uses the following CRD fields to enable TLS/SSL encryption in `Qdrant`.

- `spec:`
  - `tls:`
    - `issuerRef`
    - `certificates`
    - `client`
    - `p2p`

Read about the fields in detail from the [Qdrant Concepts](/docs/guides/qdrant/concepts/qdrant.md#spectls) page.

`KubeDB` uses the `Issuer` or `ClusterIssuer` referenced in the `tls.issuerRef` field, and the certificate specs provided in `tls.certificates` to generate certificate secrets. These certificate secrets including `ca.crt`, `server.crt`, `tls.key`, etc. are used to configure the `Qdrant` server.

Here,

- `issuerRef` is a reference to the `Issuer` or `ClusterIssuer` CR of [cert-manager](https://cert-manager.io/docs/concepts/issuer/) that will be used by `KubeDB` to generate necessary certificates.
  - `apiGroup` is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
  - `kind` is the type of resource that is being referenced. `KubeDB` supports both `Issuer` and `ClusterIssuer` as values for this field.
  - `name` is the name of the resource (`Issuer` or `ClusterIssuer`) being referenced.

- `certificates` (optional) is a list of additional certificates used to configure the Qdrant server. You can specify custom `dnsNames`, `ipAddresses`, and `subject` for server certificates.

- `client` (optional, default `false`) enables TLS for client-to-server communication. When set to `true`, the Qdrant server will accept TLS-encrypted connections from clients.

- `p2p` (optional, default `false`) enables TLS for peer-to-peer communication between Qdrant nodes. When set to `true`, inter-node communication within the Qdrant cluster will be encrypted using TLS.

## How TLS/SSL Configures in Qdrant

The following figure shows how `KubeDB` configures TLS/SSL in Qdrant. Open the image in a new tab to see the enlarged version.

<figure align="center">
<img alt="Deploy Qdrant with TLS/SSL" src="/docs/guides/qdrant/images/qdrant-tls.png">
<figcaption align="center">Fig: Deploy Qdrant with TLS/SSL</figcaption>
</figure>

Deploying Qdrant with TLS/SSL configuration process consists of the following steps:

1. At first, a user creates a `Issuer/ClusterIssuer` CR.

2. Then the user creates a `Qdrant` CR which refers to the `Issuer/ClusterIssuer` CR that the user created in the previous step.

3. `KubeDB-Provisioner` operator watches for the `Qdrant` CR.

4. When it finds one, it creates `Secret`, `Service`, etc. for the `Qdrant`.

5. `KubeDB` Ops-manager operator watches for `Qdrant`(5c), `Issuer/ClusterIssuer`(5b), `Secret` and `Service`(5a).

6. When it finds all the resources (`Qdrant`, `Issuer/ClusterIssuer`, `Secret`, `Service`), it creates `Certificates` by using `tls.issuerRef` and `tls.certificates` field specification from `Qdrant` CR.

7. `cert-manager` watches for certificates.

8. When it finds one, it creates certificate secrets `tls-secrets` (server, client secrets, etc.) that hold the actual certificates signed by the CA.

9. `KubeDB-Provisioner` operator watches for the certificate secrets `tls-secrets`.

10. When it finds all the tls-secrets, it creates the related `PetSet` so that the Qdrant database can be configured with TLS/SSL.

In the next doc, we are going to show a step-by-step guide on how to configure a `Qdrant` database with TLS/SSL.
