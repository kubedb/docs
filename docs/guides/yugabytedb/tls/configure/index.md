---
title: YugabyteDB TLS/SSL (Transport Encryption)
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-tls-configure
    name: YugabyteDB TLS/SSL Configuration
    parent: yugabytedb-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS/SSL in YugabyteDB

This guide creates a universe with TLS from the start. To add TLS to a running universe, see [Reconfigure TLS](/docs/guides/yugabytedb/reconfigure-tls/reconfigure-tls.md) and read the downtime warning there.

> Note: YAML files used in this tutorial are stored in [docs/examples/yugabytedb/tls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/yugabytedb/tls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Follow the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md) up to the point where the `demo` namespace and the `spread-across-zones` PlacementPolicy exist.
- Install cert-manager.

## Create a CA Issuer

The manifest creates a self-signed issuer, a CA certificate `yb-ca`, and an `Issuer` named `yb-ca-issuer` that signs with it. (It also creates a second CA `yb-ca2` and `yb-ca2-issuer`, used by the issuer change in [Reconfigure TLS](/docs/guides/yugabytedb/reconfigure-tls/reconfigure-tls.md).)

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata: {name: selfsigned, namespace: demo}
spec: {selfSigned: {}}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata: {name: yb-ca, namespace: demo}
spec:
  isCA: true
  commonName: yb-ca
  secretName: yb-ca
  issuerRef: {name: selfsigned, kind: Issuer}
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata: {name: yb-ca-issuer, namespace: demo}
spec: {ca: {secretName: yb-ca}}
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata: {name: yb-ca2, namespace: demo}
spec:
  isCA: true
  commonName: yb-ca2
  secretName: yb-ca2
  issuerRef: {name: selfsigned, kind: Issuer}
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata: {name: yb-ca2-issuer, namespace: demo}
spec: {ca: {secretName: yb-ca2}}
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/tls/issuer.yaml
```

## Create a TLS enabled YugabyteDB

The manifest is the quickstart manifest plus `spec.tls`:

```yaml
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: yb-ca-issuer
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/tls/yugabytedb-tls.yaml
yugabytedb.kubedb.com/yb-demo created

$ kubectl get yb -n demo yb-demo
NAME      VERSION    STATUS   AGE
yb-demo   2025.2.8   Ready    80s
```

Check the Secrets created for the certificates:

```bash
$ kubectl get secret -n demo | grep yb-demo
yb-demo-auth                  kubernetes.io/basic-auth   2      80s
yb-demo-ca-bundle             Opaque                     1      79s
yb-demo-client-cert           kubernetes.io/tls          3      80s
yb-demo-config                Opaque                     2      80s
yb-demo-master-server-cert    kubernetes.io/tls          3      79s
yb-demo-tserver-server-cert   kubernetes.io/tls          3      80s
```

## Connect with TLS

From a tserver pod, the CA is mounted at `/opt/certs/yugabyte/ca.crt`. A plaintext connection is refused:

```bash
$ PW=$(kubectl get secret -n demo yb-demo-auth -o jsonpath='{.data.password}' | base64 -d)

$ kubectl exec -n demo yb-demo-tserver-0 -c yb-tserver -- env PGPASSWORD="$PW" PGSSLMODE=disable ysqlsh -h yb-demo.demo.svc -U yugabyte -c "select 1"
ysqlsh: error: connection to server at "yb-demo.demo.svc" (10.96.134.216), port 5433 failed: FATAL:  no pg_hba.conf entry for host "10.244.1.87", user "yugabyte", database "yugabyte", no encryption
```

A `verify-full` connection through the client Service works, because the Service name is in the SANs of the server certificates:

```bash
$ kubectl exec -n demo yb-demo-tserver-0 -c yb-tserver -- env PGPASSWORD="$PW" PGSSLMODE=verify-full PGSSLROOTCERT=/opt/certs/yugabyte/ca.crt ysqlsh -h yb-demo.demo.svc -U yugabyte -c "select ssl, version from pg_stat_ssl where pid=pg_backend_pid()"
 ssl | version 
-----+---------
 t   | TLSv1.3
(1 row)
```

For a client outside the pods, take the CA from the `yb-demo-ca-bundle` Secret (key `ca.crt`).

## Cleaning up

```bash
$ kubectl delete yb -n demo yb-demo
```

## Next Steps

- [Reconfigure TLS](/docs/guides/yugabytedb/reconfigure-tls/reconfigure-tls.md)
