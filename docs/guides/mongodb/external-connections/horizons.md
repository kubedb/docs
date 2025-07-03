---
title: External Connections using MongoDB Horizons Guide
menu:
  docs_{{ .version }}:
    identifier: mg-horizons-guides
    name: MongoDB Horizons
    parent: mg-horizons
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB External Connections Outside Kubernetes using MongoDB Horizons

MongoDB Horizons is a feature in MongoDB that enables external connections to MongoDB replica sets deployed within Kubernetes. It allows applications or clients outside the Kubernetes cluster to connect to individual replica set members by mapping internal Kubernetes DNS names to externally accessible hostnames or IP addresses. This is useful for scenarios where external access is needed, such as hybrid deployments or connecting from outside the cluster.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates. As We must need tls enabled MongoDB for this tutorial.

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: YAML files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prerequisites

We need to have the following prerequisites to run this tutorial:

### Install Voyager Gateway

Install voyager gateway using the following command:
```bash
helm install ace oci://ghcr.io/appscode-charts/voyager-gateway \
  --version v2025.6.30 \
  -n ace-gw --create-namespace \
  --set gateway-converter.enabled=false \
  --wait --burst-limit=10000 --debug
```

### Create EnvoyProxy and GatewayClass
We need to setup `EnvoyProxy` and `GatewayClass` to use voyager gateway.

Create `EnvoyProxy` using the following command:
```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyProxy
metadata:
  name: ace
  namespace: ace-gw
spec:
  logging:
    level:
      default: warn
  mergeGateways: true
  provider:
    kubernetes:
      envoyDeployment:
        container:
          image: ghcr.io/voyagermesh/envoy:v1.34.1-ac
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
              - ALL
            privileged: false
            runAsNonRoot: true
            runAsUser: 65534
            seccompProfile:
              type: RuntimeDefault
        patch:
          value:
            spec:
              template:
                spec:
                  containers:
                  - name: shutdown-manager
                    securityContext:
                      allowPrivilegeEscalation: false
                      capabilities:
                        drop:
                        - ALL
                      privileged: false
                      runAsNonRoot: true
                      runAsUser: 65534
                      seccompProfile:
                        type: RuntimeDefault
      envoyService:
        externalTrafficPolicy: Cluster
        type: LoadBalancer
    type: Kubernetes
```


> If you want to use `NodePort` service. Update `.spec.provider.kubernetes.envoyService.type` to `NodePort` in the above YAML.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/horizons/envoyproxy.yaml
envoyproxy.gateway.envoyproxy.io/ace created
```

> Before creating `GatewayClass`, create a certificate secret named `ace-gw-cert` in ace namespace.

Create `GatewayClass` using the following command:
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  annotations:
    catalog.appscode.com/gateway-config: |-
      frontendTLSSecretRef:
        name: ace-gw-cert
        namespace: ace
      service:
        externalTrafficPolicy: Cluster
        nodeportRange: 30000-32767
        portRange: 10000-12767
        seedBackendPort: 8080
        type: LoadBalancer
      vaultServer:
        name: vault
        namespace: ace
    catalog.appscode.com/is-default-gatewayclass: "true"
  name: ace
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
  description: Default Service GatewayClass
  parametersRef:
    group: gateway.envoyproxy.io
    kind: EnvoyProxy
    name: ace
    namespace: ace-gw
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/horizons/gatewayclass.yaml
gatewayclass.gateway.networking.k8s.io/ace created
```

Check the `GatewayClass` status `True`.
```bash
$ kubectl get gatewayclass 
NAME   CONTROLLER                                      ACCEPTED   AGE
ace    gateway.envoyproxy.io/gatewayclass-controller   True       16s
```

### Install `FluxCD` in your cluster
Install `FluxCD` in your cluster using the following command:
```bash
helm upgrade -i flux2 \
  oci://ghcr.io/appscode-charts/flux2 \
  --version 2.15.0 \
  --namespace flux-system --create-namespace \
  --wait --debug --burst-limit=10000
```

