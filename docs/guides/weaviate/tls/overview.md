---
title: Weaviate TLS Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-tls-overview
    name: Overview
    parent: weaviate-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate TLS Encryption

**Prerequisite:** To configure TLS in `Weaviate`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. Install `cert-manager` in your cluster following the steps [here](https://cert-manager.io/docs/installation/).

To issue a certificate, the following CRDs of `cert-manager` are used:

- `Issuer/ClusterIssuer`: Issuers and ClusterIssuers represent certificate authorities (CAs) that are able to generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a referenced issuer that is in a ready condition to attempt to honor the request. You can learn more details [here](https://cert-manager.io/docs/concepts/issuer/).

- `Certificate`: cert-manager has the concept of Certificates that define a desired x509 certificate which will be renewed and kept up to date. You can learn more details [here](https://cert-manager.io/docs/concepts/certificate/).

**Weaviate CRD Specification:**

KubeDB uses the following CRD fields to enable TLS encryption in `Weaviate`.

- `spec:`
  - `tls:`
    - `issuerRef`
    - `certificates`
    - `clientAuth`

`KubeDB` uses the `Issuer` or `ClusterIssuer` referenced in the `tls.issuerRef` field to generate the certificate secrets. These certificate secrets (the `server` and `client` certificates, each holding `ca.crt`, `tls.crt`, and `tls.key`) are used to configure the `Weaviate` server. When TLS is enabled, the REST service is served over HTTPS on port `8443` (instead of plain HTTP on `8080`).

Here,

- `issuerRef` is a reference to the `Issuer` or `ClusterIssuer` CR of [cert-manager](https://cert-manager.io/docs/concepts/issuer/) that will be used by `KubeDB` to generate the necessary certificates.
  - `apiGroup` is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
  - `kind` is the type of resource that is being referenced. `KubeDB` supports both `Issuer` and `ClusterIssuer` as values for this field.
  - `name` is the name of the resource (`Issuer` or `ClusterIssuer`) being referenced.

- `certificates` (optional) is a list of additional certificate specifications used to configure the Weaviate server. You can specify custom `dnsNames`, `ipAddresses`, and `subject` for the certificates.

- `clientAuth` (optional) controls whether the REST HTTPS listener requires clients to present a valid client certificate (mutual TLS). When unset or `true`, client certificate authentication is enforced; set it to `false` to allow clients to connect without a client certificate.

## How TLS Configures in Weaviate

Deploying Weaviate with TLS configuration consists of the following steps:

1. At first, a user creates an `Issuer/ClusterIssuer` CR.

2. Then the user creates a `Weaviate` CR which refers to the `Issuer/ClusterIssuer` CR created in the previous step through `spec.tls.issuerRef`.

3. `KubeDB-Provisioner` operator watches for the `Weaviate` CR.

4. `KubeDB` Ops-manager operator creates `Certificate` resources by using the `tls.issuerRef` and `tls.certificates` specification from the `Weaviate` CR.

5. `cert-manager` watches for the certificates and issues the certificate secrets (`server` and `client`) that hold the actual certificates signed by the CA.

6. `KubeDB-Provisioner` operator watches for the certificate secrets and configures the `Weaviate` PetSet so that the database serves traffic over TLS.

In the next doc, we are going to show a step-by-step guide on how to configure a `Weaviate` database with TLS.
