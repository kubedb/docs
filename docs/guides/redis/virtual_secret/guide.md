---
title: Virtual Secret Guide Redis
menu:
  docs_{{ .version }}:
    identifier: rd-virtual-secret-guide
    name: Virtual Secret Guide
    parent: rd-virtual-secret
    weight: 10
menu_name: docs_{{ .version }}
---

> New to KubeDB? Please start [here](/docs/README.md).

# Virtual Secrets For Redis: Secure Kubernetes Secrets
KubeDB's Virtual Secrets feature enhances the security of your database credentials by allowing you to use external secret management systems instead of storing sensitive information directly 
in Kubernetes Secrets. This guide will walk you through the steps to set up and use Virtual Secrets with your Redis database in KubeDB.

## Virtual Secrets Design
`Virtual Secrets` extends Kubernetes by introducing a new `Secret` resource under the `virtual-secrets.dev` API group. From a user perspective, it behaves similarly to the native Kubernetes Secret
resource, providing familiar workflows for managing sensitive data. Unlike standard Kubernetes Secrets, Virtual Secrets does not store secret data in `etcd`. Instead, it securely stores the 
actual secret data in an `external secret manager`, ensuring enhanced security and compliance.

The Virtual Secret resource is structured into two distinct components:

- **Secret Data**– The sensitive information itself, stored externally to protect against unauthorized access.

- **Secret Metadata** – Non-sensitive information retained within the Kubernetes cluster to improve performance and support standard API operations.

This design ensures a seamless Kubernetes experience while providing enterprise-grade security for managing secrets.

## Prerequisites
Before you begin, ensure you have the following prerequisites in place:
- A running Kubernetes cluster with KubeDB installed. If you haven't set up KubeDB yet, follow the installation guide [here](/docs/setup/README.md).
- Familiarity with Kubernetes concepts, including Secrets and Custom Resource Definitions (CRDs).