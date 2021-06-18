---
title: Kubedb-Operator
menu:
  docs_{{ .version }}:
    identifier: kubedb-operator
    name: Kubedb-Operator
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

### Options

```
      --bypass-validating-webhook-xray   if true, bypasses validating webhook xray checks
      --enable-analytics                 Send analytical events to Google Analytics (default true)
  -h, --help                             help for kubedb-operator
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
```

### SEE ALSO

* [kubedb-operator run](/docs/reference/operator/kubedb-operator_run.md)	 - Run kubedb operator in Kubernetes
* [kubedb-operator version](/docs/reference/operator/kubedb-operator_version.md)	 - Prints binary version number.