###  Install Keda

Install `Keda` in your cluster using the following command:
```bash
$ kubectl create ns kubeops
namespace/kubeops created

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/horizons/helmrepo.yaml
helmrepository.source.toolkit.fluxcd.io/appscode-charts-oci created

$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/horizons/keda.yaml
helmrelease.helm.toolkit.fluxcd.io/keda created
helmrelease.helm.toolkit.fluxcd.io/keda-add-ons-http created
```

### Install `Catalog Manager`

Install `Catalog Manager` in your cluster using the following command:
```bash
helm install catalog-manager oci://ghcr.io/appscode-charts/catalog-manager \
  --version=v2025.6.30 \
  -n ace --create-namespace \
  --set helmrepo.name=appscode-charts-oci \
  --set helmrepo.namespace=kubeops \
  --wait --burst-limit=10000 --debug
```

## Overview

KubeDB uses following crd fields to enable MongoDB Horizons:

- `spec:`
    - `replicaSet:`
        - `name`
        - `horizons`
            - `dns`
            - `pods`

Read about the fields in details in [mongodb concept](/docs/guides/mongodb/concepts/mongodb.md),

## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable SSL/TLS in MongoDB. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mongo/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls mongo-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mongo-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: mongo-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/tls/issuer.yaml
issuer.cert-manager.io/mongo-ca-issuer created
```

## MongoDB Replicaset with Horizons

### Create DNS Records
Create dns `A`/`CNAME` records for mongodb replicaset pods, let's say, `MongoDB` has `3` replicas.

Example:
- `DNS`: `kubedb.cloud`, this will be used to connect to the MongoDB replica set using `mongodb+srv`.
- `A/CNAME Record` for each MongoDB replicas with exposed Envoy Gateway `LoadBalancer/NodePort` IP/Host: 
    - `mongo-0.kubedb.cloud`
    - `mongo-1.kubedb.cloud`
    - `mongo-2.kubedb.cloud`

Below is the YAML for MongoDB Replicaset Horizons. Here, [`spec.replicaSet.horizons`](/docs/guides/mongodb/concepts/mongodb.md#specreplicaset) specifies `horizons` for `replicaset`.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongodb-horizons
  namespace: demo
spec:
  clusterAuthMode: x509
  deletionPolicy: WipeOut
  replicaSet:
    horizons:
      dns: kubedb.cloud
      pods:
        - mongo-0.kubedb.cloud
        - mongo-1.kubedb.cloud
        - mongo-2.kubedb.cloud
    name: rs0
  replicas: 3
  sslMode: requireSSL
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  storageEngine: wiredTiger
  storageType: Durable
  tls:
    certificates:
      - alias: server
        dnsNames:
          - kubedb.cloud
          - mongo-0.kubedb.cloud
          - mongo-1.kubedb.cloud
          - mongo-2.kubedb.cloud
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mongo-ca-issuer
  version: 7.0.16
```

Here,
- `.spec.replicaSet.horizons.dns` specifies the DNS `SRV` record for the MongoDB replica set. It serves as the base domain for the SRV record used in `mongodb+srv` connection strings.
- `.spec.replicaSet.horizons.pods` specifies the DNS names for each pod in the replica set. These pod-specific DNS names are used to create `SRV` records that map to the individual MongoDB pods (e.g., `mongo-0`, `mongo-1`, `mongo-2`).
- `.spec.tls.certificates` specifies the certificate details for the MongoDB replica set. The dnsNames field under `.spec.tls.certificates` must include the replica set’s primary DNS (e.g., `kubedb.cloud`) and the DNS names of all pods listed in `.spec.replicaSet.horizons.pods`

>> **Note**: If you don't want to use `mongodb+srv` connection string, you can connect to the MongoDB replica set using the individual pod DNS names (e.g., `mongo-0.kubedb.cloud:10000`, `mongo-1.kubedb.cloud:10001`, etc.).

