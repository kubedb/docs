---
title: SingleStore CRD
menu:
  docs_{{ .version }}:
    identifier: sdb-singlestore-concepts
    name: SingleStore
    parent: sdb-concepts-singlestore
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---
# SingleStore

## What is MongoDB

`MongoDB` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [SingleStore](https://www.singlestore.com/) in a Kubernetes native way. You only need to describe the desired database configuration in a MongoDB object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## SingleStore Spec

As with all other Kubernetes objects, a SingleStore needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section. Below is an example SingleStore object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-sample
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 2
      configSecret:
        name: sdb-configuration
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "4Gi"
                cpu: "1000m"
              requests:
                memory: "2Gi"
                cpu: "500m"
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 3
      configSecret:
        name: sdb-configuration
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "5Gi"
                  cpu: "1100m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"                     
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 40Gi
  storageType: Durable
  licenseSecret:
    name: license-secret
  authSecret:
    name: given-secret
  init:
    script:
      configMap:
        name: sdb-init-script
  monitor:
    agent: prometheus.io/operator
    prometheus:
      exporter:
        port: 9104
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  deletionPolicy: WipeOut
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: sdb-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  serviceTemplates:
  - alias: primary
    metadata:
      annotations:
        passMe: ToService
    spec:
      type: NodePort
      ports:
      - name:  http
        port:  9200
```
