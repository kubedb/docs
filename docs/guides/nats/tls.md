---
title: Run NATS with TLS
menu:
  docs_{{ .version }}:
    identifier: nats-tls
    name: TLS/SSL
    parent: nats-guides
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run NATS with TLS

KubeDB uses [cert-manager](https://cert-manager.io) to issue certificates. Two certificate aliases are supported:

- `server`: presented by every server on the client, route, gateway, leafnode and monitor listeners.
- `client`: used by KubeDB (health checker, ops-manager) and by applications to connect.

> Note: The yaml files used in this tutorial are stored in [docs/examples/nats/tls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/nats/tls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not have one, create it with [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- Install KubeDB following the steps [here](/docs/setup/README.md), with `--set global.featureGates.NATS=true` so the NATS CRDs are installed.
- This tutorial uses the `demo` namespace:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

Install cert-manager [v1.0.0 or later](https://cert-manager.io/docs/installation/).

## Create an Issuer

```bash
$ openssl req -newkey rsa:2048 -keyout ca.key -nodes -x509 -days 365 -out ca.crt -subj "/CN=nats/O=kubedb"
$ kubectl create secret tls nats-ca -n demo --cert=ca.crt --key=ca.key
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/nats/tls/issuer.yaml
issuer.cert-manager.io/nats-ca-issuer created
```

## Create a NATS cluster with TLS

```yaml
apiVersion: kubedb.com/v1alpha2
kind: NATS
metadata:
  name: nats-tls
  namespace: demo
spec:
  version: "2.15.1"
  replicas: 3
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: nats-ca-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb
      - alias: client
        subject:
          organizations:
            - kubedb
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/nats/tls/nats-tls.yaml
```

When `spec.tls` is set, `issuerRef` is required. TLS is on by default for every listener. Switch one off with `spec.clientTLS`, `clusterTLS`, `gatewayTLS`, `leafNodeTLS` or `monitorTLS` set to `false`.

```bash
$ kubectl get secrets -n demo | grep nats-tls
nats-tls-server-cert   kubernetes.io/tls   3
nats-tls-client-cert   kubernetes.io/tls   3
```

Connect with the CA and client certificate:

```bash
$ kubectl get secret -n demo nats-tls-client-cert -o jsonpath='{.data.ca\.crt}' | base64 -d > ca.crt
$ nats --server tls://127.0.0.1:4222 --tlsca ca.crt --user admin --password "$PASS" server check connection
```

## Day-2 changes

Use `ReconfigureTLS` ([ops requests](/docs/guides/nats/ops-requests.md#reconfiguretls)):

- **Rotate certificates**: zero downtime. The reloader signals the server, which re-reads the certificate files; new handshakes use the new certificate and existing connections keep their session.
- **Change issuer or CA**: the new CA is added to the CA bundle next to the old one, certificates are rotated, then the old CA is dropped. No restart.
- **Add or remove TLS on the client port**: two zero-downtime rolls. The first accepts both TLS and plaintext clients (`allow_non_tls`). KubeDB then waits for plaintext clients to disappear, or for the request `timeout`. The second roll removes the plaintext allowance. Move your applications to TLS between the two.

## Cleanup

```bash
$ kubectl delete -n demo nats/nats-tls issuer/nats-ca-issuer secret/nats-ca
```
