[![Go Report Card](https://goreportcard.com/badge/kubedb.dev/operator)](https://goreportcard.com/report/kubedb.dev/operator)
![CI](https://github.com/kubedb/operator/workflows/CI/badge.svg)
[![Docker Pulls](https://img.shields.io/docker/pulls/kubedb/operator.svg)](https://hub.docker.com/r/kubedb/operator/)
[![Slack](http://slack.kubernetes.io/badge.svg)](http://slack.kubernetes.io/#kubedb)
[![Twitter](https://img.shields.io/twitter/follow/kubedb.svg?style=social&logo=twitter&label=Follow)](https://twitter.com/intent/follow?screen_name=KubeDB)

# KubeDB by AppsCode

> Run production-grade databases on Kubernetes

Kubernetes has emerged as the de-facto way to deploy modern containerized apps on cloud or on-premises. *"Despite all that growth on the application layer, the data layer hasn’t gotten as much traction with containerization"* - [Google](https://cloud.google.com/blog/products/databases/to-run-or-not-to-run-a-database-on-kubernetes-what-to-consider). That’s not surprising, since handling things like state (the database), availability to other layers of the application, and redundancy for a database makes it challenging to run a database in a distributed environment like Kubernetes.

However, many developers want to treat data infrastructure the same as application stacks. Operators want to use the same tools for databases and applications and get the same benefits as the application layer in the data layer: rapid spin-up and repeatability across environments. This is where KubeDB comes as a solution.

KubeDB by AppsCode is a production-grade cloud-native database management solution for Kubernetes. KubeDB simplifies and automates routine database tasks such as provisioning, patching, backup, recovery, failure detection, and repair for various popular databases on private and public clouds. It frees you to focus on your applications so you can give them the fast performance, high availability, security and compatibility they need.

KubeDB provides you with many familiar database engines to choose from, including **PostgreSQL**, **MySQL**, **MongoDB**, **Elasticsearch**, **Redis**, **Memcached**, and **Percona XtraDB**. KubeDB’s native integration with Kubernetes makes a unique solution compared to competitive solutions from cloud providers and database vendors.

## Features

|                                                                   | Community                            | Enterprise                                |
| ----------------------------------------------------------------- | ------------------------------------ | ----------------------------------------- |
|                                                                   | Open source KubeDB Free for everyone | Open Core KubeDB for production databases |
| PostgreSQL                                                        | √                                    | √                                         |
| MySQL                                                             | √                                    | √                                         |
| Elasticsearch                                                     | √                                    | √                                         |
| MongoDB                                                           | √                                    | √                                         |
| Redis                                                             | √                                    | √                                         |
| Memcached                                                         | √                                    | √                                         |
| MariaDB                                                           | √                                    | √                                         |
| Percona XtraDB                                                    | √                                    | √                                         |
| PgBouncer                                                         | x                                    | √                                         |
| ProxySQL                                                          | x                                    | √                                         |
| Database Clustering                                               | √                                    | √                                         |
| Cloud / On-prem / Air-gapped clusters                             | √                                    | √                                         |
| Multizone Cluster                                                 | √                                    | √                                         |
| Private Registry                                                  | √                                    | √                                         |
| CLI                                                               | √                                    | √                                         |
| Halt & resume database                                            | √                                    | √                                         |
| Custom Configuration                                              | √                                    | √                                         |
| Custom Extensions                                                 | √                                    | √                                         |
| Prometheus Metrics                                                | √                                    | √                                         |
| Protect against accidental deletion                               | x                                    | √                                         |
| Managed Backup/Recovery using [Stash](https://stash.run)          | x                                    | √                                         |
| Managed Patch Upgrades                                            | x                                    | √                                         |
| Managed Horizontal Scaling                                        | x                                    | √                                         |
| Managed Vertical Scaling                                          | x                                    | √                                         |
| Managed Volume Expansion                                          | x                                    | √                                         |
| Managed Reconfiguration                                           | x                                    | √                                         |
| Managed Restarts                                                  | x                                    | √                                         |
| Role Based Access Control (RBAC)                                  | √                                    | √                                         |
| Open Policy Agent (OPA)                                           | √                                    | √                                         |
| Pod Security Policy (PSP)                                         | √                                    | √                                         |
| Network Policy                                                    | √                                    | √                                         |
| User & Secret Management using [KubeVault](https://kubevault.com) | x                                    | √                                         |
| Managed TLS using [cert-manager](https://cert-manager.io)         | x                                    | √                                         |

## Installation

To install KubeDB, please follow the guide [here](https://kubedb.com/docs/latest/setup/).

## Using KubeDB

Want to learn how to use KubeDB? Please start [here](https://kubedb.com/docs/latest/guides/).

## Contribution guidelines

Want to help improve KubeDB? Please start [here](https://kubedb.com/docs/latest/welcome/contributing/).

## Support

To speak with us, please leave a message on [our website](https://appscode.com/contact/).

To join public discussions with the KubeDB community, join us in the [Kubernetes Slack team](https://kubernetes.slack.com/messages/C8149MREV/) channel `#kubedb`. To sign up, use our [Slack inviter](http://slack.kubernetes.io/).

To receive product announcements, follow us on [Twitter](https://twitter.com/KubeDB).

If you have found a bug with KubeDB or want to request for new features, please [file an issue](https://github.com/kubedb/project/issues/new).
