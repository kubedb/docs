---
title: KubeDB-Operator
menu:
  docs_{{ .version }}:
    identifier: kubedb-operator
    name: KubeDB-Operator
    parent: reference-operator
    weight: 0

menu_name: docs_{{ .version }}
section_menu_id: reference
url: /docs/{{ .version }}/reference/operator/
aliases:
- /docs/{{ .version }}/reference/operator/kubedb-operator/
---
## kubedb-operator

KubeDB operator by AppsCode

### Synopsis

KubeDB operator by AppsCode

### Options

```
      --alsologtostderr                  log to standard error as well as files
      --bypass-validating-webhook-xray   if true, bypasses validating webhook xray checks
      --enable-analytics                 Send analytical events to Google Analytics (default true)
  -h, --help                             help for kubedb-operator
      --log-flush-frequency duration     Maximum number of seconds between log flushes (default 5s)
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files (default true)
      --stderrthreshold severity         logs at or above this threshold go to stderr
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO

* [kubedb-operator run](/docs/reference/operator/kubedb-operator_run.md)	 - Run kubedb operator in Kubernetes
* [kubedb-operator version](/docs/reference/operator/kubedb-operator_version.md)	 - Prints binary version number.

