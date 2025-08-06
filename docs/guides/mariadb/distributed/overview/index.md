---
title: MariaDB Galera Cluster Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-distributed-overview
    name: Distributed MariaDB Overview
    parent: guides-mariadb-distributed
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

- Kubernetes node with minimum of 4 vCPUs and 16GB of RAM.

- Open Cluster Management (OCM): Clusteradm should be installed. See the [documentaion](https://open-cluster-management.io/docs/getting-started/quick-start/)

- kubectl: Installed and configured to interact with your Kubernetes clusters.

- Helm: Installed for deploying the KubeDB Operator and related resources.

- Persistent Storage: A storage class configured for persistent volumes (e.g., local-path or cloud provider-specific storage).

# Configuration Steps

Follow these steps to deploy a distributed MariaDB instance across multiple clusters:

### Set Up OCM for Multi-Cluster Management:

In this demonstration, we will utilize two clusters: `demo-controller` and `demo-worker`. The demo-controller will serve as the hub cluster, while the demo-worker will function as the spoke cluster. Additionally, the demo-controller hub cluster will also be configured to operate as a spoke cluster.
```bash
➤ kubectl config get-contexts
CURRENT   NAME              CLUSTER           AUTHINFO          NAMESPACE
*         demo-controller   demo-controller   demo-controller   
          demo-worker       demo-worker       demo-worker
```

Initialize the OCM hub cluster by executing the clusteradm init command.
```bash 

clusteradm init --wait
```

Check the deployment 
```bash
kubectl get ns
kubectl get pods -n open-cluster-management-hub
```

Obtain the token required to register the spoke cluster.

```bash

➤ clusteradm get token
token=eyJhbGciOiJSUzI1NiIsImtpZCI6Ikg2NlF2cDJVVFRyNUR5TTI3N0k4NG1aWVR3b015SnpRSjlLMTAzSkdIRGMifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzU0NDY2NTYyLCJpYXQiOjE3NTQ0NjI5NjIsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiOTJkNzFhNjMtMGVlYS00MDYzLWI0ZjEtZTk4ODRhYzAxNmEyIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJvcGVuLWNsdXN0ZXItbWFuYWdlbWVudCIsInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJhZ2VudC1yZWdpc3RyYXRpb24tYm9vdHN0cmFwIiwidWlkIjoiNDhmMjhkNDktMTM3OC00ZTFjLTk0NDMtNjQzNTMyOGZhNmJmIn19LCJuYmYiOjE3NTQ0NjI5NjIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpvcGVuLWNsdXN0ZXItbWFuYWdlbWVudDphZ2VudC1yZWdpc3RyYXRpb24tYm9vdHN0cmFwIn0.ANuDWLhvJ3mvxdSJjBQ4naBPgf8l--hr55JiQa2AXIeO8Ohb-nW9szNHp9KlmyKFBDcII7oS0QT2bt4Ldr-Vc79MLS_RnyhSJ6bS4_cJ_NfMSzpPUmpF5E3kkdBEmlVKdGfHYiVrXQbm7B_xCDkoSIs7avyMv6eZUzdljqp9ajGdQjRXmzYeqAHEObL5DaafZRJ8pk3rYdOfNSZRuzNZsgc7rOtFwE24LNormVwpLDdReAcEg-_pR1_55vlnfaiNJ6yCxKCRZ9S-Ht469U5DPS3DY0_qwR8SPc2vcds13gfMsJ04RSAIikHZaEZpp9QHHSH3HYXch8OFXtJ0Vs3Iig
please log on spoke and run:
clusteradm join --hub-token eyJhbGciOiJSUzI1NiIsImtpZCI6Ikg2NlF2cDJVVFRyNUR5TTI3N0k4NG1aWVR3b015SnpRSjlLMTAzSkdIRGMifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzU0NDY2NTYyLCJpYXQiOjE3NTQ0NjI5NjIsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiOTJkNzFhNjMtMGVlYS00MDYzLWI0ZjEtZTk4ODRhYzAxNmEyIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJvcGVuLWNsdXN0ZXItbWFuYWdlbWVudCIsInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJhZ2VudC1yZWdpc3RyYXRpb24tYm9vdHN0cmFwIiwidWlkIjoiNDhmMjhkNDktMTM3OC00ZTFjLTk0NDMtNjQzNTMyOGZhNmJmIn19LCJuYmYiOjE3NTQ0NjI5NjIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpvcGVuLWNsdXN0ZXItbWFuYWdlbWVudDphZ2VudC1yZWdpc3RyYXRpb24tYm9vdHN0cmFwIn0.ANuDWLhvJ3mvxdSJjBQ4naBPgf8l--hr55JiQa2AXIeO8Ohb-nW9szNHp9KlmyKFBDcII7oS0QT2bt4Ldr-Vc79MLS_RnyhSJ6bS4_cJ_NfMSzpPUmpF5E3kkdBEmlVKdGfHYiVrXQbm7B_xCDkoSIs7avyMv6eZUzdljqp9ajGdQjRXmzYeqAHEObL5DaafZRJ8pk3rYdOfNSZRuzNZsgc7rOtFwE24LNormVwpLDdReAcEg-_pR1_55vlnfaiNJ6yCxKCRZ9S-Ht469U5DPS3DY0_qwR8SPc2vcds13gfMsJ04RSAIikHZaEZpp9QHHSH3HYXch8OFXtJ0Vs3Iig --hub-apiserver https://10.2.0.56:6443 --cluster-name <cluster_name>
```
Execute the clusteradm join command on the demo-worker spoke cluster, replacing <cluster_name> with `demo-worker` in the provided command.

```bash
➤ kubectl config use-context demo-worker
Switched to context "demo-worker".

➤ clusteradm join --hub-token eyJhbGciOiJSUzI1NiIsImtpZCI6Ikg2NlF2cDJVVFRyNUR5TTI3N0k4NG1aWVR3b015SnpRSjlLMTAzSkdIRGMifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzU0NDY2NTYyLCJpYXQiOjE3NTQ0NjI5NjIsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiOTJkNzFhNjMtMGVlYS00MDYzLWI0ZjEtZTk4ODRhYzAxNmEyIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJvcGVuLWNsdXN0ZXItbWFuYWdlbWVudCIsInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJhZ2VudC1yZWdpc3RyYXRpb24tYm9vdHN0cmFwIiwidWlkIjoiNDhmMjhkNDktMTM3OC00ZTFjLTk0NDMtNjQzNTMyOGZhNmJmIn19LCJuYmYiOjE3NTQ0NjI5NjIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpvcGVuLWNsdXN0ZXItbWFuYWdlbWVudDphZ2VudC1yZWdpc3RyYXRpb24tYm9vdHN0cmFwIn0.ANuDWLhvJ3mvxdSJjBQ4naBPgf8l--hr55JiQa2AXIeO8Ohb-nW9szNHp9KlmyKFBDcII7oS0QT2bt4Ldr-Vc79MLS_RnyhSJ6bS4_cJ_NfMSzpPUmpF5E3kkdBEmlVKdGfHYiVrXQbm7B_xCDkoSIs7avyMv6eZUzdljqp9ajGdQjRXmzYeqAHEObL5DaafZRJ8pk3rYdOfNSZRuzNZsgc7rOtFwE24LNormVwpLDdReAcEg-_pR1_55vlnfaiNJ6yCxKCRZ9S-Ht469U5DPS3DY0_qwR8SPc2vcds13gfMsJ04RSAIikHZaEZpp9QHHSH3HYXch8OFXtJ0Vs3Iig --hub-apiserver https://10.2.0.56:6443 --cluster-name demo-worker

```

Now get back to the hub cluster('demo-controller') and accept the spoke cluster('demo-worker').
```bash
➤ kubectl config use-context demo-controller
Switched to context "demo-controller".

➤ clusteradm accept --clusters demo-worker
Starting approve csrs for the cluster demo-worker
CSR demo-worker-2p2pb approved
set hubAcceptsClient to true for managed cluster demo-worker

 Your managed cluster demo-worker has joined the Hub successfully. Visit https://open-cluster-management.io/scenarios or https://github.com/open-cluster-management-io/OCM/tree/main/solutions for next steps.

```

Verify the namespace on the hub cluster. A namespace corresponding to the cluster name (demo-worker) should have been created.

```bash

➤ kubectl get ns
NAME                          STATUS   AGE
default                       Active   99m
demo-worker                   Active   58s
kube-node-lease               Active   99m
kube-public                   Active   99m
kube-system                   Active   99m
open-cluster-management       Active   6m7s
open-cluster-management-hub   Active   5m32s

```
So, The `demo-worker` is successfully registered as spoke cluster.
Now register the 'demo-controller' as spoke cluster.

Run the following command on `demo-controller` cluster

```bash

clusteradm join --hub-token eyJhbGciOiJSUzI1NiIsImtpZCI6Ikg2NlF2cDJVVFRyNUR5TTI3N0k4NG1aWVR3b015SnpRSjlLMTAzSkdIRGMifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWwiLCJrM3MiXSwiZXhwIjoxNzU0NDY2NTYyLCJpYXQiOjE3NTQ0NjI5NjIsImlzcyI6Imh0dHBzOi8va3ViZXJuZXRlcy5kZWZhdWx0LnN2Yy5jbHVzdGVyLmxvY2FsIiwianRpIjoiOTJkNzFhNjMtMGVlYS00MDYzLWI0ZjEtZTk4ODRhYzAxNmEyIiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJvcGVuLWNsdXN0ZXItbWFuYWdlbWVudCIsInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJhZ2VudC1yZWdpc3RyYXRpb24tYm9vdHN0cmFwIiwidWlkIjoiNDhmMjhkNDktMTM3OC00ZTFjLTk0NDMtNjQzNTMyOGZhNmJmIn19LCJuYmYiOjE3NTQ0NjI5NjIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpvcGVuLWNsdXN0ZXItbWFuYWdlbWVudDphZ2VudC1yZWdpc3RyYXRpb24tYm9vdHN0cmFwIn0.ANuDWLhvJ3mvxdSJjBQ4naBPgf8l--hr55JiQa2AXIeO8Ohb-nW9szNHp9KlmyKFBDcII7oS0QT2bt4Ldr-Vc79MLS_RnyhSJ6bS4_cJ_NfMSzpPUmpF5E3kkdBEmlVKdGfHYiVrXQbm7B_xCDkoSIs7avyMv6eZUzdljqp9ajGdQjRXmzYeqAHEObL5DaafZRJ8pk3rYdOfNSZRuzNZsgc7rOtFwE24LNormVwpLDdReAcEg-_pR1_55vlnfaiNJ6yCxKCRZ9S-Ht469U5DPS3DY0_qwR8SPc2vcds13gfMsJ04RSAIikHZaEZpp9QHHSH3HYXch8OFXtJ0Vs3Iig --hub-apiserver https://10.2.0.56:6443 --cluster-name demo-controller

```

Accept the `demo-controller cluster`

```bash

clusteradm accept --clusters demo-controller
```

Check the namespce if  `demo-controller` is created or not

```bash

➤ kubectl get ns
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

### Configure KubeSlice for Network Connectivity:

You can follow the installation process described [here](https://kubeslice.io/documentation/open-source/1.4.0/install-kubeslice/yaml/yaml-controller-install)

We will deploy kubeslice controller on `demo-controller`

As MariaDB pods will be deployed both `demo-controller` and `demo-worker` cluster. So kubeslice worker controller will be deployed on both cluster.

Install kubeslice controller on `demo-controller`

Get the cluster info
```bash

➤ kubectl cluster-info
Kubernetes control plane is running at https://10.2.0.56:6443
CoreDNS is running at https://10.2.0.56:6443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.

```
create a controller.yaml file with the following content.
```yaml
kubeslice:
  controller:
    loglevel: info
    rbacResourcePrefix: kubeslice-rbac
    projectnsPrefix: kubeslice
    endpoint: https://10.2.0.56:6443

```
Deploy controller using helm chart
```bash

➤ helm install kubeslice-controller ./charts/kubeslice-controller  -f /home/arman/go/src/github.com/sheikh-arman/my-library/poc/kubeslice/demo/k3s/yaml-deploy/controller.yaml --namespace kubeslice-controller --create-namespace

NAME: kubeslice-controller
LAST DEPLOYED: Wed Aug  6 14:10:54 2025
NAMESPACE: kubeslice-controller
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
kubeslice controller installation successful!

```

```bash

helm install kubeslice-controller kubeslice/kubeslice-controller -f controller.yaml --namespace kubeslice-controller --create-namespace
```

Verify the installation
```bash

➤ kubectl get pods -n kubeslice-controller
NAME                                            READY   STATUS    RESTARTS   AGE
kubeslice-controller-manager-7fd756fff6-5kddd   2/2     Running   0          98s
```

Create project.yaml file with the following content and deploy it.
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

```bash
➤ kubectl apply -f project.yaml 
project.controller.kubeslice.io/demo-distributed-mariadb created
```

```bash

➤ kubectl get project -n kubeslice-controller
NAME                       AGE
demo-distributed-mariadb   31s
```

```bash

➤ kubectl get sa -n kubeslice-demo-distributed-mariadb
NAME                      SECRETS   AGE
default                   0         69s
kubeslice-rbac-rw-admin   1         68s
```

Set `kubeslice.io/node-type=gateway` labels on the nodes where the worker controller schedule. Do this for all worker cluster.

Run the following command on `demo-controller` cluster
```bash
➤ kubectl config use-context demo-controller
Switched to context "demo-controller".

➤ kubectl get nodes
NAME          STATUS   ROLES                  AGE    VERSION
demo-master   Ready    control-plane,master   143m   v1.33.3+k3s1

➤ kubectl label node demo-master kubeslice.io/node-type=gateway
node/demo-master labeled
```



run the following command on `demo-worker` cluster
```bash
➤ kubectl config use-context demo-worker
Switched to context "demo-worker".

➤ kubectl get nodes
NAME          STATUS   ROLES                  AGE    VERSION
demo-worker   Ready    control-plane,master   152m   v1.33.3+k3s1

➤ kubectl label node demo-worker kubeslice.io/node-type=gateway
node/demo-worker labeled
```

Create registration.yaml file with the following content and deploy it on `demo-controller`.

```bash


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

```

To get networkInterface run this command on your node.
```bash

ubuntu@demo-controller:~$ ip route get 8.8.8.8 | awk '{ print $5 }'
enp1s0

```

```bash
➤ kubectl config use-context demo-controller
Switched to context "demo-controller".

➤ kubectl apply -f registration.yaml 
cluster.controller.kubeslice.io/demo-controller created
cluster.controller.kubeslice.io/demo-worker created

```

Verify the clusters
```bash

➤ kubectl get clusters -n kubeslice-demo-distributed-mariadb 
NAME              AGE
demo-controller   61s
demo-worker       61s
```

### Lets register the kubeslice worker cluster to kubeslice controller

Get the secrets of project namespace `kubeslice-demo-distributed-mariadb`.
```bash
➤ kubectl get secrets -n kubeslice-demo-distributed-mariadb
NAME                                    TYPE                                  DATA   AGE
kubeslice-rbac-rw-admin                 kubernetes.io/service-account-token   3      17m
kubeslice-rbac-worker-demo-controller   kubernetes.io/service-account-token   5      5m8s
kubeslice-rbac-worker-demo-worker       kubernetes.io/service-account-token   5      5m8s
```

To get cluster endpoints, run the following command on `demo-worker` cluster . 
```bash

➤ kubectl cluster-info --context demo-worker --kubeconfig $HOME/.kube/config
Kubernetes control plane is running at https://10.2.0.60:6443
CoreDNS is running at https://10.2.0.60:6443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.

```
Here the endpoint is `https://10.2.0.60:6443`

Now run this script on `demo-controller` and create sliceoperator-worker.yaml with the following script output
```bash
kubectl config use-context demo-controller
Switched to context "demo-controller".

➤ sh secrets.sh kubeslice-rbac-worker-demo-worker demo-worker kubeslice-demo-distributed-mariadb enp1s0 https://10.2.0.60:6443

---
## Base64 encoded secret values from controller cluster
controllerSecret:
  namespace: a3ViZXNsaWNlLWRlbW8tZGlzdHJpYnV0ZWQtbWFyaWFkYg==
  endpoint: aHR0cHM6Ly8xMC4yLjAuNTY6NjQ0Mw==
  ca.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkekNDQVIyZ0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdGMyVnkKZG1WeUxXTmhRREUzTlRRME5UY3lOVEV3SGhjTk1qVXdPREEyTURVeE5ERXhXaGNOTXpVd09EQTBNRFV4TkRFeApXakFqTVNFd0h3WURWUVFEREJock0zTXRjMlZ5ZG1WeUxXTmhRREUzTlRRME5UY3lOVEV3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFRZ0d0VVc3bFA5aWZLajNzN01rZmFwU1NxZFptYXJaN0tsYjBzZmIxUksKU2tkMkR5YVB2Q01BQkZoZ2EvRlJSd3pIZGxCL3kxMHEvcUtGNm85VXBKMjdvMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVWlRNkxlekFRbERGSHF3SndxVHpFClpnNGxzTTh3Q2dZSUtvWkl6ajBFQXdJRFNBQXdSUUloQUlEUjlwZmZYcWFqd0VXd3U2cWpYVkFmNkNvVGZaRXEKa0NUN1dMOXZ1NjErQWlBOHhFTFVxSXNHSXc1eTlQM21rRnVHdDQzNGJDYkhraDF6OHJQT3RsZ2tDUT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
  token: ZXlKaGJHY2lPaUpTVXpJMU5pSXNJbXRwWkNJNklrZzJObEYyY0RKVlZGUnlOVVI1VFRJM04wazRORzFhV1ZSM2IwMTVTbnBSU2psTE1UQXpTa2RJUkdNaWZRLmV5SnBjM01pT2lKcmRXSmxjbTVsZEdWekwzTmxjblpwWTJWaFkyTnZkVzUwSWl3aWEzVmlaWEp1WlhSbGN5NXBieTl6WlhKMmFXTmxZV05qYjNWdWRDOXVZVzFsYzNCaFkyVWlPaUpyZFdKbGMyeHBZMlV0WkdWdGJ5MWthWE4wY21saWRYUmxaQzF0WVhKcFlXUmlJaXdpYTNWaVpYSnVaWFJsY3k1cGJ5OXpaWEoyYVdObFlXTmpiM1Z1ZEM5elpXTnlaWFF1Ym1GdFpTSTZJbXQxWW1WemJHbGpaUzF5WW1GakxYZHZjbXRsY2kxa1pXMXZMWGR2Y210bGNpSXNJbXQxWW1WeWJtVjBaWE11YVc4dmMyVnlkbWxqWldGalkyOTFiblF2YzJWeWRtbGpaUzFoWTJOdmRXNTBMbTVoYldVaU9pSnJkV0psYzJ4cFkyVXRjbUpoWXkxM2IzSnJaWEl0WkdWdGJ5MTNiM0pyWlhJaUxDSnJkV0psY201bGRHVnpMbWx2TDNObGNuWnBZMlZoWTJOdmRXNTBMM05sY25acFkyVXRZV05qYjNWdWRDNTFhV1FpT2lJMU9URmxNR0UyWlMwd01EWmpMVFJrWkdZdFlqVmhNQzA0TkRsbVpqQTBOalkzTTJFaUxDSnpkV0lpT2lKemVYTjBaVzA2YzJWeWRtbGpaV0ZqWTI5MWJuUTZhM1ZpWlhOc2FXTmxMV1JsYlc4dFpHbHpkSEpwWW5WMFpXUXRiV0Z5YVdGa1lqcHJkV0psYzJ4cFkyVXRjbUpoWXkxM2IzSnJaWEl0WkdWdGJ5MTNiM0pyWlhJaWZRLkhmaE5Ba1J5VmZCQ0pVV1lLTDN3MXhRdVVNakJzMUZFLVotelNRX3pPNTU3NWttUkczUzhyMUlZV1FJdzNOLWZydzhvUlpTRURwWWhsdnhVVTlxYUFsXzJuZkczVDE5OUZZc2hMNEt4U0JWZlhaT0puaGdaS1ZuOW40MTY2VHAwV0NRVTZ1WE1MZE80TTgwV21neGVORWRzVUtYU05iekVHRWRFY3oteWg3dkRteThQUmwyVFZjUFFTamJkQ0l0UWxBMTQ0STVraWMxSGRCNVdiTHR1WDZ0N2hIaHNOYzlNZkYwZXBBZkd4a0YwVDFTWHdNcDBHY3c5cW52OVEzSk91d2liemdXTGF0aUIyWHQtVEdnNWtSZTlCMzduelBlQi1uTmtRY3FGeWEzb2x2Q01QNDRMcW9CbGdqMTJlRGhjOFB6dEFseXRiQ3Z1S1l3Y0lsYk9PUQ==
cluster:
  name: demo-worker
  endpoint: https://10.2.0.60:6443
netop:
  networkInterface: enp1s0

```

Now install worker by the helm command on `demo-worker` cluster
```bash
➤ kubectl config use-context demo-worker
Switched to context "demo-worker".

➤ helm install kubeslice-worker ./charts/kubeslice-worker -f /home/arman/go/src/github.com/sheikh-arman/my-library/poc/kubeslice/demo/k3s/yaml-deploy/sliceoperator-worker.yaml -n kubeslice-system --create-namespace

NAME: kubeslice-worker
LAST DEPLOYED: Wed Aug  6 14:53:04 2025
NAMESPACE: kubeslice-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
Kubeslice Operator installation successful!

```

Now register `demo-controller` cluster as worker cluster

Run the following script on `demo-controller`, and create sliceoperator-controller.yaml  with the following script output
```bash
➤ kubectl config use-context demo-controller
Switched to context "demo-controller".

➤ sh secrets.sh kubeslice-rbac-worker-demo-controller demo-controller kubeslice-demo-distributed-mariadb enp1s0 https://10.2.0.56:6443

---
## Base64 encoded secret values from controller cluster
controllerSecret:
  namespace: a3ViZXNsaWNlLWRlbW8tZGlzdHJpYnV0ZWQtbWFyaWFkYg==
  endpoint: aHR0cHM6Ly8xMC4yLjAuNTY6NjQ0Mw==
  ca.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJkekNDQVIyZ0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdGMyVnkKZG1WeUxXTmhRREUzTlRRME5UY3lOVEV3SGhjTk1qVXdPREEyTURVeE5ERXhXaGNOTXpVd09EQTBNRFV4TkRFeApXakFqTVNFd0h3WURWUVFEREJock0zTXRjMlZ5ZG1WeUxXTmhRREUzTlRRME5UY3lOVEV3V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFRZ0d0VVc3bFA5aWZLajNzN01rZmFwU1NxZFptYXJaN0tsYjBzZmIxUksKU2tkMkR5YVB2Q01BQkZoZ2EvRlJSd3pIZGxCL3kxMHEvcUtGNm85VXBKMjdvMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVWlRNkxlekFRbERGSHF3SndxVHpFClpnNGxzTTh3Q2dZSUtvWkl6ajBFQXdJRFNBQXdSUUloQUlEUjlwZmZYcWFqd0VXd3U2cWpYVkFmNkNvVGZaRXEKa0NUN1dMOXZ1NjErQWlBOHhFTFVxSXNHSXc1eTlQM21rRnVHdDQzNGJDYkhraDF6OHJQT3RsZ2tDUT09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
  token: ZXlKaGJHY2lPaUpTVXpJMU5pSXNJbXRwWkNJNklrZzJObEYyY0RKVlZGUnlOVVI1VFRJM04wazRORzFhV1ZSM2IwMTVTbnBSU2psTE1UQXpTa2RJUkdNaWZRLmV5SnBjM01pT2lKcmRXSmxjbTVsZEdWekwzTmxjblpwWTJWaFkyTnZkVzUwSWl3aWEzVmlaWEp1WlhSbGN5NXBieTl6WlhKMmFXTmxZV05qYjNWdWRDOXVZVzFsYzNCaFkyVWlPaUpyZFdKbGMyeHBZMlV0WkdWdGJ5MWthWE4wY21saWRYUmxaQzF0WVhKcFlXUmlJaXdpYTNWaVpYSnVaWFJsY3k1cGJ5OXpaWEoyYVdObFlXTmpiM1Z1ZEM5elpXTnlaWFF1Ym1GdFpTSTZJbXQxWW1WemJHbGpaUzF5WW1GakxYZHZjbXRsY2kxa1pXMXZMV052Ym5SeWIyeHNaWElpTENKcmRXSmxjbTVsZEdWekxtbHZMM05sY25acFkyVmhZMk52ZFc1MEwzTmxjblpwWTJVdFlXTmpiM1Z1ZEM1dVlXMWxJam9pYTNWaVpYTnNhV05sTFhKaVlXTXRkMjl5YTJWeUxXUmxiVzh0WTI5dWRISnZiR3hsY2lJc0ltdDFZbVZ5Ym1WMFpYTXVhVzh2YzJWeWRtbGpaV0ZqWTI5MWJuUXZjMlZ5ZG1salpTMWhZMk52ZFc1MExuVnBaQ0k2SWpJeE9UTmtNamd4TFRRMVl6Z3RORGcxTlMwNU56QXdMVEF4T1RJM04yRmhaakpqWlNJc0luTjFZaUk2SW5ONWMzUmxiVHB6WlhKMmFXTmxZV05qYjNWdWREcHJkV0psYzJ4cFkyVXRaR1Z0Ynkxa2FYTjBjbWxpZFhSbFpDMXRZWEpwWVdSaU9tdDFZbVZ6YkdsalpTMXlZbUZqTFhkdmNtdGxjaTFrWlcxdkxXTnZiblJ5YjJ4c1pYSWlmUS5ZU1c3Sjl1N2NtUjVBTWhmMHY1ZVFtdmVpTUh5VlVJelBfTXBqV1NfY0pwcG0yWXJUOWZSNHRqejR4MGx3OHlSR2hrb1JuUGJXLXVMU1pDOExWNm0zX2Zxa0dHN3l3MFhQM3hCOWJsOEdxaGNVaG1rd0JCWHEtbEEybkZ6T0MtVTRqMEhOMnlRS0EtdzJQRUE4dnFGUlByUWVLckJwN0pLRXFUOFExbHMtTUNiUWNYZHJ0UDVNN255QXhSTHVsNnZJRlFkM0cwZU1RMzMwU3JNVlVsV2FaZ0NSbmJHZ2FGbnlwdks2RnlLZG9XUzg2ZzR6Sk1hZ0NRY2N3QnNibEN2anJEUHR4X1h6cVo2RWwwblpYanZTTFEyOGJGdU5DdmJ1QlI1T1JGSzI2aVZyQ3MxNnJYUlpSM2NTQXh3MTN0NDRGQVBraWg3ZlRJUEV6bjhnN0RTV0E=
cluster:
  name: demo-controller
  endpoint: https://10.2.0.56:6443
netop:
  networkInterface: enp1s0

```

```bash

➤ helm install kubeslice-worker ./charts/kubeslice-worker -f /home/arman/go/src/github.com/sheikh-arman/my-library/poc/kubeslice/demo/k3s/yaml-deploy/sliceoperator-controller.yaml -n kubeslice-system --create-namespace

NAME: kubeslice-worker
LAST DEPLOYED: Wed Aug  6 15:08:26 2025
NAMESPACE: kubeslice-system
STATUS: deployed
REVISION: 1
TEST SUITE: None
NOTES:
Kubeslice Operator installation successful!

```
```bash

➤ helm install kubeslice-worker kubeslice/kubeslice-worker -f /home/arman/go/src/github.com/sheikh-arman/my-library/poc/kubeslice/demo/k3s/yaml-deploy/sliceoperator-controller.yaml -n kubeslice-system --create-namespace

```

```bash

➤ kubectl get pods -n kubeslice-system 
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

We have successfully registered worker.

Now create sliceconfig on `demo-controller` to onboard namespace

create sliceconfig.yaml filw with the following contents.
```yaml
apiVersion: controller.kubeslice.io/v1alpha1
kind: SliceConfig
metadata:
  name: demo-slice
  namespace: kubeslice-demo-distributed-mariadb
spec:
  sliceSubnet: 10.1.0.0/16
  maxClusters: 16       #Ex: 5. By default, the maxClusters value is set to 16
  sliceType: Application
  sliceGatewayProvider:
#    sliceGatewayType: OpenVPN
    sliceGatewayType: Wireguard
    sliceCaType: Local
  sliceIpamType: Local
  rotationInterval: 60    # If not provided, by default key rotation interval is 30 days
  vpnConfig:
     cipher: AES-128-CBC       # If not provided, by default cipher is AES-256-CBC
  clusters:
    - demo-controller
    - demo-worker
  qosProfileDetails:
    queueType: HTB
    priority: 1                      #keep integer values from 0 to 3
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
     - namespace: kubeops
       clusters:
         - '*'
     - namespace: monitoring
       clusters:
         - '*'
    isolationEnabled: false                   #make this true in case you want to enable isolation
    allowedNamespaces:
     - namespace: kube-system
       clusters:
       - '*'

```

```bash

➤ kubectl apply -f sliceconfig.yaml 
sliceconfig.controller.kubeslice.io/demo-slice created
```

### Install the KubeDB Operator:

Install KubeDB operator

```bash
You can follow the instruction [here](https://kubedb.com/docs/v2025.6.30/setup/install/kubedb/)
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
    --version v2025.7.30-rc.0 \
    --namespace kubedb --create-namespace \
    --set-file global.license=$HOME/Downloads/kubedb-license-b8a604fd-bc99-430c-a5fb-abbe4b0d989e.txt \
    --wait --burst-limit=10000 --debug
```

### Define a PodPlacementPolicy:
Create a PodPlacementPolicy custom resource in the hub cluster to specify which clusters should host MariaDB pods. 

Create pod-placement-policy.yaml file with the following yaml and deploy it.
```yaml

apiVersion: apps.k8s.appscode.com/v1
kind: PlacementPolicy
metadata:
  labels:
    app.kubernetes.io/managed-by: Helm
  name: distributed-mariadb
spec:
  nodeSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway
  ocm:
    distributionRules:
      - clusterName: demo-controller
        replicas:
          - 0
          - 2
          - 4
      - clusterName: demo-worker
        replicas:
          - 1
          - 3
          - 5
    sliceName: demo-slice
  zoneSpreadConstraint:
    maxSkew: 1
    whenUnsatisfiable: ScheduleAnyway

```

```bash

kubectl apply -f pod-placement-policy.yaml --context demo-controller --kubeconfig $HOME/.kube/config
```



### Create a Distributed MariaDB Instance:

Define a MariaDB custom resource with spec.distributed set to true and reference the PodPlacementPolicy by name in spec.podTemplate.spec.podPlacementPolicy.name.

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

Apply the MariaDB resource on `demo-controller`
```bash

➤ kubectl apply -f mariadb.yaml --context demo-controller --kubeconfig $HOME/.kube/config
mariadb.kubedb.com/mariadb created
```

### Verify the Deployment:

Check the mariadb resource and pod  on `demo-controller`
```bash

➤ kubectl get md,pods,secret -n demo --context demo-controller --kubeconfig $HOME/.kube/config
NAME                         VERSION   STATUS   AGE
mariadb.kubedb.com/mariadb   11.5.2    Ready    99s

NAME            READY   STATUS    RESTARTS   AGE
pod/mariadb-0   3/3     Running   0          95s
pod/mariadb-2   3/3     Running   0          95s

NAME                  TYPE                       DATA   AGE
secret/mariadb-auth   kubernetes.io/basic-auth   2      95s
```

Verify pod,secret on `demo-worker`
```bash

➤ kubectl get pods,secrets -n demo --context demo-worker --kubeconfig $HOME/.kube/config
NAME        READY   STATUS    RESTARTS   AGE
mariadb-1   3/3     Running   0          95s

NAME                  TYPE                       DATA   AGE
secret/mariadb-auth   kubernetes.io/basic-auth   2      95s

```

So distributed MariaDB is up and running.



