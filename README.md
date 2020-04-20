[![Go Report Card](https://goreportcard.com/badge/kubedb.dev/operator)](https://goreportcard.com/report/kubedb.dev/operator)
[![Build Status](https://github.com/kubedb/operator/workflows/CI/badge.svg)](https://github.com/kubedb/operator/actions?workflow=CI)
[![codecov](https://codecov.io/gh/kubedb/operator/branch/master/graph/badge.svg)](https://codecov.io/gh/kubedb/operator)
[![Docker Pulls](https://img.shields.io/docker/pulls/kubedb/operator.svg)](https://hub.docker.com/r/kubedb/operator/)
[![Slack](http://slack.kubernetes.io/badge.svg)](http://slack.kubernetes.io/#kubedb)
[![mailing list](https://img.shields.io/badge/mailing_list-join-blue.svg)](https://groups.google.com/forum/#!forum/kubedb)
[![Twitter](https://img.shields.io/twitter/follow/kubedb.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=kubedb)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fkubedb%2Foperator.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fkubedb%2Foperator?ref=badge_shield)

# KubeDB by AppsCode

> Making running production-grade databases easy on Kubernetes

Kubernetes has emerged as the de-facto way to deploy modern containerized apps on cloud or on-premises. “Despite all that growth on the application layer, the data layer hasn’t gotten as much traction with containerization” - [Google](https://cloud.google.com/blog/products/databases/to-run-or-not-to-run-a-database-on-kubernetes-what-to-consider). That’s not surprising, since handling things like state (the database), availability to other layers of the application, and redundancy for a database makes it challenging to run a database in a distributed environment like Kubernetes.

However, many developers want to treat data infrastructure the same as application stacks. Operators want to use the same tools for databases and applications and get the same benefits as the application layer in the data layer: rapid spin-up and repeatability across environments. This is where KubeDB comes as a solution.

KubeDB by AppsCode is a production-grade cloud-native database management solution for Kubernetes. KubeDB simplifies and automates routine database tasks such as provisioning, patching, backup, recovery, failure detection, and repair for various popular databases on private and public clouds. It frees you to focus on your applications so you can give them the fast performance, high availability, security and compatibility they need.

KubeDB provides you with many familiar database engines to choose from, including **PostgreSQL**, **MySQL**, **MongoDB**, **Elasticsearch**, **Redis**, **Memcached**, and **Percona XtraDB**. KubeDB’s native integration with Kubernetes makes a unique solution compared to competitive solutions from cloud providers and database vendors.

## Supported Versions

Please pick a version of KubeDB that matches your Kubernetes installation.

| KubeDB Version                                                                     | Docs                                                        | Kubernetes Version |
| ---------------------------------------------------------------------------------- | ----------------------------------------------------------- | ------------------ |
| [v0.13.0-rc.0](https://github.com/kubedb/cli/releases/tag/v0.13.0-rc.0) (uses CRD) | [User Guide](https://kubedb.com/docs/v0.13.0-rc.0/)         | 1.11.x +           |
| [0.12.0](https://github.com/kubedb/cli/releases/tag/0.12.0) (uses CRD)             | [User Guide](https://kubedb.com/docs/0.12.0/)               | 1.9.x +            |
| [0.11.0](https://github.com/kubedb/cli/releases/tag/0.11.0) (uses CRD)             | [User Guide](https://kubedb.com/docs/0.11.0/)               | 1.9.x +            |
| [0.10.0](https://github.com/kubedb/cli/releases/tag/0.10.0) (uses CRD)             | [User Guide](https://kubedb.com/docs/0.10.0/)               | 1.9.x +            |
| [0.9.0](https://github.com/kubedb/cli/releases/tag/0.9.0) (uses CRD)               | [User Guide](https://kubedb.com/docs/0.9.0/)                | 1.9.x +            |
| [0.8.0](https://github.com/kubedb/cli/releases/tag/0.8.0) (uses CRD)               | [User Guide](https://kubedb.com/docs/0.8.0/)                | 1.9.x +            |
| [0.6.0](https://github.com/kubedb/cli/releases/tag/0.6.0) (uses TPR)               | [User Guide](https://github.com/kubedb/cli/tree/0.6.0/docs) | 1.5.x - 1.7.x      |

## Installation

To install KubeDB, please follow the guide [here](https://kubedb.com/docs/latest/setup/install/).

## Using KubeDB

Want to learn how to use KubeDB? Please start [here](https://kubedb.com/docs/latest/guides/).

## KubeDB API Clients

You can use KubeDB api clients to programmatically access its CRD objects. Here are the supported clients:

- Go: [https://github.com/kubedb/apimachinery](https://github.com/kubedb/apimachinery/tree/master/client/clientset/versioned)
- Java: https://github.com/kubedb-client/java

## Contribution guidelines

Want to help improve KubeDB? Please start [here](https://kubedb.com/docs/latest/welcome/contributing/).

## Support

We use Slack for public discussions. To chit chat with us or the rest of the community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product annoucements, please join our [mailing list](https://groups.google.com/forum/#!forum/kubedb) or follow us on [Twitter](https://twitter.com/KubeDB). Our mailing list is also used to share design docs shared via Google docs.

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).

## License

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fkubedb%2Foperator.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fkubedb%2Foperator?ref=badge_large)
