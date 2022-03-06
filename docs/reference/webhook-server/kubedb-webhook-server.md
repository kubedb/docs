---
title: Kubedb-Webhook-Server
menu:
  docs_{{ .version }}:
    identifier: kubedb-webhook-server
    name: Kubedb-Webhook-Server
    parent: reference-webhook-server
    weight: 0

menu_name: docs_{{ .version }}
section_menu_id: reference
url: /docs/{{ .version }}/reference/webhook-server/
aliases:
- /docs/{{ .version }}/reference/webhook-server/kubedb-webhook-server/
---
## kubedb-webhook-server



### Options

```
      --bypass-validating-webhook-xray   if true, bypasses validating webhook xray checks
  -h, --help                             help for kubedb-webhook-server
      --use-kubeapiserver-fqdn-for-aks   if true, uses kube-apiserver FQDN for AKS cluster to workaround https://github.com/Azure/AKS/issues/522 (default true)
```

### SEE ALSO

* [kubedb-webhook-server run](/docs/reference/webhook-server/kubedb-webhook-server_run.md)	 - Launch KubeDB Webhook Server
* [kubedb-webhook-server version](/docs/reference/webhook-server/kubedb-webhook-server_version.md)	 - Prints binary version number.