### Deploy MongoDB Replicaset Horizons

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/horizons/mongodb.yaml
mongodb.kubedb.com/mongodb-horizons created
```

Now, wait until `mongodb-horizons` has status `Ready`. i.e,

```bash
$ watch kubectl get mg -n demo
Every 2.0s: kubectl get mg -n demo
NAME               VERSION     STATUS    AGE
mongodb-horizons   7.0.16      Ready     4m10s
```

Now, create `MongoDBBinding` object to configure the whole process.
```yaml
apiVersion: catalog.appscode.com/v1alpha1
kind: MongoDBBinding
metadata:
  name: mongodb-bind
  namespace: demo
spec:
  sourceRef:
    name: mongodb-horizons
    namespace: demo
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/horizons/binding.yaml
mongodbbinding.catalog.appscode.com/mongodb-bind created
```

Now, check the status of `mongodbbinding` objects and ops requests.

```bash
$ kubectl get mongodbbinding,mongodbopsrequest -n demo
NAME                                               SRC_NS   SRC_NAME           STATUS   AGE
mongodbbinding.catalog.appscode.com/mongodb-bind   demo     mongodb-horizons   Current  3m28s

NAME                                                       TYPE       STATUS       AGE
mongodbopsrequest.ops.kubedb.com/mongodb-horizons-jddiql   Horizons   Successful   2m58s
```

### Connect to MongoDB as Replicaset

To connect to the MongoDB replica set, you can use the following command:

Collect the replicas from the `mongodb-horizons` object:
```bash
$ kubectl get mongodb -n demo mongodb-horizons -ojson | jq .spec.replicaSet.horizons.pods
[
  "mongo-0.kubedb.cloud:10000",
  "mongo-1.kubedb.cloud:10001",
  "mongo-2.kubedb.cloud:10002"
]

$ mongosh "mongodb://root:<password>@mongo-0.kubedb.cloud:10000,mongo-1.kubedb.cloud:10001,mongo-2.kubedb.cloud:10002/admin?authSource=admin&tls=true&tlsCAFile=<ca.crt-path>"
rs0 [primary] admin>
```

## Connect Using MongoDB `SRV`
To connect to the MongoDB replica set using `mongodb+srv`, you need to create `srv` records with the `A/CNAME` records you created earlier like,
```
srv-host=_mongodb._tcp.kubedb.cloud,mongo-0.kubedb.cloud,10000
srv-host=_mongodb._tcp.kubedb.cloud,mongo-1.kubedb.cloud,10001
srv-host=_mongodb._tcp.kubedb.cloud,mongo-2.kubedb.cloud,10002
```

Create a `TXT` record for the `SRV` records you created above.
```
txt-record=kubedb.cloud,"replicaSet=rs0&authSource=admin"
```
You can keep it empty.

Now, you can connect to the MongoDB replica set using the following command:

```bash
$ mongosh "mongodb+srv://root:<password>@kubedb.cloud/admin?tls=true&tlsCAFile=<ca.crt-path>"
rs0 [primary] admin>
```

> You can use `ca.crt` from default path.
```bash
sudo cp ca.crt /usr/local/share/ca-certificates/ca.crt
sudo update-ca-certificates
```

Now, you can connect without specifying `tlsCAFile` in the connection string.

```bash
$ mongosh "mongodb+srv://root:<password>@kubedb.cloud/admin"
rs0 [primary] admin>
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mongodbbinding -n demo mongodb-bind
kubectl delete mongodb -n demo mongodb-horizons

kubectl delete gatewayclass ace
kubectl delete -n ace-gw envoyproxy ace

helm uninstall -n ace catalog-manager
```

If you would like to uninstall the KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart MongoDB](/docs/guides/mongodb/quickstart/quickstart.md) with KubeDB Operator.
- [Backup and Restore](/docs/guides/mongodb/backup/stash/overview/index.md) MongoDB instances using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB instance with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB instance with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Use [kubedb cli](/docs/guides/mongodb/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

