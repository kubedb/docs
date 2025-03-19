---
title: Kubedb-Webhook-Server Run
menu:
  docs_{{ .version }}:
    identifier: kubedb-webhook-server-run
    name: Kubedb-Webhook-Server Run
    parent: reference-webhook-server
menu_name: docs_{{ .version }}
section_menu_id: reference
---
## kubedb-webhook-server run

Launch KubeDB Webhook Server

```
kubedb-webhook-server run [flags]
```

### Options

```
      --burst int                          The maximum burst for throttle (default 1000000)
      --cert-dir string                    The directory that contains the webhook and metrics server certificate.
      --enable-http2                       If set, HTTP/2 will be enabled for the metrics and webhook servers
      --health-probe-bind-address string   The address the probe endpoint binds to. (default ":8081")
  -h, --help                               help for run
      --label-key-blacklist strings        list of keys that are not propagated from a CRD object to its offshoots (default [app.kubernetes.io/name,app.kubernetes.io/version,app.kubernetes.io/instance,app.kubernetes.io/managed-by])
      --leader-elect                       Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.
      --metrics-bind-address string        The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.
      --metrics-secure                     If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead. (default true)
      --qps float                          The maximum QPS to the master from this client (default 1e+06)
```

### Options inherited from parent commands

```
      --default-seccomp-profile-type string   Default seccomp profile
```

### SEE ALSO

* [kubedb-webhook-server](/docs/reference/webhook-server/kubedb-webhook-server.md)	 - 

