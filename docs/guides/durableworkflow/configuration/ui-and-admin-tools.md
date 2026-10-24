---
title: UI and Admin Tools
menu:
  docs_{{ .version }}:
    identifier: dw-ui-admin-configuration
    name: UI and Admin Tools
    parent: dw-configuration
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# UI and Admin Tools

Both are optional and off by default. They are plain clients of the frontend, rolled like the servers.

## Web UI

`spec.ui.enabled: true` runs the web UI of the engine: the Temporal UI or cadence-web. It is a PetSet with `spec.ui.replicas` pods (1 by default) behind the Service `<name>-ui`, on port `8080` for Temporal and `8088` for Cadence. A cluster with the UI is highly available only if the UI also has at least two replicas.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-ui
  namespace: demo
spec:
  version: temporal-1.32.1
  topology:
    frontend:
      replicas: 2
    history:
      replicas: 2
    matching:
      replicas: 2
    worker:
      replicas: 2
  persistence:
    default:
      name: pg
      namespace: demo
  ui:
    enabled: true
    replicas: 2
  adminTools:
    enabled: true
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/configuration/temporal-ui-admin.yaml
$ kubectl port-forward -n demo svc/temporal-ui-ui 8080
```

Open `http://localhost:8080`.

- **TLS.** When TLS is enabled the UI gets the CA and the `client` certificate. The Temporal UI presents the certificate, so it works with `frontendClientAuth: Required`. cadence-web does one-way TLS only: it verifies the server and cannot present a client certificate. That is why a Cadence cluster does not accept `frontendClientAuth: Required`.
- **Cadence authentication.** `spec.ui.auth` is `Disabled` (the default) or `JWT`. With `JWT`, the UI reads a raw JWT from the `cadence-authorization` cookie and sends it to the frontend, which validates it. The UI has no login with an identity provider of its own: put an authenticating proxy in front of it that sets the cookie, or paste a token on the `/login` page. See [Authorization](/docs/guides/durableworkflow/authorization/authorization.md). `spec.ui.auth` is rejected for Temporal.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/configuration/cadence-ui-admin.yaml
```

## Admin tools

`spec.adminTools.enabled: true` runs one pod, `<name>-admin-tools-0`, that sleeps and is there to `exec` into. It has the CLI of the engine, already pointed at the frontend, with the TLS settings configured when TLS is on.

- **Temporal.** The `temporal` CLI. `TEMPORAL_ADDRESS` is set, and so are `TEMPORAL_TLS_*` and `TEMPORAL_CLI_TLS_*` (the CA, the `client` certificate and key, the server name) when TLS is on.

  ```bash
  $ kubectl exec -n demo temporal-ui-admin-tools-0 -- temporal operator namespace list
  ```

- **Cadence.** The `cadence` CLI, from the server image. Pass the address. With TLS, add the CA, which is mounted at `/etc/kubedb/certs/frontend/ca.crt`:

  ```bash
  $ kubectl exec -n demo cadence-ui-admin-tools-0 -- cadence --address cadence-ui:7833 --transport grpc admin cluster describe
  ```

The certificates of the other aliases are in `/etc/kubedb/certs/<alias>/` (`tls.crt`, `tls.key`, `ca.crt`).

## Next Steps

- Secure it with [TLS](/docs/guides/durableworkflow/tls/tls.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
