---
title: Reconfigure YugabyteDB TLS
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-reconfigure-tls-guide
    name: Reconfigure TLS
    parent: yugabytedb-reconfigure-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure TLS of YugabyteDB

> Note: YAML files used in this tutorial are stored in [docs/examples/yugabytedb/reconfigure-tls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/yugabytedb/reconfigure-tls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> **Warning:** adding TLS to, or removing TLS from, a running YugabyteDB is **not** zero downtime on this YugabyteDB version (2025.2.8). Expect connection failures and a stall of about 20 seconds. Rotating certificates and changing the issuer are online.

## Before You Begin

- Deploy `yb-demo` **without TLS** from the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md), and install cert-manager.
- Create the issuers (a CA `yb-ca` with `yb-ca-issuer`, and a second CA `yb-ca2` with `yb-ca2-issuer`):

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/reconfigure-tls/issuer.yaml
```

The numbers below were measured under a load generator (4 workers of insert plus select through the client Service). "Failed calls" counts only failures, "gap" is the longest period without any successful call.

## Add TLS

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-tls-add
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef: {name: yb-demo}
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: yb-ca-issuer
  rolling: {masterWait: 20s, tserverWait: 20s}
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/reconfigure-tls/ops-request-add.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-demo-tls-add created
```

Conditions: `IssueCertificatesSucceeded`, `CertificateSynced`, `RollingRoundOne`, `RollMasters`, `LeaderDrainTimedOut`, `RollTServers`, `TLSConverged`, `RequireSecureConnections`, `UpdatePetSets`, `RollTServersRound2`, `Successful`.

| Measure | Result |
|---------|--------|
| Duration | 492 s |
| Successful calls | 596,938 |
| Failed calls | 21 (10 timeouts, 8 terminated connections, 3 TLS errors), no other kind |
| Longest gap | 20.1 s |

`LeaderDrainTimedOut` is expected in this operation (see the [overview](/docs/guides/yugabytedb/reconfigure-tls/overview.md)). When the operation finishes, plaintext clients are refused.

## Rotate certificates

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-tls-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef: {name: yb-demo}
  tls:
    rotateCertificates: true
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/reconfigure-tls/ops-request-rotate.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-demo-tls-rotate created

$ kubectl get ybops -n demo yb-demo-tls-rotate
NAME                 TYPE             STATUS       AGE
yb-demo-tls-rotate   ReconfigureTLS   Successful   45s
```

Conditions: `IssueCertificatesSucceeded`, `CertificateSynced`, `ReloadCertificates`, `Successful`. The certificates are hot reloaded, no pod restarts. Measured under load: Successful in 101 s, 142,640 successful calls, **0 failed calls**, longest gap 16 ms. (A later run on an idle cluster took 45 s and no pod restarted.)

## Change the issuer

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-tls-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef: {name: yb-demo}
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: yb-ca2-issuer
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/reconfigure-tls/ops-request-issuer.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-demo-tls-issuer created
```

Conditions: `IssueProbeCertificate`, `CABundleExtended`, `CertificatesReissued`, `CABundleReduced`, `CleanupCertificates`, `Successful`. Measured under load on a TLS cluster: Successful in 313 s, 372,270 successful calls, **0 failed calls**, longest gap 0.7 s, no restart. The pods trust the CA bundle in the Secret `<name>-ca-bundle`, which is extended with the old and the new CA before the leaf certificates are re-issued, so peers always trust each other during the change.

## Remove TLS

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-tls-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef: {name: yb-demo}
  tls:
    remove: true
  rolling: {masterWait: 20s, tserverWait: 20s}
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/reconfigure-tls/ops-request-remove.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-demo-tls-remove created
```

Conditions: `AllowInsecureConnections`, `UpdatePetSets`, `LeaderDrainTimedOut`, `RollTServers`, `RollMasters`, `TLSConverged`, `CleanupCertificates`, `Successful`. Measured under load: Successful in 272 s, 288,286 successful calls, 21 failed calls (16 timeouts, 2 terminated connections, 2 TLS errors, 1 EOF), longest gap 20.1 s. The tservers are restarted first, then the masters.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-demo-tls-add yb-demo-tls-rotate yb-demo-tls-issuer yb-demo-tls-remove
```
