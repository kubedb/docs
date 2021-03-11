---
title: Run Elasticsearch with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: es-overview-configuration
    name: Overview
    parent: es-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch with Custom Configuration Files

The KubeDB operator allows a user to deploy an Elasticsearch cluster with custom configuration files. The operator also allows the user to configure the security plugins such as X-Pack, SearchGurad, and OpenDistro. If the custom configuration files are not provided, the operator will start the cluster with default configurations.

## Overview

Elasticsearch has three configuration files:

- `elasticsearch.yml`: for configuring Elasticsearch
- `jvm.options`: for configuring Elasticsearch JVM settings
- `log4j2.properties`: for configuring Elasticsearch logging

In KubeDB managed Elasticsearch cluster, the configuration files are located at `/usr/share/elasticsearch/config` directory of Elasticsearch pods. To know more about configuring the Elasticsearch cluster see [here](https://www.elastic.co/guide/en/elasticsearch/reference/7.10/settings.html).

The `X-Pack` security plugin has the following configuration files:

- `roles.yml` - define roles and the associated permissions.
- `role_mapping.yml` - define which roles should be assigned to each user based on their username, groups, or other metadata.

The `SearchGuard` security plugin has the following configuration files:

- `sg_config.yml` - configure authenticators and authorization backends.
- `sg_roles.yml` - define roles and the associated permissions.
- `sg_roles_mapping.yml` - map backend roles, hosts, and users to roles.
- `sg_internal_users.yml` - stores users, and hashed passwords in the internal user database.
- `sg_action_groups.yml` - define named permission groups.
- `sg_tenants.yml` - defines tenants for configuring the Kibana access.
- `sg_blocks.yml` -  defines blocked users and IP addresses.

The `OpenDistro` security plugin has the following configuration files:

- `internal_users.yml` - contains any initial users that you want to add to the security plugin’s internal user database.
- `roles.yml` - contains any initial roles that you want to add to the security plugin.
- `roles_mapping.yml` - maps backend roles, hosts, and users to roles.
- `action_groups.yml` - contains any initial action groups that you want to add to the security plugin.
- `tenants.yml` - contains the tenant configurations.
- `nodes_dn.yml` - contains nodesDN mapping name and corresponding values.

## Custom Config Seceret

The custom configuration files are passed via a Kubernetes secret. The **file names are the keys** of the Secret with the **file-contents as the values**. The secret name needs to be mentioned in `spec.configSecret.name` of the [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md) object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-custom-config
  namespace: demo
spec:
  version: searchguard-7.9.3
  configSecret:
    name: es-custom-config
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: es-custom-config
  namespace: demo
stringData:
  elasticsearch.yml: |-
    logger.org.elasticsearch.discovery: DEBUG
```

**How the resultant configuration files are generated?**

- `YML`: The default configuration file pre-stored at config directories is overwritten by the operator-generated configuration file (if any). Then the resultant configuration file is overwritten by the user-provided custom configuration file (if any). The [yq](https://github.com/mikefarah/yq) tool is used to merge two YAML files.

  ```bash
  $ yq merge -i --overwrite file1.yml file2.yml
  ```

- `Non-YML`: The default configuration file is replaced by the operator-generated one (if any). Then the resultant configuration file is replaced by the user-provided custom configuration file (if any).

  ```bash
  $ cp -f file2 file1
  ```

**How to provide node-role specific configurations?**

If an Elasticsearch cluster is running in the topology mode (ie. `spec.topology` is set), a user may want to provide node-role specific configurations, say configurations that will only be merged to `master` nodes. To achieve this, users need to add the node role as a prefix to the file name.

- Format: `<node-role>-<file-name>.extension`
- Samples:
  - `data-elasticsearch.yml`: Only applied to `data` nodes.
  - `master-jvm.options`: Only applied to `master` nodes.
  - `ingest-log4j2.properties`: Only applied to `ingest` nodes.
  - `elasticsearch.yml`: applied to all nodes.

**How to provide additional files that are referenced from the configurations?**

All these files provided via `configSecret` is stored in each Elasticsearch node (i.e. pod) at `ES_CONFIG_DIR/custom_config/` ( i.e. `/usr/share/elasticsearch/config/custom_config/`) directory. So, user can use this path while configuring the Elasticsearch.

## Next Steps

- Learn how to use custom configuration in combined cluster from [here](/docs/guides/elasticsearch/configuration/combined-cluster/index.md).
- Learn how to use custom configuration in topology cluster from [here](/docs/guides/elasticsearch/configuration/topology-cluster/index.md).
