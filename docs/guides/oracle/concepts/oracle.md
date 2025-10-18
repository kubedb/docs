---
title: Oracle CRD
menu:
  docs_{{ .version }}:
    identifier: orc-concepts
    name: Oracle
    parent: orc-concepts-oracle
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Oracle CRD

## What is Oracle CRD?

`Oracle` is a Kubernetes Custom Resource Definition (CRD) maintained by KubeDB. It provides a declarative configuration for Oracle database instances in your Kubernetes cluster, enabling you to manage Oracle databases using native Kubernetes tools and practices.

## Key Features

- **Native Kubernetes Integration**: Manage Oracle databases using kubectl and Kubernetes APIs
- **Flexible Deployment Modes**: Support for both Standalone and DataGuard configurations
- **High Availability**: Built-in support for Oracle DataGuard for disaster recovery
- **Resource Management**: Fine-grained control over CPU, memory, and storage resources
- **Security Controls**: Built-in security context configuration for enhanced security
- **Automated Operations**: Simplified deployment and management of complex Oracle setups

## Deployment Modes

KubeDB Oracle supports two primary deployment modes:

### 1. Standalone Mode
- Single instance Oracle database deployment
- Suitable for development and testing environments
- Simplified configuration and management

### 2. DataGuard Mode
- High availability configuration with primary and standby databases
- Built-in disaster recovery capabilities
- Support for synchronous and asynchronous replication
- Observer process for automated failover

## Oracle CRD Specification

Like any official Kubernetes resource, an `Oracle` object has standard fields:
- `apiVersion`: Specifies the API version (kubedb.com/v1alpha2)
- `kind`: Defines the resource type (Oracle)
- `metadata`: Contains name, namespace, and other metadata
- `spec`: Defines the desired state of your Oracle instance

### Sample Oracle Configuration

#### 1. Standalone Mode Configuration

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle
  namespace: demo
spec:
  version: 21.3.0
  mode: Standalone
  edition: enterprise
  replicas: 1
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 30Gi
  podTemplate:
    spec:
      securityContext:
        fsGroup: 54321
        runAsGroup: 54321
        runAsUser: 54321
      containers:
        - name: oracle
          resources:
            limits:
              cpu: "4"
              memory: 10Gi
            requests:
              cpu: "2"
              memory: 3Gi
  deletionPolicy: Delete
```
```shell
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/quickstart/standalone.yaml
oracle.kubedb.com/oracle created

```
#### 2. DataGuard Mode Configuration

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sample
  namespace: demo
spec:
  version: 21.3.0
  mode: DataGuard
  edition: enterprise
  replicas: 3
  
  # DataGuard Specific Configuration
  dataGuard:
    protectionMode: MaximumProtection
    standbyType: PHYSICAL
    syncMode: SYNC
    applyLagThreshold: 0
    transportLagThreshold: 0
    fastStartFailover:
      fastStartFailoverThreshold: 15
    observer:
      podTemplate:
        spec:
          containers:
          - name: observer
            resources:
              limits:
                cpu: "1"
                memory: 2Gi
              requests:
                cpu: 500m
                memory: 2Gi
      storage:
        resources:
          requests:
            storage: 1Gi
            
  # Storage Configuration
  storageType: Durable
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 30Gi
        
  # Pod Configuration
  podTemplate:
    spec:
      securityContext:
        fsGroup: 54321
        runAsGroup: 54321
        runAsUser: 54321
      serviceAccountName: oracle-sample
      containers:
      - name: oracle
        resources:
          limits:
            cpu: "4"
            memory: 10Gi
          requests:
            cpu: "1500m"
            memory: 4Gi
      - name: oracle-coordinator
        resources:
          limits:
            memory: 256Mi
          requests:
            cpu: 200m
            memory: 256Mi
  
  deletionPolicy: Delete
```

### Configuration Parameters Explained

#### Core Parameters
- `version`: Oracle database version (e.g., 21.3.0)
- `mode`: Deployment mode (Standalone or DataGuard)
- `edition`: Oracle edition (enterprise, standard)
- `replicas`: Number of database instances

#### DataGuard Specific Parameters
- `protectionMode`: Defines the data protection mode (MaximumProtection, MaximumAvailability, MaximumPerformance)
- `standbyType`: Type of standby database (PHYSICAL)
- `syncMode`: Synchronization mode between primary and standby (SYNC, ASYNC)
- `applyLagThreshold`: Maximum acceptable lag in applying changes
- `transportLagThreshold`: Maximum acceptable transport lag
- `fastStartFailover`: Configuration for automated failover
- `observer`: Configuration for the DataGuard observer process

#### Storage Configuration
- `storageType`: Type of storage (Durable for persistent storage)
- `storage`: Kubernetes PVC configuration
  - `accessModes`: Volume mount access modes
  - `resources`: Storage resource requests and limits

#### Pod Configuration
- `podTemplate`: Pod customization options
  - `securityContext`: Security settings
  - `containers`: Resource configurations for various containers
  - `serviceAccountName`: Kubernetes service account to use

#### Lifecycle Management
- `deletionPolicy`: Resource cleanup policy on deletion

## Best Practices

1. **Deployment Mode Selection**:
   - Use Standalone mode for development/testing
   - Choose DataGuard mode for production environments requiring high availability

2. **Resource Planning**: 
   - Allocate sufficient resources based on workload
   - Consider overhead for DataGuard synchronization
   - Plan observer resources carefully

3. **Storage Configuration**:
   - Use enterprise-grade storage for production
   - Configure appropriate storage class
   - Plan for backup storage

4. **Security**:
   - Implement proper security contexts
   - Use service accounts with minimum required permissions
   - Regular security patches and updates

5. **Monitoring and Maintenance**:
   - Monitor replication lag in DataGuard setups
   - Regular backup testing
   - Keep track of resource usage patterns

6. **High Availability**:
   - Configure appropriate FastStart Failover thresholds
   - Regular failover testing
   - Monitor observer health
