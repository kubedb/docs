---
title: Kubectl-Dba Completion
menu:
  docs_{{ .version }}:
    identifier: kubectl-dba-completion
    name: Kubectl-Dba Completion
    parent: reference-cli
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubectl-dba completion

Generate completion script

### Synopsis

To load completions:

Bash:

$ source <(kubectl-dba completion bash)

# To load completions for each session, execute once:
Linux:
  $ kubectl-dba completion bash > /etc/bash_completion.d/kubectl-dba
MacOS:
  $ kubectl-dba completion bash > /usr/local/etc/bash_completion.d/kubectl-dba

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ kubectl-dba completion zsh > "${fpath[1]}/_kubectl-dba"

# You will need to start a new shell for this setup to take effect.

Fish:

$ kubectl-dba completion fish | source

# To load completions for each session, execute once:
$ kubectl-dba completion fish > ~/.config/fish/completions/kubectl-dba.fish


```
kubectl-dba completion [bash|zsh|fish|powershell]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
      --as string                        Username to impersonate for the operation
      --as-group stringArray             Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string                 Default HTTP cache directory (default "/home/runner/.kube/http-cache")
      --certificate-authority string     Path to a cert file for the certificate authority
      --client-certificate string        Path to a client certificate file for TLS
      --client-key string                Path to a client key file for TLS
      --cluster string                   The name of the kubeconfig cluster to use
      --context string                   The name of the kubeconfig context to use
      --enable-analytics                 Send analytical events to Google Analytics (default true)
      --insecure-skip-tls-verify         If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                Path to the kubeconfig file to use for CLI requests.
      --log-backtrace-at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log-dir string                   If non-empty, write log files in this directory
      --log-flush-frequency duration     Maximum number of seconds between log flushes (default 5s)
      --logtostderr                      log to standard error instead of files
      --match-server-version             Require server version to match client version
  -n, --namespace string                 If present, the namespace scope for this CLI request
      --request-timeout string           The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                    The address and port of the Kubernetes API server
      --stderrthreshold severity         logs at or above this threshold go to stderr
      --tls-server-name string           Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                     Bearer token for authentication to the API server
      --user string                      The name of the kubeconfig user to use
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [kubectl-dba](/docs/reference/cli/kubectl-dba.md)	 - kubectl plugin for KubeDB

