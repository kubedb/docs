---
title: Backup & Restore Elasticsearch Using Snapshot Plugins
menu:
  docs_{{ .version }}:
    identifier: guides-es-plugins-backup-overview
    name: Overview
    parent: guides-es-plugins-backup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Snapshot and Restore Using Repository Plugins

A snapshot is a backup taken from a running Elasticsearch cluster. You can take snapshots of an entire cluster, including all its data streams and indices. You can also take snapshots of only specific data streams or indices in the cluster.

Snapshots can be stored in remote repositories like Amazon S3, Microsoft Azure, Google Cloud Storage, and other platforms supported by a repository plugin.

Find more details at the official docs: [Snapshot and Restore](https://www.elastic.co/guide/en/elasticsearch/reference/7.14/snapshot-restore.html#snapshot-restore)

## KubeDB Managed Elasticsearch Docker Images

To enable the snapshot and restore feature, users need to install the respective repositroy plugin. For example, if user needs to installed [S3 Repository](https://www.elastic.co/guide/en/elasticsearch/plugins/7.14/repository-s3.html), the following needed to be run as root user:

```bash
sudo bin/elasticsearch-plugin install repository-s3
```

While running the Elasticsearch cluster in k8s, you don't always have the previliage to run as root user. Moreover, the plugin must be installed on every node in the cluster, and each node must be restarted after installation which bring more operational complexities. Here comes the KubeDB with Elasticsearch docker images (i.e. `Distribution=KubeDB`) with the pre-installed plugins; repository-s3, repository-azure, repository-hdfs, and repository-gcs.

```bash
$ kubectl get elasticsearchversions
NAME                   VERSION   DISTRIBUTION   DB_IMAGE                                          DEPRECATED   AGE
kubedb-xpack-7.12.0    7.12.0    KubeDB         kubedb/elasticsearch:7.12.0-xpack-v2021.08.23                  4h44m
kubedb-xpack-7.13.2    7.13.2    KubeDB         kubedb/elasticsearch:7.13.2-xpack-v2021.08.23                  4h44m
kubedb-xpack-7.14.0    7.14.0    KubeDB         kubedb/elasticsearch:7.14.0-xpack-v2021.08.23                  4h44m
kubedb-xpack-7.9.1     7.9.1     KubeDB         kubedb/elasticsearch:7.9.1-xpack-v2021.08.23                   4h44m
```

In case, you want to build your own custom Elasticsearch image with your own custom set of Elasticsearch plugins, visit the [elasticsearch-docker](https://github.com/kubedb/elasticsearch-docker/tree/release-7.14-xpack) github repository.

## What's Next?

- Snapshot and restore Elasticsearch cluster data using [S3 Repository Plugin](/docs/guides/elasticsearch/plugins-backup/s3-repository/index.md).