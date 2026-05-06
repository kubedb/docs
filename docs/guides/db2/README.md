---
title: DB2
menu:
  docs_{{ .version }}:
    identifier: db2-readme
    name: DB2
    parent: db2-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/db2/
aliases:
  - /docs/{{ .version }}/guides/db2/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports IBM DB2 using the `DB2` Custom Resource Definition (CRD). You can declare the desired DB2 configuration, and KubeDB provisions and manages the required Kubernetes resources.

KubeDB simplifies deploying and managing DB2 on Kubernetes with a declarative API. It automates common operational tasks such as:

- Creating and provisioning standalone DB2 instances
- Managing persistent storage and data recovery
- Handling authentication and authorization
- Customizing pod templates and service configurations
- Monitoring database health

## Supported DB2 Features

| Features                       | Availability |
|--------------------------------|:------------:|
| Standalone DB2 deployment      |   &#10003;   |
| Persistent volume              |   &#10003;   |
| Authentication secret          |   &#10003;   |
| Pod and service customization  |   &#10003;   |
| Health checker                 |   &#10003;   |
| Custom RBAC                    |   &#10003;   |
| Private registry               |   &#10003;   |


## Example DB2 Manifest

Here's a simple example of a DB2 deployment:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DB2
metadata:
  name: db2
  namespace: demo
spec:
  version: 11.5.8.0
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: Delete
  healthChecker:
    periodSeconds: 10
    timeoutSeconds: 10
    failureThreshold: 3
```

For a complete list of available configuration options, see the [DB2 CRD Documentation](/docs/guides/db2/concepts/db2.md).

## User Guide

Learn about DB2 features and how to use them:

### Quickstart
- [Quickstart](/docs/guides/db2/quickstart/quickstart.md) - Deploy your first DB2 instance with KubeDB operator

### Concepts
- [DB2 CRD](/docs/guides/db2/concepts/db2.md) - Understand the DB2 Custom Resource Definition and all available configuration options
- [DB2Version CRD](/docs/guides/db2/concepts/catalog.md) - Learn about specifying DB2 versions and docker images

### Setup & Configuration
- [Custom RBAC](/docs/guides/db2/custom-rbac/using-custom-rbac.md) - Setup custom ServiceAccount, Role, and RoleBinding for DB2 instances
- [Using Private Registry](/docs/guides/db2/private-registry/using-private-registry.md) - Deploy DB2 using images from a private docker registry



## Architecture

KubeDB uses a PetSet (similar to StatefulSet) to manage DB2 instances, ensuring stable network identities and persistent storage. Each DB2 instance:

- Runs as a single pod in a PetSet
- Uses a PersistentVolume for data storage
- Comes with automated health checking via a coordinator container

## What's Next

- **Want to learn more?** Check out the [DB2 Concepts](/docs/guides/db2/concepts/db2.md) page and explore the [Configuration](/docs/guides/db2/concepts/db2.md#spec-podtemplate) options.
- **Want to deploy DB2?** Follow the [Quickstart](/docs/guides/db2/quickstart/quickstart.md) guide.
- **Want to set up custom RBAC?** See the [Custom RBAC](/docs/guides/db2/custom-rbac/using-custom-rbac.md) guide.
- **Want to use a private registry?** Follow the [Private Registry](/docs/guides/db2/private-registry/using-private-registry.md) guide.

## Support

To speak with us, use the [Slack channel](http://slack.appscode.com) in the `#kubedb` room.

## Contributing

Want to help improve KubeDB? Please start [here](/docs/CONTRIBUTING.md).
