---
title: MariaDB Galera Cluster Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-clustering-overview
    name: MariaDB Clustering Overview
    parent: guides-mariadb-clustering
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Distributed MariaDB

KubeDB Supports distributed MariaDB deployments across multiple Kubernetes clusters, enabling scalable and resilient database architectures. By leveraging Open Cluster Management (OCM) for streamlined multi-cluster management and KubeSlice for seamless pod-to-pod network connectivity, you can deploy MariaDB instances across clusters with high availability and efficient resource utilization. The deployment uses a PodPlacementPolicy to control pod scheduling on specific clusters, ensuring precise workload distribution.

This guide walks you through the prerequisites, configuration steps, and an example for deploying a distributed MariaDB instance with these technologies.

# Prerequisites

Before deploying a distributed MariaDB instance, ensure the following are in place:

- Kubernetes Clusters: Multiple Kubernetes clusters (version 1.26 or higher) configured and accessible.

- Open Cluster Management (OCM): Clusteradm should be installed. See the [documentaion](https://open-cluster-management.io/docs/getting-started/quick-start/)

- kubectl: Installed and configured to interact with your Kubernetes clusters.

- Helm: Installed for deploying the KubeDB Operator and related resources.

- Persistent Storage: A storage class configured for persistent volumes (e.g., local-path or cloud provider-specific storage).


# Configuration Steps

Follow these steps to deploy a distributed MariaDB instance across multiple clusters:

### Set Up OCM for Multi-Cluster Management:

We have two cluster `demo-controller` and `demo-worker`. Here we will use `demo-controller` as hub cluster and `demo-worker` as spoke cluster. The hub `demo-controller` cluster will also be used as spoke cluster.

```bash
➤ kubectl config get-contexts
CURRENT   NAME              CLUSTER           AUTHINFO          NAMESPACE
*         demo-controller   demo-controller   demo-controller   
          demo-worker       demo-worker       demo-worker
```

Initialize the OCM hub cluster using the clusteradm init command.
```bash 

clusteradm init --wait
```

Check the deployment 
```bash
kubectl get ns
kubectl get pods -n open-cluster-management-hub
```

Now get the tokoen which will be used to register spoke cluster.

```bash

clusteradm get token
```
Now run the command


- Join managed clusters to the hub using the clusteradm join command with the hub's token and API server endpoint. Ensure the double opt-in mechanism is completed for secure registration.

- Verify that the managed clusters are registered and their namespaces are created in the hub cluster (kubectl get managedclusters -A).

### Configure KubeSlice for Network Connectivity:

- Install the KubeSlice controller on the hub cluster and KubeSlice workers on managed clusters.

- Create a KubeSlice slice configuration to enable pod-to-pod communication across clusters. Ensure the slice includes the MariaDB service ports (default: 3306).

- Validate network connectivity by checking that pods in different clusters can communicate via their cluster IPs or service endpoints.

### Install the KubeDB Operator:

- Install the MariaDB Operator on each managed cluster using Helm:

- helm repo add mariadb-operator https://helm.mariadb.com/mariadb-operator
- helm install mariadb-operator mariadb-operator/mariadb-operator
- Verify the operator is running (kubectl get pods -n mariadb-operator).

### Define a PodPlacementPolicy:

- Create a PodPlacementPolicy custom resource in the hub cluster to specify which clusters should host MariaDB pods. The policy uses labels, taints, or tolerations to control scheduling.

Example PodPlacementPolicy YAML:

apiVersion: cluster.open-cluster-management.io/v1alpha1
kind: PodPlacementPolicy
metadata:
name: mariadb-placement
namespace: open-cluster-management
spec:
clusterSelector:
matchLabels:
environment: production
region: us-east
taints:
- key: gpu
value: available
effect: PreferNoSelect
tolerations:
- key: gpu
operator: Exists
effect: PreferNoSelect

Apply the policy:

kubectl apply -f pod-placement-policy.yaml



### Create a Distributed MariaDB Instance:

Define a MariaDB custom resource with spec.distributed set to true and reference the PodPlacementPolicy by name in spec.podTemplate.spec.podPlacementPolicy.name.

Example MariaDB YAML:
apiVersion: k8s.mariadb.com/v1alpha1
kind: MariaDB
metadata:
name: mariadb-distributed
namespace: default
spec:
distributed: true
podTemplate:
spec:
podPlacementPolicy:
name: mariadb-placement
rootPasswordSecretKeyRef:
name: mariadb-root-password
key: password
storage:
size: 1Gi
storageClassName: local-path
resources:
requests:
memory: 512Mi
cpu: 500m
limits:
memory: 1Gi
cpu: 1000m
myCnf: |
[mariadb]
bind-address=0.0.0.0
skip-log-bin
innodb_buffer_pool_size=512M
max_connections=50
galera:
enabled: true
affinity:
antiAffinityEnabled: true
service:
type: ClusterIP
ports:
- port: 3306
targetPort: 3306
protocol: TCP



Apply the MariaDB resource:

kubectl apply -f mariadb-distributed.yaml

### Verify the Deployment:

Check that MariaDB pods are scheduled on the target clusters according to the PodPlacementPolicy:

kubectl get pods -A -o wide

Verify Galera cluster replication across nodes:

kubectl exec -it mariadb-distributed-0 -n default -- mariadb -uroot -p -e "SHOW STATUS LIKE 'wsrep_cluster_size';"










