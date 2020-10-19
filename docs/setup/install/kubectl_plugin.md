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

## Install KubeDB kubectl Plugin

KubeDB provides a `kubectl` plugin to interact with KubeDB resources. You can download the pre-build binaries from [kubedb/cli](https://github.com/kubedb/cli/releases) releases and put it into one of your installation directory denoted by `$PATH` variable.

Here is a simple Linux command to install the latest 64-bit Linux binary directly into your `/usr/local/bin` directory:

```bash
# Linux amd 64-bit
wget -O kubectl-dba https://github.com/kubedb/cli/releases/download/{{< param "info.cli" >}}/kubectl-dba-linux-amd64 \
  && chmod +x kubectl-dba \
  && sudo mv kubectl-dba /usr/local/bin/

# Mac OSX 64-bit
wget -O kubectl-dba https://github.com/kubedb/cli/releases/download/{{< param "info.cli" >}}/kubectl-dba-darwin-amd64 \
  && chmod +x kubectl-dba \
  && sudo mv kubectl-dba /usr/local/bin/
```

If you prefer to install kubectl KubeDB cli from source code, make sure that your go development environment has been setup properly. Then, just run:

```bash
go get github.com/kubedb/cli/...
```

>Please note that this will install kubectl kubedb cli from master branch which might include breaking and/or undocumented changes.
