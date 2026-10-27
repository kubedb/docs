---
title: TLS
menu:
  docs_{{ .version }}:
    identifier: dw-tls-guide
    name: TLS
    parent: dw-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run DurableWorkflow with TLS

KubeDB issues the certificates of a `DurableWorkflow` with [cert-manager](https://cert-manager.io/docs/). TLS is on when `spec.tls.issuerRef` is set (`issuerRef` is required, and the webhook rejects a `tls` block without it).

## Before You Begin

- Complete [Prepare Dependencies](/docs/guides/durableworkflow/quickstart/prerequisites.md).
- Install [cert-manager](https://cert-manager.io/docs/installation/) in the cluster.
- You should be familiar with [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Certificates

Three certificates are issued, with these aliases in `spec.tls.certificates`:

| Alias       | Used for                                                                                          | Secret                      |
|-------------|---------------------------------------------------------------------------------------------------|-----------------------------|
| `internode` | The traffic between the roles. One certificate for every role, with server and client usage.         | `<name>-internode-cert`     |
| `frontend`  | The frontend, towards its clients.                                                                 | `<name>-frontend-cert`      |
| `client`    | The clients that the operator runs: the UI, the admin tools and the worker.                         | `<name>-client-cert`        |

A `certificates` entry can set `duration`, `renewBefore`, `subject` and `dnsNames`, and `secretName` to name the Secret. Each Secret has `tls.crt`, `tls.key` and `ca.crt`.

`spec.frontendClientAuth` decides if the frontend requires client certificates. It is `Required` for Temporal and `NotRequired` for Cadence by default.

## Create an Issuer

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=dw/O=kubedb"
$ kubectl create secret tls dw-ca -n demo --cert=ca.crt --key=ca.key
secret/dw-ca created
```

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: dw-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: dw-ca
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/tls/issuer.yaml
issuer.cert-manager.io/dw-ca-issuer created
```

## Temporal

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-tls
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
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: dw-ca-issuer
    certificates:
    - alias: frontend
      duration: 2160h
      renewBefore: 360h
  adminTools:
    enabled: true
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/tls/temporal-tls.yaml
durableworkflow.kubedb.com/temporal-tls created
```

What you get:

- **Mutual TLS between the roles**, with the `internode` certificate.
- **TLS on the frontend.** With the default `frontendClientAuth: Required`, a client must present a certificate signed by the CA. The admin tools pod already has its settings (`TEMPORAL_TLS_*`), so the CLI works in it. The certificate of an application is issued from the same issuer, and the server name to verify is `<name>.<namespace>.svc`.
- **Reload without restart.** The servers re-read the certificates from disk periodically (`global.tls.refreshInterval`, 1 minute), so a renewed certificate is used without restarting any pod. The [rotation](/docs/guides/durableworkflow/reconfigure_tls/overview.md) is therefore not disruptive.

## Cadence

{{< notice type="warning" message="Cadence TLS needs a server image that is built with the AppsCode public-client TLS patch. The upstream Cadence 1.4.2 image cannot be used, and the webhook rejects spec.tls for a version that does not set publicClientTLS: true. The cadence-1.4.2 version in the KubeDB catalog is the upstream image, so create your own DurableWorkflowVersion first." >}}

Why: the worker service of Cadence calls the frontend with its public client, which upstream builds without any TLS settings. The patched build reads `publicClient.tls`, and the catalog field `spec.publicClientTLS: true` tells the operator that the image has it.

1. Create a `DurableWorkflowVersion` with `publicClientTLS: true` that points to your patched image. Replace the image of the sample:

   ```yaml
   apiVersion: catalog.kubedb.com/v1alpha1
   kind: DurableWorkflowVersion
   metadata:
     name: cadence-1.4.2-tls
   spec:
     version: 1.4.2
     distribution: Cadence
     defaultHistoryShards: 16
     publicClientTLS: true
     server:
       image: registry.example.com/kubedb/cadence-server:v1.4.2-tls
     adminTools:
       image: registry.example.com/kubedb/cadence-server:v1.4.2-tls
     ui:
       image: docker.io/ubercadence/web:v4.0.17
     ports:
       frontend: {tchannel: 7933, grpc: 7833}
       history: {tchannel: 7934, grpc: 7834}
       matching: {tchannel: 7935, grpc: 7835}
       worker: {tchannel: 7939}
       metrics: 9090
       ui: 8088
     schema:
       tool: cadence-server
       dir: /etc/cadence/schema
       stores:
       - {type: Postgres, default: true, visibility: true}
       - {type: MySQL, default: true, visibility: true}
       - {type: Cassandra, default: true, visibility: true}
       - {type: Elasticsearch, visibility: true}
       - {type: OpenSearch, visibility: true}
     securityContext:
       runAsUser: 1000
     rollout:
       roleOrder: [history, matching, frontend, worker]
       frontendShutdownDrainDuration: 10s
       matchingShutdownDrainDuration: 10s
       historyShutdownDrainDuration: 30s
       preStopSleep: 10s
       historyMinInterval: 60s
       onlineSchemaUpdate: true
     updateConstraints:
       allowlist:
       - "> 1.4.2, < 1.5.0"
   ```

   ```bash
   $ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/tls/cadence-version-tls.yaml
   ```

2. Create the cluster. `frontendClientAuth` must be `NotRequired` (it is the default for Cadence), because the Cadence web UI and the readiness probe cannot present a client certificate. A Cadence cluster with TLS and `Required` is rejected.

   ```yaml
   apiVersion: kubedb.com/v1alpha2
   kind: DurableWorkflow
   metadata:
     name: cadence-tls
     namespace: demo
   spec:
     version: cadence-1.4.2-tls
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
     tls:
       issuerRef:
         apiGroup: cert-manager.io
         kind: Issuer
         name: dw-ca-issuer
     frontendClientAuth: NotRequired
     ui:
       enabled: true
     deletionPolicy: WipeOut
   ```

   ```bash
   $ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/tls/cadence-tls.yaml
   ```

What differs from Temporal:

- **Only the gRPC listeners have TLS.** History and matching require client certificates from the other roles (mutual TLS between the roles), and the frontend verifies the server for its clients. **The TChannel listeners stay plaintext.** Cadence carries its membership on TChannel, so it cannot be closed. The operator therefore creates a `NetworkPolicy` named `<name>-tchannel` that allows the TChannel ports (`7933` to `7939`) only from the pods of the same `DurableWorkflow`. Your cluster needs a CNI that enforces `NetworkPolicy`. Only the gRPC port `7833` is exposed by the frontend Service.
- **No reload.** Cadence reads the certificates at start-up. A renewed certificate is picked up by [rolling the pods one at a time](/docs/guides/durableworkflow/reconfigure_tls/overview.md), without downtime on a highly available cluster.
- **The UI verifies the server** with the CA and does not present a certificate.
- With the CLI, pass the CA: `cadence --address cadence-tls:7833 --transport grpc --tls_cert_path /etc/kubedb/certs/frontend/ca.crt ...`.

## Enabling or removing TLS later

You can add TLS to a running cluster, remove it, or change the issuer with a [Reconfigure TLS](/docs/guides/durableworkflow/reconfigure_tls/overview.md) ops request. **These three changes replace all pods at once and cause downtime**, because plaintext and TLS pods cannot talk to each other. The webhook warns you. Only certificate rotation and certificate changes under the same issuer are rolling and have no downtime.

## Cleaning up

```bash
kubectl delete dw -n demo temporal-tls cadence-tls
kubectl delete dwversion cadence-1.4.2-tls
kubectl delete issuer -n demo dw-ca-issuer
kubectl delete secret -n demo dw-ca
```

## Next Steps

- [Reconfigure TLS](/docs/guides/durableworkflow/reconfigure_tls/overview.md).
- Add [JWT authorization](/docs/guides/durableworkflow/authorization/authorization.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
