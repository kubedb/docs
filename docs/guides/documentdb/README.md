---
title: DocumentDB
menu:
  docs_{{ .version }}:
    identifier: documentdb-readme
    name: DocumentDB
    parent: documentdb-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/documentdb/
aliases:
  - /docs/{{ .version }}/guides/documentdb/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

KubeDB supports MongoDB-compatible documentdb database using the DocumentDB Custom Resource Definition (CRD). You can define the desired DocumentDB configuration in a YAML manifest, and KubeDB provisions and manages the necessary Kubernetes resources.

KubeDB simplifies deploying and managing DocumentDB on Kubernetes with a declarative API. It automates common operational tasks such as:

- Creating and provisioning standalone DocumentDB instances
- Managing persistent storage and data recovery
- Handling authentication and authorization
- Customizing pod templates and service configurations
- Monitoring database health

## Supported DocumentDB Features

| Features                         | Availability |
|----------------------------------|:------------:|
| Standalone DocumentDB deployment |   &#10003;   |
| Persistent volume                |   &#10003;   |
| Authentication secret            |   &#10003;   |
| Pod and service customization    |   &#10003;   |
| Health checker                   |   &#10003;   |
| Custom RBAC                      |   &#10003;   |
| Private registry                 |   &#10003;   |


## Example DocumentDB Manifest
Here's a simple example of a DocumentDB deployment:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: documentdb
spec:
  version: "pg17-0.109.0"
  storageType: Durable
  deletionPolicy: Delete
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

For a complete list of available configuration options, see the [DocumentDB CRD Documentation](/docs/guides/documentdb/concepts/documentdb.md).

## User Guide

Learn about DocumentDB features and how to use them:

### Quickstart
- [Quickstart](/docs/guides/documentdb/quickstart/quickstart.md) - Deploy your first DocumentDB instance with KubeDB operator

### Concepts
- [DocumentDB CRD](/docs/guides/documentdb/concepts/documentdb.md) - Understand the DocumentDB Custom Resource Definition and all available configuration options
- [DocumentDBVersion CRD](/docs/guides/documentdb/concepts/catalog.md) - Learn about specifying DocumentDB versions and docker images

### Setup & Configuration
- [Custom RBAC](/docs/guides/documentdb/custom-rbac/using-custom-rbac.md) - Setup custom ServiceAccount, Role, and RoleBinding for DocumentDB instances
- [Using Private Registry](/docs/guides/documentdb/private-registry/using-private-registry.md) - Deploy DocumentDB using images from a private docker registry



## Architecture

KubeDB uses a PetSet (similar to StatefulSet) to manage DocumentDB instances, ensuring stable network identities and persistent storage. Each DocumentDB instance:

- Runs as a single pod in a PetSet
- Uses a PersistentVolume for data storage

## What's Next

- **Want to learn more?** Check out the [DocumentDB Concepts](/docs/guides/documentdb/concepts/documentdb.md) page and explore the [Configuration](/docs/guides/documentdb/concepts/documentdb.md) options.
- **Want to deploy DocumentDB?** Follow the [Quickstart](/docs/guides/documentdb/quickstart/quickstart.md) guide.
- **Want to set up custom RBAC?** See the [Custom RBAC](/docs/guides/documentdb/custom-rbac/using-custom-rbac.md) guide.
- **Want to use a private registry?** Follow the [Private Registry](/docs/guides/documentdb/private-registry/using-private-registry.md) guide.

## Support

To speak with us, use the [Slack channel](http://slack.appscode.com) in the `#kubedb` room.

## Contributing

Want to help improve KubeDB? Please start [here](/docs/CONTRIBUTING.md).