---
title: YugabyteDB Webserver TLS
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-reconfigure-tls-webserver
    name: Webserver TLS
    parent: yugabytedb-reconfigure-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDB Webserver TLS

> **Not verified on a cluster.** This feature has not been run on a real cluster. What was checked: the admission webhook rules of the API (`apimachinery`, branch `yugabytedb`) have unit tests, and they pass. Everything below describes what the API types and those rules say, not an observed run, so expect to find differences when you try it.

`spec.tls.webserver` serves the web UIs and the metrics endpoint of the masters (port 7000) and the tservers (port 9000) over HTTPS, with the node certificate (`--webserver_certificate_file` and related flags). It defaults to `false`, and it needs `spec.tls.issuerRef`.

- At creation, set `webserver: true` next to `issuerRef` (see the [Connection Manager example](/docs/guides/yugabytedb/connection-manager/connection-manager.md), which does).
- On a running database it can only be changed by a `ReconfigureTLS` request. The step is `ApplyWebserverTLS`, which needs a rolling restart.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-webserver-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: yb-demo
  tls:
    # Serve the web UIs (7000, 9000) and the metrics endpoints over HTTPS.
    webserver: true
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/reconfigure-tls/ops-request-webserver.yaml
```

The webhook rules for the request: on a database that has no TLS, the request is an "add TLS" request and needs `spec.tls.issuerRef`; `spec.tls.webserver: true` cannot be combined with `remove` or `rotateCertificates`; and on a database with TLS, a `webserver` value equal to the current setting is rejected. Only one kind of change (remove, rotate, or a configuration change) is allowed per request.

## Effect on monitoring

The metrics endpoint moves from HTTP to HTTPS together with the web UIs. How the Prometheus scrape configuration of the monitoring setup follows that change was not verified, so check your `ServiceMonitor` after the switch.

## Not changed by this page

The warning of the [ReconfigureTLS overview](/docs/guides/yugabytedb/reconfigure-tls/overview.md) stands: adding TLS to, or removing TLS from, a running database is **not** zero downtime. Certificate rotation and issuer changes are online.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-webserver-tls
```
