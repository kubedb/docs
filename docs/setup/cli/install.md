---
title: Install KubeDB CLI
menu:
  docs_{{ .version }}:
    identifier: install-cli
    name: Install
    parent: cli-setup
    weight: 10
product_name: KubeDB
menu_name: docs_{{ .version }}
section_menu_id: setup
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Installation Guide

In order to install KubeDB CLI as [kubectl-plugin](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/), download the `kubectl-dba` binary and move it anywhere on you PATH.

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="mac-tab" data-toggle="tab" href="#mac" role="tab" aria-controls="mac" aria-selected="true">macOS</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="linux-tab" data-toggle="tab" href="#linux" role="tab" aria-controls="linux" aria-selected="false">Linux</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="windows-tab" data-toggle="tab" href="#windows" role="tab" aria-controls="windows" aria-selected="false">Windows</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="mac" role="tabpanel" aria-labelledby="mac-tab">

## macOS

```console
$ wget -O kubectl-dba https://github.com/kubedb/cli/releases/download/{{< param "info.version" >}}/kubectl-dba-darwin-amd64 \
   && chmod +x kubectl-dba \
   && sudo mv kubectl-dba /usr/local/bin/
```

</div>
<div class="tab-pane fade" id="linux" role="tabpanel" aria-labelledby="linux-tab">

## Linux

```console
$ wget -O kubectl-dba https://github.com/kubedb/cli/releases/download/{{< param "info.version" >}}/kubectl-dba-linux-amd64 \
   && chmod +x kubectl-dba \
   && sudo mv kubectl-dba /usr/local/bin/
```

</div>
<div class="tab-pane fade" id="windows" role="tabpanel" aria-labelledby="windows-tab">

## Windows

1. Download the latest release v{{< param "info.version" >}} from this [link](https://github.com/kubedb/cli/releases/download/{{< param "info.version" >}}/kubectl-dba.exe).
2. Add the binary in to your PATH.

</div>
</div>

Now, check installation using:

```console
$ kubectl vault -h
KubeDB cli by AppsCode

Usage:
  vault [command]

Available Commands:
  approve     Approve request
  deny        Deny request
  help        Help about any command
  version     Prints binary version number.

Flags:
      --alsologtostderr                  log to standard error as well as files
      --analytics                        Send analytical events to Google Analytics (default true)
      --as string                        Username to impersonate for the operation
      --as-group stringArray             Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string                 Default HTTP cache directory (default "/home/tamal/.kube/http-cache")
      --certificate-authority string     Path to a cert file for the certificate authority
      --client-certificate string        Path to a client certificate file for TLS
      --client-key string                Path to a client key file for TLS
      --cluster string                   The name of the kubeconfig cluster to use
      --context string                   The name of the kubeconfig context to use
      --enable-status-subresource        If true, uses sub resource for crds. (default true)
  -h, --help                             help for vault
      --insecure-skip-tls-verify         If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                Path to the kubeconfig file to use for CLI requests.
      --log-backtrace-at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log-dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
      --match-server-version             Require server version to match client version
  -n, --namespace string                 If present, the namespace scope for this CLI request
      --request-timeout string           The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                    The address and port of the Kubernetes API server
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 0)
      --token string                     Bearer token for authentication to the API server
      --user string                      The name of the kubeconfig user to use
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging

Use "dba [command] --help" for more information about a command.
```
