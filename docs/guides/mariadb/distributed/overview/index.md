---
title: Distributed MariaDB Galera Cluster Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-distributed-overview
    name: Distributed MariaDB Overview
    parent: guides-mariadb-distributed
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
# Distributed MariaDB Galera Cluster Overview

## Introduction

KubeDB enables distributed MariaDB deployments using the Galera clustering technology across multiple Kubernetes clusters, providing a scalable, highly available, and resilient database solution. By integrating **Open Cluster Management (OCM)** for multi-cluster orchestration and **KubeSlice** for seamless pod-to-pod network connectivity, KubeDB simplifies the deployment and management of MariaDB instances across clusters. A **PodPlacementPolicy** ensures precise control over pod scheduling, allowing you to distribute MariaDB pods across clusters for optimal resource utilization and fault tolerance.

This guide provides a step-by-step process to deploy a distributed MariaDB Galera cluster, including prerequisites, configuration, and verification steps. It assumes familiarity with Kubernetes and basic database concepts.

> **New to KubeDB?** Start with the [KubeDB documentation](https://kubedb.com/docs/v2025.7.31/welcome/) for an introduction.

## Understanding OCM Hub and Spoke Clusters

In an **Open Cluster Management (OCM)** setup, clusters are categorized as:

- **Hub Cluster**: The central control plane where policies, applications, and resources are defined and managed. It orchestrates the lifecycle of applications deployed across spoke clusters.
- **Spoke Cluster**: Managed clusters registered with the hub, running the actual workloads (e.g., MariaDB pods).

When a spoke cluster (e.g., `demo-worker`) is joined to the hub using the `clusteradm join` command, OCM creates a namespace on the hub cluster matching the spoke cluster's name (e.g., `demo-worker`). This namespace is used to manage resources specific to the spoke cluster from the hub.

## Prerequisites

Before deploying a distributed MariaDB Galera cluster, ensure the following:

- **Kubernetes Clusters**: Multiple Kubernetes clusters (version 1.26 or higher) configured and accessible.
- **Node Requirements**: Each Kubernetes node should have at least 4 vCPUs and 16GB of RAM.
- **Open Cluster Management (OCM)**: Install `clusteradm` as per the [OCM Quick Start Guide](https://open-cluster-management.io/docs/getting-started/quick-start/).
- **kubectl**: Installed and configured to interact with all clusters.
- **Helm**: Installed for deploying the KubeDB Operator and KubeSlice components.
- **Persistent Storage**: A storage class (e.g., `local-path` or cloud provider-specific) configured for persistent volumes.

## Configuration Steps

Follow these steps to deploy a distributed MariaDB Galera cluster across multiple Kubernetes clusters.

### Step 1: Set Up Open Cluster Management (OCM)

1. **Configure KUBECONFIG**:
   Ensure your `KUBECONFIG` is set up to switch between clusters. This guide uses two clusters: `demo-controller` (hub and spoke) and `demo-worker` (spoke).

   ```bash
   kubectl config get-contexts
   ```

   **Output**:
   ```
   CURRENT   NAME              CLUSTER           AUTHINFO          NAMESPACE
   *         demo-controller   demo-controller   demo-controller   
             demo-worker       demo-worker       demo-worker
   ```

2. **Initialize the OCM Hub**:
   On the `demo-controller` cluster, initialize the OCM hub.

   ```bash
   kubectl config use-context demo-controller
   clusteradm init --wait --feature-gates=ManifestWorkReplicaSet=true
   ```

3. **Verify Hub Deployment**:
   Check the pods in the `open-cluster-management-hub` namespace to ensure all components are running.

   ```bash
   kubectl get pods -n open-cluster-management-hub
   ```

   **Output**:
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

4. **Register Spoke Cluster (`demo-worker`)**:
   Obtain the join token from the hub cluster.

   ```bash
   clusteradm get token
   ```

   **Output**:
   ```
   token=<Your_Clusteradm_Join_Token>
   please log on spoke and run:
   clusteradm join --hub-token <Your_Clusteradm_Join_Token> --hub-apiserver https://10.2.0.56:6443 --cluster-name <cluster_name>
   ```

   On the `demo-worker` cluster, join it to the hub, including the `RawFeedbackJsonString` feature gate for resource feedback.

   ```bash
   kubectl config use-context demo-worker
   clusteradm join --hub-token <Your_Clusteradm_Join_Token> --hub-apiserver https://10.2.0.56:6443 --cluster-name demo-worker --feature-gates=RawFeedbackJsonString=true
   ```

5. **Accept Spoke Cluster**:
   On the `demo-controller` cluster, accept the `demo-worker` cluster.

   ```bash
   kubectl config use-context demo-controller
   clusteradm accept --clusters demo-worker
   ```

   **Note**: It may take a few attempts (e.g., retry every 10 seconds) if the cluster is not immediately available.

   **Output** (on success):
   ```
   Starting approve csrs for the cluster demo-worker
   CSR demo-worker-2p2pb approved
   set hubAcceptsClient to true for managed cluster demo-worker
   Your managed cluster demo-worker has joined the Hub successfully.
   ```

6. **Verify Namespace Creation**:
   Confirm that a namespace for `demo-worker` was created on the hub cluster.

   ```bash
   kubectl get ns
   ```

   **Output**:
   ```
   NAME                          STATUS   AGE
   default                       Active   99m
   demo-worker                   Active   58s
   kube-node-lease               Active   99m
   kube-public                   Active   99m
   kube-system                   Active   99m
   open-cluster-management       Active   6m7s
   open-cluster-management-hub   Active   5m32s
   ```

7. **Register `demo-controller` as a Spoke Cluster**:
   Repeat the join and accept process for `demo-controller` to also act as a spoke cluster.

   ```bash
   kubectl config use-context demo-controller
   clusteradm join --hub-token <Your_Clusteradm_Join_Token> --hub-apiserver https://10.2.0.56:6443 --cluster-name demo-controller --feature-gates=RawFeedbackJsonString=true
   clusteradm accept --clusters demo-controller
   ```

   Verify the namespace for `demo-controller`.

   ```bash
   kubectl get ns
   ```

   **Output**:
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

### Step 2: Configure OCM WorkConfiguration (Optional)

If you did not follow the provided OCM installation steps, update the `klusterlet` resource to enable feedback retrieval.

```bash
kubectl edit klusterlet klusterlet
```

Add the following under the `spec` field:

```yaml
workConfiguration:
  featureGates:
    - feature: RawFeedbackJsonString
      mode: Enable
  hubKubeAPIBurst: 100
  hubKubeAPIQPS: 50
  kubeAPIBurst: 100
  kubeAPIQPS: 50
```

Verify the configuration:

```bash
kubectl get klusterlet klusterlet -oyaml
```

**Sample Output** (abridged):
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

1. **Install KubeSlice Controller**:
   On `demo-controller`, create a `controller.yaml` file:

   ```yaml
   kubeslice:
     controller:
       loglevel: info
       rbacResourcePrefix: kubeslice-rbac
       projectnsPrefix: kubeslice
       endpoint: https://10.2.0.56:6443
   ```

   Deploy the controller using Helm:

   ```bash
   helm upgrade -i kubeslice-controller oci://ghcr.io/appscode-charts/kubeslice-controller \
       --version v2025.7.31 \
       -f controller.yaml \
       --namespace kubeslice-controller \
       --create-namespace \
       --wait --burst-limit=10000 --debug
   ```

   Verify the installation:

   ```bash
   kubectl get pods -n kubeslice-controller
   ```

   **Output**:
   ```
   NAME                                            READY   STATUS    RESTARTS   AGE
   kubeslice-controller-manager-7fd756fff6-5kddd   2/2     Running   0          98s
   ```

2. **Create a KubeSlice Project**:
   Create a `project.yaml` file:

   ```yaml
   apiVersion: controller.kubeslice.io/v1alpha1
   kind: Project
   metadata:
     name: demo-distributed-mariadb
     namespace: kubeslice-controller
   spec:
     serviceAccount:
       readWrite:
         - admin
   ```

   Apply the project:

   ```bash
   kubectl apply -f project.yaml
   ```

   Verify:

   ```bash
   kubectl get project -n kubeslice-controller
   ```

   **Output**:
   ```
   NAME                       AGE
   demo-distributed-mariadb   31s
   ```

   Check service accounts:

   ```bash
   kubectl get sa -n kubeslice-demo-distributed-mariadb
   ```

   **Output**:
   ```
   NAME                      SECRETS   AGE
   default                   0         69s
   kubeslice-rbac-rw-admin   1         68s
   ```

3. **Label Nodes for KubeSlice**:
   Assign the `kubeslice.io/node-type=gateway` label to node(where worker operator will deploy) in both clusters.

   On `demo-controller`:

   ```bash
   kubectl get nodes
   kubectl label node demo-master kubeslice.io/node-type=gateway
   ```

   On `demo-worker`:

   ```bash
   kubectl config use-context demo-worker
   kubectl get nodes
   kubectl label node demo-worker kubeslice.io/node-type=gateway
   ```

4. **Register Clusters with KubeSlice**:
   Identify the network interface for both clusters by running the command on the node.

   ```bash
   ip route get 8.8.8.8 | awk '{ print $5 }'
   ```

   **Output** (example):
   ```
   enp1s0
   ```

   Create a `registration.yaml` file:

   Cluster name should match the OCM cluster names(spoke). The managedClusterAddOn resource should be created in the namespace with the same name as the cluster.
   ```yaml
   apiVersion: controller.kubeslice.io/v1alpha1
   kind: Cluster
   metadata:
     name: demo-controller
     namespace: kubeslice-demo-distributed-mariadb
   spec:
     networkInterface: enp1s0
     clusterProperty: {}
   ---
   apiVersion: controller.kubeslice.io/v1alpha1
   kind: Cluster
   metadata:
     name: demo-worker
     namespace: kubeslice-demo-distributed-mariadb
   spec:
     networkInterface: enp1s0
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
         namespace: kubeslice-demo-distributed-mariadb
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
         namespace: kubeslice-demo-distributed-mariadb
         group: controller.kubeslice.io
         resource: clusters
   ---
   ```

   Apply on `demo-controller`:

   ```bash
   kubectl apply -f registration.yaml
   ```

   Verify:

   ```bash
   kubectl get clusters -n kubeslice-demo-distributed-mariadb
   ```

   **Output**:
   ```
   NAME              AGE
   demo-controller   9s
   demo-worker       9s
   ```

   Verify the worker installation:

   ```bash
   kubectl get pods -n kubeslice-system
   ```

   **Output**:
   ```
   NAME                                 READY   STATUS      RESTARTS   AGE
   forwarder-kernel-bw5l4               1/1     Running     0          4m43s
   kubeslice-dns-6bd9749f4d-pvh7g       1/1     Running     0          4m43s
   kubeslice-install-crds-szhvc         0/1     Completed   0          4m56s
   kubeslice-netop-g4dfn                1/1     Running     0          4m43s
   kubeslice-operator-949b7d6f7-9wj7h   2/2     Running     0          4m43s
   kubeslice-postdelete-job-ctlzt       0/1     Completed   0          20m
   nsm-delete-webhooks-ndksl            0/1     Completed   0          20m
   nsm-install-crds-5z4j9               0/1     Completed   0          4m53s
   nsmgr-zzwgh                          2/2     Running     0          4m43s
   registry-k8s-979455d6d-q2j8x         1/1     Running     0          4m43s
   spire-install-clusterid-cr-qwqlr     0/1     Completed   0          4m47s
   spire-install-crds-cnbjh             0/1     Completed   0          4m50s
   ```
6. **Onboard Application Namespace**:

Create a `SliceConfig` to onboard the `demo` (application) and `kubedb` (operator) namespaces for network connectivity. Create a `sliceconfig.yaml` file:

```yaml
apiVersion: controller.kubeslice.io/v1alpha1
kind: SliceConfig
metadata:
  name: demo-slice
  namespace: kubeslice-demo-distributed-mariadb
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
kubectl apply -f sliceconfig.yaml
```

update the 


### Step 4: Install the KubeDB Operator

Install the KubeDB Operator on the `demo-controller` cluster to manage the MariaDB instance.

#### Get a Free License
Download a FREE license from [AppsCode License Server](https://appscode.com/issue-license?p=kubedb). Get the license for `demo-controller` cluster.

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
    --version v2025.8.31 \
    --namespace kubedb --create-namespace \
    --set-file global.license=$HOME/Downloads/kubedb-license-cd548cce-5141-4ed3-9276-6d9578707f12.txt \
    --set petset.features.ocm.enabled=true \
    --wait --burst-limit=10000 --debug
```
Note: `--set petset.features.ocm.enabled=true` must be set to enable MariaDB Distributed feature.

Follow the [KubeDB Installation Guide](https://kubedb.com/docs/v2025.6.30/setup/install/kubedb/) for additional details.

Verify the pods are running:

```bash
kubectl get pods -n kubedb
```
**Output**:
   ```
   NAME                                           READY   STATUS    RESTARTS   AGE
kubedb-kubedb-autoscaler-0                     1/2     Running   0          44s
kubedb-kubedb-ops-manager-0                    1/2     Running   0          43s
kubedb-kubedb-provisioner-0                    1/2     Running   0          43s
kubedb-kubedb-webhook-server-df667cd85-tjdp9   2/2     Running   0          44s
kubedb-petset-cf9f5b6f4-d9558                  2/2     Running   0          44s
kubedb-sidekick-5dbf7bcf64-4b8cw               2/2     Running   0          44s
   ```


### Step 5: Define a PodPlacementPolicy

Create a `PodPlacementPolicy` to control pod distribution across clusters. Create a `pod-placement-policy.yaml` file:

```yaml
apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
   labels:
      app.kubernetes.io/managed-by: Helm
   name: distributed-mariadb
spec:
   clusterSpreadConstraint:
      distributionRules:
         - clusterName: demo-controller
           replicaIndices:
              - 0
              - 2
         - clusterName: demo-worker
           replicaIndices:
              - 1
      slice:
         projectNamespace: kubeslice-demo-distributed-mariadb
         sliceName: demo-slice
   nodeSpreadConstraint:
      maxSkew: 1
      whenUnsatisfiable: ScheduleAnyway
   zoneSpreadConstraint:
      maxSkew: 1
      whenUnsatisfiable: ScheduleAnyway
```

This policy schedules:
- `mariadb-0` and `mariadb-2` on `demo-controller`.
- `mariadb-1` on `demo-worker`.

Apply the policy on `demo-controller`:

```bash
kubectl apply -f pod-placement-policy.yaml --context demo-controller --kubeconfig $HOME/.kube/config
```

### Step 6: Create a Distributed MariaDB Instance

Define a MariaDB custom resource with `spec.distributed` set to `true` and reference the `PodPlacementPolicy`. Create a `mariadb.yaml` file:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb
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
        storage: 500Mi
  storageType: Durable
  version: 11.5.2
  podTemplate:
    spec:
      podPlacementPolicy:
        name: distributed-mariadb
```

Apply the resource on `demo-controller`:

```bash
kubectl apply -f mariadb.yaml --context demo-controller --kubeconfig $HOME/.kube/config
```

### Step 7: Verify the Deployment

1. **Check MariaDB Resource and Pods on `demo-controller`**:

   ```bash
   kubectl get md,pods,secret -n demo --context demo-controller --kubeconfig $HOME/.kube/config
   ```

   **Output**:
   ```
   NAME                         VERSION   STATUS   AGE
   mariadb.kubedb.com/mariadb   11.5.2    Ready    99s

   NAME            READY   STATUS    RESTARTS   AGE
   pod/mariadb-0   3/3     Running   0          95s
   pod/mariadb-2   3/3     Running   0          95s

   NAME                  TYPE                       DATA   AGE
   secret/mariadb-auth   kubernetes.io/basic-auth   2      95s
   ```

2. **Check Pods and Secrets on `demo-worker`**:

   ```bash
   kubectl get pods,secrets -n demo --context demo-worker --kubeconfig $HOME/.kube/config
   ```

   **Output**:
   ```
   NAME        READY   STATUS    RESTARTS   AGE
   mariadb-1   3/3     Running   0          95s

   NAME                  TYPE                       DATA   AGE
   secret/mariadb-auth   kubernetes.io/basic-auth   2      95s
   ```

3. **Verify Galera Cluster Status**:
   Connect to a MariaDB pod and check the Galera cluster status:
   The primary service DNS will be formatted as <database-name>.<database-namespace>.svc

   ```bash
   kubectl exec -it -n demo pod/mariadb-0 --context demo-controller -- bash
   mariadb -uroot -p$MYSQL_ROOT_PASSWORD -hmariadb.demo.svc
   ```

   Run the following query:

   ```sql
   SHOW STATUS LIKE 'wsrep_cluster_status';
   ```

   **Output**:
   ```
   +----------------------+---------+
   | Variable_name        | Value   |
   +----------------------+---------+
   | wsrep_cluster_status | Primary |
   +----------------------+---------+
   1 row in set (0.001 sec)
   ```

   Check additional Galera status variables:

   ```sql
   SHOW STATUS LIKE 'wsrep%';
   ```

   **Key Indicators**:
    - `wsrep_cluster_status: Primary`: The cluster is fully operational.
    - `wsrep_cluster_size: 3`: All three nodes are part of the cluster.
    - `wsrep_connected: ON`: The node is connected to the cluster.
    - `wsrep_ready: ON`: The node is ready to accept queries.
    - `wsrep_incoming_addresses`: Lists the IP addresses of all nodes (e.g., `10.1.0.3:0,10.1.0.4:0,10.1.16.4:0`).

## Troubleshooting Tips

- **Pods Not Running**: Check pod logs (`kubectl logs -n demo mariadb-0`) for errors related to storage, networking, or configuration.
- **Cluster Not Joining**: Ensure the `RawFeedbackJsonString` feature gate is enabled and verify network connectivity between clusters.
- **KubeSlice Issues**: Confirm that the network interface (`enp1s0`) matches your cluster's configuration and that sidecar containers are injected.
- **MariaDB Not Synced**: Check `wsrep_local_state_comment` (should be `Synced`) and ensure all nodes have the same `wsrep_cluster_state_uuid`.

## Next Steps

- **Accessing the Database**: Use the `mariadb-auth` secret to retrieve credentials and connect to the MariaDB instance.
- **Scaling**: Adjust the `PodPlacementPolicy` to add or remove replicas across clusters.
- **Monitoring**: Integrate KubeDB with monitoring tools like Prometheus for cluster health insights.

For further details, refer to the [KubeDB Documentation](https://kubedb.com/docs/v2025.7.31/) 