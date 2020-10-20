---
title: Install KubeDB kubectl Plugin
description: Installation guide for KubeDB kubectl Plugin
menu:
  docs_{{ .version }}:
    identifier: install-kubedb-kubectl-plugin
    name: KubeDB kubectl Plugin
    parent: installation-guide
    weight: 30
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Install KubeDB kubectl Plugin

KubeDB provides a `kubectl` plugin to interact with KubeDB resources.

## Install using Krew

KubeDB `kubectl` plugin can be installed using Krew. [Krew](https://krew.sigs.k8s.io/) is the plugin manager for kubectl command-line tool. To install follow the steps below:

- Install `krew` following the steps [here](https://krew.sigs.k8s.io/docs/user-guide/setup/install/).

- If you have already installed `krew`, please upgrade `krew` to version v0.4.0 or later so that you can use [custom plugin indexes](https://krew.sigs.k8s.io/docs/user-guide/custom-indexes/).

```bash
kubectl krew upgrade
kubectl krew version
```

- Add [AppsCode's kubectl plugin index](https://github.com/appscode/krew-index). If you have already added the index, update the index.

```bash
kubectl krew index add appscode https://github.com/appscode/krew-index.git
kubectl krew index list
kubectl krew update
```

- Install KubeDB `kubectl` plugin following the commands below:

```bash
kubectl krew install appscode/dba
kubectl dba version
```

- If KubeDB `kubectl` plugin is already installed, run the following command to upgrade the plugin:

```bash
kubectl krew upgrade
kubectl dba version
```

## Install using pre-built binary

You can download the pre-build binaries from [kubedb/cli](https://github.com/kubedb/cli/releases) releases and put it into one of your installation directory denoted by `$PATH` variable.

Here is a simple Linux command to install the latest 64-bit Linux binary directly into your `/usr/local/bin` directory:

```bash
# Linux amd 64-bit
curl -o kubectl-dba.tar.gz -fsSL https://github.com/kubedb/cli/releases/download/{{< param "info.cli" >}}/kubectl-dba-linux-amd64.tar.gz \
  && tar zxvf kubectl-dba.tar.gz \
  && chmod +x kubectl-dba-linux-amd64 \
  && sudo mv kubectl-dba-linux-amd64 /usr/local/bin/kubectl-dba \
  && rm kubectl-dba.tar.gz LICENSE.md

# Mac OSX 64-bit
curl -o kubectl-dba.tar.gz -fsSL https://github.com/kubedb/cli/releases/download/{{< param "info.cli" >}}/kubectl-dba-darwin-amd64.tar.gz \
  && tar zxvf kubectl-dba.tar.gz \
  && chmod +x kubectl-dba-darwin-amd64 \
  && sudo mv kubectl-dba-darwin-amd64 /usr/local/bin/kubectl-dba \
  && rm kubectl-dba.tar.gz LICENSE.md
```

If you prefer to install kubectl KubeDB cli from source code, make sure that your go development environment has been setup properly. Then, just run:

```bash
go get github.com/kubedb/cli/...
```

>Please note that this will install KubeDB cli from master branch which might include breaking and/or undocumented changes.
