---
title: Rotate Authentication Overview
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-rotate-auth-overview
    name: Overview
    parent: guides-proxysql-rotate-auth
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of ProxySQL

This guide will give an overview on how KubeDB Ops-manager operator rotates the authentication credentials of a `ProxySQL` server.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [ProxySQL](/docs/guides/proxysql/concepts/proxysql/index.md)
    - [ProxySQLOpsRequest](/docs/guides/proxysql/concepts/opsrequest/index.md)

## How Rotate ProxySQL Authentication Process Works

The auth secret of a KubeDB managed `ProxySQL` holds the credentials of the ProxySQL admin interface. KubeDB uses these credentials for the `admin-admin_credentials`, `admin-cluster_username` and `admin-cluster_password` admin variables, so the operator and the ProxySQL cluster nodes use them to manage and sync the servers.

The authentication rotation process for ProxySQL using KubeDB involves the following steps:

1. A user first creates a `ProxySQL` Custom Resource Object (CRO) that points to a MySQL backend.

2. The `KubeDB Provisioner operator` continuously watches for `ProxySQL` CROs.

3. When the operator detects a `ProxySQL` CR, it provisions the required `PetSet`, along with related resources such as the auth secret, the configuration secret, services and other dependencies.

4. To initiate authentication rotation, the user creates a `ProxySQLOpsRequest` CR of type `RotateAuth`.

5. The `KubeDB Ops-manager` operator watches for `ProxySQLOpsRequest` CRs.

6. Upon detecting a `ProxySQLOpsRequest`, the operator pauses the referenced `ProxySQL` object, ensuring that the Provisioner operator does not perform any operations during the authentication rotation process.

7. The `Ops-manager` operator then updates the auth secret with the new credentials. It either generates a new password or uses the user provided secret, and keeps the previous credentials under the `username.prev` and `password.prev` keys.

8. The operator updates the ProxySQL configuration secret and the `PetSet` with the new credentials, then restarts all `ProxySQL` Pods so they come up with the new admin credentials.

9. Once the authentication rotation is completed successfully, the operator updates `spec.authSecret` of the `ProxySQL` object and resumes it, allowing the Provisioner operator to continue its usual operations.

In the next section, we will walk you through a step-by-step guide to rotating ProxySQL authentication using the `ProxySQLOpsRequest` CRD.
