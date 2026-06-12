---
title: Distributed Postgres Cluster Overview
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-overview
    name: Distributed Postgres Overview
    parent: guides-postgres-distributed
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Distributed Postgres Cluster Overview

## Introduction

KubeDB enables distributed Postgres deployments across multiple Kubernetes clusters, providing a scalable, highly available, and resilient database solution. By integrating **Open Cluster Management (OCM)** for multi-cluster orchestration and **KubeSlice** for seamless pod-to-pod network connectivity, KubeDB simplifies the deployment and management of Postgres instances across clusters. A **PlacementPolicy** ensures precise control over pod scheduling, allowing you to distribute Postgres pods across clusters for optimal resource utilization and fault tolerance.

This guide provides a step-by-step process to deploy a distributed Postgres cluster, including prerequisites, configuration, and verification steps. It assumes familiarity with Kubernetes and basic database concepts.

> **New to KubeDB?** Start with the [KubeDB documentation](https://kubedb.com/docs/v2025.7.31/welcome/) for an introduction.

## Understanding OCM Hub and Spoke Clusters

In an **Open Cluster Management (OCM)** setup, clusters are categorized as follows:

- **Hub Cluster**: The central control plane where policies, applications, and resources are defined and managed. It orchestrates the lifecycle of applications deployed across spoke clusters.
- **Spoke Cluster**: Managed clusters registered with the hub that run the actual workloads (e.g., Postgres pods).

When a spoke cluster (e.g., `demo-worker`) is joined to the hub using the `clusteradm join` command, OCM creates a namespace on the hub cluster that matches the spoke cluster's name (e.g., `demo-worker`). This namespace is used to manage resources specific to the spoke cluster from the hub.

## Prerequisites

Before deploying a distributed Postgres cluster, ensure the following requirements are met:

- **Kubernetes Clusters**: Multiple Kubernetes clusters (version 1.26 or higher) configured and accessible.
- **Node Requirements**: Each Kubernetes node should have at least 4 vCPUs and 16 GB of RAM.
- **Open Cluster Management (OCM)**: Install `clusteradm` as described in the [OCM Quick Start Guide](https://open-cluster-management.io/docs/getting-started/quick-start/).
- **kubectl**: Installed and configured to interact with all clusters.
- **Helm**: Installed for deploying the KubeDB Operator and KubeSlice components.
- **Persistent Storage**: A storage class (e.g., `local-path` or a cloud provider-specific option) configured for persistent volumes.

## Configuration Steps

Follow these steps to deploy a distributed Postgres cluster across multiple Kubernetes clusters.

### Step 1: Set Up Open Cluster Management (OCM)

#### 1. Configure KUBECONFIG

Ensure your `KUBECONFIG` is set up to switch between clusters. This guide uses two clusters: `demo-controller` (hub and spoke) and `demo-worker` (spoke).

```bash
$ kubectl config get-contexts
```

**Output:**

```bash
CURRENT   NAME              CLUSTER           AUTHINFO          NAMESPACE
*         demo-controller   demo-controller   demo-controller   
          demo-worker       demo-worker       demo-worker
```

#### 2. Initialize the OCM Hub

On the `demo-controller` cluster, initialize the OCM hub:

```bash
$ kubectl config use-context demo-controller
$ clusteradm init --wait --feature-gates=ManifestWorkReplicaSet=true
```

#### 3. Verify Hub Deployment

Check the pods in the `open-cluster-management-hub` namespace to ensure all components are running:

```bash
$ kubectl get pods -n open-cluster-management-hub
```

**Output:**

```
NAME                                                        READY   STATUS    RESTARTS   AGE
cluster-manager-addon-manager-controller-5f99f56896-qpzj8   1/1     Running   0          7m2s
cluster-manager-placement-controller-597d5ff644-wqjq2       1/1     Running   0          7m2s
cluster-manager-registration-controller-6d79d7dcc6-b8h9p    1/1     Running   0          7m2s
cluster-manager-registration-webhook-5d88cf97c7-2sq5m       1/1     Running   0          7m2s
cluster-manager-work-controller-7468bf4dc-5qn6q             1/1     Running   0          7m2s
cluster-manager-work-webhook-c5875947-d272b                 1/1     Running   0          7m2s
```

All pods should be in the `Running` state with `1/1` readiness and no restarts, indicating a successful hub deployment.

#### 4. Register Spoke Cluster (`demo-worker`)

Obtain the join token from the hub cluster:

```bash
$ clusteradm get token
```

**Output:**

```
token=<Your_Clusteradm_Join_Token>
please log on spoke and run:
clusteradm join --hub-token <Your_Clusteradm_Join_Token> --hub-apiserver https://<hub-apiserver-ip>:6443 --cluster-name <cluster_name>
```

On the `demo-worker` cluster, join it to the hub. Include the `RawFeedbackJsonString` feature gate for resource feedback:

```bash
$ kubectl config use-context demo-worker
$ clusteradm join --hub-token <Your_Clusteradm_Join_Token> --hub-apiserver https://<hub-apiserver-ip>:6443 --cluster-name demo-worker --feature-gates=RawFeedbackJsonString=true
```

#### 5. Accept Spoke Cluster

On the `demo-controller` cluster, accept the `demo-worker` cluster:

```bash
$ kubectl config use-context demo-controller
$ clusteradm accept --clusters demo-worker
```

> **Note:** It may take a few attempts (e.g., retry every 10 seconds) if the cluster is not immediately available.

**Output (on success):**


- Starting approve csrs for the cluster demo-worker
- CSR demo-worker-2p2pb approved
- set hubAcceptsClient to true for managed cluster demo-worker
- Your managed cluster demo-worker has joined the Hub successfully.


#### 6. Verify Namespace Creation

Confirm that a namespace for `demo-worker` was created on the hub cluster:

```bash
$ kubectl get ns
```

**Output:**

```bash
NAME                          STATUS   AGE
default                       Active   99m
demo-worker                   Active   58s
kube-node-lease               Active   99m
kube-public                   Active   99m
kube-system                   Active   99m
open-cluster-management       Active   6m7s
open-cluster-management-hub   Active   5m32s
```

#### 7. Register `demo-controller` as a Spoke Cluster

Repeat the join and accept process for `demo-controller` so it can also act as a spoke cluster:

```bash
$ kubectl config use-context demo-controller
$ clusteradm join --hub-token <Your_Clusteradm_Join_Token> --hub-apiserver https://<hub-apiserver-ip>:6443 --cluster-name demo-controller --feature-gates=RawFeedbackJsonString=true
$ clusteradm accept --clusters demo-controller
```

Verify the namespace for `demo-controller`:

```bash
$ kubectl get ns
```

**Output:**

```
NAME                                  STATUS   AGE
default                               Active   104m
demo-controller                       Active   3s
demo-worker                           Active   6m7s
kube-node-lease                       Active   104m
kube-public                           Active   104m
kube-system                           Active   104m
open-cluster-management               Active   11m
open-cluster-management-agent         Active   37s
open-cluster-management-agent-addon   Active   34s
open-cluster-management-hub           Active   10m
```

#### 8. Verify OCM Roles

After registration, use these commands to confirm which cluster is the hub and which are spokes:

```bash
# Hub: lists all registered spoke clusters
$ kubectl get managedclusters

# Spoke: shows this cluster's registered name
$ kubectl get klusterlet klusterlet -o jsonpath='{.spec.clusterName}'

# Hub components run only on the hub cluster
$ kubectl get pods -n open-cluster-management-hub

# Spoke agent runs on every spoke cluster
$ kubectl get pods -n open-cluster-management-agent
```

### Step 2: Configure OCM WorkConfiguration

Run this on **every spoke cluster** (`demo-controller` and `demo-worker`). This enables the `RawFeedbackJsonString` feature gate that KubeDB requires to read pod status across clusters, and raises the API rate limits to prevent throttling.

> **Note:** Even if you passed `--feature-gates=RawFeedbackJsonString=true` during `clusteradm join`, the rate limit fields are not set by that flag. Run this patch on all spokes regardless.
>
> **Why this matters:** KubeDB uses OCM's ManifestWork feedback mechanism to watch the status of Postgres pods on remote spoke clusters. Without `RawFeedbackJsonString`, the KubeDB provisioner on the hub never receives pod status updates from spokes and the distributed Postgres CR will stay in a non-Ready state indefinitely. The rate limits prevent the klusterlet agent from being API-throttled during initial cluster formation.

```bash
$ kubectl patch klusterlet klusterlet --type=merge -p '{
  "spec": {
    "workConfiguration": {
      "featureGates": [{"feature": "RawFeedbackJsonString", "mode": "Enable"}],
      "hubKubeAPIBurst": 100,
      "hubKubeAPIQPS": 50,
      "kubeAPIBurst": 100,
      "kubeAPIQPS": 50
    }
  }
}'
```

Verify the configuration:

```bash
$ kubectl get klusterlet klusterlet -oyaml
```

**Sample Output (abridged):**

```yaml
apiVersion: operator.open-cluster-management.io/v1
kind: Klusterlet
metadata:
  name: klusterlet
spec:
  clusterName: demo-worker
  workConfiguration:
    featureGates:
    - feature: RawFeedbackJsonString
      mode: Enable
    hubKubeAPIBurst: 100
    hubKubeAPIQPS: 50
    kubeAPIBurst: 100
    kubeAPIQPS: 50
```

### Step 3: Configure KubeSlice for Network Connectivity

KubeSlice enables pod-to-pod communication across clusters. Install the KubeSlice Controller on the `demo-controller` cluster and the KubeSlice Worker on both `demo-controller` and `demo-worker` clusters.

#### 1. Install KubeSlice Controller

On `demo-controller`, get the hub API server address first:

```bash
$ kubectl cluster-info | grep 'Kubernetes control plane'
```

Use the IP and port from that output as the `endpoint` value. Create a `controller.yaml` file:

```yaml
kubeslice:
  controller:
    loglevel: info
    rbacResourcePrefix: kubeslice-rbac
    projectnsPrefix: kubeslice
    endpoint: https://<hub-apiserver-ip>:6443
```

Deploy the controller using Helm:

```bash
$ helm upgrade -i kubeslice-controller oci://ghcr.io/appscode-charts/kubeslice-controller \
    --version v2026.1.15 \
    -f controller.yaml \
    --namespace kubeslice-controller \
    --create-namespace \
    --set ocm.enabled=true \
    --wait --burst-limit=10000 --debug
```

Verify the installation:

```bash
$ kubectl get pods -n kubeslice-controller
```

**Output:**

```
NAME                                            READY   STATUS    RESTARTS   AGE
kubeslice-controller-manager-7fd756fff6-5kddd   2/2     Running   0          98s
```

#### 2. Create a KubeSlice Project

Create a `project.yaml` file:

```yaml
apiVersion: controller.kubeslice.io/v1alpha1
kind: Project
metadata:
  name: demo-distributed-postgres
  namespace: kubeslice-controller
spec:
  serviceAccount:
    readWrite:
      - admin
```

Apply the project:

```bash
$ kubectl apply -f project.yaml
```

Verify:

```bash
$ kubectl get project -n kubeslice-controller
```

**Output:**

```bash
NAME                        AGE
demo-distributed-postgres   31s
```

Check service accounts:

```bash
$ kubectl get sa -n kubeslice-demo-distributed-postgres
```

**Output:**

```
NAME                      SECRETS   AGE
default                   0         69s
kubeslice-rbac-rw-admin   1         68s
```

#### 3. Label Nodes for KubeSlice

Assign the `kubeslice.io/node-type=gateway` label to the node where the worker operator will be deployed in both clusters.

On `demo-controller`:

```bash
$ kubectl get nodes
$ kubectl label node <node-name> kubeslice.io/node-type=gateway
```

On `demo-worker`:

```bash
$ kubectl config use-context demo-worker
$ kubectl get nodes
$ kubectl label node <node-name> kubeslice.io/node-type=gateway
```

#### 4. Register Clusters with KubeSlice

Identify the network interface for each cluster by running the following command **on the gateway node of each cluster**:

```bash
$ ip route get 8.8.8.8 | awk '{ print $5 }'
```

**Output (example):**

```
enp1s0
```

> **Important:** The `networkInterface` value must match the primary network interface of each cluster's gateway node. Run the command above on **each cluster separately** and use that cluster's output as its `networkInterface` value. Each cluster may have a **different interface name** — do not assume they are the same. Using the wrong interface name will cause the WireGuard gateway to silently fail and cross-cluster Postgres replication will never connect.
>
> Example: if `demo-controller` returns `enp3s0` and `demo-worker` returns `eth0`, use those exact values in the YAML below.

Create a `registration.yaml` file:

> **Note:**
> - The cluster name must exactly match the name of the OCM (spoke) cluster.
> - The corresponding `ManagedClusterAddOn` resource must be created in the namespace that bears the same name as the cluster to set up the KubeSlice worker automatically.

```yaml
apiVersion: controller.kubeslice.io/v1alpha1
kind: Cluster
metadata:
  name: demo-controller
  namespace: kubeslice-demo-distributed-postgres
spec:
  networkInterface: <demo-controller-interface>   # replace with output of ip route command
  clusterProperty: {}
---
apiVersion: controller.kubeslice.io/v1alpha1
kind: Cluster
metadata:
  name: demo-worker
  namespace: kubeslice-demo-distributed-postgres
spec:
  networkInterface: <demo-worker-interface>   # replace with output of ip route command
  clusterProperty: {}
---
apiVersion: addon.open-cluster-management.io/v1alpha1
kind: ManagedClusterAddOn
metadata:
  name: kubeslice
  namespace: demo-controller
spec:
  installNamespace: kubeslice-system
  configs:
    - name: demo-controller
      namespace: kubeslice-demo-distributed-postgres
      group: controller.kubeslice.io
      resource: clusters
---
apiVersion: addon.open-cluster-management.io/v1alpha1
kind: ManagedClusterAddOn
metadata:
  name: kubeslice
  namespace: demo-worker
spec:
  installNamespace: kubeslice-system
  configs:
    - name: demo-worker
      namespace: kubeslice-demo-distributed-postgres
      group: controller.kubeslice.io
      resource: clusters
```

Apply on `demo-controller`:

```bash
$ kubectl apply -f registration.yaml
```

Verify OCM is deploying the KubeSlice worker manifests to each cluster:

```bash
$ kubectl get managedclusteraddon -A
```

**Output:**

```
NAMESPACE         NAME        AVAILABLE   DEGRADED   PROGRESSING
demo-controller   kubeslice   Unknown                True
demo-worker       kubeslice   Unknown                True
```

`PROGRESSING: True` means OCM is actively deploying. Wait until `kubeslice-operator` shows `2/2 Running` on both clusters before proceeding:

```bash
# Run on each spoke cluster
$ kubectl get pods -n kubeslice-system --watch
```

**Expected output (after KubeSlice worker is fully deployed):**

```
NAME                                 READY   STATUS      RESTARTS   AGE
forwarder-kernel-bw5l4               1/1     Running     0          4m43s
kubeslice-dns-6bd9749f4d-pvh7g       1/1     Running     0          4m43s
kubeslice-install-crds-szhvc         0/1     Completed   0          4m56s
kubeslice-netop-g4dfn                1/1     Running     0          4m43s
kubeslice-operator-949b7d6f7-9wj7h   2/2     Running     0          4m43s
nsc-grpc-server-sbjj7                1/1     Running     0          4m43s
nsm-install-crds-5z4j9               0/1     Completed   0          4m53s
nsmgr-zzwgh                          2/2     Running     0          4m43s
registry-k8s-979455d6d-q2j8x         1/1     Running     0          4m43s
```

#### 5. Onboard Application Namespace

Create a `SliceConfig` to onboard the `demo` (application) and `kubedb` (operator) namespaces for network connectivity. Create a `sliceconfig.yaml` file:

```yaml
apiVersion: controller.kubeslice.io/v1alpha1
kind: SliceConfig
metadata:
  name: demo-slice
  namespace: kubeslice-demo-distributed-postgres
spec:
  sliceSubnet: 10.1.0.0/16
  maxClusters: 16
  sliceType: Application
  sliceGatewayProvider:
    sliceGatewayType: Wireguard
    sliceCaType: Local
  sliceIpamType: Local
  rotationInterval: 60
  vpnConfig:
    cipher: AES-128-CBC
  clusters:
    - demo-controller
    - demo-worker
  qosProfileDetails:
    queueType: HTB
    priority: 1
    tcType: BANDWIDTH_CONTROL
    bandwidthCeilingKbps: 5120
    bandwidthGuaranteedKbps: 2560
    dscpClass: AF11
  namespaceIsolationProfile:
    applicationNamespaces:
      - namespace: demo
        clusters:
          - '*'
      - namespace: kubedb
        clusters:
          - '*'
    isolationEnabled: false
    allowedNamespaces:
      - namespace: kube-system
        clusters:
          - '*'
```

Apply the `SliceConfig`:

```bash
$ kubectl apply -f sliceconfig.yaml
```

After the SliceConfig is applied, a `vl3-slice-router` pod will appear in `kubeslice-system` on each cluster, indicating the slice VPN tunnel is being established.

#### 6. Configure DNS for KubeSlice

Update CoreDNS to forward `*.slice.local` traffic to the KubeSlice DNS service. Run the following steps on **every cluster** in the slice.

> **Important:** CoreDNS must be updated and restarted on all clusters before proceeding to Step 4 (KubeDB install). Postgres nodes use `.slice.local` DNS names to discover each other across clusters. If DNS is not configured before Postgres pods start, replication will not form.

Get the KubeSlice DNS service IP address on each cluster:

```bash
$ kubectl get svc -n kubeslice-system -owide -l 'app=kubeslice-dns'
```

**Output:**

```bash
NAME            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)         AGE   SELECTOR
kubeslice-dns   ClusterIP   10.43.172.191   <none>        53/UDP,53/TCP   8d    app=kubeslice-dns
```

Add the following block to the **top** of your CoreDNS `Corefile` ConfigMap (replace the IP with the one from your cluster):

```
slice.local:53 {
    errors
    cache 30
    forward . 10.43.172.191
}
```

Example of the full CoreDNS ConfigMap after editing:

```bash
$ kubectl get cm -n kube-system coredns -oyaml
```

**Output:**

```yaml
apiVersion: v1
data:
  Corefile: |
    slice.local:53 {
        errors
        cache 30
        forward . 10.43.172.191
    }
    .:53 {
        errors
        health
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
          pods insecure
          fallthrough in-addr.arpa ip6.arpa
        }
        hosts /etc/coredns/NodeHosts {
          ttl 60
          reload 15s
          fallthrough
        }
        prometheus :9153
        cache 30
        loop
        reload
        loadbalance
        import /etc/coredns/custom/*.override
        forward . /etc/resolv.conf
    }
    import /etc/coredns/custom/*.server
  NodeHosts: |
    10.2.0.248 demo-worker
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
```

After editing the ConfigMap, restart CoreDNS to apply the change:

```bash
$ kubectl rollout restart deploy/coredns -n kube-system
```

Repeat the DNS configuration steps on every cluster in the slice.

### Step 4: Install the KubeDB Operator

> **Note:** Install the KubeDB operator only on the hub cluster (`demo-controller`). The operator manages Postgres pods on spoke clusters through OCM — no KubeDB installation is needed on `demo-worker`.

#### Get a Free License

The KubeDB license is tied to the `kube-system` namespace UID of the hub cluster and has an expiry date. Get your cluster UID and verify the license before installing:

```bash
# Get your cluster UID (required when requesting the license)
$ kubectl get ns kube-system -o jsonpath='{.metadata.uid}'

# Verify the license is not expired
$ openssl x509 -noout -enddate -in $HOME/Downloads/kubedb-license-<uid>.txt
```

If expired or not yet obtained, download a FREE license from the [AppsCode License Server](https://appscode.com/issue-license?p=kubedb) using the cluster UID above.

```bash
$ helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
    --version v2026.2.26 \
    --namespace kubedb --create-namespace \
    --set-file global.license=$HOME/Downloads/kubedb-license-<uid>.txt \
    --set petset.features.ocm.enabled=true \
    --wait --burst-limit=10000 --debug
```

> **Note:** The `--set petset.features.ocm.enabled=true` flag must be set to enable the Postgres Distributed feature.

For additional details, refer to the [KubeDB Installation Guide](https://kubedb.com/docs/v2026.4.27/setup/).

Verify that the pods are running:

```bash
$ kubectl get pods -n kubedb
```

**Output:**

```bash
NAME                                           READY   STATUS    RESTARTS   AGE
kubedb-kubedb-autoscaler-0                     1/2     Running   0          44s
kubedb-kubedb-ops-manager-0                    1/2     Running   0          43s
kubedb-kubedb-provisioner-0                    1/2     Running   0          43s
kubedb-kubedb-webhook-server-df667cd85-tjdp9   2/2     Running   0          44s
kubedb-petset-cf9f5b6f4-d9558                  2/2     Running   0          44s
kubedb-sidekick-5dbf7bcf64-4b8cw               2/2     Running   0          44s
```

### Step 5: Define a PlacementPolicy

You can define the `storageClassName` under `spec.clusterSpreadConstraint.distributionRules` for each cluster. If not explicitly specified, the clusters will automatically select an appropriate storage class. Additionally, you can control replica placement by defining which replicas belong to which cluster. When scaling the number of replicas, the distribution must also be specified at the cluster level. In this example, we are using three replicas.

To manage pod distribution across clusters, create a `PlacementPolicy`. For this purpose, define a `pod-placement-policy.yaml` file as shown below:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  labels:
    app.kubernetes.io/managed-by: Helm
  name: distributed-postgres
spec:
  clusterSpreadConstraint:
    distributionRules:
      - clusterName: demo-controller
        storageClassName: local-path   # optional; omit to use the cluster's default storage class
        replicaIndices:
          - 0
          - 2
      - clusterName: demo-worker
        storageClassName: local-path   # optional; omit to use the cluster's default storage class
        replicaIndices:
          - 1
    slice:
      projectNamespace: kubeslice-demo-distributed-postgres
      sliceName: demo-slice
  nodeSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
  zoneSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
```

This policy schedules:

- `postgres-0` and `postgres-2` on `demo-controller`
- `postgres-1` on `demo-worker`

Apply the policy on `demo-controller`:

```bash
$ kubectl apply -f pod-placement-policy.yaml --context demo-controller --kubeconfig $HOME/.kube/config
```

### Step 6: Create a Distributed Postgres Instance

Create the `demo` namespace first:

```bash
$ kubectl create namespace demo
```

Define a Postgres custom resource with `spec.distributed` set to `true` and reference the `PlacementPolicy`. Create a `postgres.yaml` file:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: postgres
  namespace: demo
spec:
  distributed: true
  deletionPolicy: WipeOut
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  storageType: Durable
  version: "17.2"
  podTemplate:
    spec:
      podPlacementPolicy:
        name: distributed-postgres
```

Apply the resource on `demo-controller`:

```bash
$ kubectl apply -f postgres.yaml --context demo-controller --kubeconfig $HOME/.kube/config
```

### Step 7: Verify the Deployment

#### 1. Check Postgres Resource and Pods on `demo-controller`

```bash
$ kubectl get pg,pods,secret -n demo --context demo-controller --kubeconfig $HOME/.kube/config
```

**Output:**

```bash
NAME                           VERSION   STATUS   AGE
postgres.kubedb.com/postgres   17.2      Ready    99s

NAME              READY   STATUS    RESTARTS   AGE
pod/postgres-0    2/2     Running   0          95s
pod/postgres-2    2/2     Running   0          95s

NAME                    TYPE                       DATA   AGE
secret/postgres-auth    kubernetes.io/basic-auth   2      95s
```

#### 2. Check Pods and Secrets on `demo-worker`

```bash
$ kubectl get pods,secrets -n demo --context demo-worker --kubeconfig $HOME/.kube/config
```

**Output:**

```bash
NAME           READY   STATUS    RESTARTS   AGE
pod/postgres-1 2/2     Running   0          95s

NAME                    TYPE                       DATA   AGE
secret/postgres-auth    kubernetes.io/basic-auth   2      95s
```

#### 3. Verify Replication Status

Connect to the primary Postgres pod and check the replication status:

```bash
$ kubectl exec -it -n demo pod/postgres-0 --context demo-controller -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
postgres-0:/$ psql -U postgres
```

Run the following query:

```sql
SELECT * FROM pg_stat_replication;
```

**Output:**

```
 pid  | usesysid | usename  | application_name | client_addr | client_hostname | client_port |         backend_start         | backend_xmin |   state   | sent_lsn  | write_lsn | flush_lsn | replay_lsn |    write_lag    |   flush_lag    |   replay_lag    | sync_priority | sync_state |          reply_time           
------+----------+----------+------------------+-------------+-----------------+-------------+-------------------------------+--------------+-----------+-----------+-----------+-----------+------------+-----------------+----------------+-----------------+---------------+------------+-------------------------------
  495 |       10 | postgres | postgres-1       | 10.1.0.12   |                 |       49492 | 2025-08-26 12:03:47.940257+00 |              | streaming | 0/50001B0 | 0/50001B0 | 0/50001B0 | 0/50001B0  | 00:00:00.000212 | 00:00:00.00207 | 00:00:00.00212  |             0 | async      | 2025-08-26 12:05:33.011941+00
 1183 |       10 | postgres | postgres-2       | 10.42.0.82  |                 |       58254 | 2025-08-26 12:05:28.257383+00 |              | streaming | 0/50001B0 | 0/50001B0 | 0/50001B0 | 0/50001B0  | 00:00:00.000179 | 00:00:00.00185 | 00:00:00.001912 |             0 | async      | 2025-08-26 12:05:33.01175+00
(2 rows)
```

Both `postgres-1` (on `demo-worker`, connected via KubeSlice) and `postgres-2` (on `demo-controller`) should appear in streaming replication state.

## Next Steps

- **Accessing the Database**: Use the `postgres-auth` secret to retrieve credentials and connect to the Postgres instance.
- **Scaling**: Adjust the `PlacementPolicy` to add or remove replicas across clusters.
- **Monitoring**: Integrate KubeDB with monitoring tools like Prometheus for cluster health insights.

