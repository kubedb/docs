---
title: Oracle TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-tls-configure
    name: Configure TLS/SSL
    parent: guides-oracle-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS/SSL in Oracle

`KubeDB` supports providing TLS/SSL encryption for Oracle using **TCPS** (Oracle Net over TLS). This tutorial will show you how to deploy a TLS secured Oracle database and connect to it over the encrypted TCPS listener.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- Install [`cert-manager`](https://cert-manager.io/docs/installation/) in your cluster. KubeDB uses cert-manager to issue the Oracle certificates.

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
```

- To keep things isolated, this tutorial uses a separate namespace called `demo`.

```bash
kubectl create ns demo
```
namespace/demo created

> Note: YAML files used in this tutorial are stored in [docs/examples/oracle/tls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/tls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`. Create an Oracle Container Registry token, if you haven't created one already, by following the instructions in the guide below: [here](/docs/guides/oracle/quickstart#create-oracle-image-pull-secret-important)

## Create an Issuer

KubeDB needs a cert-manager `Issuer` (or `ClusterIssuer`) to sign the Oracle certificates. First, generate a CA and create a TLS secret from it,

```bash
openssl req -x509 -nodes -days 3650 \
  -newkey rsa:2048 \
  -keyout ca.key \
  -out ca.crt \
  -subj "/CN=oracle-ca"
```

```bash
kubectl create secret tls oracle-ca \
  --cert=ca.crt \
  --key=ca.key \
  -n demo
```
secret/oracle-ca created

Now create an `Issuer` that uses this CA secret,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: oracle-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: oracle-ca
```

```bash
kubectl apply -f issuer.yaml
```
issuer.cert-manager.io/oracle-ca-issuer created

```bash
kubectl get issuer -n demo
```
NAME               READY   AGE
oracle-ca-issuer   True    10s

> If the Issuer is not present (or not `Ready`), the Oracle database will stay in the `Provisioning` phase.

## Deploy TLS/SSL secured Oracle

Below is the YAML of a TLS enabled Oracle standalone database. TLS is configured through `spec.tcpsConfig`,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: standalone-tls
  namespace: demo
spec:
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  tcpsConfig:
    tls:
      issuerRef:
        apiGroup: cert-manager.io
        name: oracle-ca-issuer
        kind: Issuer
    tcpsListener:
      port: 2484
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Here,

- `spec.tcpsConfig.tls.issuerRef` references the `oracle-ca-issuer` Issuer used to sign the certificates.
- `spec.tcpsConfig.tcpsListener.port: 2484` is the port for the encrypted TCPS listener. The plaintext listener remains on port `1521`.

For a **DataGuard** cluster, set `mode: DataGuard` and `replicas: 3` (see `dataguard-tls.yaml`); the TLS configuration is identical.

Let's create the `Oracle` CR,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/tls/standalone-tls.yaml
```
oracle.kubedb.com/standalone-tls created

Wait until the database is `Ready` and the pod prints the `DATABASE IS READY TO USE!!!` banner.

## Verify TLS/SSL

Once the database is ready, KubeDB has created the cert-manager `Certificate`s and the Oracle auto-login wallet. Let's check the generated certificate and wallet secrets,

```bash
kubectl get secret -n demo | grep standalone-tls
```
standalone-tls-auth                    kubernetes.io/basic-auth   2      16m
standalone-tls-client-cert             kubernetes.io/tls          4      16m
standalone-tls-metrics-exporter-cert   kubernetes.io/tls          4      16m
standalone-tls-server-cert             kubernetes.io/tls          3      16m
standalone-tls-tls-wallet              Opaque                     5      16m

Here, `standalone-tls-server-cert`, `standalone-tls-client-cert`, and `standalone-tls-metrics-exporter-cert` are the cert-manager issued certificates, and `standalone-tls-tls-wallet` is the Oracle auto-login wallet (built from those certificates) that clients use to connect over TCPS. The underlying cert-manager `Certificate` objects are all `Ready`,

```bash
kubectl get certificate -n demo | grep standalone-tls
```
standalone-tls-client-cert             True    standalone-tls-client-cert             16m
standalone-tls-metrics-exporter-cert   True    standalone-tls-metrics-exporter-cert   16m
standalone-tls-server-cert             True    standalone-tls-server-cert             16m

The TCPS listener is exposed on port `2484` of the database services (the plaintext listener remains on `1521`),

```bash
kubectl get svc -n demo -l app.kubernetes.io/instance=standalone-tls
```
NAME                  TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
standalone-tls        ClusterIP   10.43.67.172   <none>        1521/TCP,2484/TCP   16m
standalone-tls-pods   ClusterIP   None           <none>        1521/TCP,2484/TCP   16m

## Connect over TCPS

To verify the encrypted connection, create a client pod that mounts the wallet secret (`standalone-tls-tls-wallet`),

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: oracle-client-pod
  namespace: demo
spec:
  imagePullSecrets:
    - name: orclcred
  containers:
    - name: client
      image: container-registry.oracle.com/database/enterprise:21.3.0.0
      command: ["/bin/sh"]
      args: ["-c", "sleep infinity"]
      volumeMounts:
        - name: wallet
          mountPath: /opt/oracle/wallet
          readOnly: true
  securityContext:
    runAsUser: 0
    fsGroup: 0
  volumes:
    - name: wallet
      secret:
        secretName: standalone-tls-tls-wallet
```

Exec into the pod,

```bash
kubectl exec -it -n demo oracle-client-pod -- bash
```

Configure `sqlnet.ora` to point at the mounted wallet,

```bash
cat > /opt/oracle/product/21c/dbhome_1/network/admin/samples/sqlnet.ora <<'EOF'
WALLET_LOCATION =
  (SOURCE =
    (METHOD = FILE)
    (METHOD_DATA =
      (DIRECTORY = /opt/oracle/wallet)
    )
  )

SSL_CLIENT_AUTHENTICATION = FALSE
EOF
```

> Note: `SSL_CLIENT_AUTHENTICATION = FALSE` is the standard Oracle setting on the **client** side; the client still presents its certificate from the mounted wallet, so the server-side mutual TLS handshake is preserved. This flag only disables the client acting as the authenticating party during negotiation and does not turn off mutual TLS enforced by the listener.

Configure `tnsnames.ora` to connect over TCPS on port `2484`,

```bash
cat > /opt/oracle/product/21c/dbhome_1/network/admin/samples/tnsnames.ora <<'EOF'
ORCL =
  (DESCRIPTION =
    (ADDRESS = (PROTOCOL = TCPS)(HOST = standalone-tls.demo.svc.cluster.local)(PORT = 2484))
    (CONNECT_DATA =
      (SERVER = DEDICATED)
      (SERVICE_NAME = ORCL)
    )
  )
EOF
```

Copy the configs into place, set `TNS_ADMIN`, and verify,

```bash
mkdir -p /opt/oracle/network/admin
cp /opt/oracle/product/21c/dbhome_1/network/admin/samples/tnsnames.ora /opt/oracle/network/admin/
cp /opt/oracle/product/21c/dbhome_1/network/admin/samples/sqlnet.ora /opt/oracle/network/admin/
export TNS_ADMIN=/opt/oracle/network/admin

tnsping ORCL
sqlplus sys/'<your-sys-password>'@ORCL as sysdba
```

`tnsping ORCL` resolves the `TCPS` address and reaches the listener on port `2484`,

```bash
tnsping ORCL
```
TNS Ping Utility for Linux: Version 21.0.0.0.0 - Production
...
Attempting to contact (DESCRIPTION = (ADDRESS = (PROTOCOL = TCPS)(HOST = standalone-tls.demo.svc.cluster.local)(PORT = 2484)) (CONNECT_DATA = (SERVER = DEDICATED) (SERVICE_NAME = ORCL)))
OK (10 msec)

Finally, connect with `sqlplus` and confirm the session protocol is `tcps`,

```bash
sqlplus -s sys/'<your-sys-password>'@ORCL as sysdba
```
SQL> SELECT SYS_CONTEXT('USERENV','NETWORK_PROTOCOL') AS PROTOCOL FROM DUAL;

PROTOCOL
--------------------------------------------------------------------------------
tcps

The session protocol reported as `tcps` confirms that the connection is TLS/SSL encrypted.

> You can retrieve the `sys` password with:
> `kubectl get secret -n demo standalone-tls-auth -o jsonpath='{.data.password}' | base64 -d`

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete pod -n demo oracle-client-pod
kubectl patch -n demo oracle/standalone-tls -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete oracle -n demo standalone-tls
kubectl delete issuer -n demo oracle-ca-issuer
kubectl delete secret -n demo oracle-ca
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Read the [TLS/SSL overview](/docs/guides/oracle/tls/overview/index.md) for how KubeDB configures TLS.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

> ## ⚠️ Legal Notice
>
> Oracle® and Oracle Database® are registered trademarks of Oracle Corporation.
> KubeDB is not affiliated with, endorsed by, or sponsored by Oracle Corporation.
>
> KubeDB provides only orchestration and management tooling for Kubernetes.
> It does not distribute, bundle, ship, or include any Oracle Database software or binaries.
>
> Users must provide their own Oracle container images and hold valid Oracle licenses.
> Users are solely responsible for compliance with Oracle’s licensing terms, including all rules regarding containers, Docker, and Kubernetes environments.
>
> KubeDB makes no representations or warranties regarding Oracle licensing compliance.
