---
title: GitOps PostgreSQL
menu:
  docs_{{ .version }}:
    identifier: pg-using-gitops
    name: GitOps PostgreSQL
    parent: pg-gitops-postgres
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# GitOps Postgres using KubeDB GitOps Operator

This guide will show you how to use `KubeDB` GitOps operator to create postgres database and manage updates using GitOps workflow.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).  Pass `--set kubedb-crd-manager.installGitOpsCRDs=true` in the kubedb installation process to enable `GitOps` operator. 

- You need to install GitOps tools like `ArgoCD` or `FluxCD` and configure with your Git Repository to monitor the Git repository and synchronize the state of the Kubernetes cluster with the desired state defined in Git.

  ```bash
  $ kubectl create ns monitoring
  namespace/monitoring created

  $ kubectl create ns demo
  namespace/demo created
  ```
> Note: YAML files used in this tutorial are stored in [docs/examples/postgres](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

We are going to use `ArgoCD` in this tutorial. You can install `ArgoCD` in your cluster by following the steps [here](https://argo-cd.readthedocs.io/en/stable/getting_started/). Also, you need to install `argocd` CLI in your local machine. You can install `argocd` CLI by following the steps [here](https://argo-cd.readthedocs.io/en/stable/cli_installation/).

## Creating Apps via CLI

### For Public Repository
```bash
argocd app create kubedb --repo <repo-url> --path kubedb --dest-server https://kubernetes.default.svc --dest-namespace <namespace>
```

### For Private Repository
#### Using HTTPS
```bash
argocd app create kubedb --repo <repo-url> --path kubedb --dest-server https://kubernetes.default.svc --dest-namespace <namespace> --username <username> --password <github-token>
```

#### Using SSH
```bash
argocd app create kubedb --repo <repo-url> --path kubedb --dest-server https://kubernetes.default.svc --dest-namespace <namespace> --ssh-private-key-path ~/.ssh/id_rsa
```



## Next Steps

- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
